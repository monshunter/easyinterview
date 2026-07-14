# E2E.P0.036 — Resume Workshop flat list + auth boundary

> Owner: [`frontend-resume-workshop/001-listing-routing-and-detail-readonly`](../../../../docs/spec/frontend-resume-workshop/plans/001-listing-routing-and-detail-readonly/plan.md)
> 关联需求: frontend-resume-workshop C-1, C-2, C-3, C-5, C-6, C-7, C-8, C-9
> 关联 BDD: [bdd-plan.md](../../../../docs/spec/frontend-resume-workshop/plans/001-listing-routing-and-detail-readonly/bdd-plan.md) E2E.P0.036
> 状态: Ready · 执行方式: automated

## 1 范围

验证 Resume Workshop 列表视图的核心 P0 路径：

- Route shell：`resume_versions` 走 `ResumeWorkshopScreen`。
- 未登录态 route-level gate 先跳 `auth_login`，0 触发 `listResumes` / `getResume`。
- 登录态默认渲染 flat table；每个 `listResumes` fixture item 只含精确 9 个 summary 字段并对应一行，列表加载不调用 `getResume`。
- 旧分组 chrome、视图切换和 selected helper 不渲染，也不触发 "即将开放" toast。
- 点击 flat row 的 open action 进入 `resume_versions?resumeId=...` detail，默认 active tab 为 preview，并恰好调用一次 `getResume`。
- React StrictMode 下首个 `listResumes` transport 失败后 registry 释放；用户 retry 恰好发起第 2 个新 transport 并成功。
- DOM 中不出现 out-of-scope route testid（`route-welcome` / `route-mistakes` / `route-drill` / `route-followup` / `route-onboarding` / `route-experiences` / `route-star` / `route-voice`）。

## 2 触发流

- `setup.sh` 准备 `.test-output/e2e/p0-036-resume-flat-list-auth-boundary`
- `trigger.sh` 在仓库根执行 `pnpm --filter @easyinterview/frontend test src/app/scenarios/p0-036-resume-flat-list-auth-boundary.test.tsx`
- `verify.sh` 校验输出含 5 tests passed、summary/list-detail 与 reject→retry exact transport markers、1 test file passed，并对 skip/no-op/React warning/out-of-scope testid 做 negative grep。
- `cleanup.sh` 清理 setup 标记。

## 3 离线限制

vitest scenario test 全部本地执行，使用 in-tree fixtures，不依赖 Docker / Kind / 网络。
