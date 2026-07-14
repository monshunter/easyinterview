# OpenAPI v1 Contract Flat Resume Coverage Checklist

> **版本**: 1.7
> **状态**: completed
> **更新日期**: 2026-07-14

**关联计划**: [plan](./plan.md)

## Current Contract

- [x] `openapi/openapi.yaml` Resumes tag exposes only `listResumes / registerResume / getResume / getResumeSource / updateResume / duplicateResume / archiveResume / exportResume`.
- [x] `ResumeTailor` tag exposes only `requestResumeTailor / getResumeTailorRun`.
- [x] `scripts/lint/openapi_inventory.py` enforces the current 10 tag / 37 operation inventory and current idempotency / 501 / provenance rules.<!-- verified: 2026-07-10 method=make target=lint-openapi -->
- [x] Resumes fixtures exist for all 8 current Resume operations and validate against OpenAPI.<!-- verified: 2026-07-10 method=make target=validate-fixtures fixtures=37 -->
- [x] ResumeTailor fixtures exist for both current tailor operations and validate against OpenAPI.
- [x] Generated Go/TS artifacts expose the same current operationIds and no version-tree operation surface.
- [x] `exportResume` keeps the P0 typed 501 response path with `RESUME_EXPORT_NOT_AVAILABLE`.
- [x] Fixture validation rejects version-scoped request / response keys in executable fixtures.
- [x] OpenAPI README, fixtures README, B2 spec, mock-contract-suite and engineering-roadmap describe current 37-operation inventory.<!-- verified: 2026-07-10 method=targeted-grep+docs-update -->

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

## Phase 7: OPENAPI-005 Resume list summary / detail split

- [x] 7.1 OWNER/RED: accepted OPENAPI-005 + spec/history 1.59 exist；focused schema/generator/fixture tests reject old full list items, non-exact/open summary, detail/provenance extras and any getResume detail regression while list/get invariants remain.
- [x] 7.2 CONTRACT-GREEN: 001 Phase 16 generates closed required nine-field `ResumeSummary` list types；002 Phase 11 publishes summary-only list and full-detail get fixtures/examples/Prism/mock bytes；no compatibility schema or scenario exists.
  <!-- verified: 2026-07-14 evidence="closed nine-field generated ResumeSummary; 37 fixtures; Prism unit=5 and live=13/13 byte-equal; getResume remains full detail" -->
- [x] 7.3 BACKEND-GREEN: dedicated store list columns/record/service mapper/handler return summary facts without selecting detail payload；cursor/user isolation and full `getResume` lookup remain covered by focused and real PostgreSQL tests.
- [x] 7.4 FRONTEND-GREEN: inventory and migrate every `listResumes` consumer to generated `ResumeSummary` and backend `hasReadableContent`；list render/selection uses no removed field, local persistence or N+1 `getResume` fallback；navigation alone starts full detail fetch.
- [x] 7.5 AUDIT-GATE: 003 Phase 9 generates/exact-matches the declared OPENAPI-005 oracle from merge-base old baseline and preserves the audit before re-freeze；then current lint/fixture/codegen/diff and scoped negative searches pass.
  <!-- verified: 2026-07-14 evidence="OPENAPI-005 preserved exact 12/12 findings with zero errors; guarded re-freeze then diff/lint/fixtures/unit/codegen/negative gates PASS" -->
- [x] 7.6 BDD-N/A: 本 plan 不拥有真实 API/UI 用户流程；以 OpenAPI/fixture/codegen/drift gate 替代，不创建场景 wrapper。
- [x] 7.7 REGRESSION-GATE: 阶段收口从仓库根执行 `make test`，统一完成 backend 与 frontend 全量单元测试回归。
- [x] 7.8 CLOSEOUT-GATE: post-pass spec/plan/checklist/context/INDEX reconcile passes；all consumer owners are green in the same batch before restoring completed.
  <!-- verified: 2026-07-14 evidence="four B2 contexts validate; consumer-owner scenarios and final contract gates are green; batch Header/INDEX reconciliation completed" -->
