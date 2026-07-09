# Frontend Home / Parse History

> **版本**: 2.10
> **状态**: active
> **更新日期**: 2026-07-09

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-07-09 | 2.10 | Spec v2.13 and plan 001 v2.8 simplify the unified Parse detail into a readonly saved-plan receipt: no field edits, resume picker, success Re-parse, Save plan, or Cancel; Start uses the saved TargetJob / Resume / PracticePlan snapshot directly. | 001-home-jd-import-and-parse Phase 6 |
| 2026-07-09 | 2.9 | Spec v2.12 and plan 001 v2.7 reopen the owner to rename the Parse preview into the unified "面试规划详情 / 面试上下文确认" mother page and route workspace detail re-entry through the same page. | 001-home-jd-import-and-parse Phase 5 |
| 2026-07-07 | 2.8 | 压缩 Home + Parse owner 文档为当前合同：Home 输入卡内 source actions、ready 简历下拉框、最近 3 张模拟面试卡片、Parse loading / preview、真实 `resumeId` handoff、generated-client operation matrix 与 P0.014-P0.016 BDD。 | 001-home-jd-import-and-parse |
| 2026-07-07 | 2.7 | 对齐 current UI truth source 与 formal frontend：`home-jd-source-controls`、`home-resume-select`、`home-recent-more`、Parse route `resumeId` inheritance 和 `autoStartPractice` handoff 均进入 owner gate。 | 001-home-jd-import-and-parse |
