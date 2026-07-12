/** @vitest-environment jsdom */
import { act, renderHook } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it } from "vitest";
import { DEFAULT_INTERVIEW_CONTEXT, InterviewContextProvider, interviewContextReducer, useInterviewContext } from "./InterviewContext";

const wrapper = ({ children }: { children: ReactNode }) => <InterviewContextProvider>{children}</InterviewContextProvider>;

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
});
