# Practice Voice MVP History

> **版本**: 1.6
> **状态**: active
> **更新日期**: 2026-07-07

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-07-07 | 1.6 | 收敛 voice route negative gate 文案为 non-current route input 表述，不改变 voice MVP 契约。 | product-scope/001 Phase 6.59 |
| 2026-05-22 | 1.5 | Repo lint follow-up：收窄 backend-practice non-current lint 的 non-current standalone voice route 匹配，继续拦截独立 `/voice` route / alias，但允许 `createPracticeVoiceTurn`、`/voice-turns`、`voice-turn://` persisted ref 与 `practice.voice.*` profile / feature key；focused lint test、`make lint-backend-practice-non-current` 与 `make lint` 通过。 | [001-cascaded-stt-llm-tts](./plans/001-cascaded-stt-llm-tts/plan.md) |
| 2026-05-22 | 1.4 | 对齐 ADR-Q4 v1.7 / 方案 A：语音 provider secret fail-fast 边界从 dev/Kind/staging/prod 改为非测试本地 app run 与未来部署；Kind 不再是 P0 语音 smoke 默认前提。 | local-dev-stack/001 post-pass revision |
| 2026-05-17 | 1.2 | BUG-0070 follow-up：固化 voice playback runtime gates，要求 `ttsChunks[].audioRef` response 可被浏览器直接播放、持久化 event summary 使用不含音频数据的 `voice-turn://...` opaque ref、barge-in 前上报 partial `tts_chunk_played`，并从 store replay committed context 注入下一轮 prompt。 | [001-cascaded-stt-llm-tts](./plans/001-cascaded-stt-llm-tts/plan.md) |
| 2026-05-17 | 1.1 | 完成 plan 001：落地 `createPracticeVoiceTurn` contract / fixtures / generated artifacts、backend STT/chat/TTS orchestration、frontend voice surface/controller、playback events、BDD `E2E.P0.007-009` 与 lifecycle close-out。 | [001-cascaded-stt-llm-tts](./plans/001-cascaded-stt-llm-tts/plan.md) |
| 2026-05-08 | 1.0 | 初始创建：用户批准以 `stt -> chat -> tts` 级联方案作为语音面试 MVP，STT/TTS 独立 provider 配置，realtime S2S 继续 fail-closed。 | 001-cascaded-stt-llm-tts |
