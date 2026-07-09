# 003 BDD Plan

> **版本**: 1.5
> **状态**: completed
> **更新日期**: 2026-07-09

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 场景矩阵

| 场景 ID | 名称 | 类别 | 验证入口 |
|---------|------|------|----------|
| `E2E.P0.048` | assisted hint 主路径 × goal 矩阵 | primary + alternate | `backend/cmd/api/practice_http_scenario_test.go::TestE2EP0048PracticeHintAssistedAcrossGoals` |
| `E2E.P0.049` | legacy strict optional hint × goal 矩阵 | alternate + compatibility | `backend/cmd/api/practice_http_scenario_test.go::TestE2EP0049PracticeHintOptionalAcrossLegacyStrictGoals` |
| `E2E.P0.050` | AssistantAction provenance wire 边界 + ai_task_runs runtime 字段 | cross-layer contract | `backend/cmd/api/practice_http_scenario_test.go::TestE2EP0050PracticeAssistantActionProvenanceAndTaskRuns` |
| `E2E.P0.051` | hint AI graceful degrade + privacy + runtime boundary negative | failure/recovery + privacy + regression | `backend/cmd/api/practice_http_scenario_test.go::TestE2EP0051PracticeHintDegradeAndPrivacy` |

## 2 场景明细

### E2E.P0.048 assisted hint 主路径

| Given | When | Then |
|-------|------|------|
| 用户 A 拥有 `baseline` / `retry_current_round` / `next_round` 三类 ready plan，mode 均为 `assisted`，session 均为 running 且 current turn 为 asked；F3/A3 fake 返回合法 hint | 用户对每个 session 发起 `hint_requested`，随后同 turn 发起第二次 hint，再 replay 第一次 `clientEventId` | 返回 200 + `show_hint`；hint 非空；provenance 六字段；`practice_turns.hint_text` 写入；turn status / turn count / outbox / audit 不变；replay 返回原始 hint snapshot；`ai_task_runs(hint_generate)` 写 success |

### E2E.P0.049 legacy strict optional hint

| Given | When | Then |
|-------|------|------|
| 用户 A 拥有三类 current goal 的 legacy strict running session | 用户对每个 session 发起 `hint_requested` 并 replay 同一 `clientEventId` | 返回 200 + `show_hint`；hint_text 写入；replay 返回同一 hint snapshot；无 pending event row；不创建 turn-completed outbox；AI 不在 reservation transaction 内执行 |

### E2E.P0.050 provenance and task-run boundary

| Given | When | Then |
|-------|------|------|
| 用户 A 拥有 assisted running session；fake F3/A3 支持 follow-up 与 lightweight observe；task-run writer 可观测 | 用户依次触发 answer_submitted、hint_requested、session_paused、最终 answer_submitted | 每个 AssistantAction provenance JSON 只含 `promptVersion` / `rubricVersion` / `modelId` / `language` / `featureFlag` / `dataSourceVersion`；runtime 字段不出现在 wire；task-run writer 按实际 AI action 记录 typed columns |

### E2E.P0.051 graceful degrade and privacy

| Given | When | Then |
|-------|------|------|
| 用户 A 拥有多个 assisted running session，分别注入 F3 unsupported、A3 secret missing、A3 timeout、invalid output；另有 strict session 作为 boundary | 用户分别发起 `hint_requested` 并运行 runtime boundary lint | assisted failure 返回 200 + `session_wait`，session 保持 running，failure_code 为 NULL，hint_text 为 NULL；failed `ai_task_runs(hint_generate)` 写入 B1 error code；log / metric / audit / event / task-run payload 不含 question/answer/hint/prompt/response/secret；runtime boundary lint 通过 |

## 3 执行入口

```bash
cd backend && go test ./cmd/api -run 'TestE2EP0048|TestE2EP0049|TestE2EP0050|TestE2EP0051' -count=1
```

003 不维护单独 shell scenario 目录；场景编号和 Given/When/Then 由 Go HTTP scenario tests 承接。

## 4 AC 映射

| spec AC / decision | 覆盖场景 |
|--------------------|----------|
| C-7 assisted hint | `E2E.P0.048` |
| C-8 legacy strict optional hint | `E2E.P0.049` |
| C-12 provenance wire boundary | `E2E.P0.050` |
| C-16 privacy redline subset | `E2E.P0.051` |
| C-17 auxiliary AI graceful degrade | `E2E.P0.051` |
| D-11 privacy | `E2E.P0.051` |
| D-37 `hint_generate` task-run evidence | `E2E.P0.048`, `E2E.P0.050`, `E2E.P0.051` |
| D-38 hint lifecycle boundary | `E2E.P0.048` |
