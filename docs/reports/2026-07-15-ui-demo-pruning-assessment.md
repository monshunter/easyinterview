# UI Demo Pruning 交付复盘报告

> **日期**: 2026-07-15
> **审查人**: Codex

**关联计划**: [frontend-shell/003-ui-design-responsive-browser-gate](../spec/frontend-shell/plans/003-ui-design-responsive-browser-gate/plan.md)

**关联 Bug**: [BUG-0172](../bugs/BUG-0172.md)

## 1 复盘范围与成功证据

本次交付删除 `ui-design/` 可运行 Demo 及其 Playwright 对照、fixture 同步、根级 contract、源码追溯与 Demo-only hash route 适配；保留 `docs/ui-design/` 作为 UI 信息架构、页面流程、交互约束、响应式要求和设计决策 owner，并把正式实现与验证收敛到 `frontend/`。

成功证据：

- `make lint-ui-demo-pruning`：`ui_demo_directory: absent`，`active_residuals (0)`。
- `make test`：544 个 Python 测试、4241 个子断言、全部 Go package 与 123 个 frontend test files / 969 tests 通过。
- `make build`：Go build 与 frontend TypeScript/Vite production build 通过。
- `make docs-check`：Header、INDEX、孤儿与相对链接均零漂移。
- `make codegen-check`：conventions、events/jobs 与 OpenAPI generated artifacts 无差异。
- owner `context.yaml` validator 与 `git diff --check` 通过。

## 2 会话中的主要阻点/痛点

### 2.1 零残留扫描最初混淆了设计文档链接与 Demo 路径

- **证据**：pruning gate 首次报告 312 处 residual，其中包含 `../../../../ui-design/INDEX.md` 这类实际解析到 `docs/ui-design/` 的合法相对链接，以及负向 lint 测试中的禁用样例。
- **影响**：若直接按字符串删除，会误伤保留范围；需要先补测试，将路径语义和负向上下文纳入分类。

### 2.2 Demo 耦合不只存在于目录和测试命令

- **证据**：清理中发现 `#route=...` bootstrap、SPA deny prefix、fixture `prototype-baseline`、OpenAPI mapping、source-traceability tests、治理 Skills 和多个 owner context 同时依赖 Demo。
- **影响**：只删除目录会留下不可执行合同和无 owner 的兼容适配，无法达到真正降熵。

### 2.3 批量术语迁移可能破坏相对路径和 owner 名称

- **证据**：机械替换一度把 `../../../../ui-design/*.md` 改为不存在的 `docs/frontend/*.md`，并使 plan 链接指向未落盘的目录名；`make docs-check` 和 context validator 明确报错。
- **影响**：产生一次文档返工，但未进入提交；最终通过路径解析和索引门禁修正。

### 2.4 L2 review 发现 browser gate 同义词与 sibling owner 漂移

- **证据**：原 pruning lint 报告零残留，但 raw search 发现 18 个 active 文档仍要求已删除的 `responsive browser verification`；004 routing owner 同时保留 hash adapter、已删除 `bootstrapRoute.ts` 与 browser harness 当前合同。
- **影响**：结构门禁和代码回归全部可以通过，但后续 implement/review 会被 completed owner 引导到不存在的工具链；本次通过 BUG-0172、focused Red/Green 和 003 Phase 7 原地修复。

## 3 根因归类

- Demo 原合同横跨 governance、spec/plan、frontend、OpenAPI fixture 与工具链，却没有单一删除 gate。
  - **类别**：spec-plan / AGENTS.md / README / skill
- 初版扫描只理解字符串，不理解相对链接解析和负向测试上下文。
  - **类别**：spec-plan；本次已由 `ui_demo_pruning.py` 的 executable contract 修复
- pruning forbidden pattern 未覆盖迁移时引入的同义词，且没有反查引用 003 的 completed sibling owner。
  - **类别**：test / spec-plan；本次已由 BUG-0172 的同义词 Red test 与 004 owner reconcile 修复
- 批量替换没有以“链接目标可解析”和“context 可验证”作为每批即时停止条件。
  - **类别**：无需仓库改动；属于本次执行方式问题，现有 `docs-check` 与 context validator 已能可靠阻断

## 4 对流程资产的改进建议

- 后续任何“删除真理源/旧模块”的 plan 都应在 Phase 1 建立专属零目录、零 active-reference lint，并明确历史 allowlist。
  - **落点**：spec-plan
  - **优先级**：high
- UI-visible Skills 只读取 `docs/ui-design/`、active spec 与正式 frontend；不要把可运行展示工程重新加入设计前置条件。
  - **落点**：skill / AGENTS.md
  - **优先级**：high
  - **状态**：本次已同步 `design`、`implement`、`tdd`、`plan-code-review` 与根治理规则
- 批量修订 context 或 Markdown 链接时，每个 bounded batch 后立即运行 context validator 和 `make docs-check`，再继续下一批。
  - **落点**：无需仓库改动
  - **优先级**：medium
- 删除验证工具链时，把新旧命令名、描述性同义词和引用该 owner 的 sibling contexts 一起列入 executable negative matrix。
  - **落点**：spec-plan / test
  - **优先级**：high
  - **状态**：本次已补 `responsive browser verification`、`test:responsive-browser`、`serve-responsive-browser` Red/Green，并清理 004 与 18 个 active residual

## 5 建议优先级与后续动作

1. 下一轮 UI 设计修订直接以 `docs/ui-design/` 对应文档为 owner，并在正式 `frontend/` 中实施和验证；不要恢复 Demo-first 或双源同步。
2. 若后续继续做仓库降熵，复用本次“先建零残留 gate、再删除实体、最后清理消费者”的顺序，并按 bounded batch 执行链接/context 校验。
3. 后续 pruning branch 合并前复用 BUG-0172 的审查模式：先跑 owner lint，再做不依赖 allowlist 的 raw negative search，并反查所有 sibling plan contexts。
