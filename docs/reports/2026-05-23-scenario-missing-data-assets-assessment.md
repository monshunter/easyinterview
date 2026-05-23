# Scenario Missing Data Assets 交付复盘报告

> **日期**: 2026-05-23
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：补齐 `E2E.P0.014`-`E2E.P0.017` 与 `E2E.P0.065`-`E2E.P0.069` 共 9 个 Ready e2e 场景缺失的 `data/seed-input.md` 与 `data/expected-outcome.md`。
- 资产证据：新增 18 个场景 data 文件；结构检查显示 `scenario_dirs=87`、`missing_files=0`。
- runner 证据：`scenario-run-20260523T073633Z` 串行执行 9 个场景全部 PASS，summary 位于 `.test-output/runs/scenario-run-20260523T073633Z/summary.json`，`passed=9`、`failed=0`、`errored=0`。
- 辅助验证：`make docs-check` 通过，`make dev-doctor` 通过，350 个 scenario shell 脚本 `bash -n` 通过。
- 关联 Bug：[BUG-0093](../bugs/BUG-0093.md)。

## 2 会话中的主要阻点/痛点

- Ready 场景的结构资产缺失没有被前一轮 owner gate 修复发现。
  - **证据**：`INDEX.md` 中 9 个场景均为 `Ready`，但结构扫描发现缺少 `data/seed-input.md` 或 `data/expected-outcome.md`。
  - **影响**：场景虽然能通过脚本语法和部分 runner gate，但缺少固定输入/期望输出，后续审查者无法只靠目录资产复原场景意图。
- 运行证据需要临时外层 wrapper 汇总。
  - **证据**：场景脚本自身写 `.test-output/e2e/<scenario>/trigger.log`，本次额外生成 `.test-output/runs/scenario-run-20260523T073633Z/` 汇总每个阶段外层日志和 result line。
  - **影响**：复跑可行，但 run summary 仍依赖临时 shell 片段，未完全沉淀为 repo-tracked 执行入口。
- 本机 locale warning 仍会污染 shell 输出。
  - **证据**：多次 `bash` 调用输出 `LC_ALL: cannot change locale (C.UTF-8)`，但退出码为 0。
  - **影响**：不影响本次结果，但会降低长日志可读性。

## 3 根因归类

- Ready 状态缺少资产完整性 preflight。
  - **类别**：README / test framework convention
  - **说明**：`test/scenarios/README.md` 已声明必需文件，但执行路径没有共享脚本强制检查该清单。
- run artifact 聚合仍停留在会话级实现。
  - **类别**：skill / test framework helper
  - **说明**：`/scenario-run` skill 要求 run summary，但仓库暂无统一 helper，导致每次 targeted/full run 容易重写临时 shell。
- locale warning 属于本机环境噪声。
  - **类别**：无需仓库改动
  - **说明**：当前 warning 没有改变测试退出码或验证判断。

## 4 对流程资产的改进建议

- 在 `docs/bugs/PATTERNS.md` 模式 4 中加入 Ready/Verified 场景资产完整性检查。
  - **落点**：Bug pattern library
  - **优先级**：high
- 新增 repo-tracked scenario preflight helper，检查 Ready/Verified 场景必需文件清单，并让 `/scenario-env verify` 或 `/scenario-run` 在 runner 前调用。
  - **落点**：`test/scenarios/_shared/scripts/` 或 scenario skill
  - **优先级**：medium
- 将本次外层 run wrapper 固化为共享脚本，统一生成 `.test-output/runs/<run-id>/summary.json`、`results.ndjson` 和 per-scenario phase log。
  - **落点**：`test/scenarios/_shared/scripts/`
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：提交当前资产补丁，并在 commit message 中关联 `BUG-0093`。
- 下一轮建议：把结构 preflight 与 run summary wrapper 提炼成 repo-tracked helper，避免后续场景运行继续依赖临时 shell。
- 可延后处理：修复本机 `C.UTF-8` locale warning，或在环境诊断文档中记录该 warning 的无害边界。
