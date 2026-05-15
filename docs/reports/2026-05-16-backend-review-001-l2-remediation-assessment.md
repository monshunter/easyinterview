# Backend Review 001 L2 Remediation 交付复盘报告

> **日期**: 2026-05-16
> **审查人**: Codex

## 1 复盘范围与成功证据

本次复盘覆盖 `$plan-code-review backend-review/001 --fix` 的第二轮 L2 remediation：F3 prompt 输出 contract、逐题 assessment status fallback、report 持久化状态事务、target report list 用户边界，以及对应 Go HTTP BDD scenario 补强。

成功证据：

- `cd backend && go test ./internal/ai/registry -run TestF3ReportGenerateAndAssessmentPreflight -count=1`
- `cd backend && go test ./internal/review -run TestAssessQuestionsMapsScoreLevelToWireStatus -count=1`
- `cd backend && go test ./internal/store/review -run 'TestListTargetJobReportsCursorPagination|TestListTargetJobReportsRequiresOwnedTarget' -count=1`
- `cd backend && env 'DATABASE_URL=postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable' go test -tags=integration ./internal/store/review -run 'TestPersistReportRejectsStaleStatusAndRollsBack|TestPersistReportFailureRejectsStaleStatusAndRollsBack' -count=1 -v`
- `cd backend && go test ./cmd/api -run 'TestE2EP0053ReportReadAndListing|TestE2EP0055ReportPrivacyAndLegacy' -count=1`
- `cd backend && go test ./internal/ai/registry ./internal/review ./internal/store/review ./cmd/api -count=1`
- `cd backend && env 'DATABASE_URL=postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable' go test -tags=integration ./internal/store/review -count=1`
- `python3 scripts/lint/prompt_lint.py`
- `python3 scripts/lint/backend_review_legacy.py --repo-root . --phase all`
- `python3 -m pytest scripts/lint/backend_review_legacy_test.py -q`
- `git diff --check`

关联 Bug：[BUG-0062](../bugs/BUG-0062.md)。

## 2 会话中的主要阻点/痛点

- Prompt lint 通过但没有证明 prompt 输出字段能被 runtime 消费。
  - **证据**：`report.generate` prompt 使用 `strengths/gaps`，runtime 只读取 `highlights/issues`；`report.question_assessment` prompt 使用 `dimension_scores`，runtime 读取 `dimension_results`。
  - **影响**：真实模型按 prompt 输出时，报告证据和逐题 assessment 可能为空或被判 invalid。

- Store transaction tests 缺少 zero-row state transition negative case。
  - **证据**：新增 stale-ready integration tests 前，`PersistReport` 和 `PersistReportFailure` 在 `feedback_reports` 已是 `ready` 时仍返回 nil。
  - **影响**：旧 job 可能在 report 已终态后继续插入 question assessments、outbox、audit 或改写 async job。

- BDD list privacy 只覆盖 report ID，不覆盖 target ID。
  - **证据**：原 `E2E.P0.053/P0.055` 只测跨用户 `GET /reports/{reportId}`，没有测 `GET /targets/{targetJobId}/reports`。
  - **影响**：跨用户 target list 会返回 200 空列表，破坏 targetjob boundary 的 not-found 语义。

## 3 根因归类

- Prompt output schema 没有进入 L2 preflight invariant。
  - **类别**：spec-plan
  - **说明**：当前 plan 强调 prompt/rubric registry、safe input 和 provenance，但缺少“prompt 输出 key 必须与 runtime draft struct 对齐”的可执行 gate。

- Store 状态机只验证 happy path 和 retry path，没有验证 zero-row update rollback。
  - **类别**：spec-plan
  - **说明**：checklist 写到事务回滚，但测试没有覆盖 stale status 导致 `RowsAffected=0` 的场景。

- BDD scenario 对 target boundary 的负向覆盖不完整。
  - **类别**：spec-plan
  - **说明**：read by report ID 与 list by target ID 是两个入口，隐私边界需要分别断言。

## 4 对流程资产的改进建议

- 在 backend-review plan 的 F3 / report generation gate 中增加 prompt output schema preflight。
  - **落点**：spec-plan
  - **优先级**：high
  - **建议内容**：每个 prompt feature key 至少有一个测试断言 runtime 必需 output keys；prompt body、YAML hash、seed migration 必须同步更新。

- 在 store persistence checklist 中把 terminal-state zero-row update 作为必测项。
  - **落点**：spec-plan
  - **优先级**：high
  - **建议内容**：success/failure persistence 都要有 stale terminal status rollback integration test，并断言 outbox/audit/async job 不被写入。

- 在 BDD checklist 中把 target-scoped list privacy 与 resource ID privacy 分开列项。
  - **落点**：spec-plan
  - **优先级**：medium
  - **建议内容**：`GET /reports/{reportId}` 和 `GET /targets/{targetJobId}/reports` 各自需要 cross-user not-found scenario。

## 5 建议优先级与后续动作

最高优先级是把 prompt output schema preflight 和 stale-state rollback case 固化回 `backend-review/001` 的 checklist/test checklist，避免后续报告能力扩展时重复出现“prompt registry 绿但 runtime contract 漂移”的问题。

中优先级是把 target-scoped list privacy 作为 BDD 模板化条目，后续所有 target-owned list API 都应同时覆盖 owner target 与 cross-user target 两种路径。
