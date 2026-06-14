# P0.085 Expected Outcome

- Polling banner is info while phase=polling, danger with retry CTA while phase=failed/timeout/error.
- `getResumeTailorRun` is called with exactly one positional arg (no opts, no IK header).
- `requestResumeTailor` always carries an Idempotency-Key matching `/^v1\.\d+\..+/`.
- Same body fingerprint replays the same IK; mode change rotates it.
- onReady fires once when polling reaches ready.
- Component unmount cancels the next scheduled poll - fake timers advance past the next tick with no additional calls.
- Privacy: originalBullet / suggestedBullet text never appears in URL / localStorage / fetch transport log.
