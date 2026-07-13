# Shared Conventions Bootstrap

> **版本**: 1.11
> **状态**: active
> **更新日期**: 2026-07-13

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

本 plan 承接 B1 shared conventions foundation：

- `shared/conventions.yaml` 是 Go / TypeScript 共享枚举、错误码、分页结构、ID 规则、idempotency TTL 和 AI shared vocabulary 的可执行真理源。
- `backend/cmd/codegen/conventions` 从同一 YAML 生成 Go shared types/errors/AI vocabulary/ID constants 与 frontend conventions/ids files。
- Go shared module 落点为 `backend/internal/shared/{types,errors,idx,ai}`；TypeScript shared lib 落点为 `frontend/src/lib/{conventions,ids}`。
- UUIDv7、`tmp_` 前缀拒绝、24h `Idempotency-Key` 解析/生成和 `ApiError` inner object 在 Go/TS 双端保持一致。
- 本地 lint gate 约束错误码 `UPPER_SNAKE_CASE`、枚举值 `lower_snake_case`、JSON field `camelCase`、generator drift 和跨语言 parity。
- 根 `go.work` 只 use `./backend`；workspace、module 与根 tool version 使用同一 Go 版本，module metadata 保持 tidy 零漂移。
- TargetJob 当前只接受粘贴 JD：删除 `TARGET_IMPORT_SOURCE_INVALID` / `TARGET_IMPORT_SOURCE_UNAVAILABLE`，保留 `VALIDATION_FAILED` 与 `TARGET_IMPORT_FAILED` 的通用校验/失败语义。

## 2 当前合同

| surface | current behavior | truth / generated files | coverage |
|---------|------------------|-------------------------|----------|
| conventions truth source | current enum catalog, 21 error codes after Phase 10, job statuses, UUID/idempotency/pagination constants, AI vocabulary | `shared/conventions.yaml` | `make lint-conventions` |
| Go generated shared types | enums, HTTP DTOs, error codes, AI vocabulary, ID constants | `backend/internal/shared/{types,errors,idx,ai}` | `go test ./backend/internal/shared/...`, `make codegen-check` |
| TS generated shared types | enum literals, error codes, pagination, AI vocabulary, ID constants | `frontend/src/lib/{conventions,ids}` | frontend conventions Vitest/typecheck gates |
| UUID / server ID tools | UUIDv7 generation, `tmp_` rejection, shared ID regex | Go `idx`, TS `ids` | Go/TS id tests |
| Idempotency-Key tools | shared 24h TTL and strict decimal timestamp parsing | Go `idx/idempotency.go`, TS `conventions/idempotency.ts` | Go/TS idempotency tests |
| naming / drift lint | generated files match YAML and naming rules | `scripts/lint/conventions_yaml.py`, `make codegen-check` | lint and codegen gates |
| Go workspace/module | root workspace uses only backend; tool/workspace/module versions and tidy metadata agree | `.tool-versions`, `go.work`, `backend/go.mod`, `backend/go.sum` | `make lint-go-mod-tidy` |

## 3 质量门禁

- **Plan 类型**: `contract + tooling + code-internal`。
- **TDD 策略**: 适用。Focused tests cover generator idempotency, Go/TS shared parity, error-code lint, UUID/idempotency helpers and conventions drift.
- **BDD 策略**: 不适用。本 plan 是内部共享契约和工具链；用户可见 API/UI flows由消费方 owner 的 BDD gate 覆盖。
- **替代验证 gate**:
  - `make lint-conventions`
  - `make lint-go-mod-tidy`
  - `make codegen-conventions`
  - `make codegen-check`
  - `go test ./backend/internal/shared/...`
  - `go vet ./backend/...`
  - `pnpm --filter @easyinterview/frontend exec tsc --noEmit`
  - `pnpm --dir frontend test src/lib/conventions src/lib/ids`
  - `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/shared-conventions-codified/plans/001-bootstrap/context.yaml --target backend`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `make docs-check`

## 4 实施步骤

Phase 1-9 保留为历史完成证据；Phase 10 是本轮 current net-state owner，不回写历史 PASS。

### Phase 1: truth source and generator

- Maintain `shared/conventions.yaml` as the single B1 source.
- Generate Go and TS shared artifacts from the same source.
- Keep generator output idempotent and covered by `make codegen-check`.

### Phase 2: Go / TS shared helpers

- Provide Go shared `types`, `errors`, `idx` and AI vocabulary packages.
- Provide frontend `conventions` and `ids` libs.
- Keep UUIDv7 and `tmp_` prefix semantics aligned.
- Keep `Idempotency-Key` generation/parsing/TTL semantics aligned across languages.

### Phase 3: lint and naming gates

- Reject invalid error-code names/values.
- Reject enum and JSON field naming drift.
- Keep generated Go/TS parity tests current.

### Phase 4: product-scope enum contract

- Keep `PracticeGoal=baseline|retry_current_round|next_round`.
- Keep removed `PracticeMode` and `QuestionReviewStatus` types and removed practice goal values out of generated runtime surfaces.

### Phase 5: closeout

- Run focused conventions, Go shared and frontend conventions gates.
- Sync owner index and product-scope evidence.
- Leave remote CI integration to CI owner gates; B1 retains local executable gates.

### Phase 6: validator dead-code cleanup

- Delete the unreferenced `_require` helper and its dedicated `ValidationError` type from the conventions validator.
- Keep validator behavior covered by focused/full lint and codegen gates.

### Phase 7: Go workspace/module metadata convergence

- Keep root `go.work` limited to `./backend` and align its Go version with `.tool-versions` and `backend/go.mod` at `1.24.5`.
- Let standard `go mod tidy` own direct/indirect classification and checksums without changing dependency versions.
- Enforce workspace/module version agreement and tidy zero drift through `make lint-go-mod-tidy`.

### Phase 8: generator decision documentation convergence

- Replace the two implemented choices still listed as pending in the active spec with locked decisions that match the current `yaml.v3` loader, hand-written renderer and generated TypeScript file boundaries.
- Delete the empty pending-decision section without a historical marker or replacement compatibility note.
- Verify the current generator source, focused tests, drift gate, owner contexts and docs/index/diff/pruning gates; do not change generated artifacts.

### Phase 9: report oversized-context error contract

- Add exactly one canonical entry to `shared/conventions.yaml`: `REPORT_CONTEXT_TOO_LARGE`, message `report context exceeds supported generation size`, `retryable: false`. B1 owns this cross-language literal and retryability; backend-review owns the 48,000-byte boundary and terminal behavior.
- Drive the change Red-Green through conventions validator/generator tests, then regenerate Go `backend/internal/shared/errors`, frontend `frontend/src/lib/conventions/errors.ts` and parity fixtures. Reject a missing/duplicate/misspelled code, changed message, or `retryable: true`.
- Emit `REPORT_CONTEXT_TOO_LARGE_CONVENTIONS_PASS` after B1 conventions lint/codegen idempotency, Go/TS parity and owner context validation pass. The marker means only that the canonical cross-language literal and retryability are ready for downstream consumption; it does not wait for B2 and therefore cannot encode OpenAPI parity.
- B2 consumes that marker, synchronizes `ApiErrorCode` from B1 without hand-maintaining a second error list, and independently proves through the OPENAPI-001 normalized base-ref audit that the single `enum_value_added` finding for `REPORT_CONTEXT_TOO_LARGE` is additive, outside the breaking allowset, with no unrecorded enum delta. User-visible terminal behavior and 48,000/+1 byte tests remain backend-review-owned.

### Phase 10: TargetJob paste-only error vocabulary

- **RED**: add exact contract assertions proving the current YAML/generated/OpenAPI parity still exposes `TARGET_IMPORT_SOURCE_INVALID` and `TARGET_IMPORT_SOURCE_UNAVAILABLE`; also lock positive assertions for `VALIDATION_FAILED`, `TARGET_IMPORT_FAILED`, `TARGET_JOB_NOT_FOUND` and `TARGET_INVALID_STATE_TRANSITION` so the cleanup cannot over-delete shared behavior.
- **GREEN**: remove only the two source-specific entries from `shared/conventions.yaml`, regenerate Go/TS artifacts, and hand off the same exact enum subtraction to B2 OpenAPI codegen/parity. Do not add compatibility aliases or replacement source-specific codes.
- **BDD 不适用**: this is an internal cross-language vocabulary contraction rather than a standalone user flow. Substitute gates are conventions lint, generator unit/parity tests, two-pass idempotent codegen, Go/TS focused tests, OpenAPI parity handoff, and exact zero-reference searches for both removed literals with positive retained-code probes.

## 5 验收标准

| ID | 验收点 | 验证 |
|----|--------|------|
| A-1 | `shared/conventions.yaml` validates current enum/error/job/id contract | `make lint-conventions` |
| A-2 | Go and TS generated artifacts match the truth source | `make codegen-check` |
| A-3 | Go shared helpers pass UUID/idempotency/error tests | `go test ./backend/internal/shared/...` |
| A-4 | Frontend conventions and ids helpers pass type/test gates | `pnpm --filter @easyinterview/frontend exec tsc --noEmit`; `pnpm --dir frontend test src/lib/conventions src/lib/ids` |
| A-5 | Error code and enum naming remain enforceable locally | conventions lint and parity tests |
| A-6 | Current product-scope enum values stay aligned across generated surfaces | conventions parity tests and pruning-surface lint |
| A-7 | `REPORT_CONTEXT_TOO_LARGE` is single-source and non-retryable before B2 consumption | conventions lint/codegen + Go/TS parity + owner context validation |
| A-8 | TargetJob paste-only vocabulary removes both source-specific errors while retaining generic validation and retryable import-failure semantics | conventions RED/GREEN + codegen idempotency + Go/TS/OpenAPI parity + exact zero-reference/positive-presence probes |

## 6 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-13 | 1.11 | Reopen Phase 10 to contract TargetJob errors to the paste-only vocabulary and require generated/OpenAPI zero-reference closure. |
| 2026-07-12 | 1.10 | Remove the B1/B2 cycle: B1 source-ready marker covers YAML/Go/TS only; B2 independently proves OpenAPI parity/oracle. |
| 2026-07-12 | 1.9 | Reopen Phase 9 for the single-source REPORT_CONTEXT_TOO_LARGE YAML/codegen/OpenAPI enum contract. |
| 2026-07-10 | 1.8 | Reconcile implemented generator and TypeScript output choices as current locked decisions. |
| 2026-07-10 | 1.7 | Lock the root Go workspace to backend, align all Go version declarations, and enforce tidy zero drift. |
| 2026-07-10 | 1.6 | 删除 conventions validator 中无调用方的异常式校验 helper 与专用异常类型。 |
| 2026-07-07 | 1.5 | Compress owner docs to current shared truth source, generator, Go/TS helper and local lint contract. |
| 2026-05-04 | 1.4 | Complete quality gate classification for shared conventions foundation. |
