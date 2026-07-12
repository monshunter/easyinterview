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
): CreatePracticePlanRequest {
  const targetJobId = normalizeServerBoundId(ctx.targetJobId);
  if (!targetJobId) {
    throw new Error("invalid targetJobId");
  }

  const resumeId = normalizeServerBoundId(ctx.resumeId);
  if (!resumeId) {
    throw new Error("invalid resumeId");
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
    timeBudgetMinutes: 30,
    resumeId,
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
