/**
 * @vitest-environment jsdom
 *
 * Phase 5.4: Workspace handoff tests — CompanyIntelEmbed handoff + sessionHistory
 * placeholder + negative assertions (getCompanyIntel/getFeedbackReport not called).
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
import getResumeFixture from "../../../../../openapi/fixtures/Resumes/getResume.json";
import createPracticePlanFixture from "../../../../../openapi/fixtures/PracticePlans/createPracticePlan.json";
import startPracticeSessionFixture from "../../../../../openapi/fixtures/PracticeSessions/startPracticeSession.json";
import getPracticePlanFixture from "../../../../../openapi/fixtures/PracticePlans/getPracticePlan.json";
import getMeFixture from "../../../../../openapi/fixtures/Auth/getMe.json";
import getRuntimeConfigFixture from "../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";

const BASE_FIXTURES = [
  getRuntimeConfigFixture,
  getMeFixture,
  getTargetJobFixture,
  getResumeFixture,
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
    resumeVersionId: "01918fa0-0000-7000-8000-000000001000",
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
  it("CompanyIntelEmbed 'open' navigates to company_intel with targetJobId/jdId", async () => {
    const nav = vi.fn();
    renderWorkspace(nav);

    const user = userEvent.setup();

    await waitFor(() => {
      expect(screen.getByTestId("workspace-companyintel-open")).toBeDefined();
    });

    await user.click(screen.getByTestId("workspace-companyintel-open"));

    expect(nav).toHaveBeenCalled();
    const navCall = nav.mock.calls[0]![0] as Record<string, unknown>;
    expect(navCall.name).toBe("company_intel");
    const params = navCall.params as Record<string, string>;
    expect(params.targetJobId).toBe(WORKSPACE_ROUTE.params.targetJobId);
    expect(params.jdId).toBe(WORKSPACE_ROUTE.params.jdId);
  });

  it("getFeedbackReport is NOT called during workspace render", async () => {
    const nav = vi.fn();
    const { client } = renderWorkspace(nav);
    const fbSpy = vi.spyOn(client, "getFeedbackReport");

    await waitFor(() => {
      expect(screen.getByTestId("workspace-cta-start")).toBeDefined();
    });

    // The workspace should never call getFeedbackReport
    expect(fbSpy).not.toHaveBeenCalled();
  });

  it("sessionHistory renders EmptyHistory placeholder", async () => {
    const nav = vi.fn();
    renderWorkspace(nav);

    await waitFor(() => {
      expect(screen.getByTestId("workspace-history-card")).toBeDefined();
    });

    expect(screen.getByTestId("workspace-history-empty")).toBeDefined();
  });

  it("sessionHistory empty placeholder clicking does NOT trigger report nav", async () => {
    const nav = vi.fn();
    renderWorkspace(nav);

    const user = userEvent.setup();

    await waitFor(() => {
      expect(screen.getByTestId("workspace-history-empty")).toBeDefined();
    });

    await user.click(screen.getByTestId("workspace-history-empty"));

    // Verify no report navigation happened
    const reportCalls = nav.mock.calls.filter(
      ([call]) => (call as Record<string, unknown>).name === "report",
    );
    expect(reportCalls).toHaveLength(0);
  });

  it("workspace empty state CTA navigates to home", async () => {
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
      expect(screen.getByTestId("workspace-empty")).toBeDefined();
    });

    const user = userEvent.setup();
    await user.click(screen.getByTestId("workspace-empty-cta"));

    expect(nav).toHaveBeenCalledWith(
      expect.objectContaining({ name: "home" }),
    );
  });

  it("workspace missing resume state CTA navigates to resume_versions?flow=create", async () => {
    const nav = vi.fn();
    const missingResumeRoute: Route = {
      name: "workspace",
      params: {
        targetJobId: "01918fa0-0000-7000-8000-000000002000",
        // No resumeVersionId
      },
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
              <HydrateContext route={missingResumeRoute} />
              <WorkspaceScreen route={missingResumeRoute} />
            </NavigationProvider>
          </AppRuntimeProvider>
        </InterviewContextProvider>
      </DisplayPreferencesProvider>,
    );

    await waitFor(() => {
      expect(screen.getByTestId("workspace-missing-resume")).toBeDefined();
    });

    const user = userEvent.setup();
    await user.click(screen.getByTestId("workspace-missing-resume-cta"));

    expect(nav).toHaveBeenCalledWith(
      expect.objectContaining({ name: "resume_versions" }),
    );
    const navCall = nav.mock.calls[0]![0] as Record<string, unknown>;
    const params = navCall.params as Record<string, string>;
    expect(params.flow).toBe("create");
  });
});
