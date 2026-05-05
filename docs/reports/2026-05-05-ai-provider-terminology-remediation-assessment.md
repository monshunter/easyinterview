# AI Provider Terminology Remediation 交付复盘报告

> **日期**: 2026-05-05
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：按用户确认的方案 A，原地修订 A3 `ai-provider-and-model-routing` spec/plan/checklist、ADR-Q6、A2/A4/B1/roadmap 关联文档，并把 runtime config、Model Profile schema、deploy assets、generated AI vocabulary 从 gateway 口径硬切到 AI provider 口径。
- 成功证据：focused Go tests、`make lint-config`、`make lint`、`make test`、`make build`、`make docs-check`、AI provider terminology lint、active-scope retired terminology 搜索均通过；`make codegen-check` 使用临时 git index 复核生成物无额外漂移。
- 生命周期证据：A3 001 plan/checklist 已恢复 `completed`，`docs/spec/INDEX.md` 与 A3 plans INDEX 已由 `sync-doc-index --fix-index` 同步。

## 2 会话中的主要阻点/痛点

- 旧口径不是局部代码命名，而是 active spec/plan/config/code/generated artifacts 一致地传播错误语义。
  - **证据**：初始 review 同时命中 A3 spec、AIClient config、Model Profile schema 和 A4 env bindings。
  - **影响**：必须先修订文档真理源，再修改代码，否则实现仍会按旧 spec 回归。
- 原有 `env_dict` gate 只验证结构一致性，无法发现它保护的是旧 provider 术语。
  - **证据**：新增 `scripts/lint/ai_provider_terminology.py` 前，`env_dict` 可以通过但 active docs/code 仍保留旧口径。
  - **影响**：需要补语义负向 gate，而不是只补一次性 `rg` 验证。
- 批量替换波及历史报告索引，导致 `docs-check` 出现一个不存在的历史报告链接。
  - **证据**：首次 `make docs-check` 失败于 `docs/reports/INDEX.md` 中不存在的 2026-04-30 report filename。
  - **影响**：历史证据目录必须作为例外处理，active truth source 与历史记录不能同一批替换。
- `make codegen-check` 最后使用真实 index 做 `git diff --exit-code`，会把本轮有意修改的 generated artifacts 当成未提交 drift。
  - **证据**：前置 conventions drift 检查通过，但 final diff 列出本轮 intentional AI vocabulary 注释变更。
  - **影响**：需要临时 index 才能验证“生成后无额外漂移”，该 gate 对未提交的代码生成修订不够友好。

## 3 根因归类

- **spec/plan**：A3/A4 active docs 把 provider 连接描述成旧连接抽象，导致实现和 fixture 沿旧真理源落地。
- **skill / gate**：已有 review gate 没有强制为旧 AI provider 术语添加 repo-tracked 负向 lint。
- **no repo change needed**：本轮批量替换的历史索引问题已在当前变更中修复，但它是执行方式问题，不需要单独修改历史规范。
- **README / Makefile**：`codegen-check` 的真实 index diff 策略可能需要后续优化，避免 intentional generated changes 在提交前无法直接通过。

## 4 对流程资产的改进建议

- 在后续 `/plan-code-review` 类似 terminology remediation 中，要求先列出 active truth source 与 historical exception scope，再做批量替换。
  - **落点**：skill
  - **优先级**：medium
- 为语义迁移类任务优先补 repo-tracked negative lint，并把一次性 `rg` 作为验证而非唯一 gate。
  - **落点**：spec-plan / Makefile
  - **优先级**：high
- 评估 `make codegen-check` 是否应改为比较 codegen 前后 diff，而不是直接要求真实 index 无 diff。
  - **落点**：Makefile / ci-pipeline-baseline
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：下一个相关 owner 应继续用 `scripts/lint/ai_provider_terminology.py` 作为 provider terminology gate，避免 A3/A4/F1/B1 后续文档或 generated artifacts 回写旧口径。
- 次优先级：在后续 A5 / ci-pipeline-baseline 修订中评估 `codegen-check` 的 diff 策略，减少 intentional generated changes 的验证摩擦。
- 可延后：把“active truth source / history exception scope”写成 plan-code-review 的固定输出小节，等下一次术语迁移或大规模 rename 再落地。
