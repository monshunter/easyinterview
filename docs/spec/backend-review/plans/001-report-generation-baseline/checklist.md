# 001 — Conversation-level Report Generation Checklist

> **版本**: 2.1
> **状态**: completed
> **更新日期**: 2026-07-12

**关联计划**: [plan](./plan.md)

## Phase 1: Contract/prompt
- [x] 1.1 RED-GREEN: replace QuestionAssessment/retryFocusTurnIds with DimensionAssessment/retryFocusCompetencyCodes across OpenAPI/fixtures/generated types.
- [x] 1.2 RED-GREEN: remove question_assessments/report.question_assessment and update report.generate prompt/rubric/schema/evals/seeds.

## Phase 2: Generate/store
- [x] 2.1 RED-GREEN: load ordered practice_messages and generate/persist session-level report.
- [x] 2.2 RED-GREEN: readiness/dimension validation and AI retry/failure matrix pass.
- [x] 2.3 BDD-Gate: P0.056/P0.058/P0.099 pass for generate/failure/real integration.

## Phase 3: Read/replay
- [x] 3.1 RED-GREEN: queued/generating/ready/failed get/list mappings use new shape.
- [x] 3.2 RED-GREEN: retry plan uses competency codes and no turn IDs.
- [x] 3.3 BDD-Gate: P0.056/P0.057/P0.058 pass for read/retry states.

## Phase 4: Privacy/closeout
- [x] 4.1 RED-GREEN: isolation/redaction/current-scope negative tests pass.
- [x] 4.2 Substitute gate: focused isolation/redaction/current-scope negative tests pass.
- [x] 4.3 Run focused/full backend, prompt/eval, migration/codegen/fixture/docs/diff gates.

## Phase 5: Review remediation
- [x] 5.1 RED-GREEN: report prompt/output schema declare candidate score range `1.0-5.0` and distinguish it from evaluator rubric thresholds. (`python3 -m pytest scripts/lint/practice_conversation_contract_test.py -q -k candidate_score`; `python3 -m pytest scripts/lint/prompt_lint_test.py -q -k numeric_schema_bounds`; `make lint-prompts`)
- [x] 5.2 RED-GREEN: runtime rejects missing, duplicate or out-of-range dimension scores before readiness/persistence and maps valid boundaries deterministically. (`go test ./backend/internal/review -count=1`)
- [x] 5.3 BDD-Gate: P0.056 and P0.058 valid-report and invalid-output scenarios pass. (serial `setup.sh` → `trigger.sh` → `verify.sh` → `cleanup.sh`, both PASS)
