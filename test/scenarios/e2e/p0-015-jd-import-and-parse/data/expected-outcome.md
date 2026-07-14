# Expected Outcome

- `targetJob.realApiMode.test.ts` proves paste-only TargetJob import, TargetJob read, and update operations route to the real backend base URL with cookie credentials, side-effect `Idempotency-Key`, and exact `{rawText,targetLanguage,resumeId}` body.
- `HomeScreen.test.tsx`, `HomeLayout.test.tsx`, `HomeResumeSelection.test.tsx`, `HomeImport.test.tsx`, `HomeAuthGate.test.tsx`, `pendingImportState.test.ts`, `ParseScreen.test.tsx`, `ParseFlow.test.tsx`, `ParseFailedState.test.tsx`, and `ParseEdit.test.tsx` run and report passed tests.
- Home omits the out-of-scope hero sub copy and `解析并确认面试` CTA; the main CTA renders as `立即面试` / `Start interview now`.
- Home contains one paste textarea and no source controls, upload / URL trigger, or assist modal; it keeps import disabled before non-blank text and explicit resume selection, and never creates `importTargetJob` or pending import without a real selected `resumeId`.
- Successful paste import navigates to parse with `targetJobId + resumeId` only.
- Signed-out continuation stores only `opaquePendingImportId`; normal consume dispatches once with the original idempotency key, while missing / expired / duplicate consume dispatches zero imports.
- `tests/pixel-parity/home.spec.ts` runs `paste-only Home matches the UI truth and captures desktop/mobile evidence` and emits viewport-specific `E2E.P0.014 home paste-only browser gate` markers; `tests/pixel-parity/parse.spec.ts` runs both the internal-metadata negative gate and `parse loading matches the UI truth at desktop and mobile`, emitting viewport-specific formal screenshots plus formal/prototype DOM, computed-style, bounding-box, viewport and pixel-parity markers.
- Paste success reaches parse loading and preview; blank / 4xx import and failed parse variants show deterministic error UI.
- Polling cadence and preview fields are exercised through fixture-backed tests.
- Trigger log contains no JD raw text, source URL, console logging, prompt registry, provider key, AIClient, or LLM endpoint content.
- Source-level negative gate under `frontend/src/app/screens/home` and `frontend/src/app/screens/parse` remains clean for provider/model/prompt hard-coding.
