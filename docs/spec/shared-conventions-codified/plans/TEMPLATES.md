# Subspec Plans 模板

## 1 `plan.md` 模板

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

## 2 `checklist.md` 模板

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

## 3 `context.yaml` 模板

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

## 4 BDD 模板

### 4.1 `bdd-plan.md`

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

### 4.2 `bdd-checklist.md`

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
