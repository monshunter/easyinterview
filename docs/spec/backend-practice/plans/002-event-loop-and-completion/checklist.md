# 002 — Conversation Message Loop and Completion Checklist

> **版本**: 2.0
> **状态**: active
> **更新日期**: 2026-07-12

**关联计划**: [plan](./plan.md)

## Phase 1: Message reservation
- [ ] 1.1 RED-GREEN: add message domain/store reservation/replay tests and implementation.
- [ ] 1.2 RED-GREEN: enforce client/reply uniqueness and concurrent-new-message conflict.

## Phase 2: Assistant reply
- [ ] 2.1 RED: service/API tests require ordinary message pair and no AssistantAction/turn fields.
- [ ] 2.2 GREEN: implement chat context, AI call, assistant persistence and replay.
- [ ] 2.3 BDD-Gate: P0.044 happy conversation passes.

## Phase 3: Failure and repair
- [ ] 3.1 RED-GREEN: timeout/config/provider/schema/language matrix keeps user message retryable and writes no invalid reply.
- [ ] 3.2 RED-GREEN: same ID retry and request mismatch behavior is deterministic.
- [ ] 3.3 BDD-Gate: P0.046 failure/recovery passes.

## Phase 4: Completion
- [ ] 4.1 RED-GREEN: completion creates one conversation-level report/job/outbox with no turn focus.
- [ ] 4.2 BDD-Gate: P0.047 completion/generating handoff passes.

## Phase 5: Privacy and closeout
- [ ] 5.1 RED-GREEN: ownership/privacy/race/redaction tests pass.
- [ ] 5.2 Run focused/full backend, codegen/fixture/migration/prompt/docs/diff gates.
