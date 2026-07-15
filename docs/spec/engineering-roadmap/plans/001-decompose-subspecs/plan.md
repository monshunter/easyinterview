# Roadmap Rebaseline and Subspec Governance

> **版本**: 3.7
> **状态**: completed
> **更新日期**: 2026-07-10

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

`engineering-roadmap` 作为当前产品、UI、契约和编码 truth source 的实施地图。本计划固化以下治理合同：

- `docs/spec/INDEX.md` 只投影真实存在的 `docs/spec/*/spec.md`。
- 已存在的基础、契约、质量和产品 owner spec 作为后续实施依赖。
- 候选 workstream 只有在产品范围、UI 设计文档或用户指令明确进入设计 / 实施时才创建 child spec / plan / checklist / context。
- P1/P2 方向只保留在 roadmap spec 的 future candidate 区域，不创建空 spec、空 plan 或 pending INDEX 行。
- 范围外 root spec、route、模块和技术草稿只作为边界输入，不能作为当前 owner 或恢复依据。

## 2 背景

当前产品范围已经收敛到 `JD / 简历 -> 模拟面试 -> 报告 -> 复练当前轮 / 进入下一轮`。Roadmap 的职责不是承载 backlog 空壳，而是把当前可执行 owner、依赖顺序和创建规则写清楚，避免后续设计或实现绕过 `product-scope`、`docs/ui-design/`、`frontend/` 和编码 truth source。

本计划保持 docs-only：它只维护 roadmap spec、plan、checklist、context 和索引投影，不修改运行时代码。

## 3 质量门禁分类

- **Plan 类型**: `docs-only`。
- **TDD 策略**: 不适用；本计划只修订 `docs/spec/` 文档和索引，不修改前端、后端、工具脚本、迁移、codegen 或测试辅助逻辑。
- **BDD 策略**: 不适用；本计划不新增用户可见 UI、API 行为、业务流程或端到端功能。后续用户可见 workstream 的实施计划必须维护 BDD gate。
- **替代验证 gate**: `validate_context.py` 校验本 plan context；`sync-doc-index --check` 校验 Header / INDEX；`make docs-check` 校验 Markdown 链接；technical-draft zero-reference search 校验范围外草稿目录名、文件名和 shorthand；`git diff --check` 校验文本格式。

## 4 实施步骤

### Phase 1: 当前 truth source 投影

#### 1.1 ADR 约束投影

ADR-Q1..Q6 作为认证、异步、分析、部署、隐私和 AI 路由的架构约束，由 `engineering-roadmap/decisions/` 持有，并在 roadmap spec 中摘要为当前约束。

#### 1.2 Active owner spec 投影

确认当前基础、契约、质量、产品和 UI owner spec 与 `product-scope`、`docs/ui-design/`、`frontend/` 保持一致；后续实施直接引用这些 owner spec 与编码 truth source。

#### 1.3 INDEX 真实投影

`docs/spec/INDEX.md` 只列真实存在且 Header 合规的 spec。未创建实体的候选项只出现在 roadmap spec 正文的 candidate / sequence 区域。

### Phase 2: Roadmap 当前实施地图

#### 2.1 对齐产品与 UI 设计文档

读取并对齐 `docs/spec/product-scope/spec.md`、`docs/ui-design/` 和 `frontend/src` 的当前模块、route、上下文和范围外边界。

#### 2.2 修订 roadmap spec

`engineering-roadmap/spec.md` 只描述当前 truth source 关系、active owner、P0 workstream、future candidates、实施顺序和验收标准。

#### 2.3 修订本 plan / checklist / context

本 plan、checklist 和 context discovery 只表达 roadmap rebaseline、按需 child 创建和 no-pending INDEX 合同。

#### 2.4 同步交叉引用

`product-scope`、`docs/spec/INDEX.md` 和本 subject plans INDEX 必须指向当前 roadmap 版本和当前 plan 版本。

#### 2.5 验证文档一致性

运行本 plan §3 的替代验证 gate；任何失败都在当前 owner 文档内原地修正后复跑。

### Phase 3: 后续 child 创建规则

#### 3.1 创建前置条件

任一 child spec / plan 创建前必须满足：

- `product-scope` 与 UI 设计文档已明确保留该用户行为或工程能力。
- 同主题 owner 不存在；若已存在，必须原地修订现有 owner。
- `context.yaml`、`plan.md`、`checklist.md` 成对完整。
- 涉及代码逻辑时写明 TDD 策略；涉及用户行为时维护 BDD plan / checklist。

#### 3.2 推荐创建顺序

后续实现优先按当前产品闭环推进：

1. App shell、auth、settings 与 fixture-backed mock。
2. Home / Parse、Workspace / Practice、Report Dashboard、Resume Workshop。
3. Backend auth、upload、target job、practice、review、resume、async runner。
4. E2E scenarios、analytics funnel、release gate 与 rollout。

#### 3.3 Future candidates 延后

Readiness、retrieval、privacy export、production voice、company/source intel 和 multi-platform job search 只有在产品 / UI / 合规设计确认后才创建 owner 文档。

### Phase 4: 技术草稿引用边界

#### 4.1 继承统一 owner matrix

技术契约职责由 `product-scope` §1.5 owner matrix 持有。本 roadmap 消费该矩阵，不复制第二套 API / DB / event / metrics / logging / config 映射。

#### 4.2 技术草稿 zero-reference

当前项目文档、代码注释、生成源、生成物、日志与报告不得把范围外技术草稿目录名或文件名作为 truth source。需要描述责任来源时，正文只能引用当前 owner spec、history 或编码 truth source。

#### 4.3 编码 truth source 注释规范

`shared/conventions.yaml`、lint 脚本、codegen source 与 generated artifacts 必须把共享约定、OpenAPI、DB、event、observability 的字段和 gate 归属到当前 owner。

#### 4.4 固化 zero-reference gate

技术草稿处置前必须重新运行本 plan §3 gate，并额外确认目录名、文件名、shorthand、Markdown 链接、当前实施前置和外部 truth-source 口径均为零匹配。

## 5 验收标准

- `engineering-roadmap/spec.md` 只描述当前实施地图、active owner、候选项和创建规则。
- `docs/spec/INDEX.md` 只包含真实存在的 `docs/spec/*/spec.md`。
- `product-scope` 对 roadmap 的交叉引用指向当前版本。
- P1/P2 future candidates 没有空 spec、空 plan 或 pending INDEX 行。
- 技术契约 owner matrix 由 `product-scope` §1.5 持有，roadmap 和 child spec 只消费当前 owner spec / 编码 truth source。
- 当前文档和编码 truth source 不把范围外技术草稿目录名、文件名、Markdown 链接或外部 truth source 作为当前依据。
- 本 plan checklist 与 Header / INDEX 投影一致。
- 本 plan §3 的替代验证 gate 全部通过。
- `context.yaml` discovery lists contain no duplicate scalar entries.

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 未来方向缺少可见入口 | 未来方向只放在 roadmap spec candidate 区域，进入设计 / 实施时再创建 owner 文档 |
| 后续实现绕过当前 owner | child 创建规则要求先检查 product-scope、UI 设计文档和现有 owner |
| P1/P2 能力提前空壳化 | child 创建规则要求进入设计或实现时才创建 spec / plan |
| A/B/F active spec 与 roadmap 表述漂移 | 通过 `sync-doc-index --check`、链接检查和后续 plan-review 原地修订 |
| 技术草稿处置后出现断链或上游缺口 | product-scope §1.5 owner matrix + Phase 4 zero-reference gate |

## 7 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-07-10 | 3.7 | Remove the duplicate AI provider discovery keyword and verify context list uniqueness. | tech-debt pruning |
| 2026-07-10 | 3.6 | Reword out-of-scope route and technical-draft boundaries without changing child-creation rules. | tech-debt pruning |
| 2026-07-10 | 3.5 | Reword backlog wording to empty-shell governance without changing child-creation rules. | tech-debt pruning |
| 2026-07-06 | 3.4 | Reword plan to current execution-map governance, no-pending INDEX model, child-creation rules, and technical-draft zero-reference gates. | product-scope v1.38 / engineering-roadmap v3.29 |
| 2026-05-05 | 3.3 | Codify technical-draft zero-reference gates and current owner-spec responsibility. | product-scope v1.7 / engineering-roadmap v3.4 |
| 2026-05-05 | 3.2 | Add technical-draft reference governance and source-comment ownership rules. | product-scope v1.6 / engineering-roadmap v3.3 |
| 2026-05-05 | 3.1 | L1 plan-review remediation: confirm Phase 3 only records future child-creation governance. | docs-only L1 remediation |
| 2026-05-03 | 3.0 | Rebaseline roadmap to real spec index projection, active owner truth sources, and on-demand child creation. | product-scope v1.5 / docs-ui current |
