# OpenAPI v1 Contract Breaking-Change Gate

> **版本**: 1.21
> **状态**: active
> **更新日期**: 2026-07-14

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

本 plan 承接 [openapi-v1-contract spec](../../spec.md) 的 v1.0.0 freeze gate；Phase 10 是当前 OPENAPI-006 RuntimeConfig content-limits correction owner：

- `openapi/baseline/openapi-v1.0.0.yaml` 是当前 37 operation / 10 tag contract 的 baseline snapshot。
- `make openapi-diff` 比对 baseline 与 `openapi/openapi.yaml`，并按 B2 additive-only rules 拦截 breaking changes。
- wrapper rules reject endpoint deletion, method/path change, schema field deletion, type change, required-field addition and enum value removal.
- additive changes are allowed only when OpenAPI, fixtures, generated artifacts, inventory lint and spec records move together.
- `POST /api/v1/privacy/exports` P0 `501` to P1 `202` is the only response-status whitelist; it still requires a same-change contract record.
- breaking changes outside the whitelist require an accepted ADR before OpenAPI changes land.

本 plan 不实现 API handler、frontend consumer、fixtures or codegen outputs; those remain under their respective owner plans. It owns the local gate that prevents contract drift from passing silently.

## 2 当前合同

### 2.1 Gate Matrix

| Surface | Current contract | Verification |
|---------|------------------|--------------|
| Baseline snapshot | `openapi/baseline/openapi-v1.0.0.yaml` mirrors current freeze shape | `make openapi-diff` confirms baseline/current compatibility |
| Diff wrapper | `scripts/lint/openapi_diff.py` reclassifies raw diff output to B2 rules | wrapper unit tests + `make openapi-diff` |
| Diff config | `openapi/diff-config.yaml` pins 37 operation inventory and privacy export whitelist | `make openapi-diff`, openapi diff unit tests |
| ADR workflow | `docs/spec/openapi-v1-contract/decisions/TEMPLATE.md` defines accepted ADR shape for breaking changes | docs link checks + owner review |
| Baseline guide | `openapi/baseline/README.md` defines baseline selection and versioning rules | docs link checks + owner review |
| Contract handoff | B2 001/002/003 gates must pass before implementation owners consume generated clients | `make lint-openapi`, `make validate-fixtures`, `make codegen-check`, `make openapi-diff` |

### 2.2 Breaking / Additive Rules

Breaking:

- remove endpoint, method or path
- remove schema field
- change field type
- add a required request/response field
- remove enum value
- change response status semantics outside explicit whitelist

Additive:

- add endpoint or tag with updated inventory
- add optional schema field
- add enum value
- add optional query/header parameter
- add fixture scenario or example without changing existing semantics

### 2.3 Privacy Export Whitelist

The only status-transition whitelist is:

```yaml
path: /api/v1/privacy/exports
method: POST
from: "501"
to: "202"
```

The wrapper must verify the exact path, method and statuses. Any other status transition is breaking unless accepted through ADR and spec revision.

## 3 质量门禁分类

- **Plan 类型**: `contract + tooling + governance`。
- **TDD 策略**: 适用。Wrapper unit tests cover breaking/additive reclassification, composition schema diff, privacy export whitelist and contract-record requirements.
- **BDD 策略**: 不创建本地 BDD；本 plan 是内部 contract evolution tooling。涉及用户可见 correction 时，guarded re-freeze 必须等待业务 owner 的 BDD evidence；Phase 9 复用 P0.034/P0.036/P0.037。
- **替代验证 gate**:
  - `make openapi-diff`
  - `python3 -m unittest scripts.lint.openapi_diff_test`
  - `make lint-openapi`
  - `make validate-fixtures`
  - `make codegen-check`
  - `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/openapi-v1-contract/plans/003-breaking-change-gate/context.yaml --target contract`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `make docs-check`

## 4 实施步骤

### Phase 1: baseline and diff entrypoint

- Keep `openapi/baseline/openapi-v1.0.0.yaml` aligned to the current freeze baseline.
- Keep `make openapi-diff` wired through the repo Makefile.
- Support explicit baseline version selection when future baselines are added.

### Phase 2: ruleset and wrapper

- Enforce B2 breaking/additive rules through `scripts/lint/openapi_diff.py`.
- Detect composition changes under `oneOf` / `allOf` / `anyOf`.
- Keep privacy export whitelist exact and isolated.

### Phase 3: contract governance

- Maintain the ADR template for accepted breaking changes.
- Maintain baseline README guidance for baseline selection and versioning.
- Require contract records for response status transitions and baseline changes.

### Phase 4: closeout

- Run OpenAPI validation, fixture validation, codegen drift check and diff gate.
- Keep 001/002/003 B2 handoff green before downstream implementation plans consume generated clients.
- Keep product-scope pruning evidence tied to current OpenAPI truth source.

## 5 验收标准

| ID | 验收点 | 验证 |
|----|--------|------|
| A-1 | Current OpenAPI and baseline are compatible | `make openapi-diff` |
| A-2 | Breaking schema/endpoint changes fail | `python3 -m unittest scripts.lint.openapi_diff_test` |
| A-3 | Additive optional changes pass | `python3 -m unittest scripts.lint.openapi_diff_test` |
| A-4 | Privacy export whitelist is exact and isolated | `make openapi-diff`, wrapper unit tests |
| A-5 | Generated artifacts and fixtures remain current | `make lint-openapi`, `make validate-fixtures`, `make codegen-check` |

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-14 | 1.20 | Add Phase 9 for OPENAPI-005 exact Resume list-summary audit and all-consumer guarded re-freeze. |
| 2026-07-14 | 1.14 | Add OPENAPI-004 exact old-baseline audit and guarded report-overview re-freeze phase. |
| 2026-07-14 | 1.13 | Correct OPENAPI-002 from 15 to 17 exact findings and define the separate Practice machine oracle as a non-ADR projection of D-35/history 1.54/方案 A. |
| 2026-07-13 | 1.12 | Fix OPENAPI-002 proposed/audit order and exact invariants; add a separate Practice message recovery correction audit phase. |
| 2026-07-13 | 1.11 | Reopen Phase 6 for OPENAPI-002 exact TargetJob paste-only audit and guarded v1.0.0 re-freeze. |
| 2026-07-12 | 1.10 | Require the normalized conditional-contract finding to encode baseline source prohibition and derived non-null source-only branches. |
| 2026-07-12 | 1.9 | Exact-match OPENAPI-001 across severity and classify REPORT_CONTEXT_TOO_LARGE enum widening as additive-only. |
| 2026-07-12 | 1.8 | Reopen Phase 5 for OPENAPI-001 merge-base breaking authorization and guarded pre-release baseline re-freeze. |
| 2026-07-10 | 1.7 | Re-freeze the breaking-change baseline and diff inventory to the current 37-operation contract. |
| 2026-07-07 | 1.6 | Compress owner docs to the 2026-07-07 36-operation breaking-change gate contract. |
| 2026-05-04 | 1.5 | Add quality-gate classification for the completed breaking-change gate. |

## 6 Phase 5: OPENAPI-001 pre-release correction gate

### 5.1 Authorize before mutation

Require accepted `OPENAPI-001-report-direct-semantics.md` and B1 `REPORT_CONTEXT_TOO_LARGE_CONVENTIONS_PASS` before changing current OpenAPI. Capture the merge-base `openapi/baseline/openapi-v1.0.0.yaml` and compare it to the proposed new OpenAPI before baseline re-freeze.

### 5.2 Exact finding audit

Extend the wrapper with a base-ref mode that emits a deterministic normalized JSON artifact. Findings must exact-match `OPENAPI-001-report-direct-semantics.expected-findings.json` by `severity + JSON pointer + kind + before + after`; an unlisted/missing finding, severity drift, non-accepted ADR or missing spec/history increment fails. The synthetic conditional finding must encode baseline sourceReportId prohibition plus retry/next non-null source-only branches, not merely the existence of a `oneOf`. Closed objects and constraint tightening are audited even if the underlying diff library omits them. `REPORT_CONTEXT_TOO_LARGE` must appear exactly once as additive `enum_value_added`; treating it as breaking, informational or a wildcard authorization fails.

### 5.3 Guarded re-freeze

Only after 001 schema/codegen, 002 fixtures/prototype/Prism and downstream consumer gates pass may `openapi-v1.0.0.yaml` be re-frozen. Final verification requires both the preserved old-baseline finding artifact and a clean current-baseline `make openapi-diff`; the clean result alone is insufficient.

## 8 Phase 6: OPENAPI-002 pre-release paste-only gate

### 6.1 Authorize and snapshot before mutation

Require accepted [OPENAPI-002](../../decisions/OPENAPI-002-targetjob-paste-only.md) v1.2, product-owner approval, spec/history 1.56 and the machine oracle. Resolve the merge-base from `main` and snapshot the old `openapi/baseline/openapi-v1.0.0.yaml`; then let 001 update the proposed OpenAPI while the worktree baseline remains byte-unchanged. Only after the proposed source exists may this gate compare old baseline → proposed and persist the exact audit artifact.

### 6.2 Exact finding audit

Extend/reuse base-ref mode to exact-match `OPENAPI-002-targetjob-paste-only.expected-findings.json` by the full five-key tuple. The expected set is exactly 17 breaking findings. `rawText` is a new constrained required property: normalize its initial `minLength=1,pattern=\S` into the single `required_property_added.after` signature and do not emit a second constraint finding. The two source-only `ApiErrorCode` removals—`TARGET_IMPORT_SOURCE_INVALID` and `TARGET_IMPORT_SOURCE_UNAVAILABLE`—must each remain an independent `enum_value_removed` finding. Focused RED must prove both the unnormalized extra constraint and a stale 15-finding oracle fail exact-set; GREEN must equal 17. Missing/unexpected findings, severity/path/kind/before/after drift, wildcard authorizations, edited old baseline or a zero-finding simultaneous replacement must fail.

The audit must separately enforce unchanged invariants: 37 operations / 10 tags；`POST /api/v1/targets/import`, operationId `importTargetJob`, `202 + TargetJobWithJob`；and `POST /api/v1/uploads/presign`, operationId `createUploadPresign`, `201 + UploadPresign`, with remaining resume/privacy purposes.

### 6.3 Guarded consumer handoff and re-freeze

Do not re-freeze until 001 source/codegen, 002 fixtures/prototype/Prism, mock-contract-suite runtime, frontend Home, backend TargetJob/upload, persistence/event owners and P0.010/P0.015 all pass paste-only positive/negative gates. Require zero positive/runtime references for `TargetJobImportSource*`, `target_job_attachment`, TargetJob `sourceType/sourceUrl`, URL/file/manual-form import branches and compatibility aliases. Accepted ADR/oracle and exact negative declarations may retain rejected tokens; whole-file/test-directory exclusions are forbidden.

Preserve the deterministic old-baseline audit artifact under `openapi/baseline/audits/` before re-freeze. Final completion requires both that artifact and a clean current-baseline `make openapi-diff`, followed independently by fixture, codegen and consumer gates; clean current diff alone is never sufficient evidence.

## 9 Phase 7: Practice durable message recovery correction

### 7.1 Separate authority and audit

Treat spec D-35 / history 1.54 and the product-approved方案 A as the sole correction authority. The separate Practice machine oracle is only D-35's executable full five-key finding projection; it is not and must not require a third `OPENAPI-NNN` ADR. Snapshot the old baseline, keep it unchanged while 001 produces the role-discriminated Practice schema, and preserve the deterministic manifest before any re-freeze. Practice findings must be audited separately and must never be merged into OPENAPI-002's exact 17 allowset; missing or wildcard authorization fails.

### 7.2 Guarded re-freeze

Do not re-freeze the Practice correction until 001 typed schema/codegen/`ApiClientError` gates, 002 get/send recovery fixtures, mock-contract-suite parity, backend-practice persistence, frontend typed consumer and P0.046 reload/same-ID retry all pass. Final proof requires the preserved old-baseline artifact plus clean current diff and independent fixture/codegen/consumer gates.

## 10 Phase 8: OPENAPI-004 TargetJob report overview correction

### 8.1 Accepted authority and exact oracle

Require accepted OPENAPI-004 and spec/history 1.57 before proposed schema mutation. Snapshot the merge-base old baseline, keep the worktree baseline unchanged, and create a separate deterministic expected-findings oracle for only the report-overview correction. It must include removal of cursor/pageSize, response/full-pagination shape and TargetJob pointer plus the new required/closed schemas exactly as emitted by the normalizer; missing/extra/severity/path/kind/before/after drift or wildcard authorization fails.

### 8.2 Invariants and guarded re-freeze

Audit must prove 37 operations/10 tags and unchanged `GET /api/v1/targets/{targetJobId}/reports`, operationId `listTargetJobReports`, 200 response. Preserve the old-baseline artifact before mutation. Do not re-freeze until 001/002, migration/TargetJob, backend-review, target-scoped ReportsScreen/P0.059, Parse/Report/Generating zero-list-consumer and mock gates pass. Final proof requires preserved audit + clean current diff + independent lint/fixture/codegen/consumer gates; a clean diff alone is insufficient.

## 11 Phase 9: OPENAPI-005 Resume list summary correction

### 9.1 Accepted authority and RED oracle generation

Require accepted [OPENAPI-005](../../decisions/OPENAPI-005-resume-list-summary.md) and spec/history 1.59 before proposed schema mutation. Snapshot the merge-base old baseline and keep the worktree baseline unchanged. Phase 9 must generate `decisions/OPENAPI-005-resume-list-summary.expected-findings.json` from old baseline → proposed OpenAPI only after focused RED proves the old full-Resume list shape; the path declaration in the decision is not evidence and this document revision intentionally does not create the JSON.

The generated oracle must exact-match the response item ref change, new closed/required `ResumeSummary` schema and nullable/typed property constraints by `severity + path + kind + before + after`. Missing/extra finding, severity/path/kind/before/after drift, wildcard authorization, hand-authored placeholder, edited old baseline or simultaneous current/baseline replacement fails.

### 9.2 Invariants and guarded re-freeze

Audit separately locks 37 operations/10 tags；`GET /api/v1/resumes`, operationId `listResumes`, 200 `PaginatedResume` pagination envelope；and `GET /api/v1/resumes/{resumeId}`, operationId `getResume`, 200 full `Resume`. Preserve the deterministic old-baseline artifact before mutation. Do not re-freeze until 001 Phase 16, 002 Phase 11, 004 Phase 7, backend list projection, mock parity, every frontend consumer and P0.034/P0.036/P0.037 all pass without compatibility fields or N+1 detail fetch. Final proof requires preserved audit + clean current diff + independent lint/fixture/codegen/consumer gates.

## 12 Phase 10: OPENAPI-006 Runtime content limits

### 10.1 Accepted authority and exact oracle

Require accepted [OPENAPI-006](../../decisions/OPENAPI-006-runtime-content-limits.md), spec D-38 and history 1.60. Keep the merge-base baseline unchanged while 001 Phase 17 produces proposed `RuntimeConfig.contentLimits`. Exact-match required-property addition, closed required `ContentLimits`, its exact five positive-int64 properties and the runtime `$ref` by `severity + path + kind + before + after`; missing/extra/wildcard/type/minimum/required drift fails.

### 10.2 Invariants and guarded re-freeze

Audit preserves 37 operations/10 tags and unchanged `GET /runtime-config`, operationId `getRuntimeConfig`, 200 `RuntimeConfig`. Do not re-freeze until fixture/generated/backend builder, Resume/Home/Practice consumers and focused/full gates pass with report/HTTP/provider/profile limits absent from the public schema. Preserve the audit artifact before baseline mutation; final proof requires clean current diff plus independent lint/fixture/codegen/consumer gates.
