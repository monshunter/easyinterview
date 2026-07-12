# 003 — Remove Dedicated Assistance Modes

> **版本**: 2.0
> **状态**: completed
> **更新日期**: 2026-07-12

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

删除已经没有独立产品价值的 hint/strict/assisted 分支：用户通过普通聊天请求帮助；后端、OpenAPI、DB、Prompt、feature flags、frontend context 和报告不再维护专用 assistance 状态或 provenance action。

## 2 删除范围

- `PracticeMode` shared enum 与 plan.mode。
- `hintsEnabled`、`hintRequested`、`show_hint`、`session_wait` action、hint text/count/usage context。
- `practice.turn.lightweight_observe` prompt/rubric/profile/eval/task type。
- 已删除的 practice hint/assistance flags 与 runtime allowlist 投影。
- HintBanner/button/hook/tests/fixtures/scenarios。
- P0.048-P0.050 positive hint scenario assets；P0.051 重写为 stale assistance negative gate。

## 3 质量门禁分类

- **Plan 类型**: product simplification + contract deletion + code cleanup。
- **TDD 策略**: negative tests first assert removed fields/actions/flags/routes cannot compile or pass lint; then delete implementation and generated artifacts.
- **BDD 策略**: P0.051 verifies user can ask for help as ordinary message and no dedicated hint/mode behavior appears.
- **替代验证 gate**: codegen/config/prompt/migration/current-scope negative searches.

## 4 Coverage Matrix

| Area | Category | Phase | Verification |
|------|----------|-------|--------------|
| ordinary help request | primary | 2 | message-loop test + P0.051 |
| shared/API/DB deletion | contract | 1 | codegen/migration negative |
| frontend context deletion | cross-layer | 2 | typecheck/navigation tests |
| config/prompt deletion | regression | 1 | config/prompt/profile/eval lint |
| stale scenarios/docs | negative | 3 | zero-reference search/docs-check |

## 5 实施步骤

### Phase 1: Contract/config deletion
- Delete enum/fields/event/actions/prompt/profile/flags/DB columns and regenerate.

### Phase 2: Runtime/frontend/report deletion
- Delete hint service/store/frontend UI/context/handoff/report display.
- Prove “给我一点提示” goes through ordinary `sendPracticeMessage` with no special metadata.

### Phase 3: Scenario and docs closeout
- Remove P0.048-P0.050 directories/index rows.
- Rewrite P0.051 as assistance-mode stale-contract negative scenario.
- Run full negative search and owner gates.

## 6 验收标准

- Current tree has no positive dedicated hint/mode contract.
- Ordinary conversation can contain user requests for help without special classification.
- Generated/shared/config/prompt/migration assets stay internally consistent.

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-12 | 2.0 | Reopen to delete dedicated hint/strict/assisted behavior and route help through chat. |
