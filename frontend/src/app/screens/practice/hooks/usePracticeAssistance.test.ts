/**
 * @vitest-environment jsdom
 *
 * Item 3.1 — usePracticeAssistance derives UI display flags from
 * `practiceMode==='strict'` ONLY. `practiceGoal` MUST NOT affect display.
 */

import { describe, expect, it } from "vitest";
import { renderHook } from "@testing-library/react";

import { usePracticeAssistance } from "./usePracticeAssistance";

describe("usePracticeAssistance", () => {
  it.each([
    ["assisted", "baseline"],
    ["assisted", "debrief"],
    ["assisted", "retry_current_round"],
    ["assisted", "next_round"],
  ] as const)(
    "assisted × %s × %s shows live notes / hint button / experience cards",
    (mode, goal) => {
      const { result } = renderHook(() =>
        usePracticeAssistance({ practiceMode: mode, practiceGoal: goal }),
      );
      expect(result.current.showLiveNotes).toBe(true);
      expect(result.current.showHintButton).toBe(true);
      expect(result.current.showExperienceCards).toBe(true);
      expect(result.current.showStrictBanner).toBe(false);
    },
  );

  it.each([
    ["strict", "baseline"],
    ["strict", "debrief"],
    ["strict", "retry_current_round"],
    ["strict", "next_round"],
  ] as const)(
    "strict × %s × %s hides live notes / hint button / experience cards",
    (mode, goal) => {
      const { result } = renderHook(() =>
        usePracticeAssistance({ practiceMode: mode, practiceGoal: goal }),
      );
      expect(result.current.showLiveNotes).toBe(false);
      expect(result.current.showHintButton).toBe(false);
      expect(result.current.showExperienceCards).toBe(false);
      expect(result.current.showStrictBanner).toBe(true);
    },
  );

  it("baseline vs debrief snapshot under assisted is identical", () => {
    const a = renderHook(() =>
      usePracticeAssistance({
        practiceMode: "assisted",
        practiceGoal: "baseline",
      }),
    ).result.current;
    const b = renderHook(() =>
      usePracticeAssistance({
        practiceMode: "assisted",
        practiceGoal: "debrief",
      }),
    ).result.current;
    expect(b).toEqual(a);
  });

  it("baseline vs debrief snapshot under strict is identical", () => {
    const a = renderHook(() =>
      usePracticeAssistance({
        practiceMode: "strict",
        practiceGoal: "baseline",
      }),
    ).result.current;
    const b = renderHook(() =>
      usePracticeAssistance({
        practiceMode: "strict",
        practiceGoal: "debrief",
      }),
    ).result.current;
    expect(b).toEqual(a);
  });
});
