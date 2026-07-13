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
      context({ resumeId: VALID_RESUME_ID, roundId: "round-2-technical" }),
      "en",
      60,
    );

    expect(body.targetJobId).toBe(VALID_TARGET_JOB_ID);
    expect(body.resumeId).toBe(VALID_RESUME_ID);
    expect(body.goal).toBe("baseline");
    expect(body.sourceReportId).toBeUndefined();
    expect(body.timeBudgetMinutes).toBe(60);
    expect(body.roundId).toBe("round-2-technical");
    expect(body).not.toHaveProperty("roundSequence");
  });

  it.each([
    ["en", "en"],
    ["zh", "zh-CN"],
    ["zh_cn", "zh-CN"],
    ["zh-cn", "zh-CN"],
    ["zh-CN", "zh-CN"],
  ])("canonicalizes baseline practice language %s to %s", (input, expected) => {
    const body = buildCreatePlanRequest(
      context({ resumeId: VALID_RESUME_ID, roundId: "round-2-technical" }),
      input,
      60,
    );

    expect(body.language).toBe(expected);
  });

  it("rejects an unknown non-empty baseline practice language", () => {
    expect(() =>
      buildCreatePlanRequest(
        context({ resumeId: VALID_RESUME_ID, roundId: "round-2-technical" }),
        "fr",
        60,
      ),
    ).toThrow("invalid language");
  });

  it("creates a closed next_round request from only goal + sourceReportId", () => {
    const body = buildCreatePlanRequest(
      context({
        targetJobId: "",
        jobId: "",
        practiceGoal: "next_round",
        sourceReportId: VALID_REPORT_ID,
      }),
      "fr",
      60,
    );

    expect(body).toEqual({
      goal: "next_round",
      sourceReportId: VALID_REPORT_ID,
    });
  });

  it("creates a closed retry_current_round request without client focus or identity", () => {
    const body = buildCreatePlanRequest(
      context({
        targetJobId: "",
        jobId: "",
        practiceGoal: "retry_current_round",
        sourceReportId: VALID_REPORT_ID,
      }),
      "en",
      0,
    );

    expect(body).toEqual({
      goal: "retry_current_round",
      sourceReportId: VALID_REPORT_ID,
    });
  });

  it("rejects derived report plans without a valid sourceReportId", () => {
    expect(() =>
      buildCreatePlanRequest(
        context({
          resumeId: VALID_RESUME_ID,
          roundId: "round-2-technical",
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
        context({ resumeId: VALID_RESUME_ID, roundId: "round-2-technical" }),
        "en",
        0,
      ),
    ).toThrow("invalid timeBudgetMinutes");
  });

  it("rejects a missing round intention instead of asking the server to infer from UI state", () => {
    expect(() =>
      buildCreatePlanRequest(
        context({ resumeId: VALID_RESUME_ID }),
        "en",
        60,
      ),
    ).toThrow("invalid roundId");
  });
});
