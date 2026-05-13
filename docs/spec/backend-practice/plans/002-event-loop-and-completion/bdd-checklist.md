# 002 — Event Loop and Completion BDD Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-13

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.034 appendSessionEvent answer_submitted 主路径 + AssistantAction 三分支

- [ ] 创建场景目录 `test/scenarios/e2e/p0-034-practice-event-loop-answer-flow/`（含 `README.md` / `scripts/{setup,trigger,verify,cleanup}.sh` / `data/{seed-input.md,expected-outcome.md}` 骨架）
- [ ] 准备 seed-input：独立 user / target_job / resume_asset / plan(baseline) / session(running, turnIndex=1, status=asked) + F3 + A3 fake AIClient 接入命名 fixture `follow-up` 与 `default`
- [ ] 实现 setup.sh：创建用户与 session，断言 currentTurn 同步首题已就绪
- [ ] 实现 trigger.sh：连续 3 次 `POST /practice/sessions/{sessionId}/events kind=answer_submitted`（不同 `clientEventId`），覆盖 ask_question / ask_follow_up / session_completed 三分支
- [ ] 实现 verify.sh：断言每次 HTTP 200 + `SessionEventResult` 形状；DB `practice_session_events.seq_no` 严格连续；`practice_turns.status` DB 5 值推进；`outbox_events practice.turn.completed` 行数 == 完成的 turn 数；AssistantAction provenance 仅含 B2 wire 字段；wire `PracticeTurn.status` 暴露 5 值
- [ ] 实现 cleanup.sh：按 BDD plan §7 顺序删除自身资源
- [ ] 执行并通过场景验证；`result.json` 中 `validBddEvidence=true`
- [ ] 记录验证证据并将场景标记 `Ready` 同步到 `test/scenarios/e2e/INDEX.md`

## E2E.P0.035 appendSessionEvent clientEventId replay / mismatch + 5 kind exhaustive + hint 默认 strict + Idempotency-Key header 拒收

- [ ] 创建场景目录 `test/scenarios/e2e/p0-035-practice-event-idempotency-and-kind-router/`
- [ ] 准备 seed-input：用户 A + 用户 B + 5 kind 的 clientEventId 矩阵 + 命名 fixtures `replay` / `mismatch` / `hint-strict-conflict` / `turn-skipped` / `pause-resume`；assisted 与 strict 两个 mode 各跑一次 hint
- [ ] 实现 setup.sh：创建用户 A 与 session（mode=strict 与 assisted 两次），用户 B 独立账号
- [ ] 实现 trigger.sh：依次发送 6 组请求，覆盖 replay / mismatch / hint default / 5 kind / header reject / cross-user
- [ ] 实现 verify.sh：断言 ① replay 返回首次结果且 DB / outbox 不重复、append 不写 audit；② mismatch 409 不泄露首次 payload；③ hint 在 assisted 与 strict 下均返回 409 + `detail.policy='hint_disabled_in_mode'` + `detail.mode`；④ 5 kind 全部 200 / 409 路由正确，`practice_sessions.status` 在 paused / resumed 切换；⑤ 携带 Idempotency-Key → 400 + `detail.policy='use_client_event_id'`；⑥ 用户 B 同 sessionId → 404
- [ ] 实现 cleanup.sh
- [ ] 执行并通过场景验证
- [ ] 记录验证证据并同步 INDEX

## E2E.P0.038 appendSessionEvent 并发 seq_no 序列化

- [ ] 创建场景目录 `test/scenarios/e2e/p0-038-practice-event-concurrent-seq-no/`
- [ ] 准备 seed-input：用户 A + session + 2 个独立 clientEventId，两个请求均指向同一个 currentTurn
- [ ] 实现 setup.sh：创建 session 并保留 currentTurn 为 `asked`，确保两个客户端会竞争同一 turn
- [ ] 实现 trigger.sh：用 xargs / curl 或等价 repo-tracked 脚本近同时发送 2 个 `POST events kind='answer_submitted'`（每个 clientEventId 唯一）
- [ ] 实现 verify.sh：断言恰好一个请求 200 + acknowledged，另一个 409 `PRACTICE_SESSION_CONFLICT` 或等价 stale-turn conflict 且 envelope 不泄露另一请求 payload；DB 已接受事件的 `seq_no` 严格连续无重无丢；UNIQUE(session_id, seq_no) 零违反；`practice_turns` / `outbox_events` 行数 1:1。N≥10 全成功 seq_no 压力验证由 repository integration test 承接，不作为 BDD gate
- [ ] 实现 cleanup.sh
- [ ] 执行并通过场景验证
- [ ] 记录验证证据并同步 INDEX

## E2E.P0.039 completePracticeSession 202 + ReportWithJob + queued report / job + outbox emit

- [ ] 创建场景目录 `test/scenarios/e2e/p0-039-practice-session-complete-async-report/`
- [ ] 准备 seed-input：用户 A + session(turn_count=question_budget)；002 阶段无 runtime outbox→asynq dispatcher，scenario 仅依赖 handler-side INSERT + `async_jobs(job_type, dedupe_key)` UNIQUE INDEX + D-35 反向查 `feedback_reports` + repo grep 兜底
- [ ] 实现 setup.sh：通过 setup 推进 session 到 `running` + 满足 question_budget
- [ ] 实现 trigger.sh：`POST /practice/sessions/{sessionId}/complete` 携带 `Idempotency-Key`
- [ ] 实现 verify.sh：断言 HTTP 202 + `ReportWithJob{reportId, job{status:'queued', jobType:'report_generate'}}`，wire `Job` 不暴露 `dedupeKey`；DB `practice_sessions(status='completing', updated_at>=started_at)`、`feedback_reports` placeholder（UNIQUE INDEX `idx_feedback_reports_session_unique` 兜底）、`async_jobs(job_type='report_generate', dedupe_key=sessionId)` queued row（UNIQUE INDEX `idx_async_jobs_active_dedupe` 兜底）、`practice_session_events(event_type='session_completed', seq_no=MAX+1)`、`outbox_events WHERE event_name='practice.session.completed' AND aggregate_id=sessionId` 行数为 1、`idempotency_records(status='succeeded')` 各唯一；002 阶段无 runtime outbox→asynq dispatcher，验证 `report_generate` INSERT 不被二次创建依赖 (a) handler 二次进入不触发新 INSERT（D-35 反向查兜底）+ (b) `async_jobs` UNIQUE INDEX + (c) repo grep 断言 `report_generate` INSERT 仅出现在 `backend/internal/store/practice/complete_session.go`
- [ ] 实现 cleanup.sh
- [ ] 执行并通过场景验证
- [ ] 记录验证证据并同步 INDEX

## E2E.P0.040 completePracticeSession idempotency 矩阵 + D-35 双 key replay + cross-user 404 + concurrent single-executor

- [ ] 创建场景目录 `test/scenarios/e2e/p0-040-practice-session-complete-idempotency-matrix/`
- [ ] 准备 seed-input：用户 A 已完成 complete（reportId=R1, jobId=J1） + 用户 B + 6 组测试请求
- [ ] 实现 setup.sh：用户 A 完成第一次 complete（建立 R1 / J1 / outbox），保留 idempotency_key=K1 与 fingerprint snapshot
- [ ] 实现 trigger.sh：6 组：① 同 user K1 同 fingerprint replay；② 同 user K1 不同 fingerprint mismatch；③ 同 user K2 重新 complete（D-35）；④ 用户 B 携带 K1 complete sessionId=A；⑤ 同 user K1 同 fingerprint 并发；⑥ 用户 B GET sessionId=A
- [ ] 实现 verify.sh：断言 ① replay HTTP 202 + reportId=R1 + DB 不变；② 409 + envelope 不泄露 R1；③ D-35 HTTP 202 + reportId=R1 + jobId=J1 + 新 `idempotency_records(idempotency_key='K2', resource_id=R1)` 存在；④ 用户 B 走自己路径返回 404 PRACTICE_SESSION_NOT_FOUND；⑤ 仅一个执行者成功，另一个 replay 同 response；⑥ 用户 B 404；`feedback_reports` / `async_jobs` / `outbox_events practice.session.completed` 全程行数为 1
- [ ] 实现 cleanup.sh
- [ ] 执行并通过场景验证
- [ ] 记录验证证据并同步 INDEX

## E2E.P0.041 隐私红线 + AI metric label 边界 + D-32 forward-binding 反查 + D-33 wire enum drift 反查 + legacy-negative grep

- [ ] 创建场景目录 `test/scenarios/e2e/p0-041-practice-event-loop-privacy-and-legacy-negative/`
- [ ] 准备 seed-input：依赖 `E2E.P0.034` / `E2E.P0.035` / `E2E.P0.038` / `E2E.P0.039` / `E2E.P0.040` 已执行；准备 grep allowlist 与 negative list
- [ ] 实现 setup.sh：必要时通过 `scripts/lint/snapshot_metrics.sh` 或同等收集器抓取 metric / log / audit / outbox snapshot
- [ ] 实现 trigger.sh：触发额外 `answer_submitted`（命中 follow_up AI）与 `hint_requested`（默认 strict 409）调用；运行 `make lint-events` / `make codegen-events-check` / `python3 scripts/lint/conventions_drift.py --repo-root .` / repo grep
- [ ] 实现 verify.sh：断言 ① 全集合无 `question_text` / `answer_text` / `hint_text` / prompt / response / secret，且 append 路径不写 `audit_events`；② `ai_task_runs` 包含 spec §4 typed columns；③ metric label 命中 F1 allowlist；④ D-32 forward-binding 验证集：`make lint-events` 与 `make codegen-events-check` 通过 + `report_generate` 在 jobs.yaml 标注 `triggerEventSemantic: source_event_only` + `cd backend && go test ./internal/shared/jobs/...` 断言 `JobTriggerEventSemantic*` 常量与 `IsSourceEventOnly("report_generate")=true` 谓词 + repo grep 断言 `report_generate` INSERT 仅出现在 `backend/internal/store/practice/complete_session.go`（runtime outbox→asynq dispatcher 集成 gate 归 future `backend-async-runner` plan）；⑤ OpenAPI `PracticeTurn.status` 暴露 5 值，generated Go / TS const 无 drift；⑥ repo grep `warmup` / `single_drill` / `drill_builder` / `mistake_queue` / `growth_center` / 独立 `voice` route / `practiceModeCard` / `compressTurnStatus` / `mapInternalToWire` / `legacyTurnStatusMapping` 在 002 输出 diff 零出现
- [ ] 实现 cleanup.sh
- [ ] 执行并通过场景验证
- [ ] 记录验证证据并同步 INDEX
