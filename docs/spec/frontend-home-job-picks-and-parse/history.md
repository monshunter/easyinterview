# Frontend Home / Parse History

> **版本**: 2.23
> **状态**: active
> **更新日期**: 2026-07-14

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-07-14 | 2.23 | 用户修正确认：报告列表从 Parse 嵌入区迁移到独立 `/reports?targetJobId=...`；Parse 仅在内容区右上角保留页面级入口，并删除列表请求、嵌入 UI 与 `section=reports` 兼容逻辑。 | [001-home-jd-import-and-parse](./plans/001-home-jd-import-and-parse/plan.md) |
| 2026-07-14 | 2.22 | 用户确认 R-A：不新增顶层报告中心，在 Parse 统一详情内按 canonical round 接入最小 `currentReport/latestAttempt` 概览；锁定 join、状态链接、独立错误、不阻断 Start 与 `section=reports` 返回锚点。 | [001-home-jd-import-and-parse](./plans/001-home-jd-import-and-parse/plan.md) |
| 2026-07-13 | 2.21 | Home JD intake 收敛为唯一 textarea 与 `{ rawText, targetLanguage, resumeId }`；删除其他 JD intake 正向合同并保留 Resume 上传。 | 001-home-jd-import-and-parse Phase 18 |
| 2026-07-13 | 2.20 | Parse loading 删除用户界面中的 model/rubric/provenance/latency 内部元数据并增加 desktop/mobile 截图 gate。 | 001-home-jd-import-and-parse Phase 17 |
| 2026-07-10 | 2.19 | Parse success detail ignores route-only `resumeId` for binding; Start requires saved `TargetJob.resumeId`. | 001-home-jd-import-and-parse |
| 2026-07-10 | 2.18 | Home empty state 文案口径收敛为不展示示例业务数据。 | tech-debt pruning |
| 2026-07-09 | 2.10 | Spec v2.13 and plan 001 v2.8 simplify the unified Parse detail into a readonly saved-plan receipt: no field edits, resume picker, success Re-parse, Save plan, or Cancel; Start uses the saved TargetJob / Resume / PracticePlan snapshot directly. | 001-home-jd-import-and-parse Phase 6 |
| 2026-07-09 | 2.9 | Spec v2.12 and plan 001 v2.7 reopen the owner to rename the Parse preview into the unified "面试规划详情 / 面试上下文确认" mother page and route workspace detail re-entry through the same page. | 001-home-jd-import-and-parse Phase 5 |
| 2026-07-07 | 2.8 | 压缩 Home + Parse owner 文档为当前合同：Home 输入卡内 source actions、ready 简历下拉框、最近 3 张模拟面试卡片、Parse loading / preview、真实 `resumeId` handoff、generated-client operation matrix 与 P0.014-P0.016 BDD。 | 001-home-jd-import-and-parse |
| 2026-07-07 | 2.7 | 对齐 current UI truth source 与 formal frontend：`home-jd-source-controls`、`home-resume-select`、`home-recent-more`、Parse `resumeId` handoff 和 `autoStartPractice` handoff 均进入 owner gate。 | 001-home-jd-import-and-parse |
