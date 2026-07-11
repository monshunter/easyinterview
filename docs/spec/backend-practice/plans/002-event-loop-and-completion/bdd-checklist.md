# Backend Practice Event Loop and Completion BDD Checklist

> **版本**: 1.6
> **状态**: completed
> **更新日期**: 2026-07-11

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 可执行 BDD 证据

- [x] `E2E.P0.038` append answer flow is covered（验证：`cd backend && go test ./cmd/api -run TestE2EP0038PracticeEventLoopAnswerFlow -count=1`）
- [x] `E2E.P0.039` append replay/mismatch/kind/header/cross-user matrix is covered（验证：`cd backend && go test ./cmd/api -run TestE2EP0039PracticeEventIdempotencyKindRouterAndHeaderPolicy -count=1`）
- [x] `E2E.P0.040` append sequencing and stale-turn boundary is covered（验证：`cd backend && go test ./cmd/api -run TestE2EP0040PracticeEventConcurrentSeqNoStaleTurnConflict -count=1`）
- [x] `E2E.P0.041` completion handoff is covered（验证：`cd backend && go test ./cmd/api -run TestE2EP0041PracticeSessionCompleteCreatesQueuedReportJob -count=1`）
- [x] `E2E.P0.042` completion idempotency matrix is covered（验证：`cd backend && go test ./cmd/api -run TestE2EP0042PracticeSessionCompleteIdempotencyMatrix -count=1`）
- [x] `E2E.P0.043` privacy and runtime boundary is covered（验证：`cd backend && go test ./cmd/api -run TestE2EP0043PracticeEventLoopPrivacyAndOutOfScopeSurface -count=1`）

## 场景断言

- [x] P0.038 verifies `answer_submitted` follow-up, next-question and session-completed branches, server-owned follow-up state, current 4-value turn status and turn-completed outbox behavior.
- [x] Phase 7 P0.038 revision verifies follow-up/next-question generation kind, canonical server-owned context/session language, client question/intent/count/next-question override negative, parser/language invalid exactly-one repair, provider no-repair, and `session_wait` on second invalid output with pre-event turn control state restored, no completion outbox and no canned question.
  <!-- verified: 2026-07-11 command="cd backend && go test ./cmd/api -run 'TestE2EP0038|TestE2EP0040' -count=1" result="PASS" -->
- [x] P0.039 verifies append replay/mismatch semantics, all four current text event kinds, strict-mode hint `show_hint`, `Idempotency-Key` header refusal and cross-user 404.
- [x] P0.040 verifies accepted event `seq_no` continuity, stale-turn conflict and sanitized conflict envelope.
- [x] P0.041 verifies HTTP 202 `ReportWithJob`, queued report/job rows, session completion event, outbox row and idempotency snapshot.
- [x] P0.042 verifies same-key replay, mismatch conflict, another-key completed-session replay, cross-user 404 and illegal status conflict without duplicate report/job/outbox rows.
- [x] P0.043 verifies redaction, append no-audit boundary, source-event-only job boundary, current 4-value turn-status drift gate and runtime boundary lint.

## 收口命令

- [x] `cd backend && go test ./cmd/api -run 'TestE2EP0038|TestE2EP0039|TestE2EP0040|TestE2EP0041|TestE2EP0042|TestE2EP0043' -count=1`
- [x] `python3 scripts/lint/backend_practice_out_of_scope.py --repo-root . --phase all`
- [x] `python3 -m pytest scripts/lint/backend_practice_out_of_scope_test.py -q`
- [x] `make lint-events`
- [x] `make codegen-events-check`
- [x] `make codegen-check`
- [x] `make validate-fixtures`
- [x] Re-run updated P0.038 plus focused follow-up/voice repair tests and existing P0.039-P0.043 privacy/contract gates.
  <!-- verified: 2026-07-11 evidence="P0.038-P0.043 PASS; focused append second-invalid session_wait, wrong-language hint and voice double-invalid no-TTS/no-persistence tests PASS." -->
