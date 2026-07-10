# Backend Practice Event Loop and Completion

> **版本**: 1.8
> **状态**: completed
> **更新日期**: 2026-07-10

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

本 plan 承接 Practice API 的答题事件循环与会话完成合同：

- `POST /practice/sessions/{sessionId}/events` (`appendSessionEvent`) 通过 body `clientEventId` 做 per-session replay，拒绝 `Idempotency-Key` header，处理 `answer_submitted` / `hint_requested` / `session_paused` / `session_resumed` 四种文本 event kind；`turn_skipped` 不再是正向用户动作。
- `POST /practice/sessions/{sessionId}/complete` (`completePracticeSession`) 通过 shared idempotency middleware 返回 `202 + ReportWithJob`，同事务创建 queued `feedback_reports`、`async_jobs(job_type='report_generate')`、`practice.session.completed` outbox 和 session completion event。
- `PracticeTurn.status` wire enum 与 DB 4 值保持一致：`asked` / `answered` / `follow_up_requested` / `assessed`。
- `report_generate` job 由 completion handler 创建；`practice.session.completed` 是 source event / analytics fact，`shared/jobs.yaml` 用 `triggerEventSemantic: source_event_only` 固化该边界。
- 事件、outbox、audit、log、metric 和 task-run payload 不包含 question、answer、hint、prompt、response 或 provider secret 明文。

## 2 当前合同

### 2.1 Operation Matrix

| operationId | fixture / scenario | backend behavior | persistence | AI dependency | coverage |
|-------------|--------------------|------------------|-------------|---------------|----------|
| `appendSessionEvent` answer flow | `appendSessionEvent.json` answer / follow-up / pause / resume variants | returns `200 + SessionEventResult`; routes all 4 current text event kinds; rejects `Idempotency-Key` with `400 VALIDATION_FAILED` | `practice_session_events`, `practice_turns`, `practice_sessions`, `practice.turn.completed` outbox when a turn is assessed | F3 `practice.session.follow_up` only for follow-up branch | `E2E.P0.038`, `E2E.P0.039`, `E2E.P0.040`, unit/store tests |
| `appendSessionEvent` replay / mismatch | `appendSessionEvent.json` replay and mismatch variants | same `clientEventId` + same fingerprint returns original result; changed fingerprint returns 409 | no duplicate event/outbox/audit rows | no repeated AI call on replay | `E2E.P0.039`, repository tests |
| `appendSessionEvent` hint optional | `appendSessionEvent.json` `show-hint` / `hint-assisted-show` | strict and assisted sessions keep hint available; AI-backed `show_hint` behavior is owned by plan 003 | sanitized event response only | `practice.turn.lightweight_observe` via plan 003 | `E2E.P0.039`, mode tests |
| `completePracticeSession` create | `completePracticeSession.json` `default` | returns `202 + ReportWithJob{jobType:'report_generate', status:'queued'}` | `practice_sessions`, `practice_session_events`, `feedback_reports`, `async_jobs`, `outbox_events`, `audit_events`, `idempotency_records` | none | `E2E.P0.041`, store/handler tests |
| `completePracticeSession` replay / mismatch / cross-user | `completePracticeSession.json` replay, mismatch, session-already-completed, cross-user variants | same key replays; same key changed fingerprint returns 409; completed session with another key returns existing report/job; cross-user returns 404 | no duplicate report/job/outbox rows | none | `E2E.P0.042`, middleware/store tests |
| privacy and runtime boundary | no public fixture | source-event-only job boundary, 4-value turn status, no raw text leakage, no duplicate `report_generate` insert path | sanitized payloads and typed rows only | observed AI labels stay bounded | `E2E.P0.043`, lint and redaction tests |

### 2.2 Persistence Boundary

- Append events serialize per session with `SELECT FOR UPDATE` and monotonically increasing `seq_no`.
- Append replay is scoped by `(session_id, client_event_id)` and stored fingerprint.
- Completion creates exactly one active queued report/job pair per session and reuses it on later completion requests.
- `async_jobs.dedupe_key=sessionId` and D-35 service replay protect against duplicate report generation handoff.
- Append path does not write domain audit events; completion path writes audit metadata without question/answer text.

### 2.3 Wire Boundary

- `AssistantAction.provenance` exposes only B2 `GenerationProvenance` fields.
- `PracticeTurn.status` exposes all four current turn statuses.
- `Job` response does not expose internal `dedupe_key`.
- Error envelopes use B1/B2 codes and do not disclose another user's resources or a prior request body.

## 3 质量门禁

- **Plan 类型**: `feature-behavior + contract + code-internal`。
- **TDD 策略**: 适用。Focused tests cover state machine branches, replay/mismatch, row-lock sequencing, completion idempotency, source-event-only job semantics, generated event/job constants, OpenAPI generated types and privacy redaction.
- **BDD 策略**: 适用。`E2E.P0.038` - `E2E.P0.043` cover answer flow, event replay/mismatch, sequencing, completion handoff, completion idempotency and privacy/runtime boundary.
- **替代验证 gate**:
  - `cd backend && go test ./cmd/api -run 'TestE2EP0038|TestE2EP0039|TestE2EP0040|TestE2EP0041|TestE2EP0042|TestE2EP0043' -count=1`
  - `cd backend && go test ./internal/api/practice ./internal/practice ./internal/store/practice ./internal/middleware/idempotency ./internal/shared/jobs ./cmd/api -count=1`
  - `python3 scripts/lint/backend_practice_out_of_scope.py --repo-root . --phase all`
  - `python3 -m pytest scripts/lint/backend_practice_out_of_scope_test.py -q`
  - `make lint-events`
  - `make codegen-events-check`
  - `make validate-fixtures`
  - `make codegen-check`
  - `python3 scripts/lint/conventions_drift.py --repo-root .`
  - `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/backend-practice/plans/002-event-loop-and-completion/context.yaml --target backend`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `make docs-check`

## 4 实施步骤

### Phase 0: contract preflight

- Confirm `shared/jobs.yaml` marks `report_generate` as `source_event_only` and generated jobs constants expose `IsSourceEventOnly`.
- Confirm OpenAPI `PracticeTurn.status` contains the four current values and generated Go/TS artifacts are in sync.
- Confirm `appendSessionEvent` and `completePracticeSession` fixtures cover main, replay, mismatch and cross-user variants.
- Confirm F3 `practice.session.follow_up` resolves for follow-up generation.

### Phase 1: append event state machine

- Route all four current text event kinds.
- Generate `ask_question`, `ask_follow_up`, `session_wait` or `session_completed` actions from server-owned session/turn state.
- Preserve the current turn-status wire boundary.
- Reject malformed answer payloads before AI or persistence side effects.

### Phase 2: append event vertical slice

- Persist event, turn/session updates and outbox in one repository transaction.
- Enforce `clientEventId` replay/mismatch semantics.
- Refuse `Idempotency-Key` on append.
- Keep append audit-free and redact logs/events/metrics.

### Phase 3: completion vertical slice

- Wrap completion with shared idempotency middleware.
- Reuse existing report/job for completed sessions across idempotency keys.
- Create queued report/job/outbox/audit rows only once.
- Reject illegal session states when no existing report/job can be replayed.

### Phase 4: privacy, contract drift and closeout

- Enforce redaction and bounded metric labels.
- Run event/job/OpenAPI/generated drift gates.
- Run BDD scenario tests and backend-practice runtime boundary lint.
- Update plan/index evidence.

### Phase 5: Handler dead helper cleanup

- Delete the unreferenced `derefString` helper from `backend/internal/api/practice/handler.go`; retain the used `stringValue` normalizer.
- Keep all handler surfaces and request/response behavior unchanged.
- Use backend-wide `staticcheck` U1000 as the red signal; verify with scoped staticcheck and the backend-practice package gate.

### Phase 6: Turn status helper cleanup

- 保留生产状态机实际使用的四个 `TurnStatus` 常量与 OpenAPI 四值枚举。
- 删除仅由自测调用的 parse / wire / valid helper；直接常量集合测试继续锁定四个 wire 值。
- 修正 owner checklist/test plan 中错误的“五值”表述，不增加兼容转换层。

## 5 验收标准

| ID | 验收点 | 验证 |
|----|--------|------|
| A-1 | Answer events drive follow-up, next-question and completion branches from server-owned state | `TestE2EP0038PracticeEventLoopAnswerFlow`, service tests |
| A-2 | `clientEventId` replay/mismatch, five-kind routing, header policy and cross-user isolation hold | `TestE2EP0039PracticeEventIdempotencyKindRouterAndHeaderPolicy`, repository tests |
| A-3 | Concurrent/stale turn submission preserves contiguous accepted `seq_no` and returns conflict for stale input | `TestE2EP0040PracticeEventConcurrentSeqNoStaleTurnConflict`, store tests |
| A-4 | Completion creates one queued report/job handoff and source event | `TestE2EP0041PracticeSessionCompleteCreatesQueuedReportJob`, store/handler tests |
| A-5 | Completion idempotency matrix and D-35 replay do not duplicate report/job/outbox rows | `TestE2EP0042PracticeSessionCompleteIdempotencyMatrix`, middleware tests |
| A-6 | Privacy/runtime boundary has no real residuals | `TestE2EP0043PracticeEventLoopPrivacyAndOutOfScopeSurface`, backend-practice lint, pruning-surface lint |

## 6 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-10 | 1.8 | Point strict-mode hint evidence at plan 003's single canonical service test. |
| 2026-07-10 | 1.7 | Remove test-only turn-status conversion helpers and align owner evidence to the four-value contract. |
| 2026-07-10 | 1.6 | Remove the unreferenced duplicate string-pointer helper from the Practice API handler. |
| 2026-07-07 | 1.3 | Compress owner docs to current event-loop, completion, idempotency, event/job and privacy contract. |
| 2026-05-14 | 1.2 | Complete implementation and verification for event loop and completion. |
