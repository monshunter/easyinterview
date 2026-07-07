# Shared Conventions Bootstrap Checklist

> **版本**: 1.5
> **状态**: completed
> **更新日期**: 2026-07-07

**关联计划**: [plan](./plan.md)

## Phase 1: truth source and generator

- [x] 1.1 `shared/conventions.yaml` contains the current enum/error/job/id/idempotency/pagination/AI vocabulary contract（验证：`make lint-conventions`）
- [x] 1.2 `backend/cmd/codegen/conventions` generates Go and TS artifacts from the same source（验证：`make codegen-conventions`、`make codegen-check`）
- [x] 1.3 generated output remains idempotent（验证：second-pass codegen/check gates）

## Phase 2: Go / TS shared helpers

- [x] 2.1 Go generated and handwritten shared packages cover `types`, `errors`, `idx` and AI vocabulary（验证：`go test ./backend/internal/shared/...`）
- [x] 2.2 frontend conventions and ids libs expose generated enum/error/pagination/AI vocabulary and ID helpers（验证：`pnpm --filter @easyinterview/frontend exec tsc --noEmit`、`pnpm --dir frontend test src/lib/conventions src/lib/ids`）
- [x] 2.3 UUIDv7 and `tmp_` prefix rejection semantics match across Go and TS（验证：Go/TS ID tests）
- [x] 2.4 `Idempotency-Key` generation/parsing/24h TTL semantics match across Go and TS（验证：Go/TS idempotency tests）

## Phase 3: lint and naming gates

- [x] 3.1 error codes are `UPPER_SNAKE_CASE` and generated on both sides（验证：`make lint-conventions`）
- [x] 3.2 enum values are `lower_snake_case` and JSON fields are `camelCase`（验证：conventions YAML lint）
- [x] 3.3 generated Go/TS parity tests protect the current catalog（验证：`go test ./backend/internal/shared/types -count=1`、frontend conventions parity tests）

## Phase 4: product-scope enum contract

- [x] 4.1 `PracticeMode` current values are `assisted` and `strict`（验证：generated Go/TS enum tests）
- [x] 4.2 `PracticeGoal` current values are `baseline`, `retry_current_round` and `next_round`（验证：generated Go/TS enum tests）
- [x] 4.3 `QuestionReviewStatus` current values are `open`, `queued_for_retry` and `resolved`（验证：generated Go/TS enum tests）
- [x] 4.4 removed practice mode/goal/status values stay out of generated runtime surfaces（验证：conventions parity tests and pruning-surface lint）

## Phase 5: closeout

- [x] 5.1 focused conventions gates pass（验证：`make lint-conventions`、`make codegen-conventions`、`make codegen-check`）
- [x] 5.2 focused Go/TS shared gates pass（验证：`go test ./backend/internal/shared/...`、`go vet ./backend/...`、frontend typecheck/conventions tests）
- [x] 5.3 owner context and docs indexes are current（验证：`validate_context.py shared-conventions-codified/001 backend`、`sync-doc-index --check`、`make docs-check`）
