# Backend Review 001 Report Generation Baseline 交付复盘报告

> **日期**: 2026-05-16
> **审查人**: Codex

## 1 复盘范围与成功证据

本次复盘覆盖 `backend-review/001-report-generation-baseline` 的端到端实施：报告生成 inline runner、lease / reaper、F3 / A3 报告与逐题 assessment 编排、readiness / retry-focus / next-action 算法、报告持久化 / outbox / audit、`getFeedbackReport` / `listTargetJobReports` 读取 API、OpenAPI / conventions / migration / fixture / frontend contract generated artifacts，以及 `E2E.P0.052~055` Go HTTP BDD scenario。

成功证据：

- `cd backend && go test ./... -count=1`
- `cd backend && go test ./internal/review ./internal/store/review ./cmd/api -count=1`
- `cd backend && env 'DATABASE_URL=postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable' go test -tags=integration ./internal/store/review -run 'TestPersistReportWritesQuestionAssessments|TestPersistReportFailureRetryAndPermanent|TestLeaseSkipLocked' -count=1 -v`
- `make codegen-check`
- `make validate-fixtures`
- `migrations/lint.sh`
- `make lint-events`
- `make codegen-events-check`
- `python3 scripts/lint/conventions_drift.py --repo-root .`
- `python3 scripts/lint/prompt_lint.py`
- `python3 scripts/lint/rubric_lint.py`
- `python3 scripts/lint/backend_review_legacy.py --repo-root . --phase all`
- `python3 -m pytest scripts/lint/backend_review_legacy_test.py -q`
- `pnpm --filter @easyinterview/frontend typecheck`
- `make docs-check`
- `git diff --check`

最终提交：`feat(backend-review): implement report generation baseline`。

## 2 会话中的主要阻点/痛点

- Runner 与 service 的 `async_jobs` finalization 所有权一度不清晰。
  - **证据**：收口后复查发现 `Runner.RunOnce` 会根据 `ReportOutcome` 调用 `UpdateAsyncJobSucceeded/Failed`，而真实 `Service.GenerateReport` 已在 `PersistReportResult/PersistReportFailure` 中更新同一条 `async_jobs`。修正为 `ReportOutcome.AsyncJobFinalized` 后，原 `E2E.P0.052` fake 仍断言 runner 二次写 `succeededJobID`，测试失败并暴露旧断言与新所有权冲突。
  - **影响**：若未在收尾复查中发现，真实路径会产生重复 async job 写入；虽然部分字段幂等，但失败路径会让职责边界变得脆弱。

- 大型跨契约计划的 codegen gate 需要明确 staged / generated artifact 顺序。
  - **证据**：`make codegen-check` 需要在 intended generated changes staged 后执行，避免 clean-worktree diff assertion 把预期生成物当成漂移；本次收口按该顺序执行才得到稳定通过。
  - **影响**：后续同类计划若未提前说明，容易把正常生成物误判为 codegen 漂移或在错误阶段运行全量 codegen gate。

- 本地集成测试命令中的 `DATABASE_URL` 需要 shell-safe 引号。
  - **证据**：第一次执行 `env DATABASE_URL=postgres://...?... go test` 被 zsh 报 `no matches found`，因为 `?sslmode=disable` 被当作 glob；加引号后同一测试通过。
  - **影响**：这是低成本的命令模板问题，但会浪费一次定位循环。

## 3 根因归类

- `async_jobs` finalization 所有权未被 plan/checklist 固化。
  - **类别**：spec-plan
  - **说明**：plan 明确了 runner、service、store 的存在，但没有把“真实 service persistence 成功后 runner 不得二次 finalization”写成可执行 gate。

- codegen gate 的 staged 语义依赖执行经验。
  - **类别**：skill / README
  - **说明**：实施流程知道要运行 codegen check，但没有把“generated artifact 已 staged 时再跑 clean-worktree 型 gate”的顺序写成显式提示。

- `DATABASE_URL` 的 shell quoting 属于命令示例健壮性问题。
  - **类别**：README
  - **说明**：不是功能缺陷；可通过 scenario / integration test README 中的命令模板避免重复踩坑。

## 4 对流程资产的改进建议

- 在 future `backend-async-runner` 或后续 `backend-review` plan gate 中新增 async job ownership invariant。
  - **落点**：spec-plan
  - **优先级**：high
  - **建议内容**：明确 runner 只负责占位 service 或未持久化 outcome 的 job finalization；真实 service 若已完成 `PersistReportResult/PersistReportFailure`，必须通过 outcome 字段阻止 runner 二次写 `async_jobs`，并配套 BDD / unit test 断言。

- 在 `/implement` 或 TDD phase close-out 提示中补充 generated artifact / codegen-check 顺序。
  - **落点**：skill
  - **优先级**：medium
  - **建议内容**：当计划涉及 OpenAPI、conventions、events、fixtures、generated clients 时，先确认 intended generated diff 已 staged，再运行带 clean-worktree 语义的 codegen check；否则应使用 targeted generator + diff review。

- 在本地集成测试 README 或 plan evidence 示例中使用带引号的 `DATABASE_URL`。
  - **落点**：README
  - **优先级**：low
  - **建议内容**：将示例写成 `env 'DATABASE_URL=postgres://...?...' go test ...`，避免 zsh glob 误拦截。

## 5 建议优先级与后续动作

最高优先级是把 async job finalization ownership 写入下一轮 backend async runner / backend-review 计划，因为它直接影响 report runner 与未来通用 runner 抽象的责任边界。

中优先级是强化 `/implement` 或 `/tdd` 的 generated artifact 收口提示，减少跨 OpenAPI / conventions / migrations / frontend generated clients 计划中的 gate 顺序误判。

低优先级是修正文档中的本地集成测试命令模板；这不影响当前交付，但能减少重复环境噪音。
