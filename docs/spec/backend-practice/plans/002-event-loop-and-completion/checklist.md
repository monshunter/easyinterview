# 002 — Conversation Message Loop and Completion Checklist

> **版本**: 2.4
> **状态**: completed
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

## 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-12 | 2.4 | 完成事实限定 TargetJob 绑定 resume，并增加 system policy / JSON 不可信 follow-up 上下文分层 gate。 |
