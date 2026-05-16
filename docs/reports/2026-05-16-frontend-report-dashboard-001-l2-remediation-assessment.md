# Frontend Report Dashboard 001 L2 Remediation 交付复盘报告

> **日期**: 2026-05-16
> **审查人**: Codex

## 1 复盘范围与成功证据

本次复盘覆盖 `frontend-report-dashboard/plans/001-report-screen-and-generating-handoff` 的最新 `main` 合并、L2 code review 修复、测试验证和文档收口。`git fetch origin main` 后 `git merge --no-edit main` 结果为 `Already up to date.`，没有冲突和 merge commit。

成功证据：

- `pnpm --filter @easyinterview/frontend test`：最终全量重跑 171 个 test files / 985 tests passed。
- `pnpm --filter @easyinterview/frontend typecheck`
- `pnpm --filter @easyinterview/frontend build`
- `pnpm --filter @easyinterview/frontend test:pixel-parity tests/pixel-parity/generating.spec.ts tests/pixel-parity/report.spec.ts`：14 passed。
- `make codegen-check`
- `make validate-fixtures`
- `python3 scripts/lint/frontend_report_dashboard_legacy.py --repo-root . --phase all`
- `python3 -m pytest scripts/lint/frontend_report_dashboard_legacy_test.py -q`：3 passed。
- `E2E.P0.056` / `057` / `058` / `059` setup → trigger → verify → cleanup 全部通过。
- `E2E.P0.044` / `045` / `046` / `047` setup → trigger → verify → cleanup 全部通过。
- `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`、`make docs-check`、`git diff --check` 均通过。

## 2 会话中的主要阻点/痛点

- Pixel parity 的 hash route bootstrap 漂移会让静态预览落到 Home，而不是真实 `generating` / `report` 路由。
  - **证据**：补齐 `parseInitialRouteHash` 后，Playwright 才真实进入 report/generating，并暴露移动端 report overflow。
  - **影响**：旧 gate 可能误判为覆盖了目标页面，实际没有审到目标 DOM。
- UI token negative gate 漏掉了 prototype short CSS token。
  - **证据**：L2 review 发现 implementation 中仍有 `--ei-bg` / `--ei-ink` / `--ei-accent` 等短 token；已替换为正式 token 并扩展 lint。
  - **影响**：样式来源可追溯性不完整，容易绕过 `ui-design` → 正式前端的 token 约束。
- 移动端 no-overflow 断言依赖真实路由后才暴露。
  - **证据**：hash bootstrap 修复后，mobile 390px report `scrollWidth` 超出视口；已通过 responsive grid、`minWidth: 0` 和 `overflowWrap` 修复。
  - **影响**：单独看 jsdom / unit tests 无法证明移动端报告页面布局闭合。
- checklist 与 INDEX 生命周期未同步到最新验证事实。
  - **证据**：主 checklist 已完成，但 `test-checklist.md` / `bdd-checklist.md` 仍未勾选，plan header 与 plans INDEX 曾存在日期/状态漂移。
  - **影响**：后续 owner 会误判 plan 是否仍待收口。

## 3 根因归类

- Pixel parity hash bootstrap：**类别** `spec-plan`。当前 gate 需要把“真实目标 route 是否启动”作为前置断言，而不是只依赖测试 URL 形态。
- Prototype short token 漏检：**类别** `spec-plan`。负向 lint 已覆盖旧 route/旧模块字面量，但未把 UI token provenance 纳入同一 gate。
- Mobile overflow：**类别** `spec-plan`。parity gate 需要保留 `documentElement.scrollWidth <= innerWidth` 这类可执行几何断言。
- 文档生命周期漂移：**类别** `spec-plan`。completed plan 的主 checklist、test checklist、BDD checklist 和 `plans/INDEX.md` 必须同批更新并跑 sync-doc-index。

## 4 对流程资产的改进建议

- 在 frontend-report-dashboard 后续 UI plan 的 pixel parity gate 中保留 hash route bootstrap / route identity 断言。
  - **落点**：spec-plan
  - **优先级**：high
- 将 prototype short CSS token 纳入 UI legacy-negative / provenance lint 的固定禁止项。
  - **落点**：spec-plan
  - **优先级**：high
- 对移动端用户可见 dashboard 固定执行 `scrollWidth <= innerWidth` 与关键容器 bounding box no-overlap。
  - **落点**：spec-plan
  - **优先级**：high
- completed 状态推进时，把 `test-checklist.md` / `bdd-checklist.md` 与 `plans/INDEX.md` 作为同一收口单元，随后运行 `sync-doc-index --check`。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

下一步建议先执行 `/work-journal`，将本次 L2 review 修复、docs/checklist 收口和复盘报告打包成一个英文 ASCII commit。之后再进入 push / PR 回复阶段，PR 回复应突出：main merge 无冲突、L2 修复项、pixel/mobile/legacy-negative gate 以及 P0.044-047 / P0.056-059 场景证据。
