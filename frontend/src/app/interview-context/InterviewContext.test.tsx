/**
 * @vitest-environment jsdom
 */

import { describe, expect, it } from "vitest";
import { renderHook, act } from "@testing-library/react";
import type { ReactNode } from "react";

import {
  INTERVIEW_CONTEXT_ROUTES,
  shouldCarryInterviewContext,
} from "../routes";
import {
  InterviewContextProvider,
  useInterviewContext,
  interviewContextReducer,
  DEFAULT_INTERVIEW_CONTEXT,
  type InterviewContextAction,
  type InterviewContextState,
} from "./InterviewContext";

function wrapper({ children }: { children: ReactNode }) {
  return (
    <InterviewContextProvider>{children}</InterviewContextProvider>
  );
}

describe("InterviewContext reducer", () => {
  it("DEFAULT_INTERVIEW_CONTEXT has correct defaults per plan §3.7", () => {
    expect(DEFAULT_INTERVIEW_CONTEXT.targetJobId).toBe("");
    expect(DEFAULT_INTERVIEW_CONTEXT.jobId).toBe("");
    expect(DEFAULT_INTERVIEW_CONTEXT.planId).toBeUndefined();
    expect(DEFAULT_INTERVIEW_CONTEXT.jdId).toBeUndefined();
    expect(DEFAULT_INTERVIEW_CONTEXT.resumeVersionId).toBeUndefined();
    expect(DEFAULT_INTERVIEW_CONTEXT.roundId).toBeUndefined();
    expect(DEFAULT_INTERVIEW_CONTEXT.roundName).toBeUndefined();
    expect(DEFAULT_INTERVIEW_CONTEXT.mode).toBe("text");
    expect(DEFAULT_INTERVIEW_CONTEXT.modality).toBe("text");
    expect(DEFAULT_INTERVIEW_CONTEXT.practiceMode).toBe("strict");
    expect(DEFAULT_INTERVIEW_CONTEXT.practiceGoal).toBe("baseline");
    expect(DEFAULT_INTERVIEW_CONTEXT.hintUsed).toBe("false");
    expect(DEFAULT_INTERVIEW_CONTEXT.hintCount).toBe("0");
    expect(DEFAULT_INTERVIEW_CONTEXT.sessionId).toBeUndefined();
    expect(DEFAULT_INTERVIEW_CONTEXT.autoStartPractice).toBeUndefined();
  });

  it("HYDRATE_FROM_ROUTE populates all fields from route params", () => {
    const state: InterviewContextState = { ...DEFAULT_INTERVIEW_CONTEXT };
    const action: InterviewContextAction = {
      type: "HYDRATE_FROM_ROUTE",
      params: {
        targetJobId: "tj-1",
        jdId: "jd-1",
        planId: "plan-1",
        resumeVersionId: "rv-1",
        roundId: "round-hr",
        roundName: "HR 初筛",
        practiceMode: "assisted",
        practiceGoal: "debrief",
      },
    };
    const next = interviewContextReducer(state, action);
    expect(next.targetJobId).toBe("tj-1");
    expect(next.jobId).toBe("tj-1");
    expect(next.jdId).toBe("jd-1");
    expect(next.planId).toBe("plan-1");
    expect(next.resumeVersionId).toBe("rv-1");
    expect(next.roundId).toBe("round-hr");
    expect(next.roundName).toBe("HR 初筛");
    expect(next.practiceMode).toBe("assisted");
    expect(next.practiceGoal).toBe("debrief");
    // unchanged defaults
    expect(next.mode).toBe("text");
    expect(next.modality).toBe("text");
  });

  it("HYDRATE_FROM_ROUTE derives jdId fallback from targetJobId", () => {
    const state: InterviewContextState = { ...DEFAULT_INTERVIEW_CONTEXT };
    const action: InterviewContextAction = {
      type: "HYDRATE_FROM_ROUTE",
      params: { targetJobId: "tj-2" },
    };
    const next = interviewContextReducer(state, action);
    expect(next.jdId).toBe("jd-tj-2");
  });

  it("HYDRATE_FROM_ROUTE derives planId fallback from targetJobId", () => {
    const state: InterviewContextState = { ...DEFAULT_INTERVIEW_CONTEXT };
    const action: InterviewContextAction = {
      type: "HYDRATE_FROM_ROUTE",
      params: { targetJobId: "tj-3" },
    };
    const next = interviewContextReducer(state, action);
    expect(next.planId).toBe("plan-tj-3");
  });

  it("MERGE_TARGET_JOB updates jobId from targetJob.id and merges round info", () => {
    const state: InterviewContextState = {
      ...DEFAULT_INTERVIEW_CONTEXT,
      targetJobId: "tj-1",
    };
    const action: InterviewContextAction = {
      type: "MERGE_TARGET_JOB",
      targetJob: {
        id: "tj-1",
        title: "Senior Frontend Engineer",
        companyName: "Acme Corp",
        locationText: "Shanghai",
        sourceType: "linkedin",
      } as any,
    };
    const next = interviewContextReducer(state, action);
    expect(next.jobId).toBe("tj-1");
    expect(next.targetJobId).toBe("tj-1");
  });

  it("MERGE_RESUME sets resumeVersionId from resume data", () => {
    const state: InterviewContextState = {
      ...DEFAULT_INTERVIEW_CONTEXT,
      targetJobId: "tj-1",
    };
    const action: InterviewContextAction = {
      type: "MERGE_RESUME",
      resume: { id: "rv-1", title: "FE Resume v3" } as any,
    };
    const next = interviewContextReducer(state, action);
    expect(next.resumeVersionId).toBe("rv-1");
  });

  it("MERGE_PRACTICE_PLAN sets planId from plan data", () => {
    const state: InterviewContextState = {
      ...DEFAULT_INTERVIEW_CONTEXT,
      targetJobId: "tj-1",
    };
    const action: InterviewContextAction = {
      type: "MERGE_PRACTICE_PLAN",
      plan: { id: "plan-1" } as any,
    };
    const next = interviewContextReducer(state, action);
    expect(next.planId).toBe("plan-1");
  });

  it("MERGE_SESSION sets sessionId from session data", () => {
    const state: InterviewContextState = {
      ...DEFAULT_INTERVIEW_CONTEXT,
      targetJobId: "tj-1",
    };
    const action: InterviewContextAction = {
      type: "MERGE_SESSION",
      session: { id: "sess-1" } as any,
    };
    const next = interviewContextReducer(state, action);
    expect(next.sessionId).toBe("sess-1");
  });

  it("CLEAR resets to DEFAULT_INTERVIEW_CONTEXT", () => {
    const state: InterviewContextState = {
      targetJobId: "tj-1",
      jobId: "tj-1",
      planId: "plan-1",
      resumeVersionId: "rv-1",
      sessionId: "sess-1",
      jdId: "jd-1",
      roundId: "round-hr",
      roundName: "HR",
      mode: "text",
      modality: "text",
      practiceMode: "assisted",
      practiceGoal: "baseline",
      hintUsed: "true",
      hintCount: "2",
    };
    const action: InterviewContextAction = { type: "CLEAR" };
    const next = interviewContextReducer(state, action);
    expect(next).toEqual(DEFAULT_INTERVIEW_CONTEXT);
  });
});

describe("InterviewContextProvider + useInterviewContext", () => {
  it("provides default context", () => {
    const { result } = renderHook(() => useInterviewContext(), { wrapper });
    expect(result.current.ctx.targetJobId).toBe("");
    expect(result.current.ctx.mode).toBe("text");
  });

  it("dispatch HYDRATE_FROM_ROUTE updates context", () => {
    const { result } = renderHook(() => useInterviewContext(), { wrapper });
    act(() => {
      result.current.dispatch({
        type: "HYDRATE_FROM_ROUTE",
        params: { targetJobId: "tj-5", roundName: "经理面" },
      });
    });
    expect(result.current.ctx.targetJobId).toBe("tj-5");
    expect(result.current.ctx.roundName).toBe("经理面");
    expect(result.current.ctx.jdId).toBe("jd-tj-5");
    expect(result.current.ctx.planId).toBe("plan-tj-5");
  });

  it("dispatch CLEAR resets context", () => {
    const { result } = renderHook(() => useInterviewContext(), { wrapper });
    act(() => {
      result.current.dispatch({
        type: "HYDRATE_FROM_ROUTE",
        params: { targetJobId: "tj-6" },
      });
    });
    expect(result.current.ctx.targetJobId).toBe("tj-6");
    act(() => {
      result.current.dispatch({ type: "CLEAR" });
    });
    expect(result.current.ctx.targetJobId).toBe("");
  });
});

describe("INTERVIEW_CONTEXT_ROUTES parity with ui-design/src/app.jsx", () => {
  it("contains exact set from ui-design/src/app.jsx line 76", () => {
    const expected = new Set([
      "workspace",
      "practice",
      "generating",
      "report",
      "debrief",
      "company_intel",
    ]);
    expect(INTERVIEW_CONTEXT_ROUTES).toEqual(expected);
  });

  it("shouldCarryInterviewContext returns true for context routes", () => {
    expect(shouldCarryInterviewContext("workspace")).toBe(true);
    expect(shouldCarryInterviewContext("practice")).toBe(true);
    expect(shouldCarryInterviewContext("generating")).toBe(true);
    expect(shouldCarryInterviewContext("report")).toBe(true);
    expect(shouldCarryInterviewContext("debrief")).toBe(true);
    expect(shouldCarryInterviewContext("company_intel")).toBe(true);
  });

  it("shouldCarryInterviewContext returns false for non-context routes", () => {
    expect(shouldCarryInterviewContext("home")).toBe(false);
    expect(shouldCarryInterviewContext("settings")).toBe(false);
    expect(shouldCarryInterviewContext("profile")).toBe(false);
    expect(shouldCarryInterviewContext("auth_login")).toBe(false);
  });
});
