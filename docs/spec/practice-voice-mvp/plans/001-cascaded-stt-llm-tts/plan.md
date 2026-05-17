# Cascaded STT LLM TTS Voice MVP

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-05-17

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

在现有 `PracticeScreen` 会话骨架内落地 P0 语音面试 MVP：前端采集用户音频，后端执行 `STT -> LLM -> TTS` 级联编排，前端播放 TTS 并上报播放 / 打断事件，后端只把已播放完成的 assistant chunk 提交到上下文。

## 2 背景

S2S / realtime voice 成本高且 provider 形态差异大；面试训练的 P0 价值在追问质量、证据化 transcript、报告和复练闭环。级联方案允许先用豆包 STT、DeepSeek chat、豆包或 MiniMax TTS 组合完成可用语音体验，同时保持 realtime S2S 后续可插拔。

本计划依赖 A3 `004-cascaded-speech-provider-foundation` 提供 `tts` capability 与 speech adapters。若 A3 004 未完成，本计划可先以 fixture/stub 进行 UI/contract 开发，但不得宣称真实 provider 闭环。

2026-05-17 L1 预检结论：A3 004 checklist 已 17/17 完成，当前代码已存在 `practice.voice.stt.default` / `practice.voice.tts.default` profile、`AIClient.Transcribe` / `AIClient.Synthesize`、豆包 STT/TTS 与 MiniMax TTS adapter。当前实现仍缺 `createPracticeVoiceTurn` OpenAPI operation、fixtures、generated artifacts、backend handler/service、正式 frontend voice surface、以及 `E2E.P0.007`-`E2E.P0.009` 场景资产。本计划必须直接承接 `backend/internal/api/practice/README.md` 中的 voice/audio route handoff，不再创建同主题 sibling plan。

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior + contract + frontend + backend + ai-orchestration`。
- **TDD 策略**: 通过 `/implement` -> `/tdd` 顺序执行。每个实现项必须有 OpenAPI/fixture/codegen gate、backend service/handler tests、frontend component/controller tests、privacy/negative tests 或 BDD-Gate 作为断言来源。
- **BDD 策略**: 需要 BDD。用户可见语音面试流程、打断恢复和 provider failure fallback 分别由 `E2E.P0.007`、`E2E.P0.008`、`E2E.P0.009` 覆盖。
- **替代验证 gate**: 不适用；本计划是用户行为功能计划。

## 4 Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `createPracticeVoiceTurn` | `openapi/fixtures/PracticeSessions/createPracticeVoiceTurn.json` scenarios `default` / `stt-config-missing` / `chat-failed` / `tts-failed` | `PracticeScreen` voice turn controller / audio capture hook | `backend/internal/practice` voice turn handler/service mounted by `backend/internal/api/practice` | session events；transient in-memory audio metadata only；no long-term audio retention by default | `practice.voice.stt.default` + `practice.followup.default` + `practice.voice.tts.default` | `E2E.P0.007` / `E2E.P0.009` |
| `appendSessionEvent` | `openapi/fixtures/PracticeSessions/appendSessionEvent.json` scenarios `voice-tts-started` / `voice-tts-played` / `voice-barge-in` / `voice-context-committed` | voice player progress reporter | existing `appendSessionEvent` handler/service extended with voice event kinds | session events | none; records playback, barge-in, and committed context events | `E2E.P0.008` |

## 5 Coverage Matrix

| Row | Category | Source | Phase | Verification | Negative scope |
|-----|----------|--------|-------|--------------|----------------|
| PV-MVP-C1 | Primary path | spec C-1 | Phase 2-4 | backend/frontend tests + `E2E.P0.007` | provider key in frontend |
| PV-MVP-C2 | Cross-layer contract | docs/development §2 | Phase 1 | OpenAPI fixture validation + `make codegen-check` | ad hoc fetch shape |
| PV-MVP-C3 | Alternate path | spec C-2 | Phase 2 | A3 profile resolver tests + backend orchestration tests | STT/TTS bound to same provider |
| PV-MVP-C4 | Failure/recovery | spec C-4/C-5 | Phase 2-4 | service tests + `E2E.P0.009` | TTS failure drops text |
| PV-MVP-C5 | Boundary condition | spec C-3 | Phase 3 | committed context unit tests + `E2E.P0.008` | unplayed draft in prompt |
| PV-MVP-C6 | Privacy/security/observability | spec C-7 | Phase 2/5 | privacy grep + backend tests | raw audio/transcript/TTS text in log/DB/metric |
| PV-MVP-C7 | UX quality | docs/ui-design/module-practice-review | Phase 4 | frontend tests + visual parity gates | independent voice route/page |
| PV-MVP-C8 | Regression/legacy-negative | product-scope D-6 | Phase 5 | scope tests + negative search | `voice` route alias, S2S marked active |
| PV-MVP-C9 | Current drift preflight | current code truth source | Phase 0 | source grep + focused smoke tests | `VoiceSurfaceComingSoon` remains active after voice MVP; backend README points to a legacy placeholder owner |

## 6 实施步骤

### Phase 0: Current-state preflight and handoff lock

#### 0.1 A3 handoff and owner boundary

Confirm A3 004 outputs are available in code before changing the voice MVP contract: `tts` shared capability, `AIClient.Synthesize`, `AIClient.Transcribe`, `practice.voice.stt.default`, `practice.voice.tts.default`, provider-specific speech adapters, profile coverage lint, and privacy observability tests. If any of these regress, stop and repair A3 before continuing.

#### 0.2 Existing negative tests to invert

Record the current implementation gaps that must change during this plan: `frontend/src/app/screens/practice/components/VoiceSurfaceComingSoon.tsx` and its tests currently assert that real voice surface DOM is absent; `backend/internal/api/practice/README.md` must name this plan as the voice/audio route owner. This plan is the current owner for that handoff and must update those code/docs surfaces while preserving the product rule that there is no independent `voice` route.

### Phase 1: Contract and fixture

#### 1.1 OpenAPI voice turn contract

新增 `POST /practice/sessions/{sessionId}/voice-turns` / `createPracticeVoiceTurn` operation。该 endpoint 是 side-effect endpoint，必须携带 `Idempotency-Key`；请求体必须包含 `clientVoiceTurnId`、`turnId`、`audio.contentBase64`、`audio.contentType`、`audio.durationMs`、`language`、`practiceMode` 与可选 `manualTranscriptFallback`。输出必须包含 `voiceTurnId`、`userTranscriptFinal`、`assistantTextDraft`、`ttsChunks[]`、`providerMetaSummary`、`session` 与可空 `ttsError`。`ttsChunks[]` 只保存 chunk id、content type、duration、byte length/hash 和播放引用或测试 fixture handle，不保存音频明文。

#### 1.2 Fixtures and generated clients

新增 `createPracticeVoiceTurn` fixture scenarios：`default`、`stt-config-missing`、`chat-failed`、`tts-failed`；扩展 `appendSessionEvent` fixture scenarios：`voice-tts-started`、`voice-tts-played`、`voice-barge-in`、`voice-context-committed`。运行 codegen 后，前端只能通过 generated client 和 fixture-backed transport 消费，不允许在 `PracticeScreen` 内手写 ad hoc fetch shape。

### Phase 2: Backend orchestration

#### 2.1 Voice turn service

实现 voice turn service：解析 STT profile、调用 `Transcribe`，把 transcript 交给 chat profile，调用 `Synthesize` 生成 TTS chunk metadata。STT/TTS/chat profile name 必须独立配置。

#### 2.2 Failure isolation

实现 STT failure、chat failure、TTS failure 的独立错误路径。TTS 失败返回 assistant text + TTS error；STT 失败不调用 chat/TTS；chat 失败不调用 TTS。

#### 2.3 Privacy and event persistence

session event 只保存必要 transcript / committed text / event摘要；AI/audit metadata 只写 hash、长度、duration、provider、profile、cost。若 transcript 属于业务记录正文，其持久化必须走明确 session event schema，而不是 AI metadata。

### Phase 3: Playback progress and barge-in context

#### 3.1 Playback event model

扩展或复用 `appendSessionEvent` 记录 `tts_chunk_started`、`tts_chunk_played`、`barge_in_detected`、`assistant_context_committed`。这些事件继续使用 body-level `clientEventId` 做 replay key，禁止携带 `Idempotency-Key`；payload 必须包含 `voiceTurnId`、`chunkId`、`playedTextHash` / `playedTextLength`、`playbackOffsetMs`、`occurredAt` 和必要的 committed assistant text schema。

#### 3.2 Committed context builder

实现 committed context builder：只有完整播放的 chunk 可提交；部分播放 chunk 默认不提交；未播放 draft 永不进入下一轮 prompt。

#### 3.3 Prompt interruption note

下一轮 prompt 必须显式说明上一条助理回复被打断、用户实际听到哪些内容、未播放内容不得视为已告知用户。

### Phase 4: Frontend voice controller

#### 4.1 Voice UI source parity

在 `PracticeScreen` 内复刻 `ui-design/src/screen-practice.jsx` 语音 Surface：live 状态、暂停、转写、AI 透明度、语音现场提示、结束并生成报告入口。不得新增独立 `voice` route。

必须删除或反转当前 `VoiceSurfaceComingSoon` placeholder 语义：voice mode 不再展示 coming-soon 卡片，且 `practice-voice-waveform`、`practice-voice-annotated-waveform`、`practice-voice-expression-panel` 等 DOM 锚点必须进入正式前端 parity gate。保留 `voice` route fallback 到 `home` 的负向测试。

#### 4.2 Audio capture and STT submission

实现音频采集状态、提交 voice turn、展示 STT partial/final transcript、错误恢复和手动输入 fallback。测试使用 fixtures/stub，不打真实 provider。

#### 4.3 TTS playback and barge-in

实现 TTS chunk 播放、播放完成回报、VAD/用户输入触发 barge-in、停止播放和下一轮输入。误触发阈值和 echo 处理在 MVP 中以可配置策略实现。

### Phase 5: Verification and negative gates

#### 5.1 BDD scenarios

创建并执行 `E2E.P0.007`、`E2E.P0.008`、`E2E.P0.009` 场景资产，覆盖完整语音 turn、打断上下文提交和 provider failure fallback。

场景资产必须同时更新 `test/scenarios/e2e/INDEX.md`，并遵守 `test/scenarios/README.md` wrapper contract：`trigger.sh` 将真实 runner 输出写入 `.test-output/`，`verify.sh` 检查 runner marker、目标测试路径和 pass marker，拒绝只检查文件存在的 false-green。

#### 5.2 Regression gates

重跑 app shell / practice 相关 frontend tests、OpenAPI fixture validation、codegen drift、A3 profile coverage、privacy grep、旧 route negative search。

## 7 验收标准

- `createPracticeVoiceTurn` contract、fixtures、generated client/server artifacts 完成并无 drift。
- 后端 voice turn service 覆盖 STT/chat/TTS happy path 与独立失败路径。
- 前端 voice controller 可展示 transcript、播放 TTS、处理 barge-in，并保留文本 fallback。
- Committed context builder 证明未播放 assistant draft 不进入下一轮 prompt。
- `E2E.P0.007` / `E2E.P0.008` / `E2E.P0.009` BDD-Gate 通过。

## 8 风险与应对

| 风险 | 应对措施 |
|------|----------|
| A3 004 未完成阻塞真实 provider | 前端/contract 可用 fixture 开发；真实 provider smoke 不得标 PASS |
| 打断误触发 | MVP 使用持续语音阈值 + 播放起始保护窗；误触发仍只影响当前播放，不污染上下文 |
| TTS 延迟影响面试节奏 | 前端先展示文本回复，TTS 可异步播放；TTS failure 保持文本继续 |
| 隐私边界不清 | transcript 作为业务 session event 明确建模，AI metadata 只存摘要 |
