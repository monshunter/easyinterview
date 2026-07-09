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
import {
  InterviewContextProvider,
  useInterviewContext,
} from "../../interview-context/InterviewContext";
import { AppRuntimeProvider } from "../../runtime/AppRuntimeProvider";
import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";
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
    planId: "01918fa0-0000-7000-8000-000000004000",
    resumeId: "01918fa0-0000-7000-8000-000000001000",
    roundId: "round-hr",
  },
};

function HydrateRoute({ route }: { route: Route }) {
  const { dispatch } = useInterviewContext();
  useEffect(() => {
    dispatch({ type: "HYDRATE_FROM_ROUTE", params: route.params });
  }, [dispatch, route.params]);
  return null;
}

function buildClient() {
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

function renderWorkspace(route: Route = WORKSPACE_ROUTE) {
  const nav = vi.fn();
  return render(
    <DisplayPreferencesProvider initial={{ lang: "en" }}>
      <InterviewContextProvider>
        <AppRuntimeProvider
          client={buildClient()}
          requestOptions={{ getMe: { headers: { Prefer: "example=authenticated" } } }}
        >
          <NavigationProvider value={{ navigate: nav }}>
            <HydrateRoute route={route} />
            <WorkspaceScreen route={route} />
          </NavigationProvider>
        </AppRuntimeProvider>
      </InterviewContextProvider>
    </DisplayPreferencesProvider>,
  );
}

describe("Workspace unified plan detail data rendering", () => {
  it("renders TargetJob basics, requirements, risks, and strengths through the unified detail", async () => {
    renderWorkspace();

    await waitFor(() => {
      expect(screen.getByTestId("unified-plan-detail")).toBeDefined();
    });

    expect(screen.getByTestId("route-workspace")).toBeDefined();
    expect(screen.getByTestId("unified-plan-detail-title")).toHaveTextContent(
      "Interview plan detail",
    );
    const titleInput = screen
      .getByTestId("parse-basics-title")
      .querySelector("input");
    const companyInput = screen
      .getByTestId("parse-basics-company")
      .querySelector("input");
    expect(titleInput).toHaveValue("Senior Frontend Engineer");
    expect(companyInput).toHaveValue("Acme");
    expect(screen.getByTestId("parse-requirement-must_have-0")).toHaveTextContent(
      "5+ years building component libraries",
    );
    expect(screen.getByTestId("parse-requirement-nice_to_have-0")).toHaveTextContent(
      "edge runtime deployments",
    );
    expect(screen.getByTestId("parse-hidden-signal-0")).toHaveTextContent(
      "scaling design systems",
    );
    expect(screen.getByTestId("parse-launch")).toBeDefined();
  });

  it("does not render the retired independent workspace detail anchors", async () => {
    renderWorkspace();

    await waitFor(() => {
      expect(screen.getByTestId("unified-plan-detail")).toBeDefined();
    });

    expect(screen.queryByTestId("workspace-header")).toBeNull();
    expect(screen.queryByTestId("workspace-jd-card")).toBeNull();
    expect(screen.queryByTestId("workspace-prep-card")).toBeNull();
    expect(screen.queryByTestId("workspace-history-card")).toBeNull();
  });
});
