# 002 — Event Loop and Completion BDD Checklist

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-05-14

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 可执行 BDD 证据

002 本次交付使用 `backend/cmd/api/practice_http_scenario_test.go` 中的 `cmd/api` HTTP scenario tests 承接 `E2E.P0.038` / `E2E.P0.039` / `E2E.P0.040` / `E2E.P0.041` / `E2E.P0.042` / `E2E.P0.043`。这些测试以真实 route / middleware / handler / service 边界验证用户可见 API 行为；本 plan 不新增 `test/scenarios/e2e/p0-NNN-*` shell 目录。

## E2E.P0.038 appendSessionEvent answer_submitted 主路径 + AssistantAction 三分支

- [x] 在 `TestE2EP0038PracticeEventLoopAnswerFlow` 中准备独立 user / target_job / resume_asset / plan / running session 与 fake follow-up AI。
- [x] 连续触发 `answer_submitted`，覆盖 DB-owned `follow_up_count=0` → `ask_follow_up`、追问答案 → `ask_question`、budget 边界 → `session_completed` 三分支。
- [x] 断言客户端 `payload.followUpCount` 不驱动状态机，HTTP 200、`SessionEventResult` shape、seq_no 连续、turn status 5 值推进、`practice.turn.completed` outbox 一次/turn 与 provenance wire 字段。
- [x] 执行通过：`cd backend && go test ./cmd/api -run TestE2EP0038PracticeEventLoopAnswerFlow -count=1`（也随 `go test ./cmd/api` 与 `go test ./...` 通过）。

## E2E.P0.039 appendSessionEvent clientEventId replay / mismatch + 5 kind exhaustive + hint 默认 strict + Idempotency-Key header 拒收

- [x] 在 `TestE2EP0039PracticeEventIdempotencyKindRouterAndHeaderPolicy` 中准备用户 A / 用户 B、5 kind（含 `answer_submitted` 成功分支）、replay / mismatch / header reject / cross-user 矩阵。
- [x] 断言 replay 不重复写 event / outbox / audit，mismatch 返回 409 且不泄露首次 payload。
- [x] 断言 hint 在 002 默认返回 409 + `detail.policy='hint_disabled_in_mode'`，5 kind 路由与 pause/resume 状态切换正确。
- [x] 断言携带 `Idempotency-Key` 返回 400 + `detail.policy='use_client_event_id'`，跨用户访问返回 404。
- [x] 执行通过：`cd backend && go test ./cmd/api -run TestE2EP0039PracticeEventIdempotencyKindRouterAndHeaderPolicy -count=1`（也随 `go test ./cmd/api` 与 `go test ./...` 通过）。

## E2E.P0.040 appendSessionEvent 并发 seq_no 序列化

- [x] 在 `TestE2EP0040PracticeEventConcurrentSeqNoStaleTurnConflict` 中先让 currentTurn 进入 `follow_up_requested`，再准备同一旧 turn 的双客户端竞争输入。
- [x] 断言一个请求成功，另一个 stale-turn / conflict 返回 409，错误 envelope 不泄露另一请求 payload。
- [x] 断言已接受事件的 seq_no 连续，turn / outbox 行数保持 1:1；row-lock / UNIQUE 约束由 store 层测试与 DB schema 兜底。
- [x] 执行通过：`cd backend && go test ./cmd/api -run TestE2EP0040PracticeEventConcurrentSeqNoStaleTurnConflict -count=1`（也随 `go test ./cmd/api` 与 `go test ./...` 通过）。

## E2E.P0.041 completePracticeSession 202 + ReportWithJob + queued report / job + outbox emit

- [x] 在 `TestE2EP0041PracticeSessionCompleteCreatesQueuedReportJob` 中准备用户 A + running session + turn budget 满足条件。
- [x] 触发 `POST /practice/sessions/{sessionId}/complete` 并断言 HTTP 202 + `ReportWithJob{jobType:'report_generate', status:'queued'}`。
- [x] 断言 session=`completing`、`feedback_reports` placeholder、`async_jobs(report_generate, dedupe_key=sessionId)`、`practice.session.completed` outbox、`session_completed` event 与 idempotency snapshot。
- [x] 执行通过：`cd backend && go test ./cmd/api -run TestE2EP0041PracticeSessionCompleteCreatesQueuedReportJob -count=1`（也随 `go test ./cmd/api` 与 `go test ./...` 通过）。

## E2E.P0.042 completePracticeSession idempotency 矩阵 + D-35 双 key replay + cross-user 404

- [x] 在 `TestE2EP0042PracticeSessionCompleteIdempotencyMatrix` 中准备首次 complete 结果 R1 / J1 与第二个用户。
- [x] 断言同 key replay 返回 R1 / J1 且 DB 不变；同 key 不同 fingerprint 返回 409 且不泄露 R1。
- [x] 断言不同 `Idempotency-Key` 的已完成 session complete 走 D-35 replay，新增 idempotency record 指向同一 report/job，不重复创建 feedback report / async job / outbox。
- [x] 断言跨用户 complete / get 返回 404，不能命中用户 A 资源。
- [x] 断言没有既有 report/job 的 failed session complete 返回 409 `PRACTICE_SESSION_CONFLICT`，不创建新的 feedback report / async job / outbox。
- [x] 执行通过：`cd backend && go test ./cmd/api -run TestE2EP0042PracticeSessionCompleteIdempotencyMatrix -count=1`（也随 `go test ./cmd/api` 与 `go test ./...` 通过）。

## E2E.P0.043 隐私红线 + AI metric label 边界 + D-32 / D-33 / legacy-negative 反查

- [x] 在 `TestE2EP0043PracticeEventLoopPrivacyAndLegacyNegativeSurface` 中触发 follow-up AI 与 strict hint conflict。
- [x] 断言 outbox / audit / ai task payload 不含 `question_text` / `answer_text` / `hint_text` / prompt / response / secret，append 路径不写 `audit_events`。
- [x] 断言 D-32 source-event-only forward-binding、D-33 5 值 turn status、legacy-negative surface 与 `report_generate` 二次创建路径不回归。
- [x] 执行通过：`cd backend && go test ./cmd/api -run TestE2EP0043PracticeEventLoopPrivacyAndLegacyNegativeSurface -count=1`（也随 `go test ./cmd/api` 与 `go test ./...` 通过）。

## 收口命令

- [x] `cd backend && go test ./internal/api/practice ./internal/practice ./internal/store/practice ./internal/middleware/idempotency ./cmd/api -count=1`
- [x] `cd backend && go test ./...`
- [x] `make lint-events`
- [x] `make codegen-events-check`
- [x] `make codegen-check`
- [x] `make validate-fixtures`
- [x] `python3 scripts/lint/conventions_drift.py --repo-root .`
- [x] `python3 scripts/lint/backend_practice_legacy.py --repo-root .`
- [x] `make docs-check`
- [x] `git diff --check`
