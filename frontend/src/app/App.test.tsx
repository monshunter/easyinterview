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
    const contextRoutes = ["parse", "report", "company_intel"] as const;
    for (const name of contextRoutes) {
      const { unmount } = render(<App initialRoute={{ name, params: {} }} />);
      expect(screen.getByTestId("app-shell-topbar")).toBeInTheDocument();
      expect(screen.getByTestId(`route-${name}`)).toBeInTheDocument();
      unmount();
    }
  });

  it("hides chrome for immersive practice / generating routes", () => {
    const immersiveRoutes = ["practice", "generating"] as const;
    for (const name of immersiveRoutes) {
      const { unmount } = render(<App initialRoute={{ name, params: {} }} />);
      expect(screen.queryByTestId("app-shell-topbar")).not.toBeInTheDocument();
      expect(screen.getByTestId(`route-${name}`)).toBeInTheDocument();
      unmount();
    }
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

  it("propagates route params to the rendered route view", () => {
    render(
      <App
        initialRoute={{
          name: "practice",
          params: { mode: "voice", planId: "plan-tj-1" },
        }}
      />,
    );
    const view = screen.getByTestId("route-practice");
    expect(view).toHaveAttribute(
      "data-route-params",
      JSON.stringify({ mode: "voice", planId: "plan-tj-1" }),
    );
  });

  it("renders WorkspaceScreen on workspace route instead of PlaceholderScreen", () => {
    render(
      <App
        initialRoute={{
          name: "workspace",
          params: { targetJobId: "tj-1" },
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

  it("practice route still renders PlaceholderScreen", () => {
    render(
      <App
        initialRoute={{
          name: "practice",
          params: { sessionId: "sess-1" },
        }}
      />,
    );
    expect(screen.getByTestId("route-practice")).toBeInTheDocument();
  });

  it("generating route still renders PlaceholderScreen", () => {
    render(
      <App
        initialRoute={{
          name: "generating",
          params: { sessionId: "sess-1", reportId: "rpt-1" },
        }}
      />,
    );
    expect(screen.getByTestId("route-generating")).toBeInTheDocument();
  });
});
