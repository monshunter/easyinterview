# E2E.P0.009 Expected Outcome

- STT failure returns a shared AI error and short-circuits chat/TTS calls.
- Chat failure short-circuits TTS.
- A second wrong-language chat output returns top-level `AI_OUTPUT_INVALID`, skips TTS/persistence, and shows a localized same-session recovery path back to text mode.
- TTS failure returns `PracticeVoiceTurnResult` with transcript, assistant text, empty `ttsChunks`, and retryable `ttsError`.
- Frontend renders the assistant text and does not start TTS playback instead of blocking the interview.
- A3 rejects realtime profile use in STT path before provider invocation.
- A3 rejects a disabled TTS profile before provider invocation, and the wrapper requires that exact test to execute.
- Runtime source grep confirms voice turn code does not use realtime or stub shortcuts, and privacy markers do not leak into metadata or summaries.
