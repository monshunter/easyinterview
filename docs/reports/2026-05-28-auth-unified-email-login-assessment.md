# Auth Unified Email Login 交付复盘报告

> **日期**: 2026-05-28
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付把 auth 从“注册页 + 登录页”调整为单一邮箱验证码登录入口：邮箱是唯一账号标识，displayName 不是唯一字段，也不参与账号去重。
- 新邮箱首次完成验证码验证后创建资料未补全账号，`GET /me.profileCompletionRequired=true` 强制进入 `auth_profile_setup`；完成 displayName + 条款确认后 `PATCH /me` 返回 `profileCompletionRequired=false`，后续同邮箱登录直接进入正常登录态。
- 覆盖范围包括 backend auth store/service/handler/migration、OpenAPI/generated client/fixtures、frontend auth screens/routes/runtime guard/dev mock/i18n、ui-design 静态原型、P0.101 real frontend/backend/Mailpit 场景，以及 backend-auth/frontend-shell owner spec/plan/checklist/BDD。
- 成功证据：
  - `cd backend && go test ./...`
  - `pnpm --filter @easyinterview/frontend typecheck`
  - `pnpm --filter @easyinterview/frontend test`
  - `pnpm --filter @easyinterview/frontend exec vitest run src/app/AppAuthDispatch.test.tsx -t "redirects unauthenticated auth_profile_setup visits"`
  - `node --test ui-design/ui-design-contract.test.mjs`
  - `python3 scripts/lint/openapi_inventory.py openapi/openapi.yaml`
  - `python3 scripts/lint/validate_fixtures.py --repo-root .`
  - `python3 -m unittest scripts.lint.openapi_inventory_test scripts.lint.validate_fixtures_test scripts.lint.validate_fixtures_cli_test`
  - `make lint-mock-contract`
  - `make openapi-diff`
  - `make lint-config`
  - `python3 scripts/lint/migrations_lint.py --repo-root .`
  - `bash -lc 'set -a; . deploy/dev-stack/.env; set +a; make migrate-status'`
  - `bash test/scenarios/env-redeploy.sh all`
  - `bash test/scenarios/env-verify.sh`
  - `bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/cleanup.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/setup.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/trigger.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/verify.sh && bash test/scenarios/e2e/p0-101-auth-email-code-login-register/scripts/cleanup.sh`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `make docs-check`
  - `git diff --check`
- `/bug-report` 评估结论：本轮主体是用户确认后的产品语义变更；收口中发现的未登录直开 `auth_profile_setup` 守卫缺口已在同一交付内用 focused red/green test 修复，未形成逃逸缺陷或需要复用诊断模式，未单独创建 BUG 记录。

## 2 会话中的主要阻点/痛点

- Auth 语义变更横跨两个 completed owner plan。
  - **证据**：同一变更必须同时重开并收口 `backend-auth/001-passwordless-session-bootstrap` Phase 8 与 `frontend-shell/001-app-shell-auth-settings` Phase 9，并同步 OpenAPI、fixtures、generated artifacts、ui-design、正式前端和 P0.101 场景。
  - **影响**：只改 UI 或只改后端都会产生旧注册入口、旧 purpose 字段或资料补全 gate 漏洞，必须按 owner spec 反向闭环。

- 本地场景环境的迁移前置条件一开始只在运行时暴露。
  - **证据**：P0.101 首轮 real-stack 验证发现 profile completion migration column 尚未应用；随后补强 `scripts/setup.sh` 的 migration column preflight，并通过 `env-setup.sh --with-migrations` 与 `make migrate-status` 确认 `version=13 dirty=false`。
  - **影响**：真实场景可以在业务步骤中才失败，反馈成本高于 setup 阶段的明确失败。

- `make codegen-check` 在 dirty tree 下容易被误读。
  - **证据**：该 target 最终依赖 `git diff --exit-code`；当前任务存在未提交的 OpenAPI/generated 预期 diff 时，直接运行会被 clean-tree 检查拦住。通过 `make codegen-openapi` 前后 diff 比对和临时 index 等价检查确认生成器无漂移。
  - **影响**：收尾证据需要明确区分“生成器不幂等”和“当前变更尚未提交”。

- 资料补全 route guard 的未登录边缘路径在收口审查才被补测发现。
  - **证据**：新增 `AppAuthDispatch.test.tsx` focused 用例先红灯，暴露未登录直开 `auth_profile_setup` 会短暂停在资料补全表单；随后前端 runtime guard 将该路径回退到 `auth_login`，focused test、全量 Vitest 和 P0.101 real scenario 均通过。
  - **影响**：`profileCompletionRequired` 主路径已有覆盖，但“未登录访问资料补全页”这种反向边界需要与强制补全主路径一起固化。

## 3 根因归类

- completed plan 跨 owner 重开属于 `spec/plan`：auth identity、profile completion 与 frontend route guard 是同一个用户流程，但当前分布在 backend-auth 与 frontend-shell 两个 owner plan，需要显式双 owner gate。
- migration 前置条件属于 `test` / `README`：real-stack 场景 setup 应优先检查所需 schema column，而不是让浏览器流程暴露数据库结构缺口。
- codegen-check dirty-tree 语义属于 `README` / Make target 帮助文档：本地实施中应提供 clean-tree drift check 与 dirty-tree idempotence check 的区分。
- auth profile setup 的反向访问边界属于 `spec/plan`：同一个 gate 需要同时覆盖 authenticated incomplete 强制进入、authenticated complete 不重复填写、unauthenticated 不能停留在资料补全表单。

## 4 对流程资产的改进建议

- 在 auth 相关 owner plan 的 operation matrix 中增加“profile completion owner map”：列出 `UserContext.profileCompletionRequired`、`PATCH /me`、前端 route guard、dev mock、P0 场景和 migration column 的对应关系。
  - **落点**：spec-plan
  - **优先级**：high

- 为需要新 migration 的 real-stack scenario 建立通用 setup preflight 模板：列出必须存在的 table/column/version，缺失时直接提示运行共享 migration setup。
  - **落点**：test/scenarios README 或 scenario-env skill
  - **优先级**：medium

- 在 codegen 文档或 Make target 帮助中补充 dirty-tree 说明：`make codegen-check` 是 clean-tree gate；开发中验证幂等可使用“记录 diff -> codegen -> 比对 diff”或临时 index 方式。
  - **落点**：README / Make target help
  - **优先级**：low

- 在 frontend-shell auth checklist 中把 `auth_profile_setup` 的三态 route guard 固化为同一个验收点：未登录回登录、已登录未补全停留、已补全恢复 pendingAction / Home。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：下一轮对 auth/account 相关计划做一次 `/plan-code-review backend-auth/001-passwordless-session-bootstrap --fix`，重点确认 completed plan 中是否仍有旧 `purpose=signup/login`、独立注册页、duplicate-register 409 或 displayName-before-verify 口径残留在 active truth source。
- 中优先级：把 frontend-shell auth guard 的三态测试要求固化到计划模板或下一次计划审查 gate，并把 P0.101 的 migration column preflight 抽成可复用 helper，避免后续带新 migration 的 real-stack 场景重复写 schema 检查。
- 低优先级：补充 codegen-check dirty-tree 使用说明，降低本地开发阶段对最终 `git diff --exit-code` 的误判。
