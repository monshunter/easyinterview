# Backend Practice Spec

> **版本**: 1.27
> **状态**: completed
> **更新日期**: 2026-07-12

## 1 背景与目标

`backend-practice` 承接基于 JD、简历和面试轮次上下文的连续模拟面试会话。当前 P0 不再维护预设题数、当前题、追问/下一题分类或逐题状态机；AI 与用户只通过按序消息推进对话，用户主动结束后进入会话级报告生成。

当前目标：

- 会话启动时生成一条普通 assistant opening message。
- 用户消息与 assistant 回复写入统一 `practice_messages`，不标记“题目 / 回答 / 追问”。
- `baseline / retry_current_round / next_round` 只决定上下文来源与能力重点，不决定题目集合。
- 暂时禁用电话模式：语音 endpoint 必须 fail-closed，且不得调用 STT / chat / TTS provider 或写入语音事件。
- 会话结束后只生成会话级准备度、能力维度、证据和下一步建议。

## 2 当前合同

### 2.1 Operation Matrix

| operationId | HTTP 行为 | backend handler | persistence | AI / job dependency | scenario coverage |
|-------------|-----------|-----------------|-------------|---------------------|-------------------|
| `createPracticePlan` | `POST /practice/plans`，要求 `Idempotency-Key`；请求不包含 `questionBudget`、`mode` 或 hint 配置，可携带 `roundId` 轮次意图 | `backend/internal/api/practice.CreatePracticePlan` + `backend/internal/practice.CreatePracticePlan` | `practice_plans.round_id/round_sequence`, `idempotency_records`, `audit_events` | none；服务端从 TargetJob summary 推导 sequence 并验证当前轮 | `E2E.P0.022`, `E2E.P0.070`, `E2E.P0.072` |
| `getPracticePlan` | `GET /practice/plans/{planId}`，用户隔离读取 | `backend/internal/api/practice.GetPracticePlan` | `practice_plans` | none | `E2E.P0.022`, `E2E.P0.070` |
| `listPracticeSessions` | `GET /practice/sessions`，按 cursor / targetJob / status 列表 | `backend/internal/api/practice.ListPracticeSessions` | `practice_sessions` | none | workspace / report owner gates |
| `startPracticeSession` | `POST /practice/sessions`，要求 `Idempotency-Key`；同步返回 session 与 opening assistant message | `backend/internal/api/practice.StartPracticeSession` + `backend/internal/practice.StartPracticeSession` | `practice_sessions`, `practice_messages`, lifecycle event, outbox, `idempotency_records`, `ai_task_runs` | `practice.session.chat` + `AIClient.Complete` | `E2E.P0.023`-`E2E.P0.026` |
| `getPracticeSession` | `GET /practice/sessions/{sessionId}`，返回 session 与有序 messages | `backend/internal/api/practice.GetPracticeSession` | `practice_sessions`, `practice_messages` | none | `E2E.P0.023`, `E2E.P0.025`, `E2E.P0.044` |
| `sendPracticeMessage` | `POST /practice/sessions/{sessionId}/messages`；body `clientMessageId` 幂等，返回 user/assistant message pair | `backend/internal/api/practice.SendPracticeMessage` + `backend/internal/practice.SendPracticeMessage` | `practice_messages`, lifecycle events, `ai_task_runs` | `practice.session.chat`; recent ordered messages + plan/session context | `E2E.P0.044`, `E2E.P0.046` |
| `completePracticeSession` | `POST /practice/sessions/{sessionId}/complete`，要求 `Idempotency-Key`，返回 `202 ReportWithJob` | existing practice handler/service/store | `practice_sessions`, `feedback_reports`, `async_jobs`, outbox, `idempotency_records` | `report_generate` queued job | `E2E.P0.047`, `E2E.P0.056` |
| `createPracticeVoiceTurn` | 当前禁用；任何请求 fail-closed 为 typed `AI_UNSUPPORTED_CAPABILITY`，不得读取音频后调用 provider | existing voice handler boundary | none | none while disabled | `E2E.P0.007` |

### 2.2 数据模型

`practice_plans` 只保存会话所需的稳定上下文：

- `target_job_id`
- `resume_id`
- `source_report_id`
- `goal`
- `interviewer_persona`
- `difficulty`
- `language`
- `time_budget_minutes`
- `round_id`
- `round_sequence`
- `focus_competency_codes`
- `status`

以下字段删除：`question_budget`、`mode`。专用 hint / strict-assistance 语义不再存在。

`practice_sessions` 保存生命周期、语言和关联 ID，不保存 `turn_count`、`current_turn` 或 hint 配置。

`practice_messages` 是对话唯一正文真理源：

| 字段 | 约束 |
|------|------|
| `id` | UUID 主键 |
| `session_id` | session FK，级联删除 |
| `seq_no` | session 内单调递增，唯一 |
| `role` | `user / assistant` |
| `content` | 用户可见消息正文 |
| `client_message_id` | user message 必填；session 内唯一，作为 replay key |
| `reply_to_message_id` | assistant message 指向 user message；唯一，保证每条用户消息最多一个回复 |
| `created_at` | server timestamp |

不存在 `PracticeTurn`、`question_text`、`question_intent`、`answer_text`、`follow_up_count`、题目状态或题目编号。

### 2.3 连续聊天 AI 边界

- 唯一 feature key 为 `practice.session.chat`。空历史生成 opening message，非空历史生成下一条普通 assistant message。
- Prompt 输入包含 persisted session language、plan goal、targetJob/resume/round context、能力重点和按序消息历史；简历事实源按 `parsed_text_snapshot → original_text → structured_profile` 选择第一份非空完整内容，业务代码不得做字符/token 截断；不得包含题号、题目预算、generation kind、question intent 或 follow-up count。
- Prompt 必须把不可被业务数据覆盖的 `system` 策略与不可信业务上下文分层：JD、完整简历、轮次、persona 和历史对话只能作为 JSON 编码后的 user data 传入，任何其中出现的指令式文本都不得进入或改写 system policy。JSON 编码必须覆盖引号、换行、标签和类似 prompt-injection 的内容，避免上下文逃逸。完整历史可用于保持对话连续性，但只有 persisted resume 与 candidate-authored `user` message 能建立候选人事实；`assistant` message 永远不能证明履历事实或把上一轮模型臆造放大为下一轮依据。
- `round context` 必须由 plan 的 exact `roundId/roundSequence` 回查 TargetJob summary 得到 round name/type/focus；不得把 `interviewerPersona` 当作 interview round，也不得只把 persona/generalist 文案写进 `{{interview_round}}`。start/send 必须使用同一 round 投影；新 session 对 legacy null/mismatch identity fail closed。
- `interviewerPersona` 只控制语气、视角和追问风格，不能创造候选人事实、替代真实轮次、改变完成台账或决定当前/下一轮。
- start/send 两条 store reservation 必须使用相同的 resume context 投影，并再次验证 `practice_plans.resume_id = target_jobs.resume_id`；TargetJob 改绑后，旧 plan/session 即使仍属同一用户也必须 fail closed，不能继续使用旧简历。若三种简历内容均为空，必须在 prompt resolve / AI 调用前 fail closed 为 typed `VALIDATION_FAILED`，不得注入 `resume context unavailable` 后继续生成，也不得写 assistant message。
- 面试官只能把 persisted Resume context 或 candidate-authored `user` message 中明确出现的公司、项目、产品和技术栈当作候选人事实；`assistant` history 只用于连续性，不是事实来源。不得声称简历包含实际不存在的项目，也不得继续追问仅由上一条 assistant 引入的项目。用户只说“几个项目”却未给出名称时，应先请用户命名或描述，不能自行补造项目。
- 输出保持最小结构化 envelope `{messageText}`，只用于 schema 与语言校验；它不是题目结构。
- `finish_reason=length` 即使返回内容碰巧是合法 JSON 也视为截断的 `AI_OUTPUT_INVALID`；schema / business parse、语言不一致或 length 截断只 repair 一次，provider/config/secret/timeout/unsupported 不做 business repair。
- 二次 invalid output 返回 typed `AI_OUTPUT_INVALID`，不生成 canned message。

### 2.4 消息幂等与恢复

`sendPracticeMessage` 使用三段式边界：

1. 短事务按 `(session_id, client_message_id)` reserve 或 replay user message，并拒绝前一条 user message 尚无 assistant reply 时的并发新消息。
2. 事务外执行 `practice.session.chat`。
3. 短事务写 assistant reply；`reply_to_message_id` 唯一约束防止重复生成结果落库。提交时必须再次锁定并校验 session 仍处于 `running / waiting_user_input`；若完成请求已把 session 推进到 `completing / completed`，则回滚迟到的 assistant reply 并返回 typed conflict，不得把 session 改回 `running`。

AI 失败时 user message 保留；同一 `clientMessageId` 重试不得重复写 user message，成功后返回唯一 assistant reply。前端在请求期间禁用输入和结束 CTA，并在失败时保留原消息和重试动作。

### 2.5 生命周期事件

`practice_session_events` 只保存生命周期事实：

- `session_started`
- `session_completed`

`session_completed` 在完成事务内与 `practice_sessions.completed_at/status='completing'`、report/job/outbox 一起提交，是轮次已完成的唯一台账事实。报告生成的 queued/generating/ready/failed 不得反向改变这项完成事实；同一 session 的完成重放不得重复贡献轮次进度。

消息正文只存在于 `practice_messages`，不得复制进 event、outbox、audit、log、metric 或 task-run payload。

### 2.6 电话模式禁用

- `practice.voice.stt.default`、`practice.voice.tts.default` 和 realtime profile 保持 disabled / unsupported。
- `createPracticeVoiceTurn` 在读取/解析业务音频前 fail-closed；禁止调用 STT、chat、TTS，禁止写 voice event / committed context / audio metadata。
- 正式前端只展示不可点击的置灰电话图标，不产生 `phone` route state。
- 通用 speech adapter 可保留；重新启用必须重新修订 Product/UI/OpenAPI owner 与真实 provider gate。

### 2.7 隐私与隔离

- 所有读取和写入按 `user_id` 隔离；跨用户访问返回 404。
- message content 只进入授权的 session/read/report 输入，不进入 URL、localStorage、outbox、audit、log、metric label、task-run payload 或 provider metadata。
- `DELETE /me` 通过 FK cascade 删除 plans、sessions、messages、events 和 reports。

### 2.8 轮次身份、创建校验与进度事实

- canonical round 只来自带完整非空 provenance 的 `TargetJob.summary.interviewRounds[]`：`promptVersion`、`rubricVersion`、`modelId`、`language`、`dataSourceVersion` 均不得为空。每个 sequence 必须是正 `int32`，按数组顺序严格递增且唯一，但允许 `1,2,4` 这类非连续目录；type 必须保持小写并属于 `hr / technical / manager / cross_functional / culture / final / other`。服务端按 sequence 派生 `round-{sequence}-{type}`，不得用 TargetJob lifecycle `status`、数组第一项、固定轮次表或时长碰撞猜测当前轮。
- baseline plan 必须选择完成台账投影出的第一个未完成轮次；`retry_current_round` 必须复用 source report 所属 plan 的 exact round pair；`next_round` 必须选择 source round 之后 canonical 列表中的紧邻项，并且该项仍是当前第一个未完成轮次。sequence 不连续时选择下一条已存在 canonical round，不硬编码 `sourceSequence + 1`。
- 请求 `roundId` 仅用于校验客户端意图；`roundSequence` 由服务端推导。请求 `resumeId` 必须等于当前用户 TargetJob 持久化绑定的 `resume_id`；新 plan 必须成对持久化 round identity，time budget 必须等于该 round 的 `durationMinutes`。即使另一份简历也属于同一用户，也不得用于该 TargetJob 创建 plan。
- baseline 已完成台账、retry/next 的 source report/session/plan 事实都必须来自同一 TargetJob 绑定 resume；wrong-resume completion、source plan 或 ready plan 必须被忽略或 fail closed，不得贡献完成前缀、授权派生计划或成为当前 plan。
- session start 与后续 message prompt 必须从 plan pair 加载同一真实 round 的 name/type/focus，并在每次 reservation 时重新校验 plan resume 仍等于 TargetJob 当前绑定 resume；persona 是独立的面试官风格字段，不是轮次身份。
- legacy null round identity 不参与当前 plan 复用或完成投影；只有 migration/backfill owner 证明唯一匹配后才能补齐。
- 除主题/外观偏好外，轮次、计划、完成事实和当前进度均为后端业务数据。前端内存、URL、local/session storage 或 fixture 不得充当事实源。

## 3 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| API contract | `openapi-v1-contract` | message schemas、fixtures、generated client/server |
| Persistence | `db-migrations-baseline` | `practice_messages` 与 session/report schema |
| Prompt | `prompt-rubric-registry` | `practice.session.chat` schema/hash/eval |
| Frontend | `frontend-workspace-and-practice` | 单聊天窗口、消息发送、本地暂停、结束、语音 disabled UI |
| Report | `backend-review` / `frontend-report-dashboard` | 会话级维度、证据、建议和复练 |
| Voice | `practice-voice-mvp` | disabled fail-closed 边界与 generic provider 保留范围 |

## 4 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 创建 baseline plan | 有 JD 和 resume | 创建 plan | 不读写 question/mode/hint 字段 | 001 |
| C-2 | 启动会话 | plan ready | start session | 返回 opening message，无 currentTurn | 001 |
| C-3 | 连续聊天 | session running | 用户发送多条消息 | 每次只追加 user/assistant message pair，无题目分类 | 002 |
| C-4 | 幂等恢复 | assistant generation 首次失败 | 同 clientMessageId 重试 | user message 不重复，最终只有一个 assistant reply | 002 |
| C-5 | 生命周期 | session running，可能存在进行中的 AI reply | 本地暂停、恢复、完成 | 暂停只影响前端输入；完成写生命周期事实并进入 report job；迟到 reply 不得重新打开 session | 002 |
| C-6 | 无专用提示 | session running | 用户需要提示 | 通过普通聊天表达，不存在 hint event/action/count | 003 |
| C-7 | 语音禁用 | 任意 session | 点击电话按钮或调用 voice endpoint | 前端不可点击；后端 fail-closed 且零 provider/持久化副作用 | practice-voice-mvp/001 |
| C-8 | 隐私 | 完成聊天与报告 | 检查非正文存储面 | raw messages 不泄漏 | 001/002/003 |
| C-9 | 真实简历 grounding | resume 的 `structured_profile` 为空或 parse failed，但存在完整 `parsed_text_snapshot` / `original_text` | 启动会话并继续发送消息 | start/send prompt 均包含完整 source snapshot 尾部 marker；面试官不得发明简历未出现的项目；所有简历内容都为空时返回 `VALIDATION_FAILED` 且零 AI / assistant message | 001/002 |
| C-10 | 轮次持久化与完成事实 | TargetJob 有结构化轮次，可能相邻轮次时长相同、sequence 不连续、完成请求重放或存在 legacy plan | 创建 baseline/retry/next plan 并完成 session | 新 plan 持久化 exact `roundId/roundSequence`；请求轮次/预算不匹配、全轮完成和非法 source fail closed；`session_completed` 每个 round pair 去重贡献进度，report 状态不影响完成事实 | 001/002 |
| C-11 | 绑定、目录与 prompt 信任边界 | TargetJob 绑定 resume A；同用户另有 resume B；summary 可能缺 provenance、含大于 int32 / 非连续 sequence 或大小写错误 type；简历/JD/历史可能含指令式文本，assistant history 可能已臆造项目 | 创建 plan 并启动/继续会话 | 只有绑定 resume A 的 source/completion/ready-plan 事实有效；`1,2,4` 目录按 canonical successor 推进，溢出/非法 type/缺 provenance fail closed；system policy 不被 JSON 编码的不可信上下文覆盖，persona 只影响风格，assistant-only claim 不成为候选人事实 | 001/002 |

## 5 关联计划

- [001-plan-and-session-orchestration](./plans/001-plan-and-session-orchestration/plan.md)
- [002-event-loop-and-completion](./plans/002-event-loop-and-completion/plan.md)
- [003-mode-policies-and-provenance](./plans/003-mode-policies-and-provenance/plan.md)

## 6 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-12 | 1.27 | Restrict candidate facts to the persisted resume and candidate-authored user messages; assistant history remains continuity-only and cannot amplify model-invented projects. |
| 2026-07-12 | 1.26 | Tighten TargetJob-resume binding, canonical round provenance/type/int32 constraints, non-contiguous successor semantics, and the system-policy versus JSON-encoded untrusted-context prompt boundary. |
| 2026-07-12 | 1.25 | Add normalized practice-plan round identity, server-side current/next validation, and session-completion ledger semantics for backend progress projection. |
| 2026-07-12 | 1.24 | Reopen 001/002 so start and send use the complete resume source snapshot, fail closed without evidence, and forbid invented resume projects. |
| 2026-07-12 | 1.23 | 重新打开 002：关闭 message commit 与 completion 的竞态，并强化 P0.046/P0.047 的失败恢复与生命周期证据。 |
| 2026-07-12 | 1.22 | 按奥卡姆剃刀将 Practice 从题目/turn 状态机重构为连续 message conversation；删除专用 hint/mode/question 合同并暂时禁用语音。 |
