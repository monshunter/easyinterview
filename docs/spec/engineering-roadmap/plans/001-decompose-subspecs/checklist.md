# Roadmap Rebaseline and Subspec Governance Checklist

> **版本**: 3.1
> **状态**: completed
> **更新日期**: 2026-05-05

**关联计划**: [plan](./plan.md)

## Phase 1: 历史完成事实保留

- [x] 1.1 保留 ADR-Q1..Q6 作为当前认证、异步、分析、部署、隐私和 AI 路由架构约束；验证: roadmap spec v3.0 §3.2 仍链接并摘要 6 项 ADR
- [x] 1.2 保留当前已存在的 active foundation / contract / quality spec，不删除 A1-A5、B1-B4、F1、F3；验证: roadmap spec v3.0 §5.1 列出当前 active spec 清单
- [x] 1.3 将旧 “38 child pending INDEX” 任务改为历史事实，不再作为当前执行模型；验证: roadmap plan v3.0 §4 Phase 1 / Phase 2 明确关闭 pending 模型

## Phase 2: Roadmap rebaseline

- [x] 2.1 对齐产品与 UI 真理源：读取 product-scope、docs/ui-design 与 ui-design/src/app.jsx 当前模块 / 路由 / 删除范围；验证: roadmap spec v3.0 §4.1 只保留当前 UI 范围
- [x] 2.2 重写 `engineering-roadmap/spec.md` 为当前实施地图；验证: spec Header 版本 3.0，§5.2 只列 P0 workstream 候选，§5.3 只列 future candidates
- [x] 2.3 修订本 plan、checklist、context 和 plans/INDEX；验证: 本文件、plan.md、context.yaml、plans/INDEX Header / row 均为 3.0
- [x] 2.4 删除 `docs/spec/INDEX.md` 中所有 `_pending_` 行和待 spawn 分组；验证: INDEX 只包含真实存在的 spec link row
- [x] 2.5 同步 product-scope 中指向旧 engineering-roadmap v2.2 的交叉引用；验证: product-scope spec/history 更新到 v1.5
- [x] 2.6 验证文档一致性：`validate_context.py`、`sync-doc-index --check`、`check_md_links.py docs`、`git diff --check` 全部通过（2026-05-03：四项均通过；`sync-doc-index --check` zero drift）

## Phase 3: 后续 P0 workstream 创建规则

- [x] 3.1 创建任一 P0 workstream child spec 前，确认 product-scope 和 UI 文档已保留对应用户行为或工程能力
  <!-- verified: 2026-05-05 evidence=plan.md Phase 3.1 and engineering-roadmap spec §4.2/§5.2; this item records the future creation rule only and does not create a child spec -->
- [x] 3.2 创建任一 child plan 时同步生成 `context.yaml`、`plan.md`、`checklist.md`，涉及用户行为时同步 BDD plan / checklist
  <!-- verified: 2026-05-05 evidence=plan.md Phase 3.1 and engineering-roadmap spec §4.2/§7 C-5; this item records the future plan-completeness rule only -->
- [x] 3.3 涉及代码逻辑的 child plan 必须通过 `/implement` -> `/tdd` 执行，并按 checklist 顺序即时更新
  <!-- verified: 2026-05-05 evidence=plan.md Phase 3.1 and engineering-roadmap spec §4.2; this item records the future execution rule only -->
- [x] 3.4 Future candidates（readiness、retrieval、privacy export、source intel、production voice、multi-platform job search）不得提前创建空 spec / empty plan / INDEX pending row
  <!-- verified: 2026-05-05 evidence=engineering-roadmap spec §5.3/§6.5; no new future candidate spec or pending INDEX row was created -->
