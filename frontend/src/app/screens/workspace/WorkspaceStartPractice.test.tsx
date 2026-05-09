/**
 * @vitest-environment jsdom
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

function buildClient(startScenario = "default") {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([
        getTargetJobFixture,
        getResumeFixture,
        createPracticePlanFixture,
        {
          ...startPracticeSessionFixture,
          scenarios: {
            ...startPracticeSessionFixture.scenarios,
            default: startPracticeSessionFixture.scenarios[startScenario]!,
          },
        },
        getPracticePlanFixture,
      ]),
      { scenario: "default" },
    ),
  });
}

function renderScreen(route: Route) {
  const client = buildClient();
  const nav = vi.fn();
  return {
    client,
    nav,
    ...render(
      <DisplayPreferencesProvider>
        <InterviewContextProvider>
          <AppRuntimeProvider client={client}>
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

function HydrateContext({ route }: { route: Route }) {
  const { dispatch } = useInterviewContext();
  useEffect(() => {
    dispatch({ type: "HYDRATE_FROM_ROUTE", params: route.params });
  }, []);
  return null;
}

const FULL_ROUTE: Route = {
  name: "workspace",
  params: {
    targetJobId: "01918fa0-0000-7000-8000-000000002000",
    jdId: "jd-1",
    resumeVersionId: "01918fa0-0000-7000-8000-000000001000",
    roundId: "round-hr",
    planId: "",
  },
};

describe("WorkspaceStartPractice (Phase 4.7)", () => {
  it("happy path: no plan → createPracticePlan → startPracticeSession → nav practice", async () => {
    const { nav, client } = renderScreen(FULL_ROUTE);
    const createSpy = vi.spyOn(client, "createPracticePlan");
    const startSpy = vi.spyOn(client, "startPracticeSession");
    const user = userEvent.setup();

    await waitFor(() => {
      expect(screen.getByTestId("workspace-cta-start")).toBeDefined();
    });

    await user.click(screen.getByTestId("workspace-cta-start"));

    await waitFor(() => {
      expect(nav).toHaveBeenCalled();
    });

    expect(createSpy).toHaveBeenCalled();
    expect(startSpy).toHaveBeenCalled();

    const navCall = nav.mock.calls[0]![0] as Record<string, unknown>;
    expect(navCall.name).toBe("practice");
    expect(navCall.params).toHaveProperty("sessionId");
  });

  it("idempotency keys are stable across retries", async () => {
    const aiTimeoutClient = new EasyInterviewClient({
      fetch: createFixtureBackedFetch(
        createFixtureRegistry([
          getTargetJobFixture,
          getResumeFixture,
          createPracticePlanFixture,
          {
            ...startPracticeSessionFixture,
            scenarios: {
              ...startPracticeSessionFixture.scenarios,
              default: startPracticeSessionFixture.scenarios["ai-timeout-502"]!,
            },
          },
          getPracticePlanFixture,
        ]),
        { scenario: "default" },
      ),
    });

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

    await waitFor(() => {
      expect(screen.getByTestId("workspace-cta-start")).toBeDefined();
    });

    await user.click(screen.getByTestId("workspace-cta-start"));

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
});
