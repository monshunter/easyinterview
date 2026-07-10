# Backend Practice Event Loop and Completion Checklist

> **版本**: 1.6
> **状态**: completed
> **更新日期**: 2026-07-10

**关联计划**: [plan](./plan.md)

## Phase 0: contract preflight

- [x] 0.1 `shared/jobs.yaml` marks `report_generate` as `triggerEventSemantic: source_event_only` and generated jobs constants expose the source-event-only predicate（验证：`make lint-events`、`make codegen-events-check`、`go test ./backend/internal/shared/jobs -count=1`）
- [x] 0.2 OpenAPI `PracticeTurn.status` contains `asked` / `answered` / `follow_up_requested` / `assessed` and generated Go/TS artifacts are current（验证：`make codegen-check`、`python3 scripts/lint/conventions_drift.py --repo-root .`）
- [x] 0.3 PracticeSessions fixtures cover append main/replay/mismatch/completed variants and completion main/replay/mismatch/cross-user variants（验证：`make validate-fixtures`）
- [x] 0.4 F3 `practice.session.follow_up` is resolvable for follow-up generation（验证：backend-practice preflight tests）

## Phase 1: append event state machine

- [x] 1.1 `SessionEventService` routes `answer_submitted` / `hint_requested` / `session_paused` / `session_resumed`（验证：`TestSessionEventServiceRouteCoversAllKinds`）
- [x] 1.2 `answer_submitted` chooses `ask_follow_up`, `ask_question` or `session_completed` from DB-owned turn/session state（验证：`TestHandleAnswerSubmittedDecisionBranches`）
- [x] 1.3 `hint_requested` remains available for strict and assisted sessions; plan 003 owns AI-backed hint behavior（验证：`TestAppendSessionEventHintStrictModeRunsAIAndAppends`）
- [x] 1.4 turn status mapping preserves all five current values and rejects unknown values（验证：`TestTurnStatus`）
- [x] 1.5 malformed answer payloads and missing client timestamps fail before AI or repository side effects（验证：`TestAppendSessionEventRejectsMissingAnswerText`、`TestAppendSessionEventRequiresOccurredAt`）

## Phase 2: append event vertical slice

- [x] 2.1 Repository persists event, turn/session updates and outbox in one transaction with per-session `seq_no` ordering（验证：`go test ./backend/internal/store/practice -run TestAppendSessionEvent -count=1`）
- [x] 2.2 `clientEventId` replay returns the original result; changed fingerprint returns 409 without duplicate event/outbox/audit rows（验证：repository replay/mismatch tests）
- [x] 2.3 handler rejects `Idempotency-Key` on append and maps generated request/response shapes to B2 contracts（验证：`go test ./backend/internal/api/practice -run TestAppendSessionEvent -count=1`）
- [x] 2.4 BDD-Gate: `E2E.P0.038` answer event main flow is covered（验证：`cd backend && go test ./cmd/api -run TestE2EP0038PracticeEventLoopAnswerFlow -count=1`）
- [x] 2.5 BDD-Gate: `E2E.P0.039` replay/mismatch/kind/header/cross-user matrix is covered（验证：`cd backend && go test ./cmd/api -run TestE2EP0039PracticeEventIdempotencyKindRouterAndHeaderPolicy -count=1`）
- [x] 2.6 BDD-Gate: `E2E.P0.040` stale-turn and sequencing boundary is covered（验证：`cd backend && go test ./cmd/api -run TestE2EP0040PracticeEventConcurrentSeqNoStaleTurnConflict -count=1`）

## Phase 3: completion vertical slice

- [x] 3.1 Repository completes sessions by creating queued `feedback_reports`, queued `async_jobs(report_generate)`, session event, outbox and audit rows in one transaction（验证：`go test ./backend/internal/store/practice -run TestCompleteSession -count=1`）
- [x] 3.2 D-35 replay returns an existing report/job before creating any new handoff rows and binds lookup to `job_type='report_generate'` + `resource_type='feedback_report'` + `dedupe_key=sessionId`（验证：store replay/status-guard tests）
- [x] 3.3 handler uses shared idempotency middleware with `domain=practice` and `operation=completePracticeSession`（验证：`go test ./backend/internal/api/practice -run TestCompletePracticeSession -count=1`）
- [x] 3.4 illegal completion status without an existing report/job returns conflict and creates no report/job/outbox rows（验证：`TestSQLRepositoryCompleteSessionRejectsIllegalStatusWithoutReport`）
- [x] 3.5 BDD-Gate: `E2E.P0.041` completion handoff is covered（验证：`cd backend && go test ./cmd/api -run TestE2EP0041PracticeSessionCompleteCreatesQueuedReportJob -count=1`）
- [x] 3.6 BDD-Gate: `E2E.P0.042` completion idempotency and cross-user matrix is covered（验证：`cd backend && go test ./cmd/api -run TestE2EP0042PracticeSessionCompleteIdempotencyMatrix -count=1`）

## Phase 4: privacy, contract drift and closeout

- [x] 4.1 outbox, audit, log, metric and task-run payloads do not contain question, answer, hint, prompt, response or provider secret text（验证：outbox/redaction tests and P0.043）
- [x] 4.2 runtime boundary lint keeps removed practice terms, duplicate `report_generate` handoff paths and turn-status compression helpers out of runtime/scenario/generated surfaces（验证：`python3 scripts/lint/backend_practice_out_of_scope.py --repo-root . --phase all`）
- [x] 4.3 BDD-Gate: `E2E.P0.043` privacy/runtime boundary is covered（验证：`cd backend && go test ./cmd/api -run TestE2EP0043PracticeEventLoopPrivacyAndOutOfScopeSurface -count=1`）
- [x] 4.4 owner closeout gates are current（验证：`validate_context.py backend-practice/002 backend`、`sync-doc-index --check`、`make docs-check`、`git diff --check`）

## Phase 5: Handler dead helper cleanup

- [x] 5.1 删除 `backend/internal/api/practice/handler.go` 中无调用且与 `stringValue` 重复的 `derefString`；验证: scoped `staticcheck ./internal/api/practice/...`、Practice handler/package gate 与 owner docs gates 通过。
  <!-- verified: 2026-07-10 method=practice-handler-dead-helper-removal evidence="RED: backend staticcheck reported U1000 for the unreferenced derefString duplicate. GREEN: removed the helper; staticcheck ./internal/api/practice/... ./internal/practice/... PASS; go test ./internal/api/practice ./internal/practice ./internal/store/practice ./internal/middleware/idempotency ./internal/ai/aiclient ./internal/ai/registry ./cmd/api -count=1 PASS after the separately owned canonical scenario fixture repair; owner contexts, docs/diff/pruning gates PASS." -->

## 收口命令

- [x] `cd backend && go test ./internal/api/practice ./internal/practice ./internal/store/practice ./internal/middleware/idempotency ./cmd/api -count=1`
- [x] `cd backend && go test ./...`
- [x] `make lint-events`
- [x] `make codegen-events-check`
- [x] `make codegen-check`
- [x] `make validate-fixtures`
- [x] `python3 scripts/lint/conventions_drift.py --repo-root .`
- [x] `python3 scripts/lint/backend_practice_out_of_scope.py --repo-root . --phase all`
- [x] `python3 -m pytest scripts/lint/backend_practice_out_of_scope_test.py -q`
