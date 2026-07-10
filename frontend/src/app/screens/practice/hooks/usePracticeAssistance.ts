import { useMemo } from "react";

export interface UsePracticeAssistanceInput {
  practiceMode: string;
  practiceGoal: string;
}

export interface UsePracticeAssistanceResult {
  showHintButton: boolean;
  showStrictBanner: boolean;
}

/**
 * Current practice UI treats assistance as an in-session optional action.
 *
 * `practiceGoal` (baseline / retry_current_round / next_round)
 * affects question sourcing only; handoff `practiceMode` is treated as
 * opaque input and must not hide hints or create a visible banner.
 */
export function usePracticeAssistance(
  _input: UsePracticeAssistanceInput,
): UsePracticeAssistanceResult {
  return useMemo<UsePracticeAssistanceResult>(
    () => ({
      showHintButton: true,
      showStrictBanner: false,
    }),
    [],
  );
}
