# DB Migrations Baseline History

> **版本**: 1.5
> **状态**: active
> **更新日期**: 2026-04-29

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-04-29 | 1.5 | 收口 A/B spec 全面审查 remediation：同步 B3 v1.3 `email_dispatch` internal-only jobType 到 `async_jobs.job_type` check；为 `ai_task_runs` 增补 A3/F1 所需 typed columns；新增 P0 privacy deletion table matrix，锁定 hard delete / cascade / retain / audit tombstone 策略。 | plan-review remediation |
| 2026-04-29 | 1.4 | 同步 B3 v1.2 outbox remediation：将 `publish_attempts` / `next_attempt_at` / `locked_at` / `last_error_code` / `last_error_message` 与 pending due 查询索引纳入 B4 baseline migration 输入，避免 C8 dispatcher retry 语义缺少表字段承载。 | event-and-outbox-contract plan-review remediation |
| 2026-04-29 | 1.3 | 修复 L1 review findings：把 ADR-Q1 的 `auth_challenges` / `sessions` / `external_identities` 纳入 B4 baseline；新增 `schema_backfills` 与 `backend/cmd/migrate` backfill ledger / dry-run / apply 契约；明确 pgvector up/down 生命周期与 A2 dev-stack 责任边界；把 enum/check 约束拆为 B1 / B2 / B3 / ADR-Q1-C1 / B4 来源矩阵；对齐当前 `go.work` + `backend/go.mod` 拓扑。 | plan-review remediation |
| 2026-04-27 | 1.2 | 清理剩余 CI drift / backfill 表述：B4 当前用本地 drift 与 migrate-check gate 收口，远端 CI 执行仅在 A5 触发条件成立后再接入。 | ci-pipeline-baseline spec-contract remediation |
| 2026-04-27 | 1.1 | 对齐 A5 单人开发阶段决策：B4 当前只要求本地 migrate check 完整闭环，远端 CI 迁移校验不作为 P0 前置。 | ci-pipeline-baseline spec-contract remediation |
| 2026-04-27 | 1.0 | 初始创建：锁定 `golang-migrate` 工具、`migrations/` 目录与 `NNNNNN_<verb>_<noun>.{up,down}.sql` 文件命名、27 张 P0 应用表（与 [03 §4](../../../easyinterview-tech-docs/03-db-definition.md#4-表清单) 一致）+ `schema_migrations` 元数据 + pgvector 扩展、03 §7 全部索引、`make migrate-{up,down,status,create}` target、可逆 + 数据回填策略、prod 防呆与 enum 与 B1 同源约束；引用 [03 §9 迁移策略](../../../easyinterview-tech-docs/03-db-definition.md#9-迁移策略) 与 [B1 D-6 枚举](../shared-conventions-codified/spec.md#31-已锁定决策)。 | engineering-roadmap/001 Phase 3 |
