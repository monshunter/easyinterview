# E2E.P0.024 Practice Session AI Failure Retry

> **场景 ID**: E2E.P0.024
> **执行方式**: automated
> **隔离级别**: shared-cluster
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

已登录用户 A 拥有 `practice_plans(status='ready', goal='baseline')`；F3 baseline active；A3 fake AIClient 第一次返回 `AI_PROVIDER_TIMEOUT`，第二次返回合法 first-question JSON。

## 2 When

用户 A 用同一个 `Idempotency-Key` 执行两次 `POST /api/v1/practice/sessions`：第一次触发 AI timeout，第二次用同 key 和同 body 重试。

## 3 Then

第一次返回 `502 + AI_PROVIDER_TIMEOUT`，错误 envelope 不含 prompt / response 明文；store snapshot 记录 `idempotency_records.status=failed_retryable`、一条 failed session、无 `practice.session.started` outbox。第二次返回 `201 + PracticeSession{status:'running', currentTurn:{turnIndex:1,status:'asked'}}`，同一 idempotency record 升级为 `succeeded`，且 `practice.session.started` outbox 只出现一次。

## 4 执行

```bash
./test/scenarios/e2e/p0-024-practice-session-ai-failure-retry/scripts/setup.sh
./test/scenarios/e2e/p0-024-practice-session-ai-failure-retry/scripts/trigger.sh
./test/scenarios/e2e/p0-024-practice-session-ai-failure-retry/scripts/verify.sh
./test/scenarios/e2e/p0-024-practice-session-ai-failure-retry/scripts/cleanup.sh
```

## 5 污染控制

当前脚本使用 `cmd/api` HTTP 场景测试与 in-process store snapshot，覆盖 auth middleware、practice handler/service、fake F3/A3、failed_retryable 状态机和 retry success commit。`cleanup.sh` 只清理 setup marker，保留 `.test-output/e2e/p0-024-practice-session-ai-failure-retry/trigger.log` 与 `result.json` 作为 BDD evidence。
