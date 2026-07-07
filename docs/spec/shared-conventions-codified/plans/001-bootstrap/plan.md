# Shared Conventions Bootstrap

> **版本**: 1.5
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

本 plan 承接 B1 shared conventions foundation：

- `shared/conventions.yaml` 是 Go / TypeScript 共享枚举、错误码、分页结构、ID 规则、idempotency TTL 和 AI shared vocabulary 的可执行真理源。
- `backend/cmd/codegen/conventions` 从同一 YAML 生成 Go shared types/errors/AI vocabulary/ID constants 与 frontend conventions/ids files。
- Go shared module 落点为 `backend/internal/shared/{types,errors,idx,ai}`；TypeScript shared lib 落点为 `frontend/src/lib/{conventions,ids}`。
- UUIDv7、`tmp_` 前缀拒绝、24h `Idempotency-Key` 解析/生成和 `ApiError` inner object 在 Go/TS 双端保持一致。
- 本地 lint gate 约束错误码 `UPPER_SNAKE_CASE`、枚举值 `lower_snake_case`、JSON field `camelCase`、generator drift 和跨语言 parity。

## 2 当前合同

| surface | current behavior | truth / generated files | coverage |
|---------|------------------|-------------------------|----------|
| conventions truth source | current enum catalog, 22 error codes, job statuses, UUID/idempotency/pagination constants, AI vocabulary | `shared/conventions.yaml` | `make lint-conventions` |
| Go generated shared types | enums, HTTP DTOs, error codes, AI vocabulary, ID constants | `backend/internal/shared/{types,errors,idx,ai}` | `go test ./backend/internal/shared/...`, `make codegen-check` |
| TS generated shared types | enum literals, error codes, pagination, AI vocabulary, ID constants | `frontend/src/lib/{conventions,ids}` | frontend conventions Vitest/typecheck gates |
| UUID / server ID tools | UUIDv7 generation, `tmp_` rejection, shared ID regex | Go `idx`, TS `ids` | Go/TS id tests |
| Idempotency-Key tools | shared 24h TTL and strict decimal timestamp parsing | Go `idx/idempotency.go`, TS `conventions/idempotency.ts` | Go/TS idempotency tests |
| naming / drift lint | generated files match YAML and naming rules | `scripts/lint/conventions_yaml.py`, `make codegen-check` | lint and codegen gates |

## 3 质量门禁

- **Plan 类型**: `contract + tooling + code-internal`。
- **TDD 策略**: 适用。Focused tests cover generator idempotency, Go/TS shared parity, error-code lint, UUID/idempotency helpers and conventions drift.
- **BDD 策略**: 不适用。本 plan 是内部共享契约和工具链；用户可见 API/UI flows由消费方 owner 的 BDD gate 覆盖。
- **替代验证 gate**:
  - `make lint-conventions`
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

- Keep `PracticeMode=assisted|strict`.
- Keep `PracticeGoal=baseline|retry_current_round|next_round`.
- Keep `QuestionReviewStatus=open|queued_for_retry|resolved`.
- Keep removed practice mode/goal/status values out of generated runtime surfaces.

### Phase 5: closeout

- Run focused conventions, Go shared and frontend conventions gates.
- Sync owner index and product-scope evidence.
- Leave remote CI integration to CI owner gates; B1 retains local executable gates.

## 5 验收标准

| ID | 验收点 | 验证 |
|----|--------|------|
| A-1 | `shared/conventions.yaml` validates current enum/error/job/id contract | `make lint-conventions` |
| A-2 | Go and TS generated artifacts match the truth source | `make codegen-check` |
| A-3 | Go shared helpers pass UUID/idempotency/error tests | `go test ./backend/internal/shared/...` |
| A-4 | Frontend conventions and ids helpers pass type/test gates | `pnpm --filter @easyinterview/frontend exec tsc --noEmit`; `pnpm --dir frontend test src/lib/conventions src/lib/ids` |
| A-5 | Error code and enum naming remain enforceable locally | conventions lint and parity tests |
| A-6 | Current product-scope enum values stay aligned across generated surfaces | conventions parity tests and pruning-surface lint |

## 6 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-07 | 1.5 | Compress owner docs to current shared truth source, generator, Go/TS helper and local lint contract. |
| 2026-05-04 | 1.4 | Complete quality gate classification for shared conventions foundation. |
