# DB Migrations Baseline Resume Versions Additive Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-11

**关联计划**: [plan](./plan.md)

## Phase 1: Migration up - 新表与字段补充

- [ ] 1.1 修订前跑 `make migrate-up && make migrate-down && make migrate-status` 确认 baseline PASS（验证：exit 0）
- [ ] 1.2 创建 `migrations/000NNN_resume_versions.up.sql`，包含 `resume_versions` 表 + idx `(user_id, updated_at DESC)` / `(resume_asset_id, version_type)` / `(parent_version_id) WHERE parent_version_id IS NOT NULL`（验证：SQL lint + `make migrate-up` 干净 DB 成功）
- [ ] 1.3 same SQL 创建 `resume_version_suggestions` 表 + idx `(resume_version_id, status)` / `(tailor_run_id)`，FK ON DELETE CASCADE 到 resume_versions（验证：`psql -c "\d+ resume_version_suggestions"` 字段与约束）
- [ ] 1.4 same SQL 追加 `ALTER TABLE resume_assets ADD COLUMN source_type / original_text / guided_answers / parsed_text_snapshot`，含 `source_type` check constraint；`guided_answers` 为 `jsonb NULL`（验证：`psql -c "\d+ resume_assets"` 新字段存在且 NULL）
- [ ] 1.5 跑 `make migrate-up` 干净 DB 成功（验证：exit 0 + schema_migrations 含新 version）
- [ ] 1.6 跑 `make migrate-up` 在含 `resume_assets` 历史数据的 DB 上成功，不破坏现有行（验证：count(*) before/after 一致 + 新字段全 NULL）
- [ ] 1.7 写 PG check constraint negative test：插入非法 `version_type='foo'` / `seed_strategy='bar'` / `source_type='unknown'` / `suggestion.status='unknown'`，PG 必须拒绝（错误码 23514）（验证：`go test ./internal/migrations/... -run TestResumeVersionsCheckConstraints` PASS）

## Phase 2: Migration down - 反向操作 + 幂等

- [ ] 2.1 创建 `migrations/000NNN_resume_versions.down.sql`，drop 顺序：suggestions（FK 子表）→ versions → resume_assets ADD COLUMN 反向（验证：SQL lint + `make migrate-down` 在已 apply 状态成功）
- [ ] 2.2 跑 `make migrate-down`（必要时 `MIGRATE_DOWN_FORCE=1` + APP_ENV=dev）成功（验证：exit 0 + `\d+ resume_assets` 无新字段）
- [ ] 2.3 跑 `make migrate-up` 重新 apply 成功（验证：双向幂等 IF NOT EXISTS / IF EXISTS 路径正确）
- [ ] 2.4 跑 `make migrate-down && make migrate-up` 循环 3 次，每次成功（验证：完全幂等）
- [ ] 2.5 写 FK cascade test：在 `resume_versions` 删除一行时，相关 `resume_version_suggestions` 行被级联删除（验证：`go test ./internal/migrations/... -run TestResumeVersionsCascadeDelete` PASS）
- [ ] 2.6 写 FK orphan negative test：仍存在 `resume_versions` 引用时直接删除 `resume_assets` 必须被拒绝，证明 privacy delete 必须按 suggestions → versions → assets 顺序执行（验证：`go test ./internal/migrations/... -run TestResumeAssetDeleteRequiresVersionCleanup` PASS）

## Phase 3: enum-sources + privacy matrix + spec 同步

- [ ] 3.1 修订 `migrations/enum-sources.yaml`，登记 4 项新 enum source：`resume_versions.version_type` / `resume_versions.seed_strategy` / `resume_version_suggestions.status` 归 B1，`resume_assets.source_type` 归 B2 `RegisterResumeRequest.sourceType`；不得引用不存在的 B1 sourceType enum（验证：`migrations/lint.sh` PASS + yaml lint）
- [ ] 3.2 跑 `migrations/lint.sh`，并通过专用 lint fixture / unit test 验证未登记 enum 会失败；不得手工破坏当前 `enum-sources.yaml` 后再依赖人工 revert（验证：lint negative fixture exit code 2）
- [ ] 3.3 验证 B4 spec §3.1.2 privacy deletion matrix 已含 `resume_versions` / `resume_version_suggestions` 行（在 B4 spec 1.14 中追加）（验证：手工检查 + 引用 B4 spec line 102+ 后续行）
- [ ] 3.4 修订 B4 spec.md 1.14 → 1.15：D-17 措辞从"声明阶段"改为"已落地"；§2.1 表 inventory 从"26 → 28 拟新增"改为"28 表"，删除"拟"字（验证：手工检查）
- [ ] 3.5 B4 history.md 追加 1.15 行（关联本 plan，标记 "落地阶段"）（验证：`sync-doc-index --check`）

## Phase 4: 验收与下游同步

- [ ] 4.1 运行 §3 全部替代验证 gate PASS：`make migrate-up && make migrate-down && make migrate-up` + `make migrate-check` + `go test` + `migrations/lint.sh` + `sync-doc-index --check`
- [ ] 4.2 修订 `docs/spec/INDEX.md` db-migrations-baseline 版本与日期 1.14 → 1.15（验证：`sync-doc-index --check`）
- [ ] 4.3 在 `openapi-v1-contract/004-resume-additive-coverage` plan checklist 中追加 "B4 D-17 resume_versions / resume_version_suggestions 已落地" 引用（验证：cross-plan 引用 commit）
- [ ] 4.4 在 `event-and-outbox-contract/002-resume-tailor-mode-drift-fix` plan checklist 中追加 B4 D-17 落地信号（验证：cross-plan 引用 commit）
- [ ] 4.5 同步 `docs/spec/engineering-roadmap/spec.md` 中 "25 应用表 / 26 应用表" 文字描述升级到 "28 应用表"（如有）；spec.md / history.md 视需求同步（验证：grep negative search）
