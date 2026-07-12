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
    practiceProgress: {
      status: "in_progress",
      completedRounds: [
        { roundId: "round-1-technical", roundSequence: 1 },
      ],
      currentRound: { roundId: "round-2-technical", roundSequence: 2 },
    },
    summary: {
      interviewRounds: [
        {
          sequence: 1,
          type: "technical",
          name: "Technical one",
          durationMinutes: 60,
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
        roundId: "round-2-technical",
        roundSequence: 2,
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
      createPracticePlan: vi.fn().mockResolvedValue({
        id: replacementPlanId,
        targetJobId: TARGET_JOB_ID,
        resumeId: RESUME_ID,
        timeBudgetMinutes: 60,
        status: "ready",
        roundId: "round-2-technical",
        roundSequence: 2,
      }),
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

  it.each([
    ["equal-duration adjacent round", { roundId: "round-1-technical", roundSequence: 1 }],
    ["legacy null identity", { roundId: null, roundSequence: null }],
  ])("does not reuse a ready %s plan", async (_label, identity) => {
    const replacementPlanId = "01918fa0-0000-7000-8000-000000004002";
    const client = {
      getTargetJob: vi.fn().mockResolvedValue(targetJob()),
      getPracticePlan: vi.fn().mockResolvedValue({
        id: PLAN_ID,
        targetJobId: TARGET_JOB_ID,
        resumeId: RESUME_ID,
        timeBudgetMinutes: 60,
        status: "ready",
        ...identity,
      }),
      createPracticePlan: vi.fn().mockResolvedValue({
        id: replacementPlanId,
        targetJobId: TARGET_JOB_ID,
        resumeId: RESUME_ID,
        timeBudgetMinutes: 60,
        status: "ready",
        roundId: "round-2-technical",
        roundSequence: 2,
      }),
      startPracticeSession: vi.fn().mockResolvedValue({ id: SESSION_ID }),
    } as unknown as EasyInterviewClient;

    const result = await startPracticeFromParams(client, {
      planId: PLAN_ID,
      targetJobId: TARGET_JOB_ID,
      resumeId: RESUME_ID,
      roundId: "round-2-technical",
      practiceGoal: "baseline",
    }, "en");

    expect(client.createPracticePlan).toHaveBeenCalledWith(
      expect.objectContaining({ roundId: "round-2-technical", timeBudgetMinutes: 60 }),
      expect.anything(),
    );
    expect(result.planId).toBe(replacementPlanId);
  });

  it("rejects a mismatched create response before starting a session", async () => {
    const client = {
      getTargetJob: vi.fn().mockResolvedValue(targetJob()),
      getPracticePlan: vi.fn(),
      createPracticePlan: vi.fn().mockResolvedValue({
        id: PLAN_ID,
        targetJobId: TARGET_JOB_ID,
        resumeId: RESUME_ID,
        timeBudgetMinutes: 60,
        status: "ready",
        roundId: "round-1-technical",
        roundSequence: 1,
      }),
      startPracticeSession: vi.fn(),
    } as unknown as EasyInterviewClient;

    await expect(startPracticeFromParams(client, {
      targetJobId: TARGET_JOB_ID,
      resumeId: RESUME_ID,
      roundId: "round-2-technical",
      practiceGoal: "baseline",
    }, "en")).rejects.toThrow("practice plan round mismatch");
    expect(client.startPracticeSession).not.toHaveBeenCalled();
  });

  it.each([
    ["fully completed", {
      status: "completed" as const,
      completedRounds: [
        { roundId: "round-1-technical", roundSequence: 1 },
        { roundId: "round-2-technical", roundSequence: 2 },
      ],
      currentRound: null,
    }],
    ["missing projection", undefined],
  ])("fails closed for %s baseline progress with zero plan/session calls", async (_label, practiceProgress) => {
    const client = {
      getTargetJob: vi.fn().mockResolvedValue({ ...targetJob(), practiceProgress }),
      getPracticePlan: vi.fn(),
      createPracticePlan: vi.fn(),
      startPracticeSession: vi.fn(),
    } as unknown as EasyInterviewClient;

    await expect(startPracticeFromParams(client, {
      targetJobId: TARGET_JOB_ID,
      resumeId: RESUME_ID,
      roundId: "round-2-technical",
      practiceGoal: "baseline",
    }, "en")).rejects.toThrow("invalid practice progress");
    expect(client.getPracticePlan).not.toHaveBeenCalled();
    expect(client.createPracticePlan).not.toHaveBeenCalled();
    expect(client.startPracticeSession).not.toHaveBeenCalled();
  });

  it("allows retry-current-round after final completion and leaves validation to the server", async () => {
    const completedJob = {
      ...targetJob(),
      practiceProgress: {
        status: "completed",
        completedRounds: [
          { roundId: "round-1-technical", roundSequence: 1 },
          { roundId: "round-2-technical", roundSequence: 2 },
        ],
        currentRound: null,
      },
    };
    const client = {
      getTargetJob: vi.fn().mockResolvedValue(completedJob),
      getPracticePlan: vi.fn(),
      createPracticePlan: vi.fn().mockResolvedValue({
        id: PLAN_ID,
        targetJobId: TARGET_JOB_ID,
        resumeId: RESUME_ID,
        timeBudgetMinutes: 60,
        status: "ready",
        roundId: "round-2-technical",
        roundSequence: 2,
      }),
      startPracticeSession: vi.fn().mockResolvedValue({ id: SESSION_ID }),
    } as unknown as EasyInterviewClient;

    await expect(startPracticeFromParams(client, {
      targetJobId: TARGET_JOB_ID,
      resumeId: RESUME_ID,
      sourceReportId: REPORT_ID,
      roundId: "round-2-technical",
      practiceGoal: "retry_current_round",
    }, "en")).resolves.toMatchObject({ planId: PLAN_ID, sessionId: SESSION_ID });
  });

  it("rejects stale next-round intent that no longer matches backend current", async () => {
    const client = {
      getTargetJob: vi.fn().mockResolvedValue(targetJob()),
      getPracticePlan: vi.fn(),
      createPracticePlan: vi.fn(),
      startPracticeSession: vi.fn(),
    } as unknown as EasyInterviewClient;

    await expect(startPracticeFromParams(client, {
      targetJobId: TARGET_JOB_ID,
      resumeId: RESUME_ID,
      sourceReportId: REPORT_ID,
      roundId: "round-1-technical",
      practiceGoal: "next_round",
    }, "en")).rejects.toThrow("round is not backend current");
    expect(client.createPracticePlan).not.toHaveBeenCalled();
    expect(client.startPracticeSession).not.toHaveBeenCalled();
  });

  it("starts the next existing round when canonical sequences are non-contiguous", async () => {
    const nonContiguousJob = {
      ...targetJob(),
      practiceProgress: {
        status: "in_progress",
        completedRounds: [
          { roundId: "round-1-technical", roundSequence: 1 },
          { roundId: "round-2-technical", roundSequence: 2 },
        ],
        currentRound: { roundId: "round-4-manager", roundSequence: 4 },
      },
      summary: {
        interviewRounds: [
          ...targetJob().summary.interviewRounds,
          {
            sequence: 4,
            type: "manager",
            name: "Manager round",
            durationMinutes: 45,
            focus: "Ownership",
          },
        ],
      },
    };
    const client = {
      getTargetJob: vi.fn().mockResolvedValue(nonContiguousJob),
      getPracticePlan: vi.fn(),
      createPracticePlan: vi.fn().mockResolvedValue({
        id: PLAN_ID,
        targetJobId: TARGET_JOB_ID,
        resumeId: RESUME_ID,
        timeBudgetMinutes: 45,
        status: "ready",
        roundId: "round-4-manager",
        roundSequence: 4,
      }),
      startPracticeSession: vi.fn().mockResolvedValue({ id: SESSION_ID }),
    } as unknown as EasyInterviewClient;

    await expect(startPracticeFromParams(client, {
      targetJobId: TARGET_JOB_ID,
      resumeId: RESUME_ID,
      sourceReportId: REPORT_ID,
      roundId: "round-4-manager",
      practiceGoal: "next_round",
    }, "en")).resolves.toMatchObject({ planId: PLAN_ID, sessionId: SESSION_ID });
    expect(client.createPracticePlan).toHaveBeenCalledWith(
      expect.objectContaining({ roundId: "round-4-manager", timeBudgetMinutes: 45 }),
      expect.anything(),
    );
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
