import { describe, expect, it } from "vitest";

import type { FeedbackReport } from "../../../../api/generated/types";
import {
  ACTION_LABEL_WIRE_MAX_CODE_POINTS,
  isValidReadyReport,
} from "../reportContract";

function readyReport(): FeedbackReport {
  return {
    id: "01918fa0-0000-7000-8000-000000007000",
    sessionId: "01918fa0-0000-7000-8000-000000005000",
    targetJobId: "01918fa0-0000-7000-8000-000000002000",
    status: "ready",
    errorCode: null,
    summary: "Grounded summary.",
    context: {
      sourcePlanId: "01918fa0-0000-7000-8000-000000004000",
      targetJobTitle: "Platform Engineer",
      targetJobCompany: "Acme",
      resumeId: "01918fa0-0000-7000-8000-000000001000",
      resumeDisplayName: "Platform resume",
      roundId: "round-1-technical",
      roundSequence: 1,
      roundName: "Technical interview",
      roundType: "technical",
      language: "en",
      hasNextRound: true,
    },
    preparednessLevel: "needs_practice",
    dimensionAssessments: [
      { code: "technical_depth", label: "Technical depth", status: "needs_work", confidence: "medium" },
    ],
    highlights: [],
    issues: [
      { dimensionCode: "technical_depth", evidence: "The result needs a measurable anchor.", confidence: "medium" },
    ],
    nextActions: [
      { type: "retry_current_round", label: "Practice the current round with a measurable result." },
    ],
    retryFocusDimensionCodes: ["technical_depth"],
    provenance: {
      promptVersion: "v0.2.0",
      rubricVersion: "v0.2.0",
      modelId: "fixture",
      language: "en",
      featureFlag: "none",
      dataSourceVersion: "fixture.v1",
    },
    createdAt: "2026-07-12T08:30:00Z",
    updatedAt: "2026-07-12T08:31:00Z",
  };
}

function mutate(
  change: (value: Record<string, unknown>) => void,
): FeedbackReport {
  const value = structuredClone(readyReport()) as unknown as Record<string, unknown>;
  change(value);
  return value as unknown as FeedbackReport;
}

describe("ready FeedbackReport fail-closed contract", () => {
  it("accepts the direct report and legal empty generic replay focus", () => {
    expect(isValidReadyReport(readyReport())).toBe(true);
    expect(isValidReadyReport(mutate((value) => {
      value.retryFocusDimensionCodes = [];
    }))).toBe(true);
  });

  it.each([
    ["missing top-level field", (value: Record<string, unknown>) => { delete value.summary; }],
    ["unknown top-level field", (value: Record<string, unknown>) => { value.legacyScore = 88; }],
    ["null context", (value: Record<string, unknown>) => { value.context = null; }],
    ["unknown context field", (value: Record<string, unknown>) => {
      (value.context as Record<string, unknown>).legacyRound = "round-1";
    }],
    ["null dimension", (value: Record<string, unknown>) => { value.dimensionAssessments = [null]; }],
    ["unknown dimension field", (value: Record<string, unknown>) => {
      (value.dimensionAssessments as Array<Record<string, unknown>>)[0]!.score = 90;
    }],
    ["null evidence", (value: Record<string, unknown>) => { value.issues = [null]; }],
    ["unknown evidence field", (value: Record<string, unknown>) => {
      (value.issues as Array<Record<string, unknown>>)[0]!.quote = "legacy";
    }],
    ["null action", (value: Record<string, unknown>) => { value.nextActions = [null]; }],
    ["unknown action field", (value: Record<string, unknown>) => {
      (value.nextActions as Array<Record<string, unknown>>)[0]!.payload = {};
    }],
    ["unknown provenance field", (value: Record<string, unknown>) => {
      (value.provenance as Record<string, unknown>).providerSecret = "must-not-pass";
    }],
  ])("rejects %s without throwing", (_label, change) => {
    expect(() => isValidReadyReport(mutate(change))).not.toThrow();
    expect(isValidReadyReport(mutate(change))).toBe(false);
  });

  it("rejects cross-field semantic shapes that the direct report contract forbids", () => {
    const sevenEvidence = mutate((value) => {
      const issue = (value.issues as Array<Record<string, unknown>>)[0];
      value.highlights = Array.from({ length: 3 }, (_, index) => ({ ...issue, evidence: `Highlight ${index}` }));
      value.issues = Array.from({ length: 4 }, (_, index) => ({ ...issue, evidence: `Issue ${index}` }));
    });
    expect(isValidReadyReport(sevenEvidence)).toBe(false);

    const noNeedsWork = mutate((value) => {
      value.dimensionAssessments = [
        { code: "technical_depth", label: "Technical depth", status: "meets_bar", confidence: "medium" },
      ];
      value.retryFocusDimensionCodes = [];
    });
    expect(isValidReadyReport(noNeedsWork)).toBe(false);
  });

  it("enforces two actions and language-aware user-facing label limits", () => {
    expect(ACTION_LABEL_WIRE_MAX_CODE_POINTS).toBe(200);

    const twentyFourWords = Array.from({ length: 24 }, (_, index) => `word${index + 1}`).join(" ");
    expect(isValidReadyReport(mutate((value) => {
      (value.nextActions as Array<Record<string, unknown>>)[0]!.label = twentyFourWords;
    }))).toBe(true);
    expect(isValidReadyReport(mutate((value) => {
      (value.nextActions as Array<Record<string, unknown>>)[0]!.label = `${twentyFourWords} word25`;
    }))).toBe(false);
    const feffSeparated = Array.from({ length: 25 }, (_, index) => `word${index + 1}`).join("\uFEFF");
    expect(isValidReadyReport(mutate((value) => {
      (value.nextActions as Array<Record<string, unknown>>)[0]!.label = feffSeparated;
    }))).toBe(false);
    const nelSeparated = Array.from({ length: 25 }, (_, index) => `word${index + 1}`).join("\u0085");
    expect(isValidReadyReport(mutate((value) => {
      (value.nextActions as Array<Record<string, unknown>>)[0]!.label = nelSeparated;
    }))).toBe(true);

    expect(isValidReadyReport(mutate((value) => {
      const context = value.context as Record<string, unknown>;
      const provenance = value.provenance as Record<string, unknown>;
      context.language = "zh-CN";
      provenance.language = "zh-CN";
      (value.nextActions as Array<Record<string, unknown>>)[0]!.label = "字".repeat(64);
    }))).toBe(true);
    expect(isValidReadyReport(mutate((value) => {
      const context = value.context as Record<string, unknown>;
      const provenance = value.provenance as Record<string, unknown>;
      context.language = "zh-CN";
      provenance.language = "zh-CN";
      (value.nextActions as Array<Record<string, unknown>>)[0]!.label = "字".repeat(65);
    }))).toBe(false);

    expect(isValidReadyReport(mutate((value) => {
      const context = value.context as Record<string, unknown>;
      const provenance = value.provenance as Record<string, unknown>;
      context.language = "zh-CN";
      provenance.language = "zh-CN";
      (value.nextActions as Array<Record<string, unknown>>)[0]!.label = "🧪".repeat(64);
    }))).toBe(true);
    expect(isValidReadyReport(mutate((value) => {
      const context = value.context as Record<string, unknown>;
      const provenance = value.provenance as Record<string, unknown>;
      context.language = "zh-CN";
      provenance.language = "zh-CN";
      (value.nextActions as Array<Record<string, unknown>>)[0]!.label = "🧪".repeat(65);
    }))).toBe(false);

    expect(isValidReadyReport(mutate((value) => {
      value.nextActions = [
        { type: "retry_current_round", label: "Retry this round" },
        { type: "review_evidence", label: "Review the cited evidence" },
        { type: "next_round", label: "Start the next round" },
      ];
    }))).toBe(false);
  });

  it("fails closed on a 201-code-point malformed action", () => {
    const malformed = "x".repeat(ACTION_LABEL_WIRE_MAX_CODE_POINTS + 1);
    const value = mutate((report) => {
      (report.nextActions as Array<Record<string, unknown>>)[0]!.label = malformed;
    });

    expect(isValidReadyReport(value)).toBe(false);
  });
});
