import type { FeedbackReport } from "../../../api/generated/types";
import type { Route } from "../../routes";

const ROUND_ORDER = [
  "round-hr",
  "round-tech-1",
  "round-tech-2",
  "round-manager",
] as const;

const DEFAULT_NEXT_ROUND = "round-tech-2";

/**
 * Infers the next round identifier from an InterviewContext `roundId` using
 * the canonical interview ladder. Unknown / missing roundIds fall back to the
 * tech-2 default so the prototype source's behavior is preserved.
 *
 * Plan §3.7 Open Question: the round identifier scheme is provisional — when
 * backend-targetjob owner finalizes round metadata, swap this helper for the
 * real lookup. The contract stays the same: input `roundId | undefined`,
 * output a non-empty next round string.
 */
export function inferNextRoundId(currentRoundId: string | undefined): string {
  if (!currentRoundId) return DEFAULT_NEXT_ROUND;
  const idx = ROUND_ORDER.indexOf(currentRoundId as (typeof ROUND_ORDER)[number]);
  if (idx < 0) return DEFAULT_NEXT_ROUND;
  return ROUND_ORDER[Math.min(idx + 1, ROUND_ORDER.length - 1)]!;
}

export interface ReplayPayloadInput {
  route: Route;
  report: FeedbackReport | null;
  sessionId: string;
}

/**
 * Path A — `复练当前轮 / Replay current round`. Constructs the route params
 * forwarded to `nav("practice", ...)` so the practice screen restarts on the
 * same round but injects the report-derived retry-focus turn IDs.
 *
 * The payload contains owner IDs + display knobs only — no raw text, no
 * promptHash, no AI model identifier. Tests guard the negative red lines.
 */
export function buildReplayPayload(
  input: ReplayPayloadInput,
): Record<string, string> {
  const { route, report, sessionId } = input;
  const params = route.params;
  const replayItems = (report?.retryFocusTurnIds ?? []).join(",");
  const evidenceGaps = (report?.issues ?? [])
    .map((issue) => issue.dimension)
    .filter((value): value is string => Boolean(value))
    .join("|");
  return omitEmpty({
    sourceSessionId: sessionId,
    replayItems,
    evidenceGaps,
    planId: params.planId ?? "",
    targetJobId: params.targetJobId ?? "",
    jdId: params.jdId ?? "",
    resumeVersionId: params.resumeVersionId ?? "",
    roundId: params.roundId ?? "",
    mode: "text",
    modality: "text",
    practiceMode: params.practiceMode ?? "strict",
    practiceGoal: "retry_current_round",
  });
}

/**
 * Path B — `进入下一轮 / Start next round`. Same shape as path A but rotates
 * to the next round and drops the per-turn retry list.
 */
export function buildNextRoundPayload(
  input: ReplayPayloadInput,
): Record<string, string> {
  const { route, sessionId } = input;
  const params = route.params;
  const nextRoundId = inferNextRoundId(params.roundId);
  const roundName = params.roundName ?? "";
  return omitEmpty({
    sourceSessionId: sessionId,
    nextRoundId,
    roundName,
    roundId: nextRoundId,
    planId: params.planId ?? "",
    targetJobId: params.targetJobId ?? "",
    jdId: params.jdId ?? "",
    resumeVersionId: params.resumeVersionId ?? "",
    mode: "text",
    modality: "text",
    practiceMode: params.practiceMode ?? "strict",
    practiceGoal: "next_round",
  });
}

function omitEmpty(input: Record<string, string>): Record<string, string> {
  const next: Record<string, string> = {};
  for (const [key, value] of Object.entries(input)) {
    if (value !== "" && value !== undefined && value !== null) next[key] = value;
  }
  return next;
}
