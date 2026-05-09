# 001 — Plan and Session Orchestration Checklist

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-05-09

**关联计划**: [plan](./plan.md)

## Phase 0: 跨 spec 前置修订 + Preflight

- [x] 0.1 修订 `shared/conventions.yaml#enums[PracticeMode]` 删除 `legacy debrief replay value`，保留 `assisted` / `strict`；运行 `make codegen-conventions` regenerate Go / TS shared types；验证: `python3 scripts/lint/conventions_drift.py --repo-root .` + `cd backend && go test ./internal/shared/types ./internal/shared/errors ./internal/shared/idx ./internal/shared/ai` + frontend conventions focused tests
- [x] 0.2 修订 `shared/conventions.yaml` 错误码注册表新增 `PRACTICE_PLAN_NOT_FOUND` / `PRACTICE_SESSION_NOT_FOUND`；regenerate Go / TS / OpenAPI 错误码常量；验证: shared types 单元测试 + drift check
- [x] 0.3 修订 `openapi/openapi.yaml#components.schemas.PracticeMode` 删除 `legacy debrief replay value`（pre-launch baseline rebase per backend-practice spec D-21，无需 ADR）；新增两个错误码到 `ApiError.code` enum；regenerate Go server / TS client；如 `openapi/baseline/openapi-v1.0.0.yaml` 涉及 enum 变更，按 D-21 wording 原地 rebase 而非新建 v1.1.0；验证: `make openapi-diff` + `make docs-openapi` + fixture parity test + generated artifact drift check
- [x] 0.4 修订 B3 generated event refs（`shared/events/*` 中引用 `PracticeMode` 的 surface）二值化；regenerate B3 events；验证: `make codegen-events` + `make lint-events` + B3 contract test
- [x] 0.5 新建 migration `migrations/NNNNNN_practice_idempotency_baseline.up.sql` / `.down.sql`：① shared `idempotency_records` 表 + UNIQUE(user_id, domain, operation, idempotency_key_hash) + INDEX(expires_at)；② `practice_plans.mode` CHECK 收敛到 `IN ('assisted','strict')`；③ 同步 B4 spec §2.1 表 list 加入 `idempotency_records`；验证: golang-migrate up/down idempotent + `make migrate-status` + 单元测试断言 CHECK 拒绝 `legacy debrief replay value`
- [x] 0.6 同步 owner spec history append + Header bump（按 D-30 Q1=A integrator 模式）：`backend-practice/history.md` v1.4 行 + Header 1.4 / 2026-05-09；`shared-conventions-codified/history.md` v1.15 + Header 1.15；`openapi-v1-contract/history.md` v1.14 + Header 1.14；`event-and-outbox-contract/history.md` v2.2 + Header 2.2；`db-migrations-baseline/history.md` v1.13 + Header 1.13；同步 `docs/spec/INDEX.md` 5 个 spec 版本；验证: `/sync-doc-index --check` 无 drift
- [x] 0.7 同步更新 `backend-practice/spec.md` 到 v1.4：新增 D-30（Q1+Q2 决策）；§7 line 216 wording 微调以反映 integrator 模式；§5 模块边界表 "DB schema" 行登记 shared `idempotency_records` 归属说明；验证: spec lint + drift check
- [x] 0.8 F3 baseline preflight assert：grep `prompt-rubric-registry/001-baseline` Header 状态 = `completed`，且其 `Resolve` 对 3 个 practice feature_key 返回 valid `(prompt_version, rubric_version, model_profile_name)` 三元组；验证: preflight script + 单元测试断言（fail-fast 在 plan 001 启动 AI 实现前）
- [x] 0.9 Phase 0 legacy-negative grep：仓库范围 removed legacy mode literal 在 `PracticeMode` / `mode` / `practice_plans.mode` / `session.mode` / `PracticeDisplayContext.practiceMode` 上下文零出现；`PracticeGoal` 上下文允许保留；验证: CI grep gate
- [x] 0.10 Phase 0 commit + work-journal：单 phase commit `phase 0: backend-practice plan-001 contract preflight`；运行 `/work-journal` 一条记录

## Phase 1: Plan + Session 主流程 (success path) + idempotency replay 基础

- [x] 1.1 在 `backend/internal/middleware/idempotency/` 实现 shared idempotency middleware：消费 `idempotency_records` 表，提供 `(domain, operation)` 注入入口，支持 pending lock + success replay + per-user 隔离 + TTL expire；验证: middleware 单元测试 5 项（pending lock / success replay / per-user 隔离 / TTL expire / domain namespace 隔离）
- [x] 1.2 在 `backend/internal/api/practice/` 实现 `createPracticePlan` handler（仅 `goal='baseline'`，其它 goal 返回 `422 VALIDATION_FAILED` + detail 指向 future `004-derived-plans-debrief` owner）；走 idempotency middleware；写 `practice_plans(status='ready')` 短事务 + audit_events；返回 `201`；验证: handler 单元测试 + `openapi/fixtures/PracticePlans/createPracticePlan.json` parity contract test + repository 单元测试
- [x] 1.3 在 `backend/internal/api/practice/` 实现 `getPracticePlan` handler：user_id 过滤、404 + `PRACTICE_PLAN_NOT_FOUND`；验证: handler 单元测试（含越权 404 返回，不泄露存在性）+ contract test
- [x] 1.4 在 `backend/internal/practice/` 实现 startPracticeSession 三段式（baseline goal）：reservation 短事务写 `practice_sessions.status='queued'` → 事务外 F3 `Resolve("practice.session.first_question", language)` + A3 observed `AIClient.Complete` → 短事务 commit `practice_turns(turn_index=1, status='asked')` + `practice_session_events(seq_no=1, event_type='session_started')` + outbox `practice.session.started` + `practice_sessions.status='running'`；返回 `201` 含 `currentTurn`；验证: 集成单元测试（fake F3 + fake AIClient + DB lock 检测断言外部 AI 调用不在 tx 内）+ contract test
- [x] 1.5 在 `backend/internal/api/practice/` 实现 `getPracticeSession` handler：user_id 过滤、404 + `PRACTICE_SESSION_NOT_FOUND`；验证: handler 单元测试 + contract test
- [x] 1.6 outbox emit 与 B3 payload 对齐：`outbox_emitter` 单元断言 `practice.session.started` payload 与 `shared/events/practice.session.started.json` 一致 + piiBoundary 通过；验证: outbox 序列化单元测试
- [x] 1.7 BDD-Gate: 验证 `E2E.P0.022` 通过（`test/scenarios/e2e/p0-022-practice-plan-baseline-create-and-read/`）
- [x] 1.8 BDD-Gate: 验证 `E2E.P0.023` 通过（`test/scenarios/e2e/p0-023-practice-session-start-and-first-question/`）
- [x] 1.9 Phase 1 commit + work-journal

## Phase 2: 错误路径 + Idempotency 完备性

- [x] 2.1 实现 AI 失败映射（D-19）：timeout → `502 AI_PROVIDER_TIMEOUT`、invalid output → `502 AI_OUTPUT_INVALID`、secret missing → `502 AI_PROVIDER_SECRET_MISSING`、fallback exhausted → `503 AI_FALLBACK_EXHAUSTED`；session reservation 进入 `failed` + `failure_code`；envelope 不含 prompt / response 明文；验证: error_mapping 单元测试每错误码一例
- [x] 2.2 实现 reservation `failed_retryable` 语义（D-23）：首题失败时把 `idempotency_records.status='failed_retryable'` + `practice_sessions.status='failed'` + `failure_code` 记录；同 Idempotency-Key 重试可重新生成首题，成功后固化为 `succeeded` 并升级为 `running`；验证: 集成单元测试（fake AIClient 注入失败 → 重试成功）
- [x] 2.3 实现 idempotency body mismatch 拒绝（C-22）：同 key 不同 fingerprint 返回 `409 PRACTICE_SESSION_CONFLICT`，不执行第二个副作用；envelope 不泄露首次资源内容；验证: middleware 单元测试
- [x] 2.4 实现跨用户 idempotency 隔离（C-23）：DB UNIQUE 约束包含 `user_id`，repository 全部带 `user_id` 过滤；用户 A / B 同 key 各自独立 record；验证: middleware 单元测试 + repository 集成测试
- [x] 2.5 实现并发单执行者（C-24）：DB UNIQUE + row lock 串行化；并发 goroutine 测试断言只一个执行者写 `pending` row，其它请求 replay 或阻塞；验证: 集成测试（goroutine 并发 fixture）
- [x] 2.6 实现同 plan 多 key 并发约束（C-25）：partial UNIQUE INDEX 保证同一 `(user_id, plan_id)` 最多一个 `practice_sessions.status IN ('queued','running')`；第二个返回 `409 PRACTICE_SESSION_CONFLICT`；验证: 集成测试
- [x] 2.7 BDD-Gate: 验证 `E2E.P0.024` 通过（`test/scenarios/e2e/p0-024-practice-session-ai-failure-retry/`）
- [x] 2.8 BDD-Gate: 验证 `E2E.P0.025` 通过（`test/scenarios/e2e/p0-025-practice-idempotency-and-isolation-matrix/`）
- [x] 2.9 Phase 2 commit + work-journal

## Phase 3: 观测 / 隐私 / 收尾

- [x] 3.1 验证 A3 observed AIClient metric label 边界：不私造 `practice_*` metric，不把 `feature_key` / `prompt_version` / `rubric_version` 放入 metric label；practice-specific business metrics 若需要，必须留给 F1 owner 先落可执行 lint/helper；验证: metric label allowlist 单元测试
- [x] 3.2 接线 A3 observed AIClient + `AITaskRunContext`：业务包不得直接写 `ai_task_runs` 或绕过 `backend/internal/ai/aiclient/observability` decorator；每次 AI 调用结束由 fake `AITaskRunWriter` 捕获一行，含 `feature_key` / `model_profile_name` / `model_family` / `fallback_chain` / `validation_status` / `route` / `feature_flag` / `data_source_version` 完整；验证: 集成单元测试断言每列非空 + 与 A3 AICallMeta payload 一致
- [x] 3.3 实现 audit_events 写入：`createPracticePlan` / `startPracticeSession` 触发；metadata 仅 `plan_id` / `session_id` / `goal` / `mode` / `language` / `target_job_id`，不含 question / answer 文本；验证: audit 单元测试
- [x] 3.4 实现并断言隐私红线（D-11）：log / metric label / audit / outbox payload 中 `question_text` / `answer_text` / `hint_text` / AI prompt body / response body / provider secret 全部零出现；用 negative-fixture 单元测试覆盖；验证: redaction 单元测试 + scenario verify.sh 内 grep 断言
- [x] 3.5 Phase 3 legacy-negative grep（收口子集）：`warmup` / `single_drill` / `drill_builder` / `mistake_queue` / `growth_center` / 独立 `voice` route（除 `practice-voice-mvp` 内部占位）/ `practiceModeCard` 在本 plan 输出文档与代码 diff 零出现；验证: CI grep gate
- [x] 3.6 BDD-Gate: 验证 `E2E.P0.026` 通过（`test/scenarios/e2e/p0-026-practice-observability-and-privacy-redlines/`）
- [x] 3.7 文档收口：plan / checklist / test-plan / test-checklist / bdd-plan / bdd-checklist 状态 `active` → `completed`（待 §5 验收 gate 全绿后）；同步 `plans/INDEX.md` 与 `docs/spec/INDEX.md`；运行 `/sync-doc-index --check` 无 drift
- [x] 3.8 Phase 3 commit + work-journal + `/retrospective`（成功交付后沉淀复盘建议）
