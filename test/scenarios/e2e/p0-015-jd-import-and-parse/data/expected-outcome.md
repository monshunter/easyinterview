# Expected Outcome

- `targetJob.realApiMode.test.ts` proves paste-only TargetJob import, TargetJob read, and update operations route to the real backend base URL with cookie credentials, side-effect `Idempotency-Key`, and exact `{rawText,targetLanguage,resumeId}` body.
- `HomeScreen.test.tsx`, `HomeLayout.test.tsx`, `HomeResumeSelection.test.tsx`, `HomeImport.test.tsx`, `HomeAuthGate.test.tsx`, `pendingImportState.test.ts`, `ParseScreen.test.tsx`, `ParseFlow.test.tsx`, `ParseFailedState.test.tsx`, and `ParseEdit.test.tsx` run and report passed tests.
- Home omits the out-of-scope hero sub copy and `解析并确认面试` CTA; the main CTA renders as `立即面试` / `Start interview now`.
- Home contains one paste textarea and no source controls, upload / URL trigger, or assist modal; it keeps import disabled before non-blank text and explicit resume selection, and never creates `importTargetJob` or pending import without a real selected `resumeId`.
- Successful paste import navigates to `/parse?targetJobId=...`; `resumeId` and all parsed business data stay out of the URL.
- Signed-out continuation stores only `opaquePendingImportId`; normal consume dispatches once with the original idempotency key, while missing / expired / duplicate consume dispatches zero imports.
- `tests/pixel-parity/home.spec.ts` runs `paste-only Home matches the UI truth and captures desktop/mobile evidence` and emits viewport-specific `E2E.P0.014 home paste-only browser gate` markers; `tests/pixel-parity/parse.spec.ts` runs both the processing-response metadata negative gate and `parse loading matches the UI truth at desktop and mobile`, emitting viewport-specific formal screenshots plus formal/prototype DOM, computed-style, bounding-box, viewport and pixel-parity markers.
- Paste success reaches Parse command progress; ready immediately replaces to target-scoped Workspace detail. Blank / 4xx import and failed parse variants show deterministic error UI.
- StrictMode queued/ready initial read emits one underlying `getTargetJob`; each scheduler tick emits exactly one later transport, and ready-to-Workspace terminal handoff is exercised through fixture-backed tests; saved-plan fields never render on the Parse route.
- Trigger log contains no JD raw text, source URL, console logging, prompt registry, provider key, AIClient, or LLM endpoint content.
- Source-level negative gate under `frontend/src/app/screens/home` and `frontend/src/app/screens/parse` remains clean for provider/model/prompt hard-coding.
