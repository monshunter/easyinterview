# OpenAPI v1 Contract Breaking-Change Gate Checklist

> **版本**: 1.6
> **状态**: completed
> **更新日期**: 2026-07-07

**关联计划**: [plan](./plan.md)

## Phase 1: baseline and diff entrypoint

- [x] `openapi/baseline/openapi-v1.0.0.yaml` represents current 35 operation / 10 tag freeze baseline（验证：`make openapi-diff` PASS）
- [x] `make openapi-diff` runs the wrapper-backed local diff gate from the repo root（验证：Makefile target exists and passes）
- [x] baseline selection supports explicit override for future baselines（验证：wrapper/openapi baseline tests cover version selection）

## Phase 2: ruleset and wrapper

- [x] breaking endpoint/path/method/schema/required/enum changes fail（验证：`python3 -m unittest scripts.lint.openapi_diff_test` PASS）
- [x] additive endpoint/tag/optional-field/enum-value/query/example changes pass（验证：openapi diff unit tests PASS）
- [x] `oneOf` / `allOf` / `anyOf` composition changes are inspected recursively（验证：openapi diff unit tests PASS）
- [x] privacy export status whitelist is limited to `POST /api/v1/privacy/exports` `501 -> 202`（验证：openapi diff unit tests + `make openapi-diff` PASS）

## Phase 3: contract governance

- [x] ADR template for accepted breaking changes exists under `docs/spec/openapi-v1-contract/decisions/`（验证：docs link checks PASS）
- [x] baseline README documents baseline selection and versioning rules（验证：docs link checks PASS）
- [x] response-status whitelist and baseline changes require same-change contract records（验证：wrapper tests cover missing-record failure）

## Phase 4: closeout

- [x] `make lint-openapi` passes with current 10 tags / 35 operations
- [x] `make validate-fixtures` passes
- [x] `make codegen-check` passes
- [x] `make openapi-diff` passes
- [x] context, INDEX and docs links pass（验证：`validate_context.py ...003...`, `sync-doc-index --check`, `make docs-check` PASS）
