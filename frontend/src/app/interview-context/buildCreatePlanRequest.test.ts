import { describe, expect, it } from "vitest";

import { buildCreatePlanRequest } from "./buildCreatePlanRequest";
import { DEFAULT_INTERVIEW_CONTEXT, type InterviewContextState } from "./InterviewContext";

const VALID_TARGET_JOB_ID = "01918fa0-0000-7000-8000-000000002000";
const VALID_RESUME_ID = "01918fa0-0000-7000-8000-000000001000";

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
      context({ resumeVersionId: VALID_RESUME_ID }),
      "en",
    );

    expect(body.targetJobId).toBe(VALID_TARGET_JOB_ID);
    expect(body.resumeAssetId).toBe(VALID_RESUME_ID);
  });

  it("treats synthetic resume placeholders as absent before API calls", () => {
    const body = buildCreatePlanRequest(
      context({ resumeVersionId: "resume-unbound" }),
      "en",
    );

    expect(body.targetJobId).toBe(VALID_TARGET_JOB_ID);
    expect(body.resumeAssetId).toBeUndefined();
  });

  it("rejects synthetic target ids instead of sending them to generated APIs", () => {
    expect(() =>
      buildCreatePlanRequest(
        context({ targetJobId: "target-job-draft", jobId: "target-job-draft" }),
        "en",
      ),
    ).toThrow("invalid targetJobId");
  });
});
