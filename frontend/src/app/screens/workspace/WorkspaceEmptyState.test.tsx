/**
 * @vitest-environment jsdom
 */

import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
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
import { AppRuntimeContext } from "../../runtime/AppRuntimeProvider";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import type { Route } from "../../routes";
import { WorkspaceScreen } from "./WorkspaceScreen";

import getTargetJobFixture from "../../../../../openapi/fixtures/TargetJobs/getTargetJob.json";
import listTargetJobsFixture from "../../../../../openapi/fixtures/TargetJobs/listTargetJobs.json";
import getResumeFixture from "../../../../../openapi/fixtures/Resumes/getResume.json";
import listResumesFixture from "../../../../../openapi/fixtures/Resumes/listResumes.json";

function clientWithScenarios(opts: {
  targetJobScenario?: string;
  targetJobsScenario?: string;
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
            ...listTargetJobsFixture,
            scenarios: {
              ...listTargetJobsFixture.scenarios,
              default: listTargetJobsFixture.scenarios[(opts.targetJobsScenario ?? "default") as keyof typeof listTargetJobsFixture.scenarios]!,
            },
          },
          {
            ...getResumeFixture,
            scenarios: {
              ...getResumeFixture.scenarios,
              default: getResumeFixture.scenarios[(opts.resumeScenario ?? "default") as keyof typeof getResumeFixture.scenarios]!,
            },
          },
          listResumesFixture,
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
          <AppRuntimeContext.Provider
            value={{
              client,
              runtime: { status: "ready", config: {} as never },
              auth: { status: "authenticated", user: {} as never },
              refreshAuth: vi.fn(),
            }}
          >
            <NavigationProvider value={{ navigate: nav }}>
              <HydrateContext route={route} />
              <WorkspaceScreen route={route} />
            </NavigationProvider>
          </AppRuntimeContext.Provider>
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
  it("renders the interview plan list when targetJobId is missing", async () => {
    const client = clientWithScenarios();
    const listSpy = vi
      .spyOn(client, "listTargetJobs")
      .mockResolvedValue(listTargetJobsFixture.scenarios.default.response.body as Awaited<ReturnType<EasyInterviewClient["listTargetJobs"]>>);
    const targetSpy = vi.spyOn(client, "getTargetJob");
    renderScreen({ name: "workspace", params: {} }, client);

    await waitFor(() => {
      expect(screen.getByTestId("workspace-plan-list")).toBeDefined();
    });

    expect(listSpy).toHaveBeenCalled();
    expect(targetSpy).not.toHaveBeenCalled();
    expect(screen.queryByTestId("workspace-empty")).toBeNull();
    expect(screen.getByTestId("workspace-plan-list-title")).toBeDefined();
    await waitFor(() => {
      expect(screen.getByTestId("workspace-plan-list-card-01918fa0-0000-7000-8000-000000002000")).toBeDefined();
    });
    const planCard = screen.getByTestId("workspace-plan-list-card-01918fa0-0000-7000-8000-000000002000");
    expect(planCard).toBeDefined();
    expect((planCard as HTMLElement).style.background).toBe("var(--ei-color-bg-card)");
    expect((planCard as HTMLElement).style.border).toBe("1px solid var(--ei-color-rule-strong)");
    expect((planCard as HTMLElement).style.boxShadow).toBe("var(--ei-shadow-elev2)");
    expect(within(planCard).queryByText(/URL import|Manual input|ZH-CN/i)).toBeNull();
    expect(screen.getByTestId("workspace-plan-list-card-body-01918fa0-0000-7000-8000-000000002000")).toBeDefined();
    const cardFooter = screen.getByTestId("workspace-plan-list-card-footer-01918fa0-0000-7000-8000-000000002000");
    expect(cardFooter).toBeDefined();
    expect((cardFooter as HTMLElement).style.borderTop).toBe("1px solid var(--ei-color-rule-strong)");
    expect((cardFooter as HTMLElement).style.background).toBe("var(--ei-color-bg-card)");
    expect((cardFooter as HTMLElement).style.justifyContent).toBe("flex-end");
    const openButton = screen.getByTestId("workspace-plan-list-open-01918fa0-0000-7000-8000-000000002000");
    expect((openButton as HTMLElement).style.background).toBe("var(--ei-color-accent)");
    expect((openButton as HTMLElement).style.border).toBe("1px solid var(--ei-color-accent)");
  });

  it("plan-list create CTA navigates to home", async () => {
    const { nav } = renderScreen({ name: "workspace", params: {} });

    await waitFor(() => {
      expect(screen.getByTestId("workspace-plan-list-create")).toBeDefined();
    });

    screen.getByTestId("workspace-plan-list-create").click();
    expect(nav).toHaveBeenCalledWith({ name: "home", params: {} });
  });

  it("plan card selection opens current-plan detail without fabricating resume or report ids", async () => {
    const user = userEvent.setup();
    const { nav } = renderScreen({ name: "workspace", params: {} });

    await waitFor(() => {
      expect(screen.getByTestId("workspace-plan-list-card-01918fa0-0000-7000-8000-000000002000")).toBeDefined();
    });

    await user.click(screen.getByTestId("workspace-plan-list-open-01918fa0-0000-7000-8000-000000002000"));

    expect(nav).toHaveBeenCalledWith({
      name: "workspace",
      params: {
        targetJobId: "01918fa0-0000-7000-8000-000000002000",
        jobId: "01918fa0-0000-7000-8000-000000002000",
        jdId: "jd-01918fa0-0000-7000-8000-000000002000",
        planId: "01918fa0-0000-7000-8000-000000004000",
        resumeId: "01918fa0-0000-7000-8000-000000001000",
      },
    });
  });

  it("plan card selection carries target-job resume binding even before a practice plan exists", async () => {
    const user = userEvent.setup();
    const client = clientWithScenarios();
    vi.spyOn(client, "listTargetJobs").mockResolvedValue({
      items: [
        {
          ...listTargetJobsFixture.scenarios.default.response.body.items[0],
          currentPracticePlanId: null,
          resumeId: "01918fa0-0000-7000-8000-000000001000",
        },
      ],
      pageInfo: {
        hasMore: false,
        nextCursor: null,
        pageSize: 12,
      },
    } as Awaited<ReturnType<EasyInterviewClient["listTargetJobs"]>>);
    const { nav } = renderScreen({ name: "workspace", params: {} }, client);

    await waitFor(() => {
      expect(screen.getByTestId("workspace-plan-list-card-01918fa0-0000-7000-8000-000000002000")).toBeDefined();
    });

    await user.click(screen.getByTestId("workspace-plan-list-open-01918fa0-0000-7000-8000-000000002000"));

    expect(nav).toHaveBeenCalledWith({
      name: "workspace",
      params: {
        targetJobId: "01918fa0-0000-7000-8000-000000002000",
        jobId: "01918fa0-0000-7000-8000-000000002000",
        jdId: "jd-01918fa0-0000-7000-8000-000000002000",
        resumeId: "01918fa0-0000-7000-8000-000000001000",
      },
    });
  });

  it("renders a friendly plan-list empty state when no plans exist", async () => {
    renderScreen(
      { name: "workspace", params: {} },
      clientWithScenarios({ targetJobsScenario: "empty" }),
    );

    await waitFor(() => {
      expect(screen.getByTestId("workspace-plan-list-empty")).toBeDefined();
    });

    expect(screen.queryByTestId("workspace-empty")).toBeNull();
    expect(screen.getByTestId("workspace-plan-list-create")).toBeDefined();
  });

  it("renders unified detail error state when getTargetJob returns not found", async () => {
    renderScreen(
      {
        name: "workspace",
        params: {
          targetJobId: "01918fa0-0000-7000-8000-000000009999",
          resumeId: "01918fa0-0000-7000-8000-000000001000",
        },
      },
      clientWithScenarios({ targetJobScenario: "not-found" }),
    );

    await waitFor(() => {
      expect(screen.getByTestId("parse-error")).toBeDefined();
    });

    expect(screen.getByTestId("route-workspace")).toBeDefined();
    expect(screen.queryByTestId("workspace-empty")).toBeNull();
    expect(screen.queryByTestId("workspace-launcher")).toBeNull();
  });

  it("renders unified detail error state when getTargetJob returns 5xx", async () => {
    const client = clientWithScenarios({ targetJobScenario: "5xx" });
    const spy = vi.spyOn(client, "getTargetJob");
    renderScreen(
      {
        name: "workspace",
        params: {
          targetJobId: "01918fa0-0000-7000-8000-000000002000",
          resumeId: "01918fa0-0000-7000-8000-000000001000",
        },
      },
      client,
    );

    await waitFor(() => {
      expect(screen.getByTestId("parse-error")).toBeDefined();
    });

    expect(spy).toHaveBeenCalled();
    expect(screen.queryByTestId("workspace-empty")).toBeNull();
    expect(screen.queryByTestId("workspace-launcher")).toBeNull();
  });
});

describe("WorkspaceMissingResumeState", () => {
  it("does not render when getTargetJob returns target job-level resume binding", async () => {
    renderScreen({
      name: "workspace",
      params: {
        targetJobId: "01918fa0-0000-7000-8000-000000002000",
      },
    });

    await waitFor(() => {
      expect(screen.getByTestId("unified-plan-detail")).toBeDefined();
    });

    expect(screen.queryByTestId("workspace-missing-resume")).toBeNull();
    expect(screen.getByTestId("parse-action-start-interview")).toBeEnabled();
  });

  it("blocks Save/Start in unified detail when targetJobId exists but resumeId is missing", async () => {
    const client = clientWithScenarios();
    vi.spyOn(client, "getTargetJob").mockResolvedValue({
      ...getTargetJobFixture.scenarios.default.response.body,
      resumeId: null,
    } as Awaited<ReturnType<EasyInterviewClient["getTargetJob"]>>);
    renderScreen({
      name: "workspace",
      params: {
        targetJobId: "01918fa0-0000-7000-8000-000000002000",
      },
    }, client);

    await waitFor(() => {
      expect(screen.getByTestId("parse-resume-required")).toBeDefined();
    });

    expect(screen.queryByTestId("workspace-missing-resume")).toBeNull();
    expect(screen.getByTestId("parse-action-save-plan")).toBeDisabled();
    expect(screen.getByTestId("parse-action-start-interview")).toBeDisabled();
  });

  it("resume create CTA is only shown by unified detail when no selectable resume exists", async () => {
    const client = clientWithScenarios();
    vi.spyOn(client, "getTargetJob").mockResolvedValue({
      ...getTargetJobFixture.scenarios.default.response.body,
      resumeId: null,
    } as Awaited<ReturnType<EasyInterviewClient["getTargetJob"]>>);
    vi.spyOn(client, "listResumes").mockResolvedValue({
      items: [],
      pageInfo: { hasMore: false, nextCursor: null, pageSize: 20 },
    } as Awaited<ReturnType<EasyInterviewClient["listResumes"]>>);
    const { nav } = renderScreen({
      name: "workspace",
      params: {
        targetJobId: "01918fa0-0000-7000-8000-000000002000",
      },
    }, client);

    await waitFor(() => {
      expect(screen.getByTestId("parse-resume-create")).toBeDefined();
    });

    screen.getByTestId("parse-resume-create").click();
    expect(nav).toHaveBeenCalledWith({
      name: "resume_versions",
      params: {
        flow: "create",
        targetJobId: "01918fa0-0000-7000-8000-000000002000",
      },
    });
  });
});
