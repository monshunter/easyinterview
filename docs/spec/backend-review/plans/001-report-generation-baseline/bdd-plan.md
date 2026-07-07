# 001 - Report Generation Baseline BDD Plan

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-07-07

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 场景矩阵

| 场景 ID | 名称 | 类别 | 验证入口 |
|---------|------|------|----------|
| `E2E.P0.052` | Report generation happy path | primary | `TestE2EP0052ReportGenerationHappyPath` |
| `E2E.P0.053` | Report read and listing | alternate + API UX | `TestE2EP0053ReportReadAndListing` |
| `E2E.P0.054` | Report AI failure and retry | failure + observability | `TestE2EP0054ReportAIFailureAndRetry` |
| `E2E.P0.055` | Report privacy and current-scope boundary | privacy + regression | `TestE2EP0055ReportPrivacyAndNonCurrent` |

执行入口：

```bash
cd backend && go test ./cmd/api -run 'TestE2EP0052|TestE2EP0053|TestE2EP0054|TestE2EP0055' -count=1
```

## 2 场景定义

| 场景 ID | Given | When | Then |
|---------|-------|------|------|
| `E2E.P0.052` | 当前用户拥有 completed practice session、queued feedback report、queued `report_generate` job、3 个 assessed turns、F3 report prompts/rubrics 和 A3 fake client | runner kernel handles the report job | job is leased once; report becomes `generating` then `ready`; F3/A3 calls happen in order; report content, assessments, retry focus, provenance columns, outbox, audit, async job success, and `ai_task_runs` rows are persisted; raw QA/prompt/response/secret material is absent |
| `E2E.P0.053` | 当前用户拥有 ready/queued/generating/failed reports and a paginated target report set; another user has no ownership | user calls `getFeedbackReport` and `listTargetJobReports` as both users | ready/placeholder/failed shapes are correct; pagination and invalid cursor behavior are correct; cross-user report and target reads return `REPORT_NOT_FOUND`; wire provenance has exactly 6 fields |
| `E2E.P0.054` | Several queued reports are configured for prompt, provider, timeout, invalid output, parse-empty, and permanent-failure cases | runner handles each report job | each report becomes failed with bounded B1 error code; retry/permanent job state is correct; failure outbox and audit rows are emitted; `ai_task_runs` rows use current enums; HTTP read path returns failed report shape rather than 5xx |
| `E2E.P0.055` | Two users, one ready report, one happy-path report generation fixture, and runtime/audit/outbox/AI task collectors | cross-user reads run, then current user generates one report and the backend-review current-scope lint runs | cross-user reads return `REPORT_NOT_FOUND`; persisted/runtime outputs exclude raw QA/prompt/response/secret material; metric labels stay bounded; wire provenance stays at 6 fields; backend-review current-scope lint passes |

## 3 数据隔离与恢复

- 每个 scenario uses unique user/report/session/target/job identifiers.
- Cleanup removes scenario-owned assessments, reports, async jobs, practice rows, AI task rows, audit rows, outbox rows, target rows, and users.
- A failed scenario first cleans its own data, then checks shared F3/A3/runtime state. Full environment rebuild is only a last resort.

## 4 AC 映射

| Spec AC | 覆盖场景 |
|---------|----------|
| C-1 report happy path | `E2E.P0.052` |
| C-2 assessment mapping | `E2E.P0.052` |
| C-3 readiness / retry / next action | `E2E.P0.052` |
| C-4 placeholder read | `E2E.P0.053` |
| C-5 failed read | `E2E.P0.053`, `E2E.P0.054` |
| C-6 report listing | `E2E.P0.053` |
| C-7 failure and retry | `E2E.P0.054` |
| C-8 provenance boundary | `E2E.P0.053`, `E2E.P0.055` |
| C-9 cross-user isolation | `E2E.P0.053`, `E2E.P0.055` |
| C-10 privacy boundary | `E2E.P0.055` |
| C-11 async job concurrency | unit/runtime tests referenced by [test-plan](./test-plan.md) |
