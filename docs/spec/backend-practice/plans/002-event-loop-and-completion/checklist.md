# 002 — Conversation Message Loop and Completion Checklist

> **版本**: 2.0
> **状态**: active
> **更新日期**: 2026-07-12

**关联计划**: [plan](./plan.md)

## Phase 1: Message reservation
- [x] 1.1 RED-GREEN: add message domain/store reservation/replay tests and implementation.
- [x] 1.2 RED-GREEN: enforce client/reply uniqueness and concurrent-new-message conflict.

## Phase 2: Assistant reply
- [x] 2.1 RED: service/API tests require ordinary message pair and no AssistantAction/turn fields.
- [x] 2.2 GREEN: implement chat context, AI call, assistant persistence and replay.
- [x] 2.3 BDD-Gate: P0.044 happy conversation passes.

## Phase 3: Failure and repair
- [x] 3.1 RED-GREEN: timeout/config/provider/schema/language matrix keeps user message retryable and writes no invalid reply.
- [x] 3.2 RED-GREEN: same ID retry and request mismatch behavior is deterministic.
- [x] 3.3 BDD-Gate: P0.046 failure/recovery passes.

## Phase 4: Completion
- [x] 4.1 RED-GREEN: completion creates one conversation-level report/job/outbox with no turn focus.
- [x] 4.2 BDD-Gate: P0.047 completion/generating handoff passes.

## Phase 5: Privacy and closeout
- [x] 5.1 RED-GREEN: ownership/privacy/race/redaction tests pass.
- [ ] 5.2 Run focused/full backend, codegen/fixture/migration/prompt/docs/diff gates.
