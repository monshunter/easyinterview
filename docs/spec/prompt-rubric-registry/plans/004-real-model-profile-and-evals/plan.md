# F3 Real Model Profile and Evals

> **版本**: 1.24
> **状态**: completed
> **更新日期**: 2026-07-13

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

维护 F3 当前离线评估闭环：`judge.default` 使用真实 judge profile，`LLMJudge` 按 rubric dimension 返回 `[]Score + Reasoning`，Promptfoo 通过 registry single-source 消费业务 prompts；当前 suite 从 27 个录制 fixture case扩展为 28 个，`make eval-offline` 默认零网络运行。

本 plan 不新增用户可见 UI，不改变 HTTP API，不进入业务请求链路，也不拥有 A3 chat 业务 profile 的 status。它唯一拥有 `config/rubrics/report.generate/v0.2.0.yaml`、context-aware report judge/eval；F3 `002` 唯一拥有 v0.2 prompt/schema、多版本 parser 与最终激活。

## 2 当前合同

- `backend/internal/ai/registry.Judge` 返回 `([]Score, Reasoning, error)`；每个 `Score` 对应一个 rubric dimension，`Value` 位于 `[0,1]`。
- `LLMJudge` 注入 rubric provider 与 judge model client，通过 `judge.default` profile 调用 judge capability；profile 不可用、output schema invalid、模型输出无法解析、维度缺失或维度不匹配时 fail-close。
- `config/ai-profiles.yaml` 中 `judge.default` 当前为 `active`，精确使用 `judge-deepseek` / `deepseek-v4-pro`、non-thinking JSON、max tokens 6144、60s timeout、无 fallback、profile v1.2.0；A3 owns profile/wire/final-content fail-close，F3 只消费该坐标与 `JUDGE_FINAL_CONTENT_V120_PASS`。当前 chat business profiles 保持各自 A3 owner 状态，只由 coverage gate 校验 runnable provider / model。
- `config/ai-providers.yaml` 提供 `judge-deepseek`，protocol 为 `judge_compatible`，capability 为 `judge`。
- `config/evals/` 当前覆盖 6 个 chat feature_key、精确 28 个录制 fixture case，并包含 `en -> multi` fallback；5 个 report case 覆盖 complete(en)、evidence-limited(zh-CN)、short generic retry/empty focus(en)、pending(en)、injection(en)。
- Promptfoo pinned version 为 `0.121.12`，通过仓库依赖执行；eval output、Promptfoo config、logs 和 state 默认写入 `.test-output/evals/`，可由 `EVAL_OUTPUT_DIR` 覆盖。
- `make eval-offline` 默认使用录制 fixture，不发网络请求；`EVAL_LIVE=1` 才允许真实 provider 调用，且不纳入 `make test`。
- Promptfoo custom provider / grader 只能通过 registry resolved prompt、schema、rubric 与 `LLMJudge`；不得复制业务 prompt 正文。
- 方案 A action schema maxLength200 code points只作fuse；24 whitespace words/64 Unicode code points是quality gate；targeted action-label repair内部生成目标18/52。P0.099验收desktop+390与typed-invalid/no-raw。
- Action support is type-specific and evidence-bounded: `retry_current_round` only turns cited missing behavior into something to add；`review_evidence` may revisit cited positive or explicitly evidence-limited content without inventing an artifact, corrective gap, new scenario or transfer task；`next_round` requires frozen `hasNextRound=true` and permitted readiness。Every type treats a mechanism, threshold, tool, sequence, framework or example absent from cited candidate messages as `unsupported`, never `partial`。
- `report.generate/v0.2.0` judge request 从 strict-decoded report 机械派生 ordered `expected_item_verdicts(path/kind)` 与按 dimension assessment 顺序排列的 `expected_causal_dimension_codes`；judge response 必须精确回填。空 highlights/issues 无集合 verdict，`$.retryFocusDimensionCodes` 是唯一数组整体 verdict。
- Evalkit generation与judge使用相互独立的max4-call budgets。Generation每轮完整validate并按当前violations选择action_labels/whole_report；judge仅重试retryable provider或protocol/schema invalid，valid negative typed terminal。最终 prompt 的 live run `e2e-p0-100-20260713T101214Z-59381` 机械9/9、语义8/9、代表场景4/5；严格11/11诊断因第9次unsupported summary而FAIL，不伪装为当前PASS。

## 3 质量门禁分类

- **Plan 类型**: `code-internal + contract + tooling + config`
- **TDD 策略**: Judge signature、judge dispatch、LLMJudge fail-close、profile coverage、eval count、Promptfoo single-source、no-network default 和 output-dir routing 均由 focused tests / lint / Make gate 覆盖。重进本 plan 时先运行对应 gate 暴露 drift，再最小修复。
- **BDD 策略**: BDD 不适用。本 plan 不产生用户行为流；用户可见 AI 输出质量由业务 owner 的 BDD / scenario gate 承接。
- **替代验证 gate**: `go test ./backend/internal/ai/registry -count=1`、`go test ./backend/internal/ai/aiclient/... -run 'Test.*Judge|Test.*judge' -count=1`、`go test ./backend/internal/ai/aiclient/providers/judge_compatible -count=1`、`go test ./backend/internal/ai/aiclient/profile -count=1`、`make lint-ai-profile-coverage`、`make eval-offline`、`make lint-prompts-hardcode`、registry single-source drift check、`validate_context.py`、`sync-doc-index --check`、`make docs-check`、`git diff --check`。

## 4 交付范围

### 4.1 Judge interface and LLMJudge

The registry Judge contract is per-dimension. `FailClosedJudge` remains the safe default for unconfigured callers, while `LLMJudge` is the active eval implementation. `LLMJudge` loads rubric dimensions, validates evaluated output against the feature output schema, calls the judge-capability model client and rejects malformed judge output.

### 4.2 Judge profile and provider coverage

`judge.default` is active and routes to a runnable `judge_compatible` provider. `scripts/lint/ai_profile_coverage.py` verifies judge/default capability alignment and rejects non-runnable provider/model markers for judge and current chat business profiles.

### 4.3 Eval fixtures and Promptfoo runner

`config/evals/<feature_key>/cases.yaml` owns fixture eval cases. The runner generates Promptfoo config under `EVAL_OUTPUT_DIR`, grades through registry/LLMJudge, and stores runtime state outside `config/evals/`. Promptfoo assets must not duplicate business prompt bodies.

### 4.4 Offline/live execution split

Offline eval is deterministic and safe for local gates. Live eval is opt-in through `EVAL_LIVE=1`, uses platform config/secrets bootstrap, and remains outside `make test`.

### 4.5 Eval wire score conversion simplification

`DimensionScore` and the private JSON `wireScore` remain field-identical apart from serialization tags. The eval harness uses an explicit type conversion instead of repeating the field mapping, while `RunOffline` and `JudgeTranscript` continue to exercise the strict `LLMJudge` JSON contract. Recorded outputs follow the current `resume.parse` and `target.import.parse` schemas instead of weakening validation, and `resolved-prompts.json` is regenerated from the registry single source. Owner discovery includes both `backend/internal/eval` and `backend/cmd/evalkit`.

### 4.6 Fixed judge profile state removal

`judge.default` is the locked F3 profile and has no runtime override consumer. Remove the unreachable `WithJudgeProfile` option and the constant-valued `LLMJudge.profileName` field; `LLMJudge` calls the model client with the locked profile directly while the existing judge contract test continues to assert that value.

### 4.7 Context-aware report evaluation

Create the immutable-content `config/rubrics/report.generate/v0.2.0.yaml` as this plan's sole report-rubric artifact. It retains the established dimensions and weights (`report_evidence=0.35`, `report_specificity=0.25`, `report_action_quality=0.25`, `report_calibration=0.15`) while rewriting descriptions/score levels around context-grounded facts, item-level support, executable advice and needs-work→issue→focus/action causal calibration. It uses the exact `status: inactive` metadata contract introduced by F3 `002`, but this plan does not switch the active pair. After both GREEN markers, 002 may mutate only this file's `status`; dimensions/weights/score levels remain 004-owned immutable content.

Extend eval cases and judge input with optional structured `context` and `transcript`; report.generate requires both, while other feature keys retain current output-only behavior. Each judge dimension also carries its exact source `weight` and ordered `score_levels(label/threshold/description)` so numeric bands are reproducible. Evalkit and Promptfoo provider/grader preserve this payload without duplicating business prompts. For v0.2 reports, the registry derives ordered `expected_item_verdicts(path/kind)` from actual summary/preparedness and indexed dimension/highlight/issue/action items, appends exactly one array-level `$.retryFocusDimensionCodes:advice`, and derives `expected_causal_dimension_codes` from needs-work dimensions in assessment order. The judge copies both expected lists exactly; empty highlights/issues contribute no collection verdict. Report rubric/judge returns only this exact item-level fact/judgment/advice `supported|partial|unsupported` coverage and causal checks in addition to dimension scores; tracked outputs are redacted.

Evalkit generation uses an in-memory budget of4 calls。Every output receives the product full validator；each invalid round recomputes all violations and chooses`action_labels`or`whole_report`。Attempts2-4 may recover and all usage/latency aggregate。Judge uses a separate4-call budget：only retryable provider failure or judge protocol/schema invalid can consume another attempt。A structurally valid unsupported/causal/zero-tolerance/critical negative verdict is typed content rejection and terminates immediately。Both manifests record redacted attempt_count/retry_count/reason/scope。

Historical runs retain their diagnostic value: `36625` exposed the generic-replay rubric contradiction, `80338` exposed a full-validator escape, `59906` exposed unsupported injection-summary wording, `75753` exposed an exact empty-focus misclassification, and `25849` was deliberately aborted after a contract change. Run `35103` passed the then-current 11-attempt prompt and remains historical evidence only.

After retaining and grounding the complete prompt example, the current run is `e2e-p0-100-20260713T101214Z-59381`. All nine emitted final outputs passed JSON/schema/200/24/64/focus/action/cross-field validation. Eight of nine judge attempts passed; all five fixed representative case categories were reached and four passed. The ninth attempt failed terminally at `$.summary` with `unsupported_item`, so the strict runner stopped before attempts 10-11 and blind review. Product acceptance uses the user-approved fixed five-case semantic sample (4/5 = 80%) plus deterministic mechanical fail-close; the stricter 11/11 scenario remains an explicit FAIL diagnostic and is never promoted to PASS.

Replace the four report cases that reuse one output anchor with five distinct complete/partial/short/control-only+pending/injection contexts and outputs, bringing the exact suite count from 27 to 28 and covering all four preparedness tiers. The injection inputs include fake role/schema/XML delimiters; the well-prepared gold requires two strong/high dimensions from two distinct substantive answers. Item-level `unsupported` always fails; `partial` is allowed only when the report explicitly says evidence is limited and does not use it for a negative readiness/action claim. Every action must be immediately executable under the user's control and each type appears at most once. Retry advice only turns cited missing behavior into something to add；review advice may revisit cited positive/explicit evidence-limit without inventing artifact/corrective gap/new scenario/transfer task；next requires hasNextRound+permitted readiness。Every type rejects a new uncited mechanism, threshold, tool, sequence, framework or example rather than marking it partial. Lower tiers contain exactly one retry, optionally one review, never next-round. Generic empty focus is valid only for final exact single `answer_depth` brief or exact single `answer_relevance` control-only issue；all other retry focus equals the complete ascending unique same-code needs-work issue set，with subset/superset and `I>=2` empty rejected。For every selected focus code, the first retry label names at least one directly cited missing behavior without extending it into a new solution；multi-focus uses one short semicolon-separated fragment per code in focus order；English labels use at most 24 whitespace words and zh-CN labels at most 64 Unicode code points。Targeted action-label repair instructions aim for18/52 before full200+24/64 revalidation。An umbrella-only label is unsupported regardless of a high aggregate judge score. An explicitly unsafe current-round approach is blocking, not an ordinary detail gap. Every rubric dimension must score at least 0.70 and the weighted mean at least 0.80; fabricated fact, unsupported readiness/negative, irrelevant or externally contingent advice, causal mismatch or critical safety miss is zero-tolerance. Contract-safe alone is insufficient.

The five-case matrix must include a complete Chinese `zh-CN` evidence-limited report and a short-answer `retry_current_round` case whose exact single `answer_depth` issue permits empty focus. The pending/control-only case uses exact single `answer_relevance` for the other empty exception. The prompt retains a complete JSON exemplar paired with a synthetic prioritization answer whose stated gap is an unexplained tie-breaking rule；the adjacent anti-copy rule limits it to JSON shape/cross-field coherence and requires regeneration from current evidence. P0.100 runs the two non-critical cases once and the three critical cases three times (11 attempts total). Before any judge call, P0.100 mechanically rejects non-exception empty focus, subset/superset, `I>=2` empty and duplicate action types；negative fixtures prove zero judge calls and preserve only bounded redacted issue/needs-work/action/focus/token counts, focus mode and digests. Because generation and judge currently use the same DeepSeek V4 Pro model family, a strict PASS additionally requires the same representative context/transcript/output for all five cases to reach a separate Agent reviewer through OS-only 0700/0600 temporary files. The reviewer does not read judge verdicts；if a terminal negative stops the strict runner before review, the run remains FAIL while its completed redacted sample counts remain valid diagnostic evidence.

When v0.2 rubric lint, strict judge contract, all five distinct report cases, the exact 28-case offline gate and critical 3/3 gate pass, emit `REPORT_RUBRIC_V020_PASS` and `REPORT_CONTEXT_AWARE_EVAL_PASS`. Hand those markers to F3 `002`; only 002 may perform final prompt/rubric activation.

## 5 验收标准

- Judge interface, LLMJudge, judge adapter and profile catalog focused Go tests pass.
- `make lint-ai-profile-coverage` proves `judge.default` active, judge provider protocol/capability valid, and current chat profiles runnable.
- `make eval-offline` runs 28 cases without network in default mode after Phase 8.
- Evalkit generation/judge retry matrices pass with independent max4 caps，typed protocol-invalid/content-rejected outcomes and aggregate audit manifest；valid negative verdict is never retried。
- Current product acceptance records mechanical 9/9, semantic judge 8/9 and fixed representative cases 4/5 (80%) without replacement sampling；strict P0.100 remains FAIL because it did not complete 11/11 or blind review。maxLength200 proves fuse only；18/52 proves targeted-repair margin only；P0.099 desktop+390 proves24/64 UX。
- Promptfoo version is pinned in repo dependency metadata.
- `make lint-prompts-hardcode` and registry single-source drift gate pass.
- Docs/context/index gates pass and active docs describe the current judge/eval contract.
- `deadcode -test` and symbol inventory report no judge profile override surface.

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| Judge profile active but unusable | Judge dispatch, provider adapter and profile tests cover capability/protocol alignment |
| Eval runner duplicates prompt text | Runner resolves prompts through registry and drift gate compares resolved prompt content |
| Live eval leaks into default local gates | `make eval-offline` defaults to fixture mode; no-network tests enforce that default |
| Eval output pollutes source tree | Promptfoo runtime output is routed under `EVAL_OUTPUT_DIR`, default `.test-output/evals/` |
| A3 chat profiles are changed by this owner | Coverage gate reads chat profiles but this plan only owns `judge.default` status/provider changes |
| Generation and judge share one provider/model and self-confirm | P0.100 requires a second Agent to inspect the same temporary raw context/output without judge verdict access; digest-bound redacted audit is mandatory and raw is deleted |
| Judge retry hides a real quality failure | Retry only provider/protocol invalid；valid negative is typed terminal rejection and its attempt_count cannot increase |

## 7 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-07-13 | 1.23 | Independent max4 evalkit generation/judge budgets；dynamic generation repair scope，typed judge retry boundary，aggregate attempt manifest；run25849 aborted10/11，markers pending. | backend-review/001 + F3/002 + P0.100 |
| 2026-07-13 | 1.22 | Reuse product full validator for evalkit repair scope；record80338 validator escape、59906 injection grounding fix and75753 exact generic empty-focus fix；focused GREEN only，full matrix/markers pending. | backend-review/001 + P0.100 |
| 2026-07-13 | 1.21 | Repair exact generic-replay versus report_action_quality contradiction；single short-conservative retest0.82/0.70 zero violations，full live matrix remains pending. | backend-review/001 + P0.100 |
| 2026-07-13 | 1.21 | Finalize A：fuse200；semantic/quality24 whitespace words/64 Unicode code points；targeted repair margin18/52；keep markers/P0.100 pending. | F3 002 + P0.099/P0.100 |
| 2026-07-13 | 1.20 | A-200 fuse200；keep14/40 and reopen F3 markers/P0.100 live rerun. | F3 002 + P0.100 |
| 2026-07-13 | 1.19 | Record live P0.100 FAIL：label>120 was misclassified whole_report；require normalized all-label schema120/14-40 violation sets to use action_labels. | P0.100 |
| 2026-07-13 | 1.18 | Add one-budget whole-report versus sole-action-length label-only LLM repair，labels-only merge，full revalidation and one-shot judge contract. | backend-review/001 + P0.100 |
| 2026-07-13 | 1.17 | Separate the 120-char fuse, P0.100 5-case/11-attempt reliability, and P0.099 current-run canonical 390x844 UX audit. | backend-review/001 + F3/002 + P0.099/P0.100 |
| 2026-07-13 | 1.16 | Add evalkit same-source output-schema validation, one `$ / output_schema_invalid` product repair, aggregate generation usage/latency + repair_used and zero judge repair/retry；bound action length and multi-focus semicolon fragments. | backend-review/001 + F3/002 + P0.100 |
| 2026-07-13 | 1.14 | Tighten empty-focus, multi-issue focus, lower-tier action and unsafe blocking calibration; add the non-blocking incident exemplar, mechanically derived ordered expected judge lists and P0.100 pre-judge/redacted gates. | backend-review/001 + P0.100 |
| 2026-07-12 | 1.13 | Make judge score bands reproducible, require preparedness/focus item coverage and cover all readiness tiers plus control-only/fake-schema injection before live P0.100. | backend-review/001 + P0.100 |
| 2026-07-12 | 1.12 | Add multilingual/generic-empty-focus report cases, consume the A3 judge final-content marker and require same-output independent Agent audit for P0.100. | A3/003 Phase 9 + backend-review/001 + P0.100 |
| 2026-07-12 | 1.11 | Revalidate report context-aware judge/eval markers after F3/002 recursive closed-schema and runtime bounds enforcement. | F3 002 Phase 14 |
| 2026-07-12 | 1.10 | Clarify v0.2 rubric immutable content versus 002-owned status-only activation metadata. | F3 002 Phase 14 |
| 2026-07-12 | 1.9 | Make 004 the sole v0.2 report rubric/context-aware judge/eval owner and hand two GREEN markers to 002 for final activation; lock current suite 27→28. | F3 002 Phase 14 + backend-review/001 |
| 2026-07-12 | 1.8 | Reopen Phase 8 for context-aware report eval/judge and content-level reliability verdicts. | backend-review/001 + P0.100 |
| 2026-07-10 | 1.7 | 删除零消费者 judge profile override 与恒定 profile 字段，直接使用已锁定的 `judge.default`。 | tech-debt pruning |
| 2026-07-10 | 1.6 | 简化 eval score wire conversion，修复 current-schema fixtures / resolved prompt projection，并补齐 eval package / command 的 owner discovery。 | tech-debt pruning |
| 2026-07-10 | 1.5 | 同步当前 `FailClosedJudge` / `LLMJudge` 代码事实，并将 profile coverage 表述收敛为 runnable / non-runnable marker。 | tech-debt pruning |
| 2026-07-07 | 1.4 | 压缩 owner 文档为当前 judge.default active、LLMJudge、36-case eval-offline and Promptfoo single-source contract。 | product-scope/001-core-loop-module-pruning |
| 2026-05-24 | 1.3 | 完成 real model profile、judge implementation and eval-offline delivery。 | prompt-rubric-registry |
