# Conversation-level Report Test Checklist

> **版本**: 2.1
> **状态**: completed
> **更新日期**: 2026-07-12

**关联 Test Plan**: [test-plan](./test-plan.md)

## Phase 1
- [x] Contract/prompt/schema/migration tests pass.
## Phase 2
- [x] Generate/persist/readiness/failure tests pass.
## Phase 3
- [x] Read/replay competency tests pass.
## Phase 4
- [x] Privacy/full regression gates pass.
## Phase 5
- [x] Candidate score prompt/schema/runtime boundary tests pass. (`make lint-prompts`; `go test ./backend/internal/review -count=1`)
