// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { EasyInterviewClient } from "../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../api/mockTransport";
import { App } from "./App";

import getRuntimeConfigFixture from "../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../openapi/fixtures/Auth/getMe.json";
import getTargetJobFixture from "../../../openapi/fixtures/TargetJobs/getTargetJob.json";
import listTargetJobsFixture from "../../../openapi/fixtures/TargetJobs/listTargetJobs.json";
import getResumeFixture from "../../../openapi/fixtures/Resumes/getResume.json";
import listResumesFixture from "../../../openapi/fixtures/Resumes/listResumes.json";
import getPracticePlanFixture from "../../../openapi/fixtures/PracticePlans/getPracticePlan.json";

function buildWorkspaceClient(): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([
        getRuntimeConfigFixture,
        getMeFixture,
        getTargetJobFixture,
        listTargetJobsFixture,
        getResumeFixture,
        listResumesFixture,
        getPracticePlanFixture,
      ]),
      { scenario: "default" },
    ),
  });
}

describe("App shell", () => {
  it("hydrates account display preferences once and does not refetch /me across routes", async () => {
    const client = buildWorkspaceClient();
    const getMeSpy = vi.spyOn(client, "getMe");
    const user = userEvent.setup();

    render(
      <App
        client={client}
        requestOptions={{ getMe: { headers: { Prefer: "example=authenticated" } } }}
        initialRoute={{ name: "settings", params: {} }}
      />,
    );

    await waitFor(() =>
      expect(screen.getByTestId("topbar-user-area")).toHaveAttribute(
        "data-signed-in",
        "true",
      ),
    );
    expect(screen.getByTestId("settings-appearance")).toBeInTheDocument();
    expect(getMeSpy).toHaveBeenCalledTimes(1);

    await user.click(screen.getByTestId("topbar-nav-home"));
    expect(screen.getByTestId("route-home")).toBeInTheDocument();
    await user.click(screen.getByTestId("topbar-settings"));
    expect(screen.getByTestId("settings-appearance")).toBeInTheDocument();
    expect(getMeSpy).toHaveBeenCalledTimes(1);
  });

  it("defaults to the home route with App chrome rendered", () => {
    render(<App />);
    expect(screen.getByTestId("app-shell-topbar")).toBeInTheDocument();
    expect(screen.getByTestId("route-home")).toBeInTheDocument();
  });

  it("keeps chrome rendered for context routes (parse / report)", () => {
    // `parse` exposes its route marker while keeping chrome visible.
    // `report` is now wired to ReportScreen; with no
    // reportId it falls back to ReportMissingState which still keeps
    // App chrome visible (per frontend-report-dashboard/001 §4 routing).
    const routeShellContextRoutes = ["parse"] as const;
    for (const name of routeShellContextRoutes) {
      const { unmount } = render(<App initialRoute={{ name, params: {} }} />);
      expect(screen.getByTestId("app-shell-topbar")).toBeInTheDocument();
      expect(screen.getByTestId(`route-${name}`)).toBeInTheDocument();
      unmount();
    }
    const { unmount: unmountReport } = render(
      <App initialRoute={{ name: "report", params: {} }} />,
    );
    expect(screen.getByTestId("app-shell-topbar")).toBeInTheDocument();
    expect(screen.getByTestId("report-missing-report")).toBeInTheDocument();
    expect(screen.queryByTestId("route-report")).not.toBeInTheDocument();
    unmountReport();

    const { unmount: unmountConversation } = render(
      <App initialRoute={{ name: "report_conversation", params: {} }} />,
    );
    expect(screen.getByTestId("app-shell-topbar")).toBeInTheDocument();
    expect(screen.getByTestId("report-conversation-unavailable")).toBeInTheDocument();
    unmountConversation();
  });

  it("renders the target-scoped Reports screen with chrome and no global nav entry", () => {
    render(
      <App
        initialRoute={{
          name: "reports",
          params: { targetJobId: "01918fa0-0000-7000-8000-000000002000" },
        }}
      />,
    );
    expect(screen.getByTestId("reports-screen")).toBeInTheDocument();
    expect(screen.getByTestId("app-shell-topbar")).toBeInTheDocument();
    expect(screen.queryByTestId("topbar-nav-reports")).not.toBeInTheDocument();
    expect(screen.queryByTestId("route-reports")).not.toBeInTheDocument();
  });

  it("keeps Practice chrome visible and hides it only for Generating", () => {
    // Practice keeps the shared App TopBar even when the session locator is missing.
    const practiceRender = render(
      <App initialRoute={{ name: "practice", params: {} }} />,
    );
    expect(screen.getByTestId("app-shell-topbar")).toBeInTheDocument();
    expect(screen.getByTestId("practice-session-lost")).toBeInTheDocument();
    practiceRender.unmount();

    // generating route is now wired to GeneratingScreen (frontend-report-dashboard/001).
    // Without reportId it short-circuits to GeneratingErrorState — both layouts
    // are immersive so the TopBar must still be hidden.
    render(<App initialRoute={{ name: "generating", params: {} }} />);
    expect(screen.queryByTestId("app-shell-topbar")).not.toBeInTheDocument();
    expect(screen.getByTestId("generating-error-state")).toBeInTheDocument();
    expect(screen.queryByTestId("route-generating")).not.toBeInTheDocument();
  });

  it("renders HomeScreen on the home route instead of the route shell", () => {
    render(<App />);
    expect(screen.getByTestId("home-hero-label")).toBeInTheDocument();
    expect(screen.getByTestId("home-hero-title")).toBeInTheDocument();
    expect(screen.getByTestId("home-jd-textarea")).toBeInTheDocument();
  });

  it("renders ParseScreen on the parse route instead of the route shell", () => {
    render(
      <App
        initialRoute={{
          name: "parse",
          params: { targetJobId: "01918fa0-0000-7000-8000-000000002000" },
        }}
      />,
    );
    expect(screen.getByTestId("parse-loading-step-0")).toBeInTheDocument();
    expect(screen.queryByText("route shell")).not.toBeInTheDocument();
  });

  it("ignores legacy phone params and keeps voice disabled on chat", () => {
    render(
      <App
        initialRoute={{
          name: "practice",
          params: {
            sessionId: "01918fa0-0000-7000-8000-000000005000",
            mode: "phone",
            modality: "phone",
            planId: "plan-tj-1",
          },
        }}
      />,
    );
    expect(screen.getByTestId("practice-conversation")).toBeInTheDocument();
    expect(screen.getByTestId("practice-topbar-phone-toggle")).toBeDisabled();
    expect(screen.queryByTestId("practice-phone-surface")).not.toBeInTheDocument();
  });

  it("renders ResumeWorkshopScreen on resume_versions route instead of the route shell", () => {
    render(
      <App initialRoute={{ name: "resume_versions", params: {} }} />,
    );
    expect(screen.getByTestId("resume-workshop-screen")).toBeInTheDocument();
    expect(screen.queryByTestId("route-resume_versions")).not.toBeInTheDocument();
    expect(screen.queryByText("route shell")).not.toBeInTheDocument();
  });

  it("renders a target-scoped workspace as read-only detail with one getTargetJob", async () => {
    const client = buildWorkspaceClient();
    const getTargetJobSpy = vi.spyOn(client, "getTargetJob");
    const listTargetJobsSpy = vi.spyOn(client, "listTargetJobs");
    const listResumesSpy = vi.spyOn(client, "listResumes");
    render(
      <App
        client={client}
        requestOptions={{ getMe: { headers: { Prefer: "example=authenticated" } } }}
        initialRoute={{
          name: "workspace",
          params: { targetJobId: "01918fa0-0000-7000-8000-000000002000" },
        }}
      />,
    );
    await waitFor(() => {
      expect(screen.getByTestId("unified-plan-detail")).toBeInTheDocument();
    });
    expect(getTargetJobSpy).toHaveBeenCalledTimes(1);
    expect(listTargetJobsSpy).not.toHaveBeenCalled();
    expect(listResumesSpy).not.toHaveBeenCalled();
    expect(screen.queryByTestId("parse-loading-step-0")).not.toBeInTheDocument();
    expect(screen.queryByTestId("workspace-plan-list")).not.toBeInTheDocument();
    expect(screen.queryByText("route shell")).not.toBeInTheDocument();
  });

  it("uses only targetJobId as workspace detail authority", async () => {
    const client = buildWorkspaceClient();
    const getTargetJobSpy = vi.spyOn(client, "getTargetJob");
    render(
      <App
        client={client}
        requestOptions={{ getMe: { headers: { Prefer: "example=authenticated" } } }}
        initialRoute={{
          name: "workspace",
          params: {
            targetJobId: "01918fa0-0000-7000-8000-000000002000",
            jdId: "jd-1",
            resumeId: "01918fa0-0000-7000-8000-000000001000",
            roundId: "round-hr",
            planId: "01918fa0-0000-7000-8000-000000004000",
          },
        }}
      />,
    );

    await waitFor(() => {
      expect(screen.getByTestId("unified-plan-detail")).toBeInTheDocument();
    });
    expect(getTargetJobSpy).toHaveBeenCalledTimes(1);
    expect(getTargetJobSpy).toHaveBeenCalledWith(
      "01918fa0-0000-7000-8000-000000002000",
    );
    expect(screen.queryByTestId("workspace-empty")).not.toBeInTheDocument();
  });

  it("practice route renders PracticeScreen instead of the route shell", () => {
    render(
      <App
        initialRoute={{
          name: "practice",
          params: {
            sessionId: "01918fa0-0000-7000-8000-000000005000",
            planId: "plan-1",
            targetJobId: "01918fa0-0000-7000-8000-000000002000",
            jdId: "jd-1",
            resumeId: "01918fa0-0000-7000-8000-000000001000",
            roundId: "round-tech1",
            mode: "text",
            modality: "text",
            practiceMode: "assisted",
            practiceGoal: "baseline",
            hintUsed: "false",
            hintCount: "0",
          },
        }}
      />,
    );
    expect(screen.getByTestId("practice-screen")).toBeInTheDocument();
    expect(screen.getByTestId("practice-topbar")).toBeInTheDocument();
    // route-practice testid is the route-shell marker and must not appear here.
    expect(screen.queryByTestId("route-practice")).not.toBeInTheDocument();
    expect(screen.queryByText("route shell")).not.toBeInTheDocument();
  });

  it("generating route mounts GeneratingScreen with reportId in params (frontend-report-dashboard/001 Phase 1)", () => {
    render(
      <App
        initialRoute={{
          name: "generating",
          params: { sessionId: "sess-1", reportId: "rpt-1" },
        }}
      />,
    );
    expect(screen.getByTestId("generating-screen")).toBeInTheDocument();
    expect(screen.queryByTestId("route-generating")).not.toBeInTheDocument();
  });

  it("standalone insight route falls back to home", () => {
    render(<App initialRoute={{ name: "standalone_insight", params: {} }} />);
    expect(screen.queryByTestId("route-standalone_insight")).not.toBeInTheDocument();
    expect(screen.getByTestId("route-home")).toBeInTheDocument();
  });

  it("report route mounts ReportScreen and requires reportId as its only locator", () => {
    render(<App initialRoute={{ name: "report", params: {} }} />);
    expect(screen.getByTestId("report-missing-report")).toBeInTheDocument();
    expect(screen.queryByTestId("route-report")).not.toBeInTheDocument();
  });

  it("report route ignores caller-selected status and reads state by reportId", async () => {
    const reportId = "01918fa0-0000-7000-8000-000000007000";
    const getFeedbackReport = vi.fn().mockResolvedValue({
      id: reportId,
      sessionId: "01918fa0-0000-7000-8000-000000005000",
      targetJobId: "01918fa0-0000-7000-8000-000000002000",
      status: "generating",
      errorCode: null,
      summary: null,
      context: {
        sourcePlanId: "01918fa0-0000-7000-8000-000000004000",
        targetJobTitle: "Platform Engineer",
        targetJobCompany: "Acme",
        resumeId: "01918fa0-0000-7000-8000-000000001000",
        resumeDisplayName: "Platform resume",
        roundId: "round-1-technical",
        roundSequence: 1,
        roundName: "Technical interview",
        roundType: "technical",
        language: "en",
        hasNextRound: true,
      },
      preparednessLevel: null,
      highlights: [],
      issues: [],
      nextActions: [],
      dimensionAssessments: [],
      retryFocusDimensionCodes: [],
      provenance: null,
      createdAt: "2026-07-12T08:00:00Z",
      updatedAt: "2026-07-12T08:01:00Z",
    });
    const client = {
      async getRuntimeConfig() { return { aiProviderProfile: "stub" } as never; },
      async getMe() { return { id: "user-1" } as never; },
      getFeedbackReport,
    } as unknown as EasyInterviewClient;
    render(
      <App
        client={client}
        initialRoute={{
          name: "report",
          params: {
            sessionId: "route-session-must-drop",
            reportId,
            reportStatus: "failed",
            errorCode: "AI_PROVIDER_TIMEOUT",
          },
        }}
      />,
    );
    expect(await screen.findByTestId("report-pending-state")).toBeInTheDocument();
    expect(screen.queryByTestId("report-failure-state")).not.toBeInTheDocument();
    expect(getFeedbackReport).toHaveBeenCalledWith(reportId);
    expect(screen.queryByTestId("route-report")).not.toBeInTheDocument();
  });
});
