# Practice Voice MVP History

> **版本**: 1.14
> **状态**: active
> **更新日期**: 2026-07-11

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-07-11 | 1.14 | 重新打开 001 的电话会话交互：删除重开与文字“切断”，用单一电话图标和圆形挂断返回同一 session 文本模式；补充 VAD/TTS 自动推进、真实 speech-start 打断及同语言追问单次 repair 合同。 | [001-cascaded-stt-llm-tts](./plans/001-cascaded-stt-llm-tts/plan.md) |
| 2026-07-10 | 1.13 | 将 001 regression gate 中的 route 负向搜索统一为 out-of-scope 口径；行为不变。 | tech-debt pruning |
| 2026-07-10 | 1.12 | 将 `voice` route/query 负向输入统一为 out-of-scope 口径；行为不变。 | tech-debt pruning |
| 2026-07-10 | 1.10 | 用户可见形态进一步收敛为电话模式；“语音面试”仅保留为底层 voice 能力历史语义，不作为当前产品表达。 | tech-debt pruning |
| 2026-07-10 | 1.9 | 电话模式正向入口收敛为 `practice?mode=phone&modality=phone`；out-of-scope `voice` route/query 只作为负向输入处理，不再归一为电话模式。 | tech-debt pruning |
| 2026-07-07 | 1.6 | 收敛 voice route negative gate 文案为 out-of-scope route input 表述，不改变 voice MVP 契约。 | product-scope/001 Phase 6.59 |
| 2026-05-22 | 1.5 | Repo lint follow-up：收窄 backend-practice out-of-scope lint 的 out-of-scope standalone voice route 匹配，继续拦截独立 `/voice` route / alias，但允许 `createPracticeVoiceTurn`、`/voice-turns`、`voice-turn://` persisted ref 与 `practice.voice.*` profile / feature key；focused lint test、`make lint-backend-practice-out-of-scope` 与 `make lint` 通过。 | [001-cascaded-stt-llm-tts](./plans/001-cascaded-stt-llm-tts/plan.md) |
| 2026-05-22 | 1.4 | 对齐 ADR-Q4 v1.7 / 方案 A：语音 provider secret fail-fast 边界从 dev/Kind/staging/prod 改为非测试本地 app run 与未来部署；Kind 不再是 P0 语音 smoke 默认前提。 | local-dev-stack/001 post-pass revision |
| 2026-05-17 | 1.2 | BUG-0070 follow-up：固化 voice playback runtime gates，要求 `ttsChunks[].audioRef` response 可被浏览器直接播放、持久化 event summary 使用不含音频数据的 `voice-turn://...` opaque ref、barge-in 前上报 partial `tts_chunk_played`，并从 store replay committed context 注入下一轮 prompt。 | [001-cascaded-stt-llm-tts](./plans/001-cascaded-stt-llm-tts/plan.md) |
| 2026-05-17 | 1.1 | 完成 plan 001：落地 `createPracticeVoiceTurn` contract / fixtures / generated artifacts、backend STT/chat/TTS orchestration、frontend voice surface/controller、playback events、BDD `E2E.P0.007-009` 与 lifecycle close-out。 | [001-cascaded-stt-llm-tts](./plans/001-cascaded-stt-llm-tts/plan.md) |
| 2026-05-08 | 1.0 | 初始创建：用户批准以 `stt -> chat -> tts` 级联方案作为语音面试 MVP，STT/TTS 独立 provider 配置，realtime S2S 继续 fail-closed。 | 001-cascaded-stt-llm-tts |
