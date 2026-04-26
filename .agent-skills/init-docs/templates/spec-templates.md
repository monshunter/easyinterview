# Spec-Centric 文档模板

## 1 Subject 目录模板

```text
docs/spec/${subspec}/
├── spec.md
├── history.md
└── plans/
    ├── README.md
    ├── TEMPLATES.md
    ├── INDEX.md
    └── ${NNN-plan}/
        ├── context.yaml
        ├── plan.md
        ├── checklist.md
        ├── bdd-plan.md
        └── bdd-checklist.md
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

`plans/README.md`、`plans/TEMPLATES.md`、`plans/INDEX.md` 使用 `/init-docs` 的 `subspec-plans-*` 模板生成，作用域只限当前 `subspec`。

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

## 3 实施步骤

### Phase 1: 基础准备

#### 1.1 任务名称

具体步骤描述。

### Phase 2: 验证收口

#### 2.1 任务名称

具体步骤描述。

## 4 验收标准

- 本计划列出的实现 / 测试项全部通过
- 关联 BDD-Gate / 场景验证全部通过

## 5 风险与应对

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

## 6 `context.yaml` 模板

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

## 7 BDD 模板

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
