import type { TargetJob } from "../../api/generated/types";

/**
 * Interview context derived from a TargetJob for workspace navigation.
 *
 * Plan §3.7: derives workspace params from TargetJob plus declared server
 * bindings. It must not fabricate practice plan or resume IDs.
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
  return {
    targetJobId: id,
    jobId: id,
    jdId: `jd-${id}`,
    planId,
    resumeId,
    roundId: "round-technical-1",
    roundName: "Technical Round 1",
  };
}
