# Backend Async Runner Scheduler Backoff Fix 交付复盘报告

> **日期**: 2026-05-22
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付覆盖 `backend-async-runner/001-internal-job-outbox-runner` 的 L2 review 后续修复：runtime fresh finalize timestamp、production per-job lease loop 防止 `email_dispatch` starvation、`report_generate` failure 归一化到 kernel shared backoff、BUG-0088 与 owner plan/spec/test gate 同步。
- 已通过 focused regression：`cd backend && go test ./internal/runner ./internal/review ./cmd/api -run '^(TestRuntime_FinalizeUsesTimestampAfterHandlerReturns|TestRuntime_StartDoesNotLetCriticalJobStarveEmailDispatch|TestGenerateHandler_NormalizesFinalizedRetryableFailureThroughKernel|TestE2EP0052ReportGenerationHappyPath|TestE2EP0054ReportAIFailureAndRetry)$' -count=1`。
- 已通过后端包级回归：`cd backend && go test ./internal/runner/... ./internal/review/... ./internal/store/review ./cmd/api -count=1` 与 `cd backend && go test ./... -count=1`。
- 已通过 legacy gate：`make lint-runner-legacy`。Integration gate `go test -tags=integration ./internal/store/review -run '^TestPersistReportFailure' -count=1 -v` 可发现测试，但本机 `DATABASE_URL` 未设置，按仓库约定 skip。

## 2 会话中的主要阻点/痛点

- 原 L2 修复后仍遗漏 scheduler fairness 的生产语义。
  - **证据**：review comment 指出 `runOnce` fixed-priority synchronous drain 会让 low-priority `email_dispatch` 等待 long-running critical/default handler；此前只验证 priority bucket 选择，没有验证 handler 执行期间的并发扫描。
  - **影响**：magic-link 投递可能不满足一个 scan 周期可见的用户承诺。
- Shared backoff 的证明停在 runner package，没有反查 report domain failure path。
  - **证据**：`Service.GenerateReport` retryable failure 仍能通过 `AsyncJobFinalized` 让 kernel 跳过 finalize；`PersistReportFailure` 还持有 async job update/backoff 逻辑。
  - **影响**：`report_generate` 继续走旧 review-store retry schedule，完成态计划对“全部 handler 共用 shared BackoffPolicy”的证据不成立。
- 时间戳缺陷需要模拟长耗时 handler 才能暴露。
  - **证据**：instant handler happy path 不会发现 pre-handle `now` 被用于 retry `available_at` / terminal `completed_at`。
  - **影响**：审计时间和 retry schedule 会在 AI-backed 长调用失败时漂移。

## 3 根因归类

- Scheduler gate 覆盖粒度不足。
  - **类别**：spec-plan
  - **说明**：计划只写了 priority bucket 与 queueWeights 注入，没有把 long-running handler 与 low-priority user-visible job 的可见延迟作为必须测试的不变量。
- Domain-side finalize escape hatch 没有设退出条件。
  - **类别**：spec-plan / skill
  - **说明**：`AsyncJobFinalized` 是迁移期 escape hatch，但 code review gate 没有逐 handler 反查它是否仍绕过 shared retry policy。
- deterministic time regression 不够系统。
  - **类别**：skill
  - **说明**：异步 runner 审查需要把 handler 前后时间推进纳入 fixture，而不是只看行级计算是否合理。

## 4 对流程资产的改进建议

- 在 backend async/runtime 类 plan 的 test-plan 中强制列出 starvation / long-running handler gate。
  - **落点**：spec-plan
  - **优先级**：high
  - **建议**：凡计划承诺 scan-cycle SLA 或 queueWeights/fair scheduling，必须至少有一条测试让高优先级 handler 阻塞，并证明低优先级用户可见 job 仍会被处理。
- 增强 `/plan-code-review` 对 escape hatch 的反查。
  - **落点**：`.agent-skills/plan-code-review/SKILL.md`
  - **优先级**：high
  - **建议**：遇到 `AsyncJobFinalized`、`AlreadyHandled`、`SkipFinalize` 等 migration-only 字段时，必须反查每个 handler/service/store 是否仍绕过 owner kernel 或 shared policy。
- 为 runner/retry 审查增加 deterministic clock checklist。
  - **落点**：`.agent-skills/plan-code-review/SKILL.md`
  - **优先级**：medium
  - **建议**：涉及 retry/backoff/completed_at/locked_at 的代码，review gate 应覆盖 handler 执行前后时间变化，而不是只运行 instant handler happy path。

## 5 建议优先级与后续动作

- 最高优先级是把 scheduler starvation 与 escape-hatch 反查写进 `/plan-code-review`，它们直接对应这次 P2 review comments。
- 次优先级是在 backend-async-runner 后续计划中继续缩小 `AsyncJobFinalized` 的兼容面，最终让所有 handler 的 async job finalization 都由 kernel 单点负责。
