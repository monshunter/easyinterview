# Historical Spec Implementation Review Spec

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-04

## 1 背景与目标

`engineering-roadmap` v3.0 已把旧的 38 child / W0-W5 pending spawn 模型收敛为当前 active spec、P0 workstream 候选和 on-demand child 创建规则。当前仓库仍保留多个历史 active spec、completed plan、生成物 truth source 与静态 UI 原型。后续如果直接对这些历史 spec 执行 `/implement`，容易把旧文档口径、旧 route、旧 API / DB / event / config 协议或旧 UI 交互重新带入代码。

本 subject 定义一条可重复执行的历史 spec 实施与 L2 code review 编排：先清点与 L1 文档对齐，再按依赖顺序实施，最后对每个已完成 plan 执行 `/plan-code-review --fix` 与全局收口。

目标是：

1. **入口统一**：为新对话提供稳定执行入口，避免每次重新讨论 `/implement` 与 `/plan-code-review` 的顺序。
2. **文档先行**：任何历史 spec / plan / checklist / context 漂移必须先通过 L1 修订解决，再进入实现。
3. **TDD / BDD 不降级**：本编排 plan 不替代具体 plan 的 TDD / BDD gate；任何代码或用户行为实现仍由原 subject 的 plan 承担。
4. **并行有边界**：只读发现可并行；写入阶段必须按 disjoint scope 与 truth source 风险分类执行。
5. **证据闭环**：每个实施批次都有 context validation、review/fix 证据、测试证据、retrospective 与 work journal。

## 2 范围

### 2.1 In Scope

- 对当前 `docs/spec/INDEX.md` 中真实存在的 active / completed historical spec 进行实施 readiness 清点。
- 将 [product-scope](../product-scope/spec.md)、[engineering-roadmap](../engineering-roadmap/spec.md)、[docs/ui-design](../../ui-design/INDEX.md) 与 `ui-design/` 静态原型列为所有 spec / plan 校正的强制参考资料。
- 为每个候选 plan 记录 context validation、owner subject、状态、target、写入范围、依赖和并行安全分类。
- 在进入 `/implement` 前执行必要的 `/plan-review --fix`，使 spec、plan、checklist、context 与当前 product-scope、engineering-roadmap、UI 设计和编码 truth source 对齐。
- 对可实施 plan 执行 `/implement`，并由原 plan 的 `/tdd` checklist 顺序推进代码、测试、文档和 lifecycle 更新。
- 对已完成实现或 remediation 的 plan 执行 `/plan-code-review --fix`，核对代码事实、接口、协议、生成物和测试证据。
- 对 shared conventions、OpenAPI、events/jobs、migrations、config、AI gateway、observability、prompt/rubric、UI prototype、frontend/backend domain 等跨域 truth source 做串行收口。
- 记录最终验证证据、retrospective、bug-report 判断和 work journal。

### 2.2 Out of Scope

- 本 subject 不直接定义任何业务 API schema、DB schema、事件 payload、UI 页面或用户交互细节。
- 本 subject 不替代各原始 subject 的 spec / plan / checklist；发现具体设计漂移时，必须回到原 subject 原地修订。
- 本 subject 不创建 `test/scenarios/` 场景资产；若某后续实现 plan 引入用户可感知行为，BDD 文件和场景资产由该 implementation plan 生成和维护。
- 本 subject 不承诺所有 plan 可并行执行；并行资格必须在执行期基于当前 diff、写入范围和 truth source 风险重新判定。
- 本 subject 不处理未创建真实 `spec.md` 的 future workstream，除非用户在执行期明确进入对应设计或实现。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 执行入口 | 先 inventory / L1，再 `/implement`，再 `/plan-code-review --fix` | 同一 plan 不允许 implement 与 L2 review 并行 |
| D-2 | L1 优先 | spec / plan / checklist / context 漂移必须先修订 | 防止按旧产品或旧协议编码 |
| D-3 | TDD owner | 具体代码实现仍由原 plan 的 `/implement` -> `/tdd` 执行 | 本 subject 不绕过原 checklist |
| D-4 | BDD owner | 用户可感知 UI / API / workflow 的 BDD 由具体 feature plan 维护 | 本 subject 不生成全局 BDD |
| D-5 | 并行模型 | 只读发现可并行；写入只允许 disjoint scope；truth source 串行 | 避免 shared/generated/index 冲突 |
| D-6 | Truth source 串行区 | `shared/`、`openapi/`、`migrations/`、`config/`、events/jobs、generated artifacts 串行 | 这些区域影响跨域接口和生成物 |
| D-7 | 完成定义 | 目标 plan 的 L1/L2、测试、build、docs/index、retrospective/work-journal 均有证据 | 不以“代码已改”单独视为完成 |

### 3.2 执行期决策

| ID | 决策 | 默认 | 何时重新判断 |
|----|------|------|--------------|
| Q-1 | 首批 historical spec 批次 | 从 active spec 与最近 remediation 风险最高的 subject 开始 | 新对话启动时根据 `git status`、INDEX、用户目标确定 |
| Q-2 | 是否并行实现多个 plan | 默认不并行 | 只有写入范围不重叠且不触碰 truth source 串行区时才允许 |
| Q-3 | 已 completed plan 是否 reopen | 默认原地 reopen 后修订，再验证后恢复 completed | 当 review 发现完成态 plan 与当前事实冲突时 |

## 4 设计约束

### 4.1 顺序约束

执行顺序固定为：

1. Scope Inventory：清点 historical spec / plan / checklist / target / status / write scope。
2. L1 Reconcile：对文档漂移执行 `/plan-review --fix`。
3. Implementation Runway：按依赖顺序执行 `/implement`，由原 plan 进入 `/tdd`。
4. L2 Remediation：每个完成 phase 或 completed plan 后执行 `/plan-code-review --fix`。
5. Final Reconcile：执行全局 docs、contract、codegen、test、build、retrospective、bug-report、work-journal 收口。

同一 plan 内，后一步依赖前一步证据；不能跳过 L1 或把 L2 review 与 implementation 同时启动。

### 4.2 校正参考资料

每个 historical spec / plan 进入 L1 或 L2 前，必须按以下参考资料矩阵校正语义：

| 参考资料 | 用途 | 约束 |
|----------|------|------|
| [product-scope](../product-scope/spec.md) | 产品范围、阶段边界、丢弃规则、用户价值承载方式 | 任何旧功能、旧模块或旧业务流恢复前必须先修订本 spec |
| [engineering-roadmap](../engineering-roadmap/spec.md) | active spec 地图、P0 workstream 候选、on-demand child 创建规则、并行边界 | 不得恢复旧 pending child / wave 执行模型 |
| [docs/ui-design](../../ui-design/INDEX.md) | 当前 UI 模块、用户流程、交互契约、删除模块边界 | 任何 UI / 用户交互 plan 必须与这里的模块文档一致 |
| `ui-design/` 静态原型 | 当前可视化交互事实和 UI contract tests | 文档不能引用已删除 screen、route、画板标签或旧入口作为当前目标 |
| 编码 truth source | `openapi/`、`shared/`、`migrations/`、`config/`、generated outputs | 接口、协议、枚举、事件、DB、feature flag 以当前编码 truth source 为准 |

若以上参考资料之间出现冲突，默认优先级为：

1. product-scope 的产品范围与丢弃规则。
2. docs/ui-design 与 `ui-design/` 的当前用户交互事实。
3. engineering-roadmap 的实施边界和 on-demand 创建规则。
4. 各 active engineering spec 与编码 truth source。

冲突不能靠实现侧自行解释；必须回到对应 owner spec / plan 原地修订后再继续。

### 4.3 并行约束

允许并行：

- 多个 subject 的只读 `validate_context.py`、INDEX 扫描、旧词/旧 route 搜索、L1/L2 finding 草稿。
- 不同 subject 的文档审查，只要不同时写同一个 `docs/spec/INDEX.md` 或同一个 subject `plans/INDEX.md`。

谨慎并行：

- 不同 subject 的实现，且写入范围完全 disjoint，例如一个只改 `docs/ui-design/`，另一个只改 `backend/internal/ai/aiclient/`。
- 多 agent 协作时，每个 worker 必须有明确 path ownership，且不得改同一 generated output 或同一 INDEX。

禁止并行：

- 同一 plan 的 `/implement` 与 `/plan-code-review`。
- 同一 checklist 的多个 phase。
- 任何会同时修改 `shared/conventions.yaml`、`openapi/openapi.yaml`、`shared/events.yaml`、`shared/jobs.yaml`、`migrations/`、`config/feature-flags.yaml` 或 generated artifacts 的任务。
- 会同时修改全局 `docs/spec/INDEX.md` 的多个写入任务，除非由单一 owner 集成。

### 4.4 质量约束

- 文档 gate：`sync-doc-index --check`、Markdown link check、必要时执行 docs/spec heading anchor audit。
- Contract gate：按目标 subject 执行 OpenAPI、fixture、event/job、migration、config、conventions、codegen drift gate。
- Test gate：按目标 plan 的 TDD checklist 执行 focused tests；最终执行 `make test`。
- Build gate：最终执行 `make build`；占位 build 必须如实记录。
- Review gate：每个完成 plan 必须有 `/plan-code-review --fix` 或明确的“不适用 + 替代验证”记录。
- 收尾 gate：成功交付后执行 `/retrospective` 判断；涉及 bugfix 时执行 `/bug-report` 判断；提交前执行 `/work-journal`。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| 本 subject spec / plan | historical-spec-implementation-review | 只定义执行编排、并行规则和全局收口 |
| 原始 product / engineering spec | product-scope / engineering-roadmap | 产品范围、roadmap、active spec 与 future workstream 边界 |
| 原始 implementation subject | 各 `docs/spec/*` owner | 具体 API、DB、event、config、UI、backend/frontend 行为真理源 |
| L1 文档修复 | `/plan-review` | 修订 spec / plan / checklist / context 的一致性 |
| 实现执行 | `/implement` -> `/tdd` | 按原 plan checklist 串行实施代码、测试和文档 |
| L2 code review | `/plan-code-review` | 审查代码事实、接口、协议、测试证据与 plan 完成项一致性 |
| 全局验证 | 当前 session owner | 集成所有 review/fix 后运行 docs、contract、test、build 和收尾技能 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | Scope inventory 完整 | 新对话启动历史 spec 执行 | 完成 Phase 1 | 每个候选 plan 有 context validation、状态、target、依赖、写入范围、并行安全分类 | 001 Phase 1 |
| C-2 | L1 先于实现 | 某 plan 存在 spec/plan/checklist/context 漂移 | 尝试进入 `/implement` | 先通过 `/plan-review --fix` 修订并验证；未修订不得实施 | 001 Phase 2 |
| C-3 | 串行实施 | 某 plan 进入 `/implement` | 执行 checklist | 按 phase 顺序进入 `/tdd`；完成项立即同步 checklist 与证据 | 001 Phase 3 |
| C-4 | L2 审查闭环 | 某 plan 已完成实现或 remediation | 执行 `/plan-code-review --fix` | 代码事实、接口、协议、测试证据与 spec/plan/checklist 一致；finding 均已修复或明确降级记录 | 001 Phase 4 |
| C-5 | 并行边界明确 | 多个 subject 可能并行 | 做并行判定 | 只读任务可并行；写入任务只有 disjoint scope 才可并行；truth source 串行区不并行 | 001 Phase 5 |
| C-6 | 全局收口完成 | 所有目标 plan 完成 L2 | 执行 Final Reconcile | docs、contract、codegen、target tests、`make test`、`make build`、retrospective、bug-report 判断和 work-journal 均有证据 | 001 Phase 6 |

## 7 关联计划

- [001-implement-review-runway](./plans/001-implement-review-runway/plan.md)：历史 spec 实施与 L2 review 的执行跑道。

## 8 关联文档

- [product-scope](../product-scope/spec.md)
- [engineering-roadmap](../engineering-roadmap/spec.md)
- [docs/ui-design](../../ui-design/INDEX.md)
- [`ui-design/` 静态原型](../../../ui-design/)
- [plan-review skill](../../../.agent-skills/plan-review/SKILL.md)
- [implement skill](../../../.agent-skills/implement/SKILL.md)
- [plan-code-review skill](../../../.agent-skills/plan-code-review/SKILL.md)
- [tdd skill](../../../.agent-skills/tdd/SKILL.md)
