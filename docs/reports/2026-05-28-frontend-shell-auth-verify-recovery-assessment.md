# Frontend Shell Auth Verify Recovery 交付复盘报告

> **日期**: 2026-05-28
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：修复 `frontend-shell/001-app-shell-auth-settings` L2 review 后续发现的两个 auth verify 回归风险：direct verified user commit 后的 initial auth probe skip 状态，以及 verify code 成功后 `/me` 刷新失败的恢复语义。
- 成功证据：
  - Red: `pnpm --filter @easyinterview/frontend test src/app/runtime/AppRuntimeProvider.test.tsx src/app/AppAuthDispatch.test.tsx` 先复现 requestOptions 变化后认证丢失、post-verify `/me` 失败被当作 code failure。
  - Green: `pnpm --filter @easyinterview/frontend test src/app/runtime/AppRuntimeProvider.test.tsx src/app/AppAuthDispatch.test.tsx`。
  - Green: `pnpm --filter @easyinterview/frontend typecheck`。
  - Green: `pnpm --filter @easyinterview/frontend test`，214 个 test files / 1320 个 tests 通过。
  - Green: `pnpm --filter @easyinterview/frontend build`，仅保留既有 Vite chunk size warning。
  - Green: `validate_context.py --target frontend`、`sync-doc-index --check`、`make docs-check`、`git diff --check`。
  - 收尾资产：新增 [BUG-0117](../bugs/BUG-0117.md)，并把 frontend-shell spec / plan / checklist / context / INDEX 原地更新到 Phase 11。

## 2 会话中的主要阻点/痛点

- public auth route 的初始 skip 状态不是一次性状态。
  - **证据**：review 指出 `skipInitialAuthProbe` 在 `auth_login` / `auth_verify` provider 生命周期内持续为 true；新增 red test 复现 direct user commit 后 requestOptions 变化会重置 auth。
  - **影响**：用户登录成功后，切换 TopBar 语言这类正常操作可能把 authenticated state 重新置为 unauthenticated。
- verify code 与 post-verify `/me` 刷新失败共用错误语义。
  - **证据**：新增 red test 让 `verifyAuthEmailChallenge` 成功但 `/api/v1/me` 返回 503，页面仍停留在 `auth_verify` 并显示 code failure。
  - **影响**：一次性 code 可能已经被消费，用户重试同一 code 会失败，形成不可恢复的 verify 页面卡住体验。
- 原 plan 已完成，但 auth 状态机边界仍缺少这两个负向 gate。
  - **证据**：`frontend-shell/001-app-shell-auth-settings` 既有 Phase 9/10 覆盖了主路径、资料补全和场景 wrapper，但没有覆盖 `requestOptions` dependency 变化与 post-verify profile refresh failure。
  - **影响**：review 后续发现需要原地重开 owner plan，并补充 Phase 11 文档和 focused tests 才能闭环。

## 3 根因归类

- auth probe skip 生命周期建模不足：
  - **类别**：spec-plan
  - 计划要求 public auth route 可跳过首次 `/me` 探测，但没有明确“跳过只能消费一次，direct auth commit 必须消耗该状态”。
- verify 成功与 profile refresh 恢复边界不清：
  - **类别**：spec-plan
  - 计划没有把一次性凭证消费和可重试 `/me` refresh 作为两个错误域分别验收。
- 现有测试缺少 auth effect dependency 回归：
  - **类别**：spec-plan
  - 测试覆盖主路径更充分，但未强制覆盖 `requestOptions` / language switch 这类正常 UI 操作触发 auth effect rerun 的边界。

## 4 对流程资产的改进建议

- auth 状态机相关 checklist 增加固定 gate：direct user commit 后改变 `requestOptions`，必须保持 authenticated state 且必要时执行真实 `/me` refresh。
  - **落点**：spec-plan
  - **优先级**：high
- verify/login/profile 计划增加错误域拆分约束：一次性 code 验证成功后，profile refresh 失败只能进入可恢复 auth/profile-loading 路径，不能被展示为 code invalid/expired。
  - **落点**：spec-plan
  - **优先级**：high
- 后续 L2 review auth 变更时，把 `skipInitialAuthProbe`、`authNonce`、`refreshAuth(user)`、post-verify `/me` 纳入固定负向搜索清单。
  - **落点**：skill
  - **优先级**：medium

## 5 建议优先级与后续动作

- 下一轮最值得做：在 auth/profile 后续 review 或 plan-code-review 清单中固化这四个 auth runtime 边界，避免只看 happy path 与现有 completed checklist。
- 可以延后：为 `frontend-shell/001-app-shell-auth-settings` 增加一个轻量场景级 auth verify recovery wrapper；本次已有 focused Vitest、全量 frontend test 和 build 证据，场景级 wrapper 可在下一次 auth runtime 计划中补齐。
