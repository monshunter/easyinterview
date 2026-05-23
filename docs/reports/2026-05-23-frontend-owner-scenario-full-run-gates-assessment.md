# Frontend Owner Scenario Full-Run Gates 交付复盘报告

> **日期**: 2026-05-23
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：执行并修复 frontend owner real-backend gate 场景全量复跑，覆盖 P0.018-P0.021、P0.044-P0.047、P0.056-P0.059、P0.065-P0.069、P0.081-P0.087。
- 成功证据：`scenario-run-full-20260522T172858Z` 完整 24 场景串行复跑全部 PASS，summary 位于 `.test-output/runs/scenario-run-full-20260522T172858Z/e2e/summary.json`，`failures=0`。
- 修复证据：`scenario-run-fixcheck-20260522T172739Z` 对 P0.065-P0.069 + P0.087 受影响场景复跑全部 PASS。
- 关联 Bug：[BUG-0090](../bugs/BUG-0090.md)。

## 2 会话中的主要阻点/痛点

- P0.066 / P0.067 初始失败并非目标业务流失败，而是 wrapper 把 scoped debrief gate 扩大为全 frontend Vitest。
  - **证据**：第一轮 run `scenario-run-20260522T172049Z` 中 P0.066 / P0.067 trigger 失败，日志显示 213 个 frontend test files 参与，失败点落在无关 `DebriefPickerRegression.test.tsx`。
  - **影响**：全量场景证据一开始不可用，且 failure symptom 容易被误判为 debrief 功能回归。
- P0.087 Playwright pixel parity 使用的 hash route harness 与当前 routeStore 优先级不一致。
  - **证据**：第一轮 P0.087 desktop/mobile 均等待 `resume-branch-flow-form` / `resume-rewrites-tab` 超时；`gotoHashRoute` 生成 `/?pixelRoute=N#route=...`，routeStore 会先解析 canonical search。
  - **影响**：pixel parity 没有进入目标页面，导致视觉 gate 对当前路由机制不成立。

## 3 根因归类

- Root cause 1：scenario wrapper 的 package-script 参数透传未被 runner log 反查。
  - **类别**：README / test framework convention
  - **说明**：脚本中出现 `pnpm test -- --run ...` 时，必须证明最终 runner command 仍按 scoped 文件执行。
- Root cause 2：pixel parity harness 没有随着 URL-addressable routing 迁移同步校准。
  - **类别**：spec-plan / test harness
  - **说明**：hash adapter 仍可用于 bare `/` 静态预览，但不能叠加 query nonce。

## 4 对流程资产的改进建议

- 已落实：更新 `docs/bugs/PATTERNS.md` 模式 4，把 package script 透传扩大范围和 hash-route query 抢占写入检查清单。
  - **落点**：Bug pattern library
  - **优先级**：high
- 建议后续把 `/scenario-run` 外层 runner 固化为 repo-tracked helper，统一保存 per-scenario phase log、scenario trigger log 和 summary JSON。
  - **落点**：`test/scenarios/_shared/scripts/` 或 `scenario-run` skill
  - **优先级**：medium
- 建议后续在 scenario verify 中统一拒绝 `vitest run -- --run` 形态或等价 widened-scope marker。
  - **落点**：scenario README / shared verify helper
  - **优先级**：medium

## 5 建议优先级与后续动作

- 下一步最高价值动作：把这次 shell 外层 full-run runner 提炼为仓库共享脚本，避免后续全量场景复跑靠临时 shell 片段。
- 可延后动作：对历史 frontend scenario wrapper 做一次静态 lint，扫描 `pnpm test -- --run`、缺少 target test file grep、缺少 Playwright pass marker 等 false-green 风险。
