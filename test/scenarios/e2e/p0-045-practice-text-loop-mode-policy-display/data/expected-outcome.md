# Expected Outcome — E2E.P0.045

Trigger output evidence:

- `Test Files  N passed`
- `practiceGoalParity.test.tsx` runs and passes
- `practiceHints.test.tsx` runs and passes
- `practiceVoiceTurn.test.tsx` runs and passes
- `practiceSessionContinuity.test.tsx` runs and passes
- `practicePauseResume.test.tsx` runs and passes
- `practiceModeSwitch.test.tsx` runs and passes
- `SessionMap.test.tsx` runs and passes
- session loader, target display, phone controller, VAD/monitor, playback, and microphone lifecycle tests run and pass

Verify gates:

- out-of-scope practice goal literal not present in `frontend/src/app/screens/practice/`.
- `切到语音` and `Switch to voice` literals not present in i18n catalogs.
- `Idempotency-Key.*appendSessionEvent` reverse-grep zero hits.
- Current-boundary grep returns zero hits for strict/role/skip testids and user-visible voice transcription / expression-metric copy in practice runtime files.
- Exactly one production TopBar handset exists; the center hang-up and TopBar handset share `exitPhoneMode`, preserving the current session.
- Voice turn A advances to turn B before the next voice submission; switching route session A to B resets all session-owned screen state before B renders.
- Old segmented/live/restart/call-ended markers are absent across frontend, backend and contract surfaces; production Practice UI does not render raw `questionIntent`.
