# Frontend Workspace and Practice History

> **版本**: 1.2
> **状态**: active
> **更新日期**: 2026-05-08

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-08 | 1.2 | 与 backend-practice v1.2 对齐：把 frontend practiceMode 从三值收敛为 `assisted/strict` 二值，移除 live spec 中的 `legacy debrief replay value` 正向口径；`goal='debrief'` 改为 plan/session 数据来源轴，可与 assisted 或 strict 任意组合；更新 D-3/D-12、PracticeScreen 显隐约束、C-5 与 002 plan 说明，明确只有 strict 隐藏 hint/实时观察，assisted+debrief 仍允许提示；同时把 `ReportGenerationParams` 改为 route-only `PracticeDisplayContext`，`completePracticeSession` body 严格遵守 B2 `CompletePracticeSessionRequest{clientCompletedAt}`。 | 暂无（spec-only 修订） |
| 2026-05-08 | 1.1 | 终稿候选修订：按 spec review 结论收敛 owner 范围为 `workspace / practice / generating`；`report` 与 `company_intel` 改为外部 owner handoff；补充 route 最小上下文矩阵、OpenAPI operation matrix、UI 真理源 anchor 修正、legacy negative gate 分类、acceptance criteria 到 `001`～`004` 预留计划映射；明确 `listResumes` 与 `createPracticeVoiceTurn` 是后续 contract blocker，不在本次 spec 内伪造 | 暂无（spec-only 修订） |
| 2026-05-08 | 1.0 | 初始创建：从 `engineering-roadmap/spec.md` §5.2 预占行 + `frontend-shell/spec.md` §2.1（5 路由壳） + `frontend-home-job-picks-and-parse/spec.md`（parse confirm 跳转契约） + 已 active 的 [`backend-practice`](./../backend-practice/spec.md) v1.0（6 operation + AssistantAction + 双轨 idempotency + 异步 report 触发） + `practice-voice-mvp/spec.md` §5（frontend voice controller owner 指向） + `module-job-workspace.md` v1.8 + `module-practice-review.md` v1.13 + `module-map.md` v2.4 派生新 subspec；定义 5 路由（workspace / practice / generating / report / company_intel）前端 owner 范围、决策 D-1～D-14（含 5 路由全收 + workspace 语义 + practice `practiceMode` 三值（assisted/strict/legacy debrief replay value）+ TopBar 隐藏 + InterviewContext 跨路由 + report session-scoped + voice 入口唯一 + 公司情报合规边界 + 立即面试 createPracticePlan→startPracticeSession 双步契约 + backend-practice 契约消费 + voice 协作面 + appendSessionEvent 单 endpoint kind 路由 + completePracticeSession 202 异步流 + getCompanyIntel fixture-only 红线）、设计约束、模块边界、acceptance criteria C-1～C-13；第一份 plan 推迟，待 `/plan-review` 通过 spec 后由后续 `/design` 拆 D1（建议编号 `001-workspace-and-interview-context`，对接 backend-practice plan `001-plan-and-session-orchestration`）；保留编号建议 002 ~ 005 覆盖 practice text/voice / generating-report / company-intel | 暂无（待后续 /design 拆 plan） |
