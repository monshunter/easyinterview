/**
 * Item 4.2 — derive the stable handoff params for the `generating` route
 * from the InterviewContext + PracticeDisplayContext.
 *
 * Privacy red lines (spec §4 + plan §3.5): raw answer/question/hint/prompt/
 * provenance text must NEVER appear in URL params; only stable IDs and
 * display knobs are propagated.
 */

import type { InterviewContextState } from "../../../interview-context/InterviewContext";

export interface PracticeHandoffSource {
  ctx: InterviewContextState;
  reportId: string;
  /** mode at the time of handoff (text / voice). */
  mode: string;
  /** modality at the time of handoff (text / voice). */
  modality: string;
  /** practiceMode at the time of handoff (assisted / strict). */
  practiceMode: string;
  /** practiceGoal at the time of handoff. */
  practiceGoal: string;
  /** hintCount as integer (will be stringified for the URL). */
  hintCount: number;
}

export interface PracticeHandoffParams {
  // Stable owner IDs.
  planId: string;
  targetJobId: string;
  jdId: string;
  resumeId: string;
  roundId: string;
  sessionId: string;
  reportId: string;
  // PracticeDisplayContext (display-only, not in backend body).
  mode: string;
  modality: string;
  practiceMode: string;
  practiceGoal: string;
  hintUsed: string;
  hintCount: string;
}

const FORBIDDEN_KEYS = new Set([
  "answerText",
  "questionText",
  "hint",
  "prompt",
  "promptVersion",
  "rubricVersion",
  "modelId",
  "language",
  "featureFlag",
  "dataSourceVersion",
  "provenance",
]);

export function buildPracticeHandoffParams(
  source: PracticeHandoffSource,
): PracticeHandoffParams {
  const params: PracticeHandoffParams = {
    planId: source.ctx.planId ?? "",
    targetJobId: source.ctx.targetJobId ?? "",
    jdId: source.ctx.jdId ?? "",
    resumeId: source.ctx.resumeId ?? "",
    roundId: source.ctx.roundId ?? "",
    sessionId: source.ctx.sessionId ?? "",
    reportId: source.reportId,
    mode: source.mode,
    modality: source.modality,
    practiceMode: source.practiceMode,
    practiceGoal: source.practiceGoal,
    hintUsed: source.hintCount > 0 ? "true" : "false",
    hintCount: String(source.hintCount),
  };
  return params;
}

/**
 * Sanity gate: returns `null` if all keys are allowed; otherwise returns
 * an array of disallowed keys that must be stripped before navigating.
 */
export function findForbiddenHandoffKeys(
  params: Record<string, unknown>,
): string[] | null {
  const offenders: string[] = [];
  for (const key of Object.keys(params)) {
    if (FORBIDDEN_KEYS.has(key)) offenders.push(key);
  }
  return offenders.length > 0 ? offenders : null;
}
