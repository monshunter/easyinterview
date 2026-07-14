# Grounded Conversation Report Test Plan

> **版本**: 2.21
> **状态**: active
> **更新日期**: 2026-07-14

## Phase 1-5: Historical baseline

- Existing conversation-level, read/replay, privacy and numeric-score boundary tests remain regression evidence; they do not prove the revised semantic design.

## Phase 6: Frozen context and contract

- Merge-base OpenAPI audit requires the accepted ADR finding set before current baseline re-freeze; closed-schema/codegen/fixture tests assert exact summary/context/code+label/dimensionCode/focus shape and reject old/additional fields.
- Migration up/down/up and DB integration prove `generation_context`, summary, dimension focus, current-shape reset policy, content-bearing placement and privacy cascade while explicitly proving `llm_attempt_count`/synonymous product retry columns are absent.
- `backend-practice/002` owns completion/store snapshot tests and publishes schema `practice-completion-evidence.v1`; this plan only validates and consumes its three markers.
- Review load tests mutate current TargetJob/Resume after the owner snapshot and prove report payload remains unchanged; count/last-seq mismatch fails and no mutable-entity fallback is called.
- Prompt payload tests prove trusted policy/untrusted JSON split and no raw context in job/outbox/audit/log/metric.
- The historical committed input-boundary fixtures are superseded by Phase 11 in-memory exact-payload construction. Current-schema worst-case zh/en JSON and its manifest remain; reconstruction drift fails before `REPORT_BOUNDARY_FIXTURES_READY` is emitted. Runtime boundary tests send in-memory exact payloads; oversized content makes zero provider/repair calls and persists REPORT_CONTEXT_TOO_LARGE. A3 separately proves provider context capacity and actual-token output fit before returning the 6,144-token profile marker.

## Phase 7: Direct semantics and reliability

- `maxLength=200` test is wire-fuse only。Semantic tests use English whitespace-word24/25 and zh-CN Unicode-code-point64/65；English delimiter parity additionally proves ECMAScript `/\s/u` U+FEFF splits while U+0085 does not。frontend over-limit is typed invalid/no raw；legal fixtures wrap on desktop+390。Targeted repair tests use the internal18/52 generation margin but acceptance remains200+24/64。
- Validator table uses action wire boundary200 with24/64 semantic and all existing cross-field/focus/action invariants。
- Generation recovery tests use an injected context-aware waiter and invocation-local counter to cover invalid→valid on attempts2/3/4, attempt4 invalid typed failure, dynamic scope/full revalidation, provider timeout/rate-limit/protocol handling, exact waits10s/20s/40s, cancellation and a second independent invocation reset. Runner/outbox schedules and async job attempts are negative dependencies, not product retry inputs.
- Persistence tests assert model-owned summary/assessments/evidence/actions/dimension focus/provenance and frozen public context are byte-semantically lossless and transactional while internal anchors stay out of API.
- Negative tests require zero positive active-contract hits for numeric score/average readiness/observation-count confidence/evidence-score conversion/default action plus `dimension_scores`, `retry_round`, `retryFocusCompetencyCodes`, `retry_focus_competency_codes`, `focusCompetencyCodes`, `focus_competency_codes`, `retryFocusTurnIds`, `retry_focus_turn_ids`, `questionAssessments`, `question_assessments` and `DimensionResult`; history/migrations/explicit negative fixtures are allowlisted.

## Phase 8: Replay, eval and UAT

- `backend-practice/004` integration owns generic empty-focus retry, issue-backed non-empty projection, cross-user/target/resume/round mismatch, non-ready source, next empty focus, server-derived settings/identity and idempotent replay; this plan only validates named owner markers.
- Frontend route/request negatives belong to frontend owner; review neither accepts copied create-plan fields nor validates client copies.
- Eval contract tests extend case schema, evalkit/Promptfoo and LLM judge request to pass original context + transcript + output plus dimension weight/ordered score levels; exact item verdict coverage includes preparedness and retry focus. Five distinct report cases cover all four readiness tiers, limited evidence, short answer, control-only+pending, and genuine-answer injection with fake role/schema/XML; separate negative fixtures cover contract drift.
- Product generation tests create a fresh in-memory retry context per `GenerateReport` invocation。RED/GREEN matrix covers initial+3、dynamic targeted→whole and whole→targeted scope across invalid rounds、success on attempts2/3/4、attempt4 terminal、retryable provider/protocol failures with exact10s/20s/40s waits、nonretryable immediate terminal、context cancellation、return-time destruction and second-invocation attempt1 reset。
- Producer/runner tests assert `async_jobs.attempts/max_attempts` are infrastructure-only and do not set, restore or consume product retry state; report product behavior no longer depends on explicit max_attempts4.
- DB integration interleaves attempt1 worker、reap、attempt2 takeover and delayed attempt1 success/retry/failure。Persistence receives job ID + claimed attempts；stale result/failure writes zero report/outbox/audit/job side effects, while no pre-call product reservation or `llm_attempt_count` increment exists.
- Evalkit tests keep independent in-memory generation/judge budgets=4，aggregate all usage/latency，and write bounded attempt_count/retry_count/reason/scope manifest entries。Generation reuses the full validator every round。Judge retries retryable provider and protocol/schema invalid，but a structurally valid unsupported/causal/zero-tolerance/critical verdict emits terminal typed content rejection with exactly one call。
- Product acceptance runs fixed complete/partial/short/pending-question/injection cases without replacement；all emitted final outputs must pass the mechanical contract and at least4/5 categories pass the existing per-sample judge thresholds/zero-tolerance rules。Strict P0.100 additionally repeats critical cases three times and requires11/11+blind review。Generic empty focus is accepted only for exact single `answer_depth` brief or single `answer_relevance` control-only issue；all other retries copy the full ascending unique needs-work issue-code set and reject subset/superset or `I >= 2` empty。For every selected focus code, the first retry label names at least one directly cited missing behavior；multi-focus uses one short semicolon-separated fragment per selected code；English action labels use at most 24 whitespace words and zh-CN labels at most 64 Unicode code points。Umbrella-only labels fail even when the judge score is high。review_evidence may revisit cited positive/explicit evidence-limit without inventing artifact/gap/new scenario/transfer task；next_round requires hasNextRound+permitted readiness。Any action type containing a mechanism, threshold, tool, sequence, framework or example absent from cited messages is unsupported rather than partial。
- Regression fixture preserves the real red shape as only `invalid_partial + $.nextActions[0]` and redacted structural coordinates; prompt/preflight and judge-instruction GREEN lock the exact no-uncited-specificity rule without retaining raw model output or reason prose.
- P0.100 product validation includes multi-issue empty focus、duplicate/incompatible action types and all current mechanical cross-field invariants before judge。Invalid output may consume up to remaining generation attempts；only a fully valid output enters judge。Failure evidence keeps only bounded redacted counts/reason/scope/digests，and the verifier rejects missing attempts、duplicate attempt numbers、retry_count drift or attempt_count>4。
- P0.100 product acceptance/strict diagnostic and P0.099 current-run audit remain independent；desktop+390 prove legal24/64 wrapping，200 schema PASS or18/52 repair margin is insufficient。Current run59381 satisfies product acceptance but remains strict FAIL。

## Backend Evidence Contract for P0.056 / P0.058

- P0.056 exact command: `cd backend && go test ./internal/review ./internal/store/review ./internal/api/reports -run '^TestE2EP0056ReportBackendEvidence$' -count=1 -v`.
- P0.058 exact command: `cd backend && go test ./internal/review ./internal/store/review ./internal/api/reports -run '^TestE2EP0058ReportFailureBackendEvidence$' -count=1 -v`.
- Each scenario writes `backend-evidence.json` in its `.test-output/e2e/<scenario>/` directory. P0.056 keeps exact top-level keys `schemaVersion`, `scenarioId`, `command`, `tests`, `consumedOwnerEvidence`, `markers`, `database`, `result` and schema `report-backend-evidence.v1`. P0.058 uses exact top-level keys `schemaVersion`, `scenarioId`, `command`, `tests`, `consumedOwnerEvidence`, `markers`, `database`, `runtime`, `result` and schema `report-backend-evidence.v3`. `database` contains only redacted report status/ready-column fail-closed facts; `runtime` contains provider-call counts, exact action waits, reset/destruction and async-attempt separation booleans.
- P0.056 required markers are `REPORT_COMPLETION_OWNER_EVIDENCE_CONSUMED_PASS`, `REPORT_DIRECT_READY_PASS`, `REPORT_FROZEN_CONTEXT_READ_PASS`, `REPORT_REVIEW_LEGACY_IDENTIFIER_NEGATIVE_PASS`. P0.058 requires `REPORT_CONTEXT_MISMATCH_FAIL_CLOSED_PASS`, `REPORT_CONTEXT_TOO_LARGE_PASS`, `REPORT_OUTPUT_RETRY_PASS`, `REPORT_FOUR_INVALID_FAIL_CLOSED_PASS`, `REPORT_ACTION_RETRY_RESET_PASS`, `REPORT_RETRY_LAYER_SEPARATION_PASS`.
- P0.058 `database` exact keys are `contextMismatchFailClosed`, `contextTooLargeStatus`, `fourInvalidStatus`, `failedReadyColumnsEmpty`. Its `runtime` exact keys are `contextTooLargeProviderCalls`, `outputRetryProviderCalls`, `fourInvalidProviderCalls`, `firstActionCallCount`, `secondActionInitialAttempt`, `retryStateDestroyedAfterAction`, `actionRetryScheduleSeconds`, `asyncAttemptsAffectProductAttempt`, `attemptFourTerminal`.
- Exact Go tests own log/DB marker production; the corresponding scenario `verify.sh` is the sole writer of each `backend-evidence.json` artifact after validating the complete marker set.
- `result=PASS` requires command exit 0, the exact test's `=== RUN` and `--- PASS:`, package `ok`, every required marker and DB case, schema-valid consumed P0.047 owner evidence, and zero `--- FAIL:`, package `FAIL`, `no tests to run`, raw cookie/JD/resume/transcript/prompt/output content. Frontend markers are composed separately and cannot replace backend PASS.

## Phase 10: Canonical-round report overview

- Contract tests reject pagination/full-report fields and assert the closed minimal objects: `round: PracticeRoundRef`, nullable `currentReport{id,generatedAt}` and nullable `latestAttempt{id,status,errorCode,createdAt}`.
- Store tests enumerate every canonical round in order and exercise empty、prior-ready+newer-failed、generating-only、latest-ready and deterministic tie cases. Assertions lock `currentReport` ordering to `generated_at DESC, created_at DESC, id DESC` and `latestAttempt` ordering to `created_at DESC, id DESC`.
- Failure/security tests cover hidden 404 plus invalid TargetJob summary、missing/invalid frozen context、row user/target/session mismatch、noncanonical round pair and ready-null-generatedAt. Every invalid case rejects the whole overview, makes no mutable/URL fallback call and leaks no partial identity.
- Consumer/negative tests prove only target-scoped ReportsScreen uses `listTargetJobReports`; Parse/Report/Generating do not, Report/Generating continue `getFeedbackReport(reportId)`, and active runtime/generated/fixtures contain no paginated full-list or TargetJob latest-report pointer semantics.

## Phase 11: Configured report input boundary

- Deterministic test helpers construct exact 62,397-byte regression, 917,504-byte limit and 917,505-byte limit+1 framed payloads in memory and prove canonical JSON round-trips; no `input-*.json` files are committed.
- Config injection tests cover default, legal override and invalid values; service tests prove limit calls provider unchanged while limit+1 makes zero provider/repair calls and persists `REPORT_CONTEXT_TOO_LARGE`.
- A3 capacity test proves `917504+2048+6144=925696<1000000`; no TPM arithmetic may satisfy the gate.
- P0.056 exercises the regression/default path; P0.058 exercises oversized terminal receipt and recovery.
