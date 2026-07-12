# Backend Review Spec

> **版本**: 1.5
> **状态**: active
> **更新日期**: 2026-07-12

## 1 背景与目标

`backend-review` 将一次已完成 Practice conversation 转换为会话级证据化报告。当前报告不再按题目/turn 分组，也不生成题目回顾、逐题评分或题目重练清单。

报告只回答四类问题：

1. 当前准备度如何。
2. 各能力维度表现如何。
3. 哪些会话证据支持亮点和问题判断。
4. 下一次应复练当前轮还是进入下一轮，以及重点能力是什么。

## 2 当前合同

### 2.1 Operation Matrix

| operation / async path | frontend consumer | persistence | AI dependency | scenario coverage |
|------------------------|-------------------|-------------|---------------|-------------------|
| `getFeedbackReport` | generating/report | `feedback_reports` | none on read | `E2E.P0.056`, `E2E.P0.058`, `E2E.P0.099` |
| `listTargetJobReports` | report records | `feedback_reports` | none on read | focused handler/store contract gate |
| `report_generate` | async runner | `feedback_reports`, jobs/outbox/audit/task-runs | `report.generate` | `E2E.P0.056`, `E2E.P0.058`, `E2E.P0.099` |

### 2.2 输入

- PracticePlan / PracticeSession 稳定上下文。
- 按 `seq_no` 排序的 `practice_messages`。
- target job、resume、round、goal 与 focus competency context。
- 不读取 `practice_turns`、question intent、question status、hint 或 voice committed context。

### 2.3 输出

`FeedbackReport` ready shape：

- `preparednessLevel`
- `dimensionAssessments[] { dimension, status, confidence }`
- `highlights[] { dimension, evidence, confidence }`
- `issues[] { dimension, evidence, confidence }`
- `nextActions[]`
- `retryFocusCompetencyCodes[]`
- `provenance`

删除：

- `questionAssessments`
- `retryFocusTurnIds`
- `QuestionAssessment`
- `QuestionReviewStatus`
- `question_assessments` table
- `report.question_assessment` task/prompt/rubric

### 2.4 Readiness 与复练

- readiness 由会话级维度与证据计算，不按题目数量或逐题状态聚合。
- `retry_current_round` 使用 `retryFocusCompetencyCodes` / issues dimensions 创建新 plan。
- `next_round` 保持 round transition，不携带 turn/question IDs。

### 2.5 失败与恢复

- report generation timeout / provider / invalid output 保持现有 async retry 与 failed report 语义。
- queued / generating / failed read shape 不伪造维度或证据。
- ready output 缺少必需 session-level dimensions 时按 `AI_OUTPUT_INVALID` 处理。

### 2.6 隐私

- report JSON 可以包含面向用户的 evidence 摘要，但不得复制完整 transcript。
- raw message content、prompt、response 和 provider secret 不进入 outbox、audit、log、metric label 或 task-run payload。
- 跨用户读取返回 404。

## 3 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| Practice transcript | `backend-practice` | `practice_messages` |
| API schema | `openapi-v1-contract` | session-level FeedbackReport |
| DB | `db-migrations-baseline` | 删除 question assessments，更新 feedback report focus field |
| Prompt | `prompt-rubric-registry` | 单一 report.generate schema/rubric |
| UI | `frontend-report-dashboard` | readiness / dimensions / evidence / next 四区 |

## 4 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 生成报告 | 完成会话含多轮 messages | runner 执行 | ready report 含会话级 dimensions/evidence，无 question shape | 001 |
| C-2 | 复练当前轮 | report needs practice | 创建 retry plan | 使用 competency codes，不使用 turn IDs | 001 |
| C-3 | 读取报告 | queued/generating/ready/failed | 前端读取 | 各状态 shape 稳定，跨用户 404 | 001 |
| C-4 | AI 失败 | provider/invalid output | runner 重试或终止 | job/report 状态正确，无部分 ready 数据 | 001 |
| C-5 | 隐私 | report 完成 | 检查非正文存储面 | 无 raw transcript/prompt/response 泄漏 | 001 |

## 5 关联计划

- [001-report-generation-baseline](./plans/001-report-generation-baseline/plan.md)

## 6 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-12 | 1.5 | 删除逐题评估与 turn focus，报告改为 conversation-level dimensions/evidence/actions。 |
