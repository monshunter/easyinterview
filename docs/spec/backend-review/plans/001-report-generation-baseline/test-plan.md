# 001 — Report Generation Baseline Test Plan

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-05-16

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)
**关联 Test Checklist**: [test-checklist](./test-checklist.md)

## 0 范围与证据原则

本 test-plan 将 backend-review spec v1.0 的 R1-R22 覆盖矩阵映射到当前仓库实际存在的测试函数、contract gate 与 drift gate。执行结果由 [test-checklist](./test-checklist.md) 记录；本文件只定义覆盖范围、测试入口与完成口径。

完成口径：

- 所有列出的 Go 单元 / 集成 / HTTP scenario 测试必须真实执行通过，不能接受 `[no tests to run]` 证据。
- Contract / fixture / migration / event / legacy-negative / prompt-rubric / docs gate 必须在 repo root 或指定工作目录执行通过。
- 不设置硬编码覆盖率百分比门槛；覆盖率只可作为观察信息。

## 1 覆盖矩阵

| 覆盖行 | 范围 | 当前可执行证据 |
|--------|------|----------------|
| R1 | queued → generating → ready 主路径 | `TestRunnerLeasesAndAdvancesToGenerating`, `TestGenerateReportServiceOrchestratesAndPersists`, `TestPersistReportWritesQuestionAssessments`, `TestE2EP0052ReportGenerationHappyPath` |
| R2 | question_assessments 写入 | `TestAssessQuestionsForAllTurns`, `TestPersistReportWritesQuestionAssessments`, `TestE2EP0052ReportGenerationHappyPath` |
| R3 | ReadinessTier / DimensionStatus 映射 | `TestComputeReadinessTier`, `TestComputeReadinessTierScoreLevelsAndDimensionStatusMapping`, `TestComputeReadinessTierPropertyRandomDimensions` |
| R4 | retry_focus 选择 | `TestSelectRetryFocusTurns`, `TestE2EP0052ReportGenerationHappyPath` |
| R5 | next_action 决策 | `TestDecideNextAction`, `TestE2EP0052ReportGenerationHappyPath` |
| R6 | queued/generating/failed placeholder | `TestGetFeedbackReportHandlerReturnsReadyReport`, `TestGetFeedbackReportHandlerMapsCrossUserToReportNotFound`, `TestGetFeedbackReportUserScopedWithAssessments`, `TestE2EP0053ReportReadAndListing` |
| R7 | listTargetJobReports 分页 / 空列表 / target ownership | `TestListTargetJobReportsHandlerParsesCursorAndPageSize`, `TestListTargetJobReportsCursorPagination`, `TestListTargetJobReportsRequiresOwnedTarget`, `TestReportsFixturesIncludeFailureAndEmptyVariants`, `TestE2EP0053ReportReadAndListing`, `TestE2EP0055ReportPrivacyAndLegacy` |
| R8 | AI 失败 6 路径 | `TestGenerateReportFailedMatrix`, `TestE2EP0054ReportAIFailureAndRetry` |
| R9 | retry policy + permanent fail | `TestRunnerRetryPolicyAndPermanentFail`, `TestPersistReportFailureRetryAndPermanent`, `TestE2EP0054ReportAIFailureAndRetry` |
| R10 | SKIP LOCKED 并发 | `TestLeaseSkipLocked`, integration command in checklist |
| R11 | reaper stale lease 回收 | `TestReaperReclaimsExpiredLease` in `backend/internal/review` and `backend/internal/store/review` |
| R12 | cursor encode/decode | `TestCursorEncodeDecodeRoundTrip`, `TestCursorRejectsTampered`, `TestCursorRejectsLegacyFormat`, `TestListTargetJobReportsHandlerMapsInvalidCursor` |
| R13 | B2 schema + provenance wire | `TestFeedbackReportSchemaIncludesErrorCode`, `TestGetFeedbackReportNotFoundResponseContract`, `TestGetFeedbackReportUserScopedWithAssessments`, `make codegen-check`, `python3 scripts/lint/conventions_drift.py --repo-root .` |
| R14 | B3 event payload schema | `TestReportGeneratedPayload`, `TestReportGenerationFailedPayload`, `TestReportOutboxPayloadRejectsPII`, `make lint-events`, `make codegen-events-check` |
| R15 | `REPORT_NOT_FOUND` + 404 contract | `TestGetFeedbackReportNotFoundResponseContract`, `TestGetFeedbackReportHandlerMapsCrossUserToReportNotFound`, `TestGetFeedbackReportMapsNoRowsToReportNotFound`, `make codegen-check` |
| R16 | B4 enum + feedback_reports provenance columns | `TestBaselineMigrationAcceptsReportAssessmentTaskType`, `TestFeedbackReportsContainsProvenancePersistenceColumns`, `TestValidTaskTypesIncludesReportAssessment`, `migrations/lint.sh` |
| R17 | F3 report prompt/rubric preflight + output schema alignment | `TestF3ReportGenerateAndAssessmentPreflight`, `TestAssessQuestionsMapsScoreLevelToWireStatus`, `python3 scripts/lint/prompt_lint.py`, `python3 scripts/lint/rubric_lint.py` |
| R18 | Privacy / cross-user isolation | `TestPersistReportRedactsRawText`, `TestQuestionAssessmentParsesWithoutLeaks`, `TestReportOutboxPayloadRejectsPII`, `TestGetFeedbackReportHandlerMapsCrossUserToReportNotFound`, `TestListTargetJobReportsRequiresOwnedTarget`, `TestE2EP0055ReportPrivacyAndLegacy` |
| R19 | Observability / ai_task_runs typed columns | `TestDecorator_ReportMetricLabelsExcludeProvenanceAndRawModelID`, `TestGenerateReportFailedMatrix`, `TestE2EP0054ReportAIFailureAndRetry` |
| R20 | API status code and placeholder UX | `TestGetFeedbackReportHandlerReturnsReadyReport`, `TestListTargetJobReportsHandlerMapsInvalidCursor`, `TestE2EP0053ReportReadAndListing` |
| R21 | scoped legacy-negative grep | `python3 scripts/lint/backend_review_legacy.py --repo-root . --phase all`, `test_backend_review_legacy_includes_terms`, `test_backend_review_legacy_allows_negative_docs` |
| R22 | 002/003 owner out-of-scope boundary | `python3 scripts/lint/backend_review_legacy.py --repo-root . --phase all`, `TestE2EP0055ReportPrivacyAndLegacy` |

## 2 分阶段执行入口

### Phase 0: Contract preflight

- `cd backend && go test ./internal/migrations ./internal/ai/aiclient ./internal/ai/registry ./internal/api/reports -count=1`
- `cd backend && go test ./internal/ai/registry -run TestF3ReportGenerateAndAssessmentPreflight -count=1`
- `cd backend && go test ./internal/review -run TestAssessQuestionsMapsScoreLevelToWireStatus -count=1`
- `make codegen-check`
- `make validate-fixtures`
- `migrations/lint.sh`
- `python3 scripts/lint/conventions_drift.py --repo-root .`

### Phase 1: Runner / lease / reaper

- `cd backend && go test ./internal/review ./internal/store/review -count=1`
- `cd backend && env 'DATABASE_URL=postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable' go test -tags=integration ./internal/store/review -run 'TestPersistReportWritesQuestionAssessments|TestPersistReportFailureRetryAndPermanent|TestLeaseSkipLocked' -count=1 -v`

### Phase 2: AI content generation

- `cd backend && go test ./internal/review ./internal/ai/aiclient/observability -count=1`

### Phase 3: Algorithms

- `cd backend && go test ./internal/review -run 'TestComputeReadinessTier|TestSelectRetryFocusTurns|TestDecideNextAction' -count=1`

### Phase 4: Persistence and outbox

- `cd backend && go test ./internal/review ./internal/store/review -count=1`
- Integration command from Phase 1 must include `TestPersistReportWritesQuestionAssessments` and `TestPersistReportFailureRetryAndPermanent`.
- L2 stale-state gate: `cd backend && env 'DATABASE_URL=postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable' go test -tags=integration ./internal/store/review -run 'TestPersistReportRejectsStaleStatusAndRollsBack|TestPersistReportFailureRejectsStaleStatusAndRollsBack' -count=1 -v`

### Phase 5: Read handlers

- `cd backend && go test ./internal/api/reports ./internal/review ./internal/store/review -count=1`
- `cd backend && go test ./internal/store/review -run 'TestListTargetJobReportsCursorPagination|TestListTargetJobReportsRequiresOwnedTarget' -count=1`
- `cd backend && go test ./cmd/api -run 'TestE2EP0053ReportReadAndListing' -count=1`

### Phase 6: Failure / privacy / legacy-negative

- `cd backend && go test ./internal/review ./internal/store/review ./internal/api/reports ./internal/ai/aiclient/observability ./cmd/api -count=1`
- `python3 scripts/lint/backend_review_legacy.py --repo-root . --phase all`
- `python3 -m pytest scripts/lint/backend_review_legacy_test.py -q`
- `python3 scripts/lint/prompt_lint.py`
- `python3 scripts/lint/rubric_lint.py`

## 3 Runtime wiring gate

L2 review 追加 runtime wiring gate，防止 `cmd/api` 只挂 read handler 而没有启动 report worker：

- `TestBuildReportRuntimeWiresRoutesRunnerReaperAndAI`
- `TestBuildReportRuntimeRejectsMissingAIClient`
- `cd backend && go test ./cmd/api -run 'TestBuildReportRuntime' -count=1`

## 3.5 L2 remediation gate

L2 review 追加 report contract / persistence / privacy gate，防止 prompt registry、store 状态机和 target-scoped read 只覆盖 happy path：

- `TestF3ReportGenerateAndAssessmentPreflight` 断言 prompt output keys 与 runtime draft struct 对齐。
- `TestAssessQuestionsMapsScoreLevelToWireStatus` 断言 `score_level` fallback 写入 B2 wire `DimensionStatus`。
- `TestPersistReportRejectsStaleStatusAndRollsBack` / `TestPersistReportFailureRejectsStaleStatusAndRollsBack` 断言 stale terminal status 的 zero-row update 回滚整事务。
- `TestListTargetJobReportsRequiresOwnedTarget` 断言 target 不属于当前用户时返回 `ErrReportNotFound`。
- `TestE2EP0053ReportReadAndListing` / `TestE2EP0055ReportPrivacyAndLegacy` 断言 cross-user target list 404。

## 4 全局收口命令

```bash
cd backend && go test ./... -count=1
cd backend && env 'DATABASE_URL=postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable' go test -tags=integration ./internal/store/review -run 'TestPersistReportWritesQuestionAssessments|TestPersistReportFailureRetryAndPermanent|TestLeaseSkipLocked' -count=1 -v
make codegen-check
make validate-fixtures
migrations/lint.sh
make lint-events
make codegen-events-check
python3 scripts/lint/conventions_drift.py --repo-root .
python3 scripts/lint/backend_review_legacy.py --repo-root . --phase all
python3 -m pytest scripts/lint/backend_review_legacy_test.py -q
python3 scripts/lint/prompt_lint.py
python3 scripts/lint/rubric_lint.py
pnpm --filter @easyinterview/frontend typecheck
make docs-check
git diff --check
```
