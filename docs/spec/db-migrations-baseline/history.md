# DB Migrations Baseline History

> **版本**: 1.17
> **状态**: active
> **更新日期**: 2026-05-15

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-15 | 1.17 | 授权 backend-review/001 Phase 0 pre-launch baseline rebase（同 commit）：(a) `migrations/000001_create_baseline.up.sql` 中 `ai_task_runs.task_type` CHECK 扩值 `report_assessment`（与 `report_generate` 并列）；(b) `feedback_reports` 表新增 4 列 `language text NOT NULL DEFAULT 'en'` / `feature_flag text NOT NULL DEFAULT 'none'` / `data_source_version text NOT NULL DEFAULT 'not_applicable'` / `retry_focus_turn_ids jsonb NOT NULL DEFAULT '[]'::jsonb`，使 wire `GenerationProvenance` 6 字段可由 `feedback_reports` 单表 round-trip。同步 `migrations/enum-sources.yaml`、`migrations/lint.sh`、`backend/internal/ai/aiclient/writers.go` 的 `AITaskRunTaskReportAssessment` 常量与 `allowedAITaskRunCapabilities` 集合。Additive 默认值不破坏既有数据；与 [B1 1.18](../shared-conventions-codified/history.md) + [B2 1.20](../openapi-v1-contract/history.md) 同 commit 闭合 plan-review --fix。 | backend-review/001-report-generation-baseline Phase 0.3 / 0.5 |
| 2026-05-12 | 1.16 | L2 remediation：`resume_versions` live integration test cleanup 改为按 FK 顺序删除 suggestions / versions / tailor runs / target jobs / assets / users，并把清理错误作为测试失败；新增 C-14 rerun-safe cleanup gate。 | db-migrations-baseline/002-resume-versions-additive Phase 5 |
| 2026-05-12 | 1.15 | D-17 Resume Workshop additive 表与字段落地阶段：新增 `migrations/000005_resume_versions.{up,down}.sql`，实际创建 `resume_versions` / `resume_version_suggestions` 2 张表、`resume_assets` 4 个 additive 字段、4 项 enum-source 登记、privacy deletion matrix 与 baseline inventory 回填；当前 P0 应用表升至 28 张。 | db-migrations-baseline/002-resume-versions-additive |
| 2026-05-11 | 1.14 | D-17 Resume Workshop additive 表与字段声明阶段：拟新增 `resume_versions` / `resume_version_suggestions` 2 张表 + `resume_assets` additive 字段（`source_type` / `original_text` / `guided_answers` / `parsed_text_snapshot`）；同步声明 §3.1.2 privacy deletion matrix 与 §2.1 baseline 表 inventory（26 → 28，`resume_version_edits` 归 P1 延后）；具体 migration up/down + idx + enum-sources 同步由 002 plan 落地。 | db-migrations-baseline/002-resume-versions-additive（声明阶段，docs-only） |
| 2026-05-09 | 1.13 | 对齐 backend-practice Phase 0：PracticeMode 收敛为 `assisted` / `strict`，新增共享 `idempotency_records` 表、幂等唯一键与过期索引，并将 baseline 表范围更新为 26 应用表 + 3 auth 支撑表 + 2 迁移元数据表。 | backend-practice/001 Phase 0 |
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
