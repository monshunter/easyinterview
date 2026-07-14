import { describe, expect, it } from "vitest";

import type { TargetJob } from "../../api/generated/types";
import { interviewContextFromTargetJob } from "./interviewContext";

function targetJob(overrides: Partial<TargetJob> = {}): TargetJob {
  return {
    id: "01918fa0-0000-7000-8000-000000002000",
    status: "interviewing",
    analysisStatus: "ready",
    title: "Senior Frontend Engineer",
    companyName: "Acme",
    locationText: "Shanghai",
    targetLanguage: "zh-CN",
    requirements: [],
    openQuestionIssueCount: 0,
    createdAt: "2026-04-22T09:30:00Z",
    updatedAt: "2026-04-28T12:00:00Z",
    ...overrides,
  };
}

const provenance = {
  modelId: "fixture-model:target-import-parse",
  promptVersion: "v0.1.0",
  rubricVersion: "v0.1.0",
  dataSourceVersion: "registry.v1",
  featureFlag: "none",
  language: "en",
};

describe("interviewContextFromTargetJob", () => {
  it("uses server-declared practice plan and resume IDs", () => {
    const ctx = interviewContextFromTargetJob(
      targetJob({
        currentPracticePlanId: "01918fa0-0000-7000-8000-000000004000",
        resumeId: "01918fa0-0000-7000-8000-000000001000",
      }),
    );

    expect(ctx.planId).toBe("01918fa0-0000-7000-8000-000000004000");
    expect(ctx.resumeId).toBe("01918fa0-0000-7000-8000-000000001000");
    expect(JSON.stringify(ctx)).not.toContain("plan-01918fa0");
    expect(JSON.stringify(ctx)).not.toContain("resume-unbound");
  });

  it("does not fabricate plan or resume IDs when the target job has no binding", () => {
    const ctx = interviewContextFromTargetJob(targetJob());

    expect(ctx.planId).toBe("");
    expect(ctx.resumeId).toBe("");
    expect(JSON.stringify(ctx)).not.toContain("plan-01918fa0");
    expect(JSON.stringify(ctx)).not.toContain("resume-unbound");
  });

  it("allows a caller-supplied selected resume to override the target-job binding", () => {
    const ctx = interviewContextFromTargetJob(
      targetJob({
        currentPracticePlanId: "01918fa0-0000-7000-8000-000000004000",
        resumeId: "01918fa0-0000-7000-8000-000000001000",
      }),
      { resumeId: "01918fa0-0000-7000-8000-000000001010" },
    );

    expect(ctx.planId).toBe("01918fa0-0000-7000-8000-000000004000");
    expect(ctx.resumeId).toBe("01918fa0-0000-7000-8000-000000001010");
  });

  it("derives route round context through target-job round assumptions", () => {
    const ctx = interviewContextFromTargetJob(
      targetJob({
        status: "interviewing",
        summary: {
          coreThemes: [],
          interviewRounds: [
            {
              sequence: 1,
              type: "hr",
              name: "Recruiter screen",
              durationMinutes: 30,
              focus: "LLM HR screen probes motivation fit",
            },
            {
              sequence: 2,
              type: "technical",
              name: "Frontend architecture interview",
              durationMinutes: 55,
              focus: "LLM technical round probes frontend architecture",
            },
          ],
          provenance,
        },
        practiceProgress: {
          status: "in_progress",
          completedRounds: [
            { roundId: "round-1-hr", roundSequence: 1 },
          ],
          currentRound: {
            roundId: "round-2-technical",
            roundSequence: 2,
          },
        },
      }),
    );

    expect(ctx.roundId).toBe("round-2-technical");
    expect(ctx.roundName).toBe("Frontend architecture interview · 55m");
    expect(JSON.stringify(ctx)).not.toContain("Technical Round 1");
  });

  it("fails closed for final or invalid backend progress instead of using lifecycle status", () => {
    const base = targetJob({
      status: "interviewing",
      resumeId: "01918fa0-0000-7000-8000-000000001000",
      summary: {
        interviewRounds: [
          { sequence: 1, type: "hr", name: "Recruiter", durationMinutes: 30, focus: "Fit" },
        ],
        provenance,
      },
    });
    const completed = interviewContextFromTargetJob({
      ...base,
      practiceProgress: {
        status: "completed",
        completedRounds: [{ roundId: "round-1-hr", roundSequence: 1 }],
        currentRound: null,
      },
    });
    const missing = interviewContextFromTargetJob({ ...base, status: "offer" });

    expect(completed.roundId).toBe("");
    expect(missing.roundId).toBe("");
  });
});
