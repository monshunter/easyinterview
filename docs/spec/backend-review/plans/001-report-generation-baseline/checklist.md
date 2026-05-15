# 001 — Report Generation Baseline Checklist

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-05-16

**关联计划**: [plan](./plan.md)

## Phase 0: 跨 spec 前置修订 + Preflight

- [x] 0.1 修订 `shared/conventions.yaml#errors` 新增 `REPORT_NOT_FOUND` 错误码（`httpStatus: 404`, `retryable: false`, `message: "feedback report not found or not accessible"`）；同步 [B1 spec](../../../shared-conventions-codified/spec.md) Header bump（next minor）+ `history.md` 追加 row："授权 backend-review/001 Phase 0 新增 `REPORT_NOT_FOUND` 错误码（cross-user 隔离 404）"；验证：`python3 scripts/lint/conventions_drift.py --repo-root .` 通过 + `make codegen-check` 通过 + generated Go `ErrReportNotFound` 常量 + generated TS 等价常量出现
- [x] 0.2 修订 `openapi/openapi.yaml` 中 `getFeedbackReport` 的 404 response 显式与 `REPORT_NOT_FOUND` 关联（与 `getPracticeSession` 404 `PRACTICE_SESSION_NOT_FOUND` 同模式）；同步 `openapi/baseline/openapi-v1.0.0.yaml`（pre-launch baseline rebase）；regenerate `make codegen-openapi` + `make codegen-check`；同步 [B2 spec](../../../openapi-v1-contract/spec.md) Header bump + `history.md` 追加 row："授权 backend-review/001 Phase 0 把 `getFeedbackReport` 404 response 显式与 `REPORT_NOT_FOUND` 关联（pre-launch baseline rebase）"；验证：`make codegen-check` + `cd backend && go build ./...` + `pnpm --filter @easyinterview/frontend typecheck` 通过
- [x] 0.3 **无条件**修订 B4：grep 已确认 `migrations/000001_create_baseline.up.sql#ai_task_runs.task_type CHECK` 当前不含 `report_assessment`；修订 CHECK 扩值 `report_assessment`（pre-launch baseline rebase），同步 `migrations/enum-sources.yaml` + `migrations/lint.sh`；同步 [B4 spec](../../../db-migrations-baseline/spec.md) Header bump + `history.md` 追加 row + `backend/internal/ai/aiclient/writers.go` 新增 `AITaskRunTaskReportAssessment AITaskRunCapability = "report_assessment"` 常量到 `allowedAITaskRunCapabilities`；验证：`cd backend && go test ./internal/migrations ./internal/ai/aiclient -count=1` + `migrations/lint.sh` 通过
- [x] 0.4 **无条件**修订 B2 `FeedbackReport.errorCode` 字段：grep 已确认 `openapi/openapi.yaml#FeedbackReport` schema 未声明 `errorCode`；与 0.2 同 commit 扩 B2 schema 增加 `errorCode: oneOf[ApiErrorCode|null]` 字段并 regenerate `make codegen-openapi` + `make codegen-check`；同步 `openapi/baseline/openapi-v1.0.0.yaml`；同步 [B2 spec](../../../openapi-v1-contract/spec.md) `history.md` 追加 row；`feedback_reports.error_code` B4 列已存在（已 active），无需扩 B4 错误码列
- [x] 0.5 **无条件**修订 B4 `feedback_reports` 新增 4 列：grep 已确认当前 baseline 17 列不含 `language` / `feature_flag` / `data_source_version` / `retry_focus_turn_ids`；与 0.3 同 commit 在 `migrations/000001_create_baseline.up.sql` `feedback_reports` 表定义中追加 4 列（`language text NOT NULL DEFAULT 'en'` / `feature_flag text NOT NULL DEFAULT 'none'` / `data_source_version text NOT NULL DEFAULT 'not_applicable'` / `retry_focus_turn_ids jsonb NOT NULL DEFAULT '[]'::jsonb`）；同步 [B4 spec](../../../db-migrations-baseline/spec.md) `history.md` 追加 row；migration test 新增 `TestFeedbackReportsContainsProvenancePersistenceColumns` 断言 4 列存在 + 默认值
- [x] 0.6 扩展 `openapi/fixtures/Reports/getFeedbackReport.json` 命名场景：`report-failed`（status='failed', errorCode 非空, 空 highlights/issues/nextActions/questionAssessments, provenance:null）；扩展 `openapi/fixtures/Reports/listTargetJobReports.json` 命名场景：`empty`（空 items + pageInfo.hasMore=false + nextCursor=null）；验证：`make validate-fixtures` 通过 + contract test 通过；fixture 在 0.4 schema 扩展后再写
- [x] 0.7 F3 preflight assert：读取 `docs/spec/prompt-rubric-registry/spec.md` 当前版本 + `plans/001-baseline/checklist.md` `状态: completed` + work-journal `close 001-baseline lifecycle` commit；新增 `backend/internal/ai/registry/backend_review_preflight_test.go::TestF3ReportGenerateAndAssessmentPreflight` 断言 `RegistryClient.ResolveActive(ctx, "report.generate", "en")` / `("report.question_assessment", "en")` 返回非空 ResolvedPrompt（含 `score_levels` 4 档 + weight），并直接读取 `config/prompts/report.generate/v0.1.0*.md` / `config/prompts/report.question_assessment/v0.1.0*.md` 断言 prompt body 不要求 raw `{{question}}` / `{{answer}}` / `{{transcript}}` 输入、不要求 `evidence_quotes` / verbatim quote 输出；若失败则停止并回 F3 owner 修订，不进入 Phase 1；验证：`cd backend && go test ./internal/ai/registry -count=1` 通过
- [x] 0.8 Phase 0 收口：`make codegen-check` / `make validate-fixtures` / `migrations/lint.sh` / `python3 scripts/lint/conventions_drift.py --repo-root .` / `cd backend && go build ./...` / `make docs-check`（含 `/sync-doc-index --check`） 全部通过；B1/B2/B4 owner spec history append 与 Header bump 完成；本 plan 状态保持 `active`

## Phase 1: Inline review runner + lease + status state machine

- [x] 1.1 新增 `backend/internal/review/` 包骨架：`runner.go`（Start/Stop + poll loop + max concurrency=1 默认；worker 身份通过结构化 logger field，不写 DB）；`service.go`（GenerateReport 接口占位返回 fake outcome）；`lease.go`（LeaseNextJob + ReleaseJobAfterSuccess/Failure，SQL 使用 `attempts` / `locked_at` 列）；`reaper.go`（Start/Stop + 周期性 reclaim 超时 lease 行，**仅作用于 `job_type='report_generate'`**，UPDATE 条件 `locked_at < now() - lease_timeout`）；不引入 AI 调用（service 占位）；spec D-16 决策：与 `backend/internal/targetjob/drainer.go` 共存但不复用其抽象
- [x] 1.2 新增 `backend/internal/store/review/` 包骨架：`lease_async_job.go`（LeaseAsyncJob + UpdateAsyncJobRunning/Succeeded/Failed，SQL 严格使用 B4 列名 `attempts` / `locked_at`；UPDATE on succeeded 同时 `locked_at=null`）；`feedback_reports_status.go`（UpdateFeedbackReportStatus 含 D-7 状态机乐观锁）；`reaper.go`（ReclaimExpiredLeases by job_type='report_generate'）；新增 `backend/internal/api/reports/` 空 skeleton 等待 Phase 5
- [x] 1.3 实现 `runner_test.go::TestRunnerLeasesAndAdvancesToGenerating` 用 fake store + fake service 断言 lease 成功后 status `queued → generating`；断言 `locked_at` 列被 set（不断言 `worker_id` 列）；验证 lease 失败时 worker 继续 poll（不 panic）
- [x] 1.4 实现 `lease_async_job_test.go::TestLeaseSkipLocked` 用 multi-goroutine 真实 Postgres（test stack）断言 SKIP LOCKED 行为；两个 worker 不会同时 lease 同一行；同一 row 单 worker 持有 lease 直到 release（通过 `locked_at != null` + transaction lock 验证；不依赖 `worker_id` 列）
- [x] 1.5 实现 `feedback_reports_status_test.go::TestStatusStateMachineEnforcement` 断言 `queued → ready` 直接迁移被拒绝（ErrIllegalTransition）；`failed → ready` 被拒绝；`generating → ready` / `generating → failed` / `failed → queued`（保留出口）允许
- [x] 1.6 实现 `reaper_test.go::TestReaperReclaimsExpiredLease` 用 fake clock 验证 reaper 周期回收 stale running 行；不动 succeeded / failed 行；不动未超时 running 行；不动其它 job_type 的 running 行（dual-runner 边界）；reaper UPDATE 条件 `locked_at < now() - lease_timeout` AND `job_type='report_generate'`
- [x] 1.7 Phase 1 收口：`cd backend && go test ./internal/review/... ./internal/store/review/... -count=1` 全部通过；lease + reaper + status 推进 + double-worker SKIP LOCKED 单元测试覆盖；不引入 AI 调用

## Phase 2: AI 调用与内容生成

- [x] 2.1 新增 `backend/internal/review/generate_report.go::generateReportContent(ctx, session, plan, turns) (ReportContentDraft, error)`：调 `f3.ResolveActive(ctx, "report.generate", language)` 拿 ResolvedPrompt；构造 system+user message（**不**含 raw `question_text` / `answer_text` / `hint_text` literal）；调 `a3.Complete(...)`；解析 AI JSON 响应到 ReportContentDraft；A3 observability decorator 自动写 `ai_task_runs(task_type='report_generate', ...)` 行
- [x] 2.2 新增 `backend/internal/review/question_assessment.go::assessQuestionsForAllTurns(ctx, session, plan, turns) ([]QuestionAssessmentDraft, error)`：对每个 turn（按 turn_index 升序）调 `f3.ResolveActive(ctx, "report.question_assessment", language)` + `a3.Complete(...)`；解析 JSON 到 QuestionAssessmentDraft；A3 decorator 写 `ai_task_runs(task_type='report_assessment')` typed columns
- [x] 2.3 新增 `generate_report_test.go::TestGenerateReportContentSuccess` 用 fake F3 + fake AIClient 返回合法 JSON 断言 outcome 结构（含 highlights/issues/next_actions 数组）
- [x] 2.4 新增 `generate_report_test.go::TestGenerateReportContentBuildsPromptWithoutLeaks` 反向断言 user message 不含 `question_text` / `answer_text` / `hint_text` literal（用 grep 字段或 reflect 拼装段）
- [x] 2.5 新增 `question_assessment_test.go::TestAssessQuestionsForAllTurns` 用 fake F3 + fake AIClient 断言 N 行 outcome，按 turn_index 升序，每行含 dimension_results map + overall_status + confidence + strengths + gaps + recommended_framework + review_status
- [x] 2.6 新增 `question_assessment_test.go::TestAssessQuestionsBuildsPromptWithoutLeaks` 反向断言
- [x] 2.7 Phase 2 收口：`cd backend && go test ./internal/review/... -count=1` 全部通过；F3 / A3 success 路径 + prompt redaction 单元测试覆盖；AI 失败路径占位（Phase 6 完整 6 路径 matrix）

## Phase 3: ReadinessTier / retry_focus / next_action 算法

- [x] 3.1 新增 `backend/internal/review/readiness.go::computeReadinessTier(assessments, rubric) ReadinessTier`：按 dimension `score_level` numeric 值（weak=0.2/developing=0.5/proficient=0.8/strong=1.0）× rubric weight 算单题分；同时新增 helper 把 score_level 映射为 B2 wire `DimensionStatus`（weak/developing → `needs_work`，proficient → `meets_bar`，strong → `strong`），不得把 weak/developing/proficient 写入 B2 `DimensionStatus`；对所有 turn 取均值；阈值 < 0.30 → not_ready / [0.30, 0.55) → needs_practice / [0.55, 0.75) → basically_ready / ≥ 0.75 → well_prepared
- [x] 3.2 实现 `readiness_test.go::TestComputeReadinessTier` 表驱动覆盖 4 档阈值边界（测试值 0.29 / 0.30 / 0.54 / 0.55 / 0.74 / 0.75 / 0.99）+ property 测试随机 N 维度 × N turn + 空 assessments fallback not_ready
- [x] 3.3 新增 `backend/internal/review/retry_focus.go::selectRetryFocusTurnIDs(assessments) []uuid.UUID`：筛选 `overall_status='needs_work'` OR `review_status='queued_for_retry'`；按 turn_index 升序；最多 5 个；mutate `assessments[i].IncludedInRetryPlan=true`
- [x] 3.4 实现 `retry_focus_test.go::TestSelectRetryFocusTurns` 表驱动覆盖全 needs_work / 全 strong / 混合（含 queued_for_retry）/ 超过 5 个 / 空数组
- [x] 3.5 新增 `backend/internal/review/next_action.go::decideNextAction(readiness, retryFocusCount) NextActionType`：readiness ∈ {not_ready, needs_practice} 且 retryFocusCount ≥ 1 → retry_current_round；readiness ∈ {basically_ready, well_prepared} 且 retryFocusCount < 3 → next_round；其他 → review_evidence（fallback）
- [x] 3.6 实现 `next_action_test.go::TestDecideNextAction` 表驱动覆盖 4 档 readiness × retry_focus_count 矩阵（0/1/2/3/5）
- [x] 3.7 Phase 3 收口：`cd backend && go test ./internal/review/... -count=1` 全部通过；4 档 readiness × N retry_focus × 决策矩阵 + property 测试 + 边界值覆盖

## Phase 4: 持久化 + outbox emit

- [x] 4.1 新增 `backend/internal/store/review/persist_report.go::PersistReport(ctx, input) (result, error)` 单事务：SELECT FOR UPDATE feedback_reports + UPDATE 设 status='ready' + UPDATE 内容字段（含 Phase 0.5 新增的 `language` / `feature_flag` / `data_source_version` / `retry_focus_turn_ids` 4 列）+ INSERT N 行 question_assessments + INSERT outbox `report.generated` + INSERT `audit_events` + UPDATE async_jobs.status='succeeded', completed_at=now(), locked_at=null
- [x] 4.2 新增 `backend/internal/store/review/persist_failure.go::PersistReportFailure(ctx, input) (result, error)` 单事务：UPDATE feedback_reports 设 status='failed' + error_code + generated_at + UPDATE async_jobs 使用当前 lease 后的 `attempts` 计算 available_at 退避 + status (queued/failed) + locked_at=null（failure finalize 不再二次递增 attempts）+ INSERT outbox `report.generation.failed` + INSERT `audit_events`；computeBackoff = `min(2^attempts * 30s, 30min)`（B4 列名 `attempts`）
- [x] 4.3 新增 `backend/internal/store/review/outbox_emitter.go::BuildReportGeneratedPayload` + `BuildReportGenerationFailedPayload` + `assertNoReviewOutboxPII` helper；payload 严格按 B3 `shared/events.yaml` schema
- [x] 4.4 在 `backend/internal/review/service.go::GenerateReport` 实现完整 orchestrate：advanceToGenerating → 拉 session/turns/plan/target_jobs（其中 `language` 取 `practice_sessions.language`）→ generateReportContent + assessQuestionsForAllTurns → computeReadinessTier + selectRetryFocusTurnIDs + decideNextAction → PersistReport 或 Failure 分支；F3 resolve / parse 失败时显式调 `aiclient.AITaskRunWriter.WriteAITaskRun(Status=AITaskRunStatusFailed, ValidationStatus=ValidationStatusInvalid 仅 parse 路径)`（A3 decorator 不会自动写这两类）
- [x] 4.5 实现 `persist_report_test.go::TestPersistReportWritesQuestionAssessments` 真实 Postgres 集成测试：断言 feedback_reports 字段（含 4 新列）+ N 行 question_assessments + outbox + audit + async_jobs.status='succeeded'+locked_at=null 全部在单事务内写入；UPDATE 行失败时整体回滚
- [x] 4.6 实现 `persist_failure_test.go::TestPersistReportFailureRetryAndPermanent` 真实 Postgres 集成测试：`attempts` 1/2/3/4 → status='queued' + locked_at=null；`attempts` 5 → status='failed'（permanent） + locked_at=null；available_at 按 backoff 计算
- [x] 4.7 实现 `outbox_emitter_test.go::TestReportGeneratedPayload` + `TestReportGenerationFailedPayload` 断言与 `shared/events/report.generated.json` / `report.generation.failed.json` schema 一致 + piiBoundary 通过；payload 不含 raw text
- [x] 4.8 BDD-Gate: 验证 `E2E.P0.052` 通过（report 主路径 happy path 含 ai_task_runs + outbox + question_assessments 写入；Go HTTP scenario `TestE2EP0052ReportGenerationHappyPath`）
- [x] 4.9 Phase 4 收口：`cd backend && go test ./internal/review/... ./internal/store/review/... -count=1` 全部通过；PersistReport / PersistReportFailure 单元 + 集成 + outbox PII 单元测试通过

## Phase 5: Read handler 接入

- [x] 5.1 新增 `backend/internal/api/reports/get_feedback_report.go`：通过 generated `ServerInterface` 注册 `GET /reports/{reportId}` handler；从 session middleware 拿 user_id；调用 service；不存在 / 越权 → 404 + ApiError{code:'REPORT_NOT_FOUND'}
- [x] 5.2 实现 `backend/internal/review/get_feedback_report_service.go::GetFeedbackReport` 与 `backend/internal/store/review/get_feedback_report.go::GetFeedbackReport(ctx, userID, reportID)` user-scoped read by reportId + join question_assessments；status placeholder 透传（queued/generating/failed 各自构造合法 wire shape）
- [x] 5.3 新增 `backend/internal/api/reports/list_target_job_reports.go`：通过 generated `ServerInterface` 注册 `GET /targets/{targetJobId}/reports` handler；解析 query `cursor` (base64 encoded `(created_at, id)` tuple) + `pageSize` (默认 20，最大 50)；cursor 非法 → 400 + VALIDATION_FAILED
- [x] 5.4 实现 `backend/internal/review/list_target_job_reports_service.go::ListTargetJobReports` 与 `backend/internal/store/review/list_target_job_reports.go::ListTargetJobReports(ctx, input)` cursor 分页 + user-scoped；SELECT * FROM feedback_reports WHERE target_job_id=$1 AND user_id=$2 AND (created_at, id) < ($cursor_created_at, $cursor_id) ORDER BY created_at DESC, id DESC LIMIT $pageSize+1；额外一行用于 hasMore 检测；返回 PaginatedFeedbackReport
- [x] 5.5 新增 `backend/internal/review/cursor.go::EncodeCursor` + `DecodeCursor`：编码契约 = `base64url(JSON({"createdAt": ISO8601(UTC, ns), "id": uuid}))`，无 padding；Decode 严格反序列化（拒绝 trailing bytes、未知 key、非法 ISO8601）；解析失败返回 ErrInvalidCursor；契约由 `TestCursorEncodeDecodeRoundTrip` / `TestCursorRejectsTampered` / `TestCursorRejectsLegacyFormat` 锁定
- [x] 5.6 在 `backend/internal/api/server.go`（或当前 router 入口）挂接 2 个 handler；替换可能存在的 `Reports` tag handler stub
- [x] 5.7 实现 `get_feedback_report_test.go`：5 case（ready 完整字段 / queued placeholder / generating placeholder / failed + errorCode / cross-user 404）；listTargetJobReports 6 case（空 / 单页 / 多页 + cursor / cursor decode round-trip / 非法 cursor 400 / 越权 user 隔离）
- [x] 5.8 BDD-Gate: 验证 `E2E.P0.053` 通过（status 转换 + 分页 + 404；Go HTTP scenario `TestE2EP0053ReportReadAndListing`）
- [x] 5.9 Phase 5 收口：`cd backend && go test ./internal/api/reports/... ./internal/review/... ./internal/store/review/... -count=1` 全部通过；read handler 200 + 404 + placeholder 路径完整

## Phase 6: 失败语义 / retry policy / 隐私 / observability / legacy-negative

- [x] 6.1 在 `service.go` 完善 6 路径 AI 失败 graceful matrix：F3 ErrPromptUnsupported / F3 ErrLanguageUnsupported / A3 secret missing / A3 timeout / A3 invalid output / parsed empty；每路径 mapped 到 B1 error code（F3 → AI_PROVIDER_CONFIG_INVALID；A3 secret missing → AI_PROVIDER_SECRET_MISSING；A3 timeout → AI_PROVIDER_TIMEOUT；A3 invalid output / parsed empty → AI_OUTPUT_INVALID）；写 feedback_reports.status='failed' + error_code；F3 resolve / parsed empty 时显式调 `aiclient.AITaskRunWriter.WriteAITaskRun(Status=AITaskRunStatusFailed, ValidationStatus=ValidationStatusInvalid 仅 parsed-empty 路径)`（A3 decorator 不会自动写 F3-before-A3 / parse-after-success 两类）
- [x] 6.2 实现 `generate_report_service_test.go::TestGenerateReportFailedMatrix` 表驱动覆盖 6 路径；每路径断言：(a) outcome.Status='failed' + 正确 B1 error_code；(b) `ai_task_runs.task_type` 区分（F3 主调失败 → `task_type='report_generate'`；F3 assessment 失败 → `task_type='report_assessment'`；A3 路径按当前阶段归属）；(c) `ai_task_runs.status='failed'` 字面量严格（B4 enum，非 'succeeded'）；(d) 写入路径（decorator vs service）通过 fake AIClient call counter + fake AITaskRunWriter 拦截区分
- [x] 6.3 实现 `runner_test.go::TestRunnerRetryPolicyAndPermanentFail` 覆盖 `async_jobs.attempts` 1→2→3→4 → status='queued' + locked_at=null + backoff 递增；attempts 5 → status='failed' + permanent + locked_at=null；fake clock 验证 available_at 计算；computeBackoff 参数命名为 `attempts`
- [x] 6.4 Redaction 单元测试：`feedback_reports.highlights/issues/next_actions` jsonb + `question_assessments.strengths/gaps/recommended_framework/dimension_results` jsonb + outbox payload + structured log + A3 metric label 不含 `question_text` / `answer_text` / `hint_text` / AI prompt body / response body / provider secret；用 `persist_report_test.go::TestPersistReportRedactsRawText` + `question_assessment_test.go::TestQuestionAssessmentParsesWithoutLeaks` + outbox emitter PII 单元测试覆盖
- [x] 6.5 Metric label allowlist 单元测试：A3 `ai_task_*` metric label 命中 F1 allowlist；不含 `feature_key` / prompt-rubric version / provider raw model id；ai_task_runs row `task_type='report_generate'` / `'report_assessment'` 合法（复用 A3 observed AIClient 既有 allowlist 测试 + 001 新增 report-specific assertion，包括 `status` ∈ B4 enum {`success`,`failed`,`timeout`,`fallback`} 字面量断言）
- [x] 6.6 新增 `scripts/lint/backend_review_legacy.py`：扫描 `backend/internal/review/`、`backend/internal/api/reports/`、`backend/internal/store/review/`、`openapi/fixtures/Reports/`、`test/scenarios/e2e/p0-{052,053,054,055}-*/`：`reportLayout` / 5 档 readiness / `readiness_score` numeric / `mistakes_queue` / `drill_builder` / `growth_center` / `report_timeline` / `report_form` / 旧 next_action 取值 / `review_method_version` / 错误列名 `leased_at` / `attempt_count` / `worker_id` / ai_task_runs 上下文的 `'succeeded'` 字面量 零出现；本 plan / BDD / test docs / spec prohibition rows 不属于实现 / runtime 范围；新增 `scripts/lint/backend_review_legacy_test.py` 覆盖 `test_backend_review_legacy_includes_terms` / `test_backend_review_legacy_allows_negative_docs`
- [x] 6.7 新增 `backend/internal/review/README.md`：简明 handoff 段落，记录 001 新增 endpoint / inline runner / 中间件挂法 / 接口签名 / handoff 给 frontend-report-dashboard 与 future backend-async-runner 的边界；引用 D-13 inline runner 与未来收干路径
- [x] 6.8 BDD-Gate: 验证 `E2E.P0.054` 通过（AI 失败 graceful + retry policy + ai_task_runs failed 行 + report.generation.failed event；Go HTTP scenario `TestE2EP0054ReportAIFailureAndRetry`）
- [x] 6.9 BDD-Gate: 验证 `E2E.P0.055` 通过（cross-user 404 隔离 + 隐私红线 + retired 术语 0 出现；Go HTTP scenario `TestE2EP0055ReportPrivacyAndLegacy`）
- [x] 6.10 收口 gate：`cd backend && go test ./... -count=1` 全绿；`make codegen-check`、`make validate-fixtures`、`migrations/lint.sh`（若 0.3/0.5 修订）、`make lint-events`、`make codegen-events-check`、`python3 scripts/lint/conventions_drift.py --repo-root .` 全通过
- [x] 6.11 `python3 scripts/lint/backend_review_legacy.py --repo-root . --phase all` 通过；`python3 -m pytest scripts/lint/backend_review_legacy_test.py -q` 通过
- [x] 6.12 更新 `docs/spec/backend-review/plans/INDEX.md`：001 状态保持 `active`（plan-review / sync-doc-index 推进到 completed 由后续动作完成）

## 收口证据

- `cd backend && go test ./internal/api/reports ./internal/review ./internal/store/review ./internal/ai/aiclient ./internal/ai/registry ./cmd/api -count=1`
- `cd backend && go test ./...`
- `cd backend && env 'DATABASE_URL=postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable' go test -tags=integration ./internal/store/review -run 'TestPersistReportWritesQuestionAssessments|TestPersistReportFailureRetryAndPermanent|TestLeaseSkipLocked' -count=1 -v`
- `make codegen-check`
- `make validate-fixtures`
- `migrations/lint.sh`
- `make lint-events`
- `make codegen-events-check`
- `python3 scripts/lint/conventions_drift.py --repo-root .`
- `python3 scripts/lint/backend_review_legacy.py --repo-root . --phase all`
- `python3 -m pytest scripts/lint/backend_review_legacy_test.py -q`
- `python3 scripts/lint/prompt_lint.py`
- `python3 scripts/lint/rubric_lint.py`
- `pnpm --filter @easyinterview/frontend typecheck`
- `make docs-check`
- `git diff --check`
