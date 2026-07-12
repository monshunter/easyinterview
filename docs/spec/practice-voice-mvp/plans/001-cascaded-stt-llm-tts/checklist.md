# 001 — Practice Voice Disabled Boundary Checklist

> **版本**: 2.0
> **状态**: completed
> **更新日期**: 2026-07-12

**关联计划**: [plan](./plan.md)

## Phase 1: Frontend/prototype
- [x] 1.1 RED-GREEN: phone icon is native disabled with unavailable a11y/copy and no handler/route change.
- [x] 1.2 RED-GREEN: remove PhoneSurface/controllers/hooks/prototype and positive UI tests.

## Phase 2: Backend guard
- [x] 2.1 RED-GREEN: voice endpoint returns AI_UNSUPPORTED_CAPABILITY before audio/provider/store.
- [x] 2.2 RED-GREEN: disabled profiles and zero side effects are proven; fixture is disabled-only.

## Phase 3: Scenarios/closeout
- [x] 3.1 Rewrite P0.007; delete P0.008/P0.009 positive assets/index rows.
- [x] 3.2 BDD-Gate: P0.007 disabled boundary passes.
  <!-- verified: 2026-07-12 method=voice-disabled-scenario evidence="PracticeScreen native disabled control and real voice handler 422 AI_UNSUPPORTED_CAPABILITY pass" -->
- [x] 3.3 Run frontend/backend/profile/codegen/privacy/parity/negative/docs gates.
