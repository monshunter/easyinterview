// @vitest-environment jsdom
import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import type { EasyInterviewClient } from "../../../api/generated/client";
import type { Resume, TargetJob } from "../../../api/generated/types";
import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import {
  AppRuntimeContext,
  type AppRuntimeValue,
} from "../../runtime/AppRuntimeProvider";
import { ParseScreen } from "./ParseScreen";

const TARGET_JOB_ID = "01918fa0-0000-7000-8000-000000002000";
const RESUME_ID = "resume-ready";

const targetJob: TargetJob = {
  id: TARGET_JOB_ID,
  title: "Frontend Engineer",
  companyName: "Acme",
  analysisStatus: "ready",
  status: "draft",
  targetLanguage: "en",
  requirements: [],
  openQuestionIssueCount: 0,
  resumeId: RESUME_ID,
  practiceProgress: {
    status: "not_started",
    completedRounds: [],
    currentRound: { roundId: "round-1-technical", roundSequence: 1 },
  },
  summary: {
    coreThemes: [],
    interviewRounds: [
      {
        sequence: 1,
        type: "technical",
        name: "Architecture",
        durationMinutes: 50,
        focus: "System boundaries",
      },
    ],
    provenance: {
      modelId: "model",
      promptVersion: "prompt",
      rubricVersion: "rubric",
      dataSourceVersion: "source",
      featureFlag: "flag",
      language: "en",
    },
  },
  createdAt: "2026-07-14T08:00:00Z",
  updatedAt: "2026-07-14T08:00:00Z",
};

const resume: Resume = {
  id: RESUME_ID,
  title: "Frontend resume",
  displayName: "Frontend resume",
  language: "en",
  parseStatus: "ready",
  sourceType: "paste",
  status: "active",
  createdAt: "2026-07-14T08:00:00Z",
  updatedAt: "2026-07-14T08:00:00Z",
};

function runtimeValue(client: EasyInterviewClient): AppRuntimeValue {
  return {
    client,
    runtime: { status: "ready", config: {} as never },
    auth: { status: "authenticated", user: {} as never },
    refreshAuth: vi.fn(),
  };
}

function renderPreview(
  client: EasyInterviewClient,
  navigate = vi.fn(),
  params: Record<string, string> = {},
  job: TargetJob = targetJob,
  stage: "loading" | "preview" | "error" | "failed" = "preview",
): ReactNode {
  return (
    <DisplayPreferencesProvider>
      <AppRuntimeContext.Provider value={runtimeValue(client)}>
        <NavigationProvider value={{ navigate }}>
          <ParseScreen
            route={{
              name: "parse",
              params: { targetJobId: job.id, ...params },
            }}
            _mockStage={stage}
            _mockTargetJob={job}
          />
        </NavigationProvider>
      </AppRuntimeContext.Provider>
    </DisplayPreferencesProvider>
  );
}

function clientWithReportSpy(): {
  client: EasyInterviewClient;
  listTargetJobReports: ReturnType<typeof vi.fn>;
} {
  const listTargetJobReports = vi.fn();
  return {
    client: {
      listResumes: vi.fn(async () => ({ items: [resume], pageInfo: {} })),
      listTargetJobReports,
    } as unknown as EasyInterviewClient,
    listTargetJobReports,
  };
}

afterEach(() => {
  localStorage.removeItem("ei-lang");
});

describe("Parse report handoff", () => {
  it("renders one page-level entry and never requests or embeds reports", async () => {
    const navigate = vi.fn();
    const { client, listTargetJobReports } = clientWithReportSpy();
    render(renderPreview(client, navigate));

    const entry = await screen.findByTestId("parse-reports-entry");
    expect(entry).toBeVisible();
    expect(screen.queryByTestId("parse-reports")).not.toBeInTheDocument();
    expect(screen.queryByTestId("parse-report-section")).not.toBeInTheDocument();
    expect(listTargetJobReports).not.toHaveBeenCalled();

    fireEvent.click(within(entry).getByRole("button"));
    expect(navigate).toHaveBeenCalledWith({
      name: "reports",
      params: { targetJobId: TARGET_JOB_ID },
    });
  });

  it("ignores the retired section=reports parameter without scrolling or focusing", async () => {
    const scrollIntoView = vi.fn();
    const original = HTMLElement.prototype.scrollIntoView;
    Object.defineProperty(HTMLElement.prototype, "scrollIntoView", {
      configurable: true,
      value: scrollIntoView,
    });
    const { client, listTargetJobReports } = clientWithReportSpy();
    render(renderPreview(client, vi.fn(), { section: "reports" }));

    await screen.findByTestId("parse-reports-entry");
    await waitFor(() => expect(client.listResumes).toHaveBeenCalled());
    expect(scrollIntoView).not.toHaveBeenCalled();
    expect(listTargetJobReports).not.toHaveBeenCalled();
    expect(screen.queryByTestId("parse-reports")).not.toBeInTheDocument();

    if (original) {
      Object.defineProperty(HTMLElement.prototype, "scrollIntoView", {
        configurable: true,
        value: original,
      });
    } else {
      delete (HTMLElement.prototype as { scrollIntoView?: unknown }).scrollIntoView;
    }
  });

  it.each(["loading", "failed"] as const)(
    "never requests the report list in the %s detail state",
    async (stage) => {
      const { client, listTargetJobReports } = clientWithReportSpy();
      render(renderPreview(client, vi.fn(), {}, targetJob, stage));

      await waitFor(() => expect(listTargetJobReports).not.toHaveBeenCalled());
      expect(screen.queryByTestId("parse-reports")).not.toBeInTheDocument();
    },
  );

  it("does not start a report-list request while the trusted TargetJob switches", async () => {
    const { client, listTargetJobReports } = clientWithReportSpy();
    const targetJobB: TargetJob = {
      ...targetJob,
      id: "01918fa0-0000-7000-8000-000000002001",
      title: "Backend Engineer",
    };
    const rendered = render(renderPreview(client));

    await screen.findByTestId("parse-reports-entry");
    rendered.rerender(renderPreview(client, vi.fn(), {}, targetJobB));
    await screen.findByTestId("parse-reports-entry");

    expect(listTargetJobReports).not.toHaveBeenCalled();
    expect(screen.queryByTestId("parse-reports")).not.toBeInTheDocument();
  });
});
