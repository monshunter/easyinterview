# Grounded Conversation Report Test Checklist

> **版本**: 2.22
> **状态**: active
> **更新日期**: 2026-07-14

**关联 Test Plan**: [test-plan](./test-plan.md)

## Contract and frozen context

- [x] Closed OpenAPI/codegen/fixtures and lossless persistence/API projection tests pass.
- [x] Frozen-context mutation/mismatch tests reject mutable fallback and keep raw context out of non-content stores/logs.
- [x] Historical committed input fixtures are deleted; only small output fixtures and their manifest remain where required.

## Validator, retry and persistence

- [x] Wire、24/64、delimiter、focus/action and fail-closed table tests pass.
- [x] Invocation-local initial+3、10s/20s/40s recorder、dynamic repair、cancellation and second-invocation reset tests pass.
- [x] PostgreSQL lease-takeover tests preserve stale-worker zero report/outbox/audit/job side effects without a report-level retry counter.

## Eval and real UI separation

- [x] Evalkit generation/judge independent budgets and terminal-valid-negative behavior pass as code/eval gates.
- [x] P0.099 independently binds current-run real report/generating API/DB/screenshot evidence; eval output digest is not required.
- [x] Exact 24/64 is proven by code tests; real P0.099 images prove only current legal content is fully visible.

## Canonical-round overview

- [x] Minimal wire、canonical order、independent current/latest selection、tie-break、nullable/error enum and whole-response fail-closed tests pass.
- [ ] Generated/fixture handoff、ReportsScreen-only consumer negative and scoped stale-pointer/pagination searches pass.

## Injected report input guard

- [x] A4 single owner contract covers default/override/invalid/cross-field rules.
- [x] One small injected admitted/overflow provider call/no-call guard passes without reconstructing the historical 62,397-byte input, default-sized material or `input-*.json`.
- [x] A3 loader/coverage gates require six active profiles at 16K or above and report context at 1M without a byte/token capacity formula.
- [x] Configuration guard is not represented as BDD/E2E.

## Full regression

- [x] Development uses focused tests for feedback; phase completion runs root `make test` for the complete backend/frontend unit regression.
