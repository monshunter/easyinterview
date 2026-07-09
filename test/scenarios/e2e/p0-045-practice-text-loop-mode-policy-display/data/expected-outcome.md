# Expected Outcome — E2E.P0.045

Trigger output evidence:

- `Test Files  N passed`
- `usePracticeAssistance.test.ts` runs and passes
- `practiceGoalParity.test.tsx` runs and passes
- `practiceHints.test.tsx` runs and passes
- `practiceVoiceTurn.test.tsx` runs and passes
- `practicePauseResume.test.tsx` runs and passes
- `practiceModeSwitch.test.tsx` runs and passes
- `SessionMap.test.tsx` runs and passes

Verify gates:

- non-current practice goal literal not present in `frontend/src/app/screens/practice/`.
- `切到语音` and `Switch to voice` literals not present in i18n catalogs.
- `Idempotency-Key.*appendSessionEvent` reverse-grep zero hits.
- Old voice/strict/skip DOM (`VoiceSessionSurface`, `PracticeAnnotatedWaveform`, `VoiceExpressionPanel`, `practice-topbar-strict`, `practice-input-skip`) zero hits in practice runtime files.
