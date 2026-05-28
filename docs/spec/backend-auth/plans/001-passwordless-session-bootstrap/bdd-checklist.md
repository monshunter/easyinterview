# Backend Auth BDD Checklist

> **版本**: 1.3
> **状态**: completed
> **更新日期**: 2026-05-28

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.003 Passwordless session cookie

- [x] 创建场景目录 `test/scenarios/e2e/p0-003-passwordless-session-cookie/`，并确认 `test/scenarios/e2e/INDEX.md` 保持 `E2E.P0.003` 指向该目录
- [x] 准备测试数据：干净用户邮箱、C1 backend-internal mail dispatcher / dev mail sink delivery retrieval、session cookie jar、无效 token、重复 verify token、logout 后 cookie、deleteMe `Idempotency-Key`
- [x] 实现 setup / trigger / verify / cleanup：start challenge -> C1 后台派发器写入 dev mail sink -> 读取 6 位 code -> verify -> `/me` -> `/runtime-config` -> logout -> repeated logout -> logout 后 `/me`；独立登录分支执行 repeated `DELETE /me`
- [x] 断言错误路径：无效 token、重复 verify、缺 cookie / 无效 session 返回 B1 error envelope，响应不泄露账号存在性
- [x] 断言 deleteMe 幂等：相同 `Idempotency-Key` 或等价 active-request dedupe 返回同一 active `privacy_delete` job 或同义终态，不创建重复 job，并撤销当前 session
- [x] 执行并通过场景验证，确认未启动独立 worker 进程也能读取邮件链接，日志、in-process queue、dev sink、future outbox / async payload、audit 证据不含 token、完整 URL、邮箱明文、session cookie 或 secret
  <!-- verified: 2026-05-06 method=scenario scripts=setup,trigger,verify,cleanup run=.test-output/runs/20260506T1911-backend-auth-p0-003/e2e/E2E.P0.003/result.json -->
- [x] 记录验证证据
  <!-- evidence: .test-output/e2e/p0-003-passwordless-session-cookie/trigger.log -->

## E2E.P0.101 Auth email-code single-entry login + profile setup

- [x] 更新场景目录说明与 `test/scenarios/e2e/INDEX.md`，保留目录名但把场景语义改为 single-entry login + profile setup
- [x] 准备测试数据：唯一新邮箱、资料未补全状态、资料补全 displayName、第二 browser context / 无 cookie context、Mailpit code-only 邮件、real frontend/backend API base、session cookie jar
- [x] 实现 setup / trigger / verify / cleanup：single login -> Mailpit code -> manual verify -> `/me.profileCompletionRequired=true` -> refresh profile setup -> close/switch browser and relogin same email -> still profile setup -> `PATCH /me` via frontend -> `/me.profileCompletionRequired=false` -> logout -> login same email -> no profile setup
- [x] 断言 backend contract：start 阶段不返回 duplicate-register / unknown-login 差异；new email verify 创建唯一账号但不把 displayName 写入 verify 前 challenge；`PATCH /me` 要求 session、trimmed displayName 和 `acceptedTerms=true`
- [x] 断言错误/隐私/旧口径路径：邮件、URL、console、scenario evidence 不含旧 URL callback、`/auth/verify?token=`、raw session cookie、`purpose=signup/login` request body、注册按钮、`auth_register` live page、`刘哲` / `Liu Zhe` / `liuzhe@example.com`
- [x] 执行并通过场景验证，记录验证证据
  <!-- verified: 2026-05-28 command="bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/cleanup.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/setup.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/trigger.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/verify.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/cleanup.sh" evidence="first-login-profile-setup profileCompletionRequired=true; cross-browser-relogin-profile-setup true; logout-relogin-profile-setup true; existing-email-login profileCompletionRequired=false; consoleErrors=0 pageErrors=0 httpFailures=0" -->
