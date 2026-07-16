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
import archiveTargetJobFixture from "../../../../../openapi/fixtures/TargetJobs/archiveTargetJob.json";
import listTargetJobsFixture from "../../../../../openapi/fixtures/TargetJobs/listTargetJobs.json";
import getResumeFixture from "../../../../../openapi/fixtures/Resumes/getResume.json";
import listResumesFixture from "../../../../../openapi/fixtures/Resumes/listResumes.json";
import getPracticePlanFixture from "../../../../../openapi/fixtures/PracticePlans/getPracticePlan.json";
import createPracticePlanFixture from "../../../../../openapi/fixtures/PracticePlans/createPracticePlan.json";
import startPracticeSessionFixture from "../../../../../openapi/fixtures/PracticeSessions/startPracticeSession.json";

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
          archiveTargetJobFixture,
          createPracticePlanFixture,
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
    expect((planCard as HTMLElement).style.padding).toBe("20px");
    expect((planCard as HTMLElement).style.position).toBe("relative");
    expect(within(planCard).queryByText(/URL import|Manual input|ZH-CN/i)).toBeNull();
    const rail = screen.getByTestId("workspace-plan-list-rail-01918fa0-0000-7000-8000-000000002000");
    expect(rail).toHaveTextContent("Frontend architecture screen · 45m");
    expect(rail).toHaveTextContent("Hiring manager impact interview · 50m");
    expect(screen.getByTestId("workspace-plan-list-card-body-01918fa0-0000-7000-8000-000000002000")).toBeDefined();
    const cardFooter = screen.getByTestId("workspace-plan-list-card-footer-01918fa0-0000-7000-8000-000000002000");
    expect(cardFooter).toBeDefined();
    expect((cardFooter as HTMLElement).style.borderTop).toBe("1px solid var(--ei-color-rule-strong)");
    expect((cardFooter as HTMLElement).style.background).toBe("var(--ei-color-bg-card)");
    expect((cardFooter as HTMLElement).style.justifyContent).toBe("flex-end");
    expect(cardFooter).toHaveTextContent("Start mock interview");
    expect(cardFooter).not.toHaveTextContent("Open plan");
    expect(
      cardFooter.querySelector(
        "[data-testid='workspace-plan-list-delete-01918fa0-0000-7000-8000-000000002000']",
      ),
    ).toBeNull();
    expect(
      screen.queryByTestId("workspace-plan-list-open-01918fa0-0000-7000-8000-000000002000"),
    ).toBeNull();
    const startButton = screen.getByTestId("workspace-plan-list-start-01918fa0-0000-7000-8000-000000002000");
    expect((startButton as HTMLElement).style.background).toBe("var(--ei-color-accent)");
    expect((startButton as HTMLElement).style.border).toBe("1px solid var(--ei-color-accent)");
    const deleteButton = screen.getByTestId("workspace-plan-list-delete-01918fa0-0000-7000-8000-000000002000");
    expect(deleteButton).toHaveAttribute("aria-label", "Delete");
    expect((deleteButton as HTMLElement).style.position).toBe("absolute");
    expect((deleteButton as HTMLElement).style.right).toBe("14px");
    expect(deleteButton.querySelector('[data-icon="trash"]')).not.toBeNull();
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

  it("plan card selection opens workspace detail with targetJobId only", async () => {
    const user = userEvent.setup();
    const { nav } = renderScreen({ name: "workspace", params: {} });

    await waitFor(() => {
      expect(screen.getByTestId("workspace-plan-list-card-01918fa0-0000-7000-8000-000000002000")).toBeDefined();
    });

    await user.click(screen.getByTestId("workspace-plan-list-card-01918fa0-0000-7000-8000-000000002000"));

    expect(nav).toHaveBeenCalledWith({
      name: "workspace",
      params: {
        targetJobId: "01918fa0-0000-7000-8000-000000002000",
      },
    });
  });

  it("plan card selection never copies list-item resume binding into the detail URL", async () => {
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

    await user.click(screen.getByTestId("workspace-plan-list-card-01918fa0-0000-7000-8000-000000002000"));

    expect(nav).toHaveBeenCalledWith({
      name: "workspace",
      params: {
        targetJobId: "01918fa0-0000-7000-8000-000000002000",
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

  it("quick-starts a plan card without opening parse detail", async () => {
    const user = userEvent.setup();
    const client = clientWithScenarios();
    const getPlanSpy = vi
      .spyOn(client, "getPracticePlan")
      .mockResolvedValue(
        getPracticePlanFixture.scenarios.default.response.body as Awaited<
          ReturnType<EasyInterviewClient["getPracticePlan"]>
        >,
      );
    const createPlanSpy = vi.spyOn(client, "createPracticePlan");
    const startSpy = vi
      .spyOn(client, "startPracticeSession")
      .mockResolvedValue(
        startPracticeSessionFixture.scenarios.default.response.body as Awaited<
          ReturnType<EasyInterviewClient["startPracticeSession"]>
        >,
      );
    const { nav } = renderScreen({ name: "workspace", params: {} }, client);

    await waitFor(() => {
      expect(screen.getByTestId("workspace-plan-list-start-01918fa0-0000-7000-8000-000000002000")).toBeDefined();
    });

    await user.click(screen.getByTestId("workspace-plan-list-start-01918fa0-0000-7000-8000-000000002000"));

    await waitFor(() => {
      expect(startSpy).toHaveBeenCalled();
    });

    expect(getPlanSpy).toHaveBeenCalledWith("01918fa0-0000-7000-8000-000000004000");
    expect(createPlanSpy).not.toHaveBeenCalled();
    expect(nav).toHaveBeenCalledWith({
      name: "practice",
      params: expect.objectContaining({
        targetJobId: "01918fa0-0000-7000-8000-000000002000",
        planId: "01918fa0-0000-7000-8000-000000004000",
        resumeId: "01918fa0-0000-7000-8000-000000001000",
        roundId: "round-2-manager",
        roundName: "Hiring manager impact interview · 50m",
        sessionId: "01918fa0-0000-7000-8000-000000005000",
      }),
    });
    expect(nav).not.toHaveBeenCalledWith(expect.objectContaining({ name: "parse" }));
  });

  it("does not expose a backend error when plan quick-start fails", async () => {
    const user = userEvent.setup();
    const client = clientWithScenarios();
    vi.spyOn(client, "getPracticePlan").mockRejectedValue(
      new Error("HTTP 503 PRACTICE_STORE_UNAVAILABLE"),
    );
    const { nav } = renderScreen({ name: "workspace", params: {} }, client);

    const start = await screen.findByTestId(
      "workspace-plan-list-start-01918fa0-0000-7000-8000-000000002000",
    );
    await user.click(start);

    expect(await screen.findByTestId("workspace-plan-list-start-error")).toHaveTextContent(
      "We couldn't start the mock interview. Try again in a moment.",
    );
    expect(screen.queryByText("HTTP 503 PRACTICE_STORE_UNAVAILABLE")).not.toBeInTheDocument();
    expect(start).toBeEnabled();
    expect(nav).not.toHaveBeenCalledWith(expect.objectContaining({ name: "practice" }));
  });

  it("renders final rail as done and disables quick-start with zero plan/session calls", async () => {
    const client = clientWithScenarios();
    const finished = {
      ...listTargetJobsFixture.scenarios.default.response.body.items[0]!,
      currentPracticePlanId: null,
      practiceProgress: {
        status: "completed" as const,
        completedRounds: [
          { roundId: "round-1-technical", roundSequence: 1 },
          { roundId: "round-2-manager", roundSequence: 2 },
          { roundId: "round-3-culture", roundSequence: 3 },
        ],
        currentRound: null,
      },
    };
    vi.spyOn(client, "listTargetJobs").mockResolvedValue({
      items: [finished],
      pageInfo: { hasMore: false, nextCursor: null, pageSize: 12 },
    } as Awaited<ReturnType<EasyInterviewClient["listTargetJobs"]>>);
    const getTargetSpy = vi.spyOn(client, "getTargetJob");
    const createSpy = vi.spyOn(client, "createPracticePlan");
    const startSpy = vi.spyOn(client, "startPracticeSession");

    renderScreen({ name: "workspace", params: {} }, client);

    const start = await screen.findByTestId(`workspace-plan-list-start-${finished.id}`);
    const rail = screen.getByTestId(`workspace-plan-list-rail-${finished.id}`);
    expect(start).toBeDisabled();
    expect(rail.querySelectorAll('[data-round-state="done"]')).toHaveLength(3);
    expect(rail.querySelectorAll('[data-round-state="current"]')).toHaveLength(0);
    expect(getTargetSpy).not.toHaveBeenCalled();
    expect(createSpy).not.toHaveBeenCalled();
    expect(startSpy).not.toHaveBeenCalled();
  });

  it("archives the plan card through the generated client before removing it", async () => {
    const user = userEvent.setup();
    const client = clientWithScenarios();
    const archiveSpy = vi
      .spyOn(client, "archiveTargetJob")
      .mockResolvedValue(
        archiveTargetJobFixture.scenarios.default.response.body as Awaited<
          ReturnType<EasyInterviewClient["archiveTargetJob"]>
        >,
      );
    const getPlanSpy = vi.spyOn(client, "getPracticePlan");
    const startSpy = vi.spyOn(client, "startPracticeSession");
    const { nav } = renderScreen({ name: "workspace", params: {} }, client);

    await waitFor(() => {
      expect(screen.getByTestId("workspace-plan-list-delete-01918fa0-0000-7000-8000-000000002000")).toBeDefined();
    });

    await user.click(screen.getByTestId("workspace-plan-list-delete-01918fa0-0000-7000-8000-000000002000"));

    await waitFor(() => {
      expect(archiveSpy).toHaveBeenCalled();
    });
    expect(archiveSpy).toHaveBeenCalledWith(
      "01918fa0-0000-7000-8000-000000002000",
      expect.objectContaining({
        idempotencyKey: expect.stringMatching(/^v1\.\d+\.[0-9a-f-]{36}$/),
      }),
    );
    expect(nav).not.toHaveBeenCalled();
    expect(getPlanSpy).not.toHaveBeenCalled();
    expect(startSpy).not.toHaveBeenCalled();
    expect(screen.queryByTestId("workspace-plan-list-card-01918fa0-0000-7000-8000-000000002000")).toBeNull();
  });

  it("keeps the card visible and hides backend details when archiveTargetJob fails", async () => {
    const user = userEvent.setup();
    const client = clientWithScenarios();
    vi.spyOn(client, "archiveTargetJob").mockRejectedValue(new Error("archive failed"));
    const { nav } = renderScreen({ name: "workspace", params: {} }, client);

    await waitFor(() => {
      expect(screen.getByTestId("workspace-plan-list-delete-01918fa0-0000-7000-8000-000000002000")).toBeDefined();
    });

    await user.click(screen.getByTestId("workspace-plan-list-delete-01918fa0-0000-7000-8000-000000002000"));

    await waitFor(() => {
      expect(screen.getByTestId("workspace-plan-list-delete-error")).toHaveTextContent(
        "We couldn't delete the interview plan. Try again in a moment.",
      );
    });
    expect(screen.queryByText("archive failed")).not.toBeInTheDocument();
    expect(
      screen.getByTestId("workspace-plan-list-delete-01918fa0-0000-7000-8000-000000002000"),
    ).toBeEnabled();
    expect(screen.getByTestId("workspace-plan-list-card-01918fa0-0000-7000-8000-000000002000")).toBeDefined();
    expect(nav).not.toHaveBeenCalled();
  });

});
