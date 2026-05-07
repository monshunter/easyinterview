# Frontend Shell L2 Remediation 交付复盘报告

> **日期**: 2026-05-07
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`frontend-shell/001-app-shell-auth-settings` L2 code review remediation，覆盖 `voice` route alias 删除、`auth_verify` token query wire、pendingAction auth-only params 隔离，以及 owner plan/checklist v1.2 原地修订。
- 关联 Bug：[BUG-0018](../bugs/BUG-0018.md)。
- 成功证据：
  - `pnpm --filter @easyinterview/frontend test src/app/normalizeRoute.test.ts src/app/AppNormalize.test.tsx src/app/scope.test.ts` 通过。
  - `pnpm --filter @easyinterview/frontend test src/app/AppAuthDispatch.test.tsx src/app/auth/pendingAction.test.ts src/app/auth/AppPendingAction.test.tsx src/app/auth/authContractGate.test.ts` 通过。
  - `pnpm --filter @easyinterview/frontend typecheck` 通过。
  - `pnpm --filter @easyinterview/frontend test` 通过（26 files / 127 tests）。
  - `E2E.P0.001` 与 `E2E.P0.002` 场景均完成 `setup -> trigger -> verify -> cleanup`，全部 0 退出。
  - context validator、`make docs-check`、`git diff --check` 均通过。

## 2 会话中的主要阻点/痛点

- 历史 checklist 口径与当前产品 / UI 真理源冲突。
  - **证据**：checklist 1.2 仍把 `voice` 放入旧 route 映射集合；`product-scope` 与 `removed-modules-and-scope` 已明确删除 `voice` route alias。
  - **影响**：若只按历史 checklist 验收，会把已被删除的 route alias 当成通过证据。

- Generated client wire gate 覆盖了 operationId，但没有覆盖必填 query。
  - **证据**：原 `AppAuthDispatch.test.tsx` 只断言 login start 命中 `POST /auth/email/start`；`verifyAuthEmailChallenge` 在 App 层未传 token 仍可通过全量测试。
  - **影响**：真实后端调用会缺少 OpenAPI 必填 `token` query，认证验证链路不可用。

- pendingAction 恢复测试只断言业务上下文保留，没有断言认证临时字段剥离。
  - **证据**：修复前 `email` 随 `auth_verify` params 进入 `practice` route params，测试未覆盖负向字段。
  - **影响**：认证页临时信息进入业务 route 状态，增加隐私与后续 owner 误用风险。

## 3 根因归类

- `spec-plan`：已完成 checklist 没有随着产品 / UI 真理源的 `voice` alias 删除及时原地修订，导致历史验收口径滞后。
- `spec-plan`：Phase 3.3 的 auth contract gate 描述了“不新增错误 auth 形态”，但没有明确 generated client 必填 query/header/path 参数也必须被断言。
- `no repo change needed`：本次通过追加 focused negative tests 已封住 pendingAction auth-only params 泄漏；暂不需要改治理规则。

## 4 对流程资产的改进建议

- 在未来 L2 review 的 `Deep Evidence` 中固定增加“当前 UI / 产品真理源覆盖旧 checklist 口径”的差异扫描。
  - **落点**：`plan-code-review` skill
  - **优先级**：high

- 对 generated client 调用增加测试检查提示：不仅检查 operationId / path，还要检查 OpenAPI 必填 query、path、header 参数是否由调用层传入。
  - **落点**：`plan-code-review` skill 或 frontend README 测试约定
  - **优先级**：medium

- pendingAction / route-state 类测试采用双向断言：必须保留的业务上下文 + 必须剥离的认证/临时字段。
  - **落点**：frontend README 或后续 frontend-shell owner plan gate
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：在下一次修改 `plan-code-review` skill 时，把“历史 checklist 与当前 truth source 冲突时，以当前 truth source 反向审计并产出 remediation item”写成固定审查动作。
- 可延后：为 frontend README 增补 generated client 测试约定，作为 D2-D6 owner 接入 API 调用前的 handoff 规则。
