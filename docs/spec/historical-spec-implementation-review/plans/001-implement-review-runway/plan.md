# Historical Spec Implement Review Runway

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-05-04

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

建立一条可在新对话中直接执行的历史 spec 实施与 plan-code-review 跑道：先清点当前真实存在的 spec / plan / checklist / context，再对文档漂移执行 L1 修复，然后按依赖顺序通过 `/implement` 进入原 plan 的 `/tdd`，每个完成 plan 再执行 `/plan-code-review --fix`，最终以全局 docs、contract、codegen、test、build、retrospective、bug-report 判断和 work-journal 收口。

本计划只定义执行编排和证据要求，不替代任何原始 subject 的具体实现计划。

## 2 背景

当前仓库已经完成 product-scope、engineering-roadmap、OpenAPI、shared conventions、events/jobs、migrations、config、AI gateway、observability、prompt/rubric 和 UI design 的多轮对齐。后续仍需要对历史 active spec 重新执行 `/implement` 与 `/plan-code-review --fix`，确保旧计划和已落地代码事实继续与最新产品 spec、工程 spec、UI 设计和用户交互逻辑一致。

本 plan 的校正参考资料不是单一 roadmap。每个 historical spec / plan 进入 L1 或 L2 前，都必须显式对照：

- [product-scope](../../../product-scope/spec.md)：产品范围、丢弃规则、阶段边界和用户价值承载方式。
- [engineering-roadmap](../../../engineering-roadmap/spec.md)：active spec 地图、P0 workstream 候选、on-demand child 创建规则和并行边界。
- [docs/ui-design](../../../../ui-design/INDEX.md)：当前 UI 模块、用户流程、交互契约和删除模块边界。
- `ui-design/` 静态原型：当前可视化交互事实、route / screen 实现和 UI contract tests。
- 编码 truth source：`openapi/`、`shared/`、`migrations/`、`config/` 与 generated outputs。

如果直接并行执行多个 `/implement` 或把 `/plan-code-review` 与实现同时运行，容易出现以下问题：

- 按旧 spec / plan 写代码，随后再由 L2 review 大面积返工。
- 多个任务同时修改 shared / OpenAPI / migrations / config / generated artifacts，互相覆盖 truth source。
- completed plan 中的旧验收口径被误认为当前执行要求。
- docs/spec/INDEX、plans/INDEX、work journal、retrospective 等全局文档由多个任务同时写入，产生合并漂移。

因此，本 plan 把“可并行的只读发现”和“必须串行的写入 / truth source 变更”明确拆开。

## 3 质量门禁分类

- **Plan 类型**: docs-only / governance-orchestration
- **TDD 策略**: 不适用：本 plan 不实现前端、后端、工具脚本、迁移、codegen 或测试辅助逻辑。后续具体 code plan 必须通过对应 subject 的 `/implement` -> `/tdd` 执行，并在原 checklist 中记录 Red / Green / Refactor 与测试证据。
- **BDD 策略**: 不适用：本 plan 不引入用户可感知 UI、API 行为、业务流程或端到端功能。若后续某 implementation plan 引入用户行为，该 plan 必须生成 `bdd-plan.md` / `bdd-checklist.md` 并在主 checklist 加 `BDD-Gate:`。
- **替代验证 gate**: `validate_context.py`、`/plan-review --fix`、`/plan-code-review --fix`、`sync-doc-index --check`、Markdown link check、docs/spec heading anchor audit、目标 contract gates、focused tests、`make test`、`make build`、retrospective、bug-report 判断、work-journal。

## 4 实施步骤

### Phase 1: Scope Inventory

#### 1.1 读取执行前置上下文

读取 `docs/work-journal/INDEX.md` 与最新工作日志，确认最近完成的 spec / plan / commit 和未完成事项。读取 `docs/spec/INDEX.md` 与所有 active subject 的 `plans/INDEX.md`，形成候选清单。

同时读取强制校正参考资料：

- `docs/spec/product-scope/spec.md`
- `docs/spec/engineering-roadmap/spec.md`
- `docs/ui-design/INDEX.md`
- 与候选 plan 相关的 `docs/ui-design/*.md`
- 与候选 plan 相关的 `ui-design/` 静态原型源码和 UI contract tests

#### 1.2 构建候选 plan 表

为每个候选 plan 记录：

- subject / plan id / status / version / target
- checklist 完成状态
- context.yaml 路径与 validation 结果
- 主要写入范围
- 依赖的 truth source
- 对应 product-scope / engineering-roadmap / docs-ui-design / ui-design 参考章节
- 是否 completed plan 需要 reopen

#### 1.3 分类并行安全级别

把候选 plan 分为：

- `read-only-parallel-safe`：只读审查和 context validation。
- `docs-write-serial`：会改 `docs/spec/INDEX.md`、plans/INDEX 或同一 subject 文档。
- `truth-source-serial`：会改 shared/OpenAPI/events/migrations/config/generated artifacts。
- `implementation-disjoint`：写入范围与其它 plan 不重叠，可在 owner 明确时并行。
- `implementation-serial`：默认串行，直到证明 disjoint。

### Phase 2: L1 Reconcile

#### 2.1 对候选 plan 执行 L1 文档审查

对每个目标 plan 先执行 `/plan-review` 或等价语义审查，确认 spec、plan、checklist、context 是否与当前 product-scope、engineering-roadmap、docs/ui-design、ui-design、OpenAPI、shared conventions、events/jobs、migrations、config 等事实一致。审查输出必须明确引用所使用的 product / roadmap / UI 参考章节，避免只按单一 roadmap 判断。

#### 2.2 对文档漂移执行原地修订

发现漂移时，执行 `/plan-review --fix`，并按原 subject 原地修订 spec / plan / checklist / context / INDEX。completed plan 若需要修订，先切回 active 或追加 remediation 项，验证后再恢复 completed。

#### 2.3 L1 验证

每个修订后的 subject 必须通过：

- context validation
- sync-doc-index
- Markdown link check
- 必要时执行 docs/spec heading anchor audit

### Phase 3: Implementation Runway

#### 3.1 按依赖顺序选择下一个 plan

默认顺序：

1. product-scope / engineering-roadmap 相关文档前置项。
2. shared conventions。
3. OpenAPI、events/jobs、DB migrations、config。
4. AI gateway、observability、prompt/rubric。
5. UI / frontend / backend domain。
6. mock / E2E / release gate。

执行期可根据实际 plan 依赖调整，但必须记录原因。

#### 3.2 通过 `/implement` 进入原 plan

每个 plan 的代码逻辑只能通过 `/implement` 解析 context，再由 `/tdd` 按 checklist phase 顺序执行。不得跳 phase，不得在 checklist 未同步时继续下一项。

#### 3.3 维护实现证据

每个 phase 完成后必须记录：

- Red / Green / Refactor 证据
- focused tests
- 受影响 generated artifacts / contract gates
- checklist 更新
- 若涉及用户行为，BDD-Gate 与场景证据

### Phase 4: L2 Plan Code Review

#### 4.1 对已完成范围执行 L2 审查

每个实现完成或 remediation 完成的 plan 必须执行 `/plan-code-review --fix`，审查范围至少包含已勾选 checklist phase 与对应代码 / 生成物 / 文档事实。

#### 4.2 通过原 checklist section 修复 finding

L2 finding 必须映射回具体 checklist section。可自动映射的 finding 通过 `/tdd --section` 修复；无法映射的 finding 降级为 preview-only，并要求用户确认 owner section。

#### 4.3 L2 验证

每个 L2 修复后必须运行 focused tests、相邻 regression tests 和受影响 contract gates。任何失败都阻止该 plan 进入完成态。

### Phase 5: Parallelism Gate

#### 5.1 并行只读发现

允许对多个 subject 并行执行 context validation、文本搜索、diff 分析和 finding 草稿。并行结果必须由当前 session owner 集成。

#### 5.2 写入并行审批

只有满足以下条件才允许并行写入：

- 写入路径完全 disjoint。
- 不修改同一 INDEX、同一 generated output 或同一 truth source。
- 每个 worker 有明确 path ownership。
- 当前 owner 在合并后执行全局验证。

#### 5.3 Truth source 串行执行

以下区域永远串行：

- `shared/conventions.yaml` 与 conventions generated outputs
- `openapi/openapi.yaml`、fixtures、baseline、OpenAPI generated outputs
- `shared/events.yaml`、`shared/jobs.yaml` 与 event/job generated outputs
- `migrations/` 与 migration lint truth source
- `config/` runtime config、feature flags、env dictionary
- 全局 `docs/spec/INDEX.md`、work journal、reports INDEX

### Phase 6: Final Reconcile

#### 6.1 全局验证

最终至少执行：

- `sync-doc-index --check`
- Markdown link check
- docs/spec heading anchor audit
- `make docs-check`
- `make codegen-check`
- 受影响 contract gates
- focused tests
- `make test`
- `make build`
- `git diff --check`

#### 6.2 收尾技能判断

成功交付后执行 `/retrospective` 判断是否需要报告；涉及 bugfix 或回归时执行 `/bug-report` 判断；需要提交时执行 `/work-journal`，按逻辑边界拆分 commit。

#### 6.3 交接新对话

输出下一轮可继续执行的 subject / plan / target / checklist 状态、剩余风险、验证证据和不适用说明。

## 5 验收标准

- Phase 1 输出的候选 plan 清单覆盖当前目标范围，并包含 context validation、状态、target、依赖、写入范围和并行安全分类。
- Phase 2 修复所有进入实施前的 L1 文档漂移，且 Header / INDEX / links / anchors 无漂移。
- Phase 3 的所有代码实现都由原 plan `/implement` -> `/tdd` 执行，checklist 与测试证据同步。
- Phase 4 的 L2 finding 均已修复、明确降级或交给用户确认 owner section。
- Phase 5 的并行执行记录能解释为什么并行安全，且没有 truth source 冲突。
- Phase 6 的全局验证、retrospective / bug-report 判断和 work-journal 证据齐全。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 历史 completed plan 被按旧口径直接实现 | Phase 2 先 L1 review；需要修订时原地 reopen/remediation |
| 多任务同时修改 generated truth source | Phase 5 把 truth source 区域列为串行 |
| L2 finding 无法映射 checklist section | 降级 preview-only，并要求用户确认 owner section |
| BDD 被全局 plan 漏掉 | 本 plan不生成全局 BDD；具体 feature plan 若引入用户行为，必须在原 plan 生成 BDD |
| 全局 INDEX / work journal 发生冲突 | 由当前 session owner 串行更新并最终验证 |
| 新对话丢失上下文 | Phase 6 输出可继续执行的 subject / plan / target / checklist 状态 |
