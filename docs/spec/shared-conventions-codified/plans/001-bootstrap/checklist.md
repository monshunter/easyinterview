# Shared Conventions Bootstrap Checklist

> **版本**: 1.8
> **状态**: completed
> **更新日期**: 2026-07-10

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

## Phase 6: validator dead-code cleanup

- [x] 6.1 删除 `scripts/lint/conventions_yaml.py` 中零调用的 `_require` helper 及其专用 `ValidationError`；验证 AST/symbol inventory、focused/full conventions tests、lint/codegen drift、Go/TS consumers、owner contexts 与 docs/diff/pruning gates。
  <!-- verified: 2026-07-10 method=conventions-validator-dead-code-removal evidence="AST RED found _require as the sole unreferenced production Python symbol; its only dependency was ValidationError. Deleted both without replacement. Python passes 17 tests plus 5 subtests; conventions lint/codegen drift, Go shared tests/vet, frontend conventions/ids 40 tests/typecheck, zero generated diff, owner contexts and docs/diff/pruning gates PASS." -->

## Phase 7: Go workspace/module metadata convergence

- [x] 7.1 根 `go.work` 只 use `./backend`，并与 `.tool-versions`、`backend/go.mod` 统一使用 Go `1.24.5`；标准 tidy 修正直接依赖分类与 checksum，不改 dependency version
- [x] 7.2 根 `lint-go-mod-tidy` 同时验证三处 Go 版本与 module tidy zero drift；验证: RED/GREEN、backend full test/build、B1/A1/product contexts 与 docs/diff/pruning gates
  <!-- red: 2026-07-10 method=go-module-tidy-and-workspace-version-gate evidence="The initial gate failed on go directive/directness/checksum drift. Full compilation then exposed the stale root workspace directive, and the strengthened gate failed with tool=1.24.5, workspace=1.24.0, module=1.24.5." -->
  <!-- verified: 2026-07-10 method=go-workspace-module-convergence evidence="The root workspace uses only ./backend; tool/workspace/module versions all equal 1.24.5, dependency versions are unchanged, and no toolchain directive was added. The tidy gate, full root lint, backend tests/build, three owner contexts and docs/diff/pruning gates pass." -->

## Phase 8: generator decision documentation convergence

- [x] 8.1 Record a RED inventory proving the active spec still labels the implemented generator and TypeScript output boundaries as pending.
  <!-- verified: 2026-07-10 method=shared-conventions-current-decision-red evidence="The scoped absence gate failed first on the active spec's 3.2 pending-decision heading; source inventory separately confirmed yaml.v3 typed loading, hand-written render functions and the already-split TS targets." -->
- [x] 8.2 Move both facts into locked decisions, delete the pending section, and run generator source/test/drift plus owner/product context and docs/index/diff/pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=shared-conventions-generator-decision-convergence evidence="Spec v1.27 contains D-11/D-12 and no pending section or alternative-tool wording. Focused generator tests, conventions lint and the B1 drift gate pass; all 10 generated outputs match and their scoped git diff is empty. Both contexts, links, docs/index/diff/pruning gates pass. Aggregate codegen-check passed its conventions checks before reporting the already-intentional dirty deletion of frontend generated spec.ts from an earlier batch, so this batch used the owner-specific drift gate. No Bug/retrospective report, environment restart or data cleanup was needed." -->
