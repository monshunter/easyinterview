import { describe, expect, it, vi } from "vitest";

import type { EasyInterviewClient } from "../../api/generated/client";
import { startPracticeFromParams } from "./startPractice";

const TARGET_JOB_ID = "01918fa0-0000-7000-8000-000000002000";
const RESUME_ID = "01918fa0-0000-7000-8000-000000001000";
const PLAN_ID = "01918fa0-0000-7000-8000-000000004000";
const SESSION_ID = "01918fa0-0000-7000-8000-000000005000";
const REPORT_ID = "01918fa0-0000-7000-8000-000000007000";

function targetJob() {
  return {
    id: TARGET_JOB_ID,
    summary: {
      interviewRounds: [
        {
          sequence: 1,
          type: "technical",
          name: "Technical one",
          durationMinutes: 45,
          focus: "Coding",
        },
        {
          sequence: 2,
          type: "technical",
          name: "Technical two",
          durationMinutes: 60,
          focus: "Architecture",
        },
      ],
    },
  };
}

describe("startPracticeFromParams route output", () => {
  it("returns only current practice route params", async () => {
    const client = {
      getTargetJob: vi.fn().mockResolvedValue(targetJob()),
      getPracticePlan: vi.fn().mockResolvedValue({
        id: PLAN_ID,
        targetJobId: TARGET_JOB_ID,
        resumeId: RESUME_ID,
        timeBudgetMinutes: 60,
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
        roundId: "round-2-technical",
        roundName: "Technical two · 60m",
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
      roundId: "round-2-technical",
      roundName: "Technical two · 60m",
      practiceGoal: "baseline",
      language: "en",
      planId: PLAN_ID,
      sessionId: SESSION_ID,
    });
    expect(client.createPracticePlan).not.toHaveBeenCalled();
  });

  it("creates a plan with the selected round duration when the existing budget drifted", async () => {
    const replacementPlanId = "01918fa0-0000-7000-8000-000000004001";
    const client = {
      getTargetJob: vi.fn().mockResolvedValue(targetJob()),
      getPracticePlan: vi.fn().mockResolvedValue({
        id: PLAN_ID,
        targetJobId: TARGET_JOB_ID,
        resumeId: RESUME_ID,
        timeBudgetMinutes: 45,
        status: "ready",
      }),
      createPracticePlan: vi.fn().mockResolvedValue({ id: replacementPlanId }),
      startPracticeSession: vi.fn().mockResolvedValue({ id: SESSION_ID }),
    } as unknown as EasyInterviewClient;

    const result = await startPracticeFromParams(client, {
      planId: PLAN_ID,
      targetJobId: TARGET_JOB_ID,
      resumeId: RESUME_ID,
      roundId: "round-2-technical",
      roundName: "Technical two · 60m",
      practiceGoal: "baseline",
    }, "en");

    expect(client.createPracticePlan).toHaveBeenCalledWith(
      expect.objectContaining({ timeBudgetMinutes: 60 }),
      expect.anything(),
    );
    expect(result.planId).toBe(replacementPlanId);
  });

  it("fails closed before plan/session creation when roundId is unknown", async () => {
    const client = {
      getTargetJob: vi.fn().mockResolvedValue(targetJob()),
      getPracticePlan: vi.fn(),
      createPracticePlan: vi.fn(),
      startPracticeSession: vi.fn(),
    } as unknown as EasyInterviewClient;

    await expect(startPracticeFromParams(client, {
      targetJobId: TARGET_JOB_ID,
      resumeId: RESUME_ID,
      roundId: "round-9-other",
      practiceGoal: "baseline",
    }, "en")).rejects.toThrow("invalid roundId");
    expect(client.getPracticePlan).not.toHaveBeenCalled();
    expect(client.createPracticePlan).not.toHaveBeenCalled();
    expect(client.startPracticeSession).not.toHaveBeenCalled();
  });
});
