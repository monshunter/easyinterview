# Plans 索引

> 本索引只记录当前 subspec 下的 `plans/*/plan.md`。

---

## 1 进行中（Active）

| 计划 | 文件 | 版本 | 状态 | 更新日期 |
|------|------|------|------|----------|
| [002-practice-text-event-loop](./002-practice-text-event-loop/plan.md) | [plan](./002-practice-text-event-loop/plan.md) / [checklist](./002-practice-text-event-loop/checklist.md) / [context](./002-practice-text-event-loop/context.yaml) / [test-plan](./002-practice-text-event-loop/test-plan.md) / [test-checklist](./002-practice-text-event-loop/test-checklist.md) / [bdd-plan](./002-practice-text-event-loop/bdd-plan.md) / [bdd-checklist](./002-practice-text-event-loop/bdd-checklist.md) | 1.6 | active | 2026-06-13 |
| [001-workspace-and-interview-context](./001-workspace-and-interview-context/plan.md) | [plan](./001-workspace-and-interview-context/plan.md) / [checklist](./001-workspace-and-interview-context/checklist.md) / [context](./001-workspace-and-interview-context/context.yaml) / [bdd-plan](./001-workspace-and-interview-context/bdd-plan.md) / [bdd-checklist](./001-workspace-and-interview-context/bdd-checklist.md) | 1.6 | active | 2026-06-13 |

## 2 已完成（Completed）

| 计划 | 文件 | 版本 | 状态 | 更新日期 |
|------|------|------|------|----------|

<!--
保留编号建议（启动条件由后续 plan-review 阶段确认；与 backend-practice §7 plan 序列对接见 spec.md §7）：

- 003-practice-voice-surface — PracticeScreen 语音 surface + audio/playback/barge-in UI 事件上报 + BDD；full voice turn flow 等 createPracticeVoiceTurn contract landed 后启用
- 004-generating-report-handoff — GeneratingScreen 异步轮询 getFeedbackReport(reportId) + succeeded/failed handoff 到 external report owner + BDD

本 subspec 不预留 report 或 company_intel plan；ReportScreen / CompanyIntelScreen / getCompanyIntel 由外部 owner spec 原地承接。
-->
