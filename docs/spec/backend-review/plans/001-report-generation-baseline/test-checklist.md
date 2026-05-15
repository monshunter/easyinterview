# 001 — Report Generation Baseline Test Checklist

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-05-16

**关联 Test Plan**: [test-plan](./test-plan.md)

## Phase 0: 跨 spec 前置修订 + Preflight

- [x] Phase 0 本计划定义的契约 / preflight 测试项全部通过：
  - `TestBaselineMigrationAcceptsReportAssessmentTaskType`
  - `TestFeedbackReportsContainsProvenancePersistenceColumns`
  - `TestValidTaskTypesIncludesReportAssessment`
  - `TestFeedbackReportSchemaIncludesErrorCode`
  - `TestGetFeedbackReportNotFoundResponseContract`
  - `TestReportsFixturesIncludeFailureAndEmptyVariants`
  - `TestF3ReportGenerateAndAssessmentPreflight`
  - `TestComputeReadinessTierScoreLevelsAndDimensionStatusMapping`
  - `make codegen-check`
  - `make validate-fixtures`
  - `migrations/lint.sh`
  - `python3 scripts/lint/conventions_drift.py --repo-root .`

## Phase 1: Inline review runner + lease + status state machine

- [x] Phase 1 本计划定义的 runner / lease / reaper 测试项全部通过：
  - `TestRunnerLeasesAndAdvancesToGenerating`
  - `TestRunnerSkipsAsyncJobUpdateWhenServiceFinalized`
  - `TestRunnerContinuesPollingAfterLeaseError`
  - `TestRunnerRetryPolicyAndPermanentFail`
  - `TestLeaseAsyncJobUsesAttemptsAndLockedAt`
  - `TestLeaseSkipLocked`
  - `TestUpdateAsyncJobSucceededClearsLockedAt`
  - `TestUpdateAsyncJobFailedClearsLockedAt`
  - `TestReviewJobTypeConstant`
  - `TestStatusStateMachineEnforcement`
  - `TestReaperReclaimsExpiredLease`（`backend/internal/review` + `backend/internal/store/review`）

## Phase 2: AI 调用与内容生成

- [x] Phase 2 本计划定义的 F3 / A3 调用与隐私测试项全部通过：
  - `TestGenerateReportContentSuccess`
  - `TestGenerateReportContentBuildsPromptWithoutLeaks`
  - `TestAssessQuestionsForAllTurns`
  - `TestAssessQuestionsBuildsPromptWithoutLeaks`
  - `TestQuestionAssessmentParsesWithoutLeaks`
  - `TestDecorator_ReportMetricLabelsExcludeProvenanceAndRawModelID`

## Phase 3: ReadinessTier / retry_focus / next_action 算法

- [x] Phase 3 本计划定义的算法测试项全部通过：
  - `TestComputeReadinessTier`
  - `TestComputeReadinessTierScoreLevelsAndDimensionStatusMapping`
  - `TestComputeReadinessTierPropertyRandomDimensions`
  - `TestSelectRetryFocusTurns`
  - `TestDecideNextAction`

## Phase 4: 持久化 + outbox emit

- [x] Phase 4 本计划定义的持久化 / outbox / service 编排测试项全部通过：
  - `TestPersistReportWritesQuestionAssessments`
  - `TestPersistReportFailureRetryAndPermanent`
  - `TestPersistReportRedactsRawText`
  - `TestReportGeneratedPayload`
  - `TestReportGenerationFailedPayload`
  - `TestReportOutboxPayloadRejectsPII`
  - `TestGenerateReportServiceOrchestratesAndPersists`
  - `TestGenerateReportFailedMatrix`

## Phase 5: Read handler 接入

- [x] Phase 5 本计划定义的 read handler / service / store / cursor 测试项全部通过：
  - `TestGetFeedbackReportHandlerReturnsReadyReport`
  - `TestGetFeedbackReportHandlerMapsCrossUserToReportNotFound`
  - `TestListTargetJobReportsHandlerParsesCursorAndPageSize`
  - `TestListTargetJobReportsHandlerMapsInvalidCursor`
  - `TestGetFeedbackReportServiceUserScopedNotFound`
  - `TestListTargetJobReportsServicePagination`
  - `TestGetFeedbackReportUserScopedWithAssessments`
  - `TestGetFeedbackReportMapsNoRowsToReportNotFound`
  - `TestListTargetJobReportsCursorPagination`
  - `TestCursorEncodeDecodeRoundTrip`
  - `TestCursorRejectsTampered`
  - `TestCursorRejectsLegacyFormat`

## Phase 6: 失败语义 / retry policy / 隐私 / observability / legacy-negative

- [x] Phase 6 本计划定义的 failure matrix / privacy / legacy-negative 测试项全部通过：
  - `TestGenerateReportFailedMatrix`
  - `TestRunnerRetryPolicyAndPermanentFail`
  - `TestPersistReportRedactsRawText`
  - `TestQuestionAssessmentParsesWithoutLeaks`
  - `TestReportOutboxPayloadRejectsPII`
  - `TestDecorator_ReportMetricLabelsExcludeProvenanceAndRawModelID`
  - `test_backend_review_legacy_includes_terms`
  - `test_backend_review_legacy_allows_negative_docs`
  - `python3 scripts/lint/backend_review_legacy.py --repo-root . --phase all`
  - `python3 -m pytest scripts/lint/backend_review_legacy_test.py -q`
  - `python3 scripts/lint/prompt_lint.py`
  - `python3 scripts/lint/rubric_lint.py`

## Runtime wiring L2 closure

- [x] `TestBuildReportRuntimeWiresRoutesRunnerReaperAndAI` 覆盖 `cmd/api` runtime 真实挂接：HTTP report routes、review runner、review reaper 与 report service 使用同一个 F3/A3-backed service。
- [x] `TestBuildReportRuntimeRejectsMissingAIClient` 防止 report runtime 在缺少 AI client 时静默降级为只读 handler。
- [x] `go test ./cmd/api -run 'TestBuildReportRuntime' -count=1` 通过。
- [x] `go test ./cmd/api ./internal/review ./internal/store/review -count=1` 通过。

## 全局收口

- [x] `cd backend && go test ./... -count=1`
- [x] `cd backend && env 'DATABASE_URL=postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable' go test -tags=integration ./internal/store/review -run 'TestPersistReportWritesQuestionAssessments|TestPersistReportFailureRetryAndPermanent|TestLeaseSkipLocked' -count=1 -v`
- [x] `make codegen-check`
- [x] `make validate-fixtures`
- [x] `migrations/lint.sh`
- [x] `make lint-events`
- [x] `make codegen-events-check`
- [x] `python3 scripts/lint/conventions_drift.py --repo-root .`
- [x] `python3 scripts/lint/backend_review_legacy.py --repo-root . --phase all`
- [x] `python3 -m pytest scripts/lint/backend_review_legacy_test.py -q`
- [x] `python3 scripts/lint/prompt_lint.py`
- [x] `python3 scripts/lint/rubric_lint.py`
- [x] `pnpm --filter @easyinterview/frontend typecheck`
- [x] `make docs-check`
- [x] `git diff --check`
