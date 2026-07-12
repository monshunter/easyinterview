# 001 Plan and Session Orchestration BDD Checklist

> **版本**: 2.3
> **状态**: completed
> **更新日期**: 2026-07-12

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.022 Plan contract
- [x] Revise data/scripts/README, run and record evidence.

## E2E.P0.023 Session opening
- [x] Revise data/scripts/README, run and record evidence.

## E2E.P0.024 AI failure retry
- [x] Revise data/scripts/README, run and record evidence.

## E2E.P0.025 Idempotency and isolation
- [x] Revise data/scripts/README, run and record evidence.

## E2E.P0.026 Privacy and stale-contract negative
- [x] Revise data/scripts/README, run and record evidence.

## Phase 6 resume grounding
- [x] P0.023 executes and verifies complete source snapshot tail-marker grounding.<!-- verified: 2026-07-12 method=scenario -->
- [x] P0.024 executes and verifies empty-context typed failure, zero AI, and no opening assistant message.<!-- verified: 2026-07-12 method=scenario -->

## Phase 7 persisted round identity

- [x] P0.022 executes create/read exact canonical round identity, equal-duration non-reuse and non-contiguous canonical successor.<!-- verified: 2026-07-12 method=scenario-run result=PASS -->
- [x] P0.070 executes same-bound-resume retry/next source semantics plus exact idempotency replay.<!-- verified: 2026-07-12 method=scenario-run result=PASS -->
- [x] P0.072 executes wrong-resume/provenance/type/int32/mismatched/all-complete/legacy/cross-user fail-closed paths.<!-- verified: 2026-07-12 method=unit+scenario-run result=PASS -->
