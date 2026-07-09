import type { EasyInterviewClient } from "../../api/generated/client";
import { newIdempotencyBatch } from "../../lib/conventions/idempotency";
import { buildCreatePlanRequest } from "./buildCreatePlanRequest";
import {
  DEFAULT_INTERVIEW_CONTEXT,
  type InterviewContextState,
} from "./InterviewContext";
import { normalizeServerBoundId } from "./apiIds";

export interface StartPracticeResult {
  sessionId: string;
  planId: string;
  params: Record<string, string>;
}

export async function startPracticeFromParams(
  client: EasyInterviewClient,
  params: Record<string, string>,
  lang: string,
): Promise<StartPracticeResult> {
  const ctx = interviewContextStateFromParams(params);
  const batch = newIdempotencyBatch();
  const shouldCreateDerivedPlan =
    ctx.practiceGoal === "retry_current_round" ||
    ctx.practiceGoal === "next_round";

  let planId = shouldCreateDerivedPlan
    ? undefined
    : normalizeServerBoundId(ctx.planId);

  if (planId) {
    try {
      const existingPlan = await client.getPracticePlan(planId);
      const matchesContext =
        existingPlan.targetJobId === ctx.targetJobId &&
        existingPlan.resumeId === ctx.resumeId;
      planId =
        existingPlan.status === "ready" && matchesContext
          ? existingPlan.id
          : undefined;
    } catch (err: unknown) {
      if (!isNotFound(err)) throw err;
      planId = undefined;
    }
  }

  if (!planId) {
    const plan = await client.createPracticePlan(
      buildCreatePlanRequest(ctx, lang),
      { idempotencyKey: batch.create },
    );
    planId = plan.id;
  }

  const session = await client.startPracticeSession(
    { planId, hintsEnabled: ctx.practiceMode === "assisted" },
    { idempotencyKey: batch.start },
  );

  return {
    sessionId: session.id,
    planId,
    params: omitEmpty({
      ...withoutAutoStart(params),
      targetJobId: ctx.targetJobId,
      jobId: ctx.jobId || ctx.targetJobId,
      jdId: ctx.jdId ?? "",
      resumeId: ctx.resumeId ?? "",
      roundId: ctx.roundId ?? "",
      roundName: ctx.roundName ?? "",
      mode: ctx.mode,
      modality: ctx.modality,
      practiceMode: ctx.practiceMode,
      practiceGoal: ctx.practiceGoal,
      hintUsed: ctx.hintUsed,
      hintCount: ctx.hintCount,
      planId,
      sessionId: session.id,
    }),
  };
}

export function interviewContextStateFromParams(
  params: Record<string, string>,
): InterviewContextState {
  const targetJobId = params.targetJobId || params.jobId || "";
  return {
    ...DEFAULT_INTERVIEW_CONTEXT,
    planId: params.planId || undefined,
    targetJobId,
    jobId: params.jobId || targetJobId,
    jdId: params.jdId || (targetJobId ? `jd-${targetJobId}` : undefined),
    resumeId: params.resumeId || undefined,
    sourceReportId: params.sourceReportId || undefined,
    roundId: params.roundId || undefined,
    roundName: params.roundName || undefined,
    mode: params.mode || DEFAULT_INTERVIEW_CONTEXT.mode,
    modality: params.modality || DEFAULT_INTERVIEW_CONTEXT.modality,
    practiceMode:
      params.practiceMode || DEFAULT_INTERVIEW_CONTEXT.practiceMode,
    practiceGoal:
      params.practiceGoal || DEFAULT_INTERVIEW_CONTEXT.practiceGoal,
    hintUsed: params.hintUsed || DEFAULT_INTERVIEW_CONTEXT.hintUsed,
    hintCount: params.hintCount || DEFAULT_INTERVIEW_CONTEXT.hintCount,
    sessionId: params.sessionId || undefined,
    autoStartPractice: undefined,
  };
}

function withoutAutoStart(
  params: Record<string, string>,
): Record<string, string> {
  const { autoStartPractice: _autoStartPractice, ...rest } = params;
  return rest;
}

function omitEmpty(input: Record<string, string>): Record<string, string> {
  const next: Record<string, string> = {};
  for (const [key, value] of Object.entries(input)) {
    if (value !== "") next[key] = value;
  }
  return next;
}

function isNotFound(err: unknown): boolean {
  const message = err instanceof Error ? err.message : String(err);
  return /^HTTP 404\b/.test(message);
}
