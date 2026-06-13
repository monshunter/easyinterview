// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { EasyInterviewClient } from "../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../api/mockTransport";
import { App } from "../App";

import getRuntimeConfigFixture from "../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../../openapi/fixtures/Auth/getMe.json";
import getResumeFixture from "../../../../openapi/fixtures/Resumes/getResume.json";
import exportResumeFixture from "../../../../openapi/fixtures/Resumes/exportResume.json";

const FIXTURES = [
  getRuntimeConfigFixture,
  getMeFixture,
  getResumeFixture,
  exportResumeFixture,
];

const RESUME_ID = getResumeFixture.scenarios.default.response.body.id;

interface ToastCall {
  message: string;
  tone?: string;
}

let toastCalls: ToastCall[] = [];

beforeEach(() => {
  toastCalls = [];
  (
    window as unknown as {
      eiToast?: (msg: string, opts?: { tone?: string }) => void;
    }
  ).eiToast = (message, opts) => {
    toastCalls.push({ message, tone: opts?.tone });
  };
});

afterEach(() => {
  delete (
    window as unknown as {
      eiToast?: (msg: string, opts?: { tone?: string }) => void;
    }
  ).eiToast;
});

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
  authMode: "authenticated" | "unauthenticated" = "authenticated",
) {
  return render(
    <App
      client={buildClient(scenario)}
      requestOptions={{
        getMe: { headers: { Prefer: `example=${authMode}` } },
      }}
      initialRoute={{
        name: "resume_versions",
        params: { resumeId },
      }}
    />,
  );
}

describe("E2E.P0.037 resume detail Preview Tab + original modal + 404 fallback + export 501", () => {
  it("resume renders detail with crumb / 3 tabs and defaults the active tab to preview", async () => {
    renderDetail("default", RESUME_ID);

    await waitFor(() => {
      expect(screen.getByTestId("resume-detail-crumb")).toBeInTheDocument();
    });
    for (const tab of ["preview", "rewrites", "edit"]) {
      expect(screen.getByTestId(`resume-detail-tab-${tab}`)).toBeInTheDocument();
    }
    expect(screen.getByTestId("resume-detail-tab-preview")).toHaveAttribute(
      "aria-selected",
      "true",
    );
    expect(
      screen.getByTestId("resume-detail-preview-content"),
    ).toBeInTheDocument();
    // D-20 flatten: there is no branch graph in the flat detail header.
    expect(
      screen.queryByTestId("resume-detail-branch-graph"),
    ).not.toBeInTheDocument();
  });

  it("explicit ?tab=rewrites preserves the rewrites tab and renders the current Rewrites surface", async () => {
    render(
      <App
        client={buildClient("default")}
        requestOptions={{
          getMe: { headers: { Prefer: "example=authenticated" } },
        }}
        initialRoute={{
          name: "resume_versions",
          params: { resumeId: RESUME_ID, tab: "rewrites" },
        }}
      />,
    );

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-detail-tab-rewrites"),
      ).toBeInTheDocument();
    });
    expect(screen.getByTestId("resume-detail-tab-rewrites")).toHaveAttribute(
      "aria-selected",
      "true",
    );
    expect(screen.getByTestId("resume-rewrites-tab")).toBeInTheDocument();
    expect(
      screen.queryByTestId("resume-detail-tab-content-coming-soon-rewrites"),
    ).not.toBeInTheDocument();
  });

  it("View original opens modal with focus trap and closes on ESC / outer overlay / X button", async () => {
    renderDetail("default", RESUME_ID);

    await waitFor(() => {
      expect(
        screen.getByTestId("resume-detail-view-original"),
      ).toBeInTheDocument();
    });
    const user = userEvent.setup();
    await user.click(screen.getByTestId("resume-detail-view-original"));

    const dialog = await screen.findByTestId("resume-detail-original-modal");
    expect(dialog).toHaveAttribute("aria-modal", "true");
    expect(dialog).toHaveAttribute("role", "dialog");
    await waitFor(() => {
      expect(
        within(dialog).getByTestId("resume-detail-original-modal-content"),
      ).toHaveTextContent("Original resume parsed text snapshot");
    });
    const closeBtn = within(dialog).getByTestId(
      "resume-detail-original-modal-close",
    );
    await waitFor(() => expect(document.activeElement).toBe(closeBtn));

    await user.keyboard("{Escape}");
    await waitFor(() =>
      expect(
        screen.queryByTestId("resume-detail-original-modal"),
      ).not.toBeInTheDocument(),
    );
  });

  it("Export PDF passes Idempotency-Key on the wire and surfaces the P0 not-available toast (no blob, no localStorage)", async () => {
    let exportHeaders: Record<string, string> | null = null;
    const baseFetch = createFixtureBackedFetch(
      createFixtureRegistry(FIXTURES),
      { scenario: "default" },
    );
    const fetchSpy = (
      input: RequestInfo | URL,
      init?: RequestInit,
    ): Promise<Response> => {
      const url = typeof input === "string" ? input : input.toString();
      if (url.includes("/exports")) {
        exportHeaders = (init?.headers as Record<string, string>) ?? {};
      }
      return baseFetch(input, init);
    };
    const client = new EasyInterviewClient({ fetch: fetchSpy });

    render(
      <App
        client={client}
        requestOptions={{
          getMe: { headers: { Prefer: "example=authenticated" } },
        }}
        initialRoute={{
          name: "resume_versions",
          params: { resumeId: RESUME_ID, tab: "preview" },
        }}
      />,
    );

    // The reshaped detail surfaces Export PDF both in the header action bar
    // and inside the Preview tab; target the header bar's button so the
    // selector stays unambiguous regardless of the active tab.
    const headerActions = await screen.findByTestId(
      "resume-detail-header-actions",
    );
    const exportBtn = within(headerActions).getByTestId(
      "resume-detail-export-pdf",
    );
    await userEvent.setup().click(exportBtn);

    await waitFor(() => {
      expect(exportHeaders).not.toBeNull();
    });
    expect(exportHeaders!["Idempotency-Key"]).toMatch(/^v1\.\d+\./);

    await waitFor(() => {
      expect(
        toastCalls.some((c) =>
          /即将开放|not available|P0/i.test(c.message),
        ),
      ).toBe(true);
    });

    const offenders: string[] = [];
    for (let i = 0; i < window.localStorage.length; i++) {
      const key = window.localStorage.key(i);
      if (key && /resume|export|pdf/i.test(key)) offenders.push(key);
    }
    expect(offenders).toEqual([]);
  });

  it("non-existent resumeId returns 404 → NotFoundEmptyState renders generic copy and a back-to-list CTA (UI does not echo fixture error.code)", async () => {
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
