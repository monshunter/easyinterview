# DB Migrations Baseline History

> **版本**: 1.12
> **状态**: active
> **更新日期**: 2026-05-08

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-08 | 1.12 | 对齐 A2 用户决策：B4 本地迁移验证前提升级为 Postgres 18。 | local-dev-stack/001 post-pass revision |
| 2026-05-08 | 1.11 | 对齐 A3 003 Phase 6：删除向量扩展、向量检索表/索引、extension drop gate 与对应 privacy/enum/source 条目；当前 baseline 收敛为 25 应用表 + 3 auth 支撑表 + 2 迁移元数据表。 | ai-provider-and-model-routing/003 Phase 6 |
| 2026-05-06 | 1.10 | 对齐 backend-runtime-topology：privacy deletion matrix 的执行方从 C8 worker 改为 backend internal runner，表级删除契约保持不变。 | backend-runtime-topology/001-worker-consolidation |
| 2026-05-05 | 1.9 | B4 migration baseline 改为只由本 spec、`migrations/`、`migrations/enum-sources.yaml` 与 migration gate 承接；移除旧技术草稿名称和旧 shorthand 依赖。 | engineering-roadmap/001-decompose-subspecs |
| 2026-05-03 | 1.8 | 明确当前迁移 baseline 由本 spec、`migrations/` 与 product-scope 当前范围决定，避免旧 27 应用表 / 旧 enum / 旧索引口径被当作当前迁移依据。 | docs-only |
| 2026-05-03 | 1.7 | 修正 v1.6 后残留的历史表数量文案：当时 baseline 统一表述为应用表 + auth 支撑表 + 迁移元数据表。 | readiness reconcile |
| 2026-05-03 | 1.6 | 对齐 product-scope v1.2：删除独立 `mistake_entries` 表，报告题目回顾 / 本轮复练由 `question_assessments` / `feedback_reports` 承载；同步当时 baseline 表数量 gate。 | 001-bootstrap Phase 5 remediation |
| 2026-04-29 | 1.5 | 收口 A/B spec 全面审查 remediation：同步 B3 v1.3 `email_dispatch` internal-only jobType 到 `async_jobs.job_type` check；为 `ai_task_runs` 增补 A3/F1 所需 typed columns；新增 P0 privacy deletion table matrix，锁定 hard delete / cascade / retain / audit tombstone 策略。 | plan-review remediation |
| 2026-04-29 | 1.4 | 同步 B3 v1.2 outbox remediation：将 `publish_attempts` / `next_attempt_at` / `locked_at` / `last_error_code` / `last_error_message` 与 pending due 查询索引纳入 B4 baseline migration 输入，避免 C8 dispatcher retry 语义缺少表字段承载。 | event-and-outbox-contract plan-review remediation |
| 2026-04-29 | 1.3 | 修复 L1 review findings：把 ADR-Q1 的 `auth_challenges` / `sessions` / `external_identities` 纳入 B4 baseline；新增 `schema_backfills` 与 `backend/cmd/migrate` backfill ledger / dry-run / apply 契约；明确当时 DB extension 生命周期与 A2 dev-stack 责任边界；把 enum/check 约束拆为 B1 / B2 / B3 / ADR-Q1-C1 / B4 来源矩阵；对齐当前 `go.work` + `backend/go.mod` 拓扑。 | plan-review remediation |
| 2026-04-27 | 1.2 | 清理剩余 CI drift / backfill 表述：B4 当前用本地 drift 与 migrate-check gate 收口，远端 CI 执行仅在 A5 触发条件成立后再接入。 | ci-pipeline-baseline spec-contract remediation |
| 2026-04-27 | 1.1 | 对齐 A5 单人开发阶段决策：B4 当前只要求本地 migrate check 完整闭环，远端 CI 迁移校验不作为 P0 前置。 | ci-pipeline-baseline spec-contract remediation |
| 2026-04-27 | 1.0 | 初始创建：锁定 `golang-migrate` 工具、`migrations/` 目录与 `NNNNNN_<verb>_<noun>.{up,down}.sql` 文件命名、当时 P0 应用表 + `schema_migrations` 元数据 + DB extension、B4 索引 inventory、`make migrate-{up,down,status,create}` target、可逆 + 数据回填策略、prod 防呆与 enum 与 B1 同源约束。 | engineering-roadmap/001 Phase 3 |
