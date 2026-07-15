# Report Context Grid Alignment 交付复盘报告

> **日期**: 2026-07-15
> **审查人**: Codex

## 1 复盘范围与成功证据

- 交付范围：在 `frontend-report-dashboard/001-report-screen-and-generating-handoff` 原 owner 内，将 ready report Context Strip 调整为 target/round/resume/interview record 四个同级子项；为 frozen resume 提供 canonical URL；把两个 desktop detail row 固化为同行等高，同时保持 390 mobile 自然高度单列。
- 行为证据：Chrome 使用真实 backend 持久化的 synthetic ready report 完成 exact 1440x1200 与 390x844 full-page 截图、几何测量、resume navigation/back、report conversation navigation/back 和 DOM privacy audit；固定两图与脱敏 manifest 位于 `.test-output/acceptance/report-context-grid/`。
- 代码证据：前端 126 files / 1003 tests、P0.099 evidence tests 8 tests、OPENAPI-001 focused tests 4 tests 全部通过。
- 完整门禁：根 `make test` 通过 559 Python tests / 4481 subtests、Go 全包与前端全量；`make lint`、`make build`、frontend typecheck、`make docs-check`、Header/INDEX 与 `git diff --check` 通过。

## 2 会话中的主要阻点/痛点

- Chrome viewport capability 初次设置后页面仍保持桌面 `innerWidth`。
  - **证据**：设置 390x844 后首次审计仍返回约 1447px 宽；在实际缩窄 Chrome 窗口后，同一 capability 才稳定返回 exact 390x844，随后 exact 1440x1200 也可复现。
  - **影响**：产生一次无效“移动端”截图，必须丢弃并重新捕获；若只相信设置调用而不读 `innerWidth`，会形成假验收。
- 原 Phase 12 把任何 ready 布局微调都绑定到 P0.099 的三状态、双语言、三套 provider 资源完整矩阵。
  - **证据**：本次实际变更仅影响 ready Context Strip 与 detail grid；generating/failure、provider 输出和语言状态未改变，而用户明确要求奥卡姆剃刀与 Chrome 截图闭环。
  - **影响**：若机械执行完整矩阵，会引入与变更无关的 provider 资源生成、状态捕获和 cookie 证据成本。当前计划已改为 focused real-backend ready acceptance，同时更新 P0.099 后续完整矩阵合同，不伪造新 E2E PASS。
- 根回归暴露 OPENAPI-001 refreeze 后的测试事实源漂移。
  - **证据**：首次 `make test` 3 failures；[BUG-0177](../bugs/BUG-0177.md) 证明 helper 读取可前进的 `main` baseline，而 preserved audit 已记录正确 pre-refreeze source。
  - **影响**：UI 交付被无关根门禁阻塞，但也避免带着已坏的 repo gate 收口。

## 3 根因归类

- 视口设置调用缺少“设置后必须读取实际 `window.innerWidth/innerHeight`”的显式验收约束。
  - **类别**：skill
- Phase 12 的 E2E gate 没有按 changed behavior 划分 focused acceptance 与完整状态矩阵。
  - **类别**：spec/plan
- OPENAPI-001 test helper 把 `main` 当成不可变历史快照。
  - **类别**：test README / shared test helper；具体缺陷已由 BUG-0177 修复。

## 4 对流程资产的改进建议

- 浏览器响应式验收应把实际 `innerWidth/innerHeight`、`scrollWidth` 与截图像素宽度列为 capture 前置 gate，设置调用成功本身不算证据。
  - **落点**：Chrome/browser skill 的 responsive capture guidance
  - **优先级**：high
- 用户可见但状态范围窄的 UI 修订，owner plan 应先列 changed states；只对 changed states 做当前 real-browser acceptance，同时更新完整 E2E owner 合同，避免重复无关 provider 资源。
  - **落点**：spec/plan（本次 owner 已原地修订）
  - **优先级**：medium
- OpenAPI decision tests 可抽取一个读取 preserved audit `baselineSource` 的共享 helper，减少后续 refreeze decision 重复实现与分支快照误用。
  - **落点**：scripts/lint test helper
  - **优先级**：medium

## 5 建议优先级与后续动作

- 下一轮最高价值动作：在浏览器响应式验收 skill 中固化“设置视口后读取实际尺寸再截图”的 fail-closed 规则。
- 后续 OpenAPI gate 维护时，把各 decision 的 preserved baseline 读取收敛为共享 helper，并用 post-refreeze regression 覆盖。
- 本次报告 UI owner 已完成 focused acceptance 与 P0.099 contract handoff，无需再为未变化状态补造当前 provider 资源；下一次明确执行 P0.099 全矩阵时会自动承接新增的四项上下文与响应式对齐 checks。
