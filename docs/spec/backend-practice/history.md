# Backend Practice History

> **版本**: 1.17
> **状态**: active
> **更新日期**: 2026-07-07

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-07-07 | 1.17 | 压缩 active spec 为当前 PracticePlans / PracticeSessions / VoiceTurn 合同；001 计划完成 flat Resume `resumeId` 绑定与首题 `resumes.structured_profile` prompt context。 | [001-plan-and-session-orchestration](./plans/001-plan-and-session-orchestration/plan.md) |
| 2026-07-07 | 1.16 | 002 文档收敛为当前 text event loop、completion handoff、source-event-only report job 与双轨幂等合同。 | [002-event-loop-and-completion](./plans/002-event-loop-and-completion/plan.md) |
| 2026-07-06 | 1.15 | 004 owner 路径与正文收敛为当前 `sourceReportId` 派生计划合同。 | [004-report-derived-practice-plans](./plans/004-report-derived-practice-plans/plan.md) |
| 2026-06-29 | 1.13 | PracticeGoal 收敛为 `baseline` / `retry_current_round` / `next_round`。 | product-scope/001-core-loop-module-pruning |
| 2026-05-15 | 1.9 | 003 增补 `show_hint` replay 不变量与 hint lifecycle 边界。 | [003-mode-policies-and-provenance](./plans/003-mode-policies-and-provenance/plan.md) |
| 2026-05-13 | 1.7 | 002 落地 append event loop、complete queued report/job handoff、practice turn/session outbox 与双轨幂等。 | [002-event-loop-and-completion](./plans/002-event-loop-and-completion/plan.md) |
| 2026-05-09 | 1.4 | 001 派生 baseline plan/session foundation、shared idempotency、PracticeMode 二值和 first-question AI flow。 | [001-plan-and-session-orchestration](./plans/001-plan-and-session-orchestration/plan.md) |
