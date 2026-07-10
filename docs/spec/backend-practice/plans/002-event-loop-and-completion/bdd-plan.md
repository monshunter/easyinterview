# Backend Practice Event Loop and Completion BDD Plan

> **版本**: 1.5
> **状态**: completed
> **更新日期**: 2026-07-10

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 场景范围

本 BDD plan 用 `backend/cmd/api/practice_http_scenario_test.go` 的 cmd/api HTTP scenario tests 覆盖 `appendSessionEvent` 与 `completePracticeSession` 的用户可见 API 行为。

- 套件: `e2e`
- 阶段: `P0`
- 场景编号: `E2E.P0.038` / `E2E.P0.039` / `E2E.P0.040` / `E2E.P0.041` / `E2E.P0.042` / `E2E.P0.043`
- 编号分配: `E2E.P0.038` / `E2E.P0.039` / `E2E.P0.040` / `E2E.P0.041` / `E2E.P0.042` / `E2E.P0.043`
- 执行入口: `cd backend && go test ./cmd/api -run 'TestE2EP0038|TestE2EP0039|TestE2EP0040|TestE2EP0041|TestE2EP0042|TestE2EP0043' -count=1`

## 2 场景矩阵

| 场景 ID | 名称 | 类别 | 覆盖 |
|---------|------|------|------|
| `E2E.P0.038` | append answer flow | primary | answer_submitted 三分支、server-owned follow-up state、turn status 4 值、turn-completed outbox |
| `E2E.P0.039` | append replay / mismatch / kind router / header policy | boundary + failure | `clientEventId` replay/mismatch、4 kind routing、strict-mode hint returns `show_hint`、append header refusal、cross-user 404 |
| `E2E.P0.040` | append sequencing boundary | boundary | stale-turn conflict, accepted event `seq_no` continuity, outbox count alignment |
| `E2E.P0.041` | completion handoff | primary | HTTP 202, queued report/job, session completion event, outbox, idempotency snapshot |
| `E2E.P0.042` | completion idempotency matrix | boundary + failure | same-key replay, mismatch, another-key completed-session replay, cross-user 404, illegal status conflict |
| `E2E.P0.043` | privacy and runtime boundary | privacy + regression | redaction, bounded AI metrics/task rows, source-event-only job boundary, 4-value turn status drift check |

## 3 Given / When / Then

| 场景 ID | Given | When | Then | 验证入口 |
|---------|-------|------|------|----------|
| `E2E.P0.038` | 已登录用户拥有 ready practice plan、running session、asked current turn 和 fake follow-up AI | 连续提交 `answer_submitted` 事件覆盖 follow-up、next-question 和 session-completed 分支 | HTTP 200；assistant action 类型正确；client payload 不驱动 `follow_up_count`；accepted events `seq_no` 连续；turn status 暴露当前 4 值；`practice.turn.completed` outbox 一次/turn；provenance 仅含 wire 字段 | `TestE2EP0038PracticeEventLoopAnswerFlow` |
| `E2E.P0.039` | 用户 A 拥有 running session；用户 B 独立；矩阵覆盖 replay、mismatch、kind、header 和 cross-user | 同 payload 重放、变更 payload 重放、提交 4 种当前 text kind、携带 `Idempotency-Key`、用户 B 访问用户 A session | replay 返回原结果且无重复 side effect；mismatch 返回 409；4 kind 均路由；strict-mode hint 返回 `show_hint`；header 返回 400；cross-user 返回 404 | `TestE2EP0039PracticeEventIdempotencyKindRouterAndHeaderPolicy` |
| `E2E.P0.040` | current turn 先进入 `follow_up_requested`；两个客户端持有同一 turn 上下文 | 一个请求推进 turn，另一个请求继续提交同一 turn | 一个请求成功，另一个 stale-turn conflict；错误 envelope 不泄露另一请求 payload；accepted events `seq_no` 连续；turn-completed outbox 行数与 completed turn 对齐 | `TestE2EP0040PracticeEventConcurrentSeqNoStaleTurnConflict` |
| `E2E.P0.041` | 用户 A 的 running session 已满足完成条件 | 调用 `POST /practice/sessions/{sessionId}/complete` 并携带 `Idempotency-Key` | HTTP 202 + `ReportWithJob`；session=`completing`；queued `feedback_reports` 和 `async_jobs(report_generate)` 各一行；`practice.session.completed` outbox 一行；idempotency snapshot 与响应一致 | `TestE2EP0041PracticeSessionCompleteCreatesQueuedReportJob` |
| `E2E.P0.042` | 已存在 completion 结果 R1/J1；用户 B 独立；另有 failed session | 同 key replay、同 key mismatch、另一 key complete 已完成 session、cross-user complete/get、failed session complete | replay 返回 R1/J1；mismatch 409；另一 key 写新的 idempotency snapshot 但复用 R1/J1；cross-user 404；failed session conflict；report/job/outbox 不重复 | `TestE2EP0042PracticeSessionCompleteIdempotencyMatrix` |
| `E2E.P0.043` | P0.038-P0.042 的 API 行为已可执行；日志、metric、audit、outbox 和 task-run payload 可检查 | 触发 follow-up AI、completion handoff，并运行 runtime boundary lint / generated drift gate | question/answer/hint/prompt/response/secret 文本零泄露；append 不写 audit；source-event-only `report_generate` boundary 成立；turn status 4 值无 drift；runtime boundary lint 通过 | `TestE2EP0043PracticeEventLoopPrivacyAndOutOfScopeSurface` |

## 4 与单元测试边界

BDD 只验证用户可见 API 行为、持久化副作用和跨层红线。纯函数状态机、repository row-lock detail、middleware reserve state、fixture schema parity、event/job codegen 和 generated type drift 由 [test-plan](./test-plan.md) 中的 focused gates 覆盖。
