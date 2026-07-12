# 002 Conversation Message Loop Test Plan

> **版本**: 2.4
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

## Phase 7: Resume grounding
- Send SQL precedence matches start exactly and ignores empty JSON `{}` / `null` profiles.
- Long source snapshot tail marker survives into follow-up AI payload without slicing.
- Empty resume context returns typed validation before prompt resolve/AI; user reservation remains retryable and no assistant reply is committed.
- Payload-role tests keep policy in `system` and JSON-encode JD/resume/round/persona/history in the untrusted user message, including quote/newline/closing-tag injection cases; persona is style-only.
- Follow-up reservation re-checks the TargetJob's current resume binding; a stale same-user plan/session cannot append a user or assistant message. Shared generation rejects `finish_reason=length` before committing a reply.

## Phase 8: Completion ledger projection

- Store tests assert completion event/report/job/outbox atomicity and exact replay with no duplicate event.
- Integration tests create a same-user wrong-resume completion plus duplicate completed sessions/events for one round and prove only TargetJob-bound-resume facts reach the read-side distinct projection.
- Report queued/ready/failed variants leave the same completion fact; final round and privacy deletion boundaries are covered.
