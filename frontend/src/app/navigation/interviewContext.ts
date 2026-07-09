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
