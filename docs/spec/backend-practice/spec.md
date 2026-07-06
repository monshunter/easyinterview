# Backend Practice Spec

> **版本**: 1.14
> **状态**: active
> **更新日期**: 2026-07-06

## 1 背景与目标

`backend-practice` 承接 [engineering-roadmap §5.2](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 中 `Mock Interview + Practice` workstream 的后端域，落地 P0 用户路径：用户在模拟面试规划页点击 `立即面试` 后，后端创建 `practice_plans`、启动 `practice_sessions`、维护 turn-by-turn 的问答与提示状态、在用户点击 `结束并生成报告` 时通过 outbox 触发异步报告生成，让前端 [Interview Session](../../ui-design/module-practice-review.md) 与 [Report Dashboard](../../ui-design/report-dashboard.md) 形成完整闭环。

[B2 OpenAPI v1](../openapi-v1-contract/spec.md) 已冻结本域承接的 6 个 operation：`createPracticePlan` / `getPracticePlan` / `startPracticeSession` / `getPracticeSession` / `appendSessionEvent` / `completePracticeSession`；[B3 `event-and-outbox-contract`](../event-and-outbox-contract/spec.md) 已冻结 `practice.session.started` / `practice.turn.completed` / `practice.session.completed` 三个事件，以及 `report_generate` 异步 job 的 `triggerEvent: practice.session.completed`；[B4 `db-migrations-baseline`](../db-migrations-baseline/spec.md) 已落地 `practice_plans` / `practice_sessions` / `practice_session_events` / `practice_turns`、shared `idempotency_records` 与外键到 `feedback_reports` 的反向 FK；[B1 `shared-conventions-codified`](../shared-conventions-codified/spec.md) 已锁定 `PracticeGoal` / `InterviewerRole` / `SessionStatus` 枚举、通用 AI 错误码与 `PRACTICE_SESSION_CONFLICT`；[F3 `prompt-rubric-registry`](../prompt-rubric-registry/spec.md) 已为本域分配 `practice.session.first_question` / `practice.session.follow_up` / `practice.turn.lightweight_observe` 三个 feature_key。本 subject 把这些契约、表结构、事件与 feature_key 缝合成一个 P0 后端域，并明确仍需在 implementation Phase 0 收口的编码真理源漂移：`PracticeMode` 二值化、practice not-found 错误码、derived plan source 字段、`GenerationProvenance` wire shape、DB/internal turn status 映射、Practice API idempotency 存储、D-22 下的 B3 `report_generate` job ownership 语义，以及独立 F3 baseline 完成状态。

本 subject 不私自绕过 OpenAPI / DB / event / feature_key owner。D-21 / D-24 / D-25 / D-26 / D-27 / D-28 / D-29 记录的契约修订必须先回到对应 owner spec 与编码真理源，再由本 subject 的 plan 实施；未完成这些前置修订前，不得进入依赖该契约的 handler / store / scenario 实现。

## 2 范围

### 2.1 In Scope

- 6 个 Practice operation 的 backend handler + service + store：
  - `POST /practice/plans` `createPracticePlan`：201 + `PracticePlan`，要求 `Idempotency-Key`；同步写入 `practice_plans`，根据 `goal` 派生 `source_report_id` 引用，初始 `status='ready'`。
  - `GET /practice/plans/{planId}` `getPracticePlan`：返回完整 `PracticePlan`。
  - `POST /practice/sessions` `startPracticeSession`：201 + `PracticeSession`，要求 `Idempotency-Key`；先用短事务按 D-27 的 user-scoped idempotency record reserve session，再在事务外同步生成首题（调用 F3 `Resolve("practice.session.first_question", language)` + A3 `AIClient.Complete()`），成功后用短事务写入 `practice_turns(turn_index=1, status='asked')` + `seq_no=1` 的 `session_started` event + `practice.session.started` outbox，并把 `currentTurn` 嵌入响应。
  - `GET /practice/sessions/{sessionId}` `getPracticeSession`：返回完整 `PracticeSession` 用于刷新 / 断网恢复，必须按 `user_id` 过滤。
  - `POST /practice/sessions/{sessionId}/events` `appendSessionEvent`：200 + `SessionEventResult{acknowledged, session, assistantAction}`；通过 body `clientEventId` 实现 per-session idempotency，必须**禁止**携带 `Idempotency-Key` header；按 event `kind` 路由到状态机分支并生成下一个 `AssistantAction`。
  - `POST /practice/sessions/{sessionId}/complete` `completePracticeSession`：202 + `ReportWithJob`，要求 `Idempotency-Key`；事务内把 session 状态推进到 `completing`，写 `session_completed` event，创建 `feedback_reports(status='queued')` placeholder 与 `async_jobs(job_type='report_generate')` queued row，发出 outbox `practice.session.completed`；本域只创建报告占位与 job，不生成报告内容。
- Plan goal 与数据来源派生规则（goal 决定 turn 来源，与 mode 正交）：
  - `baseline`：仅依赖 `target_jobs` + `resumes` 上下文（D-20 扁平简历）；`source_report_id` 为空；首题与后续 turn 由 AI 生成。
  - `retry_current_round`：要求 `source_report_id NOT NULL`，`practice_plans.focus_competency_codes` 来自报告 `next_actions` 中标记 `included_in_retry_plan=true` 的 `question_assessments`；turn 仍由 AI 生成，但聚焦于上一轮缺口。
  - `next_round`：要求 `source_report_id NOT NULL`，`interviewer_persona` / `difficulty` 按 InterviewRound 进阶规则切换；turn 由 AI 生成。
  - `debrief` 已随 product-scope D-22 删除，不再是合法 `PracticeGoal` 或 plan source。
- Session 状态机与首题同步生成：`startPracticeSession` 必须先 reserve session，再在事务外生成首题，最后短事务提交首题与 started event；baseline / retry / next_round 均由 AI 生成首题。AI 调用失败映射到 B1 `AI_PROVIDER_TIMEOUT` / `AI_OUTPUT_INVALID` / `AI_PROVIDER_SECRET_MISSING` / `AI_PROVIDER_CONFIG_INVALID`，session reservation 进入 `failed` 并记录 `failure_code`；同 `Idempotency-Key` 对 failed 且无 currentTurn 的 reservation 允许重试首题生成，不把失败响应固化为永久 dedupe 结果。
- AssistantAction 决策树：`answer_submitted` → 决定下一个动作为 `ask_question`（新 turn）/ `ask_follow_up`（同 turn 加深）/ `session_wait`（队列空闲）/ `session_completed`（达到 `question_budget`）；`hint_requested` → `show_hint`（mode=`assisted`，复用 `practice.turn.lightweight_observe` 的低延迟辅助输出）或 `409 PRACTICE_SESSION_CONFLICT`（mode=`strict`）；`turn_skipped` → baseline/retry/next_round 调 AI 生成下一题；`session_paused` / `session_resumed` → 仅切换 `waiting_user_input`/`running` 状态，不生成新 AI 输出。
- Mode 策略执行（mode 仅控制辅助度，与 goal 正交）：
  - `assisted`：允许 hint，启用 `practice.turn.lightweight_observe` 周期性观察反馈；适用任何 goal。
  - `strict`：拒绝 `hint_requested`（返回 `409 PRACTICE_SESSION_CONFLICT`），禁用 lightweight observe，AssistantAction 仅生成 `ask_question` / `ask_follow_up` / `session_completed`；适用任何 goal。
- Voice operation 服务端 owner 边界：本 spec 拥有 voice 相关新增 operation（如 `createPracticeVoiceTurn`）的 OpenAPI 契约入口、handler 注册、`practice_session_events` 持久化与 idempotency；STT / LLM / TTS profile 选择、committed-context 推进、barge-in 语义由 [practice-voice-mvp](../practice-voice-mvp/spec.md) plan 落地。voice session 复用本 spec 的 6 个 operation 状态机，不新建独立 session 类型。
- 异步执行边界：`completePracticeSession` 在事务内推进 session 到 `completing` + 创建 `feedback_reports` placeholder + 创建 `async_jobs(report_generate)` + outbox 写 `practice.session.completed`；本 spec **不**消费 `report_generate` job（归 `backend-review` future owner），但 D-28 要求 B3 / `shared/jobs.yaml` 先把 `practice.session.completed` 解释为 report job 的 source event / analytics fact，而不是由 outbox dispatcher 再创建第二个 job。job type / dedupe key 必须与 B3 `report_generate` mapping 一致，outbox 事件 payload 与 B3 `practice.session.completed` schema 一致。
- 并发与 idempotency：
  - 副作用 endpoint（`createPracticePlan` / `startPracticeSession` / `completePracticeSession`）按 D-27 的 `(user_id, operation, idempotency_key, request_fingerprint)` 去重。相同 fingerprint 重试返回首次成功结果；相同 key 但 path/body/fingerprint 不同返回 `409 PRACTICE_SESSION_CONFLICT` 或 B1/B2 对齐后的 idempotency mismatch 错误；不同用户同 key 隔离；pending 并发只能有一个执行者。`startPracticeSession` 的首题失败 reservation 是唯一可重试失败态：失败且未产生 `currentTurn` 时，同 key 可重试并覆盖 retryable failure 状态，成功后才固化 201 响应。
  - `appendSessionEvent` 按 `(session_id, client_event_id)` 去重（B4 已落 UNIQUE 约束），同 payload 重试返回首次结果，不重复写 event / 不重复触发 AI；同 `clientEventId` 但 `kind` / payload fingerprint 不同返回冲突；事务内 `SELECT FOR UPDATE practice_sessions WHERE id=$1` → 计算 `seq_no = COALESCE(MAX(seq_no), 0) + 1` → 写 `practice_session_events` → 释放锁；保证多端并发提交不重号、不丢序。
- 用户隔离：所有 read / write SQL 必须按 `user_id` 过滤；越权访问 `getPracticePlan` / `getPracticeSession` / `appendSessionEvent` / `completePracticeSession` 返回 HTTP 404 + B1 `PRACTICE_PLAN_NOT_FOUND` / `PRACTICE_SESSION_NOT_FOUND`（D-26 前置新增），避免泄露存在性；DELETE /me 通过 ON DELETE CASCADE 级联清理 `practice_plans` / `practice_sessions` / `practice_session_events` / `practice_turns`。
- 隐私 / 观测红线：`practice.session.*` / `practice.turn.*` 事件、F1 metric label、log、audit 不得包含 `question_text` / `answer_text` / `hint_text` / AI prompt / response 明文 / provider secret；只允许 IDs、length、count、status、profile、provider、model_id、cost micros、error code 摘要。
- F1 metric 注册边界：所有新增的 practice metric / audit 类型必须先在 [F1 `observability-stack`](../observability-stack/spec.md) baseline 字典登记或由 F1 owner 承接，不得在本域私造 metric / label。
- AI 调用形态：业务侧调用 [F3 `RegistryClient.Resolve`](../prompt-rubric-registry/spec.md) → 拿三元组 → 调用 [A3 `AIClient.Complete`](../ai-provider-and-model-routing/spec.md)；payload metadata 必须携带 `feature_key + prompt_version + rubric_version + model_profile_name + language + data_source_version`。D-29 要求 F3 `001-baseline` 独立派生并完成后，backend-practice 才能进入依赖首题 / 追问 / hint AI 输出的实现阶段。OpenAPI 响应中的 `AssistantAction.provenance` 使用 B2 `GenerationProvenance` 当前 wire 字段，额外 runtime 字段只进入 `ai_task_runs` / audit 摘要，不绕过 B2 私自扩展响应。
- 失败语义与状态机退出：session-survival AI（`practice.session.first_question` / `practice.session.follow_up`）失败 → session=`failed`，写 `failure_code` 为对应 B1 错误码；辅助 AI（`practice.turn.lightweight_observe` / hint）失败按 D-36 graceful degrade，session 保持 `running` 且不写 `failure_code`；超时未活跃（platform sweep 触发，本 spec 仅约定阈值）→ session=`cancelled`；DELETE /me / cascade → 物理删除。

### 2.2 Out of Scope

- 不实现 report 生成、证据回收、维度评估、ReadinessTier 计算、题目回顾页 payload；归 `backend-review` (future) owner。本 spec 仅创建 queued report/job handoff 与 source event，不生成报告内容。
- 不实现真实面试复盘的 intake、分析、复盘面试问题导出；该模块已随 product-scope D-22 删除。本 spec 不再消费 `source_debrief_id` 或 debrief-derived turns。
- 不实现 JD 解析、`target_jobs` 生命周期；归 [`backend-targetjob`](./../backend-targetjob/spec.md) owner。本 spec 仅引用 `target_job_id`。
- 不实现简历解析、改写；归 `backend-resume` owner。本 spec 仅引用 `resumeId`（D-20 扁平简历，原 `resumeAssetId`）。
- 不实现 STT / LLM / TTS 编排、committed-context 推进、barge-in 处理、TTS chunk 播放语义；归 [`practice-voice-mvp`](../practice-voice-mvp/spec.md) owner 与其 plan。
- 不实现独立 worker / Asynq dispatcher / 生产级 outbox consumer；001/002 范围只负责 handler/store 同事务写 `outbox_events` 与 `async_jobs`，真实 dispatcher / runtime drainer 由 future `backend-async-runner` 或对应 owner 接管，不把不存在的 runtime 包作为 002 实施依赖。
- 不暴露独立 cancel API（OpenAPI v1 不提供 `DELETE /practice/sessions/{id}`）；`status='cancelled'` 仅由 platform sweep（超时阈值由本 spec §6 约定，实现归 platform owner）和 DELETE /me cascade 触发。
- 不实现 plan 自动 archival / 用户主动 dismiss；`practice_plans.status='archived'` 列保留以便后续 plan 演进，v1 不写入。
- 不在本 spec 文档内 inline 修改 B2 OpenAPI、B3 events/jobs、B4 baseline 表结构、A3 provider 协议或 F3 baseline prompt / rubric 文本；D-21 / D-24 / D-26 / D-27 / D-28 / D-29 指向的契约修订必须由对应 owner spec / truth source / plan 先落地，再由 backend-practice implementation 消费。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | API 契约来源 | 本域只消费 [B2 OpenAPI](../openapi-v1-contract/spec.md) 已定义的 6 个 Practice operation；不私造 endpoint、不重写 schema | 任何字段 / 新 operation 先在 B2 spec / `openapi.yaml` 修订 |
| D-2 | DB 真理源 | 复用 [B4 baseline](../db-migrations-baseline/spec.md) 的 `practice_plans` / `practice_sessions` / `practice_session_events` / `practice_turns`、与现有索引、CHECK 约束、UNIQUE 约束 | 不在本 spec 内 inline 落 migration；新增列必须先修订 B4 |
| D-3 | 事件契约 | 复用 [B3](../event-and-outbox-contract/spec.md) 已冻结的 `practice.session.started` / `practice.turn.completed` / `practice.session.completed` 与 `report_generate` job mapping；D-28 要求先修订 mapping ownership，避免 outbox dispatcher 二次创建 job | 事件 payload 与 PII 边界不得扩张；新增字段或 job ownership 语义先回到 B3 spec / `shared/jobs.yaml` |
| D-4 | Plan goal 三值 | `baseline` / `retry_current_round` / `next_round`，对应不同 source 引用与 turn seeding 规则；`debrief` 已随 D-22 删除 | 不引入第四种 goal；新增 goal 先回到 B1 / B4 / product-scope |
| D-5 | Mode 二值（仅辅助度） | `assisted` / `strict`；strict 禁用 hint 与 lightweight observe；mode 与 goal 正交，不承担数据来源语义 | 不引入第三种 mode；旧 `debrief` mode 只能作为负向测试输入 |
| D-6 | 完成是异步流 | `completePracticeSession` 返回 202 + `ReportWithJob`，事务内推进 session=`completing` + 创建 `feedback_reports(status='queued')` placeholder + 创建 `async_jobs(report_generate)` + outbox `practice.session.completed`；report 内容由 backend-review 异步 worker 消费 | 用户在 `generating` 屏等待异步报告，不阻塞 HTTP；API 响应中的 `reportId/job` 有真实 DB row 承接 |
| D-7 | 双轨 idempotency | 副作用 endpoint 用 `Idempotency-Key` header；`appendSessionEvent` 用 body `clientEventId`（per session 唯一），不允许携带 `Idempotency-Key` header；`startPracticeSession` 首题失败 reservation 可用同 key 重试，成功后才固化 dedupe 结果 | 防止重复创建 plan / session / 报告 job；防止多端事件重号或重写；避免 AI 临时失败把 session 永久卡死 |
| D-8 | AI 调用形态 | 通过 F3 `RegistryClient.Resolve("practice.session.first_question" / ".follow_up" / ".turn.lightweight_observe", language)` + A3 `AIClient.Complete`；不 hardcode prompt | 业务包不得直接持有 prompt 文本或 model 字符串 |
| D-9 | 用户隔离 | 所有 API 走 [backend-auth](../backend-auth/spec.md) session middleware + user_id 过滤；越权返回 404，不泄露存在性；DELETE /me CASCADE 级联清理 | 与 backend-targetjob 一致的隔离协议 |
| D-10 | AssistantAction provenance | OpenAPI wire 形态复用 B2 `GenerationProvenance{promptVersion,rubricVersion,modelId,language,featureFlag,dataSourceVersion}`；内部 AI metadata / `ai_task_runs` / audit 可保留 `feature_key`、`model_profile_name`、provider、cost、latency 等运行时字段；非评分动作 `rubricVersion='not_applicable'`，`session_wait` / `session_completed` 写 `featureFlag='none'` | 与当前 B2 生成物一致；不在 backend-practice 私自扩展 response schema |
| D-11 | 隐私红线 | 事件 / metric label / log / audit / async payload 不得包含 `question_text` / `answer_text` / `hint_text` / AI prompt / response 明文 / provider secret；只允许 hash、length、count、status、profile、provider、model_id、cost micros、error code | 与 product-scope §9.3 / B3 piiBoundary / F1 一致 |
| D-12 | Spec 粒度 | 单一 `backend-practice` spec 覆盖 plan + session + turn + assistantAction 全部 6 个 operation；与 OpenAPI tag `PracticePlans`+`PracticeSessions` 边界一致 | 不拆 sibling spec；后续按 phase 拆 plan |
| D-13 | 首题同步生成 | `startPracticeSession` 对调用方同步返回 `201 + currentTurn`；实现上禁止把外部 AI 调用包在长 DB transaction 内，必须采用 session reservation + 事务外 AI + 成功/失败短事务收口 | 不破坏 OpenAPI 当前 201 语义；前端不需要轮询；避免锁表等待 provider |
| D-14 | debrief 数据来源退役 | `source_debrief_id` 与 `sourceDebriefId` 已随 product-scope D-22 删除；`goal='debrief'` 不再是合法计划目标 | 相关旧输入必须被 contract / handler / migration / scenario 负向 gate 拒绝或归零 |
| D-15 | Voice operation owner | `createPracticeVoiceTurn` 等 voice 新 operation 的 OpenAPI 契约 + handler 入口 + session event 持久化归本 spec；STT / LLM / TTS profile 选择、committed-context、barge-in 归 practice-voice-mvp | voice 与文本闭环共享同一状态机 / 事件 / idempotency 框架 |
| D-16 | Strict 下 hint 拒绝码 | 在 `mode='strict'` 的 session 上 `hint_requested` 返回 `409 PRACTICE_SESSION_CONFLICT`；ApiError detail 携带 `mode` 与 `policy='hint_disabled_in_mode'`；goal 与 hint 拒绝无关 | 与 B1 已有 conflict code 语义一致；避免与 422 结构性错误混淆 |
| D-17 | 并发序列化 | `appendSessionEvent` 事务必须 `SELECT FOR UPDATE practice_sessions WHERE id=$1 AND user_id=$2`，再计算 `seq_no = COALESCE(MAX(seq_no), 0) + 1` 并写 `practice_session_events`；与 UNIQUE(session_id, seq_no) 共同保护 | 多端并发不重号、不乱序、不丢事件 |
| D-18 | v1 不暴露独立 cancel API | OpenAPI v1 不提供 `DELETE /practice/sessions/{id}`；`status='cancelled'` 仅由 platform timeout sweep（默认 24h 无活跃事件）与 DELETE /me CASCADE 触发；超时阈值实现归 platform / future plan | 减少 v1 分支；保留 SessionStatus.cancelled 出口 |
| D-19 | AI 失败映射 | session-survival AI（first_question / follow_up）：timeout → `502 AI_PROVIDER_TIMEOUT`；invalid output → `502 AI_OUTPUT_INVALID`；upstream unavailable / fallback exhausted → `503 AI_FALLBACK_EXHAUSTED` 或 A3 返回的 B1 `AI_*` 错误；secret missing → `502 AI_PROVIDER_SECRET_MISSING`；session=`failed`，`failure_code` 记录对应 B1 错误码；不静默回退 stub（除 `APP_ENV=test`）。辅助 AI（hint / lightweight_observe）失败例外，按 D-36 graceful degrade（session 保持 running，不写 failure_code，不返回 502/503） | 全部错误码来自 B1，避免私造通用 availability 错误码 |
| D-20 | 文档治理 | 后续修订原地更新本 spec，不创建 sibling spec；删除任何决策必须先修订本 spec 再修代码；旧术语（warmup / single_drill / mistake_queue / drill_builder / 独立 voice route）不得在本 spec 与衍生 plan 中出现 | 与 product-scope §4.5 一致 |
| D-21 | PracticeMode 收敛前置契约修订 | 因 D-5 把 mode 从三值降为二值，`legacy debrief replay value` 不再是合法 PracticeMode 取值；任何 backend-practice 实施 plan 的 Phase 0 必须先修订 [B1 `shared/conventions.yaml#enums[PracticeMode]`](../shared-conventions-codified/spec.md)、[B2 `openapi/openapi.yaml#components.schemas.PracticeMode`](../openapi-v1-contract/spec.md)、[B4 `migrations/000001_create_baseline.up.sql:170` `practice_plans.mode` CHECK](../db-migrations-baseline/spec.md)、B3 `shared/events.yaml` / generated event refs 中引用 B1 PracticeMode 的 surface，把 `legacy debrief replay value` 取值删除（pre-launch 直接改 baseline，不引入 deprecated alias）；前端 mock fixture 与 generated client/server 必须随同更新 | 编码真理源与本 spec 决策对齐；保持 mode 与 goal 两轴正交；契约前置不通过则 implementation plan 不得进入 Phase 1 |
| D-22 | `completePracticeSession` report/job 创建边界 | 本域在 complete 事务内创建 `feedback_reports` placeholder 与 `async_jobs(report_generate)`，并以 `session_id` / `report_id` / idempotency key 防重复；backend-review 只消费 queued job 生成内容 | 让 202 response 的 `ReportWithJob` 可执行，不等待 dispatcher 异步补行 |
| D-23 | 首题失败重试边界 | `startPracticeSession` failed reservation 不视为最终 dedupe 结果；同 user + key + plan 可重试，成功后返回同一 session id 或明确迁移到成功 session，且不得产生两个 active session | 防止 provider 临时错误导致用户无法开始面试 |
| D-24 | Derived plan source 字段前置 | `retry_current_round` / `next_round` 需要 B2 `sourceReportId` + B4 `practice_plans.source_report_id`；D-22 后不再存在 `sourceDebriefId` / `source_debrief_id` | baseline start 可先实施；复练/下一轮不能靠隐藏字段或本地 mock |
| D-25 | Turn status API/DB 映射 | DB `practice_turns.status` 允许内部状态 `asked/answered/follow_up_requested/assessed/skipped`；OpenAPI `PracticeTurn.status` 当前只暴露 `asked/answered/skipped`，handler 必须做映射，除非先回 B2 扩展 schema。由 002 plan-level D-33 落实为 wire enum 扩 5 值（pre-launch baseline rebase），handler 不再做"压缩到 3 值"映射；002 落地后 D-25 的"映射"分支视为已淘汰备选 | 避免前端 SDK 与 DB internal state 漂移；不把内部评估状态强塞进 wire enum |
| D-26 | Practice 错误码前置 | B1/B2 必须新增 `PRACTICE_PLAN_NOT_FOUND` / `PRACTICE_SESSION_NOT_FOUND` 并同步 generated Go/TS/OpenAPI 后，backend-practice 才能在 404 隔离路径使用这些 code；未认证统一使用既有 `AUTH_UNAUTHORIZED` | 保持错误码单一真理源；避免 spec 中出现未注册字面量 |
| D-27 | Practice idempotency 存储与 replay 语义 | B4 / backend-practice Phase 0 必须新增 user-scoped API idempotency 表（载体已由 D-30 收敛为 shared `idempotency_records`，含 `domain` / `operation` namespace 字段），至少保存 `user_id`、`domain`、`operation`、`idempotency_key_hash`、`request_fingerprint`、`status(pending/succeeded/failed_retryable/failed_terminal)`、`resource_type`、`resource_id`、`response_body`、`error_code`、`expires_at`；副作用 endpoint 用该记录锁定单执行者、replay 成功响应、拒绝同 key 不同 fingerprint、隔离跨用户；`startPracticeSession` 首题前失败可写 `failed_retryable` 并允许同 key 重试 | 让 D-7/D-23/C-10 可落库、可测试、可并发；避免仅靠业务表猜测重复请求；shared 载体让 backend-targetjob / backend-review / 002 复用同款基建 |
| D-28 | D-22 下 B3 job ownership 前置 | 用户已确认坚持 D-22：`completePracticeSession` 同事务创建 `feedback_reports` placeholder 与 `async_jobs(report_generate)`。因此 B3 `event-and-outbox-contract` / `shared/jobs.yaml` / generated jobs docs 必须在 Phase 0 明确 `practice.session.completed` 是 report job 的 source event 与 analytics fact，不能再由 outbox dispatcher 根据同一事件创建第二个 `report_generate` job；dispatcher / backend-review 只消费既有 queued job 或按 dedupe key 查找既有 job | 保持 `ReportWithJob` 可执行，同时避免重复 job、重复报告与 dispatcher 职责冲突 |
| D-29 | F3 baseline 独立前置 | `prompt-rubric-registry/001-baseline` 必须独立派生、完成并通过其 Resolve / prompt / rubric / lint gates 后，backend-practice 才能进入依赖 `practice.session.first_question` / `practice.session.follow_up` / `practice.turn.lightweight_observe` 的实现阶段；未完成前只允许契约 / migration / 非 AI store work | 防止 backend-practice hardcode prompt 或在 F3 真理源未落地时启动 AI handler |
| D-30 | 001-Phase-0 跨 spec 修订归属与 idempotency 表载体 | (a) `001-plan-and-session-orchestration` Phase 0 直接修订 B1 `shared/conventions.yaml` / B2 `openapi/openapi.yaml` / B3 generated event refs / B4 migrations 编码真理源（integrator 模式），同时同步追加各 owner spec 的 `history.md` 与 `spec.md` Header 授权记录与版本号；不再为 D-21 / D-26 / D-27 各派 sibling owner spec plan。(b) D-27 idempotency 存储载体收敛为 **shared** `idempotency_records` 表（含 `domain` / `operation` namespace 字段），由 001 Phase 0 引入并设计为可被 backend-targetjob / backend-review / 自身 002 等未来 backend domain 直接复用 | 缩短 critical path；shared idempotency 基建在引入第一个 caller 时一并设计，避免后续重构；ownership 软化由各 owner spec 在 history 显式登记"协调修订模式 / 关联计划: backend-practice/001 Phase 0"兜底 |
| D-31 | baseline plan 必须绑定简历资产（**D-39 重塑（D-20）**：`resumeAssetId`→`resumeId`） | `createPracticePlan` 的 baseline 路径要求 `resumeId`（D-20 前为 `resumeAssetId`）为 schema 必填字段，并且该 resume 必须属于当前用户且未删除；缺失、空值或不可用 resume 均返回 `422 VALIDATION_FAILED`。B2 `CreatePracticePlanRequest.required`、fixtures、Go/TS generated artifacts、frontend request builder 与 backend service/store 必须保持同一口径 | 当前 workspace UI 已把缺简历作为进入面试前的阻塞空状态；首题上下文和后续报告证据链依赖目标岗位 + 简历绑定，避免契约允许客户端发送不完整 plan |
| D-36 | hint / lightweight_observe AI 失败按 graceful degrade（由 003 plan-level 派生） | 辅助 AI（`practice.turn.lightweight_observe` / hint）失败 → `appendSessionEvent` 返回 `200 + AssistantAction{type:'session_wait', hint:null, sessionStatus:'running'}`；DB `practice_sessions.status` 保持 `running`，`failure_code` 保持 NULL；`practice_turns.hint_text` 保持 NULL；不返回 502/503，不写 `audit_events`；service-local `SessionEventOutcome.AuditMetadata["hint_degrade_reason"]` 只允许携带 sanitized B1 error_code 且不得落入 `audit_events.metadata` 或 wire envelope；F3 resolve 类失败统一映射为 B1 `AI_PROVIDER_CONFIG_INVALID`，A3 / parse 类失败使用 A3 返回或解析得到的 B1 `AI_*` code；ai_task_runs 仍写 `task_type='hint_generate', validation_status='failed', error_code` 来自 B1 enum 作为运维可观测兜底 | hint 是 session-running 期间用户主动触发的辅助 AI；强制 fail-closed 会因临时 AI 故障中断答题循环，并与 002 D-19 follow_up fallback 模式不一致；session-survival AI 仍按 C-17 / D-19 narrowed 规则 fail-closed |
| D-38 | hint turn-lifecycle 边界（由 003 plan-level 派生） | hint_requested 路径在 `practice_session_events` 上写入 `kind='hint_requested'` 留痕，但不递增 `practice_sessions.turn_count`、不改 `practice_turns.status` / `turn_index` / `follow_up_count`、不发 `practice.turn.completed` outbox、不写 `audit_events`；assisted 成功路径 UPDATE `practice_turns.hint_text`，degrade / strict 路径 hint_text 保持 NULL；hint 不计入 `question_budget` | 把 hint 与 answer_submitted 在 turn 主表上的写入路径彻底解耦，避免 hint 路径误占预算 / 误推进 turn 状态 / 误触发完成事件；002 已固定 strict-default 不写 hint_text 等行为，003 在 assisted 接入后必须保持同一边界 |
| D-39 | 简历扁平化绑定适配（product-scope D-20） | `createPracticePlan` / `practice_plans` 的简历绑定从 `resumeAssetId` / `resume_asset_id` 改为 `resumeId` / `resume_id`（指向扁平 `resumes` 表：[B4 D-22](../db-migrations-baseline/spec.md) 已 rename `practice_plans.resume_asset_id`→`resume_id`，[B2 D-26](../openapi-v1-contract/spec.md) `CreatePracticePlanRequest.resumeAssetId`→`resumeId` 随全局 resumeId 重命名）；移除「简历主版本 / 岗位定制版本」上下文口径（D-20 简历无版本）；baseline session 首题 prompt 引用扁平 resume 的 `structured_profile`。 | 由 backend-practice/001 D-20 phase 落地 handler / service / store / generated 类型 rename；与 [B4 D-22](../db-migrations-baseline/spec.md) / [B2 D-26](../openapi-v1-contract/spec.md) / [backend-resume D-13](../backend-resume/spec.md) 同步 |

### 3.2 非后端 owner 决策

| ID | 事项 | Owner | 本域处理 |
|----|------|-------|----------|
| Q-1 | 报告 next_actions → retry plan focus_competency_codes 派生算法 | backend-review (future) | 本 spec 仅约定 `practice_plans.focus_competency_codes` 字段从报告读取；具体算法在 backend-review plan |
| Q-2 | debrief-derived turn seeding | 已关闭 | product-scope D-22 删除该路径；不再等待 backend-debrief owner |
| Q-3 | session 24h timeout sweep 实现 | platform / future plan | 本 spec 仅约定 SessionStatus.cancelled 的退出路径与阈值默认值；不在本 spec 内派 plan |

### 3.3 待确认事项

| ID | 待确认事项 | 影响 | 默认处理 |
|----|------------|------|----------|
| Q-4 | `practice.turn.lightweight_observe` 触发频率 | assisted 模式下 AI 周期性观察是否每 turn 触发一次还是按时间窗 | 默认每 turn 完成时触发一次；后续 plan 在 BDD 中确认体验阈值 |
| Q-5 | follow-up 在 `question_budget` 中的计数方式 | 追问是否占用题目预算 | 默认追问不占预算；`turn_count` 增长以独立 turn 为准，follow-up 在同 turn 内通过 `practice_turns.follow_up_count` 累加 |

## 4 设计约束

### 4.1 API 契约约束

- 必须使用 [B2 generated `ServerInterface`](../openapi-v1-contract/spec.md) 注册 6 个 handler，不得绕过 generated types 自造 router。
- 入参反序列化必须使用 generated request types；响应必须使用 generated response types（`PracticePlan` / `PracticeSession` / `PracticeTurn` / `SessionEventResult` / `AssistantAction` / `ReportWithJob`）；fixture 与真实 handler 共用同一 schema。
- `startPracticeSession` 必须采用 D-13/D-23 的三段式：短事务 reserve session row，事务外生成首题，成功短事务写 `seq_no=1` 的 `session_started` event + `turn_index=1` 的 `practice_turns(status='asked')` 并在响应里嵌入 `currentTurn`；AI 失败用短事务把 reservation 标记为 `failed` + `failure_code`，同 key 可重试，不把外部 AI 调用包进长 DB transaction。
- `appendSessionEvent` 必须根据 `kind` 路由到状态机分支：
  - `answer_submitted`：要求 `payload.turnId` + `payload.answerText`，写 `practice_turns.answer_text` / `answer_summary`，DB internal turn 状态可推进 `asked → answered → assessed`；OpenAPI `PracticeTurn.status` 按 D-25 映射到 wire enum，根据 budget 决定下一动作。
  - `hint_requested`：mode=`assisted` 时调用 F3 `practice.turn.lightweight_observe` + A3 生成 hint，写 `practice_turns.hint_text`；mode=`strict` 时返回 `409 PRACTICE_SESSION_CONFLICT`。
  - `turn_skipped`：写 `practice_turns.status='skipped'`，推进到下一题。
  - `session_paused`：session=`waiting_user_input`，AssistantAction=`session_wait`。
  - `session_resumed`：session=`running`，AssistantAction 重发当前 turn 的 question。
- `completePracticeSession` 必须验证 session 状态机迁移合法（`running` / `waiting_user_input` → `completing`），非法迁移返回 `409 PRACTICE_SESSION_CONFLICT`；事务内创建 `feedback_reports(status='queued')` placeholder + `async_jobs(job_type='report_generate', status='queued')`，并写 outbox `practice.session.completed`。
- 所有副作用 endpoint 必须先经过 D-27 idempotency middleware / service layer：同一 user + operation + key + fingerprint 只执行一次；pending 并发用 DB row lock 或唯一约束串行化；commit 后 response 丢失时重试返回持久化 response；同 key 不同 fingerprint 必须拒绝，不得执行第二个副作用。

### 4.2 数据约束

- 所有 DB 写入必须在各自短事务内完成；外部 AI 调用不得包在 DB transaction 内。`practice.session.started` / `practice.turn.completed` / `practice.session.completed` 必须通过 outbox 与对应业务写入同事务发出，避免双写。
- `practice_session_events.event_type` 仅允许 B4 已有 CHECK 列表；新增事件 type 先回到 B4 修订。
- `practice_turns.status` 仅允许 DB internal 值 `asked` / `answered` / `follow_up_requested` / `assessed` / `skipped`；`answered → assessed` 由本 spec 在 turn 结束时基于 `practice.turn.lightweight_observe` 反馈推进，handler 必须按 D-25 映射到 B2 当前 `PracticeTurn.status` wire enum。
- `practice_plans.source_report_id` 只用于 report-derived `retry_current_round` / `next_round`；D-22 后不存在 `source_debrief_id` 或 debrief-derived CHECK 分支。
- D-27 idempotency 存储必须有唯一约束保护 `(user_id, operation, idempotency_key_hash)`，并保存 request fingerprint 与 terminal response；response body 可保存为 B2 response JSON 摘要 / envelope，禁止保存 prompt、answer、hint 或 provider raw response。
- 软删 / 物理删除：v1 不引入软删；DELETE /me 通过 ON DELETE CASCADE 物理清理；session timeout sweep 仅修改 `status='cancelled'` + `cancelled_at`，不删除行。

### 4.3 安全 / 隐私约束

- 所有 API 走 [backend-auth](../backend-auth/spec.md) session middleware；未认证返回 `401 AUTH_UNAUTHORIZED`；越权返回 404 + `PRACTICE_PLAN_NOT_FOUND` / `PRACTICE_SESSION_NOT_FOUND`（D-26 前置新增），不泄露资源存在性。
- AI 调用 fail-closed 边界按调用类别区分：session-survival AI（`practice.session.first_question` / `practice.session.follow_up`）在 F3 `Resolve` 返回 unsupported / disabled profile 或 A3 缺 provider secret 时，整个 operation 返回 B1 错误并把 session 置 `failed`；不得静默回退到 stub provider（除 `APP_ENV=test`）。辅助 AI（`practice.turn.lightweight_observe` / hint）失败按 D-36 graceful degrade：session 保持 `running`，wire 返回 `200 + AssistantAction{type:'session_wait', hint:null}`，运维仍可通过 `ai_task_runs(task_type='hint_generate', validation_status='failed', error_code=<B1 AI_*>)` 观测真实失败；service-local degrade reason 不得进入 `audit_events.metadata` 或响应体。
- log / metric label / audit / 事件 payload 不得包含 `question_text` / `answer_text` / `hint_text` / AI prompt body / AI response body / provider secret；仅允许 hash、length、count、status、profile、provider、model_id、cost micros、error code 摘要。
- `audit_events` 触发集合：`createPracticePlan` / `startPracticeSession` / `completePracticeSession` 必须写 audit；`appendSessionEvent` 不写 audit（高频 + B3 outbox 已记录）；audit metadata 仅含 `plan_id` / `session_id` / `goal` / `mode` / `language` / `target_job_id`，不含问答文本。

### 4.4 异步 / 可观测约束

- 本域 P0 只创建 `report_generate` queued job、不消费该 job；002 不引入 runtime dispatcher / drainer 包，但必须按 D-28 保证 job row / outbox emit 路径与未来 dispatcher 兼容，且 outbox 重放不会创建第二个 report/job。未来 `backend-async-runner` plan 接管 dispatcher 时必须读取 B3 job ownership 语义并验证 graceful shutdown / drain timeout。
- F1 metric 字典登记前置：practice-specific business metrics（如 session started / completed / duration）实施前必须先在 [F1 baseline](../observability-stack/spec.md) 字典登记，并且 label 只能使用 F1 allowed labels 与 B1 有界枚举；AI 调用 metric 复用 A3 已登记的 7 个 `ai_task_*` metric family，`feature_key` / `prompt_version` / `rubric_version` 只进入 `ai_task_runs` typed columns 或审计摘要，不进入 metric label。
- 同步首题与 follow-up 的 P95 延迟作为观测目标登记，但不作为本 spec 验收 gate。

### 4.5 文档治理约束

- 本 spec 后续修订必须原地更新；不允许创建同主题 sibling spec。
- 涉及 OpenAPI / events / migrations / shared enums / runtime config 的修改必须先回 owner spec / `*.yaml` truth source。
- 涉及用户行为流的 plan 必须维护 BDD gate；本域所有 backend operation 都属于用户可见 API 行为，必须有 BDD 场景。
- 命中 `completed` plan 时不创建同主题 sibling follow-up plan，原地修订即可。
- 旧术语（`warmup` / `single_drill` / `drill_builder` / `mistake_queue` / `growth_center` / 独立 `voice` route / `practiceModeCard`）不得在本 spec 与衍生 plan 中出现，违反必须先修订 [product-scope](../product-scope/spec.md)。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| API contract | [B2 `openapi-v1-contract`](../openapi-v1-contract/spec.md) | 6 个 Practice operation 与 voice 扩展 operation 的 schema、fixtures、generated client / server |
| Backend domain | `backend-practice`（本 spec） | handler / service / store / async handoff row creation / state machine / AssistantAction generator / outbox emit |
| DB schema | [B4 `db-migrations-baseline`](../db-migrations-baseline/spec.md) | `practice_plans` / `practice_sessions` / `practice_session_events` / `practice_turns` 列与索引；shared `idempotency_records` 表由 B4 承载，backend-practice 是首个 caller，必须使用 `domain` / `operation` namespace 保持 future backend domain 可复用；`source_debrief_id` 已由 D-22 删除 |
| Event / job contract | [B3 `event-and-outbox-contract`](../event-and-outbox-contract/spec.md) | `practice.session.started` / `practice.turn.completed` / `practice.session.completed` 与 `report_generate` job mapping；D-28 要求 report job row 由 `completePracticeSession` 创建，事件只作 source fact / analytics |
| Shared enums / errors | [B1 `shared-conventions-codified`](../shared-conventions-codified/spec.md) | `PracticeMode` / `PracticeGoal` / `InterviewerRole` / `SessionStatus` / `PRACTICE_SESSION_CONFLICT` / `PRACTICE_PLAN_NOT_FOUND` / `PRACTICE_SESSION_NOT_FOUND` |
| AI provider | [A3 `ai-provider-and-model-routing`](../ai-provider-and-model-routing/spec.md) | `AIClient.Complete`、provider registry、model profile、observability decorator |
| Prompt / rubric | [F3 `prompt-rubric-registry`](../prompt-rubric-registry/spec.md) | `practice.session.first_question` / `practice.session.follow_up` / `practice.turn.lightweight_observe` feature_key、Resolve 实现、baseline prompt / rubric |
| Config / secret | [A4 `secrets-and-config`](../secrets-and-config/spec.md) | provider secret、feature flag |
| Observability | [F1 `observability-stack`](../observability-stack/spec.md) | metric / audit 类型登记、label allowlist、隐私红线 |
| Auth / isolation | [`backend-auth`](../backend-auth/spec.md) | session middleware、user-scoped read/write、idempotency framework、DELETE /me CASCADE 协议 |
| Upstream — TargetJob | [`backend-targetjob`](../backend-targetjob/spec.md) | 提供 `target_job_id` 与解析后的 requirements / fitSummary / company language |
| Upstream — Resume | `backend-resume` | 提供 `resumeId` 与扁平简历（`structured_profile`）上下文（D-20，原 `resume_asset_id` + 版本树） |
| Downstream — Report | future `backend-review` | 消费既有 `async_jobs(report_generate)` queued job，关联 `practice.session.completed` source event / analytics fact，填充 `feedback_reports` 内容并生成 `question_assessments`；本 spec 仅创建 placeholder / job / source event |
| Voice extension | [`practice-voice-mvp`](../practice-voice-mvp/spec.md) | STT / LLM / TTS profile 选择、committed-context、barge-in；本 spec 提供 voice operation 契约入口与 session event 持久化 |
| Frontend consumer | future `frontend-workspace-and-practice` | Interview Session 与 Report Dashboard mock → real 切换；本 spec 不直接耦合前端组件 |
| Scenario coverage | scenarios owner + 本 subject | `E2E.P0.0NN-practice-session-*` 套件 setup / trigger / verify / cleanup（具体编号在各 plan 内分配） |
| Async runner replacement | future `backend-async-runner` | 接管 runtime dispatcher / drainer，必须沿用 B3 payload red-line 与 D-32 job ownership 语义 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | baseline plan 创建 | 已登录用户拥有 `target_job_id` 与 `resumeId`（D-20 扁平简历），`goal='baseline'` | `POST /practice/plans` 携带 `Idempotency-Key` | 返回 201 + `PracticePlan{status:'ready'}`，DB 写入 `practice_plans`，`source_report_id` 为空，audit_events 写入元数据摘要 | 001-plan-and-session-orchestration |
| C-2 | retry / next_round plan 派生 | 用户对某 `feedback_reports` 选择 `复练当前轮` 或 `进入下一轮` | `POST /practice/plans` 携带 `goal IN ('retry_current_round','next_round')` 与 `sourceReportId` | DB 写入 `source_report_id`；`focus_competency_codes` 来源于 report `next_actions` 中 `included_in_retry_plan=true` 的 turn id 集合 | 004-derived-plans-debrief |
| C-3 | debrief source 已退役 | 请求携带旧 `goal='debrief'` 或 `sourceDebriefId` | `POST /practice/plans` | 返回 validation/conflict 错误，不写 `source_debrief_id`，不创建 debrief-derived turns | product-scope/001-core-loop-module-pruning |
| C-4 | 同步启动 session 与首题 | plan 处于 `status='ready'`，F3 / A3 active | `POST /practice/sessions` 携带 `Idempotency-Key` 与 `planId` | 返回 201 + `PracticeSession{status:'running', currentTurn:{turnIndex:1, status:'asked', questionText, questionIntent, askedAt}}`，事件 `practice.session.started` 出 outbox | 001 |
| C-5 | startPracticeSession AI 失败重试 | F3 active 但 A3 返回 `AI_PROVIDER_TIMEOUT` | `POST /practice/sessions` 携带 `Idempotency-Key` | 短事务 reservation 被标记为 session=`failed` + `failure_code='AI_PROVIDER_TIMEOUT'`，无 `currentTurn` / started outbox；同 `Idempotency-Key` 重试可重新生成首题并最终成功固化 201；client 可见 502 错误 envelope 不包含 prompt / response 明文 | 001 |
| C-6 | 答题循环与下一题 | session=`running`，currentTurn=`asked` | `POST /practice/sessions/{id}/events` 携带 `clientEventId` + `kind='answer_submitted'` + `payload.{turnId,answerText}` | 200 + `SessionEventResult{acknowledged:true, session, assistantAction.type IN ('ask_question','ask_follow_up','session_completed')}`；`practice_turns` 写 answer_text，DB internal 状态可推进 `asked → answered → assessed`，API response 按 D-25 映射；事件 `practice.turn.completed` 出 outbox | 002-event-loop-and-completion |
| C-7 | hint 在 assisted 下生效 | session.mode=`assisted`，currentTurn=`asked` | `POST /practice/sessions/{id}/events` 携带 `kind='hint_requested'` | 200 + `assistantAction.type='show_hint'`；`practice_turns.hint_text` 落库；hint 使用 F3 `practice.turn.lightweight_observe`，wire provenance 只暴露 B2 `GenerationProvenance` 当前字段 | 003-mode-policies-and-provenance |
| C-8 | hint 在 strict 下被拒 | session.mode=`strict`（无论 goal 取值） | `POST /practice/sessions/{id}/events` 携带 `kind='hint_requested'` | 409 + `ApiError{code:'PRACTICE_SESSION_CONFLICT', detail:{mode:'strict', policy:'hint_disabled_in_mode'}}`；session 状态不变 | 003 |
| C-9 | 重复 clientEventId idempotency | 同一 `(session_id, client_event_id)` 已处理过 | 重新 `POST /practice/sessions/{id}/events` | 返回首次结果（`acknowledged:true`），不写新 event row，不重复触发 AI 调用 | 002 |
| C-10 | 重复 Idempotency-Key dedupe | 同一 `(user_id, idempotency_key)` 已被 createPracticePlan / startPracticeSession / completePracticeSession 处理 | 重复请求 | 返回首次结果，DB 不出现重复 plan / session row，outbox 不重复发事件 | 001 + 002 |
| C-11 | 完成 session 触发异步 report | session=`running`，达到 question_budget 或用户主动结束 | `POST /practice/sessions/{id}/complete` 携带 `Idempotency-Key` | 202 + `ReportWithJob{reportId, job{jobType:'report_generate', status:'queued'}}`；同事务创建 `feedback_reports(status='queued')` placeholder 与 `async_jobs(report_generate)` queued row，session=`completing`，事件 `practice.session.completed` 出 outbox；report 内容由 backend-review 异步生成（本 spec 不验证 report payload） | 002 |
| C-12 | AssistantAction provenance 完整 | 任意 AssistantAction 生成（ask_question / ask_follow_up / show_hint） | 检查响应 payload | `provenance` 仅含 B2 当前 wire 字段 `promptVersion` / `rubricVersion` / `modelId` / `language` / `featureFlag` / `dataSourceVersion`；`feature_key` / `model_profile_name` / provider / cost / latency 等运行时字段只进入 `ai_task_runs` / audit 摘要 | 003 |
| C-13 | Cross-user 隔离 | 用户 A 持有 planX / sessionY，用户 B 携带相同 `Idempotency-Key` | 用户 B 调 6 个 operation 任意一个 | 用户 B 不能读到 / 改到 planX / sessionY；越权返回 HTTP 404 + `PRACTICE_PLAN_NOT_FOUND` / `PRACTICE_SESSION_NOT_FOUND`；Idempotency dedupe 仅在同 user 范围生效 | 001 + 002 |
| C-14 | 并发 appendSessionEvent 序列化 | 用户在两个 tab 同时提交不同 `clientEventId` 的事件 | 并发 `POST /practice/sessions/{id}/events` | DB 中 `practice_session_events.seq_no` 严格递增、不重号、不丢；UNIQUE(session_id, seq_no) 约束生效；事务通过 `SELECT FOR UPDATE` 序列化 | 002 |
| C-15 | DELETE /me CASCADE 清理 | 用户 A 调用 `DELETE /me`（backend-auth 入口） | 级联删除 | 用户 A 的所有 `practice_plans` / `practice_sessions` / `practice_session_events` / `practice_turns` 行被物理删除；外部已发出的 outbox 事件不回滚（B3 已固化） | 006-privacy-cascade-and-cleanup |
| C-16 | 隐私红线 | 任意 plan / session / turn / complete 操作完成 | 检查 log / metric label / audit / 事件 payload | 不含 question_text / answer_text / hint_text / AI prompt / response 明文 / provider secret；只含 IDs / length / count / status / profile / provider / model_id / cost micros / error code 摘要 | 006 |
| C-17 | F3 / A3 fail-closed（session-survival AI）；辅助 AI graceful degrade | 选中的 practice profile 不可解析或 provider 缺 secret，且 `APP_ENV` 不在 `test` | startPracticeSession 或 appendSessionEvent 触发 AI | session-survival AI（first_question / follow_up）：整个 operation 返回 B1 错误（`AI_PROVIDER_SECRET_MISSING` / `AI_PROVIDER_CONFIG_INVALID` / `AI_PROVIDER_TIMEOUT` 等）并写 session=`failed` + `failure_code`；不静默回退 stub。辅助 AI（hint / lightweight_observe）：按 D-36 graceful degrade（200 + session_wait，session 保持 running，不写 failure_code，不返回 502/503），ai_task_runs 仍记录 failed row 供运维观测 | 003 |
| C-18 | Voice operation 入口语义 | practice-voice-mvp plan 引入 `createPracticeVoiceTurn` operation | 调用该 operation | 由本 spec owner 注册 handler、写 `practice_session_events`（kind 扩展由 voice plan 与 B3/B4 协作）、走双轨 idempotency；STT/LLM/TTS profile 选择与 committed-context 推进归 voice plan 实现 | 005-voice-turn-extension |
| C-19 | session 24h 超时取消 | session=`waiting_user_input` 且 `updated_at < now() - 24h` | platform sweep 触发（实现归 platform owner） | session=`cancelled`，`cancelled_at=now()`；不发出新 outbox 事件；本 spec 仅约定阈值与状态出口 | 006 |
| C-20 | 文档与历史治理 | 本 spec 状态变更或决策调整 | 更新 spec / history / `plans/INDEX.md` / `docs/spec/INDEX.md` | 文档保持单一 owner，无 sibling spec；旧术语（`warmup` / `single_drill` / `mistake_queue` / `drill_builder` / `growth_center` / 独立 `voice` route / `practiceModeCard`）零引用 | docs-only |
| C-21 | 副作用 endpoint 成功 replay | `createPracticePlan` / `startPracticeSession` / `completePracticeSession` 首次请求已成功提交，但客户端未收到响应 | 同一用户用相同 `Idempotency-Key` 和相同 request fingerprint 重试 | 返回首次成功响应；不重复创建 plan/session/turn/report/job/outbox；idempotency record 从 `pending` 进入 `succeeded` 并保留 replay response | 001 + 002 |
| C-22 | idempotency key body mismatch | 同一用户已用某 `Idempotency-Key` 提交过请求 A | 复用同 key 但 path/body/fingerprint 不同的请求 B | 返回冲突错误，不执行 B，不写新业务副作用；错误 envelope 不泄露 A 的资源内容 | 001 + 002 |
| C-23 | idempotency 跨用户隔离 | 用户 A 与用户 B 使用相同 `Idempotency-Key` | 两人分别调用同一副作用 endpoint | 两个用户的 idempotency records 独立；用户 B 不能拿到用户 A 的 response/resource；各自 user-scoped 副作用按权限执行 | 001 + 002 |
| C-24 | idempotency 并发单执行者 | 同一用户两个 tab 同时提交同 endpoint + 同 key + 同 fingerprint | 并发请求进入 handler | DB 唯一约束 / row lock 保证只有一个执行者；另一个请求等待 replay 或返回 retryable pending 状态；最终只产生一份业务副作用 | 001 + 002 |
| C-25 | startPracticeSession 同 plan 多 key 并发 | 同一 ready plan 被同一用户两个不同 `Idempotency-Key` 同时启动 | 两个 `POST /practice/sessions` 并发 | 最多一个 active/running session；另一个返回 `PRACTICE_SESSION_CONFLICT` 或复用既有 active session（由 001 plan 锁定），不得绕过 idempotency 生成双 session | 001 |
| C-26 | appendSessionEvent clientEventId replay / mismatch | session 已处理某 `clientEventId` | 同 payload 重试，或同 `clientEventId` 换 `kind/payload` 重试 | 同 payload 返回首次 `SessionEventResult`，不重复写 event / 不重复触发 AI；payload mismatch 返回 conflict；不同 session 可独立使用相同 clientEventId；携带 `Idempotency-Key` header 被拒绝；`show_hint` replay 必须绑定原事件 response snapshot，不得从后续可变的 `practice_turns.hint_text` 重建 | 002 + 003 L2 |
| C-27 | completePracticeSession D-22 去重 | session 首次 complete 已创建 queued report/job/outbox | 同 key、不同 key、或 outbox 重放再次触发 report generation handoff | 同 key 返回同一 `ReportWithJob`；不同 key 不创建第二个 report/job（返回既有结果或 conflict，由 002 plan 锁定）；outbox 重放只查找既有 job 或被 dedupe key 拦截；事务失败不留下 session=`completing` 但无 report/job 的半状态 | 002 |

## 7 关联计划

`001-plan-and-session-orchestration` 已派 plan（spec v1.4 同会话），其余 plan 按 phase closability 与 owner 边界依次派生。全局前置：必须先独立派生并完成 `prompt-rubric-registry/001-baseline`（D-29），否则 backend-practice 只能推进不依赖 AI 输出的契约 / migration / store 准备工作。

1. [`001-plan-and-session-orchestration`](./plans/001-plan-and-session-orchestration/plan.md)：D-12 + D-13 + D-21 + D-23 + D-26 + D-27 + D-30 主流程（createPracticePlan baseline + startPracticeSession reservation/首题/失败重试 + getPracticePlan/getPracticeSession + practice.session.started outbox + shared `idempotency_records` 表）；Phase 0 按 D-30 Q1=A integrator 模式直接修订 B1/B2/B3/B4 编码真理源，并同步追加各 owner spec history append 与 Header bump；含 PracticeMode 二值化、practice not-found 错误码、B3 PracticeMode event surface、Practice idempotency storage/replay 语义与 F3 baseline preflight 检查
2. [`002-event-loop-and-completion`](./plans/002-event-loop-and-completion/plan.md)（completed, 2026-05-13）：D-6 + D-7 + D-22 + D-25 + D-27 + D-28 全 5 种 event kind 状态机 + completePracticeSession 202 + placeholder report/job + practice.turn.completed / practice.session.completed outbox + 双轨 idempotency + **B2 `PracticeTurn.status` wire enum 扩 5 值（D-33 落实 D-25，pre-launch baseline rebase）** + **B3 `shared/jobs.yaml#report_generate` `triggerEventSemantic: source_event_only`（D-32 落实 D-28，未来 dispatcher 必须按 generated `JobTriggerEventSemantic*` 常量在 dispatch-time 跳过；002 阶段无 runtime dispatcher，由 jobs.yaml lint + 常量 + handler 端 `async_jobs(job_type, dedupe_key)` UNIQUE + repo grep 兜底）** + **D-34 plan-level 决策：`hint_requested` 在 002 默认 strict 409，等待 003 接手 assisted 分支** + **D-35 plan-level 决策：已完成 session 的二次 complete 不论 `Idempotency-Key` 是否一致都返回既有 `ReportWithJob`，idempotency key 仅控制 inflight 单执行者**
3. [`003-mode-policies-and-provenance`](./plans/003-mode-policies-and-provenance/plan.md)：D-5 + D-10 mode 策略（仅 assisted/strict 两支） + AssistantAction provenance + show_hint feature_key + lightweight observe + **D-36 plan-level（hint / lightweight_observe AI 失败 graceful degrade narrowing；同步 inline-narrow C-17 / D-19 / §4.3 / §2.1 失败语义文字）** + **D-37 plan-level（B4 `ai_task_runs.task_type` CHECK 扩值 `hint_generate` pre-launch baseline rebase + `AITaskRunTaskHintGenerate` writers.go enum）** + **D-38 plan-level（hint turn-lifecycle 边界：不递增 turn_count / 不发 practice.turn.completed outbox / 不写 audit / 不改 turn.status）**
4. [`004-derived-plans-debrief`](./plans/004-derived-plans-debrief/plan.md)：历史目录名保留；D-22 后当前有效范围只剩 retry/next_round report-derived plan 派生 + B2 `sourceReportId` / B4 `source_report_id`；debrief/sourceDebriefId/source_debrief_id 分支已退役，负向 gate 归 product-scope/001 承接
5. `005-voice-turn-extension`：D-15 与 practice-voice-mvp 协作 + voice operation OpenAPI 修订
6. `006-privacy-cascade-and-cleanup`：D-9 + D-11 + D-18 DELETE /me CASCADE + audit_events + timeout sweep contract

每个 plan 通过 `/design` 落地时单独配 BDD/test plan；本 spec §6 AC 是这些 plan 的统一来源。

## 8 相关文档

- [Product Scope §6.8 M3 模拟面试编排器](../product-scope/spec.md#67-m3模拟面试编排器)
- [Product Scope §6.9 M4 证据化报告](../product-scope/spec.md#68-m4证据化报告)（下游 owner 边界参考）
- [docs/ui-design/module-practice-review.md](../../ui-design/module-practice-review.md)
- [docs/ui-design/report-dashboard.md](../../ui-design/report-dashboard.md)
- [openapi-v1-contract](../openapi-v1-contract/spec.md)
- [event-and-outbox-contract](../event-and-outbox-contract/spec.md)
- [db-migrations-baseline](../db-migrations-baseline/spec.md)
- [shared-conventions-codified](../shared-conventions-codified/spec.md)
- [ai-provider-and-model-routing](../ai-provider-and-model-routing/spec.md)
- [prompt-rubric-registry](../prompt-rubric-registry/spec.md)
- [secrets-and-config](../secrets-and-config/spec.md)
- [observability-stack](../observability-stack/spec.md)
- [backend-auth](../backend-auth/spec.md)（idempotency framework 与 auth/session middleware 先例）
- [backend-targetjob](../backend-targetjob/spec.md)（上游 target_job_id 提供方）
- [practice-voice-mvp](../practice-voice-mvp/spec.md)（voice 协作 owner）
- [docs/development.md §2 Frontend / Backend Contract Workflow](../../development.md)
