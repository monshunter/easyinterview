# Practice Plan and Session Orchestration BDD Checklist

> **版本**: 2.6
> **状态**: active
> **更新日期**: 2026-07-15

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## `BDD.PRACTICE.PLAN.001` plan/session 幂等与失败恢复

- [x] Owner behavior tests 覆盖 identity match、幂等、隔离、opening failure 与零重复事实。
- [x] 根 `make test` 已执行对应 Go tests；该结果是代码层行为证据，不是 E2E PASS。
- [x] 当前无 plan 创建/session start 真实 E2E owner；不创建 shell wrapper 或场景编号。

## Phase 8 BDD 适用性

- [x] 确认 `listPracticeSessions` 删除不创建新的 Behavior ID、BDD/E2E 场景或伪 PASS；代码层替代 gate 全部由主 checklist/test checklist 执行。
  <!-- verified: 2026-07-15 method=plan+scenario-negative result=PASS -->
- [x] 回归 `BDD.PRACTICE.PLAN.001`，证明保留的 `startPracticeSession` / `getPracticeSession` 行为未因公共列表删除而退化。
  <!-- verified: 2026-07-15 method=focused-owner-behavior-tests result=PASS -->
