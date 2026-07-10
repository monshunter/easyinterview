# E2E.P0.009 Seed Input

- `createPracticeVoiceTurn` fixture scenarios:
  - `stt-config-missing`
  - `chat-failed`
  - `tts-failed`
- Backend service failure fixtures:
  - STT returns `AI_PROVIDER_SECRET_MISSING`
  - Chat returns `AI_FALLBACK_EXHAUSTED`
  - TTS returns retryable `AI_PROVIDER_TIMEOUT`
- A3 profile fixtures:
  - `practice.voice.stt.default`
  - `practice.voice.tts.default`
  - `practice.voice.realtime.default` remains unsupported
- Frontend fixture:
  - `createPracticeVoiceTurn=tts-failed`
  - assistant text remains visible and no TTS playback starts after TTS failure
