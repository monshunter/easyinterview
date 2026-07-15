# UI Demo Pruning and Documentation-Owned Design Checklist

> **版本**: 2.0
> **状态**: active
> **更新日期**: 2026-07-15

**关联计划**: [plan](./plan.md)

## Phase 1: 建立删除合同并移除 Demo 实体

- [x] 1.1 新增 `scripts/lint/ui_demo_pruning.py`、`scripts/lint/ui_demo_pruning_test.py` 与 `make lint-ui-demo-pruning`，建立 `ui-design/` 零目录与 active-reference allowlist 合同（验证：`python3 -m pytest scripts/lint/ui_demo_pruning_test.py -q` Red=`ModuleNotFoundError`，Green=`6 passed`）
- [x] 1.2 删除 `ui-design/` 全部实体资产（验证：`test ! -d ui-design`；`python3 -m pytest scripts/lint/ui_demo_pruning_test.py -q`=`6 passed`；完整 pruning lint 保持后续 active-reference Red）

## Phase 2: 让 `docs/ui-design/` 成为纯设计文档

- [ ] 2.1 改写 `docs/ui-design/README.md`、`INDEX.md` 与 `TEMPLATES.md`，移除 Demo 运行/同步合同并明确文档 owner（验证：`make docs-check` + Demo path negative scan）
- [ ] 2.2 修订 `docs/ui-design/*.md` 中的原型源码、hash route、source replication 和 parity 口径，保留 UI 架构/流程/交互语义（验证：`make docs-check` + focused `rg` negative scan）

## Phase 3: 删除双源工具链

- [ ] 3.1 删除 Playwright Demo parity config/server/spec/package 入口与 scaffold tests（验证：frontend package tests + package script/dependency negative scan）
- [ ] 3.2 删除 prototype fixture sync script/test/mapping/Make target，保持 OpenAPI fixture owner 独立（验证：相关 Python/Make tests + `make codegen-check`）
- [ ] 3.3 从根 `make test`、lint 与辅助脚本移除 Demo 合同测试和扫描路径并接入删除合同（验证：Makefile dry-run tests + Demo pruning lint）

## Phase 4: 解耦正式前端测试与源码注释

- [ ] 4.1 删除或改写读取 `ui-design/src` 的 source-traceability tests，保留正式 token/DOM/control/route/responsive/a11y 断言（验证：相关 focused Vitest；Phase 结束执行根 `make test`）
- [ ] 4.2 清理 `frontend/` 源码注释、模块 README 和失去价值的 Demo import 负向断言（验证：frontend focused tests + active-reference negative scan；Phase 结束执行根 `make test`）

## Phase 5: 修订当前治理与 owner 文档

- [ ] 5.1 修订 `AGENTS.md`、`docs/development.md`、`docs/README.md`、`design` / `implement` / `plan-code-review` / `tdd` skills 和仍被当前 context 使用的 spec/plan/checklist/context，确立 `docs/ui-design/` 设计文档 → `frontend/` 直接实施流程（验证：skill tests + context validation + `make docs-check`）
- [ ] 5.2 清理非历史 active 资产中的 Demo-first、pixel parity、source-level replication、golden preview 和“UI 真理源”合同（验证：Demo pruning lint + allowlisted historical-only repository scan）

## Phase 6: 验证与生命周期收口

- [ ] 6.1 执行 Demo pruning lint、`make test`、`make build`、`make docs-check`、`make codegen-check` 与 `git diff --check` 并记录当前 PASS
- [ ] 6.2 完成 post-pass doc reconcile、INDEX 同步和 retrospective 后恢复 spec/plan/checklist `completed` 生命周期（验证：context validator + sync-doc-index check + checklist zero-open）
