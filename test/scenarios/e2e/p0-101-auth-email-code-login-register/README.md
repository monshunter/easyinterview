# E2E.P0.101 Auth Email-Code Same-Email Register/Login

> **场景 ID**: E2E.P0.101
> **执行方式**: automated
> **隔离级别**: shared-cluster + host-run frontend/backend + Mailpit
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

本地共享环境已通过 `test/scenarios/env-redeploy.sh all` 或等价 `/scenario-env redeploy all` 刷新并启动：

- frontend dev server: `http://127.0.0.1:5173`
- backend API: `http://127.0.0.1:8080/api/v1`
- Mailpit Web/API: `http://127.0.0.1:8025`
- `deploy/dev-stack/.env` 中 `VITE_EI_API_MODE=real` 且 `VITE_EI_API_BASE_URL` 指向 backend。

场景使用 setup 生成的一个 synthetic `.example.test` 邮箱。该邮箱先注册并创建用户，随后作为同一个用户的登录邮箱再次登录；重复注册同一邮箱必须被拒绝，且不得覆盖原显示名。它不使用真实外部邮箱账号，不直接写 `sessions` 表，也不启动场景专属 backend helper。

## 2 When

从仓库根目录执行：

```bash
bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/setup.sh
bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/trigger.sh
bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/verify.sh
bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/cleanup.sh
```

`trigger.sh` 调用 repo-tracked Playwright spec：

```bash
pnpm --filter @easyinterview/frontend exec playwright test \
  --config=playwright.auth-email-code.config.ts \
  --reporter=list \
  --workers=1 \
  auth-email-code.spec.ts
```

## 3 Then

- `/auth/register` 填写显示名、邮箱并勾选条款后，前端进入 `auth_verify` 页面并显示邮箱提示。
- Mailpit 邮件只包含 6 位验证码，不包含 frontend `/auth/verify?token=...` URL callback 或 backend verify API URL。
- 输入验证码后，frontend 调用 backend `verifyAuthEmailChallenge` 兑换 `ei_session`，随后回到 `/`。
- `/auth/login` 使用同一邮箱再次发送 6 位验证码并登录同一个用户。
- 重复注册同一邮箱时，注册 start 请求返回冲突错误，页面停留在 `auth_register` 并显示失败状态；Mailpit 不新增验证码邮件；原显示名仍为 `Runtime Verify`，不会被 `Runtime Duplicate` 覆盖。
- 注册与登录成功后的 `GET /me` 均返回 200。
- Playwright 捕获的 console errors / page errors / 非预期 HTTP >=400 failures 均为 0。
- `trigger.log` 与 `result.json` 不包含 raw 验证码、`ei_session` cookie 值、auth secret 或旧 URL callback。

## 4 Cleanup

`cleanup.sh` 只删除本场景 setup 生成邮箱对应的 users / auth_challenges / sessions / idempotency / auth email async job 记录；不清理共享 dev-stack，也不删除 Mailpit 全局邮件。输出证据保留在：

```text
.test-output/e2e/p0-101-auth-email-code-login-register/
```
