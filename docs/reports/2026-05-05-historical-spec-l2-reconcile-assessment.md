# Historical Spec L2 Reconcile 交付复盘报告

> **日期**: 2026-05-05
> **审查人**: Codex

> **说明**: 本报告记录本轮早期 L2 reconcile 快照。用户随后要求忽略历史 PASS / checklist 状态并执行 artifact-level deep reimplementation；最终结论以 [Historical Spec Deep Reimplementation Ledger](./2026-05-05-historical-spec-deep-reimplementation-ledger.md) 和 [Historical Spec Deep Reimplementation 交付复盘](./2026-05-05-historical-spec-deep-reimplementation-assessment.md) 为准。

## 1 复盘范围与成功证据

本次交付覆盖 `historical-spec-implementation-review` 指定的现存 historical spec / plan L1/L2 收口：

- 重新验证 15 个现存 `docs/spec/*/plans/*/context.yaml`，全部 PASS。
- 对 `engineering-roadmap/001-decompose-subspecs` 执行 L1 修复：Phase 3 明确为后续 child 创建治理规则，不创建当前 P0 workstream，plan/checklist 从 active 9/13 收口为 completed 13/13。
- 对 A3 AI gateway 执行 L2 修复：`openai_compatible` adapter 去除 provider-specific model naming assumption，并补足 `ModelFamily` 断言。
- 对 A1 repo scaffold 入口事实执行轻量修复：根 README 的 `test/` 行改为当前 `test/scenarios/` 尚未落地。

成功证据：

- `make docs-check`
- `make codegen-check`
- `make test`
- `make build`
- `git diff --check`
- `cd backend && go test ./internal/ai/aiclient/... -count=1`
- `make lint-openapi && make validate-fixtures && make openapi-diff`
- `make codegen-events-check`
- `node --test ui-design/ui-design-contract.test.mjs`

## 2 会话中的主要阻点/痛点

- `engineering-roadmap/001-decompose-subspecs` 已经完成主体 rebaseline，但 Header 仍是 `active` 且 checklist Phase 3 保留 4 个未勾选规则项。
  - **证据**：`list_context_candidates.py` 最初把该 plan 推荐为 `/implement` 候选，状态为 active、进度 9/13。
  - **影响**：容易把“后续 P0 workstream 创建规则”误当成当前 implementation target，触发用户明确禁止的新 child 创建路径。

- A3 adapter contract tests 使用 vendor-style model id，掩盖了 `modelFamily()` 的通用命名问题。
  - **证据**：Red phase 失败于 `ModelFamily="chat-primary-2026-05"` / `"embed"`，而期望是 `"chat-primary"` / `"embed-small"`。
  - **影响**：如果进入真实 OpenAI-compatible gateway 或不同模型命名约定，AI observability / cost label 会错误聚合。

- 根 README 对 `test/` 的描述暗示场景测试框架已经存在。
  - **证据**：`README.md` 原文写“跨服务测试根目录与场景测试框架”，但 `test/README.md` 明确当前尚未落地 `test/scenarios/`。
  - **影响**：新会话或新开发者可能误以为 scenario framework 已可执行。

## 3 根因归类

- Roadmap plan 生命周期漂移。
  - **类别**：spec-plan
  - **根因**：Phase 3 是治理规则记录，但 checklist 用未勾选项表达 future rule，导致 `/implement` 候选筛选无法区分“未执行规则”和“未完成实施”。

- Model family derivation 测试不足。
  - **类别**：spec-plan / no repo change needed
  - **根因**：A3 计划写了“不锁 model_id 命名”，但测试 fixture 使用具体 vendor/model branding，且没有直接断言 derived `ModelFamily`。

- README 入口文案二义性。
  - **类别**：README
  - **根因**：`test/README.md` 已修订为当前事实，根 README 未同步到同样精确的未落地表述。

## 4 对流程资产的改进建议

- 在 plan-review 检查中把“active + unchecked future-rule items”列为 L1 finding。
  - **落点**：`/plan-review` skill 或 spec-plan 审查清单
  - **优先级**：high

- 在 AI gateway / provider-neutral adapter 的 future plan 中禁止使用真实 vendor branding 作为默认 fixture。
  - **落点**：A3 spec-plan / test convention
  - **优先级**：medium

- 调整 candidate listing 的展示语义，把 draft-gated plan 与 active implementation candidate 分开。
  - **落点**：`/implement` shared script
  - **优先级**：medium

- 入口 README 变更时同步核对对应根目录 README。
  - **落点**：README maintenance checklist
  - **优先级**：low

## 5 建议优先级与后续动作

最高优先级是把 future-rule checklist 的 L1 识别固化，避免 completed governance plan 被再次推荐为当前 implementation owner。

中优先级是让 provider-neutral test fixture 成为 AI gateway 后续计划的默认写法，并把 draft-gated plan 从 `/implement` 推荐列表中降级显示。`ai-gateway-and-model-routing/002-tools-streaming-and-stt` 当前仍保持 draft-gated，未进入实现。
