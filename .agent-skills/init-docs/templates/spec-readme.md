# Spec-Centric 文档规范

## 1 目录结构

```text
docs/spec/
├── README.md
├── TEMPLATES.md
├── INDEX.md
└── ${subspec}/
    ├── spec.md
    ├── history.md
    └── plans/
        ├── INDEX.md
        └── ${NNN-plan}/
            ├── context.yaml
            ├── plan.md
            ├── checklist.md
            ├── bdd-plan.md
            └── bdd-checklist.md
```

`spec.md` 是该 subject 的设计真理源；`plans/` 下每个目录是一个边界清晰、可执行的交付计划。

## 2 核心规则

1. **Spec owns plan**：计划必须挂在对应 `docs/spec/${subspec}/plans/` 下。
2. **边界清晰**：一个 plan 只覆盖一个可独立实施/验证的目标。
3. **成对创建**：每个 `plan.md` 必须有对应 `checklist.md`。
4. **Context 必填**：每个 plan 目录必须有 `context.yaml`，作为 `/implement`、`/plan-review`、`/plan-code-review` 的入口。
5. **局部计划索引**：每个 `plans/` 目录只保留 `INDEX.md`；规则与模板统一使用 `docs/spec/README.md` 和 `docs/spec/TEMPLATES.md`。
6. **索引投影**：`docs/spec/INDEX.md` 和 `docs/spec/${subspec}/plans/INDEX.md` 只反映 Header，不作为状态真理源。
7. **原地修订**：同主题后续修订优先原地更新原 spec 和原 plan，不创建同主题 sibling bugfix/follow-up。

## 3 文档元信息

所有 `spec.md`、`history.md`、`plan.md`、`checklist.md`、`bdd-*.md` 必须在头部包含：

```markdown
> **版本**: 1.0
> **状态**: active
> **更新日期**: YYYY-MM-DD
```

状态值：

| 状态 | 含义 |
|------|------|
| `draft` | 草稿，尚未正式生效 |
| `active` | 生效中或正在执行 |
| `completed` | 已完成，作为历史交付记录保留 |
| `superseded` | 已被取代，需注明新文档路径 |
| `deprecated` | 已废弃 |

## 4 命名规范

| 对象 | 命名模式 | 示例 |
|------|----------|------|
| Subspec | `${subspec}` kebab-case | `target-job-workspace` |
| Spec | `spec.md` | `docs/spec/target-job-workspace/spec.md` |
| History | `history.md` | `docs/spec/target-job-workspace/history.md` |
| Plan | `${NNN-plan}` | `001-frontend` |
| Plan 文件 | `plan.md` / `checklist.md` | `plans/001-frontend/plan.md` |

## 5 Plan 目标拆分

推荐按交付边界拆分，而不是按旧 target 文件堆叠：

- `001-foundation`
- `002-api-contract`
- `003-backend`
- `004-frontend`
- `005-mock-contract`
- `006-integration`
- `007-unit-test`（仅当测试计划足够独立时单独拆）

具体编号由 subject 内部递增，不能跨 subject 复用同一 plan 目录。

## 6 检查清单

创建或修改 spec/plan 后确认：

- [ ] 文件位于 `docs/spec/${subspec}/...`
- [ ] Header 字段完整且顺序正确
- [ ] `plan.md` 与 `checklist.md` 成对存在
- [ ] `context.yaml` target 路径全部有效
- [ ] `docs/spec/INDEX.md` 已同步

## 7 协作约定

协作前必须先阅读本目录 `README.md` 的命名与状态规则。
起草或修改正文时，必须参考同目录 `TEMPLATES.md`。
创建或修改 plan 时，不得在 `docs/spec/${subspec}/plans/` 下复制 `README.md` 或 `TEMPLATES.md`；统一参考本目录 `README.md` 与 `TEMPLATES.md`。
不得把 `README.md` 当作可复制模板。
