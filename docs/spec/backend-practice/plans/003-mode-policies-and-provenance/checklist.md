# 003 — Remove Dedicated Assistance Modes Checklist

> **版本**: 2.1
> **状态**: completed
> **更新日期**: 2026-07-14

**关联计划**: [plan](./plan.md)

## Phase 1: Contract/config deletion
- [x] 1.1 RED-GREEN: remove PracticeMode/hint fields/actions/events from shared/OpenAPI/DB/generated artifacts.
- [x] 1.2 RED-GREEN: remove lightweight-observe prompt/rubric/profile/eval/seed/task references.
- [x] 1.3 RED-GREEN: remove practice hint/assistance feature flags and runtime allowlist/tests.

## Phase 2: Runtime/frontend/report deletion
- [x] 2.1 RED-GREEN: delete backend hint service/store branches and frontend hint UI/hook/context/handoff.
- [x] 2.2 RED-GREEN: delete special metadata/classification for help requests；ordinary help remains an ordinary `sendPracticeMessage` flow owned by `frontend-workspace-and-practice/002-practice-text-event-loop`.

## Phase 3: Scenario/docs closeout
- [x] 3.1 BDD-N/A: this plan only deletes the dedicated assistance contract；remove `bdd-plan.md` / `bdd-checklist.md` and their context/index references, while the text-loop owner retains ordinary-help behavior.
- [x] 3.2 仓库根 `make test` 完成前后端全量单测回归；zero-reference、config/prompt/codegen/migration/docs/diff 作为独立 gates。
