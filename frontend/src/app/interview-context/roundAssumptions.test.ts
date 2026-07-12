import { describe, expect, it } from "vitest";

import type { TargetJob } from "../../api/generated/types";
import { resolveTargetJobRoundContext } from "./roundAssumptions";

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
