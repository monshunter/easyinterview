# Backend Review Follow-up Correctness Fix 交付复盘报告

> **日期**: 2026-05-26
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付修复三条 review P2 findings：OpenAI-compatible provider array schema 不再强制 object mode；report assessment fallback status 聚合与 map iteration 顺序无关；privacy delete completed tombstone retry 幂等返回成功。
- 成功证据：
  - `go test ./internal/ai/aiclient/providers/openai_compatible -run 'TestComplete_(RequestsJSONObjectWhenOutputSchemaIsPresent|LeavesResponseFormatUnsetForArrayOutputSchema)' -count=1 -v` PASS。
  - `go test ./internal/review -run 'TestOverallStatusFromDimensionResultsNeedsWorkDominatesMixedStatuses' -count=1 -v` PASS。
  - `go test ./internal/privacy/runner -run 'TestPrivacyDeleteHandlerCompletedTombstoneRetrySucceedsIdempotently|TestSQLStoreLookupDeleteRequestUserTreatsCompletedTombstoneAsIdempotent' -count=1 -v` PASS。
  - `go test ./internal/ai/aiclient/... ./internal/review/... ./internal/privacy/runner -count=1` PASS。
  - `go test ./cmd/api -run 'Test(BuildReportRuntime|BuildTargetJobRuntimeRegistersPrivacyDeleteHandler|PrivacyDeleteRemovesAccountIdentityAfterJobCompletion|JDMatchA3F3AdapterUsesRegistryProfilesForSearchAndRecommendation|E2EP0052ReportGenerationHappyPath|E2EP0054ReportAIFailureAndRetry)' -count=1 -v` PASS。

## 2 会话中的主要阻点/痛点

- 三条 review findings 分属 AI provider、review report、privacy runner 三个 owner surface。
  - **证据**：`change-intake` matcher 首选 backend-async-runner，但候选同时包含 backend-jobs-recommendations 与 backend-review；实际修复跨 `backend/internal/ai/aiclient`、`backend/internal/review`、`backend/internal/privacy/runner`。
  - **影响**：需要手动补齐多 owner focused tests，不能只按单一 plan owner 验证。
- `backend/cmd/api` 全包测试仍受已知残余 `TestE2EP0050PracticeAssistantActionProvenanceAndTaskRuns` 影响。
  - **证据**：扩大运行 `go test ./internal/ai/aiclient/... ./internal/review/... ./internal/privacy/runner ./cmd/api -count=1` 时，除本次修复触达的 privacy sqlmock 已同步外，仍失败于工作日志已记录的 E2EP0050 practice task-run 断言。
  - **影响**：本次只能用 narrower cmd/api production wiring / report / JD-Match / privacy regex 作为验收证据，无法把 `go test ./cmd/api` 全包作为最终绿灯。

## 3 根因归类

- 原始缺陷的根因是测试矩阵缺了三个边界：array schema provider wire payload、mixed dimension dominance、crash-after-completion retry。
  - **类别**：no repo change needed；本次已用 focused regression tests 固化。
- `cmd/api` 存在已知 unrelated failing test，削弱了 broad package gate 的可用性。
  - **类别**：spec-plan；应由 practice task-run owner 或当前已记录残余项继续收口。

## 4 对流程资产的改进建议

- 对 AI provider L2 review，增加“按 active output schema 顶层类型捕获 provider wire request”的检查点。
  - **落点**：`plan-code-review` skill 或 prompt-rubric/provider plan checklist。
  - **优先级**：medium。
- 对 tombstone / hard delete 类隐私修复，增加“domain cleanup committed but async job finalize not yet committed”的 retry-idempotence regression。
  - **落点**：backend-async-runner / privacy delete checklist。
  - **优先级**：medium。
- 单独修复或隔离 `TestE2EP0050PracticeAssistantActionProvenanceAndTaskRuns`，恢复 `backend/cmd/api` 全包测试作为宽门禁。
  - **落点**：practice task-run owner plan 或 BUG follow-up。
  - **优先级**：high。

## 5 建议优先级与后续动作

- 最高优先级：继续处理已知 `TestE2EP0050PracticeAssistantActionProvenanceAndTaskRuns` 残余，使 `go test ./cmd/api` 能重新作为 broad gate。
- 次优先级：在下一次 provider/schema L2 review 中，把 object schema 与 array schema 的 request-format 差异纳入固定检查。
