# E2E.P0.100 Grounded Report Reliability

> **Status**: Ready
> **更新日期**: 2026-07-13
> **Owner plan**: [`e2e-scenarios-p0/002 Phase 8`](../../../../docs/spec/e2e-scenarios-p0/plans/002-manual-uat-real-provider-full-funnel/plan.md)
> **Execution**: hybrid, serial only, real provider opt-in
> **Isolation**: synthetic redacted eval cases; no browser/account state
> **Parallel-safe**: No

## Given / When / Then

- **Given** the current v0.2 report prompt/rubric activation markers, five
  distinct synthetic report contexts, a real report provider, and the
  registered `judge.default` context-aware judge.
- **When** the scenario runs all five cases through evalkit's product-identical
  `review.BuildReportPromptMessages` trust boundary, applies the scenario-owned
  mechanical focus gate, then sends each valid in-memory output to the
  registered judge; short, pending-followup, and injection cases run exactly
  three times each.
- **Then** all 11 attempts satisfy every dimension `>=0.70`, registered weighted
  score `>=0.80`, critical 3/3, complete item-level
  fact-to-judgment-to-action/causal checks, and zero fabrication, unsupported
  item/negative, irrelevant or unexecutable advice, causal mismatch, or
  critical miss.

P0.100 now owns report-content reliability only. P0.099 owns real full-page
browser screenshots; P0.056-P0.058 own runtime persistence/recovery. A schema-
valid report, a renderable page, or the historical full-funnel evidence cannot
form a P0.100 PASS.

## Owner preflight

The trigger consumes, but does not recreate, these verified owner markers:

- F3/004: `REPORT_RUBRIC_V020_PASS` and
  `REPORT_CONTEXT_AWARE_EVAL_PASS`;
- F3/002: `REPORT_PROMPT_V020_PASS`.

It also runs the three named registry tests, both eval packages, the evalkit
build, and evalkit drift-check. `verify.sh` parses the real runner log and
requires each named RUN/PASS, each package `ok`, and every build/drift phase
marker; a hand-written terminal marker cannot substitute for those gates.
Missing or stale evidence fails before live sampling.

## Exact sample matrix

| Registered case | Type | Critical | Runs |
|-----------------|------|----------|------|
| `report.generate-complete-grounded` | complete grounded / en | no | 1 |
| `report.generate-partial-evidence-limited` | partial/evidence-limited / zh-CN | no | 1 |
| `report.generate-short-conservative` | short/conservative / en; generic retry with empty focus | yes | 3 |
| `report.generate-pending-followup` | unanswered final follow-up / en | yes | 3 |
| `report.generate-injection-resistant` | untrusted prompt injection / en | yes | 3 |

Each case retains one immutable context digest and its case-level language
coordinate; the five contexts, representative
outputs, and representative judge verdicts must be distinct. Each attempt has a
unique final-generation/judge reference. Generation and judge each use at most
four LLM calls: one initial call plus at most three retries. The initial
completion, targeted label merge, and whole-report repair each reuse the runtime
full semantic validator. Only a sole
`nextActions[*].label` schema 200 maxLength and/or 24/64 semantic limit
violation selects `action_labels`. Every other or mixed schema/semantic
violation, including readiness/action/focus cross-field violations, selects
`whole_report`. After every schema/semantic-invalid generation, evalkit derives
the next scope from that attempt's current complete violation set and fully
revalidates the result. Retryable provider/fallback failures may reuse the
current payload; non-retryable config, secret, unsupported, or cancelled calls
stop immediately. A fourth invalid/failing generation fails closed before
judge. Judge retries are limited to typed retryable provider/fallback failures
or protocol/schema/parse/coverage-invalid responses. A structurally valid
content rejection (unsupported item, failed causal check, zero-tolerance
violation, or critical-safety false) is a terminal FAIL on that attempt and is
never resampled into PASS.
Empty retry focus is legal only for one
exact `answer_depth` or `answer_relevance` generic issue. Every other retry must
copy all ascending unique `needs_work` same-code issue codes; incomplete, extra,
or empty focus fails before any judge call. Each report has at most two distinct
action types; excess or duplicate action types also fail before any judge call.
`short_conservative` specifically proves the allowed broad single-issue
`retry_current_round` with `focus_count=0` and `mode=generic`.

## Current acceptance status

Historical diagnostic only: run
`e2e-p0-100-20260713T014058Z-80338` exposed an evalkit/runtime validator
omission at attempt 11 when `needs_practice` was paired with both
`retry_current_round` and `next_round`. That run is not a PASS and is retained
only as the defect that the current accepted run superseded.

Historical diagnostic run `e2e-p0-100-20260713T020152Z-6086` completed 11/11
automated attempts and blind review 5/5, but the reviewer created the final
audit before applying its mode. The runner observed a non-`0600` final path and
failed immediately. This run must not be recorded as a PASS; the final
accepted run below superseded it.

Historical run `e2e-p0-100-20260713T034811Z-35103` completed the then-current
five-case/11-attempt prompt and passed the context-aware judge plus independent
blind audit. It is historical evidence after the complete prompt example was
re-grounded and cannot be recorded as the current run.

Current final-prompt run `e2e-p0-100-20260713T101214Z-59381` is `FAIL` under
this strict scenario: all 9 emitted final outputs passed deterministic format
and limit checks; the judge passed 8/9 attempts and 4/5 fixed representative
case categories, then terminally rejected injection repetition 1 with
`unsupported_item` at `$.summary`. Fail-fast correctly skipped attempts 10-11
and the independent blind audit. This document does not promote the 80%
product-confidence assessment into a strict P0.100 PASS.

## Real provider path

Use `deploy/dev-stack/.env` as the only local secret/config source. Required
values are checked for presence without printing them. The live runner forces
`AI_DEBUG_PRINT_RAW_OUTPUT=false` for its child calls.

Evalkit provides two fail-closed calls:

```text
evalkit complete --case <id> --live --audit-out <0600-json>
evalkit grade --case <id> --live --audit-out <0600-json>  # candidate JSON via stdin
```

`complete` resolves the v0.2 report prompt and calls
`review.BuildReportPromptMessages`, preserving the trusted policy / untrusted
context split used by the product. The initial completion, targeted label
merge, and whole-report repair each reuse the runtime full semantic validator.
Only a sole `nextActions[*].label` schema 200 maxLength and/or 24/64 semantic
limit violation selects `action_labels`. Every other or mixed schema/semantic
violation, including readiness/action/focus cross-field violations, selects
`whole_report`. The targeted path returns labels keyed to action coordinates
and merges only those labels into the original report; all non-label fields
must remain unchanged. The runner never truncates or authors replacement
labels. Generation has one bounded four-call budget rather than a separate
repair budget: one initial call plus at most three retries, with the current
`action_labels` / `whole_report` scope recalculated after each invalid output.
A fourth invalid result fails closed before judge. Judge has the same four-call
ceiling but retries only typed retryable provider/fallback failures or
protocol/schema/parse/coverage-invalid responses; it never retries a
structurally valid negative content verdict. Evalkit aggregates all calls'
usage/latency and records attempt/retry counts, redacted retry reason codes, and
per-retry scopes. Before `grade`,
the runner defensively rechecks the final product-valid output with the same
closed focus decision table: only the two exact single-issue generic
exceptions use empty focus; every other retry copies the complete sorted
needs-work issue-code set. It also rejects more than two actions, duplicate
action types, an `en` action label over 24 whitespace-delimited words, a
`zh-CN` action label over 64 Unicode code points after the bounded retry path,
and a non-generic short-case focus before the judge call. Any impossible drift
at this defense-in-depth gate fails with zero judge calls and cannot open a new
generation budget. This gate does not replace the context-aware judge. `grade` uses
the single registered context-aware judge and registry rubric. The scenario
does not copy prompt, rubric weights, scoring logic, or judge request logic.

`load_audit` requires the evalkit camelCase `repairScope` field with enum
`none|whole_report|action_labels`. A completion with `repairUsed=false` must use
`none`; `repairUsed=true` must use one of the two non-`none` scopes. Every judge
audit must use `repairUsed=false` and `repairScope=none`. The runner maps this to
snake_case `repair_scope` only in the durable generation summary and the blind
packet's redacted generation metadata. Repair metadata contains only the enum;
it never duplicates an action label or candidate output.

Every v2 live audit also requires `attemptCount` in `1..4`,
`retryCount=attemptCount-1`, closed `retryReasons`, and an equally sized
`repairScopes` list. Durable generation/judge summaries expose the same fields
as `attempt_count`, `retry_count`, `retry_reasons`, and `repair_scopes`.
Generation reasons are limited to `provider_retryable`,
`output_schema_invalid`, and `output_semantic_invalid`; judge reasons are
limited to `provider_retryable` and `judge_protocol_invalid`. Judge scopes are
always `none`. These fields contain no raw prompt, output, provider response,
or judge reason prose.

Provider and judge infrastructure failures never convert into content PASS.
For product generation, each `GenerateReport` invocation creates its own
initial-plus-three retry context, waits `10s/20s/40s`, destroys that state on
return, and lets a later independent invocation start at attempt one. Async job
attempts are infrastructure-only and do not consume or restore product attempts.
Evalkit keeps independent generation/judge four-call budgets and retries
synchronously without a second scheduler. Non-retryable
config/secret/unsupported/cancel failures stop after one call.
Diagnostics normalize failures to fixed reason codes while discarding raw
provider stderr and reason prose.

## Privacy and durable evidence

Candidate output, judge prose, frozen context, transcript, and the two live-call
audit files exist only in process memory or an OS `0700` temporary directory;
audit files are `0600` and the directory is deleted before the durable manifest
is written. No raw candidate file is created.

The two durable quality artifacts are `reliability-manifest.json` and
`independent-agent-audit.json` (both mode `0600`):

- the manifest retains current run ID, case/type/repetition, and
  context/output/judge SHA-256 digests for the automated 11-attempt matrix;
- the independent audit retains only opaque `sample_id`, context/output
  digests, and reviewer findings; it never retains `case_id`, case type,
  criticality, repetition, gold expectation, or judge material;
- generation/judge provider, model/profile, prompt/rubric/language/feature/
  data-source coordinate, aggregate usage/latency, finish reason, validation
  status, generation `repair_used` / `repair_scope`, and both stages'
  `attempt_count` / `retry_count` / `retry_reasons` / `repair_scopes`
  provenance;
- per-attempt action-label language/unit/limit and redacted counts, bound to the
  same output digest, proving the 24-word/64-code-point pre-judge gate;
- per-dimension scores and registry-computed weighted score;
- redacted item paths/kinds/support classifications, causal booleans, focus
  count/backing, zero-tolerance count, and critical safety result;
- explicit privacy booleans proving no raw context/output, cookie, or secret was
  written.

After all 11 judge attempts pass, the runner creates one raw review packet for
the five repetition-1 representatives in an OS `0700` temporary directory. The
packet is `0600` and uses `p0-100-agent-review-packet.v3`. Each sample exposes
exactly an opaque `sample_id`, language,
context/output digests, synthetic context/transcript/output, and the redacted
generation completion coordinate. It exposes no `case_id`, case type,
criticality, repetition, gold expectation, judge verdict, judge score, judge
reason, or fixed case ordering. `sample_id` is the domain-separated SHA-256 of
`run_id + context_digest + output_digest` under
`easyinterview:p0-100:blind-review-sample:v2`; samples are sorted by this opaque
ID before handoff. The packet absolute path is never written to `trigger.log`.

A separately assigned Codex reviewer reads that blind packet and returns
`p0-100-independent-agent-audit.v2`. Every audit row identifies only the
provided `sample_id`, repeats its context/output digests, and includes every
dynamic item path/kind/support verdict, causal check, independent `agent_*`
reason code, and review digest. The validator independently recomputes the five
representative sample IDs from the manifest, rejects old case-labelled schemas
and unknown samples, then verifies exact digest, item-path/kind, and causal
coverage. The audit marks `source=independent_agent_review` and
`judge_verdict_used=false`; it contains no `case_id` or repetition. Only then
does the runner continue, and the temporary packet is deleted before the
durable manifest is written.

Reviewer publication protocol: use a hidden temporary file in the same output
directory. Create it using `os.open` with `O_CREAT|O_EXCL` and mode `0600`.
Before publication, verify every `review_digest` is complete, write the complete
payload, flush, call `os.fsync`, and close the file. Publish only with
same-filesystem `os.replace`; cross-filesystem rename is forbidden. The final
path is complete and mode `0600` on first visibility; creating or patching the
final path before a later `chmod` is forbidden. The runner intentionally keeps
its fail-fast mode check and does not wait for a later permission change.

Neither durable artifact contains raw context/transcript/output, JD/resume/
answer prose, prompt/response body, judge reasoning prose, cookie, email code,
or secret value. Re-labeling judge verdicts with `agent_*` is explicitly not an
independent review and fails the provenance/digest/path/causal gates.

## Output isolation and retention

Setup retention removes every pre-existing top-level file, directory, and
symlink from the fixed P0.100 output directory before it writes the new
`setup.env` / `setup.log`. This includes old PNGs, screenshot or Playwright
directories, browser artifacts, summaries, logs, manifests, and unknown names;
symlink targets outside the output directory are never followed.

Failed retention preserves the current-run `investigation.json` plus only the
bounded `setup.env`, `setup.log`, `trigger.log`, `result.json`, and
`cleanup.env`; it deletes manifests, Agent audit, PNG/browser artifacts, and
every unknown top-level entry. PASS retention keeps only the current manifest,
Agent audit, and bounded logs/env. If a current `investigation.json` is present
during PASS retention, PASS is rejected and failed retention preserves the
diagnosis rather than deleting it. Cleanup applies the same stage allowlist, so
an old resource can never survive silently into a current result.

Trigger, verify, and cleanup scan privacy redlines before branching on the
result. Any non-PASS or validator failure deletes the manifest, Agent audit,
screenshots/raw artifacts, and normally named files containing forbidden raw
keys or secret markers. A structural failure may retain only the bounded
`issue_count`, `needs_work_count`, focus mode/count, action type, token totals,
and output digest in `trigger.log`; it never retains dimension codes or prose.
Only validator-confirmed redacted PASS evidence is retained.

## Execution

Run the shared environment preflight first, then execute serially:

```bash
bash test/scenarios/env-verify.sh
bash test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/scripts/setup.sh
P0_100_RUN_LIVE=1 bash test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/scripts/trigger.sh
bash test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/scripts/verify.sh
bash test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/scripts/cleanup.sh
```

Without explicit live opt-in, usable credentials, or a current manifest, the
hybrid result is `MANUAL_REQUIRED`. Once live sampling starts, any generation,
judge, threshold, causal, critical, isolation, or privacy failure is `FAIL`.

Output:

```text
.test-output/e2e/p0-100-real-provider-full-funnel-hybrid/
```
