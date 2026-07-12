# 003 — Remove Dedicated Assistance Modes Checklist

> **版本**: 2.0
> **状态**: completed
> **更新日期**: 2026-07-12

**关联计划**: [plan](./plan.md)

## Phase 1: Contract/config deletion
- [x] 1.1 RED-GREEN: remove PracticeMode/hint fields/actions/events from shared/OpenAPI/DB/generated artifacts.
- [x] 1.2 RED-GREEN: remove lightweight-observe prompt/rubric/profile/eval/seed/task references.
- [x] 1.3 RED-GREEN: remove practice hint/assistance feature flags and runtime allowlist/tests.

## Phase 2: Runtime/frontend/report deletion
- [x] 2.1 RED-GREEN: delete backend hint service/store branches and frontend hint UI/hook/context/handoff.
- [x] 2.2 RED-GREEN: ordinary help-request message uses sendPracticeMessage with no special metadata.
- [x] 2.3 BDD-Gate: P0.051 assistance negative scenario passes.

## Phase 3: Scenario/docs closeout
- [x] 3.1 Delete P0.048-P0.050 positive scenario assets and index rows; rewrite P0.051.
- [x] 3.2 Run zero-reference, focused/full, config/prompt/codegen/migration/docs/diff gates.
