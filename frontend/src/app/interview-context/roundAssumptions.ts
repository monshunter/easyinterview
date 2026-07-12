import type {
  PracticeRoundRef,
  TargetJob,
  TargetJobInterviewRound,
} from "../../api/generated/types";
import type { MessageKey } from "../i18n/messages";

type Translate = (key: MessageKey) => string;

const canonicalRoundTypes = new Set<string>([
  "hr",
  "technical",
  "manager",
  "cross_functional",
  "culture",
  "final",
  "other",
]);

export interface TargetJobRoundAssumption {
  id: string;
  sequence: number;
  name: string;
  focus: string;
  type: TargetJobInterviewRound["type"];
  durationMinutes: number;
}

export interface TargetJobRoundContext {
  currentRound: TargetJobRoundAssumption | null;
  nextRound: TargetJobRoundAssumption | null;
}

export interface TargetJobPracticeProgressContext {
  valid: boolean;
  rounds: TargetJobRoundAssumption[];
  completedCount: number;
  currentIndex: number | null;
  currentRound: TargetJobRoundAssumption | null;
  completed: boolean;
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
  const rounds = (job?.summary?.interviewRounds ?? [])
    .slice()
    .sort((a, b) => a.sequence - b.sequence);
  if (
    rounds.length < 2 ||
    rounds.length > 5 ||
    rounds.some((round, index) =>
      !Number.isInteger(round.sequence) ||
      round.sequence <= 0 ||
      (index > 0 && round.sequence <= rounds[index - 1]!.sequence) ||
      !canonicalRoundTypes.has(round.type) ||
      !nonBlank(round.name) ||
      !nonBlank(round.focus) ||
      !Number.isInteger(round.durationMinutes) ||
      round.durationMinutes < 10 ||
      round.durationMinutes > 180)
  ) {
    return [];
  }

  const normalized = rounds
    .map((round) => ({
      id: `round-${round.sequence}-${round.type}`,
      sequence: round.sequence,
      name: displayRoundName(round),
      focus: nonBlank(round.focus) ?? round.name,
      type: round.type,
      durationMinutes: round.durationMinutes,
    }));
  if (new Set(normalized.map((round) => round.id)).size !== normalized.length) {
    return [];
  }
  return normalized;
}

export function resolveTargetJobRoundContext(
  job: Pick<TargetJob, "summary"> | null | undefined,
  roundId: string | undefined,
): TargetJobRoundContext {
  const rounds = buildTargetJobRoundAssumptions(job);
  if (!roundId || new Set(rounds.map((round) => round.id)).size !== rounds.length) {
    return { currentRound: null, nextRound: null };
  }

  const currentIndex = rounds.findIndex((round) => round.id === roundId);
  if (currentIndex < 0) {
    return { currentRound: null, nextRound: null };
  }
  return {
    currentRound: rounds[currentIndex] ?? null,
    nextRound: rounds[currentIndex + 1] ?? null,
  };
}

function invalidProgress(
  rounds: TargetJobRoundAssumption[],
): TargetJobPracticeProgressContext {
  return {
    valid: false,
    rounds,
    completedCount: 0,
    currentIndex: null,
    currentRound: null,
    completed: false,
  };
}

function isExactRoundRef(
  ref: PracticeRoundRef | null | undefined,
  round: TargetJobRoundAssumption | undefined,
): boolean {
  return Boolean(
    ref &&
    round &&
    ref.roundId === round.id &&
    ref.roundSequence === round.sequence,
  );
}

/**
 * Validates the backend practice-progress read model against the canonical
 * structured rounds. Invalid or missing projections deliberately expose no
 * current round so callers cannot manufacture a frontend round cursor.
 */
export function resolveTargetJobPracticeProgress(
  job: Pick<TargetJob, "summary" | "practiceProgress"> | null | undefined,
): TargetJobPracticeProgressContext {
  const rounds = buildTargetJobRoundAssumptions(job);
  const progress = job?.practiceProgress;
  const ids = rounds.map((round) => round.id);
  const sequences = rounds.map((round) => round.sequence);
  if (
    rounds.length === 0 ||
    new Set(ids).size !== rounds.length ||
    new Set(sequences).size !== rounds.length ||
    !progress
  ) {
    return invalidProgress(rounds);
  }

  const completedCount = progress.completedRounds.length;
  if (
    completedCount > rounds.length ||
    progress.completedRounds.some((ref, index) =>
      !isExactRoundRef(ref, rounds[index]))
  ) {
    return invalidProgress(rounds);
  }

  if (completedCount === rounds.length) {
    if (progress.status !== "completed" || progress.currentRound !== null) {
      return invalidProgress(rounds);
    }
    return {
      valid: true,
      rounds,
      completedCount,
      currentIndex: rounds.length,
      currentRound: null,
      completed: true,
    };
  }

  const expectedStatus = completedCount === 0 ? "not_started" : "in_progress";
  const currentRound = rounds[completedCount];
  if (
    progress.status !== expectedStatus ||
    !isExactRoundRef(progress.currentRound, currentRound)
  ) {
    return invalidProgress(rounds);
  }

  return {
    valid: true,
    rounds,
    completedCount,
    currentIndex: completedCount,
    currentRound: currentRound ?? null,
    completed: false,
  };
}

/** Report next-round is valid only while its immediate successor is backend current. */
export function resolvePersistedNextRound(
  job: Pick<TargetJob, "summary" | "practiceProgress"> | null | undefined,
  completedRoundId: string | undefined,
): TargetJobRoundAssumption | null {
  const { nextRound } = resolveTargetJobRoundContext(job, completedRoundId);
  const progress = resolveTargetJobPracticeProgress(job);
  if (!nextRound || !progress.valid || progress.currentRound?.id !== nextRound.id) {
    return null;
  }
  return nextRound;
}
