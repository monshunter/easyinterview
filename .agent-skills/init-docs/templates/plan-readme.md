# 计划文档规范

## 1 目录结构

```
docs/plan/
├── README.md                              # 本文件，规范说明
├── TEMPLATES.md                           # 模板资产
├── INDEX.md                               # 文档索引
└── ${subject}/                            # 计划目录
    ├── context.yaml                       # 上下文声明（必填，计划上下文契约）
    ├── implementation.md                  # 实施计划
    ├── implementation-checklist.md        # 实施 Checklist
    ├── ${type}-plan.md                    # 其他计划（如测试计划）
    └── ${type}-plan-checklist.md          # 对应 Checklist
```

**示例**：
```
docs/plan/
├── target-workspace/
│   ├── context.yaml
│   ├── implementation.md
│   ├── implementation-checklist.md
│   ├── unit-test-plan.md
│   └── unit-test-plan-checklist.md
└── ...
```

## 2 核心规则

1. **成对创建**：每个计划文档必须有对应的 Checklist
2. **版本一致**：计划和 Checklist 版本号必须一致
3. **唯一真理**：Checklist 是任务完成状态的唯一真理来源
4. **原子更新**：修改计划内容时，必须同步更新 Checklist
5. **串行默认**：新计划默认采用顺序 phase 闭环
6. **上下文声明**：每个计划目录必须包含 `context.yaml`
7. **模板分离**：规范写在 README，复制示例放在 `TEMPLATES.md`
8. **测试完成以计划测试项为准**：禁止把代码覆盖率百分比写成 Checklist、提交或 phase exit 的硬门槛

## 3 文档元信息

所有计划文档必须在头部包含元信息：

```markdown
> **版本**: 1.0
> **状态**: active
> **更新日期**: YYYY-MM-DD
```

**状态值说明**：

| 状态 | 含义 |
|------|------|
| `draft` | 草稿，尚未正式生效 |
| `active` | 生效中，正在执行 |
| `completed` | 已完成 |
| `superseded` | 已被取代，需注明新文档路径 |
| `deprecated` | 已废弃，不再适用 |

## 4 命名规范

| 类型 | 计划文档 | Checklist |
|------|----------|-----------|
| 实施计划 | `implementation.md` | `implementation-checklist.md` |
| 测试计划 | `unit-test-plan.md` / `bdd-test-plan.md` / `bdd-plan.md` | 对应 `*-checklist.md` |
| 其他计划 | `${type}-plan.md` | `${type}-plan-checklist.md` |

## 5 写作约定

- Phase 标题格式：`### Phase {序号}: 描述`
- 任务标题格式：`#### {Phase序号}.{任务序号} 描述`
- Checklist section 标题格式：`## Phase {序号}: 描述`
- Checklist 任务 ID 必须与 plan task ID 一致（如 `1.1`、`2.1`）
- `BDD-Gate:` item 使用目标测试套件的场景编号，如 `E2E.P0.001`
- test checklist 默认通过 section 标题直接映射 implementation phase
- 测试计划 / test checklist 的完成条件必须写成“本计划定义的单元测试项全部通过”这类可执行验证；禁止写 `coverage >= 75%`、`覆盖率 ≥ 80%` 一类硬阈值
- 如需记录覆盖率，只能作为观测信息或背景数据，不能作为完成态、提交或 phase exit 的判定条件
- 完整模板与 `context.yaml` 示例位于 [TEMPLATES.md](./TEMPLATES.md)

## 5.1 协作约定

协作前必须先阅读本目录 `README.md` 的生命周期与命名规则。
起草或修改正文时，必须参考同目录 `TEMPLATES.md`。
不得把 `README.md` 当作可复制模板。

## 6 上下文声明与原地修订约定

每个计划目录必须包含 `context.yaml`，声明该计划的可执行目标、文档关联和检索索引元数据。

- `/implement`、`/plan-review`、`/plan-code-review` 读取 `context.yaml` 的执行字段
- `/change-intake` 读取 `context.yaml` 的 discovery 字段做问题匹配
- `context.yaml` 是 plan 上下文契约的单一真理源，但不是执行命令说明书

### 6.1 关键字段说明

- `apiVersion`：固定为 `plancontext.agent.dev/v1alpha1`
- `spec.defaultTarget`：`/implement` 默认执行目标
- `spec.discovery.aliases/keywords`：`/change-intake` 候选计划匹配信号
- `spec.targets.<name>.plan/checklist/spec`：该 target 的核心文档
- `spec.targets.<name>.testPlan`：可选，映射测试计划文档（`/tdd --references`）
- `spec.targets.<name>.testChecklist`：可选，映射测试 Checklist（`/tdd --test-checklist`）
- `spec.targets.<name>.bddPlan`：可选，BDD-Gate 场景源（`/tdd` BDD-Gate 参考）
- `spec.targets.<name>.bddChecklist`：可选，BDD 场景资产与执行清单（`/tdd` BDD-Gate 前置参考）
- `spec.targets.<name>.references`：可选字符串数组，其他只读参考文档
- `spec.targets.<name>.discovery.packages/uiRoutes/apiNames`：当前 target 的稳定检索标识

> YAML 字段键名统一采用 camelCase；规范化输出的 role 名称沿用 kebab-case（`test-plan` / `test-checklist` / `bdd-plan` / `bdd-checklist` / `reference`），两者不可混用。完整示例见 [TEMPLATES.md §5.1](./TEMPLATES.md)。

### 6.2 Discovery 字段使用约定

- 顶层仅允许：`aliases`、`keywords`、`relatedSpecs`、`relatedBugs`
- target 层仅允许：`packages`、`uiRoutes`、`apiNames`
- `aliases`：简短名、模块名、常见缩写
- `keywords`：现象词、路由、API 名、关键字段
- `packages`：写当前 target 最相关的代码/文档路径，不做全仓库堆砌
- `uiRoutes` / `apiNames`：仅写稳定的用户入口或接口标识
- 禁止写入 `commands`、脚本名、Make target 或人工操作步骤；具体命令必须写在对应 README / INDEX / 场景文档中

### 6.3 Plan 生命周期说明

| 状态 | 语义 | 允许动作 |
|------|------|----------|
| `draft` | 草稿，尚未进入执行 | 可补齐 spec/plan/checklist/context |
| `active` | 当前执行中的计划 | 可由 `/implement` / `/tdd` 按 checklist 推进 |
| `completed` | 同主题的最近完成基线 | 可在后续修订前递增版本并切回 `active` |
| `superseded` | 已被新计划取代 | 应指向后续计划 |
| `deprecated` | 已废弃 | 仅保留历史信息 |

对于 `completed` 计划，同主题后续修订应直接在原计划目录内更新 spec / plan / checklist。
需要继续执行时，先递增原 plan/checklist 版本并将 Header `状态` 调整回 `active`；验证完成后再恢复 `completed`。修订记录模板见 [TEMPLATES.md](./TEMPLATES.md)。

## 7 Checklist

创建或修改计划文档后，确认：

- [ ] 计划与 checklist 成对存在
- [ ] Header 字段完整且顺序正确
- [ ] checklist 与 plan 版本号保持一致
- [ ] 已更新 `INDEX.md` 索引
- [ ] `context.yaml` 路径和 target 均有效
