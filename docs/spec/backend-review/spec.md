# Backend Review Spec

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-15

## 1 背景与目标

`backend-review` 承接 [engineering-roadmap §5.2](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 中 `Report Dashboard` workstream 的后端域。它是 P0 用户路径中"练习完成 → 报告生成 → 复练 / 下一轮 / 真实复盘"闭环的最后一段后端实现：在 [backend-practice](../backend-practice/spec.md) 已经把 `completePracticeSession` 同事务创建好 `feedback_reports(status='queued')` placeholder + `async_jobs(job_type='report_generate', status='queued')` + `practice.session.completed` outbox 的前提下，把那一行 queued report 推进到 `ready`（或 `failed`），落实证据化报告、维度评估、ReadinessTier、retry-focus turn 选择、next_action 决策、`getFeedbackReport` / `listTargetJobReports` 读取入口与 `report.generated` / `report.generation.failed` outbox。

[B2 OpenAPI v1](../openapi-v1-contract/spec.md) 已冻结本域承接的 2 个读取 operation：`getFeedbackReport` / `listTargetJobReports`；写入路径上的 `completePracticeSession` 由 backend-practice owner 维护，本 spec 不重新声明该 operation。[B3 `event-and-outbox-contract`](../event-and-outbox-contract/spec.md) 已冻结 `report.generation.requested` / `report.generated` / `report.generation.failed` 三个事件 v1，以及 `report_generate` 异步 job 的 `triggerEvent: practice.session.completed`（D-32 落实 `triggerEventSemantic: source_event_only`，即由 backend-practice 在 complete 同事务创建 queued job、外部 dispatcher 不再二次创建）。[B4 `db-migrations-baseline`](../db-migrations-baseline/spec.md) 已落地 `feedback_reports`（当前 baseline 17 列：含 highlights / issues / next_actions jsonb、preparedness_level CHECK、`prompt_version` / `rubric_version` / `model_id` / `provider` provenance 列、`error_code`、`generated_at`、UNIQUE(session_id)；wire `GenerationProvenance` 必需的 `language` / `featureFlag` / `dataSourceVersion` 三字段对应的列与 `retry_focus_turn_ids` jsonb 列**不在当前 baseline**，由 plan 001 Phase 0 通过 pre-launch baseline rebase 在 B4 同 commit 新增 4 列）、`question_assessments`（dimension_results jsonb + review_status CHECK + UNIQUE(report_id, turn_id)）、`ai_task_runs.task_type` CHECK 已含 `report_generate`、`async_jobs.job_type` CHECK 已含 `report_generate`、`async_jobs` 列名为 `attempts` / `locked_at`（无 `worker_id` 列）、`ai_task_runs.status` CHECK 枚举为 `success` / `failed` / `timeout` / `fallback`（注意：B4 `ai_task_runs.status` 用 `success` 而非 `succeeded`；`async_jobs.status` 用 `succeeded`，与 B4 baseline CHECK 一致）。[B1 `shared-conventions-codified`](../shared-conventions-codified/spec.md) 已锁定 `ReportStatus` / `ReadinessTier` / `DimensionStatus` / `Confidence` / `QuestionReviewStatus` 五个枚举、`feedback_report` resource、`report_generate` job、`REPORT_NOT_READY` 错误码。[F3 `prompt-rubric-registry`](../prompt-rubric-registry/spec.md) 已为本域分配 `report.generate` / `report.question_assessment` 两个 `feature_key` 与 `report.generate.default` / `report.assessment.default` model profile（含 multi + en + zh baseline prompt 与 rubric v0.1.0 active）。本 subject 把这些契约、表结构、事件与 feature_key 缝合成一个独立的 P0 后端域，并明确仍需在 implementation Phase 0 收口的编码真理源前置：runtime config 中 review worker poll 周期与 lease timeout 的 owner 边界、`feedback_reports.error_code` 与 B1 enum 的对齐、retry policy 上限的运维参数登记、B4 列名 `locked_at` / `attempts` 在 plan 与 SQL 中统一引用，以及 F3 prompt / rubric 当前编码真理源是否满足本 spec 的无 raw question / answer / transcript 输入与无 verbatim quote 输出红线；若 F3 preflight 发现 prompt 仍要求 `{{question}}` / `{{answer}}` / `{{transcript}}` 或 verbatim evidence quotes，implementation 必须先回到 F3 owner 修订，不得进入 review worker AI 路径。

本 subject 不私自绕过 OpenAPI / DB / event / feature_key owner。任何契约修订必须先回到对应 owner spec 与编码真理源，再由本 subject 的 plan 实施；未完成 F3 baseline `report.generate` / `report.question_assessment` 的可用性验证前，不得进入依赖该 AI 输出的 worker 路径实现（与 backend-practice D-29 一致原则）。

## 2 范围

### 2.1 In Scope

- 2 个 Report 读取 operation 的 backend handler + service + store：
  - `GET /reports/{reportId}` `getFeedbackReport`：返回完整 `FeedbackReport`；按 `user_id` 过滤；越权 / 不存在均返回 `404 REPORT_NOT_FOUND`（本 spec D-15 派生 B1/B2 前置）；status ∈ {`queued`,`generating`} 时返回 `200 + FeedbackReport{status,preparednessLevel:null,highlights:[],issues:[],nextActions:[],questionAssessments:[],provenance:null}` placeholder；status='ready' 时填充全部字段且 `provenance` 必须非空；status='failed' 时 `error_code` 必填、其它内容字段为空数组、`provenance` 可空。
  - `GET /targets/{targetJobId}/reports` `listTargetJobReports`：返回 `PaginatedFeedbackReport`，按 `target_job_id` + `user_id` 过滤；cursor 分页（`pageSize` 默认 20，最大 50）；按 `created_at DESC` 排序；空列表返回 `{items:[], pageInfo:{nextCursor:null,pageSize,hasMore:false}}`；不暴露内部 `error_code` 摘要给跨用户。
- Report 生成 worker：消费 `async_jobs(job_type='report_generate', status='queued')` 行，遵循 D-32 forward-binding 语义不被外部 dispatcher 重复创建。
  - 行租约：`SELECT * FROM async_jobs WHERE job_type='report_generate' AND status='queued' AND available_at <= now() ORDER BY available_at ASC LIMIT 1 FOR UPDATE SKIP LOCKED` 取一行 + UPDATE 为 `status='running', attempts=attempts+1, locked_at=now()`（B4 baseline 实际列名；worker 身份只进 logger / metric label，不写 DB）。
  - 推进 `feedback_reports.status='queued' → 'generating'`（与 job lease 同事务）；写入 generating 时间标记。
  - 通过 [F3 `RegistryClient.Resolve("report.generate", language)`](../prompt-rubric-registry/spec.md) → [A3 `AIClient.Complete`](../ai-provider-and-model-routing/spec.md) 生成 highlights / issues / next_actions / preparedness_level signal 草稿（详细 schema 解析在 plan 内落实）。
  - 通过 F3 `RegistryClient.Resolve("report.question_assessment", language)` → A3 `AIClient.Complete` 对每个 `practice_turns` 行（按 `turn_index` 升序）生成 `QuestionAssessment` 字段（dimension_results map + overall_status + confidence + strengths / gaps / recommended_framework）。
  - 计算 `preparedness_level` (`ReadinessTier`)：基于 `question_assessments.dimension_results` 加权汇总到 ReadinessTier 四档（D-4 决策细节）。
  - 计算 `retry_focus_turns`：选出 `overall_status='needs_work'` 或 `review_status='queued_for_retry'` 的 turn 集合（D-5 决策细节）；写入 `feedback_reports.question_assessments[].included_in_retry_plan=true`。
  - 计算 `next_action` 推荐：基于 `preparedness_level` × `retry_focus_turns.length` × 是否首轮决策 `retry_current_round / next_round / review_evidence`（D-6 决策细节），落入 `feedback_reports.next_actions` 中第一行 `type` 字段。
  - 单事务持久化 `feedback_reports`（status='generating' → 'ready'，含新增列 `language` / `feature_flag` / `data_source_version` / `retry_focus_turn_ids` 写值）+ N 行 `question_assessments`（关联 report_id + turn_id）+ outbox `report.generated`（payload 含 reportId / sessionId / targetJobId / preparednessLevel / questionIssueCount / promptVersion / rubricVersion / modelId）+ UPDATE `async_jobs.status='succeeded', completed_at=now(), locked_at=null`（B4 baseline `async_jobs.status` enum 用 `'succeeded'`，与 drainer `FinalizeAsyncJob` 一致）。
  - AI 失败 graceful：F3 ResolveActive 失败 / A3 secret missing / A3 timeout / A3 invalid JSON / parsed empty 等映射到 B1 `AI_*` 错误码；UPDATE `feedback_reports.status='failed', error_code=<B1 code>, generated_at=now()`；UPDATE `async_jobs SET available_at=now() + backoff(attempts), status=CASE WHEN attempts >= 5 THEN 'failed' ELSE 'queued' END, locked_at=null`；发出 outbox `report.generation.failed`（payload 含 reportId / sessionId / errorCode / retryable）。所有失败行同步写 `ai_task_runs(task_type='report_generate' 或 'report_assessment', status='failed', validation_status='invalid' 仅在 AI 输出 schema 失败时使用, error_code=<B1>)`；不在 handler 路径返回 5xx 给客户端（report 生成在异步路径，客户端通过 `getFeedbackReport` 拿到 failed status）。
  - Retry policy：`async_jobs.attempts`（B4 列名）只在 lease 时递增一次，failure finalize 使用 lease 后当前 attempts 计算退避与永久失败状态，不二次递增；`available_at = now() + min(2^attempts * 30s, 30min)` 指数退避；`attempts >= 5` 时把 `feedback_reports.status='failed'` 与 `async_jobs.status='failed'` 同时置为终态（permanent failure），不再 reschedule；终态时发出 `report.generation.failed`。
- 状态机推进规则：`feedback_reports.status` 只允许 `queued → generating`、`generating → ready`、`generating → failed`、`generating → queued`（retry）、`failed → queued`（手工重试入口暂不暴露，保留状态出口）；越权迁移返回 internal error；非 worker 路径只允许 `getFeedbackReport` / `listTargetJobReports` 读取。
- Inline runner：在 `backend/internal/review/` 包内**新建独立** SELECT FOR UPDATE polling worker（Runner + Reaper + LeaseAsyncJob 抽象），与 `backend/internal/targetjob/drainer.go` 的 `targetjob.Drainer` 共存于同一 backend 进程但不共享抽象（D-16 决策）；列名严格对齐 B4 baseline（`locked_at` / `attempts`，无 `worker_id`）；poll 周期、lease timeout、max worker concurrency 由 A4 runtime config 与 F1 metric 字典登记决定（D-13 决策细节）。`backend/internal/privacy/runner/` 在仓库中仅是 `targetjob.JobHandler` 实现（不是 polling worker），本 spec 不基于 privacy/runner 套用 lifecycle。
- AI 调用形态：业务侧只调用 [F3 `RegistryClient.Resolve`](../prompt-rubric-registry/spec.md) → 拿三元组 → 调用 [A3 `AIClient.Complete`](../ai-provider-and-model-routing/spec.md)；payload metadata 必须携带 `feature_key + prompt_version + rubric_version + model_profile_name + language + data_source_version`；`language` 来源以 `practice_sessions.language`（与 backend-practice spec 一致）为唯一真理源，read handler 不消费 `Accept-Language` header 切换 wire；A3 observability decorator 自动写 `ai_task_runs(task_type='report_generate' 或 'report_assessment', ...)` typed columns；不在 backend-review 私自调用 provider SDK 或 hardcode prompt / model 字符串。`AssistantAction` 不在 backend-review 范围（无人工干预 turn），但 `FeedbackReport.provenance` 必须使用 B2 `GenerationProvenance` wire 6 字段，runtime 字段（`feature_key` / `model_profile_name` / provider / cost / latency）只进入 `ai_task_runs` typed columns 与 audit 摘要。
- 用户隔离：所有 read / write SQL 必须按 `user_id` 过滤；越权访问 `getFeedbackReport` / `listTargetJobReports` 返回 HTTP 404 + B1 `REPORT_NOT_FOUND`（D-15 派生），避免泄露存在性；DELETE /me 通过 ON DELETE CASCADE 级联清理 `feedback_reports` / `question_assessments`。
- 隐私 / 观测红线：`report.generated` / `report.generation.failed` 事件、F1 metric label、log、audit、`feedback_reports` jsonb 列（highlights / issues / next_actions）、`question_assessments` jsonb 列（dimension_results / strengths / gaps / recommended_framework）的 value 内容由 AI 输出生成；spec 强制禁止 raw `question_text` / `answer_text` / `hint_text` / AI prompt body / response body / provider secret 出现在 report jsonb 中。允许出现的内容是 AI 解析后的"摘要 / 证据片段引用 / 维度判断"，不得逐字回放用户答案或问题正文。具体 redaction 规则在 plan 内通过解析层与单元测试固化。
- F1 metric 注册边界：所有新增的 review metric / audit 类型必须先在 [F1 `observability-stack`](../observability-stack/spec.md) baseline 字典登记或由 F1 owner 承接，不得在本域私造 metric / label。
- 失败语义与状态机退出：worker 路径所有 AI 失败按 graceful failed 处理，永不让 worker panic / crash 流程（与 backend-practice D-19 narrowed 一致原则）；DELETE /me / cascade → 物理删除。

### 2.2 Out of Scope

- 不实现 `completePracticeSession` 或任何 Practice operation；归 [`backend-practice`](../backend-practice/spec.md) owner。本 spec 仅消费已存在的 `async_jobs(report_generate)` 行与 `feedback_reports(status='queued')` placeholder。
- 不实现 JD 解析、`target_jobs` 生命周期；归 [`backend-targetjob`](./../backend-targetjob/spec.md) owner。本 spec 仅引用 `target_job_id`。
- 不实现简历资产、`resume_assets` 解析、岗位定制版本；归 [`backend-resume`](../backend-resume/spec.md) owner。
- 不实现 STT / LLM / TTS 编排、committed-context 推进、barge-in 处理；归 [`practice-voice-mvp`](../practice-voice-mvp/spec.md) owner。
- 不实现独立 worker / Asynq dispatcher / 生产级 outbox consumer / 集中式 retry orchestrator。本 spec 早期在 backend-review 包内实现的 inline runner 形态（D-13 / D-16）已由 [`backend-async-runner/001`](../backend-async-runner/spec.md) 接管：`report_generate` 现由 `runner.Runtime` kernel 统一 lease / retry / reaper，handler 业务实现迁移为 `review.GenerateHandler`，`review.Runner` / `review.Reaper` 已删除。
- 不暴露独立 `POST /reports/{reportId}/regenerate` 或手工 retry API；status='failed' 的 report 只能等待 platform 触发或下一次 backend-review plan 决定开放手工入口。
- 不实现 report 高级排序 / filter（按 readiness / round / date / score 等）；P0 listTargetJobReports 仅按 `created_at DESC` 排序。
- 不实现 report 导出 / 分享 / 时间线；归 product-scope D-7 dashboard-only 边界。
- 不实现 report 评分质量反馈、LLM judge、A/B grayscale；归 [`prompt-rubric-registry`](../prompt-rubric-registry/spec.md) future `005-grayscale-and-quality-feedback` plan。
- 不实现高级 retry-focus 算法（情感 / 难度加权 / 历史趋势比较）；P0 只按 needs_work + queued_for_retry 简单集合，留 plan 002 future。
- 不实现 voice-specific report 维度；P0 报告内容字段不区分 text / voice modality（modality 信息仅在 `provenance.dataSourceVersion` 中体现）；voice-specific 维度等 voice MVP 上线后再设计。
- 不在本 spec 文档内 inline 修改 B2 OpenAPI、B3 events/jobs、B4 baseline 表结构、A3 provider 协议或 F3 baseline prompt / rubric 文本；任何契约修订必须由对应 owner spec / truth source / plan 先落地，再由 backend-review implementation 消费。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | API 契约来源 | 本域只消费 [B2 OpenAPI](../openapi-v1-contract/spec.md) 已定义的 2 个 Report 读取 operation；不私造 endpoint、不重写 schema | 任何字段 / 新 operation 先在 B2 spec / `openapi.yaml` 修订 |
| D-2 | DB 真理源 | 复用 [B4 baseline](../db-migrations-baseline/spec.md) 的 `feedback_reports` / `question_assessments` / `ai_task_runs` / `async_jobs` / `outbox_events` 表与 CHECK 约束、UNIQUE 约束、外键 | 不在本 spec 内 inline 落 migration；新列必须先修订 B4 |
| D-3 | 事件契约 | 复用 [B3](../event-and-outbox-contract/spec.md) 已冻结的 `report.generation.requested` / `report.generated` / `report.generation.failed` 与 `report_generate` job mapping；与 backend-practice D-28 / D-32 一致 — `practice.session.completed` 是 source event，`async_jobs(report_generate)` 由 backend-practice 同事务创建，本域只消费 | 事件 payload 与 PII 边界不得扩张；新增字段先回到 B3 spec / `shared/jobs.yaml` |
| D-4 | ReadinessTier 算法 | 基于 `question_assessments.dimension_results` 加权汇总：每个维度保留 F3 rubric `score_level`（weak / developing / proficient / strong）作为内部评分输入，按 rubric `weight` × score_level numeric 值（weak=0.2 / developing=0.5 / proficient=0.8 / strong=1.0）算单题得分，对 turn 取均值得到 session 得分；B2 wire `DimensionResult.status` 仍使用 B1 `DimensionStatus`（weak/developing → `needs_work`，proficient → `meets_bar`，strong → `strong`）以保持 OpenAPI schema 不漂移；阈值 `[0.30, 0.55, 0.75]` 映射到 `[not_ready, needs_practice, basically_ready, well_prepared]` 四档 | 阈值、权重和 score_level→DimensionStatus 映射作为 D-4 锁定值；后续调整必须修订本 spec 决策行 |
| D-5 | retry_focus_turns 选择 | P0 简单集合：选择 `question_assessments.overall_status='needs_work'` 或 `question_assessments.review_status='queued_for_retry'` 的 turn id 集合；按 `turn_index` 升序，最多 5 个；写入 `feedback_reports.question_assessments[].included_in_retry_plan=true` 与 `feedback_reports.retry_focus_turn_ids`（**该 jsonb 列不在当前 B4 baseline；由 plan 001 Phase 0 在 B4 pre-launch baseline rebase 时同 commit 新增**） | 高级加权算法留 plan 002 future；保持 P0 选择策略对 backend-practice `goal='retry_current_round'` plan 派生消费足够 |
| D-6 | next_action enum 决策 | 基于 `preparedness_level` × `retry_focus_turns.length` 决定 `feedback_reports.next_actions` 第一行 `type`：`needs_practice` 或 `not_ready` 且 retry_focus ≥ 1 → `retry_current_round`；`basically_ready` / `well_prepared` 且 retry_focus < 3 → `next_round`；其他 → `review_evidence`（保留 fallback） | 与 product-scope §6.9 "复练优先" 一致；具体阈值锁定在本决策行 |
| D-7 | status 状态机 | `feedback_reports.status` 只允许 `queued → generating`、`generating → ready`、`generating → failed`、`generating → queued`（retry）、`failed → queued`（保留出口，P0 不暴露入口）；非法迁移在 store / worker 内 ErrIllegalTransition；handler 路径只读取，不修改 | 单调推进 + 失败退避；防止脏数据态 |
| D-8 | AI 失败语义 | worker 路径所有 AI 失败 graceful：feedback_reports.status='failed' + error_code from B1 `AI_*` enum；不返回 5xx 给 HTTP；客户端通过 `getFeedbackReport.status='failed'` + `error_code` 感知；ai_task_runs 写 failed row 留供观测 | 异步流不阻塞用户；与 backend-practice D-36 narrowed 同款 graceful 原则（不过本域属于纯 worker 路径） |
| D-9 | provenance wire 边界 | `FeedbackReport.provenance` 严格只暴露 B2 `GenerationProvenance` 6 wire 字段（`promptVersion` / `rubricVersion` / `modelId` / `language` / `featureFlag` / `dataSourceVersion`）；runtime 字段（`feature_key` / `model_profile_name` / provider / cost / latency）仅写入 `ai_task_runs` typed columns 与 audit 摘要。**持久化锚点**：6 字段统一从 `feedback_reports` 表读取；当前 baseline 已有 `prompt_version` / `rubric_version` / `model_id`，plan 001 Phase 0 通过 B4 pre-launch baseline rebase 同 commit 新增 `language` / `feature_flag` / `data_source_version` 三列；getFeedbackReport 不依赖 cross-table JOIN | 与 backend-practice D-10 一致；保持单一 wire schema 来源 + 单表读取 |
| D-10 | ai_task_runs 行 | A3 observability decorator 自动写 `ai_task_runs(task_type='report_generate' 或 'report_assessment', ...)` 行；report.generate 主调一行（含 input_tokens / output_tokens / latency_ms / `status` ∈ B4 enum {`success`,`failed`,`timeout`,`fallback`}）；question_assessment 按每个 turn 一行；失败行 `status='failed'`、`validation_status` 视语义可选 `invalid`；F3 resolve 失败 / parse 失败由 worker 显式调 `aiclient.AITaskRunWriter` 写 `status='failed'` row（与 backend-practice D-37 同模式）。**列名取自 B4 baseline 真理源**：`status` 是 B4 CHECK 列（success/failed/timeout/fallback），`validation_status` 是独立的 `ok` / `invalid` 自由文本列，不混用 | 失败可观测；与 backend-practice 一致 |
| D-11 | 隐私红线 | `feedback_reports.highlights` / `issues` / `next_actions` / `question_assessments.strengths` / `gaps` / `recommended_framework` / `dimension_results` jsonb 不得包含 `question_text` / `answer_text` / `hint_text` / AI prompt body / AI response body / provider secret 的逐字回放；允许的内容是 AI 解析后的摘要、维度判断、引用证据片段的 turn_id + 偏移；event payload / metric label / log / audit 同款约束；具体 redaction 规则在 plan 内由解析层 + 单元测试固化 | 与 product-scope §9.3 / backend-practice D-11 / F1 一致 |
| D-12 | Outbox event payload schema | `report.generated.payload` = `{reportId, sessionId, targetJobId, preparednessLevel, questionIssueCount, promptVersion, rubricVersion, modelId}`（与 B3 当前 schema 完全一致）；`report.generation.failed.payload` = `{reportId, sessionId, errorCode, retryable}`；`report.generation.requested.payload` 由 backend-practice 写出（D-32 source_event_only），本域不重新发；payload 严格无 raw text | 与 B3 spec / `shared/events.yaml` 一致 |
| D-13 | Inline runner（已收干） | 本 spec P0 曾在 `backend/internal/review/runner.go` 内实现 SELECT FOR UPDATE polling worker（poll 周期 5s、concurrency 1、lease timeout 5min）。该形态已由 [`backend-async-runner/001`](../backend-async-runner/spec.md) 接管：`report_generate` 的 lease / retry / reaper 由 `runner.Runtime` kernel 统一持有，业务实现迁移为 `review.GenerateHandler`，`review.Runner` 文件已删除 | 历史记录保留以追溯；当前真理源为 backend-async-runner kernel |
| D-14 | listTargetJobReports 分页 | cursor 用 base64 encoded `(created_at, id)` tuple；`pageSize` 默认 20、最大 50；按 `created_at DESC, id DESC` 排序；空列表返回 `pageInfo.hasMore=false, nextCursor=null` | 与 B2 当前 `PaginatedFeedbackReport` schema 一致 |
| D-15 | Practice / Report 错误码前置 | B1/B2 必须新增 `REPORT_NOT_FOUND` 错误码并同步 generated Go/TS/OpenAPI 后，backend-review 才能在 404 隔离路径使用该 code；当前 B1 已有 `REPORT_NOT_READY` / `PRACTICE_SESSION_NOT_FOUND`，但 `REPORT_NOT_FOUND` 尚缺；若 B1/B2 owner 选择复用 `REPORT_NOT_READY` + 404 status 作为 cross-user 隔离信号需明确 status 与 detail 含义 | 保持错误码单一真理源；避免 spec 中出现未注册字面量 |
| D-16 | 异步 worker 抽象边界 | 本 spec P0 在 `backend/internal/review/` 内**新建独立** Runner / Reaper / `LeaseAsyncJob` 抽象，**不**复用 `backend/internal/targetjob/drainer.go` 的 `targetjob.Drainer` / `JobHandler` 模式。二者共存于同一 backend 进程：drainer 拥有 `target_import` / `source_refresh` / `privacy_delete` / `resume_parse` / `resume_tailor` / `debrief_generate` 等已注册 job_type 的 claim + finalize；review.Runner 拥有 `report_generate` job_type 的 lease + finalize + reaper。SQL 列名以 B4 baseline 为真理源：`async_jobs.locked_at`（不写 `leased_at`）、`async_jobs.attempts`（不写 `attempt_count`）、`async_jobs.status` ∈ {`queued`,`running`,`succeeded`,`failed`,`cancelled`,`dead`}；P0 阶段**不**新增 `worker_id` 列（worker 身份通过结构化 logger + metric label 暴露，不进 DB）。该 dual-runner 形态已由 [`backend-async-runner/001`](../backend-async-runner/spec.md) 收干：`review.Runner` / `review.Reaper` 删除，`report_generate` 与原 drainer 注册的 job_type 统一由单一 `runner.Runtime` kernel lease / retry / reaper，业务 handler（含 `review.GenerateHandler`）保留各 owner 实现。SQL 列名仍以 B4 baseline 为真理源：`async_jobs.locked_at`、`async_jobs.attempts`、`async_jobs.status` ∈ {`queued`,`running`,`succeeded`,`failed`,`cancelled`,`dead`} | 列名 / status 枚举与 baseline 完全对齐避免 spec → 实现 drift；dual-runner 并发治理已由 kernel 收干 |

### 3.2 非后端 owner 决策

| ID | 事项 | Owner | 本域处理 |
|----|------|-------|----------|
| Q-1 | `practice.report.requested` 是否独立 source event（与 `practice.session.completed` 分离） | backend-practice D-28 / D-32 (已锁定 source_event_only) | 本 spec 直接消费 backend-practice 创建的 `async_jobs(report_generate)` queued row，不期待 dispatcher 二次创建 |
| Q-2 | F3 baseline `report.generate` / `report.question_assessment` prompt body 内容与 rubric 维度具体 weight | prompt-rubric-registry future `004-real-model-profile-and-evals` | 本 spec 仅约定通过 F3 Resolve 拿三元组；rubric weight × score_level 算法在本 spec D-4 锁定 |
| Q-3 | DELETE /me CASCADE 实现是否 atomic 跨表（feedback_reports + question_assessments + ai_task_runs） | platform / future privacy plan | 本 spec 仅约定 ON DELETE CASCADE 已落地，平台触发 atomic 删除 |
| Q-4 | review worker 是否多实例水平扩展 | platform / future backend-async-runner | 本 spec P0 默认 max concurrency=1，多实例由未来 plan 验证 lock & lease 正确性 |

### 3.3 待确认事项

| ID | 待确认事项 | 影响 | 默认处理 |
|----|------------|------|----------|
| Q-5 | report 失败后是否自动重试 5 次还是 3 次 | retry 总耗时 vs provider 临时故障容忍 | 默认 5 次（指数退避到 30 min 上限）；后续根据 production observability 调整 |
| Q-6 | `provenance.dataSourceVersion` 应该包含哪些信号（session id？practice plan snapshot version？） | report 可追溯性 vs 隐私 | 默认 `dataSourceVersion='practice-session/<sessionId>@<schema-version>'`；plan 内固化 |

## 4 设计约束

### 4.1 API 契约约束

- 必须使用 [B2 generated `ServerInterface`](../openapi-v1-contract/spec.md) 注册 2 个 Report 读取 handler，不得绕过 generated types 自造 router。
- 入参反序列化必须使用 generated request types；响应必须使用 generated response types（`FeedbackReport` / `PaginatedFeedbackReport`）；fixture 与真实 handler 共用同一 schema。
- `getFeedbackReport` 必须按 `(user_id, report_id)` 过滤；越权 / 不存在均返回 `404 REPORT_NOT_FOUND`（D-15 前置）；`status` ∈ `queued`/`generating` 时返回 `200 + placeholder` 而非 `REPORT_NOT_READY`（因 OpenAPI 已声明 status 字段，placeholder 是合法 wire shape）；客户端可据 `status` 字段决定是否轮询 / 渲染 generating 屏。
- `listTargetJobReports` 必须按 `(user_id, target_job_id)` 过滤；越权 / 不存在 target 返回空列表（不暴露 target_job 存在性，404 仅在 target_job_id 本身越权时由 backend-targetjob middleware 拒绝，本 spec 不在 backend-review handler 内做 target_job 存在性校验）。
- 所有 read 路径无副作用，不要求 `Idempotency-Key`。
- worker 路径写入路径必须先 lease async_jobs 行（SELECT FOR UPDATE SKIP LOCKED），后续 `feedback_reports` 写入与 outbox emit 必须在同一短事务内；F3 / A3 调用在事务外执行。

### 4.2 数据约束

- 所有 DB 写入必须在各自短事务内完成；外部 AI 调用不得包在 DB transaction 内。`report.generated` / `report.generation.failed` 必须通过 outbox 与对应业务写入同事务发出，避免双写。
- `feedback_reports.status` 仅允许 D-7 状态机；非法迁移在 store 层 ErrIllegalTransition；handler 路径 read-only。
- `question_assessments` 必须 UNIQUE(report_id, turn_id)；批量 INSERT 前需要确认 `practice_turns` 行存在（cross-table read in worker context）。
- `feedback_reports.preparedness_level` 仅允许 D-4 四档值；`question_assessments.overall_status` 仅允许 `strong / meets_bar / needs_work`（B4 已 CHECK）。
- `async_jobs(report_generate).dedupe_key=sessionId`（与 backend-practice complete 同事务创建保持一致）；重复 INSERT 由 UNIQUE 兜底。
- D-13 inline runner 必须使用 `SELECT ... FOR UPDATE SKIP LOCKED` 避免行锁冲突；lease timeout 后由后台 reaper 把 stale `running` 行回退到 `queued`（reaper 实现在 plan 内落实）。
- 软删 / 物理删除：v1 不引入软删；DELETE /me 通过 ON DELETE CASCADE 物理清理 `feedback_reports` / `question_assessments`。

### 4.3 安全 / 隐私约束

- 所有 API 走 [backend-auth](../backend-auth/spec.md) session middleware；未认证返回 `401 AUTH_UNAUTHORIZED`；越权返回 404 + `REPORT_NOT_FOUND`（D-15 前置），不泄露资源存在性。
- AI 调用 fail-closed 边界：本 spec worker 路径无用户阻塞调用，所有 AI 失败按 graceful failed 处理；不静默回退 stub provider（除 `APP_ENV=test`）；F3 / A3 缺 secret 或 provider unreachable 同样落 `feedback_reports.status='failed' + error_code='AI_PROVIDER_*'`。
- log / metric label / audit / 事件 payload / `feedback_reports` jsonb 列 / `question_assessments` jsonb 列 不得包含 `question_text` / `answer_text` / `hint_text` / AI prompt body / AI response body / provider secret；只允许 AI 解析后的摘要、turn_id 引用、维度判断、IDs、length、count、status、profile、provider、model_id、cost micros、error code 摘要。
- `audit_events` 触发集合：`generateFeedbackReport` worker 路径在 status 推进到 `ready` 或 `failed` 时写一行 audit；read handler 不写 audit（高频）。audit metadata 仅含 `report_id` / `session_id` / `target_job_id` / `status` / `preparedness_level` / `language` / `error_code`，不含问答文本。

### 4.4 异步 / 可观测约束

- 本域 `report_generate` 消费已由 [`backend-async-runner/001`](../backend-async-runner/spec.md) kernel 接管：`runner.OutboxDispatcher` 通过 `IsSourceEventOnly(report_generate)` 谓词遵守 D-32 forward-binding，不二次创建 report job；`review.GenerateHandler` 注册到单一 `runner.Runtime`，与 backend-practice / privacy 等共存于同一 backend 进程且不重复消费。
- F1 metric 字典登记前置：report-specific business metrics（如 report generated count / generation duration / failed count by error_code）实施前必须先在 [F1 baseline](../observability-stack/spec.md) 字典登记，并且 label 只能使用 F1 allowed labels 与 B1 有界枚举；AI 调用 metric 复用 A3 已登记的 7 个 `ai_task_*` metric family，`feature_key` / `prompt_version` / `rubric_version` 只进入 `ai_task_runs` typed columns 或审计摘要，不进入 metric label。
- worker P95 延迟（从 lease 到 ready）作为观测目标登记，但不作为本 spec 验收 gate。

### 4.5 文档治理约束

- 本 spec 后续修订必须原地更新；不允许创建同主题 sibling spec。
- 涉及 OpenAPI / events / migrations / shared enums / runtime config 的修改必须先回 owner spec / `*.yaml` truth source。
- 涉及用户行为流的 plan 必须维护 BDD gate；本域 2 个 read operation 属于用户可见 API 行为；worker 触发后端 outbox / DB state 推进，前端可通过 `getFeedbackReport` 感知，归属用户行为流。
- 命中 `completed` plan 时不创建同主题 sibling follow-up plan，原地修订即可。
- 旧术语（`reportLayout` / 5 档 readiness / 独立 `mistakes` 错题本 / `drill` / `growth_center` / 报告时间线 / 多形态 report）不得在本 spec 与衍生 plan 中出现，违反必须先修订 [product-scope](../product-scope/spec.md)。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| API contract | [B2 `openapi-v1-contract`](../openapi-v1-contract/spec.md) | `getFeedbackReport` / `listTargetJobReports` operation 与 `FeedbackReport` / `PaginatedFeedbackReport` / `QuestionAssessment` / `ReportHighlight` / `ReportIssue` / `ReportNextAction` / `DimensionResult` / `GenerationProvenance` schema、fixtures、generated client / server |
| Backend domain | `backend-review`（本 spec） | 2 个 read handler + service + store + inline review runner + AI 生成 / 维度评估 / readiness / retry_focus / next_action / status 状态机 / outbox emit |
| DB schema | [B4 `db-migrations-baseline`](../db-migrations-baseline/spec.md) | `feedback_reports` / `question_assessments` / `ai_task_runs` / `async_jobs` / `outbox_events` 列与索引；shared `idempotency_records` 表（本 spec 不要求 idempotency，read-only） |
| Event / job contract | [B3 `event-and-outbox-contract`](../event-and-outbox-contract/spec.md) | `report.generation.requested` / `report.generated` / `report.generation.failed` 事件；`report_generate` job mapping 与 `triggerEventSemantic: source_event_only` |
| Shared enums / errors | [B1 `shared-conventions-codified`](../shared-conventions-codified/spec.md) | `ReportStatus` / `ReadinessTier` / `DimensionStatus` / `Confidence` / `QuestionReviewStatus` / `REPORT_NOT_READY` / `REPORT_NOT_FOUND`（D-15 前置） |
| AI provider | [A3 `ai-provider-and-model-routing`](../ai-provider-and-model-routing/spec.md) | `AIClient.Complete`、provider registry、model profile、observability decorator 写 `ai_task_runs` |
| Prompt / rubric | [F3 `prompt-rubric-registry`](../prompt-rubric-registry/spec.md) | `report.generate` / `report.question_assessment` feature_key、Resolve 实现、baseline prompt / rubric v0.1.0 active |
| Config / secret | [A4 `secrets-and-config`](../secrets-and-config/spec.md) | provider secret、feature flag、review runner poll/lease/max-concurrency 配置 |
| Observability | [F1 `observability-stack`](../observability-stack/spec.md) | metric / audit 类型登记、label allowlist、隐私红线 |
| Auth / isolation | [`backend-auth`](../backend-auth/spec.md) | session middleware、user-scoped read、DELETE /me CASCADE 协议 |
| Upstream — Practice | [`backend-practice`](../backend-practice/spec.md) | 提供 `feedback_reports(status='queued')` placeholder + `async_jobs(report_generate, status='queued', dedupe_key=sessionId)` queued row + `practice.session.completed` outbox source event；本 spec 仅消费 |
| Upstream — TargetJob | [`backend-targetjob`](../backend-targetjob/spec.md) | 提供 `target_jobs` 行；本 spec 仅引用 `target_job_id` 用于过滤 |
| Upstream — Resume | [`backend-resume`](../backend-resume/spec.md) | 提供 resume asset 行；本 spec 不直接消费 resume 内容，由 backend-practice 在 plan 上下文中已 hydrate |
| Downstream — Report UI | [`frontend-report-dashboard`](../frontend-report-dashboard/spec.md) | `ReportScreen` / `GeneratingScreen` 渲染、复练 CTA、报告失败态；本 spec 提供 schema + data，不耦合 UI |
| Downstream — Debrief | future `backend-debrief` | 真实复盘流程；不消费本 spec 产出 |
| Frontend consumer | [`frontend-report-dashboard`](../frontend-report-dashboard/spec.md) + [`frontend-workspace-and-practice`](../frontend-workspace-and-practice/spec.md) plan 004 | `getFeedbackReport` 轮询入口在 generating 屏；`getFeedbackReport` / `listTargetJobReports` 详情入口在 report 屏 |
| Scenario coverage | scenarios owner + 本 subject | `E2E.P0.0NN-report-*` 套件 setup / trigger / verify / cleanup（具体编号在各 plan 内分配） |
| Async runner replacement | future `backend-async-runner` | 接管 runtime dispatcher / drainer / 多 worker；必须沿用 D-32 forward-binding 与 backend-review D-13 inline runner 边界 |

### 5.1 Operation Matrix

| `operationId` | Fixture | Frontend consumer | Backend handler | Persistence | AI dependency | Scenario coverage |
|---------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `getFeedbackReport` | `openapi/fixtures/Reports/getFeedbackReport.json`（`default` / `report-generating` / `prototype-baseline`；plan 001 计划新增 `report-failed` variant 用于 status='failed' 路径） | generating 轮询 + report dashboard 详情 | Plan 001 Phase 5：`backend/internal/api/reports.GetFeedbackReport` + `backend/internal/review.Service.GetFeedbackReport` + `backend/internal/store/review.GetFeedbackReport`（user-scoped） | `feedback_reports` + `question_assessments` read | none in handler path | `E2E.P0.053` |
| `listTargetJobReports` | `openapi/fixtures/Reports/listTargetJobReports.json`（`default`；plan 001 计划新增 `empty` variant） | report dashboard 列表 / 历史入口（本 spec scope；frontend D-7 dashboard-only 不在一级导航暴露） | Plan 001 Phase 5：`backend/internal/api/reports.ListTargetJobReports` + `backend/internal/review.Service.ListTargetJobReports` + `backend/internal/store/review.ListTargetJobReports`（cursor 分页 + user-scoped） | `feedback_reports` cursor read | none | `E2E.P0.053` |
| `(worker: report_generate job)` | N/A | N/A（异步 worker，不暴露 API） | Plan 001 Phase 1-4：`backend/internal/review/runner.go` + `backend/internal/review.GenerateReportService` + `backend/internal/store/review.PersistReport` | `feedback_reports` write + `question_assessments` write + `async_jobs` lease + `outbox_events` + `ai_task_runs` + `audit_events` | F3 `report.generate` + `report.question_assessment` v0.1.0；A3 `AIClient.Complete` × 2 类调用 | `E2E.P0.052` + `E2E.P0.054` + `E2E.P0.055` |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | report 主路径（queued → ready） | backend-practice 已 completePracticeSession，创建 `feedback_reports(status='queued')` + `async_jobs(report_generate, status='queued')`；F3 active；A3 active | review runner lease 该 job 并处理 | `feedback_reports.status` 推进 `queued → generating → ready`；填充 `highlights` / `issues` / `next_actions` / `preparedness_level` / `provenance`；`question_assessments` 按每个 turn 写一行；`async_jobs.status='succeeded'`；outbox `report.generated` 发出 1 行 | 001-report-generation-baseline |
| C-2 | 逐题维度评估写入 | review runner 进入 Phase 2（已 lease + generating）；session 有 N 个 turn | runner 对每个 turn 调 F3 `report.question_assessment` + A3 | `question_assessments` 写入 N 行，每行含 `dimension_results` (map of {dimension → B2 `DimensionResult.status` / `confidence` + internal `score_level`}) + `overall_status` + `confidence` + `strengths` + `gaps` + `recommended_framework` + `review_status='open'/'queued_for_retry'` + `included_in_retry_plan` 标记；read handler 只暴露 B2 schema 允许的字段，不把 internal `score_level` 扩成新的 wire 字段 | 001 |
| C-3 | ReadinessTier 计算 | C-2 已完成 N 行 question_assessments | runner 进入 Phase 3 | `feedback_reports.preparedness_level` ∈ {`not_ready`,`needs_practice`,`basically_ready`,`well_prepared`} 按 D-4 阈值算出；非空 | 001 |
| C-4 | getFeedbackReport status placeholder | report status='queued' 或 'generating'，用户尚未完成生成 | `GET /reports/{reportId}` | 200 + `FeedbackReport{status, preparednessLevel:null, highlights:[], issues:[], nextActions:[], questionAssessments:[], provenance:null, error_code:null}`；前端可据 status 决定是否继续轮询 | 001 |
| C-5 | listTargetJobReports 分页 + 空列表 | 用户 A 对 target_job=T 有 0 个 / 1 个 / 25 个报告 | `GET /targets/{T}/reports?pageSize=20` | 0 个 → `{items:[],pageInfo:{nextCursor:null,pageSize:20,hasMore:false}}`；25 个 → 第一页 20 个 + `nextCursor` 非空 + `hasMore=true`，按 `created_at DESC` 排序；越权 target_job 走 backend-targetjob middleware（本 spec 不在此层校验） | 001 |
| C-6 | AI 失败 graceful（F3 resolve / A3 timeout / parse empty） | F3 ResolveActive 失败 / A3 timeout / parsed empty 等任一失败 | review runner lease 后调用失败 | `feedback_reports.status='failed', error_code=<B1 AI_*>, generated_at=now()`；`async_jobs` 使用 lease 后当前 `attempts` 计算 `available_at=now()+backoff(attempts)` 并清空 `locked_at`（failure finalize 不二次递增 attempts）；attempts < 5 时 `async_jobs.status='queued'`（retry），attempts >= 5 时 `status='failed'`（permanent）；outbox `report.generation.failed` 发出；`ai_task_runs` 写一行 `status='failed'`（B4 enum，非 'succeeded'）+ B1 error_code；不返回 5xx 给 HTTP | 001 |
| C-7 | Retry policy + max attempts | C-6 命中且 `async_jobs.attempts` 已累计到 4（B4 列名） | 第 5 次失败 | `async_jobs.status='failed'`（permanent）+ `locked_at=null`；`feedback_reports.status='failed'`；不再 reschedule；outbox `report.generation.failed{retryable:false}` | 001 |
| C-8 | FeedbackReport.provenance wire 完整 | report status='ready' | `GET /reports/{reportId}` | `provenance` 仅含 B2 wire 6 字段：`promptVersion` / `rubricVersion` / `modelId` / `language` / `featureFlag` / `dataSourceVersion`；`feature_key` / `model_profile_name` / provider / cost / latency 等运行时字段不在 wire JSON 出现 | 001 |
| C-9 | Cross-user 隔离 | 用户 A 持有 reportX | 用户 B 调 `GET /reports/X` 或 `GET /targets/T/reports` 其中 T 属于 A | 用户 B 不能读到 reportX；getFeedbackReport 返回 `404 REPORT_NOT_FOUND`；listTargetJobReports 在 target_job_id 属于 A 时由 backend-targetjob middleware 拦截（404）；本 spec handler 内 user_id 过滤兜底 | 001 |
| C-10 | 隐私红线 | report status='ready' | 检查 `feedback_reports.highlights/issues/next_actions` jsonb / `question_assessments.dimension_results/strengths/gaps/recommended_framework` jsonb / outbox payload / metric label / log / audit | 不含 `question_text` / `answer_text` / `hint_text` / AI prompt body / AI response body / provider secret 的逐字回放；只含 AI 解析后的摘要、维度判断、turn_id 引用、IDs、length、count、status、profile、provider、model_id、cost micros、error code 摘要 | 001 |
| C-11 | ai_task_runs 行 | report.generate / report.question_assessment AI 调用完成或失败 | 检查 `ai_task_runs` | 每个 report.generate 调用一行 `task_type='report_generate'`；每个 question_assessment 调用一行 `task_type='report_assessment'`；行含 `feature_key` / `model_profile_name` / `input_tokens` / `output_tokens` / `latency_ms` / `validation_status` / `error_code`；失败行 `error_code` 来自 B1 enum，不含 raw provider message | 001 |
| C-12 | 单 async_jobs 行双 worker 抢占 | 两个 review worker 同时 poll | 并发执行 lease | `SELECT FOR UPDATE SKIP LOCKED` 保证最多一个 worker 拿到该行；另一个 worker 立即返回继续 poll 下一行；同一 report 不会双触发 AI 生成 | 001 |

## 7 关联计划

`001-report-generation-baseline` 已派 plan（spec v1.0 同会话），其余 plan 按 phase closability 与 owner 边界依次派生。全局前置：依赖 [`prompt-rubric-registry/001-baseline`](../prompt-rubric-registry/plans/001-baseline/plan.md) 已 completed（`report.generate` / `report.question_assessment` v0.1.0 baseline 已 active）；依赖 backend-practice/002 已 completed（complete 同事务创建 queued job + placeholder）。

1. [`001-report-generation-baseline`](./plans/001-report-generation-baseline/plan.md)：D-1 ~ D-16 全部决策落地；inline review runner + Phase 0 跨 spec 前置（B1 `REPORT_NOT_FOUND` 新增 + B4 reaper 间接需求评估 + F3 baseline preflight）+ Phase 1 runner skeleton + Phase 2 AI 生成 + Phase 3 ReadinessTier / retry_focus / next_action 算法 + Phase 4 持久化 + outbox + ai_task_runs + Phase 5 read handler + Phase 6 失败语义 / retry / 隐私 / observability / legacy negative。
2. 保留编号建议 `002-advanced-retry-focus-and-listing`：D-5 高级 retry-focus 加权算法 + listTargetJobReports 高级 filter / 排序 + 手工 retry API（如产品需要）。
3. 保留编号建议 `003-report-retention-and-cascade`：DELETE /me CASCADE 验证 + retention policy（超期 report 删除）+ 隐私 export 占位（与 product-scope Q-5 一致）。

每个 plan 通过 `/design` 落地时单独配 BDD/test plan；本 spec §6 AC 是这些 plan 的统一来源。

## 8 关联文档

- [Product Scope §6.9 M4 证据化报告](../product-scope/spec.md#69-m4证据化报告)
- [Product Scope §4.1 产品原则](../product-scope/spec.md#41-产品原则)（证据优先、复练优先）
- [docs/ui-design/report-dashboard.md](../../ui-design/report-dashboard.md)
- [openapi-v1-contract](../openapi-v1-contract/spec.md)
- [event-and-outbox-contract](../event-and-outbox-contract/spec.md)
- [db-migrations-baseline](../db-migrations-baseline/spec.md)
- [shared-conventions-codified](../shared-conventions-codified/spec.md)
- [ai-provider-and-model-routing](../ai-provider-and-model-routing/spec.md)
- [prompt-rubric-registry](../prompt-rubric-registry/spec.md)
- [secrets-and-config](../secrets-and-config/spec.md)
- [observability-stack](../observability-stack/spec.md)
- [backend-auth](../backend-auth/spec.md)（auth/session middleware）
- [backend-practice](../backend-practice/spec.md)（上游 handoff source）
- [backend-targetjob](../backend-targetjob/spec.md)（target_job_id 来源）
- [frontend-report-dashboard](../frontend-report-dashboard/spec.md)（下游 UI consumer）
- [docs/development.md §2 Frontend / Backend Contract Workflow](../../development.md)
