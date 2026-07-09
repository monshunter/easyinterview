/**
 * @vitest-environment jsdom
 */

import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
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
import getTargetJobFixture from "../../../../../openapi/fixtures/TargetJobs/getTargetJob.json";
import listResumesFixture from "../../../../../openapi/fixtures/Resumes/listResumes.json";

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
        getTargetJobFixture,
        listResumesFixture,
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
  return render(
    <DisplayPreferencesProvider>
      <InterviewContextProvider>
        <AppRuntimeProvider
          client={client}
          requestOptions={{ getMe: { headers: { Prefer: "example=authenticated" } } }}
        >
          <NavigationProvider value={{ navigate: vi.fn() }}>
            <HydrateContext route={route} />
            <WorkspaceScreen route={route} />
          </NavigationProvider>
        </AppRuntimeProvider>
      </InterviewContextProvider>
    </DisplayPreferencesProvider>,
  );
}

describe("Workspace retired modal integration", () => {
  it("ordinary unified detail does not expose the old plan switcher or resume picker actions", async () => {
    renderWorkspace();

    await waitFor(() => {
      expect(screen.getByTestId("unified-plan-detail")).toBeInTheDocument();
    });

    expect(screen.queryByTestId("workspace-plan-action-switch")).toBeNull();
    expect(screen.queryByTestId("workspace-binding-resume-change")).toBeNull();
    expect(screen.queryByTestId("workspace-plan-modal-card")).toBeNull();
    expect(screen.queryByTestId("workspace-resume-modal-card")).toBeNull();
  });
});
