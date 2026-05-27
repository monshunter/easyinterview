# Backend Auth BDD Checklist

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-05-27

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

## E2E.P0.101 Auth email-code register-then-login

- [x] 将场景目录迁移为 `test/scenarios/e2e/p0-101-auth-email-code-login-register/`，并确认 `test/scenarios/e2e/INDEX.md` 指向该目录
- [x] 准备测试数据：唯一注册/登录邮箱、注册 displayName、重复注册 displayName、Mailpit code-only 邮件、real frontend/backend API base、session cookie jar
- [x] 实现 setup / trigger / verify / cleanup：register -> Mailpit code -> manual verify -> `/me` + TopBar displayName -> logout -> login same email -> Mailpit code -> manual verify -> `/me` same displayName
- [x] 断言重复注册路径：同一邮箱再次走 register 在 start 阶段返回错误，不创建 challenge、不发送新 code、不得覆盖既有 displayName，不得隐式登录成成功注册态，必须提示回到登录路径
- [x] 断言错误/隐私路径：邮件、URL、console、scenario evidence 不含旧 URL callback、`/auth/verify?token=`、raw session cookie、`刘哲` / `Liu Zhe` / `liuzhe@example.com`
- [x] 执行并通过场景验证，记录验证证据
  <!-- evidence: .test-output/e2e/p0-101-auth-email-code-login-register/trigger.log (1 Playwright test passed; register/login same email lifecycle PASS; duplicate-register start rejected without new mail; verify.sh ok) -->
