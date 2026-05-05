# AI Provider Runtime Bootstrap Remediation 交付复盘报告

> **日期**: 2026-05-05
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`ai-provider-and-model-routing/003-provider-registry-and-capability-profiles` 的 L2 remediation，覆盖生产 registry/profile bootstrap、profile reload warn、active profile anti-stub gate、repo-tracked active profile 迁移与 lifecycle closeout。
- 关联 Bug：[BUG-0008](../bugs/BUG-0008.md)。
- 成功证据：`cd backend && go test ./internal/ai/aiclient/... -count=1`、`python3 -m pytest scripts/lint/ai_profile_coverage_test.py scripts/lint/env_dict_test.py scripts/lint/ai_provider_terminology_test.py -q`、`make lint-ai-profile-coverage`、`make lint-config`、context validation、`make docs-check`、`make lint`、`make test`、`make build`、`git diff --check` 均通过。
- 状态证据：003 plan / checklist 已恢复 `completed`，`plans/INDEX.md` 和 A3 spec §7 已把 003 状态投影改回 completed。

## 2 会话中的主要阻点/痛点

- 原交付证据没有证明生产 runtime bootstrap。
  - **证据**：L2 review 发现 `ResolveSelectedProviders` 只在测试中调用；`aiclient.New` 只保存注入 resolver/provider，没有实际读取 registry/profile path 或 materialize provider-ref adapter。
  - **影响**：历史 Phase 1 / 3 checklist 可以在 loader/router 单测通过的情况下遗漏真实启动路径 secret fail-fast。
- profile coverage 正向 fixture 继续使用 `unit-test-stub`。
  - **证据**：新增 `test_fails_when_active_profile_uses_stub_provider` 后，修复前 lint 返回 OK；真实 `config/ai-profiles/` active defaults 也指向 `unit-test-stub`。
  - **影响**：F3/Product UI coverage gate 证明了结构存在，却没有证明 active P0 defaults 符合 D-5/C-9 的非 stub 部署语义。
- profile hot reload failure 可用但不可见。
  - **证据**：profile loader poller 保留旧快照，但丢弃 `Reload` error；新增 `OnWarn` focused test 后才覆盖 warning 信号。
  - **影响**：配置漂移会被静默吞掉，运维和测试都缺少失败观测点。

## 3 根因归类

- runtime bootstrap 漏证属于 `spec-plan` 根因：003 plan 的原始 verification 更偏 loader/router 单元语义，没有把“生产入口必须消费 truth source”列成独立 gate。
- active anti-stub 漏证属于 `spec-plan` 根因：coverage lint 只覆盖 profile/provider 结构一致性，没有覆盖 active/default 状态下的禁止项。
- reload warning 漏证属于 `spec-plan` 根因：热加载任务写了“失败不污染旧快照”，但没有把“失败必须可观测”单独转成测试断言。
- 无需修改 `AGENTS.md` 或 `/plan-code-review` skill：本次问题正是 deep L2 review 找出并通过原计划原地 remediation 收口，流程入口有效；后续改进应落在相似 provider/config plans 的 checklist gate。

## 4 对流程资产的改进建议

- 在后续 provider/config runtime plan 中加入 bootstrap truth-source gate。
  - **落点**：spec-plan
  - **优先级**：high
  - **建议**：任何配置 truth source loader 交付时，checklist 必须同时覆盖 loader 单测、runtime bootstrap/entrypoint 调用点、secret fail-fast 与 README wiring 示例。
- 将 active/default 禁止项写成 lint 负向 fixture。
  - **落点**：spec-plan
  - **优先级**：high
  - **建议**：当 spec 声明 test-only、offline-only、disabled-only 语义时，lint 或 unit test 必须含一个 repo-like negative fixture，而不是只靠 `rg` 或人工 review。
- 热加载语义拆成两个断言。
  - **落点**：spec-plan
  - **优先级**：medium
  - **建议**：后续 hot reload checklist 同时写明“失败不污染旧快照”和“失败有 warn/metric/log 信号”，避免只测状态保留。

## 5 建议优先级与后续动作

- 下一轮最高价值动作：在 `002-tools-streaming-and-stt` 激活 speech/realtime adapter 前，把 bootstrap truth-source gate、active/default 禁止项负向 fixture、hot reload warning 断言复制到该 plan 的 verification 阶段。
- 可延后动作：等下一个 provider/config 类 workstream 再评估是否把这些规则抽进 `/tdd` 或 `/plan-code-review` skill；当前 BUG-0008、003 remediation checklist 与本报告已经能为近期 follow-up 提供足够具体的检查点。
