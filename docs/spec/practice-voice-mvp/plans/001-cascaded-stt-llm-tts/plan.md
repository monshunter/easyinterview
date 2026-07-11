# Cascaded STT LLM TTS Voice MVP

> **版本**: 1.14
> **状态**: active
> **更新日期**: 2026-07-11

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

在现有 `PracticeScreen` 会话骨架内落地 P0 电话模式：前端用单一电话图标进入/退出，电话 Surface 以红色圆形挂断按钮回到同一 session 文本模式，不再提供重新开始；底层后端继续执行 `STT -> LLM -> TTS` 级联编排，按 session language 和统一上下文生成真实追问，前端自动推进 VAD/TTS turn 并只在真实 speech-start 时上报打断事件。

## 2 背景

S2S / realtime voice 成本高且 provider 形态差异大；面试训练的 P0 价值在追问质量、证据化 transcript、报告和复练闭环。级联方案允许先用豆包 STT、DeepSeek chat、豆包或 MiniMax TTS 组合完成可用电话模式体验，同时保持 realtime S2S 后续可插拔。

本计划依赖 A3 `004-cascaded-speech-provider-foundation` 提供 `tts` capability 与 speech adapters。若 A3 004 未完成，本计划可先以 fixture/stub 进行 UI/contract 开发，但不得宣称真实 provider 闭环。

2026-05-17 L1 预检结论：A3 004 checklist 已 17/17 完成，当前代码已存在 `practice.voice.stt.default` / `practice.voice.tts.default` profile、`AIClient.Transcribe` / `AIClient.Synthesize`、豆包 STT/TTS 与 MiniMax TTS adapter。当前实现仍缺 `createPracticeVoiceTurn` OpenAPI operation、fixtures、generated artifacts、backend handler/service、正式 frontend phone surface、以及 `E2E.P0.007`-`E2E.P0.009` 场景资产。本计划必须直接承接 `backend/internal/api/practice/README.md` 中的 voice/audio route handoff，不再创建同主题 sibling plan。

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior + contract + frontend + backend + ai-orchestration`。
- **TDD 策略**: 通过 `/implement` -> `/tdd` 顺序执行。每个实现项必须有 OpenAPI/fixture/codegen gate、backend service/handler tests、frontend component/controller tests、privacy/negative tests 或 BDD-Gate 作为断言来源。
- **BDD 策略**: 需要 BDD。用户可见电话模式流程、打断恢复和 provider failure fallback 分别由 `E2E.P0.007`、`E2E.P0.008`、`E2E.P0.009` 覆盖。
- **替代验证 gate**: 不适用；本计划是用户行为功能计划。
- **Review-fix runtime gate**: BUG-0070 后续要求 voice playback 证据覆盖 response `audioRef` 浏览器可播放、persisted session event 不保存 audio data、barge-in 前 partial `tts_chunk_played`、store replay committed context into next prompt；证据命令：`go test ./internal/practice ./internal/store/practice -count=1` + `pnpm --dir frontend test src/app/screens/practice/__tests__/practiceVoiceTurn.test.tsx --run`。
- **Review-fix fixture gate**: BUG-0072 后续要求 `createPracticeVoiceTurn` HTTP fixture 与真实 service audioRef 语义一致；fixture/default response 的 `ttsChunks[].audioRef` 必须为浏览器可播放 `data:audio/...;base64,...` 或同计划 resolver URL，禁止 `fixture-audio://...` 这类 mock-only scheme 进入 generated fixture client。
- **Review-fix lint precision gate**: 2026-05-22 后续要求 backend-practice out-of-scope lint 继续禁止独立 `/voice` route / alias，但必须允许本计划拥有的 `POST /practice/sessions/{sessionId}/voice-turns`、`createPracticeVoiceTurn`、`practice.voice.stt.default` / `practice.voice.tts.default` profile 与 `practice.voice.stt` / `practice.voice.tts` feature key；证据命令：`python3 -m pytest scripts/lint/backend_practice_out_of_scope_test.py -q` + `make lint-backend-practice-out-of-scope` + `make lint`。
- **Real-interview phone gate**: 用户可见 UI / docs / scenarios 使用 `电话模式 / Phone`；Top Bar 只保留单一电话图标，电话 surface 只提供通话状态、字幕和红色圆形挂断图标，不提供分段切换、live chip、切断文字、重新开始、`callEnded`、语音分析、手动转写、开始录音或提交本轮。底层 `createPracticeVoiceTurn`、`practice.voice.*` profile 和 persisted `voice-turn://` ref 仍允许作为工程能力名。

## 4 Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `createPracticeVoiceTurn` | `openapi/fixtures/PracticeSessions/createPracticeVoiceTurn.json` scenarios `default` / `stt-config-missing` / `chat-failed` / `chat-output-invalid` / `tts-failed` | `PracticeScreen` phone controller / audio capture hook；VAD silence submit + TTS-ended rearm | `backend/internal/practice` voice turn handler/service mounted by `backend/internal/api/practice` | success stores session-event metadata and opaque `voice-turn://...` refs；HTTP response `ttsChunks[].audioRef` must be browser-playable data URL or documented resolver；double-invalid chat returns top-level `AI_OUTPUT_INVALID` before result/TTS persistence | `practice.voice.stt.default` + `practice.followup.default` + `practice.voice.tts.default`；canonical current-turn/transcript/committed context + persisted session language + exactly one structured/language repair | `E2E.P0.007` / `E2E.P0.009` + BUG-0070 audioRef gate |
| `appendSessionEvent` | `openapi/fixtures/PracticeSessions/appendSessionEvent.json` scenarios `voice-tts-started` / `voice-tts-played` / `voice-barge-in` / `voice-context-committed` | phone player progress reporter；hang-up commits heard prefix without barge-in；speech-start reports partial playback then barge-in | existing `appendSessionEvent` handler/service extended with voice event kinds | session events；store replay loads latest voice turn + subsequent playback events into next prompt | none; records playback, real speech-start barge-in, and committed context events | `E2E.P0.008` + BUG-0070 store replay gate |

## 5 Coverage Matrix

| Row | Category | Source | Phase | Verification | Negative scope |
|-----|----------|--------|-------|--------------|----------------|
| PV-MVP-C1 | Primary path | spec C-1 | Phase 2-4 | backend/frontend tests + `E2E.P0.007` | provider key in frontend |
| PV-MVP-C2 | Cross-layer contract | docs/development §2 | Phase 1 | OpenAPI fixture validation + `make codegen-check` | ad hoc fetch shape |
| PV-MVP-C3 | Alternate path | spec C-2 | Phase 2 | A3 profile resolver tests + backend orchestration tests | STT/TTS bound to same provider |
| PV-MVP-C4 | Failure/recovery | spec C-4/C-5 | Phase 2-4 | service tests + `E2E.P0.009` | TTS failure drops text |
| PV-MVP-C5 | Boundary condition | spec C-3 | Phase 3 | committed context unit tests + store replay tests + frontend partial playback event test + `E2E.P0.008` | unplayed draft in prompt |
| PV-MVP-C6 | Privacy/security/observability | spec C-7 | Phase 2/5 | privacy grep + backend tests + persisted audioRef summary gate | raw audio/transcript/TTS text in log/DB/metric/session event summary |
| PV-MVP-C7 | UX quality | docs/ui-design/module-practice-review | Phase 7 | frontend tests + visual parity gates | independent voice route/page, user-visible Voice copy, segmented/live/restart/callEnded controls, voice analysis panel |
| PV-MVP-C8 | Regression/out-of-scope-negative | product-scope D-6 | Phase 5 | scope tests + negative search | `voice` route alias, S2S marked active |
| PV-MVP-C9 | Current drift preflight | current code truth source | Phase 0 | source grep + focused smoke tests | phone surface still relies on a coming-soon card; backend README points to an out-of-scope voice-surface owner |
| PV-MVP-C10 | Phone lifecycle | spec C-8/C-9 | Phase 7 | frontend controller tests + `E2E.P0.007` / `E2E.P0.008` + parity/browser gate | restart/callEnded survives; hang-up emits barge-in; TTS resumes after exit |
| PV-MVP-C11 | Conversational integrity | spec C-10 | Phase 7 | backend service tests + `E2E.P0.009` | mixed-language or canned follow-up after repair failure |

## 6 实施步骤

### Phase 0: Current-state preflight and handoff lock

#### 0.1 A3 handoff and owner boundary

Confirm A3 004 outputs are available in code before changing the voice MVP contract: `tts` shared capability, `AIClient.Synthesize`, `AIClient.Transcribe`, `practice.voice.stt.default`, `practice.voice.tts.default`, provider-specific speech adapters, profile coverage lint, and privacy observability tests. If any of these regress, stop and repair A3 before continuing.

#### 0.2 Existing negative tests to invert

Record the current implementation gaps that must change during this plan: `frontend/src/app/screens/practice/components/VoiceSurfaceComingSoon.tsx` and its tests currently assert that real voice surface DOM is absent; `backend/internal/api/practice/README.md` must name this plan as the voice/audio route owner. This plan is the current owner for that handoff and must update those code/docs surfaces while preserving the product rule that there is no independent `voice` route.

### Phase 1: Contract and fixture

#### 1.1 OpenAPI voice turn contract

新增 `POST /practice/sessions/{sessionId}/voice-turns` / `createPracticeVoiceTurn` operation。该 endpoint 是 side-effect endpoint，必须携带 `Idempotency-Key`；请求体必须包含 `clientVoiceTurnId`、`turnId`、`audio.contentBase64`、`audio.contentType`、`audio.durationMs`、`language` 和 `practiceMode`，不接受手动转写替代字段。输出必须包含 `voiceTurnId`、`userTranscriptFinal`、`assistantTextDraft`、`ttsChunks[]`、`providerMetaSummary`、`session` 与可空 `ttsError`。HTTP response 的 `ttsChunks[].audioRef` 必须是浏览器可播放 data URL 或已落地 resolver URL；持久化 session event summary 只保存 chunk id、content type、duration、byte length/hash 和 opaque `voice-turn://...` ref，不保存音频数据。

P0 producer 必须保持每次 response `ttsChunks` 长度为 0 或 1；multi-chunk 需要先新增 per-chunk assistant text span / committed-context contract、OpenAPI 与 BDD，不能由当前前端自行推断文本分段。

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

实现 committed context builder：完整播放 chunk 可提交；barge-in 前已上报 `playedTextLength` 的部分播放文本只能按长度截断提交；无 `tts_chunk_played` 证据的 partial chunk 默认不提交；未播放 draft 永不进入下一轮 prompt。生产 service 必须从已持久化的最新 voice turn 与后续 playback events 加载 committed context，而不是只依赖请求体。

#### 3.3 Prompt interruption note

下一轮 prompt 必须显式说明上一条助理回复被打断、用户实际听到哪些内容、未播放内容不得视为已告知用户。

### Phase 4: Frontend voice controller

> Historical implementation phase. Phase 7 supersedes its former restart/call-ended UI anchors; this section is retained only to explain the implemented voice foundation.

#### 4.1 Phone UI source parity

在 `PracticeScreen` 内复刻 `ui-design/src/screen-practice.jsx` 电话 Surface：通话状态、字幕、切断、重新开始和结束并生成报告入口。不得新增独立 `voice` route，也不得把 out-of-scope `voice` query 作为电话模式入口。

必须删除或反转当前 `VoiceSurfaceComingSoon` coming-soon 语义：phone mode 不再展示 coming-soon 卡片，且 `practice-phone-surface`、`practice-phone-waveform`、`practice-phone-captions-toggle`、`practice-phone-hangup`、`practice-phone-restart` 等 DOM 锚点必须进入正式前端 parity gate。保留 `voice` route fallback 到 `home` 与 out-of-scope `mode=voice` query 被过滤的负向测试。

#### 4.2 Audio capture and STT submission

实现音频采集状态、提交 voice turn、展示 STT final transcript / 字幕、电话错误恢复和切换文本面试路径；不得提供手动转写替代字段或入口。测试使用 fixtures/stub，不打真实 provider。

#### 4.3 TTS playback and barge-in

实现 TTS chunk 播放、播放完成回报、VAD/用户输入触发 barge-in、停止播放和下一轮输入。barge-in 必须先发送 partial `tts_chunk_played`（含 `playedTextLength` / `playedTextHash` / `playbackOffsetMs`），再发送 `barge_in_detected`。误触发阈值和 echo 处理在 MVP 中以可配置策略实现。

### Phase 5: Verification and negative gates

#### 5.1 BDD scenarios

创建并执行 `E2E.P0.007`、`E2E.P0.008`、`E2E.P0.009` 场景资产，覆盖完整语音 turn、打断上下文提交和 provider failure fallback。

场景资产必须同时更新 `test/scenarios/e2e/INDEX.md`，并遵守 `test/scenarios/README.md` wrapper contract：`trigger.sh` 将真实 runner 输出写入 `.test-output/`，`verify.sh` 检查 runner marker、目标测试路径和 pass marker，拒绝只检查文件存在的 false-green。

#### 5.2 Regression gates

重跑 app shell / practice 相关 frontend tests、OpenAPI fixture validation、codegen drift、A3 profile coverage、privacy grep、out-of-scope route negative search。BUG-0070 后续 gate 必须额外验证 response `audioRef` 可播放、stored TTS summary 不含 audio data、store replay committed context、barge-in partial playback event。

### Phase 6: Real-interview phone-mode simplification

> Historical implementation phase. Phase 7 supersedes the former restart/call-ended control contract and is the only current UI acceptance source.

#### 6.1 Phone-mode UI language and controls

将用户可见面试形式固定为 `电话模式 / Phone`，删除 `开始录音` / `提交本轮` 主按钮，改为真实电话控制：接通状态、切断、重新开始、显示字幕。历史 `voice` API/profile 命名仅作为底层能力保留。

#### 6.2 Keep phone surface focused on current controls

电话 surface 只保留通话状态、字幕、切断和重新开始；用户可见层不提供语速、停顿、口头禅、音量等语音分析，也不提供手动转写或 speech-to-text 替代入口。STT 失败只能走电话错误恢复、切断/重开或切换文本面试。

#### 6.3 Contract and scenario negative gates

更新 OpenAPI 描述、fixtures、generated client/server tests、scenario README/expected outcome，证明 phone mode 只暴露当前电话控件，同时底层 voice turn privacy、audioRef 和 committed-context gates 继续通过。

### Phase 7: Phone lifecycle and conversational integrity

#### 7.1 Single handset entry and shared exit

Revise the UI truth source and formal frontend so the Top Bar uses one accessible handset icon. Text mode click enters phone mode; phone mode click and the center red circular hang-up button call the same `exitPhoneMode` coordinator. Exit immediately stops microphone/TTS, permits non-empty capture settlement, suppresses all later phone TTS, and navigates to text mode for the same session. Remove restart, `callEnded`, segmented mode controls, live chip and the visible “切断” label; delete any frontend/backend restart action, state, event or handler residue, and prove zero positive restart surface without adding an HTTP schema.

#### 7.2 Automatic turn progression and interruption truth

VAD silence submits a non-empty capture automatically; TTS ended re-arms listening. Only real speech-start during TTS emits partial `tts_chunk_played` followed by `barge_in_detected`. Hang-up may commit the heard prefix through existing playback/commit events but must not emit barge-in.

#### 7.3 Contextual, language-consistent follow-up

The voice chat step must use the shared canonical question renderer and persisted session language. Only invalid structured output, business parsing failure or question-language mismatch receives one repair attempt; provider/config/timeout errors do not trigger business repair. A second invalid result returns the existing top-level `AI_OUTPUT_INVALID` error envelope before `PracticeVoiceTurnResult`/TTS persistence, leaves the session row unchanged, and keeps text-mode exit available.

#### 7.4 Current evidence

Update existing `E2E.P0.007`-`E2E.P0.009` assertions, focused frontend/backend tests, source/parity gates and a real-mode browser proof. Do not add sibling scenarios or expand the HTTP schema.

## 7 验收标准

- `createPracticeVoiceTurn` contract、fixtures、generated client/server artifacts 完成并无 drift。
- 后端 voice turn service 覆盖 STT/chat/TTS happy path 与独立失败路径。
- 前端 phone controller 可按需展示字幕、自动推进 VAD/TTS turn、处理真实 speech-start barge-in，并通过单一电话图标或圆形挂断回到同一 session 文本模式；用户可见层不提供重新开始或手动转写替代入口。
- HTTP response 的 TTS `audioRef` 可被浏览器播放，持久化 session event summary 不保存 audio bytes。
- Committed context builder + store replay 证明已播放文本进入下一轮 prompt，未播放 assistant draft 不进入下一轮 prompt。
- `E2E.P0.007` / `E2E.P0.008` / `E2E.P0.009` BDD-Gate 通过。
- Voice follow-up 使用 canonical current-turn/transcript/committed context + persisted session language；parser/language invalid 只 repair 一次，provider/config/timeout 不 repair；第二次 invalid 返回顶层 `AI_OUTPUT_INVALID`，session 行不变且无 result/canned question/TTS。

## 8 风险与应对

| 风险 | 应对措施 |
|------|----------|
| A3 004 未完成阻塞真实 provider | 前端/contract 可用 fixture 开发；真实 provider smoke 不得标 PASS |
| 打断误触发 | MVP 使用持续语音阈值 + 播放起始保护窗；误触发仍只影响当前播放，不污染上下文 |
| TTS 延迟影响面试节奏 | 前端先展示文本回复，TTS 可异步播放；TTS failure 保持文本继续 |
| 隐私边界不清 | transcript 作为业务 session event 明确建模，AI metadata 只存摘要 |

## 9 修订记录

| 日期 | 版本 | 说明 |
|------|------|------|
| 2026-07-11 | 1.14 | Reopen Phase 7 for single-handset phone exit, no restart/callEnded, VAD/TTS auto progression, speech-start-only barge-in, and language-consistent follow-up with one repair. |
| 2026-07-10 | 1.13 | Align regression gate route negative search wording to out-of-scope terminology without behavior changes. |
| 2026-07-10 | 1.12 | Align `voice` route/query negative input wording to out-of-scope terminology without behavior changes. |
| 2026-07-10 | 1.11 | Replace remaining VoiceSurfaceComingSoon wording with coming-soon/unavailable-surface terminology. |
| 2026-07-10 | 1.10 | Align remaining user-visible modality wording with phone mode in the active plan. |
| 2026-07-10 | 1.9 | Align active plan parity targets to current phone route/query and `practice-phone-*` anchors; out-of-scope `voice` query is negative input only. |
| 2026-07-10 | 1.8 | Remove the old manual transcription request field from the active voice turn contract and align Phase 6 evidence gates. |
| 2026-07-10 | 1.7 | Tighten Phase 6 wording around the current phone surface contract and remove old surface labels from active gates. |
| 2026-07-09 | 1.6 | Reopen for real-interview phone-mode simplification: user-visible voice becomes phone mode, old analysis and manual transcription substitute surfaces are removed, and phone hang-up/restart/captions become the target UI. |
