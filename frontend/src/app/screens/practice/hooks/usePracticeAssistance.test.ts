/**
 * @vitest-environment jsdom
 *
 * usePracticeAssistance keeps hints optional inside the session. Out-of-scope
 * practiceMode values and practiceGoal must not hide assistance controls.
 */

import { describe, expect, it } from "vitest";
import { renderHook } from "@testing-library/react";

import { usePracticeAssistance } from "./usePracticeAssistance";

describe("usePracticeAssistance", () => {
  it.each([
    ["assisted", "baseline"],
    ["assisted", "retry_current_round"],
    ["assisted", "next_round"],
    ["strict", "baseline"],
    ["strict", "retry_current_round"],
    ["strict", "next_round"],
  ] as const)(
    "%s × %s keeps hint controls available",
    (mode, goal) => {
      const { result } = renderHook(() =>
        usePracticeAssistance({ practiceMode: mode, practiceGoal: goal }),
      );
      expect(result.current.showHintButton).toBe(true);
      expect(result.current.showStrictBanner).toBe(false);
    },
  );

  it("baseline vs retry_current_round snapshot under assisted is identical", () => {
    const a = renderHook(() =>
      usePracticeAssistance({
        practiceMode: "assisted",
        practiceGoal: "baseline",
      }),
    ).result.current;
    const b = renderHook(() =>
      usePracticeAssistance({
        practiceMode: "assisted",
        practiceGoal: "retry_current_round",
      }),
    ).result.current;
    expect(b).toEqual(a);
  });

  it("baseline vs next_round snapshot under out-of-scope strict input is identical", () => {
    const a = renderHook(() =>
      usePracticeAssistance({
        practiceMode: "strict",
        practiceGoal: "baseline",
      }),
    ).result.current;
    const b = renderHook(() =>
      usePracticeAssistance({
        practiceMode: "strict",
        practiceGoal: "next_round",
      }),
    ).result.current;
    expect(b).toEqual(a);
  });
});
