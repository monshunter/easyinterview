# Backend Auth BDD Checklist

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-05-06

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.003 Passwordless session cookie

- [x] 创建场景目录 `test/scenarios/e2e/p0-003-passwordless-session-cookie/`，并确认 `test/scenarios/e2e/INDEX.md` 保持 `E2E.P0.003` 指向该目录
- [x] 准备测试数据：干净用户邮箱、C1 backend-internal mail dispatcher / dev mail sink delivery retrieval、session cookie jar、无效 token、重复 verify token、logout 后 cookie、deleteMe `Idempotency-Key`
- [x] 实现 setup / trigger / verify / cleanup：start challenge -> C1 后台派发器写入 dev mail sink -> 读取链接 -> verify -> `/me` -> `/runtime-config` -> logout -> repeated logout -> logout 后 `/me`；独立登录分支执行 repeated `DELETE /me`
- [x] 断言错误路径：无效 token、重复 verify、缺 cookie / 无效 session 返回 B1 error envelope，响应不泄露账号存在性
- [x] 断言 deleteMe 幂等：相同 `Idempotency-Key` 或等价 active-request dedupe 返回同一 active `privacy_delete` job 或同义终态，不创建重复 job，并撤销当前 session
- [x] 执行并通过场景验证，确认未启动独立 C8 worker 进程也能读取邮件链接，日志、in-process queue、dev sink、future outbox / async payload、audit 证据不含 token、完整 URL、邮箱明文、session cookie 或 secret
  <!-- verified: 2026-05-06 method=scenario scripts=setup,trigger,verify,cleanup run=.test-output/runs/20260506T1911-backend-auth-p0-003/e2e/E2E.P0.003/result.json -->
- [x] 记录验证证据
  <!-- evidence: .test-output/e2e/p0-003-passwordless-session-cookie/trigger.log -->
