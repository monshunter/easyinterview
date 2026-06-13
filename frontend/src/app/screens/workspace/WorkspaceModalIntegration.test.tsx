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
import { NavigationProvider } from "../../navigation/NavigationProvider";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";
import type { Route } from "../../routes";
import { WorkspaceScreen } from "./WorkspaceScreen";

import getRuntimeConfigFixture from "../../../../../openapi/fixtures/Auth/getRuntimeConfig.json";
import getMeFixture from "../../../../../openapi/fixtures/Auth/getMe.json";
import listTargetJobsFixture from "../../../../../openapi/fixtures/TargetJobs/listTargetJobs.json";
import getTargetJobFixture from "../../../../../openapi/fixtures/TargetJobs/getTargetJob.json";
import getResumeFixture from "../../../../../openapi/fixtures/Resumes/getResume.json";
import getPracticePlanFixture from "../../../../../openapi/fixtures/PracticePlans/getPracticePlan.json";
import createPracticePlanFixture from "../../../../../openapi/fixtures/PracticePlans/createPracticePlan.json";
import startPracticeSessionFixture from "../../../../../openapi/fixtures/PracticeSessions/startPracticeSession.json";

const WORKSPACE_ROUTE: Route = {
  name: "workspace",
  params: {
    targetJobId: "01918fa0-0000-7000-8000-000000002000",
    jdId: "jd-1",
    resumeId: "01918fa0-0000-7000-8000-000000001000",
    roundId: "round-hr",
    planId: "01918fa0-0000-7000-8000-000000004000",
  },
};

function buildClient(): EasyInterviewClient {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([
        getRuntimeConfigFixture,
        getMeFixture,
        listTargetJobsFixture,
        getTargetJobFixture,
        getResumeFixture,
        getPracticePlanFixture,
        createPracticePlanFixture,
        startPracticeSessionFixture,
      ]),
      { scenario: "default" },
    ),
  });
}

function HydrateContext({ route }: { route: Route }) {
  const { dispatch } = useInterviewContext();
  useEffect(() => {
    dispatch({ type: "HYDRATE_FROM_ROUTE", params: route.params });
  }, [dispatch, route.params]);
  return null;
}

function renderWorkspace(route: Route = WORKSPACE_ROUTE) {
  const client = buildClient();
  const nav = vi.fn();
  return {
    client,
    nav,
    ...render(
      <DisplayPreferencesProvider>
        <InterviewContextProvider>
          <AppRuntimeProvider
            client={client}
            requestOptions={{ getMe: { headers: { Prefer: "example=authenticated" } } }}
          >
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

describe("WorkspaceScreen modal integration", () => {
  it("opens PlanSwitcherModal from the switch plan action", async () => {
    renderWorkspace();
    const user = userEvent.setup();

    await waitFor(() => {
      expect(screen.getByTestId("workspace-plan-action-switch")).toBeInTheDocument();
    });

    await user.click(screen.getByTestId("workspace-plan-action-switch"));

    await waitFor(() => {
      expect(screen.getByTestId("workspace-plan-modal-card")).toBeInTheDocument();
    });
    expect(screen.getByTestId("workspace-plan-modal-confirm")).toBeInTheDocument();
  });

  it("opens ResumePickerModal from the change resume action", async () => {
    renderWorkspace();
    const user = userEvent.setup();

    await waitFor(() => {
      expect(screen.getByTestId("workspace-binding-resume-change")).toBeInTheDocument();
    });

    await user.click(screen.getByTestId("workspace-binding-resume-change"));

    await waitFor(() => {
      expect(screen.getByTestId("workspace-resume-modal-card")).toBeInTheDocument();
    });
    expect(screen.getByTestId("workspace-resume-modal-confirm")).toBeInTheDocument();
  });
});
