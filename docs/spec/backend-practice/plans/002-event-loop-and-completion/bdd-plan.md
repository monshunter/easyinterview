# 002 — Event Loop and Completion BDD Plan

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-05-14

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 0 BDD 框架与编号

本 plan 的 6 个 BDD 场景保留 `E2E.P0.xxx` 编号与 Given / When / Then 语义；002 本次代码交付的可执行入口落在 `backend/cmd/api/practice_http_scenario_test.go` 的 HTTP scenario tests，用 `cmd/api` 真实路由、middleware、service/store fake 与 response/side-effect 断言覆盖用户可见 API 行为。

- 套件: `e2e`
- 阶段: `P0`
- 已占用编号现状（[`test/scenarios/e2e/INDEX.md`](../../../../../test/scenarios/e2e/INDEX.md)）：`001-006`, `010-037`；本 plan 在空闲号段 `038-043` 中分配 6 个场景（`034-035` 已由 backend-resume 占用）
- 编号分配: `E2E.P0.038` / `E2E.P0.039` / `E2E.P0.040` / `E2E.P0.041` / `E2E.P0.042` / `E2E.P0.043`
- 执行入口: `cd backend && go test ./cmd/api -run 'TestE2EP0038|TestE2EP0039|TestE2EP0040|TestE2EP0041|TestE2EP0042|TestE2EP0043' -count=1`
- 外部 Kind / shell 场景资产: 002 不新增 `test/scenarios/e2e/p0-NNN-*` 目录；若未来需要 Kind live BDD，可由 scenarios owner 按当前编号把同一 Given / When / Then 提升为 `test/scenarios` 资产，不改变本 plan 的 API 行为语义

每个场景的执行证据在 [bdd-checklist](./bdd-checklist.md) 跟踪；本文件只记录场景的 Given / When / Then 与覆盖范围，不出现执行 checkbox。

## 1 场景矩阵

| 场景 ID | 名称 | 类别 | 关联 Plan Phase | 关联 spec AC / D |
|---------|------|------|----------------|-------------------|
| `E2E.P0.038` | appendSessionEvent answer_submitted 主路径 + AssistantAction 三分支 | primary | Phase 2 | C-6, C-13（append 子集）, D-25 / D-33 wire status |
| `E2E.P0.039` | appendSessionEvent clientEventId replay / mismatch + 5 kind exhaustive + hint 默认 strict + Idempotency-Key header 拒收 | boundary + failure + alternate + regression | Phase 2 | C-9, C-26, spec §2.1 5 kind router, D-7, D-34 |
| `E2E.P0.040` | appendSessionEvent 并发 seq_no 序列化 | boundary | Phase 2 | C-14 |
| `E2E.P0.041` | completePracticeSession 202 + ReportWithJob + queued report / job + outbox emit | primary | Phase 3 | C-11, D-6, D-22, D-28 / D-32 |
| `E2E.P0.042` | completePracticeSession idempotency 矩阵 + D-35 双 key replay + cross-user 404 + concurrent single-executor | boundary + failure + alternate | Phase 3 | C-21（complete 子集）, C-22（complete 子集）, C-23（complete 子集）, C-24（complete 子集）, C-27, D-35 |
| `E2E.P0.043` | 隐私红线 + AI metric label 边界 + D-32 forward-binding 反查 + D-33 wire enum drift 反查 + legacy-negative grep | privacy + observability + regression | Phase 4 | C-16（002 子集）, D-11, D-20, D-32, D-33, D-34 |

## 2 Phase 2 — AppendSessionEvent 主路径与状态机

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| `E2E.P0.038` | appendSessionEvent answer_submitted 主路径 + AssistantAction 三分支 | 用户 A 已通过 `E2E.P0.022` / `E2E.P0.023` 同款 setup 拥有 `practice_plans(goal='baseline', status='ready')` + `practice_sessions(status='running', currentTurn.turnIndex=1, status='asked')`；F3 active；A3 fake AIClient 配置可返回 follow_up 与 next-question | 用户 A 连续 `POST /practice/sessions/{sessionId}/events` 三次：① 首次 `kind='answer_submitted'` 即使携带恶意 `payload.followUpCount` 也由 DB `follow_up_count=0` 触发 `ask_follow_up`（同 turn 加深）；② 同 turn 的追问答案触发 `ask_question`（进入下一题）；③ 当 `turn_count` 达到 `question_budget` 后再 `kind='answer_submitted'` 触发 `session_completed` | 每次返回 `200 + SessionEventResult{acknowledged:true, session{status:'running'│'completed'}, assistantAction.type IN ('ask_question','ask_follow_up','session_completed')}`；DB `practice_session_events` 行按 seq_no 严格递增；`practice_turns.status` 推进按 DB 5 值 `asked → follow_up_requested → assessed`；`practice.turn.completed` outbox 仅在 turn assessed 后出现，且一次/turn；wire `PracticeTurn.status` 暴露完整 5 值（D-33）；AssistantAction `provenance` 仅含 B2 wire 字段 | `backend/cmd/api/practice_http_scenario_test.go::TestE2EP0038PracticeEventLoopAnswerFlow` |

## 3 Phase 2 — AppendSessionEvent 错误与边界

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| `E2E.P0.039` | appendSessionEvent clientEventId replay / mismatch + 5 kind exhaustive + hint 默认 strict + Idempotency-Key header 拒收 | 用户 A 拥有 `practice_sessions(status='running')`；session.mode 任意（assisted 或 strict）；session.goal 任意；用户 B 独立账号；测试矩阵覆盖 replay / mismatch / header reject / cross-user 与 5 kind router | ① 同 user + 同 `clientEventId` + 同 payload `session_resumed` 重试 → replay；② 同 `clientEventId` + 不同 payload → mismatch；③ `kind='hint_requested'` 任意 mode → 默认 409（D-34）；④ 分别发送 `answer_submitted` / `turn_skipped` / `session_paused` / `session_resumed` / `hint_requested` → 每种 kind 都被识别且返回合规 200 / 409；⑤ 携带 `Idempotency-Key` header → 400 + `detail.policy='use_client_event_id'`；⑥ 用户 B 同 `clientEventId` 同 `sessionId` 访问 → 404 + `PRACTICE_SESSION_NOT_FOUND` | ① replay 返回首次结果，DB / outbox 不重复，append 不写 audit，AI 不再调用；② 409 + envelope 不泄露首次 payload；③ 409 + `detail.policy='hint_disabled_in_mode'` + `detail.mode`；④ 5 kind 全部路由：`answer_submitted → ask_follow_up`、`turn_skipped → ask_question`、`session_paused → session_wait + practice_sessions.status='waiting_user_input'`、`session_resumed → session_wait + practice_sessions.status='running'`、`hint_requested → 409`；⑤ 400；⑥ 404 不泄露存在性 | `backend/cmd/api/practice_http_scenario_test.go::TestE2EP0039PracticeEventIdempotencyKindRouterAndHeaderPolicy` |
| `E2E.P0.040` | appendSessionEvent 并发 seq_no 序列化 | 用户 A 拥有 `practice_sessions(status='running')` 与 currentTurn；先用一次 `answer_submitted` 让当前 turn 进入 `follow_up_requested`；两个客户端请求均携带不同 `clientEventId`，但都针对同一个旧 currentTurn | 一个客户端提交追问答案并推进到下一题，另一个客户端继续用旧 turnId 提交 `POST /practice/sessions/{sessionId}/events kind='answer_submitted'` | 服务端通过 current-turn 校验与 repository 序列化约束：推进请求被接受并返回 200 + `SessionEventResult.acknowledged:true`，旧 turn 请求返回 409 `PRACTICE_SESSION_CONFLICT` 或等价 stale-turn conflict，且 envelope 不泄露另一请求 payload；DB `practice_session_events.seq_no` 对已接受事件严格连续无重号，UNIQUE(session_id, seq_no) 零违反；完成的 turn 数与 `outbox_events WHERE event_name='practice.turn.completed' AND aggregate_id=sessionId` 行数 1:1。N≥10 全成功的 seq_no 压力验证留在 repository integration test，不作为 BDD 用户流 gate | `backend/cmd/api/practice_http_scenario_test.go::TestE2EP0040PracticeEventConcurrentSeqNoStaleTurnConflict` |

## 4 Phase 3 — CompletePracticeSession 主路径

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| `E2E.P0.041` | completePracticeSession 202 + ReportWithJob + queued report / job + outbox emit | 用户 A 拥有 `practice_sessions(status='running', turn_count=question_budget)`；session 已有 N 个完成 turn；002 阶段无 runtime outbox→asynq dispatcher（dispatch-time 跳过 binding 归 future `backend-async-runner` plan） | 用户 A `POST /practice/sessions/{sessionId}/complete` 携带 `Idempotency-Key` | `202 + ReportWithJob{reportId, job{jobType:'report_generate', status:'queued'}}`，wire `Job` 不暴露 `dedupeKey`；DB `practice_sessions(status='completing', updated_at>=started_at)`；`feedback_reports(status='queued', user_id, session_id, language, target_job_id)` placeholder 行存在且唯一（UNIQUE INDEX `idx_feedback_reports_session_unique` 兜底）；`async_jobs(job_type='report_generate', status='queued', dedupe_key=sessionId)` 行存在且唯一（UNIQUE INDEX `idx_async_jobs_active_dedupe` 兜底）；`outbox_events WHERE event_name='practice.session.completed' AND aggregate_id=sessionId` 行数为 1（应用层 D-35 Replay 路径保证，DB 当前无 outbox 唯一约束），payload 与 B3 `shared/events/practice.session.completed.*` schema 一致；`practice_session_events(event_type='session_completed', seq_no=MAX+1)` 行存在；`idempotency_records(domain='practice', operation='completePracticeSession', status='succeeded', resource_id=reportId)` 行存在且 `response_body` snapshot 与 HTTP 响应一致；handler 二次进入（不同 Idempotency-Key 或同 key replay）均不触发新 INSERT — 由 D-35 + `async_jobs` UNIQUE INDEX + idempotency middleware 共同保证 | `backend/cmd/api/practice_http_scenario_test.go::TestE2EP0041PracticeSessionCompleteCreatesQueuedReportJob` |

## 5 Phase 3 — CompletePracticeSession idempotency 矩阵

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| `E2E.P0.042` | completePracticeSession idempotency 矩阵 + D-35 双 key replay + cross-user 404 + concurrent single-executor | 用户 A 已通过 `E2E.P0.041` setup 完成 `feedback_reports(reportId=R1, status='queued')` + `async_jobs(jobId=J1)`；用户 B 独立账号；另有一个没有既有 report/job 的 failed session | 多组请求：① 同 user + 同 `Idempotency-Key=K1` + 同 fingerprint replay；② 同 user + 同 `Idempotency-Key=K1` + 不同 fingerprint mismatch；③ 同 user + 不同 `Idempotency-Key=K2` 重新 complete（已完成 session）→ D-35；④ 用户 B 携带 `Idempotency-Key=K1` complete `sessionId=A` → 跨用户；⑤ 同 user + 同 key + 同 fingerprint 并发；⑥ 用户 B `GET /practice/sessions/{sessionId}` 取用户 A session；⑦ complete failed session | ① replay 返回首次 `202 + ReportWithJob{reportId=R1}`，DB 不重复；② 409 + `PRACTICE_SESSION_CONFLICT`，envelope 不泄露 R1；③ D-35：返回相同 `ReportWithJob{reportId=R1, jobId=J1}`（HTTP 202），新 `idempotency_records(idempotency_key='K2')` 行写入且 `resource_id=R1` 与 K1 一致，`feedback_reports` / `async_jobs` / `outbox_events practice.session.completed` 行数仍为 1；④ 用户 B 走自己 idempotency record，complete 用户 B 自己的（不存在的）session 时返回 404 + `PRACTICE_SESSION_NOT_FOUND`，不能命中用户 A R1；⑤ 单执行者：UNIQUE + row lock 仅一个执行者成功，另一个 replay 同一 response；⑥ 用户 B 拿到 404；⑦ 返回 409 `PRACTICE_SESSION_CONFLICT` 且不创建 report/job/outbox | `backend/cmd/api/practice_http_scenario_test.go::TestE2EP0042PracticeSessionCompleteIdempotencyMatrix` |

## 6 Phase 4 — 观测 / 隐私 / 回归

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| `E2E.P0.043` | 隐私红线 + AI metric label 边界 + D-32 forward-binding 反查 + D-33 wire enum drift 反查 + legacy-negative grep | `E2E.P0.038` / `E2E.P0.039` / `E2E.P0.040` / `E2E.P0.041` / `E2E.P0.042` 已运行；log / metric / complete audit / outbox 收集器可读 | 触发额外 AI follow_up 与 hint_requested（默认 strict）调用；采集本 phase 的 log / metric label / complete `audit_events` / `outbox_events` / `ai_task_runs` payload；运行 repo-wide grep + lint | 全集合中 `question_text` / `answer_text` / `hint_text` / AI prompt body / response body / provider secret 零出现；append 路径不写 `audit_events`；`ai_task_runs` 行包含 spec §4 typed columns；A3 `ai_task_*` metric label 命中 F1 allowlist 且不含 `feature_key` / prompt-rubric version / provider raw model id；D-32 验证集：`make lint-events` 与 `make codegen-events-check` 通过 + `report_generate` 在 jobs.yaml 标注 `triggerEventSemantic: source_event_only` + generated `JobTriggerEventSemantic*` 常量与 `IsSourceEventOnly("report_generate")=true` 谓词单元测试通过 + repo grep 断言 `report_generate` INSERT 仅出现在 `backend/internal/store/practice/complete_session.go`（runtime outbox→asynq dispatcher 集成 gate 由 future `backend-async-runner` plan 承接）；OpenAPI `PracticeTurn.status` enum 含 5 值 + generated TS / Go const 无 drift（D-33）；repo grep 中 `warmup` / `single_drill` / `drill_builder` / `mistake_queue` / `growth_center` / 独立 `voice` route / `practiceModeCard` / `compressTurnStatus` / `mapInternalToWire` / `legacyTurnStatusMapping` 在 002 输出文档与代码 diff 零出现 | `backend/cmd/api/practice_http_scenario_test.go::TestE2EP0043PracticeEventLoopPrivacyAndLegacyNegativeSurface` |

## 7 数据隔离与污染恢复

每个场景按 `test/scenarios/e2e/README.md` §5 / §3 / §6 / §8 约定：

- 数据隔离：每个场景使用独立的 `user_id` / `plan_id` / `session_id` / `clientEventId` / `idempotency_key` 命名空间；不复用其它场景（含 `E2E.P0.022 ~ E2E.P0.026`）的 fixture
- 清理顺序：cleanup 先删自身 `practice_session_events` → `practice_turns` → `feedback_reports`（顺序在 `async_jobs` 之前，避免 FK 冲突）→ `async_jobs` → `practice_sessions` → `practice_plans` → `idempotency_records` → `audit_events` → `outbox_events` → users
- 污染恢复：场景失败时按 README §8 顺序：① 清理场景自身资源；② 定位并恢复 shared 组件（F3 cache、idempotency_records pending；002 阶段无 runtime dispatcher，无需恢复 dispatcher 状态）；③ 仅在 ① ② 失败时 `test/scenarios/env-cleanup.sh && env-setup.sh` 全量重建
- 不预设 Helm chart / 外部 Git 平台名称；所有命令以本仓库脚本为真理源

## 8 与单元测试边界

本 BDD plan 验证用户可见行为切片（HTTP API + DB 状态 + outbox emit + log/metric/complete audit 红线 + append 无 audit + generated artifact drift + jobs.yaml `triggerEventSemantic` lint + `IsSourceEventOnly` 谓词 + `report_generate` INSERT 位点 grep）；不重复内部接口签名、序列化结构、错误映射、状态机内部分支等单元测试覆盖（见 [test-plan](./test-plan.md)）。002 阶段不存在 runtime outbox→asynq dispatcher，"dispatcher 集成测试" 不在本 BDD plan 范围内，由 future `backend-async-runner` plan 承接。
