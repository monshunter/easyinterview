# 001 - OpenAPI v1 Contract Bootstrap Checklist

> **版本**: 1.11
> **状态**: completed
> **更新日期**: 2026-07-10

**关联计划**: [plan](./plan.md)

## Completed Owner Gates

- [x] Current OpenAPI inventory is 10 tags and 37 operations.<!-- verified: 2026-07-10 method=make target=lint-openapi -->
- [x] OpenAPI generated Go and TS artifacts are reproducible from `openapi/openapi.yaml`.
  <!-- verified: 2026-05-04 method=codegen evidence="make codegen-openapi and make codegen-check passed for generated Go/TS artifacts." -->
- [x] B1 shared conventions and B2 error envelope ownership are represented in generated artifacts.
  <!-- verified: 2026-05-04 method=generator-tests evidence="OpenAPI generator tests and generated package checks covered B1 shared type reuse and ApiErrorResponse envelope." -->
- [x] Local API docs renderer is wired through the current `docs-openapi` Make target.
  <!-- verified: 2026-05-04 method=docs-openapi evidence="make docs-openapi generated openapi/dist/index.html with current renderer target." -->
- [x] Child plan handoff is clear: 002 owns fixtures/mock source, 003 owns breaking-change baseline/gate, and 004 owns resume additive coverage.
  <!-- verified: 2026-05-04 method=handoff evidence="plans INDEX and context manifests identify downstream B2 owners." -->
- [x] Current owner docs describe only the 37-operation / 10-tag OpenAPI contract and executable evidence index.<!-- verified: 2026-07-10 method=targeted-grep+context-validation -->
- [x] Test-only snapshot hash cleanup: 删除 production `sha256.go`，在 `run_test.go` 内联 SHA-256 snapshot 计算；验证：production `deadcode` RED/GREEN、OpenAPI generator tests、codegen drift 与 owner docs gates。
  <!-- verified: 2026-07-10 method=openapi-test-only-snapshot-hash-cleanup evidence="Production deadcode RED identified sha256hex as test-only. Inlined SHA-256 calculation in snapshotHashes and deleted sha256.go. Generator tests, file/symbol inventory, deadcode and make codegen-check PASS with no generated drift." -->
- [x] Inventory linter dead constant cleanup: 删除零读取的 `PROVENANCE_REF`；验证 AST/symbol inventory、OpenAPI inventory tests/lint、codegen/fixtures/mock consumers、owner contexts 与 docs/diff/pruning gates。
  <!-- verified: 2026-07-10 method=openapi-inventory-dead-constant-removal evidence="AST RED identified PROVENANCE_REF as the sole unreferenced production Python constant. Deleted it without replacement; real GenerationProvenance schema-name traversal remains. Inventory 19 tests, lint, 37 fixtures, mock contract, codegen drift, Go generator, frontend mock 10 tests/typecheck, zero generated diff, owner contexts and docs/diff/pruning gates PASS." -->
- [x] Frontend raw-spec snapshot removal: 删除无消费方的 generated snapshot、专用 template 与转义 helper；验证 Go/Python RED/GREEN、generator/codegen/openapi、frontend、owner contexts 与 docs/diff/pruning gates。
  <!-- red: 2026-07-10 method=main-entry-reachability+generator-contract evidence="The frontend main-entry graph reported the raw spec snapshot as an unreachable non-test file. The Python pruning suite failed only the new snapshot-absence contract while the prior 10 tests passed, and the new Go Run contract failed because the generator still emitted spec.ts." -->
  <!-- verified: 2026-07-10 method=frontend-raw-openapi-snapshot-removal evidence="Deleted generated spec.ts, its dedicated template, the render branch and the snapshot-only string escaping helper. Python pruning passes 11/11 and the OpenAPI generator package passes including the new positive client/types and negative snapshot contract. make codegen-openapi leaves the snapshot absent; the original make codegen-check passes under an isolated temporary index containing only the expected generated deletion, without touching the real index. Fixture validation passes 37; full frontend passes 136 files/836 tests plus typecheck/build. B2/product contexts, git diff check and pruning surface pass with real_residuals=0. No wire/API behavior, Bug/retrospective report, environment restart or data cleanup was involved." -->

## BDD-Gate

> **BDD 不适用**: 本 plan 交付内部 API contract、codegen pipeline 和 local contract gates，不新增用户可见 UI 或业务 workflow。用户可见 API behavior 由消费该 generated contract 的 backend/frontend/scenario owner 承接。

## Evidence Commands

```bash
make lint-openapi
make codegen-openapi
make codegen-check
cd backend && go test ./cmd/codegen/openapi -count=1
make docs-openapi
python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/openapi-v1-contract/plans/001-bootstrap/context.yaml --target contract
python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check
make docs-check
git diff --check
```
