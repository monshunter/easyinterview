# OpenAPI v1 Contract Breaking-Change Gate Checklist

> **版本**: 1.23
> **状态**: active
> **更新日期**: 2026-07-15

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

- [x] 5.1 OWNER/GOVERNANCE-GATE: OPENAPI-001 is accepted, product-owner approval is recorded, B1 conventions gate passes, and spec/history/baseline README describe the same pre-release correction rule.
  <!-- verified: 2026-07-12 method=governance-preflight evidence="accepted OPENAPI-001 v1.1 records 方案 A approval; B1 marker exists; spec D-32, history 1.45 and baseline README align; baseline blob 352e7605a86ed85aa396446749bcf117dd70a200 is untouched" -->
- [x] 5.2 RED-GREEN: base-ref audit compares merge-base old baseline to proposed OpenAPI, fails without accepted ADR, and exact-matches severity/path/kind/before/after against OPENAPI-001 expected-findings JSON, including closed/constraint findings. Conditional finding encodes baseline source prohibition and derived retry/next non-null source-only branches. `REPORT_CONTEXT_TOO_LARGE` occurs exactly once as additive enum_value_added and never enters the breaking allowset. (openapi_diff unit + oracle fixture tests)
  <!-- verified: 2026-07-12 method=tdd-base-ref-oracle evidence="RED lacked decision/oracle normalizer and old current produced 36 missing findings. GREEN resolves merge-base baseline, requires accepted matching ADR, validates conditional source-only semantics and exact-matches 36 findings by five key fields; wrapper suite passes." -->
- [x] 5.3 Preserve the deterministic old-baseline JSON finding artifact before any baseline edit; simultaneous current/baseline replacement cannot satisfy this gate.
  <!-- verified: 2026-07-12 method=tracked-audit-artifact path=openapi/baseline/audits/OPENAPI-001-report-direct-semantics.json evidence="baselineSource is git:c3c9902a37b1aaefe0c4fb154296d711c8a6332d; findingCount=36; errors=[]; current baseline worktree has zero diff" -->
- [x] 5.4 方案 A 重新打开本 gate：expected finding 保持 `maxLength=200` code points；FeedbackReport ready/failed state closure与 current freeze 同步；重新生成 preserved old-baseline audit、re-freeze v1.0.0 并要求 clean `make openapi-diff`，随后独立执行 codegen-check。24/64、18/52、generation/judge max4与internal attempt audit均不进入OpenAPI finding；负向确认没有attempt/retry/progress字段或retry endpoint。旧合同clean-baseline PASS/sha不再是当前完成证据。
  <!-- verified: 2026-07-13 commands="make lint-openapi validate-fixtures openapi-diff; make codegen-check" result="preserved OPENAPI-001 audit regenerated from merge-base and remains exact at 36 findings; ready/non-ready/errorCode state conditions are in source, baseline and generated schemas; re-frozen v1.0.0 baseline matches current with zero findings; codegen byte-stable; no attempt/retry/progress wire surface or retry endpoint" -->

## Phase 6: OPENAPI-002 TargetJob paste-only correction

- [x] 6.1 OWNER/GOVERNANCE-GATE: OPENAPI-002 v1.2 is accepted, product-owner approval and spec/history 1.56 are recorded; capture merge-base old baseline, then update proposed OpenAPI while worktree baseline remains byte-unchanged, then audit old → proposed.
- [x] 6.2 RED-GREEN: base-ref audit folds new `rawText` minLength/pattern into `required_property_added.after` and exact-matches 17 findings, including independent removal findings for `TARGET_IMPORT_SOURCE_INVALID` / `TARGET_IMPORT_SOURCE_UNAVAILABLE`; RED proves an extra constraint finding and a stale 15-finding oracle fail. Missing/unexpected/drifted finding, wildcard, edited baseline and simultaneous zero-finding replacement all fail.
- [x] 6.3 INVARIANT-GATE: audit proves 37/10 plus exact method/path/operationId/status/response for `importTargetJob` and `createUploadPresign`, with resume/privacy purposes retained.
  <!-- verified: 2026-07-14 artifact="OPENAPI-002-targetjob-paste-only.json" findings=17 expected=17 errors=0 invariants="37 operations/10 tags" -->
- [x] 6.4 Preserve deterministic old-baseline JSON finding artifact before baseline mutation; 001/002/mock/frontend/backend/persistence/event paste-only and scoped zero-reference gates all pass.
- [x] 6.5 Re-freeze `openapi-v1.0.0.yaml` only after 6.4; require clean current-baseline `make openapi-diff`, then independently run lint, fixture, codegen and downstream consumer gates. Historical clean PASS is not current evidence.
  <!-- verified: 2026-07-14 evidence="baseline/current sha256=6e81b656...9bc6; independent diff=0, lint=37, fixtures=37, unit=122 and codegen changed=0 PASS" -->

## Phase 7: Practice durable message recovery correction

- [x] 7.1 GOVERNANCE/RED: spec D-35/history 1.54 and the product-approved方案 A are the sole authority；snapshot old baseline and fail until a separate deterministic five-key Practice machine oracle exists. The oracle is D-35's executable projection, not a third `OPENAPI-NNN` ADR. Never add Practice findings to OPENAPI-002's exact 17.
- [x] 7.2 AUDIT-GATE: old baseline → proposed role-discriminated message schema exact-matches the Practice manifest while baseline remains unchanged; missing/extra/wildcard findings fail.
  <!-- verified: 2026-07-14 artifact="D-35-practice-durable-recovery.json" findings=11 expected=11 errors=0 isolation="OPENAPI-002 remains exact 17" -->
- [x] 7.3 HANDOFF-GATE: 001 schema/codegen/typed `ApiClientError`, 002 fixtures, mock runtime, backend persistence and frontend typed consumer reload/same-ID retry tests pass before re-freeze.
- [x] 7.4 RE-FREEZE: preserve the old-baseline artifact, then re-freeze and independently run current diff, lint, fixture, codegen and consumer gates.
  <!-- verified: 2026-07-14 evidence="preserved D-35 audit retained; guarded re-freeze and independent contract/codegen/consumer gates PASS" -->

## Phase 8: OPENAPI-004 TargetJob report overview correction

- [x] 8.1 GOVERNANCE/RED: accepted OPENAPI-004 + spec/history 1.57 exist; old baseline snapshot is byte-stable; wrapper fails until a separate exact five-key report-overview oracle exists.
- [x] 8.2 AUDIT-GATE: old baseline → proposed schema exact-matches cursor/pageSize/flat response/TargetJob pointer removals and new closed required overview schemas; missing/extra/wildcard/drift fails.
- [x] 8.3 INVARIANT-GATE: 37/10 and exact listTargetJobReports method/path/operationId/200 remain unchanged.
  <!-- verified: 2026-07-14 artifact="OPENAPI-004-targetjob-report-overview.json" findings=15 expected=15 errors=0 invariants="37 operations/10 tags" -->
- [x] 8.4 HANDOFF-GATE: 001/002, db/targetjob, backend-review, target-scoped ReportsScreen, Parse/Report/Generating zero-list-consumer and mock gates pass before baseline edit.
- [x] 8.5 RE-FREEZE: preserve old-baseline artifact, re-freeze, then independently run current diff/lint/fixture/codegen/consumer and old-shape zero-reference gates.
  <!-- verified: 2026-07-14 evidence="preserved OPENAPI-004 audit retained; guarded re-freeze plus independent diff/lint/fixture/codegen/consumer gates PASS" -->

## Phase 9: OPENAPI-005 Resume list summary correction

- [x] 9.1 GOVERNANCE/RED: accepted OPENAPI-005 + spec/history 1.59 exist；snapshot merge-base old baseline and fail while proposed `PaginatedResume.items` still references full `Resume` or the summary is not exact/closed/required.
  <!-- verified: 2026-07-14 method=audit-red evidence="Focused audit tests failed first on the absent OPENAPI-005 normalizer/oracle and contract tests reject the old full-Resume ref plus non-exact/open/optional summary shapes against the immutable merge-base baseline." -->
- [x] 9.2 ORACLE-GATE: generate `decisions/OPENAPI-005-resume-list-summary.expected-findings.json` during this Phase from old baseline → proposed schema；reject a missing file, placeholder/hand-authored wildcard, missing/extra finding or any five-key drift. The path declaration alone is not PASS evidence.
  <!-- verified: 2026-07-14 method=exact-set-oracle evidence="Generated OPENAPI-005 exact-set oracle and preserved audit contain the same 12 five-key findings; missing/unexpected findings fail, order is insignificant, and full openapi_diff suite 50 tests PASS." -->
- [x] 9.3 INVARIANT-GATE: audit preserves 37/10, listResumes method/path/operationId/200/pagination and getResume method/path/operationId/200 + full `Resume`.
  <!-- verified: 2026-07-14 method=invariant-audit evidence="Audit reports one intentional PaginatedResume item-ref break plus eleven ResumeSummary additions, with 37 operations/10 tags and exact list/get methods, paths, operationIds, 200 responses, parameters and full getResume response unchanged." -->
- [x] 9.4 HANDOFF-GATE: 001/002/004, backend list projection, mock and all frontend list/detail consumer tests pass without compatibility fields or N+1 detail fetch before baseline edit.
- [x] 9.5 BDD-N/A/REGRESSION: 本 plan 不拥有真实 API/UI 流程；阶段收口从仓库根执行 `make test`。
- [x] 9.5 RE-FREEZE: preserve the deterministic old-baseline artifact, then re-freeze and independently run current diff/lint/fixture/codegen/downstream gates；clean current diff alone is insufficient.
  <!-- verified: 2026-07-14 artifact="OPENAPI-005-resume-list-summary.json" findings=12 expected=12 errors=0 evidence="baseline/current sha256=6e81b656...9bc6; independent gates PASS" -->

## Phase 10: OPENAPI-006 Runtime content limits

- [x] 10.1 GOVERNANCE/RED: accepted OPENAPI-006 + spec D-38/history 1.60 exist；merge-base baseline stays unchanged and generic diff reports the authorized required-field break before re-freeze.
- [x] 10.2 ORACLE-GATE: exact five-key oracle covers required `RuntimeConfig.contentLimits`, closed required `ContentLimits`, exact five positive-int64 properties and runtime `$ref`; missing/extra/wildcard/type/minimum/required drift fails.
- [x] 10.3 INVARIANT/HANDOFF: 37/10 and getRuntimeConfig method/path/operationId/200 stay unchanged；fixture/generated/backend builder/Resume/Home/Practice pass and internal limits remain absent.
  <!-- verified: 2026-07-14 evidence="Preserved OPENAPI-006 audit has 9 exact findings/errors=[]; current diff 0 and 52 wrapper tests pass." -->
- [x] 10.4 RE-FREEZE: preserve the audit artifact, re-freeze, then independently run current diff/lint/fixture/codegen/consumer/docs gates.
  <!-- verified: 2026-07-14 evidence="OPENAPI-006 preserved audit contains 9 exact findings/errors=[]; re-frozen current diff is 0 and independent lint/fixture/wrapper/consumer/docs plus post-commit codegen-check pass." -->

## Phase 11: OPENAPI-001 v1.7 report conversation correction

- [x] 11.1 GOVERNANCE/RED: accepted OPENAPI-001 v1.7 + spec/history 1.61 exist；snapshot merge-base old baseline and fail while public `listPracticeSessions` remains or the report-owned replacement is absent/non-closed.
  <!-- verified: 2026-07-15 method=tdd-red evidence="v1.7 normalizer/oracle entrypoint was absent; focused contract tests reject restored public GET and internal message locator; merge-base baseline b5f99de3 remains byte-stable" -->
- [x] 11.2 ORACLE-GATE: generate the expanded OPENAPI-001 expected-findings JSON during implementation after RED；exact-match every five-key deletion/addition finding and reject placeholder, wildcard, missing/extra/drift or edited old baseline.
  <!-- verified: 2026-07-15 method=generated-oracle evidence="--emit-openapi-001-v17-oracle generated 15 five-key findings; OPENAPI-001-report-conversation.expected-findings.json exact-matches; wrapper test rejects edited worktree baseline" -->
- [x] 11.3 INVARIANT-GATE: preserve 37/10 and exact start/get live-session operations；lock protected getReportConversation method/path/operationId/200 while rejecting public list compatibility.
  <!-- verified: 2026-07-15 method=base-ref-audit evidence="OPENAPI-001 v1.7 audit PASS: 37 operations, 10 tags, unchanged start/get Practice operations, protected Reports GET, 2 breaking + 13 additive findings, errors=[]" -->
- [x] 11.4 HANDOFF-GATE: 001/002, backend-practice, backend-review, frontend-report, mock and downstream BDD gates pass before baseline edit with zero positive/runtime list surface.
  <!-- verified: 2026-07-15 method=current-handoff+base-ref-audit evidence="001 18.1-18.5, 002 Phase 12, practice/review/frontend/mock/BDD owners PASS; OPENAPI001V17ConversationTests 4 PASS; positive/runtime list surface zero" -->
- [x] 11.5 RE-FREEZE: preserve deterministic old-baseline artifact, re-freeze, then independently run current diff/lint/fixture/codegen/downstream/root-test/docs gates；clean current diff alone is insufficient.
  <!-- verified: 2026-07-15 method=guarded-refreeze+independent-gates evidence="preserved audit old baseline b5f99de3/15 findings/errors=[] unchanged; baseline/current sha256=88d77c6a; diff 0; lint/fixtures/codegen/consumers/root/context/docs/diff PASS" -->

## Phase 12: OPENAPI-007 Settings UserContext pruning

- [ ] 12.1 GOVERNANCE/RED: accepted OPENAPI-007 + spec/history 1.63 exist；merge-base baseline remains byte-stable and focused diff proves the two old UserContext required properties before any baseline edit.
- [ ] 12.2 ORACLE-GATE: generate exact five-key findings from old baseline → proposed source for both old fields leaving required/properties plus `additionalProperties: false` closure；reject missing/extra/wildcard/placeholder, edited baseline, optional aliases or constant-filled compatibility fields.
- [ ] 12.3 INVARIANT-GATE: preserve 37/10 and exact getMe/completeMyProfile/deleteMe method/path/operationId/status/security；lock four-field closed required UserContext, email masking and profile completion.
- [ ] 12.4 HANDOFF-GATE: 001/002, backend-auth, frontend-shell, B4 migration, mock and downstream Settings BDD pass before baseline edit with production old-field references at zero.
- [ ] 12.5 RE-FREEZE/REGRESSION: preserve audit, re-freeze, then independently run diff/lint/fixture/codegen/consumer/migration/root-test/contexts/docs gates；clean current diff alone is insufficient.
