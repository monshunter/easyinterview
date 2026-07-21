# Offline Evaluation Suite

This directory is the F3 offline evaluation harness truth source
(`prompt-rubric-registry/004-real-model-profile-and-evals` §4). It holds the
recorded evaluation cases and the LLM-judge scoring instruction; the business
prompts themselves are **never** copied here — they are resolved from
`config/prompts/` through the registry single source. `config/` is owned by
[A4 secrets-and-config](../../docs/spec/secrets-and-config/spec.md); F3 owns the
`config/evals/` namespace.

## 1 File layout

```
config/evals/
  README.md                          (this file)
  judge-instruction.md               (LLM judge scoring instruction; injected, not a business prompt)
  resolved-prompts.json              (committed registry single-source export; drift baseline)
  <feature_key>/
    cases.yaml                       (recorded evaluation cases for one feature_key)
  promptfooconfig.yaml               (Promptfoo runner config template; exec-bridges to evalkit)
  promptfoo_provider.js              (custom provider: shells to `evalkit complete`)
  promptfoo_assert.js                (javascript assertion: shells to `evalkit grade`)
```

Promptfoo runtime output is not part of `config/`. `make eval-offline`
regenerates tests at `.test-output/evals/promptfoo_tests.yaml`, renders the
concrete Promptfoo config to `.test-output/evals/promptfooconfig.yaml`, and
keeps Promptfoo state, SQLite DB, and logs under `.test-output/evals/promptfoo/`.
When `EVAL_OUTPUT_DIR` is overridden, all three paths move together under that
directory. Do not create or ignore `config/evals/.generated/`; if that path
appears, the runner output path has regressed.

## 2 Case schema

Each `<feature_key>/cases.yaml` is:

```yaml
feature_key: <feature_key>           # must match the directory name and a §3.1.1 chat feature_key
cases:
  - id: <unique case id>
    language: multi | en | zh-CN     # request language; report.generate must be en/zh-CN and match context.language
    prompt_version: <optional exact candidate version>
    rubric_version: <optional exact candidate version>
    input: <short description of the business input>
    context: { ... }                 # required for report.generate
    transcript: [ ... ]              # required for report.generate
    critical: true | false           # report critical-safety sampling
    redacted: true                   # required for tracked report cases
    output: { ... }                  # recorded golden model output (must satisfy the output schema)
    judge:                           # recorded judge verdict replayed in offline mode
      scores:
        - dimension: <rubric dimension name>   # one entry per rubric dimension, names verbatim
          value: 0.0                            # in [0,1]
      reasoning:
        summary: <non-empty>
        evidence_quotes: []
      item_verdicts: []               # required strict item verdicts for report.generate
      causal_checks: []               # one per needs-work dimension
      zero_tolerance_violations: []
      critical_safety_pass: true
```

The committed suite has exactly 32 cases covering the 6 current §3.1.1 chat
`feature_key`s, including at least one `en -> multi` fallback case. The five
`report.generate` cases are distinct complete/partial/short/pending/injection
fixtures; all carry synthetic redacted context + transcript, and exactly three
are critical. The eleven `practice.session.chat` cases pin the candidate
`v0.3.0` prompt/rubric coordinate and include named-target, anonymous-target,
resume-employer impersonation, and assistant identity-drift coverage.

## 3 Runner and gates

The Go `backend/cmd/evalkit` binary owns registry single-source resolution and
grading via the single `registry.LLMJudge`; Promptfoo is the runner that
orchestrates and reports.

- `make eval-offline` — single-source drift gate + exact `32` count gate + the
  Promptfoo runner over recorded fixtures. Deterministic, zero-cost, no network.
  **Not** part of `make test`. Runtime artifacts are confined to
  `$(EVAL_OUTPUT_DIR)` (default `.test-output/evals/`).
- `make eval-offline-resolve` — regenerate the committed `resolved-prompts.json`
  after a deliberate registry prompt change; the drift gate fails until it is
  regenerated.
- `EVAL_LIVE=1 make eval-offline` — opt in to real provider/judge calls
  (`judge.default` profile). Requires provider secrets; excluded from CI and the
  default offline run.
- `evalkit complete --case <id> --live --audit-out <file>` — write the candidate
  JSON to stdout for an ephemeral `0600` pipe/file and persist only a redacted
  `evalkit-live-call-audit.v1` completion audit (coordinate, model/provider,
  usage, latency, validation, `finishReason`, output digest/size).
- `evalkit grade --case <id> --live --output <json> --audit-out <file>` — run the
  context-aware judge, return scores, registry-derived `weighted_score`, item
  verdicts, causal checks and critical safety on stdout, and persist the same
  redacted call-audit schema for the judge. The audit never contains prompt,
  frozen context, transcript, candidate output, judge response, cookie or
  secret values; callers must delete the ephemeral stdout payload after merging
  a redacted UAT manifest.

The drift gate proves Promptfoo consumes the same prompt the registry resolves
(no second prompt copy); `make lint-prompts-hardcode` proves no prompt body was
copied into business code.

For report v0.2, every dimension must score at least `0.70` and the weighted
mean must be at least `0.80`. Unsupported items, fabrication, unsupported
negative claims, irrelevant/unexecutable advice, causal mismatch, or critical
safety failure have zero tolerance. Partial support passes only when the report
explicitly marks evidence limits and does not use the item negatively.

## 4 References

- [prompt-rubric-registry spec](../../docs/spec/prompt-rubric-registry/spec.md) (D-9, D-15, §3.2, §5, C-10/C-15)
- [004-real-model-profile-and-evals plan](../../docs/spec/prompt-rubric-registry/plans/004-real-model-profile-and-evals/plan.md)
- [config/prompts truth source](../prompts/README.md)
- [config/rubrics truth source](../rubrics/README.md)
