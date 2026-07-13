# 001 - OpenAPI v1 Contract Bootstrap Checklist

> **版本**: 1.21
> **状态**: completed
> **更新日期**: 2026-07-13

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

- [x] 12.1 OWNER/GOVERNANCE-GATE: consume B1 `REPORT_CONTEXT_TOO_LARGE_CONVENTIONS_PASS` (canonical literal/retryability only); 003 records accepted OPENAPI-001 and snapshots merge-base old baseline; baseline remains untouched. OpenAPI parity is proved here, not required to emit the B1 marker.
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
