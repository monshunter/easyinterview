# 001 - OpenAPI v1 Contract Bootstrap Checklist

> **版本**: 1.34
> **状态**: completed
> **更新日期**: 2026-07-19

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

## Phase 11: Practice round identity and progress projection

- [x] 11.1 RED: contract tests require `PracticeRoundRef`, `PracticeProgress`, optional `CreatePracticePlanRequest.roundId`, optional paired `PracticePlan.roundId/roundSequence`, and optional `TargetJob.practiceProgress`.<!-- verified: 2026-07-12 method=python-contract-red -->
- [x] 11.2 GREEN: update `openapi/openapi.yaml` and baseline, regenerate Go/TS artifacts, and keep the 37-operation / 10-tag inventory unchanged.<!-- verified: 2026-07-12 method=codegen+lint inventory="10 tags/37 operations" -->
- [x] 11.3 Verify additive compatibility with `make lint-openapi`, `make codegen-check`, `make openapi-diff`; no required request field, endpoint, status code, or existing property is removed.<!-- verified: 2026-07-12 method=openapi-diff evidence="HEAD breaking=0 additive=6; rebased baseline drift=0" -->
- [x] 11.4 Handoff generated types to backend-practice/backend-targetjob/frontend owners and run their focused contract tests.<!-- verified: 2026-07-12 method=go-compile+frontend-typecheck+practice-handler-tests -->

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

## Phase 12: OPENAPI-001 grounded direct report

- [x] 12.1 OWNER/GOVERNANCE-GATE: consume the B1 conventions gate (canonical literal/retryability only); 003 records accepted OPENAPI-001 and snapshots merge-base old baseline; baseline remains untouched.
  <!-- verified: 2026-07-12 method=owner-handoff evidence="B1 source-ready marker consumed; accepted OPENAPI-001/product approval verified; old baseline loaded from merge-base commit c3c9902a37b1aaefe0c4fb154296d711c8a6332d and worktree baseline blob remains untouched" -->
- [x] 12.2 RED-GREEN: edit proposed OpenAPI for summary + frozen context/hasNextRound + code/label/dimensionCode/retryFocusDimensionCodes and typed-object conditional oneOf CreatePracticePlanRequest. Matrix requires baseline-with-source fail; retry/next missing/null/blank/malformed source fail; retry/next with any extra fail; minimal retry/next `{goal,sourceReportId}` pass. Remove old fields/DimensionResult, close objects and enforce bounds. Sync only B1-sourced `REPORT_CONTEXT_TOO_LARGE` into `ApiErrorCode`. (inventory/schema/error-enum/conditional negative tests + `make lint-openapi`)
  <!-- verified: 2026-07-12 method=tdd-red-green evidence="RED failed on DimensionResult/old fields, derived minimal requests, baseline+source, unknown fields and maxLength. GREEN passes exact direct-report inventory, 12-case baseline/derived matrix, closed/bounds validator, Swagger validation and 10-tag/37-operation inventory with B1 error parity." -->
- [x] 12.3 AUDIT-GATE: with old baseline still untouched, 003 Phase 5.2 exact-matches old-baseline→proposed findings by severity/path/kind/before/after; the new error enum is additive-only and no codegen/fixture/baseline re-freeze occurs before PASS.
  <!-- verified: 2026-07-12 method=openapi-001-exact-set artifact=openapi/baseline/audits/OPENAPI-001-report-direct-semantics.json findings=36 breaking=33 additive=3 evidence="errorCount=0; REPORT_CONTEXT_TOO_LARGE appears once as additive enum_value_added; baseline file remains untouched" -->
- [x] 12.4 RED-GREEN: regenerate Go/TS artifacts; exact tests reject old names/compatibility aliases and assert CreatePracticePlanRequest is a typed struct/interface, never Go/TS `any`; derived-extra runtime fixtures fail. (`make codegen-openapi`; generator tests; `make codegen-check`)
  <!-- verified: 2026-07-12 method=generator-structure-red-green+isolated-index-codegen-check evidence="Consumer compilation exposed FeedbackReport incorrectly generated as PaginatedEnvelope because any allOf was classified as pagination. A focused struct/interface RED now locks direct fields; generator only classifies an explicit PaginatedEnvelope ref, Go/TS direct report types regenerate deterministically, generator tests pass, and isolated-index make codegen-check passes without treating expected current generated changes as drift." -->
- [x] 12.5 HANDOFF-GATE: 002 Phase 7 fixtures/prototype/Prism pass before backend-review/frontend consumers begin generated-shape implementation; baseline re-freeze remains deferred to 003 Phase 5.4.
  <!-- verified: 2026-07-12 method=owner-handoff evidence="002 Phase 7 passes fixture sync/validation/example projection and live Prism byte parity for both report operations plus createPracticePlan. Generated direct-shape compile drift was handed to backend-review/backend-practice and frontend owners; 003 Phase 5.4 remains intentionally unchecked and the old baseline remains untouched." -->
- [x] 12.6 CONTRACT-BOUNDARY-FOUNDATION: wire fuse 与 semantic/UX 职责已分离；旧具体边界由 12.7 当前方案 A 取代，本项不证明 200/24/64/18-52 已实现或 codegen 通过。
- [x] 12.7 A-GATE: `ReportNextAction.label`必须为1..200 code points；expected finding after=`minLength=1,maxLength=200`；FeedbackReport ready/failed/non-terminal conditional state machine必须拒绝 nullable-ready payload和非failed errorCode；OpenAPI/schema/generated/fixture同步并实际通过audit、re-freeze、openapi-diff与codegen-check。下游产品完整validator的24/64、18/52、sole-label/whole-report repair不改变OpenAPI finding，也不得替代codegen证据。
  <!-- verified: 2026-07-13 commands="make lint-openapi validate-fixtures openapi-diff; make codegen-check" result="37 operations/10 tags; 37 fixtures; ready requires non-null summary/preparedness/provenance and non-empty dimensions/actions; failed alone requires non-null errorCode; baseline diff breaking=0/additive=0; generated conventions/events/OpenAPI byte-stable; ReportNextAction.label minLength=1,maxLength=200" -->
- [x] 12.8 RESPONSIBILITY-NEGATIVE: generated/OpenAPI/fixtures contain zero `attemptCount|retryCount|repairReason|repairScope|generationProgress|retryReportGeneration` positive surfaces；max4/internal audit/status polling create no new expected finding，and maxAttempts49 exhaustion is not encoded as server failed.
  <!-- verified: 2026-07-13 evidence="OpenAPI inventory/diff/codegen and scoped negative search keep attempt/retry/repair/progress fields and retryReportGeneration endpoint absent; frontend maxAttempts49 remains local polling exhaustion only" -->

## Phase 13: OPENAPI-002 TargetJob paste-only contract

- [x] 13.1 OWNER/GOVERNANCE-GATE: consume accepted OPENAPI-002 v1.2 and its 17-finding exact oracle; preserve the merge-base old baseline before any OpenAPI/baseline mutation.
  <!-- verified: 2026-07-14 method=approved-design+merge-base-baseline-byte-proof mergeBase=2550c1b58472803755147dd648a4149632e86d8f sha256=e92eea4ca25618c9e6300b104ad6aea85b2b9e6094de8571349898bdecf29527 evidence="O-A corrects OPENAPI-002 to 17 authorized findings including two source-only ApiErrorCode removals; merge-base baseline and worktree baseline remain byte-identical before audit/re-freeze" -->
- [x] 13.2 RED: focused schema/inventory tests fail until `ImportTargetJobRequest` is closed, exactly requires `rawText,targetLanguage,resumeId`, and `rawText` has `minLength: 1` + `pattern: '\S'`; empty/space/tab/newline-only text plus old source wrapper/URL/file/manual-form/title/company/extra payloads, TargetJob source fields and `target_job_attachment` are rejected while both operation invariants and 37/10 inventory remain.
  <!-- verified: 2026-07-13 method=schema+generator-red evidence="inventory test failed on missing additionalProperties closure; generated type test failed on missing rawText; semantic-linter RED failed because the dedicated invariant did not exist" -->
- [x] 13.3 GREEN: update `openapi/openapi.yaml`, delete all `TargetJobImportSource*`, remove `TargetJob.sourceType/sourceUrl` and only the TargetJob upload purpose, regenerate typed Go/TS artifacts, and prove no union/discriminator/compatibility alias remains. (`make lint-openapi`; generator tests; `make codegen-openapi`; isolated pre-freeze `make codegen-check`)
  <!-- verified: 2026-07-13 commands="python3 -m unittest scripts.lint.openapi_inventory_test; go test ./backend/cmd/codegen/openapi -count=1; make lint-openapi; make codegen-openapi; isolated-index make codegen-check; validate_context openapi/contract" result="24 Python tests PASS; Go generator PASS; OpenAPI 3.1 valid; 37 operations/10 tags; typed flattened Go/TS request; source union/discriminator/TargetJob provenance/target_job_attachment absent; resume/privacy upload purposes retained; old baseline byte-unchanged" -->
- [x] 13.4 AUDIT/HANDOFF-GATE: 003 Phase 6 wrapper RED proves unconsolidated new-property constraints and a stale 15-finding oracle do not match, then GREEN exact-matches 17 findings with `rawText` constraints folded into `required_property_added.after` and both source-only `ApiErrorCode` removals present; audit passes before baseline edit. 002 canonical 422 fixture and mock runtime pass before frontend/backend consumers compile; re-freeze remains deferred until every owner gate is green.
  <!-- verified: 2026-07-14 method=tdd-exact-oracle+preserved-audit+mock evidence="RED: stale 15 tests failed against actual/expected 17. GREEN: focused exact-17 and stale-oracle-negative tests PASS; preserved OPENAPI-002 artifact replays merge-base 2550c1b with 17 breaking findings/errors=[]; 37 fixtures, 24 inventory tests and lint-mock-contract PASS. Baseline remains unchanged pending all-owner re-freeze." -->
- [x] 13.5 BDD-N/A/HANDOFF: downstream backend/frontend consumer tests cover paste submit, accepted/failed import and persisted readback using the flattened request; URL/file/manual-form positive assets are deleted, not treated as compatibility coverage.
- [x] 13.6 ZERO-REFERENCE-GATE: current OpenAPI/generated/positive-fixture/frontend/backend/mock/positive-scenario surfaces contain zero positive/runtime `TargetJobImportSource*|target_job_attachment|sourceType/sourceUrl|url/file/manual_form` TargetJob-import branches. ADR/oracle and exact negative test/fixture declarations are allowed; no whole-file/directory exclusion.
  <!-- verified: 2026-07-13 method=artifact-level-negative-search evidence="Current positive/runtime OpenAPI, generated, fixture, Home, TargetJob, mock and active scenario surfaces contain zero obsolete import branches; explicit negative gates remain; canonical success scenario is renamed paste-primary." -->

## Phase 14: Practice durable message recovery

- [x] 14.1 RED: schema/generator tests require a closed user/assistant discriminated union; user requires `clientMessageId + replyStatus(pending|retryable_failed|terminal_failed|complete)`, assistant forbids both, and generated Go/TS remains typed without `any`.
  <!-- verified: 2026-07-13 method=schema+generator-red evidence="Old generic PracticeMessage and generators failed the closed role union and no-any assertions before the union IR/template support was added." -->
- [x] 14.2 GREEN: update OpenAPI source and generated artifacts; backend-practice handoff persists user client ID/reply status and proves pending/retryable/terminal/complete transitions, unique reservation and at-most-one assistant.
  <!-- verified: 2026-07-13 method=openapi+codegen+real-postgres evidence="Strict generated Go/TS union compiles; API projection omits assistant recovery fields; real PostgreSQL proves user four-state atomic transitions, unique reservation and at-most-one assistant." -->
- [x] 14.3 REPLAY-GATE: `getPracticeSession` reload after AI failure returns the same user ID/status; only retryable failure accepts same-ID/same-text retry, complete replays existing result, mismatch/conflict return typed 409 without duplicates.
  <!-- verified: 2026-07-13 method=api+store+fixtures evidence="Authorized readback returns stable client ID/status; retryable same-ID/text CAS retries; complete replays without AI; pending/terminal/mismatch return typed 409 and create no duplicate rows." -->
- [x] 14.4 TS-ERROR-GATE: generated `ApiClientError(status,apiError)` exact tests cover valid JSON, non-JSON, empty, Abort and transport failures; non-JSON/empty/Abort/transport use `apiError=null`, raw body is not leaked and frontend consumers contain zero error-message parsing.
  <!-- verified: 2026-07-13 method=generated-client+consumer-tests evidence="Exact valid/non-JSON/empty/Abort/transport cases PASS; raw HTTP body is discarded, typed metadata is preserved, and Practice maps failures without parsing or displaying Error.message." -->
- [x] 14.5 HANDOFF-GATE: 002 and mock owners publish exact get/send recovery fixtures including planned validation/auth/not-found/conflict/mismatch/retryable cases; 003 preserves a separate D-35 Practice machine-oracle audit before baseline re-freeze. The oracle is only the executable projection of D-35 + history 1.54 + 方案 A, not a third `OPENAPI-NNN` ADR, and its findings never enter OPENAPI-002's exact 17 allowset.
  <!-- verified: 2026-07-14 method=machine-oracle-red-green evidence="D-35 focused 7 tests and full openapi_diff 41 tests PASS; the 11-finding oracle/audit retains 8 breaking plus 3 additive findings, and scoped OPENAPI-003 naming is zero-residual." -->
- [x] 14.6 BDD-N/A/HANDOFF: frontend-workspace-and-practice/002 and backend-practice/002 consumer tests cover optimistic row, pending lock/thinking, retryable-only affordance, reload recovery and same-ID retry without duplicates.

## Phase 15: OPENAPI-004 TargetJob report overview

- [x] 15.1 OWNER/RED: accepted OPENAPI-004 exists before schema mutation; tests reject cursor/pageSize, flat `PaginatedFeedbackReport`, TargetJob latest-report pointer and non-closed/missing overview fields while preserving endpoint and 37/10 invariants.
  <!-- verified: 2026-07-14 method=contract-red evidence="Accepted OPENAPI-004 is present; the new focused source/schema validator test fails first on legacy cursor/pageSize and also locks response, closed required summaries, failed-only errorCode, no full-report fields and 37/10 invariants." -->
- [x] 15.2 GREEN: add typed closed overview/current/latest schemas with explicit nullable required fields and failed-only errorCode; remove pagination/full-list/pointer surface and regenerate Go/TS without `any` or compatibility aliases.
  <!-- verified: 2026-07-14 method=openapi+codegen-green evidence="Focused overview source/semantic tests PASS, full inventory 27 tests and Go generator tests PASS, lint-openapi preserves 37/10, generated Go/TS are typed with no old aliases, and a second codegen run is byte-idempotent at sha 1262419c6ebe18e3f27c168b52fbf2764a369dfe. Nested PracticeRoundRef was RED-proven open then closed in place; overview rounds preserve the canonical 2..5 bound." -->
- [x] 15.3 AUDIT-GATE: 003 Phase 8 exact-matches old baseline → proposed OPENAPI-004 findings before baseline edit and preserves the deterministic artifact.
  <!-- verified: 2026-07-14 method=old-baseline-exact-audit evidence="Focused RED failed on the absent OPENAPI-004 normalizer/CLI; GREEN emits a deterministic zero-error 15-finding artifact (6 breaking, 9 additive). Full openapi_diff 48 tests PASS and live isolation keeps OPENAPI-002=17, D-35=11 and OPENAPI-004=15." -->
- [x] 15.4 HANDOFF-GATE: 002 canonical fixtures/Prism, db/targetjob zero-ref, backend-review real PostgreSQL selection, target-scoped ReportsScreen and Parse/Report/Generating zero-list-consumer gates pass before re-freeze.
- [x] 15.5 REGRESSION-GATE: `make lint-openapi validate-fixtures codegen-check openapi-diff` and scoped old-shape zero-reference pass after guarded re-freeze.
  <!-- verified: 2026-07-14 evidence="baseline/current sha256=6e81b656...9bc6; openapi-diff findings=0; lint=37 operations; fixtures=37; OpenAPI unit=122 PASS; codegen 69 files changed=0" -->

## Phase 16: OPENAPI-005 Resume list summary projection

- [x] 16.1 OWNER/RED: accepted OPENAPI-005 + spec/history 1.59 exist；schema/inventory/generator tests reject the old full-Resume list item, any non-nine-field summary, detail/provenance fields and extras while preserving list/get method/path/operationId/200 plus 37/10.
  <!-- verified: 2026-07-14 method=tdd-red evidence="Focused inventory, fixture, generator and diff-oracle tests first failed on the legacy full Resume list item, missing ResumeSummary, list provenance expectation and absent OPENAPI-005 oracle/normalizer; mutation tests cover missing/extra/wrong-type fields plus list/detail substitution." -->
- [x] 16.2 GREEN: add closed required `ResumeSummary` with nullable `summaryHeadline` and boolean `hasReadableContent`, switch only `PaginatedResume.items`, regenerate typed Go/TS `ResumeSummary[]` / `[]ResumeSummary`, and keep every `getResume`/mutation response on full `Resume` without alias or `any`.
  <!-- verified: 2026-07-14 method=openapi+codegen-green evidence="Closed exact nine-field ResumeSummary is generated as Go []ResumeSummary and TS ResumeSummary[] with typed nullable headline/boolean readability; only PaginatedResume.items changed while get/update/duplicate/archive remain full Resume. Full inventory 28 tests, generator package and lint-openapi PASS." -->
- [x] 16.3 AUDIT-GATE: 003 Phase 9 generates and exact-matches `decisions/OPENAPI-005-resume-list-summary.expected-findings.json` from merge-base old baseline before baseline mutation；the declared path is not treated as evidence until the JSON exists and passes the five-key audit.
  <!-- verified: 2026-07-14 method=old-baseline-exact-audit evidence="OPENAPI-005 exact-set oracle and preserved audit PASS with 12 findings (1 breaking, 11 additive), 37/10 plus list/get invariants, and zero errors; frozen baseline sha256 remains e92eea4ca25618c9e6300b104ad6aea85b2b9e6094de8571349898bdecf29527." -->
- [x] 16.4 HANDOFF-GATE: 002 Phase 11 fixture/example/Prism/mock, 004 Phase 7, backend dedicated list projection/service/handler and every frontend `listResumes` consumer pass without compatibility fields or N+1 `getResume` fallback.
  <!-- verified: 2026-07-14 evidence="Resume list fixture/Prism parity, scalar backend projection and generated ResumeSummary consumers PASS; no compatibility field or list-row getResume fallback remains" -->
- [x] 16.5 BDD-N/A/HANDOFF: backend/frontend consumer tests prove register/list summary projection, flat list/auth navigation and full detail fetch only after navigation.
- [x] 16.6 REGRESSION-GATE: 阶段收口从仓库根执行 `make test`。
- [x] 16.6 REGRESSION-GATE: after guarded re-freeze, `make lint-openapi validate-fixtures codegen-check openapi-diff`, downstream focused tests and scoped list-detail zero-reference gates all pass with current evidence.
  <!-- verified: 2026-07-14 evidence="guarded re-freeze is byte-equal; independent diff/lint/fixture/unit/Prism/codegen and downstream consumer gates PASS" -->

## Phase 17: RuntimeConfig content limits projection

- [x] 17.1 RED: schema/generator tests require exact five-field closed `ContentLimits`, positive int64, required RuntimeConfig field and no internal limits/any/optional fallback.
- [x] 17.2 GREEN: source/fixture/generated Go/TS/backend builder use small positive representative values；`ContentLimits` 与五个子字段 required，consumer 无 per-field fallback。
- [x] 17.3 HANDOFF: Resume/Home/Practice frontend consumers compile and use generated fields；仅整体 runtime source 不可用时保留既有 bootstrap fallback；report/HTTP/provider/profile values remain absent.
- [x] 17.4 BDD-N/A/SUBSTITUTE-GATE: Phase 17 不新增用户流程；以 closed-schema lint、fixture validation、Go/TS codegen drift、builder focused test、frontend compile/consumer test 与 internal-field negative search 收口，不运行或扩展任何 E2E 来证明配置传播。
  <!-- verified: 2026-07-14 evidence="OPENAPI-006 closed contract, fixture, generated, builder and consumer gates pass; configuration-only scenario dependency removed." -->
- [x] 17.5 REGRESSION: 37/10 inventory, lint/fixtures/Prism/codegen-check/openapi-diff, focused/full consumers, contexts/docs/diff pass.
  <!-- verified: 2026-07-14 evidence="Post-commit codegen-check is byte-stable; 37/10 lint, 37 fixtures, diff 0, 52 wrapper tests, consumer suites, 11 contexts and docs/diff gates pass." -->

## Phase 18: Report-owned conversation replacement

- [x] 18.1 OWNER/RED: OPENAPI-001 v1.7 + spec/history 1.61 exist；tests reject the public list path/query/schema and require the reportId-only protected replacement while preserving 37/10 plus start/get live-session invariants.
  <!-- verified: 2026-07-15 evidence="Focused inventory contract test failed against the old list surface, then passed with 37/10 and preserved start/get session operations." -->
- [x] 18.2 GREEN: add exact closed `ReportConversation` / `ReportConversationMessage`, replace `listPracticeSessions` with `getReportConversation`, and regenerate typed Go/TS without compatibility method, locator leakage, optional alias or `any`.
  <!-- verified: 2026-07-15 evidence="make lint-openapi plus focused inventory/generated Go+TS type test pass after codegen-openapi." -->
- [x] 18.3 AUDIT-GATE: 003 Phase 11 generates the exact five-key OPENAPI-001 v1.7 oracle from merge-base old baseline after RED and before baseline mutation；missing/extra/wildcard/hand-authored placeholder fails.
  <!-- verified: 2026-07-15 method=generated-base-ref-audit evidence="--emit-openapi-001-v17-oracle generated 15 findings from main merge-base; exact audit and preserved artifact match with errors=[]; focused wrapper test rejects wildcard/edited baseline." -->
- [x] 18.4 HANDOFF-GATE: 002 Phase 12 fixture/example/Prism, backend-practice list deletion, backend-review strict read model, frontend reportId-only consumer, mock parity and downstream BDD gates pass before re-freeze.
  <!-- verified: 2026-07-15 method=current-owner-handoff evidence="002 Phase 12, backend-practice 8.1-8.5, backend-review 12, frontend-report 13, mock 10 and scoped BDD owner gates PASS" -->
- [x] 18.5 ZERO-REFERENCE-GATE: current positive/runtime OpenAPI, fixture, generated, backend, frontend, mock and scenario surfaces contain zero `listPracticeSessions` / public session-list route/query/schema/consumer references；accepted decision/oracle and explicit negative tests are the only allowed occurrences.
  <!-- verified: 2026-07-15 method=structured-openapi+classified-production-search evidence="current positive/runtime operation/schema/fixture/generated/backend/frontend/mock/scenario surfaces zero; decision/oracle/negative tests classified separately" -->
- [x] 18.6 REGRESSION-GATE: `make lint-openapi validate-fixtures codegen-check openapi-diff`、scoped consumers、root `make test`、contexts/docs/diff all pass with 37/10 unchanged.
  <!-- verified: 2026-07-15 method=post-refreeze-independent-regression evidence="diff 0; lint/37 fixtures/codegen PASS; Python 7, Go 4 packages, frontend 22 focused PASS; root Python 551/4493, Go all, frontend 125/993; context/docs/diff PASS" -->

## Phase 19: OPENAPI-007 Settings UserContext pruning

- [x] 19.1 OWNER/RED: accepted OPENAPI-007 + spec/history 1.64 exist；focused schema/generated tests reject `uiLanguage/preferredPracticeLanguage/emailMasked` in `UserContext` and require complete `email`, while locking 37/10 and unchanged getMe/completeMyProfile/deleteMe method/path/operationId/status/security. Evidence (2026-07-15): the approved full-email correction makes the current masked source/generated contract RED.
- [x] 19.2 GREEN: source explicitly sets `UserContext.additionalProperties: false` and generated Go/TS expose exact required `{id,email,displayName,profileCompletionRequired}`；no optional alias/default/compatibility mapping, Settings receives the complete authenticated account email, and logs/E2E evidence remain redacted.
  <!-- verified: 2026-07-15 method=source-codegen-focused-contract evidence="source/embedded/Go/TS exact four-field contract PASS; 37/10 and protected Auth operation invariants unchanged" -->
- [x] 19.3 FIXTURE/CODEGEN-GATE: 002 Phase 13 Auth fixtures, embedded schema, dev mock and generated builders migrate together；fixture validation and `make codegen-check` pass.
- [x] 19.4 HANDOFF-GATE: backend-auth/001 Phase 10, frontend-shell/001 Phase 14, B4 001 Phase 13 and mock consumers compile/pass before baseline edit；Settings reuses runtime user without duplicate `getMe`.
- [x] 19.5 AUDIT/RE-FREEZE: 003 Phase 12 generates and exact-matches all 9 findings, including the `emailMasked` → `email` replacement and closed-object contract, from unchanged merge-base baseline；preserve the audit and re-freeze only after all consumers pass. Clean current diff alone is insufficient.
- [x] 19.6 BDD-HANDOFF/REGRESSION: reference Settings BDD + extended `E2E.P0.101` without creating B2 E2E；run lint/fixtures/codegen/diff, root `make test`, contexts/docs/diff and scoped old-field zero-reference gates.

## Phase 20: Failed report manual regeneration

- [x] 20.1 RED: inventory/schema/generated tests require exact protected POST path, operationId, required IK, body absence and `202 + ReportWithJob`; current inventory is 38/10.
  <!-- verified: 2026-07-16 method=focused-red evidence="inventory test failed because regenerateFeedbackReport was absent from EXPECTED_OPERATIONS before source mutation" -->
- [x] 20.2 GREEN: update OpenAPI source and regenerate Go/TS artifacts with `REPORT_INVALID_STATE_TRANSITION` parity and no attempt/progress fields.<!-- verified: 2026-07-16 method=openapi-codegen-green evidence="protected bodyless POST + IK + 202 ReportWithJob generated in embedded Go and typed TS clients; conventions parity and 30-test inventory PASS; scoped attempt/progress search empty" -->
- [x] 20.3 HANDOFF: 002 Phase 14, 003 Phase 13, backend-review and frontend-report-dashboard consume the exact operation matrix.<!-- verified: 2026-07-16 method=operation-handoff evidence="38/38 source/fixture/generated operation matrix; backend handler/store and frontend client/devMock consumers PASS; preserved D-40 audit exact" -->
- [x] 20.4 BDD-N/A/REGRESSION: contract gates plus downstream BDD references, `make lint-openapi validate-fixtures codegen-check openapi-diff` and root `make test` pass.

## Phase 21: OPENAPI-008 account theme and generic updateMe

- [x] 21.1 OWNER/RED: accepted OPENAPI-008 + spec/history 1.66；unchanged baseline reports exact 3 breaking + 4 additive findings and separately locks operationId rename with 38/10 Auth invariants.
- [x] 21.2 GREEN: source/generated Go+TS replace completeMyProfile/CompleteProfileRequest with closed updateMe/UpdateMeRequest and required closed displayPreferences.
- [x] 21.3 FIXTURE/CODEGEN: 002 Phase 15 replaces the fixture one-for-one；getMe returns theme；validators/Prism/dev mock/generated artifacts stay in parity.
- [x] 21.4 HANDOFF: backend-auth Phase 14、B4 Phase 14、frontend-shell focused tests prove atomic persistence, one bootstrap GET and one-save PATCH without follow-up GET.
- [x] 21.5 AUDIT/RE-FREEZE: preserve audit, re-freeze after consumer/migration gates, require zero diff and no old production operation/schema/fixture references.
- [x] 21.6 REGRESSION: lint/fixtures/codegen/diff、root `make test`、docs/context/diff and downstream BDD/E2E gates pass.

## Phase 22: OPENAPI-008 active spec review remediation

- [x] 22.1 RED/GREEN: focused owner-doc test fails on stale §4.2/handoff wording, then current schema inventory and owner association use only `UpdateMeRequest` / `updateMe` / required `displayPreferences`。（证据：RED 1 failed；GREEN focused PASS。）
- [x] 22.2 REGRESSION: OpenAPI inventory/fixture/codegen/diff、root tests、docs/index/context/diff 与 scoped old-positive reference classification pass before restoring `completed`。（证据：38 operations / 38 fixtures、codegen/diff 0 findings、root tests、docs/index/context/diff 与 scoped residual gate PASS。）
