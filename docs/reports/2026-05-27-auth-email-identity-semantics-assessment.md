# Auth Email Identity Semantics 交付复盘报告

> **日期**: 2026-05-27
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付修复 auth 注册/登录身份语义：注册邮箱就是后续登录邮箱，email 是唯一账号标识，displayName 不唯一且不参与账号去重。
- 覆盖范围包括 OpenAPI/generated client、backend auth service/store/handler、frontend auth screens/TopBar、Mailpit email-code 文案、P0.101 real frontend/backend/Mailpit 场景、backend-auth/frontend-shell owner spec/plan/checklist/BDD。
- 成功证据：
  - `cd backend && go test ./internal/auth -count=1`
  - `cd backend && go test ./cmd/api -run 'Test(AuthEmail|BuildAuth|LocalDevCORS|BuildAPIHandlerMountsAuthRoutesAndSessionAwareRuntimeConfig)' -count=1`
  - `cd backend && go test ./...`
  - `pnpm --filter @easyinterview/frontend test src/app/auth/AuthScreens.test.tsx src/app/AppAuthDispatch.test.tsx src/app/routeUrl.test.ts src/app/topbar/TopBar.test.tsx`
  - `pnpm --filter @easyinterview/frontend build`
  - `bash test/scenarios/env-redeploy.sh all && bash test/scenarios/env-verify.sh`
  - `bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/setup.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/trigger.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/verify.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/cleanup.sh`
  - `make docs-check`, `make lint-config`, `make lint-openapi`, `make lint-mock-contract`, `git diff --check`
- 关联 Bug: [BUG-0114](../bugs/BUG-0114.md)。

## 2 会话中的主要阻点/痛点

- 身份不变量没有一开始成为 owner gate。
  - **证据**：用户明确指出“登录邮箱和注册邮箱是同一个邮箱”，随后需要同时修订 OpenAPI、backend-auth、frontend-shell、P0.101 场景和 UI 文案。
  - **影响**：最初实现容易把 displayName 注册持久化当成主问题，而真正必须收紧的是 email uniqueness 与 signup/login purpose 分流。

- P0.101 第一次真实运行暴露了重复注册路径与 rate-limit 的语义冲突。
  - **证据**：重复注册同一邮箱作为第三次 challenge 请求时，服务端按 rate-limit accepted 路径处理，Mailpit 没有新邮件，场景等待第三封邮件失败。
  - **影响**：需要把重复 signup 的 email uniqueness 检查前移到 start 阶段，确保 409、不创建 challenge、不发 code。

- 场景失败统计一开始把预期 auth 状态码当成失败。
  - **证据**：P0.101 后续运行里，浏览器对预期 `401 /me`、`401 /targets`、`409 /auth/email/start` 产生 network warning / response event，场景误计为 console/http failure。
  - **影响**：业务路径已通过但 runner false-negative，需要把 expected auth lifecycle failure 从“非预期失败”统计中过滤。

- `make codegen-check` 在 dirty tree 中的语义容易误读。
  - **证据**：该 target 会运行 codegen 和 lint，然后对 OpenAPI/source/generated 路径执行 `git diff --exit-code`；本次未提交 contract 变更存在时最终失败，但 `make codegen-openapi` 与 OpenAPI lint 已通过。
  - **影响**：收尾报告需要明确区分“生成器不幂等”与“当前任务尚未提交导致 drift gate 失败”。

## 3 根因归类

- 身份不变量缺失属于 `spec/plan`：auth owner plan 需要显式声明 email 是账号身份，displayName 是 profile 字段。
- 重复 signup 与 rate-limit 冲突属于 `spec/plan` + `backend`：服务端必须先判断业务 identity 冲突，再进入 challenge/rate-limit 行为。
- 场景 false-negative 属于 `test` README / scenario convention：真实 auth 场景需要区分 expected auth-state HTTP responses 与 unexpected failures。
- codegen-check 语义属于 `README` / Make target 帮助文档：dirty-tree 下失败不等价于生成器漂移。

## 4 对流程资产的改进建议

- 在 auth 相关 plan/checklist 的 operation matrix 中强制列出 identity 字段、profile 字段、唯一性字段和重复提交行为。
  - **落点**：spec-plan
  - **优先级**：high

- 为 real-mode auth scenario 增加 runner 约定：expected `401`/`409` 必须显式白名单并保留业务断言，`consoleErrors=0` 只统计非预期错误。
  - **落点**：test/scenarios README 或 scenario-run skill
  - **优先级**：medium

- 在 codegen-check 的本地使用说明中补一句：该 target 适合 clean-tree drift check；未提交 contract 改动时最终 `git diff --exit-code` 失败是预期信号。
  - **落点**：README / Make target help
  - **优先级**：low

## 5 建议优先级与后续动作

- 下一轮最高价值动作：对 auth/account 相关 owner spec 做一次 `/plan-code-review backend-auth/001-passwordless-session-bootstrap --fix`，重点反查身份字段与 profile 字段是否还在其他 API、fixtures、场景或 docs 中混用。
- 可延后动作：把 real-mode auth scenario 的 expected auth-state HTTP 过滤规则抽成测试 helper，减少后续 P0 auth 场景重复维护。
- 可延后动作：补充 codegen-check clean-tree 使用说明，降低本地未提交开发过程中的误判成本。
