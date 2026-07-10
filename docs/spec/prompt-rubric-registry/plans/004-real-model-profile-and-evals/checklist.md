# F3 Real Model Profile and Evals Checklist

> **版本**: 1.6
> **状态**: completed
> **更新日期**: 2026-07-10

**关联计划**: [plan](./plan.md)

## 1 Judge contract

- [x] 1.1 `Judge` interface returns per-rubric-dimension `[]Score` plus `Reasoning`.
- [x] 1.2 `FailClosedJudge` remains a safe default and returns `ErrJudgeUnavailable`.
- [x] 1.3 `LLMJudge` validates evaluated output schema, calls judge capability through `judge.default`, returns one score per rubric dimension and fail-closes malformed judge output.

## 2 Judge profile and provider coverage

- [x] 2.1 `judge.default` is active in `config/ai-profiles.yaml`.
- [x] 2.2 `judge.default` routes to runnable `judge-deepseek` / `deepseek-v4-pro`.
- [x] 2.3 `config/ai-providers.yaml` includes `judge-deepseek` with `judge_compatible` protocol and `judge` capability.
- [x] 2.4 `make lint-ai-profile-coverage` verifies judge profile/provider alignment and rejects non-runnable provider/model markers for judge and current chat business profiles.

## 3 Eval fixtures and Promptfoo

- [x] 3.1 `config/evals/` covers current 9 chat feature_key values with 36 recorded fixture cases.
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
