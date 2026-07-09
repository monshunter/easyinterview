/**
 * @vitest-environment jsdom
 */

import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { useEffect, type ReactNode } from "react";

import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import {
  InterviewContextProvider,
  useInterviewContext,
} from "../../interview-context/InterviewContext";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import type { Route } from "../../routes";
import { WorkspaceScreen } from "./WorkspaceScreen";

function HydrateRoute({ route }: { route: Route }) {
  const { dispatch } = useInterviewContext();
  useEffect(() => {
    dispatch({ type: "HYDRATE_FROM_ROUTE", params: route.params });
  }, [dispatch, route.params]);
  return null;
}

function withProviders(ui: ReactNode, route: Route) {
  const nav = vi.fn();
  return {
    nav,
    ...render(
      <DisplayPreferencesProvider>
        <InterviewContextProvider>
          <NavigationProvider value={{ navigate: nav }}>
            <HydrateRoute route={route} />
            {ui}
          </NavigationProvider>
        </InterviewContextProvider>
      </DisplayPreferencesProvider>,
    ),
  };
}

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

describe("WorkspaceScreen route split", () => {
  it("renders the no-context workspace as the interview plan-list landing", async () => {
    const route = { name: "workspace", params: {} } as const;
    withProviders(<WorkspaceScreen route={route} />, route);

    await waitFor(() => {
      expect(screen.getByTestId("workspace-plan-list")).toBeDefined();
    });
    expect(screen.getByTestId("workspace-plan-list-empty")).toBeDefined();
    expect(screen.queryByTestId("workspace-empty")).toBeNull();
    expect(screen.queryByTestId("workspace-plan-eyebrow")).toBeNull();
  });

  it("ordinary context route delegates to the unified plan-detail mother page", () => {
    withProviders(<WorkspaceScreen route={WORKSPACE_ROUTE} />, WORKSPACE_ROUTE);

    expect(screen.getByTestId("route-workspace")).toBeDefined();
    expect(screen.getByTestId("parse-loading-step-0")).toBeDefined();
    expect(screen.queryByTestId("workspace-header")).toBeNull();
    expect(screen.queryByTestId("workspace-launcher")).toBeNull();
    expect(screen.queryByTestId("workspace-jd-card")).toBeNull();
    expect(screen.queryByTestId("workspace-prep-card")).toBeNull();
    expect(screen.queryByTestId("workspace-history-card")).toBeNull();
  });

  it("keeps non-current prototype testids out of the workspace route", () => {
    withProviders(<WorkspaceScreen route={WORKSPACE_ROUTE} />, WORKSPACE_ROUTE);

    expect(screen.queryByTestId("practice-mode-card-warmup")).toBeNull();
    expect(screen.queryByTestId("practice-mode-card-single_drill")).toBeNull();
    expect(screen.queryByTestId("growth-center")).toBeNull();
    expect(screen.queryByTestId("drill-builder-daily")).toBeNull();
    expect(screen.queryByTestId("mistake-queue-entry")).toBeNull();
  });
});
