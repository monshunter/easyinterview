# Auth Gated Interview Routes 交付复盘报告

> **日期**: 2026-05-28
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付修复未登录 Home 与面试相关路由的认证边界：Home 不再展示 Recent mock interviews 或触发受保护 API；非 Home、非 `auth_*` 的业务路由在未登录时统一进入 `auth_login` 并携带 pendingAction；后端业务 API 继续由 session middleware 保护。
- Owner 文档已同步：`docs/spec/frontend-shell/spec.md`、`docs/ui-design/auth-and-entry.md`、`docs/spec/frontend-shell/plans/001-app-shell-auth-settings/*`、`test/scenarios/e2e/INDEX.md`。
- 成功证据：
  - `node --test ui-design/ui-design-contract.test.mjs`
  - `pnpm --filter @easyinterview/frontend test`
  - `pnpm --filter @easyinterview/frontend typecheck`
  - `go test ./internal/auth -run TestSessionPolicyClassifiesPublicOptionalAndProtectedOperations -count=1`
  - `go test ./cmd/api -run 'TestBuildAPIHandlerMounts(TargetJobRoutes|UploadPresign|ResumeRoutes|PracticeAndProfileRoutes|ReportRoutes|JobRoute)BehindSessionMiddleware|TestJDMatchRoutesRequireSessionOnAllRoutes' -count=1`
  - `test/scenarios/e2e/p0-102-auth-gated-interview-routes/scripts/{setup,trigger,verify,cleanup}.sh`
  - `make docs-check`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `git diff --check`
- 已建 Bug 记录：[BUG-0115](../bugs/BUG-0115.md)。

## 2 会话中的主要阻点/痛点

- Owner plan 原有 auth 口径没有把“protected route 必须在业务屏幕挂载前进入登录”写成显式 invariant。
  - **证据**：本次需要在 `frontend-shell` spec/plan/BDD 中追加 C-15、Phase 10 与 E2E.P0.102，才能覆盖 Home Recent、入口动作、直接路由访问和后端 middleware 的组合边界。
  - **影响**：修复前验证容易停留在单页面内部 auth gate，而不是 App route dispatch 的统一行为。
- 旧前端测试仍默认未登录也能先挂载业务屏幕。
  - **证据**：全量 Vitest 首轮剩余失败集中在 `ReportScreen.test.tsx` 与 `p0-005-app-shell-visual-system-smoke.test.tsx`，原因是同步断言业务页面，未等待认证解析或未按新 route guard 预期进入登录。
  - **影响**：需要额外修正测试 harness 与场景断言，才能让测试表达当前产品策略，而不是锁住旧行为。
- 未登录 Home 的公开/私有数据边界缺少场景级负向用例。
  - **证据**：本次新增 `HomeRecentMocks.test.tsx`、`HomeAuthGate.test.tsx`、`AppAuthDispatch.test.tsx` 和 E2E.P0.102 后，才同时断言“不显示、不请求、不泄漏 raw unauthorized”。
  - **影响**：仅靠后端 middleware pass 不能发现前端公开入口把 raw unauthorized 暴露给用户的问题。

## 3 根因归类

- `spec-plan`：前端 shell 认证边界缺少 route-level protected-route matrix，导致实现和测试可以各自保留页面内部 gate 口径。
- `spec-plan`：Home 的 Recent mock interviews 属于用户历史数据，但 owner 文档没有明确 signed-in only display 与 API short-circuit。
- `no repo change needed`：全量测试修正属于本次实现的必要同步，已在同一提交内完成；现阶段不需要额外改测试框架。

## 4 对流程资产的改进建议

- 在后续涉及前端认证边界的 plan 中，固定加入 protected-route matrix：route、是否公开、未登录渲染目标、pendingAction 类型、是否允许业务 API 调用。
  - **落点**：spec-plan
  - **优先级**：high
- 对 Home、TopBar、登录页等公开入口，新增“未登录零业务请求”负向 gate，避免公开页面展示用户历史或 backend raw error。
  - **落点**：spec-plan
  - **优先级**：high
- 若未来多次出现测试默认认证状态漂移，再考虑抽取 shared authenticated/unauthenticated App harness；本次不提前抽象。
  - **落点**：README / test helper
  - **优先级**：low

## 5 建议优先级与后续动作

- 下一轮最值得做的是对 `frontend-shell` 后续 auth/profile 相关计划补齐 protected-route matrix，尤其是 profile completion、logout、auth verify resume 这些容易与 pendingAction 交叉的路径。
- 可以延后处理的是共享测试 harness 抽象；当前失败点已用具体测试修正，尚未形成必须抽象的重复成本。
