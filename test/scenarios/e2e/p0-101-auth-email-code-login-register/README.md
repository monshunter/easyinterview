# E2E.P0.101 Auth Email-Code Single-Entry Login/Profile Setup

> **场景 ID**: E2E.P0.101
> **执行方式**: automated
> **隔离级别**: shared-cluster + host-run frontend/backend + Mailpit
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

本地共享环境已通过 `test/scenarios/env-setup.sh --with-migrations` 应用当前 schema，并通过 `test/scenarios/env-redeploy.sh all` 或等价 `/scenario-env redeploy all` 刷新并启动：

- frontend dev server: `http://127.0.0.1:5173`
- backend API: `http://127.0.0.1:8080/api/v1`
- Mailpit Web/API: `http://127.0.0.1:8025`
- `deploy/dev-stack/.env` 中 `VITE_EI_API_MODE=real` 且 `VITE_EI_API_BASE_URL` 指向 backend。
- Postgres `users` 表包含 `profile_completed_at` 与 `terms_accepted_at`，否则 `setup.sh` 会在消费验证码前失败并提示先运行 migrations。

场景使用 setup 生成的一个 synthetic `.example.test` 邮箱。该邮箱从单一登录入口首次登录时创建用户，但用户资料保持未补全；在完成 displayName + 条款确认前，刷新、深链、关闭浏览器后重登、换浏览器后重登、退出后重登都必须停在 `auth_profile_setup`。资料补全后，同一邮箱后续登录直接进入正常账号。它不使用真实外部邮箱账号，不直接写 `sessions` 表，也不启动场景专属 backend helper。

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

- 未登录 TopBar 只有登录入口；`/auth/register` 或旧 `auth_register` 不 materialize live page。
- `/auth/login` 输入新邮箱后，前端进入 `auth_verify` 页面并显示邮箱提示。
- Mailpit 邮件只包含 6 位验证码，不包含 frontend `/auth/verify?token=...` URL callback 或 backend verify API URL。
- 输入验证码后，frontend 调用 backend `verifyAuthEmailChallenge` 兑换 `ei_session`，`GET /me` 返回 200 且 `profileCompletionRequired=true`，页面进入 `auth_profile_setup`。
- 资料补全前刷新 `auth_profile_setup`、深链到业务 route、关闭 browser context 后重登同一邮箱、换 browser context 后重登同一邮箱、退出后重登同一邮箱，均继续停在 `auth_profile_setup`，不得恢复 pending action。
- `auth_profile_setup` 提交 trimmed displayName 和 `acceptedTerms=true` 后调用 `PATCH /me`，`GET /me.profileCompletionRequired=false`，TopBar 展示该 displayName。
- 退出后使用同一邮箱再次登录时不再进入 `auth_profile_setup`，直接进入正常账号。
- 所有 `POST /auth/email/start` request body 只包含 email，不包含 `purpose`、`displayName`、password 或 OAuth 字段。
- Playwright 捕获的 console errors / page errors / 非预期 HTTP >=400 failures 均为 0。
- `trigger.log` 与 `result.json` 不包含 raw 验证码、`ei_session` cookie 值、auth secret 或旧 URL callback。

## 4 Cleanup

`cleanup.sh` 只删除本场景 setup 生成邮箱对应的 users / auth_challenges / sessions / idempotency / auth email async job 记录；不清理共享 dev-stack，也不删除 Mailpit 全局邮件。输出证据保留在：

```text
.test-output/e2e/p0-101-auth-email-code-login-register/
```
