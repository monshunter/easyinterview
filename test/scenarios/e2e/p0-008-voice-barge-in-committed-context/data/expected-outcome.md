# E2E.P0.008 Expected Outcome

- `barge_in_detected` is sent before the next recording starts.
- The `appendSessionEvent` calls for voice playback events do not include `Idempotency-Key`.
- Complete playback commits the full assistant text.
- Partial barge-in commits only the played prefix and records an interruption note.
- No-playback state leaves committed context empty.
- The next voice follow-up prompt includes committed assistant context and interruption note, but never includes an unplayed assistant draft.
