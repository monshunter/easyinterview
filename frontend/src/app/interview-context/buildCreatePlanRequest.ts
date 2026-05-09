import type { CreatePracticePlanRequest } from "../../api/generated/types";
import type { InterviewContextState } from "./InterviewContext";

/**
 * Builds a CreatePracticePlanRequest from InterviewContext per plan §4.2 mapping.
 */
export function buildCreatePlanRequest(
  ctx: InterviewContextState,
  lang: string,
): CreatePracticePlanRequest {
  return {
    targetJobId: ctx.targetJobId,
    goal: "baseline",
    mode: "assisted",
    interviewerPersona: "hiring_manager",
    difficulty: "standard",
    language: lang,
    questionBudget: 6,
    timeBudgetMinutes: 30,
    resumeAssetId: ctx.resumeVersionId ?? undefined,
    focusCompetencyCodes: [],
  };
}
