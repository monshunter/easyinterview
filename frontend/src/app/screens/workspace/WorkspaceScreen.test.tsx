/**
 * @vitest-environment jsdom
 */

import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { useEffect, type ReactNode } from "react";

import type { TargetJob } from "../../../api/generated/types";
import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import {
  InterviewContextProvider,
  useInterviewContext,
} from "../../interview-context/InterviewContext";
import { NavigationProvider } from "../../navigation/NavigationProvider";
import type { Route } from "../../routes";
import { WorkspaceScreen } from "./WorkspaceScreen";

const workspaceTargetJobsMock = vi.hoisted(() => ({
  result: {
    loading: false,
    jobs: [] as TargetJob[],
    error: null as Error | null,
  },
}));

vi.mock("./hooks/useWorkspaceTargetJobs", () => ({
  useWorkspaceTargetJobs: () => workspaceTargetJobsMock.result,
}));

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

const READY_TARGET_JOB: TargetJob = {
  id: "01918fa0-0000-7000-8000-000000002000",
  title: "大模型应用开发工程师",
  companyName: "杭州某大型互联网电商平台上市公司",
  locationText: "Location not set",
  status: "draft",
  analysisStatus: "ready",
  sourceType: "manual_text",
  targetLanguage: "zh-CN",
  requirements: [],
  summary: {
    coreThemes: [],
    interviewRounds: [
      {
        sequence: 1,
        type: "technical",
        name: "技术电话面",
        durationMinutes: 30,
        focus: "技术电话面会确认基础工程能力",
      },
      {
        sequence: 2,
        type: "manager",
        name: "主管面",
        durationMinutes: 45,
        focus: "主管面会确认项目影响力",
      },
    ],
    provenance: {
      promptVersion: "v0.1.0",
      rubricVersion: "v0.1.0",
      modelId: "fixture-model:target-import-parse",
      language: "zh-CN",
      featureFlag: "none",
      dataSourceVersion: "registry.v1",
    },
  },
  openQuestionIssueCount: 0,
  createdAt: "2026-07-09T08:00:00Z",
  updatedAt: "2026-07-09T09:00:00Z",
  currentPracticePlanId: null,
  resumeId: "01918fa0-0000-7000-8000-000000001000",
};

beforeEach(() => {
  workspaceTargetJobsMock.result = {
    loading: false,
    jobs: [],
    error: null,
  };
});

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

  it("workspace remains the plan-list landing even when stale detail params are present", async () => {
    withProviders(<WorkspaceScreen route={WORKSPACE_ROUTE} />, WORKSPACE_ROUTE);

    await waitFor(() => {
      expect(screen.getByTestId("workspace-plan-list")).toBeDefined();
    });
    expect(screen.queryByTestId("parse-loading-step-0")).toBeNull();
    expect(screen.queryByTestId("workspace-header")).toBeNull();
    expect(screen.queryByTestId("workspace-launcher")).toBeNull();
    expect(screen.queryByTestId("workspace-jd-card")).toBeNull();
    expect(screen.queryByTestId("workspace-prep-card")).toBeNull();
    expect(screen.queryByTestId("workspace-history-card")).toBeNull();
  });

  it("keeps out-of-scope prototype testids out of the workspace route", () => {
    withProviders(<WorkspaceScreen route={WORKSPACE_ROUTE} />, WORKSPACE_ROUTE);

    expect(screen.queryByTestId("practice-mode-card-warmup")).toBeNull();
    expect(screen.queryByTestId("practice-mode-card-single_drill")).toBeNull();
    expect(screen.queryByTestId("growth-center")).toBeNull();
    expect(screen.queryByTestId("drill-builder-daily")).toBeNull();
    expect(screen.queryByTestId("mistake-queue-entry")).toBeNull();
  });

  it("keeps desktop plan cards on a stable fixed-width grid when only one card is present", async () => {
    workspaceTargetJobsMock.result = {
      loading: false,
      jobs: [READY_TARGET_JOB],
      error: null,
    };
    const route = { name: "workspace", params: {} } as const;
    withProviders(<WorkspaceScreen route={route} />, route);

    await waitFor(() => {
      expect(
        screen.getByTestId(
          "workspace-plan-list-card-01918fa0-0000-7000-8000-000000002000",
        ),
      ).toBeDefined();
    });
    const grid = screen.getByTestId("workspace-plan-list-grid");
    const rail = screen.getByTestId(
      "workspace-plan-list-rail-01918fa0-0000-7000-8000-000000002000",
    );

    expect(grid.style.gridTemplateColumns).toContain("360px");
    expect(grid.style.gridTemplateColumns).not.toContain("1fr");
    expect(grid.style.justifyContent).toBe("start");
    const card = screen.getByTestId(
      "workspace-plan-list-card-01918fa0-0000-7000-8000-000000002000",
    );
    expect((card as HTMLElement).style.position).toBe("relative");
    expect(rail).toHaveTextContent("技术电话面 · 30m");
    expect(rail).toHaveTextContent("主管面 · 45m");
    const footer = screen.getByTestId(
      "workspace-plan-list-card-footer-01918fa0-0000-7000-8000-000000002000",
    );
    expect(footer).toHaveTextContent("Start interview now");
    expect(footer).not.toHaveTextContent("Open plan");
    expect(
      footer.querySelector(
        "[data-testid='workspace-plan-list-delete-01918fa0-0000-7000-8000-000000002000']",
      ),
    ).toBeNull();
    expect(
      screen.queryByTestId("workspace-plan-list-open-01918fa0-0000-7000-8000-000000002000"),
    ).toBeNull();
    expect(
      screen.getByTestId("workspace-plan-list-start-01918fa0-0000-7000-8000-000000002000"),
    ).toBeDefined();
    const deleteButton = screen.getByTestId(
      "workspace-plan-list-delete-01918fa0-0000-7000-8000-000000002000",
    );
    expect((deleteButton as HTMLElement).style.position).toBe("absolute");
    expect((deleteButton as HTMLElement).style.right).toBe("14px");
    expect(deleteButton.querySelector('[data-icon="trash"]')).not.toBeNull();
  });
});
