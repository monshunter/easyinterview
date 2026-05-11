# Frontend Shell Workspace Pixel Gate 交付复盘报告

> **日期**: 2026-05-10
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付修复 `frontend-shell/003-ui-design-pixel-parity-gate` 的 workspace / screenshot remediation：完整 `test:pixel-parity` 不再依赖 ignored screenshot baseline，也不再通过 Home recent card 的 `resume-unbound` 前提进入 hydrated workspace。
- 成功证据：Red 聚焦运行 `screenshot.spec.ts + workspace.spec.ts` 复现 12 failed / 24 passed；修复后同一聚焦运行 36 passed；完整 `pnpm --filter @easyinterview/frontend test:pixel-parity` 110 passed；E2E.P0.006 setup→trigger→verify→cleanup PASS；`pnpm --filter @easyinterview/frontend test` 108 files / 681 tests PASS；`typecheck`、`build`、`make docs-check`、context validator、`git diff --check` 均通过。
- 原 owner 已原地修订并恢复 completed：`docs/spec/frontend-shell/plans/003-ui-design-pixel-parity-gate/` v1.2 / checklist v1.3，BDD 口径同步到 8 spec / 110 tests。

## 2 会话中的主要阻点/痛点

- P0.006 scenario contract 落后于真实 test suite。
  - **证据**：Playwright `--list` 显示 8 spec / 110 tests，但 P0.006 README / verify 仍写 7 spec / 68 tests，早期 checklist evidence 仍是 4 spec / 48 tests。
  - **影响**：scenario verify 不能证明 workspace spec 已被执行，也容易让 full gate 看起来“应该依赖旧 hydrated 前提”。
- Screenshot baseline 口径在 README 与测试实现之间分裂。
  - **证据**：`frontend/README.md` 已说明 clean checkout 不能依赖 ignored baseline，但 `screenshot.spec.ts` / `workspace.spec.ts` 仍直接使用 `toHaveScreenshot` 作为常规 PASS gate。
  - **影响**：新机器或清理 snapshot 后首跑失败，且失败会自动写 ignored PNG，造成诊断噪音。
- Workspace full-state 测试复用了产品入口而不是测试态入口。
  - **证据**：Home recent card 按业务语义携带 `resume-unbound`；workspace hooks 正确通过 `normalizeServerBoundId` 过滤 synthetic id，pixel test 却等待 full workspace anchor。
  - **影响**：测试前提与产品语义冲突，容易诱导错误修复 `interviewContextFromTargetJob`。

## 3 根因归类

- P0.006 verify 脚本 hard-code 用例数量和 spec marker，缺少随 Playwright suite 扩展而同步的 owner gate。
  - **类别**：spec-plan / README
- Screenshot baseline 是否可作为 hard gate 的条件没有固化为可执行 lint，只停留在 README 文本。
  - **类别**：spec-plan / test
- Full-state pixel harness 缺少稳定 route bootstrap 约定，导致测试借用普通用户入口。
  - **类别**：README / spec-plan

## 4 对流程资产的改进建议

- 为 P0.006 或 pixel parity owner 增加一个 suite manifest / verify helper，自动反查 `playwright test --list` 的 spec marker 与 total count，避免 README、verify、真实 suite 三处漂移。
  - **落点**：test/scenarios README 或 pixel owner plan
  - **优先级**：high
- 增加静态检查：默认 `frontend/tests/pixel-parity/*.spec.ts` 不允许使用 `toHaveScreenshot`，除非同一 plan 明确声明 baseline artifact 来源。
  - **落点**：frontend test tooling / docs/spec/frontend-shell/plans/003
  - **优先级**：medium
- 抽出 pixel harness helper（如 `installInitialRoute(page, route)`），统一 server-bound route bootstrap，避免各 spec 自行写全局变量。
  - **落点**：frontend/tests/pixel-parity helper README 或 test utility
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高价值下一步：给 P0.006 增加自动 suite manifest 校验，减少未来新增 pixel spec 后再次出现 `68 passed` / `110 passed` 这种脚本漂移。
- 次优先级：将 screenshot hard-gate 条件转成静态 lint，直接阻止 ignored baseline 回流到默认 clean checkout gate。
