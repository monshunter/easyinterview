# 计划文档模板

## 1 串行计划文档模板

```markdown
# 计划名称

> **版本**: 1.0
> **状态**: active
> **更新日期**: YYYY-MM-DD

**关联 Checklist**: [implementation-checklist](./implementation-checklist.md)

## 1 目标

描述本计划要达成的目标。

## 2 背景

说明为什么需要这个计划。

## 3 实施步骤

### Phase 1: 基础设施准备

#### 1.1 创建基础配置

具体步骤描述。

#### 1.2 校验默认值

具体步骤描述。

### Phase 2: 服务集成

#### 2.1 接入 Service 层

具体步骤描述。

## 4 验收标准

- 本计划列出的实现 / 单元测试项全部通过
- 关联的 BDD-Gate / 场景验证全部通过

## 5 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 风险 1 | 措施 1 |

## 6 关联文档

- [相关设计](../../spec/xxx-design.md)
```

## 2 串行 Checklist 模板

```markdown
# 计划名称 Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: YYYY-MM-DD

**关联计划**: [计划文档](./implementation.md)

## Phase 1: 基础设施准备

- [ ] 1.1 创建基础配置
- [ ] 1.2 校验默认值

## Phase 2: 服务集成

- [ ] 2.1 接入 Service 层
- [ ] 2.2 BDD-Gate: 验证 E2E.P0.001, E2E.P0.004 通过
```

## 3 测试 Checklist 模板

```markdown
# 单元测试 Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: YYYY-MM-DD

**关联计划**: [测试计划](./unit-test-plan.md)

## Phase 1: 登录页测试

- [ ] 1.1 渲染模式切换测试
- [ ] 1.2 本地登录表单校验

## Phase 2: 注册页面测试

- [ ] 2.1 Token 解析测试
- [ ] 2.2 密码校验测试
- [ ] 2.3 Phase 2 本计划定义的单元测试项全部通过
```

测试计划 / test checklist 写作约束：

- 完成条件必须绑定计划正文中列出的测试项或 BDD 场景
- 禁止创建 `coverage >= N%`、`覆盖率 ≥ N%`、`line coverage` 等百分比阈值型 Checklist 项
- 如需记录覆盖率，只能写在备注、报告或背景说明中，不能作为完成态、提交或 phase exit gate

## 4 BDD-Gate 示例

```markdown
- [ ] 1.3 BDD-Gate: 验证 E2E.P0.001, E2E.P0.004 通过
```

### 4.1 BDD Checklist 模板

`bdd-plan.md` 定义场景，不记录执行进度；`bdd-checklist.md` 记录每个场景的资产准备与执行状态。主 `implementation-checklist.md` 只保留阶段级 `BDD-Gate` 汇总项。

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

## 5 `context.yaml` 最小模板

```yaml
apiVersion: plancontext.agent.dev/v1alpha1
kind: PlanContext
metadata:
  name: ${subject}
spec:
  defaultTarget: backend
  discovery:
    aliases:
      - ${subject}
    keywords:
      - TODO: add user-facing issue keywords
  targets:
    backend:
      plan: ./implementation.md
      checklist: ./implementation-checklist.md
      spec: ../../spec/${subject}-design.md
      bddPlan: ./bdd-plan.md                # 可选：BDD 场景定义
      bddChecklist: ./bdd-checklist.md      # 可选：BDD 资产与执行清单
      discovery:
        packages:
          - TODO: add main packages or modules
```

`context.yaml` 只承载稳定检索标识，不承载 `commands`。如有需要，可补充 `uiRoutes` / `apiNames`，但具体执行命令必须写回 README 或场景文档。

### 5.1 完整 target 模板（含测试与 BDD 可选字段）

target 字段键名统一采用 camelCase。`testPlan` / `testChecklist` / `bddPlan` / `references` 均为可选；未使用时直接省略。

```yaml
apiVersion: plancontext.agent.dev/v1alpha1
kind: PlanContext
metadata:
  name: ${subject}
spec:
  defaultTarget: backend
  discovery:
    aliases:
      - ${subject}
    keywords:
      - TODO: add user-facing issue keywords
  targets:
    backend:
      plan: ./implementation.md
      checklist: ./implementation-checklist.md
      spec: ../../spec/${subject}-design.md
      testPlan: ./test-plan.md              # 可选：测试计划
      testChecklist: ./test-checklist.md    # 可选：/tdd --test-checklist 入参
      bddPlan: ./bdd-plan.md                # 可选：/tdd BDD-Gate 场景源
      bddChecklist: ./bdd-checklist.md      # 可选：BDD 资产与执行清单
      references:                           # 可选：其他只读参考
        - ./release-notes.md
      discovery:
        packages:
          - TODO: add main packages or modules
```

| YAML 键 | 规范化 role | 消费者 |
|---------|------------|--------|
| `plan` | `plan` | review + execution |
| `checklist` | `checklist` | `/tdd --file` |
| `spec` | `spec` | review + execution references |
| `testPlan` | `test-plan` | `/tdd --references` |
| `testChecklist` | `test-checklist` | `/tdd --test-checklist` |
| `bddPlan` | `bdd-plan` | `/tdd` BDD-Gate 参考 |
| `bddChecklist` | `bdd-checklist` | `/tdd` BDD-Gate 前置参考 |
| `references[]` | `reference` | review + execution references |

> YAML 字段键名统一 camelCase；规范化输出 role 名称沿用 kebab-case，两者不可混用。详见 `.agent-skills/implement/shared/references/plan-context-contract.md`。

## 6 修订记录模板

```markdown
## 修订记录

- [2026-04-15] v1.1 — 补齐 README 协作护栏并同步模板引用
```
