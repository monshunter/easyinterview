# 002 — Event Loop and Completion Test Plan

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-05-14

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)

## 1 测试策略概览

本 plan 是 feature-behavior + contract + code-internal 多类型组合。TDD 与 contract gate 必须共同收口；BDD scenario 覆盖用户可见行为切片（appendSessionEvent / completePracticeSession 的 HTTP API + DB 状态 + outbox + 红线）。本测试计划只覆盖 **单元 / 集成 / contract / drift / lint** 层；BDD 场景见 [bdd-plan](./bdd-plan.md) 与 [bdd-checklist](./bdd-checklist.md)。

测试目标按 phase 与 [plan §3.5 Coverage Matrix](./plan.md#35-coverage-matrix) 行号映射；不引入硬编码代码覆盖率百分比作为 gate；如团队需要观测覆盖率，仅作背景指标。

## 2 Coverage Matrix（测试视角投射）

| 测试源 | Coverage Matrix 行 | Phase | 测试形态 | 文件 / 命令 |
|--------|-------------------|-------|----------|------------|
| shared/jobs.yaml `triggerEventSemantic` enum + lint | R13 | Phase 0 + Phase 4 | drift + 单元 | `make lint-events` + `make codegen-events-check` + generated jobs Go const + `IsSourceEventOnly` 谓词单元测试 |
| `JobTriggerEventSemantic*` 常量 + `IsSourceEventOnly` 谓词暴露 | R13 | Phase 0 | 单元 | `cd backend && go test ./internal/shared/jobs/...` 断言常量与谓词；002 阶段无 runtime outbox→asynq dispatcher（dispatch-time 跳过 binding 由 future `backend-async-runner` plan 接管） |
| `report_generate` INSERT 仅出现在 `complete_session.go` + async_jobs UNIQUE 兜底 | R13 / R15 | Phase 3 + Phase 4 | 集成 + grep | `async_jobs(job_type, dedupe_key)` UNIQUE INDEX 集成测试（并发 / D-35 双 key）+ repo grep gate 断言 `report_generate` INSERT 位点 |
| OpenAPI `PracticeTurn.status` enum 扩 5 值 | R14 | Phase 0 + Phase 1 | drift + 单元 | `make codegen-check` + `python3 scripts/lint/conventions_drift.py --repo-root .` + handler 单元测试用 generated types |
| openapi fixtures extension (events / complete) | R14 / R15 | Phase 0 | contract + drift | `make validate-fixtures`（或 `python3 scripts/lint/validate_fixtures.py --repo-root .`）+ Prism / contract 测试 |
| F3 baseline preflight | (Plan §6 风险 + R16) | Phase 0 | preflight script + 单元 | F3 baseline 状态读取断言 + `RegistryClient.Resolve` mock test |
| SessionEventService state machine | R1 / R3 / R6 / R16 | Phase 1 + Phase 2 | 单元 | `cd backend && go test ./internal/practice/...` (`session_event_test.go`) |
| `answer_submitted` payload + server-owned follow-up contract | R1 / R3 | Phase 2 | 单元 + HTTP scenario | `cd backend && go test ./internal/practice -run 'TestHandleAnswerSubmittedDecisionBranches|TestAppendSessionEventRejectsMissingAnswerText|TestAppendSessionEventRejectsStaleTurnID' -count=1` + `cd backend && go test ./cmd/api -run 'TestE2EP0038|TestE2EP0039|TestE2EP0040' -count=1` |
| required client timestamp fields | R1 / R7 / R19 | Phase 2 + Phase 3 | handler 单元 | `cd backend && go test ./internal/api/practice -run 'TestAppendSessionEventRequiresOccurredAt|TestCompletePracticeSessionRequiresClientCompletedAt' -count=1` |
| turn-status mapping (D-33) | R14 | Phase 1 | 单元 | `cd backend && go test ./internal/practice -run TestTurnStatus -count=1` |
| AppendSessionEvent repository | R1 / R2 / R3 / R4 / R5 / R10 | Phase 2 | 集成 + 单元 | `cd backend && go test ./internal/store/practice -run TestAppendSessionEvent -count=1` |
| AppendSessionEvent handler | R1 / R3 / R4 / R6 / R10 / R16 | Phase 2 | 单元 + contract | `cd backend && go test ./internal/api/practice -run TestAppendSessionEvent -count=1` + fixture parity |
| outbox emitter practice.turn.completed | R16 / R17 | Phase 2 | 单元（序列化 + 红线断言） | `outbox_emitter_test.go` |
| AssistantAction provenance | R16 | Phase 2 | 单元 | `session_event_test.go` 子集 |
| F3 follow_up AI 调用与降级 | R1 / R16 / R17 | Phase 2 | 单元（fake F3 + fake AIClient） | `append_session_event_service_test.go` |
| CompleteSession repository | R7 / R11 / R12 / R15 / R19 | Phase 3 | 集成 + 单元 | `cd backend && go test ./internal/store/practice -run TestCompleteSession -count=1` |
| CompleteSession status guard | R7 / R12 / R19 | Phase 3 | 单元 + HTTP scenario | `cd backend && go test ./internal/store/practice -run 'TestSQLRepositoryCompleteSession|TestCanCompletePracticeSessionStatus' -count=1` + `cd backend && go test ./cmd/api -run TestE2EP0042PracticeSessionCompleteIdempotencyMatrix -count=1` |
| CompleteSession service replay (D-35) | R12 / R7 | Phase 3 | 单元 | `complete_session_service_test.go` |
| CompletePracticeSession handler + idempotency middleware | R8 / R9 / R11 / R19 | Phase 3 | 单元 + contract | `cd backend && go test ./internal/api/practice -run TestCompletePracticeSession -count=1` + fixture parity |
| outbox emitter practice.session.completed | R15 / R17 | Phase 3 | 单元（序列化 + 红线断言） | `outbox_emitter_test.go` |
| Idempotency middleware reservation + replay + mismatch + cross-user (复用 001 抽象) | R8 / R9 / R10 / R11 | Phase 3 | 单元 + 集成（复用 001 base） | `cd backend && go test ./internal/middleware/idempotency/...` |
| async_jobs dedupe_key UNIQUE | R12 / R15 | Phase 3 | 集成 | `append_complete_test.go` + cmd/api HTTP scenario D-35 assertions |
| Privacy redaction | R17 | Phase 2 + Phase 3 + Phase 4 | 单元 + HTTP scenario | `outbox_emitter_test.go` + cmd/api scenario privacy assertions |
| Metric label allowlist | R18 | Phase 2 + Phase 3 | 单元 | A3 observed AIClient existing allowlist + 002 service tests no new label surface |
| Out-of-scope boundary（hint / lightweight observe / derived plan / voice / cascade） | R21 | Phase 1 + Phase 2 | 单元 + 反查 | `session_event_test.go` + repo grep |
| Legacy-negative grep | R20 | Phase 4 | repo grep gate | `python3 scripts/lint/backend_practice_legacy.py --repo-root .` |
| BDD scenario ID collision gate | R20 | Phase 4 | lint + 单元 | `python3 -m pytest scripts/lint/backend_practice_legacy_test.py -q` + `python3 scripts/lint/backend_practice_legacy.py --repo-root .` |

## 3 Phase 0: 跨 spec 前置修订 + Preflight

| 任务 | 测试文件 / 命令 | 预期 Red/Green 证据 |
|------|----------------|---------------------|
| `shared/jobs.yaml` `triggerEventSemantic` enum + lint | `make lint-events` + `make codegen-events-check` + jobs Go const 单元测试 | Red: enum 校验缺失 / `report_generate` 未标注 / 生成物未更新；Green: `report_generate` 显式 `source_event_only`，其它 8 个 job 缺省值视为 `trigger_creates_job`；generated `JobTriggerEventSemanticSourceEventOnly` / `JobTriggerEventSemanticTriggerCreatesJob` 常量与 `IsSourceEventOnly(jobType) bool` 谓词在 `backend/internal/shared/jobs/` 暴露 |
| `IsSourceEventOnly` 谓词 + repo grep 兜底（替代 002 阶段不可达成的 dispatcher 集成测试） | `cd backend && go test ./internal/shared/jobs/...` 谓词单元测试 + repo grep 脚本 + `async_jobs(job_type, dedupe_key) WHERE status IN ('queued','running')` UNIQUE INDEX 集成测试 | Red: 谓词缺失 / `report_generate` INSERT 出现在 `complete_session.go` 以外 / UNIQUE INDEX 缺失 / handler 二次进入产生多份 row；Green: 谓词正确 + grep 只命中 `complete_session.go` + UNIQUE 阻断二次 INSERT。备注：002 阶段无 runtime outbox→asynq dispatcher 包，"dispatcher 在 dispatch-time 跳过 `source_event_only`" 是 forward-binding 契约，集成测试归 future `backend-async-runner` plan；生成文件只能由 `backend/cmd/codegen/events` 渲染 |
| OpenAPI `PracticeTurn.status` enum 扩 5 值 | `make codegen-check` + `python3 scripts/lint/conventions_drift.py --repo-root .` + handler 编译 | Red: 5 值缺失 / generated types 不一致 / baseline 未同步；Green: 5 值齐全 + Go / TS 生成物无 drift + baseline 同步 rebase |
| 旧 wire turn enum 压缩 grep | repo grep gate（`backend_practice_legacy.py`） | Red: 旧 mapping 函数残留；Green: `compressTurnStatus` / `mapInternalToWire` / 类似命名零出现 |
| openapi fixtures 扩展（events + complete 命名场景） | `make validate-fixtures`（或 `python3 scripts/lint/validate_fixtures.py --repo-root .`）+ contract test | Red: 命名场景缺失 / fixture 不匹配 generated schema；Green: 13 个命名场景齐全 + Prism / contract 测试 PASS |
| 各 owner spec header / history append | `/sync-doc-index --check` 或 `python3 .agent-skills/sync-doc-index/scripts/check_index.py` | Red: B2 / B3 Header 或 history 缺失 002 记录；Green: 两个 spec Header 与 history 一致 |
| F3 baseline preflight | F3 baseline 状态读取断言 + `RegistryClient.Resolve` mock test | Red: F3 baseline 非 `completed` / Resolve 不返回 valid 三元组；Green: 三个 feature_key Resolve 通过 |

## 4 Phase 1: AppendSessionEvent state machine 与 turn-status 域

| 任务 | 测试文件 / 命令 | 预期 Red/Green 证据 |
|------|----------------|---------------------|
| SessionEventService skeleton + types | `backend/internal/practice/session_event_test.go` | Red: 类型 / 接口不存在；Green: 5 kind exhaustive switch、AssistantAction shape、provenance 默认值断言 PASS |
| handleAnswerSubmitted 3 分支 | 同上（`TestHandleAnswerSubmitted_*`） | Red: 任一 `ask_question` / `ask_follow_up` / `session_completed` 决策错误，或客户端 `payload.followUpCount` 可影响状态机；Green: budget / DB `follow_up_count` 边界值覆盖 PASS，首次 follow-up 不发 `practice.turn.completed`，turn assessed 后一次/turn 发 outbox |
| handleHintRequested 默认 strict 409（D-34） | 同上（`TestHandleHintRequested_DefaultsToStrictConflict`） | Red: 002 返回 200 / show_hint / 调用 AI；Green: 所有 mode / goal 输入返回 409 + `detail.policy='hint_disabled_in_mode'` |
| handleTurnSkipped / handleSessionPaused / handleSessionResumed | 同上 | Red: status 推进错误 / AssistantAction 错；Green: turn / session status 推进与 AssistantAction shape 断言 PASS |
| unknown kind → VALIDATION_FAILED | 同上 | Red: panic / 500；Green: outcome 形式正确，handler 层映射到 422 |
| AssistantAction provenance 默认值 | `session_event_test.go` 子集 | Red: 出现 runtime-only 字段（feature_key / cost / latency / provider）；Green: 仅含 B2 wire 字段 |
| turn-status mapping（D-33） | `backend/internal/practice/turn_status_test.go` | Red: 出现 D-25 备选的"压缩到 3 值"映射 / unknown 输入返回默认；Green: 5 值往返 + unknown 返回 error |

## 5 Phase 2: AppendSessionEvent vertical slice

| 任务 | 测试文件 / 命令 | 预期 Red/Green 证据 |
|------|----------------|---------------------|
| Repository AppendSessionEvent 主流程 | `backend/internal/store/practice/append_complete_test.go` | Red: tx 分裂 / outbox 缺失 / append 写入 audit；Green: 单事务写 event / turn / session / outbox，不写 `audit_events`；`SELECT FOR UPDATE` 起效 |
| Repository stale-turn / seq_no boundary | 同上 + cmd/api `TestE2EP0040PracticeEventConcurrentSeqNoStaleTurnConflict` | Red: stale turn 被接受或 seq_no 重号 / 丢序；Green: stale-turn conflict 返回 409，已接受事件 seq_no 单调连续；UNIQUE(session_id, seq_no) 约束生效 |
| Repository replay（同 clientEventId 同 fingerprint） | 同上（`TestAppendSessionEvent_Replay`） | Red: 第二次写入新 event row；Green: 第二次返回首次结果，DB 行数不变 |
| Repository mismatch（同 clientEventId 不同 fingerprint） | 同上（`TestAppendSessionEvent_FingerprintMismatch`） | Red: 接受第二次 payload；Green: 返回 `ErrClientEventFingerprintMismatch`，envelope 不泄露首次 payload |
| Repository cross-user 404 | 同上（`TestAppendSessionEvent_CrossUser`） | Red: 用户 B 命中用户 A 的 session；Green: `ErrSessionNotFound` 触发 404 |
| outbox_emitter practice.turn.completed | `backend/internal/store/practice/outbox_emitter_test.go` (`TestBuildPracticeTurnCompletedPayload`) | Red: payload 与 B3 schema 不一致 / 含明文；Green: 与 `shared/events/practice.turn.completed.*` 一致 + piiBoundary 通过 |
| Service AppendSessionEvent（含 F3 follow_up） | `backend/internal/practice/append_session_event_service_test.go` | Red: AI 调用在事务内 / AI 失败阻塞用户；Green: AI 在事务外、失败时退化到 `ask_question` placeholder 并写 `failure_code` 到 outcome / structured log 摘要（不写 append audit） |
| `answer_submitted` 缺失 answerText | `backend/internal/practice/append_session_event_service_test.go` (`TestAppendSessionEventRejectsMissingAnswerText`) | Red: 缺失 / 空白 `payload.answerText` 进入 Route / AI / append；Green: 返回 `VALIDATION_FAILED` + `field='payload.answerText'`，只执行 reservation，不写 event |
| `answer_submitted` server-owned follow-up state | `backend/internal/practice/session_event_test.go` + `backend/cmd/api/practice_http_scenario_test.go` | Red: 客户端 `payload.followUpCount` 可跳过 follow-up，或首次 follow-up 提前写 `practice.turn.completed`；Green: Route 只读 latest turn 的 DB `FollowUpCount`，HTTP scenario 覆盖 `ask_follow_up` → `ask_question` → `session_completed` 当前顺序 |
| AppendSessionEvent required `occurredAt` | `backend/internal/api/practice/session_event_handlers_test.go` (`TestAppendSessionEventRequiresOccurredAt`) | Red: 缺失 `occurredAt` 被 handler 接受并由 service 默认 server time；Green: 缺失返回 422 `VALIDATION_FAILED` + `field='occurredAt'`，service 不被调用 |
| Handler AppendSessionEvent | `backend/internal/api/practice/session_event_handlers_test.go` | Red: 接受 `Idempotency-Key` header / 错误码映射错；Green: 拒绝 header → 400；正常返回 200 + B2 wire shape |
| Error mapping | handler/service error mapping tests | Red: 新增映射缺失；Green: clientEvent mismatch / idempotency header policy 映射 PASS |
| Router 注册 | router test | Red: 路径未挂接 / 错误地挂上 idempotency middleware；Green: 路径命中 + 无 idempotency wrapper |

## 6 Phase 3: CompletePracticeSession vertical slice

| 任务 | 测试文件 / 命令 | 预期 Red/Green 证据 |
|------|----------------|---------------------|
| Repository CompleteSession 主流程 | `backend/internal/store/practice/append_complete_test.go` | Red: 单事务漏写表 / outbox 缺失 / async_jobs 缺失；Green: 单事务写 session / event / feedback_reports / async_jobs / outbox / audit |
| Repository D-35 replay 路径 | 同上（`TestCompleteSession_ReplayReturnsExistingReport`） | Red: 第二次创建新 feedback_reports / async_jobs，或 replay 查询不绑定 `async_jobs.dedupe_key=sessionId` / `resource_type='feedback_report'`；Green: 反向查到既有 report + matching job + outbox dedupe，返回 Replay=true |
| Repository status guard | 同上（`TestSQLRepositoryCompleteSessionRejectsIllegalStatusWithoutReport` / `TestSQLRepositoryCompleteSessionReplaysExistingReportBeforeStatusGuard` / `TestCanCompletePracticeSessionStatusAllowsRunningWaitingAndCompleted`） | Red: 无既有 report/job 时 `failed` / `queued` / `completing` / `cancelled` 也创建 queued report/job，或 D-35 既有 report/job replay 被状态 guard 阻断；Green: D-35 replay 优先，缺失 report/job 时只允许 `running` / `waiting_user_input` / `completed` |
| Repository concurrent single-executor | 同上（`TestCompleteSession_Concurrent`） | Red: 并发产生多份 feedback_reports；Green: UNIQUE(session_id) 约束 + row lock 序列化 |
| Repository cross-user 404 | 同上（`TestCompleteSession_CrossUser`） | Red: 用户 B 命中用户 A；Green: 404 |
| outbox_emitter practice.session.completed | `outbox_emitter_test.go` | Red: payload 与 B3 schema 不一致；Green: 与 `shared/events/practice.session.completed.*` 一致 |
| async_jobs dedupe_key | `append_complete_test.go` + cmd/api D-35 scenario | Red: dedupe_key 为空 / 不写 INDEX；Green: dedupe_key=sessionId 写入；UNIQUE INDEX / D-35 replay 阻止二次创建 |
| Service CompleteSession (D-35) | `backend/internal/practice/complete_session_service_test.go` | Red: replay 路径返回新 reportId；Green: replay 返回原始 reportId / jobId / response_body |
| Handler CompletePracticeSession + idempotency middleware | `backend/internal/api/practice/session_event_handlers_test.go` | Red: middleware 未挂接 / 双 key 行为错；Green: 正常返回 202 + ReportWithJob；replay 返回首次 response；mismatch 返回 409；cross-user 返回 404；D-35 双 key 走 service replay 路径 |
| CompletePracticeSession required `clientCompletedAt` | `backend/internal/api/practice/session_event_handlers_test.go` (`TestCompletePracticeSessionRequiresClientCompletedAt`) | Red: 缺失 `clientCompletedAt` 被 handler 接受并由 service 默认 server time；Green: 缺失返回 422 `VALIDATION_FAILED` + `field='clientCompletedAt'`，service 不被调用 |
| Idempotency middleware复用（complete 端） | `backend/internal/middleware/idempotency/middleware_test.go`（必要时扩展） | Red: complete 操作未走 middleware；Green: domain=`practice` + operation=`completePracticeSession` 在 middleware 中存档 |
| Error mapping | handler/service error mapping tests | Red: 缺映射；Green: session not found → 404 PRACTICE_SESSION_NOT_FOUND；middleware mismatch → 409 |

## 7 Phase 4: 隐私 / 观测 / Legacy-Negative

| 任务 | 测试文件 / 命令 | 预期 Red/Green 证据 |
|------|----------------|---------------------|
| Redaction 单元测试 | `backend/internal/store/practice/outbox_emitter_test.go` + cmd/api `TestE2EP0043PracticeEventLoopPrivacyAndLegacyNegativeSurface` | Red: outbox / complete audit / log 出现明文，或 append 写入 audit；Green: `question_text` / `answer_text` / `hint_text` / prompt / response / secret 在 fixture 输入下零出现，且 append 路径无 `audit_events` |
| Metric label allowlist | A3 observed AIClient existing allowlist + 002 service tests no new label surface | Red: metric label 含 `feature_key` / prompt-rubric version / provider raw model id；Green: 命中 F1 allowlist |
| Repo-wide legacy-negative grep | `python3 scripts/lint/backend_practice_legacy.py --repo-root .` | Red: `report_generate` 二次创建 / 旧 wire turn enum 压缩 / `warmup` 等 retired 术语出现；Green: 全部零出现 |
| Out-of-scope boundary 反查 | `session_event_test.go` (`TestHintNotImplementedInPhase2`) + repo grep | Red: 002 出现 `practice.turn.lightweight_observe` 调用 / assisted hint 实现 / derived plan goal 处理 / voice operation handler / CASCADE / sweep；Green: 全部零出现 |
| `make codegen-check` 收口 | `make codegen-check` | Red: generated artifacts 漂移；Green: 全 PASS |
| `cd backend && go test ./...` 收口 | go test | Red: 任一包失败；Green: 全 PASS |
| `python3 scripts/lint/conventions_drift.py --repo-root .` | drift | Red: conventions Go / TS drift；Green: PASS |
