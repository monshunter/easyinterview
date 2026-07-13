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

function derivedPlan(
  goal: "retry_current_round" | "next_round",
  roundId = goal === "next_round" ? "round-2-technical" : "round-1-technical",
  roundSequence = goal === "next_round" ? 2 : 1,
  timeBudgetMinutes = 60,
) {
  return {
    id: PLAN_ID,
    targetJobId: TARGET_JOB_ID,
    resumeId: RESUME_ID,
    sourceReportId: REPORT_ID,
    goal,
    interviewerPersona: "hiring_manager",
    difficulty: "standard",
    language: "en",
    timeBudgetMinutes,
    status: "ready",
    roundId,
    roundSequence,
    createdAt: "2026-07-12T08:00:00Z",
  };
}

function derivedSession() {
  return {
    id: SESSION_ID,
    planId: PLAN_ID,
    targetJobId: TARGET_JOB_ID,
    language: "en",
    status: "running",
    messages: [],
    createdAt: "2026-07-12T08:01:00Z",
    updatedAt: "2026-07-12T08:01:00Z",
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

  it("creates retry-current-round from only goal + sourceReportId and trusts server-derived context", async () => {
    const client = {
      getTargetJob: vi.fn(),
      getPracticePlan: vi.fn(),
      createPracticePlan: vi.fn().mockResolvedValue(
        derivedPlan("retry_current_round", "round-2-technical", 2),
      ),
      startPracticeSession: vi.fn().mockResolvedValue(derivedSession()),
    } as unknown as EasyInterviewClient;

    const result = await startPracticeFromParams(client, {
      sourceReportId: REPORT_ID,
      practiceGoal: "retry_current_round",
      targetJobId: "route-target-must-be-ignored",
      roundId: "round-99-other",
    }, "en");

    expect(client.createPracticePlan).toHaveBeenCalledWith(
      { goal: "retry_current_round", sourceReportId: REPORT_ID },
      expect.anything(),
    );
    expect(client.getTargetJob).not.toHaveBeenCalled();
    expect(result).toMatchObject({ planId: PLAN_ID, sessionId: SESSION_ID });
    expect(result.params).toMatchObject({
      targetJobId: TARGET_JOB_ID,
      resumeId: RESUME_ID,
      roundId: "round-2-technical",
      practiceGoal: "retry_current_round",
      sourceReportId: REPORT_ID,
    });
  });

  it("propagates a backend rejection for stale next-round intent without mutable context reads", async () => {
    const client = {
      getTargetJob: vi.fn(),
      getPracticePlan: vi.fn(),
      createPracticePlan: vi.fn().mockRejectedValue(new Error("HTTP 409 TARGET_INVALID_STATE_TRANSITION")),
      startPracticeSession: vi.fn(),
    } as unknown as EasyInterviewClient;

    await expect(startPracticeFromParams(client, {
      sourceReportId: REPORT_ID,
      practiceGoal: "next_round",
    }, "en")).rejects.toThrow("TARGET_INVALID_STATE_TRANSITION");
    expect(client.createPracticePlan).toHaveBeenCalledWith(
      { goal: "next_round", sourceReportId: REPORT_ID },
      expect.anything(),
    );
    expect(client.getTargetJob).not.toHaveBeenCalled();
    expect(client.startPracticeSession).not.toHaveBeenCalled();
  });

  it("uses the server-derived non-contiguous next round without sending round identity", async () => {
    const client = {
      getTargetJob: vi.fn(),
      getPracticePlan: vi.fn(),
      createPracticePlan: vi.fn().mockResolvedValue(
        derivedPlan("next_round", "round-4-manager", 4, 45),
      ),
      startPracticeSession: vi.fn().mockResolvedValue(derivedSession()),
    } as unknown as EasyInterviewClient;

    const result = await startPracticeFromParams(client, {
      sourceReportId: REPORT_ID,
      practiceGoal: "next_round",
    }, "en");
    expect(client.createPracticePlan).toHaveBeenCalledWith(
      { goal: "next_round", sourceReportId: REPORT_ID },
      expect.anything(),
    );
    expect(client.getTargetJob).not.toHaveBeenCalled();
    expect(result.params.roundId).toBe("round-4-manager");
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
