# Backend Practice Event Loop and Completion Test Checklist

> **版本**: 1.3
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 Test Plan**: [test-plan](./test-plan.md)

## Contract / Drift

- [x] Event/job source-event-only contract is covered（验证：`make lint-events`、`make codegen-events-check`、`go test ./backend/internal/shared/jobs -count=1`）
- [x] OpenAPI turn status and generated artifacts are current（验证：`make codegen-check`、`python3 scripts/lint/conventions_drift.py --repo-root .`）
- [x] PracticeSessions fixtures cover current append and completion variants（验证：`make validate-fixtures`）

## Append Event Tests

- [x] State machine tests cover all five event kinds, answer branches, strict hint default, provenance defaults and malformed payload fail-fast（验证：`cd backend && go test ./internal/practice -count=1`）
- [x] Store tests cover transaction writes, replay/mismatch, row lock sequencing, cross-user boundary and no-audit append behavior（验证：`cd backend && go test ./internal/store/practice -run TestAppendSessionEvent -count=1`）
- [x] Handler tests cover generated request/response mapping, `Idempotency-Key` rejection, required `occurredAt` and error mapping（验证：`cd backend && go test ./internal/api/practice -run TestAppendSessionEvent -count=1`）

## Completion Tests

- [x] Store tests cover queued report/job creation, D-35 replay, status guard, cross-user boundary, outbox and audit rows（验证：`cd backend && go test ./internal/store/practice -run TestCompleteSession -count=1`）
- [x] Handler and middleware tests cover idempotency reserve/replay/mismatch, required `clientCompletedAt` and resource handoff（验证：`cd backend && go test ./internal/api/practice -run TestCompletePracticeSession -count=1`、`cd backend && go test ./internal/middleware/idempotency -count=1`）

## BDD / Runtime Boundary

- [x] BDD P0.038-P0.043 cmd/api scenario suite is covered（验证：`cd backend && go test ./cmd/api -run 'TestE2EP0038|TestE2EP0039|TestE2EP0040|TestE2EP0041|TestE2EP0042|TestE2EP0043' -count=1`）
- [x] Runtime boundary lint is covered（验证：`python3 scripts/lint/backend_practice_non_current.py --repo-root . --phase all`、`python3 -m pytest scripts/lint/backend_practice_non_current_test.py -q`）
- [x] Privacy redaction is covered by unit tests and `TestE2EP0043PracticeEventLoopPrivacyAndNonCurrentNegativeSurface`

## Closeout

- [x] Owner context, docs index and repo whitespace gates are covered（验证：`validate_context.py backend-practice/002 backend`、`sync-doc-index --check`、`make docs-check`、`git diff --check`）
