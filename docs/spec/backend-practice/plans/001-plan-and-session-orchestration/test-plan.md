# 001 — Plan and Session Orchestration Test Plan

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-05-10

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)

## 1 测试策略概览

本 plan 是 feature-behavior + contract + migration + code-internal 多类型组合，TDD 与 contract gate 必须共同收口；BDD scenario 覆盖用户可见行为切片。本测试计划只覆盖 **单元 / 集成 / contract / drift / lint** 层；BDD 场景见 [bdd-plan](./bdd-plan.md) 与 [bdd-checklist](./bdd-checklist.md)。

测试目标按 phase 与 [plan §3.5 Coverage Matrix](./plan.md#35-coverage-matrix) 行号映射；不引入硬编码代码覆盖率百分比作为 gate；如团队需要观测覆盖率，仅作背景指标。

## 2 Coverage Matrix（测试视角投射）

| 测试源 | Coverage Matrix 行 | Phase | 测试形态 | 文件 / 命令 |
|--------|-------------------|-------|----------|------------|
| shared conventions generated artifacts | R9 / R18 | Phase 0 | 单元 + drift | `make codegen-conventions` + `python3 scripts/lint/conventions_drift.py --repo-root .` + `cd backend && go test ./internal/shared/types ./internal/shared/errors ./internal/shared/idx ./internal/shared/ai` + frontend conventions focused tests |
| openapi diff | R9 / R18 | Phase 0 | drift gate | `make openapi-diff` (with HISTORY_REF base) + `make docs-openapi` |
| events drift | R10 / R18 | Phase 0 | drift gate | `make codegen-events` + `make lint-events` |
| migration check | R5 / R7 / R8 / R18 | Phase 0 | 集成 + drift | `make migrate-up` + `make migrate-down` + `make migrate-status` + 自定义 CHECK 约束断言测试 |
| F3 baseline preflight | (Plan §6 风险) | Phase 0 | preflight script + 单元测试 | `prompt-rubric-registry/001-baseline` 状态读取断言 + `RegistryClient.Resolve` mock test |
| legacy-negative grep | R18 / R19 | Phase 0 + Phase 3 | repo grep gate | CI grep script |
| idempotency middleware | R5 / R6 / R7 / R13 | Phase 1 + Phase 2 | 单元 + 集成 | `cd backend && go test ./internal/middleware/idempotency/...` |
| createPracticePlan handler | R1 / R9 / R20 | Phase 1 | 单元 + contract | `cd backend && go test ./internal/api/practice/...` + `openapi/fixtures/PracticePlans/createPracticePlan.json` parity |
| getPracticePlan handler | R1 / R13 | Phase 1 | 单元 + contract | 同上 + cross-user fixture |
| startPracticeSession 三段式 | R2 / R10 / R11 / R16 | Phase 1 | 集成 + contract | `cd backend && go test ./internal/practice/...` (fake F3 + fake AIClient + DB lock 检测) |
| first-question template render + F3 resolution cleanup | R21 / R24 | Phase 1 | 单元 | `cd backend && go test ./internal/practice -run 'TestStartPracticeSession(Render|FailsReservationWhenPromptResolutionFails)' -count=1` |
| getPracticeSession handler | R2 / R13 | Phase 1 | 单元 + contract | `cd backend && go test ./internal/api/practice/...` + cross-user fixture |
| outbox emit | R10 | Phase 1 | 单元（序列化）+ 事务断言 | `outbox_emitter` unit test |
| AI 失败映射 | R3 / R17 | Phase 2 | 单元（每错误码一例） | `error_mapping_test.go` |
| reservation failed_retryable | R4 | Phase 2 | 集成（fake AIClient 注入失败 → 重试成功） | `reservation_retry_test.go` |
| idempotency body mismatch | R6 | Phase 2 | 单元 | `conflict_test.go` |
| 跨用户隔离 | R13 / R14 | Phase 2 | 单元 + repository 集成 | repository_test.go cross-user fixture |
| 并发单执行者 | R7 | Phase 2 | 集成（goroutine 并发） | concurrency_test.go |
| 同 plan 多 key 并发 | R8 | Phase 2 | 集成 | session_repository_test.go partial UNIQUE INDEX 测试 |
| non-2xx idempotency recovery | R22 | Phase 2 | 单元 | `cd backend && go test ./internal/middleware/idempotency -run TestMiddlewareFinalizesNon2xxAndAllowsCorrectedSameKey -count=1` |
| startPracticeSession TTL expiry | R23 | Phase 2 | SQL mock | `cd backend && go test ./internal/store/practice -run 'TestSQLRepositoryReserveSessionStartResetsExpired' -count=1` |
| A3 observed AIClient + ai_task_runs writer | R11 / R15 | Phase 3 | 集成 | observed_client_test.go / observability wiring test with fake `AITaskRunWriter` |
| audit_events 写入 | R15 | Phase 3 | 单元 | audit_test.go |
| 隐私红线 | R12 / R17 | Phase 3 | 单元 + scenario verify.sh grep | redaction_test.go + verify.sh assertion |
| Legacy-negative grep | R18 / R19 | Phase 3 | repo grep gate | CI grep script |

## 3 Phase 0: 跨 spec 前置修订 + Preflight

| 任务 | 测试文件 / 命令 | 预期 Red/Green 证据 |
|------|----------------|---------------------|
| shared types regenerate after PracticeMode 二值 | `make codegen-conventions` + `python3 scripts/lint/conventions_drift.py --repo-root .` + `cd backend && go test ./internal/shared/types ./internal/shared/errors ./internal/shared/idx ./internal/shared/ai` + frontend conventions focused tests | Red: drift 检测发现 `legacy debrief replay value` 残留；Green: drift 通过 + `PracticeMode` 仅含 `assisted/strict` |
| 错误码注册 | shared types unit + drift | Red: 错误码常量缺失；Green: `PRACTICE_PLAN_NOT_FOUND` / `PRACTICE_SESSION_NOT_FOUND` 在 Go / TS / OpenAPI 一致 |
| OpenAPI enum 删除 + ApiError 错误码新增 | `make openapi-diff` + fixture parity | Red: 检测到 enum 删除（按 D-21 标记 pre-launch baseline rebase）；Green: baseline 同步 rebase + ApiError additive 通过 |
| B3 events 二值化 | `make codegen-events` + `make lint-events` | Red: generated event refs 含旧 PracticeMode；Green: regenerate 后 lint 通过 |
| `idempotency_records` 表 + `practice_plans.mode` CHECK | `make migrate-up` + `make migrate-down` + 单元测试 | Red: CHECK 接受 `legacy debrief replay value`；Green: CHECK 拒绝 + idempotency_records UNIQUE 约束生效 |
| owner spec history append | `/sync-doc-index --check` | Red: Header / INDEX drift；Green: 5 个 spec Header 与 history.md 版本同步 |
| F3 baseline preflight | preflight script + 单元测试 | Red: F3 baseline 状态非 `completed` 时 preflight 抛错；Green: 状态 `completed` + `Resolve` 三个 feature_key 都返回 valid 三元组 |
| Legacy-negative grep | CI grep script | Red: `legacy debrief replay value` 在 PracticeMode 上下文出现；Green: 仅 PracticeGoal 上下文保留 |

## 4 Phase 1: Plan + Session 主流程

| 任务 | 测试文件 / 命令 | 预期 Red/Green 证据 |
|------|----------------|---------------------|
| idempotency middleware | `backend/internal/middleware/idempotency/middleware_test.go` | Red: pending lock / replay / per-user / TTL / domain namespace 任一未实现；Green: 5 个测试用例全部 PASS |
| createPracticePlan handler | `backend/internal/api/practice/create_practice_plan_test.go` + `openapi/fixtures/PracticePlans/createPracticePlan.json` parity contract test | Red: handler 未走 idempotency / 非 baseline goal 未返回 `VALIDATION_FAILED` / 响应字段缺失；Green: 全部断言 PASS + fixture 与 generated server schema 一致 |
| getPracticePlan handler | `backend/internal/api/practice/get_practice_plan_test.go` | Red: 越权未返回 404 / 泄露存在性；Green: cross-user fixture 验证 404 + `PRACTICE_PLAN_NOT_FOUND` |
| startPracticeSession 三段式 | `backend/internal/practice/session_starter_test.go` (fake F3 + fake AIClient + DB lock 检测) | Red: 外部 AI 调用在 DB tx 内 / 三段式合并到单事务；Green: lock 检测显示 reservation tx 与 commit tx 之间 AI 调用无锁；session row + turn row + outbox row 同 commit tx 写入 |
| first-question template render + F3 resolution cleanup | `backend/internal/practice/session_starter_test.go` | Red: `ResolveActive` 失败后不调用 `FailSessionStart` / payload 仍含 raw `{{...}}`；Green: failed reservation 持久化 + prompt body 携带 language / role / skills / rubric / goal |
| getPracticeSession handler | `backend/internal/api/practice/get_practice_session_test.go` | Red: 越权未返回 404；Green: cross-user 404 + `PRACTICE_SESSION_NOT_FOUND` |
| outbox emit | `backend/internal/store/practice/outbox_emitter_test.go` | Red: payload schema 与 B3 不一致 / piiBoundary 未通过；Green: 序列化结果与 `shared/events/practice.session.started.json` 一致 + 不含 raw question text |

## 5 Phase 2: 错误路径 + Idempotency 完备性

| 任务 | 测试文件 / 命令 | 预期 Red/Green 证据 |
|------|----------------|---------------------|
| AI 失败映射 | `backend/internal/practice/error_mapping_test.go` | Red: 错误码映射缺失；Green: timeout / invalid output / secret missing / fallback exhausted 各一例 PASS；envelope 不含明文 |
| reservation failed_retryable | `backend/internal/practice/reservation_retry_test.go` | Red: 同 key 重试无法成功；Green: failed_retryable → 重试 → succeeded 状态机 PASS |
| idempotency body mismatch | `backend/internal/middleware/idempotency/conflict_test.go` | Red: 同 key 不同 fingerprint 仍执行；Green: 返回 `409 PRACTICE_SESSION_CONFLICT`，不执行第二个副作用 |
| 跨用户隔离 | `backend/internal/middleware/idempotency/cross_user_test.go` + repository test | Red: 用户 A response 被用户 B 看到；Green: 各自独立 record |
| 并发单执行者 | `backend/internal/middleware/idempotency/concurrency_test.go` (goroutine fixture) | Red: 多个执行者写 pending row；Green: UNIQUE 约束保证仅一个 |
| 同 plan 多 key 并发 | `backend/internal/store/practice/session_repository_concurrency_test.go` | Red: 同 user + plan 出现两个 active session；Green: partial UNIQUE INDEX 覆盖 `queued/running` 并拒绝第二个，返回 `PRACTICE_SESSION_CONFLICT` |
| non-2xx idempotency recovery | `backend/internal/middleware/idempotency/middleware_test.go` | Red: 422 后同 key corrected body 仍 pending / mismatch；Green: 非 2xx reservation 进入 failed terminal / reset path，corrected body 可重新执行 |
| startPracticeSession TTL expiry | `backend/internal/store/practice/create_plan_test.go` | Red: 过期 pending / succeeded record 仍 conflict / replay；Green: 读取 `expires_at` 并 reset pending 后重新 reserve |

## 6 Phase 3: 观测 / 隐私 / 收尾

| 任务 | 测试文件 / 命令 | 预期 Red/Green 证据 |
|------|----------------|---------------------|
| A3 observed AIClient + ai_task_runs 写入 | `backend/internal/practice/observed_client_test.go` | Red: 业务包直接写 `ai_task_runs` / 绕过 A3 decorator / typed columns 缺失；Green: fake `AITaskRunWriter` 捕获每行 D-15 typed columns，且 metric label 不含 `feature_key` / prompt-rubric version |
| audit_events 写入 | `backend/internal/practice/audit_test.go` | Red: metadata 含 question_text / answer_text；Green: metadata 仅 plan_id / session_id / goal / mode / language / target_job_id |
| 隐私红线 | `backend/internal/practice/redaction_test.go` (negative fixture) | Red: log / metric / audit / outbox payload 出现明文；Green: 全 zero-trace 断言 PASS |
| Legacy-negative grep | CI grep script | Red: retired 术语命中；Green: 全 zero-trace |

## 7 测试命令汇总

```bash
# 全包单元 / 集成测试
cd backend && go test ./internal/api/practice/... \
        ./internal/practice/... \
        ./internal/store/practice/... \
        ./internal/middleware/idempotency/... \
        ./internal/shared/types \
        ./internal/shared/errors \
        ./internal/shared/idx \
        ./internal/shared/ai

# 契约 / drift / lint
make openapi-diff
make codegen-conventions
python3 scripts/lint/conventions_drift.py --repo-root .
cd backend && go test ./internal/shared/types ./internal/shared/errors ./internal/shared/idx ./internal/shared/ai
make codegen-events
make lint-events
make migrate-status

# 文档一致性
/sync-doc-index --check
```

## 8 与 BDD 边界

本测试 plan 不重复 BDD 场景验证；BDD 由 [bdd-plan](./bdd-plan.md) 拥有，包括：

- 用户行为切片（创建 plan、启动 session、看到首题、错误重试）
- 端到端契约（OpenAPI handoff、real DB、real F3 + real / fake A3）
- 隐私红线 / legacy-negative repo 级 grep（与单元测试互补）

单元测试聚焦内部接口、状态机、序列化、事务边界、错误映射；BDD 聚焦"用户得到了什么、下一步能做什么"。
