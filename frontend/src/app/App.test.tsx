// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";

import { EasyInterviewClient } from "../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../api/mockTransport";
import { App } from "./App";

import getRuntimeConfigFixture from "../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../openapi/fixtures/Auth/getMe.json";
import getTargetJobFixture from "../../../openapi/fixtures/TargetJobs/getTargetJob.json";
import getResumeFixture from "../../../openapi/fixtures/Resumes/getResume.json";
import getPracticePlanFixture from "../../../openapi/fixtures/PracticePlans/getPracticePlan.json";

function buildWorkspaceClient(): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([
        getRuntimeConfigFixture,
        getMeFixture,
        getTargetJobFixture,
        getResumeFixture,
        getPracticePlanFixture,
      ]),
      { scenario: "default" },
    ),
  });
}

describe("App shell", () => {
  it("defaults to the home route with App chrome rendered", () => {
    render(<App />);
    expect(screen.getByTestId("app-shell-topbar")).toBeInTheDocument();
    expect(screen.getByTestId("route-home")).toBeInTheDocument();
  });

  it("keeps chrome rendered for context routes (parse / report / company_intel)", () => {
    // `parse` and `company_intel` still go through PlaceholderScreen — assert
    // via route-${name}. `report` is now wired to ReportScreen; with no
    // sessionId it falls back to ReportMissingSessionState which still keeps
    // App chrome visible (per frontend-report-dashboard/001 §4 routing).
    const placeholderContextRoutes = ["parse", "company_intel"] as const;
    for (const name of placeholderContextRoutes) {
      const { unmount } = render(<App initialRoute={{ name, params: {} }} />);
      expect(screen.getByTestId("app-shell-topbar")).toBeInTheDocument();
      expect(screen.getByTestId(`route-${name}`)).toBeInTheDocument();
      unmount();
    }
    const { unmount: unmountReport } = render(
      <App initialRoute={{ name: "report", params: {} }} />,
    );
    expect(screen.getByTestId("app-shell-topbar")).toBeInTheDocument();
    expect(screen.getByTestId("report-missing-session")).toBeInTheDocument();
    expect(screen.queryByTestId("route-report")).not.toBeInTheDocument();
    unmountReport();
  });

  it("hides chrome for immersive practice / generating routes", () => {
    // practice route now mounts PracticeScreen; without sessionId it falls
    // back to PracticeSessionLostState (still chrome-less).
    const practiceRender = render(
      <App initialRoute={{ name: "practice", params: {} }} />,
    );
    expect(screen.queryByTestId("app-shell-topbar")).not.toBeInTheDocument();
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

  it("renders HomeScreen on the home route instead of PlaceholderScreen", () => {
    render(<App />);
    expect(screen.getByTestId("home-hero-label")).toBeInTheDocument();
    expect(screen.getByTestId("home-hero-title")).toBeInTheDocument();
    expect(screen.getByTestId("home-jd-textarea")).toBeInTheDocument();
  });

  it("renders ParseScreen on the parse route instead of PlaceholderScreen", () => {
    render(
      <App
        initialRoute={{
          name: "parse",
          params: { targetJobId: "01918fa0-0000-7000-8000-000000002000" },
        }}
      />,
    );
    expect(screen.getByTestId("parse-loading-step-0")).toBeInTheDocument();
    expect(screen.queryByText("D2-D6")).not.toBeInTheDocument();
  });

  it("propagates voice mode route params into PracticeScreen voice placeholder", () => {
    render(
      <App
        initialRoute={{
          name: "practice",
          params: {
            sessionId: "01918fa0-0000-7000-8000-000000005000",
            mode: "voice",
            modality: "voice",
            planId: "plan-tj-1",
          },
        }}
      />,
    );
    expect(screen.getByTestId("practice-voice-coming-soon")).toBeInTheDocument();
  });

  it("renders ResumeWorkshopScreen on resume_versions route instead of PlaceholderScreen", () => {
    render(
      <App initialRoute={{ name: "resume_versions", params: {} }} />,
    );
    expect(screen.getByTestId("resume-workshop-screen")).toBeInTheDocument();
    expect(screen.queryByTestId("route-resume_versions")).not.toBeInTheDocument();
    expect(screen.queryByText("D2-D6")).not.toBeInTheDocument();
  });

  it("renders WorkspaceScreen on workspace route instead of PlaceholderScreen", () => {
    render(
      <App
        initialRoute={{
          name: "workspace",
          params: { targetJobId: "01918fa0-0000-7000-8000-000000002000" },
        }}
      />,
    );
    expect(screen.getByTestId("workspace-crumbs")).toBeInTheDocument();
    expect(screen.queryByTestId("route-workspace")).not.toBeInTheDocument();
  });

  it("hydrates workspace route params into InterviewContext and loads fixture data", async () => {
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
            resumeVersionId: "01918fa0-0000-7000-8000-000000001000",
            roundId: "round-hr",
            planId: "01918fa0-0000-7000-8000-000000004000",
          },
        }}
      />,
    );

    await waitFor(() => {
      expect(screen.getByTestId("workspace-header-title")).toHaveTextContent(
        "Senior Frontend Engineer",
      );
    });
    expect(getTargetJobSpy).toHaveBeenCalledWith(
      "01918fa0-0000-7000-8000-000000002000",
    );
    expect(screen.queryByTestId("workspace-empty")).not.toBeInTheDocument();
  });

  it("practice route renders PracticeScreen instead of PlaceholderScreen", () => {
    render(
      <App
        initialRoute={{
          name: "practice",
          params: {
            sessionId: "01918fa0-0000-7000-8000-000000005000",
            planId: "plan-1",
            targetJobId: "01918fa0-0000-7000-8000-000000002000",
            jdId: "jd-1",
            resumeVersionId: "01918fa0-0000-7000-8000-000000001000",
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
    // route-practice testid is the PlaceholderScreen marker — must NOT appear.
    expect(screen.queryByTestId("route-practice")).not.toBeInTheDocument();
    expect(screen.queryByText("D2-D6")).not.toBeInTheDocument();
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

  it("company_intel route still renders PlaceholderScreen (out of scope)", () => {
    render(<App initialRoute={{ name: "company_intel", params: {} }} />);
    expect(screen.getByTestId("route-company_intel")).toBeInTheDocument();
  });

  it("report route mounts ReportScreen — dispatches missingSession when sessionId absent (frontend-report-dashboard/001 Phase 2)", () => {
    render(<App initialRoute={{ name: "report", params: {} }} />);
    expect(screen.getByTestId("report-missing-session")).toBeInTheDocument();
    expect(screen.queryByTestId("route-report")).not.toBeInTheDocument();
  });

  it("report route dispatches ReportFailureState when reportStatus=failed (frontend-report-dashboard/001 Phase 2)", () => {
    render(
      <App
        initialRoute={{
          name: "report",
          params: {
            sessionId: "sess-1",
            reportId: "rpt-1",
            reportStatus: "failed",
            errorCode: "AI_PROVIDER_TIMEOUT",
          },
        }}
      />,
    );
    expect(screen.getByTestId("report-failure-state")).toBeInTheDocument();
    expect(screen.queryByTestId("route-report")).not.toBeInTheDocument();
  });
});
