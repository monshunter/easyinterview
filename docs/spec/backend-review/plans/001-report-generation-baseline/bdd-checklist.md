# Grounded Conversation Report BDD Checklist

> **版本**: 2.16
> **状态**: completed
> **更新日期**: 2026-07-13

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.056 Frozen direct report

- [x] Setup validates/consumes P0.047 `completion-backend-evidence.json` schema `practice-completion-evidence.v1` with all three owner markers; it creates no duplicate completion path and records redacted correlation only.
- [x] Trigger runs `cd backend && go test ./internal/review ./internal/store/review ./internal/api/reports -run '^TestE2EP0056ReportBackendEvidence$' -count=1 -v` plus existing frontend Vitest files; backend log must contain the exact test RUN/PASS and `REPORT_COMPLETION_OWNER_EVIDENCE_CONSUMED_PASS`, `REPORT_DIRECT_READY_PASS`, `REPORT_FROZEN_CONTEXT_READ_PASS`, `REPORT_REVIEW_LEGACY_IDENTIFIER_NEGATIVE_PASS`.
- [x] Registry-owned `verify.sh` is the sole writer of `.test-output/e2e/p0-056-generating-to-report-happy-path/backend-evidence.json` with exact `report-backend-evidence.v1` keys `schemaVersion/scenarioId/command/tests/consumedOwnerEvidence/markers/database/result` after the complete marker set exists.
- [x] Verify sets `result=PASS` only when command exit is 0, exact backend/frontend markers and redacted DB assertions pass, and no FAIL/no-test/raw cookie/JD/resume/transcript/prompt/output appears; cleanup deletes rows and keeps only approved redacted artifacts.

## E2E.P0.058 Repair and fail closed

- [x] Setup validates/consumes the P0.047 owner artifact and prepares isolated missing/mismatched/48,001-byte context plus action-local attempts2/3/4, exact waiter, second-invocation reset, async-attempt separation and nonretryable cases without recreating completion ownership.
  <!-- verified: 2026-07-13 evidence="P0.058 setup PASS with isolated v3 owner cases" -->
- [x] Trigger runs `cd backend && go test ./internal/review ./internal/store/review ./internal/api/reports -run '^TestE2EP0058ReportFailureBackendEvidence$' -count=1 -v` plus frontend failure-state Vitest and emits the six v3 markers from test-plan.
  <!-- verified: 2026-07-13 evidence="exact backend test and seven frontend files/51 tests PASS; six v3 markers emitted" -->
- [x] Write `.test-output/e2e/p0-058-report-failure-and-missing-session/backend-evidence.json` with exact `report-backend-evidence.v3` top-level keys; `database` contains only status/ready-column fail-closed facts and `runtime` contains calls/waits/reset/separation facts.
  <!-- verified: 2026-07-13 evidence="v3 artifact separates database fail-closed facts from runtime calls/waits/reset/separation" -->
- [x] Verify requires exact RUN/PASS, every backend/frontend marker, failed-case ready-column null/empty assertions, zero FAIL/no-test/raw content and schema-valid consumed owner evidence before `result=PASS`; cleanup removes rows and keeps only approved redacted artifacts.
  <!-- verified: 2026-07-13 evidence="P0.058 verify and cleanup PASS with no false-PASS or raw-content finding" -->

## E2E.P0.070 Server-owned replay projection

- [x] Setup prepares ready owned source-report data with both empty generic-retry focus and issue-backed dimension focus plus frozen canonical successor.
- [x] Trigger/verify require generic-retry, issue-backed projection, server-derived request settings/identity, next-round-empty-focus and idempotency markers from the registry-owned P0.070 runner.
- [x] Cleanup removes source/derived plans, sessions and reports.

## E2E.P0.072 Replay isolation and failure

- [x] Setup prepares missing/cross-user/mismatch/non-ready/missing-context and unsupported/duplicate non-empty-focus data; empty focus remains a valid generic retry case in P0.070.
- [x] Trigger/verify require fail-closed and zero-leak markers from the registry-owned P0.072 runner.
- [x] Cleanup proves no derived plan/session was created.

## E2E.P0.099 Real full-stack report

- [x] Setup creates current-run en/zh ready rows for P0.099；each row must bind DB/API `canonical_report_content_digest`、`action_length_audit`、`content_audit`、`screenshot_sha256` and report/session/context digest。P0.100 output digest is not a prerequisite.
  <!-- verified: 2026-07-13 run="e2e-p0-099-20260713T095144Z-12381" evidence="two ready rows plus generating state bind current DB/API canonical report, screenshot, report/session and frozen-context digests; manual content audit closes fact-to-judgment-to-action" -->
- [x] Trigger/verify capture the exact six-image matrix；the 390x844 real report images fully cover action regions and show their actual `<=24-whitespace-word` English / `<=64-Unicode-code-point` zh-CN labels without clipping/ellipsis/hiding/horizontal overflow；manifest binds every screenshot to the consumed digest.
  <!-- verified: 2026-07-13 evidence="trigger+verify PASS; exact six redacted full-page images cover desktop and 390x844 for ready-needs-practice, ready-well-prepared and generating; raw-debug absent" -->
- [x] Deterministic ui-design/OpenAPI boundary fixtures use exactly 24 English whitespace words / 64 zh-CN Unicode code points and pass prototype/formal pixel parity with complete wrapping；this proof is separate from the real current-run six-image audit.
  <!-- verified: 2026-07-13 evidence="ui-design contract 54/54 and focused Playwright 34/34 PASS for exact en24/zh64 desktop/mobile wrapping" -->
- [x] Cleanup removes scenario data and preserves only the redacted manifest/screenshots/audit.
  <!-- verified: 2026-07-13 evidence="cleanup completed with redacted evidence retained; scenario privacy tests 38/38 and script tests 9/9 PASS" -->

## E2E.P0.100 Real-provider quality

- [x] Setup creates distinct complete/partial/short/pending-question/injection context+transcript inputs and expected forbidden/causal assertions.
- [x] Trigger product action-local generation：each independent invocation owns initial+up to3 retries and10s/20s/40s waits；per-round full validator selects action_labels/whole_report；attempt2/3/4 can recover；attempt4 ends that action；nonretryable zero retry；return destroys state and a second invocation resets.
  <!-- verified: 2026-07-13 evidence="P0.058 v3 supplies product action-local/reset proof; P0.100 current wording composes it without resampling" -->
- [x] Judge/Agent classify each fact/judgment/advice and repeat critical grounding/pending/injection cases three times；focused multi-code actions use one directly cited semicolon fragment per code, en<=24 whitespace words and zh-CN<=64 Unicode code points.
- [x] Trigger independent evalkit judge max4：provider retryable and protocol/schema invalid can recover；structurally valid unsupported/causal/zero-tolerance/critical negative produces typed terminal content rejection and exactly zero retry。
- [x] Verify generation/judge manifest chains contain redacted attempt_count/retry_count/reason/scope plus aggregate usage/latency，with no raw content/secrets and no duplicate/missing attempt index。
- [x] Verify async job attempts/max_attempts are infrastructure-only，while reap/takeover stale worker still has zero report/outbox/audit/job side effects without a pre-call product reservation.
  <!-- verified: 2026-07-13 evidence="generic producer max_attempts=5 and current PostgreSQL/race fencing regressions PASS" -->
- [x] Complete final-prompt product acceptance：all emitted finals mechanically valid and fixed-five semantic categories at least4/5；retain strict5-case/11-attempt+blind-review P0.100 as a separate diagnostic。200 fuse、18/52 margin or retry recovery are not content-quality evidence.
  <!-- verified: 2026-07-13 run="e2e-p0-100-20260713T101214Z-59381" evidence="mechanical9/9; semantic8/9; fixed categories4/5; strict P0.100 FAIL on unsupported summary; blind audit skipped; privacy cleanup PASS" -->
- [x] Preserve focused run evidence without promoting it：80338 attempt11 full-validator escape is historical and accepted by the fresh full-validator PASS；59906 direct injection3x and75753 same-digest+5 generic-empty-focus regressions remain focused evidence rather than matrix evidence。
- [x] Record `e2e-p0-100-20260713T022140Z-25849` as aborted after10/11 due max4 contract replacement；never count it as PASS or marker evidence；the fresh current-contract run supersedes it.
- [x] Record `e2e-p0-100-20260713T030100Z-35622` as aborted after7/11 due L2 job/fencing/frontend-resume findings；never count it as PASS or marker evidence；the fresh post-L2 run supersedes it.
- [x] Verify fails on product action attempt5、missing/wrong10s/20s/40s waits、failed reset、async/product attempt coupling、stale-worker mutation、invalid output、wrong scope、non-label mutation、valid-negative judge retry、fabrication or threshold miss；cleanup retains only redacted audit results.
  <!-- verified: 2026-07-13 evidence="P0.058 v3 negatives plus current P0.100 validators; final run59381 retains redacted strict FAIL rather than resampling" -->
