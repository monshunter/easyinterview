# OpenAPI v1 Contract Flat Resume Coverage

> **版本**: 1.5
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

本 plan 固化 OpenAPI 当前 flat Resume 覆盖面：`openapi/openapi.yaml` 保持 10 tag / 36 operation 合同，`Resumes` tag 只保留 flat `Resume` 资产读写、PDF source 预览、复制、归档、导出，`ResumeTailor` tag 只保留 tailor run 请求与读取。版本树、分支、suggestion accept/reject 和 version-scoped export 不再作为当前 OpenAPI operation、schema、fixture、generated client 或 inventory lint 正向面存在。

## 2 当前 Contract Surface

| Area | Current truth |
|------|---------------|
| OpenAPI inventory | `scripts/lint/openapi_inventory.py` enforces 10 tags and 36 operations |
| Resume schemas | `Resume`, `PaginatedResume`, `ResumeWithJob`, `RegisterResumeRequest`, `UpdateResumeRequest`, `DuplicateResumeRequest`, `ResumeTailorRun`, `ResumeTailorRunWithJob` |
| Resume fixtures | `openapi/fixtures/Resumes/{listResumes,registerResume,getResume,getResumeSource,updateResume,duplicateResume,archiveResume,exportResume}.json` |
| ResumeTailor fixtures | `openapi/fixtures/ResumeTailor/{requestResumeTailor,getResumeTailorRun}.json` |
| Generated clients | `frontend/src/api/generated/client.ts` and `backend/internal/api/generated/*` expose the same operationIds |
| Export behavior | `exportResume` is the current P0 501 typed response path with `RESUME_EXPORT_NOT_AVAILABLE` |
| Provenance | AI-generated resume/tailor fields remain reachable from `GenerationProvenance`; inventory and fixture gates enforce reachability |

## 3 Operation Matrix

| operationId | Fixture | Frontend consumer | Backend handler | Persistence | AI dependency | Gate |
|-------------|---------|-------------------|-----------------|-------------|---------------|------|
| `listResumes` | `Resumes/listResumes.json` | Resume picker / Resume Workshop list | backend-resume real handler | `resumes` | none | fixture + generated client + inventory |
| `registerResume` | `Resumes/registerResume.json` | upload / paste create flow | backend-resume real handler | `resumes`, upload object link | parse job backend-only | fixture + generated client + B4 contract |
| `getResume` | `Resumes/getResume.json` | resume detail, report context, workspace bound summary | backend-resume real handler | `resumes` | none | fixture + generated client |
| `getResumeSource` | `Resumes/getResumeSource.json` | PDF resume detail preview object | backend-resume real handler | `resumes.file_object_id` + `file_objects` + object bytes | none | fixture + generated client + inventory |
| `updateResume` | `Resumes/updateResume.json` | Resume Workshop edit overwrite | backend-resume real handler | `resumes` | none | idempotency + fixture |
| `duplicateResume` | `Resumes/duplicateResume.json` | Resume Workshop save-as-new | backend-resume real handler | `resumes` | none | idempotency + fixture |
| `archiveResume` | `Resumes/archiveResume.json` | archive action / privacy boundary | backend-resume real handler | `resumes.deleted_at` | none | idempotency + fixture |
| `exportResume` | `Resumes/exportResume.json` | export fallback UI | P0 typed 501 response | none | none | 501 allowlist + fixture |
| `requestResumeTailor` | `ResumeTailor/requestResumeTailor.json` | Rewrites run trigger | backend-resume tailor runner | `ai_task_runs`, outbox job | resume tailor profile | idempotency + fixture |
| `getResumeTailorRun` | `ResumeTailor/getResumeTailorRun.json` | Rewrites polling | backend-resume tailor runner | `ai_task_runs` | resume tailor profile output | fixture |

## 4 Negative Boundary

Current B2 gates must reject version-tree operationIds, version-scoped path params and version-only schemas in executable OpenAPI, fixtures, generated clients and lint truth. The executable boundary is enforced by:

- `scripts/lint/openapi_inventory.py`
- `scripts/lint/validate_fixtures.py`
- `make lint-openapi`
- `make validate-fixtures`
- `make codegen-check`
- `make openapi-diff`

## 5 Verification

Current closeout evidence:

- `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/openapi-v1-contract/plans/004-resume-additive-coverage/context.yaml --target contract`
- `python3 scripts/lint/openapi_inventory.py openapi/openapi.yaml` → 10 tags / 36 operations
- `make lint-openapi`
- `make validate-fixtures`
- `make codegen-check`
- `make openapi-diff`
- Targeted grep over `openapi/`, generated clients and fixtures for version-tree operationIds / params returns no executable residuals
- `sync-doc-index --check`
- `make docs-check`
- `git diff --check`

## 6 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-07 | 1.5 | Add `getResumeSource` PDF source preview to the current flat Resume OpenAPI contract, fixture set, generated clients and inventory gate. |
| 2026-07-07 | 1.4 | Compress owner docs to the then-current flat Resume / ResumeTailor OpenAPI contract, fixtures, generated clients and 501 export gate. |
