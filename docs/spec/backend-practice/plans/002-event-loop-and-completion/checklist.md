# 002 — Conversation Message Loop and Completion Checklist

> **版本**: 2.7
> **状态**: active
> **更新日期**: 2026-07-13

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
- [x] 5.2 Run focused/full backend, codegen/fixture/migration/prompt/docs/diff gates.

## Phase 6: Review remediation
- [x] 6.1 RED-GREEN: assistant commit locks/checks mutable session state, rolls back after completion wins, and maps the conflict without reopening the session. (`go test ./backend/internal/practice -count=1`; `go test ./backend/internal/store/practice -count=1`)
- [x] 6.2 RED-GREEN: P0.046/P0.047 trigger and verifier require provider-failure, replay, mismatch, pending-retry and late-reply lifecycle PASS markers. (`python3 -m pytest scripts/lint/scenario_script_contract_test.py -q -k practice_failure_and_completion`; focused Go tests; `bash -n`)
- [x] 6.3 BDD-Gate: P0.046 and P0.047 failure/recovery and completion scenarios pass. (serial `setup.sh` → `trigger.sh` → `verify.sh` → `cleanup.sh`, both PASS)

## Phase 7: Complete resume grounding for follow-up messages
- [x] 7.1 RED: send store/service tests require the same snapshot/original/profile precedence, complete long-input tail marker, and zero AI/assistant reply for empty context.<!-- verified: 2026-07-12 method=go-test-red tests=TestSQLRepositoryReservePracticeMessageRetriesPendingUserMessage,TestSendPracticeMessageFailsClosedWithoutResumeContextAndSkipsAI -->
- [x] 7.2 GREEN: message reservation returns shared `ResumeContext` with no slicing and common generation fails closed before prompt resolve/AI when empty.<!-- verified: 2026-07-12 method=go-test packages=internal/practice,internal/store/practice -->
- [x] 7.3 E2E.P0.044/P0.046 trigger/verify require named tail-marker and empty-context tests, preserving retry/replay evidence and rejecting skip/no-op.<!-- verified: 2026-07-12 method=scenario both=PASS -->
- [x] 7.4 BDD-Gate: P0.044/P0.046 pass with complete follow-up grounding and empty-context recovery evidence.<!-- verified: 2026-07-12 method=scenario bddChecklist=complete -->
- [x] 7.5 RED-GREEN: follow-up payload tests plus prompt lint/eval prove system-role policy, JSON escaping of untrusted JD/resume/round/persona/history and persona-style-only behavior.<!-- verified: 2026-07-12 method=go+pytest test=TestSendPracticeMessageUsesOrdinaryConversationHistory evidence="full tail marker, system/user roles, closing-tag JSON escape" -->

## Phase 8: Completion ledger as round-progress fact

- [x] 8.1 RED-GREEN: completion commits exactly one `session_completed` event with `completed_at`, report/job/outbox and returns exact replay without duplicate lifecycle facts.<!-- verified: 2026-07-12 method=unit+P0.047 test=TestSQLRepositoryCompleteSessionReplayDoesNotAppendSecondCompletedFact -->
- [x] 8.2 RED-GREEN: only completion facts whose plan resume equals `target_jobs.resume_id` are admitted; duplicate completed sessions/events for one round and report status changes project one completed round and never mutate a TargetJob progress column.<!-- verified: 2026-07-12 method=P0.098 real-postgres markers="wrong-resume-completion-ignored,target-report-status-independent,out-of-order-gap-hidden" -->
- [x] 8.3 BDD-Gate: P0.047 executes completion/replay event evidence; P0.098 proves first-to-next and final-round projection after real completion.<!-- verified: 2026-07-12 method=scenario-run both=PASS -->
- [x] 8.4 Run focused/full backend, migration/OpenAPI, privacy, context/docs/index/diff and no-frontend-business-persistence gates.<!-- verified: 2026-07-12 evidence="completion replay+bound-resume integration; P0.098; storage negative search; make test; migration/OpenAPI/context/docs/index/diff" -->

## Phase 9: Reportable completion and frozen context

- [x] 9.1 RED-GREEN: sole-owner test `TestE2EP0047RejectsZeroAnswerCompletion` proves zero committed user messages or pending assistant reply returns VALIDATION_FAILED, keeps session mutable and writes no completion/report/job/outbox/idempotency success; one committed user message succeeds. (`cd backend && go test ./internal/api/practice ./internal/practice ./internal/store/practice -run '^TestE2EP0047RejectsZeroAnswerCompletion$' -count=1 -v`; real PostgreSQL)
  <!-- verified: 2026-07-12 method=red-green+postgres evidence="RED undefined reportability contract; GREEN exact test PASS in api/practice/store packages, adjacent package suites PASS, integration TestIntegrationE2EP0047RejectsZeroAnswerCompletion PASS against version=17 Postgres with zero invalid side effects and one-answer success; marker ZERO_ANSWER_COMPLETION_REJECTED_PASS" -->
- [x] 9.2 RED-GREEN: sole-owner tests `TestE2EP0047FreezesReportContext` and `TestE2EP0047CompletionReplayPreservesReportContext` prove successful completion atomically writes full report-context.v1 + terminal coordinate from one consistent DB view; concurrent target/resume mutation, mismatch and replay pass with no AI call. (`cd backend && go test ./internal/api/practice ./internal/practice ./internal/store/practice -run '^(TestE2EP0047FreezesReportContext|TestE2EP0047CompletionReplayPreservesReportContext)$' -count=1 -v`)
  <!-- verified: 2026-07-12 method=red-green+postgres-v18 evidence="exact API/domain/store tests emit REPORT_CONTEXT_SNAPSHOT_PASS and REPORT_CONTEXT_REPLAY_PASS; completion uses read-write repeatable-read; a tagged Postgres advisory gate proves a concurrent TargetJob mutation blocks behind completion row locks, then report-context.v1/terminal 3:3 persists from the pre-mutation view; replay stays byte-equivalent after further TargetJob/Resume edits, mismatch writes zero side effects, and a second isolated rerun passes" -->
- [x] 9.3 RED-GREEN: frontend Practice Finish is disabled before the first committed user message and exposes a localized accessible reason; backend remains authoritative.
  <!-- verified: 2026-07-12 method=focused-vitest evidence="24/24 PASS across PracticeScreen, useCompletePracticeSession, target display and zh/en coverage; TestE2EP0047RejectsZeroAnswerCompletion proves opening-only and draft input do not count, native disabled + aria-describedby exposes localized reason, pending assistant reply also disables, and no completion request is sent" -->
- [x] 9.4 BDD-Gate: P0.047 runs the three exact owner tests and writes `completion-backend-evidence.json` schema `practice-completion-evidence.v1` with `ZERO_ANSWER_COMPLETION_REJECTED_PASS`, `REPORT_CONTEXT_SNAPSHOT_PASS`, `REPORT_CONTEXT_REPLAY_PASS`; PASS requires command exit 0, exact RUN/PASS markers, no FAIL/no-test marker, zero-answer no-side-effect DB assertions and same-snapshot replay. P0.056/058 consume rather than duplicate it.
  <!-- verified: 2026-07-12 method=scenario-run evidence="E2E.P0.047 setup/trigger/verify/cleanup PASS; exact 3-package owner command and tagged v18 PostgreSQL test pass with no FAIL/no-test; verifier alone writes schema-valid redacted artifact; cleanup leaves only completion-backend-evidence.json" -->

## Phase 10: Server-recoverable message reply state

- [ ] 10.1 RED: store/service/API/OpenAPI tests prove failed reservations lack durable/public recovery status and `getPracticeSession` cannot return the original replay identity; generated TS error tests prove `retryable` is dropped.
- [ ] 10.2 GREEN: baseline migration and store add user-only `reply_status=pending|retryable_failed|terminal_failed|complete`; reserve/fail/commit transitions are atomic, user-scoped and preserve unique user/reply rows.
- [ ] 10.3 GREEN: generated `PracticeMessage` exposes user `clientMessageId/replyStatus`; `getPracticeSession` fixtures cover pending/retryable/terminal/complete and assistant messages omit recovery fields.
- [ ] 10.4 GREEN: generated TS `ApiClientError` preserves HTTP status plus parsed `ApiErrorResponse`; JSON/non-JSON/empty/Abort/transport tests pass and no consumer parses `Error.message`.
- [ ] 10.5 BDD-Gate: P0.046 proves AI failure → reload/readback → same-ID retry → one assistant reply, plus pending/terminal/cross-user/privacy gates; P0.044 remains the immediate-send pending/success owner.
- [ ] 10.6 Run focused/full backend, OpenAPI/codegen/fixture, migration, frontend composed owner, context/docs/index/diff gates and restore completed only after current evidence is recorded.

## 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-13 | 2.7 | Reopen for durable reply status and refresh-safe same-ID recovery. |
| 2026-07-12 | 2.6 | 锁定 002 completion 唯一 owner、精确 P0.047 tests/markers/artifact。 |
| 2026-07-12 | 2.5 | 要求至少一条 candidate user message 后才能 completion，并原子冻结 report-context.v1。 |
| 2026-07-12 | 2.4 | 完成事实限定 TargetJob 绑定 resume，并增加 system policy / JSON 不可信 follow-up 上下文分层 gate。 |
