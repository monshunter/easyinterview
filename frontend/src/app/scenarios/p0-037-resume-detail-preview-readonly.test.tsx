// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";

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

function renderDetail(
  scenario: string,
  resumeId: string,
  params: Record<string, string> = {},
) {
  return render(
    <App
      client={buildClient(scenario)}
      requestOptions={{
        getMe: { headers: { Prefer: "example=authenticated" } },
      }}
      initialRoute={{
        name: "resume_versions",
        params: { resumeId, ...params },
      }}
    />,
  );
}

describe("E2E.P0.037 resume detail read-only view + 404 fallback", () => {
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
    const client = buildClient("default");
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
    const getResumeSpy = vi
      .spyOn(client, "getResume")
      .mockResolvedValueOnce(queued)
      .mockResolvedValueOnce(ready);

    render(
      <App
        client={client}
        requestOptions={{
          getMe: { headers: { Prefer: "example=authenticated" } },
        }}
        initialRoute={{
          name: "resume_versions",
          params: { resumeId: RESUME_ID },
        }}
      />,
    );

    await waitFor(
      () => {
        expect(getResumeSpy).toHaveBeenCalledTimes(2);
      },
      { timeout: 2000 },
    );
    expect(
      screen.getAllByRole("heading", { name: "谭章毓 - 后端工程师 AI" })
        .length,
    ).toBeGreaterThanOrEqual(1);
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
  });

  it("does not poll again when a PDF upload has failed but the source page stack and displayName are available", async () => {
    const client = buildClient("default");
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
    const getResumeSpy = vi.spyOn(client, "getResume").mockResolvedValue(failed);

    render(
      <App
        client={client}
        requestOptions={{
          getMe: { headers: { Prefer: "example=authenticated" } },
        }}
        initialRoute={{
          name: "resume_versions",
          params: { resumeId: RESUME_ID },
        }}
      />,
    );

    await waitFor(() => {
      expect(screen.getByTestId("resume-detail-pdf-preview-stack")).toBeInTheDocument();
    });
    await new Promise((resolve) => setTimeout(resolve, 350));

    expect(getResumeSpy).toHaveBeenCalledTimes(1);
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
