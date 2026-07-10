# 001 — Plan and Session Orchestration BDD Plan

> **版本**: 1.3
> **状态**: completed
> **更新日期**: 2026-07-10

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 场景矩阵

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| `E2E.P0.022` | Baseline plan create/read | 登录用户拥有 ready target job 与 ready flat resume | `createPracticePlan` then `getPracticePlan` | 201 + ready plan；`sourceReportId` 为空；audit metadata 无明文；跨用户 read 返回 404 | `test/scenarios/e2e/p0-022-practice-plan-baseline-create-and-read/` |
| `E2E.P0.023` | Start session first turn | ready baseline plan + active first-question profile | `startPracticeSession` then `getPracticeSession` | 201 + running session + current turn；started event/outbox 写入一次 | `test/scenarios/e2e/p0-023-practice-session-start-and-first-question/` |
| `E2E.P0.024` | AI failure retry | first-question AI first call fails, second succeeds | two `startPracticeSession` calls with same key | first response maps to B1 `AI_*` code；retry returns running session and one started outbox | `test/scenarios/e2e/p0-024-practice-session-ai-failure-retry/` |
| `E2E.P0.025` | Idempotency and isolation matrix | users A/B, ready plans and repeated keys | replay, mismatch, cross-user and concurrent requests | replay / conflict / 404 / single-executor behavior matches B1/B4 | `test/scenarios/e2e/p0-025-practice-idempotency-and-isolation-matrix/` |
| `E2E.P0.026` | Privacy and observability redlines | plan/session flows produce audit, outbox and AI task rows | inspect DB/log/metric/audit/outbox | no prompt/response/question/answer/hint/secret text; AI task typed columns populated | `test/scenarios/e2e/p0-026-practice-observability-and-privacy-redlines/` |

## 2 数据隔离

Each scenario owns its user IDs, plan IDs, session IDs and idempotency keys. Cleanup removes scenario-owned practice rows, idempotency rows, audit rows, outbox rows and users in dependency order.

## 3 单元测试边界

BDD verifies observable HTTP / DB / outbox behavior. Go tests own internal state machine, SQL scan, prompt rendering, error mapping, redaction and idempotency edge cases.

## 4 Phase 9 regression focus

`E2E.P0.022` and `E2E.P0.023` retain the complete plan/session GET behavior gate after the test-only fixture harness consolidation. Both scenarios must run setup / trigger / verify / cleanup serially; no scenario asset or product behavior changes.
