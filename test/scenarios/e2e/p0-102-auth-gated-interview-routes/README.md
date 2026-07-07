# E2E.P0.102 Auth-Gated Interview Routes

> **场景 ID**: E2E.P0.102
> **执行方式**: automated
> **隔离级别**: local runner
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

用户未登录时可以打开 `home` 和 auth 页面，但所有面试相关入口必须先进入登录页：

- Home 首页不展示 `Recent mock interviews` 区块，也不调用 `listTargetJobs`。
- Home 上的 JD 导入、简历工作台在未登录时编码 pendingAction 并跳转 `auth_login`；Home 不再提供复盘入口。
- 直接访问 `parse`、`workspace`、`resume_versions`、`practice`、`generating`、`report`、`settings` 时，App 在 `/me` 判定前不得挂载业务 screen；判定未登录后跳转 `auth_login`。Non-current `jd_match`、`debrief`、`debrief_full`、`profile` 输入先归一到当前 route，不作为独立保护 route。
- 后端除 auth start/verify、runtime-config、logout optional 之外的业务 API 均保持 session middleware 保护。

## 2 When

从仓库根目录执行：

```bash
bash test/scenarios/e2e/p0-102-auth-gated-interview-routes/scripts/setup.sh
bash test/scenarios/e2e/p0-102-auth-gated-interview-routes/scripts/trigger.sh
bash test/scenarios/e2e/p0-102-auth-gated-interview-routes/scripts/verify.sh
bash test/scenarios/e2e/p0-102-auth-gated-interview-routes/scripts/cleanup.sh
```

`trigger.sh` 调用 repo-tracked runner：

```bash
node --test ui-design/ui-design-contract.test.mjs
pnpm --filter @easyinterview/frontend test \
  src/app/screens/home/HomeRecentMocks.test.tsx \
  src/app/screens/home/HomeAuthGate.test.tsx \
  src/app/AppAuthDispatch.test.tsx
cd backend && go test ./internal/auth -run TestSessionPolicyClassifiesPublicOptionalAndProtectedOperations -count=1
cd backend && go test ./cmd/api -run 'TestBuildAPIHandlerMounts(TargetJobRoutes|UploadPresign|ResumeRoutes|PracticeRoutes|ReportRoutes|JobRoute)BehindSessionMiddleware|TestBuildAPIHandlerDoesNotMountNonCurrentDebriefOrProfileRoutes|TestJDMatchRoutesRemainUnmountedPerD17' -count=1
```

## 3 Then

- `ui-design/src/app.jsx` 将 `signedIn` 传给 Home，`ui-design/src/screen-home.jsx` 只在 `signedIn` 时渲染 Recent mock interviews。
- 未登录 Home 不渲染 `home-recent-mocks`，不展示后端 raw unauthorized / missing fixture 错误，不调用 `listTargetJobs`。
- 未登录 Home 不渲染复盘 CTA，也不会产生 `pendingRoute=debrief`。
- 未登录点击 Home 业务入口会导航到 `auth_login`，并保留 `pendingType=open_protected_route` 或 `pendingType=import_jd`。
- 未登录或鉴权 loading 直接进入保护路由时，业务 screen 不挂载，前端只允许 runtime-config 与 `/me` 调用。
- 后端 session policy 与 cmd/api route mount 测试证明业务 API 缺 session 返回 `AUTH_UNAUTHORIZED` envelope。

## 4 Cleanup

本场景只运行 unit/contract runner，不创建共享环境资源。输出证据保留在：

```text
.test-output/e2e/p0-102-auth-gated-interview-routes/
```
