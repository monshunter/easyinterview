# 001 - Report Generation Baseline Test Plan

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-07-07

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)
**关联 Test Checklist**: [test-checklist](./test-checklist.md)

## 1 覆盖矩阵

| 范围 | 当前可执行证据 |
|------|----------------|
| Report happy path | `TestRunnerLeasesAndAdvancesToGenerating`, `TestGenerateReportServiceOrchestratesAndPersists`, `TestPersistReportWritesQuestionAssessments`, `TestE2EP0052ReportGenerationHappyPath` |
| Question assessments | `TestAssessQuestionsForAllTurns`, `TestPersistReportWritesQuestionAssessments`, `TestE2EP0052ReportGenerationHappyPath` |
| Readiness / DimensionStatus | `TestComputeReadinessTier`, `TestComputeReadinessTierScoreLevelsAndDimensionStatusMapping`, `TestComputeReadinessTierPropertyRandomDimensions` |
| Retry focus / next action | `TestSelectRetryFocusTurns`, `TestDecideNextAction`, `TestE2EP0052ReportGenerationHappyPath` |
| Placeholder / failed read | report handler/service/store tests, `TestE2EP0053ReportReadAndListing` |
| Listing pagination / ownership | cursor, ownership, handler/service/store tests, `TestE2EP0053ReportReadAndListing`, `TestE2EP0055ReportPrivacyAndNonCurrent` |
| AI failure matrix | `TestGenerateReportFailedMatrix`, `TestE2EP0054ReportAIFailureAndRetry` |
| Retry / permanent failure | `TestRunnerRetryPolicyAndPermanentFail`, `TestPersistReportFailureRetryAndPermanent`, `TestE2EP0054ReportAIFailureAndRetry` |
| Runtime wiring | `TestBuildReportRuntimeWiresRoutesRunnerReaperAndAI`, `TestBuildReportRuntimeRejectsMissingAIClient` |
| Contract / generated artifacts | `make codegen-check`, `make validate-fixtures`, `migrations/lint.sh`, `make lint-events`, `make codegen-events-check`, `python3 scripts/lint/conventions_drift.py --repo-root .` |
| Prompt / rubric | `TestF3ReportGenerateAndAssessmentPreflight`, `python3 scripts/lint/prompt_lint.py`, `python3 scripts/lint/rubric_lint.py` |
| Privacy / observability | redaction tests, outbox PII tests, metric-label tests, `TestE2EP0055ReportPrivacyAndNonCurrent` |
| Current-scope boundary | `python3 scripts/lint/backend_review_non_current.py --repo-root . --phase all`, `python3 -m pytest scripts/lint/backend_review_non_current_test.py -q` |

## 2 Owner Evidence Commands

```bash
cd backend && go test ./internal/api/reports ./internal/review ./internal/store/review ./internal/ai/aiclient ./internal/ai/registry ./cmd/api -count=1
cd backend && go test ./cmd/api -run 'TestE2EP0052|TestE2EP0053|TestE2EP0054|TestE2EP0055|TestBuildReportRuntime' -count=1
make codegen-check
make validate-fixtures
migrations/lint.sh
make lint-events
make codegen-events-check
python3 scripts/lint/conventions_drift.py --repo-root .
python3 scripts/lint/backend_review_non_current.py --repo-root . --phase all
python3 -m pytest scripts/lint/backend_review_non_current_test.py -q
python3 scripts/lint/prompt_lint.py
python3 scripts/lint/rubric_lint.py
make docs-check
git diff --check
```

## 3 Completion Criteria

- Listed Go tests must execute real tests, not `[no tests to run]`.
- Contract, fixture, migration, event, prompt/rubric, docs, and diff gates must pass from the documented working directory.
- Current-scope negative lint must pass with no backend-review runtime residuals.
