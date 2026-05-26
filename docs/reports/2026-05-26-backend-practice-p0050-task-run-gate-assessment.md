# Backend Practice P0.050 Task-run Gate 交付复盘报告

> **日期**: 2026-05-26
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：确认并修复 `TestE2EP0050PracticeAssistantActionProvenanceAndTaskRuns` 的当前失败，闭环 `backend-practice/003` 的 BDD 文档漂移与 BUG 记录。
- 成功证据：
  - `cd backend && go test ./cmd/api -run '^TestE2EP0050PracticeAssistantActionProvenanceAndTaskRuns$' -count=1 -v`
  - `cd backend && go test ./cmd/api -run 'TestE2EP0048|TestE2EP0049|TestE2EP0050|TestE2EP0051' -count=1 -v`
  - `test/scenarios/e2e/p0-050-practice-hint-provenance-task-runs/scripts/trigger.sh`
  - `test/scenarios/e2e/p0-050-practice-hint-provenance-task-runs/scripts/verify.sh`
  - `git diff --check`
  - `make docs-check`

## 2 会话中的主要阻点/痛点

- `P0.050` 的旧断言把全局 `ai_task_runs` 行数和固定下标当作证明。
  - **证据**：focused test 失败输出显示当前合法路径已经包含 `question_generate`、answer summary observation 的 `hint_generate`、`followup_generate` 与 `hint_requested` 的 `hint_generate`。
  - **影响**：BUG-0105 修复 answer_summary 后，旧门禁把真实语义误判成 task-run 数量漂移。
- fake AI 成功响应没有覆盖当前 parser 需要的 observation shape。
  - **证据**：`scenarioPracticeAIClient` 对 `practice.turn.lightweight_observe` 只返回 `hint`；`parseTurnObservation` 在 answer observation 场景需要 `answerSummary`，从而制造 `AI_OUTPUT_INVALID` failed row。
  - **影响**：测试同时混入了真实语义变化和测试 fixture 失真，增加诊断成本。

## 3 根因归类

- 旧 BDD gate 未随 BUG-0105 的 runtime semantic fix 更新。
  - **类别**：spec-plan
- 测试 fixture 没有模拟 `practice.turn.lightweight_observe` 的当前多用途输出契约。
  - **类别**：spec-plan
- 现有流程资产已有 Pattern 4（completed checklist / runner gate 不能当历史证据），本次不需要新增 AGENTS 或 Skill 规则；已在 BUG-0108 与 BDD 修订中固化。
  - **类别**：无需仓库改动

## 4 对流程资产的改进建议

- 后续涉及真实 provider / full funnel runtime 语义修复时，把相邻 completed BDD gate 的固定行数、固定下标、fixture shape 假设列入 post-pass reconcile。
  - **落点**：spec-plan
  - **优先级**：high
- provenance / task-run 测试优先按 action 前后增量证明，而不是使用全局行数证明。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 推荐下一步：在后续 BUG-0105 类 full funnel 修复收口时，主动扫 `backend/cmd/api/*scenario_test.go` 中与该语义相邻的 BDD focused tests，尤其是对 `ai_task_runs`、event count、outbox count 的固定下标断言。
- 可延后：如果同类问题再次出现，再把 “task-run gates use per-trigger deltas” 提炼进 `.agent-skills/plan-code-review` 或 `docs/bugs/PATTERNS.md`。
