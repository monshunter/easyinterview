# Workspace Persistent Archive Delete 交付复盘报告

> **日期**: 2026-07-09
> **审查人**: Codex

## 1 复盘范围与成功证据

- 交付范围：把 workspace 面试规划卡片删除从本地隐藏升级为持久 `archiveTargetJob`；删除图标固定到卡片右上角，footer 只保留 `立即面试`；Home 最近模拟面试复用同一 `MockInterviewCard` 主体但不展示删除。
- 后端范围：新增 `POST /targets/{targetJobId}/archive` B2 additive contract、fixture、generated client/server、TargetJob store/service/handler 持久归档，以及 `cmd/api` production mux route。
- 成功证据：
  - `make lint-openapi`、`make validate-fixtures`、`make lint-mock-contract` PASS。
  - `node --test ui-design/ui-design-contract.test.mjs` PASS。
  - `pnpm --filter @easyinterview/frontend test src/app/screens/home/MockInterviewCard.test.tsx src/app/screens/home/HomeRecentMocks.test.tsx src/app/screens/workspace/WorkspaceScreen.test.tsx src/app/screens/workspace/WorkspaceEmptyState.test.tsx` PASS。
  - `pnpm --filter @easyinterview/frontend typecheck` PASS。
  - `cd backend && go test ./internal/targetjob -count=1` PASS。
  - `cd backend && go test ./cmd/api -count=1` PASS。
  - `test/scenarios/e2e/p0-018-workspace-default-render/scripts/setup.sh && .../trigger.sh && .../verify.sh && .../cleanup.sh` PASS，输出 `E2E.P0.018 PASS`。
  - local real-backend browser smoke 通过，删除前截图为 `.test-output/e2e/workspace-archive-real-browser/workspace-card-before-delete.png`，删除后截图为 `.test-output/e2e/workspace-archive-real-browser/workspace-after-delete.png`，DB readback `archive-db-state.txt=archived|t`，刷新后文本不含目标岗位 title/id。

## 2 会话中的主要阻点/痛点

- 真实浏览器联调发现 production mux 漏挂 archive route。
  - **证据**：点击右上角删除时，前端已请求 `POST /api/v1/targets/{targetJobId}/archive`，但后端返回裸 `404 page not found`；OpenAPI/generated/handler focused tests 均已通过。
  - **影响**：需要回到 `backend/cmd/api/main.go` 挂载 route，并补 `cmd/api` route mounting regression；见 [BUG-0150](../bugs/BUG-0150.md)。

- BDD 证据跨 frontend scenario wrapper 与 local real-backend smoke 两层。
  - **证据**：`E2E.P0.018` wrapper 覆盖 generated client、右上角 delete、failure preserving card 和 no-bubble；真实 browser smoke 覆盖登录、DB seed、删除、刷新和截图。
  - **影响**：checklist 需要明确哪些证据来自自动脚本，哪些来自本次 local smoke，避免把未脚本化的 direct API/DB 断言误标为 scenario-only 完成。

- 测试数据清理脚本一开始误按 `async_jobs.user_id` 清理。
  - **证据**：本地清理 SQL 失败后反查 DDL，`async_jobs/outbox_events` 没有 `user_id` 列；最终改为按 owned resource id 和 auth challenge id 清理，`cleanup-db-state.txt=0`。
  - **影响**：只影响验收现场清理，不影响功能；后续若要把该 smoke 脚本化，应复用 full-funnel scenario 的 owned resources 清理模式。

## 3 根因归类

- 新增 OpenAPI operation 缺少 production mux 级别 gate。
  - **类别**：spec-plan。

- P0.018 当前是 frontend scenario wrapper + local browser smoke 的组合 gate。
  - **类别**：spec-plan。

- 手工 smoke 清理 SQL 没有先复用现有 scenario cleanup 模式。
  - **类别**：no repo change needed。

## 4 对流程资产的改进建议

- Backend TargetJob plan 后续新增 operation 时，把 `cmd/api` route mounting test 列为强制 checklist 项，不能只依赖 generated interface 和 handler focused tests。
  - **落点**：spec-plan。
  - **优先级**：high。

- 若 workspace archive smoke 需要长期回归，应把当前 local browser smoke 固化成 P0.018 的真实后端子场景，包含登录、seed、click、refresh、DB readback 和 cleanup。
  - **落点**：spec-plan / scenario README。
  - **优先级**：medium。

- 脚本化清理时复用 owned resources CTE：`resumes`、`target_jobs`、`practice_plans`、`practice_sessions`、`feedback_reports`，再删 `async_jobs.resource_id` 和 `outbox_events.aggregate_id`。
  - **落点**：scenario README。
  - **优先级**：medium。

## 5 建议优先级与后续动作

- 最高优先级：在 backend-targetjob owner checklist 保留 `cmd/api` route mounting regression，防止后续 additive route 重复出现 handler 已有但 runtime 不通。
- 次优先级：把本次 local real-backend browser smoke 脚本化为 P0.018 real-backend 子场景，减少后续截图验收的人工作业。
- 可延后：整理一个共享 SQL cleanup helper，复用 full-funnel owned resources 模式，避免每个浏览器 smoke 手写清理语句。
