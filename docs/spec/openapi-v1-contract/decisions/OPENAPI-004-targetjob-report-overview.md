# OPENAPI-004 · TargetJob canonical-round report overview

> **ID**: OPENAPI-004
> **状态**: accepted
> **日期**: 2026-07-14
> **版本**: 1.1

## 1 背景

当前 `GET /api/v1/targets/{targetJobId}/reports` 返回 cursor-paginated full `FeedbackReport` 列表，而规划详情真正需要的是“每个标准轮次当前可打开的报告，以及最近一次生成尝试的状态”。flat list 迫使前端自行解析冻结 context、归组、排序并判断失败/生成中的新尝试是否覆盖旧 ready；`TargetJob.latestReportId` 又维护了第二份会漂移的可变指针。

项目尚未上线。用户于 2026-07-14 确认 R-A：原地修改既有 endpoint，删除分页与 TargetJob 指针，由后端基于 owned TargetJob canonical rounds 和 immutable report context 返回最小 overview；不保留兼容层。

## 2 决策

- 保持 method/path/operationId/status：`GET /api/v1/targets/{targetJobId}/reports`、`listTargetJobReports`、`200`。
- 删除可选 query `cursor/pageSize`，response 从 `PaginatedFeedbackReport` 改为 closed `TargetJobReportsOverview`：
  - `targetJobId` required UUID；
  - `rounds` required array，覆盖当前 TargetJob 的全部 canonical rounds，按 sequence 升序；
  - 每项 required `round: PracticeRoundRef`；
  - `currentReport` required nullable，非空时只含 required `{id,generatedAt}`；
  - `latestAttempt` required nullable，非空时只含 required `{id,status,errorCode,createdAt}`。`errorCode` 仅 failed 可非空，其余状态必须为 null。
- 同批删除 `TargetJob.latestReportId`、DB `target_jobs.latest_report_id` 与 `PaginatedFeedbackReport`；不新增 alias、兼容 query、替代 pointer 或 top-level report center。
- `currentReport` 只从合法 ready report 中按 `generated_at DESC, created_at DESC, id DESC` 选择；`latestAttempt` 从全部合法状态按 `created_at DESC, id DESC` 选择。更新的 queued/generating/failed 不得挤掉旧 ready；同一最新 ready 可以同时出现在两处。
- owned TargetJob 不存在、越权或软删返回 404。TargetJob canonical summary 无效，或任一关联 report 的 frozen context 缺失/非法、user/target/session identity 不一致、round pair 不属于当前 catalog、ready 缺 `generated_at` 时，整份 overview fail closed；不得 partial、route fallback 或查询 mutable session/plan 重建 context。

## 3 影响

| 边界 | 受影响的项 | Owner |
|------|-----------|-------|
| 契约 | list endpoint query/response、TargetJob field、new closed summaries、baseline diff | openapi-v1-contract 001/003 |
| Fixtures | `Reports/listTargetJobReports.json`、TargetJob fixtures、prototype projection、Prism | openapi-v1-contract 002 |
| 数据库 | 删除 `target_jobs.latest_report_id` | db-migrations-baseline/001 |
| 后端 | canonical catalog validation、report selection/read model、TargetJob store cleanup | backend-review/001 + backend-targetjob/001 |
| 前端 | Parse 内容区右上报告入口、独立 target-scoped ReportsScreen、Report/Generating trusted-context Back；Parse 无列表请求或 `section=reports` | frontend-home/001 + frontend-report-dashboard/001 + frontend-shell/004 |
| BDD | current ready/latest attempt/invalid context/back navigation | E2E.P0.016 + E2E.P0.059 |

## 4 迁移与回滚

- **迁移路径**：先记录本 accepted decision，并在旧 baseline 未变时建立 exact finding oracle；再同步 OpenAPI、fixtures/codegen、baseline SQL、backend/frontend/scenario；所有 owner green 后保存 old-baseline audit 并 re-freeze。
- **放行条件**：37 operations/10 tags 与 endpoint method/path/operationId/200 不变；exact finding 无缺失/额外；fixture/Prism/codegen、real PostgreSQL selection/isolation、Parse 入口/Reports/Report desktop+mobile BDD 全通过；旧 pagination/full-list/pointer 与 Parse 内嵌列表/`section=reports` zero-reference。
- **回滚**：任一 invalid-context fail-closed、user isolation、stable selection 或 frontend consumer 未同批完成时整体回滚本 correction；不得仅恢复 pointer 或兼容 response。
- **SemVer**：v1.0.0 尚未发布，作为 accepted pre-release freeze correction 原地 re-freeze；发布后同类变更必须使用 major version。

## 5 相关

- [openapi-v1-contract spec](../spec.md) D-36
- [backend-review spec](../../backend-review/spec.md)
- [backend-targetjob spec](../../backend-targetjob/spec.md)
- [frontend-home-job-picks-and-parse spec](../../frontend-home-job-picks-and-parse/spec.md)
- [frontend-report-dashboard spec](../../frontend-report-dashboard/spec.md)

## 6 审计

| 项 | 内容 |
|----|------|
| 提议人 | interview UX owner |
| Review | product owner confirmed R-A on 2026-07-14 |
| 实施分支 | `fix/interview-turn-ux-0713` |
| `make openapi-diff` 证据 | 003 Phase 8 在 schema mutation 后、baseline re-freeze 前记录 exact five-key artifact |
| baseline | `openapi/baseline/openapi-v1.0.0.yaml` pre-release correction |
| history | `2026-07-14 | 1.57 | OPENAPI-004 TargetJob report overview` |

## 7 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-07-14 | 1.1 | 将前端影响口径同步为 Parse 页面级入口、独立 ReportsScreen 与 trusted-context Back，删除已被修订方案取代的 Parse section 口径。 | 入口方案 A（修订版） |
| 2026-07-14 | 1.0 | 接受 canonical-round overview、independent current/latest selection 与 no-pointer/no-pagination 边界。 | R-A |
