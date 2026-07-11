# Practice Voice MVP Spec

> **版本**: 1.14
> **状态**: active
> **更新日期**: 2026-07-11

## 1 背景与目标

`practice-voice-mvp` 定义 P0 电话模式的底层语音落地方案。用户可见产品形态统一称为 `电话模式 / Phone`；工程层仍可用 `voice` 命名承接 STT / TTS capability、provider profile、OpenAPI operation 和历史测试资产。用户已明确选择用低成本 `STT -> LLM -> TTS` 级联方案替代 S2S / realtime voice 首发，同时要求支持多 provider，并确认 `stt` 与 `tts` 是独立 profile / provider 配置，不绑定同一家供应商。

本 spec 的目标是让用户在现有 `PracticeScreen` 内进入电话模式，完成一轮或多轮：

1. 用户语音经 STT 转写为文本。
2. 文本回答进入现有 chat profile，由 LLM 生成追问 / 反馈 / 下一题。
3. LLM 文本回复经 TTS 合成并播放。
4. 系统记录 transcript、已播放 assistant chunk、打断事件和可用于报告的 committed context。

该 MVP 不把级联语音伪装成 realtime S2S。`realtime` capability 继续代表 S2S / realtime multimodal voice，并保持 fail-closed。

## 2 范围

### 2.1 In Scope

- `PracticeScreen` 内的电话模式形态；用户可见参数 / 文案写作 `phone`，正向入口只接受 `mode=phone` / `modality=phone`，out-of-scope `voice` query 不作为电话模式入口。
- 后端 voice turn 编排：`stt profile -> chat profile -> tts profile`。
- STT / chat / TTS 三类 profile 独立选择：默认建议 STT 豆包、chat DeepSeek、TTS 豆包，MiniMax `speech-02-turbo` 作为 TTS 备选或 fallback。
- AI assistant 回复 draft 与 committed context 分离：未播放内容不得进入下一轮 prompt。
- 电话退出 / 用户插话：Top Bar 单一电话图标和中间挂断按钮复用同一退出语义，立即停止麦克风与 TTS 并回到同一 session 的文本模式；只有真实 speech-start 才触发 barge-in，已播放且有 `playedTextLength` 证据的 assistant 文本可提交，未播放 draft 必须丢弃。
- 电话 turn 自动推进：VAD 静音自动提交非空采集，TTS 播放结束自动重新监听；退出后即使采集结算完成也不得继续电话 TTS。
- API / fixture / generated client / backend handler / frontend consumer / scenario coverage operation matrix。
- 隐私与观测：raw audio、TTS audio、transcript 明文、provider secret 不进入 log / DB metadata / metric label。

### 2.2 Out of Scope

- S2S / realtime voice provider 接入；`practice.voice.realtime.default` 继续 fail-closed。
- 独立 `voice` route 或独立语音页面骨架。
- 用户可见语音分析：语速、停顿、口头禅、音量和类似表达层指标不属于本 spec 输出。
- 文本面试中的 `语音转文字`、`插入转写`、麦克风转写或手动转写替代入口。
- 长期音频留存、对象存储生命周期、24h 删除链路和用户级录音保留开关的完整生产化；本计划只保留 session 内必要播放 / 转写状态。
- 复盘语音添加流程；它后续可复用 A3 `stt` / `tts` 底座，但不由本 MVP 实现。
- Provider 质量评测和供应商选型最终结论；MVP 只锁接口和默认建议。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 电话模式底层语音形态 | P0 电话模式采用 `stt -> chat -> tts` 级联方案，不采用 S2S 首发 | 成本可控，工程可拆分 |
| D-2 | Provider 配置 | STT、chat、TTS 三者独立 profile，不绑定同一家 provider | 可用豆包 STT + DeepSeek chat + 豆包或 MiniMax TTS |
| D-3 | Realtime 边界 | `realtime` 只表示 S2S / realtime multimodal voice，MVP 不打开 | 防止语义混淆和成本误判 |
| D-4 | 打断提交 | 完整播放 chunk 或 barge-in 前已上报 `playedTextLength` 的部分播放文本可进入 committed context；未播放 draft 不得进入下一轮 prompt | 用户已听到的内容可延续上下文，未播放内容不会污染下一轮 prompt |
| D-5 | 路由边界 | 电话模式只能通过 `practice` 显式 `mode=phone` / `modality=phone` 参数进入，不恢复 `voice` route 或 out-of-scope `voice` query 入口 | 保持 UI 真理源一致 |
| D-6 | TTS 失败语义 | TTS 失败只影响语音播放，已生成文本仍展示并可继续文本面试 | 防止语音 provider 故障导致会话丢失 |
| D-7 | 用户可见命名 | UI / docs / report 展示统一为 `电话模式 / Phone`；`voice` 只作为底层工程能力名 | 防止用户误解为调试型语音功能或独立语音页面 |
| D-8 | 真实电话交互 | Top Bar 只保留单一电话图标；电话态中间只保留红色圆形挂断图标与字幕，不提供“切断”文字、重新开始、`callEnded`、`开始录音` 或 `提交本轮` | 对齐真实电话面试心智，减少状态分叉 |
| D-9 | 无语音分析 | 不展示或生成语速、停顿、口头禅、音量等分析指标 | 删除低价值干扰项 |
| D-10 | 退出电话模式 | Top Bar 电话图标与中间挂断按钮复用 `exitPhoneMode`：立即停止麦克风/TTS，允许非空采集结算，然后回到同一 session 文本模式；退出后不再播放电话 TTS | 不把形式切换误当成结束或重开会话 |
| D-11 | 自动对话节奏 | VAD 静音自动提交，TTS 结束自动重新监听；仅真实 speech-start 触发 barge-in | 让级联语音更接近连续通话并防止挂断伪造打断事件 |
| D-12 | 追问生成失败 | 追问使用 canonical server-owned context 和 persisted session language；结构或语言校验失败只 repair 一次，第二次失败返回既有顶层 `AI_OUTPUT_INVALID` error envelope，不生成 `PracticeVoiceTurnResult`、canned question 或 TTS，session 行保持原状态 | 防止 mock 感、语言混杂和无依据兜底问题 |

### 3.2 待确认事项

- 首版 TTS 是否默认启用 MiniMax fallback，还是只作为可选 profile 由配置切换。
- 首版是否保存任何临时音频对象；默认不落长期存储，只保留播放和测试所需的 session 内临时状态。

## 4 设计约束

### 4.1 UI / 路由约束

- UI 必须复刻 `ui-design/src/screen-practice.jsx` 的电话模式 Surface 和 `docs/ui-design/module-practice-review.md` 的电话边界。
- 文本面试输入框不得出现“语音转文字”、插入转写或麦克风转写。
- 不得新增或恢复 `voice` route / route alias / 独立语音页。
- 不得展示语速、停顿、口头禅、音量等语音分析面板。
- 电话模式默认不展示文字；用户点击显示字幕时才展示同一会话的字幕层。
- Top Bar 不得使用文本/电话分段控件或额外 `live` chip，只保留单一电话图标；电话 Surface 不得显示“切断”文字、重新开始或 `callEnded` 状态。
- 文本态点击电话图标进入电话模式；电话态点击同一图标或中间红色圆形挂断按钮均调用共享 `exitPhoneMode` 语义并回到同一 session 文本模式。

### 4.2 业务状态约束

Voice turn 必须区分：

- `user_transcript_final`：STT 完成后的用户回答文本。
- `assistant_text_draft`：LLM 完整或增量生成的文本回复，尚未全部播放。
- `tts_chunk_started` / `tts_chunk_played`：TTS chunk 播放状态。
- `assistant_context_committed`：已播放并允许进入下一轮 prompt 的 assistant 文本；partial playback 只能按 `playedTextLength` 截断提交。
- `barge_in_detected`：只由真实 speech-start 触发的用户插话 / 打断事件，记录打断时刻和已播放 chunk；用户点击挂断或切回文本不得生成该事件。

下一轮 prompt 只能读取 committed user messages、committed assistant messages 与当前用户输入；service 必须从已持久化的 `follow_up_generated` 与后续 playback events 回放 committed context，不得读取未播放的 `assistant_text_draft`。

### 4.3 API / Contract 约束

新增或修订 API 必须遵守 `docs/development.md` §2 的 Frontend / Backend Contract Workflow。本计划锁定以下 P0 HTTP 边界：

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `createPracticeVoiceTurn` | `openapi/fixtures/PracticeSessions/createPracticeVoiceTurn.json` scenarios `default` / `stt-config-missing` / `chat-failed` / `chat-output-invalid` / `tts-failed` | `PracticeScreen` phone controller：VAD 静音自动提交，TTS 结束自动重新监听 | `backend/internal/practice` voice turn handler/service mounted by `backend/internal/api/practice` | session events + transient in-memory audio metadata only; no long-term audio retention by default | `practice.voice.stt.default` + `practice.followup.default` + `practice.voice.tts.default`；chat 输入使用统一会话上下文和 session language，格式错误只 repair 一次 | `E2E.P0.007` / `E2E.P0.009` |
| `appendSessionEvent` | `openapi/fixtures/PracticeSessions/appendSessionEvent.json` extended with `voice-tts-started` / `voice-tts-played` / `voice-barge-in` / `voice-context-committed` | phone player progress reporter；挂断可提交已听到范围但不发送 barge-in | existing `appendSessionEvent` handler/service extended with voice event kinds | session events | none, records playback/interrupt events；只有真实 speech-start 记录 barge-in | `E2E.P0.008` |

前端不得直连豆包或 MiniMax provider，也不得持有 provider key。

`createPracticeVoiceTurn` 是会产生会话事件的 side-effect endpoint，必须携带 `Idempotency-Key`。请求体必须显式携带 `clientVoiceTurnId`、`turnId`、`audio.contentBase64`、`audio.contentType`、`audio.durationMs`、`language` 与 `practiceMode`；不接受手动转写替代字段。不得把 raw audio 写入 URL、日志、AI metadata 或 audit metadata。响应体必须区分 `userTranscriptFinal`、`assistantTextDraft`、`ttsChunks[]`、`voiceTurnId`、`providerMetaSummary` 与可空 `ttsError`。`ttsChunks[]` 只包含 chunk id、content type、duration、byte length/hash 与 `audioRef`；`audioRef` 的播放承载与持久化边界见下段。

`ttsChunks[].audioRef` 的 HTTP response 值必须是浏览器可直接播放的 `data:audio/...;base64,...` 或同计划落地的 resolver URL；持久化到 `practice_session_events` 的 voice turn summary 必须改写为不含音频数据的 opaque `voice-turn://{voiceTurnId}/chunks/{chunkId}` 引用，并由测试证明 response playback ref 与 persisted summary ref 分离。

当前 P0 producer 的可执行不变量是每个 voice turn 返回 `0..1` 个 TTS chunk：TTS 不可用时为 0，成功时为 1。`ttsChunks[]` 的数组形状不代表已支持 multi-chunk 顺序播放；后续若扩展到多 chunk，必须先补充每段 assistant 文本跨度/哈希映射、partial playback committed-context 语义、OpenAPI 约束和对应 BDD，不得只让前端循环播放数组。

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
- Follow-up 结构或语言校验失败：使用同一 canonical renderer、当前服务端可用的 turn/transcript/committed context 和 persisted session language 重试恰好一次；第二次失败返回既有顶层 `AI_OUTPUT_INVALID` error envelope，不调用 TTS、不生成 `PracticeVoiceTurnResult` 或 canned question，session 行保持原状态，前端允许退出到同一 session 的文本模式。
- 挂断 / 切回文本：立即停止麦克风与 TTS，不发送 `barge_in_detected`；非空采集允许结算，但结算结果不得在退出后继续触发电话 TTS。

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
| C-1 | 完整电话 turn | 用户进入电话模式，STT/chat/TTS profile 均 active | 用户说出回答并等待 AI 回复 | 页面以电话模式展示通话状态并可按需显示字幕；session event 记录底层 voice turn；可继续下一题 | 001 |
| C-2 | STT/TTS 独立 provider | STT profile 指向豆包，TTS profile 指向豆包或 MiniMax，chat profile 指向 DeepSeek | 后端执行 voice turn | 三个 profile 分别解析，任何一步不要求与另一能力同 provider；meta 可区分 provider/model/cost | 001 + A3 004 |
| C-3 | 打断不污染上下文 | AI TTS 播放中，前端已报告完整 chunk 或 partial `playedTextLength` | 用户插话 | 后端只提交已播放文本范围；未播放 assistant draft 不进入下一轮 prompt；下一轮 prompt 明确上一条回复被打断 | 001 |
| C-4 | TTS 失败降级 | STT 与 chat 成功，TTS provider 失败 | 用户等待回复 | 前端可显示字幕错误并允许挂断回同一 session 文本模式或重试语音播放；session 不失败 | 001 |
| C-5 | Secret fail-fast | 非测试本地 app run 或未来 staging/prod 选中 active speech profile 但缺 provider secret | 启动或调用 voice turn | 返回配置错误或启动失败；不得静默回退 stub | 001 + A3 004 |
| C-6 | UI route negative | 用户访问 out-of-scope `voice` route/query input 或文档/代码出现独立 voice page 口径 | 路由归一或 scope test 执行 | 不进入独立 voice 页面；电话模式只能从 `practice` 显式 `phone` 参数进入 | 001 |
| C-7 | 隐私红线 | 任意 voice turn 完成或失败 | 查询 log / DB metadata / metric / audit | 不含 raw audio、TTS audio、transcript 明文、provider secret；只含 hash/长度/duration/profile/provider/cost 摘要 | 001 + A3 004 |
| C-8 | 单一电话切换 | 用户在文本或电话模式的同一 session | 点击 Top Bar 电话图标或电话 Surface 挂断按钮 | 文本态进入电话；电话态立即停麦克风/TTS并回到文本；无分段控件、live chip、切断文字、重新开始或 `callEnded`；挂断不发送 barge-in | 001 |
| C-9 | 自动电话 turn | 电话模式正在监听或播放 TTS | 用户说完后静音，或 TTS 播放结束 | VAD 自动提交非空回答；TTS 结束自动恢复监听；只有真实 speech-start 在播放中触发 partial playback + barge-in | 001 |
| C-10 | 同语言真实追问 | session language 已确定，AI 首次返回 parser/language invalid 或连续 invalid | 后端生成下一条电话追问 | 使用 canonical current-turn/transcript/committed context 与 persisted session language；只 repair 一次；第二次 invalid 返回顶层 `AI_OUTPUT_INVALID`，session 行不变且无 result/canned question/TTS，前端可回同一 session 文本模式 | 001 + backend-practice/002 |

## 7 关联计划

- [001-cascaded-stt-llm-tts](./plans/001-cascaded-stt-llm-tts/plan.md)（active）：落地用户可见电话模式 MVP 的 API、backend orchestration、frontend phone controller、barge-in committed context 与 BDD 场景；Phase 7 当前负责单一电话切换、无重开退出、VAD/TTS 自动推进和同语言真实追问。

## 8 相关文档

- [A3 AI Provider and Model Routing](../ai-provider-and-model-routing/spec.md)
- [Product Scope](../product-scope/spec.md)
- [docs/ui-design module-practice-review](../../ui-design/module-practice-review.md)
- [docs/development.md](../../development.md)
