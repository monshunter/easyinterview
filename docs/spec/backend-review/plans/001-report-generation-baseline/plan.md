# 001 - Report Generation Baseline

> **版本**: 1.6
> **状态**: completed
> **更新日期**: 2026-07-10

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

落地当前 backend-review 报告生成与读取合同：

- 消费 backend-practice 已创建的 `async_jobs(job_type='report_generate')` 与 `feedback_reports(status='queued')`。
- 由 backend-async-runner kernel 调度 `review.GenerateHandler`，完成 report generation、question assessment、readiness、retry focus、next action、outbox、audit 与 `ai_task_runs` 记录。
- 提供 `getFeedbackReport` 与 `listTargetJobReports` 两个 generated OpenAPI handler。
- 在 AI / prompt / parse failure 时写入 failed report shape 与 retry/permanent job 状态，不向 read handler 泄露 5xx。
- 保持 user-scoped read/write、wire provenance 6 字段边界、raw QA/prompt/response/secret privacy boundary。

本 plan 是 completed owner 文档，当前用途是作为 implementation / review 的 executable evidence index，不再承载阶段演进说明。v1.4 将 queued / generating 报告读取语义收敛为当前状态元数据，不再描述为空报告壳。

## 2 当前合同

| Surface | 当前合同 | Gate |
|---------|----------|------|
| API | `getFeedbackReport`, `listTargetJobReports`, `REPORT_NOT_FOUND`, `FeedbackReport.errorCode` | `make codegen-check`, report handler tests, BDD P0.053/P0.055 |
| DB | `feedback_reports` with provenance/retry columns, `question_assessments`, `async_jobs`, `ai_task_runs(report_generate/report_assessment)` | migration contract tests, store tests |
| Runtime | `runner.Runtime` owns lease/retry/reaper; backend-review owns `review.GenerateHandler` business logic | `TestBuildReportRuntimeWiresRoutesRunnerReaperAndAI` |
| AI | F3 `report.generate` and `report.question_assessment`; A3 observed `AIClient.Complete` | F3 preflight, AI profile/config lint |
| Events | `report.generated`, `report.generation.failed` | event codegen/lint and outbox emitter tests |
| Privacy | no raw QA, prompt body, AI response body, or provider secret in report JSON, outbox, audit, logs, or metric labels | redaction tests, BDD P0.055, out-of-scope report surface lint |

## 3 Operation Matrix

| Operation / async path | Fixture | Frontend consumer | Backend owner | Persistence | AI dependency | Scenario coverage |
|------------------------|---------|-------------------|---------------|-------------|---------------|-------------------|
| `getFeedbackReport` | `openapi/fixtures/Reports/getFeedbackReport.json` | `frontend-report-dashboard` generating and report views | `backend/internal/api/reports`, `backend/internal/review`, `backend/internal/store/review` | `feedback_reports`, `question_assessments` | none on read path | `E2E.P0.053`, `E2E.P0.055` |
| `listTargetJobReports` | `openapi/fixtures/Reports/listTargetJobReports.json` | report records within dashboard owner | same | target ownership check, `feedback_reports` cursor read | none on read path | `E2E.P0.053`, `E2E.P0.055` |
| `report_generate` job | N/A | observed through `getFeedbackReport` | `runner.Runtime`, `review.GenerateHandler`, review store | `feedback_reports`, `question_assessments`, `async_jobs`, `outbox_events`, `audit_events`, `ai_task_runs` | F3 report prompts/rubrics, A3 Complete | `E2E.P0.052`, `E2E.P0.054`, `E2E.P0.055` |

## 4 Coverage Matrix

| Area | Required evidence |
|------|-------------------|
| Report happy path | runner/service/store tests plus `TestE2EP0052ReportGenerationHappyPath` |
| Read handlers | API/service/store tests plus `TestE2EP0053ReportReadAndListing` |
| Failure and retry | service failure matrix, runner retry policy, persist failure tests, `TestE2EP0054ReportAIFailureAndRetry` |
| Privacy and cross-user isolation | redaction tests, ownership tests, `TestE2EP0055ReportPrivacyAndOutOfScope` |
| Contract and generated artifacts | OpenAPI codegen, fixture validation, conventions drift, migration lint/tests, event codegen/lint |
| Runtime wiring | `TestBuildReportRuntimeWiresRoutesRunnerReaperAndAI`, `TestBuildReportRuntimeRejectsMissingAIClient` |
| Current-scope negative surface | `python3 scripts/lint/backend_review_out_of_scope.py --repo-root . --phase all` and its pytest suite |

## 5 Completed Implementation Slices

- Contract and fixture alignment: `REPORT_NOT_FOUND`, `FeedbackReport.errorCode`, report fixtures, migration enum/column contract, and F3 preflight are covered by generated-code, fixture, migration, conventions, prompt, and rubric gates.
- Runner and job handling: current runtime uses backend-async-runner kernel, with backend-review business logic in `review.GenerateHandler`; lease/retry/reaper behavior is covered by runtime and runner tests.
- AI content generation: report and question assessment prompts are resolved through F3 and executed through A3; output schema, score-level mapping, token accounting, and failure rows are covered by review/registry/aiclient tests.
- Persistence: report success/failure writes are transactional across report rows, assessments, async job state, outbox, and audit.
- Read APIs: handlers preserve user isolation, queued/generating/ready/failed shapes, pagination, cursor validation, and target ownership.
- Privacy and observability: redaction, metric-label allowlist, bounded error codes, outbox payload schema, and current-scope negative lint are covered.

## 6 Verification Commands

These commands define the current owner evidence set:

```bash
cd backend && go test ./internal/api/reports ./internal/review ./internal/store/review ./internal/ai/aiclient ./internal/ai/registry ./cmd/api -count=1
cd backend && go test ./cmd/api -run 'TestE2EP0052|TestE2EP0053|TestE2EP0054|TestE2EP0055|TestBuildReportRuntime' -count=1
make codegen-check
make validate-fixtures
migrations/lint.sh
make lint-events
make codegen-events-check
python3 scripts/lint/conventions_drift.py --repo-root .
python3 scripts/lint/backend_review_out_of_scope.py --repo-root . --phase all
python3 -m pytest scripts/lint/backend_review_out_of_scope_test.py -q
python3 scripts/lint/prompt_lint.py
python3 scripts/lint/rubric_lint.py
make docs-check
git diff --check
```

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-10 | 1.6 | Consolidate retryable GenerateHandler test setup while preserving exact regression names. |
| 2026-07-10 | 1.5 | Remove the production-only readiness-tier test helper and keep the enum property assertion in test data. |
| 2026-07-07 | 1.2 | Compress owner plan to the current report generation/read contract and executable evidence index. |
| 2026-05-16 | 1.1 | Complete report generation baseline delivery. |

## 8 Test-only readiness validator removal

删除 production `validReadinessTier`，随机 readiness property test 直接使用四个 shared `ReadinessTier` 常量组成的测试集合验证输出。真实 readiness 计算与阈值映射保持不变。

## 9 GenerateHandler retryable-failure test harness consolidation

Keep both top-level retryable failure test names, including the BUG-0088/backend-async-runner exact gate. Move their shared store/service/job setup and output assertions into one test-only helper parameterized only by the service's incoming `AsyncJobFinalized` value; both paths must still return retryable timeout with handler output `AsyncJobFinalized=false`.
