import { describe, expect, it } from "vitest";

import type { TargetJob } from "../../api/generated/types";
import {
  resolvePersistedNextRound,
  resolveTargetJobPracticeProgress,
  resolveTargetJobRoundContext,
} from "./roundAssumptions";

function targetJobWithRounds(
  rounds: NonNullable<TargetJob["summary"]>["interviewRounds"],
): Pick<TargetJob, "summary"> {
  return {
    summary: {
      interviewRounds: rounds,
    } as NonNullable<TargetJob["summary"]>,
  };
}

describe("resolveTargetJobRoundContext", () => {
  const rounds = [
    {
      sequence: 3,
      type: "manager" as const,
      name: "Manager round",
      durationMinutes: 40,
      focus: "Influence",
    },
    {
      sequence: 1,
      type: "technical" as const,
      name: "Technical one",
      durationMinutes: 45,
      focus: "Coding",
    },
    {
      sequence: 2,
      type: "technical" as const,
      name: "Technical two",
      durationMinutes: 60,
      focus: "Architecture",
    },
  ];

  it("resolves the current round and its immediate ordered successor", () => {
    const result = resolveTargetJobRoundContext(
      targetJobWithRounds(rounds),
      "round-1-technical",
    );

    expect(result.currentRound).toMatchObject({
      id: "round-1-technical",
      durationMinutes: 45,
    });
    expect(result.nextRound).toMatchObject({
      id: "round-2-technical",
      durationMinutes: 60,
    });
  });

  it.each([
    ["last round", targetJobWithRounds(rounds), "round-3-manager"],
    ["unknown round", targetJobWithRounds(rounds), "round-9-other"],
    ["missing round", targetJobWithRounds(rounds), undefined],
    ["empty rounds", targetJobWithRounds([]), "round-1-technical"],
  ])("fails closed for %s", (_label, job, roundId) => {
    const result = resolveTargetJobRoundContext(job, roundId);
    expect(result.currentRound?.id ?? null).toBe(
      roundId === "round-3-manager" ? "round-3-manager" : null,
    );
    expect(result.nextRound).toBeNull();
  });

  it("fails closed when structured rounds derive duplicate ids", () => {
    const duplicate = targetJobWithRounds([
      rounds[0]!,
      { ...rounds[0]!, name: "Duplicate manager round" },
    ]);

    expect(resolveTargetJobRoundContext(duplicate, "round-3-manager")).toEqual({
      currentRound: null,
      nextRound: null,
    });
  });
});

describe("resolveTargetJobPracticeProgress", () => {
  const rounds = [
    {
      sequence: 1,
      type: "hr" as const,
      name: "Recruiter screen",
      durationMinutes: 30,
      focus: "Motivation",
    },
    {
      sequence: 2,
      type: "technical" as const,
      name: "Technical round",
      durationMinutes: 45,
      focus: "Architecture",
    },
    {
      sequence: 3,
      type: "manager" as const,
      name: "Manager round",
      durationMinutes: 45,
      focus: "Influence",
    },
  ];

  function job(
    practiceProgress: TargetJob["practiceProgress"],
    status: TargetJob["status"] = "interviewing",
  ): Pick<TargetJob, "summary" | "practiceProgress" | "status"> {
    return {
      ...targetJobWithRounds(rounds),
      practiceProgress,
      status,
    };
  }

  it("maps an ordered completed prefix and exact first-incomplete pair", () => {
    const result = resolveTargetJobPracticeProgress(job({
      status: "in_progress",
      completedRounds: [{ roundId: "round-1-hr", roundSequence: 1 }],
      currentRound: { roundId: "round-2-technical", roundSequence: 2 },
    }));

    expect(result.valid).toBe(true);
    expect(result.completedCount).toBe(1);
    expect(result.currentIndex).toBe(1);
    expect(result.currentRound?.id).toBe("round-2-technical");
    expect(result.completed).toBe(false);
  });

  it("maps not-started and fully-completed projections without a fallback cursor", () => {
    const notStarted = resolveTargetJobPracticeProgress(job({
      status: "not_started",
      completedRounds: [],
      currentRound: { roundId: "round-1-hr", roundSequence: 1 },
    }));
    const completed = resolveTargetJobPracticeProgress(job({
      status: "completed",
      completedRounds: [
        { roundId: "round-1-hr", roundSequence: 1 },
        { roundId: "round-2-technical", roundSequence: 2 },
        { roundId: "round-3-manager", roundSequence: 3 },
      ],
      currentRound: null,
    }));

    expect(notStarted).toMatchObject({ valid: true, completedCount: 0, currentIndex: 0, completed: false });
    expect(completed).toMatchObject({ valid: true, completedCount: 3, currentIndex: 3, currentRound: null, completed: true });
  });

  it("is independent of TargetJob lifecycle status", () => {
    const progress = {
      status: "in_progress" as const,
      completedRounds: [{ roundId: "round-1-hr", roundSequence: 1 }],
      currentRound: { roundId: "round-2-technical", roundSequence: 2 },
    };

    expect(resolveTargetJobPracticeProgress(job(progress, "draft")))
      .toEqual(resolveTargetJobPracticeProgress(job(progress, "offer")));
  });

  it.each([
    ["missing projection", undefined],
    ["id/sequence mismatch", {
      status: "in_progress" as const,
      completedRounds: [{ roundId: "round-1-hr", roundSequence: 2 }],
      currentRound: { roundId: "round-2-technical", roundSequence: 2 },
    }],
    ["non-prefix completion", {
      status: "in_progress" as const,
      completedRounds: [{ roundId: "round-2-technical", roundSequence: 2 }],
      currentRound: { roundId: "round-1-hr", roundSequence: 1 },
    }],
    ["duplicate completion", {
      status: "in_progress" as const,
      completedRounds: [
        { roundId: "round-1-hr", roundSequence: 1 },
        { roundId: "round-1-hr", roundSequence: 1 },
      ],
      currentRound: { roundId: "round-2-technical", roundSequence: 2 },
    }],
    ["wrong current round", {
      status: "in_progress" as const,
      completedRounds: [{ roundId: "round-1-hr", roundSequence: 1 }],
      currentRound: { roundId: "round-3-manager", roundSequence: 3 },
    }],
    ["status disagrees with facts", {
      status: "completed" as const,
      completedRounds: [{ roundId: "round-1-hr", roundSequence: 1 }],
      currentRound: { roundId: "round-2-technical", roundSequence: 2 },
    }],
  ])("fails closed for %s", (_label, progress) => {
    expect(resolveTargetJobPracticeProgress(job(progress))).toMatchObject({
      valid: false,
      completedCount: 0,
      currentIndex: null,
      currentRound: null,
      completed: false,
    });
  });

  it("allows report next only when the immediate successor is backend current", () => {
    const current = job({
      status: "in_progress",
      completedRounds: [{ roundId: "round-1-hr", roundSequence: 1 }],
      currentRound: { roundId: "round-2-technical", roundSequence: 2 },
    });
    const stale = job({
      status: "in_progress",
      completedRounds: [
        { roundId: "round-1-hr", roundSequence: 1 },
        { roundId: "round-2-technical", roundSequence: 2 },
      ],
      currentRound: { roundId: "round-3-manager", roundSequence: 3 },
    });

    expect(resolvePersistedNextRound(current, "round-1-hr")?.id).toBe("round-2-technical");
    expect(resolvePersistedNextRound(stale, "round-1-hr")).toBeNull();
  });

  it("treats the next existing non-contiguous sequence as the canonical successor", () => {
    const nonContiguousRounds = [
      rounds[0]!,
      rounds[1]!,
      { ...rounds[2]!, sequence: 4 },
    ];
    const target = {
      ...targetJobWithRounds(nonContiguousRounds),
      practiceProgress: {
        status: "in_progress" as const,
        completedRounds: [
          { roundId: "round-1-hr", roundSequence: 1 },
          { roundId: "round-2-technical", roundSequence: 2 },
        ],
        currentRound: { roundId: "round-4-manager", roundSequence: 4 },
      },
    };

    expect(resolveTargetJobPracticeProgress(target)).toMatchObject({
      valid: true,
      completedCount: 2,
      currentRound: { id: "round-4-manager", sequence: 4 },
    });
    expect(resolvePersistedNextRound(target, "round-2-technical")?.id).toBe("round-4-manager");
  });

  it.each([
    ["too few rounds", [rounds[0]!]],
    ["zero sequence", [{ ...rounds[0]!, sequence: 0 }, rounds[1]!]],
    ["duplicate sequence", [rounds[0]!, { ...rounds[1]!, sequence: 1 }]],
    ["blank name", [{ ...rounds[0]!, name: "  " }, rounds[1]!]],
    ["blank focus", [rounds[0]!, { ...rounds[1]!, focus: "  " }]],
    ["duration below minimum", [{ ...rounds[0]!, durationMinutes: 5 }, rounds[1]!]],
    ["duration above maximum", [rounds[0]!, { ...rounds[1]!, durationMinutes: 181 }]],
    ["unknown runtime type", [rounds[0]!, { ...rounds[1]!, type: "sales" as never }]],
  ])("fails closed for non-canonical structured rounds: %s", (_label, invalidRounds) => {
    const invalid = {
      ...targetJobWithRounds(invalidRounds),
      practiceProgress: {
        status: "not_started" as const,
        completedRounds: [],
        currentRound: {
          roundId: `round-${invalidRounds[0]!.sequence}-${invalidRounds[0]!.type}`,
          roundSequence: invalidRounds[0]!.sequence,
        },
      },
    };

    expect(resolveTargetJobPracticeProgress(invalid)).toMatchObject({
      valid: false,
      currentIndex: null,
      currentRound: null,
    });
    expect(resolveTargetJobRoundContext(invalid, invalid.practiceProgress.currentRound.roundId)).toEqual({
      currentRound: null,
      nextRound: null,
    });
  });
});
