import {
  createContext,
  useContext,
  useReducer,
  type Dispatch,
  type FC,
  type ReactNode,
} from "react";

/**
 * InterviewContext shared store — truth source for workspace / practice /
 * generating route context per plan §3.7 mapping.
 *
 * Fields that are route-param or fetched-data-derived; CLEAR resets to
 * DEFAULT_INTERVIEW_CONTEXT.
 */
export interface InterviewContextState {
  planId?: string;
  targetJobId: string;
  jobId: string;
  jdId?: string;
  resumeVersionId?: string;
  roundId?: string;
  roundName?: string;
  mode: string;
  modality: string;
  practiceMode: string;
  practiceGoal: string;
  hintUsed: string;
  hintCount: string;
  sessionId?: string;
  autoStartPractice?: string;
  /**
   * Frontend-debrief plan 001 Phase 5.4: ID of the debrief record returned by
   * `createDebrief` (202). Stays scoped to the debrief workflow — never written
   * back into `jobId`, which remains the target-job alias.
   */
  debriefId?: string;
  /**
   * Async job id returned alongside the debrief record. Used by the
   * `useDebriefPolling` Phase A loop to poll `getJob`.
   */
  debriefJobId?: string;
}

export const DEFAULT_INTERVIEW_CONTEXT: InterviewContextState = {
  planId: undefined,
  targetJobId: "",
  jobId: "",
  jdId: undefined,
  resumeVersionId: undefined,
  roundId: undefined,
  roundName: undefined,
  mode: "text",
  modality: "text",
  practiceMode: "strict",
  practiceGoal: "baseline",
  hintUsed: "false",
  hintCount: "0",
  sessionId: undefined,
  autoStartPractice: undefined,
  debriefId: undefined,
  debriefJobId: undefined,
};

export type InterviewContextAction =
  | {
      type: "HYDRATE_FROM_ROUTE";
      params: Record<string, string>;
    }
  | {
      type: "MERGE_TARGET_JOB";
      targetJob: { id: string; [key: string]: unknown };
    }
  | {
      type: "MERGE_RESUME";
      resume: { id: string; [key: string]: unknown };
    }
  | {
      type: "MERGE_PRACTICE_PLAN";
      plan: { id: string; [key: string]: unknown };
    }
  | {
      type: "MERGE_SESSION";
      session: { id: string; [key: string]: unknown };
    }
  | { type: "INCREMENT_HINT_COUNT" }
  | { type: "CLEAR_RESUME" }
  | { type: "CLEAR_PRACTICE_PLAN" }
  | { type: "CLEAR_AUTO_START" }
  | {
      /**
       * frontend-debrief plan 001 Phase 5.4. Writes `debriefId` /
       * `debriefJobId` (and optionally `practiceGoal`) without touching
       * `jobId`. Reducer test
       * `TestInterviewContext_DoesNotOverwriteJobId` guards the boundary.
       */
      type: "SET_DEBRIEF_CONTEXT";
      payload: {
        debriefId?: string;
        debriefJobId?: string;
        practiceGoal?: string;
      };
    }
  | { type: "CLEAR" };

export function interviewContextReducer(
  state: InterviewContextState,
  action: InterviewContextAction,
): InterviewContextState {
  switch (action.type) {
    case "HYDRATE_FROM_ROUTE": {
      const p = action.params;
      const targetJobId = p.targetJobId || p.jobId || state.targetJobId;
      return {
        ...state,
        planId:
          p.planId !== undefined ? (p.planId || undefined) : (targetJobId ? `plan-${targetJobId}` : state.planId),
        targetJobId,
        jobId: targetJobId,
        jdId:
          p.jdId !== undefined ? (p.jdId || undefined) : (targetJobId ? `jd-${targetJobId}` : state.jdId),
        resumeVersionId: p.resumeVersionId || state.resumeVersionId,
        roundId: p.roundId || state.roundId,
        roundName: p.roundName || state.roundName,
        mode: p.mode || state.mode,
        modality: p.modality || state.modality,
        practiceMode: p.practiceMode || state.practiceMode,
        practiceGoal: p.practiceGoal || state.practiceGoal,
        hintUsed: p.hintUsed || state.hintUsed,
        hintCount: p.hintCount || state.hintCount,
        sessionId: p.sessionId || state.sessionId,
        autoStartPractice: p.autoStartPractice ?? state.autoStartPractice,
        debriefId: p.debriefId || state.debriefId,
        debriefJobId: p.debriefJobId || state.debriefJobId,
      };
    }
    case "MERGE_TARGET_JOB":
      return {
        ...state,
        jobId: action.targetJob.id,
        targetJobId: action.targetJob.id,
      };
    case "MERGE_RESUME":
      return {
        ...state,
        resumeVersionId: action.resume.id || state.resumeVersionId,
      };
    case "MERGE_PRACTICE_PLAN":
      return {
        ...state,
        planId: action.plan.id || state.planId,
      };
    case "MERGE_SESSION":
      return {
        ...state,
        sessionId: action.session.id || state.sessionId,
      };
    case "INCREMENT_HINT_COUNT": {
      const parsed = Number(state.hintCount);
      const next = Number.isFinite(parsed) ? parsed + 1 : 1;
      return {
        ...state,
        hintUsed: "true",
        hintCount: String(next),
      };
    }
    case "CLEAR_RESUME":
      return {
        ...state,
        resumeVersionId: undefined,
      };
    case "CLEAR_PRACTICE_PLAN":
      return {
        ...state,
        planId: undefined,
      };
    case "CLEAR_AUTO_START":
      return {
        ...state,
        autoStartPractice: undefined,
      };
    case "SET_DEBRIEF_CONTEXT": {
      // Phase 5.4 guard: never overwrite `jobId` (target-job alias).
      return {
        ...state,
        debriefId: action.payload.debriefId ?? state.debriefId,
        debriefJobId: action.payload.debriefJobId ?? state.debriefJobId,
        practiceGoal: action.payload.practiceGoal ?? state.practiceGoal,
      };
    }
    case "CLEAR":
      return { ...DEFAULT_INTERVIEW_CONTEXT };
  }
}

interface InterviewContextValue {
  ctx: InterviewContextState;
  dispatch: Dispatch<InterviewContextAction>;
}

const InterviewCtx = createContext<InterviewContextValue | null>(null);

export const InterviewContextProvider: FC<{ children: ReactNode }> = ({
  children,
}) => {
  const [ctx, dispatch] = useReducer(
    interviewContextReducer,
    DEFAULT_INTERVIEW_CONTEXT,
  );
  return (
    <InterviewCtx.Provider value={{ ctx, dispatch }}>
      {children}
    </InterviewCtx.Provider>
  );
};

export function useInterviewContext(): InterviewContextValue {
  const value = useContext(InterviewCtx);
  if (!value) {
    throw new Error(
      "useInterviewContext must be used inside <InterviewContextProvider>",
    );
  }
  return value;
}

/**
 * Derives the context needed for starting a practice session — used by
 * workspace > useStartPractice() hook.
 */
export function useStartPracticeContext(): InterviewContextState {
  const { ctx } = useInterviewContext();
  return ctx;
}
