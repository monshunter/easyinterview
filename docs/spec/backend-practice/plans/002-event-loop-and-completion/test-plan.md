# 002 Conversation Message Loop Test Plan

> **版本**: 2.9
> **状态**: completed
> **更新日期**: 2026-07-14

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

## Phase 9: Reportable completion and frozen context
- Focused completion tests reject zero-user/pending-reply completion before writes/provider and accept one committed user message.
- Focused snapshot/replay tests assert one-view report-context.v1, concurrent mutation isolation, terminal coordinate and exact replay.
- Focused runs are development feedback only；phase completion is reported by repository-root `make test`, with real PostgreSQL checks as a separate integration gate.
- Frontend test asserts Finish disabled/accessibly explained until a committed user message exists.

## Phase 10: Durable reply-state recovery

- Migration/store transition table covers new reservation, same-ID retry, replay, retryable/terminal failure, concurrent new ID, assistant commit and rollback; every transition is user/session scoped.
- Service/API tests require failure status to be committed before the error response and `getPracticeSession` to expose user `clientMessageId/replyStatus` while assistant messages omit both.
- OpenAPI/fixture/generated tests cover pending, retryable-failed, terminal-failed and complete read projections plus typed `ApiClientError` JSON/non-JSON/empty/Abort/transport behavior.
- Integration/BDD proves AI failure → reload → same-ID retry → one user/assistant pair, with no browser business-storage dependency and no raw-content leakage.

## Phase 11: Lease, generation fence and evidence freshness

- Migration RED/GREEN covers role/status/generation/lease constraints: user generation is positive, exactly pending has a lease, assistant has neither, and every legacy/direct-SQL user fixture supplies a valid generation.
- Store unit RED/GREEN uses an injected server clock and covers `new → pending(G1,+90s)`, GET expiry to `retryable_failed(G1)`, same-ID expired/retryable takeover to `pending(G2,+90s)`, Fail/Commit lease clearing, unexpired/different-ID conflicts and stale-generation zero-write conflicts.
- Real PostgreSQL runs exactly the four Phase 11 integration tests with independent connections and a start barrier. Sequential reserve calls do not satisfy the concurrency gate. The stale-worker case must pause G1, expire through GET, reserve G2, release both stale Commit and Fail, then commit G2 and count one assistant reply.
- Service/API tests prove `Now` reaches GET/reserve consistently, generation is carried only across internal Reserve/Commit/Fail calls, stale conflicts map deterministically, and the public response never includes generation/lease.

## Phase 12: Message/session injected guards

- Unit fixtures use small injected values to cover ASCII/multibyte acceptance and overflow without constructing default-sized strings.
- Store/service tests prove authoritative aggregate, replay behavior, zero write/provider on overflow and concurrent total-limit fencing.
