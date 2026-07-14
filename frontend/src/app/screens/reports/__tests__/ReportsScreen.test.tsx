// @vitest-environment jsdom
import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { useLayoutEffect, type ReactNode } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import type { EasyInterviewClient } from "../../../../api/generated/client";
import type {
  TargetJob,
  TargetJobReportsOverview,
} from "../../../../api/generated/types";
import { DisplayPreferencesProvider } from "../../../display/DisplayPreferencesProvider";
import { NavigationProvider } from "../../../navigation/NavigationProvider";
import type { Route } from "../../../routes";
import {
  AppRuntimeContext,
  type AppRuntimeValue,
} from "../../../runtime/AppRuntimeProvider";
import { ReportsScreen } from "../ReportsScreen";

const TARGET_A_ID = "01918fa0-0000-7000-8000-000000002000";
const TARGET_B_ID = "01918fa0-0000-7000-8000-000000002001";
const TARGET_OTHER_ID = "01918fa0-0000-7000-8000-000000002099";
const REPORT_IDS = {
  a: {
    current: "01918fa0-0000-7000-8000-000000007101",
    currentSecond: "01918fa0-0000-7000-8000-000000007106",
    failed: "01918fa0-0000-7000-8000-000000007102",
    generating: "01918fa0-0000-7000-8000-000000007103",
    same: "01918fa0-0000-7000-8000-000000007104",
    latestReady: "01918fa0-0000-7000-8000-000000007105",
  },
  b: {
    current: "01918fa0-0000-7000-8000-000000007201",
    currentSecond: "01918fa0-0000-7000-8000-000000007206",
    failed: "01918fa0-0000-7000-8000-000000007202",
    generating: "01918fa0-0000-7000-8000-000000007203",
    same: "01918fa0-0000-7000-8000-000000007204",
    latestReady: "01918fa0-0000-7000-8000-000000007205",
  },
} as const;

function targetJob(
  id = TARGET_A_ID,
  roundName = "Architecture",
): TargetJob {
  return {
    id,
    title: id === TARGET_B_ID ? "Backend Engineer B" : "Frontend Engineer A",
    companyName: id === TARGET_B_ID ? "Beta" : "Acme",
    analysisStatus: "ready",
    status: "draft",
    targetLanguage: "en",
    requirements: [],
    openQuestionIssueCount: 0,
    resumeId: "resume-ready",
    summary: {
      coreThemes: [],
      interviewRounds: [
        {
          sequence: 1,
          type: "technical",
          name: roundName,
          durationMinutes: 50,
          focus: "System boundaries",
        },
        {
          sequence: 2,
          type: "manager",
          name: "Manager",
          durationMinutes: 40,
          focus: "Influence",
        },
        {
          sequence: 3,
          type: "culture",
          name: "Culture",
          durationMinutes: 30,
          focus: "Values",
        },
        {
          sequence: 4,
          type: "final",
          name: "Final",
          durationMinutes: 45,
          focus: "Decision",
        },
      ],
      provenance: {
        modelId: "DO-NOT-RENDER-MODEL",
        promptVersion: "DO-NOT-RENDER-PROMPT",
        rubricVersion: "DO-NOT-RENDER-RUBRIC",
        dataSourceVersion: "DO-NOT-RENDER-SOURCE",
        featureFlag: "DO-NOT-RENDER-FLAG",
        language: "en",
      },
    },
    createdAt: "2026-07-14T08:00:00Z",
    updatedAt: "2026-07-14T08:00:00Z",
  };
}

function overview(
  targetJobId = TARGET_A_ID,
  prefix: keyof typeof REPORT_IDS = "a",
): TargetJobReportsOverview {
  const ids = REPORT_IDS[prefix];
  return {
    targetJobId,
    rounds: [
      {
        round: { roundId: "round-1-technical", roundSequence: 1 },
        currentReport: {
          id: ids.current,
          generatedAt: "2026-07-13T14:20:00Z",
        },
        latestAttempt: {
          id: ids.latestReady,
          status: "ready",
          errorCode: null,
          createdAt: "2026-07-14T09:10:00Z",
        },
      },
      {
        round: { roundId: "round-2-manager", roundSequence: 2 },
        currentReport: {
          id: ids.currentSecond,
          generatedAt: "2026-07-13T14:20:00Z",
        },
        latestAttempt: {
          id: ids.failed,
          status: "failed",
          errorCode: "AI_PROVIDER_TIMEOUT",
          createdAt: "2026-07-14T09:14:00Z",
        },
      },
      {
        round: { roundId: "round-3-culture", roundSequence: 3 },
        currentReport: null,
        latestAttempt: {
          id: ids.generating,
          status: "generating",
          errorCode: null,
          createdAt: "2026-07-14T09:15:00Z",
        },
      },
      {
        round: { roundId: "round-4-final", roundSequence: 4 },
        currentReport: {
          id: ids.same,
          generatedAt: "2026-07-14T09:20:00Z",
        },
        latestAttempt: {
          id: ids.same,
          status: "ready",
          errorCode: null,
          createdAt: "2026-07-14T09:18:00Z",
        },
      },
    ],
  };
}

function emptyOverview(targetJobId = TARGET_A_ID): TargetJobReportsOverview {
  return {
    targetJobId,
    rounds: [
      ["round-1-technical", 1],
      ["round-2-manager", 2],
      ["round-3-culture", 3],
      ["round-4-final", 4],
    ].map(([roundId, roundSequence]) => ({
      round: { roundId: String(roundId), roundSequence: Number(roundSequence) },
      currentReport: null,
      latestAttempt: null,
    })),
  };
}

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}

function clientWith({
  getTargetJob = async (id: string) => targetJob(id),
  listTargetJobReports = async (id: string) => overview(id),
}: {
  getTargetJob?: (targetJobId: string) => Promise<TargetJob>;
  listTargetJobReports?: (
    targetJobId: string,
  ) => Promise<TargetJobReportsOverview>;
} = {}): EasyInterviewClient & {
  getTargetJob: ReturnType<typeof vi.fn>;
  listTargetJobReports: ReturnType<typeof vi.fn>;
  listTargetJobs: ReturnType<typeof vi.fn>;
} {
  return {
    getTargetJob: vi.fn(getTargetJob),
    listTargetJobReports: vi.fn(listTargetJobReports),
    listTargetJobs: vi.fn(async () => {
      throw new Error("global target list is forbidden");
    }),
  } as unknown as EasyInterviewClient & {
    getTargetJob: ReturnType<typeof vi.fn>;
    listTargetJobReports: ReturnType<typeof vi.fn>;
    listTargetJobs: ReturnType<typeof vi.fn>;
  };
}

function runtimeValue(client: EasyInterviewClient): AppRuntimeValue {
  return {
    client,
    runtime: { status: "ready", config: {} as never },
    auth: { status: "authenticated", user: {} as never },
    refreshAuth: vi.fn(),
  };
}

function reportsRoute(targetJobId?: string): Route {
  return {
    name: "reports",
    params: targetJobId ? { targetJobId } : {},
  } as Route;
}

function view(
  client: EasyInterviewClient,
  targetJobId: string | undefined,
  navigate = vi.fn(),
  replaceRoute = vi.fn(),
): ReactNode {
  return (
    <DisplayPreferencesProvider>
      <AppRuntimeContext.Provider value={runtimeValue(client)}>
        <NavigationProvider value={{ navigate, replaceRoute }}>
          <ReportsScreen route={reportsRoute(targetJobId)} />
        </NavigationProvider>
      </AppRuntimeContext.Provider>
    </DisplayPreferencesProvider>
  );
}

function CommitProbe({
  token,
  onCommit,
}: {
  token: string;
  onCommit: (snapshot: {
    title: string | null;
    currentLinks: number;
    generatingLinks: number;
  }) => void;
}) {
  useLayoutEffect(() => {
    onCommit({
      title: document.querySelector("[data-testid='reports-target-title']")
        ?.textContent ?? null,
      currentLinks: document.querySelectorAll("[data-testid='reports-current']")
        .length,
      generatingLinks: document.querySelectorAll(
        "[data-testid='reports-generating']",
      ).length,
    });
  }, [onCommit, token]);
  return null;
}

afterEach(() => {
  localStorage.removeItem("ei-lang");
});

describe("ReportsScreen", () => {
  it("strictly joins the current target and renders only current/latest actions", async () => {
    const navigate = vi.fn();
    const client = clientWith();
    render(view(client, TARGET_A_ID, navigate));

    expect(screen.getByTestId("reports-loading")).toHaveAttribute(
      "role",
      "status",
    );
    expect(await screen.findByTestId("reports-target-title")).toHaveTextContent(
      "Frontend Engineer A",
    );
    expect(client.getTargetJob).toHaveBeenCalledWith(TARGET_A_ID);
    expect(client.listTargetJobReports).toHaveBeenCalledWith(TARGET_A_ID);
    expect(client.listTargetJobs).not.toHaveBeenCalled();

    const differentReady = screen.getByTestId("reports-round-1");
    expect(differentReady).toHaveTextContent("Architecture · 50m");
    expect(differentReady).toHaveTextContent("The latest generation is ready");
    expect(within(differentReady).getAllByRole("button")).toHaveLength(1);
    fireEvent.click(within(differentReady).getByRole("button"));
    expect(navigate).toHaveBeenLastCalledWith({
      name: "report",
      params: { reportId: REPORT_IDS.a.current },
    });

    const failed = screen.getByTestId("reports-round-2");
    expect(failed).toHaveTextContent("The latest generation failed");
    expect(failed).not.toHaveTextContent("AI_PROVIDER_TIMEOUT");
    expect(within(failed).queryByRole("button", { name: /retry/i })).toBeNull();

    const generating = screen.getByTestId("reports-round-3");
    expect(generating).toHaveTextContent("A new report is being prepared");
    fireEvent.click(
      within(within(generating).getByTestId("reports-generating")).getByRole(
        "button",
      ),
    );
    expect(navigate).toHaveBeenLastCalledWith({
      name: "generating",
      params: { reportId: REPORT_IDS.a.generating },
    });

    const sameReady = screen.getByTestId("reports-round-4");
    expect(within(sameReady).getAllByRole("button")).toHaveLength(1);
    expect(sameReady).not.toHaveTextContent("Latest generation completed");

    const page = screen.getByTestId("reports-screen");
    expect(page).not.toHaveTextContent(REPORT_IDS.a.latestReady);
    expect(page).not.toHaveTextContent("DO-NOT-RENDER-MODEL");
    expect(page).not.toHaveTextContent("DO-NOT-RENDER-RUBRIC");

    fireEvent.click(screen.getByTestId("reports-back-button"));
    expect(navigate).toHaveBeenLastCalledWith({
      name: "workspace",
      params: { targetJobId: TARGET_A_ID },
    });
  });

  it("renders an explicit empty state without inventing report history", async () => {
    const client = clientWith({
      listTargetJobReports: async () => emptyOverview(),
    });
    render(view(client, TARGET_A_ID));

    expect(await screen.findByTestId("reports-empty")).toHaveTextContent(
      "NO REPORTS YET",
    );
    expect(screen.queryByTestId("reports-current")).toBeNull();
    expect(screen.queryByTestId("reports-generating")).toBeNull();
  });

  it("uses the same actionable status for queued and generating attempts", async () => {
    const queued = overview();
    queued.rounds[2]!.latestAttempt!.status = "queued";
    const client = clientWith({
      listTargetJobReports: async () => queued,
    });
    render(view(client, TARGET_A_ID));

    const row = await screen.findByTestId("reports-round-3");
    expect(row).toHaveTextContent("A new report is being prepared");
    expect(within(row).getByTestId("reports-generating")).toBeInTheDocument();
  });

  it("keeps the valid current-plan Back path available while data is loading", async () => {
    const pendingJob = deferred<TargetJob>();
    const pendingReports = deferred<TargetJobReportsOverview>();
    const navigate = vi.fn();
    const client = clientWith({
      getTargetJob: async () => pendingJob.promise,
      listTargetJobReports: async () => pendingReports.promise,
    });
    render(view(client, TARGET_A_ID, navigate));

    expect(screen.getByTestId("reports-loading")).toBeInTheDocument();
    expect(screen.getByTestId("reports-target-title")).toHaveTextContent(
      "Current interview plan",
    );
    fireEvent.click(screen.getByTestId("reports-back-button"));
    expect(navigate).toHaveBeenCalledWith({
      name: "workspace",
      params: { targetJobId: TARGET_A_ID },
    });
  });

  it("fails closed on network and contract errors without leaking raw details", async () => {
    const failedClient = clientWith({
      listTargetJobReports: async () => {
        throw new Error("HTTP 503 SECRET-UPSTREAM-DETAIL");
      },
    });
    const rendered = render(view(failedClient, TARGET_A_ID));

    expect(await screen.findByTestId("reports-error")).toHaveTextContent(
      "This plan's reports could not be verified",
    );
    expect(screen.getByTestId("reports-screen")).not.toHaveTextContent(
      "SECRET-UPSTREAM-DETAIL",
    );
    expect(screen.queryByTestId("reports-current")).toBeNull();

    const mismatchClient = clientWith({
      listTargetJobReports: async () => overview(TARGET_OTHER_ID),
    });
    rendered.rerender(view(mismatchClient, TARGET_A_ID));
    expect(await screen.findByTestId("reports-error")).toBeInTheDocument();
    expect(screen.queryByTestId("reports-round-1")).toBeNull();
  });

  it("does not regress from error to loading when the target request settles late", async () => {
    const delayedTarget = deferred<TargetJob>();
    const client = clientWith({
      getTargetJob: async () => delayedTarget.promise,
      listTargetJobReports: async () => {
        throw new Error("report overview unavailable");
      },
    });
    render(view(client, TARGET_A_ID));

    expect(await screen.findByTestId("reports-error")).toBeInTheDocument();
    await act(async () => delayedTarget.resolve(targetJob()));

    expect(screen.getByTestId("reports-error")).toBeInTheDocument();
    expect(screen.queryByTestId("reports-loading")).toBeNull();
    expect(screen.getByTestId("reports-target-title")).toHaveTextContent(
      "Acme · Frontend Engineer A",
    );
  });

  it.each([
    [
      "target response mismatch",
      () => targetJob(TARGET_OTHER_ID),
      () => overview(TARGET_A_ID),
    ],
    [
      "missing canonical round",
      () => targetJob(TARGET_A_ID),
      () => ({ ...overview(), rounds: overview().rounds.slice(0, 3) }),
    ],
    [
      "extra overview round",
      () => targetJob(TARGET_A_ID),
      () => ({
        ...overview(),
        rounds: [
          ...overview().rounds,
          {
            round: { roundId: "round-5-other", roundSequence: 5 },
            currentReport: null,
            latestAttempt: null,
          },
        ],
      }),
    ],
    [
      "noncanonical target rounds",
      () => ({
        ...targetJob(TARGET_A_ID),
        summary: {
          ...targetJob(TARGET_A_ID).summary!,
          interviewRounds: targetJob(TARGET_A_ID).summary!.interviewRounds!.map(
            (round, index) =>
              index === 1 ? { ...round, sequence: 1 } : round,
          ),
        },
      }),
      () => overview(TARGET_A_ID),
    ],
  ])("fails the whole screen closed for %s", async (_label, makeJob, makeOverview) => {
    const client = clientWith({
      getTargetJob: async () => makeJob() as TargetJob,
      listTargetJobReports: async () =>
        makeOverview() as TargetJobReportsOverview,
    });
    render(view(client, TARGET_A_ID));

    expect(await screen.findByTestId("reports-error")).toBeInTheDocument();
    expect(screen.queryByTestId("reports-list")).toBeNull();
    expect(screen.queryByTestId("reports-current")).toBeNull();
    expect(screen.queryByTestId("reports-generating")).toBeNull();
  });

  it("isolates A/B targets and keeps the latest-started request authoritative", async () => {
    const aJob = deferred<TargetJob>();
    const aReports = deferred<TargetJobReportsOverview>();
    const bJob = deferred<TargetJob>();
    const bReports = deferred<TargetJobReportsOverview>();
    const client = clientWith({
      getTargetJob: (id) => (id === TARGET_A_ID ? aJob.promise : bJob.promise),
      listTargetJobReports: (id) =>
        id === TARGET_A_ID ? aReports.promise : bReports.promise,
    });
    const rendered = render(view(client, TARGET_A_ID));

    await waitFor(() =>
      expect(client.getTargetJob).toHaveBeenCalledWith(TARGET_A_ID),
    );
    rendered.rerender(view(client, TARGET_B_ID));
    await waitFor(() =>
      expect(client.listTargetJobReports).toHaveBeenCalledWith(TARGET_B_ID),
    );

    await act(async () => {
      bJob.resolve(targetJob(TARGET_B_ID, "Architecture B"));
      bReports.resolve(overview(TARGET_B_ID, "b"));
    });
    expect(await screen.findByTestId("reports-target-title")).toHaveTextContent(
      "Backend Engineer B",
    );

    await act(async () => {
      aJob.resolve(targetJob(TARGET_A_ID, "Architecture A"));
      aReports.resolve(overview(TARGET_A_ID, "a"));
    });
    expect(screen.getByTestId("reports-target-title")).toHaveTextContent(
      "Backend Engineer B",
    );
    expect(screen.getByTestId("reports-screen")).not.toHaveTextContent(
      "Frontend Engineer A",
    );
    expect(screen.getByTestId("reports-screen")).not.toHaveTextContent(
      REPORT_IDS.a.current,
    );
  });

  it("clears old target content in the target-switch commit", async () => {
    const bJob = deferred<TargetJob>();
    const bReports = deferred<TargetJobReportsOverview>();
    const client = clientWith({
      getTargetJob: (id) =>
        id === TARGET_A_ID ? Promise.resolve(targetJob()) : bJob.promise,
      listTargetJobReports: (id) =>
        id === TARGET_A_ID ? Promise.resolve(overview()) : bReports.promise,
    });
    const commits: Array<{
      title: string | null;
      currentLinks: number;
      generatingLinks: number;
    }> = [];
    const onCommit = (snapshot: (typeof commits)[number]) => commits.push(snapshot);
    const tree = (id: string, token: string) => (
      <>
        {view(client, id)}
        <CommitProbe token={token} onCommit={onCommit} />
      </>
    );
    const rendered = render(tree(TARGET_A_ID, "A"));

    await screen.findByTestId("reports-target-title");
    rendered.rerender(tree(TARGET_B_ID, "B"));

    expect(commits.at(-1)).toEqual({
      title: "Current interview plan",
      currentLinks: 0,
      generatingLinks: 0,
    });
    expect(screen.getByTestId("reports-loading")).toBeInTheDocument();
  });

  it.each([undefined, "", "not-a-uuid"])(
    "safely replaces missing/invalid target %s with workspace without requests",
    async (targetJobId) => {
      const navigate = vi.fn();
      const replaceRoute = vi.fn();
      const client = clientWith();
      render(view(client, targetJobId, navigate, replaceRoute));

      await waitFor(() =>
        expect(replaceRoute).toHaveBeenCalledWith({
          name: "workspace",
          params: {},
        }),
      );
      expect(navigate).not.toHaveBeenCalled();
      expect(client.getTargetJob).not.toHaveBeenCalled();
      expect(client.listTargetJobReports).not.toHaveBeenCalled();
      expect(client.listTargetJobs).not.toHaveBeenCalled();
      expect(screen.queryByTestId("reports-current")).toBeNull();
    },
  );
});
