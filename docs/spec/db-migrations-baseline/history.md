# DB Migrations Baseline Change Log

> **版本**: 1.36
> **状态**: active
> **更新日期**: 2026-07-13

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-07-13 | 1.36 | Practice user message 增加 durable reply status，支持刷新后按原 clientMessageId 恢复 pending/retryable/terminal/complete 状态。 | db-migrations-baseline/001 Phase 11 + backend-practice/002 |
| 2026-07-13 | 1.34 | Supersede 1.31/1.33中的durable report-attempt存储：`000018`只保留summary、generation context与dimension focus，显式删除/拒绝`llm_attempt_count`及同义产品retry列；动作内retry归backend-review内存上下文。 | db-migrations-baseline/001 Phase 9 + backend-review/001 |
| 2026-07-13 | 1.33 | 将单次 repair flag 替换为 `feedback_reports.llm_attempt_count` 0..4 durable pre-call CAS，锁定 crash/replay 不回退与无第 5 次调用。 | db-migrations-baseline/001 Phase 8 + backend-review/001 |
| 2026-07-12 | 1.32 | 修复 current inventory 为 21 app + 3 auth + 2 metadata / public >=26，并补 C-13 grounded report migration 验收。 | db-migrations-baseline/001 Phase 8 |
| 2026-07-12 | 1.31 | 新增 000018 grounded report summary/context/durable repair，并将 report/plan focus 列收敛为 dimension codes。 | db-migrations-baseline/001 Phase 8 + backend-review/001 |
| 2026-07-12 | 1.30 | 新增 000017 PracticePlan round identity pair/check/index 与 v000017 唯一时长 legacy backfill；TargetJob 不新增可变 progress 列。 | db-migrations-baseline/001 Phase 7 |
| 2026-07-12 | 1.29 | `practice_messages` 取代 `practice_turns`，删除 `question_assessments`，app table 总数调整为 21。 | backend-practice 001-003 / backend-review 001 |
| 2026-07-10 | 1.28 | 删除 tracked `baseline_noop` backfill manifest 与空 registry；backfill runner 继续支持真实行级 manifest，缺失 manifest 时跳过。 | tech-debt pruning |
| 2026-07-07 | 1.27 | 压缩为当前 net-state migration contract：22 app tables、3 auth support tables、2 metadata tables、flat Resume schema、privacy matrix、backfill ledger and executable gates. | product-scope/001-core-loop-module-pruning Phase 6.119 |
| 2026-07-07 | 1.26 | 对齐 active spec 宽口径措辞，保持当前 schema inventory、job/check matrix 与 privacy matrix 一致。 | product-scope/001-core-loop-module-pruning Phase 6.84 |
| 2026-07-06 | 1.25 | 对齐 current job/check source matrix，并用 focused migration contract test 固化 DB final-state guard。 | product-scope/001-core-loop-module-pruning Phase 6.9 |
| 2026-07-06 | 1.24 | 对齐当前 22 app table inventory、25 app/auth support table inventory 与 public schema count >= 27 gate；补齐 `idempotency_records` privacy disposition. | product-scope/001-core-loop-module-pruning Phase 6 + B4 privacy matrix reconcile |
| 2026-06-13 | 1.23 | 落地 flat Resume migration net-state：`resumes` 增加 `structured_profile` / `display_name`，`source_type` 收敛为 `upload` / `paste`，practice plan binding 使用 `resume_id`。 | db-migrations-baseline/002 flat Resume phase |
| 2026-05-26 | 1.22 | `privacy_requests.user_id` 改为 nullable + `ON DELETE SET NULL`，支撑 account deletion tombstone。 | backend-async-runner/001 BUG-0106 remediation |
| 2026-05-15 | 1.19 | 授权 report generation 所需 `ai_task_runs` 与 `feedback_reports` typed columns。 | backend-review/001 Phase 0 |
| 2026-05-15 | 1.18 | 授权 `practice_session_events.replay_payload`，分离 redacted event payload 与 idempotent replay snapshot。 | backend-practice/003 L2 replay follow-up |
| 2026-05-14 | 1.17 | 授权 practice hint task type and AI task persistence columns. | backend-practice/003 Phase 0 |
| 2026-05-12 | 1.16 | 固化 live migration tests 的 fixed-UUID cleanup gate。 | db-migrations-baseline/002 Phase 5 |
| 2026-05-09 | 1.13 | 增加 shared `idempotency_records` table、unique key 与 expiration index。 | backend-practice/001 Phase 0 |
| 2026-05-08 | 1.12 | 本地迁移验证前提升级为 Postgres 18。 | local-dev-stack/001 |
| 2026-04-29 | 1.5 | 增补 outbox retry columns、AI call meta typed columns、privacy deletion table matrix。 | plan-review remediation |
| 2026-04-29 | 1.3 | 纳入 auth support tables、`schema_backfills`、backfill ledger、enum/check source matrix 与 module topology。 | plan-review remediation |
| 2026-04-27 | 1.0 | 初始创建 migration tool、directory、naming、rollback、backfill、prod guard and enum/check constraints contract. | engineering-roadmap/001 Phase 3 |
