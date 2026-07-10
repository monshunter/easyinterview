# Python Contract Test Aggregation 交付复盘报告

> **日期**: 2026-07-10
> **审查人**: Codex

**关联计划**: [Local Quality Gates Bootstrap](../spec/ci-pipeline-baseline/plans/001-local-quality-gates/plan.md)
**关联 Bug**: [BUG-0156](../bugs/BUG-0156.md)

## 1 复盘范围与成功证据

- 本批次把 `scripts/` 与 `.agent-skills/` 的 Python contracts 接入既有根 `make test`，补齐 `requirements-dev.txt`，并修正一条 work-journal 陈旧断言。
- Focused RED 为 2 failures，GREEN 为 2 / 2；全量 Python suite 为 464 passed + 4269 subtests。
- 根 `make test` 通过 Python、backend 全包和 frontend 136 files / 836 tests；零匹配失败探针证明 Python 非零会阻止后续 Go/Vitest。
- `make lint`、A5/product contexts、requirements 解析、docs/index/diff/pruning gates 全部通过。

## 2 会话中的主要阻点/痛点

- Python tests 有规模但没有根执行入口。
  - **证据**：collect-only 得到 scripts 318 tests、skills 145 tests；旧 `make test` 只包含 Go/Vitest。
  - **影响**：work-journal contract 已失败，但常规根 gate 无法发现。
- 当前规则与测试字面量发生漂移。
  - **证据**：SKILL 要求把中文 phase 翻译/概括成英文，测试仍要求直接 lowercase remainder。
  - **影响**：若按失败测试倒改 SKILL，会破坏 ASCII-only commit policy 的完整性。
- 首次 requirements dry-run 选错了系统 Python 上下文。
  - **证据**：Homebrew Python 按 PEP 668 拒绝系统环境 pip dry-run；临时 `--target` dry-run随后通过。
  - **影响**：产生一次无文件变更的验证返工，不是仓库实现缺陷。

## 3 根因归类

- A5 聚合合同未随 Python test surface 增长更新。
  - **类别**：spec/plan
  - 本次已原地修订 A5，不创建 sibling plan。
- work-journal test 锁定旧文案而非当前语义。
  - **类别**：skill
  - 本次只更新 owner test，SKILL 当前规则保持不变。
- PEP 668 dry-run 误用。
  - **类别**：无需仓库改动
  - 后续在隔离环境或临时 target 验证即可。

## 4 对流程资产的改进建议

- 后续新增顶层 Python test root 时，同步扩展 A5 Makefile contract 的 suite inventory；不要依赖 pytest 的默认全仓扫描吸收未知目录。
  - **落点**：A5 spec/plan + `scripts/lint/makefile_dry_run_test.py`
  - **优先级**：high
- 技术债重扫继续比较“可收集 tests”与“根 gate 实际命令”，覆盖 Go/TS/Python 之外的新 runner。
  - **落点**：product-scope pruning checklist
  - **优先级**：medium
- requirements 安装验证始终使用隔离环境，不在 Homebrew/system Python 上尝试写入或绕过 PEP 668。
  - **落点**：无需仓库改动；README 已要求隔离环境
  - **优先级**：low

## 5 建议优先级与后续动作

- 下一轮最高价值动作是继续扫描未进入根 gate 的 executable scripts 与 language-specific test roots；当前 Python 两个 owner root 已有固定合同。
- BUG-0156 与 A5 Phase 9 已承接本批次，不需要新增 sibling plan 或修改 work-journal SKILL。
- 可以延后为 requirements 引入独立 lock 工具；当前只有两个精确 pin，新增管理层没有收益。
