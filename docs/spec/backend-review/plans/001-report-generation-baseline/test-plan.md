# 001 — Report Generation Baseline Test Plan

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-15

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)
**关联 Test Checklist**: [test-checklist](./test-checklist.md)

## 0 范围与覆盖矩阵映射

本 test-plan 把 plan §3.5 Coverage Matrix 的 R1–R22 + plan §4 Phase 0–6 的实施步骤映射到具体测试包、文件、测试函数与命令。覆盖矩阵中的 BDD 行（场景级）由 `bdd-plan.md` 承担，本文件聚焦单元 / 集成 / contract / drift / privacy 单元测试。

| 覆盖矩阵行 | 测试包 / 文件 | 测试函数 | 测试命令 |
|------------|---------------|----------|----------|
| R1 (primary, queued → generating → ready) | `backend/internal/review/runner_test.go` + `backend/internal/review/generate_report_service_test.go` + `backend/internal/store/review/persist_report_test.go` | `TestRunnerLeaseAndPersistReady` / `TestGenerateReportSuccess` / `TestPersistReportSuccess` | `cd backend && go test ./internal/review ./internal/store/review -count=1` |
| R2 (primary, question_assessments 写入) | `backend/internal/review/question_assessment_test.go` + `backend/internal/store/review/persist_report_test.go` | `TestAssessQuestionsForAllTurns` / `TestPersistReportWritesQuestionAssessments` | `cd backend && go test ./internal/review ./internal/store/review -count=1` |
| R3 (primary, ReadinessTier 算法) | `backend/internal/review/readiness_test.go` | `TestComputeReadinessTier`（表驱动 4 档边界）+ `TestMapScoreLevelToDimensionStatus`（weak/developing/proficient/strong → B2 DimensionStatus）+ `TestComputeReadinessTierProperty`（随机 N 维度） | `cd backend && go test ./internal/review -run 'TestComputeReadinessTier|TestMapScoreLevelToDimensionStatus' -count=1` |
| R4 (primary, retry_focus 选择) | `backend/internal/review/retry_focus_test.go` | `TestSelectRetryFocusTurns`（表驱动 needs_work × review_status × 数量边界） | `cd backend && go test ./internal/review -run 'TestSelectRetryFocusTurns' -count=1` |
| R5 (primary, next_action 决策) | `backend/internal/review/next_action_test.go` | `TestDecideNextAction`（表驱动 4 档 readiness × retry_focus count 矩阵） | `cd backend && go test ./internal/review -run 'TestDecideNextAction' -count=1` |
| R6 (alternate, status placeholder) | `backend/internal/api/reports/get_feedback_report_test.go` + `backend/internal/store/review/get_feedback_report_test.go` | `TestGetFeedbackReportReturnsPlaceholderForQueuedAndGenerating` / `TestGetFeedbackReportReturnsFailedWithErrorCode` | `cd backend && go test ./internal/api/reports ./internal/store/review -count=1` |
| R7 (alternate, listTargetJobReports 分页 + 空) | `backend/internal/api/reports/list_target_job_reports_test.go` + `backend/internal/store/review/list_target_job_reports_test.go` | `TestListTargetJobReportsCursorPagination`（空/单页/多页/cursor） | 同 R6 |
| R8 (failure, AI 失败 6 路径) | `backend/internal/review/generate_report_service_test.go` | `TestGenerateReportFailedMatrix`（F3 ErrPromptUnsupported / F3 ErrLanguageUnsupported / A3 secret_missing / A3 timeout / A3 invalid output / parsed empty） | `cd backend && go test ./internal/review -count=1` |
| R9 (failure, retry policy + max attempts) | `backend/internal/review/runner_test.go` + `backend/internal/store/review/persist_failure_test.go` | `TestRunnerRetryPolicyAndPermanentFail` / `TestPersistReportFailureRetryAndPermanent` | `cd backend && go test ./internal/review ./internal/store/review -count=1` |
| R10 (boundary, SKIP LOCKED 并发) | `backend/internal/store/review/lease_async_job_test.go` | `TestLeaseSkipLocked`（multi-goroutine 真实 Postgres） | 同 R6 |
| R11 (boundary, reaper) | `backend/internal/review/reaper_test.go` + `backend/internal/store/review/reaper_test.go` | `TestReaperReclaimsExpiredLease` / `TestReaperDoesNotTouchActiveOrFinalized` | `cd backend && go test ./internal/review ./internal/store/review -count=1` |
| R12 (boundary, cursor encode/decode) | `backend/internal/review/cursor_test.go` | `TestCursorEncodeDecodeRoundTrip` / `TestCursorRejectsTampered` / `TestCursorRejectsLegacyFormat` | `cd backend && go test ./internal/review -count=1` |
| R13 (cross-layer, B2 schema + provenance wire) | `backend/internal/api/reports/feedback_report_provenance_test.go` + `make codegen-check` + `python3 scripts/lint/conventions_drift.py` | `TestFeedbackReportProvenanceFieldsAreWireOnly` / `TestFeedbackReportProvenanceJSONShape` / `TestProvenanceReadFromFeedbackReportsColumnsOnly`（反向断言：read 路径不 JOIN `ai_task_runs`） | `cd backend && go test ./internal/api/reports -count=1`; repo-root `make codegen-check`; repo-root `python3 scripts/lint/conventions_drift.py --repo-root .` |
| R14 (cross-layer, B3 event payload schema) | `backend/internal/store/review/outbox_emitter_test.go` | `TestReportGeneratedPayload` / `TestReportGenerationFailedPayload` / `TestOutboxNoPII` | `cd backend && go test ./internal/store/review -count=1` |
| R15 (cross-layer, B1 REPORT_NOT_FOUND + B2 404 schema + B2 FeedbackReport.errorCode + spec history) | `backend/internal/api/reports/error_mapping_test.go` + `backend/internal/api/reports/feedback_report_errorcode_test.go` + `scripts/lint/conventions_drift.py` + `make codegen-check` + `make docs-check` | `TestErrReportNotFoundMapsTo404` / `TestFeedbackReportSchemaIncludesErrorCode` / spec history append 文件存在断言 | `cd backend && go test ./internal/api/reports -count=1`; repo-root `make codegen-check`; repo-root `make docs-check`; repo-root `python3 scripts/lint/conventions_drift.py --repo-root .` |
| R16 (cross-layer, B4 ai_task_runs.task_type CHECK + feedback_reports 4 新列 + writers.go enum + spec history) | `backend/internal/migrations/`（既有 baseline 契约测试位于 `sql_contract_test.go`，新增 `baseline_backend_review_rebase_test.go` 或扩展现文件）+ `backend/internal/ai/aiclient/writers_test.go` + `migrations/lint.sh` + `migrations/enum-sources.yaml` | `TestBaselineMigrationAcceptsReportAssessmentTaskType` / `TestBaselineMigrationRejectsUnknownTaskType` / `TestFeedbackReportsContainsProvenancePersistenceColumns` / `TestValidTaskTypesIncludesReportAssessment` + drift lint | `cd backend && go test ./internal/migrations ./internal/ai/aiclient -count=1`; repo-root `migrations/lint.sh`; repo-root `python3 scripts/lint/conventions_drift.py --repo-root .` |
| R17 (cross-layer, F3 preflight) | `backend/internal/ai/registry/backend_review_preflight_test.go` | `TestF3ReportGenerateAndAssessmentPreflight`（ResolveActive + prompt body privacy + no verbatim evidence quote + rubric score_levels + score_level→B2 DimensionStatus readiness coverage） | `cd backend && go test ./internal/ai/registry ./internal/review -run 'TestF3ReportGenerateAndAssessmentPreflight|TestMapScoreLevelToDimensionStatus' -count=1` |
| R18 (privacy, jsonb + cross-user FK) | `backend/internal/store/review/persist_report_test.go` + `backend/internal/review/question_assessment_test.go` + `backend/internal/api/reports/get_feedback_report_test.go` + outbox emitter PII 单测 | `TestPersistReportRedactsRawText` / `TestQuestionAssessmentParsesWithoutLeaks` / `TestCrossUserNotFound` | `cd backend && go test ./internal/store/review ./internal/review ./internal/api/reports -count=1` |
| R19 (observability, ai_task_runs typed columns + metric allowlist) | `backend/internal/ai/aiclient/observability/decorator_test.go`（扩展 report A3 路径）+ `backend/internal/review/generate_report_service_test.go` | `TestObservabilityDecoratorWritesReportGenerateAndAssessmentTaskRuns` / `TestGenerateReportRecordsAITaskRunFromA3Decorator` / `TestGenerateReportWritesAITaskRunOnF3Failure` | `cd backend && go test ./internal/ai/aiclient/observability ./internal/review -count=1` |
| R20 (UX/API status code) | `backend/internal/api/reports/get_feedback_report_test.go` + `backend/internal/api/reports/list_target_job_reports_test.go` | `TestGetFeedbackReportReady200` / `TestGetFeedbackReportQueuedPlaceholder200` / `TestGetFeedbackReportFailed200WithErrorCode` / `TestGetFeedbackReportCrossUser404` / `TestListTargetJobReportsInvalidCursor400` | `cd backend && go test ./internal/api/reports -count=1` |
| R21 (regression, scoped legacy grep) | `scripts/lint/backend_review_legacy.py`（新增）+ `scripts/lint/backend_review_legacy_test.py`（pytest） | `test_backend_review_legacy_includes_terms` / `test_backend_review_legacy_allows_negative_docs` | `python3 scripts/lint/backend_review_legacy.py --repo-root . --phase all && python3 -m pytest scripts/lint/backend_review_legacy_test.py -q` |
| R22 (out-of-scope, 002/003 owner 范围) | `backend/internal/review/`（grep）+ unit 断言 | grep `regenerateReport` / `retryReport` / `retainReportDays` 在 backend-review 包零出现 | manual grep + scope unit test |

## Phase 0: 跨 spec 前置修订 + Preflight

- **测试目标**：B1 `REPORT_NOT_FOUND` 错误码已新增到 conventions.yaml errors + generated Go/TS 常量；B2 `getFeedbackReport` 404 response 与 `REPORT_NOT_FOUND` 关联；B2 `FeedbackReport.errorCode` 字段已被 codegen 出；B4 `ai_task_runs.task_type` CHECK 接受 `report_assessment`（无条件扩）；B4 `feedback_reports` 含 `language` / `feature_flag` / `data_source_version` / `retry_focus_turn_ids` 4 新列（无条件扩）；writers.go enum 与 `allowedAITaskRunCapabilities` 集合一致；fixtures 命名场景通过 validator；F3 `report.generate` / `report.question_assessment` baseline 可被 ResolveActive 找回，且 prompt body 不要求 raw `{{question}}` / `{{answer}}` / `{{transcript}}` 输入、不要求 `evidence_quotes` / verbatim quote 输出；rubric score_levels 与 score_level→B2 DimensionStatus 映射测试就位；`make docs-check` 通过。若 F3 preflight fail，Phase 0 停止并回 F3 owner。
- **测试文件**：
  - `backend/internal/migrations/baseline_backend_review_rebase_test.go`（新增）或在 `sql_contract_test.go` 中扩展
  - `backend/internal/ai/aiclient/writers_test.go`（扩展 `TestValidTaskTypesIncludesReportAssessment`）
  - `backend/internal/ai/registry/backend_review_preflight_test.go`（新增 `TestF3ReportGenerateAndAssessmentPreflight`）
  - `backend/internal/api/reports/feedback_report_errorcode_test.go`（新增 `TestFeedbackReportSchemaIncludesErrorCode`）
  - `openapi/fixtures/Reports/getFeedbackReport.json`（fixtures，新增 `report-failed` variant）
  - `openapi/fixtures/Reports/listTargetJobReports.json`（fixtures，新增 `empty` variant）
- **测试命令**：
  - `cd backend && go test ./internal/migrations ./internal/ai/aiclient ./internal/ai/registry ./internal/api/reports -count=1`
  - `make validate-fixtures`
  - `make codegen-check`
  - `migrations/lint.sh`
  - `python3 scripts/lint/conventions_drift.py --repo-root .`
  - `make docs-check`
- **预期 Red / Green 证据**：
  - Red：修订 B1 / B2 / B4 / writers.go 前，新测试 fail（`REPORT_NOT_FOUND` 未注册 / `FeedbackReport.errorCode` 字段缺失 / CHECK 不含 `report_assessment` / `feedback_reports` 4 新列缺失 / 常量未定义 / fixture validator 缺少命名场景）
  - Green：Phase 0 完成后所有命令通过

## Phase 1: Inline review runner + lease + status state machine

- **测试目标**：Runner.Start 启动 poll loop（worker 身份通过 logger 字段，不写 DB 列）；LeaseAsyncJob 用 SKIP LOCKED 单 worker lease，SQL 使用 B4 列名 `attempts` / `locked_at`（不写 `leased_at` / `attempt_count` / `worker_id`）；status state machine 强制 D-7 单调迁移；reaper 周期回收 stale running 行（仅作用于 `job_type='report_generate'`，不动 `targetjob.Drainer` 处理的其它 job_type）；多 worker 并发不重复抢占。
- **测试文件**：
  - `backend/internal/review/runner_test.go`（新增）：`TestRunnerLeasesAndAdvancesToGenerating` / `TestRunnerHandlesLeaseFailureGracefully`
  - `backend/internal/store/review/lease_async_job_test.go`（新增）：`TestLeaseSkipLocked`（multi-goroutine 真实 Postgres）
  - `backend/internal/store/review/feedback_reports_status_test.go`（新增）：`TestStatusStateMachineEnforcement`（D-7 单调迁移）
  - `backend/internal/review/reaper_test.go` + `backend/internal/store/review/reaper_test.go`（新增）：`TestReaperReclaimsExpiredLease` / `TestReaperDoesNotTouchActiveOrFinalized`
- **测试命令**：
  - `cd backend && go test ./internal/review ./internal/store/review -count=1`
- **预期 Red / Green 证据**：
  - Red：包未实现前测试 fail（lease 函数不存在 / status state machine 接受非法迁移 / reaper 不存在）
  - Green：Phase 1 完成后命令通过

## Phase 2: AI 调用与内容生成

- **测试目标**：generateReportContent 通过 F3 + A3 调用解析合法 JSON 返回 ReportContentDraft；assessQuestionsForAllTurns 对每个 turn 独立调用；prompt user message 不含 raw text；A3 observability decorator 自动写 `ai_task_runs` 行。
- **测试文件**：
  - `backend/internal/review/generate_report.go`（新增；主调实现）
  - `backend/internal/review/generate_report_test.go`（新增）：`TestGenerateReportContentSuccess` / `TestGenerateReportContentBuildsPromptWithoutLeaks` / `TestGenerateReportContentUsesReportGenerateCapability`
  - `backend/internal/review/question_assessment.go`（新增；逐题评估实现）
  - `backend/internal/review/question_assessment_test.go`（新增）：`TestAssessQuestionsForAllTurns` / `TestAssessQuestionsBuildsPromptWithoutLeaks` / `TestAssessQuestionsUsesReportAssessmentCapability`
- **测试命令**：
  - `cd backend && go test ./internal/review -count=1`
- **预期 Red / Green 证据**：
  - Red：Phase 1 完成后 `TestGenerateReportContentSuccess` 等 fail（函数未实现）
  - Green：Phase 2 完成后命令通过；A3 decorator 写 `ai_task_runs(task_type='report_generate' or 'report_assessment')` succeeded 行

## Phase 3: ReadinessTier / retry_focus / next_action 算法

- **测试目标**：computeReadinessTier 加权阈值正确映射 4 档；score_level helper 将 weak/developing/proficient/strong 映射到 B2 DimensionStatus（needs_work/meets_bar/strong）且禁止把 score_level 字面量写入 wire status；selectRetryFocusTurnIDs 选择 needs_work + queued_for_retry 最多 5 个并升序；decideNextAction 4×N 矩阵决策。
- **测试文件**：
  - `backend/internal/review/readiness.go`（新增）+ `backend/internal/review/readiness_test.go`（新增）：`TestComputeReadinessTier`（表驱动 4 档边界 0.29/0.30/0.54/0.55/0.74/0.75/0.99）+ `TestMapScoreLevelToDimensionStatus`（weak/developing → needs_work；proficient → meets_bar；strong → strong；invalid score_level rejected）+ `TestComputeReadinessTierProperty`（随机 N 维度 × N turn）+ `TestComputeReadinessTierEmptyAssessmentsFallback`
  - `backend/internal/review/retry_focus.go`（新增）+ `backend/internal/review/retry_focus_test.go`（新增）：`TestSelectRetryFocusTurns`（全 needs_work / 全 strong / 混合含 queued_for_retry / 超过 5 个 / 空数组）+ `TestSelectRetryFocusMutatesIncludedInRetryPlan`
  - `backend/internal/review/next_action.go`（新增）+ `backend/internal/review/next_action_test.go`（新增）：`TestDecideNextAction`（4 档 readiness × retry_focus count 0/1/2/3/5）
- **测试命令**：
  - `cd backend && go test ./internal/review -count=1`
- **预期 Red / Green 证据**：
  - Red：Phase 2 完成后算法测试 fail
  - Green：Phase 3 完成后命令通过；4 档阈值 + 决策矩阵覆盖

## Phase 4: 持久化 + outbox emit

- **测试目标**：PersistReport 单事务写入 feedback_reports（含 Phase 0.5 新增的 4 列：`language` / `feature_flag` / `data_source_version` / `retry_focus_turn_ids`）+ N 行 question_assessments + outbox + audit + async_jobs (`status='succeeded'` + `locked_at=null`)；PersistReportFailure 单事务更新 report failed + 使用 lease 后当前 `attempts` 计算退避/status + `locked_at=null` + outbox failed event（failure finalize 不二次递增 attempts）；outbox payload 与 B3 schema 一致且无 PII；Service.GenerateReport orchestrate 完整流程（`language` 取自 `practice_sessions.language`）；F3 / parsed-empty 失败时显式写 ai_task_runs row (`Status=AITaskRunStatusFailed`，B4 `status='failed'`)。
- **测试文件**：
  - `backend/internal/store/review/persist_report.go`（新增）+ `backend/internal/store/review/persist_report_test.go`（新增；真实 Postgres 集成）：`TestPersistReportSuccess` / `TestPersistReportWritesQuestionAssessments` / `TestPersistReportRollbackOnFailure` / `TestPersistReportRedactsRawText`
  - `backend/internal/store/review/persist_failure.go`（新增）+ `backend/internal/store/review/persist_failure_test.go`（新增）：`TestPersistReportFailureRetryAndPermanent`（B4 列名 `attempts` 1/2/3/4 → queued + backoff + locked_at=null；attempts 5 → failed permanent + locked_at=null） / `TestPersistReportFailureWritesOutbox`
  - `backend/internal/store/review/outbox_emitter.go`（新增）+ `backend/internal/store/review/outbox_emitter_test.go`（新增）：`TestReportGeneratedPayload` / `TestReportGenerationFailedPayload` / `TestOutboxNoPII`
  - `backend/internal/review/service.go`（实现 GenerateReport orchestrate）+ `backend/internal/review/generate_report_service_test.go`（扩展）：`TestServiceGenerateReportHappyPath` / `TestServiceGenerateReportFailureBranch` / `TestServiceWritesAITaskRunOnF3Failure` / `TestServiceWritesAITaskRunOnParseFailure`
- **测试命令**：
  - `cd backend && go test ./internal/review ./internal/store/review -count=1`
- **预期 Red / Green 证据**：
  - Red：repository 未实现前测试 fail；outbox payload schema mismatch
  - Green：Phase 4 完成后命令通过；BDD `E2E.P0.052` 通过

## Phase 5: Read handler 接入

- **测试目标**：getFeedbackReport 按 status 返回正确 wire shape（ready/queued/generating/failed/cross-user 404）；listTargetJobReports cursor 分页 + 空列表 + 越权隔离 + 非法 cursor 400；cursor encode/decode round-trip。
- **测试文件**：
  - `backend/internal/api/reports/get_feedback_report.go`（新增；handler）+ `backend/internal/api/reports/get_feedback_report_test.go`（新增）：5 case
  - `backend/internal/api/reports/list_target_job_reports.go`（新增；handler）+ `backend/internal/api/reports/list_target_job_reports_test.go`（新增）：6 case
  - `backend/internal/review/get_feedback_report_service.go` + `backend/internal/review/list_target_job_reports_service.go`（新增；service）
  - `backend/internal/store/review/get_feedback_report.go` + `backend/internal/store/review/list_target_job_reports.go`（新增；store）
  - `backend/internal/review/cursor.go`（新增）+ `backend/internal/review/cursor_test.go`（新增）：`TestCursorEncodeDecodeRoundTrip` / `TestCursorRejectsTampered`
  - `backend/internal/api/reports/feedback_report_provenance_test.go`（新增）：`TestFeedbackReportProvenanceFieldsAreWireOnly` / `TestFeedbackReportProvenanceJSONShape`（reflect + json.Marshal 断言）
  - `backend/internal/api/reports/error_mapping_test.go`（新增）：`TestErrReportNotFoundMapsTo404`
- **测试命令**：
  - `cd backend && go test ./internal/api/reports ./internal/review ./internal/store/review -count=1`
- **预期 Red / Green 证据**：
  - Red：handler 未实现前测试 fail
  - Green：Phase 5 完成后命令通过；BDD `E2E.P0.053` 通过

## Phase 6: 失败语义 / retry policy / 隐私 / observability / legacy-negative

- **测试目标**：6 路径 AI 失败 graceful matrix；retry policy + permanent fail；log / metric label / audit / ai_task_runs / feedback_reports jsonb / question_assessments jsonb / outbox payload 不含 raw text；F1 metric label allowlist 命中；scoped legacy grep 反查实现 / runtime 输出中的 retired 术语。
- **测试文件**：
  - `backend/internal/review/generate_report_service_test.go`（扩展）：`TestGenerateReportFailedMatrix`（表驱动 6 路径）+ `TestGenerateReportRecordsAITaskRunFromA3Decorator` + `TestGenerateReportWritesAITaskRunOnF3Failure` + `TestGenerateReportWritesAITaskRunOnParseFailure`
  - `backend/internal/review/runner_test.go`（扩展）：`TestRunnerRetryPolicyAndPermanentFail`（fake clock 验证 backoff 与 `async_jobs.attempts`（B4 列名））
  - `backend/internal/review/observability_test.go`（新增）：`TestObservabilityReportRedaction`
  - `backend/internal/ai/aiclient/observability/decorator_test.go`（扩展）：`TestObservabilityDecoratorWritesReportGenerateAndAssessmentTaskRuns`
  - `scripts/lint/backend_review_legacy.py`（新增）+ `scripts/lint/backend_review_legacy_test.py`（pytest 新增）：`test_backend_review_legacy_includes_terms` + `test_backend_review_legacy_allows_negative_docs`
  - `backend/internal/review/README.md`（新增；handoff doc）
  - `backend/cmd/api/reports_http_scenario_test.go`（新增 4 个 Go HTTP scenario）：`TestE2EP0052ReportGenerationHappyPath` / `TestE2EP0053ReportReadAndListing` / `TestE2EP0054ReportAIFailureAndRetry` / `TestE2EP0055ReportPrivacyAndLegacy`
- **测试命令**：
  - `cd backend && go test ./internal/review ./internal/store/review ./internal/api/reports ./internal/ai/aiclient/observability ./cmd/api -count=1`
  - `python3 scripts/lint/backend_review_legacy.py --repo-root . --phase all`
  - `python3 -m pytest scripts/lint/backend_review_legacy_test.py -q`
- **预期 Red / Green 证据**：
  - Red：未补 6 路径 / retry policy / scoped legacy 反查项时 test fail
  - Green：Phase 6 完成后 plan 收口命令全绿

## 全局收口测试命令

```bash
cd backend && go test ./...
make codegen-check
make validate-fixtures
migrations/lint.sh
make lint-events
make codegen-events-check
python3 scripts/lint/conventions_drift.py --repo-root .
python3 scripts/lint/backend_review_legacy.py --repo-root . --phase all
python3 -m pytest scripts/lint/backend_review_legacy_test.py -q
make docs-check
git diff --check
```

不引入硬编码覆盖率门槛（observational only）。
