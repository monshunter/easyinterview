import { describe, expect, it } from "vitest";

import { buildCreatePlanRequest } from "./buildCreatePlanRequest";
import { DEFAULT_INTERVIEW_CONTEXT, type InterviewContextState } from "./InterviewContext";

const VALID_TARGET_JOB_ID = "01918fa0-0000-7000-8000-000000002000";
const VALID_RESUME_ID = "01918fa0-0000-7000-8000-000000001000";
const VALID_REPORT_ID = "01918fa0-0000-7000-8000-000000003000";

function context(overrides: Partial<InterviewContextState>): InterviewContextState {
  return {
    ...DEFAULT_INTERVIEW_CONTEXT,
    targetJobId: VALID_TARGET_JOB_ID,
    jobId: VALID_TARGET_JOB_ID,
    ...overrides,
  };
}

describe("buildCreatePlanRequest", () => {
  it("keeps valid server-bound ids in the generated client request body", () => {
    const body = buildCreatePlanRequest(
      context({ resumeId: VALID_RESUME_ID }),
      "en",
      60,
    );

    expect(body.targetJobId).toBe(VALID_TARGET_JOB_ID);
    expect(body.resumeId).toBe(VALID_RESUME_ID);
    expect(body.goal).toBe("baseline");
    expect(body.sourceReportId).toBeUndefined();
    expect(body.timeBudgetMinutes).toBe(60);
  });

  it("creates next_round plans from the source report id", () => {
    const body = buildCreatePlanRequest(
      context({
        resumeId: VALID_RESUME_ID,
        practiceGoal: "next_round",
        sourceReportId: VALID_REPORT_ID,
      }),
      "en",
      60,
    );

    expect(body.goal).toBe("next_round");
    expect(body.sourceReportId).toBe(VALID_REPORT_ID);
  });

  it("rejects derived report plans without a valid sourceReportId", () => {
    expect(() =>
      buildCreatePlanRequest(
        context({
          resumeId: VALID_RESUME_ID,
          practiceGoal: "next_round",
        }),
        "en",
        60,
      ),
    ).toThrow("invalid sourceReportId");
  });

  it("rejects synthetic resume identifiers instead of sending incomplete API bodies", () => {
    expect(() =>
      buildCreatePlanRequest(
        context({ resumeId: "resume-unbound" }),
        "en",
        60,
      ),
    ).toThrow("invalid resumeId");
  });

  it("rejects synthetic target ids instead of sending them to generated APIs", () => {
    expect(() =>
      buildCreatePlanRequest(
        context({ targetJobId: "target-job-draft", jobId: "target-job-draft" }),
        "en",
        60,
      ),
    ).toThrow("invalid targetJobId");
  });

  it("rejects a non-positive selected round budget", () => {
    expect(() =>
      buildCreatePlanRequest(
        context({ resumeId: VALID_RESUME_ID }),
        "en",
        0,
      ),
    ).toThrow("invalid timeBudgetMinutes");
  });
});
