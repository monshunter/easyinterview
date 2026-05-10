# E2E.P0.023 Practice Session Start And First Question

> **场景 ID**: E2E.P0.023
> **执行方式**: automated
> **隔离级别**: shared-cluster
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

已登录用户 A 拥有 `practice_plans(status='ready', goal='baseline')`；F3 baseline active；A3 fake AIClient 返回合法 first-question JSON。`APP_ENV=test` 使用 `cmd/api` HTTP 场景测试和 in-process store snapshot。

## 2 When

用户 A 执行 `POST /api/v1/practice/sessions` 携带 `Idempotency-Key` 与 `planId`，随后 `GET /api/v1/practice/sessions/{sessionId}`。

## 3 Then

`POST` 返回 `201 + PracticeSession{status:'running', currentTurn:{turnIndex:1, status:'asked', questionText, questionIntent, askedAt}}`；store snapshot 记录 `practice_sessions(status='running')`、1 条 first turn、1 条 `session_started` event、1 条 `practice.session.started` outbox；outbox payload 符合 B3 schema 且不含 question / answer / hint / prompt / response 明文；AI 调用发生在 repository transaction window 外；`GET` 返回完整 running session。

## 4 执行

```bash
./test/scenarios/e2e/p0-023-practice-session-start-and-first-question/scripts/setup.sh
./test/scenarios/e2e/p0-023-practice-session-start-and-first-question/scripts/trigger.sh
./test/scenarios/e2e/p0-023-practice-session-start-and-first-question/scripts/verify.sh
./test/scenarios/e2e/p0-023-practice-session-start-and-first-question/scripts/cleanup.sh
```

## 5 污染控制

当前脚本使用 `cmd/api` HTTP 场景测试，覆盖 auth middleware、practice handler/service、F3 fake resolver、A3 fake AIClient、session reservation/commit 和 outbox payload builder。`cleanup.sh` 只清理 setup marker，保留 `.test-output/e2e/p0-023-practice-session-start-and-first-question/trigger.log` 与 `result.json` 作为 BDD evidence。`result.json.validBddEvidence=true`。
