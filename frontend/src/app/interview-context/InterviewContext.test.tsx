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
    expect(DEFAULT_INTERVIEW_CONTEXT.resumeId).toBeUndefined();
    expect(DEFAULT_INTERVIEW_CONTEXT.sourceReportId).toBeUndefined();
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
        resumeId: "rv-1",
        sourceReportId: "report-1",
        roundId: "round-hr",
        roundName: "HR 初筛",
        practiceMode: "assisted",
        practiceGoal: "retry_current_round",
      },
    };
    const next = interviewContextReducer(state, action);
    expect(next.targetJobId).toBe("tj-1");
    expect(next.jobId).toBe("tj-1");
    expect(next.jdId).toBe("jd-1");
    expect(next.planId).toBe("plan-1");
    expect(next.resumeId).toBe("rv-1");
    expect(next.sourceReportId).toBe("report-1");
    expect(next.roundId).toBe("round-hr");
    expect(next.roundName).toBe("HR 初筛");
    expect(next.practiceMode).toBe("assisted");
    expect(next.practiceGoal).toBe("retry_current_round");
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

  it("HYDRATE_FROM_ROUTE does not fabricate planId from targetJobId", () => {
    const state: InterviewContextState = { ...DEFAULT_INTERVIEW_CONTEXT };
    const action: InterviewContextAction = {
      type: "HYDRATE_FROM_ROUTE",
      params: { targetJobId: "tj-3" },
    };
    const next = interviewContextReducer(state, action);
    expect(next.planId).toBeUndefined();
  });

  it("MERGE_TARGET_JOB updates jobId and persisted target-job bindings", () => {
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
        resumeId: "rv-1",
        currentPracticePlanId: "plan-1",
      } as any,
    };
    const next = interviewContextReducer(state, action);
    expect(next.jobId).toBe("tj-1");
    expect(next.targetJobId).toBe("tj-1");
    expect(next.resumeId).toBe("rv-1");
    expect(next.planId).toBe("plan-1");
  });

  it("MERGE_RESUME sets resumeId from resume data", () => {
    const state: InterviewContextState = {
      ...DEFAULT_INTERVIEW_CONTEXT,
      targetJobId: "tj-1",
    };
    const action: InterviewContextAction = {
      type: "MERGE_RESUME",
      resume: { id: "rv-1", title: "FE Resume v3" } as any,
    };
    const next = interviewContextReducer(state, action);
    expect(next.resumeId).toBe("rv-1");
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

  it("INCREMENT_HINT_COUNT bumps hintCount and sets hintUsed='true'", () => {
    const state: InterviewContextState = {
      ...DEFAULT_INTERVIEW_CONTEXT,
      hintUsed: "false",
      hintCount: "0",
    };
    const action: InterviewContextAction = { type: "INCREMENT_HINT_COUNT" };
    const after1 = interviewContextReducer(state, action);
    expect(after1.hintCount).toBe("1");
    expect(after1.hintUsed).toBe("true");
    const after2 = interviewContextReducer(after1, action);
    expect(after2.hintCount).toBe("2");
    expect(after2.hintUsed).toBe("true");
  });

  it("INCREMENT_HINT_COUNT recovers to '1' when hintCount is non-numeric", () => {
    const state: InterviewContextState = {
      ...DEFAULT_INTERVIEW_CONTEXT,
      hintCount: "garbage",
    };
    const next = interviewContextReducer(state, { type: "INCREMENT_HINT_COUNT" });
    expect(next.hintCount).toBe("1");
    expect(next.hintUsed).toBe("true");
  });

  it("CLEAR resets to DEFAULT_INTERVIEW_CONTEXT", () => {
    const state: InterviewContextState = {
      targetJobId: "tj-1",
      jobId: "tj-1",
      planId: "plan-1",
      resumeId: "rv-1",
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
    expect(result.current.ctx.planId).toBeUndefined();
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

  it("ignores non-current debrief params while retaining current session context", () => {
    const next = interviewContextReducer(DEFAULT_INTERVIEW_CONTEXT, {
      type: "HYDRATE_FROM_ROUTE",
      params: {
        targetJobId: "tj-1",
        debriefId: "deb-rt",
        debriefJobId: "job-deb-rt",
        practiceGoal: "retry_current_round",
        sessionId: "sess-1",
      },
    });
    expect("debriefId" in next).toBe(false);
    expect("debriefJobId" in next).toBe(false);
    expect(next.practiceGoal).toBe("retry_current_round");
    expect(next.sessionId).toBe("sess-1");
  });

});

describe("INTERVIEW_CONTEXT_ROUTES parity with ui-design/src/app.jsx", () => {
  it("contains exact set from ui-design/src/app.jsx route mapping", () => {
    const expected = new Set([
      "practice",
      "generating",
      "report",
    ]);
    expect(INTERVIEW_CONTEXT_ROUTES).toEqual(expected);
  });

  it("shouldCarryInterviewContext returns true for context routes", () => {
    expect(shouldCarryInterviewContext("practice")).toBe(true);
    expect(shouldCarryInterviewContext("generating")).toBe(true);
    expect(shouldCarryInterviewContext("report")).toBe(true);
    expect(shouldCarryInterviewContext("standalone_insight")).toBe(false);
  });

  it("shouldCarryInterviewContext returns false for non-context routes", () => {
    expect(shouldCarryInterviewContext("home")).toBe(false);
    expect(shouldCarryInterviewContext("workspace")).toBe(false);
    expect(shouldCarryInterviewContext("settings")).toBe(false);
    expect(shouldCarryInterviewContext("debrief")).toBe(false);
    expect(shouldCarryInterviewContext("profile")).toBe(false);
    expect(shouldCarryInterviewContext("auth_login")).toBe(false);
  });
});
