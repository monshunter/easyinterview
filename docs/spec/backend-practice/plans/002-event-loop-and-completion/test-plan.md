# 002 Conversation Message Loop Test Plan

> **版本**: 2.1
> **状态**: completed
> **更新日期**: 2026-07-12

## Phase 1: Store
- Reserve/replay/mismatch/concurrency/sequence/reply uniqueness and rollback tests.

## Phase 2: Service/API
- Happy multi-message flow, canonical context, ordered history, generated DTO and fixture parity.

## Phase 3: Failure
- Provider/config/timeout/schema/language repair matrix and same-ID recovery.

## Phase 4: Completion
- Idempotent completion, report job/outbox and no question assessment handoff.

## Phase 5: Privacy
- Cross-user 404, raw content redaction and full gates.

## Phase 6: Review remediation
- Repository/service tests simulate completion winning before assistant commit and require rollback plus typed conflict.
- Scenario contract tests require P0.046/P0.047 to execute every named failure/replay/lifecycle assertion and reject no-test output.
