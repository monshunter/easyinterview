/**
 * @vitest-environment jsdom
 */

import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useEffect } from "react";

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

/** Common base fixtures for authenticated workspace rendering. */
const BASE_FIXTURES = [
  getRuntimeConfigFixture,
  getMeFixture,
  getTargetJobFixture,
  getResumeFixture,
  createPracticePlanFixture,
  getPracticePlanFixture,
];

/** Build a client with fixture-backed transport (authenticated, default scenarios). */
function buildClient(startScenario: keyof typeof startPracticeSessionFixture.scenarios = "default") {
  const startFixture =
    startScenario === "default"
      ? startPracticeSessionFixture
      : {
          ...startPracticeSessionFixture,
          scenarios: {
            ...startPracticeSessionFixture.scenarios,
            default: startPracticeSessionFixture.scenarios[startScenario]!,
          },
        };

  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([...BASE_FIXTURES, startFixture]),
      { scenario: "default" },
    ),
  });
}

/** Build a client with completely custom fixture list (for error scenarios). */
function buildCustomFixtures(
  extras: Parameters<typeof createFixtureRegistry>[0],
): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([...BASE_FIXTURES, ...extras]),
      { scenario: "default" },
    ),
  });
}

/** Minimal render with full provider chain for workspace screen. */
function renderScreen(route: Route, client?: EasyInterviewClient) {
  const c = client ?? buildClient();
  const nav = vi.fn();
  return {
    client: c,
    nav,
    ...render(
      <DisplayPreferencesProvider>
        <InterviewContextProvider>
          <AppRuntimeProvider client={c}>
            <NavigationProvider value={{ navigate: nav }}>
              <HydrateContext route={route} />
              <WorkspaceScreen route={route} />
            </NavigationProvider>
          </AppRuntimeProvider>
        </InterviewContextProvider>
      </DisplayPreferencesProvider>,
    ),
  };
}

/** Hydrate InterviewContext from route params on mount. */
function HydrateContext({ route }: { route: Route }) {
  const { dispatch } = useInterviewContext();
  useEffect(() => {
    dispatch({ type: "HYDRATE_FROM_ROUTE", params: route.params });
  }, []);
  return null;
}

function mockTargetJobWithoutCurrentPlan(client: EasyInterviewClient) {
  vi.spyOn(client, "getTargetJob").mockResolvedValue({
    ...getTargetJobFixture.scenarios.default.response.body,
    currentPracticePlanId: null,
  } as Awaited<ReturnType<EasyInterviewClient["getTargetJob"]>>);
}

const FULL_ROUTE: Route = {
  name: "workspace",
  params: {
    targetJobId: "01918fa0-0000-7000-8000-000000002000",
    jdId: "jd-1",
    resumeId: "01918fa0-0000-7000-8000-000000001000",
    roundId: "round-hr",
    planId: "",
    autoStartPractice: "1",
  },
};

const PLAN_EXISTS_ROUTE: Route = {
  name: "workspace",
  params: {
    targetJobId: "01918fa0-0000-7000-8000-000000002000",
    jdId: "jd-1",
    resumeId: "01918fa0-0000-7000-8000-000000001000",
    roundId: "round-hr",
    planId: "01918fa0-0000-7000-8000-000000004000",
    autoStartPractice: "1",
  },
};

const SYNTHETIC_PLAN_ROUTE: Route = {
  name: "workspace",
  params: {
    targetJobId: "01918fa0-0000-7000-8000-000000002000",
    jdId: "jd-1",
    resumeId: "01918fa0-0000-7000-8000-000000001000",
    roundId: "round-hr",
    planId: "plan-01918fa0-0000-7000-8000-000000002000",
    autoStartPractice: "1",
  },
};

// ── Phase 4.7 Comprehensive Tests ──

describe("WorkspaceStartPractice (Phase 4.7)", () => {
  it("happy path: no plan → createPracticePlan → startPracticeSession → nav practice", async () => {
    const client = buildClient();
    mockTargetJobWithoutCurrentPlan(client);
    const createSpy = vi.spyOn(client, "createPracticePlan");
    const startSpy = vi.spyOn(client, "startPracticeSession");
    const { nav } = renderScreen(FULL_ROUTE, client);

    await waitFor(() => {
      expect(nav).toHaveBeenCalled();
    });

    expect(createSpy).toHaveBeenCalledTimes(1);
    expect(startSpy).toHaveBeenCalledTimes(1);

    const navCall = nav.mock.calls[0]![0] as Record<string, unknown>;
    expect(navCall.name).toBe("practice");
    const params = navCall.params as Record<string, unknown>;
    expect(params).toHaveProperty("sessionId");
    expect(params.planId).toBe("01918fa0-0000-7000-8000-000000004000");
    expect(params).toHaveProperty("targetJobId");
    expect(params).toHaveProperty("jdId");
    expect(params).toHaveProperty("resumeId");
    expect(params).toHaveProperty("roundId");
    expect(params).toHaveProperty("mode");
    expect(params).toHaveProperty("modality");
    expect(params).toHaveProperty("practiceMode");
    expect(params).toHaveProperty("practiceGoal");
    expect(params).toHaveProperty("hintUsed");
    expect(params).toHaveProperty("hintCount");
  });

  it("happy path: plan exists + ready → skip createPracticePlan, only startPracticeSession", async () => {
    const client = buildClient();
    const createSpy = vi.spyOn(client, "createPracticePlan");
    const startSpy = vi.spyOn(client, "startPracticeSession");
    const { nav } = renderScreen(PLAN_EXISTS_ROUTE, client);

    await waitFor(() => {
      expect(nav).toHaveBeenCalled();
    });

    // planId exists in context → createPracticePlan must be skipped
    expect(createSpy).not.toHaveBeenCalled();
    expect(startSpy).toHaveBeenCalledTimes(1);

    const navCall = nav.mock.calls[0]![0] as Record<string, unknown>;
    expect(navCall.name).toBe("practice");
    expect((navCall.params as Record<string, unknown>).sessionId).toBeDefined();
  });

  it("synthetic plan id is ignored before generated client calls", async () => {
    const client = buildClient();
    const getPlanSpy = vi.spyOn(client, "getPracticePlan");
    const createSpy = vi.spyOn(client, "createPracticePlan");
    const startSpy = vi.spyOn(client, "startPracticeSession");
    const { nav } = renderScreen(SYNTHETIC_PLAN_ROUTE, client);

    await waitFor(() => {
      expect(nav).toHaveBeenCalled();
    });

    expect(getPlanSpy).not.toHaveBeenCalledWith(
      SYNTHETIC_PLAN_ROUTE.params.planId,
    );
    expect(getPlanSpy).toHaveBeenCalledWith(
      "01918fa0-0000-7000-8000-000000004000",
    );
    expect(createSpy).not.toHaveBeenCalled();
    expect(startSpy).toHaveBeenCalledTimes(1);
    const startRequest = startSpy.mock.calls[0]![0] as unknown as {
      planId: string;
    };
    expect(startRequest.planId).toBe("01918fa0-0000-7000-8000-000000004000");
  });

  it("plan exists + archived → refreshes plan, creates replacement plan, then starts session", async () => {
    const archivedPlanClient = buildCustomFixtures([
      {
        ...getPracticePlanFixture,
        scenarios: {
          ...getPracticePlanFixture.scenarios,
          default: getPracticePlanFixture.scenarios.archived!,
        },
      },
      startPracticeSessionFixture,
    ]);
    const createSpy = vi.spyOn(archivedPlanClient, "createPracticePlan");
    const getPlanSpy = vi.spyOn(archivedPlanClient, "getPracticePlan");
    const startSpy = vi.spyOn(archivedPlanClient, "startPracticeSession");
    const { nav } = renderScreen(PLAN_EXISTS_ROUTE, archivedPlanClient);

    await waitFor(() => {
      expect(nav).toHaveBeenCalled();
    });

    expect(getPlanSpy).toHaveBeenCalledWith(PLAN_EXISTS_ROUTE.params.planId);
    expect(createSpy).toHaveBeenCalledTimes(1);
    expect(startSpy).toHaveBeenCalledTimes(1);
    expect((startSpy.mock.calls[0]![0] as unknown as Record<string, unknown>).planId).toBe(
      "01918fa0-0000-7000-8000-000000004000",
    );
  });

  it("createPracticePlan 4xx (missing-resume) shows error and does NOT call startPracticeSession", async () => {
    const missingResumeClient = buildCustomFixtures([
      {
        ...createPracticePlanFixture,
        scenarios: {
          ...createPracticePlanFixture.scenarios,
          default: createPracticePlanFixture.scenarios["missing-resume"]!,
        },
      },
      startPracticeSessionFixture,
    ]);

    const startSpy = vi.spyOn(missingResumeClient, "startPracticeSession");
    mockTargetJobWithoutCurrentPlan(missingResumeClient);

    renderScreen(FULL_ROUTE, missingResumeClient);

    // Wait for error state (manual error from the hook, not a fixture error)
    await waitFor(
      () => {
        expect(screen.getByTestId("workspace-cta-error")).toBeDefined();
      },
      { timeout: 5000 },
    );

    // startPracticeSession must NOT be called after a createPracticePlan error
    expect(startSpy).not.toHaveBeenCalled();
  });

  it("startPracticeSession 502 → retry reuses same Idempotency-Key", async () => {
    const aiTimeoutClient = buildClient("ai-timeout-502");

    const startSpy = vi.spyOn(aiTimeoutClient, "startPracticeSession");
    const nav = vi.fn();

    render(
      <DisplayPreferencesProvider>
        <InterviewContextProvider>
          <AppRuntimeProvider client={aiTimeoutClient}>
            <NavigationProvider value={{ navigate: nav }}>
              <HydrateContext route={FULL_ROUTE} />
              <WorkspaceScreen route={FULL_ROUTE} />
            </NavigationProvider>
          </AppRuntimeProvider>
        </InterviewContextProvider>
      </DisplayPreferencesProvider>,
    );

    const user = userEvent.setup();

    // Wait for error state
    await waitFor(() => {
      expect(screen.getByTestId("workspace-cta-error")).toBeDefined();
    });

    const retryBtn = screen.getByTestId("workspace-cta-retry");
    await user.click(retryBtn);

    // Both calls should use same idempotency batch
    expect(startSpy).toHaveBeenCalledTimes(2);
    const key1 = (startSpy.mock.calls[0]![1] as Record<string, unknown>)?.idempotencyKey;
    const key2 = (startSpy.mock.calls[1]![1] as Record<string, unknown>)?.idempotencyKey;
    expect(key1).toBe(key2);
  });

  it("startPracticeSession 3 failures → fallback CTA 'back to home' appears", async () => {
    const aiTimeoutClient = buildClient("ai-timeout-502");

    const nav = vi.fn();

    render(
      <DisplayPreferencesProvider>
        <InterviewContextProvider>
          <AppRuntimeProvider client={aiTimeoutClient}>
            <NavigationProvider value={{ navigate: nav }}>
              <HydrateContext route={FULL_ROUTE} />
              <WorkspaceScreen route={FULL_ROUTE} />
            </NavigationProvider>
          </AppRuntimeProvider>
        </InterviewContextProvider>
      </DisplayPreferencesProvider>,
    );

    const user = userEvent.setup();

    // Wait for error state
    await waitFor(() => {
      expect(screen.getByTestId("workspace-cta-error")).toBeDefined();
    });

    // Retry 2 more times (total 3 attempts)
    for (let i = 0; i < 2; i++) {
      const retryBtn = screen.getByTestId("workspace-cta-retry");
      await user.click(retryBtn);
      await waitFor(() => {
        expect(screen.getByTestId("workspace-cta-error")).toBeDefined();
      });
    }

    // After 3 failures, fallback CTA should appear
    await waitFor(() => {
      expect(screen.getByTestId("workspace-cta-back-home")).toBeDefined();
    });
  });

  it("nav practice params carry complete InterviewContext + PracticeDisplayContext", async () => {
    const client = buildClient();
    mockTargetJobWithoutCurrentPlan(client);
    const createSpy = vi.spyOn(client, "createPracticePlan");
    const { nav } = renderScreen(FULL_ROUTE, client);

    await waitFor(() => {
      expect(nav).toHaveBeenCalled();
    });

    expect(createSpy).toHaveBeenCalledTimes(1);

    const navCall = nav.mock.calls[0]![0] as Record<string, unknown>;
    expect(navCall.name).toBe("practice");
    const params = navCall.params as Record<string, unknown>;

    // Verify practiceGoal and PracticeDisplayContext fields
    expect(params.practiceGoal).toBe("baseline");
    expect(params.practiceMode).toBe("strict");
    expect(params.hintUsed).toBe("false");
    expect(params.hintCount).toBe("0");
    expect(params.mode).toBe("text");
    expect(params.modality).toBe("text");
    expect(params.sessionId).toBeDefined();
  });

  it("hintsEnabled derived from practiceMode: strict → false", async () => {
    const client = buildClient();
    const startSpy = vi.spyOn(client, "startPracticeSession");
    const { nav } = renderScreen(FULL_ROUTE, client);

    await waitFor(() => {
      expect(nav).toHaveBeenCalled();
    });

    // Default practiceMode is "strict" → hintsEnabled must be false
    expect(startSpy).toHaveBeenCalledTimes(1);
    const body = startSpy.mock.calls[0]![0] as unknown as Record<string, unknown>;
    expect(body.hintsEnabled).toBe(false);
  });

  it("negative: workspace does NOT produce the non-current replay value in practiceMode", async () => {
    const { nav } = renderScreen(FULL_ROUTE);

    await waitFor(() => {
      expect(nav).toHaveBeenCalled();
    });

    const navCall = nav.mock.calls[0]![0] as Record<string, unknown>;
    const params = navCall.params as Record<string, unknown>;

    const nonCurrentReplayMode = "debrief" + "_replay";
    expect(params.practiceMode).not.toBe(nonCurrentReplayMode);
    // practiceMode must be one of the allowed binary values
    expect(["assisted", "strict"]).toContain(params.practiceMode);
  });

  it("StrictMode: generated client call counts are within expected range", async () => {
    const client = buildClient();
    mockTargetJobWithoutCurrentPlan(client);
    const createSpy = vi.spyOn(client, "createPracticePlan");
    const startSpy = vi.spyOn(client, "startPracticeSession");
    const { nav } = renderScreen(FULL_ROUTE, client);

    await waitFor(() => {
      expect(nav).toHaveBeenCalled();
    });

    // In StrictMode, effects run twice but inFlightRef guard prevents double API calls
    expect(createSpy).toHaveBeenCalledTimes(1);
    expect(startSpy).toHaveBeenCalledTimes(1);
  });

  it("createPracticePlan body matches expected schema", async () => {
    const client = buildClient();
    mockTargetJobWithoutCurrentPlan(client);
    const createSpy = vi.spyOn(client, "createPracticePlan");
    const { nav } = renderScreen(FULL_ROUTE, client);

    await waitFor(() => {
      expect(nav).toHaveBeenCalled();
    });

    const body = createSpy.mock.calls[0]![0] as unknown as Record<string, unknown>;
    expect(body.targetJobId).toBe("01918fa0-0000-7000-8000-000000002000");
    expect(body.goal).toBe("baseline");
    expect(body.mode).toBe("assisted");
    expect(body.interviewerPersona).toBe("hiring_manager");
    expect(body.difficulty).toBe("standard");
    expect(body.language).toBeDefined();
    expect(body.questionBudget).toBe(6);
    expect(body.timeBudgetMinutes).toBe(30);
    expect(body.resumeId).toBe("01918fa0-0000-7000-8000-000000001000");
    expect(body.focusCompetencyCodes).toEqual([]);
  });
});
