# DB Migrations Baseline Bootstrap

> **版本**: 1.23
> **状态**: active
> **更新日期**: 2026-07-14

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [db-migrations-baseline spec](../../spec.md) 当前锁定的迁移工具、20 张应用表、3 张 auth 支撑表、2 张迁移元数据表、B3 outbox retry operational columns、A3/F1 `ai_task_runs` typed columns、enum/check 来源矩阵、backfill ledger 与 P0 privacy deletion matrix 落到 `migrations/` 与 `backend/cmd/migrate`。TargetJob 当前 net-state 只保留 `raw_jd_text`，不保留来源列/表、JD attachment purpose 或 JD source refresh jobType；独立 `source_records` 与 resume/privacy purpose 保留。

Phase 1-10 保留为既有交付证据；当前待执行 schema 合同由 Phase 11（Practice generation/lease recovery）与 Phase 12（TargetJob report pointer removal）覆盖。

本 plan 不实现业务 repository、不实现 C8 dispatcher、不实现 privacy_delete runner；只提供 schema baseline、迁移可执行入口、lint/check gate 与下游 handoff。

## 2 背景

B4 是 Layer B contract 的 schema owner。A2 已提供 Postgres 18 本地实例；B1/B2/B3 分别提供 shared enum、API-facing async enum、event/job manifest；A3/F1 需要 `ai_task_runs` typed columns；C8 需要 `outbox_events` retry columns 与 privacy deletion matrix。B4 001 必须在后续 C/D 域 implementation 读取真实表之前完成。

## 3 质量门禁分类

- **Plan 类型**: `migration + tooling + code-internal + contract`。本 plan 修改 SQL migrations、Go migration wrapper、backfill registry、lint/probe 脚本与 Make target，属于内部 schema contract 和迁移工具链交付；不引入用户可感知 UI、HTTP API 行为、业务流程或端到端产品路径。
- **TDD 策略**: 必须通过 `/tdd --file docs/spec/db-migrations-baseline/plans/001-bootstrap/checklist.md --references docs/spec/db-migrations-baseline/plans/001-bootstrap/plan.md,docs/spec/db-migrations-baseline/spec.md --phase-commit db-migrations-baseline/001-bootstrap` 顺序执行。每个 checklist item 以本 checklist 内的 `验证:` 子句作为 Red-Green-Refactor 断言来源；涉及 migration wrapper、lint、backfill registry、privacy dry-run 或 SQL probe 的 item 必须先补 focused failing test / smoke / probe，再最小实现并复跑对应命令。
- **BDD 策略**: BDD 不适用。本 plan 只交付 DB schema baseline、migration wrapper、lint/probe 与内部 dry-run 工具，不产生浏览器 UI、外部 API、用户业务流程或 scenario-test 可观察行为，因此不创建 `bdd-plan.md` / `bdd-checklist.md`，主 checklist 也不设置 `BDD-Gate:`。
- **替代验证 gate**: 使用内部迁移与契约 gate 代替 BDD：Go wrapper/backfill 单元测试、`make migrate-check`、`APP_ENV=prod make migrate-down` fail-fast smoke、migration enum/check lint 与 drift negative case、SQL table/column/index/explain probes、privacy matrix dry-run 覆盖检查、`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`。

## 4 实施步骤

### Phase 1: migration skeleton 与工具 wrapper

#### 1.1 `golang-migrate` wrapper

落地 `backend/cmd/migrate/main.go`，包装 `golang-migrate v4.18+` 与 B4 backfill registry；根 `Makefile` 提供 `make migrate-up` / `make migrate-down` / `make migrate-status` / `make migrate-create NAME=...` / `make migrate-check`，并让根 `make migrate` 指向 `migrate-up` 或清晰 help。

#### 1.2 migration 目录与命名

在 `migrations/` 使用 `NNNNNN_<verb>_<noun>.up.sql` / `.down.sql`，编号从 `000001` 起 6 位连续；所有新增 migration 必须由 `make migrate-create NAME=...` 生成，不手敲编号。

#### 1.3 schema_backfills

`000001` baseline 不启用未使用 DB extension；down migration 不管理 extension 生命周期。落地 `schema_backfills` ledger 表与 backfill manifest contract；当前 `v000017/practice_plan_round_identity` 是已登记的真实行级回填，按 dry-run/apply 写 ledger 并保持幂等。

### Phase 2: baseline DDL 与索引

#### 2.1 24 张当前应用 / auth 支撑表

落地当前产品范围内的 21 张应用表，加 ADR-Q1 的 `auth_challenges` / `sessions` / `external_identities` 3 张支撑表；final schema 只保留这 24 张当前应用 / auth 支撑表。`make migrate-up` 后 public schema 至少 26 张表（含 `schema_migrations` / `schema_backfills`）。

#### 2.2 B3 outbox / async columns

`outbox_events` 必须包含 `publish_attempts` / `next_attempt_at` / `locked_at` / `last_error_code` / `last_error_message`；pending due 查询索引至少覆盖 `(publish_status, next_attempt_at, created_at)`。历史 Phase 2 曾交付含 `source_refresh` 的 8 项 jobType；Phase 10 已 supersede 该事实。当前 `async_jobs.job_type` check 必须精确匹配 B3 的 7 个 canonical jobType，只有 `email_dispatch` 是 B2 六项 API-facing subset 之外的 internal-only job。

#### 2.3 A3/F1 AI typed columns

`ai_task_runs` 必须包含 `model_family` / `model_profile_name` / `model_profile_version` / `fallback_chain jsonb not null default '[]'::jsonb` / `route` / `validation_status` / `output_schema_version`；dashboard 核心查询不得依赖 JSONB path scan。

#### 2.4 索引覆盖

覆盖 B4 B-Tree 索引与可选 `target_jobs` GIN 全文索引；用 SQL probe / explain 验证关键查询走索引。

### Phase 3: enum/check 来源、backfill 与 privacy lint

#### 3.1 enum/check 来源 lint

落地 `migrations/enum-sources.yaml` 与 `scripts/lint/migrations_lint.py`（或等价 Go 工具），逐列登记 `table.column -> source -> checksum`。SQL 中出现未登记 `check (col in (...))` 必须 fail；B1/B2/B3 修改 manifest 后 `make migrate-check` 必须发现漂移。

#### 3.2 backfill registry

`backend/internal/migrations/backfills/<version>/` 注册 dry-run/apply 函数；同一 `version + mode + checksum` apply 成功后不得重复执行，除非 `--force` 且 `APP_ENV!=prod`。

#### 3.3 privacy deletion dry-run

提供 `make privacy-delete-dry-run` 或 `backend/cmd/migrate privacy-matrix --dry-run` 等入口，读取 spec §3.1.2 table matrix，对测试用户输出每表 disposition（hard delete / cascade / retain / audit tombstone），不执行真实删除；C8 后续 plan 消费该矩阵。

### Phase 4: Verification + handoff

#### 4.1 migrate-check

在干净 Postgres 18 上执行 `make migrate-check`：`migrate-up -> migrate-down -> migrate-up` 全部成功；`APP_ENV=prod make migrate-down` 必须 fail-fast；`schema_backfills` ledger 无重复成功记录。

#### 4.2 table / column / index probes

运行 SQL probes：public table count ≥26；`outbox_events` retry columns 存在；`ai_task_runs` typed columns 存在；pending due / B-Tree 索引存在且 explain 命中。

#### 4.3 enum / privacy probes

临时修改 B3 job manifest 或 B1 enum，确认 `make migrate-check` / lint 报 drift；privacy deletion dry-run 输出覆盖 spec §3.1.2 所有表组，且 `prompt_versions` / `rubric_versions` / migration metadata 被 retain。

#### 4.4 文档与 INDEX

本 plan checklist 全部勾选后，将 Header 切 completed，运行 sync-doc-index check/fix，并在 work journal 记录 migrate-check、SQL probes、lint probes 与 downstream handoff。

### Phase 5: product-scope v1.2 schema remediation

#### 5.1 Red: migration inventory 期望排除 `mistake_entries`

先调整 migration contract test / lint 的期望：public schema count gate 跟随当前 B4 spec，table inventory 不再允许 `mistake_entries`，B1 enum source 不再登记 `MistakeStatus`。旧 SQL / privacy matrix / generated probes 仍引用 `mistake_entries` 时必须失败。

#### 5.2 Green: 修订 SQL、enum source 与 privacy matrix

修订 baseline migration：删除 `mistake_entries` 表及其索引 / FK / privacy disposition；`target_jobs.open_mistake_count` 改为 `open_question_issue_count`；`question_assessments.written_to_mistake_book` 改为 `included_in_retry_plan`；`practice_plans` 的 mode / goal check 跟随 B1 v1.7；如需要记录报告内题目状态，使用 `question_assessments.review_status` 引用 `QuestionReviewStatus`。

#### 5.3 Verify

运行 `make migrate-check` 或可用的 migration lint / SQL contract tests；repo 搜索确认实现侧不再出现 `mistake_entries`、`open_mistake_count`、`written_to_mistake_book`、旧 practice mode / goal check 值。

### Phase 6: Migration CLI test-double cleanup

#### 6.1 Keep test environment adapters out of production

删除 `backend/internal/migrations` 生产包中只供 `cli_test.go` 使用的导出 `StaticEnv` 类型及方法，将等价 map-backed test double 放回测试文件。`Run` 继续只依赖 `Env` 接口，`cmd/migrate` 继续提供唯一生产 `osEnv` adapter；nil-env 错误只描述接口要求，不引用测试专用具体类型。

### Phase 7: normalized PracticePlan round identity

#### 7.1 RED migration contracts

先让 migration contract / live Postgres test 失败于缺失 `practice_plans.round_id` / `round_sequence`、pair CHECK、positive-sequence CHECK 与 partial lookup index；负向覆盖只写一列、sequence <= 0、相邻轮次相同时长和重复执行。

#### 7.2 DDL and legacy backfill

必须通过 `make migrate-create NAME=practice_plan_round_identity` 创建 `000017` up/down。两列保持 nullable 只为 legacy compatibility；新写入由 backend-practice 强制成对。legacy 行级回填通过现有 backfill registry 执行：先要求 plan 与 TargetJob 的 user/current-resume 绑定一致、TargetJob 未删除、sequence 为正 int32，再仅当 `target_jobs.summary.interviewRounds[]` 中恰好一个轮次的 `durationMinutes` 等于 plan `time_budget_minutes` 时写入 canonical `round-{sequence}-{type}` 与 sequence；零/多匹配、错绑、删除或溢出保持 null，并写可审计 ledger。不得按数组第一项、TargetJob lifecycle status 或固定轮次表猜测；实现语义变更必须重算 manifest checksum。

#### 7.3 Index and rollback

建立只覆盖非 null identity 的 TargetJob/round lookup index，服务当前 plan 与完成事实投影。down 移除 index/check/columns；backfill rollback 按 registry contract 处理，不能让歧义 legacy 数据变成伪造轮次。

#### 7.4 Verification

运行 migration lint/contract tests、backfill registry tests、真实 Postgres `make migrate-check`，验证 up/down/up、pair invariant、唯一匹配、wrong-resume/deleted/int32-overflow/歧义保持 null、checksum ledger 幂等，并确认 schema 中没有 TargetJob progress 列。

### Phase 8: HISTORICAL-SUPERSEDED grounded direct report storage

#### 8.1 RED contract

Migration tests must require nullable-until-ready `feedback_reports.summary`, content-bearing `generation_context`, durable `llm_attempt_count integer NOT NULL DEFAULT 0 CHECK (0..4)`, `retry_focus_dimension_codes` and `practice_plans.focus_dimension_codes`, while rejecting the superseded boolean repair flag and old competency column names. The current final inventory remains exactly 21 app + 3 auth + 2 metadata tables.

#### 8.2 DDL

Create reversible `000018_grounded_report_context` through `make migrate-create`. Existing development rows receive an empty object only to make migration executable; runtime treats non-`report-context.v1` as invalid and never fabricates/backfills sensitive context. New completion writes current context atomically. Before every product generation provider call, repository CAS increments `llm_attempt_count` only while `<4`; a consumed attempt is never rolled back after timeout, crash, replay or worker takeover. Attempt 4 may finish ready or terminal failed, but no path may issue attempt 5. Rename report/plan focus columns to dimension semantics; no compatibility columns or trigger mirrors.

#### 8.3 Privacy and verification

The new summary/context/attempt coordinate remain covered by hard deletion of `feedback_reports`; probes confirm user content never appears in audit/job/outbox/log tables. Run migration lint/contract, clean-DB and populated-DB up/down/up, current-row invalid-context, concurrent CAS 1..4, attempt4 exhaustion, crash/replay no-rollback, rename/down restoration and full privacy matrix probes.

`db-migrations-baseline/001` is the sole producer of `REPORT_STORAGE_V18_PASS`. Emit it only after `000018` passes clean and populated PostgreSQL up/down/up, invalid-current-context, rename/down restoration, privacy/non-content leakage, `make migrate-check`, migration lint and backend migration tests. Downstream report owners only consume this marker.

> Phase 8中的summary/context/focus、privacy与可逆迁移证据继续有效；其中`llm_attempt_count`、pre-call CAS及crash/replay global-cap内容已被Phase 9取代，不再代表current contract。

### Phase 9: remove durable product retry storage

#### 9.1 Current-shape RED contract

Migration lint与SQL contract必须要求nullable-until-ready `feedback_reports.summary`、content-bearing `generation_context`、`retry_focus_dimension_codes`、`practice_plans.focus_dimension_codes`及当前21+3+2 inventory，同时断言`llm_attempt_count`和任何同义产品retry column为零命中。

#### 9.2 Reconcile `000018` in place

项目未上线，不创建兼容迁移：原地修订可逆`000018_grounded_report_context`，删除attempt列up/down DDL，只保留summary/context/focus current shape。Populated development rows仍只获得empty invalid context以完成迁移，runtime继续fail closed，不伪造敏感快照。

#### 9.3 Verification and owner marker

重新运行migration lint、clean/populated PostgreSQL up/down/up、invalid-context、rename/down restoration、privacy/non-content leakage与最终schema negative probes。`REPORT_STORAGE_V18_PASS`只有在最终schema确认无产品retry列后才能重新发出；provider调用次数、动作重置与10s/20s/40s不由DB gate证明。

### Phase 10: TargetJob paste-only schema net-state

#### 10.1 RED migration contract

先让 migration lint、SQL contract、privacy matrix 与 clean-DB inventory tests 要求 20 app + 3 auth + 2 metadata，并明确旧 TargetJob 来源列/表、JD attachment purpose 与 JD source refresh jobType 必须不存在；当前 baseline 仍含任一旧结构时记录 RED。BDD 不适用，因为本 Phase 只改变内部 schema/check；替代 gate 为 migration contract、enum-source lint、privacy dry-run 与 PostgreSQL up/down/up。

#### 10.2 GREEN baseline reconciliation

项目未上线，原地修订 baseline/up/down 与 `enum-sources.yaml`：删除 `target_jobs.source_type/source_url/source_file_object_id`、`target_job_sources`、JD attachment purpose 和 JD source refresh jobType；保留 `target_jobs.raw_jd_text`、独立 `source_records`、resume/privacy purpose。同步 migration inventory、privacy matrix 和相关 SQL contract，不创建兼容列、影子表或 sibling migration。

#### 10.3 Migration and zero-reference closure

运行 migration lint、focused backend migration tests、clean/populated PostgreSQL up/down/up 与 privacy dry-run。精确 zero-reference gate 覆盖 migrations/enum sources/backend migration probes；正向 probe 必须确认 `raw_jd_text`、`source_records`、resume/privacy purpose 仍存在，并输出最终 20+3+2 inventory。

### Phase 11: Practice reply-status net-state

#### 11.1 RED migration contract

先让 SQL contract、enum-source lint 与 populated PostgreSQL probe 失败于缺失 `practice_messages.reply_status`。RED 数据集必须同时包含 opening assistant、已完成 user/assistant pair、无 reply user 与跨 session 相同 client ID；不得只验证空表 DDL。BDD 不适用，因为本 Phase 是内部 schema；替代 gate 为 migration/store contract、privacy cascade 与真实 Postgres up/down/up。

#### 11.2 GREEN baseline reconciliation

项目未上线，在 baseline 中为 user message 增加 `reply_status`，allowlist 为 `pending / retryable_failed / terminal_failed / complete`，assistant 必须为 NULL；与 `client_message_id/reply_to_message_id` 角色约束合并校验。既有有 reply user 在 populated migration proof 中归一为 `complete`，无 reply user 保持可恢复状态；不新建平行状态表、浏览器 token 或兼容列。

#### 11.3 Migration and persistence closure

运行 migration lint、backend migration/store focused tests、clean/populated PostgreSQL up/down/up 与 privacy delete probe。正向证明 reply transition 与原唯一约束共存；负向证明非法状态、assistant 非空状态、complete 无 reply、重复 assistant 与跨用户读取均失败或不可见。

#### 11.4 Generation and lease fence remediation

在同一 baseline 表中增加仅供后端使用的 `reply_generation bigint` 与 `reply_lease_expires_at timestamptz`。user row 必须具有正 generation；仅 pending 必须具有 lease，retryable/terminal/complete 必须清空 lease；assistant 的 `client_message_id/reply_status/reply_generation/reply_lease_expires_at` 全为 NULL。新 reserve 从 generation 1 开始并写 `Now + 90s`，同 ID retry 每次加一；GET 与同 ID reserve 惰性把过期 pending 收敛为 retryable，commit/fail 使用 expected generation CAS，防止 G1 在 G2 后迟到落库。字段不进入 OpenAPI，也不引入 worker/scheduler。

### Phase 12: TargetJob report pointer removal

#### 12.1 RED schema contract

先让 SQL/store/OpenAPI contract 断言 `target_jobs.latest_report_id` 与 public `TargetJob.latestReportId` 必须不存在，同时证明 `feedback_reports`、`generation_context` 与 TargetJob ownership 仍可查询。BDD 不适用：本 Phase 只移除内部去规范化列；替代 gate 为 migration contract、real PostgreSQL up/down/up、backend-review overview integration 与 zero-reference。

#### 12.2 GREEN baseline reconciliation

项目未上线，原地从 `000001_create_baseline.up/down.sql`、TargetJob store scan/insert/update 与 fixtures/generated surface 删除该列，不创建兼容 migration、trigger 或替代 pointer。当前 ready report 与 latest attempt 由 backend-review 基于 `feedback_reports` 和冻结 canonical-round context 查询投影。

#### 12.3 Migration and zero-reference closure

运行 migration lint、clean/populated PostgreSQL up/down/up、TargetJob store/review integration 与 privacy cascade；精确搜索 production/generated/OpenAPI/fixtures/migrations 中旧列/字段为零，同时正向证明报告行、冻结 context、用户隔离与 `generated_at/created_at/id` 排序所需列仍存在。

## 5 验收标准

- spec §6 C-1..C-16 全部具备本 plan 或下游 handoff 证据；C8/F1/C11 等运行时验证由各自 owner 后续关闭。
- `make migrate-check` 可在干净本地 DB 重复执行；prod down 防呆有效。
- enum/check 来源、B3 jobType manifest、A3/F1 AI typed columns、P0 privacy deletion matrix 都有可执行 lint/probe。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| baseline DDL 过大难以 review | 分 phase 编排，保持 migration 文件编号稳定；用 table/column/index inventory probe 辅助 review |
| internal-only jobType 误暴露到 B2 API | Phase 10 后 DB job check 只跟随 B3 当前 canonical 集合；JD source refresh 不得以 internal-only 名义残留，`email_dispatch` 仍按其 owner 合同保留 |
| migration down 误在 prod 执行 | wrapper 在 `APP_ENV=prod` 时拒绝 down，除非显式 `MIGRATE_DOWN_FORCE=1` 且执行环境允许 |
| backfill 重复 apply | `schema_backfills` 以 version/mode/checksum 做幂等 ledger；重复执行必须 fail 或 skip |
| privacy deletion matrix 漏表 | Phase 3.3 dry-run 输出必须覆盖所有 baseline 表组；新增表时 lint 要求同步 matrix |

## 7 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-07-14 | 1.23 | Reopen Phase 11 for generation/90s lease fencing and Phase 12 to remove the TargetJob latest-report pointer. | backend-practice/002 + backend-review/001 |
| 2026-07-13 | 1.21 | Reopen Phase 11 for durable Practice reply status and refresh-safe same-ID recovery. | backend-practice/002 + frontend-workspace-and-practice/002 |
| 2026-07-13 | 1.20 | Reopen Phase 10 to converge the TargetJob baseline to paste-only schema, 20+3+2 inventory, and migration zero-reference gates. | TargetJob paste-only current net-state |
| 2026-07-13 | 1.19 | Reopen Phase 9 to remove durable report retry columns/CAS from migration 000018, retain summary/context/focus, and revalidate reversible current shape plus privacy with explicit retry-column absence. | backend-review/001 action-local retry contract |
| 2026-07-13 | 1.18 | Close Phase 8 after current-shape SQL lint, durable CAS/replay tests, disposable PostgreSQL migrate-check and populated up/down/up privacy probes re-emit `REPORT_STORAGE_V18_PASS`. | backend-review/001 max-four generation contract |
| 2026-07-13 | 1.17 | Replace the single repair flag with durable `llm_attempt_count` 0..4 pre-call reservation and crash/replay-safe no-fifth-call probes. | backend-review/001 max-four generation contract |
| 2026-07-12 | 1.16 | Make Phase 8 the sole producer of REPORT_STORAGE_V18_PASS after full PostgreSQL/privacy verification. | backend-review/001 |
| 2026-07-12 | 1.15 | Reconcile current 21+3+2 inventory/public >=26 gate and add exact C-13 grounded report migration proof. | db-migrations-baseline 1.32 |
| 2026-07-12 | 1.14 | Reopen Phase 8 for grounded report summary/context/durable repair and dimension-focus storage. | backend-review/001 |
| 2026-07-12 | 1.13 | Reopen Phase 7 for normalized practice-plan round identity and auditable legacy backfill. | backend-practice round progression |
| 2026-07-10 | 1.12 | 将 migration CLI 的 map-backed Env test double 从生产包下沉到测试文件。 | tech-debt pruning |
| 2026-07-10 | 1.11 | 将 baseline inventory 改为当前 25 张应用/auth 支撑表正向合同，并统一 migration 负向 gate 术语。 | tech-debt pruning |
| 2026-07-10 | 1.10 | 技术债口径清理：将 `make migrate` handoff 描述收敛为当前根 Make target 委托，不改变迁移工具合同。 | tech-debt pruning |
| 2026-07-06 | 1.7 | product-scope D-17/D-20/D-22 后续收敛：本 completed bootstrap plan 的当前正向表数、public schema gate 与 B3/B2 job type 口径更新为 22 应用表 + 3 auth 支撑表、public schema ≥27、B3 8 canonical jobs、B2 6 API-facing jobs；历史删除表只保留在 remediation / history 语境。 | product-scope/001-core-loop-module-pruning Phase 6.10 |
| 2026-05-08 | 1.6 | 对齐 A2 用户决策：本地迁移验证前提升级为 Postgres 18。 | local-dev-stack/001 post-pass revision |
| 2026-05-03 | 1.4 | 修正 Phase 2 / Phase 4 中既有表数量口径：当时 baseline 为应用表 + auth 支撑表 + 迁移元数据表。 | readiness reconcile |
| 2026-05-08 | 1.5 | 对齐 A3 003 Phase 6：删除向量扩展、向量检索表/索引与 extension drop gate；当前 baseline 为 25 应用表 + 3 auth 支撑表 + 2 迁移元数据表，public schema count gate ≥30。 | ai-provider-and-model-routing/003 Phase 6 |
| 2026-05-03 | 1.3 | 原地 reopen，新增 Phase 5 remediation：按 product-scope v1.2 删除独立 `mistake_entries` 表，迁移字段改为报告题目回顾 / 本轮复练语义。 | db-migrations-baseline v1.6 |
| 2026-04-30 | 1.2 | 原地 reopen 001-bootstrap，修复 L2 code-review 发现的 prod down fail-fast 顺序、dev-only extension drop 限制、B1/B2/B3 enum source drift gate 与 ALTER TABLE check 发现能力。 | plan-code-review remediation |
| 2026-04-30 | 1.1 | 补齐 TDD/BDD 质量门禁分类与 checklist 可执行验证断言；确认 BDD 不适用并以 migration / lint / probe / smoke gate 替代。 | implement gate remediation |
| 2026-04-29 | 1.0 | 初始物化 B4 `001-bootstrap`：migration wrapper、baseline DDL、enum/backfill/privacy lint 与 verification handoff。 | plan-review remediation |
