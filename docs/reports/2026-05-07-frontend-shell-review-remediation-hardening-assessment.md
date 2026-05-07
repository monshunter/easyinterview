# Frontend Shell Review Remediation Hardening 交付复盘报告

> **日期**: 2026-05-07
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：修复 `feat/frontend-shell-001-app-shell-auth-settings-0507` 相较 `refactor/backend-arch` 的四个 review findings，覆盖 frontend build entry、`/me` unknown scenario fail-loud、login -> register pendingAction 保留、raw `returnTo` query 恢复。
- 关联 Bug：[BUG-0019](../bugs/BUG-0019.md)。
- 成功证据：
  - `pnpm --filter @easyinterview/frontend test src/app/runtime/AppRuntimeProvider.test.tsx src/app/auth/AuthScreens.test.tsx` 通过。
  - `pnpm --filter @easyinterview/frontend typecheck` 通过。
  - `pnpm --filter @easyinterview/frontend build` 通过，Vite 输出 `dist/index.html` 与 bundle。
  - `pnpm --filter @easyinterview/frontend test` 通过（26 files / 130 tests）。
  - `make build` 通过。
  - `make test` 通过。
  - `git diff --check` 通过。

## 2 会话中的主要阻点/痛点

- Build gate 没有随 package script 真实化同步验证。
  - **证据**：`frontend/package.json` 已把 `build` 改为 `tsc --noEmit && vite build`，但缺少 `frontend/index.html`；本次 Red 阶段直接运行 build 复现 `Could not resolve entry module "index.html"`。
  - **影响**：根 `make build` 被前端构建阻断，属于跨 owner 的本地质量门禁回归。

- Runtime fail-loud 只覆盖了 `getRuntimeConfig`，没有覆盖 `/me`。
  - **证据**：原 `AppRuntimeProvider.test.tsx` 只对 runtime config unknown scenario 断言 error；`/me` unknown scenario 被 catch-all 收敛为 unauthenticated。
  - **影响**：mock fixture 拼写或 registry wiring 错误会被伪装成正常未登录状态。

- pendingAction 测试只覆盖 login 主路径，没有覆盖 auth 页间切换和 raw external fallback。
  - **证据**：原 E2E.P0.002 覆盖 login -> verify -> practice，但没有 login -> register -> verify，也没有 `returnTo=/practice?...` query 解析。
  - **影响**：用户选择注册或从外部 verify link 回来时，业务上下文仍可能丢失。

## 3 根因归类

- `spec-plan`：frontend-shell checklist 1.3 写了 unknown scenario fail loudly，但 focused test 只覆盖一个 runtime probe，缺少“每个 bootstrap operation 都要有 fail-loud 断言”的明确 gate。
- `spec-plan`：Phase 3.2 pendingAction 恢复 gate 只描述登录成功恢复主路径，没有枚举 auth 页间切换与 raw `returnTo` fallback。
- `README`：frontend handoff 说明了 Vite toolchain，但没有明确 package build 从占位切换为真实 bundler 时必须同时落地 HTML entry 并运行根 `make build`。

## 4 对流程资产的改进建议

- 在 frontend-shell 后续 owner plan 中补一条 smoke gate：package `build` script 真实化时必须运行 `pnpm --filter @easyinterview/frontend build` 和根 `make build`。
  - **落点**：spec-plan
  - **优先级**：high

- 将 runtime bootstrap fail-loud gate 扩展为 operation matrix：`getRuntimeConfig`、`getMe`、auth generated operations 的 unknown fixture / required param 行为分别覆盖。
  - **落点**：spec-plan 或 frontend README
  - **优先级**：medium

- pendingAction 测试模板增加三类入口：直接 login、login/register 切换、raw `returnTo` fallback。
  - **落点**：frontend README
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：下一次修订 `frontend-shell` 或 D2-D6 frontend owner plan 时，把真实 build smoke 和 runtime operation matrix 写入 checklist 验证项。
- 可延后：在 `frontend/README.md` 增补 pendingAction 测试模板，作为后续 owner 接入需要登录动作时的统一检查清单。
