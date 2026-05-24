import type { TargetJob } from "../../api/generated/types";

/**
 * Interview context derived from a TargetJob for workspace navigation.
 *
 * Plan §3.7: derives workspace params from TargetJob.id plus deterministic defaults.
 * Must not depend on OpenAPI-undeclared fields.
 */
export interface InterviewContext {
  targetJobId: string;
  jobId: string;
  jdId: string;
  planId: string;
  resumeVersionId: string;
  roundId: string;
  roundName: string;
}

export interface InterviewContextOptions {
  resumeVersionId?: string;
}

export function interviewContextFromTargetJob(
  job: TargetJob,
  options: InterviewContextOptions = {},
): InterviewContext {
  const id = job.id;
  const resumeVersionId = options.resumeVersionId?.trim() || "resume-unbound";
  return {
    targetJobId: id,
    jobId: id,
    jdId: `jd-${id}`,
    planId: `plan-${id}`,
    resumeVersionId,
    roundId: "round-technical-1",
    roundName: "Technical Round 1",
  };
}
