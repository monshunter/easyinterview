# Expected Outcome — E2E.P0.044

Trigger output evidence:

- `Test Files  N passed`
- `PracticeScreen.test.tsx` runs and passes
- `usePracticeEvents.test.tsx` runs and passes
- `usePracticeSessionLoader.test.tsx` runs and passes
- `AssistantActionRenderer.test.tsx` runs and passes
- `outOfScopeNegative.test.ts` runs and passes
- `practiceModeSwitch.test.tsx` runs and passes
- `idempotencyContract.test.tsx` runs and passes
- `appendSessionEventBody.test.tsx` runs and passes

Verify gates:

- `≥ 20 unique practice-* testids` rendered in `PracticeScreen.tsx` source.
- `appendSessionEvent` request init carries no `Idempotency-Key` header.
- `clientEventId` matches UUIDv7 regex.
- Runtime grep confirms practice non-test files do not import `ui-design/src/screen-practice` or prototype data helpers, do not expose out-of-scope `practice-mode-card-` / `growth-summary` / `drill-builder-` / `mistakes-queue-` testids, do not call `getFeedbackReport`, and keep `createPracticeVoiceTurn` confined to the voice owner hook. Truth-source comments may still name the static prototype path.
