# E2E.P0.007 Expected Outcome

- Frontend phone mode calls `createPracticeVoiceTurn` with `Idempotency-Key` and renders transcript / assistant text / TTS playback status.
- Backend HTTP route `POST /api/v1/practice/sessions/{sessionId}/voice-turns` returns `200 PracticeVoiceTurnResult`.
- Idempotency replay returns the same `voiceTurnId` and does not append a duplicate session event.
- Provider meta summary includes independent STT, chat, and TTS profile names and provider labels.
- Stored session event contains transcript, assistant text, TTS chunk metadata, and provider meta summary, but not raw base64 audio or generated TTS bytes.
- Current turn status becomes `follow_up_requested`, and the session remains `running` so the user can continue the interview.
