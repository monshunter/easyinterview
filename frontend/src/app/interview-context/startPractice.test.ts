import { describe, expect, it, vi } from "vitest";

import type { EasyInterviewClient } from "../../api/generated/client";
import { startPracticeFromParams } from "./startPractice";

const TARGET_JOB_ID = "01918fa0-0000-7000-8000-000000002000";
const RESUME_ID = "01918fa0-0000-7000-8000-000000001000";
const PLAN_ID = "01918fa0-0000-7000-8000-000000004000";
const SESSION_ID = "01918fa0-0000-7000-8000-000000005000";
const REPORT_ID = "01918fa0-0000-7000-8000-000000007000";

describe("startPracticeFromParams route output", () => {
  it("returns only current practice route params", async () => {
    const client = {
      getPracticePlan: vi.fn().mockResolvedValue({
        id: PLAN_ID,
        targetJobId: TARGET_JOB_ID,
        resumeId: RESUME_ID,
        status: "ready",
      }),
      createPracticePlan: vi.fn(),
      startPracticeSession: vi.fn().mockResolvedValue({ id: SESSION_ID }),
    } as unknown as EasyInterviewClient;

    const result = await startPracticeFromParams(
      client,
      {
        planId: PLAN_ID,
        targetJobId: TARGET_JOB_ID,
        resumeId: RESUME_ID,
        sourceReportId: REPORT_ID,
        roundId: "round-1",
        roundName: "Technical",
        mode: "text",
        modality: "text",
        practiceMode: "strict",
        practiceGoal: "baseline",
        hintUsed: "false",
        hintCount: "0",
        language: "en",
        autoStartPractice: "1",
        sourceSessionId: "source-session",
        replayItems: "turn-1",
        evidenceGaps: "gap-1",
        rawText: "sensitive",
      },
      "en",
    );

    expect(result.params).toEqual({
      targetJobId: TARGET_JOB_ID,
      jobId: TARGET_JOB_ID,
      jdId: `jd-${TARGET_JOB_ID}`,
      resumeId: RESUME_ID,
      sourceReportId: REPORT_ID,
      roundId: "round-1",
      roundName: "Technical",
      practiceGoal: "baseline",
      language: "en",
      planId: PLAN_ID,
      sessionId: SESSION_ID,
    });
  });
});
