# 设计文档模板

## 1 模块设计模板

```markdown
# 模块名称 设计文档

> **版本**: 1.0
> **状态**: active
> **更新日期**: YYYY-MM-DD

## 1 概述

简要描述模块的目的和职责。

## 2 设计目标

- 目标 1
- 目标 2

## 3 用户决策 / 待确认事项

- 决策项 1：待用户确认的选择与影响
- 决策项 2：待确认的兼容策略或范围边界

## 4 架构设计

描述模块的整体架构，可包含图表。

## 5 接口定义

### 5.1 公开接口

描述对外暴露的接口。

### 5.2 内部接口

描述模块内部的关键接口。

## 6 数据结构

定义核心数据结构。

## 7 错误处理

描述错误处理策略。

## 8 验收标准

| ID | 场景 | Given | When | Then | Phase |
|----|------|-------|------|------|-------|
| C-1 | 场景描述 | 前置条件 | 触发动作 | 预期结果 | 对应 Phase |

## 9 关联文档

- [相关文档1](./path1.md)
- [相关文档2](./path2.md)
```

## 2 验收标准说明

- `用户决策 / 待确认事项` 为可选 section；无待确认事项时可省略
- 验收标准为可选 section；小型无取舍的文档可省略
- `ID` 列是文档内的说明性编号；它不是 BDD 场景编号
- 需要 BDD 闭环时，应在 `bdd-plan.md` 中分配场景编号，并遵守对应 `test/scenarios/<suite>/README.md` 与 `INDEX.md` 的编号规范；主 `checklist.md` 的 `BDD-Gate` 仅引用这些场景编号

## 3 Spec-Centric Plan 模板

以下模板用于 v2 结构：

```text
docs/spec/$subspec/
├── spec.md
├── history.md
└── plans/
    └── NNN-kebab-name/
        ├── context.yaml
        ├── plan.md
        ├── checklist.md
        ├── bdd-plan.md
        └── bdd-checklist.md
```

### 3.1 `context.yaml` 模板

```yaml
apiVersion: plancontext.agent.dev/v1alpha1
kind: PlanContext
metadata:
  subspec: $subspec
  name: $subplan
  sequence: 1
  supersedes: []
  specVersion:
    from: null
    to: 1.0
spec:
  defaultTarget: backend
  discovery:
    aliases:
      - $subspec
      - $subplan
    keywords: []
  targets:
    backend:
      plan: ./plan.md
      checklist: ./checklist.md
      spec: ../../spec.md
      bddPlan: ./bdd-plan.md
      bddChecklist: ./bdd-checklist.md
      discovery:
        packages: []
```

### 3.2 `bdd-plan.md` 模板

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

### 3.3 `bdd-checklist.md` 模板

`bdd-plan.md` 定义场景，不记录执行进度；`bdd-checklist.md` 记录场景资产准备与执行状态。主 `checklist.md` 只保留阶段级 `BDD-Gate` 汇总项。

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

### 3.4 主 `checklist.md` 的 BDD-Gate 示例

```markdown
- [ ] 2.4 BDD-Gate: 验证 E2E.P0.001, E2E.P0.002 通过
```

`BDD-Gate` 通过前，相关 `bdd-checklist.md` 场景项必须全部完成；两者不能表达不同的通过结论。
