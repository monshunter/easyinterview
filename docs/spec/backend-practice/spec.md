# Backend Practice Spec

> **版本**: 1.21
> **状态**: active
> **更新日期**: 2026-07-11

## 1 背景与目标

`backend-practice` 承接模拟面试后端域：用户基于 JD 与扁平简历创建练习计划，启动文本或电话模式面试会话，逐轮提交事件，在结束时创建报告生成 handoff。域边界以 [openapi-v1-contract](../openapi-v1-contract/spec.md)、[event-and-outbox-contract](../event-and-outbox-contract/spec.md)、[db-migrations-baseline](../db-migrations-baseline/spec.md)、[shared-conventions-codified](../shared-conventions-codified/spec.md)、[prompt-rubric-registry](../prompt-rubric-registry/spec.md) 与 [ai-provider-and-model-routing](../ai-provider-and-model-routing/spec.md) 为上游契约。

当前 P0 后端闭环只服务三类目标：`baseline`、`retry_current_round`、`next_round`。`baseline` 使用 `target_jobs` + 扁平 `resumes.structured_profile`；`retry_current_round` 与 `next_round` 使用 `sourceReportId` 关联既有报告。

## 2 当前合同

### 2.1 Operation Matrix

| operationId | HTTP 行为 | backend handler | persistence | AI / job dependency | scenario coverage |
|-------------|-----------|-----------------|-------------|---------------------|-------------------|
| `createPracticePlan` | `POST /practice/plans`，要求 `Idempotency-Key`，返回 `201 PracticePlan` | `backend/internal/api/practice.CreatePracticePlan` + `backend/internal/practice.CreatePracticePlan` | `practice_plans`, `idempotency_records`, `audit_events` | none | `E2E.P0.022`, `E2E.P0.025` |
| `getPracticePlan` | `GET /practice/plans/{planId}`，用户隔离读取 | `backend/internal/api/practice.GetPracticePlan` | `practice_plans` | none | `E2E.P0.022`, `E2E.P0.025` |
| `listPracticeSessions` | `GET /practice/sessions`，按 cursor / targetJob / status 列表 | `backend/internal/api/practice.ListPracticeSessions` | `practice_sessions` | none | workspace / report owner gates |
| `startPracticeSession` | `POST /practice/sessions`，要求 `Idempotency-Key`，同步返回首题 | `backend/internal/api/practice.StartPracticeSession` + `backend/internal/practice.StartPracticeSession` | `practice_sessions`, `practice_turns`, `practice_session_events`, `outbox_events`, `idempotency_records`, `ai_task_runs` | `practice.session.first_question` + `AIClient.Complete` | `E2E.P0.023`, `E2E.P0.024`, `E2E.P0.025`, `E2E.P0.026` |
| `getPracticeSession` | `GET /practice/sessions/{sessionId}`，用户隔离读取 | `backend/internal/api/practice.GetPracticeSession` | `practice_sessions`, `practice_turns` | none | `E2E.P0.023`, `E2E.P0.025` |
| `appendSessionEvent` | `POST /practice/sessions/{sessionId}/events`，body `clientEventId` 幂等，禁止 `Idempotency-Key` header；追问生成连续失败返回既有 `session_wait` action | `backend/internal/api/practice.AppendSessionEvent` + `backend/internal/practice.AppendSessionEvent` | `practice_session_events`, `practice_turns`, `outbox_events`, `ai_task_runs` | `practice.session.follow_up`, `practice.turn.lightweight_observe`；canonical context + session language + exactly one repair，禁止 canned question | `E2E.P0.038`, `E2E.P0.044`-`E2E.P0.047` |
| `completePracticeSession` | `POST /practice/sessions/{sessionId}/complete`，要求 `Idempotency-Key`，返回 `202 ReportWithJob` | `backend/internal/api/practice.CompletePracticeSession` + `backend/internal/practice.CompletePracticeSession` | `practice_sessions`, `feedback_reports`, `async_jobs`, `outbox_events`, `idempotency_records` | `report_generate` queued job | `E2E.P0.047`, `E2E.P0.056` |
| `createPracticeVoiceTurn` | `POST /practice/sessions/{sessionId}/voice-turns`，要求 `Idempotency-Key`；连续 invalid chat output 返回顶层 `AI_OUTPUT_INVALID` | `backend/internal/api/practice.CreatePracticeVoiceTurn` + `backend/internal/practice.CreatePracticeVoiceTurn` | success-only voice session event/committed context rows；double-invalid 时 session 行不变且无 result/TTS persistence | STT / `practice.session.follow_up` / TTS profiles owned with `practice-voice-mvp`；persisted session language + exactly one parser/language repair | `E2E.P0.007`, `E2E.P0.009` |

### 2.2 数据与状态

- `practice_plans` 绑定 `target_job_id`、`resume_id`、`goal`、`mode`、`source_report_id`、`focus_competency_codes` 与计划配置。
- `PracticeGoal` 只包含 `baseline`、`retry_current_round`、`next_round`。
- `PracticeMode` 保留 `assisted`、`strict` 作为持久化枚举；`strict` 不再阻断 `hint_requested`，提示按 in-session optional assistance 处理。
- `practice_sessions` 状态集合为 `queued`、`running`、`waiting_user_input`、`completing`、`completed`、`failed`、`cancelled`。
- `practice_session_events` 记录有序会话事件；`appendSessionEvent` 以 `(session_id, client_event_id)` 去重。
- `practice_turns` 保存题目、回答、hint 与 turn 状态；事件与 outbox payload 不携带题目或回答正文。
- `completePracticeSession` 在同一事务内创建 `feedback_reports(status='queued')` 与 `async_jobs(job_type='report_generate')`，`practice.session.completed` 仅作为 source event / analytics fact。
- `idempotency_records` 为副作用 operation 提供 user-scoped 幂等、pending 单执行者、成功 replay、mismatch conflict 与 TTL reset。

### 2.3 首题与 AI 边界

`startPracticeSession` 使用三段式流程：

1. 短事务 reserve session + idempotency record。
2. 事务外解析 `practice.session.first_question` prompt / rubric / model profile，并调用 `AIClient.Complete`。
3. 短事务写首个 `practice_turns`、`session_started` event、`practice.session.started` outbox 和成功 replay response。

首题 prompt context 必须包含语言、岗位标题、岗位技能、rubric dimensions、practice goal 与扁平 `resumes.structured_profile` 摘要。AI 输出必须是严格 JSON且 `questionText` 匹配 session language；JSON/schema/business parse 或语言不一致时只重试一次，provider/config/timeout 错误不重试。最终失败写入 `failed` session 与 B1 `AI_*` 错误码。

### 2.4 追问 / 下一题 AI 边界

- `appendSessionEvent(answer_submitted)` 与 `createPracticeVoiceTurn` 的 chat step 必须复用同一 canonical question renderer。最小上下文只使用当前 reservation 已可靠提供的服务端状态：persisted session language、plan goal/mode/targetJobId、当前 turn 的问题/intent/status/follow-up count、本次已提交回答或 transcript、`generation_kind=follow_up|next_question`，以及电话路径额外读取的 voice committed context；不得伪造目标标题、轮次、resume context 或全量历史 turn，client payload 也不得覆盖问题、`questionIntent`、追问次数或下一题。
- 所有用户可见 assistant question 必须遵循 session language。内部 `questionIntent` 只用于编排/观测，不得直接拼进用户文本。
- 所有用户可见 hint cue 也必须遵循 persisted session language；结构或语言不匹配统一按 hint invalid-output graceful degrade 为 `session_wait`，不得把错误语言 cue 写入 `practice_turns.hint_text` 或返回给前端。
- Structured output 首次 JSON/schema/business parse 失败或 `questionText` 与 normalized session language 不一致时，只允许用同一上下文和语言执行恰好一次 repair；provider/config/secret/timeout/unsupported/fallback-exhausted 错误不触发 business repair。第二次失败不得回退到 hard-coded English 或 canned question。
- 文本 `appendSessionEvent` 第二次失败返回现有 schema 内 `AssistantAction{type='session_wait'}`，恢复事件前的 turn 控制状态且不写 completion outbox；前端保留原回答并用新的 `clientEventId` 重试。电话 `createPracticeVoiceTurn` 第二次失败返回既有 typed voice-turn error，跳过 TTS，前端留在同一 session 并可切回文本。
- 现有 `practice.session.follow_up` feature key 负责 session 启动后的 `follow_up|next_question` 两种 generation kind，二者复用同一 structured output；marker、repair instruction 与语言约束必须只定义在 prompt truth source。模板变化必须同步 registry hash、baseline seed migration、resolved prompt snapshot 和 eval cases，Go 代码不得拼接自然语言 prompt。
- 本修订不新增 HTTP operation、event kind 或 response schema；实现只收敛 server-owned prompt/context/repair 和现有错误/AssistantAction 分支。

### 2.5 隐私与隔离

- 所有读取与写入按 `user_id` 过滤；跨用户访问返回 404，不泄露资源存在性。
- log、metric label、audit、outbox、async payload 不得包含 `question_text`、`answer_text`、`hint_text`、AI prompt body、AI response body 或 provider secret。
- `DELETE /me` 通过 FK cascade 清理 practice plans / sessions / events / turns；已发送的 outbox 事件按 B3 不回滚。

## 3 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| API contract | [openapi-v1-contract](../openapi-v1-contract/spec.md) | PracticePlans / PracticeSessions schema、fixtures、generated artifacts |
| Backend domain | `backend/internal/api/practice`, `backend/internal/practice`, `backend/internal/store/practice` | handler、service、store、state machine、outbox emit |
| DB schema | [db-migrations-baseline](../db-migrations-baseline/spec.md) | practice tables、`idempotency_records`、report/job handoff tables |
| Event / job contract | [event-and-outbox-contract](../event-and-outbox-contract/spec.md) | `practice.session.started`, `practice.turn.completed`, `practice.session.completed`, `report_generate` |
| Shared enums / errors | [shared-conventions-codified](../shared-conventions-codified/spec.md) | `PracticeGoal`, `PracticeMode`, `SessionStatus`, practice error codes |
| Prompt / AI | [prompt-rubric-registry](../prompt-rubric-registry/spec.md), [ai-provider-and-model-routing](../ai-provider-and-model-routing/spec.md) | feature_key resolution、model profile routing、AI task run observability |
| Voice | [practice-voice-mvp](../practice-voice-mvp/spec.md) | speech profile selection and voice committed-context behavior |
| Frontend consumers | [frontend-workspace-and-practice](../frontend-workspace-and-practice/spec.md), [frontend-report-dashboard](../frontend-report-dashboard/spec.md) | workspace start flow、practice screen、generating/report handoff |

## 4 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | baseline plan 创建 | 用户拥有 ready target job 与 ready flat resume | `createPracticePlan` with `goal='baseline'` | 201，写 `practice_plans(resume_id, source_report_id null)` 与 audit metadata | 001 |
| C-2 | report-derived plan 创建 | 用户拥有 ready report | `createPracticePlan` with `goal IN retry_current_round,next_round` + `sourceReportId` | 201，写 `source_report_id` 与派生 focus / persona 配置 | 004 |
| C-3 | session 启动 | ready plan + active prompt profile | `startPracticeSession` | 201，同步返回 `currentTurn`，首题 prompt 携带 flat resume profile context | 001 |
| C-4 | AI 失败恢复 | 首题 AI 返回可映射错误 | `startPracticeSession` | 失败 reservation 写 `failed` / `failure_code`；可重试错误允许同 key 重试 | 001 |
| C-5 | event loop | running session | `appendSessionEvent` | 写有序 event，按 server-owned session/turn state 生成下一步 AssistantAction；client question/intent 不可覆盖；追问使用 canonical context + session language，格式失败只 repair 一次，第二次返回 `session_wait`、恢复事件前 turn 控制状态且无 canned question | 002 / 003 |
| C-6 | hint policy | strict 或 assisted session | `hint_requested` | strict 与 assisted 都可返回 `show_hint`；cue 必须匹配 persisted session language，AI / parse / language 失败时 graceful degrade 为 `session_wait` | 003 |
| C-7 | complete handoff | running session | `completePracticeSession` | 202，创建 queued report、queued `report_generate` job 与 completed source event | 002 |
| C-8 | 幂等与隔离 | 重试、mismatch、跨用户或并发请求 | side-effect operations | replay / conflict / 404 / single-executor 行为符合 B1/B4 | 001 / 002 |
| C-9 | 隐私红线 | 任意 practice flow | 检查 log/metric/audit/outbox/async payload | 无题目、回答、hint、prompt、response 或 secret 明文 | 001 / 002 / 003 / voice owner |
| C-10 | 文档和索引 | spec 或 plan lifecycle 变化 | sync docs | Header、plans INDEX、docs/spec INDEX 与 context manifest 一致 | docs-only |
| C-11 | voice follow-up failure | running phone session，chat structured output 连续两次无效 | `createPracticeVoiceTurn` | 首次失败只 repair 一次；第二次返回既有 typed voice-turn error，不调用 TTS、不生成 canned question，session 保持可继续/可切文本 | 002 + practice-voice-mvp/001 |

## 5 关联计划

1. [001-plan-and-session-orchestration](./plans/001-plan-and-session-orchestration/plan.md)：baseline plan/session foundation、flat resume binding、first question context、idempotency foundation。
2. [002-event-loop-and-completion](./plans/002-event-loop-and-completion/plan.md)：text event loop、canonical follow-up context/language/repair、completion handoff、report/job creation。
3. [003-mode-policies-and-provenance](./plans/003-mode-policies-and-provenance/plan.md)：mode policy、hint/lightweight observe、AssistantAction provenance。
4. [004-report-derived-practice-plans](./plans/004-report-derived-practice-plans/plan.md)：`retry_current_round` / `next_round` report-derived plans。

## 6 相关文档

- [Product Scope](../product-scope/spec.md)
- [OpenAPI v1 Contract](../openapi-v1-contract/spec.md)
- [Event and Outbox Contract](../event-and-outbox-contract/spec.md)
- [DB Migrations Baseline](../db-migrations-baseline/spec.md)
- [Shared Conventions](../shared-conventions-codified/spec.md)
- [Prompt Rubric Registry](../prompt-rubric-registry/spec.md)
- [AI Provider and Model Routing](../ai-provider-and-model-routing/spec.md)
- [Frontend Workspace and Practice](../frontend-workspace-and-practice/spec.md)
- [Frontend Report Dashboard](../frontend-report-dashboard/spec.md)
- [docs/development.md §2 Frontend / Backend Contract Workflow](../../development.md)
