# E2E.P0.008 Seed Input

- Active voice turn: `voice-turn-1`
- Current TTS chunk: `chunk-1`
- Assistant draft text: `请继续说明高风险团队试点。`
- Playback events:
  - `tts_chunk_started` at offset `0ms`
  - `tts_chunk_played` for complete or partial text length
  - `barge_in_detected` at offset `1480ms`
  - `assistant_context_committed` for fully played chunk
- Frontend fixture:
  - fake `Audio` object starts playback
  - user clicks the voice record control while playback is active
  - append event calls are captured from generated client transport
