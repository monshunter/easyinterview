# 002 — Practice Text Event Loop BDD Checklist

> **版本**: 1.10
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.044

- [x] Scenario directory exists with README, seed input, expected outcome and setup/trigger/verify/cleanup scripts.
- [x] Trigger runs real-mode owner marker plus practice screen/event/action/body tests.
- [x] Verify checks passing Vitest markers, ≥20 runtime testids, no prototype data helpers, no `getFeedbackReport` in practice runtime, and voice turn confined to owner hook.

## E2E.P0.045

- [x] Scenario directory exists with README, seed input, expected outcome and scripts.
- [x] Trigger runs assistance, goal parity, hint, skip, pause/resume, mode switch and strict lock tests.
- [x] Verify checks current goal visibility parity, append idempotency boundary, helper-control policy and forbidden input markers.

## E2E.P0.046

- [x] Scenario directory exists with README, seed input, expected outcome and scripts.
- [x] Trigger runs error, lost-state, conflict, retry and completion hook tests.
- [x] Verify checks timeout recovery, 404 lost state, 409 mismatch refresh, strict conflict and privacy redlines.

## E2E.P0.047

- [x] Scenario directory exists with README, seed input, expected outcome and scripts.
- [x] Trigger runs complete hook, handoff params, body parity, completion UI and privacy tests.
- [x] Verify checks `completePracticeSession` body/header contract, `resumeId` handoff, no `getFeedbackReport` practice call, and voice hook boundary.

## Regression

- [x] Workspace owner P0.018/P0.021 remains compatible with practice route handoff.
- [x] Backend-practice and OpenAPI fixture gates remain compatible with the current generated-client contract.
- [x] Docs, index, diff whitespace and product-scope pruning surface gates pass after owner compression.
