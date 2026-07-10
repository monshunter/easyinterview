# Expected Outcome — E2E.P0.045

Trigger output evidence:

- `Test Files  N passed`
- `practiceGoalParity.test.tsx` runs and passes
- `practiceHints.test.tsx` runs and passes
- `practiceVoiceTurn.test.tsx` runs and passes
- `practicePauseResume.test.tsx` runs and passes
- `practiceModeSwitch.test.tsx` runs and passes
- `SessionMap.test.tsx` runs and passes

Verify gates:

- out-of-scope practice goal literal not present in `frontend/src/app/screens/practice/`.
- `切到语音` and `Switch to voice` literals not present in i18n catalogs.
- `Idempotency-Key.*appendSessionEvent` reverse-grep zero hits.
- Current-boundary grep returns zero hits for strict/role/skip testids and user-visible voice transcription / expression-metric copy in practice runtime files.
