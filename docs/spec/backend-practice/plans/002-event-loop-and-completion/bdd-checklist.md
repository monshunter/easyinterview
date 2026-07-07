# Backend Practice Event Loop and Completion BDD Checklist

> **版本**: 1.3
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 可执行 BDD 证据

- [x] `E2E.P0.038` append answer flow is covered（验证：`cd backend && go test ./cmd/api -run TestE2EP0038PracticeEventLoopAnswerFlow -count=1`）
- [x] `E2E.P0.039` append replay/mismatch/kind/header/cross-user matrix is covered（验证：`cd backend && go test ./cmd/api -run TestE2EP0039PracticeEventIdempotencyKindRouterAndHeaderPolicy -count=1`）
- [x] `E2E.P0.040` append sequencing and stale-turn boundary is covered（验证：`cd backend && go test ./cmd/api -run TestE2EP0040PracticeEventConcurrentSeqNoStaleTurnConflict -count=1`）
- [x] `E2E.P0.041` completion handoff is covered（验证：`cd backend && go test ./cmd/api -run TestE2EP0041PracticeSessionCompleteCreatesQueuedReportJob -count=1`）
- [x] `E2E.P0.042` completion idempotency matrix is covered（验证：`cd backend && go test ./cmd/api -run TestE2EP0042PracticeSessionCompleteIdempotencyMatrix -count=1`）
- [x] `E2E.P0.043` privacy and runtime boundary is covered（验证：`cd backend && go test ./cmd/api -run TestE2EP0043PracticeEventLoopPrivacyAndNonCurrentNegativeSurface -count=1`）

## 场景断言

- [x] P0.038 verifies `answer_submitted` follow-up, next-question and session-completed branches, server-owned follow-up state, current 5-value turn status and turn-completed outbox behavior.
- [x] P0.039 verifies append replay/mismatch semantics, all five event kinds, strict hint conflict, `Idempotency-Key` header refusal and cross-user 404.
- [x] P0.040 verifies accepted event `seq_no` continuity, stale-turn conflict and sanitized conflict envelope.
- [x] P0.041 verifies HTTP 202 `ReportWithJob`, queued report/job rows, session completion event, outbox row and idempotency snapshot.
- [x] P0.042 verifies same-key replay, mismatch conflict, another-key completed-session replay, cross-user 404 and illegal status conflict without duplicate report/job/outbox rows.
- [x] P0.043 verifies redaction, append no-audit boundary, source-event-only job boundary, current turn-status drift gate and runtime boundary lint.

## 收口命令

- [x] `cd backend && go test ./cmd/api -run 'TestE2EP0038|TestE2EP0039|TestE2EP0040|TestE2EP0041|TestE2EP0042|TestE2EP0043' -count=1`
- [x] `python3 scripts/lint/backend_practice_non_current.py --repo-root . --phase all`
- [x] `python3 -m pytest scripts/lint/backend_practice_non_current_test.py -q`
- [x] `make lint-events`
- [x] `make codegen-events-check`
- [x] `make codegen-check`
- [x] `make validate-fixtures`
