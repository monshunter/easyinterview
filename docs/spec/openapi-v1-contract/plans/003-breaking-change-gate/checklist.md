# OpenAPI v1 Contract Breaking-Change Gate Checklist

> **版本**: 1.14
> **状态**: completed
> **更新日期**: 2026-07-13

**关联计划**: [plan](./plan.md)

## Phase 1: baseline and diff entrypoint

- [x] `openapi/baseline/openapi-v1.0.0.yaml` represents current 37 operation / 10 tag freeze baseline（验证：`make openapi-diff` PASS）<!-- verified: 2026-07-10 method=make target=openapi-diff expected=37 baseline=37 current=37 -->
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

- [x] `make lint-openapi` passes with current 10 tags / 37 operations<!-- verified: 2026-07-10 method=make target=lint-openapi -->
- [x] `make validate-fixtures` passes
- [x] `make codegen-check` passes
- [x] `make openapi-diff` passes
- [x] context, INDEX and docs links pass（验证：`validate_context.py ...003...`, `sync-doc-index --check`, `make docs-check` PASS）

## Phase 5: OPENAPI-001 pre-release correction

- [x] 5.1 OWNER/GOVERNANCE-GATE: OPENAPI-001 is accepted, product-owner approval is recorded, B1 `REPORT_CONTEXT_TOO_LARGE_CONVENTIONS_PASS` exists, and spec/history/baseline README describe the same pre-release correction rule.
  <!-- verified: 2026-07-12 method=governance-preflight evidence="accepted OPENAPI-001 v1.1 records 方案 A approval; B1 marker exists; spec D-32, history 1.45 and baseline README align; baseline blob 352e7605a86ed85aa396446749bcf117dd70a200 is untouched" -->
- [x] 5.2 RED-GREEN: base-ref audit compares merge-base old baseline to proposed OpenAPI, fails without accepted ADR, and exact-matches severity/path/kind/before/after against OPENAPI-001 expected-findings JSON, including closed/constraint findings. Conditional finding encodes baseline source prohibition and derived retry/next non-null source-only branches. `REPORT_CONTEXT_TOO_LARGE` occurs exactly once as additive enum_value_added and never enters the breaking allowset. (openapi_diff unit + oracle fixture tests)
  <!-- verified: 2026-07-12 method=tdd-base-ref-oracle evidence="RED lacked decision/oracle normalizer and old current produced 36 missing findings. GREEN resolves merge-base baseline, requires accepted matching ADR, validates conditional source-only semantics and exact-matches 36 findings by five key fields; wrapper suite passes." -->
- [x] 5.3 Preserve the deterministic old-baseline JSON finding artifact before any baseline edit; simultaneous current/baseline replacement cannot satisfy this gate.
  <!-- verified: 2026-07-12 method=tracked-audit-artifact path=openapi/baseline/audits/OPENAPI-001-report-direct-semantics.json evidence="baselineSource is git:c3c9902a37b1aaefe0c4fb154296d711c8a6332d; findingCount=36; errors=[]; current baseline worktree has zero diff" -->
- [x] 5.4 方案 A 重新打开本 gate：expected finding 保持 `maxLength=200` code points；FeedbackReport ready/failed state closure与 current freeze 同步；重新生成 preserved old-baseline audit、re-freeze v1.0.0 并要求 clean `make openapi-diff`，随后独立执行 codegen-check。24/64、18/52、generation/judge max4与internal attempt audit均不进入OpenAPI finding；负向确认没有attempt/retry/progress字段或retry endpoint。旧合同clean-baseline PASS/sha不再是当前完成证据。
  <!-- verified: 2026-07-13 commands="make lint-openapi validate-fixtures openapi-diff; make codegen-check" result="preserved OPENAPI-001 audit regenerated from merge-base and remains exact at 36 findings; ready/non-ready/errorCode state conditions are in source, baseline and generated schemas; re-frozen v1.0.0 baseline matches current with zero findings; codegen byte-stable; no attempt/retry/progress wire surface or retry endpoint" -->
