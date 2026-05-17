# Practice Voice MVP History

> **版本**: 1.2
> **状态**: active
> **更新日期**: 2026-05-08

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-17 | 1.2 | BUG-0070 follow-up：固化 voice playback runtime gates，要求 `ttsChunks[].audioRef` response 可被浏览器直接播放、持久化 event summary 使用不含音频数据的 `voice-turn://...` opaque ref、barge-in 前上报 partial `tts_chunk_played`，并从 store replay committed context 注入下一轮 prompt。 | [001-cascaded-stt-llm-tts](./plans/001-cascaded-stt-llm-tts/plan.md) |
| 2026-05-17 | 1.1 | 完成 plan 001：落地 `createPracticeVoiceTurn` contract / fixtures / generated artifacts、backend STT/chat/TTS orchestration、frontend voice surface/controller、playback events、BDD `E2E.P0.007-009` 与 lifecycle close-out。 | [001-cascaded-stt-llm-tts](./plans/001-cascaded-stt-llm-tts/plan.md) |
| 2026-05-08 | 1.0 | 初始创建：用户批准以 `stt -> chat -> tts` 级联方案作为语音面试 MVP，STT/TTS 独立 provider 配置，realtime S2S 继续 fail-closed。 | 001-cascaded-stt-llm-tts |
