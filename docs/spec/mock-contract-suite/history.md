# Mock Contract Suite History

> **版本**: 1.5
> **状态**: active
> **更新日期**: 2026-05-11

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-11 | 1.5 | B2 D-18 Resume Workshop additive 升级声明阶段同步占位：§2.1 / §6 C-1 保留 46 operation 现状，追加 D-18 声明扩到 55 operation 的预告与落地路径引用（openapi-v1-contract/004-resume-additive-coverage）；本 spec 实际 inventory 数字升级跟随 B2 plan 004 落地后同步 1.5 → 1.6。 | openapi-v1-contract/004-resume-additive-coverage（声明阶段，docs-only） |
| 2026-05-10 | 1.4 | 合并 named scenario truth-source remediation 与 frontend Vite dev preview mock wiring 要求，固化后端 mockruntime 与前端 dev preview 两类 gate。 | 001-fixture-backed-mock-runtime |
| 2026-05-10 | 1.3 | 补充 frontend Vite dev preview 默认 fixture-backed mock wiring 要求，解决无真实 backend 时已开发页面不可见的问题。 | 001-fixture-backed-mock-runtime |
| 2026-05-06 | 1.2 | 对齐 backend-runtime-topology：mock runtime out-of-scope 从后台 worker 改为 backend internal runner，避免把独立 worker 当作默认前置。 | backend-runtime-topology/001-worker-consolidation |
| 2026-05-06 | 1.1 | 补充 fixture tag 目录级旧口径拦截要求，覆盖空目录和 Git 不跟踪目录残留。 | 001-fixture-backed-mock-runtime |
| 2026-05-05 | 1.0 | 从 engineering-roadmap S1 派生 fixture-backed mock runtime subject，作为前后端 mock runway owner。 | 001-fixture-backed-mock-runtime |
