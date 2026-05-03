# Roadmap Rebaseline and Subspec Governance

> **版本**: 3.0
> **状态**: active
> **更新日期**: 2026-05-03

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 `engineering-roadmap` 从旧的“38 child + W0-W5 pending spawn”模型修订为当前产品 / UI / 契约事实驱动的实施地图：

- `docs/spec/INDEX.md` 只保留真实存在的 spec，不再承载 pending backlog。
- 已存在的 A/B/F 基础、契约和质量 spec 保持 active，作为后续实现依赖。
- 未启动的 P0 workstream 只在进入设计或实现时创建 child spec / plan / checklist / context。
- P1/P2 方向只作为 future candidates 记录，不创建空 spec、空 plan 或 draft 占位。
- 当前 product-scope 和 UI 已丢弃的旧模块不再通过 roadmap 间接恢复。

## 2 背景

旧 plan 已完成 W0/W1 阶段的历史职责：ADR-Q1..Q6 accepted，A1-A5、B1-B4、F1、F3 等基础 / 契约 / 质量文档与部分编码 truth source 已落地。之后产品范围和 UI 设计发生收敛，旧 plan 中尚未执行的 Phase 4-7 仍要求 spawn 大量 P0/P1/P2 subject，并让 `docs/spec/INDEX.md` 保留 `_pending_` 占位。

这会造成两个问题：

1. 未创建的 subject 看起来像已批准 backlog，容易绕过 product-scope 的默认丢弃规则恢复旧模块。
2. INDEX 不再是 Header 投影，而变成了混合真实文档和计划占位的二义性列表。

本次原地修订保留已完成历史事实，删除 pending 占位模型，并把后续 workstream 的创建条件写回 roadmap spec。

## 3 质量门禁分类

- **Plan 类型**: `docs-only`。
- **TDD 策略**: 不适用：本 plan 只修订 `docs/spec/` 文档和索引，不修改前端 / 后端 / 工具脚本 / 迁移 / codegen / 测试辅助逻辑。
- **BDD 策略**: 不适用：本 plan 不新增用户可见 UI、API 行为、业务流程或端到端功能；后续任何 P0 workstream 创建实现 plan 时必须单独维护 BDD gate。
- **替代验证 gate**: `validate_context.py` 校验本 plan context；`sync-doc-index --check` 校验 Header / INDEX；`check_md_links.py docs` 校验 Markdown 链接；`git diff --check` 校验文本格式。

## 4 实施步骤

### Phase 1: 历史完成事实保留

#### 1.1 保留 W0 ADR 决策

ADR-Q1..Q6 继续作为架构约束保留在 `docs/spec/engineering-roadmap/decisions/`。它们不再生成 pending child，但认证、异步、分析、部署、隐私和 AI 路由结论仍约束后续 P0 workstream。

#### 1.2 保留已落地基础 / 契约 / 质量 spec

确认 A1-A5、B1-B4、F1、F3 等现有 active spec 仍与 current product-scope 和 UI scope 相容；后续实现直接引用这些 spec 与编码 truth source。

#### 1.3 关闭旧 pending INDEX 模型

旧 plan 中 “38 行 child subspec 占位” 的任务保留为历史记录，但当前策略改为：未创建真实 `spec.md` 的 subject 不进入 `docs/spec/INDEX.md`。

### Phase 2: Roadmap rebaseline

#### 2.1 对齐产品与 UI 真理源

读取并对齐 `docs/spec/product-scope/spec.md`、`docs/ui-design/` 和 `ui-design/src/app.jsx` 的当前模块、路由、上下文和删除范围。

#### 2.2 重写 roadmap spec

把 `engineering-roadmap/spec.md` 改为当前实施地图：保留 active spec 清单、当前 P0 workstream 候选、future candidates、实施顺序和验收标准。

#### 2.3 修订本 plan / checklist / context

把本 plan 从旧 wave spawn 计划改为 roadmap rebaseline 与 child 创建治理计划；同步 checklist、context discovery keywords 和 plans/INDEX。

#### 2.4 删除 pending 索引占位

从 `docs/spec/INDEX.md` 删除所有 `_pending_` 行和“待 spawn”分组说明，只保留真实存在的 spec row。

#### 2.5 同步交叉引用

修订 product-scope 中指向旧 engineering-roadmap v2.2 的说明，使其指向当前 v3.0 的“无 pending 占位 + on-demand child 创建”策略。

#### 2.6 验证文档一致性

运行本 plan §3 的替代验证 gate；若失败，原地修正文档后再复跑。

### Phase 3: 后续 P0 workstream 创建规则

#### 3.1 创建前置条件

任何 P0 workstream 只有在明确进入设计或实现时才能创建 child spec。创建时必须满足：

- product-scope 和 UI 文档已明确保留该用户行为或工程能力。
- 已有同主题 spec / plan 不存在；若存在必须原地修订。
- `context.yaml`、`plan.md`、`checklist.md` 成对完整。
- 涉及代码逻辑时写明 TDD 策略；涉及用户行为时维护 BDD plan / checklist。

#### 3.2 推荐创建顺序

后续实现优先按以下顺序推进：

1. `mock-contract-suite` + `frontend-shell`，建立 fixture-backed mock 和 App 壳。
2. D2-D6 当前 UI workstream：Home / Job Picks / Parse、Workspace / Practice、Report、Resume、Debrief。
3. 后端基础：`backend-auth`、`backend-upload`、`backend-profile`、`backend-async-runtime`。
4. 后端业务域：`backend-targetjob`、`backend-practice`、`backend-review`、`backend-resume`、`backend-debrief`。
5. 集成与上线：`e2e-scenarios-p0`、`analytics-funnel`、`release-gate-and-rollout`。

#### 3.3 Future candidates 延后

嵌入式 readiness、retrieval、privacy export、company/source intel、production voice 和 multi-platform job search 不提前建 spec。触发时必须先确认 product-scope / UI / 合规边界。

## 5 验收标准

- `engineering-roadmap/spec.md` 不再声明 38 child / W0-W5 pending spawn 为当前执行模型。
- `docs/spec/INDEX.md` 只包含真实存在的 `docs/spec/*/spec.md`，无 `_pending_` 行。
- `product-scope` 对 roadmap 的交叉引用不再停留在 v2.2。
- 当前已存在 active spec 均保留，且没有为 P1/P2 future candidates 创建空 spec 或空 plan。
- 本 plan checklist 与 Header / INDEX 投影一致。
- 本 plan §3 的替代验证 gate 全部通过。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 删除 pending 行后看不到未来方向 | 将未来方向保留在 roadmap spec §5.2 / §5.3，而不是放入 INDEX |
| 旧 completed plan 中仍有历史术语 | 保留历史证据，但当前执行口径以本 plan v3.0 与 roadmap spec v3.0 为准 |
| 后续实现重新创建旧模块 | product-scope 默认丢弃规则 + UI 文档删除清单 + roadmap §4.1 共同拦截 |
| P1/P2 能力被提前空壳化 | child 创建规则要求进入设计或实现时才创建 spec / plan |
| A/B/F active spec 与新 roadmap 表述漂移 | 通过 `sync-doc-index --check`、链接检查和后续 plan-review 原地修订 |

## 7 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-03 | 3.0 | 原地重写为 roadmap rebaseline：删除 pending 占位模型，保留 active spec truth source，改为 on-demand child 创建。 | product-scope v1.5 / docs-ui current |
