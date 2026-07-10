# F3 Real Model Profile and Evals

> **版本**: 1.7
> **状态**: completed
> **更新日期**: 2026-07-10

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

维护 F3 当前离线评估闭环：`judge.default` 使用真实 judge profile，`LLMJudge` 按 rubric dimension 返回 `[]Score + Reasoning`，Promptfoo 通过 registry single-source 消费业务 prompts，`config/evals/` 提供 36 个录制 fixture case，`make eval-offline` 默认零网络运行。

本 plan 不新增用户可见 UI，不改变 HTTP API，不进入业务请求链路，也不拥有 A3 chat 业务 profile 的 status。它只维护 judge/eval 内部合同、profile coverage gate、Promptfoo runner、eval fixture 和相关 Make/lint gate。

## 2 当前合同

- `backend/internal/ai/registry.Judge` 返回 `([]Score, Reasoning, error)`；每个 `Score` 对应一个 rubric dimension，`Value` 位于 `[0,1]`。
- `LLMJudge` 注入 rubric provider 与 judge model client，通过 `judge.default` profile 调用 judge capability；profile 不可用、output schema invalid、模型输出无法解析、维度缺失或维度不匹配时 fail-close。
- `config/ai-profiles.yaml` 中 `judge.default` 当前为 `active`，provider ref 为 `judge-deepseek`，model 为 `deepseek-v4-pro`；当前 chat business profiles 保持各自 A3 owner 状态，只由 coverage gate 校验 runnable provider / model。
- `config/ai-providers.yaml` 提供 `judge-deepseek`，protocol 为 `judge_compatible`，capability 为 `judge`。
- `config/evals/` 覆盖 9 个 chat feature_key，共 36 个录制 fixture case，并包含 `en -> multi` fallback 覆盖。
- Promptfoo pinned version 为 `0.121.12`，通过仓库依赖执行；eval output、Promptfoo config、logs 和 state 默认写入 `.test-output/evals/`，可由 `EVAL_OUTPUT_DIR` 覆盖。
- `make eval-offline` 默认使用录制 fixture，不发网络请求；`EVAL_LIVE=1` 才允许真实 provider 调用，且不纳入 `make test`。
- Promptfoo custom provider / grader 只能通过 registry resolved prompt、schema、rubric 与 `LLMJudge`；不得复制业务 prompt 正文。

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

## 5 验收标准

- Judge interface, LLMJudge, judge adapter and profile catalog focused Go tests pass.
- `make lint-ai-profile-coverage` proves `judge.default` active, judge provider protocol/capability valid, and current chat profiles runnable.
- `make eval-offline` runs 36 cases without network in default mode.
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

## 7 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-07-10 | 1.7 | 删除零消费者 judge profile override 与恒定 profile 字段，直接使用已锁定的 `judge.default`。 | tech-debt pruning |
| 2026-07-10 | 1.6 | 简化 eval score wire conversion，修复 current-schema fixtures / resolved prompt projection，并补齐 eval package / command 的 owner discovery。 | tech-debt pruning |
| 2026-07-10 | 1.5 | 同步当前 `FailClosedJudge` / `LLMJudge` 代码事实，并将 profile coverage 表述收敛为 runnable / non-runnable marker。 | tech-debt pruning |
| 2026-07-07 | 1.4 | 压缩 owner 文档为当前 judge.default active、LLMJudge、36-case eval-offline and Promptfoo single-source contract。 | product-scope/001-core-loop-module-pruning |
| 2026-05-24 | 1.3 | 完成 real model profile、judge implementation and eval-offline delivery。 | prompt-rubric-registry |
