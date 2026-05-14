# Expected Outcome — E2E.P0.046

Trigger output evidence:

- `Test Files  N passed`
- `practiceSessionLost.test.tsx` runs and passes
- `usePracticeEvents.test.tsx` runs and passes (covers retry-key-reuse)
- `useCompletePracticeSession.test.tsx` runs and passes (covers 3-attempt fallback)

Verify gates:

- `practice.errors.aiTimeout / network / sessionConflict / strictHintConflict / unknown / retry / backToWorkspace` keys exist in both `zh.ts` and `en.ts`.
- `frontend/src/app/screens/practice/` source files do NOT contain LLM provider keys, prompt registry, AIClient, or LLM endpoints (negative grep).
- `practice.errors.*` ErrorState renders the right copy per HTTP code.
