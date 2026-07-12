# 001 — Conversation-level Report Generation

> **版本**: 2.0
> **状态**: active
> **更新日期**: 2026-07-12

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

将 report generation/read contract 从 question assessments 改为 conversation-level dimensions/evidence/actions；删除 question assessment prompt/table/task/UI fields，复练使用 competency codes。

## 2 Operation Matrix

| operationId / job | fixture | consumer | backend | persistence | AI | scenario |
|-------------------|---------|----------|---------|-------------|----|----------|
| `report_generate` job | N/A | generating poll | runner/review service/store | feedback_reports/jobs/outbox/audit/task | report.generate | P0.056/P0.058/P0.099 |
| `getFeedbackReport` | report fixtures | dashboard | reports handler/review read | feedback_reports | none | P0.056/P0.058/P0.099 |
| `listTargetJobReports` | list fixture | report records | same | feedback_reports | none | focused handler/store gate |

## 3 质量门禁分类

- **Plan 类型**: feature-behavior + contract + async backend + migration。
- **TDD 策略**: Red tests require dimensionAssessments/retryFocusCompetencyCodes and absence of question rows before service/store/schema changes.
- **BDD 策略**: 已存在的 P0.056/P0.057/P0.058/P0.099 cover generate/read/retry/failure/real integration；隐私与隔离由 focused backend contract gate 承接。
- **替代验证 gate**: prompt/schema/eval, migration, codegen/fixtures, runner/store integration, privacy lint.

## 4 Coverage Matrix

| Behavior | Category | Phase | Verification | Negative |
|----------|----------|-------|--------------|----------|
| generate report | primary | 1-2 | service/store + P0.056/P0.099 | per-question task/rows |
| read states | contract | 3 | handler/store + P0.056/P0.058 | question fields |
| AI failure/retry | recovery | 2 | matrix + P0.058 | partial ready report |
| competency replay | cross-layer | 3 | mapper/replay tests + P0.057 | turn IDs |
| privacy | security | 4 | focused isolation/redaction gates | raw transcript in non-content surfaces |

## 5 实施步骤

### Phase 1: Report contract/prompt
- Rewrite report.generate schema/rubric/evals to output readiness, dimensionAssessments, highlights/issues, nextActions and retryFocusCompetencyCodes.
- Delete report.question_assessment assets and generated/API/DB shapes.

### Phase 2: Generate/store
- Load ordered practice_messages and context.
- Generate one report AI result, calculate/validate readiness, persist one report transactionally.
- Preserve retry/failure/job/outbox/audit/task-run semantics.

### Phase 3: Read/replay
- Map queued/generating/ready/failed reports without question fields.
- Use competency focus for retry plan creation.

### Phase 4: Privacy and closeout
- Redaction/isolation/current-scope negative/full gates and P0.056/P0.057/P0.058/P0.099.

## 6 验收标准

- Ready report contains session-level dimensions/evidence/actions and no question data.
- Retry path carries competency codes, not turn IDs.
- Failure/read/privacy behavior stays deterministic.

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-12 | 2.0 | Reopen for conversation-level report and competency-focused retry. |
