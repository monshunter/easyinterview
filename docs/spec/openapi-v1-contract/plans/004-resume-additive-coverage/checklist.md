# OpenAPI v1 Contract Flat Resume Coverage Checklist

> **版本**: 1.5
> **状态**: completed
> **更新日期**: 2026-07-07

**关联计划**: [plan](./plan.md)

## Current Contract

- [x] `openapi/openapi.yaml` Resumes tag exposes only `listResumes / registerResume / getResume / getResumeSource / updateResume / duplicateResume / archiveResume / exportResume`.
- [x] `ResumeTailor` tag exposes only `requestResumeTailor / getResumeTailorRun`.
- [x] `scripts/lint/openapi_inventory.py` enforces the current 10 tag / 36 operation inventory and current idempotency / 501 / provenance rules.<!-- verified: 2026-07-07 method=make target=lint-openapi -->
- [x] Resumes fixtures exist for all 8 current Resume operations and validate against OpenAPI.<!-- verified: 2026-07-07 method=make target=validate-fixtures fixtures=36 -->
- [x] ResumeTailor fixtures exist for both current tailor operations and validate against OpenAPI.
- [x] Generated Go/TS artifacts expose the same current operationIds and no version-tree operation surface.
- [x] `exportResume` keeps the P0 typed 501 response path with `RESUME_EXPORT_NOT_AVAILABLE`.
- [x] Fixture validation rejects version-scoped request / response keys in executable fixtures.
- [x] OpenAPI README, fixtures README, B2 spec, mock-contract-suite and engineering-roadmap describe current 36-operation inventory.<!-- verified: 2026-07-07 method=targeted-grep+docs-update -->

## Verification

- [x] `validate_context.py openapi-v1-contract/004 contract`
- [x] `python3 scripts/lint/openapi_inventory.py openapi/openapi.yaml`
- [x] `make lint-openapi`
- [x] `make validate-fixtures`
- [x] `make codegen-check`
- [x] `make openapi-diff`
- [x] Targeted executable grep for version-tree operationIds / params / schemas returns no residuals.
- [x] `sync-doc-index --check`
- [x] `make docs-check`
- [x] `git diff --check`
