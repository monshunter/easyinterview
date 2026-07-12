# Backend Practice Spec

> **版本**: 1.22
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
| `createPracticePlan` | `POST /practice/plans`，要求 `Idempotency-Key`；请求不包含 `questionBudget`、`mode` 或 hint 配置 | `backend/internal/api/practice.CreatePracticePlan` + `backend/internal/practice.CreatePracticePlan` | `practice_plans`, `idempotency_records`, `audit_events` | none | `E2E.P0.022`, `E2E.P0.070`, `E2E.P0.072` |
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
- Prompt 输入包含 persisted session language、plan goal、targetJob/resume/round context、能力重点和按序消息历史；不得包含题号、题目预算、generation kind、question intent 或 follow-up count。
- 输出保持最小结构化 envelope `{messageText}`，只用于 schema 与语言校验；它不是题目结构。
- schema / business parse 或语言不一致时只 repair 一次；provider/config/secret/timeout/unsupported 不做 business repair。
- 二次 invalid output 返回 typed `AI_OUTPUT_INVALID`，不生成 canned message。

### 2.4 消息幂等与恢复

`sendPracticeMessage` 使用三段式边界：

1. 短事务按 `(session_id, client_message_id)` reserve 或 replay user message，并拒绝前一条 user message 尚无 assistant reply 时的并发新消息。
2. 事务外执行 `practice.session.chat`。
3. 短事务写 assistant reply；`reply_to_message_id` 唯一约束防止重复生成结果落库。

AI 失败时 user message 保留；同一 `clientMessageId` 重试不得重复写 user message，成功后返回唯一 assistant reply。前端在请求期间禁用输入，并在失败时保留原消息和重试动作。

### 2.5 生命周期事件

`practice_session_events` 只保存生命周期事实：

- `session_started`
- `session_completed`

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
| C-5 | 生命周期 | session running | 本地暂停、恢复、完成 | 暂停只影响前端输入；完成写生命周期事实并进入 report job | 002 |
| C-6 | 无专用提示 | session running | 用户需要提示 | 通过普通聊天表达，不存在 hint event/action/count | 003 |
| C-7 | 语音禁用 | 任意 session | 点击电话按钮或调用 voice endpoint | 前端不可点击；后端 fail-closed 且零 provider/持久化副作用 | practice-voice-mvp/001 |
| C-8 | 隐私 | 完成聊天与报告 | 检查非正文存储面 | raw messages 不泄漏 | 001/002/003 |

## 5 关联计划

- [001-plan-and-session-orchestration](./plans/001-plan-and-session-orchestration/plan.md)
- [002-event-loop-and-completion](./plans/002-event-loop-and-completion/plan.md)
- [003-mode-policies-and-provenance](./plans/003-mode-policies-and-provenance/plan.md)

## 6 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-12 | 1.22 | 按奥卡姆剃刀将 Practice 从题目/turn 状态机重构为连续 message conversation；删除专用 hint/mode/question 合同并暂时禁用语音。 |
