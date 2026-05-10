# Backend Practice 001 L2 Remediation 交付复盘报告

> **日期**: 2026-05-10
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`backend-practice/001-plan-and-session-orchestration` L2 code review remediation，覆盖首题 parser、`startPracticeSession` success replay snapshot、`000003` down migration ownership，以及 follow-up review 暴露的 reservation cleanup、generic non-2xx idempotency recovery、custom start-session TTL expiry、首问 template render。
- Plan 原地修订：`plan.md` / `checklist.md` 升至 1.2，追加并完成 `0.11` / `1.10` / `1.11` / `2.10` / `2.11` / `2.12` remediation checklist，最终恢复 `completed`。
- 验证证据：backend practice 相关 Go 包全绿；`make validate-fixtures`、migration/preflight/legacy lint、`make codegen-check`、`make lint-events` 全绿；`E2E.P0.022` ~ `E2E.P0.026` scenario setup / trigger / verify / cleanup 全绿；`sync-doc-index --check` zero drift。
- Follow-up 验证证据：`go test ./internal/practice -run 'TestStartPracticeSession(RunsThreeStepFlowWithAIOutsideTransactions|FailsReservationWhenPromptResolutionFails|RejectsMissingFirstQuestionText|RejectsNonJSONFirstQuestionResponse)' -count=1`、`go test ./internal/middleware/idempotency -run 'TestMiddleware(FinalizesNon2xxAndAllowsCorrectedSameKey|ReplaysSucceededResponse|RejectsFingerprintMismatchWithoutSideEffect|ExpiresRecordsByTTL)' -count=1`、`go test ./internal/store/practice -run 'TestSQLRepositoryReserveSessionStart(ResetsExpired|ReusesFailedRetryableRecord|ReplaysStoredResponseBody|ScopesIdempotencyByUser|RejectsFingerprintMismatch|RejectsConcurrentPendingRecord|MapsActivePlanUniqueViolationToConflict)' -count=1`、`go test ./internal/practice/... ./internal/middleware/idempotency/... ./internal/store/practice/... ./internal/api/practice/... ./cmd/api` 均 PASS。
- Bug 记录：更新 [BUG-0033](../bugs/BUG-0033.md)，记录本次 L2 finding 与修复证据。

## 2 会话中的主要阻点/痛点

- Completed plan 需要重新进入 TDD 修复路径。
  - **证据**：原 checklist 全部已勾选，必须先按 create-doc/tdd 规则将原 plan 调回 `active` 并追加 remediation items，才能执行红绿流程。
  - **影响**：增加一次文档生命周期操作，但避免了绕过 checklist 的直接修复。
- Scenario 脚本入口第一次执行路径假设错误。
  - **证据**：最初执行根级 `./setup.sh` 报 `no such file or directory`，随后按 `scripts/setup.sh` / `scripts/trigger.sh` / `scripts/verify.sh` / `scripts/cleanup.sh` 通过。
  - **影响**：轻微验证返工；仓库 README 已写明场景契约，无需修改仓库流程资产。
- 原 gate 对语义 drift 覆盖不足。
  - **证据**：历史 gate 覆盖了“有首题”“无重复副作用”“migration up contract”，但没有覆盖 F3 prompt JSON keys、replay stored snapshot、migration down ownership。
  - **影响**：L2 review 才发现真实 provider 输出与 replay/migration rollback 语义缺口。
- Follow-up review 继续暴露 idempotency/error-path 分支覆盖不足。
  - **证据**：新增 review finding 指出 `ResolveActive` 失败未 cleanup、shared middleware 非 2xx response 不 finalize、custom start-session reservation 不读 `expires_at`、首问 template placeholders 未渲染。
  - **影响**：需要把 plan 从 completed 再次调回 active，并补充 R21-R24 coverage matrix 与 focused tests。
- Store interface 扩展后测试替身容易漏改。
  - **证据**：`go test ./...` 首次暴露 `cmd/api/practice_http_scenario_test.go` 的 `scenarioPracticeStore` 缺少 `MarkFailed`，需要同步 fake store 的 shared idempotency 状态机。
  - **影响**：多跑一次全量编译；相关包测试最终通过。
- 全量 backend test suite 仍有无关残留失败。
  - **证据**：`go test ./...` 当前只剩 `internal/api/mockruntime TestHandlerSelectsNamedSeedScenariosAndFailsUnknown/missing-session`，期望 401 但实际返回 `404 PRACTICE_SESSION_NOT_FOUND`。
  - **影响**：本次 practice remediation 的相关包与 `cmd/api` 已通过，但 full-suite 绿灯需要单独处理 mockruntime contract drift。

## 3 根因归类

- Parser 与 prompt truth source 没有同测。
  - **类别**：spec-plan
- Prompt template placeholders 与 runtime payload render 没有同测。
  - **类别**：spec-plan
- Idempotency replay gate 只验证副作用数量，没有制造 mutable state drift。
  - **类别**：spec-plan
- Idempotency state-machine gate 没有覆盖 non-2xx terminal completion、registry resolution failure cleanup、custom reservation TTL expiry。
  - **类别**：spec-plan
- Migration baseline rebase 后缺少 down-path ownership negative test。
  - **类别**：spec-plan
- Scenario 脚本路径误执行属于本次操作失误。
  - **类别**：无需仓库改动
- `internal/api/mockruntime` 失败属于既存/旁路 contract drift，超出本次 practice idempotency remediation 范围。
  - **类别**：spec-plan

## 4 对流程资产的改进建议

- 后续涉及 AI prompt JSON 输出的 backend plan，应在 checklist 中明确“prompt markdown schema keys -> parser test” gate。
  - **落点**：spec-plan
  - **优先级**：high
- 后续涉及 AI prompt template 的 backend plan，应在 checklist 中明确“registry template placeholders -> runtime payload render test” gate，覆盖语言、目标岗位、技能、rubric 与 fallback。
  - **落点**：spec-plan
  - **优先级**：high
- 后续 idempotency success replay 应默认要求 stored response snapshot drift test，而不是只断言副作用不重复。
  - **落点**：spec-plan
  - **优先级**：high
- 后续 idempotency middleware 或 custom reservation path 变更，应默认列出状态机矩阵：pending、succeeded、failed_retryable、failed_terminal、expired、fingerprint mismatch，并要求 fake store 与 SQL store 同测。
  - **落点**：spec-plan
  - **优先级**：high
- migration rebase / integrator 模式中，down migration 必须有 owner zero-drop negative test。
  - **落点**：spec-plan
  - **优先级**：medium
- 单独开一个 mockruntime contract drift 修复入口，确认 `missing-session` 未认证请求应优先返回 401 还是 practice 404，并把 mockruntime scenario seed contract 与当前 auth/practice handler 顺序对齐。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：在 `backend-practice/002-event-loop-and-completion` 开始前，把 AI output parser schema gate、prompt template render gate、idempotency state-machine matrix、idempotency replay snapshot drift gate 写入该 plan 的 test/coverage matrix。
- 下一步修复：单独处理 `internal/api/mockruntime` 的 401/404 contract drift，让 `go test ./...` 恢复全绿。
- 可延后：若未来再次出现 migration ownership rebase，再把 down-path zero-drop 检查提升为共享 migration lint rule；本轮已有 focused contract test 覆盖当前缺口。
