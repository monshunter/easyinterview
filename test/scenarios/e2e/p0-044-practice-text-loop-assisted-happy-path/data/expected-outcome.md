# Expected Outcome — E2E.P0.044

Trigger output evidence:

- `Test Files  N passed`
- `PracticeScreen.test.tsx` runs and passes
- `usePracticeEvents.test.tsx` runs and passes
- `usePracticeSessionLoader.test.tsx` runs and passes
- `AssistantActionRenderer.test.tsx` runs and passes
- `practiceModeSwitch.test.tsx` runs and passes
- `idempotencyContract.test.tsx` runs and passes
- `appendSessionEventBody.test.tsx` runs and passes

Verify gates:

- `≥ 20 unique practice-* testids` rendered in `PracticeScreen.tsx` source.
- `appendSessionEvent` request init carries no `Idempotency-Key` header.
- `clientEventId` matches UUIDv7 regex.
- Negative grep for `VoiceSessionSurface | PracticeWaveformBars | PracticeAnnotatedWaveform | VoiceExpressionPanel | window.EI_DATA | getPracticeSampleQuestions | getPracticeSampleTranscript | getPracticeWaveformSamples | practice-mode-card- | growth-summary | drill-builder- | mistakes-queue- | getFeedbackReport | createPracticeVoiceTurn` returns zero hits in `frontend/src/app/screens/practice/` non-test files.
