# Conversation-level Report Test Plan

> **版本**: 2.1
> **状态**: completed
> **更新日期**: 2026-07-12

## Phase 1
- Contract/prompt/schema/fixture/migration negative and generation tests.
## Phase 2
- Ordered message context, generate/validate/persist, readiness and failure/retry tests.
## Phase 3
- Read-state mapper/list/cursor/replay competency tests.
## Phase 4
- Cross-user/privacy/outbox/audit/log/metric/task-run and full gates.
## Phase 5
- Prompt/schema contract tests require `minimum: 1` / `maximum: 5` and explicit candidate-scale prose.
- Service tests cover 1/2/3/4/5 boundaries plus missing, duplicate and out-of-range dimension failure before persistence.
