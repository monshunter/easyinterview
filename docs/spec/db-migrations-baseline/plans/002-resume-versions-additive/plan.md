# DB Migrations Baseline Resume Versions Additive

> **版本**: 1.2
> **状态**: active
> **更新日期**: 2026-06-13

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [db-migrations-baseline spec](../../spec.md) §3.1 D-17 声明的 Resume Workshop additive 表与字段落到 `migrations/` 实际 SQL artifact：

- 落地 `migrations/000NNN_resume_versions.up.sql` + `.down.sql`，引入 2 张新表：
  - `resume_versions`：id (uuidv7) / user_id (FK users) / resume_asset_id (FK resume_assets) / parent_version_id (FK resume_versions self, NULLABLE) / version_type ∈ {`structured_master`, `targeted`} / target_job_id (FK target_jobs NULLABLE) / display_name / seed_strategy NULLABLE ∈ {`copy_master`, `blank`, `ai_select`} / focus_angle NULLABLE / structured_profile jsonb not null default `'{}'::jsonb` / match_score numeric NULLABLE / prompt_version text NULLABLE / rubric_version text NULLABLE / model_id text NULLABLE / provider text NULLABLE / created_at / updated_at / deleted_at；含 idx `(user_id, updated_at DESC)` + idx `(resume_asset_id, version_type)` + idx `(parent_version_id) WHERE parent_version_id IS NOT NULL`；
  - `resume_version_suggestions`：id (uuidv7) / resume_version_id (FK resume_versions) / tailor_run_id (FK resume_tailor_runs) / original_bullet text not null / suggested_bullet text not null / reason text NULLABLE / status ∈ {`pending`, `accepted`, `rejected`} default `'pending'` / decided_at NULLABLE / created_at；含 idx `(resume_version_id, status)` + idx `(tailor_run_id)`；
- 落地 `resume_assets` additive 字段补充（保持向后兼容）：
  - `source_type text NULL CHECK (source_type IS NULL OR source_type IN ('upload', 'paste', 'guided'))`
  - `original_text text NULL`（粘贴文本原文保留）
  - `guided_answers jsonb NULL`（guided flow 的结构化答案原文；不序列化进 `original_text`）
  - `parsed_text_snapshot text NULL`（解析快照）
- 同步 `migrations/enum-sources.yaml` 登记 4 个新 enum check 来源（3 个 B1 D-10 owner + 1 个 B2 `RegisterResumeRequest.sourceType` owner）：
  - `resume_versions.version_type` → B1 `ResumeVersionType`
  - `resume_versions.seed_strategy` → B1 `ResumeSeedStrategy`
  - `resume_version_suggestions.status` → B1 `ResumeTailorSuggestionStatus`
  - `resume_assets.source_type` → B2 `RegisterResumeRequest.sourceType`
- 同步 B4 spec §3.1.2 privacy deletion matrix 已新增的 2 行规则的 backend internal runner 删除顺序文档（先 suggestions（FK 子表）→ versions → assets）；本 plan 不实现 runner，但需确保 FK 与 ON DELETE 级联策略允许 hard delete 路径无悬空；
- 通过 spec §6 验收（C-10 migrate-check / C-13 privacy deletion dry-run / 新增 C-14 resume_versions FK 完整性如有）；
- 落地后 B4 spec §2.1 表 inventory 升级到 28（声明阶段在 plan 落地后由 spec 1.15 同步）；`resume_version_edits` 表归 P1 延后，不在本 plan 范围。

本 plan 不实现 backend handler（归 `backend-resume`）；不实现 frontend Resume Workshop UI（归 `frontend-resume-workshop`）；不修订 `resume_tailor_runs.mode` enum check（已是 `[gap_review, bullet_suggestions]`，由 B3 plan 002 同步 events 层）。

## 2 背景

B4 spec §3.1 D-17 已声明 Resume Workshop 阶段 0 contract additive 升级所需 DB 范围：在 spawn `backend-resume` / `frontend-resume-workshop` 之前必须先把 `resume_versions` 与 `resume_version_suggestions` 两张承载结构化主版本 / 岗位定制版本 / 改写建议状态的表落到 baseline，避免下游业务 plan 启动时遭遇 store 层 model 缺失。Resume Workshop UI 真理源（`ui-design/src/screen-resume-workshop.jsx` 的 `ResumeListView` / `ResumeDetailView` / `ResumeBranchFlow` 全套组件）依赖 `ResumeVersion` 与 `ResumeVersionSuggestion` 模型完整存在；当前 baseline 只有 `resume_assets` + `resume_tailor_runs` 两张表，无法承载完整 version tree 与 suggestion 状态机。

本 plan 是 [db-migrations-baseline spec §7 关联计划](../../spec.md#7-关联计划) 列出的第 2 个，承担 D-17 落地：把 B4 spec 1.14 声明阶段实际投影到 `migrations/`、`migrations/enum-sources.yaml` 与 §2.1 表 inventory。

每个 phase 是可独立验证的纵向切片：Phase 1 起来就有 migration up 创建新表 + 字段；Phase 2 起来就有 down migration 幂等回退；Phase 3 起来就有 enum-sources 与 privacy matrix 同步；Phase 4 收口验收 + 解锁 backend-resume / frontend-resume-workshop 启动。

执行本 plan 前必须确认：

- [001-bootstrap](../001-bootstrap/plan.md) Phase 已完成：`migrations/000001_*.up.sql` baseline 已就位 + `make migrate-up/down/check/status` 入口可用 + `migrations/enum-sources.yaml` 框架就绪 + `schema_backfills` 工具表存在。
- [B1 D-10 声明](../../../shared-conventions-codified/spec.md#31-已锁定决策) 已就位（spec 1.16）；本 plan 引入的 3 个 shared enum 字面量必须与 B1 vocabulary 一致；`resume_assets.source_type` 的 authority 是 [B2 D-18](../../../openapi-v1-contract/spec.md#31-已锁定决策v100-freeze-范围) `RegisterResumeRequest.sourceType`，不得隐式发明第四个 B1 enum；具体 generator 输出由 openapi-v1-contract/004 落地。
- [A2 local-dev-stack](../../../local-dev-stack/spec.md) Postgres 18 已就位（B4 history.md 1.12 已记录）；本 plan 验证 migration 在 Postgres 18 上可执行。

## 3 质量门禁分类

- **Plan 类型**: `migration + contract`。本 plan 修订 `migrations/000NNN_*.sql`、`migrations/enum-sources.yaml`、B4 spec §2.1 / §3.1.2 表述；不实现 backend handler、前端 UI、用户 workflow。
- **TDD 策略**: 适用（Code plan requires TDD）。Red-Green-Refactor 入口：
  1. 修订前跑 `make migrate-up && make migrate-down && make migrate-status` 确认 baseline PASS；
  2. 写 `.up.sql` 创建表、字段、idx、FK、check constraint，验证 `make migrate-up` 在干净 DB 上成功；
  3. 写 `.down.sql` 反向操作（drop tables + drop 字段），验证 `make migrate-down` 在已 apply 状态下成功，并跑 `make migrate-up` 二次 apply 验证幂等；
  4. 写 integration test：在非空 baseline 数据（已有 `resume_assets` + `resume_tailor_runs` 行）上跑 up，验证不破坏现有数据；写 down 测试验证用户表行未受影响；
  5. 写 enum check constraint negative test：插入非法 `version_type` / `seed_strategy` / `source_type` 值，验证 PG check constraint 拒绝；
  6. 写 FK 完整性 test：删除 parent `resume_versions` 时 `resume_version_suggestions` 必须级联；删除仍被 `resume_versions` 引用的 `resume_assets` 必须被 FK 拒绝，privacy delete 由 backend-resume 先显式删除 versions / suggestions 后再删 assets。
  执行入口：`/implement db-migrations-baseline/002-resume-versions-additive` → `/tdd`。
- **BDD 策略**: 不适用。本 plan 是 DB schema additive，无用户可感知 UI / API 行为变化；后续用户可见 Resume Workshop 流程由 `frontend-resume-workshop` / `backend-resume` 维护 BDD gate。
- **替代验证 gate**:
  - `make migrate-up && make migrate-down && make migrate-up`（幂等验证）
  - `make migrate-check`（整体 baseline drift gate）
  - `cd backend && go test ./internal/migrations/... ./internal/store/...`（含本 plan 的 schema test）
  - `psql -c "\d+ resume_versions"` 手工验证字段类型与 idx 覆盖
  - 跨仓库验证：B4 spec §2.1 / §3.1.2 / `migrations/enum-sources.yaml` / docs/spec/INDEX.md 同步
  - `sync-doc-index --check`

## 4 实施步骤

### Phase 1: Migration up - 新表与字段补充

#### 1.1 新增 `migrations/000NNN_resume_versions.up.sql`

按 B4 D-2 命名规则（编号继续递增，假设当前最高 `00000X`，下一个 `00000X+1`）：

```sql
-- resume_versions: structured master + targeted branch
CREATE TABLE IF NOT EXISTS resume_versions (
    id              uuid PRIMARY KEY,
    user_id         uuid NOT NULL REFERENCES users(id),
    resume_asset_id uuid NOT NULL REFERENCES resume_assets(id),
    parent_version_id uuid REFERENCES resume_versions(id),
    version_type    text NOT NULL CHECK (version_type IN ('structured_master', 'targeted')),
    target_job_id   uuid REFERENCES target_jobs(id),
    display_name    text NOT NULL,
    seed_strategy   text CHECK (seed_strategy IS NULL OR seed_strategy IN ('copy_master', 'blank', 'ai_select')),
    focus_angle     text,
    structured_profile jsonb NOT NULL DEFAULT '{}'::jsonb,
    match_score     numeric,
    prompt_version  text,
    rubric_version  text,
    model_id        text,
    provider        text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    deleted_at      timestamptz
);
CREATE INDEX idx_resume_versions_user_updated ON resume_versions (user_id, updated_at DESC);
CREATE INDEX idx_resume_versions_asset_type   ON resume_versions (resume_asset_id, version_type);
CREATE INDEX idx_resume_versions_parent       ON resume_versions (parent_version_id) WHERE parent_version_id IS NOT NULL;

-- resume_version_suggestions: tailor run accept/reject 状态
CREATE TABLE IF NOT EXISTS resume_version_suggestions (
    id                  uuid PRIMARY KEY,
    resume_version_id   uuid NOT NULL REFERENCES resume_versions(id) ON DELETE CASCADE,
    tailor_run_id       uuid NOT NULL REFERENCES resume_tailor_runs(id),
    original_bullet     text NOT NULL,
    suggested_bullet    text NOT NULL,
    reason              text,
    status              text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected')),
    decided_at          timestamptz,
    created_at          timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_resume_suggestions_version_status ON resume_version_suggestions (resume_version_id, status);
CREATE INDEX idx_resume_suggestions_tailor_run     ON resume_version_suggestions (tailor_run_id);
```

#### 1.2 `resume_assets` additive 字段补充

```sql
ALTER TABLE resume_assets
    ADD COLUMN IF NOT EXISTS source_type           text CHECK (source_type IS NULL OR source_type IN ('upload', 'paste', 'guided')),
    ADD COLUMN IF NOT EXISTS original_text         text,
    ADD COLUMN IF NOT EXISTS guided_answers        jsonb,
    ADD COLUMN IF NOT EXISTS parsed_text_snapshot  text;
```

#### 1.3 跑 `make migrate-up` 验证

- 干净 DB 上 apply 成功
- 包含现有 `resume_assets` + `resume_tailor_runs` 数据的 baseline 上 apply 成功（非空数据兼容）
- `psql -c "\d+ resume_versions"` / `\d+ resume_version_suggestions` 输出包含期望字段与 idx
- check constraint 验证：尝试插入非法 `version_type='foo'` / `seed_strategy='bar'` / `source_type='unknown'`，PG 必须拒绝（错误码 23514）

### Phase 2: Migration down - 反向操作 + 幂等

#### 2.1 新增 `migrations/000NNN_resume_versions.down.sql`

```sql
DROP INDEX IF EXISTS idx_resume_suggestions_tailor_run;
DROP INDEX IF EXISTS idx_resume_suggestions_version_status;
DROP TABLE IF EXISTS resume_version_suggestions;

DROP INDEX IF EXISTS idx_resume_versions_parent;
DROP INDEX IF EXISTS idx_resume_versions_asset_type;
DROP INDEX IF EXISTS idx_resume_versions_user_updated;
DROP TABLE IF EXISTS resume_versions;

ALTER TABLE resume_assets
    DROP COLUMN IF EXISTS parsed_text_snapshot,
    DROP COLUMN IF EXISTS guided_answers,
    DROP COLUMN IF EXISTS original_text,
    DROP COLUMN IF EXISTS source_type;
```

#### 2.2 跑 `make migrate-down` + `make migrate-up` 幂等验证

- `make migrate-down` 在已 apply 状态下成功（必要时 `MIGRATE_DOWN_FORCE=1` + APP_ENV!=prod）
- 再跑 `make migrate-up` 重新 apply 成功（确认 IF NOT EXISTS / IF EXISTS 路径都正确）
- `psql -c "\d+ resume_assets"` 验证 down 后 `source_type` / `original_text` / `guided_answers` / `parsed_text_snapshot` 已删除

### Phase 3: enum-sources + privacy matrix + spec 同步

#### 3.1 修订 `migrations/enum-sources.yaml`

按 B4 D-6 enum source matrix 登记：

```yaml
resume_versions.version_type:
  source: shared-conventions-codified.ResumeVersionType
  values: [structured_master, targeted]
  authority: B1
resume_versions.seed_strategy:
  source: shared-conventions-codified.ResumeSeedStrategy
  values: [copy_master, blank, ai_select]
  authority: B1
  nullable: true
resume_version_suggestions.status:
  source: shared-conventions-codified.ResumeTailorSuggestionStatus
  values: [pending, accepted, rejected]
  authority: B1
resume_assets.source_type:
  source: openapi-v1-contract.RegisterResumeRequest.sourceType
  values: [upload, paste, guided]
  authority: B2  # B2 D-18 RegisterResumeRequest.sourceType
  nullable: true
```

跑 `migrations/lint.sh` 验证未登记 enum 拒绝、新登记 enum check constraint 与 SQL 一致。

#### 3.2 验证 privacy deletion matrix 与新表对齐

确认 B4 spec §3.1.2 已含 `resume_versions` / `resume_version_suggestions` 行（已在 B4 spec 1.14 中追加）；本 plan 在 down 测试时验证 backend internal runner 假设的删除顺序（先 suggestions FK 子表 → 后 versions → 后 assets）可由 PG FK + CASCADE 完成。

#### 3.3 跨 spec 同步

修订 B4 spec.md 1.14 → 1.15：
- D-17 措辞从"声明阶段"改为"已落地"
- §2.1 表 inventory 从"26 → 28 拟新增"改为"28 表"（明确数字）
- spec.md / history.md 追加 1.15 行（关联本 plan）

修订 `docs/spec/INDEX.md` db-migrations-baseline 版本与日期 1.14 → 1.15。

### Phase 4: 验收与下游同步

#### 4.1 跨 gate 收口

按 §3 替代验证 gate 依序运行：
- `make migrate-up && make migrate-down && make migrate-up` PASS（双向幂等）
- `make migrate-check` PASS
- `cd backend && go test ./internal/migrations/... ./internal/store/...` PASS（含本 plan 新 schema test）
- `migrations/lint.sh` PASS（enum-sources 同步）
- B4 spec / history / INDEX 同步 by `sync-doc-index --check`

#### 4.2 通知下游 owner

- 通知 `backend-resume` / `backend-upload` 未来 subspec owner：`resume_versions` / `resume_version_suggestions` 表 + `resume_assets` 字段已就位，可启动 store / handler 设计；
- 通知 `openapi-v1-contract/004-resume-additive-coverage` owner：B4 D-17 已落地，B2 D-18 中 `branchResumeVersion` / `updateResumeVersion` 等 schema 的 backing store 可全字段消费；
- 通知 `event-and-outbox-contract/002-resume-tailor-mode-drift-fix` owner：B4 D-17 已落地，与 events `ResumeTailorMode` 漂移修复（B3 plan 002）并行推进；
- 通知 `engineering-roadmap` owner：`docs/spec/engineering-roadmap/spec.md` §1 表 inventory 描述（如有 "25 应用表"）同步升级到 28。

### Phase 5: L2 remediation - live test cleanup hardening

#### 5.1 修复固定 UUID live test 清理顺序

`backend/internal/migrations/resume_versions_integration_test.go` 使用固定 UUID seed live Postgres。清理必须先删除 `resume_version_suggestions` / `resume_versions` 等 child rows，再删除 `resume_assets` / `target_jobs` / `users`，并在清理失败时报告测试错误；不得忽略 `users` 删除被 FK 阻止导致的残留。

#### 5.2 验证重复运行不泄漏

Focused gate 必须证明 `TestResumeVersions*` 在当前环境可重复执行；无 `DATABASE_URL` 时记录 skip 事实，不得把 no-op 当作 live DB cleanup 证据。

### Phase 6: D-20 简历扁平化 flatten migration

> product-scope D-20 / B4 本地 D-22。把简历版本树扁平化为单一 `resumes` 资产，删除 version/suggestion/tailor-run 三表；同时回填 product-scope D-17 jd_match drop 此前未在 B4 spec/history 登记的计数漂移。新增 `migrations/000015_resume_flatten.{up,down}.sql`。Red 优先：先写 check / cascade negative test，再落 migration。

#### 6.1 flatten up migration

新增 `migrations/000015_resume_flatten.up.sql`：
- `ALTER TABLE resume_assets RENAME TO resumes`；`ALTER TABLE resumes ADD COLUMN structured_profile jsonb NOT NULL DEFAULT '{}'::jsonb`、`ADD COLUMN display_name text`；
- `ALTER TABLE resumes DROP COLUMN guided_answers`；`source_type` CHECK 由 {`upload`, `paste`, `guided`} 改为 {`upload`, `paste`}（drop 旧 constraint + add 新 constraint）；保留 `original_text` / `parsed_text_snapshot` / `raw_text` / `file_object_id`（D-20 只读保留原始来源 + 解析快照）；
- `ALTER TABLE practice_plans RENAME COLUMN resume_asset_id TO resume_id`（FK 跟随到 `resumes`；如约束名固化则显式 drop/add 同名 FK ON DELETE SET NULL）；
- 按 FK 依赖反序 `DROP TABLE resume_version_suggestions;` → `DROP TABLE resume_tailor_runs;` → `DROP TABLE resume_versions;`。

（验证：`make migrate-up` 干净 DB 成功 + `\d+ resumes` 含 `structured_profile` / `display_name` + 三表不存在 + `\d practice_plans` 含 `resume_id`）

#### 6.2 structured_profile backfill

Go backfill `backend/internal/migrations/backfills/000015/`：从每个 resume 的旧 `structured_master` version（`resume_versions.version_type='structured_master' AND deleted_at IS NULL`）回填 `resumes.structured_profile` / `display_name`；无 master 时留空 `{}` / NULL。注册 `migrations/backfill/manifest.yaml`（version `000015`，reversible=false，支持 dry-run / apply，写 `schema_backfills`）。backfill 必须在 DROP `resume_versions` 之前执行（顺序：rename+addcol → backfill → drop tables），或在同一 up 内以临时读取保证数据可达。

（验证：backfill dry-run + apply 在含历史 master/targeted 数据的 DB 上正确回填 structured_master 内容、无 master 留空；`go test ./internal/migrations/backfills/...` PASS）

#### 6.3 down migration

新增 `migrations/000015_resume_flatten.down.sql`：`resumes` RENAME TO `resume_assets`；DROP `structured_profile` / `display_name`，re-add `guided_answers jsonb`，`source_type` CHECK 恢复 {`upload`, `paste`, `guided`}；`practice_plans.resume_id` RENAME TO `resume_asset_id`；恢复 `resume_versions` / `resume_version_suggestions` / `resume_tailor_runs` 表结构骨架（数据入 backfill log，不强制恢复行，符合 D-4 可逆要求）。

（验证：`make migrate-down`（`MIGRATE_DOWN_FORCE=1` + APP_ENV=dev）+ `make migrate-up` 双向幂等循环 3 次 PASS）

#### 6.4 enum-sources + check negative test

`migrations/enum-sources.yaml`：移除 `resume_versions.version_type` / `resume_versions.seed_strategy` / `resume_version_suggestions.status` 三个 enum source；`resume_assets.source_type` 行改为 `resumes.source_type`，值收敛 {`upload`, `paste`}。check negative test：插入 `resumes.source_type='guided'` 必须被 PG 拒绝（错误码 23514）。

（验证：`migrations/lint.sh` PASS + `go test ./internal/migrations/... -run TestResumeFlattenCheckConstraints` PASS + lint negative fixture exit 2）

#### 6.5 spec / history / privacy matrix 同步

B4 spec 1.22→1.23（本次修订已完成）：§2.1 表 inventory 25 张（含 jd_match drop 回填）、§3.1.2 privacy matrix（`resumes` 行 + 删除 version/jd_match 行 + debriefs 行收敛）、§6 C-1（25 张 / ≥30）、新增 D-22 决策行、标注 D-17 / B4 本地 D-20 退役；history.md 1.23 行。

（验证：`sync-doc-index --check` 零漂移 + B4 spec §2.1 / §3.1.2 引用一致）

#### 6.6 跨 gate 收口 + 下游信号

- 依序 PASS：`make migrate-up && make migrate-down && make migrate-up` + `make migrate-check` + `cd backend && go test ./internal/migrations/... ./internal/store/...` + `migrations/lint.sh` + `sync-doc-index --check`（无 `DATABASE_URL` 时 live 明确 skip，contract / lint / negative fixture 仍 PASS）。
- 下游信号：通知 `backend-resume`（store 改 `resumes` 单表 + 删 versions/tailor/suggestion store）、`openapi-v1-contract/004`（B2 resume 契约坍缩 + resumeId rename）、`shared-conventions-codified`（3 enum 退役已锁）、`backend-practice`（session resume_id binding）。

（验证：cross-plan 引用 commit + 上述 gate 全 PASS）

## 5 验收标准

- 本计划列出的 §4 所有 Phase task 全部完成
- §3 替代验证 gate 全部通过
- `migrations/000NNN_resume_versions.{up,down}.sql` 已落地并幂等可逆
- `migrations/enum-sources.yaml` 已登记 4 项新 enum source
- B4 spec.md 1.14 → 1.15，§2.1 表数 28，D-17 措辞从"声明阶段"改为"已落地"
- B4 spec §3.1.2 privacy deletion matrix 含 `resume_versions` / `resume_version_suggestions` 行
- backend-resume / openapi-v1-contract/004 / event-and-outbox-contract/002 / engineering-roadmap owner 已收到落地信号

**D-20 flatten（Phase 6）验收**（product-scope D-20 / B4 D-22）：

- `migrations/000015_resume_flatten.{up,down}.sql` 已落地并双向幂等可逆：`resume_assets`→`resumes` + `structured_profile` / `display_name` + structured_master backfill；`practice_plans.resume_asset_id`→`resume_id`；`resume_versions` / `resume_version_suggestions` / `resume_tailor_runs` 三表已 drop；`guided_answers` 列删除、`source_type` 收敛 {`upload`, `paste`}
- `migrations/enum-sources.yaml` 移除 3 个 resume version enum source（version_type / seed_strategy / suggestion status），`resumes.source_type` 来源值收敛 {`upload`, `paste`}
- B4 spec.md 1.22 → 1.23：§2.1 表数 25、§3.1.2 privacy matrix（`resumes` 行 + 删 version/jd_match 行）、§6 C-1 表数 25 / ≥30、新增 D-22 决策、回填 product-scope D-17 jd_match drop 计数漂移
- 跨 gate（`make migrate-up/down/check`、`go test ./internal/migrations/...`、`migrations/lint.sh`、`sync-doc-index --check`）PASS；零 `resume_versions` / `resume_version_suggestions` / `resume_tailor_runs` / `resume_asset_id` 残留（除负向断言与 down migration 骨架恢复）

## 6 风险与应对

| 风险 | 应对 |
|------|------|
| R1: `parent_version_id` self-FK 在 cascade delete 时形成循环 | ON DELETE SET NULL 或不设置 cascade，业务层显式删除子节点；本 plan 默认不级联（依赖 backend-resume 的 store 逻辑控制 tree 删除顺序） |
| R2: `resume_tailor_runs.mode` enum 当前 check constraint 是 `[gap_review, bullet_suggestions]`，与 B3 events `[inline, rewrite, mirror]` 漂移 | 本 plan 不修订 `resume_tailor_runs.mode`；B3 plan 002 修订 events 层；本 plan checklist 引用 B3 plan 002 同步信号 |
| R3: 非空 baseline 数据 + additive column 默认值导致全表 rewrite | `ALTER TABLE ADD COLUMN ... text NULL` 在 PG 12+ 是 metadata-only 操作（无 default 或 NULL default），不触发 rewrite；本 plan Phase 1.2 验证此假设 |
| R4: `structured_profile jsonb` 字段未来扩展可能引入 schema drift | 本 plan 不锁 jsonb 内部 schema；由 `backend-resume` 落地 JSON schema validator；F3 prompt registry 通过 `output_schema_version` 跟踪 |
| R5: P1 `resume_version_edits` 表延后引入字段 / FK 不一致 | 本 plan 在 down migration 中预留扩展点（如 `parent_version_id` 已支持版本链）；P1 plan 加表时只需 additive |
| R6: B1 D-10 enum 字面量未实际落到 `shared/conventions.yaml`（B1 仍是声明阶段） | 本 plan check constraint 字面量与 B1 D-10 声明保持一致；如 B1 D-10 在落地时改值，本 plan 必须同步修订 check constraint，重新跑 migrate |
