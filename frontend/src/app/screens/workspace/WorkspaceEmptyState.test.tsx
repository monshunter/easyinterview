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

function clientWithScenarios(opts: {
  targetJobScenario?: string;
  resumeScenario?: string;
} = {}) {
  return new EasyInterviewClient({
    fetch: createFixtureBackedFetch(
      createFixtureRegistry(
        [
          {
            ...getTargetJobFixture,
            scenarios: {
              ...getTargetJobFixture.scenarios,
              default: getTargetJobFixture.scenarios[(opts.targetJobScenario ?? "default") as keyof typeof getTargetJobFixture.scenarios]!,
            },
          },
          {
            ...getResumeFixture,
            scenarios: {
              ...getResumeFixture.scenarios,
              default: getResumeFixture.scenarios[(opts.resumeScenario ?? "default") as keyof typeof getResumeFixture.scenarios]!,
            },
          },
        ],
      ),
      { scenario: "default" },
    ),
  });
}

function renderScreen(route: Route, client = clientWithScenarios()) {
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

describe("WorkspaceEmptyState", () => {
  it("renders when targetJobId is missing", async () => {
    renderScreen({ name: "workspace", params: {} });

    // Without targetJobId, the hook returns empty immediately
    await waitFor(() => {
      expect(screen.getByTestId("workspace-empty")).toBeDefined();
    });

    expect(screen.getByTestId("workspace-empty-eyebrow")).toBeDefined();
    expect(screen.getByTestId("workspace-empty-title")).toBeDefined();
    expect(screen.getByTestId("workspace-empty-cta")).toBeDefined();
  });

  it("CTA navigates to home", async () => {
    const { nav } = renderScreen({ name: "workspace", params: {} });

    await waitFor(() => {
      expect(screen.getByTestId("workspace-empty-cta")).toBeDefined();
    });

    screen.getByTestId("workspace-empty-cta").click();
    expect(nav).toHaveBeenCalledWith({ name: "home", params: {} });
  });

  it("renders recovery empty state when getTargetJob returns not found", async () => {
    renderScreen(
      {
        name: "workspace",
        params: {
          targetJobId: "01918fa0-0000-7000-8000-000000009999",
          resumeVersionId: "01918fa0-0000-7000-8000-000000001000",
        },
      },
      clientWithScenarios({ targetJobScenario: "not-found" }),
    );

    await waitFor(() => {
      expect(screen.getByTestId("workspace-empty")).toBeDefined();
    });

    expect(screen.queryByTestId("workspace-launcher")).toBeNull();
    expect(screen.getByTestId("workspace-empty-cta")).toBeDefined();
  });

  it("renders recoverable target error state with retry when getTargetJob returns 5xx", async () => {
    const user = userEvent.setup();
    const client = clientWithScenarios({ targetJobScenario: "5xx" });
    const spy = vi.spyOn(client, "getTargetJob");
    renderScreen(
      {
        name: "workspace",
        params: {
          targetJobId: "01918fa0-0000-7000-8000-000000002000",
          resumeVersionId: "01918fa0-0000-7000-8000-000000001000",
        },
      },
      client,
    );

    await waitFor(() => {
      expect(screen.getByTestId("workspace-target-error")).toBeDefined();
    });

    expect(screen.queryByTestId("workspace-empty")).toBeNull();
    expect(screen.queryByTestId("workspace-launcher")).toBeNull();

    await user.click(screen.getByTestId("workspace-target-error-retry"));
    await waitFor(() => {
      expect(spy.mock.calls.length).toBeGreaterThanOrEqual(2);
    });
  });
});

describe("WorkspaceMissingResumeState", () => {
  it("renders when targetJobId exists but resumeVersionId missing", async () => {
    renderScreen({
      name: "workspace",
      params: {
        targetJobId: "01918fa0-0000-7000-8000-000000002000",
      },
    });

    await waitFor(() => {
      expect(screen.getByTestId("workspace-missing-resume")).toBeDefined();
    });

    expect(screen.getByTestId("workspace-missing-resume-eyebrow")).toBeDefined();
    expect(screen.getByTestId("workspace-missing-resume-title")).toBeDefined();
    expect(screen.getByTestId("workspace-missing-resume-cta")).toBeDefined();
  });

  it("CTA navigates to resume_versions with flow=create", async () => {
    const { nav } = renderScreen({
      name: "workspace",
      params: {
        targetJobId: "01918fa0-0000-7000-8000-000000002000",
      },
    });

    await waitFor(() => {
      expect(screen.getByTestId("workspace-missing-resume-cta")).toBeDefined();
    });

    screen.getByTestId("workspace-missing-resume-cta").click();
    expect(nav).toHaveBeenCalledWith({
      name: "resume_versions",
      params: { flow: "create" },
    });
  });
});
