# E2E.P0.025 Practice Idempotency And Isolation Matrix

> **场景 ID**: E2E.P0.025
> **执行方式**: automated
> **隔离级别**: in-process (focused handler/domain/store tests)
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

一个 running session 已保存 opening assistant message；用户 A 发送一条带稳定 `clientMessageId` 的消息，用户 B 尝试访问同一 session。

## 2 When

场景依次触发：完整 message pair 精确 replay；pending 同 ID 读取不重复调用 AI；AI 失败持久化为 `retryable_failed` 后同 `clientMessageId` 重试；terminal/pending same-ID 拒绝；同 ID 不同正文 mismatch；跨用户 session 访问。

## 3 Then

完整 replay 不重复调用 AI；pending same-ID 不再次调用 AI；只有 `retryable_failed` 能复用原 user message 与同一 ID 重试；terminal/pending 不可重试；mismatch 返回 `409 IDEMPOTENCY_KEY_MISMATCH`；跨用户访问返回 `404 PRACTICE_SESSION_NOT_FOUND`。

## 4 执行

```bash
./test/scenarios/e2e/p0-025-practice-idempotency-and-isolation-matrix/scripts/setup.sh
./test/scenarios/e2e/p0-025-practice-idempotency-and-isolation-matrix/scripts/trigger.sh
./test/scenarios/e2e/p0-025-practice-idempotency-and-isolation-matrix/scripts/verify.sh
./test/scenarios/e2e/p0-025-practice-idempotency-and-isolation-matrix/scripts/cleanup.sh
```

## 5 污染控制

当前脚本组合实际 Handler、domain service 与 SQL repository focused tests，并显式拒绝 `no tests to run` 假阳性。`cleanup.sh` 只清理 setup marker，保留 `.test-output/e2e/p0-025-practice-idempotency-and-isolation-matrix/trigger.log` 与 `result.json` 作为 BDD evidence。
