# 002 Conversation Message Loop Test Checklist

> **版本**: 2.9
> **状态**: active
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
- [x] P0.046/P0.047 scenario marker contract tests pass. (`python3 -m pytest scripts/lint/scenario_script_contract_test.py -q -k practice_failure_and_completion`; focused Go tests; `bash -n`)
## Phase 7: Resume grounding
- [x] Send store precedence and long-input tail-marker tests pass.<!-- verified: 2026-07-12 method=go-test -->
- [x] Empty context returns typed validation with zero AI/assistant reply while preserving user-message retry semantics.<!-- verified: 2026-07-12 method=go-test -->
- [x] P0.044/P0.046 scenario gates pass.<!-- verified: 2026-07-12 method=scenario -->
- [x] System-role policy, JSON untrusted-context escaping and persona-style-only follow-up payload/lint/eval gates pass.<!-- verified: 2026-07-12 method=go+pytest -->

## Phase 8: Completion ledger projection

- [x] Completion atomicity/replay tests prove exactly one lifecycle fact.<!-- verified: 2026-07-12 method=P0.047 result=PASS -->
- [x] Wrong-resume exclusion, duplicate session/event and report-status-independence projection tests pass.<!-- verified: 2026-07-12 method=P0.098 real-postgres marker=wrong-resume-completion-ignored=PASS -->
- [x] P0.047/P0.098 scenario gates pass with persisted backend evidence.<!-- verified: 2026-07-12 method=scenario-run -->
## Phase 9: Reportable completion and frozen context

- [x] `TestE2EP0047RejectsZeroAnswerCompletion` plus frontend disabled-reason and one-answer success tests pass.<!-- verified: 2026-07-12 method=exact-go+postgres-v18+focused-vitest evidence="backend exact zero-answer/pending-reply/one-answer gate PASS; frontend Practice/i18n completion regression 24/24 PASS with native disabled and localized aria-describedby reason" -->
- [x] `TestE2EP0047FreezesReportContext` / `TestE2EP0047CompletionReplayPreservesReportContext` DB consistency, replay and privacy tests pass.<!-- verified: 2026-07-12 method=exact-go+postgres-v18 evidence="repeatable-read snapshot/replay unit markers plus advisory-gated concurrent mutation blocking, immutable replay, mismatch zero-side-effect and rerun isolation" -->
- [x] P0.047 `completion-backend-evidence.json` matches `practice-completion-evidence.v1`, contains every required marker and has `result=PASS` with no raw content.<!-- verified: 2026-07-12 method=scenario-run evidence="exact top-level keys/tests/markers/redacted DB booleans+counts; result PASS; cleanup retains only JSON artifact" -->

## Phase 10: Durable reply-state recovery

- [x] Store transition and real PostgreSQL migration tests pass.
- [x] Service/API readback plus generated schema/fixture tests pass.
- [x] Typed TS error runtime JSON/non-JSON/empty/Abort/transport tests pass.
- [x] P0.046 reload/same-ID/unique-reply and privacy gates pass.
  <!-- verified: 2026-07-14 method=focused+full+scenario evidence="Store/service/API/generated contract and typed runtime error suites pass; P0.046 run e26ba887-5f71-4c25-834e-448b4595ede2 passes same-ID reload recovery, one assistant reply, privacy and desktop/mobile evidence." -->

## Phase 11: Lease, generation fence and evidence freshness

- [x] Migration SQL-contract and direct-SQL fixture tests pass for every role/status/generation/lease invariant.
- [x] Store/domain injected-clock transition tests pass at before/exactly-after 90 seconds, including GET and same-ID reserve lazy convergence.
- [x] All four named real PostgreSQL concurrency tests pass with independent connections/start barriers and exact row/reply/generation assertions.
- [x] Service/API/OpenAPI/codegen/fixture regression passes and proves generation/lease remain internal.
- [x] P0.044/P0.046 scenario contract, source fingerprint, screenshot hash/geometry and exact marker checks pass with fresh artifacts. (fresh serial runs `13f3b898-4054-4949-8b85-4a15df35c712` / `e26ba887-5f71-4c25-834e-448b4595ede2`)

## Phase 12: Message/session UTF-8 byte limits

- [ ] 32KiB/32KiB+1 single-message ASCII/multibyte tests pass.
- [ ] 256KiB/256KiB+1 persisted aggregate, replay and concurrent-submit tests pass with zero overflow side effects.
- [ ] RuntimeConfig/frontend and P0.046 current evidence pass.
