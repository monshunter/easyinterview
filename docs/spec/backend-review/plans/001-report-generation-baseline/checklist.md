# 001 — Conversation-level Report Generation Checklist

> **版本**: 2.0
> **状态**: active
> **更新日期**: 2026-07-12

**关联计划**: [plan](./plan.md)

## Phase 1: Contract/prompt
- [ ] 1.1 RED-GREEN: replace QuestionAssessment/retryFocusTurnIds with DimensionAssessment/retryFocusCompetencyCodes across OpenAPI/fixtures/generated types.
- [ ] 1.2 RED-GREEN: remove question_assessments/report.question_assessment and update report.generate prompt/rubric/schema/evals/seeds.

## Phase 2: Generate/store
- [ ] 2.1 RED-GREEN: load ordered practice_messages and generate/persist session-level report.
- [ ] 2.2 RED-GREEN: readiness/dimension validation and AI retry/failure matrix pass.
- [ ] 2.3 BDD-Gate: P0.056/P0.058/P0.099 pass for generate/failure/real integration.

## Phase 3: Read/replay
- [ ] 3.1 RED-GREEN: queued/generating/ready/failed get/list mappings use new shape.
- [ ] 3.2 RED-GREEN: retry plan uses competency codes and no turn IDs.
- [ ] 3.3 BDD-Gate: P0.056/P0.057/P0.058 pass for read/retry states.

## Phase 4: Privacy/closeout
- [ ] 4.1 RED-GREEN: isolation/redaction/current-scope negative tests pass.
- [ ] 4.2 Substitute gate: focused isolation/redaction/current-scope negative tests pass.
- [ ] 4.3 Run focused/full backend, prompt/eval, migration/codegen/fixture/docs/diff gates.
