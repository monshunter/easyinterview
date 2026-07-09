import type {
  TargetJob,
  TargetJobInterviewRound,
  TargetJobStatus,
} from "../../api/generated/types";
import type { MessageKey } from "../i18n/messages";

type Translate = (key: MessageKey) => string;

export interface TargetJobRoundAssumption {
  id: string;
  name: string;
  focus: string;
  type: TargetJobInterviewRound["type"];
  durationMinutes: number;
}

function nonBlank(value: string | undefined): string | null {
  const trimmed = value?.trim();
  return trimmed ? trimmed : null;
}

function displayRoundName(round: TargetJobInterviewRound): string {
  const name = nonBlank(round.name) ?? `R${round.sequence}`;
  return `${name} · ${round.durationMinutes}m`;
}

export function buildTargetJobRoundAssumptions(
  job: Pick<TargetJob, "summary"> | null | undefined,
  _t?: Translate,
): TargetJobRoundAssumption[] {
  const rounds = job?.summary?.interviewRounds ?? [];
  return rounds
    .filter((round) => round.sequence > 0 && round.durationMinutes > 0)
    .slice()
    .sort((a, b) => a.sequence - b.sequence)
    .map((round) => ({
      id: `round-${round.sequence}-${round.type}`,
      name: displayRoundName(round),
      focus: nonBlank(round.focus) ?? round.name,
      type: round.type,
      durationMinutes: round.durationMinutes,
    }));
}

export function roundIndexFromTargetJobStatus(
  status: TargetJobStatus,
  roundCount: number,
): number {
  if (roundCount === 0) return 0;
  switch (status) {
    case "draft":
    case "preparing":
      return 0;
    case "applied":
    case "interviewing":
      return Math.min(1, roundCount - 1);
    case "offer":
    case "rejected":
    case "archived":
      return roundCount - 1;
  }
}
