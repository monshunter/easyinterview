# Frontend Owner Real Backend Handoff Gates 交付复盘报告

> **日期**: 2026-05-23
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：加固 `/plan-code-review` 的 frontend-first handoff L2 gate，并反扫 `frontend-workspace-and-practice`、`frontend-report-dashboard`、`frontend-resume-workshop`、`frontend-debrief` 已完成 plans 的 real-backend drift。
- 修复内容：新增集中 real-mode generated-client gate、shared scenario trigger/verify helper，接入 P0.018-P0.021、P0.044-P0.047、P0.056-P0.059、P0.065-P0.069、P0.081-P0.087，并清理 stale backend-missing 文档口径。
- Bug 记录：[BUG-0089](../bugs/BUG-0089.md)。
- 成功证据：
  - `python3 .agent-skills/skill-creator/scripts/quick_validate.py .agent-skills/plan-code-review && python3 -m pytest .agent-skills/plan-code-review/scripts/test_plan_code_review_skill.py -q`：13 passed。
  - `VITE_EI_API_MODE=real VITE_EI_API_BASE_URL=http://localhost:8080/api/v1 pnpm --filter @easyinterview/frontend exec vitest run src/api/frontendOwners.realApiMode.test.ts`：1 test passed。
  - P0.018 / P0.044 / P0.047 / P0.056 / P0.065 / P0.081 representative `trigger.sh && verify.sh` 均通过；P0.065 触发全量 frontend Vitest，213 files / 1297 tests passed。
  - `make docs-check`、`git diff --check`、scenario shell `bash -n` 全部通过。

## 2 会话中的主要阻点/痛点

- Frontend-first plans 的历史 fixture-first wording 在 backend owner 已完成后没有自动回填，导致多个 completed plan 同时保留 future / missing / Phase 0 stale 口径。
  - **证据**：负向搜索在 workspace/report/resume/debrief specs/plans/checklists 中命中多处 stale wording。
  - **影响**：后续 L2 review 容易把 fixture UI PASS 误读为真实 backend 闭环。
- Scenario wrapper 原本只证明 fixture-backed UI variants，缺少 real-mode marker 和 generated-client request contract。
  - **证据**：新增 shared helper 前，P0 frontend owner trigger 没有统一 `VITE_EI_API_MODE=real` gate；verify 也不检查 base URL / test file marker。
  - **影响**：真实 backend landing 后，frontend scenarios 仍可能在 mock-only 证据下继续 PASS。
- P0.044/P0.047 的旧负向断言把 `createPracticeVoiceTurn` 当作缺失契约，和 `practice-voice-mvp/001` 已完成事实冲突。
  - **证据**：代表性 P0.044 首次重跑失败，指向 `hooks/usePracticeVoiceTurn.ts`；修正 verify allowlist 后 P0.044/P0.047 均通过。
  - **影响**：负向断言若不随相邻 owner landing 反转，会把当前正确实现误判为 drift。

## 3 根因归类

- `skill`：`/plan-code-review` 缺少 frontend-first handoff 后的 sibling completed plan sweep 和 scenario real-mode marker 要求。
- `spec-plan`：completed frontend plans 没有把 backend owner landing 作为原地修订 trigger，operation matrix 和 BDD evidence 容易停留在历史完成态。
- `test`：scenario verify 的负向断言过于 repo-wide，没有表达 owner allowlist，尤其是 voice owner 后续接管同一 `PracticeScreen` 目录的情况。

## 4 对流程资产的改进建议

- 已完成：把 frontend-first handoff gate 固化进 `.agent-skills/plan-code-review/SKILL.md`，并用 skill tests 覆盖。
  - **落点**：skill
  - **优先级**：high
- 已完成：把 BUG-0089 挂到 `docs/bugs/PATTERNS.md` 模式 6，作为未来同类漂移的诊断入口。
  - **落点**：bug pattern
  - **优先级**：high
- 后续建议：对其他 frontend-first completed plans 做一轮轻量巡检，重点找“backend owner 已完成但 frontend scenario 仍 mock-only”的组合。
  - **落点**：spec-plan / scenario
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：继续推进 S3 真联调，但入口应从真实 backend scenario / dev stack 证据开始，不再接受 fixture-only UI PASS 作为闭环。
- 次优先级：把 P0.065 trigger 的参数问题单独修掉；它当前会跑全量 frontend Vitest，虽然通过，但反馈成本偏高。
- 可延后：把 React `act(...)` warning 作为测试卫生项集中处理；本次不阻塞 real-backend gate 成功，但会增加日志噪音。
