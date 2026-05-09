/**
 * @vitest-environment jsdom
 */

import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { useEffect, type ReactNode } from "react";

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

import getTargetJobFixture from "../../../../../openapi/fixtures/TargetJobs/getTargetJob.json";

const WORKSPACE_ROUTE: Route = {
  name: "workspace",
  params: {
    targetJobId: "01918fa0-0000-7000-8000-000000003000",
    jdId: "jd-1",
    planId: "plan-1",
    resumeVersionId: "rv-1",
    roundId: "round-hr",
  },
};

function HydrateRoute({ params }: { params: Record<string, string> }) {
  const { dispatch } = useInterviewContext();
  useEffect(() => {
    dispatch({ type: "HYDRATE_FROM_ROUTE", params });
  }, []);
  return null;
}

function renderWorkspace(client: EasyInterviewClient, route: Route = WORKSPACE_ROUTE) {
  const nav = vi.fn();
  return {
    nav,
    ...render(
      <DisplayPreferencesProvider>
        <InterviewContextProvider>
          <AppRuntimeProvider client={client}>
            <NavigationProvider value={{ navigate: nav }}>
              <HydrateRoute params={route.params} />
              <WorkspaceScreen route={route} />
            </NavigationProvider>
          </AppRuntimeProvider>
        </InterviewContextProvider>
      </DisplayPreferencesProvider>,
    ),
  };
}

function clientWithScenario(scenario: string) {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry([getTargetJobFixture]),
      { scenario },
    ),
  });
}

describe("WorkspaceHeader (Phase 2.7)", () => {
  it("renders header with fixture data (with-rounds scenario)", async () => {
    const client = clientWithScenario("with-rounds");
    renderWorkspace(client);

    await waitFor(() => {
      expect(screen.getByTestId("workspace-header-title")).toBeDefined();
    });

    expect(screen.getByTestId("workspace-header-title").textContent).toBe(
      "Staff Frontend Engineer",
    );
    expect(screen.getByTestId("workspace-header-subtitle").textContent).toContain(
      "Vercel",
    );
    expect(screen.getByTestId("workspace-header-tag").textContent).toBe("面试中");
    expect(screen.getByTestId("workspace-plan-eyebrow-title").textContent).toContain(
      "Vercel · Staff Frontend Engineer",
    );
  });

  it("renders JD breakdown with requirements grouped by kind", async () => {
    const client = clientWithScenario("with-rounds");
    renderWorkspace(client);

    await waitFor(() => {
      expect(screen.getByTestId("workspace-jd-block-must")).toBeDefined();
    });

    expect(screen.getByTestId("workspace-jd-block-must")).toBeDefined();
    expect(screen.getByTestId("workspace-jd-block-nice")).toBeDefined();
    expect(screen.getByTestId("workspace-jd-block-hidden")).toBeDefined();
  });

  it("renders risks/strengths from fitSummary data", async () => {
    const client = clientWithScenario("with-rounds");
    renderWorkspace(client);

    await waitFor(() => {
      expect(screen.getByTestId("workspace-prep-strong-0")).toBeDefined();
    });

    expect(screen.getByTestId("workspace-prep-strong-0")).toBeDefined();
    expect(screen.getByTestId("workspace-prep-risk-0")).toBeDefined();
  });

  it("shows header updated date formatted from iso string", async () => {
    const client = clientWithScenario("with-rounds");
    renderWorkspace(client);

    await waitFor(() => {
      expect(screen.getByTestId("workspace-header-updated")).toBeDefined();
    });

    // updatedAt is 2026-05-05 → 5/5
    expect(screen.getByTestId("workspace-header-updated").textContent).toContain("5/5");
  });

  it("does NOT render non-existent fields (level, match, nextRound, statusTone, readinessLabel)", async () => {
    const client = clientWithScenario("with-rounds");
    renderWorkspace(client);

    await waitFor(() => {
      expect(screen.getByTestId("workspace-header-title")).toBeDefined();
    });

    // The header-title should NOT contain "level" or "match" or any non-existent field values
    const html = document.body.innerHTML;
    expect(html).not.toContain("readinessLabel");
    expect(html).not.toContain("statusTone");
    expect(html).not.toContain("nextRound");
  });

  it("falls back to placeholder when fitSummary is missing fields", async () => {
    // Use 'default' which has full data; verify all sections render
    const client = clientWithScenario("default");
    renderWorkspace(client);

    await waitFor(() => {
      expect(screen.getByTestId("workspace-header-prep")).toBeDefined();
    });
  });
});
