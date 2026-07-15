# UI Demo Pruning and Documentation-Owned Design

> **版本**: 2.0
> **状态**: active
> **更新日期**: 2026-07-15

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

删除 `ui-design/` 可运行前端 Demo 及所有只为 Demo 与正式前端双向同步而存在的代码、测试、脚本和合同；保留 `docs/ui-design/` 作为 UI 信息架构、页面流程、交互约束和设计决策的文档 owner。

正式 `frontend/` 不再执行源码复刻、像素对照或 Demo-first 流程。现有组件、样式和行为由对应 spec、`docs/ui-design/`、正式前端测试、构建和真实业务场景共同验证。

## 2 背景

原计划用 `ui-design/` golden preview 与 Playwright parity suite 约束正式前端。随着正式前端持续演进，Demo 已成为第二套需要同步校正的实现：它不承载真实数据、路由、鉴权或业务状态，却要求页面、fixture、样式和测试重复维护。

用户于 2026-07-15 确认：项目不再需要 UI Demo，也不再使用“UI 真理源”概念；`docs/ui-design/` 继续定义 UI 架构和流程。该变更属于原 003 计划合同的反向修订，因此在原目录原地实施，不创建 sibling plan。

## 3 质量门禁分类

- **Plan 类型**: `code-internal + tooling + docs`。
- **TDD 策略**: 适用。先用 `scripts/lint/ui_demo_pruning_test.py` 为 `scripts/lint/ui_demo_pruning.py` 与 `make lint-ui-demo-pruning` 建立 Red，再删除 Demo、parity/codegen 依赖并改写正式前端测试；每个阶段使用 focused test 验证，涉及 `frontend/` 的阶段收口执行根 `make test`。
- **BDD 策略**: `BDD-N/A`。本计划不新增或改变用户可感知的 UI、API 或业务流程，只删除重复实现和内部验证机制；不创建 `bdd-plan.md` / `bdd-checklist.md`，不分配 E2E ID。
- **替代验证 gate**: Demo 零目录/零 active-reference lint、正式前端 unit tests、`make test`、`make build`、`make docs-check`、`make codegen-check`、`git diff --check`。
- **Operation matrix**: `N/A`。不修改 OpenAPI operation、fixture wire contract、backend handler、persistence 或 AI dependency。

## 4 覆盖矩阵

| Source | 类别 | Plan phase | 验证 | 负向范围 |
|--------|------|------------|------|----------|
| 删除重复 Demo 实现 | Primary path | Phase 1 | `scripts/lint/ui_demo_pruning_test.py` + `make lint-ui-demo-pruning` | `ui-design/` 实体目录不得存在 |
| 保留 UI 架构与流程文档 | Cross-layer contract | Phase 2 | `make docs-check` + active-doc reference scan | `docs/ui-design/` 不得被误删；不得链接 Demo 源文件或运行入口 |
| 删除双源工具链 | Regression / non-current-negative | Phase 3 | Make/package/script tests + repo search | `test:pixel-parity`、`serve-pixel-parity`、prototype fixture sync、golden preview 不得残留为 active gate |
| 保留独立有价值的正式前端测试 | UX quality | Phase 4 | focused Vitest + 根 `make test` | 不以删除 Demo 为理由删除行为、a11y、responsive 或 route 正确性覆盖 |
| 清理治理和 owner 合同 | Cross-layer contract | Phase 5 | context validation + docs/index sync | 不再要求先做 Demo、源码复刻、像素 parity 或“UI 真理源” |
| 完整回归 | Failure / recovery | Phase 6 | `make test`、`make build`、`make docs-check`、`make codegen-check` | clean checkout 不依赖已删除目录、CDN 或 Demo server |

隐私、安全、持久化和 API 错误路径不适用：本计划不改变用户数据、鉴权、网络协议、后端或运行时业务流程；由现有 owner tests 保持回归。

## 5 实施步骤

### Phase 1: 建立删除合同并移除 Demo 实体

#### 1.1 建立 UI Demo 零残留断言

新增 `scripts/lint/ui_demo_pruning.py`、`scripts/lint/ui_demo_pruning_test.py` 和 `make lint-ui-demo-pruning`，断言 `ui-design/` 不存在，active code/docs 不再依赖 Demo 路径或 parity 入口；历史 work journal、Bug、report 和 history 文档作为事实证据允许保留旧表述。

#### 1.2 删除 `ui-design/`

删除静态入口、React/Babel 源、canvas、运行脚本、Demo fixture 与 Demo 自身合同测试。

### Phase 2: 让 `docs/ui-design/` 成为纯设计文档

#### 2.1 改写目录合同

保留 `docs/ui-design/` 的 README、INDEX、模板与模块文档，删除所有 Demo 运行、源文件锚点、hash 原型路由、源码复刻和 parity 前置要求；明确其只定义信息架构、页面流程、交互约束和设计决策。

#### 2.2 修订当前 UI 文档

将 prototype/formal parity 改为正式前端可执行的 component、responsive、accessibility、route、fixture 或真实场景验证；不改变本次范围外的产品行为。

### Phase 3: 删除双源工具链

#### 3.1 删除浏览器 parity 工具

删除只为挂载 Demo/正式前端对照而存在的 Playwright config、static server、pixel-parity specs、package scripts/dependencies 和 scaffold tests。

#### 3.2 删除 prototype fixture/codegen 依赖

删除从 Demo 同步 OpenAPI fixture 的脚本、测试、文档与 Make target；保留 `openapi/fixtures/` 作为 contract-owned fixture。

#### 3.3 修订根质量门禁

从 `make test`、lint 和辅助脚本中移除 Demo 合同测试及 Demo 扫描路径，接入 Phase 1 的零残留断言。

### Phase 4: 解耦正式前端测试与源码注释

#### 4.1 改写 source-traceability tests

删除读取 `ui-design/src/*.jsx` 字面量的断言；保留直接验证正式 token、DOM、control、route、responsive 与 accessibility 合同的测试。

#### 4.2 清理正式源码和 README

把正式组件、CSS、README 中的 Demo/source-level mirror 描述改为 `docs/ui-design/` 设计语义或当前组件合同；删除 Demo import 负向测试中已失去价值的专用字符串断言。

### Phase 5: 修订当前治理与 owner 文档

#### 5.1 修订全局流程合同

更新 `AGENTS.md`、`docs/development.md`、`docs/README.md`、`design` / `implement` / `plan-code-review` / `tdd` 等相关 skills，以及仍被当前 context 使用的 spec/plan/checklist/context，使 UI 设计先在 `docs/ui-design/` 收敛，正式前端直接实施；不再要求创建或读取 Demo 源。

#### 5.2 清理 active 引用

扫描非历史 active docs、代码、脚本、配置与测试，删除 `ui-design/`、pixel parity、source-level replication、golden preview 和“UI 真理源”残留；保留明确标记为历史事实的 work journal、Bug、report 与 history。

### Phase 6: 验证与生命周期收口

#### 6.1 执行完整验证

执行 Demo pruning lint、focused tests、`make test`、`make build`、`make docs-check`、`make codegen-check` 和 `git diff --check`。

#### 6.2 同步 owner 生命周期

确认所有 checklist 项有当前执行证据后，将 spec/plan/checklist 恢复为 `completed` 并同步 INDEX；执行 post-pass reconcile 与 retrospective。

## 6 验收标准

- `ui-design/` 实体目录不存在，clean checkout 的 test/build/docs/codegen gate 不依赖它。
- `docs/ui-design/` 保留，并只表达 UI 信息架构、流程、交互约束和设计决策。
- active 代码和文档不再定义“UI 真理源”、Demo-first、源码复刻、golden preview 或 pixel parity 合同。
- 只为 Demo 对照存在的 Playwright、fixture sync、scaffold 和 traceability 资产被删除。
- 正式前端的独立行为、响应式、可访问性、route 和业务状态回归测试继续通过。
- 历史 work journal、Bug、report 和 history 文档不被改写为当前事实，也不作为 active contract 使用。

## 7 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 误删仍有独立价值的浏览器/组件覆盖 | 按断言语义区分 Demo parity 与正式前端行为；能脱离 Demo 运行的覆盖优先迁移到 owner unit/browser test |
| 文档保留了不可执行的 Demo 链接 | 对 `docs/ui-design/` 和 active docs 执行零路径/零 parity 搜索与链接检查 |
| 根 gate 暗含已删除目录 | 在 clean checkout 运行 test/build/docs/codegen 全套 gate |
| 历史记录产生搜索噪声 | 零残留 lint 明确区分 active contract 与允许保留的历史证据目录 |

## 8 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-15 | 2.0 | 用户确认删除 `ui-design/` Demo，保留 `docs/ui-design/` 作为 UI 架构与流程设计文档；原 parity owner 原地重开为降熵删除计划。 |
| 2026-07-10 | 1.6 | 完成当时的 12-spec UI Demo pixel parity gate。 |
