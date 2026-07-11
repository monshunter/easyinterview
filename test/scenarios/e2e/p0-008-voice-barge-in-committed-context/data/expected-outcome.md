# E2E.P0.008 Expected Outcome

- Partial `tts_chunk_played` is sent before `barge_in_detected` and before the next recording starts.
- The `appendSessionEvent` calls for voice playback events do not include `Idempotency-Key`.
- Complete playback commits the full assistant text.
- Partial barge-in commits only the played prefix and records an interruption note.
- Committed context requires earlier matching `tts_chunk_played` evidence for the same voice turn and chunk; the full-chunk hash must match and committed length cannot exceed played length.
- Committed-only, commit-before-played, mismatched chunk/hash, and over-length events leave committed context empty.
- Center hang-up and TopBar phone icon both preserve the current session and return to text mode.
- Mode-switch settlement sends played + committed context without a fake `barge_in_detected`, and ignores late playback completion.
- No-playback state leaves committed context empty.
- The next follow-up prompt includes committed assistant context and interruption note, but never includes an unplayed assistant draft.
- The SQL repository replay gate proves persisted event ordering reaches the same committed-context builder.
