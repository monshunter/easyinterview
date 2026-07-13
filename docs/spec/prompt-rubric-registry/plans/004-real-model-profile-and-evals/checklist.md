# F3 Real Model Profile and Evals Checklist

> **版本**: 1.24
> **状态**: completed
> **更新日期**: 2026-07-13

**关联计划**: [plan](./plan.md)

## 1 Judge contract

- [x] 1.1 `Judge` interface returns per-rubric-dimension `[]Score` plus `Reasoning`.
- [x] 1.2 `FailClosedJudge` remains a safe default and returns `ErrJudgeUnavailable`.
- [x] 1.3 `LLMJudge` validates evaluated output schema, calls judge capability through `judge.default`, returns one score per rubric dimension and fail-closes malformed judge output.

## 2 Judge profile and provider coverage

- [x] 2.1 `judge.default` is active in `config/ai-profiles.yaml`.
- [x] 2.2 `judge.default` routes to runnable `judge-deepseek` / `deepseek-v4-pro`; A3 Phase 9 locks non-thinking JSON / max tokens 6144 / timeout 60000 / no fallback / v1.2.0 and emits `JUDGE_FINAL_CONTENT_V120_PASS`.
- [x] 2.3 `config/ai-providers.yaml` includes `judge-deepseek` with `judge_compatible` protocol and `judge` capability.
- [x] 2.4 `make lint-ai-profile-coverage` verifies judge profile/provider alignment and rejects non-runnable provider/model markers for judge and current chat business profiles.

## 3 Eval fixtures and Promptfoo

- [x] 3.1 Historical 9-key/36-case baseline was delivered; product-scope pruning later reduced current truth to 6 keys/27 cases, and Phase 8 replaces report cases for a 28-case current suite.
- [x] 3.2 Promptfoo is pinned as a repo dependency and runs through repo-owned commands.
- [x] 3.3 Promptfoo provider/grader uses registry resolved prompt/rubric/schema and `LLMJudge`; business prompt bodies are not duplicated in eval assets.
- [x] 3.4 Eval runtime output writes under `EVAL_OUTPUT_DIR`, default `.test-output/evals/`.

## 4 Offline/live execution gates

- [x] 4.1 `make eval-offline` runs recorded fixtures by default and does not make network calls without `EVAL_LIVE=1`.
- [x] 4.2 `EVAL_LIVE=1` is explicit opt-in and remains outside `make test`.
- [x] 4.3 `make lint-prompts-hardcode` remains green.
- [x] 4.4 Registry single-source drift gate fails when Promptfoo output diverges from registry resolved prompts.

## 5 Current owner compression gate

- [x] 5.1 `prompt-rubric-registry/spec.md`, `plan.md`, `checklist.md`, `context.yaml` and plans INDEX align to the current judge/eval contract.
  <!-- verified: 2026-07-07 method=current-owner-compression evidence="Updated prompt-rubric-registry spec.md to v2.17, plan.md/checklist.md to v1.4, context specVersion to v2.17, and synced docs/spec plus prompt-rubric plans INDEX. PASS: targeted stale-wording grep returned no matches; validate_context.py prompt-rubric-registry/004 backend PASS; sync-doc-index --check PASS; go test ./backend/internal/ai/registry -count=1 PASS after updating backend_review_preflight_test spec-version assertion to 2.17; go test ./backend/internal/ai/aiclient/... -run 'Test.*Judge|Test.*judge' -count=1 PASS; go test ./backend/internal/ai/aiclient/providers/judge_compatible -count=1 PASS; go test ./backend/internal/ai/aiclient/profile -count=1 PASS; make lint-ai-profile-coverage PASS; make lint-prompts-hardcode PASS; pnpm rebuild better-sqlite3 restored local Promptfoo sqlite binding; make eval-offline PASS (36 cases, no network, Promptfoo 36 passed)." -->

## 6 Eval wire score conversion simplification

- [x] 6.1 将同构 `DimensionScore` 显式转换为私有 `wireScore`，删除重复字段映射并保持严格 judge JSON 合同（验证：eval package tests、`make eval-offline`、scoped `staticcheck`）
- [x] 6.2 owner context discovery 补齐 `backend/internal/eval` 与 `backend/cmd/evalkit`（验证：F3/004 context、index/docs/diff/pruning gates）
- [x] 6.3 `resume.parse` 四档录制输出补齐 current schema 必填的 `displayName` / `markdownText`，不放宽 output schema（验证：`TestRealSuiteOfflineGreen`、`make eval-offline`）
- [x] 6.4 `target.import.parse` 四档录制输出改为 current `title` / `companyName` / `interviewRounds` shape，删除范围外 `interviewHypotheses`，并从 registry 重新生成 `resolved-prompts.json`（验证：36-case schema audit、`make eval-offline-resolve`、`make eval-offline`）
  <!-- verified: 2026-07-10 method=eval-wire-and-current-fixture-reconcile evidence="S1016 red identified repeated DimensionScore field mapping. Initial real-suite red exposed missing Resume displayName and TargetJob title; a 36-case schema audit isolated the 8 alias-backed failures. Current Resume/TargetJob outputs and resolved prompt projection were repaired without weakening schemas. Eval and evalkit tests, scoped staticcheck, profile/hardcode lints, F3/product contexts, docs/diff/pruning gates PASS. make eval-offline PASS: 36 graded offline with no network and Promptfoo 36 passed, 0 failed, 0 errors; real_residuals=0." -->

## 7 Fixed judge profile state removal

- [x] 7.1 删除零消费者 `WithJudgeProfile` 与恒定 `LLMJudge.profileName` 字段，直接使用 F3 锁定的 `judge.default`；验证：`deadcode -test` RED/GREEN、symbol inventory、registry judge tests、scoped `staticcheck`、profile coverage 与 owner docs gates
  <!-- verified: 2026-07-10 method=fixed-judge-profile-state-removal evidence="deadcode -test RED reported WithJudgeProfile as unreachable. Removed the option and constant-valued field; the existing registry test still asserts judge.default reaches the model client. Registry/judge tests, staticcheck, profile/hardcode lints and reachability inventory PASS; make eval-offline PASS 36/36 with no network." -->

## 8 Context-aware report reliability eval

- [x] 8.1 RED-GREEN: create only `config/rubrics/report.generate/v0.2.0.yaml` with `status: inactive` and locked dimensions/weights `report_evidence=.35`, `report_specificity=.25`, `report_action_quality=.25`, `report_calibration=.15`; rubric lint rejects status enum, weight/dimension/score-level drift. Do not create prompt/schema or activate the pair; 002 may later change status only.
  <!-- verified: 2026-07-12 red="v0.2 rubric file absent and status lint accepted retired" green="locked-weight/status tests pass; rubric_lint reports 7 files clean; v0.2 remains inactive" -->
- [x] 8.2 RED-GREEN: eval case schema, backend/internal/eval, evalkit and Promptfoo bridge preserve optional context+transcript; report.generate requires both and other feature keys remain compatible.
  <!-- verified: 2026-07-12 red="Case lacked prompt/rubric version, context, transcript, critical and redacted fields" green="LoadSuite requires redacted report context+non-empty transcript; ResolveCase pins inactive v0.2; evalkit render test proves frozen context/transcript reach the candidate prompt while non-report cases remain output-only compatible" -->
- [x] 8.3 RED-GREEN: LLMJudge request includes context+transcript+output, exact dimension weight/ordered score levels, mechanically derived ordered `expected_item_verdicts(path/kind)` and assessment-ordered `expected_causal_dimension_codes`; judge response must copy both lists exactly. Empty highlights/issues produce no collection verdict, indexed actual items are covered exactly once, and `$.retryFocusDimensionCodes` remains the sole array-level verdict. Preparedness and retry focus are mandatory；empty focus is accepted only for final exact single `answer_depth` brief or exact single `answer_relevance` control-only issue，while all other retries require the complete ascending unique same-code needs-work issue set；subset/superset and `I>=2` empty fail closed。Action types are unique，lower tier has exactly one retry/optional review/no next-round，and unsafe current-round approach is blocking；malformed/missing/extra/wrong-kind verdicts fail closed.
  <!-- verified: 2026-07-13 red="an empty highlights array induced an invalid $.highlights collection verdict; expected verdict/causal coordinates were implicit" green="grounded request carries frozen_context+transcript+output, source-exact weight/ordered score bands, ordered expected_item_verdicts and expected_causal_dimension_codes; judge copies exact coordinates, empty highlights/issues add zero verdicts, retry focus remains the only array verdict; registry full+race PASS" -->
- [x] 8.4 Replace 4 reused-output report cases with 5 distinct complete/partial/short/control-only+pending/injection contexts/outputs; suite exact count changes 27→28, all four readiness tiers are covered, evidence-limited is full zh-CN with immediate advice, short/control-only retries are generic with empty focus, complete safety has multi-issue focused retry, and repeated-output/unknown-action negatives fail. Retain the complete JSON exemplar, pair it with a non-blocking prioritization/tie-breaking synthetic input, and require anti-copy/current-context regeneration.
  <!-- verified: 2026-07-13 commands="cd backend && go test ./internal/eval ./cmd/evalkit -count=1; make eval-offline-resolve; make eval-offline; python3 -m pytest scripts/lint/prompt_lint_test.py -q" result="exact 28; all four tiers; paired complete exemplar and anti-copy gate; Promptfoo 28/28 PASS; prompt lint 24/24 PASS" -->
- [x] 8.5 SCHEME-A STATIC/OFFLINE-GATE: schema maxLength200 fuse；semantic/per-code bound24 whitespace words/64 Unicode code points；targeted repair margin18/52。Re-run offline gates and re-emit markers；P0.099 desktop+390/typed-invalid-no-raw remains separate。
  <!-- verified: 2026-07-13 red="real judge exposed both review_evidence contradicting a universal missing-behavior rule and a high-scoring umbrella-only focused retry" green="prompt preflight requires per-focus directly cited behavior and rejects umbrella terms; judge instruction uses retry/review/next support branches, forbids invented review scenario/transfer task and preserves no-uncited-specificity" -->
  <!-- verified: 2026-07-13 method=prompt-rubric-resolve+offline evidence="schema200 fuse, semantic24/64, targeted18/52 and resolved single-source gates PASS; offline suite 28/28 PASS; P0.099 remains the independent UX gate" -->
- [x] 8.6 Re-emit current `REPORT_RUBRIC_V020_PASS` and `REPORT_CONTEXT_AWARE_EVAL_PASS` after 8.5 final-scheme static/offline gates and the generic-replay rubric regression pass；historical markers below do not close this item。This plan must not activate v0.2 or edit F3/002-owned prompt/schema.
  <!-- verified: 2026-07-13 markers="REPORT_RUBRIC_V020_PASS REPORT_CONTEXT_AWARE_EVAL_PASS" basis="current prompt/rubric/resolved single-source and offline 28/28 PASS; re-emitted markers supersede historical marker evidence" -->
- [x] 8.7 HISTORICAL-FOUNDATION: evalkit resolves the same F3 schema and aggregates usage/latency。Historical single-repair/one-shot evidence is superseded and does not prove max4 generation/judge budgets。
  <!-- verified: 2026-07-13 commands="cd backend && go test ./cmd/evalkit -run 'TestValidateLiveReportOutputSchema|TestCompleteLiveReport|Test.*Judge.*Repair' -count=1; python3 -m pytest scripts/lint/prompt_lint_test.py -q -k report_v020_direct_semantics; python3 -m pytest scripts/lint/scenario_env_contract_test.py -q -k p0_100; make lint-prompts" result="PASS; evalkit validates same-source schema, one bounded repair, aggregate meta and one-shot judge; prompt label bounds/fragments pass; P0.100 contract 8/8 pass" -->
- [x] 8.8 MAX4 GENERATION RED-GREEN: independent in-memory generation budget=4；each call full-validates and dynamically selects action_labels/whole_report；provider transient and invalid attempts can recover on2/3/4；attempt4 invalid/provider failure terminal；nonretryable zero retry；aggregate usage/latency + attempt/retry/reason/scope manifest。
  <!-- historical-failure-superseded: 2026-07-13 earlier P0.100 classified label>120 as generic schema-invalid/whole_report because scope selection preceded normalization；retained as diagnosis and superseded by full-validator/max4 tests plus historical run 35103. -->
  <!-- verified: 2026-07-13 method=max4-generation-tests+live evidence="attempt2/3/4 recovery, attempt4 terminal, nonretryable zero-retry, dynamic action_labels/whole_report and aggregate manifest PASS; final-prompt run 59381 produced nine mechanically valid finals" -->
  - [x] GENERIC-REPLAY-RUBRIC RED-GREEN: `e2e-p0-100-20260713T011140Z-36625` short-conservative attempt1 returned `invalid_partial` at `$.nextActions[0]`；after excluding test/environment/consumer drift，judge instruction and `report_action_quality` rubric now share the exact answer_depth/answer_relevance generic-replay exceptions，migration + active DB are synchronized，and the same-case retest passes weighted0.82/min0.70 with zero violations。The full matrix remained pending at this focused point and was later closed by run 35103。
  - [x] FULL-VALIDATOR IMPLEMENTATION-HANDOFF: run80338 exposed an incompatible needs_practice retry+next escape；run35103 is historical strict PASS for its prompt, and final run59381 proves all nine emitted outputs pass the corrected full validator。
- [x] 8.9 Re-emit `REPORT_RUBRIC_V020_PASS` and `REPORT_CONTEXT_AWARE_EVAL_PASS` after corrected 8.8 code/static/offline gates；keep probabilistic live product acceptance and the stricter P0.100 diagnostic as separately reported evidence。
  <!-- verified: 2026-07-13 evidence="max4/full-validator tests, prompt/rubric single source and offline 28/28 PASS; final-prompt run 59381 is reported honestly as mechanical9/9, semantic8/9, cases4/5 and strict P0.100 FAIL" -->
- [x] 8.10 MAX4 JUDGE RED-GREEN: independent judge budget=4；retry provider transient and protocol/schema invalid；valid unsupported/causal/zero-tolerance/critical negative emits typed content rejection and exactly zero retry；manifest aggregates usage/latency and keeps protocol-invalid distinct。
  <!-- verified: 2026-07-13 method=max4-judge-tests+live evidence="provider/protocol retry and typed terminal content-rejection boundaries PASS; aggregate judge manifest remains protocol/content-distinct; final run terminally rejected the ninth unsupported summary without retry" -->

## 9 Real multilingual reliability and independent review

- [x] 9.1 Consume A3 `JUDGE_FINAL_CONTENT_V120_PASS`; judge adapter/profile tests prove non-thinking JSON, 6144 final budget, stop/non-empty live smoke and privacy-safe reasoning-only fail-close. F3 does not duplicate A3 wire/profile ownership.
  <!-- verified: 2026-07-12 evidence="completion stop 1928/1230; judge v1.2.0 stop 1318/791, 2999 bytes, weighted=1, item=7, causal=1, zero-tolerance=0, critical=true" -->
- [x] 9.2 P0.100 static contract requires per-case language (partial zh-CN, others en), exact focus/action gates, en<=24-whitespace-word and zh-CN<=64-Unicode-code-point action labels, targeted repair18/52 margin, multi-focus per-code semicolon fragments, bounded redacted structural counts/digests, 5 representative raw packets in OS 0700/0600 temp storage, and an independent Agent audit bound to the same context/output digests without access to judge verdicts；pre-judge failure has zero judge calls。Previous 14/40 evidence remains historical；a static contract PASS does not imply a live semantic PASS。
  <!-- verified: 2026-07-13 commands="python3 -m pytest scripts/lint/scenario_env_contract_test.py -q; python3 -m pytest scripts/lint/scenario_script_contract_test.py -q" result="privacy/static contracts 38 PASS; scenario script contracts 9 PASS; exact focus/action/zero-judge/redaction/blind-audit boundaries pass" -->
  - [x] INJECTION-GROUNDING REGRESSION: `e2e-p0-100-20260713T012359Z-59906` now enforces summary-clause candidate-evidence mapping、no action upgrade to undeclared quality attributes and W exact readiness；direct injection judge3x PASS。
  - [x] EXACT-GENERIC-EMPTY-FOCUS REGRESSION: `e2e-p0-100-20260713T013642Z-75753` repetition3 misjudge is fixed；same digest plus5 executions PASS。
- [x] 9.2a Preserve `25849` as aborted/not-PASS and `35103` as a historical strict PASS for its then-current prompt；neither is current final-prompt evidence。
  <!-- verified: 2026-07-13 evidence="historical runs remain attributed to their exact prompt/contract and are not promoted to current evidence" -->
- [x] 9.3 SCHEME-A PRODUCT ACCEPTANCE: for the final prompt, every emitted final output must pass deterministic schema/200/24/64/focus/action/cross-field validation；semantic confidence uses the fixed five representative case categories without replacement and accepts at least four. The stricter 11/11 P0.100 runner remains diagnostic and must retain FAIL if it terminates early or skips blind review。P0.099 independently proves desktop+390 UX。
  <!-- verified: 2026-07-13 run=e2e-p0-100-20260713T101214Z-59381 evidence="mechanical 9/9; semantic judge 8/9; fixed representative cases 4/5=80%; ninth injection summary unsupported; strict P0.100 FAIL; privacy cleanup PASS" -->
