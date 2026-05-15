# 001 — Report Generation Baseline Test Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-15

**关联 Test Plan**: [test-plan](./test-plan.md)

## Phase 0: 跨 spec 前置修订 + Preflight

- [ ] Phase 0 本计划定义的单元测试项全部通过：
  - `TestBaselineMigrationAcceptsReportAssessmentTaskType` / `TestBaselineMigrationRejectsUnknownTaskType`（migration apply test，`backend/internal/migrations/`；命令入口 `cd backend && go test ./internal/migrations/... -count=1`）
  - `migrations/lint.sh`（`migrations/enum-sources.yaml` 与 SQL CHECK 同步；漏改 enum manifest 时必须失败）
  - `TestValidTaskTypesIncludesReportAssessment`（writers.go enum 与 `allowedAITaskRunCapabilities` 集合一致性 + `Validate()` 接受 `report_assessment`）
  - `TestFeedbackReportsContainsProvenancePersistenceColumns`（B4 baseline `feedback_reports` 含 4 新列 `language` / `feature_flag` / `data_source_version` / `retry_focus_turn_ids` + 默认值生效）
  - `TestFeedbackReportSchemaIncludesErrorCode`（B2 generated Go/TS struct 含 `ErrorCode` 字段；wire JSON marshal 在 nil 时省略或显式 null）
  - `TestF3ReportGenerateAndAssessmentPreflight`（F3 baseline + ResolveActive 可用性 + prompt body privacy：不要求 raw `{{question}}` / `{{answer}}` / `{{transcript}}`，不要求 `evidence_quotes` / verbatim quote；`backend/internal/ai/registry/backend_review_preflight_test.go`）
  - `TestMapScoreLevelToDimensionStatus`（weak/developing → needs_work；proficient → meets_bar；strong → strong；invalid score_level rejected）
  - `make validate-fixtures` + `make codegen-check` + `make docs-check` + `python3 scripts/lint/conventions_drift.py --repo-root .` 通过
  - generated Go `ErrReportNotFound` 常量 + generated TS 等价常量出现

## Phase 1: Inline review runner + lease + status state machine

- [ ] Phase 1 本计划定义的单元测试项全部通过：
  - `TestRunnerLeasesAndAdvancesToGenerating`（fake store + fake service 断言 lease 成功后 status `queued → generating`；断言 `locked_at` 已 set，不断言 `worker_id` 列）
  - `TestRunnerHandlesLeaseFailureGracefully`（lease 错误时 worker 不 panic，继续 poll）
  - `TestLeaseSkipLocked`（multi-goroutine 真实 Postgres 断言 SKIP LOCKED；SQL 使用 B4 列名 `attempts` / `locked_at`）
  - `TestStatusStateMachineEnforcement`（`queued → ready` 拒绝；`failed → ready` 拒绝；`generating → ready` / `generating → failed` / `failed → queued` 允许）
  - `TestReaperReclaimsExpiredLease`（fake clock + reaper 周期回收 stale running 行；UPDATE 条件 `locked_at < now() - lease_timeout AND job_type='report_generate'`）
  - `TestReaperDoesNotTouchActiveOrFinalized`（reaper 不动 succeeded / failed / 未超时 running 行；不动其它 job_type 如 `target_import`）

## Phase 2: AI 调用与内容生成

- [ ] Phase 2 本计划定义的单元测试项全部通过：
  - `TestGenerateReportContentSuccess`（fake F3 + fake AIClient 成功路径 → ReportContentDraft 含 highlights/issues/next_actions）
  - `TestGenerateReportContentBuildsPromptWithoutLeaks`（user message 不拼接 `question_text` / `answer_text` / `hint_text` 明文）
  - `TestGenerateReportContentUsesReportGenerateCapability`（fake AITaskRunWriter 捕获 capability=`report_generate`）
  - `TestAssessQuestionsForAllTurns`（fake F3 + fake AIClient 返回 N 行 outcome 按 turn_index 升序，每行含 dimension_results map + overall_status + confidence + strengths + gaps + recommended_framework + review_status）
  - `TestAssessQuestionsBuildsPromptWithoutLeaks`（反向断言）
  - `TestAssessQuestionsUsesReportAssessmentCapability`（fake AITaskRunWriter 捕获 capability=`report_assessment`）

## Phase 3: ReadinessTier / retry_focus / next_action 算法

- [ ] Phase 3 本计划定义的单元测试项全部通过：
  - `TestComputeReadinessTier`（表驱动 4 档阈值边界：测试值 0.29 / 0.30 / 0.54 / 0.55 / 0.74 / 0.75 / 0.99 → not_ready / needs_practice / basically_ready / well_prepared）
  - `TestMapScoreLevelToDimensionStatus`（score_level 仅作为内部评分输入，B2 wire status 只能输出 needs_work / meets_bar / strong）
  - `TestComputeReadinessTierProperty`（随机 N 维度 × N turn property 测试）
  - `TestComputeReadinessTierEmptyAssessmentsFallback`（空 assessments → not_ready）
  - `TestSelectRetryFocusTurns`（表驱动：全 needs_work / 全 strong / 混合含 queued_for_retry / 超过 5 个 / 空数组）
  - `TestSelectRetryFocusMutatesIncludedInRetryPlan`（mutate 对应行 IncludedInRetryPlan=true）
  - `TestDecideNextAction`（表驱动 4 档 readiness × retry_focus count 0/1/2/3/5 矩阵）

## Phase 4: 持久化 + outbox emit

- [ ] Phase 4 本计划定义的单元测试项全部通过：
  - `TestPersistReportSuccess`（真实 Postgres 集成：单事务写入 feedback_reports（含 4 新列 `language` / `feature_flag` / `data_source_version` / `retry_focus_turn_ids`）+ N 行 question_assessments + outbox + audit + `async_jobs.status='succeeded'` + `locked_at=null`）
  - `TestPersistReportWritesQuestionAssessments`（N 行写入 + UNIQUE(report_id, turn_id) 约束）
  - `TestPersistReportRollbackOnFailure`（事务失败回滚整体；feedback_reports / question_assessments / outbox 都没有部分写入）
  - `TestPersistReportRedactsRawText`（feedback_reports.highlights/issues/next_actions / question_assessments jsonb 字段不含 raw `question_text` / `answer_text` / `hint_text` literal）
  - `TestPersistReportFailureRetryAndPermanent`（输入 lease 后 `async_jobs.attempts`=1/2/3/4 → status='queued' + locked_at=null + backoff 递增且 attempts 不再变化；attempts=5 → status='failed' permanent + locked_at=null）
  - `TestPersistReportFailureWritesOutbox`（report.generation.failed event + payload schema）
  - `TestReportGeneratedPayload` / `TestReportGenerationFailedPayload`（B3 schema 一致 + piiBoundary）
  - `TestOutboxNoPII`（outbox payload 不含 raw text）
  - `TestServiceGenerateReportHappyPath` / `TestServiceGenerateReportFailureBranch` / `TestServiceWritesAITaskRunOnF3Failure`（`ai_task_runs.status='failed'` 字面量，B4 enum）/ `TestServiceWritesAITaskRunOnParseFailure`（`status='failed'` + `validation_status='invalid'` 仅 parsed-empty 路径）

## Phase 5: Read handler 接入

- [ ] Phase 5 本计划定义的单元测试项全部通过：
  - `TestGetFeedbackReportReady200`（status='ready' → 200 + 完整 FeedbackReport）
  - `TestGetFeedbackReportQueuedPlaceholder200` / `TestGetFeedbackReportGeneratingPlaceholder200`（200 + placeholder：preparednessLevel=null, 内容字段=空, provenance=null, errorCode=null）
  - `TestGetFeedbackReportFailed200WithErrorCode`（status='failed' → 200 + errorCode 非空 + 内容字段=空 + provenance=null）
  - `TestGetFeedbackReportCrossUser404`（用户 B 查询用户 A 的 reportX → 404 + REPORT_NOT_FOUND）
  - `TestListTargetJobReportsCursorPagination`（空 / 单页 / 多页 + cursor / cursor decode round-trip / 越权 user 隔离）
  - `TestListTargetJobReportsInvalidCursor400`（非法 cursor → 400 + VALIDATION_FAILED）
  - `TestCursorEncodeDecodeRoundTrip` / `TestCursorRejectsTampered` / `TestCursorRejectsLegacyFormat`（编码契约 = `base64url(JSON({createdAt: ISO8601, id: uuid}))` 无 padding；strict decode）
  - `TestFeedbackReportProvenanceFieldsAreWireOnly`（reflect + struct tag 6 字段）
  - `TestFeedbackReportProvenanceJSONShape`（json.Marshal → 6 keys 严格等于 B2 GenerationProvenance）
  - `TestProvenanceReadFromFeedbackReportsColumnsOnly`（反向断言：store.GetFeedbackReport 不 JOIN `ai_task_runs` 拼装 wire；6 字段全部源自 feedback_reports 列）
  - `TestErrReportNotFoundMapsTo404`（ErrReportNotFound → 404 + ApiError{code:'REPORT_NOT_FOUND'}）

## Phase 6: 失败语义 / retry policy / 隐私 / observability / legacy-negative

- [ ] Phase 6 本计划定义的单元测试项全部通过：
  - `TestGenerateReportFailedMatrix`（表驱动 6 路径：F3 ErrPromptUnsupported / F3 ErrLanguageUnsupported / A3 secret missing / A3 timeout / A3 invalid output / parsed empty；每路径 → outcome.Status='failed' + 正确 B1 error_code + `ai_task_runs.status='failed'` + `ai_task_runs.task_type` 正确归属 + 写入路径区分（decorator vs service））
  - `TestRunnerRetryPolicyAndPermanentFail`（fake clock 验证 lease 后 `async_jobs.attempts` + backoff；failure finalize 不二次递增；computeBackoff 参数命名 `attempts`）
  - `TestObservabilityReportRedaction`（log / metric label / ai_task_runs typed columns / feedback_reports jsonb / question_assessments jsonb 红线）
  - `TestObservabilityDecoratorWritesReportGenerateAndAssessmentTaskRuns`（A3 路径成功 → `ai_task_runs.status='success'` + 失败 → `status='failed'`；B4 enum 严格）
  - `test_backend_review_legacy_includes_terms` / `test_backend_review_legacy_allows_negative_docs`（pytest）+ `python3 scripts/lint/backend_review_legacy.py --repo-root . --phase all`（scoped legacy grep）
  - `make codegen-check` / `make validate-fixtures` / `make lint-events` / `make codegen-events-check` / `python3 scripts/lint/conventions_drift.py --repo-root .` 全部通过
  - `migrations/lint.sh` 通过（若 0.3/0.5 修订 B4）
  - `cd backend && go test ./... -count=1` 全绿

## 全局收口

- [ ] `cd backend && go test ./...`
- [ ] `make codegen-check`
- [ ] `make validate-fixtures`
- [ ] `migrations/lint.sh`
- [ ] `make lint-events`
- [ ] `make codegen-events-check`
- [ ] `python3 scripts/lint/conventions_drift.py --repo-root .`
- [ ] `python3 scripts/lint/backend_review_legacy.py --repo-root . --phase all`
- [ ] `python3 -m pytest scripts/lint/backend_review_legacy_test.py -q`
- [ ] `make docs-check`
- [ ] `git diff --check`
