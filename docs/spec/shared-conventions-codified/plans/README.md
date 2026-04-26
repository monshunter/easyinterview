# Subspec Plans 规范

## 1 目录定位

本目录属于单个 `subspec`，只管理该 subject 下的可执行计划。

```text
docs/spec/${subspec}/plans/
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

## 2 核心规则

1. **局部索引**：`INDEX.md` 只索引当前 `subspec` 的计划。
2. **计划成对**：每个 `plan.md` 必须有对应 `checklist.md`。
3. **Context 必填**：每个计划目录必须包含 `context.yaml`。
4. **顺序执行**：新计划默认按 checklist 顺序串行推进。
5. **不建顶层 plan**：不得创建 `docs/plan/`。

## 3 命名规范

| 对象 | 命名 |
|------|------|
| Plan 目录 | `${NNN-plan}`，三位序号 + kebab-case |
| 主计划 | `plan.md` |
| 主 Checklist | `checklist.md` |
| Context | `context.yaml` |
| BDD | `bdd-plan.md` / `bdd-checklist.md` |

## 4 生命周期

| 状态 | 含义 |
|------|------|
| `draft` | 草稿，尚未进入执行 |
| `active` | 当前执行中 |
| `completed` | 已完成，作为最近完成基线 |
| `superseded` | 已被后续计划取代 |
| `deprecated` | 已废弃 |

同主题后续修订优先原地更新原 spec 和原 plan。若已完成计划需要继续执行，先递增 plan/checklist 版本并将状态改回 `active`；验证完成后再恢复 `completed`。

## 5 检查清单

- [ ] 每个计划目录都有 `context.yaml`
- [ ] `plan.md` 与 `checklist.md` Header 版本一致
- [ ] `INDEX.md` 链接指向本目录下的计划
- [ ] `context.yaml` 可被 shared validator 校验

## 6 协作约定

协作前必须先阅读本目录 `README.md` 的命名与状态规则。
起草或修改正文时，必须参考同目录 `TEMPLATES.md`。
不得把 `README.md` 当作可复制模板。
