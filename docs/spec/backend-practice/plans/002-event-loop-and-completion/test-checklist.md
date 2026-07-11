# Backend Practice Event Loop and Completion Test Checklist

> **版本**: 1.6
> **状态**: completed
> **更新日期**: 2026-07-11

**关联 Test Plan**: [test-plan](./test-plan.md)

## Contract / Drift

- [x] Event/job source-event-only contract is covered（验证：`make lint-events`、`make codegen-events-check`、`go test ./backend/internal/shared/jobs -count=1`）
- [x] OpenAPI turn status and generated artifacts are current（验证：`make codegen-check`、`python3 scripts/lint/conventions_drift.py --repo-root .`）
- [x] PracticeSessions fixtures cover current append and completion variants（验证：`make validate-fixtures`）

## Append Event Tests

- [x] State machine tests cover all four current text event kinds, answer branches, optional strict-mode hint, provenance defaults and malformed payload fail-fast（验证：`cd backend && go test ./internal/practice -count=1`）
- [x] Store tests cover transaction writes, replay/mismatch, row lock sequencing, cross-user boundary and no-audit append behavior（验证：`cd backend && go test ./internal/store/practice -run TestAppendSessionEvent -count=1`）
- [x] Handler tests cover generated request/response mapping, `Idempotency-Key` rejection, required `occurredAt` and error mapping（验证：`cd backend && go test ./internal/api/practice -run TestAppendSessionEvent -count=1`）

## Completion Tests

- [x] Store tests cover queued report/job creation, D-35 replay, status guard, cross-user boundary, outbox and audit rows（验证：`cd backend && go test ./internal/store/practice -run TestCompleteSession -count=1`）
- [x] Handler and middleware tests cover idempotency reserve/replay/mismatch, required `clientCompletedAt` and resource handoff（验证：`cd backend && go test ./internal/api/practice -run TestCompletePracticeSession -count=1`、`cd backend && go test ./internal/middleware/idempotency -count=1`）

## BDD / Runtime Boundary

- [x] BDD P0.038-P0.043 cmd/api scenario suite is covered（验证：`cd backend && go test ./cmd/api -run 'TestE2EP0038|TestE2EP0039|TestE2EP0040|TestE2EP0041|TestE2EP0042|TestE2EP0043' -count=1`）
- [x] Runtime boundary lint is covered（验证：`python3 scripts/lint/backend_practice_out_of_scope.py --repo-root . --phase all`、`python3 -m pytest scripts/lint/backend_practice_out_of_scope_test.py -q`）
- [x] Privacy redaction is covered by unit tests and `TestE2EP0043PracticeEventLoopPrivacyAndOutOfScopeSurface`

## Closeout

- [x] Owner context, docs index and repo whitespace gates are covered（验证：`validate_context.py backend-practice/002 backend`、`sync-doc-index --check`、`make docs-check`、`git diff --check`）

## Phase 7 Revision

- [x] Focused tests prove text append and voice chat share canonical server-owned context/session language, use `generation_kind` for follow-up vs next question, and reject client question/intent/follow-up/next-question overrides.
- [x] Parser/language tests prove JSON/schema/business-parser/wrong-language output receives at most one repair with identical context/language, while provider/config/secret/timeout/unsupported/fallback-exhausted errors receive none; a second invalid output cannot emit hard-coded English or canned question.
- [x] Prompt tests prove canonical markers, repair semantics, template hash, baseline seed migration, resolved prompt snapshot and eval cases stay synchronized; prompt hardcode lint finds no Go-built natural-language prompt.
- [x] Text tests prove second invalid output returns `session_wait`, restores pre-event turn control state and suppresses completion outbox; voice tests prove top-level `AI_OUTPUT_INVALID` occurs before result/TTS persistence and leaves the session unchanged.
- [x] Updated P0.038/P0.009, focused/full backend, prompt/eval, OpenAPI/fixture/codegen, privacy/runtime-boundary, context/docs/index/diff gates pass.
  <!-- verified: 2026-07-11 evidence="P0.038-P0.043 direct Go E2E and P0.009 wrapper PASS; full backend, prompt/eval, OpenAPI fixture/codegen, migration, privacy/runtime-boundary, four context and docs/index/diff gates PASS." -->
