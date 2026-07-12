import type {
  CreatePracticePlanRequest,
  PracticeGoal,
} from "../../api/generated/types";
import type { InterviewContextState } from "./InterviewContext";
import { normalizeServerBoundId } from "./apiIds";

function isDerivedReportGoal(goal: string): goal is PracticeGoal {
  return goal === "retry_current_round" || goal === "next_round";
}

/**
 * Builds a CreatePracticePlanRequest from InterviewContext per plan §4.2 mapping.
 */
export function buildCreatePlanRequest(
  ctx: InterviewContextState,
  lang: string,
  timeBudgetMinutes: number,
): CreatePracticePlanRequest {
  const targetJobId = normalizeServerBoundId(ctx.targetJobId);
  if (!targetJobId) {
    throw new Error("invalid targetJobId");
  }

  const resumeId = normalizeServerBoundId(ctx.resumeId);
  if (!resumeId) {
    throw new Error("invalid resumeId");
  }
  if (!Number.isInteger(timeBudgetMinutes) || timeBudgetMinutes <= 0) {
    throw new Error("invalid timeBudgetMinutes");
  }
  const roundId = ctx.roundId?.trim();
  if (!roundId || !/^round-[1-9][0-9]*-(hr|technical|manager|cross_functional|culture|final|other)$/.test(roundId)) {
    throw new Error("invalid roundId");
  }

  const goal: PracticeGoal = isDerivedReportGoal(ctx.practiceGoal)
    ? ctx.practiceGoal
    : "baseline";
  const body: CreatePracticePlanRequest = {
    targetJobId,
    goal,
    interviewerPersona: "hiring_manager",
    difficulty: "standard",
    language: lang,
    timeBudgetMinutes,
    resumeId,
    roundId,
    focusCompetencyCodes: [],
  };

  if (isDerivedReportGoal(goal)) {
    const sourceReportId = normalizeServerBoundId(ctx.sourceReportId);
    if (!sourceReportId) {
      throw new Error("invalid sourceReportId");
    }
    body.sourceReportId = sourceReportId;
  }

  return body;
}
