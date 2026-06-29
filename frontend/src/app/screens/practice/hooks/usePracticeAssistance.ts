import { useMemo } from "react";

export interface UsePracticeAssistanceInput {
  practiceMode: string;
  practiceGoal: string;
}

export interface UsePracticeAssistanceResult {
  showLiveNotes: boolean;
  showHintButton: boolean;
  showExperienceCards: boolean;
  showStrictBanner: boolean;
}

/**
 * Item 3.1 — display flags driven by `practiceMode==='strict'` ONLY.
 *
 * `practiceGoal` (baseline / retry_current_round / next_round)
 * affects question sourcing on the backend; it MUST NOT influence the
 * front-end strict-vs-assisted display switch (spec D-3 verifier).
 *
 * `practiceGoal` is intentionally not read in this hook. The argument is
 * accepted only so call sites can keep their full context object intact.
 */
export function usePracticeAssistance(
  input: UsePracticeAssistanceInput,
): UsePracticeAssistanceResult {
  const isStrict = input.practiceMode === "strict";
  return useMemo<UsePracticeAssistanceResult>(
    () => ({
      showLiveNotes: !isStrict,
      showHintButton: !isStrict,
      showExperienceCards: !isStrict,
      showStrictBanner: isStrict,
    }),
    [isStrict],
  );
}
