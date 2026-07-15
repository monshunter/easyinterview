# UI Demo Pruning and Documentation-Owned Design Checklist

> **版本**: 2.0
> **状态**: completed
> **更新日期**: 2026-07-15

**关联计划**: [plan](./plan.md)

## Phase 1: 建立删除合同并移除 Demo 实体

- [x] 1.1 新增 `scripts/lint/ui_demo_pruning.py`、`scripts/lint/ui_demo_pruning_test.py` 与 `make lint-ui-demo-pruning`，建立 `ui-design/` 零目录与 active-reference allowlist 合同（验证：`python3 -m pytest scripts/lint/ui_demo_pruning_test.py -q` Red=`ModuleNotFoundError`，Green=`6 passed`）
- [x] 1.2 删除 `ui-design/` 全部实体资产（验证：`test ! -d ui-design`；`python3 -m pytest scripts/lint/ui_demo_pruning_test.py -q`=`6 passed`；完整 pruning lint 保持后续 active-reference Red）

## Phase 2: 让 `docs/ui-design/` 成为纯设计文档

- [x] 2.1 改写 `docs/ui-design/README.md`、`INDEX.md` 与 `TEMPLATES.md`，移除 Demo 运行/同步合同并明确文档 owner（验证：`make docs-check` PASS；focused active residuals=0；pruning tests=`6 passed`）
- [x] 2.2 修订 `docs/ui-design/*.md` 中的原型源码、hash route、source replication 和 parity 口径，保留 UI 架构/流程/交互语义（验证：13 个 UI 设计文档 active residuals=0；`make docs-check` PASS；仅保留 README 负向禁用语句）

## Phase 3: 删除双源工具链

- [x] 3.1 删除 Playwright Demo parity config/server/spec/package 入口与 scaffold tests（验证：结构 test Red=`test:pixel-parity` still present，Green=`7 passed`；7 个 direct-consumer suites=`122 passed`；frontend full=`124 files / 986 tests passed`；保留真实 E2E 使用的 `@playwright/test`）
- [x] 3.2 删除 prototype fixture sync script/test/mapping/Make target，保持 OpenAPI fixture owner 独立（验证：结构/fixture tests=`52 passed, 4182 subtests`；`make validate-fixtures` PASS；`make codegen-check` PASS；5 个 derived scenarios 已删除）
- [x] 3.3 从根 `make test`、lint 与辅助脚本移除 Demo 合同测试和扫描路径并接入删除合同（验证：focused lint tests=`26 passed`；Demo pruning tests=`10 passed`；`make lint-ui-demo-pruning` active_residuals=0；根 `make test` PASS）

## Phase 4: 解耦正式前端测试与源码注释

- [x] 4.1 删除或改写读取已删除 Demo 源码的 source-traceability tests，保留正式 token/DOM/control/route/responsive/a11y 断言（验证：frontend full=`123 files / 970 tests passed`；根 `make test` PASS）
- [x] 4.2 清理 `frontend/` 源码注释、模块 README、Demo-only hash route adapter 和失去价值的 Demo import 负向断言（验证：frontend full=`123 files / 970 tests passed`；Demo pruning lint active_residuals=0；根 `make test` PASS）

## Phase 5: 修订当前治理与 owner 文档

- [x] 5.1 修订 `AGENTS.md`、`docs/development.md`、`docs/README.md`、`design` / `implement` / `plan-code-review` / `tdd` skills 和仍被当前 context 使用的 spec/plan/checklist/context，确立 `docs/ui-design/` 设计文档 → `frontend/` 直接实施流程（验证：context validator PASS；`make docs-check` PASS；根 `make test` 含 skill tests PASS）
- [x] 5.2 清理非历史 active 资产中的 Demo-first、旧 parity、source-level replication、golden preview 和可运行原型权威来源合同（验证：`make lint-ui-demo-pruning` PASS，ui_demo_directory=absent，active_residuals=0）

## Phase 6: 验证与生命周期收口

- [x] 6.1 执行 Demo pruning lint、`make test`、`make build`、`make docs-check`、`make codegen-check` 与 `git diff --check` 并记录当前 PASS（验证：全部命令当前 PASS；pruning active_residuals=0；frontend production build PASS）
- [x] 6.2 完成 post-pass doc reconcile、INDEX 同步和 retrospective 后恢复 spec/plan/checklist `completed` 生命周期（验证：retrospective=`docs/reports/2026-07-15-ui-demo-pruning-assessment.md`；context validator PASS；sync-doc-index check PASS；checklist zero-open）
