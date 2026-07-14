import type {
  ApiErrorCode,
  ReportStatus,
  TargetJobCurrentReportSummary,
  TargetJobReportAttemptSummary,
} from "./generated/types";
import { ALL_REPORT_STATUSES } from "../lib/conventions";
import { isApiErrorCode } from "./runtimeApiErrorCode";

export interface CanonicalReportRoundDisplay {
  id: string;
  sequence: number;
}

export interface ValidatedTargetJobReportRound<
  TDisplay extends CanonicalReportRoundDisplay,
> {
  displayRound: TDisplay;
  currentReport: TargetJobCurrentReportSummary | null;
  latestAttempt: TargetJobReportAttemptSummary | null;
}

export interface ValidatedTargetJobReportsOverview<
  TDisplay extends CanonicalReportRoundDisplay,
> {
  targetJobId: string;
  rounds: ValidatedTargetJobReportRound<TDisplay>[];
}

const reportStatuses = new Set<string>(ALL_REPORT_STATUSES);
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
const RFC3339_DATE_TIME = /^(\d{4})-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])[Tt]([01]\d|2[0-3]):([0-5]\d):([0-5]\d)(?:\.\d+)?(?:[Zz]|[+-](?:[01]\d|2[0-3]):[0-5]\d)$/;

function invalid(): never {
  throw new Error("Invalid TargetJob reports overview");
}

function record(value: unknown): Record<string, unknown> {
  if (typeof value !== "object" || value === null || Array.isArray(value)) {
    return invalid();
  }
  return value as Record<string, unknown>;
}

function exactKeys(value: Record<string, unknown>, expected: readonly string[]): void {
  const actual = Object.keys(value).sort();
  const sortedExpected = [...expected].sort();
  if (
    actual.length !== sortedExpected.length ||
    actual.some((key, index) => key !== sortedExpected[index])
  ) {
    invalid();
  }
}

function nonBlank(value: unknown): string {
  if (typeof value !== "string" || value.trim() === "") return invalid();
  return value;
}

function uuid(value: unknown): string {
  const parsed = nonBlank(value);
  if (!UUID.test(parsed)) return invalid();
  return parsed;
}

function timestamp(value: unknown): string {
  const parsed = nonBlank(value);
  const match = RFC3339_DATE_TIME.exec(parsed);
  if (!match) return invalid();
  const year = Number(match[1]);
  const month = Number(match[2]);
  const day = Number(match[3]);
  const leapYear = year % 4 === 0 && (year % 100 !== 0 || year % 400 === 0);
  const daysInMonth = [31, leapYear ? 29 : 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31];
  if (year === 0 || day > daysInMonth[month - 1]! || !Number.isFinite(Date.parse(parsed))) {
    return invalid();
  }
  return parsed;
}

function currentReport(value: unknown): TargetJobCurrentReportSummary | null {
  if (value === null) return null;
  const parsed = record(value);
  exactKeys(parsed, ["generatedAt", "id"]);
  return {
    id: uuid(parsed.id),
    generatedAt: timestamp(parsed.generatedAt),
  };
}

function latestAttempt(value: unknown): TargetJobReportAttemptSummary | null {
  if (value === null) return null;
  const parsed = record(value);
  exactKeys(parsed, ["createdAt", "errorCode", "id", "status"]);
  const status = nonBlank(parsed.status);
  if (!reportStatuses.has(status)) return invalid();
  let errorCode: ApiErrorCode | null = null;
  if (status === "failed") {
    const candidate = nonBlank(parsed.errorCode);
    if (!isApiErrorCode(candidate)) return invalid();
    errorCode = candidate;
  } else if (parsed.errorCode !== null) {
    return invalid();
  }
  return {
    id: uuid(parsed.id),
    status: status as ReportStatus,
    errorCode,
    createdAt: timestamp(parsed.createdAt),
  };
}

/**
 * Fail-closed boundary for the generated listTargetJobReports response.
 * Display fields stay owned by the caller's current TargetJob round mapper;
 * the overview contributes only report locators and attempt state.
 */
export function validateTargetJobReportsOverview<
  TDisplay extends CanonicalReportRoundDisplay,
>(
  value: unknown,
  expectedTargetJobId: string,
  canonicalRounds: readonly TDisplay[],
): ValidatedTargetJobReportsOverview<TDisplay> {
  if (!UUID.test(expectedTargetJobId) || canonicalRounds.length < 2 || canonicalRounds.length > 5) {
    return invalid();
  }
  if (
    new Set(canonicalRounds.map((round) => round.id)).size !== canonicalRounds.length ||
    canonicalRounds.some((round, index) =>
      !Number.isSafeInteger(round.sequence) ||
      round.sequence <= 0 ||
      (index > 0 && round.sequence <= canonicalRounds[index - 1]!.sequence)
    )
  ) {
    return invalid();
  }

  const parsed = record(value);
  exactKeys(parsed, ["rounds", "targetJobId"]);
  if (uuid(parsed.targetJobId) !== expectedTargetJobId || !Array.isArray(parsed.rounds)) {
    return invalid();
  }
  if (parsed.rounds.length !== canonicalRounds.length) return invalid();

  const reportOwners = new Map<string, number>();
  const rounds = parsed.rounds.map((rawRound, index) => {
    const displayRound = canonicalRounds[index];
    if (!displayRound) return invalid();
    const row = record(rawRound);
    exactKeys(row, ["currentReport", "latestAttempt", "round"]);
    const ref = record(row.round);
    exactKeys(ref, ["roundId", "roundSequence"]);
    if (
      ref.roundId !== displayRound.id ||
      ref.roundSequence !== displayRound.sequence
    ) {
      return invalid();
    }
    const parsedCurrentReport = currentReport(row.currentReport);
    const parsedLatestAttempt = latestAttempt(row.latestAttempt);
    if (
      (parsedLatestAttempt?.status === "ready" && parsedCurrentReport === null) ||
      (parsedCurrentReport !== null && parsedLatestAttempt === null)
    ) {
      return invalid();
    }
    if (
      parsedCurrentReport !== null &&
      parsedLatestAttempt !== null &&
      parsedCurrentReport.id === parsedLatestAttempt.id &&
      parsedLatestAttempt.status !== "ready"
    ) {
      return invalid();
    }
    for (const reportId of new Set([
      parsedCurrentReport?.id,
      parsedLatestAttempt?.id,
    ].filter((id): id is string => id !== undefined))) {
      if (reportOwners.has(reportId)) return invalid();
      reportOwners.set(reportId, index);
    }
    return {
      displayRound,
      currentReport: parsedCurrentReport,
      latestAttempt: parsedLatestAttempt,
    };
  });

  return { targetJobId: expectedTargetJobId, rounds };
}
