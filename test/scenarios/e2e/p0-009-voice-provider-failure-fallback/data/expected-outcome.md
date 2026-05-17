# E2E.P0.009 Expected Outcome

- STT failure returns a shared AI error and short-circuits chat/TTS calls.
- Chat failure short-circuits TTS.
- TTS failure returns `PracticeVoiceTurnResult` with transcript, assistant text, empty `ttsChunks`, and retryable `ttsError`.
- Frontend renders the assistant text and a fallback notice instead of blocking the interview.
- A3 rejects realtime profile use in STT path before provider invocation.
- Runtime source grep confirms voice turn code does not use realtime or stub shortcuts, and privacy markers do not leak into metadata or summaries.
