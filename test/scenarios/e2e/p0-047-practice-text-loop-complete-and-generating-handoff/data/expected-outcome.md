# Expected Outcome — E2E.P0.047

Trigger output evidence:

- `Test Files  N passed`
- `useCompletePracticeSession.test.tsx` runs and passes (5 cases)
- `practiceHandoffParams.test.ts` runs and passes (4 cases)
- `completePracticeSessionBody.test.tsx` runs and passes (3 cases)
- `practiceCompletion.test.tsx` runs and passes (2 cases)

Verify gates:

- `completePracticeSession` POST body keys = exactly `clientCompletedAt`.
- `Idempotency-Key` header always present on `completePracticeSession`; reverse-grep `Idempotency-Key.*appendSessionEvent` returns zero hits.
- `nav.generating` params include the 14 stable fields and exclude raw text / provenance.
- `getFeedbackReport` and `createPracticeVoiceTurn` runtime calls = 0.
