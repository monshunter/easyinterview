# AI Profile Catalog Consolidation 交付复盘报告

> **日期**: 2026-05-05
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：按用户确认，原地修订 A3 `003-provider-registry-and-capability-profiles`，将 Model Profile active truth source 从 per-profile YAML directory 收敛为单一 `config/ai-profiles.yaml` catalog。
- 成功证据：`go test ./internal/ai/aiclient/profile -run 'TestLoader|TestTrackedCatalog' -count=1`、`go test ./internal/ai/aiclient/... -count=1`、`python3 -m pytest scripts/lint/ai_profile_coverage_test.py -q`、`make lint-ai-profile-coverage`、`make lint-config`、context validation、legacy directory active-scope negative search、`make codegen-check`、`make docs-check`、`make lint`、`make test`、`make build`、`sync-doc-index --check` 与 `git diff --check` 均通过。
- Lifecycle：003 plan / checklist 已恢复 `completed`，A3/A4/F3 active specs、history、INDEX 与 config README 已同步。

## 2 会话中的主要阻点/痛点

- 初始设计过早采用一 profile 一文件，当前仅 17 个薄 profile，文件数量带来的审查和维护成本高于收益。
  - **证据**：用户明确质疑 `config/ai-profiles` 文件碎片，不符合最佳实践；本轮最终收敛为单一 catalog。
  - **影响**：需要原地 reopen 已完成 plan，并联动 A3/A4/F3 文档、loader、fixtures、lint 与默认配置。
- 测试 fixture 在第一次 Green 尝试中生成了无 `-` 的 `profiles` mapping，不是 sequence。
  - **证据**：focused Go / pytest 首次 rerun 报 `'profiles' must be a sequence` / `missing profiles[]`。
  - **影响**：小范围返工测试 helper；未影响生产实现边界。

## 3 根因归类

- per-profile YAML 目录的权衡没有在 003 初始设计中绑定到当前 catalog 规模。
  - **类别**：spec-plan
- 测试 helper 生成 YAML 时没有先用结构化 YAML writer，但输入很小，修复成本低。
  - **类别**：无需仓库改动

## 4 对流程资产的改进建议

- 后续 A3/F3 profile catalog 再次扩展时，先在 owner plan 中写明规模阈值和 owner 并发模型，再决定单文件或目录型 catalog。
  - **落点**：spec-plan
  - **优先级**：medium
- 若 profile catalog 后续增长到多人并行维护，再考虑引入生成器或拆分源 + single catalog projection，而不是直接恢复 runtime 目录扫描。
  - **落点**：spec-plan
  - **优先级**：low

## 5 建议优先级与后续动作

- 优先保持当前单一 `config/ai-profiles.yaml` 作为 active truth source，并让后续 002 / C14 / F3 eval 只新增 catalog entry，不新增业务侧 provider 配置。
- 可以延后评估 catalog 拆分生成器；当前 17 个 profile 不需要额外抽象。
