# Practice Voice MVP Spec

> **版本**: 1.6
> **状态**: active
> **更新日期**: 2026-07-07

## 1 背景与目标

`practice-voice-mvp` 定义 P0 语音面试的用户可见落地方案。用户已明确选择用低成本 `STT -> LLM -> TTS` 级联方案替代 S2S / realtime voice 首发，同时要求支持多 provider，并确认 `stt` 与 `tts` 是独立 profile / provider 配置，不绑定同一家供应商。

本 spec 的目标是让用户在现有 `PracticeScreen` 内以 `practice?mode=voice&modality=voice` 进入语音面试，完成一轮或多轮：

1. 用户语音经 STT 转写为文本。
2. 文本回答进入现有 chat profile，由 LLM 生成追问 / 反馈 / 下一题。
3. LLM 文本回复经 TTS 合成并播放。
4. 系统记录 transcript、已播放 assistant chunk、打断事件和可用于报告的 committed context。

该 MVP 不把级联语音伪装成 realtime S2S。`realtime` capability 继续代表 S2S / realtime multimodal voice，并保持 fail-closed。

## 2 范围

### 2.1 In Scope

- `PracticeScreen` 内的语音面试形态，入口仍为 `practice?mode=voice&modality=voice` 或等价显式参数。
- 后端 voice turn 编排：`stt profile -> chat profile -> tts profile`。
- STT / chat / TTS 三类 profile 独立选择：默认建议 STT 豆包、chat DeepSeek、TTS 豆包，MiniMax `speech-02-turbo` 作为 TTS 备选或 fallback。
- AI assistant 回复 draft 与 committed context 分离：未播放内容不得进入下一轮 prompt。
- 用户插话 / 打断：停止播放、取消未完成 TTS、提交已播放且有 `playedTextLength` 证据的 assistant 文本范围、丢弃未播放 draft。
- API / fixture / generated client / backend handler / frontend consumer / scenario coverage operation matrix。
- 隐私与观测：raw audio、TTS audio、transcript 明文、provider secret 不进入 log / DB metadata / metric label。

### 2.2 Out of Scope

- S2S / realtime voice provider 接入；`practice.voice.realtime.default` 继续 fail-closed。
- 独立 `voice` route 或独立语音页面骨架。
- 长期音频留存、对象存储生命周期、24h 删除链路和用户级录音保留开关的完整生产化；本计划只保留 session 内必要播放 / 转写状态。
- 复盘语音添加流程；它后续可复用 A3 `stt` / `tts` 底座，但不由本 MVP 实现。
- Provider 质量评测和供应商选型最终结论；MVP 只锁接口和默认建议。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 语音形态 | P0 语音面试采用 `stt -> chat -> tts` 级联方案，不采用 S2S 首发 | 成本可控，工程可拆分 |
| D-2 | Provider 配置 | STT、chat、TTS 三者独立 profile，不绑定同一家 provider | 可用豆包 STT + DeepSeek chat + 豆包或 MiniMax TTS |
| D-3 | Realtime 边界 | `realtime` 只表示 S2S / realtime multimodal voice，MVP 不打开 | 防止语义混淆和成本误判 |
| D-4 | 打断提交 | 完整播放 chunk 或 barge-in 前已上报 `playedTextLength` 的部分播放文本可进入 committed context；未播放 draft 不得进入下一轮 prompt | 用户已听到的内容可延续上下文，未播放内容不会污染下一轮 prompt |
| D-5 | 路由边界 | 语音面试只能通过 `practice` 显式参数进入，不恢复 `voice` route | 保持 UI 真理源一致 |
| D-6 | TTS 失败语义 | TTS 失败只影响语音播放，已生成文本仍展示并可继续文本面试 | 防止语音 provider 故障导致会话丢失 |

### 3.2 待确认事项

- 首版 TTS 是否默认启用 MiniMax fallback，还是只作为可选 profile 由配置切换。
- 首版是否保存任何临时音频对象；默认不落长期存储，只保留播放和测试所需的 session 内临时状态。

## 4 设计约束

### 4.1 UI / 路由约束

- UI 必须复刻 `ui-design/src/screen-practice.jsx` 的语音 Surface 和 `docs/ui-design/module-practice-review.md` 的语音边界。
- 文本面试输入框里的麦克风仍表示“语音转文字”，不是切换到语音面试。
- 不得新增或恢复 `voice` route / route alias / 独立语音页。
- 严格模拟模式下应隐藏语音现场提示，但不影响必要的听说状态和错误恢复控件。

### 4.2 业务状态约束

Voice turn 必须区分：

- `user_transcript_final`：STT 完成后的用户回答文本。
- `assistant_text_draft`：LLM 完整或增量生成的文本回复，尚未全部播放。
- `tts_chunk_started` / `tts_chunk_played`：TTS chunk 播放状态。
- `assistant_context_committed`：已播放并允许进入下一轮 prompt 的 assistant 文本；partial playback 只能按 `playedTextLength` 截断提交。
- `barge_in_detected`：用户插话 / 打断事件，记录打断时刻和已播放 chunk。

下一轮 prompt 只能读取 committed user messages、committed assistant messages 与当前用户输入；service 必须从已持久化的 `follow_up_generated` 与后续 playback events 回放 committed context，不得读取未播放的 `assistant_text_draft`。

### 4.3 API / Contract 约束

新增或修订 API 必须遵守 `docs/development.md` §2 的 Frontend / Backend Contract Workflow。本计划锁定以下 P0 HTTP 边界：

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `createPracticeVoiceTurn` | `openapi/fixtures/PracticeSessions/createPracticeVoiceTurn.json` scenarios `default` / `stt-config-missing` / `chat-failed` / `tts-failed` | `PracticeScreen` voice turn controller | `backend/internal/practice` voice turn handler/service mounted by `backend/internal/api/practice` | session events + transient in-memory audio metadata only; no long-term audio retention by default | `practice.voice.stt.default` + `practice.followup.default` + `practice.voice.tts.default` | `E2E.P0.007` / `E2E.P0.009` |
| `appendSessionEvent` | `openapi/fixtures/PracticeSessions/appendSessionEvent.json` extended with `voice-tts-started` / `voice-tts-played` / `voice-barge-in` / `voice-context-committed` | voice player progress reporter | existing `appendSessionEvent` handler/service extended with voice event kinds | session events | none, records playback/interrupt events | `E2E.P0.008` |

前端不得直连豆包或 MiniMax provider，也不得持有 provider key。

`createPracticeVoiceTurn` 是会产生会话事件的 side-effect endpoint，必须携带 `Idempotency-Key`。请求体必须显式携带 `clientVoiceTurnId`、`turnId`、`audio.contentBase64`、`audio.contentType`、`audio.durationMs`、`language`、`practiceMode` 与可选 `manualTranscriptFallback`；不允许把 raw audio 写入 URL、日志、AI metadata 或 audit metadata。响应体必须区分 `userTranscriptFinal`、`assistantTextDraft`、`ttsChunks[]`、`voiceTurnId`、`providerMetaSummary` 与可空 `ttsError`。`ttsChunks[]` 只包含 chunk id、content type、duration、byte length/hash 与 `audioRef`；`audioRef` 的播放承载与持久化边界见下段。

`ttsChunks[].audioRef` 的 HTTP response 值必须是浏览器可直接播放的 `data:audio/...;base64,...` 或同计划落地的 resolver URL；持久化到 `practice_session_events` 的 voice turn summary 必须改写为不含音频数据的 opaque `voice-turn://{voiceTurnId}/chunks/{chunkId}` 引用，并由测试证明 response playback ref 与 persisted summary ref 分离。

Fixture-backed mock responses must follow the same HTTP response playback semantics: `createPracticeVoiceTurn` fixtures may use `data:audio/...;base64,...` or a checked-in resolver URL, but must not use mock-only schemes such as `fixture-audio://...` because those cannot be consumed by browser playback paths.

`appendSessionEvent.kind` 必须扩展为 `tts_chunk_started`、`tts_chunk_played`、`barge_in_detected`、`assistant_context_committed`。这些事件继续使用 body-level `clientEventId`，不得携带 `Idempotency-Key`。payload 必须携带 `voiceTurnId`、`chunkId`、`playedTextHash` / `playedTextLength`、`playbackOffsetMs`、`occurredAt` 等摘要字段；如需提交业务正文，只能写入 session event schema 中明确允许的 committed assistant text，不得写入 AI/audit metadata。

### 4.4 隐私 / 安全 / 观测约束

- raw audio、TTS audio、transcript 明文、LLM prompt/response 明文不得进入 log / DB metadata / metric label。
- `audit_events` 与 `ai_task_runs.metadata` 只允许 hash、长度、duration、content type、profile、provider、model、cost micros、error code。
- Provider secret 只由 A4 SecretSource 注入；前端永不接触 provider secret。
- `APP_ENV=test` 允许 stub/fixture；非测试本地 app run 或未来 staging / prod 选中真实 provider 时缺 secret 必须 fail-fast。

### 4.5 Failure / Recovery 约束

- STT 失败：保留已录入 / 已输入文本，提示用户重试或继续手动输入；不得调用 chat/TTS。
- Chat 失败：保留用户 transcript，允许重试或结束并生成部分报告；不得调用 TTS。
- TTS 失败：展示 assistant 文本，允许继续文本面试或重试语音播放；不得丢失会话。
- Barge-in：停止播放，取消未完成 TTS；前端必须先上报 partial `tts_chunk_played`（含 `playedTextLength` / `playedTextHash` / `playbackOffsetMs`）再上报 `barge_in_detected`；后端只提交已播放文本范围，未播放 draft 丢弃。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| A3 AI provider foundation | `ai-provider-and-model-routing/004` | `tts` capability、speech adapters、profile catalog、observability/privacy |
| OpenAPI contract | B2 + practice owner | `createPracticeVoiceTurn`、fixtures、generated Go/TS artifacts |
| Backend practice orchestration | future `backend-practice` / this plan target | voice turn handler/service、session event persistence、AIClient orchestration |
| Frontend practice UI | future `frontend-workspace-and-practice` / this plan target | PracticeScreen voice controller、audio capture/playback、barge-in event reporting |
| Scenario assets | scenarios owner + practice owner | E2E.P0.007/P0.008/P0.009 setup/trigger/verify/cleanup |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 完整语音 turn | 用户在 `practice?mode=voice&modality=voice` 进入语音面试，STT/chat/TTS profile 均 active | 用户说出回答并等待 AI 回复 | 页面展示用户 transcript、assistant 文本并播放 TTS；session event 记录 voice turn；可继续下一题 | 001 |
| C-2 | STT/TTS 独立 provider | STT profile 指向豆包，TTS profile 指向豆包或 MiniMax，chat profile 指向 DeepSeek | 后端执行 voice turn | 三个 profile 分别解析，任何一步不要求与另一能力同 provider；meta 可区分 provider/model/cost | 001 + A3 004 |
| C-3 | 打断不污染上下文 | AI TTS 播放中，前端已报告完整 chunk 或 partial `playedTextLength` | 用户插话 | 后端只提交已播放文本范围；未播放 assistant draft 不进入下一轮 prompt；下一轮 prompt 明确上一条回复被打断 | 001 |
| C-4 | TTS 失败降级 | STT 与 chat 成功，TTS provider 失败 | 用户等待回复 | 前端展示 assistant 文本与错误提示；用户可继续文本面试或重试播放；session 不失败 | 001 |
| C-5 | Secret fail-fast | 非测试本地 app run 或未来 staging/prod 选中 active speech profile 但缺 provider secret | 启动或调用 voice turn | 返回配置错误或启动失败；不得静默回退 stub | 001 + A3 004 |
| C-6 | UI route negative | 用户访问 non-current `voice` route input 或文档/代码出现独立 voice page 口径 | 路由归一或 scope test 执行 | 不进入独立 voice 页面；语音面试只能从 `practice` 显式参数进入 | 001 |
| C-7 | 隐私红线 | 任意 voice turn 完成或失败 | 查询 log / DB metadata / metric / audit | 不含 raw audio、TTS audio、transcript 明文、provider secret；只含 hash/长度/duration/profile/provider/cost 摘要 | 001 + A3 004 |

## 7 关联计划

- [001-cascaded-stt-llm-tts](./plans/001-cascaded-stt-llm-tts/plan.md)（completed）：落地用户可见语音面试 MVP 的 API、backend orchestration、frontend voice controller、barge-in committed context 与 BDD 场景。

## 8 相关文档

- [A3 AI Provider and Model Routing](../ai-provider-and-model-routing/spec.md)
- [Product Scope](../product-scope/spec.md)
- [docs/ui-design module-practice-review](../../ui-design/module-practice-review.md)
- [docs/development.md](../../development.md)
