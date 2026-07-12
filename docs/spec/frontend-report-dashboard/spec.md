# Frontend Report Dashboard Spec

> **版本**: 1.11
> **状态**: completed
> **更新日期**: 2026-07-12

## 1 背景与目标

`frontend-report-dashboard` 承接一次连续模拟面试结束后的生成态与会话级报告。当前报告不再展示题目数量、题目回顾、逐题评分或加入本轮复练的题目标记；它只展示准备度、能力维度、会话证据和下一步行动。

## 2 范围

### 2.1 In Scope

- `generating`：轮询 report status，并展示五阶段会话级分析进度。
- `report`：Header、精简 Context Strip、四张 Summary Cards、四个 Detail Tabs、失败/缺 session 状态。
- Header 保留唯一一对 CTA：`复练当前轮` / `进入下一轮`。
- `retry_current_round` 使用 `retryFocusCompetencyCodes`，不使用 turn/question IDs。
- 完整 zh/en i18n、keyboard/a11y、desktop/mobile parity。

### 2.2 Out of Scope

- Questions tab、题目回顾、逐题分析、题号/总题数、per-question replay toggle。
- `questionAssessments` / `retryFocusTurnIds` 消费。
- hint / practiceMode / modality 展示。
- 电话/语音报告。
- 精确通过率、时间线、独立错题本或报告一级导航。

## 3 用户决策

| ID | 决策 | 当前结论 |
|----|------|----------|
| D-1 | 报告粒度 | 整场 conversation，不按题目/turn |
| D-2 | Summary Cards | 准备度 / 能力维度 / 会话证据 / 下一步 |
| D-3 | Detail Tabs | readiness / dimensions / evidence / next；默认 readiness |
| D-4 | 复练输入 | competency codes + evidence gaps，不传 turn IDs |
| D-5 | Context Strip | session / target / round / resume；不展示 phone/hint/mode |

## 4 UI 真理源

- `ui-design/src/screen-report.jsx::ReportScreen`
- `ui-design/src/screens-p0-complete.jsx::ReportGeneratingScreen`
- `docs/ui-design/report-dashboard.md`
- `docs/ui-design/module-practice-review.md`

正式前端必须源级复刻更新后的 prototype：DOM 层级、四卡/四 tab 结构、标签、按钮、spacing、typography、responsive layout、keyboard/a11y 与截图 geometry 都需要 parity gate。

## 5 Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `getFeedbackReport` | `Reports/getFeedbackReport.json` | generating poll + ReportDashboard | backend-review | `feedback_reports` | read path none | `E2E.P0.056`, `E2E.P0.058` |
| `createPracticePlan` | `PracticePlans/createPracticePlan.json` | replay/next-round CTA | backend-practice | `practice_plans` | none | `E2E.P0.057` |
| `startPracticeSession` | `PracticeSessions/startPracticeSession.json` | replay/next-round CTA | backend-practice | session + opening message | `practice.session.chat` | `E2E.P0.057` |

## 6 页面结构

### 6.1 Generating

五阶段文案：

1. 整理会话上下文。
2. 提取能力证据。
3. 评估能力维度。
4. 归纳改进重点。
5. 生成行动建议。

不得出现“逐题”“题目回顾”“每题评分”等文案。

### 6.2 Report Dashboard

```text
ReportDashboard
├─ Back
├─ Header
│  ├─ title / subtitle
│  ├─ 复练当前轮
│  └─ 进入下一轮
├─ ContextStrip
│  ├─ session
│  ├─ target / round
│  └─ resume
├─ SummaryCards
│  ├─ readiness
│  ├─ dimensions
│  ├─ evidence
│  └─ next
└─ DetailSurface
   ├─ ReadinessTab
   ├─ DimensionsTab
   ├─ EvidenceTab
   └─ NextTab
```

### 6.3 Replay CTA

- `复练当前轮` 创建 `goal=retry_current_round` plan，传递 `sourceReportId`、`focusCompetencyCodes=retryFocusCompetencyCodes` 与稳定 target/resume/round context。
- `进入下一轮` 创建 `goal=next_round` plan，并携带目标 round context。
- 两条路径均创建 fresh session，直接进入文本 `practice`，不复用旧 session。

## 7 状态与错误

- 缺 `sessionId/reportId`：渲染 `ReportMissingSessionState`，不展示假报告。
- `queued/generating`：留在 generating。
- `ready`：渲染会话级 dashboard。
- `failed/not found/timeout`：显示 typed error、retry/back CTA。
- empty dimensions/evidence：显示明确空态，不恢复 Questions tab。

## 8 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | generating | report queued | 轮询 | 展示五阶段会话级进度，无逐题文案 | 001 |
| C-2 | ready dashboard | report ready | 进入 report | 四 cards + 四 tabs，默认 readiness | 001 |
| C-3 | evidence | report 有 highlights/issues | 打开 evidence | 展示会话级证据，不按题目分组 | 001 |
| C-4 | replay | report needs practice | 点击复练当前轮 | competency codes 创建 fresh session | 001 |
| C-5 | next round | 下一轮可用 | 点击进入下一轮 | 创建 next_round fresh session | 001 |
| C-6 | error/empty | report 缺失/失败/空证据 | 页面加载 | 状态明确且无假数据 | 001 |
| C-7 | parity | prototype 已更新 | 运行 DOM/geometry/screenshot gate | desktop/mobile 与 source 一致 | 001 |
| C-8 | stale negative | 全仓扫描 | 检查当前资产 | 无 QuestionsTab/题目回顾/questionAssessments/retryFocusTurnIds 正向残留 | 001 |

## 9 关联计划

- [001-report-screen-and-generating-handoff](./plans/001-report-screen-and-generating-handoff/plan.md)

## 10 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-12 | 1.11 | 报告改为会话级四卡/四 tab，删除逐题模型、hint/phone 展示与 turn-based replay。 |
