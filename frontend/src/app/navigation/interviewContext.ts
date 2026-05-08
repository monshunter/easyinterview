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

export function interviewContextFromTargetJob(
  job: TargetJob,
): InterviewContext {
  const id = job.id;
  return {
    targetJobId: id,
    jobId: id,
    jdId: `jd-${id}`,
    planId: `plan-${id}`,
    resumeVersionId: "resume-unbound",
    roundId: "round-technical-1",
    roundName: "Technical Round 1",
  };
}
