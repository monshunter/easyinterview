# E2E.P0.025 Practice Idempotency And Isolation Matrix

> **场景 ID**: E2E.P0.025
> **执行方式**: automated
> **隔离级别**: in-process (Go HTTP tests)
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

用户 A 拥有两个 ready baseline plan；用户 B 拥有一个独立 ready baseline plan。两名用户都已登录，并故意复用一组相同 `Idempotency-Key`。

## 2 When

场景依次触发：同 user + 同 key + 同 fingerprint replay；同 user + 同 key + 不同 fingerprint mismatch；用户 B 同 key 独立 start；同 user + 不同 key + 同 plan 第二次 start；用户 B GET 用户 A 的 plan / session。

## 3 Then

replay 返回首次 session 且不重复 outbox；mismatch 返回 `409 PRACTICE_SESSION_CONFLICT` 且不泄露首次 session；用户 B 同 key 独立成功；同 plan 多 key 返回 `409 PRACTICE_SESSION_CONFLICT`；跨用户 GET plan/session 分别返回 `PRACTICE_PLAN_NOT_FOUND` / `PRACTICE_SESSION_NOT_FOUND`。

## 4 执行

```bash
./test/scenarios/e2e/p0-025-practice-idempotency-and-isolation-matrix/scripts/setup.sh
./test/scenarios/e2e/p0-025-practice-idempotency-and-isolation-matrix/scripts/trigger.sh
./test/scenarios/e2e/p0-025-practice-idempotency-and-isolation-matrix/scripts/verify.sh
./test/scenarios/e2e/p0-025-practice-idempotency-and-isolation-matrix/scripts/cleanup.sh
```

## 5 污染控制

当前脚本使用 `cmd/api` HTTP 场景测试与 in-process store snapshot，覆盖 idempotency replay / mismatch / cross-user namespace / same-plan active session conflict / cross-user GET 404。`cleanup.sh` 只清理 setup marker，保留 `.test-output/e2e/p0-025-practice-idempotency-and-isolation-matrix/trigger.log` 与 `result.json` 作为 BDD evidence。
