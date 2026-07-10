# 001 - Report Generation Baseline Checklist

> **版本**: 1.6
> **状态**: completed
> **更新日期**: 2026-07-10

**关联计划**: [plan](./plan.md)

## Completed Owner Gates

- [x] Contract and generated artifacts: B1/B2/B4/F3 truth sources expose current report errors, schema fields, fixtures, migration enum/columns, prompt/rubric keys, and generated artifacts.
  <!-- verified: 2026-05-16 method=contract-gates evidence="make codegen-check; make validate-fixtures; migrations/lint.sh; python3 scripts/lint/conventions_drift.py --repo-root .; prompt/rubric lint; migration and generated-contract tests." -->
- [x] Runtime wiring: backend report routes, runner kernel, report handler, report reaper behavior, and AI-backed service wiring are covered.
  <!-- verified: 2026-05-16 method=runtime-wiring evidence="go test ./cmd/api -run 'TestBuildReportRuntime' -count=1; go test ./cmd/api ./internal/review ./internal/store/review -count=1." -->
- [x] Report generation happy path: queued report job reaches ready report, writes assessments, outbox, audit, async job completion, and AI task rows.
  <!-- verified: 2026-05-16 method=happy-path-tests evidence="review/service/store tests and TestE2EP0052ReportGenerationHappyPath passed." -->
- [x] Report read path: ready, queued, generating, failed, pagination, invalid cursor, target ownership, and cross-user report isolation are covered.
  <!-- verified: 2026-05-16 method=read-handler-tests evidence="api/reports, review, store tests and TestE2EP0053ReportReadAndListing passed." -->
- [x] Failure and retry path: prompt, provider, timeout, invalid output, parse-empty, retry, and permanent-failure paths write failed report state and bounded AI task rows.
  <!-- verified: 2026-05-16 method=failure-tests evidence="TestGenerateReportFailedMatrix; TestRunnerRetryPolicyAndPermanentFail; TestPersistReportFailureRetryAndPermanent; TestE2EP0054ReportAIFailureAndRetry passed." -->
- [x] Privacy and observability: report JSON, assessment JSON, outbox payload, audit metadata, logs, metric labels, and AI task rows exclude raw QA, prompt, response, and secret material.
  <!-- verified: 2026-05-16 method=privacy-observability-tests evidence="redaction tests, outbox PII tests, metric-label allowlist tests, and TestE2EP0055ReportPrivacyAndOutOfScope passed." -->
- [x] Current-scope negative surface: backend-review implementation/runtime scope has no out-of-scope report module surface.
  <!-- verified: 2026-05-16 method=backend-review-out-of-scope-lint evidence="python3 scripts/lint/backend_review_out_of_scope.py --repo-root . --phase all; python3 -m pytest scripts/lint/backend_review_out_of_scope_test.py -q." -->
- [x] 2026-07-07 owner compression: plan/checklist/BDD/test/context docs describe the current report generation/read contract without stale baseline or staged implementation prose.
  <!-- verified: 2026-07-07 method=backend-review-001-owner-compression evidence="Updated backend-review spec.md to v1.3 and backend-review/001 owner docs to v1.2 completed. PASS: targeted stale-wording grep returned no matches; validate_context.py backend-review/001 backend PASS; sync-doc-index --fix-index updated docs/spec INDEX and backend-review plans INDEX; sync-doc-index --check PASS; make docs-check PASS; git diff --check PASS; make lint-core-loop-pruning-surface PASS real_residuals=0." -->
- [x] 2026-07-10 report status wording cleanup: queued/generating report reads are described as current status metadata rather than empty-report wording.
  <!-- verified: 2026-07-10 method=tech-debt-pruning evidence="OpenAPI descriptions, backend-review spec/plan/BDD, frontend-report-dashboard spec/plan, backend/frontend comments, generated artifacts and focused report tests were updated together." -->
- [x] 2026-07-10 test-only readiness validator removal: 删除 production `validReadinessTier`，property test 直接断言 shared 四值集合；验证：production `deadcode` RED/GREEN、review tests/staticcheck、owner docs gates。
  <!-- verified: 2026-07-10 method=test-only-readiness-validator-removal evidence="Production deadcode RED identified validReadinessTier as test-only. Deleted it and moved the four-value membership set into the randomized property test. Review package tests, staticcheck, reachability and symbol inventory PASS." -->

## Phase 9: GenerateHandler retryable-failure test harness consolidation

- [x] 9.1 Record scoped `internal/review` `dupl` RED and identify exact-name consumers.
  <!-- verified: 2026-07-10 method=review-generate-handler-retryable-test-dupl evidence="Scoped dupl -t 100 reports the two retryable GenerateHandler tests as internal/review's only clone group. BUG-0088 and backend-async-runner gates require TestGenerateHandler_NormalizesFinalizedRetryableFailureThroughKernel to remain callable." -->
- [x] 9.2 Keep both top-level tests and delegate their shared setup/assertions to one helper parameterized by incoming service finalization state.
  <!-- verified: 2026-07-10 method=review-generate-handler-retryable-test-helper evidence="Both exact tests pass and share one helper parameterized only by serviceFinalized. Each still asserts retryable AI_PROVIDER_TIMEOUT and handler output AsyncJobFinalized=false; scoped Review dupl is zero and staticcheck passes." -->
- [x] 9.3 Run exact BUG-0088 gate, Review owner/full backend, vet/staticcheck and owner/product/docs/pruning closeout gates.
  <!-- verified: 2026-07-10 method=review-generate-handler-retryable-test-closeout evidence="Both exact tests and the BUG-0088/async-runner five-test gate PASS; Review owner packages/full backend, go vet/staticcheck, Review/runner/product contexts and docs/index/diff/pruning gates PASS with real_residuals=0." -->

## Evidence Commands

```bash
cd backend && go test ./internal/api/reports ./internal/review ./internal/store/review ./internal/ai/aiclient ./internal/ai/registry ./cmd/api -count=1
cd backend && go test ./cmd/api -run 'TestE2EP0052|TestE2EP0053|TestE2EP0054|TestE2EP0055|TestBuildReportRuntime' -count=1
make codegen-check
make validate-fixtures
migrations/lint.sh
make lint-events
make codegen-events-check
python3 scripts/lint/conventions_drift.py --repo-root .
python3 scripts/lint/backend_review_out_of_scope.py --repo-root . --phase all
python3 -m pytest scripts/lint/backend_review_out_of_scope_test.py -q
python3 scripts/lint/prompt_lint.py
python3 scripts/lint/rubric_lint.py
make docs-check
git diff --check
```
