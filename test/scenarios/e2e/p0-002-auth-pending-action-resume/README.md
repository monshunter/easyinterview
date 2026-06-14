# E2E.P0.002 Auth Pending-Action Resume

> **场景 ID**: E2E.P0.002
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom)
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

未登录用户处于 `workspace` 路由下，有完整的 plan context（planId / targetJobId
/ jdId / resumeId / roundId）；fixture transport 提供
`getMe` `unauthenticated`、`startAuthEmailChallenge` `default`、
`verifyAuthEmailChallenge` `default`、`getMe` 切换为 `authenticated` 的
mock auth 成功响应。

## 2 When

场景模拟用户点击 `立即面试`，触发 `requestAuth(pendingAction)` 携带 5 个
interview-context 参数；用户在 `auth_login` 输入邮箱、提交挑战、跳转
`auth_verify`，再输入 6 位登录 code 完成 `verifyAuthEmailChallenge`。

## 3 Then

- 点击 `立即面试` 后立即进入 `auth_login`，TopBar `data-signed-in=false`。
- 验证成功后 App 切换到 `practice`，`route-practice` 的 `data-route-params`
  必须包含 planId / targetJobId / jdId / resumeId / roundId 全部 5 个原始值。
- `verify.sh` 负向 grep 必须确认 trigger.log 不出现 `route-auth_login`
  之外的旧 auth alias 与 prototype data 痕迹，且最终断言 5 个 interview
  context key 全部回填。

## 4 执行

```bash
./test/scenarios/e2e/p0-002-auth-pending-action-resume/scripts/setup.sh
./test/scenarios/e2e/p0-002-auth-pending-action-resume/scripts/trigger.sh
./test/scenarios/e2e/p0-002-auth-pending-action-resume/scripts/verify.sh
./test/scenarios/e2e/p0-002-auth-pending-action-resume/scripts/cleanup.sh
```

## 5 污染控制

场景在 vitest + jsdom 中运行，不写共享数据库，不启动 Kind cluster；trigger.sh
仅产生 `.test-output/e2e/p0-002-auth-pending-action-resume/trigger.log` 作为
验证证据，cleanup.sh 移除 setup marker，保留日志。
