import { describe, expect, it } from "vitest";

import type { ApiErrorCode, TargetJob } from "./generated/types";
import { buildTargetJobRoundAssumptions } from "../app/interview-context/roundAssumptions";
import { validateTargetJobReportsOverview } from "./targetJobReports";

const targetJob: TargetJob = {
  id: "01918fa0-0000-7000-8000-000000002000",
  title: "Frontend Engineer",
  companyName: "Acme",
  analysisStatus: "ready",
  status: "draft",
  targetLanguage: "en",
  requirements: [],
  openQuestionIssueCount: 0,
  createdAt: "2026-07-14T08:00:00Z",
  updatedAt: "2026-07-14T08:00:00Z",
  summary: {
    coreThemes: [],
    interviewRounds: [
      {
        sequence: 1,
        type: "technical",
        name: "Architecture",
        durationMinutes: 50,
        focus: "System boundaries",
      },
      {
        sequence: 2,
        type: "manager",
        name: "Manager",
        durationMinutes: 40,
        focus: "Influence",
      },
    ],
    provenance: {
      modelId: "model",
      promptVersion: "prompt",
      rubricVersion: "rubric",
      dataSourceVersion: "source",
      featureFlag: "flag",
      language: "en",
    },
  },
};

function overview(): Record<string, unknown> {
  return {
    targetJobId: "01918fa0-0000-7000-8000-000000002000",
    rounds: [
      {
        round: { roundId: "round-1-technical", roundSequence: 1 },
        currentReport: {
          id: "01918fa0-0000-7000-8000-000000007001",
          generatedAt: "2026-07-13T14:20:00Z",
        },
        latestAttempt: {
          id: "01918fa0-0000-7000-8000-000000007002",
          status: "failed",
          errorCode: "AI_PROVIDER_TIMEOUT",
          createdAt: "2026-07-14T09:14:00Z",
        },
      },
      {
        round: { roundId: "round-2-manager", roundSequence: 2 },
        currentReport: null,
        latestAttempt: {
          id: "01918fa0-0000-7000-8000-000000007003",
          status: "generating",
          errorCode: null,
          createdAt: "2026-07-14T09:15:00Z",
        },
      },
    ],
  };
}

describe("validateTargetJobReportsOverview", () => {
  it("joins the minimal overview to display data from the current TargetJob", () => {
    const result = validateTargetJobReportsOverview(
      overview(),
      targetJob.id,
      buildTargetJobRoundAssumptions(targetJob),
    );

    expect(result.targetJobId).toBe("01918fa0-0000-7000-8000-000000002000");
    expect(result.rounds.map((item) => item.displayRound)).toEqual([
      {
        id: "round-1-technical",
        sequence: 1,
        name: "Architecture · 50m",
        focus: "System boundaries",
        type: "technical",
        durationMinutes: 50,
      },
      {
        id: "round-2-manager",
        sequence: 2,
        name: "Manager · 40m",
        focus: "Influence",
        type: "manager",
        durationMinutes: 40,
      },
    ]);
  });

  it.each([
    "AI_PROVIDER_TIMEOUT",
    "PRIVACY_EXPORT_NOT_AVAILABLE",
  ] as const satisfies readonly ApiErrorCode[])(
    "accepts documented failed error code %s",
    (errorCode) => {
      const value = overview();
      const first = (value.rounds as Record<string, unknown>[])[0]!;
      (first.latestAttempt as Record<string, unknown>).errorCode = errorCode;

      const result = validateTargetJobReportsOverview(
        value,
        targetJob.id,
        buildTargetJobRoundAssumptions(targetJob),
      );
      expect(result.rounds[0]!.latestAttempt?.errorCode).toBe(errorCode);
    },
  );

  it.each([
    ["fractional UTC", "2026-07-14T09:00:00.123456Z"],
    ["explicit offset", "2026-07-14T17:00:00+08:00"],
  ])("accepts RFC3339 date-time with %s", (_label, value) => {
    const candidate = overview();
    const first = (candidate.rounds as Record<string, unknown>[])[0]!;
    (first.currentReport as Record<string, unknown>).generatedAt = value;
    (first.latestAttempt as Record<string, unknown>).createdAt = value;

    const result = validateTargetJobReportsOverview(
      candidate,
      targetJob.id,
      buildTargetJobRoundAssumptions(targetJob),
    );
    expect(result.rounds[0]!.currentReport?.generatedAt).toBe(value);
    expect(result.rounds[0]!.latestAttempt?.createdAt).toBe(value);
  });

  it.each([
    ["numeric-looking string", "0"],
    ["date only", "2026-07-14"],
    ["locale date", "07/14/2026 09:00:00"],
    ["missing timezone", "2026-07-14T09:00:00"],
    ["space instead of T", "2026-07-14 09:00:00Z"],
    ["invalid calendar date", "2026-02-30T09:00:00Z"],
    ["invalid timezone", "2026-07-14T09:00:00+24:00"],
  ])("rejects non-RFC3339 date-time: %s", (_label, value) => {
    const candidate = overview();
    const first = (candidate.rounds as Record<string, unknown>[])[0]!;
    (first.currentReport as Record<string, unknown>).generatedAt = value;

    expect(() => validateTargetJobReportsOverview(
      candidate,
      targetJob.id,
      buildTargetJobRoundAssumptions(targetJob),
    )).toThrow("Invalid TargetJob reports overview");
  });

  it.each([
    ["target mismatch", (value: Record<string, unknown>) => { value.targetJobId = "target-2"; }],
    ["extra root field", (value: Record<string, unknown>) => { value.pageInfo = {}; }],
    ["missing round", (value: Record<string, unknown>) => { (value.rounds as unknown[]).pop(); }],
    ["duplicate round", (value: Record<string, unknown>) => {
      const rounds = value.rounds as Record<string, unknown>[];
      rounds[1] = structuredClone(rounds[0]!);
    }],
    ["out-of-order round", (value: Record<string, unknown>) => {
      (value.rounds as unknown[]).reverse();
    }],
    ["unknown round", (value: Record<string, unknown>) => {
      const first = (value.rounds as Record<string, unknown>[])[0]!;
      (first.round as Record<string, unknown>).roundId = "round-1-hr";
    }],
    ["extra row field", (value: Record<string, unknown>) => {
      (value.rounds as Record<string, unknown>[])[0]!.provenance = {};
    }],
    ["extra round field", (value: Record<string, unknown>) => {
      const first = (value.rounds as Record<string, unknown>[])[0]!;
      (first.round as Record<string, unknown>).name = "route authority";
    }],
    ["extra report field", (value: Record<string, unknown>) => {
      const first = (value.rounds as Record<string, unknown>[])[0]!;
      (first.currentReport as Record<string, unknown>).summary = "full report leak";
    }],
    ["invalid failed error", (value: Record<string, unknown>) => {
      const first = (value.rounds as Record<string, unknown>[])[0]!;
      (first.latestAttempt as Record<string, unknown>).errorCode = null;
    }],
    ["non-failed error", (value: Record<string, unknown>) => {
      const second = (value.rounds as Record<string, unknown>[])[1]!;
      (second.latestAttempt as Record<string, unknown>).errorCode = "AI_PROVIDER_TIMEOUT";
    }],
  ])("fails closed for %s", (_label, mutate) => {
    const value = overview();
    mutate(value);
    expect(() => validateTargetJobReportsOverview(
      value,
      targetJob.id,
      buildTargetJobRoundAssumptions(targetJob),
    )).toThrow(
      "Invalid TargetJob reports overview",
    );
  });

  it.each([
    ["a current report reused by another round", (value: Record<string, unknown>) => {
      const rounds = value.rounds as Record<string, unknown>[];
      rounds[1]!.currentReport = structuredClone(rounds[0]!.currentReport);
    }],
    ["a current report reused as another round's latest attempt", (value: Record<string, unknown>) => {
      const rounds = value.rounds as Record<string, unknown>[];
      const current = rounds[0]!.currentReport as Record<string, unknown>;
      (rounds[1]!.latestAttempt as Record<string, unknown>).id = current.id;
    }],
    ["a non-ready latest attempt sharing its round's current report ID", (value: Record<string, unknown>) => {
      const first = (value.rounds as Record<string, unknown>[])[0]!;
      const current = first.currentReport as Record<string, unknown>;
      (first.latestAttempt as Record<string, unknown>).id = current.id;
    }],
    ["a ready latest attempt without a current report", (value: Record<string, unknown>) => {
      const second = (value.rounds as Record<string, unknown>[])[1]!;
      second.currentReport = null;
      second.latestAttempt = {
        id: "01918fa0-0000-7000-8000-000000007004",
        status: "ready",
        errorCode: null,
        createdAt: "2026-07-14T09:16:00Z",
      };
    }],
    ["a current report without any latest attempt", (value: Record<string, unknown>) => {
      const first = (value.rounds as Record<string, unknown>[])[0]!;
      first.latestAttempt = null;
    }],
  ])("fails closed for impossible report ownership: %s", (_label, mutate) => {
    const value = overview();
    mutate(value);

    expect(() => validateTargetJobReportsOverview(
      value,
      targetJob.id,
      buildTargetJobRoundAssumptions(targetJob),
    )).toThrow("Invalid TargetJob reports overview");
  });

  it("accepts one ready latest attempt sharing its own current report ID", () => {
    const value = overview();
    const first = (value.rounds as Record<string, unknown>[])[0]!;
    const current = first.currentReport as Record<string, unknown>;
    first.latestAttempt = {
      id: current.id,
      status: "ready",
      errorCode: null,
      createdAt: "2026-07-14T09:14:00Z",
    };

    const result = validateTargetJobReportsOverview(
      value,
      targetJob.id,
      buildTargetJobRoundAssumptions(targetJob),
    );
    expect(result.rounds[0]!.latestAttempt?.id).toBe(current.id);
  });

  it("fails closed when the current TargetJob round catalog is noncanonical", () => {
    const invalidJob = structuredClone(targetJob);
    invalidJob.summary!.interviewRounds = invalidJob.summary!.interviewRounds!.slice(0, 1);

    expect(() => validateTargetJobReportsOverview(
      overview(),
      invalidJob.id,
      buildTargetJobRoundAssumptions(invalidJob),
    )).toThrow(
      "Invalid TargetJob reports overview",
    );
  });

  it.each([
    ["starts above one", [2, 3]],
    ["contains a sequence gap", [1, 3]],
  ])("accepts a positive strictly increasing canonical round catalog that %s", (_label, sequences) => {
    const value = overview();
    const rows = value.rounds as Record<string, unknown>[];
    const canonicalRounds = buildTargetJobRoundAssumptions(targetJob).map(
      (round, index) => ({
        ...round,
        id: `round-${sequences[index]}-${round.type}`,
        sequence: sequences[index]!,
      }),
    );
    rows.forEach((row, index) => {
      row.round = {
        roundId: canonicalRounds[index]!.id,
        roundSequence: sequences[index]!,
      };
    });

    const result = validateTargetJobReportsOverview(
      value,
      targetJob.id,
      canonicalRounds,
    );
    expect(result.rounds.map((round) => [
      round.displayRound.id,
      round.displayRound.sequence,
    ])).toEqual(canonicalRounds.map((round) => [round.id, round.sequence]));
  });

  it.each([
    ["contains a non-positive sequence", [0, 2]],
    ["contains a non-integer sequence", [1, 2.5]],
    ["contains a duplicate sequence", [1, 1]],
    ["is not strictly increasing", [2, 1]],
  ])("fails closed when the canonical round catalog %s", (_label, sequences) => {
    const canonicalRounds = buildTargetJobRoundAssumptions(targetJob).map(
      (round, index) => ({ ...round, sequence: sequences[index]! }),
    );

    expect(() => validateTargetJobReportsOverview(
      overview(),
      targetJob.id,
      canonicalRounds,
    )).toThrow(
      "Invalid TargetJob reports overview",
    );
  });

  it.each([
    ["matching malformed target UUID", (value: Record<string, unknown>) => {
      value.targetJobId = "target-not-uuid";
      return "target-not-uuid";
    }],
    ["malformed current report UUID", (value: Record<string, unknown>) => {
      const first = (value.rounds as Record<string, unknown>[])[0]!;
      (first.currentReport as Record<string, unknown>).id = "report-not-uuid";
      return targetJob.id;
    }],
    ["malformed latest-attempt UUID", (value: Record<string, unknown>) => {
      const first = (value.rounds as Record<string, unknown>[])[0]!;
      (first.latestAttempt as Record<string, unknown>).id = "attempt-not-uuid";
      return targetJob.id;
    }],
    ["unknown failed error code", (value: Record<string, unknown>) => {
      const first = (value.rounds as Record<string, unknown>[])[0]!;
      (first.latestAttempt as Record<string, unknown>).errorCode = "REPORT_UNKNOWN_FAILURE";
      return targetJob.id;
    }],
  ])("fails closed for %s", (_label, mutate) => {
    const value = overview();
    const expectedTargetJobId = mutate(value);
    expect(() => validateTargetJobReportsOverview(
      value,
      expectedTargetJobId,
      buildTargetJobRoundAssumptions(targetJob),
    )).toThrow(
      "Invalid TargetJob reports overview",
    );
  });
});
