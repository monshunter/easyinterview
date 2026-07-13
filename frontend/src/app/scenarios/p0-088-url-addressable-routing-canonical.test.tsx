// @vitest-environment jsdom
/**
 * E2E.P0.088 — Canonical path deep-link / reload / browser history.
 *
 * Truth source: docs/spec/frontend-shell/plans/004-url-addressable-routing/
 * bdd-plan.md §2 (E2E.P0.088) + bdd-checklist.md.
 *
 * Given the App is built with the Browser History router and the Plan 004
 * route store, this scenario asserts that:
 *   - Direct-open of canonical workspace / practice / generating / report /
 *     resume-versions deep links lands on the correct route +
 *     params.
 *   - InterviewContext hydrates from URL safe params.
 *   - TopBar active state + chrome-hidden behaviour match the route
 *     catalog.
 *   - Reload preserves resource context (verified by re-mount).
 *   - Back / forward updates state without double-push or lost params.
 *   - Route-specific handoff keys survive while workspace detail/start params
 *     are stripped from the list route.
 *   - Unknown / malformed query falls back without crashing.
 */
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { FC } from "react";

import { App } from "../App";
import { useNavigation } from "../navigation/NavigationProvider";

const SESSION_ID = "01918fa0-0000-7000-8000-000000005000";
const REPORT_ID = "01918fa0-0000-7000-8000-00000000a000";
// Non-UUID id intentionally: WorkspaceScreen falls through to
// `workspace-plan-list` landing (no network fetch) without a runtime
// client, keeping the scenario URL-only and contract-light.
const TARGET_JOB_ID = "tj-canonical";
const RESUME_VERSION_ID = "01918fa0-0000-7000-8000-000000001000";
const PLAN_ID = "01918fa0-0000-7000-8000-000000004000";

function resetWindow(): void {
  delete (window as { __EASYINTERVIEW_INITIAL_ROUTE__?: unknown })
    .__EASYINTERVIEW_INITIAL_ROUTE__;
  window.history.replaceState(null, "", "/");
}

beforeEach(resetWindow);
afterEach(resetWindow);

const NavBatch: FC = () => {
  const { navigate } = useNavigation();
  return (
    <>
      <button
        type="button"
        data-testid="go-workspace-hostile"
        onClick={() =>
          navigate({
            name: "workspace",
            params: {
              targetJobId: TARGET_JOB_ID,
              resumeId: RESUME_VERSION_ID,
              planId: PLAN_ID,
              autoStartPractice: "1",
            },
          })
        }
      >
        workspace
      </button>
      <button
        type="button"
        data-testid="go-practice-chat"
        onClick={() =>
          navigate({
            name: "practice",
            params: {
              sessionId: SESSION_ID,
              planId: PLAN_ID,
            },
          })
        }
      >
        practice chat
      </button>
      <button
        type="button"
        data-testid="go-report"
        onClick={() =>
          navigate({
            name: "report",
            params: {
              sessionId: SESSION_ID,
              reportId: REPORT_ID,
              reportStatus: "failed",
              errorCode: "AI_PROVIDER_TIMEOUT",
            },
          })
        }
      >
        report
      </button>
    </>
  );
};

describe("E2E.P0.088 canonical path deep-link / reload / browser history", () => {
  it("direct-open /workspace?targetJobId=...&autoStartPractice=1 strips out-of-scope params and keeps TopBar active", () => {
    window.history.replaceState(
      null,
      "",
      `/workspace?targetJobId=${TARGET_JOB_ID}&resumeId=${RESUME_VERSION_ID}&planId=${PLAN_ID}&autoStartPractice=1`,
    );
    render(<App />);
    expect(screen.getByTestId("workspace-plan-list")).toBeInTheDocument();
    expect(window.location.pathname).toBe("/workspace");
    expect(window.location.search).toBe("");
    expect(screen.getByTestId("topbar-nav-workspace")).toHaveAttribute(
      "aria-current",
      "page",
    );
  });

  it("direct-open legacy phone params mounts chat with voice disabled", () => {
    window.history.replaceState(
      null,
      "",
      `/practice?sessionId=${SESSION_ID}&mode=phone&modality=phone&planId=${PLAN_ID}`,
    );
    render(<App />);
    expect(screen.getByTestId("practice-conversation")).toBeInTheDocument();
    expect(screen.getByTestId("practice-topbar-phone-toggle")).toBeDisabled();
    expect(screen.queryByTestId("app-shell-topbar")).not.toBeInTheDocument();
    expect(window.location.pathname).toBe("/practice");
  });

  it("direct-open /generating?reportId=... mounts GeneratingScreen with chrome hidden", () => {
    window.history.replaceState(
      null,
      "",
      `/generating?sessionId=${SESSION_ID}&reportId=${REPORT_ID}`,
    );
    render(<App />);
    expect(screen.getByTestId("generating-screen")).toBeInTheDocument();
    expect(screen.queryByTestId("app-shell-topbar")).not.toBeInTheDocument();
  });

  it("direct-open /report keeps only reportId and lets the API own state", () => {
    window.history.replaceState(
      null,
      "",
      `/report?sessionId=${SESSION_ID}&reportId=${REPORT_ID}&reportStatus=failed&errorCode=AI_PROVIDER_TIMEOUT`,
    );
    render(<App />);
    expect(screen.getByTestId("report-dashboard-loading")).toBeInTheDocument();
    expect(screen.getByTestId("app-shell-topbar")).toBeInTheDocument();
    expect(window.location.search).toBe(`?reportId=${REPORT_ID}`);
  });

  it("direct-open /resume-versions?tab=rewrites&tailorRunId=... filters out-of-scope resume detail tab keys", () => {
    window.history.replaceState(
      null,
      "",
      "/resume-versions?tab=rewrites&tailorRunId=01918fa0-0000-7000-8000-00000000b000",
    );
    render(<App />);
    expect(screen.getByTestId("resume-workshop-screen")).toBeInTheDocument();
    expect(window.location.pathname).toBe("/resume-versions");
    expect(window.location.search).toBe("");
  });

  it("direct-open out-of-scope /debrief and /profile paths fold back to home without out-of-scope params", () => {
    for (const path of [
      `/debrief?targetJobId=${TARGET_JOB_ID}&debriefId=01918fa0-0000-7000-8000-00000000c000`,
      "/profile",
    ]) {
      resetWindow();
      window.history.replaceState(null, "", path);
      const { unmount } = render(<App />);
      expect(screen.getByTestId("route-home")).toBeInTheDocument();
      expect(window.location.pathname).toBe("/");
      expect(window.location.search).toBe("");
      expect(screen.queryByTestId("debrief-screen")).not.toBeInTheDocument();
      expect(screen.queryByTestId("route-profile")).not.toBeInTheDocument();
      unmount();
    }
  });

  it("App navigation pushes 3 history entries and back/forward restores chrome state", async () => {
    render(
      <App>
        <NavBatch />
      </App>,
    );
    const user = userEvent.setup();
    await user.click(screen.getByTestId("go-workspace-hostile"));
    await waitFor(() => screen.getByTestId("workspace-plan-list"));
    await user.click(screen.getByTestId("go-practice-chat"));
    await waitFor(() => screen.getByTestId("practice-conversation"));
    expect(screen.queryByTestId("app-shell-topbar")).not.toBeInTheDocument();
    expect(screen.getByTestId("practice-topbar-phone-toggle")).toBeDisabled();
    await user.click(screen.getByTestId("go-report"));
    await waitFor(() => screen.getByTestId("report-dashboard-loading"));
    expect(screen.getByTestId("app-shell-topbar")).toBeInTheDocument();

    // BACK twice: report → practice → workspace
    act(() => {
      window.history.back();
      window.dispatchEvent(new PopStateEvent("popstate"));
    });
    await waitFor(() => screen.getByTestId("practice-conversation"));
    expect(screen.queryByTestId("app-shell-topbar")).not.toBeInTheDocument();

    act(() => {
      window.history.back();
      window.dispatchEvent(new PopStateEvent("popstate"));
    });
    await waitFor(() => screen.getByTestId("workspace-plan-list"));
    expect(screen.getByTestId("app-shell-topbar")).toBeInTheDocument();

    // FORWARD twice: workspace → practice → report
    act(() => {
      window.history.forward();
      window.dispatchEvent(new PopStateEvent("popstate"));
    });
    await waitFor(() => screen.getByTestId("practice-conversation"));
    act(() => {
      window.history.forward();
      window.dispatchEvent(new PopStateEvent("popstate"));
    });
    await waitFor(() => screen.getByTestId("report-dashboard-loading"));
  });

  it("malformed query (unknown keys) falls back without crashing", () => {
    window.history.replaceState(
      null,
      "",
      "/workspace?bogusKey=42&targetJobId=tj-ok&another=zz",
    );
    render(<App />);
    expect(screen.getByTestId("workspace-plan-list")).toBeInTheDocument();
    expect(window.location.search).toBe("");
  });

  it("hash `#route=workspace&targetJobId=...` boot rewrites URL to canonical /workspace", () => {
    window.history.replaceState(
      null,
      "",
      `/#route=workspace&targetJobId=${TARGET_JOB_ID}`,
    );
    render(<App />);
    expect(window.location.pathname).toBe("/workspace");
    expect(window.location.hash).toBe("");
    expect(window.location.search).toBe("");
    expect(screen.getByTestId("workspace-plan-list")).toBeInTheDocument();
  });
});
