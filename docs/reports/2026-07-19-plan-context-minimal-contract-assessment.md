# Plan Context Minimal Contract 交付复盘报告

> **日期**: 2026-07-19
> **审查人**: Codex

**关联计划**: [Harness v1 渐进迁移 Plan](../spec/harness-simplification/plans/001-harness-v1-migration/plan.md)

## 1 复盘范围与成功证据

本次交付把仍在运行的旧 `context.yaml` 收缩为过渡期最小链接合同：删除顶层与 target discovery、target references，并让 `metadata` 只保留 `name`；同时迁移 validator、generator、change-intake matcher、candidate list、分支推导、相关 Skill、治理说明、模板和 49 份现有清单。

成功证据：

- 全部 `.agent-skills` contract tests：`151 passed`；
- shared context tools、matcher、implement 与 execution-doc focused contract：`50 passed`；
- 49 份 manifest batch validation：`49 passed, 0 failed`；精确 schema audit：`invalid=0`；
- generator 二次运行：`updated=0, reconciled=49`；
- 迁移前后 default target、target identity 与一等链接语义比较：`semantic_link_drift=0`；
- `sync-doc-index --check`、`make docs-check`、Markdown link 与 `git diff --check` 均通过；
- 真实 matcher 回放中，完整 Harness 请求以 medium 置信命中不含 context 的 `harness-simplification/001-harness-v1-migration`，`PRACTICE_SESSION_CONFLICT` 仍命中 `backend-practice/001-plan-and-session-orchestration`；两者均不从已删除 discovery 字段获得额外置信度。

## 2 会话中的主要阻点/痛点

### 2.1 人工 discovery 会制造错误高置信 owner

- **证据**：本次入口 query 曾被旧 matcher 高置信路由到 `openapi-v1-contract/003-breaking-change-gate`，但 live code/docs 搜索确认实际 owner 是 Harness simplification；初次删除 discovery 后又暴露通用 `spec/context` token 产生 low 假候选，最终通过过滤通用词并扫描无 context plan 修复。
- **影响**：错误置信度比明确的“未找到”更昂贵，会诱导 Agent 读取和修订错误 plan。

### 2.2 generator 的目录推断会误改 target identity

- **证据**：首次 dry-run 把 49 份清单中的 9 类 target 名大面积推断为 `default`，并会丢失唯一的 backend/frontend 双 target；写入前被只读审计拦截。
- **影响**：若直接批量写入，会造成目标选择与 Skill 调用语义漂移，超出字段删除授权。

### 2.3 旧字段同时存在于代码、Skill、模板与治理说明

- **证据**：初始负向搜索命中 validator/generator/matcher、6 个 workflow Skill、两套模板、共享合同和 `docs/README.md`；只迁移 YAML 会被后续 generator 或文档流程重新写回。
- **影响**：单点 schema 改动无法形成 hard cut，必须同 Change 迁移 writer、reader、caller 与治理 gate。

## 3 根因归类

- 人工 discovery 同时承担索引、路由和置信度职责，重复投影路径与当前仓库事实。
  - **类别**：spec-plan / skill
- generator 过去优先从目录名猜 target，没有冻结“允许字段收缩不得改变 target identity”的迁移不变量。
  - **类别**：skill
- plan-context 合同允许未知字段与 preservation merge，使旧值能绕过模板继续存活。
  - **类别**：skill / README / AGENTS.md
- 已完成旧 plan/checklist 中仍有历史 discovery verification 文本，但它们是历史证据而非当前 caller。
  - **类别**：无需仓库改动

## 4 对流程资产的改进建议

- 延续当前无 `context.yaml` active plan 发现能力，并在 Phase 2 把它收敛到 Project Arch current-truth lookup；“未找到/低置信 + 证据链”应继续作为正常结果。
  - **落点**：Harness spec-plan / Project Arch tooling
  - **优先级**：high
- 保留 generator 的 target identity 与一等链接语义守恒测试，任何后续 schema 收缩都先 dry-run 并比较语义投影。
  - **落点**：implement shared tooling
  - **优先级**：high
- 最终删除 `context.yaml` 时，以本次 reader/writer/caller 负向清单为起点一次性退出 validator、generator、candidate list 与旧 workflow 入口，不再引入兼容 converter。
  - **落点**：Harness spec-plan
  - **优先级**：high
- 已完成旧 plan/checklist 的历史 verification 文本继续保留；负向 gate 应明确排除历史、Bug、report 和本 Spec 的禁止项描述，只阻断当前生产 caller。
  - **落点**：Harness plan verification gate
  - **优先级**：medium

## 5 建议优先级与后续动作

下一轮应恢复 Harness Phase 1.2，先实现 Project Arch v1 最小 Green；随后在 Phase 2 用不依赖 context 的 current-truth lookup 接管 change-intake/implement owner 解析，再删除本次过渡清单及其全部 reader/writer。最应冻结的迁移纪律是：批量写入前先比较 target identity 与一等链接语义，低置信路由必须显式暴露，不用人工关键词制造高置信。
