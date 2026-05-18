# Frontend Shell URL Routing Popstate Privacy 交付复盘报告

> **日期**: 2026-05-18
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`plan-code-review frontend-shell/004-url-addressable-routing --fix` 的 L2 review remediation，聚焦 C-12 / E2E.P0.089 URL privacy redline 在 Browser History `popstate` 恢复路径中的缺口。
- 修复：`routeStore` 的 `popstate` handler 解析安全 route 后，若当前 URL/hash/history state 不等于 canonical safe URL，则 `replaceState(null, "", canonicalUrl)` 清理当前历史项。
- 原地修订：`frontend-shell/004-url-addressable-routing` plan/checklist/BDD 文档提升到 v1.2，E2E.P0.089 wrapper 由 3 tests 扩为 4 tests；新增 [BUG-0079](../bugs/BUG-0079.md)。
- 通过证据：
  - `pnpm --filter @easyinterview/frontend test src/app/AppRoutingPrivacy.test.tsx src/app/scenarios/p0-089-url-routing-auth-privacy.test.tsx` -> 10 tests passed。
  - `pnpm --filter @easyinterview/frontend test src/app/AppRoutingHistory.test.tsx src/app/routeStore.test.ts src/app/routeUrl.test.ts src/app/bootstrapRoute.test.ts` -> 53 tests passed。
  - E2E.P0.088 / E2E.P0.089 / E2E.P0.090 scenario wrapper `setup -> trigger -> verify -> cleanup` 全部通过。
  - `pnpm --filter @easyinterview/frontend build`、`make docs-check`、context validator、`git diff --check` 全部通过。

## 2 会话中的主要阻点/痛点

- Completed BDD gate 未覆盖 hostile history entry。
  - **证据**：原 E2E.P0.089 只验证 auth restore、jd_match restore 和 hostile `/auth/login` direct-open；新增 hostile popstate test 前，`AppRoutingPrivacy.test.tsx` red phase 失败，URL 仍含 `rawText` 与 hash marker。
  - **影响**：如果只看 completed checklist 和 3-test scenario count，会错过地址栏层面的隐私泄露风险。
- 类型检查在后置 build 才发现测试对象声明不够精确。
  - **证据**：首次 `pnpm --filter @easyinterview/frontend build` 因 `RAW_MARKERS.rawText` 在 `noUncheckedIndexedAccess` 下可能为 `undefined` 失败。
  - **影响**：focused Vitest 已绿但 TypeScript build 未绿，说明闭环必须包含 build/typecheck。

## 3 根因归类

- URL privacy 的覆盖矩阵写得不够具体。
  - **类别**：spec-plan
  - **说明**：plan 原文提到 URL/history/state privacy，但没有显式拆成 direct-open、programmatic navigation、auth restore、browser back/forward recovery 四个入口。
- Scenario wrapper pass count 被当作完成证据时缺少风险维度解释。
  - **类别**：spec-plan
  - **说明**：E2E.P0.089 的 `Tests 3 passed (3)` 证明既有三条路径执行过，但不证明所有 URL privacy 入口都已覆盖。
- Build/typecheck gate 是必要的后置闭环。
  - **类别**：no repo change needed
  - **说明**：本次流程已经通过 build gate 捕捉并修复类型问题；不需要新增治理规则。

## 4 对流程资产的改进建议

- URL / history privacy 类计划在原 plan gate 中显式列出入口矩阵：direct-open、programmatic navigation、auth restore、popstate/back-forward。
  - **落点**：spec-plan
  - **优先级**：high
  - **状态**：本次已在 plan 004 v1.2 和 E2E.P0.089 中固化。
- 后续 `/plan-code-review` 审查 URL-addressable workstream 时，把 "parser drops unsafe params" 和 "browser address bar/history entry is rewritten" 分开验收。
  - **落点**：skill
  - **优先级**：medium
- 对 scenario wrapper 证据，报告 pass count 时同时写明覆盖的行为入口，避免 "N tests passed" 被误读为全风险矩阵覆盖。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 优先保留本次 plan 004 v1.2 的入口矩阵作为后续 URL/privacy review 样例。
- 下一轮若继续强化流程资产，可在 `/plan-code-review` skill 的 URL/privacy review 提示中加入 `popstate` canonical rewrite 检查；这属于 workflow hardening，不阻塞本次交付。
- 不建议扩大到全量前端测试作为本次必需 gate；本次改动的可执行边界已经由 focused route/privacy tests、三个 BDD wrapper、build、docs-check 和 context validator 覆盖。
