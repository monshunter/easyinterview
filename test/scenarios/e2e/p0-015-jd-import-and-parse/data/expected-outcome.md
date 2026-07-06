# Expected Outcome

- `targetJob.realApiMode.test.ts` proves upload presign, TargetJob import, TargetJob read, and update operations route to the real backend base URL with cookie credentials and side-effect `Idempotency-Key`.
- `JDAssistModal.test.tsx`, `HomeLayout.test.tsx`, `HomeResumeSelection.test.tsx`, `HomeImport.test.tsx`, `HomeAuthGate.test.tsx`, `ParseScreen.test.tsx`, `ParseFlow.test.tsx`, `ParseFailedState.test.tsx`, and `ParseEdit.test.tsx` run and report passed tests.
- Home no longer renders the retired hero sub copy or old `解析并确认面试` CTA; the main CTA renders as `立即面试` / `Start interview now`.
- Home separates paste textarea from upload/URL source actions, exposes existing ready resumes through a compact dropdown row with the create-resume CTA, keeps `立即面试` below resume selection, keeps import disabled before explicit resume selection, and never creates `importTargetJob` or pending import without a real selected `resumeId`.
- Successful paste/upload/URL import navigates to parse with the selected real `resumeId`.
- `tests/pixel-parity/parse.spec.ts` runs the ready-response loading browser gate and logs `E2E.P0.015 ready-response loading browser gate screenshotBytes=...`.
- Paste, upload, and URL import variants reach parse loading and preview paths; invalid import and failed parse variants show deterministic error UI.
- Polling cadence and preview fields are exercised through fixture-backed tests.
- Trigger log contains no JD raw text markers, source URL markers, console logging, prompt registry, provider key, AIClient, or LLM endpoint references.
- Source-level negative gate under `frontend/src/app/screens/home` and `frontend/src/app/screens/parse` remains clean for provider/model/prompt hard-coding.
