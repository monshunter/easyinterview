# AI Profile Deploy Catalog Drift Remediation 交付复盘报告

> **日期**: 2026-05-05
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：按用户确认的 L2 finding，原地修复 A3 `003-provider-registry-and-capability-profiles` 的 dev-stack env、active owner docs 与 profile coverage lint 语义 drift。
- 成功证据：`python3 -m pytest scripts/lint/ai_profile_coverage_test.py -q`、`make lint-ai-profile-coverage`、`make lint-config`、context validation、`make docs-check`、`git diff --check` 与 `cd backend && go test ./internal/ai/aiclient/... ./internal/platform/config/... ./cmd/worker -count=1` 均通过。
- Lifecycle：003 plan/checklist 已新增并完成 5.7 remediation，状态恢复 `completed`；Product Scope、Repo Scaffold、Local Dev Stack active spec 与 `docs/spec/INDEX.md` / A3 plans INDEX 已同步。
- Bug linkage：本次 drift 已建档为 [BUG-0009](../bugs/BUG-0009.md)。

## 2 会话中的主要阻点/痛点

- catalog consolidation 的 gate 没有覆盖 deploy-specific env example。
  - **证据**：修复前 `deploy/dev-stack/.env.example` 仍使用 `AI_MODEL_PROFILE_PATH=config/ai-profiles/`，而 `make lint-config` 与 `make lint-ai-profile-coverage` 均通过。
  - **影响**：AIClient-enabled 本地组件可能从样例 env 复制到 retired profile directory，偏离 A3 D-3a 和 A4 env dictionary。
- active owner matrix 滞后于 A3 catalog truth source。
  - **证据**：Product Scope 与 Repo Scaffold active spec 均仍把旧 profile directory 描述为当前 AI profile 真理源或消费路径。
  - **影响**：后续 plan review/code review 可能被 active product/repo owner 文档引回旧目录布局。
- lint 初版需要注意多 capability / 多 profile 表格的保序匹配。
  - **证据**：A3 Product/UI 表中 `Job Picks` 同时列出 `embed` / `rerank` / `chat` 和三个默认 profile；实现中改为保序解析，避免 `set` 顺序造成误报。
  - **影响**：小范围实现调整；未影响交付边界。

## 3 根因归类

- 003 catalog consolidation verification 没有把 deploy env example 和 Product Scope owner matrix 明确列入 executable gate。
  - **类别**：spec-plan
- profile coverage lint 原先只校验 profile 存在与 provider 结构合法，没有校验 Product/UI capability 到 catalog capability 的语义闭环。
  - **类别**：spec-plan
- 多 profile 表格解析需要保序是本次实现细节，不需要额外流程资产变更。
  - **类别**：无需仓库改动

## 4 对流程资产的改进建议

- 后续做 truth source 物理落点迁移时，在 plan verification 中显式列出 deploy examples、active owner matrix、README、runtime config 与 lint coverage 的搜索路径，不能只写“active-scope negative search”。
  - **落点**：spec-plan
  - **优先级**：high
- 对 Product/UI catalog、F3 feature dictionary、runtime catalog 这类三方关系，优先让 lint 比对语义列，而不是只做存在性扫描。
  - **落点**：spec-plan
  - **优先级**：medium
- 可延后评估是否把 deploy env canonical key 检查抽成通用 helper；当前只有 AI profile catalog 一处需要，不必提前抽象。
  - **落点**：无需仓库改动
  - **优先级**：low

## 5 建议优先级与后续动作

- 下一轮最高价值动作：继续用 `/plan-code-review ai-provider-and-model-routing/003-provider-registry-and-capability-profiles backend --base-rev e69ca65` 复跑无 `--fix` 审查，确认 BUG-0009 的 gate 不再漏报同类 drift。
- 备选动作：若要推进能力扩展，先让 A3 002 / C14 voice owner 只消费 `config/ai-profiles.yaml` 中的 `stt` / `realtime` unsupported profile，不在业务侧新增 provider 配置。
