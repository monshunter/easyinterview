# E2E.P0.003 Passwordless Session Cookie

> **场景 ID**: E2E.P0.003
> **执行方式**: automated
> **隔离级别**: shared-cluster
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

用户使用干净邮箱请求 passwordless challenge。C1 backend-internal mail dispatcher 可以把一次性验证链接写入 dev mail sink，场景持有 cookie jar、无效 token、重复 verify token、logout 后 cookie 和 `DELETE /me` idempotency key。

## 2 When

场景执行 `start challenge -> dev mail sink retrieval -> verify -> /me -> /runtime-config -> logout -> repeated logout -> logout 后 /me`；随后通过独立登录分支执行 repeated `DELETE /me`。

## 3 Then

服务端签发 `ei_session`，`/me` 返回 masked user context，runtime-config 只返回 A4 allowlist，logout 清除 session 且重复调用幂等；`DELETE /me` 返回 B2 `202 + PrivacyRequestWithJob` 并复用同一 active 删除请求；无效 / 重复 token 与 logout 后 `/me` 返回 B1 error envelope。日志、in-process queue、dev sink、future outbox / async payload 与 audit 证据不包含 token、完整 URL、邮箱明文、session cookie 或 secret。

## 4 执行

```bash
./test/scenarios/e2e/p0-003-passwordless-session-cookie/scripts/setup.sh
./test/scenarios/e2e/p0-003-passwordless-session-cookie/scripts/trigger.sh
./test/scenarios/e2e/p0-003-passwordless-session-cookie/scripts/verify.sh
./test/scenarios/e2e/p0-003-passwordless-session-cookie/scripts/cleanup.sh
```

## 5 污染控制

场景使用 Go `httptest` 和内存 store，不写共享数据库，不启动独立 C8 worker 进程。`cleanup.sh` 只清理场景临时 marker，保留 `.test-output/e2e/p0-003-passwordless-session-cookie/trigger.log` 作为验证证据。
