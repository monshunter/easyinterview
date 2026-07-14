import {
  createContext,
  useContext,
  useReducer,
  type Dispatch,
  type FC,
  type ReactNode,
} from "react";

/**
 * InterviewContext shared store — truth source for practice / generating /
 * report route context. Query-free Workspace clears stale context; a
 * target-scoped Workspace detail resets and hydrates its sole TargetJob
 * locator before any later practice handoff.
 *
 * Fields that are route-param or fetched-data-derived; CLEAR resets to
 * DEFAULT_INTERVIEW_CONTEXT.
 */
export interface InterviewContextState {
  planId?: string;
  targetJobId: string;
  jobId: string;
  jdId?: string;
  resumeId?: string;
  sourceReportId?: string;
  roundId?: string;
  roundName?: string;
  practiceGoal: string;
  sessionId?: string;
}

export const DEFAULT_INTERVIEW_CONTEXT: InterviewContextState = {
  planId: undefined,
  targetJobId: "",
  jobId: "",
  jdId: undefined,
  resumeId: undefined,
  sourceReportId: undefined,
  roundId: undefined,
  roundName: undefined,
  practiceGoal: "baseline",
  sessionId: undefined,
};

export type InterviewContextAction =
  | {
      type: "HYDRATE_FROM_ROUTE";
      params: Record<string, string>;
    }
  | {
      type: "MERGE_SESSION";
      session: { id: string; [key: string]: unknown };
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
          p.planId !== undefined ? (p.planId || undefined) : state.planId,
        targetJobId,
        jobId: targetJobId,
        jdId:
          p.jdId !== undefined ? (p.jdId || undefined) : (targetJobId ? `jd-${targetJobId}` : state.jdId),
        resumeId: p.resumeId || state.resumeId,
        sourceReportId: p.sourceReportId || state.sourceReportId,
        roundId: p.roundId || state.roundId,
        roundName: p.roundName || state.roundName,
        practiceGoal: p.practiceGoal || state.practiceGoal,
        sessionId: p.sessionId || state.sessionId,
      };
    }
    case "MERGE_SESSION":
      return {
        ...state,
        sessionId: action.session.id || state.sessionId,
      };
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
