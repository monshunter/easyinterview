# DB Migrations Baseline Resume Versions Additive Checklist

> **版本**: 1.3
> **状态**: active
> **更新日期**: 2026-06-14

**关联计划**: [plan](./plan.md)

## Cross-plan prerequisite signals

- [x] B3 D-14 `ResumeTailorMode` 漂移修复已由 [event-and-outbox-contract/002](../../../event-and-outbox-contract/plans/002-resume-tailor-mode-drift-fix/plan.md) 落地；events contract 与 B2/B4 `gap_review` / `bullet_suggestions` 口径一致。

## Phase 1: Migration up - 新表与字段补充

- [x] 1.1 修订前跑 `make migrate-up && make migrate-down && make migrate-status` 确认 baseline PASS（验证：exit 0）
- [x] 1.2 创建 `migrations/000005_resume_versions.up.sql`，包含 `resume_versions` 表 + idx `(user_id, updated_at DESC)` / `(resume_asset_id, version_type)` / `(parent_version_id) WHERE parent_version_id IS NOT NULL`（验证：SQL lint + `make migrate-up` 干净 DB 成功）
- [x] 1.3 same SQL 创建 `resume_version_suggestions` 表 + idx `(resume_version_id, status)` / `(tailor_run_id)`，FK ON DELETE CASCADE 到 resume_versions（验证：`psql -c "\d+ resume_version_suggestions"` 字段与约束）
- [x] 1.4 same SQL 追加 `ALTER TABLE resume_assets ADD COLUMN source_type / original_text / guided_answers / parsed_text_snapshot`，含 `source_type` check constraint；`guided_answers` 为 `jsonb NULL`（验证：`psql -c "\d+ resume_assets"` 新字段存在且 NULL）
- [x] 1.5 跑 `make migrate-up` 干净 DB 成功（验证：exit 0 + schema_migrations 含新 version）
- [x] 1.6 跑 `make migrate-up` 在含 `resume_assets` 历史数据的 DB 上成功，不破坏现有行（验证：count(*) before/after 一致 + 新字段全 NULL）
- [x] 1.7 写 PG check constraint negative test：插入非法 `version_type='foo'` / `seed_strategy='bar'` / `source_type='unknown'` / `suggestion.status='unknown'`，PG 必须拒绝（错误码 23514）（验证：`go test ./internal/migrations/... -run TestResumeVersionsCheckConstraints` PASS）

## Phase 2: Migration down - 反向操作 + 幂等

- [x] 2.1 创建 `migrations/000005_resume_versions.down.sql`，drop 顺序：suggestions（FK 子表）→ versions → resume_assets ADD COLUMN 反向（验证：SQL lint + `make migrate-down` 在已 apply 状态成功）
- [x] 2.2 跑 `make migrate-down`（必要时 `MIGRATE_DOWN_FORCE=1` + APP_ENV=dev）成功（验证：exit 0 + `\d+ resume_assets` 无新字段）
- [x] 2.3 跑 `make migrate-up` 重新 apply 成功（验证：双向幂等 IF NOT EXISTS / IF EXISTS 路径正确）
- [x] 2.4 跑 `make migrate-down && make migrate-up` 循环 3 次，每次成功（验证：完全幂等）
- [x] 2.5 写 FK cascade test：在 `resume_versions` 删除一行时，相关 `resume_version_suggestions` 行被级联删除（验证：`go test ./internal/migrations/... -run TestResumeVersionsCascadeDelete` PASS）
- [x] 2.6 写 FK orphan negative test：仍存在 `resume_versions` 引用时直接删除 `resume_assets` 必须被拒绝，证明 privacy delete 必须按 suggestions → versions → assets 顺序执行（验证：`go test ./internal/migrations/... -run TestResumeAssetDeleteRequiresVersionCleanup` PASS）

## Phase 3: enum-sources + privacy matrix + spec 同步

- [x] 3.1 修订 `migrations/enum-sources.yaml`，登记 4 项新 enum source：`resume_versions.version_type` / `resume_versions.seed_strategy` / `resume_version_suggestions.status` 归 B1，`resume_assets.source_type` 归 B2 `RegisterResumeRequest.sourceType`；不得引用不存在的 B1 sourceType enum（验证：`migrations/lint.sh` PASS + yaml lint）
- [x] 3.2 跑 `migrations/lint.sh`，并通过专用 lint fixture / unit test 验证未登记 enum 会失败；不得手工破坏当前 `enum-sources.yaml` 后再依赖人工 revert（验证：lint negative fixture exit code 2）
- [x] 3.3 验证 B4 spec §3.1.2 privacy deletion matrix 已含 `resume_versions` / `resume_version_suggestions` 行（在 B4 spec 1.15 中追加）（验证：手工检查 + 引用 B4 spec §3.1.2）
- [x] 3.4 修订 B4 spec.md 1.14 → 1.15：D-17 措辞从"声明阶段"改为"已落地"；§2.1 表 inventory 从"26 → 28 拟新增"改为"28 表"，删除"拟"字（验证：手工检查）
- [x] 3.5 B4 history.md 追加 1.15 行（关联本 plan，标记 "落地阶段"）（验证：`sync-doc-index --check`）

## Phase 4: 验收与下游同步

- [x] 4.1 运行 §3 全部替代验证 gate PASS：`make migrate-up && make migrate-down && make migrate-up` + `make migrate-check` + `go test` + `migrations/lint.sh` + `sync-doc-index --check`
- [x] 4.2 修订 `docs/spec/INDEX.md` db-migrations-baseline 版本与日期 1.14 → 1.15（验证：`sync-doc-index --check`）
- [x] 4.3 在 `openapi-v1-contract/004-resume-additive-coverage` plan checklist 中追加 "B4 D-17 resume_versions / resume_version_suggestions 已落地" 引用（验证：cross-plan 引用 commit）
- [x] 4.4 在 `event-and-outbox-contract/002-resume-tailor-mode-drift-fix` plan checklist 中追加 B4 D-17 落地信号（验证：cross-plan 引用 commit）
- [x] 4.5 同步 `docs/spec/engineering-roadmap/spec.md` 中 "25 应用表 / 26 应用表" 文字描述升级到 "28 应用表"（如有）；spec.md / history.md 视需求同步（验证：grep negative search）

## Phase 5: L2 remediation - live test cleanup hardening

- [x] 5.1 `cleanupUser` 先清理 `resume_version_suggestions` / `resume_versions` 等 child rows，再删 user，且清理错误会 fail test（验证：focused Go test）
- [x] 5.2 focused migration test 重复运行不因固定 UUID 残留冲突；无 `DATABASE_URL` 时必须明确记录 skip 而非伪 PASS（验证：`cd backend && go test ./internal/migrations/... -run 'TestResumeVersions|TestResumeAssetDeleteRequiresVersionCleanup' -count=2 -v`；当前环境 `DATABASE_URL` 未设置，live tests 明确 skip，contract test PASS）

## Phase 6: D-20 简历扁平化 flatten migration

> product-scope D-20 / B4 D-22。Red 优先；下游 contract phase 依赖本 migration。

- [x] 6.1 静态 contract test `TestResumeFlattenMigrationContract`（`sql_contract_test.go`）断言 000015 up/down 结构（rename / drop 3 表 / `practice_plans.resume_id` / `source_type` {upload,paste}）+ 无 retired `'guided'`；live PG check-constraint（insert `'guided'`→23514）与 flatten schema live 断言 Postgres-deferred <!-- verified: 2026-06-13 go test ./internal/migrations/... ok; live PG deferred (no DATABASE_URL) -->
- [x] 6.2 创建 `migrations/000015_resume_flatten.up.sql`：`resume_assets`→`resumes` RENAME + ADD `structured_profile jsonb NOT NULL DEFAULT '{}'::jsonb` / `display_name text`，DROP `guided_answers`，`source_type` CHECK 收敛 {`upload`, `paste`}，`practice_plans.resume_asset_id`→`resume_id`，按 FK 反序 DROP `resume_version_suggestions` → `resume_tailor_runs` → `resume_versions`（验证：`make migrate-up` 干净 DB 成功 + `\d+ resumes`）
- [x] 6.3 backfill 改为 000015 up.sql 内 SQL `UPDATE ... FROM` 把 structured_master → `resumes.structured_profile` / `display_name`（在 DROP 之前；**不走 Go registry**，因 `RunBackfills` 在所有 SQL up 之后运行、无法读已 drop 的源表）；无 master 留空 `{}` / NULL（验证：`go test ./internal/migrations/... -run TestResumeFlatten*` 断言 structured_master content 落盘 + 无 master 留空；无 `DATABASE_URL` 时 live skip）
- [x] 6.4 创建 `migrations/000015_resume_flatten.down.sql`：`resumes`→`resume_assets` 反向 rename、恢复 `guided_answers` + `source_type` CHECK {upload,paste,guided}、`practice_plans.resume_id`→`resume_asset_id`、恢复三表结构骨架（验证：`make migrate-down`（FORCE+dev）`&& make migrate-up` 循环 3 次幂等 PASS）
- [x] 6.5 `migrations/enum-sources.yaml` 新增 `resumes.source_type`（source `openapi-v1-contract.RegisterResumeRequest.sourceType`，values `[upload, paste]` + checksum `sha256("upload|paste")[:16]`），匹配 000015 新 CHECK；**保留**全部历史 resume 条目（`resume_assets.parse_status` / `resume_assets.source_type` / `resume_versions.*` / `resume_version_suggestions.status` / `resume_tailor_runs.*`）——沿用 D-17 jd_match 先例（lint 扫描 000001/000005 历史 CHECK 仍要求注册），**不改** `migrations_lint.py` / `B1_SOURCE_MAP`（验证：`python3 scripts/lint/migrations_lint.py --repo-root .` PASS）
- [x] 6.6 B4 spec 1.22→1.23 + history 1.23：§2.1 表数 25、§3.1.2 privacy matrix（`resumes` 行 + 删 version/jd_match 行 + debriefs 收敛）、§6 C-1 表数 25 / ≥30、新增 D-22 决策、标注 D-17 / B4 本地 D-20 退役、回填 product-scope D-17 jd_match drop 计数漂移（本次 doc 修订已完成）（验证：`sync-doc-index --check` 零漂移）
- [ ] 6.7 跨 gate 收口：`make migrate-up && make migrate-down && make migrate-up` + `make migrate-check` + `cd backend && go test ./internal/migrations/... ./internal/store/...` + `migrations/lint.sh` + `sync-doc-index --check` PASS（无 `DATABASE_URL` 时 live 明确 skip，contract / lint / negative fixture PASS）；零 `resume_versions` / `resume_version_suggestions` / `resume_tailor_runs` / `resume_asset_id` 残留 grep（除负向断言与 down 骨架）
- [ ] 6.8 下游信号：`backend-resume`（store 改 `resumes` 单表）/ `openapi-v1-contract/004`（resume 契约坍缩）/ `shared-conventions-codified`（3 enum 退役）/ `backend-practice`（session resume_id binding）已收到 D-20 flatten 落地信号（验证：cross-plan 引用）
- [x] 6.9 L2 hardening: migration contract test 固化 narrowed CHECK 前历史合法行 cleanup：`source_type='guided'` rows 先转 `paste`，retired `jd_match_agent_scan` / `jd_match_search` async jobs 先删除再添加 narrowed `async_jobs.job_type` CHECK；验证：`go test ./backend/internal/migrations -run 'TestResumeFlattenMigrationContract|TestDropJDMatchMigrationDeletesRetiredAsyncJobsBeforeNarrowingCheck' -count=1`
