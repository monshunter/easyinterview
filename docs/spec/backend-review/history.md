# Backend Review History

> **版本**: 1.6
> **状态**: active
> **更新日期**: 2026-07-12

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-07-12 | 1.6 | 重新打开 001：candidate dimension score 固定为 1.0-5.0；prompt/schema/runtime 在 readiness 计算前共同拒绝越界或不完整输出。 | [001-report-generation-baseline](./plans/001-report-generation-baseline/plan.md) |
| 2026-07-12 | 1.5 | 报告改为整场 conversation 级 readiness、dimensions、evidence 与 next，删除逐题 assessment 和 turn-ID retry 合同。 | [001-report-generation-baseline](./plans/001-report-generation-baseline/plan.md) |
| 2026-07-10 | 1.2 | 收敛 queued / generating 报告读取语义：`getFeedbackReport` 返回当前状态元数据，不再描述为空报告占位。 | tech-debt pruning |
| 2026-06-29 | 1.1 | product-scope D-22 后同步 downstream 边界：backend-review 只声明当前 practice/report downstream owner；runner job_type 描述保持 `report_generate`。 | product-scope/001-core-loop-module-pruning |
| 2026-05-15 | 1.0 | 初始创建：从 [engineering-roadmap §5.2](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) `Report Dashboard` workstream 与 [backend-practice spec D-22/D-28/D-32](../backend-practice/spec.md) 的 downstream handoff 收敛出 backend-review owner subspec。本版本把已编码契约（[B2 OpenAPI](../openapi-v1-contract/spec.md) `getFeedbackReport` / `listTargetJobReports` operation + `FeedbackReport` / `PaginatedFeedbackReport` / `QuestionAssessment` / `GenerationProvenance` schema、[B3](../event-and-outbox-contract/spec.md) `report.generation.requested` / `report.generated` / `report.generation.failed` 事件与 `report_generate` job、[B4](../db-migrations-baseline/spec.md) `feedback_reports` / `question_assessments` / `ai_task_runs.task_type='report_generate'` / `async_jobs.job_type='report_generate'` 表、[B1](../shared-conventions-codified/spec.md) `ReportStatus` / `ReadinessTier` / `DimensionStatus` / `Confidence` / `QuestionReviewStatus` 枚举与 `REPORT_NOT_READY` 错误码、[F3](../prompt-rubric-registry/spec.md) `report.generate` / `report.question_assessment` v0.1.0 baseline feature_key 与 `report.generate.default` / `report.assessment.default` model profile）缝合为 backend-review owner spec；锁定 D-1..D-15 决策（含 ReadinessTier 加权阈值算法、retry_focus_turns 选择策略、next_action enum 决策、status 状态机、AI 失败 graceful 语义、provenance wire 边界、ai_task_runs typed columns、隐私红线、outbox payload schema、inline review runner 边界、listTargetJobReports 分页、`REPORT_NOT_FOUND` B1 前置）；列出 12 条验收标准 C-1..C-12 与首个 plan `001-report-generation-baseline` + 未来保留计划编号；本版本不派 sibling plan，待 plan 001 落地后再按 owner 边界派生。 | [001-report-generation-baseline](./plans/001-report-generation-baseline/plan.md) |
