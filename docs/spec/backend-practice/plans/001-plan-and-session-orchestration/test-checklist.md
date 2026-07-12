# 001 Plan and Session Orchestration Test Checklist

> **版本**: 2.0
> **状态**: completed
> **更新日期**: 2026-07-12

**关联 Test Plan**: [test-plan](./test-plan.md)

## Phase 1: Contract tests
- [x] Phase 1 contract/migration/prompt negative and generation tests pass.
  <!-- verified: 2026-07-12 method=contract 5 cross-layer + 28 event inventory + 22 event baseline tests; OpenAPI 37 fixtures; migration lint and Go tests; 6 prompt/rubric/profile lints; offline eval 24/24; context/docs/index/diff gates -->

## Phase 2: Plan tests
- [x] Phase 2 plan validation/store/idempotency/isolation tests pass.

## Phase 3: Start tests
- [x] Phase 3 opening/failure/repair/replay/privacy tests pass.

## Phase 4: Read tests
- [x] Phase 4 ordered/empty/missing/cross-user/list tests pass.

## Phase 5: Gate set
- [x] Phase 5 focused/full/codegen/migrate/docs/diff gates pass.
