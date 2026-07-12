# Frontend Workspace and Practice Spec

> **版本**: 1.35
> **状态**: completed
> **更新日期**: 2026-07-12

## 1 背景与目标

`frontend-workspace-and-practice` 承接 `workspace` 面试规划列表、`practice` 连续文本会话和 `generating` 报告生成过渡态。当前 Practice 不再使用题目/turn 状态机，正式页面只保留 Top Bar、全宽 Transcript 和 Composer；电话入口暂时置灰。

`report` 由 `frontend-report-dashboard` / `backend-review` 承接。

## 2 范围

### 2.1 Workspace

- 一级 `面试 / workspace` 始终渲染 ready TargetJob 规划卡片列表。
- 卡片展示状态、更新时间、岗位、公司和地点；主体进入 `parse` 统一详情，右上角归档，底部只有 `立即面试`。
- `workspace` 不读取或继承详情 route context，不拥有 Resume Picker、Plan Switcher 或 route-side auto start。
- 快速启动通过 generated `createPracticePlan` / `startPracticeSession` 创建或复用 plan，然后进入 `practice`。

### 2.2 Practice

- Route 只需要稳定 `sessionId` 与 target/plan/resume/round IDs；不使用 `mode/modality/practiceMode/hintUsed/hintCount`。
- Top Bar：真实公司/岗位、面试官角色、计时、暂停、disabled phone icon、结束并生成报告。
- Conversation：全宽有序 Transcript + Error/Retry + Composer。
- opening message 和后续 assistant reply 统一来自 server messages，不是 QuestionCard。
- 用户输入通过 generated `sendPracticeMessage`，不提交 `turnId`，不标记 answer/hint/question。
- `getPracticeSession` 刷新恢复完整 ordered messages。
- 暂停/恢复只控制当前页面的 composer 与计时显示，不产生 backend 事件；结束通过 `completePracticeSession`。
- Error/Retry 必须按失败来源恢复：session loader 调用 `refresh`，message failure 使用同一 `clientMessageId` 重试 send，completion failure 使用同一 completion idempotency key 重试 finish；不得把完成重试误接到 send。
- message 发送、session loading、completion 进行中或 session 已进入 `completing / completed` 时结束 CTA 必须 disabled，避免 UI 主动制造 send/complete 竞态。
- phone icon 使用原生 disabled 控件；phone/voice route params 不得 materialize PhoneSurface。

### 2.3 Generating

- `completePracticeSession` 返回 `ReportWithJob` 后进入 `generating?sessionId&reportId`。
- 轮询 `getFeedbackReport(reportId)`；ready 后进入 `report`，failed/timeout 展示 retry/back。
- 生成进度文案是会话级上下文/证据/维度/重点/建议，不出现逐题术语。

### 2.4 Out of Scope

- SessionMap、“本轮题目”、题号/总题数、QuestionCard、question badge/topic/prompt。
- `PracticeTurn/currentTurn/turnCount/questionIntent` UI 消费。
- 专用 hint button/banner/event/count 与 strict/assisted switch。
- PhoneSurface、麦克风、字幕、VAD、TTS、barge-in、hangup。
- 独立 Voice route、右侧辅助栏、语音分析、跳过、会话内 persona switch。
- Report Dashboard 具体实现。

## 3 用户决策

| ID | 决策 | 当前合同 |
|----|------|----------|
| D-1 | Practice 交互模型 | 连续文本 conversation，不区分问题/回答/追问 |
| D-2 | 页面布局 | 删除左栏和 QuestionCard，只保留全宽聊天 |
| D-3 | 专用提示 | 删除；用户需要提示时发送普通消息 |
| D-4 | 电话模式 | 前端入口置灰，phone/voice params 归一为文本 |
| D-5 | 报告 handoff | 只传稳定 IDs；不传 modality/practiceMode/hint fields |

## 4 UI 真理源与 parity

- Workspace：`ui-design/src/screen-workspace.jsx`
- Practice：`ui-design/src/screen-practice.jsx::PracticeScreen`
- Generating：`ui-design/src/screens-p0-complete.jsx::ReportGeneratingScreen`
- Shared：`ui-design/src/app.jsx`、`ui-design/src/primitives.jsx`
- Docs：`docs/ui-design/module-job-workspace.md`、`module-practice-review.md`、`report-dashboard.md`

用户可见改动必须先更新 `ui-design/`，再源级迁移到 frontend。验证必须拆分：

1. DOM/control/a11y source parity。
2. computed style/bounding box/responsive/screenshot geometry parity。
3. stale question/hint/phone positive-contract negative search。

## 5 Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `createPracticePlan` | `PracticePlans/createPracticePlan.json` | parse/workspace/report start helpers | backend-practice | `practice_plans` | none | `E2E.P0.021`, `E2E.P0.057` |
| `getPracticePlan` | `PracticePlans/getPracticePlan.json` | start helper | backend-practice | `practice_plans` | none | focused + real-mode gate |
| `startPracticeSession` | `PracticeSessions/startPracticeSession.json` | start helper | backend-practice | session + opening message | `practice.session.chat` | `E2E.P0.023`, `E2E.P0.057` |
| `getPracticeSession` | `PracticeSessions/getPracticeSession.json` | `usePracticeSessionLoader` | backend-practice | session + messages | none | `E2E.P0.044`, `E2E.P0.046` |
| `sendPracticeMessage` | `PracticeSessions/sendPracticeMessage.json` | conversation send hook | backend-practice | `practice_messages` | `practice.session.chat` | `E2E.P0.044`, `E2E.P0.046` |
| `completePracticeSession` | existing fixture | finish hook | backend-practice | session/report/job/outbox | report job | `E2E.P0.047` |
| `getFeedbackReport` | report fixtures | generating poll | backend-review | `feedback_reports` | async owner | `E2E.P0.056`, `E2E.P0.058` |
| `createPracticeVoiceTurn` | disabled negative fixture | no frontend consumer | backend fail-closed | none | none | `E2E.P0.007` |

## 6 Conversation 状态

- Loading：conversation skeleton，不展示假 opening message。
- Running：ordered messages + enabled composer。
- Sending：保留已发送 user message，composer disabled。
- AI failure：显示 retry；同一 `clientMessageId` 重试，不重复 user message。
- Paused：仅当前页面 composer disabled、计时显示暂停，可恢复；刷新后以 server session 状态重新进入 Running。
- Completing/completed：composer disabled，finish CTA guarded。
- Missing/cross-user：session-lost state 返回 workspace。

## 7 Layout

- Desktop：Top Bar 下只有一个 conversation column；内容 max-width 居中，不留 260px sidebar 空白。
- Mobile：单列，Top Bar controls wrap；Transcript 和 Composer 不横向溢出。
- Transcript 独立滚动，Composer 保持在会话区底部。
- disabled phone icon 不得在 narrow layout 变成可点击入口。

## 8 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | Workspace | ready plans | 进入 workspace | 当前卡片列表/启动/归档行为保持 | 001 |
| C-2 | Practice 首屏 | session 有 opening message | 进入 practice | 只见 Top Bar + 全宽 Conversation | 002 |
| C-3 | 连续聊天 | session running | 连续发送消息 | server messages 按序追加，无题目分类 | 002 |
| C-4 | 消息失败恢复 | AI 首次失败 | retry | user message 不重复，最终唯一 reply | 002 |
| C-5 | 暂停/完成 | session running，可能存在加载/发送/完成失败 | pause/resume/finish/retry | 暂停为页面本地状态；retry 调用原失败操作；完成期间 CTA guarded 并进入 generating | 002 |
| C-6 | phone disabled | 任意 route params | 查看/操作 phone icon | disabled，仍为文本 conversation | 002 + voice/001 |
| C-7 | DOM parity | prototype 已更新 | Vitest | 结构/控件/a11y 与 source 一致 | 002 |
| C-8 | Visual parity | desktop/mobile | Playwright | geometry/screenshot 与 source 一致 | 002 |
| C-9 | Stale negative | current tree | lint/search | 无 SessionMap/QuestionCard/hint/PhoneSurface 正向残留 | 002 |
| C-10 | Privacy | conversation 完成 | 检查 URL/storage/log | raw messages 不泄漏 | 002 |

## 9 关联计划

- [001-workspace-and-interview-context](./plans/001-workspace-and-interview-context/plan.md)
- [002-practice-text-event-loop](./plans/002-practice-text-event-loop/plan.md)

## 10 关联文档

- [product-scope](../product-scope/spec.md)
- [backend-practice](../backend-practice/spec.md)
- [practice-voice-mvp](../practice-voice-mvp/spec.md)
- [frontend-report-dashboard](../frontend-report-dashboard/spec.md)
- [openapi-v1-contract](../openapi-v1-contract/spec.md)
- [module-practice-review](../../ui-design/module-practice-review.md)

## 11 修订记录

| 版本 | 日期 | 变更 |
|------|------|------|
| 1.35 | 2026-07-12 | 重新打开 Practice owner：按 loader/message/completion 错误来源路由 retry，并在发送/加载/完成边界禁用结束 CTA。 |
| 1.34 | 2026-07-12 | Practice 改为全宽连续文本会话；删除题目/hint/mode UI，电话入口置灰，generating 改用会话级文案。 |
