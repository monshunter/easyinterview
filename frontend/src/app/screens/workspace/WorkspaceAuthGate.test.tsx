/**
 * @vitest-environment jsdom
 *
 * Phase 4.8: Workspace auth gate — unauthenticated user clicks "立即面试"
 * → requestAuth → navigates to auth_login with full pendingAction params
 * containing autoStartPractice=1 and no sensitive fields.
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

function buildAuthenticatedClient(): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry(BASE_FIXTURES),
      { scenario: "default" },
    ),
  });
}

function buildUnauthenticatedClient(): EasyInterviewClient {
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

function renderWorkspace(
  client: EasyInterviewClient,
  nav: ReturnType<typeof vi.fn>,
  requestOptions?: { getMe?: { headers: Record<string, string> } },
) {
  return render(
    <DisplayPreferencesProvider>
      <InterviewContextProvider>
        <AppRuntimeProvider client={client} requestOptions={requestOptions}>
          <NavigationProvider value={{ navigate: nav }}>
            <HydrateContext route={WORKSPACE_ROUTE} />
            <WorkspaceScreen route={WORKSPACE_ROUTE} />
          </NavigationProvider>
        </AppRuntimeProvider>
      </InterviewContextProvider>
    </DisplayPreferencesProvider>,
  );
}

describe("WorkspaceAuthGate (Phase 4.8)", () => {
  it("unauthenticated user: click 立即面试 → navigates to auth_login with pendingRoute=workspace", async () => {
    const client = buildUnauthenticatedClient();
    const nav = vi.fn();

    renderWorkspace(client, nav, {
      getMe: { headers: { Prefer: "example=unauthenticated" } },
    });

    await waitFor(() => {
      expect(screen.getByTestId("workspace-cta-start")).toBeDefined();
    });

    const user = userEvent.setup();
    await user.click(screen.getByTestId("workspace-cta-start"));

    // requestAuth must navigate to auth_login
    await waitFor(() => {
      expect(nav).toHaveBeenCalled();
    });

    const navCall = nav.mock.calls[0]![0] as Record<string, unknown>;
    expect(navCall.name).toBe("auth_login");
  });

  it("unauthenticated user: pendingAction params carry pendingRoute=workspace and interview context keys", async () => {
    const client = buildUnauthenticatedClient();
    const nav = vi.fn();

    renderWorkspace(client, nav, {
      getMe: { headers: { Prefer: "example=unauthenticated" } },
    });

    await waitFor(() => {
      expect(screen.getByTestId("workspace-cta-start")).toBeDefined();
    });

    const user = userEvent.setup();
    await user.click(screen.getByTestId("workspace-cta-start"));

    await waitFor(() => {
      expect(nav).toHaveBeenCalled();
    });

    const navCall = nav.mock.calls[0]![0] as Record<string, unknown>;
    const params = navCall.params as Record<string, string>;

    expect(params.pendingRoute).toBe("workspace");
    expect(params.pendingType).toBe("start_practice");
    expect(params.targetJobId).toBe(WORKSPACE_ROUTE.params.targetJobId);
    expect(params.jdId).toBe(WORKSPACE_ROUTE.params.jdId);
    expect(params.resumeVersionId).toBe(WORKSPACE_ROUTE.params.resumeVersionId);
    expect(params.roundId).toBe(WORKSPACE_ROUTE.params.roundId);
    expect(params.autoStartPractice).toBe("1");
    // PracticeDisplayContext fields
    expect(params.practiceMode).toBe("strict");
    expect(params.practiceGoal).toBe("baseline");
    expect(params.mode).toBe("text");
    expect(params.modality).toBe("text");
    expect(params.hintUsed).toBe("false");
    expect(params.hintCount).toBe("0");
  });

  it("unauthenticated user: pendingAction.params does NOT contain sensitive fields", async () => {
    const client = buildUnauthenticatedClient();
    const nav = vi.fn();

    renderWorkspace(client, nav, {
      getMe: { headers: { Prefer: "example=unauthenticated" } },
    });

    await waitFor(() => {
      expect(screen.getByTestId("workspace-cta-start")).toBeDefined();
    });

    const user = userEvent.setup();
    await user.click(screen.getByTestId("workspace-cta-start"));

    await waitFor(() => {
      expect(nav).toHaveBeenCalled();
    });

    const navCall = nav.mock.calls[0]![0] as Record<string, unknown>;
    const params = navCall.params as Record<string, string>;

    // No sensitive fields
    expect(params.answerText).toBeUndefined();
    expect(params.hintText).toBeUndefined();
    expect(params.promptHash).toBeUndefined();
    expect(params.rawTranscript).toBeUndefined();
    expect(params.jdRaw).toBeUndefined();
    expect(params.resumeRaw).toBeUndefined();
    expect(params.questionText).toBeUndefined();

    // Reserved keys shouldn't leak as business params
    expect(params.pendingLabel).toBeDefined(); // reserved key, okay
  });

  it("authenticated user: click 立即面试 → proceeds with startPractice (does not redirect)", async () => {
    const client = buildAuthenticatedClient();
    const nav = vi.fn();
    const startSpy = vi.spyOn(client, "startPracticeSession");

    renderWorkspace(client, nav, {
      getMe: { headers: { Prefer: "example=authenticated" } },
    });

    await waitFor(() => {
      expect(screen.getByTestId("workspace-cta-start")).toBeDefined();
    });

    const user = userEvent.setup();
    await user.click(screen.getByTestId("workspace-cta-start"));

    // Should navigate to practice (not auth_login) after successful start
    await waitFor(() => {
      expect(nav).toHaveBeenCalled();
    });

    const navCall = nav.mock.calls[0]![0] as Record<string, unknown>;
    expect(navCall.name).toBe("practice");
    expect(startSpy).toHaveBeenCalled();
  });

  it("pendingAction label matches workspace.startCore i18n key", async () => {
    const client = buildUnauthenticatedClient();
    const nav = vi.fn();

    renderWorkspace(client, nav, {
      getMe: { headers: { Prefer: "example=unauthenticated" } },
    });

    await waitFor(() => {
      expect(screen.getByTestId("workspace-cta-start")).toBeDefined();
    });

    const user = userEvent.setup();
    await user.click(screen.getByTestId("workspace-cta-start"));

    await waitFor(() => {
      expect(nav).toHaveBeenCalled();
    });

    const navCall = nav.mock.calls[0]![0] as Record<string, unknown>;
    const params = navCall.params as Record<string, string>;

    // pendingLabel should be the i18n label (not empty, not undefined)
    expect(params.pendingLabel).toBeTruthy();
    expect(typeof params.pendingLabel).toBe("string");
  });
});
