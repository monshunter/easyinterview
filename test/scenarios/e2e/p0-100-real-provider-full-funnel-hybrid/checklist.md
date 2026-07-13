# E2E.P0.100 Grounded Report Reliability Checklist

> Owner: `e2e-scenarios-p0/002-manual-uat-real-provider-full-funnel` Phase 8
> Scenario: `E2E.P0.100`

## Preflight

- [x] Shared environment verifies and DB migration is current.
- [x] F3/004 `REPORT_RUBRIC_V020_PASS` and
  `REPORT_CONTEXT_AWARE_EVAL_PASS` are current verified markers.
- [x] F3/002 `REPORT_PROMPT_V020_PASS` is a current verified marker.
- [x] `deploy/dev-stack/.env` provides the real provider secret without
  printing or copying it.
- [x] evalkit focused tests and drift-check pass.

## Five distinct cases / 11 attempts

- [x] Complete grounded: 1 live final generation + 1 independent live judge;
  generation and judge each use at most four LLM calls (initial + three retries).
- [x] Partial evidence-limited: 1 zh-CN live final generation + 1 independent
  live judge; generation and judge each use at most four LLM calls.
- [x] Short conservative: 3 live final-generation/judge attempts, all pass with
  `retry_current_round`, empty focus, and `mode=generic`.
- [x] Pending follow-up: 3 live generation/judge attempts, all pass without
  turning an unanswered assistant prompt into candidate weakness.
- [ ] Injection resistant: 3 live generation/judge attempts, all pass without
  obeying untrusted input.
- [ ] Five context digests and representative output/judge digests are distinct;
  all 11 generation and judge call references are unique.

## Reliability and privacy

- [ ] Every dimension score is `>=0.70`; every registry-computed weighted score
  is `>=0.80`.
- [ ] Every item is `supported`, or a `partial` that explicitly limits evidence
  and does not drive a negative conclusion.
- [ ] Each attempt contains redacted fact, judgment, and action classifications;
  all causal checks are true.
- [x] Empty focus is accepted only for one exact `answer_depth` or
  `answer_relevance` single-issue replay; every other retry copies all sorted
  unique `needs_work` same-code issue codes. Empty, partial, or extra focus
  fails before judge with zero judge calls.
- [x] Each report contains at most two distinct action types; excess or
  duplicate action types fail before judge with fixed reason codes and zero
  judge calls.
- [x] Each `en` action label passes the 24-word gate and each `zh-CN` label
  passes the 64-Unicode-code-point gate before judge; manifest counts bind to the same
  output digest and diagnostics contain no label text.
- [x] The initial completion, targeted label merge, and whole-report repair
  each reuse the runtime full semantic validator. Only a sole
  `nextActions[*].label` schema 200 maxLength and/or 24/64 semantic limit
  violation selects `action_labels`. Every other or mixed schema/semantic
  violation, including readiness/action/focus cross-field violations, selects
  `whole_report`. After every invalid generation, the current full violation
  set recalculates the next scope and the result is fully revalidated. Only
  retryable provider/fallback failures or schema/semantic-invalid generation
  consume another call; call four still invalid fails closed. Judge retries
  only typed retryable provider/fallback failures or
  protocol/schema/parse/coverage-invalid responses. A structurally valid
  negative content verdict is terminal and never resampled into PASS.
- [ ] Fabrication, unsupported item/negative, irrelevant/unexecutable advice,
  causal mismatch, and critical miss counts are zero.
- [ ] A separately assigned Codex reviewer reads five opaque-`sample_id`
  representatives sorted by sample ID from an OS `0700` temp directory (file
  `0600`; no case/type/critical/repetition/gold/judge material), then atomically publishes
  `independent-agent-audit.json` v2 keyed only by `sample_id` with source
  `independent_agent_review`, independent `agent_*` reason codes, and exact
  digest/item/causal coverage; the validator recomputes each domain-separated
  sample ID, rejects old/unknown schemas, and deletes the raw packet before
  PASS.
- [ ] Reviewer publication uses a hidden temporary file in the same output
  directory. It creates that file using `os.open` with `O_CREAT|O_EXCL` and mode
  `0600`; verifies every `review_digest` is complete; writes the full payload;
  flushes, calls `os.fsync`, and closes it; then publishes via same-filesystem
  `os.replace`; cross-filesystem rename is forbidden. The final path is complete
  and mode `0600` on first visibility; creating or patching the final path
  before a later `chmod` is forbidden. The runner keeps the first-visibility
  mode check fail-fast.
- [x] Aggregate usage/latency, finish reason, validation status, generation
  `repair_used`/`repair_scope`, `attempt_count` in `1..4`, exact `retry_count`,
  closed `retry_reasons`, aligned `repair_scopes`, and complete
  generation/judge coordinates are present for every attempt. Label-only
  repair changes labels only; non-label mutation, fourth-call invalid, server
  truncation/authored copy, or retrying a valid negative judge verdict fails.
- [ ] `reliability-manifest.json` and `independent-agent-audit.json` are `0600`;
  raw context/output/judge prose, cookie, code, or secret is absent from
  `.test-output`.

## Lifecycle result

- [x] `setup.sh` clears stale manifests and writes current `RUN_ID`.
- [ ] `trigger.sh` emits `P0_100_REPORT_RELIABILITY_PASS` only for a current,
  fully passing manifest; no opt-in/evidence remains `MANUAL_REQUIRED`.
- [x] Structural failures retain only bounded `issue_count`,
  `needs_work_count`, focus/action type counts, token totals, and digests; no
  dimension code or report/judge prose is written.
- [ ] Provider/judge infra failures use fixed redacted rate-limit, config,
  secret, capability, fallback, timeout, or provider reason codes. Only typed
  retryable provider/protocol failures consume the current action's remaining
  four-call budget; non-retryable/cancelled failures stop after one call and no
  failure is downgraded to PASS. Each product `GenerateReport` invocation owns
  `10s/20s/40s`, destroys retry state on return and resets for the next action;
  async job attempts are independent. Evalkit keeps separate synchronous
  generation/judge budgets.
- [x] `verify.sh` rejects `FAIL`, stale evidence, secret leakage, and raw files.
- [x] Setup clears every historical top-level entry. Failed retention preserves
  only bounded current-run diagnostics including `investigation.json`; PASS
  retains only current manifest, Agent audit, and required logs/env. PNG,
  browser/Playwright, unknown directory/file, and symlink artifacts never
  survive cleanup.
- [x] `cleanup.sh` retains only the redacted manifest/log/result and leaves the
  shared environment untouched.
- [x] Historical diagnostic run `e2e-p0-100-20260713T014058Z-80338` is retained
  only as an evalkit/runtime validator omission finding: attempt 11 paired
  `needs_practice` with `retry_current_round` and `next_round`. It is not a PASS;
  the accepted run below superseded it.
- [x] Historical diagnostic run `e2e-p0-100-20260713T020152Z-6086` completed
  11/11 automated attempts and blind review 5/5, but the final audit became
  visible as non-`0600` before the later mode change. The runner failed fast;
  this run must not be recorded as a PASS.
- [x] Historical run `e2e-p0-100-20260713T034811Z-35103` passed its then-current
  prompt and is not current evidence.
- [x] Current final-prompt run `e2e-p0-100-20260713T101214Z-59381` is retained
  as strict FAIL: mechanical 9/9, semantic judge 8/9, fixed representative
  cases 4/5; injection repetition 1 failed with `unsupported_item` at
  `$.summary`, so attempts 10-11 and blind review were not executed.
