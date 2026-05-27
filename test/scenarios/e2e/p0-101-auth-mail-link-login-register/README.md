# E2E.P0.101 Auth Mail-Link Login/Register Callback

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
- `deploy/dev-stack/.env` 中 `EMAIL_VERIFY_BASE_URL` 指向 frontend `/auth/verify` callback，`VITE_EI_API_MODE=real` 且 `VITE_EI_API_BASE_URL` 指向 backend。

场景使用 setup 生成的两个 synthetic `.example.test` 邮箱，分别覆盖登录与注册。它不使用真实外部邮箱账号，不直接写 `sessions` 表，也不启动场景专属 backend helper。

## 2 When

从仓库根目录执行：

```bash
bash test/scenarios/e2e/p0-101-auth-mail-link-login-register/scripts/setup.sh
bash test/scenarios/e2e/p0-101-auth-mail-link-login-register/scripts/trigger.sh
bash test/scenarios/e2e/p0-101-auth-mail-link-login-register/scripts/verify.sh
bash test/scenarios/e2e/p0-101-auth-mail-link-login-register/scripts/cleanup.sh
```

`trigger.sh` 调用 repo-tracked Playwright spec：

```bash
pnpm --filter @easyinterview/frontend exec playwright test \
  --config=playwright.auth-mail-link.config.ts \
  --reporter=list \
  --workers=1 \
  auth-mail-link.spec.ts
```

## 3 Then

- `/auth/login` 提交邮箱后，前端进入 `auth_verify` 页面并显示邮箱提示。
- `/auth/register` 填写显示名、邮箱并勾选条款后，前端进入 `auth_verify` 页面并显示邮箱提示。
- 两封 Mailpit 邮件都包含 frontend callback：`http://127.0.0.1:5173/auth/verify?token=<redacted>`，不得回退到 backend verify API URL。
- 打开邮件链接后，frontend 自动调用 backend `verifyAuthEmailChallenge` 兑换 `ei_session`，随后 replace 清理 URL token 并回到 `/`。
- 登录与注册两个独立浏览器上下文中的 `GET /me` 均返回 200。
- Playwright 捕获的 console errors / page errors / HTTP >=400 failures 均为 0。
- `trigger.log` 与 `result.json` 不包含 raw magic-link token、`ei_session` cookie 值、auth secret 或 backend verify API 邮件链接。

## 4 Cleanup

`cleanup.sh` 只删除本场景 setup 生成邮箱对应的 users / auth_challenges / sessions / idempotency / auth email async job 记录；不清理共享 dev-stack，也不删除 Mailpit 全局邮件。输出证据保留在：

```text
.test-output/e2e/p0-101-auth-mail-link-login-register/
```
