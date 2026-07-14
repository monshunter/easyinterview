# OpenAPI v1 Contract Flat Resume Coverage

> **版本**: 1.7
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

本 plan 固化 OpenAPI 当前 flat Resume 覆盖面：`openapi/openapi.yaml` 保持 10 tag / 37 operation 合同，`Resumes` tag 只保留 flat `Resume` 资产读写、PDF source 预览、复制、归档、导出，`ResumeTailor` tag 只保留 tailor run 请求与读取。版本树、分支、suggestion accept/reject 和 version-scoped export 不再作为当前 OpenAPI operation、schema、fixture、generated client 或 inventory lint 正向面存在。

## 2 当前 Contract Surface

| Area | Current truth |
|------|---------------|
| OpenAPI inventory | `scripts/lint/openapi_inventory.py` enforces 10 tags and 37 operations |
| Resume schemas | `ResumeSummary` list projection；full `Resume` detail/mutation；`PaginatedResume`, `ResumeWithJob`, `RegisterResumeRequest`, `UpdateResumeRequest`, `DuplicateResumeRequest`, `ResumeTailorRun`, `ResumeTailorRunWithJob` |
| Resume fixtures | `openapi/fixtures/Resumes/{listResumes,registerResume,getResume,getResumeSource,updateResume,duplicateResume,archiveResume,exportResume}.json` |
| ResumeTailor fixtures | `openapi/fixtures/ResumeTailor/{requestResumeTailor,getResumeTailorRun}.json` |
| Generated clients | `frontend/src/api/generated/client.ts` and `backend/internal/api/generated/*` expose the same operationIds |
| Export behavior | `exportResume` is the current P0 501 typed response path with `RESUME_EXPORT_NOT_AVAILABLE` |
| Provenance | AI-generated resume/tailor fields remain reachable from `GenerationProvenance`; inventory and fixture gates enforce reachability |

## 3 质量门禁分类

- **Plan 类型**: `contract + feature-behavior + cross-layer handoff`
- **TDD 策略**: Phase 7 先以 OpenAPI inventory/generator、fixture validator、backend store/service/handler 与 frontend consumer focused tests 建立 RED，再最小修改 source/codegen/projection/consumer；每个 checklist item 必须保留实际断言来源。
- **BDD 策略**: 不适用。本 plan 只维护 OpenAPI schema、fixture、generated artifact 与 consumer contract，不拥有真实 API/UI 用户流程，也不把代码层测试包装成 E2E。
- **替代验证 gate**: `make lint-openapi`、`make validate-fixtures`、`make codegen-check`、`make openapi-diff`、scoped negative search；阶段收口由仓库根 `make test` 完成前后端全量回归。

### 3.1 Operation Matrix

| operationId | Fixture | Frontend consumer | Backend handler | Persistence | AI dependency | Gate |
|-------------|---------|-------------------|-----------------|-------------|---------------|------|
| `listResumes` | `Resumes/listResumes.json` summary-only items | Home picker / Resume Workshop list / every list consumer | backend-resume dedicated summary projection | `resumes` list-safe columns only | none | fixture + generated `ResumeSummary` + inventory |
| `registerResume` | `Resumes/registerResume.json` | upload / paste create flow | backend-resume real handler | `resumes`, upload object link | parse job backend-only | fixture + generated client + B4 contract |
| `getResume` | `Resumes/getResume.json` full detail | Resume Workshop read-only detail and explicit detail consumers | backend-resume owned full-detail handler | `resumes` detail columns | none | fixture + generated full `Resume` |
| `getResumeSource` | `Resumes/getResumeSource.json` | PDF resume detail preview object | backend-resume real handler | `resumes.file_object_id` + `file_objects` + object bytes | none | fixture + generated client + inventory |
| `updateResume` | `Resumes/updateResume.json` | Resume Workshop edit overwrite | backend-resume real handler | `resumes` | none | idempotency + fixture |
| `duplicateResume` | `Resumes/duplicateResume.json` | Resume Workshop save-as-new | backend-resume real handler | `resumes` | none | idempotency + fixture |
| `archiveResume` | `Resumes/archiveResume.json` | archive action / privacy boundary | backend-resume real handler | `resumes.deleted_at` | none | idempotency + fixture |
| `exportResume` | `Resumes/exportResume.json` | export fallback UI | P0 typed 501 response | none | none | 501 allowlist + fixture |
| `requestResumeTailor` | `ResumeTailor/requestResumeTailor.json` | Rewrites run trigger | backend-resume tailor runner | `ai_task_runs`, outbox job | resume tailor profile | idempotency + fixture |
| `getResumeTailorRun` | `ResumeTailor/getResumeTailorRun.json` | Rewrites polling | backend-resume tailor runner | `ai_task_runs` | resume tailor profile output | fixture |

## 4 Negative Boundary

Current B2 gates must reject version-tree operationIds, version-scoped path params and version-only schemas in executable OpenAPI, fixtures, generated clients and lint truth. The executable boundary is enforced by:

Phase 7 additionally rejects full `Resume` fields/provenance in `listResumes`, any non-nine-field `ResumeSummary`, compatibility aliases and frontend list-time N+1 `getResume` fallback；full detail fields remain valid only on explicit detail/mutation operations.

- `scripts/lint/openapi_inventory.py`
- `scripts/lint/validate_fixtures.py`
- `make lint-openapi`
- `make validate-fixtures`
- `make codegen-check`
- `make openapi-diff`

## 5 Verification

Prior Phase 1-6 closeout evidence remains historical；Phase 7 requires fresh current evidence:

- `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/openapi-v1-contract/plans/004-resume-additive-coverage/context.yaml --target contract`
- `python3 scripts/lint/openapi_inventory.py openapi/openapi.yaml` → 10 tags / 37 operations
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
| 2026-07-14 | 1.7 | Reopen Phase 7 for OPENAPI-005 closed ResumeSummary list projection, full detail separation and all-consumer BDD handoff. |
| 2026-07-10 | 1.6 | Align current flat Resume coverage docs with the 37-operation B2 contract. |
| 2026-07-07 | 1.5 | Add `getResumeSource` PDF source preview to the current flat Resume OpenAPI contract, fixture set, generated clients and inventory gate. |
| 2026-07-07 | 1.4 | Compress owner docs to the then-current flat Resume / ResumeTailor OpenAPI contract, fixtures, generated clients and 501 export gate. |

## 7 Phase 7: OPENAPI-005 Resume list summary / detail split

### 7.1 Contract and fixture RED

Consume accepted OPENAPI-005. OpenAPI inventory/generator and fixture tests must fail while `PaginatedResume.items` references full `Resume`, while `ResumeSummary` is not closed/required/exact, while list fixtures expose detail/provenance, or while `getResume` ceases to return full `Resume`. Lock list/get method/path/operationId/200, pagination and 37/10 inventory.

### 7.2 Backend projection GREEN

Backend-resume must introduce a dedicated list record/query/service mapper/handler projection for the exact nine summary fields. SQL and row scans may not fetch `file_object_id`, raw/parsed body, parsed summary, structured profile/provenance, created/deleted/status detail solely to serve the list. SQL may inspect JSON/text columns only inside scalar expressions: `summaryHeadline` is the first trim-nonempty string in `parsed_summary.headline`、`parsed_summary.basics.headline`、`structured_profile.headline`、`structured_profile.basics.headline` or null；`hasReadableContent=true` exactly when trim-nonempty `parsed_text_snapshot` / `original_text` exists or `structured_profile` is a nonempty object. `file_object_id`、`source_type`、`parse_status` never imply readability. Only the derived scalar/string enters the summary row；owned/cursor/isolation behavior remains. `getResume` keeps its full detail lookup and mapper.

### 7.3 Generated and frontend consumer GREEN

Go/TS generated results type list items as `ResumeSummary`; full detail operations retain `Resume`. Home, Resume Workshop and every `listResumes` consumer use only summary fields, including backend `hasReadableContent`, and do not recover removed data via N+1 detail calls or frontend storage. Row navigation then issues the explicit `getResume` detail query. Compile/type tests and focused UI tests must inventory all consumers rather than patching only the visible list.

### 7.4 Fixture, mock and audit closure

002 Phase 11 owns list/get fixtures, examples and Prism/mock parity；003 Phase 9 generates and exact-matches the declared OPENAPI-005 oracle before guarded re-freeze；001 Phase 16 owns source/codegen。所有层在同一批次迁移，不增加 alias、可选详情兼容字段、第二个列表 endpoint 或 frontend fallback；最终由根 `make test` 执行前后端全量回归。
