# 001 — Conversation Report Screen Checklist

> **版本**: 2.0
> **状态**: active
> **更新日期**: 2026-07-12

**关联计划**: [plan](./plan.md)

## Phase 1: UI truth source
- [ ] 1.1 RED-GREEN: rewrite screen-report/data/generating prototype to four conversation-level surfaces.
- [ ] 1.2 RED-GREEN: remove perQuestion/Questions/hint/phone source and update geometry.

## Phase 2: Formal structure
- [ ] 2.1 RED-GREEN: delete QuestionsTab/question body/summary and implement four cards/tabs default readiness.
- [ ] 2.2 RED-GREEN: simplify ContextStrip and i18n/a11y/responsive behavior.

## Phase 3: Data states
- [ ] 3.1 RED-GREEN: consume dimensionAssessments/retryFocusCompetencyCodes and render evidence.
- [ ] 3.2 RED-GREEN: queued/ready/failed/notFound/missing/empty states pass.
- [ ] 3.3 BDD-Gate: P0.056/P0.058 pass.

## Phase 4: Replay/next
- [ ] 4.1 RED-GREEN: competency-focused retry and next-round fresh session paths pass.
- [ ] 4.2 BDD-Gate: P0.057 passes.

## Phase 5: Parity/real
- [ ] 5.1 Run full frontend/i18n/typecheck/build/source/pixel parity/negative gates.
- [ ] 5.2 BDD-Gate: P0.059 passes.
- [ ] 5.3 Run P0.099 real browser path and capture report screenshots.
