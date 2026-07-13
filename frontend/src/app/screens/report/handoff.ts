import type { FeedbackReport } from "../../../api/generated/types";

export interface ReplayPayloadInput {
  report: FeedbackReport | null;
}

export function buildReplayPayload(input: ReplayPayloadInput): Record<string, string> {
  return buildDerivedPayload(input.report, "retry_current_round");
}

export function buildNextRoundPayload(input: ReplayPayloadInput): Record<string, string> {
  return buildDerivedPayload(input.report, "next_round");
}

function buildDerivedPayload(
  report: FeedbackReport | null,
  goal: "retry_current_round" | "next_round",
): Record<string, string> {
  if (!report?.id) return {};
  return { goal, sourceReportId: report.id };
}
