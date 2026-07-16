# Grounded Conversation Report Test Checklist

> **版本**: 2.28
> **状态**: active
> **更新日期**: 2026-07-16

**关联 Test Plan**: [test-plan](./test-plan.md)

## Contract and frozen context

- [x] Closed OpenAPI/codegen/fixtures and lossless persistence/API projection tests pass.
- [x] Frozen-context mutation/mismatch tests reject mutable fallback and keep raw context out of non-content stores/logs.
- [x] Historical committed input fixtures are deleted; only small output fixtures and their manifest remain where required.

## Validator, retry and persistence

- [x] Wire、24/64、delimiter、focus/action and fail-closed table tests pass.
- [x] Invocation-local initial+3、10s/20s/40s recorder、dynamic repair、cancellation and second-invocation reset tests pass.
- [x] Every reachable validation code maps to a meaningful repair family；multi-family intent, trusted user-seq allowlist, unknown-code fail-closed and literal marker escaping tests pass.
- [x] PostgreSQL lease-takeover tests preserve stale-worker zero report/outbox/audit/job side effects without a report-level retry counter.

## Eval and real UI separation

- [x] Evalkit generation/judge independent budgets and terminal-valid-negative behavior pass as code/eval gates.
- [x] P0.099 independently binds current-run real report/generating API/DB/screenshot evidence; eval output digest is not required.
- [x] Exact 24/64 is proven by code tests; real P0.099 images prove only current legal content is fully visible.

## Canonical-round overview

- [x] Minimal wire、canonical order、independent current/latest selection、tie-break、nullable/error enum and whole-response fail-closed tests pass.
- [x] Generated/fixture handoff、ReportsScreen-only consumer negative and scoped stale-pointer/pagination searches pass.

## Injected report input guard

- [x] A4 single owner contract covers default/override/invalid/cross-field rules.
- [x] One small injected admitted/overflow provider call/no-call guard passes without reconstructing the historical 62,397-byte input, default-sized material or `input-*.json`.
- [x] A3 loader/coverage gates require six active profiles at 16K or above and report context at 1M without a byte/token capacity formula.
- [x] Configuration guard is not represented as BDD/E2E.

## Report conversation read

- [x] Four-status owned report, owned empty `messages` 200, strict ordered closed projection and zero AI/write/new-table tests pass.
  <!-- verified: 2026-07-15 method=go-test files=backend/internal/store/review/report_conversation_test.go,backend/internal/api/reports/report_conversation_test.go -->
- [x] Missing/cross-user/identity/empty-identity/blank-content/missing-createdAt/order/role/additional-field failures are hidden or fail closed without partial response; success, business-error and session-middleware rejection responses are all `private, no-store`.
  <!-- verified: 2026-07-15 method=go-test cases=hidden-404-identity-malformed-reportId-no-read-blank-content-role-sequence-closed-json,mux-auth-no-store bug=BUG-0173 -->
- [x] Session/message/client IDs and transcript body are absent from non-content response/error/log/audit/metric/task surfaces.
  <!-- verified: 2026-07-15 method=go-test source-negative=read-only-no-raw-error-no-locators -->
- [x] Current positive surface contains no `listPracticeSessions`; generated/fixture/mock/frontend handoff and no-migration audit pass.

## Full regression

- [x] Development uses focused tests for feedback; phase completion runs root `make test` for the complete backend/frontend unit regression.
- [x] After Phase 12, root `make test` and contract/docs/context gates pass again before completion.
  <!-- verified: 2026-07-15 method=root-test+docs-context evidence="make test, make docs-check, both owner context validators, git diff --check PASS" -->
- [x] After Phase 14, rerun root `make test`, build, docs/context/index and diff gates before restoring completed lifecycle.
  <!-- verified: 2026-07-16 evidence="make test 584 Python/4583 subtests + Go all + frontend 126/1026; build/context/docs/index/diff PASS" -->

## Failed report regeneration

- [x] Handler/service/store RED covers same-ID success/replay and typed failure matrix.
- [x] Atomic reset/job/audit and job-before-report lock-order tests pass.
- [x] PostgreSQL concurrency/finalize-window tests prove one active job and no deadlock/duplicate generation.
- [x] Focused owner tests and root `make test` pass with evidence recorded after GREEN.
  <!-- verified: 2026-07-16 evidence="focused service/store/handler plus real PostgreSQL integration PASS; root regression PASS" -->
