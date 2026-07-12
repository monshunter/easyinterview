import type { FeedbackReport } from "../../../api/generated/types";
import type { TargetJobRoundAssumption } from "../../interview-context/roundAssumptions";
import type { Route } from "../../routes";

export interface ReplayPayloadInput {
  route: Route;
  report: FeedbackReport | null;
  sessionId: string;
}

/**
 * Path A — `复练当前轮 / Replay current round`. Constructs the route params
 * forwarded to `nav("practice", ...)` so the practice screen restarts on the
 * same round and carries report-derived competency focus codes.
 *
 * The payload contains owner IDs + display knobs only — no raw text, no
 * promptHash, no AI model identifier. Tests guard the negative red lines.
 */
export function buildReplayPayload(
  input: ReplayPayloadInput,
): Record<string, string> {
  const { route, report, sessionId } = input;
  const params = route.params;
  const sourceReportId = report?.id ?? params.reportId ?? "";
  const focusCompetencyCodes = (report?.retryFocusCompetencyCodes ?? []).join(",");
  const evidenceGaps = (report?.issues ?? [])
    .map((issue) => issue.dimension)
    .filter((value): value is string => Boolean(value))
    .join("|");
  return omitEmpty({
    sourceSessionId: sessionId,
    focusCompetencyCodes,
    evidenceGaps,
    planId: params.planId ?? "",
    targetJobId: params.targetJobId ?? "",
    jdId: params.jdId ?? "",
    resumeId: params.resumeId ?? "",
    sourceReportId,
    roundId: params.roundId ?? "",
    practiceGoal: "retry_current_round",
  });
}

/**
 * Path B — `进入下一轮 / Start next round`. Same shape as path A but rotates
 * to the next round.
 */
export function buildNextRoundPayload(
  input: ReplayPayloadInput,
  nextRound: TargetJobRoundAssumption,
): Record<string, string> {
  const { route, report, sessionId } = input;
  const params = route.params;
  const sourceReportId = report?.id ?? params.reportId ?? "";
  return omitEmpty({
    sourceSessionId: sessionId,
    sourceReportId,
    nextRoundId: nextRound.id,
    roundName: nextRound.name,
    roundId: nextRound.id,
    planId: params.planId ?? "",
    targetJobId: params.targetJobId ?? "",
    jdId: params.jdId ?? "",
    resumeId: params.resumeId ?? "",
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
