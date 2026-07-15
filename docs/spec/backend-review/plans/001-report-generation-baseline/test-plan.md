# Grounded Conversation Report Test Plan

> **版本**: 2.26
> **状态**: active
> **更新日期**: 2026-07-15

## 1 Owner test boundary

- Backend code tests live with `backend/internal/review`、`backend/internal/store/review` and `backend/internal/api/reports`; frontend consumer tests live with the relevant frontend modules.
- Focused package/file tests may be used during development, but phase completion and CI regression must run root `make test`, which executes the whole backend/frontend unit suite.
- Scenario shells do not invoke Go test、Vitest/npm test、pytest、lint、build or fixture parity. Only P0.099 remains as real report/generating API/UI evidence.

## 2 Frozen context and contract

- OpenAPI/codegen/fixture tests assert the closed direct-report shape and reject removed/additional fields.
- Review load tests mutate current TargetJob/Resume after completion and prove the generated payload still uses the frozen snapshot and terminal ordered messages.
- Prompt tests prove trusted policy/untrusted JSON separation and no raw context in job/outbox/audit/log/metric.
- Persistence/API tests assert model-owned summary、dimensions、evidence、actions、focus and provenance are lossless while internal anchors stay private.

## 3 Validator and retry

- Table tests cover wire fuse、24/64 language bounds、English delimiter parity、cross-field/focus/action invariants and typed invalid output without raw echo.
- Invocation-local retry tests inject a no-wait recorder and cover attempts 2/3/4 success、attempt4 terminal failure、dynamic scope/full revalidation、retryable provider/protocol errors、non-retryable immediate failure、cancellation and a second independent invocation reset.
- Store/runner integration tests prove `async_jobs.attempts/max_attempts` are infrastructure-only and stale workers cannot write report/outbox/audit/job side effects.

## 4 Eval reliability

- Evalkit and context-aware judge tests are code/eval gates, not E2E. Generation and judge own independent budgets of four; protocol-invalid may retry, valid content rejection is terminal.
- Fixed distinct contexts cover completeness、limited evidence、short answer、pending question and injection. Failure evidence retains only bounded counts、reason/scope、usage/latency and digests.
- Eval result and P0.099 are independent: eval measures content reliability; P0.099 proves the current real report/generating API/UI and visible layout.

## 5 Canonical-round overview

- Contract tests reject pagination/full-report fields and assert minimal nullable current/latest objects.
- Store tests cover canonical order、both-null、prior-ready plus newer failed、generating-only、latest-ready and deterministic tie-breaks.
- Failure/security tests cover hidden 404 and whole-response fail-closed for ownership/context/session/round/generatedAt mismatches.
- Consumer tests prove only ReportsScreen calls `listTargetJobReports`; Parse/Report/Generating do not.

## 6 Injected report input guard

- A4 typed loader/validator owns content-limit default、override、invalid and cross-field contract once.
- Review service keeps one small injected admitted/overflow call/no-call test; the historical 62,397-byte symptom is not reconstructed, and no default-sized payload or `input-*.json` is created.
- A3 loader/coverage tests require all six active profiles `max_tokens >= 16384` and keep report context at 1,000,000. These profile facts remain independent of the byte guard; no cross-unit capacity formula or provider smoke is used.
- This configuration logic has no BDD/E2E. It closes with owner contract tests、focused business tests and root `make test`.

## 7 Report conversation read

- Store/handler tests cover reportId/current-user lookup, existing FK/unique ownership, strict sequence projection and queued/generating/ready/failed success without AI or writes.
- Positive tests cover an owned empty `messages` array as a 200 projection. Negative tests cover hidden 404, report/session/user/target mismatch, empty identity, blank message content, missing createdAt, duplicate/non-increasing sequence, unknown role and any internal/additional response field.
- Privacy tests assert no session/message/client IDs, reply state or raw transcript in errors/logs/audit/metrics/task payloads; successful, business-error and session-middleware rejection responses all set `Cache-Control: private, no-store` before any report read; read count is bounded to the report lookup plus ordered messages query without list/pagination.
- Scoped removal tests reject `listPracticeSessions` in current OpenAPI/generated/router/handler/fixture/mock/frontend positive surfaces and prove no migration/table/compatibility layer is introduced.
