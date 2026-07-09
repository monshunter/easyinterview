import type { TargetJob } from "../../api/generated/types";
import {
  buildTargetJobRoundAssumptions,
  roundIndexFromTargetJobStatus,
} from "../interview-context/roundAssumptions";

/**
 * Interview context derived from a TargetJob for parse/practice navigation.
 *
 * Derives route params from TargetJob plus declared server bindings. It must
 * not fabricate practice plan or resume IDs.
 */
export interface InterviewContext {
  targetJobId: string;
  jobId: string;
  jdId: string;
  planId: string;
  resumeId: string;
  roundId: string;
  roundName: string;
}

export interface InterviewContextOptions {
  resumeId?: string;
}

export const defaultPracticeRouteParams = {
  mode: "text",
  modality: "text",
  practiceMode: "strict",
  practiceGoal: "baseline",
  hintUsed: "false",
  hintCount: "0",
} as const;

export function interviewContextFromTargetJob(
  job: TargetJob,
  options: InterviewContextOptions = {},
): InterviewContext {
  const id = job.id;
  const planId = job.currentPracticePlanId?.trim() || "";
  const resumeId = options.resumeId?.trim() || job.resumeId?.trim() || "";
  const rounds = buildTargetJobRoundAssumptions(job);
  const roundIndex = roundIndexFromTargetJobStatus(job.status, rounds.length);
  const round = rounds[roundIndex] ?? rounds[0];
  return {
    targetJobId: id,
    jobId: id,
    jdId: `jd-${id}`,
    planId,
    resumeId,
    roundId: round?.id ?? "",
    roundName: round?.name ?? "",
  };
}

export function targetJobDetailRouteParams(
  job: TargetJob,
): Record<string, string> {
  const context = interviewContextFromTargetJob(job);
  return omitEmpty({
    targetJobId: context.targetJobId,
    planId: context.planId,
    resumeId: context.resumeId,
  });
}

export function targetJobPracticeRouteParams(
  job: TargetJob,
  options: InterviewContextOptions & { language?: string } = {},
): Record<string, string> {
  const context = interviewContextFromTargetJob(job, options);
  return omitEmpty({
    ...context,
    ...defaultPracticeRouteParams,
    language: options.language?.trim() || job.targetLanguage?.trim() || "",
  });
}

export function isTargetJobPracticeStartable(job: TargetJob): boolean {
  const context = interviewContextFromTargetJob(job);
  return (
    job.analysisStatus === "ready" &&
    context.resumeId.trim().length > 0 &&
    context.roundId.trim().length > 0
  );
}

function omitEmpty(input: Record<string, string>): Record<string, string> {
  const next: Record<string, string> = {};
  for (const [key, value] of Object.entries(input)) {
    const trimmed = value.trim();
    if (trimmed) next[key] = trimmed;
  }
  return next;
}
