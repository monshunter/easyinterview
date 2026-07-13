# E2E.P0.100 Expected Outcome

- Five distinct synthetic report cases (including zh-CN partial evidence) use the product-identical
  `review.BuildReportPromptMessages` trust boundary and real provider.
- Complete/partial run once; short/pending/injection critical cases each pass
  three independent generation + context-aware judge attempts.
- Every dimension is `>=0.70`, every registry-computed weighted score is
  `>=0.80`, and critical cases are 3/3.
- Item-level support and causal checks close fact-to-judgment-to-action;
  unsupported/fabricated claims, unsupported negatives, irrelevant or
  unexecutable advice, causal mismatch, and critical misses are zero.
- Empty focus remains legal only for one exact `answer_depth` or
  `answer_relevance` single-issue replay. Every other retry copies all sorted
  unique `needs_work` same-code issue codes; empty, partial, or extra focus
  fails before the judge and records no judge usage.
- Each report has at most two distinct action types; excess or duplicate action
  types fail before the judge and record no judge usage.
- Each `en` action label is at most 24 whitespace-delimited words and each
  `zh-CN` label is at most 64 Unicode code points; the manifest retains only the
  language/unit/limit/count audit, and over-limit output fails before judge.
- The initial completion, targeted label merge, and whole-report repair each
  reuse the runtime full semantic validator. Only a sole
  `nextActions[*].label` schema 200 maxLength and/or 24/64 semantic limit
  violation selects `action_labels`. Every other or mixed schema/semantic
  violation, including readiness/action/focus cross-field violations, selects
  `whole_report`. Generation has at most four LLM calls (initial + three
  retries); every invalid output recalculates its scope from the current full
  violation set and is fully revalidated. A fourth invalid output fails closed
  before judge. Judge has the same four-call ceiling but retries only typed
  retryable provider/fallback failures or
  protocol/schema/parse/coverage-invalid responses. A structurally valid
  negative content verdict is terminal and never resampled into PASS.
- Every evalkit live audit provides camelCase `repairScope` with enum
  `none|whole_report|action_labels`: completion `repairUsed=false` requires
  `none`, completion `repairUsed=true` requires a non-`none` scope, and judge
  requires `repairUsed=false` plus `repairScope=none`. V2 audits also require
  `attemptCount`, `retryCount`, closed `retryReasons`, and aligned
  `repairScopes`. Durable generation/judge summaries use `attempt_count`,
  `retry_count`, `retry_reasons`, and `repair_scopes`; the blind packet carries
  only generation metadata. No retry field stores an action label, raw output,
  provider response, or judge reason prose.
- Generation records `repair_used` and `repair_scope`; label-only repair cannot
  mutate non-label fields. Server truncation/authored copy and fourth-call
  schema/semantic invalid remain fail-closed.
- The short case proves a generic `retry_current_round` with empty focus.
- An independent Codex reviewer classifies every dynamic item and causal chain
  for five opaque, sample-ID-sorted representative outputs without case labels,
  criticality, repetition, gold, or judge material; its v2 redacted audit
  returns only sample IDs and binds the same context/output digests.
- Reviewer publication uses a hidden temporary file in the same output
  directory. It creates that file using `os.open` with `O_CREAT|O_EXCL` and mode
  `0600`; verifies every `review_digest` is complete; writes the full payload;
  flushes, calls `os.fsync`, and closes it; then publishes via same-filesystem
  `os.replace`; cross-filesystem rename is forbidden. The final path is complete
  and mode `0600` on first visibility; creating or patching the final path
  before a later `chmod` is forbidden. The runner retains its fail-fast mode
  check.
- Durable evidence contains only redacted digests, classifications, scores,
  usage/latency/finish reason/coordinate, and privacy booleans. Raw context,
  output, judge prose, cookie, email code, and secrets are absent.
- Setup retention removes every pre-existing top-level file, directory, and
  symlink before the current `setup.env` / `setup.log` are written. Failed
  retention preserves the current-run `investigation.json` and only bounded
  logs/env/result diagnostics while deleting manifest, Agent audit, PNG,
  browser/Playwright, unknown file/directory, and symlink artifacts. PASS
  retention keeps only the current manifest, Agent audit, and bounded logs/env;
  a current investigation rejects PASS and survives failed retention.
- A structural failure diagnostic may add only bounded `issue_count`,
  `needs_work_count`, focus/action type counts, token totals, and the output
  digest; it never includes dimension codes or report prose.
- Typed retryable provider/protocol failures may consume the current product
  action's remaining four-call budget. Each `GenerateReport` invocation waits
  `10s/20s/40s`, destroys retry state on return, and a later independent action
  starts at attempt one; async job attempts are infrastructure-only.
  Non-retryable config/secret/unsupported/cancel failures stop after one call.
  All terminal provider/judge failures are fixed-code, redacted FAIL results,
  and no PASS downgrade is allowed. Evalkit keeps separate synchronous
  generation/judge budgets without a second scheduler.
- Historical diagnostic run `e2e-p0-100-20260713T014058Z-80338` exposed an
  evalkit/runtime validator omission at attempt 11 when `needs_practice` was
  paired with `retry_current_round` and `next_round`. It is not acceptance
  evidence and is retained only as the defect superseded by the accepted run.
- Historical diagnostic run `e2e-p0-100-20260713T020152Z-6086` completed 11/11
  automated attempts and blind review 5/5, but final-audit publication exposed
  a non-`0600` path before the later mode change. The runner failed fast. It
  must not be recorded as a PASS.
- Historical run `e2e-p0-100-20260713T034811Z-35103` passed its then-current
  prompt and remains historical evidence only.
- Current final-prompt run `e2e-p0-100-20260713T101214Z-59381` is a strict
  scenario FAIL: mechanical 9/9, semantic judge 8/9, fixed representative case
  categories 4/5; injection repetition 1 was terminally rejected for an
  unsupported summary, so attempts 10-11 and blind review did not run.
