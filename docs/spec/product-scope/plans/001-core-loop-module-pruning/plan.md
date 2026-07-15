# Core Loop Module Pruning Plan

> **版本**: 1.276
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

把产品范围收敛到以下可执行闭环，并删除范围外模块、兼容入口和重复测试壳：

```text
简历 + JD -> 模拟面试 -> 报告 -> 复练当前轮 / 进入下一轮
```

真实面试复盘、候选人画像、独立 Growth / Drill / Voice 模块不属于当前范围。账号设置、邮箱认证、首次资料补全与隐私删除保留，但不承担候选人画像语义。

## 2 当前合同

- 一级导航只保留首页、模拟面试、简历；用户菜单只保留设置与隐私、退出登录。
- `debrief`、`profile` 及其 API、schema、表、event、job、prompt、rubric 与正向 route 不存在。
- JD 解析命令进度与 ready Workspace 只读详情分离；ready 详情不回流到 Parse。
- 主题只保留 Ocean / Plum 与 custom hue / saturation；其余旧控件不属于当前合同。
- 未上线阶段不保留旧 route、旧 schema 或旧模块兼容层。

## 3 质量门禁

- **Plan 类型**: `feature-behavior + contract + migration + code-internal`。
- **TDD**: 开发中按 owner package 运行 focused test；阶段收口从仓库根运行 `make test`，统一执行 backend 和 frontend 全量单元测试。
- **BDD**: [bdd-plan](./bdd-plan.md) / [bdd-checklist](./bdd-checklist.md) 只描述用户可观察的范围不变量，由 domain behavior tests、范围 lint 与根 `make test` 验证；现存 E2E 不承担整站裁剪证明，也不为格式完整性虚构场景 ID。
- **静态/契约 gate**: `make lint`、`make codegen-check`、`make validate-fixtures`、migration lint、范围外 token 负向搜索、`make docs-check` 与 `git diff --check`。

## 4 实施范围

### Phase 1: 产品与 UI 设计文档

修订 product scope、engineering roadmap、`docs/ui-design/` 与 `frontend/`，只描述当前核心闭环；正式前端按静态原型复刻，不保留范围外导航、菜单、screen 或 route。

### Phase 2: 前端与 API 合同

删除范围外 frontend screen、hook、i18n、mock 分支；同步清理 OpenAPI paths/schemas/fixtures、Go/TS generated artifacts 和 shared event/job/enum。未知或旧路径归一到当前核心入口，不恢复范围外状态。

### Phase 3: 后端、数据与 AI 配置

删除范围外领域代码、wiring、migration final-state、seed、prompt、rubric、eval 与 profile。保留当前核心闭环和账号隐私所需事实，不增加兼容表或适配层。

### Phase 4: 当前 route 与主题收敛

Parse 只展示 queued/processing 命令进度；ready 后进入 target-scoped Workspace 详情。报告与练习恢复导航只使用可信 `targetJobId`。主题菜单只保留当前最小控制。

### Phase 5: 测试资产去重

删除把 `go test`、Vitest、pytest、lint、source contract、fixture parity、build 或 pixel gate 包装成 E2E 的场景目录和 BDD 引用。代码测试留在 owner package，由根 `make test` 做全量回归；真实 API/UI 场景由各自 domain owner 独立维护，不作为本计划的替代证据。

## 5 验收标准

- 当前源码、OpenAPI、generated artifacts、数据库 final-state、AI config 与文档只包含核心闭环。
- 范围外模块无用户入口、正向 route、API、持久化 owner 或测试场景。
- BDD 只描述真实用户行为；没有匹配的真实环境流程时不关联 E2E ID，代码层 gate 与 E2E 始终分开报告。
- 根 `make test`、必要 contract/lint/drift gate、文档和 diff 检查通过。

## 6 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-14 | 1.276 | Remove invalid full-funnel E2E mappings; keep module-pruning BDD as a minimal behavior contract verified by code-level gates. |
| 2026-07-14 | 1.275 | 按当前 E2E 证据边界压缩 owner plan：删除旧 wrapper/PASS 流水账，只保留当前合同、真实 API/UI BDD 与根 `make test` 回归入口。 |
