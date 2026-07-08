# 001 - OpenAPI v1 Contract Bootstrap Checklist

> **版本**: 1.7
> **状态**: completed
> **更新日期**: 2026-07-07

**关联计划**: [plan](./plan.md)

## Completed Owner Gates

- [x] Current OpenAPI inventory is 10 tags and 36 operations.<!-- verified: 2026-07-07 method=make target=lint-openapi -->
- [x] OpenAPI generated Go and TS artifacts are reproducible from `openapi/openapi.yaml`.
  <!-- verified: 2026-05-04 method=codegen evidence="make codegen-openapi and make codegen-check passed for generated Go/TS artifacts." -->
- [x] B1 shared conventions and B2 error envelope ownership are represented in generated artifacts.
  <!-- verified: 2026-05-04 method=generator-tests evidence="OpenAPI generator tests and generated package checks covered B1 shared type reuse and ApiErrorResponse envelope." -->
- [x] Local API docs renderer is wired through the current `docs-openapi` Make target.
  <!-- verified: 2026-05-04 method=docs-openapi evidence="make docs-openapi generated openapi/dist/index.html with current renderer target." -->
- [x] Child plan handoff is clear: 002 owns fixtures/mock source, 003 owns breaking-change baseline/gate, and 004 owns resume additive coverage.
  <!-- verified: 2026-05-04 method=handoff evidence="plans INDEX and context manifests identify downstream B2 owners." -->
- [x] Current owner docs describe only the 36-operation / 10-tag OpenAPI contract and executable evidence index.<!-- verified: 2026-07-07 method=targeted-grep+context-validation -->

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
