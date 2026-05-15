# 001 — Report Generation Baseline

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-05-16

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

把 [backend-review spec](../../spec.md) v1.0 §7 锁定的第一个 plan 范围落地，承接 [backend-practice/002-event-loop-and-completion](../../../backend-practice/plans/002-event-loop-and-completion/plan.md) 已交付的 completePracticeSession 同事务 `feedback_reports(status='queued')` placeholder + `async_jobs(report_generate, status='queued', dedupe_key=sessionId)` queued row + `practice.session.completed` outbox source event 基础设施，闭合 P0 用户路径中"练习完成 → 报告生成 → 报告读取"段：

- 新增 `backend/internal/review/runner.go` inline polling worker，与既有 `backend/internal/targetjob/drainer.go`（`targetjob.Drainer`）共存但**不复用其抽象**（spec D-16 决策）；用 `SELECT ... FOR UPDATE SKIP LOCKED` lease `async_jobs(job_type='report_generate', status='queued', available_at <= now())` 行，并 `UPDATE` 为 `status='running', attempts=attempts+1, locked_at=now()`（B4 baseline 实际列名，**不写** `leased_at` / `attempt_count` / `worker_id`）；同事务推进 `feedback_reports.status='queued' → 'generating'`；事务外调用 F3 `report.generate` + `report.question_assessment` v0.1.0 baseline + A3 `AIClient.Complete`；解析 AI 输出 → 算 ReadinessTier (D-4) + retry_focus_turns (D-5) + next_action (D-6) → 单事务持久化 `feedback_reports.status='ready'`（含 Phase 0 baseline 新增的 `language` / `feature_flag` / `data_source_version` / `retry_focus_turn_ids` 4 列）+ N 行 `question_assessments` + outbox `report.generated`；AI 失败按 D-8 graceful：`feedback_reports.status='failed', error_code=<B1 AI_*>` + outbox `report.generation.failed` + retry policy (D-7 + spec §2.1)。`backend/internal/privacy/runner/` 在仓库内仅是 `targetjob.JobHandler` 实现（不是 polling worker），本 plan **不**基于 privacy/runner 套用 lifecycle。
- 新增 `backend/internal/api/reports.GetFeedbackReport` + `backend/internal/api/reports.ListTargetJobReports` handler + service + store；read-only；user-scoped；越权返回 `404 REPORT_NOT_FOUND`（D-15 派生 B1/B2 前置）；status placeholder 透传给客户端用于 generating 屏轮询；`FeedbackReport.provenance` wire 6 字段直接从 `feedback_reports` 单表读取（不 JOIN `ai_task_runs`）。
- Phase 0 跨 spec 契约前置（与 backend-practice D-30 integrator 模式一致）：B1 `shared/conventions.yaml` 新增 `REPORT_NOT_FOUND` 错误码并 regenerate Go/TS；B2 `openapi/openapi.yaml` 在 `getFeedbackReport` 404 响应中允许 `REPORT_NOT_FOUND`，并把 `FeedbackReport` schema 新增 `errorCode: oneOf[ApiErrorCode|null]` 字段；B4 `migrations/000001_create_baseline.up.sql` 进行 pre-launch baseline rebase 同 commit：(a) `ai_task_runs.task_type` CHECK 新增 `report_assessment`；(b) `feedback_reports` 新增 4 列 `language` / `feature_flag` / `data_source_version` / `retry_focus_turn_ids`；同步 owner spec history append（[B1 spec](../../../shared-conventions-codified/spec.md) + [B2 spec](../../../openapi-v1-contract/spec.md) + [B4 spec](../../../db-migrations-baseline/spec.md)）；F3 baseline `report.generate` / `report.question_assessment` v0.1.0 preflight assert（已 completed，本 plan 不动）；reaper（lease 超时回退）在 backend-review 包内自实现 goroutine，不修 B4。
- Phase 0 fixtures 扩展：`openapi/fixtures/Reports/getFeedbackReport.json` 新增 `report-failed` variant（status='failed' + errorCode 非空）；`openapi/fixtures/Reports/listTargetJobReports.json` 新增 `empty` variant（空列表）；通过 `make validate-fixtures` + contract test。

完成后用户在 generating 屏轮询 `getFeedbackReport` 可以拿到从 `queued` → `generating` → `ready`（或 `failed`）的真实状态推进；frontend-report-dashboard plan 001 可以基于此切换 generating → report 真实闭环。

## 2 背景

backend-practice spec v1.8 + plan [002-event-loop-and-completion](../../../backend-practice/plans/002-event-loop-and-completion/plan.md)（已 completed, 2026-05-14）已交付：completePracticeSession HTTP 202 + `ReportWithJob`、同事务创建 `feedback_reports(status='queued')` placeholder + `async_jobs(job_type='report_generate', status='queued', dedupe_key=sessionId, resource_type='feedback_report', resource_id=reportId)` queued row + outbox `practice.session.completed`；D-32 forward-binding 锁定 `triggerEventSemantic: source_event_only`（外部 dispatcher 不再二次创建 job）；D-35 lock completed session 二次 complete 走 replay。前端 frontend-workspace-and-practice/002（completed, 2026-05-14）已经 fixture-only 消费 `completePracticeSession` 返回的 `reportId`，并 handoff 到 generating route（当前由 PlaceholderScreen 占位）。

prompt-rubric-registry 001-baseline 已 completed（2026-05-09），`report.generate` 与 `report.question_assessment` v0.1.0 baseline prompt（multi + en + zh）与 rubric（4 维度 weighted + 4 score levels）已 active；`config/ai-profiles.yaml` 的 `report.generate.default` (deepseek-v4-pro / temp 0.2 / max 4096 / 30s timeout) 与 `report.assessment.default` (deepseek-v4-pro / temp 0.1 / max 2048 / 15s timeout) profile 已 active。当前 implementation 不能只检查 ResolveActive 存在性：Phase 0 F3 preflight 必须直接读取 prompt body / rubric truth source，确认 report prompt 不要求 raw `question` / `answer` / `transcript` 输入，不要求 `evidence_quotes` 逐字输出，并能产出 D-4 所需的 `score_level` 与 B2 `DimensionStatus` 映射字段；若不满足，本 plan 必须先停止并交给 F3 owner 修订 prompt-rubric-registry，不得进入 Phase 1。

B4 migration 当前 baseline（grep 实测）：`feedback_reports`（17 列：含 highlights/issues/next_actions jsonb + preparedness_level CHECK + prompt_version/rubric_version/model_id/provider provenance 列 + error_code + generated_at + UNIQUE(session_id)；**当前 baseline 不含** `language` / `feature_flag` / `data_source_version` / `retry_focus_turn_ids` 4 列，Phase 0.5 通过 pre-launch baseline rebase 新增）；`question_assessments`（dimension_results jsonb + review_status CHECK + UNIQUE(report_id, turn_id) + ON DELETE CASCADE 反向到 report 与 turn）；`ai_task_runs.task_type` CHECK **当前仅含** `report_generate` / `jd_parse` / `resume_parse` / `question_generate` / `followup_generate` / `resume_tailor` / `debrief_generate`（**不含** `report_assessment`，Phase 0.3 通过 pre-launch baseline rebase 新增）；`async_jobs` 列名 = `attempts` / `locked_at`（无 `leased_at` / `attempt_count` / `worker_id`，本 plan 严格使用 B4 实际列名）；`ai_task_runs.status` 是 B4 CHECK 列（枚举 `success` / `failed` / `timeout` / `fallback`，**不是** `succeeded`），与 `async_jobs.status`（枚举含 `succeeded`）不同。

001 plan 在 spec v1.0 框架内推进，不引入新的 spec D-* 主决策，但固化 4 项实施级子决策：

- **D-32-related（来自 backend-practice spec D-32 + D-28）**：本 spec consume `async_jobs(report_generate)` queued row 是唯一入口；不通过 outbox dispatcher 二次创建 job；P0 阶段 backend 进程内 inline runner 通过 `SELECT FOR UPDATE SKIP LOCKED` + `WHERE job_type='report_generate' AND status='queued' AND available_at <= now()` 谓词消费。
- **plan-level D-16（implementation 子决策）**：`ai_task_runs.task_type` 取值统一用 `report_generate`（主调）和 `report_assessment`（每题维度评估）两个 value；当前 B4 baseline 的 CHECK 不含 `report_assessment`，Phase 0.3 **无条件** pre-launch baseline rebase 扩 B4 CHECK（与 backend-practice D-37 hint_generate 同模式）。
- **plan-level D-17（implementation 子决策）**：B1 `REPORT_NOT_FOUND` 错误码新增到 `shared/conventions.yaml` errors 表（与 D-15 spec 派生一致），用于 cross-user 隔离 404 响应；同步 generated Go/TS error helper；B2 `openapi.yaml` 在 `getFeedbackReport` 的 404 response schema 注册该 error code；同 commit 把 `FeedbackReport` schema 新增 `errorCode: oneOf[ApiErrorCode|null]` 字段（当前 schema 未声明此字段；D-15 + D-8 失败语义要求 wire 暴露）。同步 [B1 spec](../../../shared-conventions-codified/spec.md) 与 [B2 spec](../../../openapi-v1-contract/spec.md) history append。
- **plan-level D-18（implementation 子决策）**：worker reaper 在 backend-review 包内自实现（不修 B4），周期性 SELECT `async_jobs WHERE status='running' AND locked_at < now() - lease_timeout AND job_type='report_generate'` 把行回退到 `status='queued', locked_at=null, available_at=now()`；reaper 周期 = lease_timeout / 2，默认 lease_timeout=5min。reaper 失败不影响 worker 主循环。reaper 只对 `report_generate` job_type 作用，**不**触动 `targetjob.Drainer` 处理的其它 job_type 的 stale running 行（dual-runner 边界）。
- **plan-level D-19（implementation 子决策）**：B4 `feedback_reports` 在 Phase 0.5 通过 pre-launch baseline rebase **无条件**新增 4 列：`language text NOT NULL DEFAULT 'en'`、`feature_flag text NOT NULL DEFAULT 'none'`、`data_source_version text NOT NULL DEFAULT 'not_applicable'`、`retry_focus_turn_ids jsonb NOT NULL DEFAULT '[]'::jsonb`，使 wire `GenerationProvenance` 6 字段可由 `feedback_reports` 单表读取并由 PersistReport 单事务写入，避免读路径 JOIN `ai_task_runs`。默认值与 A3 `AITaskRunRow` 字段约定一致。

## 3 质量门禁分类

- **Plan 类型**: feature-behavior + contract（B1 `REPORT_NOT_FOUND` 新增 + B2 `FeedbackReport.errorCode` 新增 + B4 `ai_task_runs.task_type` 扩 `report_assessment` value + B4 `feedback_reports` 新增 4 列 + 跨 spec history append）+ code-internal
- **TDD 策略**: Code plan requires TDD — 每个 implementation checklist 项 Red-Green-Refactor 入口在 `backend/internal/review/`、`backend/internal/api/reports/`、`backend/internal/store/review/`、`backend/internal/ai/aiclient/`（扩 typed task type）下相应包；migration check 入口为 `cd backend && go test ./internal/migrations/...`（baseline migration 契约测试位于 `backend/internal/migrations/sql_contract_test.go`；扩 `report_assessment` + `feedback_reports` 4 列时新增 `backend/internal/migrations/baseline_backend_review_rebase_test.go` 或扩展现文件）+ `python3 scripts/lint/conventions_drift.py --repo-root .`；测试命令从 Go module 根执行（例如 `cd backend && go test ./internal/review/... ./internal/api/reports/... ./internal/store/review/... ./internal/migrations/... ./cmd/api/...`）；详细 phase / file / verification 映射见 [test-plan](./test-plan.md)
- **BDD 策略**: Feature plan requires BDD — 引用 [bdd-plan](./bdd-plan.md) 与 [bdd-checklist](./bdd-checklist.md) 中的 4 个场景 `E2E.P0.052` / `E2E.P0.053` / `E2E.P0.054` / `E2E.P0.055`；主 [checklist](./checklist.md) 在每个 user-visible behavior phase 末尾列 `BDD-Gate:` 项
- **替代验证 gate**: Phase 0 跨 spec 契约修订使用 contract test + drift check（B1 `REPORT_NOT_FOUND` 出现在 conventions.yaml errors 表 + generated Go/TS 常量 + `python3 scripts/lint/conventions_drift.py` 通过）+ B4 migration apply test（CHECK 接受 `report_assessment` + `feedback_reports` 4 新增列存在）+ B2 schema test（`FeedbackReport.errorCode` 字段被 codegen 出）+ B1 / B2 / B4 spec history append 文件存在断言 + F3 prompt/rubric privacy preflight（read-only；若 prompt body 仍要求 raw `question` / `answer` / `transcript` 或 verbatim `evidence_quotes`，必须先回 F3 owner）+ `/sync-doc-index --check` Header / INDEX 漂移检测 + legacy-negative grep（旧 `reportLayout` / 5 档 readiness / 独立错题 / Drill / 报告时间线 / 多形态 + 旧列名 `leased_at` / `attempt_count` / `worker_id` 在 backend-review 包 + fixtures 中零出现）；Phase 6 隐私 / 观测使用 metric label allowlist + repo grep + redaction assertion 作为 gate

## 3.1 Operation Matrix

| `operationId` / 异步路径 | fixture | frontend consumer | backend handler / worker | persistence | AI dependency | scenario coverage |
|--------------------------|---------|-------------------|--------------------------|-------------|---------------|-------------------|
| `getFeedbackReport` | `openapi/fixtures/Reports/getFeedbackReport.json`：Phase 0 补齐 `report-failed`（status='failed' + error_code 非空 + provenance null + 空 highlights/issues/next_actions/questionAssessments）；保留已有 `default`（status='ready', preparednessLevel='basically_ready', 完整字段）、`report-generating`（status='generating', 空内容）、`prototype-baseline`（中文示例） | frontend-report-dashboard plan 001 GeneratingScreen 轮询 + ReportScreen 详情；frontend-workspace-and-practice plan 002 不消费（负向断言） | Phase 5：`backend/internal/api/reports.GetFeedbackReport` + `backend/internal/review.Service.GetFeedbackReport` + `backend/internal/store/review.GetFeedbackReport`（user-scoped read by reportId） | `feedback_reports` + `question_assessments` read（按 report_id join，按 turn_index 升序） | none in handler path | `E2E.P0.053` |
| `listTargetJobReports` | `openapi/fixtures/Reports/listTargetJobReports.json`：Phase 0 补齐 `empty`（空列表 + hasMore=false + nextCursor=null）；保留已有 `default`（分页 + pageInfo） | frontend-report-dashboard plan 001 不消费（spec D-7 dashboard-only，不在一级导航暴露列表）；但 schema parity 必须保证 future plan 接入 | Phase 5：`backend/internal/api/reports.ListTargetJobReports` + `backend/internal/review.Service.ListTargetJobReports` + `backend/internal/store/review.ListTargetJobReports`（cursor 分页 + user-scoped） | `feedback_reports` cursor read（WHERE target_job_id=$1 AND user_id=$2 ORDER BY created_at DESC, id DESC LIMIT $3） | none | `E2E.P0.053` |
| `(worker: report_generate job)` | N/A（异步 worker，不暴露 API） | N/A | Phase 1-4：`backend/internal/review/runner.go` + `backend/internal/review/generate_report_service.go` + `backend/internal/review/question_assessment.go` + `backend/internal/review/readiness.go` + `backend/internal/review/retry_focus.go` + `backend/internal/review/next_action.go` + `backend/internal/store/review/lease_async_job.go` + `backend/internal/store/review/persist_report.go` + `backend/internal/store/review/outbox_emitter.go` | Lease：`async_jobs UPDATE status='running', attempts=attempts+1, locked_at=now()`（B4 列名，attempts 在 lease 时递增一次）；Generating：`feedback_reports.status='generating'`；Ready 持久化：`feedback_reports.status='ready'` + `language` + `feature_flag` + `data_source_version` + `retry_focus_turn_ids` 4 新增列 + N rows `question_assessments` + outbox `report.generated` + `async_jobs.status='succeeded', completed_at=now(), locked_at=null`；Failed：`feedback_reports.status='failed', error_code, generated_at` + 使用当前 `async_jobs.attempts` 计算 `available_at` / retryable / permanent status，不在 failure finalize 再次递增 attempts + outbox `report.generation.failed` + `audit_events` 一行 | F3 `report.generate` (config/prompts/report.generate/v0.1.0.{yaml,md,en,zh}) + `report.question_assessment` (config/prompts/report.question_assessment/v0.1.0.{yaml,md,en,zh}) 各 1 个 baseline；Phase 0 preflight 必须确认 prompt/rubric IO 契约满足 privacy + D-4 score_level 映射；A3 observed AIClient.Complete × N+1 调用（1 次主调 + N 次逐题评估）；A3 observability decorator 自动写 `ai_task_runs(task_type='report_generate' 或 'report_assessment', status='success' 或 'failed' 等 B4 enum)` typed columns | `E2E.P0.052` + `E2E.P0.054` + `E2E.P0.055` |

## 3.5 Coverage Matrix

| 行 | 类别 | source | plan_phase | verification | negative_scope |
|----|------|--------|-----------|--------------|----------------|
| R1 | Primary | spec C-1（happy path: queued → generating → ready） | Phase 1 + Phase 4 | `runner_test.go::TestRunnerLeaseAndPersistReady` + `generate_report_service_test.go::TestGenerateReportSuccess` + repository 集成测试 + `E2E.P0.052` | — |
| R2 | Primary | spec C-2（per-question 维度评估写入 question_assessments） | Phase 2 + Phase 4 | `question_assessment_test.go::TestAssessQuestionsForAllTurns` + `persist_report_test.go::TestPersistReportWritesQuestionAssessments` + `E2E.P0.052` 子断言 | — |
| R3 | Primary | spec C-3 + D-4（ReadinessTier 加权阈值算法；internal `score_level` → B2 `DimensionStatus` 映射） | Phase 3 | `readiness_test.go::TestComputeReadinessTier` 表驱动覆盖 4 档边界 + `score_level` 到 wire status 映射 + property 测试 + `E2E.P0.052` 子断言 | 错误档位（5 档 readiness）/ 旧 readiness numeric score 暴露 / 把 weak/developing/proficient 误写进 B2 `DimensionStatus` |
| R4 | Primary | spec D-5（retry_focus_turns 选择策略） | Phase 3 | `retry_focus_test.go::TestSelectRetryFocusTurns` 表驱动覆盖 needs_work × review_status 组合 + 最多 5 个边界 + `E2E.P0.052` 子断言 | 高级加权算法不在 P0 |
| R5 | Primary | spec D-6（next_action enum 决策） | Phase 3 | `next_action_test.go::TestDecideNextAction` 表驱动覆盖 4 档 readiness × retry_focus count 矩阵 + `E2E.P0.052` 子断言 | 旧 next_action 取值 / 不在 enum 内的 type |
| R6 | Alternate | spec C-4（getFeedbackReport status=queued/generating 占位响应） | Phase 5 | `get_feedback_report_test.go::TestGetFeedbackReportReturnsPlaceholderForQueuedAndGenerating` + `E2E.P0.053` | placeholder 不暴露 stale 内容字段 / null provenance 但 schema 合法 |
| R7 | Alternate | spec C-5（listTargetJobReports 分页 + 空列表） | Phase 5 | `list_target_job_reports_test.go::TestListTargetJobReportsCursorPagination` 覆盖空 / 单页 / 多页 / cursor decode + `E2E.P0.053` | 越权 target_job 行不出现在结果 |
| R8 | Failure | spec C-6 + D-8（F3 resolve 失败 / A3 timeout / A3 invalid JSON / parsed empty） | Phase 6 | `generate_report_service_test.go::TestGenerateReportFailedMatrix` 表驱动 6 路径（F3 ErrPromptUnsupported / F3 ErrLanguageUnsupported / A3 secret missing / A3 timeout / A3 invalid output / parsed empty）+ `E2E.P0.054` | 不返回 5xx 给 HTTP；session 不进入 backend 自定义状态；outbox 必须发 `report.generation.failed` |
| R9 | Failure | spec C-7（retry policy + max attempts → permanent failed） | Phase 6 | `runner_test.go::TestRunnerRetryPolicyAndPermanentFail` 覆盖 `async_jobs.attempts` 1→2→3→4→5 退避计算 + 第 5 次失败置 permanent + `E2E.P0.054` | retry 5 次后再 reschedule；`attempts` 不正确累加；列名漂移到 `attempt_count` 必须被 R21 legacy grep 捕获 |
| R10 | Boundary | spec C-12（同 async_jobs 行双 worker 抢占） | Phase 1 | `lease_async_job_test.go::TestLeaseSkipLocked` multi-goroutine 真实 Postgres + 反向断言两个 worker 不会同时 lease 同一行 | 行锁泄漏 / 锁未释放 |
| R11 | Boundary | Phase 0 D-18 worker reaper | Phase 1 | `reaper_test.go::TestReaperReclaimsExpiredLease` + 反向断言 reaper 不动 succeeded / failed 行 | reaper 误回退活跃行 |
| R12 | Boundary | listTargetJobReports cursor encode/decode 边界 | Phase 5 | `list_target_job_reports_test.go::TestCursorEncodeDecode` 覆盖 base64 encoded `(created_at, id)` tuple 双向 + 非法 cursor → 400 | cursor 篡改 / 旧 cursor 格式 |
| R13 | Cross-layer contract | spec D-1 + D-9（B2 schema 一致性 + `FeedbackReport.provenance` 仅含 6 wire 字段；6 字段持久化锚点 = `feedback_reports` 单表） | Phase 0 + Phase 4 | `make codegen-check` + `python3 scripts/lint/conventions_drift.py --repo-root .` + provenance JSON marshal 单元测试（断言 6 keys 严格相等且全部由 `feedback_reports` 列直接 read 出，无 JOIN）+ `E2E.P0.055` 子断言 | runtime 字段（`feature_key` / `model_profile_name` / provider / cost / latency）不得出现在 wire JSON；不通过 `ai_task_runs` JOIN 回填 wire 字段 |
| R14 | Cross-layer contract | spec D-3 + D-12（B3 event payload schema 一致） | Phase 4 + Phase 6 | `outbox_emitter_test.go::TestReportGeneratedPayload` + `TestReportGenerationFailedPayload` 断言与 `shared/events/report.generated.json` / `report.generation.failed.json` 一致 + piiBoundary lint 通过 + `E2E.P0.052` / `E2E.P0.054` | 旧 event name / 旧 payload 字段 |
| R15 | Cross-layer contract | Phase 0 D-17（B1 `REPORT_NOT_FOUND` + B2 `getFeedbackReport` 404 schema + B2 `FeedbackReport.errorCode` 字段） | Phase 0 + Phase 5 | `shared/conventions.yaml` errors 行存在断言 + generated Go `ErrReportNotFound` 常量单元测试 + `make codegen-check` + `openapi.yaml` `getFeedbackReport` 404 response 含 `REPORT_NOT_FOUND` 引用 + `FeedbackReport` schema 含 `errorCode` 字段 + B1/B2 spec history append 文件存在断言 + `make docs-check` 通过 | 旧 404 错误码（`PRACTICE_SESSION_NOT_FOUND` 误用于 report）；B2 `FeedbackReport.errorCode` 未声明而 fixture 直接出现 |
| R16 | Cross-layer contract | Phase 0 D-16 + D-19（B4 `ai_task_runs.task_type` CHECK 扩 `report_assessment` + `feedback_reports` 新增 4 列 `language` / `feature_flag` / `data_source_version` / `retry_focus_turn_ids`） | Phase 0 | migration apply test：`cd backend && go test ./internal/migrations/... -count=1` 断言 CHECK 接受 `report_assessment` 与 `report_generate`，拒绝未知 task_type；同测试断言 `feedback_reports` 列含 4 个新列且默认值生效；`migrations/lint.sh` 通过；本 plan 是 pre-launch baseline rebase（与 backend-practice D-37 `hint_generate` 同模式），无条件分支 | 不引入向后兼容 ALTER；不增加 deprecated alias；本 plan 不写 `up.sql` 之外的迁移文件 |
| R17 | Cross-layer contract | F3 baseline (`report.generate` + `report.question_assessment` v0.1.0 active) + prompt/rubric privacy IO contract | Phase 0 | F3 preflight integration test：读取 `docs/spec/prompt-rubric-registry/spec.md` v2.1 + `plans/001-baseline/checklist.md` `状态: completed` + `config/prompts/report.generate/*` + `config/prompts/report.question_assessment/*` + matching rubrics；断言 `RegistryClient.ResolveActive(ctx, "report.generate", "en")` / `("report.question_assessment", "en")` 返回非空 ResolvedPrompt，prompt body 不含 raw `{{question}}` / `{{answer}}` / `{{transcript}}` input contract，不要求 verbatim `evidence_quotes` 输出，rubric `score_levels` 覆盖 weak/developing/proficient/strong，且 implementation 有 `score_level` → B2 `DimensionStatus` 映射测试 | F3 baseline drift 必须由 F3 owner 先修订；本 plan 不私自 stub registry，也不得绕过隐私红线直接消费 raw transcript prompt |
| R18 | Privacy/security | spec D-11 + C-10（feedback_reports / question_assessments jsonb 不写 raw answer/question/prompt/response）+ cross-user FK 隔离 | Phase 4 + Phase 5 + Phase 6 | `persist_report_test.go::TestPersistReportRedactsRawText` + `assess_questions_test.go::TestQuestionAssessmentParsesWithoutLeaks` + `get_feedback_report_test.go::TestCrossUserNotFound` + outbox emitter PII 单元测试 + `E2E.P0.055` | `question_text` / `answer_text` / `hint_text` / AI prompt body / AI response body / provider secret 在 `feedback_reports.highlights/issues/next_actions` / `question_assessments.strengths/gaps/recommended_framework/dimension_results` / outbox payload / metric label / log / audit 中零出现 |
| R19 | Observability | spec D-10（ai_task_runs typed columns）+ A3/F1 已有边界 | Phase 2 + Phase 4 + Phase 6 | A3 observed AIClient wiring test（runner + service）+ `ai_task_runs` 集成测试 + metric label allowlist 单元测试 + `E2E.P0.054` 反查 | metric label 不含 `feature_key` / prompt-rubric version / provider raw model id；ai_task_runs 行包含 `task_type='report_generate' or 'report_assessment'` + `status` ∈ {`success`,`failed`,`timeout`,`fallback`}（**B4 CHECK 列；不是 'succeeded'，不是 `validation_status`**）+ `validation_status` ∈ {`ok`,`invalid`}（独立观察列，仅在 AI 输出 schema 失败时置 `invalid`）+ `latency_ms` + `input_tokens` + `output_tokens` + `model_profile_name` |
| R20 | UX (API) | spec D-15（REPORT_NOT_FOUND 404）+ status placeholder 透传 | Phase 5 | handler 单元测试断言：cross-user → 404 + `ApiError{code:'REPORT_NOT_FOUND'}`；status='queued'/'generating' → 200 + placeholder；status='ready' → 200 + 完整内容；status='failed' → 200 + error_code 非空 + 空内容；listTargetJobReports cursor 非法 → 400 + `VALIDATION_FAILED` | 不返回 502/503；不暴露 status='completing' 的 backend 内部态 |
| R21 | Regression / legacy-negative | spec §4.5（retired 术语）+ D-16 列名口径 | Phase 6 | scoped legacy grep（新增 `scripts/lint/backend_review_legacy.py` 或在 `scripts/lint/conventions_drift.py` 扩展 `--phase backend_review` 段）：实现 / runtime 输出范围（backend review/api/store、openapi Reports fixtures、scenario assets、generated/runtime tests）对 `reportLayout` / 5 档 readiness（`acceptable` / `needs_work` 作为 readiness 取值 — 注意是 dimension_status 取值，不是 readiness_tier）/ `readiness_score` numeric / `mistakes_queue` / `drill_builder` / `growth_center` / 报告时间线 (`report_timeline`) / 多形态 report (`report_form`) / 旧 next_action 取值 / 旧 review_method_version / 错误列名 `leased_at` / `attempt_count` / `worker_id`（这些不是 B4 列；plan/checklist 不应再写）/ `ai_task_runs.status='succeeded'` 字面量 零出现；negative tests / prohibition docs 允许枚举字面量作为 grep 输入；本 plan / BDD / test docs、backend-review spec §4.5 prohibition rows 不属于实现 / runtime 范围 | retired 术语不得在实现/runtime 输出中出现；错误列名漂移不得回流 |
| R22 | Out-of-scope boundary | 002 / 003 owner 范围不应被 001 实现 | Phase 6 | unit 断言 001 不实现高级 retry-focus 加权算法 / 手工 retry API / DELETE CASCADE 集成 / retention policy / report 导出；repo grep `regenerateReport` / `retryReport` / `retainReportDays` 在 backend-review 包零出现 | 001 不调用 `practice.session.first_question` / `practice.session.follow_up`（这两个 feature_key 由 backend-practice owner 触发） |

无 UI 视觉地理 parity 行（本 plan 不涉及 `ui-design/` 复刻；frontend-report-dashboard plan 001 承担 UI parity）。

## 3.6 L2 修订说明

待 L1 / L2 review pass 后追加。

## 4 实施步骤

### Phase 0: 跨 spec 前置修订 + Preflight

**目标**：把 D-15（B1 `REPORT_NOT_FOUND` + B2 schema）+ D-16（B4 `ai_task_runs.task_type='report_assessment'`）+ D-17（B2 `FeedbackReport.errorCode` 字段）+ D-19（B4 `feedback_reports` 新增 4 列）落到编码真理源；F3 baseline 仅 preflight assert。001 直接修订各 owner spec 编码真理源（按 backend-practice D-30 Q1=A integrator 模式延续），同步更新 owner spec history.md / spec.md Header。所有 B4 改动通过 pre-launch baseline rebase 同 commit，**无条件**执行（grep 已确认现状）。

#### 0.1 B1 `shared/conventions.yaml` 新增 `REPORT_NOT_FOUND` 错误码

在 `shared/conventions.yaml#errors` 表追加：

```yaml
- code: REPORT_NOT_FOUND
  httpStatus: 404
  retryable: false
  message: "feedback report not found or not accessible"
```

同步 [B1 `shared-conventions-codified` spec](../../../shared-conventions-codified/spec.md) Header bump（next minor，如 `1.17 → 1.18`）+ `history.md` 追加 row："授权 backend-review/001 Phase 0 新增 `REPORT_NOT_FOUND` 错误码（cross-user 隔离 404）；不引入 deprecated alias"。

验证：`python3 scripts/lint/conventions_drift.py --repo-root .` 通过 + `make codegen-check` 通过 + generated Go `ErrReportNotFound` 常量与 generated TS 等价常量出现。

#### 0.2 B2 OpenAPI `getFeedbackReport` 404 response 引用 `REPORT_NOT_FOUND`

修订 `openapi/openapi.yaml#paths./reports/{reportId}.get.responses.'404'`，确保 `content.application/json.schema.$ref` 指向 `ApiError`，并在该 response 的 `description` / `example` 中显式列出 `code: REPORT_NOT_FOUND`（与 `getPracticeSession` 404 `PRACTICE_SESSION_NOT_FOUND` 同模式）。`openapi/baseline/openapi-v1.0.0.yaml` 按 D-21 pre-launch baseline rebase 原地修订。

Regenerate：`make codegen-openapi` + `make codegen-check`，刷新 Go server / TS client；`python3 scripts/lint/conventions_drift.py --repo-root .` 通过。

同步 [B2 `openapi-v1-contract` spec](../../../openapi-v1-contract/spec.md) Header bump + `history.md` 追加 row："授权 backend-review/001 Phase 0 把 `getFeedbackReport` 404 response 显式与 `REPORT_NOT_FOUND` 关联（pre-launch baseline rebase）"。

#### 0.3 B4 `ai_task_runs.task_type` CHECK 扩 `report_assessment`（无条件 pre-launch baseline rebase）

grep 已确认当前 `migrations/000001_create_baseline.up.sql:386` 的 `ai_task_runs.task_type` CHECK 仅含 `jd_parse / resume_parse / question_generate / followup_generate / report_generate / resume_tailor / debrief_generate`，**不含** `report_assessment`。

- 修订 `migrations/000001_create_baseline.up.sql` 把 `report_assessment` 添加到 `ai_task_runs.task_type` CHECK 列表（pre-launch baseline rebase，与 backend-practice D-37 `hint_generate` 同模式）。
- 同步 `migrations/enum-sources.yaml` 与 `migrations/lint.sh` 校验。
- 同步 [B4 `db-migrations-baseline` spec](../../../db-migrations-baseline/spec.md) Header bump + `history.md` 追加 row："授权 backend-review/001 Phase 0 `ai_task_runs.task_type` CHECK 扩值 `report_assessment`（pre-launch baseline rebase）"。
- 同步 `backend/internal/ai/aiclient/writers.go` 添加 `AITaskRunTaskReportAssessment AITaskRunCapability = "report_assessment"` 常量到 `allowedAITaskRunCapabilities` 集合（与 `AITaskRunTaskReportGenerate` 已有常量并列）。

验证：`cd backend && go test ./internal/migrations -count=1` 通过 + `migrations/lint.sh` 通过 + `cd backend && go test ./internal/ai/aiclient -count=1` 通过 + drift lint 通过。

#### 0.4 B2 `FeedbackReport.errorCode` 字段新增 + fixtures 扩展（无条件）

grep 已确认 `openapi/openapi.yaml` 中 `FeedbackReport` schema（行 3161-3215）**未声明** `errorCode` 字段。本 Phase 0 必须先扩 B2 schema（与 0.2 同 commit），再补 fixture：

1. 修订 `openapi/openapi.yaml#components.schemas.FeedbackReport.properties` 添加：

```yaml
errorCode:
  oneOf:
    - $ref: '#/components/schemas/ApiErrorCode'
    - type: 'null'
  description: |
    Populated only when `status == 'failed'`. Carries the B1 error code
    (typically `AI_*` family) so the client can render a failure state.
```

   - 同步 `openapi/baseline/openapi-v1.0.0.yaml`（pre-launch baseline rebase）。
   - regenerate `make codegen-openapi` + `make codegen-check`，刷新 Go server / TS client。
   - 同步 [B2 `openapi-v1-contract` spec](../../../openapi-v1-contract/spec.md) Header bump + `history.md` 追加 row："授权 backend-review/001 Phase 0 把 `FeedbackReport.errorCode` 字段新增到 wire schema（与 D-15 同 commit）"。

2. 按 §3.1 在 `openapi/fixtures/Reports/getFeedbackReport.json` 补齐 `report-failed` variant（schema 已暴露 `errorCode` 后再写）：

```json
{
  "report-failed": {
    "summary": "Report generation failed",
    "value": {
      "id": "<uuid>",
      "sessionId": "<uuid>",
      "targetJobId": "<uuid>",
      "status": "failed",
      "preparednessLevel": null,
      "highlights": [],
      "issues": [],
      "nextActions": [],
      "questionAssessments": [],
      "retryFocusTurnIds": [],
      "provenance": null,
      "errorCode": "AI_PROVIDER_TIMEOUT",
      "createdAt": "<iso>",
      "updatedAt": "<iso>"
    }
  }
}
```

3. 在 `openapi/fixtures/Reports/listTargetJobReports.json` 补齐 `empty` variant：

```json
{
  "empty": {
    "summary": "Empty list",
    "value": {
      "items": [],
      "pageInfo": { "nextCursor": null, "pageSize": 20, "hasMore": false }
    }
  }
}
```

验证：`make codegen-check` 通过 + `make validate-fixtures` 通过 + contract test 通过。

#### 0.5 B4 `feedback_reports` 新增 4 列（无条件 pre-launch baseline rebase；D-19）

grep 已确认当前 `migrations/000001_create_baseline.up.sql:256-275` 的 `feedback_reports` 列只有 17 个，**不含** `language` / `feature_flag` / `data_source_version` / `retry_focus_turn_ids`。本 Phase 0 同 commit 扩 B4 baseline：

修订 `migrations/000001_create_baseline.up.sql` 在 `feedback_reports` 表定义中追加 4 列：

```sql
language text NOT NULL DEFAULT 'en',
feature_flag text NOT NULL DEFAULT 'none',
data_source_version text NOT NULL DEFAULT 'not_applicable',
retry_focus_turn_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
```

默认值与 A3 `AITaskRunRow` 字段约定一致；放在 `provider` 后、`error_code` 前的位置以保持语义聚合。同步 `openapi/baseline/openapi-v1.0.0.yaml` 与 generated 客户端不受影响（wire `FeedbackReport.retryFocusTurnIds` 字段已存在；wire provenance 6 字段已存在）。

同步 [B4 `db-migrations-baseline` spec](../../../db-migrations-baseline/spec.md) Header bump + `history.md` 追加 row："授权 backend-review/001 Phase 0 `feedback_reports` 新增 `language` / `feature_flag` / `data_source_version` / `retry_focus_turn_ids` 4 列（pre-launch baseline rebase）"（与 0.3 同 row 或独立 row 由 B4 owner 决定）。

验证：`cd backend && go test ./internal/migrations -count=1` 通过（新增 `TestFeedbackReportsContainsProvenancePersistenceColumns`）+ `migrations/lint.sh` 通过。

#### 0.6 F3 baseline preflight

读取 `docs/spec/prompt-rubric-registry/spec.md` v2.1 + `docs/spec/prompt-rubric-registry/plans/001-baseline/checklist.md` `状态: completed` + work-journal `docs(prompt-rubric-registry): close 001-baseline lifecycle and record ac self-check` 行；断言 F3 已落地 `report.generate` 与 `report.question_assessment` v0.1.0 baseline + Resolve API 可用。

新增 `backend/internal/ai/registry/backend_review_preflight_test.go::TestF3ReportGenerateAndAssessmentPreflight`，必须同时断言：

- `RegistryClient.ResolveActive(ctx, "report.generate", "en")` 与 `("report.question_assessment", "en")` 返回非空 ResolvedPrompt。
- `config/prompts/report.generate/v0.1.0*.md` 与 `config/prompts/report.question_assessment/v0.1.0*.md` 不要求 raw `{{question}}` / `{{answer}}` / `{{transcript}}` 输入，不要求 `evidence_quotes` / verbatim quote 输出；报告生成只允许消费 backend-practice 已落地的 `questionIntent` / `answerSummary` / turn status / follow_up_count 等摘要字段。
- matching rubric `score_levels` 覆盖 weak / developing / proficient / strong，且 readiness 算法测试覆盖 `score_level` → B2 `DimensionStatus` 映射（weak/developing → `needs_work`，proficient → `meets_bar`，strong → `strong`）。

若上述 prompt/rubric privacy IO preflight 失败，001 implementation 必须停止在 Phase 0，并回到 prompt-rubric-registry owner 原地修订 F3 baseline；不得在 backend-review 中私自绕过或直接消费 raw transcript prompt。

#### 0.7 Phase 0 收口 gate

- `python3 scripts/lint/conventions_drift.py --repo-root .` 通过
- `make codegen-check` 通过
- `cd backend && go build ./...` 通过
- `make validate-fixtures` 通过
- B4 baseline rebase 验证：`cd backend && go test ./internal/migrations -count=1` + `migrations/lint.sh` 通过
- B1 / B2 / B4 owner spec history append 与 Header bump 在 Phase 1 实施前完成
- `make docs-check`（即 `/sync-doc-index --check`）通过，确认 backend-review / B1 / B2 / B4 spec Header 与 `docs/spec/INDEX.md` 投影一致
- F3 prompt/rubric privacy IO preflight 断言通过；若失败则停止并回到 F3 owner；本 plan 状态保持 `active`

### Phase 1: Inline review runner + lease + status state machine

**目标**：先落 lease + status 推进 + reaper 基础设施，独立可单元测试；不接 AI / 持久化具体内容。

#### 1.1 包结构骨架

新增 `backend/internal/review/`（spec D-16 决策：与 `targetjob.Drainer` 共存但不复用其抽象）：

- `runner.go`：`Runner.Start(ctx)` / `Runner.Stop(ctx)`；poll loop（默认间隔 5s，由 A4 runtime config 注入）；max concurrency=1 默认；worker 身份通过结构化 logger 字段 `worker_label` + 进程 ID + 启动时间组合，不写 DB。
- `service.go`：`Service.GenerateReport(ctx, leasedJob)` 接受已 lease 的 async_job，返回 `GenerateReportResult{Status, ErrorCode}`；纯函数式 outcome，不直接写 DB。
- `lease.go`：`LeaseNextJob(ctx) (LeasedJob, bool, error)` 调用 store 层 `SELECT FOR UPDATE SKIP LOCKED` + `UPDATE ... SET status='running', attempts=attempts+1, locked_at=now()`（B4 列名）；ReleaseJobAfterSuccess / ReleaseJobAfterFailure 收尾，分别 `UPDATE async_jobs SET status='succeeded', completed_at=now(), locked_at=null` 或 `UPDATE ... SET available_at=computeBackoff(attempts), status=CASE WHEN attempts >= 5 THEN 'failed' ELSE 'queued' END, locked_at=null`（failure finalize 使用 lease 后 attempts，不二次递增）。
- `reaper.go`：`Reaper.Start(ctx)` 周期性 SELECT `async_jobs WHERE job_type='report_generate' AND status='running' AND locked_at < now() - lease_timeout` 把行回退到 `status='queued', locked_at=null, available_at=now()`；周期 = lease_timeout / 2 默认；**只**作用于 `report_generate` job_type，不触动 `targetjob.Drainer` 处理的其它 job_type。

新增 `backend/internal/store/review/`：

- `lease_async_job.go`：实现 `LeaseAsyncJob(ctx, jobType, leaseTimeout) (LeasedJob, bool, error)` 返回行；`UpdateAsyncJobRunning(ctx, jobID)`；`UpdateAsyncJobSucceeded(ctx, jobID)`；`UpdateAsyncJobFailed(ctx, jobID, errorCode, nextAvailableAt, isPermanent)`。SQL 使用 `attempts` / `locked_at` 列；不使用 `leased_at` / `attempt_count` / `worker_id`。
- `feedback_reports_status.go`：`UpdateFeedbackReportStatus(ctx, reportID, oldStatus, newStatus, errorCode)` 用乐观锁保证状态机 D-7 单调迁移。
- `reaper.go`：`ReclaimExpiredLeases(ctx, jobType, leaseTimeout) (reclaimedCount, error)`。

新增 `backend/internal/api/reports/`：

- 空 skeleton；Phase 5 实施 handler 注册。

#### 1.2 状态机推进单元测试（Phase 1 已可独立测试）

- `runner_test.go::TestRunnerLeasesAndAdvancesToGenerating` 用 fake store + fake service 断言 lease 成功后 status `queued → generating`；断言 `locked_at` 已设置但**不**断言 worker_id 列。
- `lease_async_job_test.go::TestLeaseSkipLocked` 用 multi-goroutine 真实 Postgres（test stack）断言 SKIP LOCKED 行为；两个 worker 不会同时 lease 同一行；同一 row 单 worker 持有 lease（通过 `locked_at` not null + transaction lock 验证）。
- `feedback_reports_status_test.go::TestStatusStateMachineEnforcement` 断言 `queued → ready` 直接迁移被拒绝；`failed → ready` 被拒绝。
- `reaper_test.go::TestReaperReclaimsExpiredLease` 用 fake clock 验证 reaper 周期回收 stale running 行（`locked_at` 列 < now - lease_timeout）；不动 succeeded / failed 行；不动未超时 running 行；不动其它 job_type（如 `target_import`）行。

#### 1.3 Phase 1 收口 gate

- `cd backend && go test ./internal/review/... ./internal/store/review/... -count=1` 全部通过
- 单元测试覆盖 lease + reaper + status 推进 + double-worker SKIP LOCKED
- 不引入 AI 调用（Phase 1 service 用 fake 占位）

### Phase 2: AI 调用与内容生成

**目标**：通过 F3 + A3 真实生成 report 主调内容 + 逐题维度评估；不持久化 — 输出 outcome struct。

#### 2.1 主调 `report.generate` 调用

新增 `backend/internal/review/generate_report.go`：

- `generateReportContent(ctx, session, plan, turns) (ReportContentDraft, error)`：
  - `f3.ResolveActive(ctx, "report.generate", language)` 拿 `ResolvedPrompt{promptVersion, rubricVersion, modelProfileName, language, dataSourceVersion, ...}`。
  - 构造 system + user message：把 session metadata + plan goal + 每个 turn 的 `questionIntent + answerSummary + status + follow_up_count` 注入到 user message（**不**包含 raw `question_text` / `answer_text` / `hint_text`，由 F3 prompt template 决定结构）。
  - 调用 `a3.Complete(ctx, AIRequest{ModelProfileName, Messages, ResponseFormat:"json"})`；A3 observability decorator 自动写 `ai_task_runs(task_type='report_generate', ...)` 行。
  - 解析 AI JSON 响应到 `ReportContentDraft{Highlights, Issues, NextActions, ContentSummary}`；解析失败 → 返回 `ParseFailureError`；空内容 → 返回 `ParsedEmptyError`。

#### 2.2 逐题维度评估 `report.question_assessment` 调用

新增 `backend/internal/review/question_assessment.go`：

- `assessQuestionsForAllTurns(ctx, session, plan, turns) ([]QuestionAssessmentDraft, error)`：
  - 对每个 turn（按 `turn_index` 升序）：
    - `f3.ResolveActive(ctx, "report.question_assessment", language)` 拿 `ResolvedPrompt`。
    - 构造 user message：注入 turn 的 `questionIntent + answerSummary + follow_up_count + status`（**不**含 raw text）。
    - `a3.Complete(...)`；A3 decorator 写 `ai_task_runs(task_type='report_assessment')` typed columns。
    - 解析 JSON → `QuestionAssessmentDraft{TurnID, OverallStatus, Confidence, Strengths, Gaps, RecommendedFramework, DimensionResults map[string]DimensionResultDraft, ReviewStatus}`。
  - 失败任一 turn → 返回 partial result + first error；caller 决定整轮失败处理。

#### 2.3 AI 调用单元测试

- `generate_report_test.go::TestGenerateReportContentSuccess` 用 fake F3 + fake AIClient 返回合法 JSON 断言 outcome 结构。
- `generate_report_test.go::TestGenerateReportContentBuildsPromptWithoutLeaks` 反向断言 user message 不含 `question_text` / `answer_text` / `hint_text` literal。
- `question_assessment_test.go::TestAssessQuestionsForAllTurns` 用 fake F3 + fake AIClient 断言 N 行 outcome，按 turn_index 升序。
- `question_assessment_test.go::TestAssessQuestionsBuildsPromptWithoutLeaks` 反向断言。

#### 2.4 Phase 2 收口 gate

- `cd backend && go test ./internal/review/... -count=1` 全部通过
- 单元测试覆盖 F3 / A3 success 路径 + prompt redaction
- AI 失败路径占位（Phase 6 完整实现 6 路径 graceful degrade matrix）

### Phase 3: ReadinessTier / retry_focus / next_action 算法

**目标**：把 Phase 2 输出的 ReportContentDraft + QuestionAssessmentDraft 转换为 ReadinessTier + retry_focus_turn_ids + next_action enum。

#### 3.1 ReadinessTier 计算

新增 `backend/internal/review/readiness.go`：

- `computeReadinessTier(assessments []QuestionAssessmentDraft, rubric ResolvedPrompt) ReadinessTier`：
  - 对每个 assessment，按 dimension 取 `DimensionResultDraft.ScoreLevel`（weak=0.2 / developing=0.5 / proficient=0.8 / strong=1.0）× rubric.weight，按维度求和得 turn 分。B2 `DimensionResult.status` 仍使用 B1 enum `strong / meets_bar / needs_work`；implementation 必须通过 helper 把 score_level 映射为 wire status（weak/developing → `needs_work`，proficient → `meets_bar`，strong → `strong`），不得把 weak/developing/proficient 字面量写入 B2 `DimensionStatus`。
  - 对所有 turn 取算术平均得 session 分。
  - 阈值映射：< 0.30 → `not_ready`；[0.30, 0.55) → `needs_practice`；[0.55, 0.75) → `basically_ready`；≥ 0.75 → `well_prepared`。
  - 空 assessments 数组 → fallback `not_ready`（不应在 happy path 出现）。

- `readiness_test.go::TestComputeReadinessTier` 表驱动覆盖 4 档阈值边界（0.29 / 0.30 / 0.54 / 0.55 / 0.74 / 0.75 / 0.99）+ property 测试（随机 N 维度 × N turn）。

#### 3.2 retry_focus_turns 选择

新增 `backend/internal/review/retry_focus.go`：

- `selectRetryFocusTurnIDs(assessments []QuestionAssessmentDraft) []uuid.UUID`：
  - 筛选 `OverallStatus == "needs_work"` OR `ReviewStatus == "queued_for_retry"` 的 turn id。
  - 按 turn_index 升序排序。
  - 最多取前 5 个。
  - 同时 mutate `assessments[i].IncludedInRetryPlan = true` 对应行。

- `retry_focus_test.go::TestSelectRetryFocusTurns` 表驱动覆盖：全 needs_work / 全 strong / 混合（含 queued_for_retry）/ 超过 5 个 / 空数组。

#### 3.3 next_action 决策

新增 `backend/internal/review/next_action.go`：

- `decideNextAction(readiness ReadinessTier, retryFocusCount int) NextActionType`：
  - readiness ∈ {`not_ready`, `needs_practice`} 且 retryFocusCount ≥ 1 → `retry_current_round`
  - readiness ∈ {`basically_ready`, `well_prepared`} 且 retryFocusCount < 3 → `next_round`
  - 其他 → `review_evidence`（fallback）

- `next_action_test.go::TestDecideNextAction` 表驱动覆盖 4 档 readiness × retry_focus_count 矩阵（0 / 1 / 2 / 3 / 5）。

#### 3.4 BDD-Gate Phase 3

无独立 BDD（Phase 3 算法通过 Phase 4 持久化路径间接被 `E2E.P0.052` 覆盖；本 Phase 仅单元测试）。

#### 3.5 Phase 3 收口 gate

- `cd backend && go test ./internal/review/... -count=1` 全部通过
- 单元测试覆盖 4 档 readiness × N retry_focus × 决策矩阵 + property 测试 + 边界值

### Phase 4: 持久化 + outbox emit

**目标**：把 Phase 2-3 输出在单事务内持久化到 DB + 发出 outbox + 写 ai_task_runs（A3 decorator 自动 + worker 显式补 F3 / parse 失败行）。

#### 4.1 Repository PersistReport

新增 `backend/internal/store/review/persist_report.go`：

- `PersistReport(ctx, PersistReportInput) (PersistReportResult, error)`，单事务：
  1. `SELECT FOR UPDATE feedback_reports WHERE id=$1 AND status='generating'`；非法迁移 → ErrStatusMismatch。
  2. UPDATE `feedback_reports` 设置 `status='ready', preparedness_level, highlights, issues, next_actions, prompt_version, rubric_version, model_id, provider, language, feature_flag, data_source_version, retry_focus_turn_ids, generated_at=now()`。`language` 取自 `practice_sessions.language`（spec §2.1 / D-9 已锁定）；`feature_flag` 取自 F3 `ResolvedPrompt.FeatureFlag`；`data_source_version` 取自 F3 `ResolvedPrompt.DataSourceVersion`；`retry_focus_turn_ids` 由 Phase 3 `selectRetryFocusTurnIDs` 输出。
  3. INSERT N 行 `question_assessments`（每行 report_id + session_id + turn_id + question_intent + overall_status + confidence + strengths + gaps + recommended_framework + dimension_results + review_status + included_in_retry_plan + related_experience_card_ids）。
  4. INSERT `outbox_events(event_name='report.generated', event_version=1, aggregate_type='feedback_report', aggregate_id=reportId, payload={reportId,sessionId,targetJobId,preparednessLevel,questionIssueCount,promptVersion,rubricVersion,modelId})`。
  5. INSERT `audit_events(event_type='report_generated', actor_id=user_id, resource_type='feedback_report', resource_id=reportId, metadata={status,preparedness_level,language,target_job_id})`。
  6. UPDATE `async_jobs SET status='succeeded', completed_at=now(), locked_at=null WHERE id=$1`（B4 enum 用 `succeeded`，与 `targetjob.SQLStore.FinalizeAsyncJob` 一致；`locked_at=null` 释放租约信号）。
  7. 返回 `PersistReportResult{ReportID, AsyncJobID}`。

依赖 `outbox_emitter.BuildReportGeneratedPayload(reportID, sessionID, targetJobID, preparednessLevel, questionIssueCount, provenance)` 与隐私 lint helper `assertNoReviewOutboxPII`。

#### 4.2 Repository PersistReportFailure

新增 `backend/internal/store/review/persist_failure.go`：

- `PersistReportFailure(ctx, PersistReportFailureInput) (PersistReportFailureResult, error)`，单事务：
  1. UPDATE `feedback_reports SET status='failed', error_code=$1, generated_at=now() WHERE id=$2 AND status IN ('generating', 'queued')`。
  2. UPDATE `async_jobs SET available_at = computeBackoff(attempts), status = CASE WHEN attempts >= 5 THEN 'failed' ELSE 'queued' END, locked_at = null WHERE id=$3`（B4 列名 `attempts`，不是 `attempt_count`；不写 `worker_id`；attempts 已在 lease 时递增，本 failure finalize 不再二次递增）；computeBackoff 实现 `min(2^attempts * 30s, 30min)`。
  3. INSERT `outbox_events(event_name='report.generation.failed', event_version=1, aggregate_type='feedback_report', aggregate_id=reportId, payload={reportId,sessionId,errorCode,retryable=(attempts<5)})`。
  4. INSERT `audit_events(event_type='report_generation_failed', metadata={error_code,session_id,target_job_id,attempts,is_permanent})`。
  5. 返回 `PersistReportFailureResult{IsPermanent, NextAvailableAt}`。

#### 4.3 Worker glue（Service.GenerateReport 完整流程）

在 `backend/internal/review/service.go` 实现 `Service.GenerateReport(ctx, leasedJob LeasedJob) (Result, error)` orchestrate：

1. 推进 `feedback_reports.status='queued' → 'generating'`（通过 `UpdateFeedbackReportStatus` Phase 1 已实现）。
2. 从 DB 拉取 `practice_sessions` + `practice_turns` + `practice_plans` + `target_jobs` snapshot（read-only）。
3. 调用 Phase 2 `generateReportContent` + `assessQuestionsForAllTurns`；失败 → goto Failure 分支。
4. 调用 Phase 3 `computeReadinessTier` + `selectRetryFocusTurnIDs` + `decideNextAction`。
5. 组装 `PersistReportInput`：highlights/issues/next_actions（其中 next_actions 第一行的 type=decideNextAction 结果）+ preparedness_level + provenance + N 行 question_assessments；调用 `store.PersistReport(...)`；成功 → 完成。
6. Failure 分支：组装 `PersistReportFailureInput{ReportID, AsyncJobID, ErrorCode=mapAIErrToB1(err)}`；调用 `store.PersistReportFailure(...)`；如果 IsPermanent=true → 不再 reschedule；如果 IsPermanent=false → async_jobs.status='queued' 等待 reaper 或下次 poll 拿到。
7. F3 resolve 失败 / parse failure 需要在 Service 内显式调 `aiclient.AITaskRunWriter.WriteAITaskRun` 写 row（`Status=AITaskRunStatusFailed` 即 B4 `status='failed'`；`ValidationStatus=ValidationStatusInvalid` 仅在 AI 输出 schema 失败时使用，F3 resolve 失败保持 `ValidationStatus=""`；A3 decorator 不会自动写，因 Complete 未被调用或 parse-after-success）。

#### 4.4 BDD-Gate Phase 4

- BDD-Gate: 验证 `E2E.P0.052` 通过（report 主路径 happy path）

#### 4.5 Phase 4 收口 gate

- `cd backend && go test ./internal/review/... ./internal/store/review/... -count=1` 全部通过
- 单元测试 + repository 集成测试覆盖 PersistReport 成功 + PersistReportFailure 路径
- outbox emitter PII 单元测试通过

### Phase 5: Read handler 接入

**目标**：getFeedbackReport + listTargetJobReports HTTP handler + service + store；user-scoped；越权 → REPORT_NOT_FOUND；status placeholder 透传。

#### 5.1 getFeedbackReport handler

新增 `backend/internal/api/reports/get_feedback_report.go`：

- 通过 generated `ServerInterface` 注册 `GET /reports/{reportId}` handler。
- 从 session middleware 拿 `user_id`；调用 `Service.GetFeedbackReport(ctx, GetFeedbackReportInput{UserID, ReportID})`。
- service 调 `store.GetFeedbackReport(ctx, UserID, ReportID)` → user-scoped FK 查询 + question_assessments join。
- 不存在 / 越权 → `ErrReportNotFound` → handler 返回 `404 + ApiError{code:'REPORT_NOT_FOUND'}`。
- 存在：
  - status='ready'：返回 200 + 完整 FeedbackReport。
  - status='queued' / 'generating'：返回 200 + placeholder（preparednessLevel=null, highlights=[], issues=[], nextActions=[], questionAssessments=[], provenance=null, errorCode=null）。
  - status='failed'：返回 200 + 部分 FeedbackReport（preparednessLevel=null, highlights=[], issues=[], nextActions=[], questionAssessments=[], provenance=null, errorCode=非空）。

#### 5.2 listTargetJobReports handler

新增 `backend/internal/api/reports/list_target_job_reports.go`：

- 通过 generated `ServerInterface` 注册 `GET /targets/{targetJobId}/reports` handler。
- 解析 query: `cursor` + `pageSize` (默认 20，最大 50)。
- **cursor 编码契约**：`base64url(JSON({"createdAt": ISO8601(UTC, ns), "id": uuid}))`，无 padding（`=` trim）。Decode 必须严格反序列化为 struct，拒绝 trailing bytes、未知 key、非法 ISO8601；任何反序列化错误 → `ErrInvalidCursor` → handler 返回 `400 + VALIDATION_FAILED`。该契约由 `backend/internal/review/cursor.go` 实现，并以 `TestCursorEncodeDecodeRoundTrip` / `TestCursorRejectsTampered` / `TestCursorRejectsLegacyFormat` 锁定。
- 调用 `Service.ListTargetJobReports(ctx, ListTargetJobReportsInput{UserID, TargetJobID, Cursor, PageSize})`。
- service 调 `store.ListTargetJobReports(...)` → `SELECT * FROM feedback_reports WHERE target_job_id=$1 AND user_id=$2 AND (created_at, id) < ($cursor_created_at, $cursor_id) ORDER BY created_at DESC, id DESC LIMIT $pageSize+1`；额外一行用于 hasMore 检测。
- 返回 `PaginatedFeedbackReport{items, pageInfo:{nextCursor, pageSize, hasMore}}`。
- cursor 非法 → `400 + VALIDATION_FAILED`。

#### 5.3 Router 挂接

在 `backend/internal/api/server.go`（或当前 router 入口）挂接 2 个 handler；如果当前已存在 `Reports` tag handler 注册位但占位 stub，本 Phase 替换为真实实现。

#### 5.4 单元 + 集成测试

- `get_feedback_report_test.go`：5 case（ready / queued placeholder / generating placeholder / failed / cross-user 404）。
- `list_target_job_reports_test.go`：6 case（空 / 单页 / 多页 + cursor / cursor decode / 非法 cursor 400 / 越权 user 隔离）。
- `cursor_test.go`：encode/decode round-trip + 篡改防御。

#### 5.5 BDD-Gate Phase 5

- BDD-Gate: 验证 `E2E.P0.053` 通过（status 转换 + 分页 + 404）

#### 5.6 Phase 5 收口 gate

- `cd backend && go test ./internal/api/reports/... ./internal/review/... ./internal/store/review/... -count=1` 全部通过
- read handler 200 + 404 路径完整

### Phase 6: 失败语义 / retry policy / 隐私 / observability / legacy-negative

**目标**：固化 D-8 / D-11 / D-21 反查 gate；为 002 / 003 / frontend-report-dashboard handoff 留干净接口。

#### 6.1 AI 失败 graceful 完整 6 路径矩阵

在 `service.go` 完善 6 路径 graceful degrade matrix（参照 backend-practice/003 Phase 3 模式）：

- F3 `registry.ErrPromptUnsupported`
- F3 `registry.ErrLanguageUnsupported`
- A3 secret missing
- A3 timeout
- A3 invalid output（parse error）
- parsed empty content / missing critical field

每路径 mapped 到 B1 error code（F3 ErrPromptUnsupported / F3 ErrLanguageUnsupported → `AI_PROVIDER_CONFIG_INVALID`；A3 secret missing → `AI_PROVIDER_SECRET_MISSING`；A3 timeout → `AI_PROVIDER_TIMEOUT`；A3 invalid output / parsed empty → `AI_OUTPUT_INVALID`），写 `feedback_reports.status='failed', error_code=<B1>`，写 `ai_task_runs(status='failed', validation_status='invalid' 仅在 A3 invalid output 与 parsed empty 路径)`（A3 decorator 写 A3 路径；service 显式写 F3 路径 + parsed empty / parse-after-success 路径）。

- `generate_report_service_test.go::TestGenerateReportFailedMatrix` 表驱动覆盖 6 路径；每路径独立断言 (a) `error_code` 对应 B1 字面量，(b) `ai_task_runs.task_type` 区分 F3-fails-before-A3 时的归属（`report_generate` 主调失败 → `task_type='report_generate'`；`report_assessment` 阶段失败 → `task_type='report_assessment'`），(c) `ai_task_runs.status` 字面量严格 = `failed`，(d) 写入路径（decorator vs service）通过 fake AIClient call counter 与 fake AITaskRunWriter 拦截区分。

#### 6.2 Retry policy

实现 `computeBackoff(attempts, leaseTimeout) time.Duration`：`min(2^attempts * 30s, 30min)`（参数命名与 B4 列名 `attempts` 一致）。

`runner_test.go::TestRunnerRetryPolicyAndPermanentFail` 覆盖 `async_jobs.attempts` 1→2→3→4→5 退避计算 + 第 5 次 permanent failed。

#### 6.3 Privacy / observability gates

补 redaction 单元测试：

- `feedback_reports.highlights/issues/next_actions` jsonb / `question_assessments.strengths/gaps/recommended_framework/dimension_results` jsonb / outbox payload / log structured fields / A3 metric label 不含 `question_text` / `answer_text` / `hint_text` / AI prompt body / response body / provider secret。
- metric label allowlist 单元测试断言命中 F1 allowlist；ai_task_runs 行包含 `task_type='report_generate'` / `'report_assessment'` + `status` ∈ B4 CHECK enum {`success`,`failed`,`timeout`,`fallback`} + `validation_status` ∈ {`ok`,`invalid`,空} + `latency_ms` + `input_tokens` + `output_tokens` + `model_profile_name`。

#### 6.4 Legacy-negative grep

新增 `scripts/lint/backend_review_legacy.py`（或在 `scripts/lint/conventions_drift.py` 内扩展 `--phase backend_review` 段）：

- 在 `backend/internal/review/`、`backend/internal/api/reports/`、`backend/internal/store/review/`、`openapi/fixtures/Reports/`、`test/scenarios/e2e/p0-{052,053,054,055}-*/` 范围扫描：
  - `reportLayout` / `report_layout`
  - 5 档 readiness（`readinessLevel\s*[=:]\s*['"]\w+['"]` 包含 5+ 档值）
  - `readiness_score` 或 `readinessScore` numeric 字段
  - `mistakes_queue` / `mistakesQueue` / `mistake_queue`
  - `drill_builder` / `drillBuilder`
  - `growth_center` / `growthCenter`
  - `report_timeline` / `reportTimeline`
  - `report_form` / `reportForm`
  - 旧 next_action 取值（不在 enum {`retry_current_round`,`next_round`,`review_evidence`} 内的字面量）
  - `review_method_version` 旧字段
  - **错误列名漂移**：`leased_at`、`attempt_count`、`worker_id`（这些不是 B4 async_jobs 列；plan/checklist/test 不应再写）
  - **ai_task_runs.status 字面量误用**：在 ai_task_runs 上下文中出现 `'succeeded'`（B4 enum 是 `success`，不是 `succeeded`；`succeeded` 仅在 `async_jobs.status` 上下文合法）

- negative tests / prohibition docs 允许枚举字面量作为 grep 输入；本 plan / BDD / test docs / backend-review spec §4.5 prohibition rows 不属于实现 / runtime 范围。

#### 6.5 Handoff doc 更新

新增 `backend/internal/review/README.md`：

- 简明 handoff 段落，记录 001 新增 endpoint / runner / 中间件挂法 / 接口签名 / handoff 给 frontend-report-dashboard 与 future backend-async-runner 的边界。
- 引用 D-13 inline runner 边界与未来收干路径。

#### 6.6 BDD-Gate Phase 6

- BDD-Gate: 验证 `E2E.P0.054` 通过（AI 失败 graceful + retry policy）
- BDD-Gate: 验证 `E2E.P0.055` 通过（cross-user + privacy + legacy-negative）

#### 6.7 Phase 6 收口 gate

- `cd backend && go test ./... -count=1` 全绿
- `make codegen-check` / `make validate-fixtures` / `migrations/lint.sh`（若 Phase 0.3 修订）/ `make lint-events` / `make codegen-events-check` / `python3 scripts/lint/conventions_drift.py --repo-root .` 全部通过
- `python3 scripts/lint/backend_review_legacy.py --repo-root . --phase all` 通过
- 001 实现 / runtime 输出范围内无 §3.5 R21 列出的 legacy 术语；本 plan / BDD / test docs 与 prohibition rows 仅允许作为负向断言出现

## 5 验收标准

- Phase 0 ~ Phase 6 checklist 全部勾选
- 关联 BDD 场景 `E2E.P0.052` / `E2E.P0.053` / `E2E.P0.054` / `E2E.P0.055` 均由对应 Go HTTP scenario tests 执行通过
- B1 `REPORT_NOT_FOUND` 错误码已新增；B2 `getFeedbackReport` 404 response 已关联；若需要扩 B4 `report_assessment` CHECK 则已 Header bump + history append
- `make codegen-check` / `make lint-events` / `make codegen-events-check` / `python3 scripts/lint/conventions_drift.py --repo-root .` / `cd backend && go test ./...` 全绿
- 001 实现 / runtime 输出范围内无 §3.5 R21 列出的 legacy 术语 / 旧 readiness 5 档 / 旧 reportLayout / 报告时间线 / 多形态 report；负向测试与 prohibition docs 不计入 runtime 输出范围

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| F3 baseline prompt / rubric 在 D-4 加权阈值上未给出明确 `score_levels` → numeric 映射，或 prompt body 仍要求 raw `question` / `answer` / `transcript` / verbatim quote | Phase 0.6 preflight assert F3 rubric `score_levels` 包含 weak/developing/proficient/strong 4 档，prompt body 不要求 raw transcript / verbatim evidence；如缺失或违反隐私红线，Phase 0 必须先回 F3 owner 修订 `report.generate` / `report.question_assessment` v0.1.0 baseline |
| A3 timeout 设置（30s 主调 + 15s/题 评估）在 turn 数较多（>10）时累积超过 worker lease timeout（5min） | Phase 1 lease_timeout 默认 5min 是单 job 总预算；如 A3 调用累积超过预算 → reaper 回退该 job → 下次 poll 拿到时 `async_jobs.attempts`++（B4 列名）；用户感知 generating 屏多轮询几次；后续 plan 002 可考虑 worker 内 progress checkpoint 或 lease renew |
| `feedback_reports.retry_focus_turn_ids` + provenance 3 列（`language`/`feature_flag`/`data_source_version`）在当前 B4 baseline 不存在 | Phase 0.5 **无条件** pre-launch baseline rebase（与 0.3 `report_assessment` CHECK 扩值同 commit；grep 已确认现状）；migration test 断言 4 列默认值生效 |
| `FeedbackReport.errorCode` 字段在 B2 schema 未声明 | grep 已确认 schema 未声明；Phase 0.4 **无条件**与 0.2 同 commit 扩 B2 schema；fixture 仅在 schema 暴露后写 |
| 并发多 worker 时 reaper 误回退正在执行的 job | Phase 1 `lease_timeout` 必须大于 worker 最长 AI 调用预算（30s 主调 + 15s × 最多 turn 数）；reaper 周期 = lease_timeout / 2；reaper UPDATE 条件 `locked_at < now() - lease_timeout`（B4 列名）严格小于；reaper 只对 `job_type='report_generate'` 行作用，避免触动 `targetjob.Drainer` 行；reaper_test.go 用 fake clock 验证 |
| review.Runner 与 `targetjob.Drainer` 共存可能导致开发者在新增 job_type 时误注册到错误 runner | spec D-16 显式 dual-runner 边界；review.Runner 只声明 `report_generate` job_type；新增 job_type 时先确认 owner 边界，并通过 `cmd/api` wire-up 选择正确 runner 注册路径；未来 `backend-async-runner` plan 收干 |
| ai_task_runs.status 列名/枚举与代码生成器或 lint 不一致（旧版可能习惯 `'succeeded'`） | plan §3.5 R21 在 legacy grep 中显式拦截 ai_task_runs 上下文的 `'succeeded'` 字面量；测试 helper 在 fake AITaskRunWriter 拒绝非 B4 enum 值 |
| `practice.session.completed` outbox 被消费两次（dispatcher + backend-review）| backend-practice D-32 已锁定 `triggerEventSemantic: source_event_only` + handler 端 `async_jobs UNIQUE(job_type, dedupe_key)` 兜底；本 plan worker 仅消费 `async_jobs` queued 行（已存在于 DB），不直接消费 outbox event；不会触发二次创建 |
| `question_assessments` 写入失败但 `feedback_reports` 已更新成 `ready` | PersistReport 必须在单事务内同时写两表 + outbox + audit；事务失败回滚整体；事务约束 + 单元测试断言 |
| F3 resolve 失败时 A3 Complete 未调用，A3 decorator 不会写 failed row | Service 必须在 F3 失败分支显式调用 `aiclient.AITaskRunWriter.WriteAITaskRun(task_type='report_generate', status='failed', validation_status='', error_code='AI_PROVIDER_CONFIG_INVALID')`；与 backend-practice/003 D-37 同模式 |
| 跨 user_id 的 `target_jobs` cross-user 查询绕过 backend-targetjob middleware | 本 spec handler 内对 `feedback_reports.user_id` 做兜底过滤；listTargetJobReports 的 `target_job_id` 越权由 backend-targetjob middleware 拦截（已 active），本 plan 不重复验证 |
| 002 / 003 future plan 时漏修 001 已 hardcoded 的阈值 / 取值 | spec §3 D-4/D-5/D-6 锁定值；002 修改阈值必须先修订 spec 决策行；运行时阈值通过 const 暴露，002 可重写但必须有 spec 修订证据 |
