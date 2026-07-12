# 001 — Practice Voice Disabled Boundary Checklist

> **版本**: 2.0
> **状态**: active
> **更新日期**: 2026-07-12

**关联计划**: [plan](./plan.md)

## Phase 1: Frontend/prototype
- [ ] 1.1 RED-GREEN: phone icon is native disabled with unavailable a11y/copy and no handler/route change.
- [ ] 1.2 RED-GREEN: remove PhoneSurface/controllers/hooks/prototype and positive UI tests.

## Phase 2: Backend guard
- [ ] 2.1 RED-GREEN: voice endpoint returns AI_UNSUPPORTED_CAPABILITY before audio/provider/store.
- [ ] 2.2 RED-GREEN: disabled profiles and zero side effects are proven; fixture is disabled-only.

## Phase 3: Scenarios/closeout
- [ ] 3.1 Rewrite P0.007; delete P0.008/P0.009 positive assets/index rows.
- [ ] 3.2 BDD-Gate: P0.007 disabled boundary passes.
- [ ] 3.3 Run frontend/backend/profile/codegen/privacy/parity/negative/docs gates.
