# 002 — Event Loop and Completion Checklist

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-05-13

**关联计划**: [plan](./plan.md)

## Phase 0: 跨 spec 前置修订 + Preflight

- [x] 0.1 在 `shared/jobs.yaml` 引入 `triggerEventSemantic` 字段并把 `report_generate` 标注为 `source_event_only`；其它 8 个 job 显式或隐式 `trigger_creates_job`（D-32）
- [x] 0.2 在 `scripts/lint/lint_events.py` / `scripts/lint/events_inventory.py shared/events.yaml shared/jobs.yaml shared/conventions.yaml` 增加 `triggerEventSemantic` 取值校验、`source_event_only` 必须 `apiFacing: true` 且 `triggerEvent` 在 events.yaml 中、generated const 暴露语义常量（验证：`make lint-events` 通过 + 单元测试断言 const 暴露）
- [x] 0.3 扩展 `backend/cmd/codegen/events`，由 codegen 在 `backend/internal/shared/jobs/` 生成 `JobTriggerEventSemanticSourceEventOnly` / `JobTriggerEventSemanticTriggerCreatesJob` Go 常量 + `IsSourceEventOnly(jobType JobType) bool` 谓词（以及对应 TS 生成物如适用；验证：`make codegen-events-check` + `cd backend && go test ./internal/shared/jobs/...` 单元测试覆盖常量值、`IsSourceEventOnly("report_generate")=true`、`IsSourceEventOnly("target_import")=false` 等；002 阶段无 runtime outbox→asynq dispatcher，dispatch-time 跳过逻辑由未来 `backend-async-runner` plan 接管，本 plan 不引入 runtime 包；禁止手改 `DO NOT EDIT` 生成文件）
- [x] 0.4 同步 B3 `event-and-outbox-contract` `spec.md` Header `2.4 → 2.5` + `history.md` 2.5 行 + §相关条目，记录 "授权 backend-practice/002 Phase 0"
- [x] 0.5 修订 `openapi/openapi.yaml#components.schemas.PracticeTurn.status` enum 为 5 值 `[asked, answered, follow_up_requested, assessed, skipped]`（D-33）
- [x] 0.6 同步原地 rebase `openapi/baseline/openapi-v1.0.0.yaml` 与生成产物：root `make codegen-openapi` / `make codegen-check` / `cd backend && go build ./...` / `pnpm --filter @easyinterview/frontend typecheck`（或当前等价 TS check 命令）全通过；生成物路径为 `backend/internal/api/generated/` 与 `frontend/src/api/generated/`
- [x] 0.7 运行 `python3 scripts/lint/conventions_drift.py --repo-root .` 通过；shared/conventions Go/TS 生成物无 drift
- [x] 0.8 同步 B2 `openapi-v1-contract` `spec.md` Header `1.18 → 1.19` + `history.md` 1.19 行，记录 "授权 backend-practice/002 Phase 0 PracticeTurn.status pre-launch baseline rebase"
- [x] 0.9 扩展 `openapi/fixtures/PracticeSessions/appendSessionEvent.json` 命名场景：`default` / `follow-up` / `hint-strict-conflict` / `turn-skipped` / `pause-resume` / `replay` / `mismatch` / `completed`（验证：fixture validator + contract test）
- [x] 0.10 扩展 `openapi/fixtures/PracticeSessions/completePracticeSession.json` 命名场景：`default` / `replay` / `mismatch` / `session-already-completed` / `cross-user-not-found`（验证：fixture validator + contract test）
- [x] 0.11 F3 preflight assert：读取 `docs/spec/prompt-rubric-registry/spec.md` v2.1 + plan `001-baseline` `completed` 与 work-journal commit "close 001-baseline lifecycle"；断言 `practice.session.first_question` / `practice.session.follow_up` / `practice.turn.lightweight_observe` baseline 行存在且 Resolve API 可用（验证：focused integration test on F3 RegistryClient）

## Phase 1: AppendSessionEvent state machine 与 turn-status 域

- [x] 1.1 新增 `backend/internal/practice/session_event.go`：`SessionEventInput` / `SessionEventOutcome` 类型 + `SessionEventService.Route(ctx, input, session, latestTurn, plan) (SessionEventOutcome, error)`（验证：`session_event_test.go` 断言 5 种 kind 的 outcome shape）
- [x] 1.2 实现 `handleAnswerSubmitted`：按 `turn_count` / `question_budget` / `follow_up_count` / `plan.goal` 决定 `ask_question` / `ask_follow_up` / `session_completed`（验证：`session_event_test.go` 覆盖 3 分支 + 边界值）
- [x] 1.3 实现 `handleHintRequested`（D-34）：无条件返回 409 PRACTICE_SESSION_CONFLICT outcome + `detail.policy='hint_disabled_in_mode'` + `detail.mode=session.mode`；002 阶段不调 AI（验证：`session_event_test.go` 断言 outcome 形状，所有 mode/goal 输入均返回 409）
- [x] 1.4 实现 `handleTurnSkipped` / `handleSessionPaused` / `handleSessionResumed`（验证：单元测试覆盖 status 推进与 AssistantAction `session_wait` / `ask_question` / `session_completed`）
- [x] 1.5 实现 unknown kind 返回 `VALIDATION_FAILED` outcome（验证：单元测试断言无 panic / 无 500）
- [x] 1.6 AssistantAction `provenance` 填充 helper：non-AI 分支用 `rubricVersion='not_applicable'` / `featureFlag='none'` / `dataSourceVersion='static'` 等 B2 wire 默认；AI 分支由 Phase 2 service 在调用 F3 后填入（验证：`session_event_test.go` 断言 provenance 仅含 B2 wire 字段）
- [x] 1.7 新增 `backend/internal/practice/turn_status.go`：typed enum 转换（DB 5 值 ↔ wire 5 值，D-33 后完全相同）+ unknown 输入返回 error；不再做 D-25 备选的"压缩到 3 值"映射（验证：`turn_status_test.go` 覆盖 5 值正向 + unknown / 空字符串负向）
- [x] 1.8 Phase 1 收口：`cd backend && go test ./internal/practice/...` 全通过

## Phase 2: AppendSessionEvent vertical slice

- [x] 2.1 新增 `backend/internal/store/practice/append_event.go`：`AppendSessionEvent(ctx, AppendSessionEventInput) (AppendSessionEventResult, error)`，单事务内 `SELECT FOR UPDATE session` → `clientEventId` replay/mismatch 检查 → `seq_no=MAX+1` → INSERT/UPDATE event/turn/session/outbox；append 路径不写 `audit_events`（验证：`append_complete_test.go` repository 测试 + `go test ./internal/store/practice -count=1`）
- [x] 2.2 扩展 `backend/internal/store/practice/outbox_emitter.go`：`BuildPracticeTurnCompletedPayload(turn, follow_up_count, answer_char_length)` + `assertNoPracticeOutboxPII` 校验（验证：`outbox_emitter_test.go` 断言 payload 与 `shared/events/practice.turn.completed.*` 一致 + PII 红线）
- [x] 2.3 新增 `backend/internal/practice/append_session_event_service.go`（或合并入 `session_event.go`）：组合 Phase 1 Route + 事务外 F3 `practice.session.follow_up` AI 调用（仅 `ask_follow_up` 分支）+ A3 observed AIClient + outcome 退化策略（验证：`append_session_event_service_test.go` 用 fake F3 + fake AIClient 覆盖成功 / AI 失败退化 / 非 AI 分支）
- [x] 2.4 新增 `backend/internal/api/practice/session_event_handlers.go` handler：拒绝 `Idempotency-Key` header → 400 `VALIDATION_FAILED` + `detail.policy='use_client_event_id'`；正常路径返回 200 + B2 `SessionEventResult`（验证：`session_event_handlers_test.go` 覆盖 5 kind + replay + mismatch + cross-user 404）
- [x] 2.5 错误映射复用 `backend/internal/practice/error_mapping.go`；新增 `clientEventIdMismatchToConflict` / `idempotencyHeaderNotAllowed` mapping（验证：handler/service 错误映射测试覆盖新增映射）
- [x] 2.6 Router 注册：`backend/internal/api/practice/handler.go` + `backend/cmd/api/main.go` 挂 `POST /practice/sessions/{sessionId}/events`，**不**挂 idempotency middleware（验证：HTTP scenario 与 handler 单元测试断言路径命中且 header policy 生效）
- [x] 2.7 BDD-Gate: 验证 `E2E.P0.034` 通过（answer_submitted → AssistantAction 主流程，Go HTTP scenario `TestE2EP0034PracticeEventLoopAnswerFlow`）
- [x] 2.8 BDD-Gate: 验证 `E2E.P0.035` 通过（clientEventId replay / mismatch + 5 kind exhaustive + Idempotency-Key header 拒绝 + hint 默认 strict 409，Go HTTP scenario `TestE2EP0035PracticeEventIdempotencyKindRouterAndHeaderPolicy`）
- [x] 2.9 BDD-Gate: 验证 `E2E.P0.038` 通过（stale-turn conflict 代理并发竞争；seq_no / row-lock 约束由 repository 与 DB UNIQUE gate 覆盖，Go HTTP scenario `TestE2EP0038PracticeEventConcurrentSeqNoStaleTurnConflict`）

## Phase 3: CompletePracticeSession vertical slice

- [x] 3.1 新增 `backend/internal/store/practice/complete_session.go`：`CompleteSession(ctx, CompleteSessionInput) (CompleteSessionResult, error)`，单事务先 `SELECT FOR UPDATE session` + 反向查 `feedback_reports` / `async_jobs(report_generate, dedupe_key=sessionId)`；存在则返回 `Replay=true` + 既有 ReportWithJob；否则 INSERT 5 张表 + outbox（验证：`append_complete_test.go` repository 测试 + replay 路径单元测试）
- [x] 3.2 扩展 `outbox_emitter.go`：`BuildPracticeSessionCompletedPayload(session, plan, turn_count, language)` + PII 红线断言（验证：`outbox_emitter_test.go` 覆盖 schema parity）
- [x] 3.3 新增 `backend/internal/practice/complete_session_service.go`：组合 repository；replay 路径返回相同 ReportWithJob；新建路径返回新 ReportWithJob（验证：`complete_session_service_test.go` 用 fake repository 覆盖两路径）
- [x] 3.4 新增 `backend/internal/api/practice/session_event_handlers.go`：装饰 `idempotency.Middleware.Handler(domain="practice", operation="completePracticeSession", resolveUser=...)`；middleware reserve 后调用 service；将 response_body snapshot 传给 `MarkSucceeded`（验证：`session_event_handlers_test.go` 端到端覆盖正常 + replay + mismatch + cross-user + D-35 双 key）
- [x] 3.5 D-35 集成验证：repository / HTTP scenario 跑两次 complete（同 session 不同 key），断言 `feedback_reports` 行数仍为 1、`async_jobs` 行数仍为 1、`outbox_events` 行数仍为 1，第二次返回首次 reportId 与 jobId
- [x] 3.6 错误映射：service 错误 → `error_mapping.go` 对应码；middleware mismatch → 409；session not found → 404 PRACTICE_SESSION_NOT_FOUND（验证：handler/service 错误映射测试覆盖）
- [x] 3.7 Router 注册：`POST /practice/sessions/{sessionId}/complete` 通过 idempotency middleware（验证：handler 单元测试 + cmd/api HTTP scenario 断言 middleware snapshot/replay 生效）
- [x] 3.8 BDD-Gate: 验证 `E2E.P0.039` 通过（completePracticeSession 202 + ReportWithJob + queued report/job + outbox emit，Go HTTP scenario `TestE2EP0039PracticeSessionCompleteCreatesQueuedReportJob`）
- [x] 3.9 BDD-Gate: 验证 `E2E.P0.040` 通过（complete idempotency 矩阵 + D-35 双 key replay + dispatcher 不二次创建，Go HTTP scenario `TestE2EP0040PracticeSessionCompleteIdempotencyMatrix`）

## Phase 4: 隐私 / 观测 / Legacy-Negative / Handoff

- [x] 4.1 Redaction 单元测试：断言 `practice.turn.completed` / `practice.session.completed` outbox payload、complete path `audit_events.metadata`、log structured fields、A3 metric label 不含 `question_text` / `answer_text` / `hint_text` / AI prompt / response 明文 / provider secret；同时断言 append 路径不写 `audit_events`（验证：`outbox_emitter_test.go`、store/HTTP scenario privacy assertions 与 `TestE2EP0041PracticeEventLoopPrivacyAndLegacyNegativeSurface`）
- [x] 4.2 Metric label allowlist 单元测试：A3 metric label 命中 F1 allowlist，不含 `feature_key` / prompt-rubric version / provider raw model id（验证：A3 observed AIClient 既有 allowlist + 002 service tests 不扩展 metric label surface）
- [x] 4.3 Repo-wide grep gate（在 plan 收口前由测试或 lint 执行）：
    - `report_generate` INSERT 在 backend 包中仅出现在 `backend/internal/store/practice/complete_session.go` 一处（handler-side INSERT 位点；002 阶段无 runtime dispatcher 包，未来 `backend-async-runner` plan 引入 dispatcher 时必须按 `IsSourceEventOnly` 谓词在 dispatch-time 跳过）
    - 旧 wire turn status 压缩函数 / mapping（`compressTurnStatus` / `mapInternalToWire` / `legacyTurnStatusMapping` 等）零出现
    - `warmup` / `single_drill` / `drill_builder` / `mistake_queue` / `growth_center` / 独立 `voice` route / `practiceModeCard` 在本 plan 输出 diff 零出现
- [x] 4.4 更新 `backend/internal/api/practice/README.md`（若不存在则按 backend/README.md 风格新增）：记录 002 新增 endpoint / 中间件挂法 / handoff 给 003 (mode-policies) / 004 (derived-plans) / 005 (voice) / 006 (privacy-cascade) / backend-review（report generation）的边界（验证：手工 review + grep `appendSessionEvent` / `completePracticeSession` 在 handoff 段落中均被提及）
- [x] 4.5 BDD-Gate: 验证 `E2E.P0.041` 通过（隐私 / 观测 / legacy-negative + D-32 dispatcher 不二次创建 + D-33 wire enum drift 反查，Go HTTP scenario `TestE2EP0041PracticeEventLoopPrivacyAndLegacyNegativeSurface`）
- [x] 4.6 收口 gate：`cd backend && go test ./...` 全绿；`make lint-events` 通过；`make codegen-events-check` 通过；`make codegen-check` 通过；`python3 scripts/lint/conventions_drift.py --repo-root .` 通过
- [x] 4.7 更新 `docs/spec/backend-practice/plans/INDEX.md` 把 002 从 active 推进到 completed；`docs/spec/backend-practice/history.md` 追加 1.7 行（"002 落地：event-loop + completion，关联 D-32/D-33/D-34/D-35 plan-level 决策"）；同步 `docs/spec/backend-practice/spec.md` Header 1.6 → 1.7 与 §7 row 2 描述（反映 002 plan-level 决策 D-32 / D-33 / D-34 / D-35：B3 `triggerEventSemantic: source_event_only`、B2 `PracticeTurn.status` wire enum 扩 5 值、hint 默认 strict 409、已完成 session 双 key replay）；D-25 末尾追加 002 落实备注

## 收口证据

- `cd backend && go test ./internal/api/practice ./internal/practice ./internal/store/practice ./internal/middleware/idempotency ./cmd/api -count=1`
- `cd backend && go test ./...`
- `make lint-events`
- `make codegen-events-check`
- `make codegen-check`
- `make validate-fixtures`
- `python3 scripts/lint/conventions_drift.py --repo-root .`
- `python3 scripts/lint/backend_practice_legacy.py --repo-root .`
- `make docs-check`
- `git diff --check`
