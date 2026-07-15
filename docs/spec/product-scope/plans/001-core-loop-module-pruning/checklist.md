# Core Loop Module Pruning Checklist

> **版本**: 1.277
> **状态**: completed
> **更新日期**: 2026-07-14

**关联计划**: [plan](./plan.md)

## Phase 1: 产品与 UI 范围

- [x] 1.1 产品、roadmap、UI 文档与静态原型只保留简历/JD、模拟面试、报告、复练/下一轮核心闭环。
- [x] 1.2 一级导航、用户菜单、route 与 screen 删除复盘、候选人画像及其它范围外入口。
- [x] 1.3 正式前端与 `frontend/` 当前 DOM、状态和主题合同一致。

## Phase 2: 跨层合同清理

- [x] 2.1 OpenAPI、fixtures、generated clients、shared event/job/enum 不含范围外正向 surface。
- [x] 2.2 后端领域、wiring、数据库 final-state 与 seed 不含范围外模块或兼容层。
- [x] 2.3 prompt、rubric、eval 与 AI profile 只承接当前核心 feature keys。
- [x] 2.4 当前 owner spec、plan、context 与 INDEX 不把已删除模块作为正向执行入口。

## Phase 3: Route 与主题合同

- [x] 3.1 Parse 只承接 queued/processing 命令进度；ready 只读详情进入 target-scoped Workspace。
- [x] 3.2 Practice/Report/Generating 的恢复与 Back 仅使用可信 identity，不构造旧 route 参数。
- [x] 3.3 主题只保留 Ocean / Plum 和 custom hue / saturation；旧 preview/value/reset 合同删除。

## Phase 4: 测试分层

- [x] 4.1 删除将 Go/Vitest/pytest/lint/source-contract/fixture/build/pixel gate 包装成 E2E 的场景和引用。
- [x] 4.2 代码层测试保留在 owner package；阶段收口由仓库根 `make test` 执行 backend/frontend 全量回归。
- [x] 4.3 BDD-Gate: `BDD.CORE.001` 按 [bdd-checklist](./bdd-checklist.md) 由代码层行为/契约 gate 验证；不复用不匹配的 E2E ID。
- [x] 4.4 E2E 场景不调用或包装 `make test`，代码回归与端到端结果分别报告。

## Phase 5: 收口

- [x] 5.1 运行当前范围的 lint、OpenAPI/fixture/codegen/migration drift 与范围外 token 负向搜索。
- [x] 5.2 运行根 `make test`。
- [x] 5.3 运行 context、Header/INDEX、`make docs-check` 与 `git diff --check`。
