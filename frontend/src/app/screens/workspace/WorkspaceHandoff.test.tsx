/**
 * @vitest-environment jsdom
 *
 * Phase 5.4: Workspace embedded insight action +
 * Records placeholder + report handoff negative assertions.
 */

import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useEffect, type ReactNode } from "react";

import { EasyInterviewClient } from "../../../api/generated/client";
import {
  createFixtureBackedFetch,
  createFixtureRegistry,
} from "../../../api/mockTransport";
import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import {
  InterviewContextProvider,
  useInterviewContext,
} from "../../interview-context/InterviewContext";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import type { Route } from "../../routes";
import { WorkspaceScreen } from "./WorkspaceScreen";

import getTargetJobFixture from "../../../../../openapi/fixtures/TargetJobs/getTargetJob.json";
import listTargetJobsFixture from "../../../../../openapi/fixtures/TargetJobs/listTargetJobs.json";
import getResumeFixture from "../../../../../openapi/fixtures/Resumes/getResume.json";
import listResumesFixture from "../../../../../openapi/fixtures/Resumes/listResumes.json";
import createPracticePlanFixture from "../../../../../openapi/fixtures/PracticePlans/createPracticePlan.json";
import startPracticeSessionFixture from "../../../../../openapi/fixtures/PracticeSessions/startPracticeSession.json";
import getPracticePlanFixture from "../../../../../openapi/fixtures/PracticePlans/getPracticePlan.json";
import updateTargetJobFixture from "../../../../../openapi/fixtures/TargetJobs/updateTargetJob.json";
import getMeFixture from "../../../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";

const BASE_FIXTURES = [
  getRuntimeConfigFixture,
  getMeFixture,
  getTargetJobFixture,
  listTargetJobsFixture,
  getResumeFixture,
  listResumesFixture,
  updateTargetJobFixture,
  createPracticePlanFixture,
  startPracticeSessionFixture,
  getPracticePlanFixture,
];

function buildClient(): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry(BASE_FIXTURES),
      { scenario: "default" },
    ),
  });
}

function HydrateContext({ route }: { route: Route }) {
  const { dispatch } = useInterviewContext();
  useEffect(() => {
    dispatch({ type: "HYDRATE_FROM_ROUTE", params: route.params });
  }, []);
  return null;
}

const WORKSPACE_ROUTE: Route = {
  name: "workspace",
  params: {
    targetJobId: "01918fa0-0000-7000-8000-000000002000",
    jdId: "jd-1",
    resumeId: "01918fa0-0000-7000-8000-000000001000",
    roundId: "round-hr",
    planId: "",
  },
};

function renderWorkspace(nav: ReturnType<typeof vi.fn>) {
  const client = buildClient();
  return {
    client,
    ...render(
      <DisplayPreferencesProvider>
        <InterviewContextProvider>
          <AppRuntimeProvider
            client={client}
            requestOptions={{
              getMe: { headers: { Prefer: "example=authenticated" } },
            }}
          >
            <NavigationProvider value={{ navigate: nav }}>
              <HydrateContext route={WORKSPACE_ROUTE} />
              <WorkspaceScreen route={WORKSPACE_ROUTE} />
            </NavigationProvider>
          </AppRuntimeProvider>
        </InterviewContextProvider>
      </DisplayPreferencesProvider>,
    ),
  };
}

describe("WorkspaceHandoff (Phase 5.4)", () => {
  it("ordinary workspace detail re-entry renders the unified plan detail mother page", async () => {
    const nav = vi.fn();
    renderWorkspace(nav);

    await waitFor(() => {
      expect(screen.getByTestId("unified-plan-detail")).toBeDefined();
    });

    expect(screen.getByTestId("route-workspace")).toBeDefined();
    expect(screen.getByTestId("unified-plan-detail-title")).toHaveTextContent(
      "Interview plan detail",
    );
    expect(screen.queryByTestId("workspace-header")).toBeNull();
    expect(screen.queryByTestId("workspace-launcher")).toBeNull();
    expect(screen.queryByTestId("workspace-jd-card")).toBeNull();
    expect(screen.queryByTestId("workspace-prep-card")).toBeNull();
    expect(screen.queryByTestId("workspace-history-card")).toBeNull();
  });

  it("unified detail Save plan stays on workspace with declared target, plan, and resume IDs", async () => {
    const nav = vi.fn();
    const { client } = renderWorkspace(nav);
    const updateSpy = vi.spyOn(client, "updateTargetJob");
    const user = userEvent.setup();

    await waitFor(() => {
      expect(screen.getByTestId("parse-action-save-plan")).toBeEnabled();
    });

    await user.click(screen.getByTestId("parse-action-save-plan"));

    await waitFor(() => {
      expect(updateSpy).toHaveBeenCalledTimes(1);
    });
    const navCall = nav.mock.calls[0]![0] as Record<string, unknown>;
    expect(navCall.name).toBe("workspace");
    const params = navCall.params as Record<string, string>;
    expect(params.targetJobId).toBe(WORKSPACE_ROUTE.params.targetJobId);
    expect(params.jobId).toBe(WORKSPACE_ROUTE.params.targetJobId);
    expect(params.planId).toBe("01918fa0-0000-7000-8000-000000004000");
    expect(params.resumeId).toBe(WORKSPACE_ROUTE.params.resumeId);
    expect(JSON.stringify(params)).not.toContain("plan-01918fa0");
    expect(JSON.stringify(params)).not.toContain("resume-unbound");
  });

  it("getFeedbackReport is NOT called during workspace render", async () => {
    const nav = vi.fn();
    const { client } = renderWorkspace(nav);
    const fbSpy = vi.spyOn(client, "getFeedbackReport");

    await waitFor(() => {
      expect(screen.getByTestId("parse-action-start-interview")).toBeDefined();
    });

    // The workspace should never call getFeedbackReport
    expect(fbSpy).not.toHaveBeenCalled();
  });

  it("ordinary unified detail no longer renders the old workspace records placeholder", async () => {
    const nav = vi.fn();
    renderWorkspace(nav);

    await waitFor(() => {
      expect(screen.getByTestId("unified-plan-detail")).toBeDefined();
    });

    expect(screen.queryByTestId("workspace-history-card")).toBeNull();
    expect(screen.queryByTestId("workspace-history-empty")).toBeNull();
  });

  it("ordinary unified detail exposes no report navigation surface", async () => {
    const nav = vi.fn();
    renderWorkspace(nav);

    await waitFor(() => {
      expect(screen.getByTestId("unified-plan-detail")).toBeDefined();
    });

    const reportCalls = nav.mock.calls.filter(
      ([call]) => (call as Record<string, unknown>).name === "report",
    );
    expect(reportCalls).toHaveLength(0);
  });

  it("workspace plan-list landing CTA navigates to home", async () => {
    const nav = vi.fn();
    const emptyRoute: Route = {
      name: "workspace",
      params: {},
    };

    const client = buildClient();
    render(
      <DisplayPreferencesProvider>
        <InterviewContextProvider>
          <AppRuntimeProvider
            client={client}
            requestOptions={{
              getMe: { headers: { Prefer: "example=authenticated" } },
            }}
          >
            <NavigationProvider value={{ navigate: nav }}>
              <HydrateContext route={emptyRoute} />
              <WorkspaceScreen route={emptyRoute} />
            </NavigationProvider>
          </AppRuntimeProvider>
        </InterviewContextProvider>
      </DisplayPreferencesProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("workspace-plan-list")).toBeDefined();
    });

    const user = userEvent.setup();
    await user.click(screen.getByTestId("workspace-plan-list-create"));

    expect(nav).toHaveBeenCalledWith(
      expect.objectContaining({ name: "home" }),
    );
  });

  it("workspace detail with no bound resume blocks Save/Start inside the unified detail", async () => {
    const nav = vi.fn();
    const missingResumeRoute: Route = {
      name: "workspace",
      params: {
        targetJobId: "01918fa0-0000-7000-8000-000000002000",
        // No resumeId
      },
    };

    const client = buildClient();
    vi.spyOn(client, "getTargetJob").mockResolvedValue({
      ...getTargetJobFixture.scenarios.default.response.body,
      resumeId: null,
    } as Awaited<ReturnType<EasyInterviewClient["getTargetJob"]>>);
    render(
      <DisplayPreferencesProvider>
        <InterviewContextProvider>
          <AppRuntimeProvider
            client={client}
            requestOptions={{
              getMe: { headers: { Prefer: "example=authenticated" } },
            }}
          >
            <NavigationProvider value={{ navigate: nav }}>
              <HydrateContext route={missingResumeRoute} />
              <WorkspaceScreen route={missingResumeRoute} />
            </NavigationProvider>
          </AppRuntimeProvider>
        </InterviewContextProvider>
      </DisplayPreferencesProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("parse-resume-required")).toBeDefined();
    });

    expect(screen.queryByTestId("workspace-missing-resume")).toBeNull();
    expect(screen.getByTestId("parse-action-save-plan")).toBeDisabled();
    expect(screen.getByTestId("parse-action-start-interview")).toBeDisabled();
    expect(nav).not.toHaveBeenCalled();
  });
});
