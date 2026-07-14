/** @vitest-environment jsdom */
import { act, render, renderHook, screen, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it } from "vitest";
import { InterviewContextRouteSync } from "../App";
import { DEFAULT_INTERVIEW_CONTEXT, InterviewContextProvider, interviewContextReducer, useInterviewContext } from "./InterviewContext";

const wrapper = ({ children }: { children: ReactNode }) => <InterviewContextProvider>{children}</InterviewContextProvider>;

const ContextProbe = () => {
  const { ctx } = useInterviewContext();
  return (
    <output
      data-testid="interview-context-probe"
      data-target-job-id={ctx.targetJobId}
      data-plan-id={ctx.planId ?? ""}
      data-session-id={ctx.sessionId ?? ""}
    />
  );
};

describe("InterviewContext", () => {
  it("stores only conversation ownership and goal context", () => {
    expect(DEFAULT_INTERVIEW_CONTEXT).toEqual({
      planId: undefined, targetJobId: "", jobId: "", jdId: undefined,
      resumeId: undefined, sourceReportId: undefined, roundId: undefined,
      roundName: undefined, practiceGoal: "baseline", sessionId: undefined,
    });
    expect(DEFAULT_INTERVIEW_CONTEXT).not.toHaveProperty("practiceMode");
    expect(DEFAULT_INTERVIEW_CONTEXT).not.toHaveProperty("hintCount");
  });

  it("hydrates IDs and goal without structured question controls", () => {
    const next = interviewContextReducer(DEFAULT_INTERVIEW_CONTEXT, { type: "HYDRATE_FROM_ROUTE", params: { targetJobId: "tj-1", planId: "plan-1", sessionId: "session-1", practiceGoal: "retry_current_round", practiceMode: "assisted", hintCount: "2" } });
    expect(next).toMatchObject({ targetJobId: "tj-1", jobId: "tj-1", jdId: "jd-tj-1", planId: "plan-1", sessionId: "session-1", practiceGoal: "retry_current_round" });
    expect(next).not.toHaveProperty("practiceMode");
    expect(next).not.toHaveProperty("hintCount");
  });

  it("merges session and clears atomically", () => {
    const merged = interviewContextReducer(DEFAULT_INTERVIEW_CONTEXT, { type: "MERGE_SESSION", session: { id: "session-1" } });
    expect(merged.sessionId).toBe("session-1");
    expect(interviewContextReducer(merged, { type: "CLEAR" })).toEqual(DEFAULT_INTERVIEW_CONTEXT);
  });

  it("provides reducer updates to route consumers", () => {
    const { result } = renderHook(() => useInterviewContext(), { wrapper });
    act(() => result.current.dispatch({ type: "HYDRATE_FROM_ROUTE", params: { targetJobId: "tj-2" } }));
    expect(result.current.ctx).toMatchObject({ targetJobId: "tj-2", jdId: "jd-tj-2" });
  });

  it("resets stale practice context, hydrates Workspace detail, and clears the list route", async () => {
    const { rerender } = render(
      <InterviewContextProvider>
        <InterviewContextRouteSync
          route={{
            name: "practice",
            params: {
              targetJobId: "tj-old",
              planId: "plan-old",
              sessionId: "session-old",
            },
          }}
        />
        <ContextProbe />
      </InterviewContextProvider>,
    );
    await waitFor(() =>
      expect(screen.getByTestId("interview-context-probe")).toHaveAttribute(
        "data-plan-id",
        "plan-old",
      ),
    );

    rerender(
      <InterviewContextProvider>
        <InterviewContextRouteSync
          route={{ name: "workspace", params: { targetJobId: "tj-current" } }}
        />
        <ContextProbe />
      </InterviewContextProvider>,
    );
    await waitFor(() => {
      const probe = screen.getByTestId("interview-context-probe");
      expect(probe).toHaveAttribute("data-target-job-id", "tj-current");
      expect(probe).toHaveAttribute("data-plan-id", "");
      expect(probe).toHaveAttribute("data-session-id", "");
    });

    rerender(
      <InterviewContextProvider>
        <InterviewContextRouteSync route={{ name: "workspace", params: {} }} />
        <ContextProbe />
      </InterviewContextProvider>,
    );
    await waitFor(() =>
      expect(screen.getByTestId("interview-context-probe")).toHaveAttribute(
        "data-target-job-id",
        "",
      ),
    );
  });
});
