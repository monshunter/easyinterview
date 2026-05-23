# Expected Outcome

- `targetJob.realApiMode.test.ts` proves upload presign, TargetJob import, TargetJob read, and update operations route to the real backend base URL with cookie credentials and side-effect `Idempotency-Key`.
- `JDAssistModal.test.tsx`, `HomeImport.test.tsx`, `HomeAuthGate.test.tsx`, `ParseScreen.test.tsx`, `ParseFlow.test.tsx`, `ParseFailedState.test.tsx`, and `ParseEdit.test.tsx` run and report passed tests.
- Paste, upload, and URL import variants reach parse loading and preview paths; invalid import and failed parse variants show deterministic error UI.
- Polling cadence and preview fields are exercised through fixture-backed tests.
- Trigger log contains no JD raw text markers, source URL markers, console logging, prompt registry, provider key, AIClient, or LLM endpoint references.
- Source-level negative gate under `frontend/src/app/screens/home` and `frontend/src/app/screens/parse` remains clean for provider/model/prompt hard-coding.
