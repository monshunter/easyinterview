# Backend Debrief Spec

> **版本**: 1.3
> **状态**: active
> **更新日期**: 2026-05-21

## 1 背景与目标

`backend-debrief` 承接 [engineering-roadmap §5.2](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 中 `Debrief` workstream 的后端域，并把 [product-scope D-9](../product-scope/spec.md#31-已锁定决策) 与 [product-scope §6.5 主流程 D](../product-scope/spec.md#65-主流程-d刚面完一轮的用户) 描述的"真实面试复盘"产品语义落到后端实现。它是 P0 用户路径中"刚面完一轮 → 选择上下文 → 文本 / 语音共享记录 → 复盘分析 → 复盘面试"闭环的核心后端服务：在 [backend-targetjob](../backend-targetjob/spec.md) 已经把 `targetjob.Drainer` 抽象出来承载 `target_import` / `source_refresh` / `privacy_delete` / `resume_parse` / `resume_tailor` / `debrief_generate` 等 job_type 的前提下，把 `debrief_generate` 这一行 queued job 推进到 `succeeded`（或永久 `failed`），落实 AI 后置分析、风险项识别、复盘面试上下文派生、`createDebrief` / `getDebrief` / `suggestDebriefQuestions` 三个 API 入口，以及 `debrief.created` / `debrief.completed` outbox。

[B2 OpenAPI v1](../openapi-v1-contract/spec.md) 已冻结本域承接的 2 个 operation：`POST /debriefs` (`createDebrief` 202 + `DebriefWithJob`，要求 `Idempotency-Key`) 与 `GET /debriefs/{debriefId}` (`getDebrief` 200 + `Debrief`)；本 spec 在 Phase 0 通过 B2 owner pre-launch additive 升级新增第 3 个 sync operation `suggestDebriefQuestions`，用于支撑前端文本模式 record 阶段的 AI 推荐问题（参见 D-6）。[B3 `event-and-outbox-contract`](../event-and-outbox-contract/spec.md) 已冻结 `debrief.created`（api producer，4 字段）+ `debrief.completed`（backend_async producer，4 字段含 `practiceFocusCount`）+ `debrief_generate` 异步 job（asynqTask `debrief.generate`、triggerEvent `debrief.created`、apiFacing、priority default、ownerDomain C9）；**注：当前 `shared/events.yaml` 中 `debrief.created.roundType` 字段引用 `$ref:b1.InterviewerRole` 与 B2 `CreateDebriefRequest.roundType` 独立 enum 漂移**，本 spec 在 Phase 0 通过 B3 owner pre-launch addendum 修订（D-14 决策细节）。[B4 `db-migrations-baseline`](../db-migrations-baseline/spec.md) 已落地 `debriefs` 单表（17 列含 `status` CHECK ∈ {`draft`,`completed`}、`round_type` CHECK 含 6 个 enum、`raw_questions` / `risk_items` / `next_round_checklist` jsonb、`thank_you_draft` text、`prompt_version` / `rubric_version` / `model_id` / `provider` provenance 4 列、`language` 默认 `en`、`user_id` ON DELETE CASCADE）+ 索引 `idx_debriefs_target_job_created`；`async_jobs.job_type` CHECK 已含 `debrief_generate`，`ai_task_runs.task_type` CHECK 当前已含 `debrief_generate` 但尚未含 `debrief_suggest_questions`；本 spec **不**新增任何 B4 列，但 Phase 0 必须通过 B4 owner pre-launch addendum 扩展 `migrations/000001_create_baseline.up.sql` 与 `migrations/enum-sources.yaml` 的 task_type 字面量，并补 migration lint / replay 证据。[B1 `shared-conventions-codified`](../shared-conventions-codified/spec.md) 已锁定 `DebriefStatus` enum = [`draft`,`completed`]（spec §5.12，仅 2 态）；本 spec 在 Phase 0 通过 B1 owner pre-launch addendum 新增 `DEBRIEF_NOT_FOUND` 错误码 + 通用 `IDEMPOTENCY_KEY_MISMATCH` 错误码 + `DebriefRoundType` enum（替代 events.yaml 中错误的 InterviewerRole 引用）+ `DebriefQuestionSource` enum（覆盖 `jd` / `resume` / `mock_report` / `manual`，用于 suggestDebriefQuestions 响应）；AI 失败只消费当前 B1 已登记的 canonical `AI_*` 错误码，不新增旧码别名。[F3 `prompt-rubric-registry`](../prompt-rubric-registry/spec.md) 已为本域分配 `debrief.generate` feature_key 与 `debrief.generate.default` model profile（基线 v0.1.0 active，001-baseline 已 completed）；本 spec 在 Phase 0 通过 F3 owner pre-launch addendum 新增 `debrief.suggest_questions` feature_key + `debrief.suggest_questions.default` model profile + 基线 prompt v0.1.0。

本 subject 不私自绕过 OpenAPI / DB / event / feature_key owner。任何契约修订必须先回到对应 owner spec 与编码真理源，再由本 subject 的 plan 实施；Phase 0 跨 owner pre-launch addendum 在 Phase 1 实施前必须以同一 PR 或紧邻 PR 落地，避免 backend-debrief 实现 reference 未注册字面量。

## 2 范围

### 2.1 In Scope

- 3 个 Debrief API operation 的 backend handler + service + store：
  - `POST /debriefs` `createDebrief`：用户提交真实面试问答记录；require `Idempotency-Key`；同事务写 `debriefs(status='draft', raw_questions=<DebriefQuestionInput[]>, notes, language, round_type, interviewer_role, target_job_id, user_id)` + `async_jobs(job_type='debrief_generate', status='queued', dedupe_key=debriefId, payload={debriefId, targetJobId, language, questionCount})` + outbox `debrief.created`（payload 含 debriefId / targetJobId / roundType / questionCount）；返回 `202 + DebriefWithJob{debriefId, job:Job}`。validation：`questions.length >= 1` 否则 `422 VALIDATION_FAILED`，单题 `questionText.length <= 4000` + `myAnswerSummary.length <= 4000` + `interviewerReaction.length <= 1000`，超长同样 `422 VALIDATION_FAILED`；越权 `target_job_id` 由 backend-targetjob middleware 拦截；本 spec 不重复校验 target_job 存在性。
  - `GET /debriefs/{debriefId}` `getDebrief`：返回 `Debrief` schema 全字段；按 `(user_id, id)` 过滤；越权 / 不存在均返回 `404 DEBRIEF_NOT_FOUND`（D-15 派生 B1/B2 前置）；`status='draft'` 时 `questions[*].aiAnalysis=null` + `riskItems=[]` + `nextRoundChecklist=[]` + `thankYouDraft=null` + `provenance=null`（前端按 D-3 通过 polling `getJob` 与 `getDebrief` 双轨感知生成进度，不依赖 placeholder semantic 字段）；`status='completed'` 时必须含 6 字段 `provenance` 与 enriched `questions[*].aiAnalysis` + `riskItems[*]`。
  - `POST /debriefs/question-suggestions` `suggestDebriefQuestions`（**Phase 0 通过 B2 owner pre-launch additive 升级新增**）：sync 200；request `{targetJobId, sessionId?, resumeVersionId?, language, count?(默认 6, 最大 10)}`；response `{suggestions: [{stage?, questionText, whyLikelyAsked, source: <DebriefQuestionSource>}]}`；handler 内部按 `(user_id, target_job_id)` 拉 target_job 摘要、可选 session 摘要、可选 resume version 摘要，组装 F3 prompt 上下文，调 `RegistryClient.Resolve("debrief.suggest_questions", language)` → `AIClient.Complete` → 解析 → 返回；失败映射必须使用当前 B1 canonical code：F3 resolve/config 失败 → `502 AI_PROVIDER_CONFIG_INVALID`，A3 secret missing → `502 AI_PROVIDER_SECRET_MISSING`，A3 timeout → `502 AI_PROVIDER_TIMEOUT`，fallback exhausted / provider unreachable → `503 AI_FALLBACK_EXHAUSTED`，AI invalid JSON / parsed empty → `502 AI_OUTPUT_INVALID`；不写 `debriefs` 行；不写 outbox；不要求 `Idempotency-Key`（结果是 ephemeral 推荐）。
- Debrief 生成 worker：消费 `async_jobs(job_type='debrief_generate', status='queued')` 行，复用 [backend-targetjob](../backend-targetjob/spec.md) 的 `targetjob.Drainer` + `targetjob.JobHandler` interface（D-5 决策）。
  - 注册：在 backend `cmd/api/main.go` （或等价 bootstrap）把新建的 `debrief.GenerateHandler` 注册到既有 drainer，job_type='debrief_generate'；不新建 polling worker / 不复制 Drainer。
  - 行租约：由 drainer 既有 `SELECT ... FOR UPDATE SKIP LOCKED` + UPDATE 推进 `status='queued' → 'running'`（drainer 已实现）；handler 接收 `JobPayload{debriefId,targetJobId,language,questionCount}`。
  - 在 handler 内：通过 `RegistryClient.Resolve("debrief.generate", language)` → `AIClient.Complete` 生成对每个 `raw_questions[i]` 的 `aiAnalysis`（轻量摘要 + 反馈，不复述原问/答）+ 整体 `risk_items`（label + severity ∈ {low,medium,high}）+ 可选 `practice_focus_signals`；P0 不生成 `next_round_checklist` / `thank_you_draft`（D-7 P1 deferred）。
  - 单事务持久化：UPDATE `debriefs SET status='completed', risk_items=<jsonb>, raw_questions=<jsonb with aiAnalysis injected>, prompt_version=<F3>, rubric_version=<F3>, model_id=<A3>, provider=<A3>, updated_at=now()` + outbox `debrief.completed`（payload 含 debriefId / targetJobId / riskItemCount=len(risk_items) / practiceFocusCount=len(risk_items)，D-8 决策语义）+ drainer 由 handler 返回 success 自动 UPDATE `async_jobs.status='succeeded'`（沿用 drainer 既有 finalize 逻辑）。
  - AI 失败 graceful：F3 ResolveActive 失败 / A3 secret missing / A3 timeout / A3 invalid JSON / parsed empty 等任一失败时 handler 返回 `targetjob.Drainer` 既有失败信号；drainer 自动按 backend-targetjob 既有 retry policy 推进 `async_jobs.attempts +=1, available_at=now()+backoff, status='queued'`；attempts >= 5 时 drainer 自动置 `async_jobs.status='failed'`；**`debriefs.status` 保持 `'draft'` 不变**（D-2 决策：debriefs 表只 draft→completed 单向），并 emit outbox `debrief.completed` **不被触发**；ai_task_runs 写一行 `status='failed'` + B1 error_code；handler 不返回 5xx 给客户端（worker 异步路径）。
  - Retry policy：完全沿用 backend-targetjob drainer 既有指数退避（默认 attempts 上限 5、`available_at=now()+min(2^attempts*30s, 30min)`）；本 spec 不重写 retry，只通过 drainer 配置消费。
- 状态机推进规则：`debriefs.status` 只允许 `draft → completed`（单向单 transition）；非法迁移在 store 层 `ErrIllegalTransition`；handler 路径 read-only。失败语义通过 `async_jobs.status` + getJob 表达，不通过 `debriefs.status='failed'`（D-2 决策）。
- AI 调用形态：业务侧只调用 [F3 `RegistryClient.Resolve`](../prompt-rubric-registry/spec.md) → 拿三元组 → 调用 [A3 `AIClient.Complete`](../ai-provider-and-model-routing/spec.md)；payload metadata 必须携带 `feature_key + prompt_version + rubric_version + model_profile_name + language + data_source_version`；`language` 来源：createDebrief / suggestDebriefQuestions 由 request body 携带；getDebrief 不消费 `Accept-Language` header 切换 wire；A3 observability decorator 自动写 `ai_task_runs(task_type='debrief_generate' 或 'debrief_suggest_questions', ...)` typed columns；不在 backend-debrief 私自调用 provider SDK 或 hardcode prompt / model 字符串。`Debrief.provenance` 必须使用 B2 `GenerationProvenance` wire 6 字段，runtime 字段（`feature_key` / `model_profile_name` / provider / cost / latency）只进入 `ai_task_runs` typed columns 与 audit 摘要。
- 用户隔离：所有 read / write SQL 必须按 `user_id` 过滤；越权访问 `getDebrief` 返回 HTTP 404 + B1 `DEBRIEF_NOT_FOUND`（D-15 派生），避免泄露存在性；`createDebrief` 由 auth middleware 注入 `user_id`，`target_job_id` 越权由 backend-targetjob middleware 拦截；`suggestDebriefQuestions` 同款 user-scoped 上下文拉取，cross-user target/session/resume 直接返回 `403 FORBIDDEN` 或不在 prompt context 中暴露；DELETE /me 通过 `users.id` ON DELETE CASCADE 级联清理 `debriefs`。
- 隐私 / 观测红线：`debrief.created` / `debrief.completed` 事件、F1 metric label、log、audit、`debriefs.raw_questions` / `risk_items` jsonb 列的 value 内容由用户输入 + AI 输出生成；spec 强制禁止 `notes` 全文、`raw_questions[].questionText` / `myAnswerSummary` / `interviewerReaction` 全文、`risk_items[].label` 全文、AI prompt body / response body / provider secret 出现在 metric label / log / audit metadata。允许的 audit / log 内容是 IDs、length、count、status、profile、provider、model_id、cost micros、error code 摘要；event payload 已 codified 为 counts only。具体 redaction 规则在 plan 内通过解析层与单元测试固化。
- F1 metric 注册边界：所有新增的 debrief metric / audit 类型必须先在 [F1 `observability-stack`](../observability-stack/spec.md) baseline 字典登记或由 F1 owner 承接；AI 调用 metric 复用 A3 已登记的 7 个 `ai_task_*` metric family，`feature_key` / `prompt_version` / `rubric_version` 只进入 `ai_task_runs` typed columns 或审计摘要，不进入 metric label。
- 失败语义与状态机退出：worker 路径所有 AI 失败按 graceful failed 处理，永不让 worker panic / crash 流程（与 backend-practice D-19 narrowed 一致原则）；DELETE /me / cascade → 物理删除。

### 2.2 Out of Scope

- 不实现 `listDebriefs` / `updateDebrief` / `deleteDebrief` / `regenerateDebrief` API：当前 B2 仅 2 个 operation + 本 spec Phase 0 新增 1 个 sync suggestion operation；未来 plan 002+ 决定是否扩 list / update。
- 不实现 `createPracticePlan` / `startPracticeSession` / 任何 Practice operation：归 [`backend-practice`](../backend-practice/spec.md) owner；本 spec 不实现"复盘面试"启动入口，复盘面试通过既有 `createPracticePlan(goal='debrief') + startPracticeSession` 跨域复用（D-17 决策），由前端 frontend-debrief 触发。
- 不实现 JD 解析、`target_jobs` 生命周期：归 [`backend-targetjob`](../backend-targetjob/spec.md) owner。本 spec 仅消费 `target_job_id`。
- 不实现简历资产、`resume_assets` 解析、岗位定制版本：归 [`backend-resume`](../backend-resume/spec.md) owner。
- 不实现 STT / LLM / TTS 编排、committed-context 推进、barge-in 处理：归 [`practice-voice-mvp`](../practice-voice-mvp/spec.md) owner；本 spec **不**为前端 voice debrief 提供 STT endpoint（voice 复盘后续 plan 与 practice-voice-mvp 整体语音上线后协调）。
- 不实现独立 worker / Asynq dispatcher / 生产级 outbox consumer / 集中式 retry orchestrator。`debrief_generate` 的运行已由 [`backend-async-runner/001`](../backend-async-runner/spec.md) 接管：`domaindebrief.NewGenerateHandler` 注册到单一 `runner.Runtime` kernel（经 `runner.FromTargetjobHandler` adapter），lease / retry / reaper 由 kernel 统一持有；本 spec 仅保留 handler 业务实现。
- 不暴露独立 `POST /debriefs/{debriefId}/regenerate` 或手工 retry API；`async_jobs.status='failed'` 的 debrief 只能等待 platform 触发或下一次 backend-debrief plan 决定开放手工入口。
- 不实现 P1 增强字段 `nextRoundChecklist` / `thankYouDraft` 的 worker 填充：B4 表中列已存在，OpenAPI 已标 P1 optional，本 spec P0 worker 留 `[]` / `null`（D-7 决策）；前端 P0 不渲染对应区块。
- 不实现 debrief 导出 / 分享 / 时间线：归 product-scope D-7 dashboard-only 边界 + D-12 P0 隐私 export 延后。
- 不实现 debrief 评分质量反馈、LLM judge、A/B grayscale：归 [`prompt-rubric-registry`](../prompt-rubric-registry/spec.md) future `005-grayscale-and-quality-feedback` plan。
- 不实现 voice-specific debrief 字段：P0 不区分 text / voice modality（modality 信息仅在 `provenance.dataSourceVersion` 中体现）；voice-specific 字段等 voice MVP 上线后再设计。
- 不在本 spec 文档内 inline 修改 B2 OpenAPI、B3 events/jobs、B4 baseline 表结构、A3 provider 协议或 F3 baseline prompt / rubric 文本；Phase 0 cross-owner pre-launch addendum 由各 owner spec / truth source / addendum plan 先落地，再由 backend-debrief implementation 消费（D-14 决策细节）。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | API 契约来源 | 本域只消费 [B2 OpenAPI](../openapi-v1-contract/spec.md) 已定义的 2 个 Debrief operation + Phase 0 新增 `suggestDebriefQuestions` sync operation；不私造 endpoint、不重写 schema；list/update/delete 不在 P0 | 任何字段 / 新 operation 先在 B2 spec / `openapi.yaml` 修订 |
| D-2 | 失败语义 | `debriefs.status` 仅 `draft → completed` 单向；AI 失败通过 `async_jobs.status='failed'` + `job.errorCode` 表达；debrief 行保持 `status='draft'` + `risk_items=[]` + `raw_questions[*].aiAnalysis=null` + `provenance=null` 不变；前端 polling `getJob(jobId)` 感知 succeeded/failed，再 polling `getDebrief(debriefId)` 拿 enriched data；无 `feedback_reports` 那样的 generating/ready/failed 多态 wire status | 与 B1 `DebriefStatus`=[draft,completed] 现状一致；无需新增 B4 列 / B1 enum 值 / B2 wire field |
| D-3 | 前端轮询契约 | createDebrief 返回 `DebriefWithJob{debriefId, job}` → 前端按 D-3 frontend-debrief spec 先轮询 `getJob(jobId)` 拿 status，job.status='succeeded' 再调 `getDebrief(debriefId)` 拿 enriched；job.status='failed' 渲染 frontend-debrief FailureState；job 永久失败（attempts>=5）时 frontend-debrief 显示重试 / 返回 workspace；不在 backend 引入新 polling endpoint | 复用 OpenAPI 既有 `GET /jobs/{jobId}` `getJob` operation；与 backend-targetjob drainer 既有 job lifecycle 完全一致 |
| D-4 | DB 真理源 | 复用 [B4 baseline](../db-migrations-baseline/spec.md) 的 `debriefs`（17 列已就位）/ `async_jobs` / `ai_task_runs` / `audit_events` / `outbox_events` 表与 CHECK 约束、UNIQUE 约束、外键；本 spec **不**新增 B4 列；Phase 0 B4 addendum 仅把 `debrief_suggest_questions` 加入 `ai_task_runs.task_type` CHECK / `enum-sources.yaml`；`raw_questions` jsonb 在 createDebrief 时存 `DebriefQuestionInput[]`，worker 完成后 in-place 注入 `aiAnalysis` 字段；`risk_items` jsonb 在 worker 完成后写入 | 不在本 spec 内 inline 落 migration；新增 task_type 必须先修订 B4 |
| D-5 | Worker 拓扑（已收干） | backend-debrief 实现 `targetjob.JobHandler` 业务逻辑（`debrief.GenerateHandler`）处理 `job_type='debrief_generate'`；运行边界已由 [`backend-async-runner/001`](../backend-async-runner/spec.md) 接管：handler 经 `runner.FromTargetjobHandler` adapter 注册到单一 `runner.Runtime` kernel，lease / retry / reaper / graceful shutdown 由 kernel 统一持有，不再实例化独立 drainer | 列名 `locked_at` / `attempts` 与 B4 baseline 一致；运行单元统一由 backend-async-runner kernel 治理 |
| D-6 | P0 AI 范围 | 2 个 AI 调用层级：(a) `suggestDebriefQuestions` sync handler 在前端文本模式 record 阶段调 F3 `debrief.suggest_questions` → A3 → 返回最多 10 条推荐问题；(b) post-create `debrief_generate` worker 调 F3 `debrief.generate` → A3 → 填充 `raw_questions[*].aiAnalysis` + `risk_items`；两个 feature_key 在 Phase 0 由 F3 owner pre-launch addendum 注册 baseline prompt v0.1.0 | 与 product-scope §9.2 "业务域不得 hardcode prompt 文本" 一致；与 backend-review 双 feature_key 模式一致 |
| D-7 | P1 字段处理 | OpenAPI `Debrief.nextRoundChecklist` / `thankYouDraft` 标记为 P1 optional/hidden；B4 表已含 `next_round_checklist` jsonb 与 `thank_you_draft` text 列；P0 worker **不**填充这两个字段，保持 `[]` / `null`；P0 前端不渲染对应区块；P1 plan 002 再启用 | 与 OpenAPI Debrief schema description "P1 optional/hidden" 一致；避免 P0 prompt rubric 复杂度 |
| D-8 | practiceFocusCount 语义 | `debrief.completed.payload.practiceFocusCount` = `len(risk_items)`（"下一轮需要重点准备的项数"）；与 B3 history note "P1 增强 mistakes_count 改为 practiceFocusCount"一致；不来自 `next_round_checklist`（P0 留空） | 与 B3 `shared/events.yaml` 既有 piiBoundary "Counts only" 一致 |
| D-9 | AI 失败 graceful | worker 路径所有 AI 失败 graceful：drainer 自动 attempts+1 + 指数退避 + permanent fail at attempts>=5；`debriefs.status` 保持 `'draft'`；`ai_task_runs` 写 failed row + B1 `AI_*` error_code；handler 不返回 5xx（worker 异步）；`suggestDebriefQuestions` sync handler 失败时返回 `502/503` + B1 error_code，让前端可降级到手工录入或显示 retry CTA | 异步流不阻塞用户；sync 流提供明确失败信号让前端处理 |
| D-10 | ai_task_runs 行 | A3 observability decorator 自动写 `ai_task_runs(task_type=<>, ...)` 行；`createDebrief` 不直接调 AI（worker 路径），不写 ai_task_runs；`suggestDebriefQuestions` sync 一行 `task_type='debrief_suggest_questions'`（Phase 0 B4 addendum 先登记）；worker `debrief_generate` 一行 `task_type='debrief_generate'`；行含 `feature_key` / `model_profile_name` / `input_tokens` / `output_tokens` / `latency_ms` / `validation_status` / `error_code`；失败行 `status='failed'` (B4 CHECK enum = `success`/`failed`/`timeout`/`fallback`) + B1 error_code | 失败可观测；与 backend-practice / backend-review 一致 |
| D-11 | provenance wire 边界 | `Debrief.provenance` 严格只暴露 B2 `GenerationProvenance` 6 wire 字段（`promptVersion` / `rubricVersion` / `modelId` / `language` / `featureFlag` / `dataSourceVersion`）；runtime 字段（`feature_key` / `model_profile_name` / provider / cost / latency）仅写入 `ai_task_runs` typed columns 与 audit 摘要。**持久化锚点**：6 字段映射到 B4 既有列：`prompt_version` / `rubric_version` / `model_id` / `language` / 新增 `feature_flag` 字段（plan 内决定走 jsonb 子键还是 audit-only，因为 B4 没有 `feature_flag` 列，本 spec **不**新增列）/ `data_source_version`（同样，plan 内决定 jsonb / audit-only）；getDebrief 不依赖 cross-table JOIN | 与 backend-review D-9 一致原则；不为 6 wire 字段强行扩 B4 列 |
| D-12 | 隐私红线 | `debrief.created` / `debrief.completed` 事件 payload / F1 metric label / log / audit metadata / 任何 outbox payload 不得包含 `raw_questions[].questionText` / `myAnswerSummary` / `interviewerReaction` 全文 / `notes` 全文 / `risk_items[].label` 全文 / AI prompt body / AI response body / provider secret 的逐字回放；允许出现的内容是 IDs、length、count、status、profile、provider、model_id、cost micros、error code 摘要；`debriefs.raw_questions` / `risk_items` jsonb 列可包含用户输入与 AI 摘要原文（这是 wire 表达的一部分），但**不得**作为 metric label / log / audit metadata 的 value 暴露；具体 redaction 规则在 plan 内由解析层 + 单元测试固化 | 与 product-scope §9.3 / B3 events.yaml piiBoundary / backend-practice D-11 / F1 一致 |
| D-13 | Outbox event payload schema | `debrief.created.payload` = `{debriefId, targetJobId, roundType: <B1 DebriefRoundType enum>, questionCount}`（与 B3 修订后 schema 完全一致）；`debrief.completed.payload` = `{debriefId, targetJobId, riskItemCount, practiceFocusCount}`（与 B3 当前 schema 一致）；payload 严格无 raw text；不出现 notes 字段 / question text / answer text；与 B3 piiBoundary "No debrief notes or transcript text" + "Counts only" 完全一致 | 与 B3 spec / `shared/events.yaml` 一致 |
| D-14 | Cross-owner pre-launch addendum 范围 | backend-debrief/001 Phase 0 必须以同一 PR（或紧邻 PR）落地以下跨 owner addendum，否则不得进入 Phase 1：(a) B1 新增 `DEBRIEF_NOT_FOUND` 错误码 + 通用 `IDEMPOTENCY_KEY_MISMATCH` 错误码 + `DebriefRoundType` enum + `DebriefQuestionSource` enum + 同步 generated Go/TS 字面量；AI 失败只使用当前 B1 canonical `AI_*` code；(b) B2 新增 `POST /debriefs/question-suggestions` `suggestDebriefQuestions` operation + `SuggestDebriefQuestionsRequest` / `SuggestDebriefQuestionsResponse` schema + fixtures `Debriefs/suggestDebriefQuestions.json` + 扩展既有 `createDebrief` / `getDebrief` fixtures + 修复 `Debrief.roundType` / `CreateDebriefRequest.roundType` enum 引用为 `b1.DebriefRoundType`（如 B1/B2 owner 选择新建独立 enum）；(c) B3 修复 `shared/events.yaml` `debrief.created.roundType: $ref:b1.InterviewerRole` → `$ref:b1.DebriefRoundType`，同步 generated `events.v1.json` + `events_inventory.py` lint；(d) B4 把 `debrief_suggest_questions` 加入 `ai_task_runs.task_type` CHECK、`migrations/enum-sources.yaml` 与 migration lint / replay fixtures，不新增列；(e) F3 新增 `debrief.suggest_questions` feature_key + `debrief.suggest_questions.default` model profile + 基线 prompt v0.1.0 + 可选 rubric；(f) 全套 `make codegen-check` + `make validate-fixtures` + `make lint-events` + `make codegen-events-check` + `migrations/lint.sh` + `make migrate-check` 通过；(g) backend-targetjob drainer 注册 `debrief.GenerateHandler` 的位置在 cmd/api 已有锚点验证 | 避免 backend-debrief 实施时 reference 未注册字面量 / 漂移 enum；显式承认本 spec 依赖 5 个 owner 的 pre-launch addendum |
| D-15 | Cross-user 隔离 + 错误码前置 | B1/B2 必须在 Phase 0 新增 `DEBRIEF_NOT_FOUND` 错误码并同步 generated Go/TS/OpenAPI 后，backend-debrief 才能在 404 隔离路径使用该 code；若 B1/B2 owner 选择复用既有 generic `RESOURCE_NOT_FOUND` 或类似 code，必须在 D-14 addendum 决策行更新本 spec | 与 backend-review D-15 / backend-practice 一致原则；保持错误码单一真理源 |
| D-16 | Idempotency 边界 | `createDebrief` require `Idempotency-Key`（OpenAPI 已声明，复用既有 `idempotency_records` 表 + handler middleware）；同 IK + 相同 user_id + 相同 request body hash → 返回 cached 202 response（同 debriefId + job）；同 IK + 不同 body → 409 `IDEMPOTENCY_KEY_MISMATCH`；`suggestDebriefQuestions` **不**要求 IK（结果 ephemeral，AI 输出非确定性，IK 反而误导客户端复用错误结果）；`getDebrief` read-only 无 IK | 与 OpenAPI parameters.IdempotencyKey 现状 + product-scope §4.4 idempotency 边界一致 |
| D-17 | 复盘面试 handoff | "复盘面试"通过既有 `createPracticePlan(goal='debrief')` + `startPracticeSession` 跨域复用；本 spec **不**实现 `startDebriefInterview` 或类似 endpoint；前端 frontend-debrief 在 step 2 只 nav 到 `practice`，route payload 含 `targetJobId` + `resumeVersionId` + `practiceGoal='debrief'` + 可选 `debriefId`，再由 frontend-workspace-and-practice 在 practice 路由内映射为 backend `createPracticePlan(goal='debrief')`；`practice_plans.goal='debrief'` 已是 B4 CHECK 现状（migration line 169）；`PracticeMode` 当前仅允许 `assisted` / `strict` 且与 goal 正交，debrief 不再是合法 mode；当前 backend-practice 实现是否已支持 `goal='debrief'` 派生与任一合法 `mode IN ('assisted','strict')` 的 session start 必须在 backend-debrief/001 Phase 0 验证（如未支持，需 backend-practice owner 同步 addendum） | 避免新增 endpoint；与 product-scope §6.5 "复盘面试是一场完整模拟面试" + backend-practice D-5/D-21 二值 mode 收敛一致 |
| D-18 | DELETE /me CASCADE | `users.id` ON DELETE CASCADE 已级联 `debriefs.user_id`（B4 现状）；删除 user 时 `debriefs` + `async_jobs(job_type='debrief_generate')` 通过外键级联清理（async_jobs 由 user_id 间接关联，具体级联由 backend-targetjob owner 验证）；本 spec P0 不实现独立 delete debrief API；不引入软删 | 与 product-scope §9.3 / backend-review D-7 一致 |

### 3.2 非后端 owner 决策

| ID | 事项 | Owner | 本域处理 |
|----|------|-------|----------|
| Q-1 | `b1.DebriefRoundType` enum 命名（`DebriefRoundType` vs 复用 `RoundType` 通用名） | B1 owner | 本 spec 默认建议 `DebriefRoundType` 与 `Debrief.roundType` 命名一致；若 B1 owner 选择 generic `RoundType` 跨域复用（practice_plans 等），同步更新 D-13 与 events.yaml；接受 owner 决议 |
| Q-2 | F3 `debrief.suggest_questions` rubric 是否必要 | F3 owner | 本 spec 不强制要求 rubric；prompt + structured output 已足够生成 6-10 条推荐问题；若 F3 owner 决定按现有 baseline 套路加 rubric，prompt 模板需考虑 schema 校验 |
| Q-3 | backend-practice 现状是否已支持 `goal='debrief'` plan 派生 + `PracticeMode IN ('assisted','strict')` session 启动 | backend-practice owner | 本 spec 仅假设 B1/B2/B4 `PracticeGoal` enum 已包含 `debrief`；`PracticeMode` 不含 `debrief`。实际 plan 派生逻辑 / session start handler 是否处理 goal=debrief 需 backend-practice owner 验证；如未实现，需 backend-practice 同步 addendum 或本 spec Phase 0 增加协调 |
| Q-4 | DELETE /me CASCADE 是否 atomic 跨表（debriefs + async_jobs + ai_task_runs + audit_events） | platform / future privacy plan | 本 spec 仅约定 `users.id` 外键 CASCADE 已落地（B4 现状），平台触发时跨表 atomic 由 future privacy plan 验证 |

### 3.3 待确认事项

| ID | 待确认事项 | 影响 | 默认处理 |
|----|------------|------|----------|
| Q-5 | suggestDebriefQuestions 是否需要 rate limit 或 quota | 防止前端死循环 / AI 成本失控 | 默认 P0 不做 rate limit；依赖前端 UX 流程（用户主动点击「生成推荐问题」按钮）；如生产观测发现滥用，plan 002 加 rate limit middleware |
| Q-6 | `provenance.dataSourceVersion` 应该包含哪些信号（debriefId @ schema-version？practice_session reference？） | debrief 可追溯性 vs 隐私 | 默认 `dataSourceVersion='debrief/<debriefId>@v1'`；plan 内固化；不暴露 sessionId / resumeVersionId 链路（避免推断用户其他资产） |

## 4 设计约束

### 4.1 API 契约约束

- 必须使用 [B2 generated `ServerInterface`](../openapi-v1-contract/spec.md) 注册 3 个 Debrief handler，不得绕过 generated types 自造 router。
- 入参反序列化必须使用 generated request types；响应必须使用 generated response types（`Debrief` / `DebriefWithJob` / `SuggestDebriefQuestionsResponse`）；fixture 与真实 handler 共用同一 schema。
- `createDebrief` 必须按 `(user_id, target_job_id)` 写入；越权 target_job 由 backend-targetjob middleware 拦截（404 / 403）；本 spec handler 仅 user_id 兜底检查；require `Idempotency-Key`（D-16）。
- `getDebrief` 必须按 `(user_id, debrief_id)` 过滤；越权 / 不存在均返回 `404 DEBRIEF_NOT_FOUND`（D-15 前置）；`status='draft'` 时返回 wire 一致的 partial（D-2），不返回 `REPORT_NOT_READY`-like 错误。
- `suggestDebriefQuestions` 必须按 `(user_id, target_job_id)` + 可选 `(user_id, session_id)` + 可选 `(user_id, resume_version_id)` 拉取上下文；越权 target / session / resume 走对应 owner middleware；本 spec handler 不做 cross-resource 存在性校验，但必须 SQL 过滤 user_id；不写 `debriefs` / outbox / `async_jobs`；写 `ai_task_runs` 一行；不要求 IK。
- 所有 read 路径无副作用；`createDebrief` 写路径必须先写 `idempotency_records` 行（中间件由 backend-auth + shared middleware 处理），随后 `debriefs` + `async_jobs` + `outbox_events` 必须在同一短事务内；F3 / A3 调用在事务外执行（worker 阶段）；`suggestDebriefQuestions` AI 调用同步在 handler 内（事务外），handler 用单独短事务写 `ai_task_runs`。

### 4.2 数据约束

- 所有 DB 写入必须在各自短事务内完成；外部 AI 调用不得包在 DB transaction 内。`debrief.created` / `debrief.completed` 必须通过 outbox 与对应业务写入同事务发出，避免双写。
- `debriefs.status` 仅允许 D-2 状态机 `draft → completed`；非法迁移在 store 层 ErrIllegalTransition；handler 路径 read-only。
- `debriefs.user_id` + `debriefs.target_job_id` 双外键，单 UNIQUE 索引仅 PK；不引入 `(user_id, target_job_id)` UNIQUE 约束（用户可对同一岗位创建多次复盘）。
- `async_jobs(debrief_generate).dedupe_key=debriefId`（一对一映射）；重复 INSERT 由 UNIQUE 兜底（B4 现状）。
- 软删 / 物理删除：v1 不引入软删；DELETE /me 通过 ON DELETE CASCADE 物理清理 `debriefs`。
- `raw_questions` jsonb 列在 createDebrief 时存 `DebriefQuestionInput[]`（用户输入），worker 完成后在原地 patch 注入 `aiAnalysis` 字段（保持顺序与索引稳定）；`risk_items` jsonb 列在 worker 完成后整体覆盖；不引入额外明细表。

### 4.3 安全 / 隐私约束

- 所有 API 走 [backend-auth](../backend-auth/spec.md) session middleware；未认证返回 `401 AUTH_UNAUTHORIZED`；越权返回 404 + `DEBRIEF_NOT_FOUND`（D-15 前置），不泄露资源存在性；`suggestDebriefQuestions` 同款 auth。
- AI 调用 fail-closed 边界：worker 路径无用户阻塞调用，所有 AI 失败按 graceful failed 处理；不静默回退 stub provider（除 `APP_ENV=test`）；F3 / A3 缺 secret 或 provider unreachable 同样在 worker 路径走 drainer retry / permanent fail；`suggestDebriefQuestions` sync 路径返回 `502/503 + AI_*` 错误码，让前端降级。
- log / metric label / audit / 事件 payload 不得包含 raw user input / AI generated prose / provider secret；只允许 IDs、length、count、status、profile、provider、model_id、cost micros、error code 摘要。`debriefs` jsonb 列允许 wire 内容暴露给用户本人（getDebrief 响应），但不进入 metric label。
- `audit_events` 触发集合：`createDebrief` 写一行 audit（draft 创建）；worker 推进到 `completed` 时写一行 audit；`suggestDebriefQuestions` 写一行 audit（含 suggestion count）；`getDebrief` read handler 不写 audit（高频）。audit metadata 仅含 `debrief_id` / `target_job_id` / `status` / `language` / `error_code` / `suggestion_count`，不含问答文本。

### 4.4 异步 / 可观测约束

- 本域 `debrief.GenerateHandler` 消费 `async_jobs(debrief_generate)` queued 行的运行边界已由 [`backend-async-runner/001`](../backend-async-runner/spec.md) kernel 接管：handler 注册到单一 `runner.Runtime`，与 backend-practice / backend-review / privacy 共存于同一 backend 进程；`debrief.created` 为 `trigger_creates_job`，job 由业务事务创建，dispatcher 不重复消费。
- F1 metric 字典登记前置：debrief-specific business metrics（如 debrief created count / generation duration / failed count by error_code / suggestion request count）实施前必须先在 [F1 baseline](../observability-stack/spec.md) 字典登记，并且 label 只能使用 F1 allowed labels 与 B1 有界枚举；AI 调用 metric 复用 A3 已登记的 7 个 `ai_task_*` metric family。
- worker P95 延迟（从 lease 到 completed）作为观测目标登记，但不作为本 spec 验收 gate；`suggestDebriefQuestions` P95 延迟（用户感知）作为 UX 验收的软目标。

### 4.5 文档治理约束

- 本 spec 后续修订必须原地更新；不允许创建同主题 sibling spec。
- 涉及 OpenAPI / events / migrations / shared enums / runtime config 的修改必须先回 owner spec / `*.yaml` truth source。
- 涉及用户行为流的 plan 必须维护 BDD gate；本域 3 个 API operation 属于用户可见 API 行为；worker 触发后端 outbox / DB state 推进，前端可通过 `getDebrief` + `getJob` 感知，归属用户行为流。
- 命中 `completed` plan 时不创建同主题 sibling follow-up plan，原地修订即可。
- 旧术语（`mistakes_count` / `generatedMistakeCount` / 独立 `mistakes` 错题本 / `drill` / `growth_center` / `experience_library` / `star_editor` / 独立 voice route / 报告时间线 / 多形态 review） 不得在本 spec 与衍生 plan 中出现，违反必须先修订 [product-scope](../product-scope/spec.md)。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| API contract | [B2 `openapi-v1-contract`](../openapi-v1-contract/spec.md) | `createDebrief` / `getDebrief` / `suggestDebriefQuestions`（Phase 0 新增）operation 与 `Debrief` / `DebriefWithJob` / `CreateDebriefRequest` / `DebriefQuestionInput` / `DebriefQuestion` / `DebriefRiskItem` / `DebriefStatus` / `SuggestDebriefQuestionsRequest` / `SuggestDebriefQuestionsResponse`（Phase 0 新增）/ `GenerationProvenance` schema、fixtures、generated client / server |
| `CountDebriefsForUser(ctx, db, userID) (int, error)` cross-owner internal API | backend-debrief | backend-jobs-recommendations/001 BuildJobMatchProfile aggregation (D-18 sources.debriefs)；read-only；cross-user 隔离；不写 audit。实现：`backend/internal/debrief/count.go` |
| Backend domain | `backend-debrief`（本 spec） | 3 个 handler + service + store + drainer-registered worker handler + AI 调用编排 + AI 推荐问题 + status 状态机 + outbox emit |
| DB schema | [B4 `db-migrations-baseline`](../db-migrations-baseline/spec.md) | `debriefs` / `async_jobs` / `ai_task_runs` / `audit_events` / `outbox_events` / `idempotency_records` 列与索引；shared `idempotency_records` 表（createDebrief 使用）；Phase 0 仅扩展 `ai_task_runs.task_type` 字面量 `debrief_suggest_questions` |
| Event / job contract | [B3 `event-and-outbox-contract`](../event-and-outbox-contract/spec.md) | `debrief.created` / `debrief.completed` 事件；`debrief_generate` job mapping 与 `triggerEvent: debrief.created`；Phase 0 修订 `debrief.created.roundType` 引用 |
| Shared enums / errors | [B1 `shared-conventions-codified`](../shared-conventions-codified/spec.md) | `DebriefStatus`（既有）+ Phase 0 新增 `DebriefRoundType` / `DebriefQuestionSource` enum + `DEBRIEF_NOT_FOUND` / `IDEMPOTENCY_KEY_MISMATCH` error code；AI 失败只消费既有 canonical `AI_*` code |
| AI provider | [A3 `ai-provider-and-model-routing`](../ai-provider-and-model-routing/spec.md) | `AIClient.Complete`、provider registry、model profile、observability decorator 写 `ai_task_runs` |
| Prompt / rubric | [F3 `prompt-rubric-registry`](../prompt-rubric-registry/spec.md) | `debrief.generate`（既有）+ Phase 0 新增 `debrief.suggest_questions` feature_key、Resolve 实现、baseline prompt / rubric |
| Config / secret | [A4 `secrets-and-config`](../secrets-and-config/spec.md) | provider secret、feature flag、debrief worker handler 配置（如有） |
| Observability | [F1 `observability-stack`](../observability-stack/spec.md) | metric / audit 类型登记、label allowlist、隐私红线 |
| Auth / isolation | [`backend-auth`](../backend-auth/spec.md) | session middleware、user-scoped read/write、DELETE /me CASCADE 协议 |
| Async runtime | [`backend-targetjob`](../backend-targetjob/spec.md) | `targetjob.Drainer` + `targetjob.JobHandler` interface；本 spec 注册 `debrief.GenerateHandler` 不新建 polling worker |
| Upstream — TargetJob | [`backend-targetjob`](../backend-targetjob/spec.md) | 提供 `target_jobs` 行、middleware 越权拦截、`getTargetJob` API 给 suggestDebriefQuestions 使用 |
| Upstream — Resume | [`backend-resume`](../backend-resume/spec.md) | 提供 resume version 行、middleware 越权拦截；suggestDebriefQuestions 可选消费 resume bullets 作为 prompt context |
| Upstream — Practice | [`backend-practice`](../backend-practice/spec.md) | 提供 practice_sessions 行；suggestDebriefQuestions 可选消费 session 摘要作为 prompt context；复盘面试启动通过 `createPracticePlan(goal='debrief')` + `startPracticeSession` 跨域复用，由 backend-practice owner 处理 `goal='debrief'` plan 派生与合法 `mode IN ('assisted','strict')` session start（Q-3 验证） |
| Downstream — Debrief UI | [`frontend-debrief`](../frontend-debrief/spec.md) | `DebriefScreen` / 3 picker modal / record / analysis / interview step、polling、复盘面试 CTA；本 spec 提供 schema + data，不耦合 UI |
| Frontend consumer | [`frontend-debrief`](../frontend-debrief/spec.md) | `createDebrief` / `getDebrief` / `suggestDebriefQuestions` 调用入口在 DebriefScreen 各 step |
| Scenario coverage | scenarios owner + 本 subject | `E2E.P0.060-064` 套件 setup / trigger / verify / cleanup（具体编号在 plan 内分配） |
| Async runner replacement | future `backend-async-runner` | 接管 runtime drainer / 多 worker；必须沿用 backend-targetjob D-* forward-binding 与 backend-review D-13 inline runner 边界 |

### 5.1 Operation Matrix

| `operationId` | Fixture | Frontend consumer | Backend handler | Persistence | AI dependency | Scenario coverage |
|---------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `createDebrief` | `openapi/fixtures/Debriefs/createDebrief.json`（既有文件，Phase 0 扩展：`default` = 202 + DebriefWithJob queued + idempotency example） | DebriefScreen step 0 submit 按钮 | Plan 001 Phase 2：`backend/internal/api/debriefs.CreateDebrief` + `backend/internal/debrief.Service.CreateDebrief` + `backend/internal/store/debrief.CreateDebrief`（user-scoped + IK） | `debriefs` write + `async_jobs(debrief_generate)` write + `outbox_events(debrief.created)` write + `idempotency_records` write | none in handler path | `E2E.P0.060` |
| `getDebrief` | `openapi/fixtures/Debriefs/getDebrief.json`（既有文件，Phase 0 扩展：`default` = completed 完整字段 / `debrief-draft` = draft + 空 risk_items / `prototype-baseline`） | DebriefScreen step 1 (analysis) 拉取数据；前端先轮询 getJob 再调 getDebrief | Plan 001 Phase 5：`backend/internal/api/debriefs.GetDebrief` + `backend/internal/debrief.Service.GetDebrief` + `backend/internal/store/debrief.GetDebrief`（user-scoped） | `debriefs` read | none in handler path | `E2E.P0.061` |
| `suggestDebriefQuestions` | `openapi/fixtures/Debriefs/suggestDebriefQuestions.json`（Phase 0 新增：`default` = 6 条 suggestions / `empty` = 0 条 / `prototype-baseline`） | DebriefScreen step 0 text mode "生成推荐问题" 按钮 | Plan 001 Phase 3：`backend/internal/api/debriefs.SuggestDebriefQuestions` + `backend/internal/debrief.Service.SuggestQuestions` + sync AI 调用 | `ai_task_runs` write + `audit_events` write；不写 `debriefs` / `async_jobs` / outbox | F3 `debrief.suggest_questions` v0.1.0；A3 `AIClient.Complete` × 1 | `E2E.P0.063` |
| `(worker: debrief_generate job)` | N/A | N/A（异步 worker，不暴露 API） | Plan 001 Phase 4：`backend/internal/debrief.GenerateHandler` 实现 `targetjob.JobHandler` interface + `backend/internal/debrief.GenerateService` + `backend/internal/store/debrief.UpdateDebriefCompleted` | `debriefs` update（draft→completed + raw_questions in-place + risk_items）+ `async_jobs` finalize（由 drainer）+ `outbox_events(debrief.completed)` write + `ai_task_runs` write | F3 `debrief.generate` v0.1.0；A3 `AIClient.Complete` × 1 | `E2E.P0.060` + `E2E.P0.062` + `E2E.P0.064` |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | createDebrief 主路径（202 + IK） | 用户已认证；target_job 属于用户；Phase 0 cross-owner addendums 已落地；F3 / A3 active | `POST /debriefs` with `Idempotency-Key=X` + 合法 questions[] | 返回 `202 + DebriefWithJob{debriefId, job:{jobType:'debrief_generate', status:'queued'}}`；同事务写 `debriefs(status='draft', raw_questions=[...], user_id, target_job_id)` + `async_jobs(debrief_generate, queued)` + outbox `debrief.created` | 001 |
| C-2 | createDebrief IK replay | C-1 已成功 | 重发同 IK + 同 body | 返回相同 202 + 相同 debriefId + 相同 jobId（cached）；不重复写 `debriefs` / `async_jobs` / outbox | 001 |
| C-3 | createDebrief IK mismatch | C-1 已成功 | 重发同 IK + 不同 body hash | 返回 `409 IDEMPOTENCY_KEY_MISMATCH` | 001 |
| C-4 | createDebrief 输入验证 | 用户已认证 | submit empty `questions=[]` 或单题 `questionText.length > 4000` | 返回 `422 VALIDATION_FAILED`；不写 `debriefs` / `async_jobs` / outbox | 001 |
| C-5 | debrief_generate worker happy（draft → completed） | C-1 已创建 queued job；F3 `debrief.generate` baseline active；A3 active | drainer lease 该 job → 调用 handler | `debriefs.status='draft' → 'completed'`；`raw_questions[*].aiAnalysis` in-place 注入；`risk_items` 非空 jsonb；`provenance` 6 字段填充（prompt_version / rubric_version / model_id 来自 debriefs 4 列 + language / featureFlag / dataSourceVersion 来自 plan 决策的位置）；`async_jobs.status='succeeded'`；outbox `debrief.completed{debriefId,targetJobId,riskItemCount=N,practiceFocusCount=N}` 发出 1 行 | 001 |
| C-6 | getDebrief draft 占位返回 | C-1 已创建但 worker 未完成 | `GET /debriefs/{debriefId}` | `200 + Debrief{status:'draft', questions:[{questionText,myAnswerSummary, aiAnalysis:null, ...}], riskItems:[], nextRoundChecklist:[], thankYouDraft:null, provenance:null}`；前端可据 status 决定是否轮询 getJob | 001 |
| C-7 | getDebrief completed 完整返回 | C-5 已完成 | `GET /debriefs/{debriefId}` | `200 + Debrief{status:'completed', questions:[{...,aiAnalysis:<text>}], riskItems:[...], provenance:{6 fields}, ...}` | 001 |
| C-8 | Cross-user 404 隔离 | 用户 A 持有 debriefX | 用户 B 调 `GET /debriefs/X` | `404 DEBRIEF_NOT_FOUND`（D-15 派生）；不泄露存在性；handler 内 user_id 过滤兜底 | 001 |
| C-9 | suggestDebriefQuestions 主路径 | 用户已认证；target_job 属于用户；可选 sessionId / resumeVersionId 属于用户；F3 / A3 active | `POST /debriefs/question-suggestions` with `{targetJobId, language:'zh', count:6}` | 返回 `200 + SuggestDebriefQuestionsResponse{suggestions:[6 items {questionText, whyLikelyAsked, source}]}`；`ai_task_runs` 写一行 `task_type='debrief_suggest_questions', status='success'`；audit 一行 | 001 |
| C-10 | suggestDebriefQuestions AI 失败 | C-9 前提；F3 ResolveActive 失败 或 A3 timeout / invalid output | `POST /debriefs/question-suggestions` | 按 B1 canonical 映射返回 `AI_PROVIDER_CONFIG_INVALID` / `AI_PROVIDER_TIMEOUT` / `AI_OUTPUT_INVALID` / `AI_FALLBACK_EXHAUSTED` / `AI_PROVIDER_SECRET_MISSING`；`ai_task_runs` 写一行 `status='failed'` + B1 error_code；audit 一行 with error_code；前端可降级到手工录入 | 001 |
| C-11 | worker AI failure graceful + retry | C-1 已创建 queued job；F3 ResolveActive 失败 或 A3 timeout 或 parsed empty | drainer lease + handler 调用失败 | `debriefs.status` 保持 `'draft'`（D-2）；`async_jobs.attempts +=1, available_at=now()+backoff, status='queued'`（attempts < 5）或 `status='failed'`（attempts >= 5）+ `locked_at=null`；outbox `debrief.completed` **不发出**；`ai_task_runs` 写一行 `status='failed'` + B1 error_code；handler 不返回 5xx | 001 |
| C-12 | Worker permanent fail (attempts=5) | C-11 已重试 4 次失败 | 第 5 次失败 | `async_jobs.status='failed'`（permanent）+ `locked_at=null`；`debriefs.status` 保持 `'draft'`；前端通过 `getJob.status='failed' + errorCode` 感知；outbox 不发出 | 001 |
| C-13 | Debrief.provenance wire 完整 | C-5 已完成 | `GET /debriefs/{debriefId}` | `provenance` 仅含 B2 wire 6 字段：`promptVersion` / `rubricVersion` / `modelId` / `language` / `featureFlag` / `dataSourceVersion`；`feature_key` / `model_profile_name` / provider / cost / latency 等运行时字段不在 wire JSON 出现；status='draft' 时 provenance=null | 001 |
| C-14 | 隐私红线 | C-5 已完成 | 检查 `debrief.created` / `debrief.completed` outbox payload / F1 metric label / log / audit metadata | 不含 `questionText` / `myAnswerSummary` / `interviewerReaction` / `notes` 全文 / `risk_items.label` 全文 / AI prompt body / AI response body / provider secret 的逐字回放；只含 IDs / length / count / status / profile / provider / model_id / cost micros / error code 摘要 | 001 |
| C-15 | ai_task_runs 行 | C-5 / C-9 / C-10 / C-11 完成 | 检查 `ai_task_runs` | `debrief_generate` 调用一行 `task_type='debrief_generate'`；`suggestDebriefQuestions` 调用一行 `task_type='debrief_suggest_questions'`；行含 `feature_key` / `model_profile_name` / `input_tokens` / `output_tokens` / `latency_ms` / `validation_status` / `error_code`；失败行 `status='failed'`（B4 enum）+ B1 error_code | 001 |
| C-16 | Cross-owner pre-launch addendum gate | Phase 0 addendums 未落地 | 进入 Phase 1 实施 | Phase 1 实施必须 BLOCK：`make codegen-check` / `make validate-fixtures` / `make lint-events` / `migrations/lint.sh` / `make migrate-check` 任一失败时拒绝继续；D-14 addendums (a)-(g) 全部通过 | 001 |
| C-17 | 复盘面试 handoff 验证 | C-7 已完成；用户在 frontend-debrief step 2 触发"开始复盘面试" | 前端 nav practice + route payload `{practiceGoal:'debrief', targetJobId, resumeVersionId?, debriefId?}` | frontend-workspace-and-practice 把 route-level `practiceGoal` 映射为 backend-practice `createPracticePlan(goal='debrief')`；`startPracticeSession` 接受合法 `mode IN ('assisted','strict')`；本 spec 不实现这些；C-17 验证 backend-practice 现状已支持（如未支持，Q-3 升级为 D-19 强制 backend-practice addendum） | 001 |

## 7 关联计划

`001-debrief-record-and-analysis` 已派 plan（spec v1.0 同会话），其余 plan 按 phase closability 与 owner 边界依次派生。全局前置：依赖 [`prompt-rubric-registry/001-baseline`](../prompt-rubric-registry/plans/001-baseline/plan.md) 已 completed（`debrief.generate` v0.1.0 已 active）+ Phase 0 F3 owner addendum 完成 `debrief.suggest_questions` v0.1.0；依赖 backend-targetjob 既有 drainer abstraction；依赖 backend-practice plan 004 或等价 addendum 支持 `goal='debrief'` 与合法 `mode IN ('assisted','strict')`（Q-3 验证）。

1. [`001-debrief-record-and-analysis`](./plans/001-debrief-record-and-analysis/plan.md)：D-1 ~ D-18 全部决策落地；Phase 0 cross-owner pre-launch addendums (B1/B2/B3/B4/F3 + backend-practice 验证) + Phase 1 createDebrief handler + Phase 2 IK + validation + Phase 3 suggestDebriefQuestions sync handler + Phase 4 debrief_generate worker handler + AI 调用 + 持久化 + outbox + Phase 5 getDebrief read handler + 隔离 + Phase 6 失败语义 / retry / 隐私 / observability / legacy negative。
2. 保留编号建议 `002-debrief-listing-and-update`：`listDebriefs` API（如产品决定开启复盘历史浏览）+ `updateDebrief` API（如产品决定支持原地修订记录）+ rate limit on suggestDebriefQuestions（如 Q-5 升级）。
3. 保留编号建议 `003-debrief-voice-and-stt-integration`：voice debrief 真实 STT 接入（依赖 [practice-voice-mvp](../practice-voice-mvp/spec.md) 整体语音上线）+ debrief voice extraction worker + 隐私链路（音频 / transcript 留存与删除）。
4. 保留编号建议 `004-debrief-retention-and-cascade`：DELETE /me CASCADE 验证 + retention policy（超期 debrief 删除）+ 隐私 export 占位（与 product-scope Q-5 一致）。

每个 plan 通过 `/design` 落地时单独配 BDD/test plan；本 spec §6 AC 是这些 plan 的统一来源。

## 8 关联文档

- [Product Scope §6.5 主流程 D：刚面完一轮的用户](../product-scope/spec.md#65-主流程-d刚面完一轮的用户)
- [Product Scope §6.11 M4 扩展：真实面试复盘](../product-scope/spec.md#611-m4-扩展真实面试复盘)
- [Product Scope §4.1 产品原则](../product-scope/spec.md#41-产品原则)（真实复盘独立成流）
- [docs/ui-design/review-module.md](../../ui-design/review-module.md)
- [docs/ui-design/module-map.md](../../ui-design/module-map.md)
- [openapi-v1-contract](../openapi-v1-contract/spec.md)
- [event-and-outbox-contract](../event-and-outbox-contract/spec.md)
- [db-migrations-baseline](../db-migrations-baseline/spec.md)
- [shared-conventions-codified](../shared-conventions-codified/spec.md)
- [ai-provider-and-model-routing](../ai-provider-and-model-routing/spec.md)
- [prompt-rubric-registry](../prompt-rubric-registry/spec.md)
- [secrets-and-config](../secrets-and-config/spec.md)
- [observability-stack](../observability-stack/spec.md)
- [backend-auth](../backend-auth/spec.md)（auth/session middleware）
- [backend-targetjob](../backend-targetjob/spec.md)（drainer + JobHandler interface 复用）
- [backend-practice](../backend-practice/spec.md)（`goal='debrief'` plan 派生 + 合法 `mode IN ('assisted','strict')` session start）
- [practice-voice-mvp](../practice-voice-mvp/spec.md)（voice debrief 后续协调）
- [frontend-debrief](../frontend-debrief/spec.md)（下游 UI consumer）
- [docs/development.md §2 Frontend / Backend Contract Workflow](../../development.md)
- 历史：[history.md](./history.md)
