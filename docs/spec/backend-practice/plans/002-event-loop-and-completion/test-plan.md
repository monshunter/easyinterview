# Backend Practice Event Loop and Completion Test Plan

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-07-09

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)

## 1 测试策略

本测试计划覆盖 `appendSessionEvent` 与 `completePracticeSession` 的单元、集成、contract、drift 和 runtime-boundary gate。BDD 场景见 [bdd-plan](./bdd-plan.md) 与 [bdd-checklist](./bdd-checklist.md)。

核心断言：

- event kind 路由、AssistantAction 决策、turn-status wire/domain 映射和 malformed payload fail-fast。
- append replay/mismatch、row lock sequencing、cross-user 404、header policy 和 no-audit boundary。
- completion idempotency middleware、D-35 replay、queued report/job handoff、status guard、outbox/audit creation and duplicate-prevention.
- source-event-only `report_generate` job semantics, generated events/jobs drift, OpenAPI generated drift and fixture validity.
- redaction for events, audit, logs, metrics and typed AI task surfaces.

## 2 Coverage Matrix

| 测试源 | 覆盖 | 命令 / 文件 |
|--------|------|-------------|
| event/job source-event-only contract | `report_generate` ownership, generated constants, B3 drift | `make lint-events`, `make codegen-events-check`, `go test ./backend/internal/shared/jobs -count=1` |
| OpenAPI turn status and generated artifacts | 4-value `PracticeTurn.status`, generated Go/TS sync | `make codegen-check`, `python3 scripts/lint/conventions_drift.py --repo-root .` |
| fixtures | append/complete named variants match current schema | `make validate-fixtures` |
| state machine | four current text event kinds, answer branches, optional legacy strict hint, provenance defaults, malformed answer fail-fast | `cd backend && go test ./internal/practice -count=1` |
| append repository | transaction writes, replay/mismatch, row lock, cross-user, outbox boundary | `cd backend && go test ./internal/store/practice -run TestAppendSessionEvent -count=1` |
| append handler | generated request/response, header policy, required `occurredAt`, error mapping | `cd backend && go test ./internal/api/practice -run TestAppendSessionEvent -count=1` |
| completion repository | queued report/job/outbox/audit, D-35 replay, status guard, cross-user | `cd backend && go test ./internal/store/practice -run TestCompleteSession -count=1` |
| completion handler/middleware | idempotency reserve/replay/mismatch, required `clientCompletedAt`, resource handoff | `cd backend && go test ./internal/api/practice -run TestCompletePracticeSession -count=1`, `cd backend && go test ./internal/middleware/idempotency -count=1` |
| HTTP scenario suite | user-visible API behavior for P0.038-P0.043 | `cd backend && go test ./cmd/api -run 'TestE2EP0038|TestE2EP0039|TestE2EP0040|TestE2EP0041|TestE2EP0042|TestE2EP0043' -count=1` |
| runtime boundary lint | removed practice terms, duplicate report handoff paths and compressed turn-status helpers stay out of runtime/scenario/generated surfaces | `python3 scripts/lint/backend_practice_non_current.py --repo-root . --phase all`, `python3 -m pytest scripts/lint/backend_practice_non_current_test.py -q` |
| privacy redaction | no question/answer/hint/prompt/response/secret text in public or operational payloads | outbox emitter tests, cmd/api P0.043 |

## 3 Focused Red / Green Gates

| Area | Red condition | Green condition |
|------|---------------|-----------------|
| `clientEventId` replay | second same-key request writes another event, AI call or outbox row | second same-key same-fingerprint request returns original result with unchanged side-effect counts |
| `clientEventId` mismatch | changed payload is accepted or prior payload leaks in error | 409 conflict with sanitized envelope and no new side effects |
| append sequencing | accepted events have duplicate or gapped `seq_no` | accepted events are contiguous per session and stale-turn requests conflict |
| completion replay | same session creates a second report/job/outbox with another idempotency key | service returns the existing report/job and records the new idempotency snapshot |
| report handoff | source event replay can create a second `report_generate` job | handler/store path is the single report job creator and active dedupe/replay blocks duplicates |
| wire boundary | turn status is compressed or runtime-only provenance fields leak | five statuses and six provenance fields are preserved exactly |
| privacy | text or provider secret appears in event/audit/log/metric/task payload | only IDs, lengths, counts, statuses, profile/model/error summaries and typed task columns are present |

## 4 Closeout Gate

Owner closeout requires all commands listed in [checklist](./checklist.md#收口命令), plus:

- `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/backend-practice/plans/002-event-loop-and-completion/context.yaml --target backend`
- `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
- `make docs-check`
- `git diff --check`
