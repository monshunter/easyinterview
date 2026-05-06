# Mock Contract Fixture Tag Directory Gate 交付复盘报告

> **日期**: 2026-05-06
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：修复 `$plan-code-review mock-contract-suite` 发现的目录级 L2 drift，确保 `lint-mock-contract` 能拒绝 `openapi/fixtures/` 下非当前 12 tag 的旧目录。
- 成功证据：
  - Red：新增 `test_retired_fixture_tag_directory_fails_even_when_empty` 后先失败，证明旧 gate 忽略空旧目录。
  - Green：`mock_runtime_boundary.py` 增加 tag 目录集合校验后 focused test 通过。
  - 聚合 gate：`make lint-mock-contract`、`make codegen-check`、`make docs-check` 均通过。
  - 文档 owner：`mock-contract-suite` spec / plan / checklist 原地修订，plan/checklist 恢复 `completed`。

## 2 会话中的主要阻点/痛点

- 空目录漂移不会出现在 Git 状态里。
  - **证据**：`git status --short --untracked-files=all` 无输出，但 filesystem 检查显示 `Growth` / `Mistakes` 目录存在。
  - **影响**：如果只看 Git 状态或 tracked diff，会误判旧目录已清零。

- retired token gate 只覆盖文件内容。
  - **证据**：`make lint-mock-contract` 在空旧目录存在时仍通过；补测后 Red 阶段返回码仍为 0。
  - **影响**：目录名本身作为 contract artifact 时缺少可执行保护。

## 3 根因归类

- 根因：mock runtime boundary gate 未把 fixture tag 目录集合纳入语义契约。
  - **类别**：spec/plan

- 根因：文件内容搜索与 Git 状态不能证明 empty-directory zero-reference。
  - **类别**：README / gate implementation

## 4 对流程资产的改进建议

- 对 fixture、migration、scenario 这类目录结构即契约的 owner，gate 应显式校验目录集合。
  - **落点**：相关 spec/plan gate 和 lint 脚本
  - **优先级**：high

- L2 复扫保留 filesystem-level negative search，而不只依赖 `rg` 和 `git status`。
  - **落点**：`plan-code-review` 执行习惯 / 后续审查记录
  - **优先级**：medium

## 5 建议优先级与后续动作

- 已完成高优先级改进：`mock-contract-suite` 的 spec、plan、checklist 与 `lint-mock-contract` 均已固化目录级 tag set gate。
- 后续可在其他目录契约 owner 中复用同类 gate，例如 scenario case 目录、migration baseline 目录或 generated baseline 目录。
