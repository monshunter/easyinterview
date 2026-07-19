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
            ├── context.yaml        # required
            ├── plan.md             # required
            ├── checklist.md        # required
            ├── test-plan.md        # conditional
            ├── test-checklist.md   # conditional
            ├── bdd-plan.md         # conditional
            └── bdd-checklist.md    # conditional
```

`spec.md` 是该 subject 的设计真理源；`plans/` 下每个目录是一个边界清晰、可执行的交付计划。

## 2 核心规则

1. **Spec owns plan**：计划必须挂在对应 `docs/spec/${subspec}/plans/` 下。
2. **边界清晰**：一个 plan 只覆盖一个可独立实施/验证的目标。
3. **成对创建**：每个 `plan.md` 必须有对应 `checklist.md`。
4. **Context 必填**：每个 plan 目录必须有 `context.yaml`，作为 `/implement`、`/plan-review`、`/plan-code-review` 的入口。
   过渡期 schema 必须精确最小化：`metadata` 只含 `name`，`spec` 只含 `defaultTarget` / `targets`，target 只含 `plan` / `checklist` / `spec` 与可选的一等 test/BDD 文档链接；禁止 discovery、references、分支/版本提示和自定义字段。
5. **局部计划索引**：每个 `plans/` 目录只保留 `INDEX.md`；规则与模板统一使用 `docs/spec/README.md` 和 `docs/spec/TEMPLATES.md`。
6. **索引投影**：`docs/spec/INDEX.md` 和 `docs/spec/${subspec}/plans/INDEX.md` 只反映 Header，不作为状态真理源。
7. **原地修订**：同主题后续修订优先原地更新原 spec 和原 plan，不创建同主题 sibling bugfix/follow-up。
8. **Code plan requires TDD**：涉及前端 / 后端 / 工具脚本 / 迁移 / codegen / 测试辅助等代码逻辑的 plan，必须写明 TDD 策略并通过 `/implement` → `/tdd` 执行。
9. **Feature plan requires BDD**：引入用户可感知 UI、API 行为或业务流程的 plan，必须包含 `bdd-plan.md`、`bdd-checklist.md` 与主 `checklist.md` 中的 `BDD-Gate:` 项；Behavior ID 可以由 code-level domain behavior test 验证，不要求一一创建 E2E。
10. **真实 E2E 单独判定**：只有通过真实 HTTP API 或浏览器访问已运行 frontend，且业务请求落到真实 backend 的流程，才能分配 `E2E.*` ID 和创建 `test/scenarios/e2e/` 资产；Go/Vitest/pytest/lint/build wrapper 不是 E2E。
11. **BDD 不适用需说明**：纯配置、内部契约、工具、迁移、codegen、lint、fixture 或 build 若不产生用户行为流，必须写明 `BDD-N/A + 替代验证 gate`，不得创建 BDD 文件或在 `context.yaml` 保留 `bddPlan` / `bddChecklist`。

## 3 文档元信息

所有 `spec.md`、`history.md`、`plan.md`、`checklist.md`、`test-*.md`、`bdd-*.md` 必须在头部包含：

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
| `completed` | 已完成，作为交付记录保留 |

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

## 6 质量门禁分类

每个 plan 的 `plan.md` 必须包含“质量门禁分类”，明确以下字段：

- **Plan 类型**：`docs-only` / `code-internal` / `feature-behavior` / `contract` / `migration` / `tooling` 等，可组合。
- **TDD 策略**：涉及代码逻辑时必须说明 Red-Green-Refactor 入口、测试文件或测试命令；纯文档计划写 `不适用：docs-only`。
- **BDD 策略**：涉及用户可感知 UI、API 行为或业务流程时必须引用 `bdd-plan.md` / `bdd-checklist.md` 与 `BDD-Gate:`；说明验证入口是 domain behavior test 还是真实 API/UI E2E。内部计划写 `BDD-N/A`。
- **替代验证 gate**：BDD 不适用的内部计划必须列出可执行替代 gate，如 contract test、lint、drift check、migration check、smoke。

## 7 检查清单

创建或修改 spec/plan 后确认：

- [ ] 文件位于 `docs/spec/${subspec}/...`
- [ ] Header 字段完整且顺序正确
- [ ] `plan.md` 与 `checklist.md` 成对存在
- [ ] `context.yaml` target 路径全部有效
- [ ] 代码逻辑计划已写明 TDD 策略；用户行为功能计划已具备 BDD 文件、Behavior ID / 真实 E2E ID 与 `BDD-Gate:`
- [ ] Behavior ID 已选择 domain behavior test 或真实 API/UI E2E，且只有后者创建 E2E 资产
- [ ] BDD 不适用的内部/配置/tooling 计划已写明 `BDD-N/A` 和替代验证 gate，且没有 BDD 文件/context 字段
- [ ] `docs/spec/INDEX.md` 已同步

## 8 协作约定

协作前必须先阅读本目录 `README.md` 的命名与状态规则。
起草或修改正文时，必须参考同目录 `TEMPLATES.md`。
创建或修改 plan 时，不得在 `docs/spec/${subspec}/plans/` 下复制 `README.md` 或 `TEMPLATES.md`；统一参考本目录 `README.md` 与 `TEMPLATES.md`。
不得把 `README.md` 当作可复制模板。
