# E2E.P0.007 Expected Outcome

- Frontend phone mode uses VAD to call `createPracticeVoiceTurn` with `Idempotency-Key`, renders transcript / assistant text / TTS playback status, and rearms listening after playback.
- The active call obtains one microphone stream, releases it on hang-up, and exposes neither manual submit nor restart/call-ended controls.
- Backend HTTP route `POST /api/v1/practice/sessions/{sessionId}/voice-turns` returns `200 PracticeVoiceTurnResult`.
- Idempotency replay returns the same `voiceTurnId` and does not append a duplicate session event.
- Provider meta summary includes independent STT, chat, and TTS profile names and provider labels.
- Current P0 producer returns exactly one playable TTS chunk on success and zero on TTS fallback; multi-chunk is not claimed by this scenario.
- Stored session event contains transcript, assistant text, TTS chunk metadata, and provider meta summary, but not raw base64 audio or generated TTS bytes.
- The generated same-language follow-up is persisted in `session.currentTurn.questionText/questionIntent`; current turn status becomes `follow_up_requested`, and the session remains `running` so the next phone or text answer uses the returned server turn state.
