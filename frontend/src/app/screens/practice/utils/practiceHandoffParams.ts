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
  /** mode at the time of handoff (text / phone). */
  mode: string;
  /** modality at the time of handoff (text / phone). */
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
