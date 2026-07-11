# E2E.P0.008 Seed Input

- Active phone turn id: `voice-turn-1`
- Current TTS chunk: `chunk-1`
- Assistant draft text: `请继续说明高风险团队试点。`
- `playedTextHash` / `committedTextHash`: SHA-256 of the complete assistant draft for the single P0 chunk
- Playback events:
  - `tts_chunk_started` at offset `0ms`
  - `tts_chunk_played` for complete or partial text length
  - `barge_in_detected` at offset `1480ms`
  - `assistant_context_committed` after matching played evidence for the same chunk
- Frontend fixture:
  - fake `Audio` object starts playback
  - real VAD speech-start interrupts active playback
  - user clicks either the center hang-up control or the single TopBar phone icon while playback is active
  - append event calls are captured from generated client transport
