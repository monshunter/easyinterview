# 002 Conversation Message Loop Test Checklist

> **版本**: 2.9
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Test Plan**: [test-plan](./test-plan.md)

## Phase 1: Store
- [x] Store reservation/replay/concurrency/uniqueness tests pass.
## Phase 2: Service/API
- [x] Happy message loop and contract tests pass.
## Phase 3: Failure
- [x] Failure/repair/retry tests pass.
## Phase 4: Completion
- [x] Completion/report handoff tests pass.
## Phase 5: Privacy
- [x] Isolation/redaction/full gates pass.
## Phase 6: Review remediation
- [x] Send/complete race and typed-conflict tests pass. (`go test ./backend/internal/practice -count=1`; `go test ./backend/internal/store/practice -count=1`)
## Phase 7: Resume grounding
- [x] Send store precedence and long-input tail-marker tests pass.<!-- verified: 2026-07-12 method=go-test -->
- [x] Empty context returns typed validation with zero AI/assistant reply while preserving user-message retry semantics.<!-- verified: 2026-07-12 method=go-test -->
- [x] System-role policy, JSON untrusted-context escaping and persona-style-only follow-up payload/lint/eval gates pass.<!-- verified: 2026-07-12 method=go+pytest -->

## Phase 8: Completion ledger projection

- [x] Wrong-resume exclusion, duplicate session/event and report-status-independence projection tests provide focused development feedback；phase completion is reported by repository-root `make test`.
## Phase 9: Reportable completion and frozen context

- [x] Focused zero-answer/pending-reply/one-answer and frontend disabled-reason tests provide development feedback.
- [x] Focused report-context consistency, replay and privacy tests provide development feedback；phase completion is reported by repository-root `make test` and PostgreSQL checks remain a separate integration gate.

## Phase 10: Durable reply-state recovery

- [x] Store transition and real PostgreSQL migration tests pass.
- [x] Service/API readback plus generated schema/fixture tests pass.
- [x] Typed TS error runtime JSON/non-JSON/empty/Abort/transport tests pass.

## Phase 11: Lease, generation fence and evidence freshness

- [x] Migration SQL-contract and direct-SQL fixture tests pass for every role/status/generation/lease invariant.
- [x] Store/domain injected-clock transition tests pass at before/exactly-after 90 seconds, including GET and same-ID reserve lazy convergence.
- [x] All four named real PostgreSQL concurrency tests pass with independent connections/start barriers and exact row/reply/generation assertions.
- [x] Service/API/OpenAPI/codegen/fixture regression passes and proves generation/lease remain internal.

## Phase 12: Message/session injected guards

- [x] Small injected single-message ASCII/multibyte acceptance/overflow tests pass.
- [x] Small injected persisted aggregate, replay and concurrent-submit tests pass with zero overflow side effects.
