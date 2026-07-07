# Backend Review Spec

> **版本**: 1.3
> **状态**: active
> **更新日期**: 2026-07-07

## 1 背景与目标

`backend-review` 承接 P0 用户路径中 "练习完成 -> 报告生成 -> 复练当前轮 / 进入下一轮" 的后端报告域。上游 [`backend-practice`](../backend-practice/spec.md) 在 `completePracticeSession` 同事务创建 `feedback_reports(status='queued')`、`async_jobs(job_type='report_generate', status='queued')` 和 `practice.session.completed` source event；本域把 report job 推进到 `ready` 或 `failed`，并提供 `getFeedbackReport` / `listTargetJobReports` 两个读取入口。

当前合同由以下 truth source 共同约束：

- [B2 OpenAPI](../openapi-v1-contract/spec.md): `getFeedbackReport` / `listTargetJobReports`、`FeedbackReport`、`PaginatedFeedbackReport`、`QuestionAssessment` 和 `GenerationProvenance`。
- [B3 Event / Job](../event-and-outbox-contract/spec.md): `report.generation.requested` / `report.generated` / `report.generation.failed` 与 `report_generate` job；`report_generate` 使用 `triggerEventSemantic: source_event_only`，外部 dispatcher 不二次创建 report job。
- [B4 Migration](../db-migrations-baseline/spec.md): `feedback_reports`、`question_assessments`、`async_jobs`、`ai_task_runs`、`outbox_events` 和 `audit_events`；`feedback_reports` 当前包含 `language`、`feature_flag`、`data_source_version`、`retry_focus_turn_ids`；`ai_task_runs.task_type` 当前包含 `report_generate` 与 `report_assessment`。
- [B1 Shared Conventions](../shared-conventions-codified/spec.md): `ReportStatus`、`ReadinessTier`、`DimensionStatus`、`Confidence`、`QuestionReviewStatus`、`REPORT_NOT_FOUND` 和 AI 错误码族。
- [A3 AI Provider](../ai-provider-and-model-routing/spec.md) 与 [F3 Prompt / Rubric](../prompt-rubric-registry/spec.md): `report.generate`、`report.question_assessment`、`report.generate.default`、`report.assessment.default` 和 A3 observed `AIClient.Complete`。
- [backend-async-runner](../backend-async-runner/spec.md): current `runner.Runtime` kernel 负责 `report_generate` lease / retry / reaper 调度；本域保留 `review.GenerateHandler` 业务实现。

本 spec 不绕过 owner truth source。OpenAPI、shared enum/error、migration、event/job、AI profile 或 prompt/rubric 的任何契约变更必须先回对应 owner，再由 backend-review 消费。

## 2 范围

### 2.1 In Scope

- `GET /reports/{reportId}` `getFeedbackReport`
  - 使用 generated handler / response type。
  - 按 `(user_id, report_id)` 过滤；不存在或越权统一返回 `404 REPORT_NOT_FOUND`。
  - `queued` / `generating` 返回合法 placeholder：内容数组为空、`preparednessLevel=null`、`provenance=null`、`errorCode=null`。
  - `ready` 返回完整报告、逐题评估、retry focus、next action 与 6 字段 `GenerationProvenance`。
  - `failed` 返回空内容、`errorCode` 非空，供前端渲染失败态。
- `GET /targets/{targetJobId}/reports` `listTargetJobReports`
  - 先确认 target 属于当前用户；不存在或越权返回 `REPORT_NOT_FOUND`。
  - 按 `created_at DESC, id DESC` cursor 分页；`pageSize` 默认 20、最大 50。
  - 空列表返回 `{items: [], pageInfo: {nextCursor: null, pageSize, hasMore: false}}`。
- `report_generate` job 业务处理
  - 由 `runner.Runtime` lease 当前 `async_jobs` 行并调用 `review.GenerateHandler`。
  - Handler 将 `feedback_reports.status` 从 `queued` 推进到 `generating`。
  - 通过 F3 `report.generate` + A3 `AIClient.Complete` 生成 report draft。
  - 通过 F3 `report.question_assessment` + A3 `AIClient.Complete` 对每个 turn 生成 question assessment draft。
  - 计算 `ReadinessTier`、`retry_focus_turn_ids` 和首条 `next_actions[].type`。
  - 单事务持久化 `feedback_reports(status='ready')`、`question_assessments`、`report.generated` outbox、audit row，并把 `async_jobs` 置为 `succeeded`。
  - AI / prompt / parse 失败时持久化 `feedback_reports(status='failed', error_code=<B1 code>)`、`report.generation.failed` outbox、audit row，并按 runner retry policy 更新 `async_jobs`。
- `ai_task_runs`
  - `report.generate` 调用写 `task_type='report_generate'`。
  - `report.question_assessment` 调用写 `task_type='report_assessment'`。
  - `status` 使用 B4 `ai_task_runs.status` enum：`success` / `failed` / `timeout` / `fallback`。
  - `validation_status='invalid'` 仅用于 AI output schema invalid / parse-empty 路径。
- Privacy / observability
  - Report JSON、question assessment JSON、outbox payload、audit metadata、log 和 metric label 不得包含 raw `question_text`、`answer_text`、`hint_text`、AI prompt body、AI response body 或 provider secret。
  - Wire `provenance` 只含 `promptVersion`、`rubricVersion`、`modelId`、`language`、`featureFlag`、`dataSourceVersion`；runtime-only 字段只进入 `ai_task_runs` 或 audit 摘要。

### 2.2 Out of Scope

- 不实现 `completePracticeSession`；该写入口归 [`backend-practice`](../backend-practice/spec.md)。
- 不实现 JD 解析、target ownership flow 或 resume asset flow；分别归 [`backend-targetjob`](../backend-targetjob/spec.md) 与 [`backend-resume`](../backend-resume/spec.md)。
- 不暴露手工 regenerate / retry API。
- 不实现 report 导出、分享、时间线、高级排序、高级筛选或高级 retry-focus 算法。
- 不直接修改 OpenAPI、shared conventions、B3 event/job、B4 migration、A3 provider 或 F3 prompt/rubric truth source。

## 3 决策

| ID | 决策 | 当前值 | 影响 |
|----|------|--------|------|
| D-1 | API 来源 | 只使用 B2 generated `getFeedbackReport` / `listTargetJobReports` | 不私造 endpoint 或 schema |
| D-2 | DB 来源 | 使用 B4 `feedback_reports` / `question_assessments` / `async_jobs` / `ai_task_runs` / `outbox_events` / `audit_events` | 表、列、CHECK 与索引以 migration truth source 为准 |
| D-3 | Event / job 来源 | 使用 B3 `report.generation.requested` / `report.generated` / `report.generation.failed` 与 `report_generate` | `practice.session.completed` 是 source event；dispatcher 不二次创建 report job |
| D-4 | ReadinessTier | `score_level` weak/developing/proficient/strong 映射为 0.2/0.5/0.8/1.0，并按 rubric weight 汇总到四档 readiness | Wire `DimensionStatus` 仍为 `needs_work` / `meets_bar` / `strong` |
| D-5 | Retry focus | 选择 `overall_status='needs_work'` 或 `review_status='queued_for_retry'` 的 turn，按 turn_index 升序最多 5 个 | 写入 `retry_focus_turn_ids` 与 `included_in_retry_plan` |
| D-6 | Next action | not_ready/needs_practice 且有 retry focus -> `retry_current_round`；basically_ready/well_prepared 且 retry focus < 3 -> `next_round`；其他 -> `review_evidence` | 支撑前端报告 CTA |
| D-7 | Report status machine | `queued -> generating -> ready/failed`；retry path 可回到 `queued` | Handler read-only，业务写入由 job handler 完成 |
| D-8 | Failure semantics | AI / prompt / parse failure 均落 `failed + error_code`，HTTP read path 返回 200 failed shape | 前端通过 status 与 errorCode 渲染失败态 |
| D-9 | Provenance | 6 个 wire 字段全部从 `feedback_reports` 单表读取 | 不通过 `ai_task_runs` JOIN 回填 wire |
| D-10 | Runtime owner | `report_generate` 由 backend-async-runner kernel 调度，backend-review 提供 `review.GenerateHandler` | 避免并行 runner ownership |
| D-11 | User isolation | 所有 read/write 均按 user scoped；越权 read 返回 `REPORT_NOT_FOUND` | 不泄露 report/target 存在性 |
| D-12 | Privacy | 持久化、事件、日志、metric 与 audit 不保存 raw QA / prompt / response / secret | 通过 unit、scenario 和 lint gate 固化 |

## 4 Operation Matrix

| Operation / Path | Fixture | Frontend consumer | Backend owner | Persistence | AI dependency | Scenario coverage |
|------------------|---------|-------------------|---------------|-------------|---------------|-------------------|
| `getFeedbackReport` | `openapi/fixtures/Reports/getFeedbackReport.json` | `frontend-report-dashboard` generating/report screens | `backend/internal/api/reports`, `backend/internal/review`, `backend/internal/store/review` | `feedback_reports`, `question_assessments` | none on read path | `E2E.P0.053`, `E2E.P0.055` |
| `listTargetJobReports` | `openapi/fixtures/Reports/listTargetJobReports.json` | report records list within current report dashboard owner | same as above | `target_jobs` ownership check, `feedback_reports` cursor read | none on read path | `E2E.P0.053`, `E2E.P0.055` |
| `report_generate` job | N/A | async path observed through `getFeedbackReport` | `runner.Runtime` + `review.GenerateHandler` | `feedback_reports`, `question_assessments`, `async_jobs`, `outbox_events`, `audit_events`, `ai_task_runs` | `report.generate`, `report.question_assessment`, A3 `AIClient.Complete` | `E2E.P0.052`, `E2E.P0.054`, `E2E.P0.055` |

## 5 验收标准

| ID | 场景 | Given | When | Then |
|----|------|-------|------|------|
| C-1 | Report happy path | completed practice session has queued report and queued report job | runner handles `report_generate` | report becomes `ready`; assessments, outbox, audit, ai_task_runs, async job completion are persisted |
| C-2 | Assessment mapping | session has assessed turns | question assessment runs per turn | each turn has `QuestionAssessment`; internal score level maps to current wire `DimensionStatus` |
| C-3 | Readiness and next action | assessments are persisted | readiness/retry/next-action calculators run | report has current `preparednessLevel`, `retryFocusTurnIds`, and first next action type |
| C-4 | Placeholder read | report is `queued` or `generating` | user calls `getFeedbackReport` | 200 placeholder shape, no stale content, nullable provenance |
| C-5 | Failed read | report is `failed` | user calls `getFeedbackReport` | 200 failed shape with `errorCode`, empty content arrays |
| C-6 | Report list | user owns target with 0/N reports | user calls `listTargetJobReports` | cursor pagination, empty list shape, and ownership gate work |
| C-7 | AI failure | F3/A3/parse failure occurs | handler processes job | report is `failed`, retry/permanent state is correct, failure event emitted |
| C-8 | Provenance boundary | report is `ready` | user reads report | wire provenance has exactly 6 current fields and no runtime-only fields |
| C-9 | Cross-user isolation | user B requests user A report/target | read handler runs | user B receives `REPORT_NOT_FOUND` |
| C-10 | Privacy boundary | report generation succeeds or fails | persisted and emitted outputs are scanned | no raw QA text, prompt body, AI response body, or provider secret appears |
| C-11 | Async job concurrency | multiple workers are available | runner kernel leases jobs | a report job is processed once and terminalized once |

## 6 关联计划

1. [`001-report-generation-baseline`](./plans/001-report-generation-baseline/plan.md): current completed owner for report generation, read handlers, failure semantics, privacy, observability, BDD, and contract gates.
2. `002-advanced-retry-focus-and-listing`: reserved only if product scope requires advanced retry focus, filtering, sorting, or manual retry.
3. `003-report-retention-and-cascade`: reserved only if product scope requires retention or privacy export expansion.

## 7 关联文档

- [Product Scope](../product-scope/spec.md)
- [docs/ui-design/report-dashboard.md](../../ui-design/report-dashboard.md)
- [openapi-v1-contract](../openapi-v1-contract/spec.md)
- [event-and-outbox-contract](../event-and-outbox-contract/spec.md)
- [db-migrations-baseline](../db-migrations-baseline/spec.md)
- [shared-conventions-codified](../shared-conventions-codified/spec.md)
- [ai-provider-and-model-routing](../ai-provider-and-model-routing/spec.md)
- [prompt-rubric-registry](../prompt-rubric-registry/spec.md)
- [backend-async-runner](../backend-async-runner/spec.md)
- [backend-auth](../backend-auth/spec.md)
- [backend-practice](../backend-practice/spec.md)
- [backend-targetjob](../backend-targetjob/spec.md)
- [backend-resume](../backend-resume/spec.md)
- [frontend-report-dashboard](../frontend-report-dashboard/spec.md)
- [docs/development.md §2 Frontend / Backend Contract Workflow](../../development.md)
