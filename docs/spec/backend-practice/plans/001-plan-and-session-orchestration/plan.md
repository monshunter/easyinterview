# 001 — Plan and Session Orchestration

> **版本**: 1.4
> **状态**: active
> **更新日期**: 2026-06-13

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

把 [backend-practice spec](../../spec.md) v1.4 §7 锁定的第一个 plan 范围落地：

- 4 个 OpenAPI Practice operation 的 backend handler / service / store：
  - `createPracticePlan`（仅 `goal='baseline'` 路径；`retry_current_round` / `next_round` / `debrief` 留给 future `004-derived-plans-debrief`）
  - `getPracticePlan`
  - `startPracticeSession`（reservation + 事务外 AI + 短事务 commit 三段式；含 first-question 同步生成与 reservation `failed_retryable` 重试，D-13 / D-23）
  - `getPracticeSession`
- shared `idempotency_records` 表与 idempotency middleware（D-30 派生决策；同时为 backend-targetjob / backend-review / 002 复用）
- `practice.session.started` outbox emit
- 跨 spec 前置真理源修订（D-30 Q1=A integrator 模式）：B1 PracticeMode 二值 + 错误码、B2 OpenAPI enum/error、B3 generated event refs、B4 migration（`idempotency_records` 表 + `practice_plans.mode` CHECK 收敛）

不含范围（留给后续 plan）：appendSessionEvent + completePracticeSession（future `002-event-loop-and-completion`）；mode policies + lightweight observe（future `003-mode-policies-and-provenance`）；retry/next_round/debrief goal 派生（future `004-derived-plans-debrief`）；voice operation（future `005-voice-turn-extension`）；DELETE /me CASCADE + timeout sweep（future `006-privacy-cascade-and-cleanup`）。

## 2 背景

`backend-practice` spec v1.4 已锁定 6 个 OpenAPI Practice operation、3 个 outbox 事件、30 条 D-* 决策、27 条 C-* 验收标准、6 个 plan 切分建议。本 plan 是 spec §7 序列的第一个，承接 D-12 + D-13 + D-21 + D-23 + D-26 + D-27 + D-30 主流程，并通过 Phase 0 解锁后续 002~006 plan 共同依赖的契约前置。

frontend-workspace-and-practice 已 ship `E2E.P0.018~021`（workspace render / context / start practice / handoff），当前正式前端已消费 `getPracticePlan` / `createPracticePlan` / `startPracticeSession` 的 fixture-backed generated client；B2 同时已冻结 6 个 Practice operationId 与 fixtures，其中 `appendSessionEvent` / `completePracticeSession` 留给 future `002-event-loop-and-completion`，`getPracticeSession` 在本 plan 落真实刷新 / 恢复读取。下表 operation matrix 明确区分 real backend 状态、mock fixture 状态和后续 owner，避免把 frontend mock handoff 误判为完整后端闭环。

D-30 在 spec v1.4 中显式记录两项关键决策（与本 plan 同步派生）：

1. **Q1=A integrator 模式**：plan 001 Phase 0 直接修订 B1 / B2 / B3 / B4 编码真理源（`shared/conventions.yaml`、`openapi/openapi.yaml`、`shared/events/*`、`migrations/*`），并在四个 owner spec 的 `history.md` 与 `spec.md` Header 同步追加授权记录与版本号；不再为 D-21 / D-26 / D-27 各派 sibling owner spec plan，缩短 critical path；ownership 软化由各 owner spec 在 history 显式登记"协调修订模式 / 关联计划: backend-practice/001 Phase 0"兜底。
2. **Q2=a shared idempotency 表**：D-27 idempotency 存储载体收敛为 shared `idempotency_records`（含 `domain` / `operation` namespace 字段），由本 plan Phase 0 引入并设计为可被未来 backend domain（targetjob / review / 自身 002 等）直接复用；表 schema 在引入第一个 caller 时一并设计，避免后续重构。

## 3 质量门禁分类

- **Plan 类型**: feature-behavior + contract + migration + code-internal（多类型组合）
- **TDD 策略**: Code plan requires TDD — 每个 implementation checklist 项 Red-Green-Refactor 入口在 `backend/internal/api/practice/`、`backend/internal/practice/`、`backend/internal/store/practice/`、`backend/internal/middleware/idempotency/` 下相应包；测试命令从 Go module 根执行，例如 `cd backend && go test ./internal/api/practice/... ./internal/practice/... ./internal/store/practice/... ./internal/middleware/idempotency/...`；详细 phase / file / verification 映射见 [test-plan](./test-plan.md)
- **BDD 策略**: Feature plan requires BDD — 引用 [bdd-plan](./bdd-plan.md) 与 [bdd-checklist](./bdd-checklist.md) 中的 5 个场景 `E2E.P0.022 ~ E2E.P0.026`；主 [checklist](./checklist.md) 在每个 user-visible behavior phase 末尾列 `BDD-Gate:` 项
- **替代验证 gate**: Phase 0 内部契约修订使用 contract test + drift check（OpenAPI / shared types / B3 events / migrations）+ legacy-negative grep（`legacy debrief replay value` 仅在 PracticeMode 上下文）+ F3 baseline preflight assert 作为 gate；这些 gate 与 BDD-Gate 共同构成 Phase 0 收口条件

## 3.1 Operation Matrix

| `operationId` | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|---------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `createPracticePlan` | `openapi/fixtures/PracticePlans/createPracticePlan.json` (`default`, `missing-resume` schema-valid unavailable resume) | `frontend/src/app/screens/workspace/hooks/useStartPractice.ts` creates a plan when no ready `planId` exists | Plan 001 Phase 1: `backend/internal/api/practice.CreatePracticePlan` + `backend/internal/practice` service + `backend/internal/store/practice` repository | `practice_plans`, `idempotency_records`, `audit_events` | none | `E2E.P0.022`, `E2E.P0.025` |
| `getPracticePlan` | `openapi/fixtures/PracticePlans/getPracticePlan.json` (`default`, `archived`, `not-found`) | `useWorkspacePracticePlan.ts` refreshes plan state; `useStartPractice.ts` checks existing plan readiness | Plan 001 Phase 1: `backend/internal/api/practice.GetPracticePlan` + user-scoped store read | `practice_plans` | none | `E2E.P0.022`, `E2E.P0.025` |
| `startPracticeSession` | `openapi/fixtures/PracticeSessions/startPracticeSession.json` (`default`, `ai-timeout-502`) | `useStartPractice.ts` starts the interview and navigates to `practice` with `sessionId` | Plan 001 Phase 1-2: `backend/internal/api/practice.StartPracticeSession` + `session_starter` three-step flow | `idempotency_records`, `practice_sessions`, `practice_turns`, `practice_session_events`, `outbox_events`, A3 `ai_task_runs` writer path | F3 `practice.session.first_question` + A3 `AIClient.Complete` via observed client; `APP_ENV=test` may use deterministic fake, non-test must fail fast without real provider config | `E2E.P0.023`, `E2E.P0.024`, `E2E.P0.025`, `E2E.P0.026` |
| `getPracticeSession` | `openapi/fixtures/PracticeSessions/getPracticeSession.json` (`default`, `missing-session`, `prototype-baseline`) | generated client exists; current workspace start flow does not call it, future `PracticeScreen` refresh / resume may consume it | Plan 001 Phase 1: `backend/internal/api/practice.GetPracticeSession` + user-scoped store read | `practice_sessions`, `practice_turns` | none on read | `E2E.P0.023`, `E2E.P0.025` |
| `appendSessionEvent` | `openapi/fixtures/PracticeSessions/appendSessionEvent.json` (`default`) | future `PracticeScreen` answer / hint / pause events; not consumed by plan 001 frontend handoff | `not-yet-implemented` in plan 001; owned by future `002-event-loop-and-completion`; plan 001 must not register a partial real route or use 501 as a placeholder response (OpenAPI v1.0.0 allows 501 only for privacy export) | `practice_session_events`, `practice_turns`, `outbox_events` in future 002 | F3 `practice.session.follow_up` / `practice.turn.lightweight_observe` in future 002/003 | substitute gate in plan 001: router / mux test proves no partial side effect; BDD deferred to future 002/003 |
| `completePracticeSession` | `openapi/fixtures/PracticeSessions/completePracticeSession.json` (`default`) | future `PracticeScreen` finish action; not consumed by plan 001 frontend handoff | `not-yet-implemented` in plan 001; owned by future `002-event-loop-and-completion`; plan 001 must not register a partial real route or use 501 as a placeholder response (OpenAPI v1.0.0 allows 501 only for privacy export) | `practice_sessions`, `feedback_reports`, `async_jobs`, `outbox_events`, `idempotency_records` in future 002 | report generation job handoff in future 002; no report content in backend-practice | substitute gate in plan 001: router / mux test proves no partial side effect; BDD deferred to future 002 |

## 3.5 Coverage Matrix

| 行 | 类别 | source | plan_phase | verification | negative_scope |
|----|------|--------|-----------|--------------|----------------|
| R1 | Primary | spec C-1（baseline plan 创建） | Phase 1 | createPracticePlan handler / service / repository 单元测试 + openapi fixture parity contract test + E2E.P0.022 | — |
| R2 | Primary | spec C-4（同步启动 session 与首题） | Phase 1 | startPracticeSession 集成单元测试（fake F3 + fake AIClient）+ contract test + E2E.P0.023 | — |
| R3 | Failure/recovery | spec C-5 + D-19（AI 失败映射） | Phase 2 | error_mapping unit test 每错误码一例 + E2E.P0.024 | — |
| R4 | Failure/recovery | spec D-13 + D-23（reservation failed_retryable 状态机） | Phase 2 | reservation_retry 集成单元测试（fake AIClient 注入失败 → 重试成功）+ E2E.P0.024 | — |
| R5 | Boundary | spec C-10 + C-21（success replay） | Phase 1 + Phase 2 | idempotency middleware unit test + E2E.P0.025 | — |
| R6 | Boundary | spec C-22（key body mismatch） | Phase 2 | conflict.go unit test + E2E.P0.025 | — |
| R7 | Boundary | spec C-24（concurrent single-executor） | Phase 2 | DB UNIQUE + row lock 集成测试（goroutine 并发 fixture）+ E2E.P0.025 | — |
| R8 | Boundary | spec C-25（multi-key same-plan concurrent） | Phase 2 | session repository 集成测试 + E2E.P0.025 | — |
| R9 | Cross-layer contract | OpenAPI fixture / generated server / shared types 一致 | Phase 0 + Phase 1 | `make codegen-conventions` + `python3 scripts/lint/conventions_drift.py --repo-root .` + `make codegen-check` + handler 单元测试用 generated request/response types | — |
| R10 | Cross-layer contract | outbox payload 与 B3 schema 一致 | Phase 1 | outbox_emitter 序列化单元测试（断言与 `shared/events/practice.session.started.json` 一致 + piiBoundary 通过）+ scenario verify.sh row 出现一次 | — |
| R11 | Cross-layer contract | F3 Resolve / A3 AIClient.Complete payload metadata 完整 | Phase 1 + Phase 3 | session_starter 集成单元测试（断言 payload 含 feature_key + prompt_version + rubric_version + model_profile_name + language + data_source_version）+ ai_task_runs DB 行断言 | — |
| R12 | Privacy/security | spec D-11 + C-16 子集（log/metric/audit/event 不含明文） | Phase 3 | redaction unit test 用 negative-fixture + scenario verify.sh grep 断言 | `question_text` / `answer_text` / `hint_text` / prompt body / response body / provider secret 在 log/metric/audit/event payload 中零出现 |
| R13 | Privacy/security | spec D-9 + C-13 + C-23（user_id 隔离 + 越权 404） | Phase 2 | repository 单元测试（cross-user fixture）+ E2E.P0.025 含跨用户断言 | — |
| R14 | Auth boundary | spec §4.3（未认证 401 / 越权 404） | Phase 2 | handler middleware 单元测试 | — |
| R15 | Observability | spec §4.4 + A3/F1 现有边界（AI metrics 由 A3 decorator 注册；`feature_key` 只进 `ai_task_runs` typed column，不进 metric label） | Phase 3 | observed AIClient wiring test + `ai_task_runs` 集成测试 + metric label allowlist 单元测试 | metric label 中不含 `question_text` / `answer_text` / `feature_key` / prompt-rubric version / provider raw model id |
| R16 | UX | spec D-13（startPracticeSession 同步首题） | Phase 1 | E2E.P0.023 断言 201 + currentTurn 同步返回 | — |
| R17 | UX | spec D-19（AI 失败 envelope 仅含 code + message） | Phase 2 | error_mapping unit test + E2E.P0.024 envelope 断言 | envelope 不含 prompt / response 明文 |
| R18 | Regression / legacy-negative | spec D-21（删除 legacy debrief replay value PracticeMode 取值） | Phase 0 + Phase 3 | repo-wide grep gate（`legacy debrief replay value` 仅允许出现在 PracticeGoal 上下文，不允许出现在 PracticeMode / mode 字段语义） | `legacy debrief replay value` 在 `PracticeMode` / `mode` / `practice_plans.mode` 上下文零出现 |
| R19 | Regression / legacy-negative | spec D-20（retired 模块术语） | Phase 3 | repo-wide grep gate | `warmup` / `single_drill` / `drill_builder` / `mistake_queue` / `growth_center` / 独立 `voice` route / `practiceModeCard` 在本 plan 输出文档与代码 diff 零出现 |
| R20 | Out-of-scope boundary | 002 / 003 / 004 / 005 / 006 拥有的 endpoint 与 goal 不应被 001 实现 | Phase 1 | router / mux 单元测试断言 `appendSessionEvent` / `completePracticeSession` 在 001 阶段不注册 real route、不产生业务副作用且不返回非合约 501；`createPracticePlan` 对非 baseline goal 返回 422 `VALIDATION_FAILED` + detail 指向 future 004 owner | 002~006 owner endpoint 与 goal 在 001 阶段不会路由到部分实现 |
| R21 | Failure/recovery | spec C-4 / C-5 / D-13 / D-29（F3 prompt resolution fail-closed） | Phase 1 | `session_starter` focused test 断言 `ResolveActive` 失败后调用 `FailSessionStart`，queued session 与 pending idempotency 不残留 | F3 resolution failure 不留下 active queued session 阻塞同 plan 重试 |
| R22 | Boundary | spec C-21 / C-22 / D-27（非 2xx idempotency reservation recovery） | Phase 2 | idempotency middleware unit test 断言非 2xx response 结束后不保留 pending，同 key corrected body 可重新执行 | validation/error response 不把 key 锁死到 TTL |
| R23 | Boundary | spec D-27（custom startSession idempotency TTL） | Phase 2 | SQL repository focused tests 覆盖 expired pending / expired succeeded record reset，不 replay / 不 conflict | `startPracticeSession` idempotency record 过期后不永久 replay / conflict |
| R24 | Cross-layer contract / UX | F3 `practice.session.first_question` prompt template + spec C-4（首题上下文） | Phase 1 | `session_starter` focused test 断言 payload 渲染 `{{language}}` / `{{role_title}}` / `{{top_skills}}` / `{{practice_goal}}` / rubric 占位且无 raw placeholder | zh-CN fallback 到 `multi` prompt 时仍携带语言与目标岗位上下文 |
| R25 | Cross-layer contract | spec D-31（`CreatePracticePlanRequest.resumeAssetId` 必填） | Phase 0 + Phase 1 | OpenAPI required + baseline rebase + Go/TS generated artifacts + fixture validation + frontend request-builder focused test | generated client 不允许省略 resumeAssetId；schema-valid unavailable resume fixture 仍覆盖 422 |
| R26 | Privacy/security | spec D-27（idempotency key hash 统一 pepper） | Phase 2 | `StartPracticeSession` handler focused test 断言 service 收到 peppered `IdempotencyKeyHash` + scenario harness 用同 pepper 查询 start reservation | `startPracticeSession` 不得用空 pepper 写 shared `idempotency_records` |

无 UI 视觉地理 parity 行（本 plan 不涉及 `ui-design/` 复刻；其消费方 frontend-workspace-and-practice 已分别承担 UI parity）。

## 4 实施步骤

### Phase 0: 跨 spec 前置修订 + Preflight

**目标**：把 D-21 / D-26 / D-27 / D-29 落到编码真理源；plan 001 直接修订 B1 / B2 / B3 / B4，**并同步更新各 owner spec history.md / spec.md Header**（按 D-30 Q1=A integrator 模式）；F3 baseline 仅 preflight assert。

#### 0.1 B1 PracticeMode 二值化

修订 `shared/conventions.yaml#enums[PracticeMode]` 删除 `legacy debrief replay value`；运行 `make codegen-conventions` regenerate Go / TS shared types，并用 `python3 scripts/lint/conventions_drift.py --repo-root .` + focused Go / TS shared conventions tests 证明生成物无漂移；同步 B1 spec.md Header 1.14 → 1.15 与 history.md 1.15 行（"授权 backend-practice/001 Phase 0 PracticeMode 二值化与 PRACTICE_*_NOT_FOUND 错误码"）。

#### 0.2 B1 错误码注册

修订 `shared/conventions.yaml` 错误码注册表新增 `PRACTICE_PLAN_NOT_FOUND` / `PRACTICE_SESSION_NOT_FOUND`；regenerate Go / TS / OpenAPI 错误码常量。

#### 0.3 B2 OpenAPI 修订

修订 `openapi/openapi.yaml#components.schemas.PracticeMode` 删除 `legacy debrief replay value`（pre-launch baseline rebase per backend-practice spec D-21，无需 ADR）；新增两个错误码到 `ApiError.code` enum；regenerate Go server / TS client；同步 B2 spec.md Header 1.13 → 1.14 与 history.md 1.14 行（"授权 backend-practice/001 Phase 0 PracticeMode pre-launch baseline rebase + 新增 PRACTICE_*_NOT_FOUND additive"）；如 `openapi/baseline/openapi-v1.0.0.yaml` 涉及 enum 变更，按 D-21 pre-launch wording 原地 rebase 而非新建 v1.1.0 文件。

#### 0.4 B3 generated event refs 二值化

修订 `shared/events/*` 中引用 `PracticeMode` 的 surface（含已 freeze 的 `practice.session.started` payload schema 引用），二值化；regenerate B3 events；同步 B3 spec.md Header 2.1 → 2.2 与 history.md 2.2 行。

#### 0.5 B4 migration：shared idempotency_records + practice_plans CHECK

新建 migration `migrations/NNNNNN_practice_idempotency_baseline.up.sql` / `.down.sql`：

1. 新建 **shared** `idempotency_records` 表，列：`id uuid (uuidv7)`, `user_id uuid`, `domain text`, `operation text`, `idempotency_key_hash text`, `request_fingerprint text`, `status text CHECK IN ('pending','succeeded','failed_retryable','failed_terminal')`, `resource_type text`, `resource_id uuid`, `response_body jsonb`, `error_code text`, `expires_at timestamptz`, `created_at timestamptz`, `updated_at timestamptz`。约束：`PRIMARY KEY (id)`、`UNIQUE (user_id, domain, operation, idempotency_key_hash)`、`INDEX (expires_at)` 用于 TTL 清理。
2. 调整 `practice_plans.mode` CHECK 约束为 `IN ('assisted', 'strict')`，移除 `legacy debrief replay value`。

同步 B4 spec.md Header 1.12 → 1.13 与 history.md 1.13 行（"授权 backend-practice/001 Phase 0 新增 shared idempotency_records 表 + practice_plans.mode CHECK 收敛"）；同步 B4 spec §2.1 表 list 加入 `idempotency_records`。

#### 0.6 F3 baseline preflight

读取 [`prompt-rubric-registry/001-baseline`](../../../prompt-rubric-registry/plans/001-baseline/plan.md) 的 Header 状态：必须为 `completed`；并验证 `RegistryClient.Resolve` 对 `practice.session.first_question` / `practice.session.follow_up` / `practice.turn.lightweight_observe` 三个 feature_key 都返回 valid `(prompt_version, rubric_version, model_profile_name)` 三元组。preflight 失败时 plan 001 暂停在 Phase 0 末并提交"前置阻塞"work-journal，**不进入 Phase 1 AI 实现步骤**。

#### 0.7 Phase 0 legacy-negative grep

仓库范围 grep：`legacy debrief replay value` 字面在 PracticeMode 上下文（`PracticeMode`、`mode`、`practice_plans.mode`、`session.mode`、`PracticeDisplayContext.practiceMode`）零出现；`PracticeGoal` 上下文允许保留。

#### 0.8 Phase 0 commit + work-journal

单 phase commit `phase 0: backend-practice plan-001 contract preflight`；运行 `/work-journal` 一条记录。

#### 0.9 Remediation: createPracticePlan resumeAssetId required contract alignment

修复 `CreatePracticePlanRequest.resumeAssetId` 契约漂移：B2 `openapi/openapi.yaml` 与 `openapi/baseline/openapi-v1.0.0.yaml` 将 `resumeAssetId` 纳入 required；重新生成 Go / TS artifacts；`openapi/fixtures/PracticePlans/createPracticePlan.json` 的 `missing-resume` 场景改为 schema-valid unavailable resume 422；frontend request builder 对 synthetic / missing resume id fail fast。同步 backend-practice D-31 与 B2 history。

### Phase 1: Plan + Session 主流程 (success path) + idempotency replay 基础

**目标**：用户可创建 baseline plan、读取、启动 session、同步看到首题、刷新 session、Idempotency-Key success replay。

#### 1.1 Shared idempotency middleware

在 `backend/internal/middleware/idempotency/` 实现 shared idempotency middleware：消费 B4 `idempotency_records` 表，提供 `(domain, operation)` 注入入口（domain 可选 `"practice"`、`"targetjob"` 等），支持 pending lock + success replay + per-user 隔离 + TTL expire；与 backend-auth 现有 active-request dedupe 不冲突（active-request 仍由 backend-auth own，此处是 generic infrastructure）。

#### 1.2 createPracticePlan handler (baseline goal)

在 `backend/internal/api/practice/` 实现 `createPracticePlan` handler；走 idempotency middleware (`domain="practice"`, `operation="createPracticePlan"`)；仅接受 `goal="baseline"`，其它 goal 返回 `422 VALIDATION_FAILED`，error detail 携带 `goal` 与 `owner='004-derived-plans-debrief'`，避免私造未注册错误码；写 `practice_plans(status='ready')` 短事务 + audit_events；返回 `201 + PracticePlan`。

#### 1.3 getPracticePlan handler

在 `backend/internal/api/practice/` 实现 `getPracticePlan` handler；user_id 过滤；越权或不存在均返回 `404 + PRACTICE_PLAN_NOT_FOUND`，不泄露资源存在性。

#### 1.4 startPracticeSession 三段式 (baseline goal)

在 `backend/internal/practice/session_starter.go` 实现 D-13 / D-23 三段式：

1. **短事务 reserve**：写 `idempotency_records(status='pending')` + 创建 `practice_sessions(status='queued', user_id, plan_id)` 行，提交事务；`queued` 是 B1/B2/B4 已冻结的 SessionStatus，plan 001 不新增 `reserved` wire / DB 枚举。
2. **事务外 AI 调用**：F3 `RegistryClient.Resolve("practice.session.first_question", language)` → A3 `AIClient.Complete(profile_name, payload)`；payload metadata 包含 `feature_key + prompt_version + rubric_version + model_profile_name + language + data_source_version`。
3. **短事务 commit**：成功路径写 `practice_turns(turn_index=1, status='asked', question_text, question_intent)` + `practice_session_events(seq_no=1, event_type='session_started')` + outbox `practice.session.started` 行 + 更新 `practice_sessions.status='running'` + 更新 `idempotency_records(status='succeeded', response_body, resource_type='practice_session', resource_id=session_id)`，全部在同一事务内。

返回 `201 + PracticeSession{status:'running', currentTurn:{turnIndex:1, status:'asked', questionText, questionIntent, askedAt}}`。

外部 AI 调用**禁止**包在 DB transaction 内；集成测试用 fake AIClient + lock 检测兜底。

#### 1.5 getPracticeSession handler

在 `backend/internal/api/practice/` 实现 `getPracticeSession` handler；user_id 过滤；越权或不存在均返回 `404 + PRACTICE_SESSION_NOT_FOUND`。

#### 1.6 outbox emit 与 B3 payload 对齐

`outbox_emitter.go` 序列化 `practice.session.started` payload 时严格按 B3 `shared/events/practice.session.started.*` schema 输出，且字段值通过 piiBoundary（不含 raw question text 等）；与业务表写入同一 DB 事务。

#### 1.7 BDD-Gate

- BDD-Gate: 验证 `E2E.P0.022` 通过（scenario path: `test/scenarios/e2e/p0-022-practice-plan-baseline-create-and-read/`）
- BDD-Gate: 验证 `E2E.P0.023` 通过（scenario path: `test/scenarios/e2e/p0-023-practice-session-start-and-first-question/`）

#### 1.8 Phase 1 commit + work-journal

#### 1.9 Remediation: F3 resolution failure cleanup + first-question template render

修复 `startPracticeSession` reservation 后的 F3 `ResolveActive` 失败路径，必须调用 `FailSessionStart` 把 queued session 与 idempotency reservation 推进到 failed terminal / retryable 状态，避免同 plan 后续重试被 active queued session 或 pending record 卡住。

同时在首题 AI payload 中渲染 F3 `practice.session.first_question` prompt template 的上下文字段：`language`、`role_title`、`seniority`、`top_skills`、`rubric_dimensions`、`practice_goal`；上下文来自 `practice_plans` 关联的 `target_jobs` / `target_job_requirements`，缺失字段用显式 fallback 值，不发送 raw `{{...}}` 占位。

### Phase 2: 错误路径 + Idempotency 完备性

**目标**：AI 失败重试、越权 404、idempotency body mismatch、跨用户隔离、并发单执行者、同 plan 多 key 并发。

#### 2.1 AI 失败映射 (D-19)

在 `backend/internal/practice/error_mapping.go` 实现 AI 错误 → B1 错误码映射：timeout → `502 AI_PROVIDER_TIMEOUT`、invalid output → `502 AI_OUTPUT_INVALID`、secret missing → `502 AI_PROVIDER_SECRET_MISSING`、fallback exhausted → `503 AI_FALLBACK_EXHAUSTED`；envelope 不含 prompt / response 明文。

#### 2.2 reservation failed_retryable 状态机 (D-23)

在 `backend/internal/practice/reservation_retry.go` 实现首题 AI 失败时把 `idempotency_records.status='failed_retryable'` + `practice_sessions.status='failed'` + `failure_code` 记录；同 `Idempotency-Key` 重试时按 `failed_retryable` 允许重新生成首题，成功后固化为 `succeeded` 并升级为 `running` session；若同 key 重试 N 次仍失败可由 retry policy 进入 `failed_terminal`（具体阈值留作 plan 实施时配置）。

#### 2.3 idempotency body mismatch 拒绝 (C-22)

在 `backend/internal/middleware/idempotency/conflict.go` 实现：同 `(user_id, domain, operation, idempotency_key_hash)` + 不同 `request_fingerprint` 返回 `409 PRACTICE_SESSION_CONFLICT`（或 B1 idempotency mismatch 错误码），不执行第二个副作用；envelope 不泄露首次资源内容。

#### 2.4 跨用户 idempotency 隔离 (C-23)

DB UNIQUE 约束包含 `user_id`，repository 层全部带 `user_id` 过滤；用户 A / 用户 B 同 key 各自独立 record。单元测试 cross-user fixture + repository 集成测试兜底。

#### 2.5 并发单执行者 (C-24)

DB `UNIQUE (user_id, domain, operation, idempotency_key_hash)` + row lock 串行化；并发 goroutine 测试断言只有一个执行者写 `pending` row，其它请求要么 replay succeeded，要么阻塞等待。

#### 2.6 同 plan 多 key 并发 (C-25)

新增 partial UNIQUE INDEX 或等价 CHECK 约束保证同一 `(user_id, plan_id)` 最多存在一个 `practice_sessions.status IN ('queued','running')`；第二个 startPracticeSession 返回 `409 PRACTICE_SESSION_CONFLICT`，不绕过 idempotency 生成双 active session。

#### 2.7 BDD-Gate

- BDD-Gate: 验证 `E2E.P0.024` 通过（scenario path: `test/scenarios/e2e/p0-024-practice-session-ai-failure-retry/`）
- BDD-Gate: 验证 `E2E.P0.025` 通过（scenario path: `test/scenarios/e2e/p0-025-practice-idempotency-and-isolation-matrix/`）

#### 2.8 Phase 2 commit + work-journal

#### 2.9 Remediation: generic idempotency non-2xx recovery

修复 shared idempotency middleware 对非 2xx handler response 的收口行为：side-effect handler 返回 validation/error envelope 后，不得留下 `pending` record；同 key + corrected body 必须可以重新进入 reservation，而不是直到 TTL 前都被 pending / fingerprint mismatch 阻塞。

#### 2.10 Remediation: startPracticeSession TTL expiry

修复 `startPracticeSession` 自定义 reservation path：读取并判断 `idempotency_records.expires_at`，过期的 pending / succeeded / failed record 必须 reset 为新 pending reservation 后重新执行，不得永久 replay / conflict。

#### 2.11 Remediation: startPracticeSession idempotency pepper alignment

修复 `startPracticeSession` 自定义 reservation path 的 key hash 口径：handler 必须接收与 shared idempotency middleware 相同的 `KeyPepper` 配置，并用该 pepper 计算 `IdempotencyKeyHash`；main wiring、scenario harness 与 focused handler test 必须证明 create/start 两条 practice side-effect path 写入 shared `idempotency_records` 时使用同一哈希策略。

### Phase 3: 观测 / 隐私 / 收尾

**目标**：A3 observed AIClient / `ai_task_runs` / audit_events / 隐私红线 / legacy-negative / 文档收口。

#### 3.1 观测边界与 metric label gate

A3 `observability` decorator 自动覆盖 7 个 `ai_task_*` metric family；本 plan 不私造新的 `practice_*` metric，不把 `feature_key` / `prompt_version` / `rubric_version` 放入 metric label。practice-specific business metrics（如 session started / completed）若需要实施，必须先由 F1 owner 修订 `observability-stack` 并落可执行 lint/helper 后再进入对应 backend-practice 后续 plan。Plan 001 的 gate 只验证 A3 metric label set 使用 F1 allowed labels，且 `feature_key` 进入 `ai_task_runs` typed column / audit 摘要而不是 metric label。

#### 3.2 AI 调用 A3 observed-client wiring + `ai_task_runs` 断言

在 `backend/internal/practice/` 的 AI 调用入口向 A3 observed `AIClient` 传入 `AITaskRunContext` 与 F3 provenance metadata；不得在业务包直接写 `ai_task_runs` 或绕过 `backend/internal/ai/aiclient/observability` decorator。集成测试用 fake `AITaskRunWriter` 断言每次 AI 调用产生一行，且 D-15 typed columns (`feature_key` / `model_profile_name` / `model_family` / `fallback_chain` / `validation_status` / `route` / `feature_flag` / `data_source_version`) 完整。

#### 3.3 audit_events 写入

`createPracticePlan` / `startPracticeSession` 触发 audit；metadata 仅 `plan_id` / `session_id` / `goal` / `mode` / `language` / `target_job_id`，不含 question / answer 文本。`appendSessionEvent` 不写 audit（高频，由 B3 outbox 已覆盖）；`completePracticeSession` audit 由 002 实现。

#### 3.4 隐私红线断言 (D-11)

在 `backend/internal/practice/redaction.go` 实现并断言 log / metric label / audit / outbox payload 中 `question_text` / `answer_text` / `hint_text` / AI prompt body / response body / provider secret 全部零出现；用 negative-fixture 单元测试覆盖；scenario verify.sh 内 grep 断言兜底。

#### 3.5 Legacy-negative grep 收口

repo-wide grep gate：`warmup` / `single_drill` / `drill_builder` / `mistake_queue` / `growth_center` / 独立 `voice` route（除 `practice-voice-mvp` 内部 voice operation 占位）/ `practiceModeCard` 在本 plan 输出文档与代码 diff 零出现。

#### 3.6 BDD-Gate

- BDD-Gate: 验证 `E2E.P0.026` 通过（scenario path: `test/scenarios/e2e/p0-026-practice-observability-and-privacy-redlines/`）

#### 3.7 文档收口

plan / checklist / test-plan / test-checklist / bdd-plan / bdd-checklist 状态从 `active` 推进到 `completed`（待 §5 验收 gate 全绿后）；同步 [plans/INDEX.md](../INDEX.md)、[docs/spec/INDEX.md](../../../INDEX.md)；运行 `/sync-doc-index --check` 无 drift。

#### 3.8 Phase 3 commit + work-journal + retrospective

成功交付后调用 `/retrospective` 沉淀复盘建议。

### Phase 4: D-20 简历扁平化 resumeId 绑定适配

> product-scope D-20 / backend-practice D-39。`createPracticePlan` 简历绑定 `resumeAssetId`→`resumeId`。依赖 [B2 004 Phase 7](../../../openapi-v1-contract/plans/004-resume-additive-coverage/plan.md)（`CreatePracticePlanRequest.resumeAssetId`→`resumeId`）+ [B4 002 Phase 6](../../../db-migrations-baseline/plans/002-resume-versions-additive/plan.md)（`practice_plans.resume_asset_id`→`resume_id`）。

#### 4.1 handler / service / store resumeId 迁移

`createPracticePlan` handler + service + `practice_plans` store：generated `CreatePracticePlanRequest.resumeAssetId`→`resumeId`、列 `resume_asset_id`→`resume_id`（指向 `resumes`）；baseline 校验「resume 属于当前用户且未删除」改查 `resumes`。

（验证：handler + store unit/integration test PASS；缺 resume 的 `422 VALIDATION_FAILED` 路径保留）

#### 4.2 首题 prompt 引用扁平 resume

baseline 首题 prompt 上下文从「resume version / 主版本」改为扁平 `resumes.structured_profile`。

（验证：first-question AI 调用 unit test stub 验证 prompt 上下文）

#### 4.3 收口

`cd backend && go test ./internal/practice/... ./cmd/api`；零 `resumeAssetId` / `resume_asset_id`（practice 包内）残留 grep（generated 由 B2 重生除外）。

（验证：全 gate PASS + 负向 grep 0 命中）

## 5 验收标准

- 本计划 §3.5 Coverage Matrix 全部 20 行的 verification 通过
- **D-20（Phase 4）**：`createPracticePlan` 简历绑定 `resumeAssetId`→`resumeId`、`practice_plans` 列 `resume_asset_id`→`resume_id`、首题 prompt 引用扁平 `structured_profile`；`go test ./internal/practice/... ./cmd/api` PASS；零 practice 包内 `resumeAssetId` 残留
- spec C-1 / C-4 / C-5 / C-13 / C-21 / C-22 / C-23 / C-24 / C-25 全部 PASS（涉及 createPracticePlan / startPracticeSession 的部分；C-10 / C-11 / C-15 / C-16 / C-17 / C-19 / C-20 等不在 001 范围或与 002+ 共担）
- 5 个 BDD 场景 `E2E.P0.022` ~ `E2E.P0.026` 全部通过
- 跨 spec 修订 history append 在 B1 / B2 / B3 / B4 已生效，且 spec.md Header 与 history.md 版本同步
- legacy-negative grep：`legacy debrief replay value`（仅在 PracticeMode 上下文）、retired 模块术语在本 plan 输出文档与代码 diff 零出现
- TDD checklist（[test-checklist](./test-checklist.md)）全部 phase 单元 / 集成 / contract test 通过
- BDD checklist（[bdd-checklist](./bdd-checklist.md)）每个 scenario 的 setup / trigger / verify / cleanup / 证据捕获项全部完成

## 6 风险与应对

| 风险 | 应对 |
|------|------|
| Q1=A integrator 模式被严格 ownership 审查质疑 | plan.md / history.md 显式登记"协调修订模式"；同步在 B1 / B2 / B3 / B4 history append 授权记录；用户已批准；plan-review 复核 |
| OpenAPI enum 删除属于 breaking change | 按 D-21 "pre-launch 直接改 baseline，不引入 deprecated alias"——pre-launch baseline rebase 不需要 ADR；history 行注明 "pre-launch baseline rebase"；如 openapi-v1-contract owner 在 plan-review 阶段坚持 ADR，则改走 ADR + 新 baseline `openapi-v1.0.1.yaml` 路径 |
| shared `idempotency_records` 抽象设计走样 | Phase 0 单独 PR review + drift check；表结构 `(domain, operation, idempotency_key_hash)` 命名空间在引入第一个 caller 时一并设计；未来增加 caller 时只追加 namespace + caller-side test，不改表 schema；如未来发现需要 schema 演进，必须新建 owner spec D-31 决策 |
| F3 baseline 未完成阻塞 Phase 1 AI 实现 | Phase 0 step 0.6 preflight 强制 fail-fast；若 F3 baseline 仍 active，plan 001 暂停在 Phase 0 末并提交"前置阻塞" work-journal，告知 F3 owner |
| AI provider 临时不可用导致 BDD 不稳定 | E2E.P0.024 用 fake AIClient + scripts/setup.sh 注入可控失败；不依赖真实 provider 故障重现 |
| outbox 双写一致性回归 | Phase 1 step 1.6 单元断言 outbox row 与业务表必须同事务；scenario verify.sh 检查 `practice.session.started` row 出现一次 + payload schema |
| reservation+三段式实现走样为长事务 | Phase 1 step 1.4 集成测试断言外部 AI 调用不在 DB tx 内（用 fake AIClient 加锁检测）；plan-review 显式 gate |
| 同 plan 多 key 并发约束实现复杂度高 | Phase 2 step 2.6 用 DB partial UNIQUE INDEX 表达"同一 user + plan 最多一个 active session"；不在应用层做软锁；集成测试 goroutine 并发 fixture 兜底 |

## 7 关联文档

- [Spec backend-practice v1.4](../../spec.md)
- [History](../../history.md)
- [Plans INDEX](../INDEX.md)
- [Test Plan](./test-plan.md) / [Test Checklist](./test-checklist.md)
- [BDD Plan](./bdd-plan.md) / [BDD Checklist](./bdd-checklist.md)
- 上游契约：[B1 shared-conventions-codified](../../../shared-conventions-codified/spec.md) · [B2 openapi-v1-contract](../../../openapi-v1-contract/spec.md) · [B3 event-and-outbox-contract](../../../event-and-outbox-contract/spec.md) · [B4 db-migrations-baseline](../../../db-migrations-baseline/spec.md)
- AI 与配置：[F3 prompt-rubric-registry](../../../prompt-rubric-registry/spec.md) · [A3 ai-provider-and-model-routing](../../../ai-provider-and-model-routing/spec.md) · [F1 observability-stack](../../../observability-stack/spec.md) · [A4 secrets-and-config](../../../secrets-and-config/spec.md)
- 协议参考：[backend-auth](../../../backend-auth/spec.md)（in-process drainer + active-request dedupe 先例）
- 消费端：[frontend-workspace-and-practice](../../../frontend-workspace-and-practice/spec.md)
- BDD 框架：[test/scenarios/README.md](../../../../../test/scenarios/README.md) · [test/scenarios/e2e/README.md](../../../../../test/scenarios/e2e/README.md) · [test/scenarios/e2e/INDEX.md](../../../../../test/scenarios/e2e/INDEX.md)

## 8 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-05-10 | 1.1 | L2 code review remediation：修复 F3 first-question JSON schema 解析、`startPracticeSession` success replay response snapshot、`000003` down migration baseline ownership gate。 |
