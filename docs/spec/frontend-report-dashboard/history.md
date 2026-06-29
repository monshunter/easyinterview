# Frontend Report Dashboard History

> **版本**: 1.4
> **状态**: active
> **更新日期**: 2026-06-29

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-06-29 | 1.4 | product-scope D-22 后同步前端 downstream 边界：真实面试复盘 UI/API 已退役，ReportScreen 不再声明 DebriefScreen / getDebrief downstream owner。 | product-scope/001-core-loop-module-pruning |
| 2026-05-23 | 1.2 | L2 real-backend drift 修订：backend-review/001 真实 handler 已落地，spec/operation matrix 不再保留 Phase 5/future 口径；P0.056-P0.059 trigger 前置 `frontendOwners.realApiMode.test.ts`，verify 检查 real-mode marker、默认 backend base URL 与测试文件 marker，同时保留 `listTargetJobReports` dashboard-only UI 负向断言。 | 001-report-screen-and-generating-handoff |
| 2026-05-16 | 1.1 | L2 code review 修订：复练 / 下一轮 CTA 不再直接进入 `practice` 或复用来源 session，改为 `workspace` + `autoStartPractice=1` fresh-session handoff；P0.059 场景 gate 从确认 Playwright spec 文件存在升级为执行 frontend build + generating/report Playwright pixel parity 并检查通过日志。 | [001-report-screen-and-generating-handoff](./plans/001-report-screen-and-generating-handoff/plan.md) |
| 2026-05-15 | 1.0 | 初始创建：从 [engineering-roadmap §5.2](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) `Report Dashboard` workstream 与 [frontend-workspace-and-practice spec D-1/D-6](../frontend-workspace-and-practice/spec.md) 的 report owner handoff 收敛出 frontend-report-dashboard owner subspec。本版本以 `ui-design/src/screen-report.jsx`（含 ReportDashboard 三态、5 个 detail tab、维度卡片、复练 CTA 路径 A/B）+ `ui-design/src/screens-p0-complete.jsx::ReportGeneratingScreen`（5 阶段进度动画 + 实时观察流 + max attempts 超时）+ `docs/ui-design/report-dashboard.md`（dashboard-only 形态 + 准备度 4 档 + 维度状态 + 复练规则）为 UI 真理源，把 [B2 OpenAPI](../openapi-v1-contract/spec.md) `getFeedbackReport` / `listTargetJobReports` operation + `FeedbackReport` / `PaginatedFeedbackReport` schema 缝合为 frontend-report-dashboard owner spec；锁定 D-1..D-14 决策（含 route owner 范围、UI 真理源 + 源级复刻、GeneratingScreen 轮询节奏（指数退避 + max attempts 30）、状态分支三态、复练 CTA payload 双路径、报告失败态语义、i18n 命名空间、InterviewContext reducer 边界 read-only、5 detail tab DOM 锚点、准备度 tier 4 档文案、维度状态三态映射、retired 术语清单、隐私红线、backend 契约消费形态）；列出 15 条验收标准 C-1..C-15 与首个 plan `001-report-screen-and-generating-handoff` + 未来保留计划编号（002 quality feedback + listing、003 export & share）；本版本不派 sibling plan，待 plan 001 落地后再按 owner 边界派生。 | [001-report-screen-and-generating-handoff](./plans/001-report-screen-and-generating-handoff/plan.md) |
