# 001 — Plan and Session Orchestration BDD Checklist

> **版本**: 1.3
> **状态**: completed
> **更新日期**: 2026-07-10

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.022 Baseline Plan Create/Read

- [x] Scenario directory, README, seed, expected outcome and setup/trigger/verify/cleanup scripts exist.
- [x] Scenario validates ready target job + flat resume, plan creation, user-scoped read and cross-user 404.
- [x] Scenario execution passed and evidence was captured.

## E2E.P0.023 Start Session First Turn

- [x] Scenario directory, README, seed, expected outcome and setup/trigger/verify/cleanup scripts exist.
- [x] Scenario validates running session, current turn, session event and started outbox.
- [x] Scenario execution passed and evidence was captured.

## E2E.P0.024 AI Failure Retry

- [x] Scenario directory, README, seed, expected outcome and setup/trigger/verify/cleanup scripts exist.
- [x] Scenario validates mapped AI failure, failed reservation, retry success and single outbox emission.
- [x] Scenario execution passed and evidence was captured.

## E2E.P0.025 Idempotency And Isolation Matrix

- [x] Scenario directory, README, seed, expected outcome and setup/trigger/verify/cleanup scripts exist.
- [x] Scenario validates replay, mismatch conflict, cross-user isolation, concurrent start and 404 boundaries.
- [x] Scenario execution passed and evidence was captured.

## E2E.P0.026 Privacy And Observability Redlines

- [x] Scenario directory, README, seed, expected outcome and setup/trigger/verify/cleanup scripts exist.
- [x] Scenario validates audit/outbox/metric/log redaction and AI task typed metadata.
- [x] Scenario execution passed and evidence was captured.

## Phase 9 GET fixture harness regression

- [x] `E2E.P0.022` completes setup / trigger / verify / cleanup after plan fixture harness consolidation.
  <!-- verified: 2026-07-10 method=p0-022-get-fixture-regression evidence="In-process HTTP scenario passes and cleanup removes only its setup marker." -->
- [x] `E2E.P0.023` completes setup / trigger / verify / cleanup after session fixture harness consolidation.
  <!-- verified: 2026-07-10 method=p0-023-get-fixture-regression evidence="In-process HTTP scenario passes and cleanup removes only its setup marker." -->
