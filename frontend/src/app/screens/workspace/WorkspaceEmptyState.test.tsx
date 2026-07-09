/**
 * @vitest-environment jsdom
 */

import { describe, expect, it, vi } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useEffect, useState, type ReactNode } from "react";

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

function renderScreen(
  route: Route,
  client = clientWithScenarios(),
  opts: { seedParams?: Record<string, string> } = {},
) {
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
              <ContextGate seedParams={opts.seedParams}>
                <HydrateContext route={route} />
                <WorkspaceScreen route={route} />
              </ContextGate>
            </NavigationProvider>
          </AppRuntimeContext.Provider>
        </InterviewContextProvider>
      </DisplayPreferencesProvider>,
    ),
  };
}

function ContextGate({
  seedParams,
  children,
}: {
  seedParams?: Record<string, string>;
  children: ReactNode;
}) {
  const { dispatch } = useInterviewContext();
  const [ready, setReady] = useState(!seedParams);
  useEffect(() => {
    if (seedParams) {
      dispatch({ type: "HYDRATE_FROM_ROUTE", params: seedParams });
      setReady(true);
    }
  }, [dispatch, seedParams]);
  return ready ? <>{children}</> : null;
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

  it("requests ready plans and filters failed or blank-title dirty records", async () => {
    const client = clientWithScenarios();
    const readyJob = listTargetJobsFixture.scenarios.default.response.body.items[0]!;
    const failedJob = {
      ...readyJob,
      id: "01918fa0-0000-7000-8000-000000009991",
      title: "Failed JD",
      analysisStatus: "failed",
    };
    const blankJob = {
      ...readyJob,
      id: "01918fa0-0000-7000-8000-000000009992",
      title: "   ",
      analysisStatus: "ready",
    };
    const listSpy = vi.spyOn(client, "listTargetJobs").mockResolvedValue({
      items: [readyJob, failedJob, blankJob],
      pageInfo: { hasMore: false, nextCursor: null, pageSize: 12 },
    } as Awaited<ReturnType<EasyInterviewClient["listTargetJobs"]>>);

    renderScreen({ name: "workspace", params: {} }, client);

    await waitFor(() => {
      expect(screen.getByTestId(`workspace-plan-list-card-${readyJob.id}`)).toBeDefined();
    });

    expect(listSpy).toHaveBeenCalledWith({
      query: expect.objectContaining({
        analysisStatus: "ready",
        pageSize: "12",
      }),
    });
    expect(screen.queryByTestId("workspace-plan-list-card-01918fa0-0000-7000-8000-000000009991")).toBeNull();
    expect(screen.queryByTestId("workspace-plan-list-card-01918fa0-0000-7000-8000-000000009992")).toBeNull();
    expect(screen.queryByText("Failed JD")).toBeNull();
  });

  it("ignores stale detail context when route has no workspace params", async () => {
    const client = clientWithScenarios();
    const listSpy = vi.spyOn(client, "listTargetJobs");
    const targetSpy = vi.spyOn(client, "getTargetJob");

    renderScreen(
      { name: "workspace", params: {} },
      client,
      {
        seedParams: {
          targetJobId: "01918fa0-0000-7000-8000-000000002000",
          jobId: "01918fa0-0000-7000-8000-000000002000",
          planId: "01918fa0-0000-7000-8000-000000004000",
          resumeId: "01918fa0-0000-7000-8000-000000001000",
        },
      },
    );

    await waitFor(() => {
      expect(screen.getByTestId("workspace-plan-list")).toBeDefined();
      expect(listSpy).toHaveBeenCalled();
    });

    expect(targetSpy).not.toHaveBeenCalled();
    expect(screen.queryByTestId("parse-error")).toBeNull();
    expect(screen.queryByText("缺少目标岗位 ID.")).toBeNull();
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
      name: "parse",
      params: {
        targetJobId: "01918fa0-0000-7000-8000-000000002000",
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
          ...listTargetJobsFixture.scenarios.default.response.body.items[0]!,
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
      name: "parse",
      params: {
        targetJobId: "01918fa0-0000-7000-8000-000000002000",
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

});
