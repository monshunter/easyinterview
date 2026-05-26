# Privacy Account Identity Cleanup 交付复盘报告

> **日期**: 2026-05-26
> **审查人**: Codex

**关联计划**: [backend-async-runner/001 Internal Job and Outbox Runner](../spec/backend-async-runner/plans/001-internal-job-outbox-runner/plan.md)
**关联 Bug**: [BUG-0106](../bugs/BUG-0106.md)

## 1 复盘范围与成功证据

- 本次交付范围是 `BUG-0106` 的 privacy delete account identity cleanup 修复：`DELETE /api/v1/me` 受理期同步软删用户并撤销全部 session，`privacy_delete` runner 在 domain cleanup 全部成功后 hard delete 账户身份，同时保留 `privacy_requests` tombstone。
- 成功证据包括 focused regression：`go test ./backend/internal/auth ./backend/internal/privacy/runner ./backend/cmd/api -run '^(TestDeleteMeSoftDeletesUserRevokesAllSessionsAndCreatesPrivacyHandoff|TestSQLStorePrivacyDeleteHandoffSoftDeletesUserAndRevokesSessions|TestSQLStoreMarkDeleteRequestCompletedDeletesAccountIdentityAndPreservesRequestTombstone|TestPrivacyDeleteHandlerHardDeletesAccountIdentityAfterDomainCleanup|TestPrivacyDeleteHandlerDomainFailureKeepsAccountIdentityForRetry|TestPrivacyDeleteRemovesAccountIdentityAfterJobCompletion)$' -count=1` PASS。
- 模块回归与契约证据包括：`go test ./backend/internal/auth ./backend/internal/privacy/runner -count=1` PASS；`go test ./backend/internal/migrations -count=1` PASS；`DATABASE_URL='postgres://easyinterview:***@localhost:5432/easyinterview?sslmode=disable' make migrate-check` PASS；local migration status 为 `version=11 dirty=false`。
- 文档与 drift gate 证据包括：`validate_context.py --context docs/spec/backend-async-runner/plans/001-internal-job-outbox-runner/context.yaml --docs-root docs --target backend` PASS；`sync-doc-index.py --check` PASS；`make docs-check` PASS；`make lint-runner-legacy` PASS；`git diff --check` PASS。
- 已知非本次范围残余：`go test ./backend/cmd/api -run '^TestE2EP0050PracticeAssistantActionProvenanceAndTaskRuns$' -count=1` 当前仍因 practice task-run 数量 / feature key 断言失败，未作为 BUG-0106 的通过证据。

## 2 会话中的主要阻点/痛点

- cleanup 完成语义原先只覆盖 `privacy_requests` / `async_jobs` terminal state，没有覆盖账户身份字段残留。
  - **证据**：manual UAT cleanup 已看到 request/job completed，但 `users.email='manual-uat-full-funnel@example.test'` 仍存在且 `deleted_at` 为空。
  - **影响**：旧 gate 会把隐私删除误判为完成，必须补用户表字段级断言。
- hard delete 用户身份与 `privacy_requests` tombstone 保存存在 schema 冲突。
  - **证据**：既有 `privacy_requests.user_id NOT NULL REFERENCES users(id) ON DELETE CASCADE` 会在删除用户时级联删除 request 记录。
  - **影响**：修复不能只改 runner，还必须新增 migration，把 tombstone 外键改成 nullable + `ON DELETE SET NULL`。
- `cmd/api` 包存在与本 BUG 无关的当前失败用例，不能用整包全绿作为本次隐私修复的唯一 close-out gate。
  - **证据**：`TestE2EP0050PracticeAssistantActionProvenanceAndTaskRuns` 单测失败点在 practice task-run 数量 / feature key 断言，本次未触碰 practice 逻辑。
  - **影响**：需要把 focused privacy runtime wiring test、auth/privacy 包回归和 migration/doc gate 作为本次完成证据，同时显式记录残余风险。

## 3 根因归类

- privacy delete 的 completion contract 缺少账户身份层最小不变量。
  - **类别**：spec-plan
- DB baseline 没有提前表达“保留 request tombstone 但允许最终删除 user row”的 FK 关系。
  - **类别**：spec-plan
- manual UAT cleanup gate 对 PII 残留字段反查不足。
  - **类别**：spec-plan
- unrelated `cmd/api` practice 用例失败属于当前仓库测试健康度问题，但不是本次 privacy 修复的根因。
  - **类别**：无需仓库改动

## 4 对流程资产的改进建议

- 后续 privacy / account deletion 计划必须固定字段级负向断言：`users.email` 不得残留，`users.deleted_at` 在受理期必须同步设置，runner 成功后 `privacy_requests.user_id` 必须为 null 且 tombstone 保留。
  - **落点**：spec-plan
  - **优先级**：high
- DB migration baseline 应把隐私 tombstone FK 关系作为显式不变量，避免未来恢复 `ON DELETE CASCADE` 或 `NOT NULL user_id`。
  - **落点**：spec-plan
  - **优先级**：high
- manual UAT cleanup 材料的最终验收应从 job/request terminal state 扩展到关键 PII 字段反查。
  - **落点**：spec-plan
  - **优先级**：medium
- 对当前无关 practice test failure 单独进入 practice owner triage，不要在 privacy 修复中扩大范围。
  - **落点**：no repo change needed
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：下一轮由 `practice` / `cmd-api` owner 处理 `TestE2EP0050PracticeAssistantActionProvenanceAndTaskRuns` 的 task-run 数量 / feature key 断言漂移，避免它继续污染后续 backend runtime close-out。
- 次优先级：在下一次 manual UAT cleanup plan 中复用本次 BUG-0106 的字段级 privacy gate，把 account identity cleanup 从“发现型检查”升级为固定验收项。
- 可延后：抽一个轻量 SQL smoke 专门检查已完成 privacy delete request 的 tombstone 与账户身份残留，作为 future manual UAT cleanup 的辅助脚本。
