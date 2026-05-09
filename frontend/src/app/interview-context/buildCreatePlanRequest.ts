import type { CreatePracticePlanRequest } from "../../api/generated/types";
import type { InterviewContextState } from "./InterviewContext";
import { normalizeServerBoundId } from "./apiIds";

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

  const resumeAssetId = normalizeServerBoundId(ctx.resumeVersionId);

  return {
    targetJobId,
    goal: "baseline",
    mode: "assisted",
    interviewerPersona: "hiring_manager",
    difficulty: "standard",
    language: lang,
    questionBudget: 6,
    timeBudgetMinutes: 30,
    resumeAssetId,
    focusCompetencyCodes: [],
  };
}
