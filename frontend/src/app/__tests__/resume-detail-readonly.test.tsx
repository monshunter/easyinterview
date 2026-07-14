// @vitest-environment jsdom
import { StrictMode } from "react";
import { describe, expect, it } from "vitest";
import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { EasyInterviewClient } from "../../api/generated/client";
import type { Resume } from "../../api/generated/types";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../api/mockTransport";
import { App } from "../App";

import getRuntimeConfigFixture from "../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../../openapi/fixtures/Auth/getMe.json";
import getResumeFixture from "../../../../openapi/fixtures/Resumes/getResume.json";

const FIXTURES = [
  getRuntimeConfigFixture,
  getMeFixture,
  getResumeFixture,
];

const RESUME_ID = getResumeFixture.scenarios.default.response.body.id;

function buildClient(scenario: string): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry(FIXTURES),
      { scenario },
    ),
  });
}

interface DetailTransportStats {
  count: number;
  inFlight: number;
  maxInFlight: number;
}

function buildSequenceClient(
  sequence: Array<Resume | Error>,
  stats: DetailTransportStats,
): EasyInterviewClient {
  const fixtureFetch = createFixtureBackedFetch(
    createFixtureRegistry(FIXTURES),
    { scenario: "default" },
  );
  return new EasyInterviewClient({
    fetch: async (input, init) => {
      const url = new URL(String(input), "http://fixture.local");
      if (
        (init?.method ?? "GET").toUpperCase() !== "GET" ||
        url.pathname !== `/api/v1/resumes/${RESUME_ID}`
      ) {
        return fixtureFetch(input, init);
      }

      const index = stats.count;
      stats.count += 1;
      stats.inFlight += 1;
      stats.maxInFlight = Math.max(stats.maxInFlight, stats.inFlight);
      try {
        await Promise.resolve();
        const value = sequence[Math.min(index, sequence.length - 1)];
        if (value instanceof Error) throw value;
        return new Response(JSON.stringify(value), {
          status: 200,
          headers: { "Content-Type": "application/json" },
        });
      } finally {
        stats.inFlight -= 1;
      }
    },
  });
}

function renderDetailWithClient(
  client: EasyInterviewClient,
  resumeId: string,
  params: Record<string, string> = {},
  options?: { strict?: boolean },
) {
  const app = (
    <App
      client={client}
      requestOptions={{
        getMe: { headers: { Prefer: "example=authenticated" } },
      }}
      initialRoute={{
        name: "resume_versions",
        params: { resumeId, ...params },
      }}
    />
  );
  return render(options?.strict ? <StrictMode>{app}</StrictMode> : app);
}

function renderDetail(
  scenario: string,
  resumeId: string,
  params: Record<string, string> = {},
) {
  return renderDetailWithClient(buildClient(scenario), resumeId, params);
}

describe("resume detail read-only view and 404 fallback", () => {
  it("issues one ready detail transport on a StrictMode mount", async () => {
    const stats = { count: 0, inFlight: 0, maxInFlight: 0 };
    const client = buildSequenceClient(
      [getResumeFixture.scenarios.default.response.body as Resume],
      stats,
    );

    renderDetailWithClient(client, RESUME_ID, {}, { strict: true });

    await screen.findByTestId("resume-detail-crumb");
    expect(stats).toMatchObject({ count: 1, maxInFlight: 1 });
    console.info(
      "resume ready detail transport PASS initial=1 maxInFlight=1",
    );
  });

  it("evicts a rejected ready-detail request and retries with one new transport", async () => {
    const stats = { count: 0, inFlight: 0, maxInFlight: 0 };
    const client = buildSequenceClient(
      [
        new TypeError("temporary detail transport failure"),
        getResumeFixture.scenarios.default.response.body as Resume,
      ],
      stats,
    );

    renderDetailWithClient(client, RESUME_ID, {}, { strict: true });

    const retry = await screen.findByTestId("resume-detail-retry");
    expect(stats).toMatchObject({ count: 1, maxInFlight: 1 });
    await userEvent.click(retry);
    await screen.findByTestId("resume-detail-crumb");
    expect(stats).toMatchObject({ count: 2, maxInFlight: 1 });
    console.info(
      "resume detail rejection retry transport PASS initialRejected=1 retrySucceeded=2 maxInFlight=1",
    );
  });

  it("renders the resume itself and exposes no tab, export, copy, edit, rewrite, or original-preview controls", async () => {
    renderDetail("default", RESUME_ID);

    await waitFor(() => {
      expect(screen.getByTestId("resume-detail-crumb")).toBeInTheDocument();
    });
    expect(screen.getByTestId("resume-detail-preview-content")).toHaveTextContent(
      "Original resume parsed text snapshot",
    );
    expect(screen.getByTestId("resume-detail-preview-content")).not.toHaveTextContent(
      "Senior frontend engineer for platform-heavy product teams",
    );
    expect(
      screen.queryByTestId("resume-detail-branch-graph"),
    ).not.toBeInTheDocument();
    expect(screen.queryByRole("tablist")).not.toBeInTheDocument();
    for (const forbidden of [
      "resume-detail-tab-preview",
      "resume-detail-tab-rewrites",
      "resume-detail-tab-edit",
      "resume-detail-header-actions",
      "resume-detail-export-pdf",
      "resume-detail-copy-text",
      "resume-detail-view-original",
      "resume-detail-original-modal",
      "resume-rewrites-tab",
      "resume-edit-tab",
    ]) {
      expect(screen.queryByTestId(forbidden)).not.toBeInTheDocument();
    }
  });

  it("out-of-scope ?tab=rewrites is ignored and cannot activate a rewrite surface", async () => {
    renderDetail("default", RESUME_ID, { tab: "rewrites" });

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-detail-preview-content"),
      ).toBeInTheDocument();
    });
    expect(screen.queryByTestId("resume-rewrites-tab")).not.toBeInTheDocument();
    expect(screen.queryByTestId("resume-detail-tab-rewrites")).not.toBeInTheDocument();
    expect(screen.getByTestId("resume-workshop-detail")).not.toHaveAttribute(
      "data-tab",
    );
  });

  it("does not write detail-only action state into localStorage", async () => {
    renderDetail("default", RESUME_ID);

    await waitFor(() => {
      expect(screen.getByTestId("resume-detail-crumb")).toBeInTheDocument();
    });
    const offenders: string[] = [];
    for (let i = 0; i < window.localStorage.length; i++) {
      const key = window.localStorage.key(i);
      if (key && /resume|export|pdf|rewrite|edit|original/i.test(key)) {
        offenders.push(key);
      }
    }
    expect(offenders).toEqual([]);
  });

  it("polls pending PDF upload detail until the source page stack and LLM displayName are shown", async () => {
    const queued: Resume = {
      ...(getResumeFixture.scenarios.default.response.body as Resume),
      id: RESUME_ID,
      title: "谭章毓简历-后端工程师AI.pdf",
      displayName: "",
      sourceType: "upload",
      parseStatus: "queued",
      originalText: null,
      parsedTextSnapshot: null,
      parsedSummary: null,
      structuredProfile: {},
    };
    const ready: Resume = {
      ...queued,
      parseStatus: "ready",
      displayName: "谭章毓 - 后端工程师 AI",
      parsedSummary: { headline: "后端工程师 AI" },
      parsedTextSnapshot:
        "谭章毓\n后端工程师 AI\nservice-registry-operator / korder / ohmykube",
    };
    const stats = { count: 0, inFlight: 0, maxInFlight: 0 };
    const client = buildSequenceClient([queued, ready], stats);

    renderDetailWithClient(client, RESUME_ID, {}, { strict: true });

    await waitFor(
      () => {
        expect(
          screen.getAllByRole("heading", { name: "谭章毓 - 后端工程师 AI" })
            .length,
        ).toBeGreaterThanOrEqual(1);
      },
      { timeout: 2000 },
    );
    expect(stats).toMatchObject({ count: 2, maxInFlight: 1 });
    const stack = screen.getByTestId("resume-detail-pdf-preview-stack");
    expect(stack).toHaveAttribute(
      "data-source-url",
      "/api/v1/resumes/01918fa0-0000-7000-8000-000000001000/source",
    );
    expect(screen.getByTestId("resume-detail-preview-content")).not.toHaveTextContent(
      "service-registry-operator / korder / ohmykube",
    );
    expect(document.querySelector("object, iframe, embed")).toBeNull();
    expect(
      screen.queryByRole("heading", { name: "谭章毓简历-后端工程师AI.pdf" }),
    ).not.toBeInTheDocument();
    console.info(
      "resume pending serial poll transport PASS initial=1 poll=2 maxInFlight=1",
    );
  });

  it("does not poll again when a PDF upload has failed but the source page stack and displayName are available", async () => {
    const failed: Resume = {
      ...(getResumeFixture.scenarios.default.response.body as Resume),
      id: RESUME_ID,
      title: "谭章毓简历-后端工程师AI.pdf",
      displayName: "谭章毓 - AI Infra DevOps 平台工程师",
      sourceType: "upload",
      parseStatus: "failed",
      originalText: null,
      parsedTextSnapshot:
        "谭章毓 | AI / Infra / DevOps 平台工程师\n核心能力：AI Workflow、Kubernetes、GitOps",
      parsedSummary: null,
      structuredProfile: {},
    };
    const stats = { count: 0, inFlight: 0, maxInFlight: 0 };
    const client = buildSequenceClient([failed], stats);

    renderDetailWithClient(client, RESUME_ID, {}, { strict: true });

    await waitFor(() => {
      expect(screen.getByTestId("resume-detail-pdf-preview-stack")).toBeInTheDocument();
    });
    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 350));
    });

    expect(stats).toMatchObject({ count: 1, maxInFlight: 1 });
    expect(screen.getByTestId("resume-detail-preview-content")).not.toHaveTextContent(
      "AI Workflow",
    );
    expect(document.querySelector("object, iframe, embed")).toBeNull();
    expect(
      screen.getAllByRole("heading", {
        name: "谭章毓 - AI Infra DevOps 平台工程师",
      }).length,
    ).toBeGreaterThanOrEqual(1);
  });

  it("non-existent resumeId returns 404 without echoing fixture error.code", async () => {
    renderDetail("not-found", "ffffffff-0000-7000-8000-00000000ff04");

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-detail-not-found"),
      ).toBeInTheDocument();
    });
    const card = screen.getByTestId("resume-detail-not-found");
    expect(card).not.toHaveTextContent("RESOURCE_NOT_FOUND");
    expect(
      screen.getByTestId("resume-detail-not-found-back"),
    ).toBeInTheDocument();
  });
});
