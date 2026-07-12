# 001 — Plan and Session Orchestration Checklist

> **版本**: 2.1
> **状态**: completed
> **更新日期**: 2026-07-12

**关联计划**: [plan](./plan.md)

## Phase 1: Contract and baseline rebase

- [x] 1.1 RED: add OpenAPI/shared/migration/prompt inventory tests for message conversation and stale-question rejection.
  <!-- verified: 2026-07-12 method=red python3 -m unittest scripts.lint.practice_conversation_contract_test failed on all four pre-change contract surfaces as expected -->
- [x] 1.2 GREEN: replace PracticeTurn/event-answer/question-review schemas with PracticeMessage/sendPracticeMessage/conversation-report schemas; run codegen/fixture gates.
  <!-- verified: 2026-07-12 method=contract make codegen; openapi inventory 20 tests; events inventory 28 tests with 13 current events; lint-openapi; lint-events; validate-fixtures 37 fixtures; git diff --check -->
- [x] 1.3 GREEN: rewrite baseline SQL/enum sources/seeds to 21 app tables with `practice_messages`, 11 shared enums and 6 prompt coordinates; run migration/prompt/rubric/profile/eval gates.
  <!-- verified: 2026-07-12 method=contract migration lint; migrations/codegen/eval Go tests; 11-enum codegen; 6 prompt/rubric/profile lints; eval-offline 24/24; cross-layer contract 4/4 -->
## Phase 2: PracticePlan simplification

- [x] 2.1 RED-GREEN: remove question budget/mode/hints across request, domain, store and fixtures; preserve baseline and derived-plan validation.
- [x] 2.2 RED-GREEN: update frontend start helpers and real-mode contract tests to send only current plan fields.
- [x] 2.3 BDD-Gate: P0.022/P0.070/P0.072 pass for create/read/derived/isolation.
  <!-- verified: 2026-07-12 method=handler-domain-store-scenarios evidence="P0.022 executes real CreatePracticePlan handler plus domain/SQL gates; P0.070/P0.072 execute derived-source behavior and isolation" -->

## Phase 3: Session start with opening message

- [x] 3.1 RED: starter tests require `practice.session.chat` and one opening assistant message, with no currentTurn/first_question.
- [x] 3.2 GREEN: implement three-stage start reservation/generation/persistence and one-repair language/schema validation.
- [x] 3.3 RED-GREEN: timeout/config/invalid-output/idempotency replay produce no duplicate session/message/outbox or canned text.
- [x] 3.4 BDD-Gate: P0.023/P0.024/P0.025 start, failure and replay scenarios pass.
  <!-- verified: 2026-07-12 method=focused-scenarios evidence="P0.023 opening message, P0.024 AI failure, P0.025 replay/pending retry/mismatch/isolation all execute named tests and reject no-test output" -->

## Phase 4: Session read model

- [x] 4.1 RED-GREEN: get/list session store/API return ordered messages and no turn fields.
- [x] 4.2 RED-GREEN: empty/cross-user/missing/deleted context tests pass.

## Phase 5: Privacy and closeout

- [x] 5.1 RED-GREEN: raw message redaction tests cover event/outbox/audit/log/metric/task-run surfaces.
- [x] 5.2 BDD-Gate: P0.026 privacy/observability passes.
  <!-- verified: 2026-07-12 method=privacy-observability-scenario evidence="lifecycle-only outbox, plaintext redaction, metric allowlist and one conversation-level report call pass" -->
- [x] 5.3 Run focused/full backend, codegen, fixture, migration, prompt/eval, context/docs/index and diff gates.

## 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-12 | 2.1 | 经用户批准，将依赖真实 backend handler 的 P0.022 从 contract rebase 后移到 Phase 2，禁止以 fixture-only 证据替代。 |
