# Backend TargetJob L2 Remediation 交付复盘报告

> **日期**: 2026-05-08
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`backend-targetjob/001-targetjob-import-and-parse-bootstrap` backend L2 remediation，修复 import idempotency race、AI/F3 failure evidence 脱敏、A3 `dataSourceVersion` metadata 缺失、handoff doc stale BDD 状态，以及后续负向搜索暴露的 completed operation `NOT_IMPLEMENTED` stub 残留。
- 成功证据：
  - `cd backend && go test ./internal/targetjob -run 'TestSQLStore_ImportTargetJob|TestParseExecutor_HappyPath|TestParseExecutor_RedactsPromptResponseAndProviderSecretInErrorMessage|TestParseExecutor_RedactsForbiddenTokensInErrorMessage|TestPackageDocReflectsCompletedScenarioGateState|TestHandlerMissingServiceReturnsInternalConfigurationError' -count=1`
  - `cd backend && go test ./internal/targetjob/... ./cmd/api ./internal/ai/aiclient/... -count=1`
  - `cd backend && go build ./...`
  - `make codegen-check`
  - `make validate-fixtures`
  - `make lint-events`
  - `make lint-config`
  - `python3 scripts/lint/migrations_lint.py --repo-root /Users/tanzhangyu/Documents/my-opensources/easyinterview`
  - `make docs-check`
  - `git diff --check`
  - `test/scenarios/e2e/p0-010-targetjob-text-import-parse-ready/scripts/{setup,trigger,verify,cleanup}.sh`
  - `test/scenarios/e2e/p0-011-targetjob-url-import-fetch-and-parse/scripts/{setup,trigger,verify,cleanup}.sh`
  - `test/scenarios/e2e/p0-012-targetjob-parse-failure-retryable/scripts/{setup,trigger,verify,cleanup}.sh`
  - `test/scenarios/e2e/p0-013-targetjob-manual-form-ready/scripts/{setup,trigger,verify,cleanup}.sh`
  - 最终场景 run id: `.test-output/runs/targetjob-plan-code-review-fix-final-20260508T090351Z`
- 关联 Bug：[BUG-0026](../bugs/BUG-0026.md)。

## 2 会话中的主要阻点/痛点

- 历史 PASS 没有覆盖 manual_form terminal marker 的并发 dedupe 边界。
  - **证据**：L2 review 发现 `ImportTargetJob` 先在事务外查 dedupe，再进入事务写入；manual_form 不受 active runner partial unique index 保护。
  - **影响**：需要补事务内 advisory lock 断言和 SQL 顺序 regression test。
- AI evidence contract 只验证了部分 provenance 字段。
  - **证据**：新增 `dataSourceVersion` 断言后，red test 先报 `CallMetadata` 缺字段。
  - **影响**：A3 调用链缺少 F3 data source 版本证据，历史 happy-path test 无法发现。
- Failure path 把 upstream error 当作可持久化证据。
  - **证据**：新增 prompt / response / provider secret redline test 后，旧路径会把原始错误文本进入 outcome / async job failure message。
  - **影响**：需要统一 code-based safe summary，避免持久化敏感 AI 错误上下文。
- 负向搜索第一轮仍偏向旧 token 清单，漏掉 completed operation stub。
  - **证据**：四项修复后再搜 `NOT_IMPLEMENTED` / `not yet implemented` 才发现 handler 仍保留 Phase stub fallback。
  - **影响**：追加修复 handler 缺 service 行为，并补非测试代码零残留搜索。
- 聚合 gate 的 shell 环境曾经误导验证。
  - **证据**：用 `bash -lc` 运行 `make validate-fixtures` 时命中 `/usr/bin/python3` 缺 `yaml`；默认 shell 下 `/opt/homebrew/bin/python3` 同一 gate 通过。
  - **影响**：多跑了一轮环境定位；最终验证改为按仓库默认 shell 分别执行 make gate。

## 3 根因归类

- 并发幂等 gate 没有要求 read / lock / write 同事务。
  - **类别**：spec-plan
- AI payload metadata gate 没有列全 F3 provenance 字段。
  - **类别**：spec-plan
- Failure evidence redline 没有覆盖 upstream wrapper error 的自由文本。
  - **类别**：spec-plan
- `/plan-code-review` 的 active-scope 搜索需要包含 completed operation stub / stale phase fallback 这类非业务 token。
  - **类别**：skill
- `bash -lc` PATH 与仓库默认 shell 不一致导致 Python 依赖误判。
  - **类别**：no repo change needed

## 4 对流程资产的改进建议

- 在 backend-targetjob plan 后续修订中，把 import idempotency gate 写成“same user / same key 的 dedupe read、advisory lock、write 必须同事务”，并保留 SQL 顺序测试。
  - **落点**：spec-plan
  - **优先级**：high
- 在所有 F3 → A3 调用计划中，把 `promptVersion` / `rubricVersion` / `modelProfile` / `dataSourceVersion` 作为完整 metadata matrix，不允许只断言部分字段。
  - **落点**：spec-plan
  - **优先级**：high
- 为 parse / AI failure evidence 增加固定 gate：持久化错误消息必须是 code-based safe summary，禁止直接使用 upstream `err.Error()`。
  - **落点**：spec-plan
  - **优先级**：high
- 给 `/plan-code-review` deep reconcile checklist 增加 completed-operation residual search：`NOT_IMPLEMENTED`、`not yet implemented`、`stub`、旧 phase 注释与缺 service fallback 需要进入 L2 搜索口径。
  - **落点**：skill
  - **优先级**：medium
- 后续本地聚合 gate 优先使用仓库默认 shell 或显式 PATH；不要用 `bash -lc` 替换用户 shell 后直接判定 Python 依赖缺失。
  - **落点**：no repo change needed
  - **优先级**：low

## 5 建议优先级与后续动作

- 优先：对 `backend-targetjob` plan 做一次 L1 文档修订，把本轮新增的并发锁、A3 metadata、安全失败摘要和 stale stub 负向搜索沉淀为 plan/checklist gate。
- 次优先：在后续 F3/A3 owner plan review 中复用 metadata matrix，检查是否还有只断言 prompt / rubric 而遗漏 `dataSourceVersion` 的调用链。
- 可延后：将 completed-operation residual search 写入 `/plan-code-review` skill；当前本报告和 BUG-0026 已先记录可执行建议。
