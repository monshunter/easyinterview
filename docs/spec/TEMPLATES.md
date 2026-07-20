# Spec-Centric 文档模板

## 1 Subject 目录模板

```text
docs/spec/${subspec}/
├── spec.md
├── history.md
└── plans/
    ├── INDEX.md
    └── ${NNN-plan}/
        ├── context.yaml        # required
        ├── plan.md             # required
        ├── checklist.md        # required unless plan.md owns inline progress
        ├── test-plan.md        # conditional
        ├── test-checklist.md   # conditional
        ├── bdd-plan.md         # conditional
        └── bdd-checklist.md    # conditional
```

## 2 `spec.md` 模板

```markdown
# ${Subject} Spec

> **版本**: 1.0
> **状态**: active
> **更新日期**: YYYY-MM-DD

## 1 背景与目标

说明 subject 的用户目标、业务边界和本阶段价值。

## 2 范围

### 2.1 In Scope

- 范围项

### 2.2 Out of Scope

- 非目标项

## 3 用户决策 / 待确认事项

- 决策项：影响与取舍

## 4 设计约束

- 产品约束
- 技术约束
- UI / API / 数据约束

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| frontend | 前端 | 页面、状态、mock flow |
| backend | 后端 | handler/service/store/background runner |
| contract | 契约 | OpenAPI、fixtures、schema |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 场景描述 | 前置条件 | 触发动作 | 预期结果 | 001-frontend |

## 7 关联计划

- [001-frontend](./plans/001-frontend/plan.md)
```

模板说明：

- `用户决策 / 待确认事项` 无待确认事项时可省略。
- 验收标准中的 `ID` 列是文档内的说明性编号；它不是 BDD 场景编号。同一 `spec.md` 内的完整 `D-*` / `C-*` ID 必须唯一，不同 subject 可独立复用编号；`make docs-check` 自动拒绝新增重复，禁止扩大 legacy baseline。

## 3 `history.md` 模板

```markdown
# ${Subject} History

> **版本**: 1.0
> **状态**: active
> **更新日期**: YYYY-MM-DD

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| YYYY-MM-DD | 1.0 | 初始创建 | ${NNN-plan} |
```

## 4 `plan.md` 模板

`plans/INDEX.md` 使用 `/init-docs` 的 `subspec-plans` scaffold 生成，作用域只限当前 `subspec`。plan / checklist / context / BDD 模板集中维护在本文，禁止复制到每个 `plans/` 目录。

```markdown
# ${Plan Name}

> **版本**: 1.0
> **状态**: active
> **更新日期**: YYYY-MM-DD

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

描述本计划要达成的独立交付目标。

## 2 背景

说明为什么需要这个计划，以及它与 subject spec 的关系。

## 3 质量门禁分类

- **Plan 类型**: docs-only / code-internal / feature-behavior / contract / migration / tooling
- **TDD 策略**: Code plan requires TDD；涉及代码逻辑时写明 Red-Green-Refactor 入口、测试文件/命令与每个 checklist item 的测试断言来源；纯文档计划写 `不适用：docs-only`
- **BDD 策略**: Feature plan requires BDD；涉及用户可感知 UI、API 行为或业务流程时，引用 `bdd-plan.md`、`bdd-checklist.md` 与主 checklist 的 `BDD-Gate:`。BDD 描述行为，不强制一一对应 E2E；验证入口可以是 domain behavior test，也可以是真实 API/UI 流程。纯内部计划写明 `BDD-N/A`，且不生成 BDD 文件
- **替代验证 gate**: BDD 不适用的内部计划必须列出 contract test、lint、drift check、migration check、smoke 或等价可执行 gate

## 4 实施步骤

### Phase 1: 基础准备

#### 1.1 任务名称

具体步骤描述。

### Phase 2: 验证收口

#### 2.1 任务名称

具体步骤描述。

## 5 验收标准

- 本计划列出的实现 / 测试项全部通过
- 适用时，关联 Behavior Gate / 真实 E2E 验证全部通过

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 风险 1 | 措施 1 |
```

## 5 `checklist.md` 模板

```markdown
# ${Plan Name} Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: YYYY-MM-DD

**关联计划**: [plan](./plan.md)

## Phase 1: 基础准备

- [ ] 1.1 任务名称

## Phase 2: 验证收口

- [ ] 2.1 任务名称
- [ ] 2.2 BDD-Gate: 验证 `${Behavior ID 或真实 E2E ID}` 通过
```

## 6 `test-plan.md` / `test-checklist.md` 模板

`test-plan.md` / `test-checklist.md` 仅当测试计划足够独立或需要跨 phase 映射时创建。普通代码 plan 仍必须在主 checklist item 中保留可执行测试断言。

```markdown
# Test Plan

> **版本**: 1.0
> **状态**: active
> **更新日期**: YYYY-MM-DD

## Phase 1: 对应实现阶段

- 测试目标、测试文件、测试命令、预期 Red/Green 证据。
```

```markdown
# Test Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: YYYY-MM-DD

**关联 Test Plan**: [test-plan](./test-plan.md)

## Phase 1: 对应实现阶段

- [ ] Phase 1 本计划定义的单元测试项全部通过
```

## 7 `context.yaml` 模板

```yaml
apiVersion: plancontext.agent.dev/v1alpha1
kind: PlanContext
metadata:
  name: ${NNN-plan}
spec:
  defaultTarget: frontend
  targets:
    frontend:
      plan: ./plan.md
      checklist: ./checklist.md
      spec: ../../spec.md
```

`context.yaml` 只承载目标与文档链接。`metadata` 只允许 `name`；禁止顶层/target `discovery`、target `references`、分支/版本提示、`commands` 或其他自定义字段。
当 plan 生成 `test-plan.md` / `test-checklist.md` 时同步写入 `testPlan` / `testChecklist`。只有实际生成 `bdd-plan.md` / `bdd-checklist.md` 时，才在对应 target 增加 `bddPlan: ./bdd-plan.md` 与 `bddChecklist: ./bdd-checklist.md`；纯内部 `BDD-N/A` 计划不得生成文件或保留字段。

single-plan/loop 模式不创建重复 `checklist.md`，由 `plan.md` 行内 checkbox 持有进度，并使用以下 target：

```yaml
spec:
  defaultTarget: harness
  targets:
    harness:
      plan: ./plan.md
      checklist: ./plan.md
      spec: ../../spec.md
```

consumer 保留 `plan` 与 `checklist` 两个角色，但读取正文时必须按解析后的绝对路径去重。

## 8 BDD 模板

BDD 文件描述可感知行为及其证据边界，不要求每个 Behavior ID 都创建 E2E 目录。domain behavior test 可以作为验证入口；只有通过已运行 frontend/backend 的真实 HTTP API 或浏览器 UI 驱动业务流程时，才创建 `test/scenarios/e2e/` 场景目录。纯内部、配置、lint、migration、codegen 或 tooling 计划应在主 plan/checklist 标记 `BDD-N/A`，不生成以下文件。

### 8.1 `bdd-plan.md`

```markdown
# BDD Plan

> **版本**: 1.0
> **状态**: active
> **更新日期**: YYYY-MM-DD

## Phase 1: 阶段名称

| Behavior ID / 真实 E2E ID | 行为 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| `${Behavior ID 或真实 E2E ID}` | 行为名称 | 前置条件 | 触发动作 | 预期结果 | `${domain behavior test 或真实 API/UI 场景目录}` |
```

### 8.2 `bdd-checklist.md`

```markdown
# BDD Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: YYYY-MM-DD

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## `${Behavior ID 或真实 E2E ID}` 行为名称

- [ ] 确认验证入口属于 domain behavior test 或真实 API/UI E2E
- [ ] 若为 domain behavior test：执行 owner test 并记录行为断言
- [ ] 若为真实 E2E：创建 `test/scenarios/e2e/` 目录，准备数据并实现 setup / trigger / verify / cleanup
- [ ] 执行并通过对应行为验证
- [ ] 记录验证证据
```

主 `checklist.md` 只保留阶段级 `BDD-Gate` 汇总项；行为证据和执行状态记录在 `bdd-checklist.md`。只有真实 E2E 才维护场景目录与其脚本资产。
