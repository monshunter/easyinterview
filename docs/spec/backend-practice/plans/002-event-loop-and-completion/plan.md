# 002 — Event Loop and Completion

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-05-14

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

把 [backend-practice spec](../../spec.md) v1.5 §7 锁定的第二个 plan 范围落地，承接 001 的 idempotency / store / outbox / observed AI 基础设施，闭合 P0 用户路径中"答题循环 + 异步完成"段：

- `POST /practice/sessions/{sessionId}/events` `appendSessionEvent`：handler + service + repository；5 种 event kind 的状态机分支（`answer_submitted` / `hint_requested` / `turn_skipped` / `session_paused` / `session_resumed`）；per-session `clientEventId` 双轨 idempotency；`SELECT FOR UPDATE` + `seq_no = COALESCE(MAX, 0) + 1` 序列化；AssistantAction 决策与 wire provenance；`practice.turn.completed` outbox emit；DB internal turn status 5 值 → OpenAPI wire turn status 5 值的契约对齐（D-33 plan-level 决策：扩 wire enum，而非 handler 压缩映射）。
- `POST /practice/sessions/{sessionId}/complete` `completePracticeSession`：handler + service + repository；shared idempotency middleware 接入；单事务推进 `practice_sessions` → `completing` + 创建 `feedback_reports(status='queued')` placeholder + 创建 `async_jobs(job_type='report_generate', status='queued', dedupe_key=sessionId)` + `outbox_events practice.session.completed` + idempotency `response_body` snapshot；HTTP 202 + `ReportWithJob`；C-27 双 key 行为锁定为 D-35 plan-level 决策：反向查 `feedback_reports` 复用既有 report/job replay。
- Phase 0 跨 spec 契约前置（D-30 integrator 模式延续）：
  - `shared/jobs.yaml#report_generate` 显式 ownership 语义（D-32 plan-level 决策：新增 `triggerEventSemantic: source_event_only` 字段 + dispatcher lint）；同步 B3 spec.md / history.md。
  - `openapi/openapi.yaml#components.schemas.PracticeTurn.status` 扩 wire enum 至 5 值并 regenerate Go server / TS client、刷新 PracticeSessions fixtures、更新 backend-practice 引用、提醒 frontend consumer（D-33 plan-level 决策）；同步 B2 spec.md / history.md。
  - F3 baseline `prompt-rubric-registry/001-baseline` preflight assert（已 completed，不动）；A3 observed AIClient 调用 `practice.session.follow_up` / `practice.turn.lightweight_observe` feature_key 复用 001 已经 wired 的 observability decorator。

不含范围（留给后续 plan）：mode 真实策略与 hint AI 输出（future `003-mode-policies-and-provenance`，本 plan 在 `hint_requested` 默认走 strict 语义 409，作为 D-34 plan-level 决策）；retry/next_round/debrief goal 派生（future `004-derived-plans-debrief`）；voice operation（future `005-voice-turn-extension`）；DELETE /me CASCADE + 24h timeout sweep（future `006-privacy-cascade-and-cleanup`）。报告内容生成、证据评分、ReadinessTier 计算归 `backend-review` future owner，本 plan 仅创建 queued report/job 占位与 source event。

## 2 背景

`backend-practice` spec v1.5 §7 把"event-loop + completion"作为 001 之后的下一个独立可部署可验证 plan。001 已交付（completed, 2026-05-10）：4 个 OpenAPI operation handler、shared `idempotency_records` 表与 idempotency middleware（含 `domain`/`operation` 命名空间与 `failed_retryable` 状态、cross-user 隔离、pepper 一致性）、`practice_sessions` / `practice_plans` / `practice_session_events` / `practice_turns` 写入与 user-scoped 读、`practice.session.started` outbox emit、Phase 0 跨 spec 契约前置（PracticeMode 二值化、`PRACTICE_*_NOT_FOUND` 错误码、shared `idempotency_records` 表与 `practice_plans.mode` CHECK 收敛）、5 个 BDD scenario（`E2E.P0.022 ~ E2E.P0.026`）。

002 不引入新的 spec D-* 主决策，但需要在 plan 层固化 4 项实施级子决策，源于本 plan 派生会话的 user 确认：

- **D-32（落实 spec D-28）**：`shared/jobs.yaml#report_generate` 新增 `triggerEventSemantic: source_event_only` 字段表达"`practice.session.completed` 是 source event / analytics fact，不是触发 dispatcher 二次创建 job 的指令"。真实 gate 走 root `make lint-events`（`scripts/lint/lint_events.py`）与 `make codegen-events-check`；必要时扩展 `scripts/lint/events_inventory.py shared/events.yaml shared/jobs.yaml shared/conventions.yaml` 的底层校验，但不得把无参 `events_inventory.py` 当作可执行 gate。`backend/cmd/codegen/events` 必须生成 `JobTriggerEventSemanticSourceEventOnly` / `JobTriggerEventSemanticTriggerCreatesJob` Go 常量 + `IsSourceEventOnly(jobType) bool` 谓词到 `backend/internal/shared/jobs/`（以及对应 TS 生成物如适用），不得手改 `DO NOT EDIT` 生成文件。002 阶段无 runtime outbox→asynq dispatcher（现仓 `backend/internal/runtime/` 不存在；现有 outbox/async_jobs 全部由 handler/store 直接 INSERT），dispatch-time 跳过 binding 留给未来 `backend-async-runner` plan 接管，届时必须读取常量在 dispatch-time 跳过 `source_event_only` 任务。002 阶段 `report_generate` 不被二次创建的兜底来自：(a) handler 端复用 `async_jobs UNIQUE INDEX (job_type, dedupe_key) WHERE status IN ('queued','running')`，(b) D-35 反向查 `feedback_reports` 复用既有 `ReportWithJob`，(c) repo grep 断言 `report_generate` INSERT 仅出现在 `complete_session.go`。其它 8 个 job 默认 `triggerEventSemantic: trigger_creates_job`（不破坏既有行为），lint 把缺省值视为 default。
- **D-33（落实 spec D-25）**：`openapi/openapi.yaml#components.schemas.PracticeTurn.status` 扩 wire enum 为 `[asked, answered, follow_up_requested, assessed, skipped]`，与 DB internal CHECK 完全对齐，去除 D-25 描述中的"handler 压缩映射"备选。`openapi/baseline/openapi-v1.0.0.yaml` 按 D-21 pre-launch baseline rebase 同步原地修订。Regenerate Go server / TS client；PracticeSessions fixtures `appendSessionEvent` / `completePracticeSession` 补齐新值；frontend consumer（`frontend/src/app/screens/workspace/`、`frontend/src/app/screens/practice/` 与生成的 client schema）需要在 plan 同步窗口内消化 enum 扩展（plan 通过 lint 与 type drift 反查）。
- **D-34（落实 spec D-5 在 002 阶段的占位）**：`appendSessionEvent` router 必须接受全部 5 种 `kind`（拒绝任何 kind 返回 422 `VALIDATION_FAILED` 即视为缺陷）。`hint_requested` 在 002 阶段默认走 strict 语义：返回 `409 PRACTICE_SESSION_CONFLICT` + `detail.policy='hint_disabled_in_mode'` + `detail.mode=session.mode`，与 spec C-8 同构；003 接手后按 mode 切分 assisted（`show_hint` + F3 `practice.turn.lightweight_observe`）与 strict（继续 409）。002 BDD 场景 E2E.P0.039 直接断言 hint 默认 409 行为，并预留 003 的 mode 切分扩展点。
- **D-35（落实 spec C-27）**：`completePracticeSession` handler 进入后先 user-scoped 查 `feedback_reports WHERE session_id = $1 AND user_id = $2`。若 session 已完成（存在 `feedback_reports` 且 `practice_sessions.status` ∈ `{completing, completed, failed}`），则不论 `Idempotency-Key` 是否一致，都返回该 session 的既有 `ReportWithJob`（原始 HTTP status 与 response_body）；新 `Idempotency-Key` 写 idempotency_records 时关联到既有 `resource_id=reportId`，状态写 `succeeded` 并 snapshot 同一 response。等同于"session-level 完成是 once-and-only-once，idempotency key 仅控制 inflight 单执行者"。

frontend-workspace-and-practice 当前 `frontend/src/app/screens/workspace/` 已经消费 `getPracticePlan` / `createPracticePlan` / `startPracticeSession` 的 fixture-backed generated client；`PracticeScreen` 与 finish action 是未来 plan 范围，但 generated client 在本 plan 完成 Phase 0 OpenAPI rebase 后必须无 schema drift。

## 3 质量门禁分类

- **Plan 类型**: feature-behavior + contract + code-internal（多类型组合）
- **TDD 策略**: Code plan requires TDD — 每个 implementation checklist 项 Red-Green-Refactor 入口在 `backend/internal/api/practice/`、`backend/internal/practice/`、`backend/internal/store/practice/`、`backend/internal/middleware/idempotency/` 下相应包；B3 lint/codegen 入口在 root `make lint-events` / `make codegen-events-check`，底层脚本为 `scripts/lint/lint_events.py` 与 `backend/cmd/codegen/events`。trigger-semantic 常量与谓词必须由事件 codegen 生成到 `backend/internal/shared/jobs/`（以及对应 TS 生成物如适用），不得手改生成文件。002 阶段无 runtime outbox→asynq dispatcher 包，dispatch-time 跳过逻辑由未来 `backend-async-runner` plan 接管。测试命令从 Go module 根执行（例如 `cd backend && go test ./internal/api/practice/... ./internal/practice/... ./internal/store/practice/... ./internal/middleware/idempotency/... ./internal/shared/jobs/...`）；详细 phase / file / verification 映射见 [test-plan](./test-plan.md)
- **BDD 策略**: Feature plan requires BDD — 引用 [bdd-plan](./bdd-plan.md) 与 [bdd-checklist](./bdd-checklist.md) 中的 6 个场景 `E2E.P0.038` / `E2E.P0.039` / `E2E.P0.040` / `E2E.P0.041` / `E2E.P0.042` / `E2E.P0.043`；主 [checklist](./checklist.md) 在每个 user-visible behavior phase 末尾列 `BDD-Gate:` 项
- **替代验证 gate**: Phase 0 内部契约修订使用 contract test + drift check（OpenAPI / shared types / B3 jobs.yaml lint / `triggerEventSemantic` 枚举校验）+ legacy-negative grep（旧 dispatcher 二次创建 `report_generate` 模式；wire turn enum 历史压缩值）+ F3 baseline preflight assert；Phase 4 隐私 / 观测使用 metric label allowlist + repo grep + redaction assertion 作为 gate

## 3.1 Operation Matrix

| `operationId` | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|---------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `appendSessionEvent` | `openapi/fixtures/PracticeSessions/appendSessionEvent.json`：Phase 0 补齐 `default`（answer_submitted → ask_question）、`follow-up`（answer_submitted → ask_follow_up）、`hint-strict-conflict`（hint_requested → 409）、`turn-skipped`（turn_skipped → ask_question）、`pause-resume`（session_paused / session_resumed → session_wait）、`replay`（重复 clientEventId 返回首次结果）、`mismatch`（同 clientEventId 不同 payload → 409）、`completed`（达到 question_budget → session_completed） | Plan 002 Phase 2 完成后，`frontend/src/app/screens/practice/` 与 generated TS client 通过 fixture-backed transport 接入；本 plan 不要求正式前端消费，但生成物必须无 schema drift | Plan 002 Phase 1-2：`backend/internal/api/practice.AppendSessionEvent` + `backend/internal/practice.SessionEventService`（state machine + AssistantAction decision）+ `backend/internal/store/practice.AppendSessionEvent` repository | `practice_session_events`（含 UNIQUE(session_id, client_event_id) 复用 + seq_no 单调）、`practice_turns`（DB 5 值 status 推进；`follow_up_count` 为 server-owned state，不接受客户端 payload 驱动）、`outbox_events`（仅在 turn assessed / skipped 等完成态产生 `practice.turn.completed` 行）；append 路径不写 `audit_events` | F3 `practice.session.follow_up`（answer_submitted → ask_follow_up 分支）；hint_requested 在 002 默认 strict 不调 AI；`practice.turn.lightweight_observe` 不在 002 调用（归 003） | `E2E.P0.038`, `E2E.P0.039`, `E2E.P0.040`, `E2E.P0.043` |
| `completePracticeSession` | `openapi/fixtures/PracticeSessions/completePracticeSession.json`：Phase 0 补齐 `default`（202 + ReportWithJob queued）、`replay`（同 key 同 fingerprint 返回首次 response）、`mismatch`（同 key 不同 fingerprint → 409）、`session-already-completed`（同 session 不同 key 走 D-35 replay）、`cross-user-not-found`（404 PRACTICE_SESSION_NOT_FOUND） | Plan 002 Phase 3 完成后，`frontend/src/app/screens/practice/` finish action 通过 fixture-backed transport 接入；本 plan 不要求正式前端消费 | Plan 002 Phase 3：`backend/internal/api/practice.CompletePracticeSession`（idempotency middleware decorated）+ `backend/internal/practice.SessionCompletionService` + `backend/internal/store/practice.CompleteSession` repository | `practice_sessions(status='completing')`、`practice_session_events(event_type='session_completed', seq_no=MAX+1)`、`feedback_reports(status='queued', user_id, session_id)` placeholder、`async_jobs(job_type='report_generate', status='queued', dedupe_key=sessionId)`、`outbox_events practice.session.completed`、`idempotency_records(status='succeeded', resource_type='feedback_report', resource_id=reportId, response_body snapshot)` | no AI in completion path（D-8 only applies to AI-driven AssistantAction generation；completion 同事务无 AI 调用） | `E2E.P0.041`, `E2E.P0.042`, `E2E.P0.043` |

## 3.5 Coverage Matrix

| 行 | 类别 | source | plan_phase | verification | negative_scope |
|----|------|--------|-----------|--------------|----------------|
| R1 | Primary | spec C-6（答题循环与下一题） | Phase 2 | `SessionEventService.handleAnswerSubmitted` 单元测试覆盖 `ask_question` / `ask_follow_up` / `session_completed` 三分支 + repository 集成测试 + `E2E.P0.038` | — |
| R2 | Primary | spec C-9（重复 clientEventId 幂等） | Phase 2 | `clientEventId` reservation 单元测试（同 payload replay 返回首次结果，不重复 AI call）+ repository UNIQUE(session_id, client_event_id) 集成测试 + `E2E.P0.039` | — |
| R3 | Alternate | spec §2.1 5 种 kind exhaustive router | Phase 1 + Phase 2 | `routeEvent(kind)` 单元测试 5 个分支 + golden test 拒绝未知 kind → 422 + `E2E.P0.039` | unknown kind 不应映射为 500；422 + `VALIDATION_FAILED` |
| R4 | Failure/recovery | spec C-26 mismatch（同 clientEventId 不同 payload fingerprint） | Phase 2 | `clientEventId` fingerprint 单元测试 → 409 + envelope 不泄露首次 payload + `E2E.P0.039` | 不重复触发 AI；不重复写 event |
| R5 | Boundary | spec C-14（并发 appendSessionEvent 序列化） | Phase 2 | repository 集成测试（multi-goroutine 真实 Postgres）+ UNIQUE(session_id, seq_no) 断言 + `E2E.P0.040` | seq_no 重号 / 丢序 |
| R6 | Failure/recovery | spec D-34（hint_requested 在 002 默认 strict） | Phase 2 | `SessionEventService.handleHintRequested` 单元测试 → 409 PRACTICE_SESSION_CONFLICT + detail.policy + `E2E.P0.039` | 002 不静默返回 200 / 不调 hint AI |
| R7 | Primary | spec C-11（完成 session 触发异步 report） | Phase 3 | `SessionCompletionService.Complete` 单元测试 + repository 集成测试（单事务 5 张表 + outbox）+ `E2E.P0.041` | — |
| R8 | Boundary | spec C-21 complete 分支（success replay） | Phase 3 | idempotency middleware unit test (replay returns snapshot) + repository integration + `E2E.P0.042` | — |
| R9 | Failure/recovery | spec C-22 complete 分支（key body mismatch） | Phase 3 | conflict.go unit test + `E2E.P0.042` | envelope 不泄露首次资源 |
| R10 | Cross-layer (auth) | spec C-13 / C-23 complete + append 分支（cross-user 隔离） | Phase 2 + Phase 3 | repository 单元测试（cross-user fixture）+ handler middleware 单元测试 + `E2E.P0.042` 含 404 跨用户断言 | 用户 B 不应通过 same idempotency_key 命中用户 A 资源 |
| R11 | Boundary | spec C-24 complete 分支（concurrent single-executor） | Phase 3 | DB UNIQUE + row lock 集成测试 + `E2E.P0.042` | 同 user + 同 key + 同 fingerprint 并发只产生一个 feedback_reports / async_jobs row |
| R12 | Boundary | spec C-27 + D-35（双 key complete 走 replay） | Phase 3 | `CompleteSessionService.findExistingReport` 单元测试 + repository 集成测试（已完成 session 用新 key complete 返回既有 ReportWithJob）+ `E2E.P0.042` | 不创建第二个 feedback_reports / async_jobs；不重复发 outbox |
| R13 | Cross-layer contract | spec D-32（shared/jobs.yaml `triggerEventSemantic` 字段 + 常量 forward-binding） | Phase 0 + Phase 3 + Phase 4 | `make lint-events` 校验 enum / ownership；`make codegen-events-check` 校验 B3 生成物无 drift；generated `JobTriggerEventSemanticSourceEventOnly` / `JobTriggerEventSemanticTriggerCreatesJob` 常量 + `IsSourceEventOnly(jobType)` 谓词单元测试（`backend/internal/shared/jobs/...`）；handler-side `async_jobs(job_type, dedupe_key) WHERE status IN ('queued','running')` UNIQUE INDEX 并发集成测试；`E2E.P0.043` grep 断言 `report_generate` INSERT 仅出现在 `backend/internal/store/practice/complete_session.go` | runtime dispatcher 集成测试归 future `backend-async-runner` plan（002 阶段无 runtime dispatcher 实体）；其它 8 个 job `trigger_creates_job` 行为保留 |
| R14 | Cross-layer contract | spec D-33（B2 OpenAPI `PracticeTurn.status` 扩 5 值 + baseline rebase + regenerate） | Phase 0 + Phase 1 | `make codegen-check` + `python3 scripts/lint/conventions_drift.py --repo-root .` + handler 单元测试用 generated request/response types + fixture parity test | wire enum 不应只暴露 3 值；`follow_up_requested` / `assessed` 必须出现在 `PracticeTurn.status` enum |
| R15 | Cross-layer contract | spec D-22 + D-28（complete 同事务 placeholder + outbox 与 B3 schema 一致 + async_jobs dedupe_key 兜底） | Phase 3 + Phase 4 | `outbox_emitter` 序列化单元测试（断言与 `shared/events/practice.session.completed.json` 一致 + piiBoundary 通过；payload 字段仅含 IDs/counts）+ `async_jobs(job_type, dedupe_key=sessionId)` UNIQUE INDEX 集成测试 + `E2E.P0.041` 断言 `outbox_events WHERE event_name='practice.session.completed' AND aggregate_id=sessionId` 行数为 1（由 D-35 应用层 Replay 路径保证，DB 当前无 outbox 唯一约束） | outbox 重放不创建第二个 `report_generate` job（async_jobs UNIQUE 兜底）；outbox dedupe 限定 `async_jobs.dedupe_key` 列，`outbox_events` 表无 dedupe_key 列 |
| R16 | Cross-layer contract | spec D-3 + D-8（`practice.turn.completed` payload 与 B3 schema + AssistantAction 仅含 B2 `GenerationProvenance` wire 字段） | Phase 2 | `outbox_emitter.PracticeTurnCompleted` 单元测试 + AssistantAction provenance 单元测试（断言仅含 `promptVersion` / `rubricVersion` / `modelId` / `language` / `featureFlag` / `dataSourceVersion`）+ `E2E.P0.038` | runtime-only 字段（`feature_key`、provider、cost、latency）不进 wire；只进 `ai_task_runs` |
| R17 | Privacy/security | spec D-11 + C-16（log / metric / audit / outbox 不含明文） | Phase 2 + Phase 3 + Phase 4 | outbox emitter PII 单元测试 + cmd/api HTTP scenario privacy 断言 + legacy lint + `E2E.P0.043` | `question_text` / `answer_text` / `hint_text` / AI prompt / response 明文 / provider secret 在 log / metric / audit_events / outbox payload 中零出现 |
| R18 | Observability | spec §4.4 + A3/F1 已有边界（AI metric label allowlist；`feature_key` 仅入 `ai_task_runs`） | Phase 2 + Phase 3 | observed AIClient wiring test + `ai_task_runs` 集成测试 + metric label allowlist 单元测试 + `E2E.P0.043` | metric label 不含 `feature_key` / prompt-rubric version / provider raw model id |
| R19 | UX (API) | spec D-6（202 + ReportWithJob）+ C-11 | Phase 3 | handler 单元测试断言 HTTP 202 + ReportWithJob shape + `E2E.P0.041` | 不返回 200 / 201；ReportWithJob.job.status 必须是 `queued` |
| R20 | Regression / legacy-negative | spec D-20 + D-32 + D-34（retired 术语 / 旧 dispatcher 二次创建 / 旧 hint 处理） | Phase 4 | repo-wide grep gate | `warmup` / `single_drill` / `drill_builder` / `mistake_queue` / `growth_center` / 独立 `voice` route / `practiceModeCard` / dispatcher 旧 `report_generate` 二次创建模式 / 旧 wire turn enum 压缩（spec 历史描述的"handler 映射"备选）在 002 输出文档与代码 diff 零出现 |
| R21 | Out-of-scope boundary | 003 / 004 / 005 / 006 owner 范围不应被 002 实现 | Phase 1 + Phase 2 | mode policy 单元测试断言 002 hint 仅返回 strict 默认 409，不分支 assisted；derived plan goal 在 002 不修改 schema；voice operation 不挂 router；CASCADE / sweep 不在 002 落地 | 002 不调用 `practice.turn.lightweight_observe`；assisted 真实 hint show_hint 不出现 |

无 UI 视觉地理 parity 行（本 plan 不涉及 `ui-design/` 复刻；其消费方 frontend-workspace-and-practice / 未来 practice screen plan 分别承担 UI parity）。

## 3.6 L2 修订说明

- 2026-05-13 `plan-code-review --fix` 修订：`answer_submitted` service gate 现在要求非空 `payload.answerText`，与 spec §2.1 事件 payload contract 对齐；缺失或空白值在 append 前返回 `422 VALIDATION_FAILED`，不触发 AI / 不写 event。
- 2026-05-13 `plan-code-review --fix` 修订：`completePracticeSession` 先保留 D-35 既有 report/job replay；没有既有 report/job 时，仅允许 `running` / `waiting_user_input` / `completed` 进入 queued report/job 创建，`queued` / `completing` / `failed` / `cancelled` 返回 `409 PRACTICE_SESSION_CONFLICT`。
- 2026-05-13 `plan-code-review --fix` 修订：BDD 编号迁移为 `E2E.P0.038` ~ `E2E.P0.043`，因为 `E2E.P0.034` / `E2E.P0.035` 已由 backend-resume 场景占用；`scripts/lint/backend_practice_legacy.py` 增加可执行重复编号 / HTTP scenario test 入口反查。
- 2026-05-14 `plan-code-review --fix` 修订：`answer_submitted` 的 follow-up 分支改为读取 DB `practice_turns.follow_up_count`，忽略客户端 `payload.followUpCount`；首次答案只进入 `follow_up_requested` 且不发 `practice.turn.completed`，turn 到 `assessed` 后才按一次/turn 发 outbox。
- 2026-05-14 `plan-code-review --fix` 修订：handler 对 OpenAPI required 字段 `occurredAt` / `clientCompletedAt` 执行 422 必填校验，不再以服务端时间静默补齐缺失客户端时间。
- 2026-05-14 `plan-code-review --fix` 修订：D-35 replay 查询精确绑定 `async_jobs(job_type='report_generate', resource_type='feedback_report', dedupe_key=sessionId)`，并补强 `E2E.P0.038` / `E2E.P0.039` / `E2E.P0.040` 的当前语义覆盖。

## 4 实施步骤

### Phase 0: 跨 spec 前置修订 + Preflight

**目标**：把 D-28（落实为 D-32）、D-25（落实为 D-33）落到编码真理源；plan 002 直接修订 B2 / B3 编码真理源，**并同步更新各 owner spec history.md / spec.md Header**（按 D-30 Q1=A integrator 模式延续）；F3 baseline 仅 preflight assert。

#### 0.1 B3 `shared/jobs.yaml` `triggerEventSemantic` 字段（D-32）

在 `shared/jobs.yaml#jobs[*]` 引入 `triggerEventSemantic` 字段，缺省值视为 `trigger_creates_job`；为 `report_generate` 显式写 `triggerEventSemantic: source_event_only`。真实验证入口是 root `make lint-events` 与 `make codegen-events-check`；实现时应扩展 `scripts/lint/lint_events.py` / `scripts/lint/events_inventory.py shared/events.yaml shared/jobs.yaml shared/conventions.yaml` 的底层校验，并扩展 `backend/cmd/codegen/events` 生成物。

1. `triggerEventSemantic` 取值仅允许 `trigger_creates_job` / `source_event_only`（缺省按 `trigger_creates_job` 接受，避免 8 个既有 job 都需要显式声明）；
2. `source_event_only` 必须同时设置 `apiFacing: true` 且对应 `triggerEvent` 必须是 events.yaml 中定义的事件名；
3. 事件 codegen 必须把 Go 常量包（`backend/internal/shared/jobs/` 内 `jobs.go` 或新增生成文件）渲染为 `JobTriggerEventSemanticSourceEventOnly` / `JobTriggerEventSemanticTriggerCreatesJob` 常量与 `IsSourceEventOnly(jobType JobType) bool` 谓词供 backend 包消费；生成文件带 `DO NOT EDIT`，不得手改；
4. 002 阶段无 runtime outbox→asynq dispatcher（现仓 `backend/internal/runtime/` 不存在；现有 outbox/async_jobs 均由 handler/store 直接 INSERT），因此 "dispatch-time 跳过 `source_event_only` 任务" 是 forward-binding 契约：未来 `backend-async-runner` plan 引入 runtime dispatcher 时必须读取上述常量并在 dispatch-time 跳过。002 阶段 `report_generate` 不被二次创建的实际兜底来自 (a) handler 端复用 `async_jobs(job_type, dedupe_key) WHERE status IN ('queued','running')` UNIQUE INDEX，(b) D-35 反向查 `feedback_reports` 复用既有 `ReportWithJob`，(c) repo grep 断言 `report_generate` INSERT 仅出现在 `complete_session.go`。

同步 B3 `event-and-outbox-contract` spec.md Header `2.4 → 2.5` 与 history.md 2.5 行（"授权 backend-practice/002 Phase 0 新增 `triggerEventSemantic` 字段 + lint + 生成常量与 `IsSourceEventOnly` 谓词，并把 `report_generate` 标注为 `source_event_only`；runtime dispatcher 集成留给 future `backend-async-runner` plan"）。

#### 0.2 B2 OpenAPI `PracticeTurn.status` 扩 5 值（D-33）

修订 `openapi/openapi.yaml#components.schemas.PracticeTurn.status` 的 enum 为 `[asked, answered, follow_up_requested, assessed, skipped]`，与 `migrations/000001_create_baseline.up.sql:242` `practice_turns.status` CHECK 完全对齐。同步：

1. `openapi/baseline/openapi-v1.0.0.yaml` 按 D-21 pre-launch baseline rebase 原地修订（D-21 的 pre-launch 表述在 D-33 同样适用）。
2. Regenerate：从 repo root 运行 `make codegen-openapi`，再运行 `make codegen-check`，刷新 Go server (`backend/internal/api/generated/`) 与 TS client（`frontend/src/api/generated/`）；运行 `python3 scripts/lint/conventions_drift.py --repo-root .`、`cd backend && go build ./...`、`pnpm --filter @easyinterview/frontend typecheck`（或对应 TS check 命令）确保无 drift。
3. 同步 B2 `openapi-v1-contract` spec.md Header `1.18 → 1.19` 与 history.md 1.19 行（"授权 backend-practice/002 Phase 0 PracticeTurn.status enum 扩 5 值 pre-launch baseline rebase"）。

#### 0.3 PracticeSessions fixtures 扩展

按 §3.1 Operation Matrix 在 `openapi/fixtures/PracticeSessions/appendSessionEvent.json` 补齐 `default` / `follow-up` / `hint-strict-conflict` / `turn-skipped` / `pause-resume` / `replay` / `mismatch` / `completed` 命名场景；在 `openapi/fixtures/PracticeSessions/completePracticeSession.json` 补齐 `default` / `replay` / `mismatch` / `session-already-completed` / `cross-user-not-found`。所有命名场景必须通过 `make validate-fixtures`（或 `python3 scripts/lint/validate_fixtures.py --repo-root .`）与 contract test。

#### 0.4 F3 baseline preflight

读取 `docs/spec/prompt-rubric-registry/spec.md` v2.1 与 `docs/spec/prompt-rubric-registry/plans/001-baseline/checklist.md` 的 `状态: completed` 与 work-journal 中 `docs(prompt-rubric-registry): close 001-baseline lifecycle and record ac self-check` 行；断言 F3 已落地 `practice.session.first_question` / `practice.session.follow_up` / `practice.turn.lightweight_observe` 三个 feature_key 的 baseline 行 + Resolve API 可用。002 不修改 F3 真理源。

#### 0.5 Phase 0 收口 gate

- `make lint-events` 与 `make codegen-events-check`（含 `triggerEventSemantic` 校验与 B3 生成物 drift gate）通过
- `make codegen-check` / `python3 scripts/lint/conventions_drift.py --repo-root .` 通过
- `cd backend && go build ./...` 通过；`make validate-fixtures` 通过
- B2 / B3 owner spec history append 与 Header bump commit 在 003 实施前完成
- F3 preflight 断言通过；本 plan 状态保持 `active`

### Phase 1: AppendSessionEvent state machine 与 turn-status 域

**目标**：先落 stateless 决策与映射层，独立可单元测试；不接入 HTTP / DB。

#### 1.1 SessionEventService skeleton

新增 `backend/internal/practice/session_event.go`：定义 `SessionEventInput{SessionID, ClientEventID, Kind, OccurredAt, Payload}`（按 B2 wire request schema 命名）、`SessionEventOutcome{AssistantAction, NextSessionStatus, NextTurn, OutboxRecord(*PracticeTurnCompletedRecord), AuditMetadata}`；接受 read-only `SessionRecord` + `LatestTurn` + `PlanRecord` 输入并返回纯函数式 outcome（无 IO）。

#### 1.2 Kind router（D-34）

在 `SessionEventService.Route` 中按 `kind` 分派到 5 个 handler：`handleAnswerSubmitted`、`handleHintRequested`（默认返回 409 PRACTICE_SESSION_CONFLICT outcome + `detail.policy='hint_disabled_in_mode'`，不调用 AI）、`handleTurnSkipped`、`handleSessionPaused`、`handleSessionResumed`。未知 kind 返回 `ValidationFailed` outcome（B2 `VALIDATION_FAILED`）。

#### 1.3 AssistantAction decision tree

`handleAnswerSubmitted` 按 turn_count vs question_budget、follow_up_count、plan goal 决定 `ask_question` / `ask_follow_up` / `session_completed`；`handleTurnSkipped` 决定 `ask_question`（若未达 budget）或 `session_completed`；`handleSessionPaused` / `handleSessionResumed` 仅切换 `waiting_user_input` / `running`，AssistantAction = `session_wait`。所有 AssistantAction `provenance` 字段从 F3 follow_up 或本地 stub 填入 B2 `GenerationProvenance` 当前 wire 字段（non-AI 分支用 `rubricVersion='not_applicable'`、`featureFlag='none'` 等 spec D-10 约定）。

#### 1.4 Turn status mapping helper（D-33）

新增 `backend/internal/practice/turn_status.go`：DB internal `practice_turns.status` 5 值与 B2 wire `PracticeTurn.status` 5 值在 D-33 后完全相同，helper 仅做 typed enum 转换（generated TS / Go 类型 → string），并对 unknown / nil 输入 panic 或返回 error 而非默认值。覆盖 `follow_up_requested → follow_up_requested`、`assessed → assessed` 不再压缩到 `asked` / `answered`。

#### 1.5 Phase 1 收口 gate

- `cd backend && go test ./internal/practice/...` 全部通过（覆盖 state machine 与 mapping helper 单元测试，per `test-plan.md` Phase 1）
- 单元测试覆盖 `handleAnswerSubmitted` 3 分支 + 5 kind exhaustive switch
- AssistantAction provenance 测试断言仅含 B2 wire 字段

### Phase 2: AppendSessionEvent vertical slice

**目标**：HTTP handler + repository + outbox + idempotency 接入，闭合 C-6 / C-9 / C-14 / C-26 + cross-user 隔离。

#### 2.1 Repository AppendSessionEvent

新增 `backend/internal/store/practice/append_event.go`：方法 `AppendSessionEvent(ctx, AppendSessionEventInput) (AppendSessionEventResult, error)`，在单事务内：

1. `SELECT FOR UPDATE practice_sessions WHERE id=$1 AND user_id=$2`；不匹配 → `ErrSessionNotFound`（映射 404 `PRACTICE_SESSION_NOT_FOUND`）。
2. `SELECT * FROM practice_session_events WHERE session_id=$1 AND client_event_id=$2`：存在 → 取首次 `payload` 与对应 response（按 B2 `SessionEventResult`）；fingerprint match → replay；mismatch → 返回 `ErrClientEventFingerprintMismatch`（映射 409）。
3. 计算 `seq_no = COALESCE(MAX(seq_no), 0) + 1`，INSERT `practice_session_events(seq_no, event_type, payload, client_event_id, ...)`；INSERT/UPDATE `practice_turns`（DB status 推进，依 outcome `NextTurn`）；UPDATE `practice_sessions.status`（如有变更）；当 `OutboxRecord` 非空，INSERT `outbox_events` 一行（`event_name='practice.turn.completed'`、aggregate / payload 按 §3.1 schema）。`appendSessionEvent` 按 spec §4.3 不写 `audit_events`，隐私证据由 outbox/log/metric/ai_task_runs 红线覆盖。
4. 提交事务，返回 `(AppendSessionEventResult{Acknowledged, Session, AssistantAction}, nil)`。

依赖 `outbox_emitter.BuildPracticeTurnCompletedPayload(turn, follow_up_count, answer_char_length)`（新增到 `backend/internal/store/practice/outbox_emitter.go`，复用 `assertNoPracticeOutboxPII`）。

#### 2.2 Service AppendSessionEvent

新增 `backend/internal/practice/append_session_event_service.go`（或合并入 `session_event.go`）：组合 Phase 1 `SessionEventService.Route` + repository；若 outcome.AssistantAction 需要 AI follow_up，在事务外调用 A3 observed AIClient + F3 `practice.session.follow_up`；调用失败时 outcome 退化为 `ask_question` placeholder 并写入 `failure_code`（与 spec D-19 一致），不阻塞用户答题循环（具体退化策略由本 plan 决议，详见 Phase 2.2 verification 单元测试）。

#### 2.3 Handler AppendSessionEvent

新增 `backend/internal/api/practice/append_session_event.go`（或在 handler.go 内挂方法）：

- 拒绝携带 `Idempotency-Key` header（D-7）：返回 `400 VALIDATION_FAILED` + `detail.policy='use_client_event_id'`。
- Decode B2 generated request；映射到 service 输入；返回 `200 + SessionEventResult`（按 B2 wire schema，含 `AssistantAction.provenance`）。
- 错误映射复用 `error_mapping.go`：`PRACTICE_SESSION_NOT_FOUND` → 404，`PRACTICE_SESSION_CONFLICT` → 409 with detail，`VALIDATION_FAILED` → 422 / 400 取决于位置，AI 错误 → 502 / 503。

#### 2.4 Router 挂接

在 `backend/internal/api/practice/handler.go` 注册 `POST /practice/sessions/{sessionId}/events` 路由；中间件链不挂 `idempotency.Middleware`。

#### 2.5 BDD-Gate Phase 2

- BDD-Gate: 验证 `E2E.P0.038` 通过
- BDD-Gate: 验证 `E2E.P0.039` 通过
- BDD-Gate: 验证 `E2E.P0.040` 通过

### Phase 3: CompletePracticeSession vertical slice

**目标**：HTTP handler + repository + outbox + async_jobs + feedback_reports placeholder + idempotency middleware 接入，闭合 C-11 / C-21 / C-22 / C-23 / C-24 / C-27 + cross-user 隔离。

#### 3.1 Repository CompleteSession

新增 `backend/internal/store/practice/complete_session.go`：方法 `CompleteSession(ctx, CompleteSessionInput) (CompleteSessionResult, error)`，在单事务内：

1. `SELECT FOR UPDATE practice_sessions WHERE id=$1 AND user_id=$2`；不匹配 → `ErrSessionNotFound`。
2. 反向查 `feedback_reports WHERE session_id=$1` 与 `async_jobs WHERE job_type='report_generate' AND dedupe_key=$1`（D-35）：若存在则返回 `CompleteSessionResult{Replay=true, Report, Job, ResponseBody}`；handler 据此 replay 不写新行。
3. 若不存在，UPDATE `practice_sessions.status='completing'`、`updated_at=now()`（B4 当前无 `completing_at` 列，002 不动 migration；`status='completing'` 与 `updated_at` 共同充当时态标记）；INSERT `practice_session_events(event_type='session_completed', seq_no=MAX+1)`；INSERT `feedback_reports(status='queued', user_id, session_id, language, target_job_id)`；INSERT `async_jobs(job_type='report_generate', status='queued', dedupe_key=sessionId, resource_type='feedback_report', resource_id=reportId, payload jsonb)`；INSERT `outbox_events(event_name='practice.session.completed', event_version=1, aggregate_type='practice_session', aggregate_id=sessionId, payload)`（`outbox_events` 表无 `dedupe_key` 列，唯一性由 D-35 应用层 Replay 路径保证，不依赖 DB UNIQUE）；INSERT `audit_events`。
4. 返回 `CompleteSessionResult{Replay=false, Report, Job}`；handler 用结果组装 `ReportWithJob`，并把 response_body snapshot 交给 idempotency middleware 写入。

依赖新增 `outbox_emitter.BuildPracticeSessionCompletedPayload`。

#### 3.2 Service CompleteSession

新增 `backend/internal/practice/complete_session_service.go`：组合 repository；若 repository 返回 `Replay=true`，service 仍正常返回相同 ReportWithJob（D-35），handler 据 idempotency middleware 选择 status code 与是否 snapshot。

#### 3.3 Handler CompletePracticeSession + idempotency middleware

新增 `backend/internal/api/practice/complete_practice_session.go`：

- 通过 `idempotency.Middleware.Handler(domain="practice", operation="completePracticeSession", resolveUser=...)` 装饰；middleware 在 `Reserve` 阶段完成 reservation + replay + mismatch + cross-user。
- `Reserve` 返回非 `StateReplay` 时，middleware 把请求 forward 给 handler/service；service 在同事务内完成 session/event/feedback_reports/async_jobs/outbox/audit 写入并返回 HTTP 202 + response body；middleware 缓冲 response body 后，2xx 调 `MarkSucceeded(snapshot=response_body)`，非 2xx 调 `MarkFailed`（参考 `backend/internal/middleware/idempotency/middleware.go` 当前实现）。
- D-35：若 service 返回 `Replay=true`，handler 返回原始 HTTP 202 + 既有 ReportWithJob；middleware 仍按上述顺序写 idempotency snapshot（同一 response_body），不重复写 feedback_reports / async_jobs / outbox。
- 错误映射：service 错误 → spec D-19 / D-26 错误码；middleware mismatch → 409。

#### 3.4 Router 挂接

在 `backend/internal/api/practice/handler.go` 注册 `POST /practice/sessions/{sessionId}/complete` 路由，包裹 idempotency middleware。

#### 3.5 BDD-Gate Phase 3

- BDD-Gate: 验证 `E2E.P0.041` 通过
- BDD-Gate: 验证 `E2E.P0.042` 通过

### Phase 4: 隐私 / 观测 / Legacy-Negative / Handoff

**目标**：固化 D-11 / D-32 / D-33 / D-34 / D-35 的反查 gate；为 003-006 / backend-review handoff 留干净接口。

#### 4.1 Privacy / observability gates

补 redaction 单元测试断言 `outbox_events` payload、complete path `audit_events.metadata`、log structured fields、A3 metric label 不含 `question_text` / `answer_text` / `hint_text` / AI prompt body / response body / provider secret；同时断言 `appendSessionEvent` 不写 `audit_events`。metric label 命中 F1 allowlist。

#### 4.2 Legacy-negative grep

repo-wide grep gate（在测试或 lint 中执行）：

- `report_generate` 二次创建 / dispatch-time 创建模式（旧 dispatcher 行为）在 backend 包零出现，唯一 INSERT 位点在 `complete_session.go`；
- 旧 wire turn status 压缩（如 spec D-25 备选中描述的 `mapInternalToWire` 或 `compressTurnStatus`）不出现；
- `warmup` / `single_drill` / `drill_builder` / `mistake_queue` / `growth_center` / 独立 `voice` route / `practiceModeCard` 在 002 输出 diff 零出现。

#### 4.3 Handoff doc 更新

更新 `backend/internal/api/practice/README.md`（若不存在则按 backend/README.md 风格新增简明 handoff 段落，仅记录 002 新增 endpoint / 中间件挂法 / 接口签名 / handoff 给 003-006 与 backend-review 的边界）。本步骤不引入新设计文档，只补包级 README。

#### 4.4 BDD-Gate Phase 4

- BDD-Gate: 验证 `E2E.P0.043` 通过

## 5 验收标准

- Phase 0 ~ Phase 4 checklist 全部勾选
- 关联 BDD 场景 `E2E.P0.038` / `E2E.P0.039` / `E2E.P0.040` / `E2E.P0.041` / `E2E.P0.042` / `E2E.P0.043` 均由 `backend/cmd/api/practice_http_scenario_test.go` 的 HTTP scenario tests 执行通过
- B2 OpenAPI `PracticeTurn.status` enum 已扩 5 值；B3 `shared/jobs.yaml#report_generate` 已带 `triggerEventSemantic: source_event_only`；两个 owner spec 已 Header bump + history append
- `make codegen-check` / `make lint-events` / `make codegen-events-check` / `python3 scripts/lint/conventions_drift.py --repo-root .` / `cd backend && go test ./...` 全绿
- 002 范围内代码与文档中无 §3.5 R20 列出的 legacy 术语 / 旧 dispatcher 模式 / 旧 wire turn enum 压缩

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| D-33 扩 wire enum 破坏 frontend 既有 generated types 假设（如硬编码 3 值 union） | Phase 0.2 后立即 `cd frontend && pnpm tsc --noEmit`（或 fixture-backed test）反查；如发现 hardcoded 3 值，写入 `frontend-workspace-and-practice` follow-up handoff，不在 002 修复 |
| 002 阶段无 runtime outbox→asynq dispatcher，D-32 `triggerEventSemantic` 仅作为 forward-binding 元数据；未来 `backend-async-runner` plan 引入 dispatcher 时若忽略该字段会回归（二次创建 `report_generate`） | 002 内可达成的兜底：(a) `make lint-events` / `make codegen-events-check` 校验 jobs.yaml 字段与生成物；(b) `backend/internal/shared/jobs/` 暴露 `JobTriggerEventSemantic*` 常量与 `IsSourceEventOnly(jobType)` 谓词单元测试；(c) handler 端 `async_jobs(job_type, dedupe_key) WHERE status IN ('queued','running')` UNIQUE INDEX 集成测试；(d) D-35 反向查 `feedback_reports` 复用 ReportWithJob；(e) repo grep 断言 `report_generate` INSERT 仅出现在 `complete_session.go`。未来 dispatcher plan 必须读取常量并在 dispatch-time 跳过 `source_event_only` 任务，否则视为回归（R13 / R15 反查行覆盖） |
| `SELECT FOR UPDATE` 在高并发下产生死锁或行锁冲突 | Phase 2 集成测试用 multi-goroutine + 真实 Postgres 验证（参考 001 reservation pattern）；如出现死锁，仅采用 PostgreSQL 默认行锁，不引入 `pg_advisory_xact_lock`，避免与 001 reservation lock 顺序混乱 |
| D-35 反向查 `feedback_reports` 与 idempotency middleware 顺序导致 race（middleware reserve 在前但 service 查到既有 report） | service 内先 `SELECT FOR UPDATE session` 再查 `feedback_reports`；middleware 在 service 成功返回后 `MarkSucceeded` 时 snapshot；若 race，UNIQUE(session_id) on feedback_reports 兜底 |
| AppendSessionEvent 大量并发写入 `outbox_events` 导致行膨胀 | 不在 002 引入分区或归档；通过 R5 集成测试断言完成的 turn 数与 `outbox_events WHERE event_name='practice.turn.completed' AND aggregate_id=sessionId` 行数 1:1；后续归档归 platform owner |
| F3 `practice.session.follow_up` AI 调用失败导致 appendSessionEvent 卡住 | Phase 2.2 service 设计：AI 失败时 fallback 到 `ask_question`（不阻塞用户答题循环），把 `failure_code` 写入 service outcome / structured log 摘要（append 不写 `audit_events`）；与 spec D-19 一致；单元测试覆盖该分支 |
| Phase 0.3 fixtures 与 002 后续 BDD 场景的 named scenario 不一致 | Phase 0.3 完成后立刻在 Phase 2 / Phase 3 BDD scenario `data/seed-input.md` 中引用同名 scenario name；fixture validator 与 contract test 联动 |
| 003 接手 hint mode 切分时打破 002 默认 strict 行为 | 002 hint 处理由 `handleHintRequested(mode)` 函数化，function signature 直接接受 mode；003 只替换函数体内 assisted 分支返回 `show_hint`；不修改 router / service 上层 |
