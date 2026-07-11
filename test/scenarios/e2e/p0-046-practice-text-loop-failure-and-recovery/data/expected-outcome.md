# Expected Outcome — E2E.P0.046

Trigger output evidence:

- `Test Files  N passed`
- `practiceSessionLost.test.tsx` runs and passes
- `usePracticeEvents.test.tsx` runs and passes (covers retry-key-reuse)
- `useCompletePracticeSession.test.tsx` runs and passes (covers 3-attempt fallback)
- `practiceErrors.test.tsx` and `practiceVoiceTurn.test.tsx` run and pass
- backend `TestAppendSessionEventSecondInvalidQuestionReturnsSessionWaitWithoutAdvancingTurn` runs and passes

Verify gates:

- `practice.errors.aiTimeout / aiOutputInvalid / network / sessionConflict / unknown / retry / backToWorkspace` keys exist in both `zh.ts` and `en.ts`.
- `frontend/src/app/screens/practice/` source files do NOT contain LLM provider keys, prompt registry, AIClient, or LLM endpoints (negative grep).
- `practice.errors.*` ErrorState renders the right copy per HTTP code.
- Text `session_wait` retains the answer without duplicate transcript and uses a new `clientEventId` on retry; voice `AI_OUTPUT_INVALID` stays in the same session and can return to text mode.
- `appendSessionEvent` replay returns the exact original successful assistant snapshot, while provider timeout degrades to the current `session_wait` policy.
