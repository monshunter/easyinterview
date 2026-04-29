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
        ├── checklist.md        # required
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
| backend | 后端 | handler/service/store/worker |
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
- 验收标准中的 `ID` 列是文档内的说明性编号；它不是 BDD 场景编号。

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
- **BDD 策略**: Feature plan requires BDD；涉及用户可感知 UI、API 行为、业务流程或端到端功能时，必须引用 `bdd-plan.md`、`bdd-checklist.md` 与主 checklist 的 `BDD-Gate:`；不适用时写明原因
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
- 关联 BDD-Gate / 场景验证全部通过

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
- [ ] 2.2 BDD-Gate: 验证 E2E.P0.001 通过
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
  subspec: ${subspec}
  name: ${NNN-plan}
  sequence: 1
  supersedes: []
  specVersion:
    from: null
    to: 1.0
spec:
  defaultTarget: frontend
  discovery:
    aliases:
      - ${subspec}
      - ${NNN-plan}
    keywords: []
  targets:
    frontend:
      plan: ./plan.md
      checklist: ./checklist.md
      spec: ../../spec.md
      bddPlan: ./bdd-plan.md
      bddChecklist: ./bdd-checklist.md
      discovery:
        packages: []
        uiRoutes: []
        apiNames: []
```

`context.yaml` 只承载稳定检索标识，不承载 `commands`、脚本名、Make target 或人工操作步骤。
当 plan 生成 `test-plan.md` / `test-checklist.md` 时同步写入 `testPlan` / `testChecklist`；当 plan 生成 `bdd-plan.md` / `bdd-checklist.md` 时必须同步写入 `bddPlan` / `bddChecklist`。没有生成对应文件时不得保留这些字段。

## 8 BDD 模板

### 7.1 `bdd-plan.md`

```markdown
# BDD Plan

> **版本**: 1.0
> **状态**: active
> **更新日期**: YYYY-MM-DD

## Phase 1: 阶段名称

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.001 | 场景名称 | 前置条件 | 触发动作 | 预期结果 | `test/scenarios/...` |
```

### 7.2 `bdd-checklist.md`

```markdown
# BDD Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: YYYY-MM-DD

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.001 场景名称

- [ ] 创建场景目录
- [ ] 准备测试数据
- [ ] 实现 setup / trigger / verify / cleanup
- [ ] 执行并通过场景验证
- [ ] 记录验证证据
```

主 `checklist.md` 只保留阶段级 `BDD-Gate` 汇总项；场景资产和执行状态只记录在 `bdd-checklist.md`。
