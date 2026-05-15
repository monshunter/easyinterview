# 001 — Report Generation Baseline BDD Plan

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-15

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 0 BDD 框架与编号

本 plan 的 4 个 BDD 场景保留 `E2E.P0.xxx` 编号与 Given / When / Then 语义；001 本次代码交付的可执行入口落在 `backend/cmd/api/reports_http_scenario_test.go` 的 HTTP scenario tests，用 `cmd/api` 真实路由、middleware、service/store fake 与 response / side-effect 断言覆盖用户可见 API 行为（与 backend-practice/002 + 003 同模式）。

- 套件: `e2e`
- 阶段: `P0`
- 已占用编号现状（[`test/scenarios/e2e/INDEX.md`](../../../../../test/scenarios/e2e/INDEX.md)）：`001-006`, `010-047`；backend-practice/003 已通过 Go HTTP scenario 预留 `048-051`（hint 四场景，未挂 e2e INDEX 目录）。本 plan 在空闲号段 `052-055` 中分配 4 个场景
- 编号分配: `E2E.P0.052` / `E2E.P0.053` / `E2E.P0.054` / `E2E.P0.055`
- 执行入口: `cd backend && go test ./cmd/api -run 'TestE2EP0052|TestE2EP0053|TestE2EP0054|TestE2EP0055' -count=1`
- 外部 Kind / shell 场景资产: 001 不新增 `test/scenarios/e2e/p0-NNN-*` 目录；若未来需要 Kind live BDD，可由 scenarios owner 按当前编号把同一 Given / When / Then 提升为 `test/scenarios` 资产，不改变本 plan 的 API 行为语义

每个场景的执行证据在 [bdd-checklist](./bdd-checklist.md) 跟踪；本文件只记录场景的 Given / When / Then 与覆盖范围，不出现执行 checkbox。

## 1 场景矩阵

| 场景 ID | 名称 | 类别 | 关联 Plan Phase | 关联 spec AC / D |
|---------|------|------|----------------|-------------------|
| `E2E.P0.052` | report 主路径（complete handoff → runner lease → AI generate → ready + outbox + question_assessments + ai_task_runs） | primary | Phase 1 + Phase 2 + Phase 3 + Phase 4 | C-1, C-2, C-3, D-4, D-5, D-6 |
| `E2E.P0.053` | getFeedbackReport status 转换（queued/generating/ready/failed placeholder）+ listTargetJobReports 分页 + cross-user 404 | alternate + UX (API) | Phase 5 | C-4, C-5, C-9, D-14, D-15 |
| `E2E.P0.054` | AI 失败 graceful 6 路径 matrix + retry policy + permanent fail + report.generation.failed event + ai_task_runs failed 行 | failure + observability | Phase 6 | C-6, C-7, D-8, D-10 |
| `E2E.P0.055` | cross-user 隔离 + 隐私红线（jsonb / outbox payload / metric label / ai_task_runs typed columns）+ retired 术语 0 出现（scoped legacy grep） | privacy/security + regression | Phase 6 | C-9, C-10, D-11, §4.5 |

## 2 Phase 1 + 2 + 3 + 4 — report 主路径

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| `E2E.P0.052` | report 主路径 | 用户 A 已通过 backend-practice/002 setup 拥有 `feedback_reports(id=R, status='queued', user_id=A, session_id=S, target_job_id=T)` placeholder + `async_jobs(id=J, job_type='report_generate', status='queued', dedupe_key=S, resource_id=R, available_at=now(), attempts=0, locked_at=null)` queued row + `practice.session.completed` outbox source event；`practice_sessions(id=S, status='completed', turn_count=3, language='en')`；3 个 `practice_turns(turn_index=1/2/3, status='assessed', answer_summary=...)`；F3 active `report.generate` + `report.question_assessment` v0.1.0 baseline；A3 fake AIClient 配置为 `report.generate` 返回合法 JSON（含 highlights/issues/next_actions 各 ≥ 1 行）+ `report.question_assessment` 对每 turn 返回 dimension_results map (4 维度 × score_level，最终 wire status 映射为 B2 DimensionStatus) + overall_status + confidence + strengths + gaps + recommended_framework | review runner Start + 自然 poll 周期内 lease J | （A）lease 成功：`async_jobs.status='queued' → 'running', attempts=0 → 1, locked_at=now()`（B4 列名；**无 worker_id 列断言**）；（B）status 推进：`feedback_reports.status='queued' → 'generating'`；（C）F3 + A3 调用：`report.generate` 1 次 + `report.question_assessment` 3 次（每 turn 一次，按 turn_index 升序）；（D）AI 输出解析成功 → ReadinessTier 算出（按 D-4 加权阈值，内部 score_level 不暴露到 B2 wire status）+ retry_focus_turn_ids 选出（按 D-5）+ next_action enum 决定（按 D-6）；（E）单事务持久化：`feedback_reports.status='generating' → 'ready', preparednessLevel='basically_ready'（或对应阈值档）, highlights/issues/next_actions/language='en'/feature_flag/data_source_version/retry_focus_turn_ids/prompt_version/rubric_version/model_id/provider/generated_at` 写入（Phase 0.5 新增的 4 列写实值）；3 行 `question_assessments(report_id=R, turn_id=$turn, ...)` 写入（含 dimension_results / overall_status / confidence / strengths / gaps / recommended_framework / review_status / included_in_retry_plan，且 dimension_results.status 仅为 needs_work / meets_bar / strong）；`outbox_events(event_name='report.generated', event_version=1, aggregate_id=R, payload={reportId:R, sessionId:S, targetJobId:T, preparednessLevel, questionIssueCount, promptVersion, rubricVersion, modelId})` 写一行；`audit_events(event_type='report_generated', metadata={status,preparedness_level,language,target_job_id})` 写一行；`async_jobs.status='running' → 'succeeded', completed_at=now(), locked_at=null`（async_jobs.status enum 用 `succeeded`，与 B4 baseline 一致）；（F）ai_task_runs 写入：1 行 `task_type='report_generate', status='success'（B4 enum，非 'succeeded'）, validation_status='ok', input_tokens, output_tokens, latency_ms, model_profile_name='report.generate.default'`；3 行 `task_type='report_assessment', status='success', validation_status='ok', model_profile_name='report.assessment.default'`；（G）所有写入字段 `feedback_reports.highlights/issues/next_actions` jsonb + `question_assessments.strengths/gaps/recommended_framework/dimension_results` jsonb 不含 raw `question_text` / `answer_text` / `hint_text` literal；（H）provenance wire 6 字段全部源自 `feedback_reports` 列读取，不 JOIN `ai_task_runs` | `backend/cmd/api/reports_http_scenario_test.go::TestE2EP0052ReportGenerationHappyPath` |

## 3 Phase 5 — read handler 路径

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| `E2E.P0.053` | getFeedbackReport status placeholder + listTargetJobReports 分页 + cross-user 404 | 用户 A 拥有 4 个 report：R1 status='ready'（完整字段）/ R2 status='queued' / R3 status='generating' / R4 status='failed' + errorCode='AI_PROVIDER_TIMEOUT'；用户 A 对同一 target_job=T 拥有 22 个 report（多页测试用，含上述 4 个）；用户 B 存在但对 T 无 report | 用户 A：① `GET /reports/R1` → 200；② `GET /reports/R2` → 200 placeholder；③ `GET /reports/R3` → 200 placeholder；④ `GET /reports/R4` → 200 + errorCode；⑤ `GET /targets/T/reports?pageSize=20`（首页）→ 200 + 20 行 + nextCursor 非空 + hasMore=true；⑥ `GET /targets/T/reports?cursor=<上一步 nextCursor>&pageSize=20` → 200 + 2 行 + hasMore=false + nextCursor=null；⑦ `GET /targets/T/reports?pageSize=20&cursor=<篡改>` → 400 VALIDATION_FAILED；⑧ `GET /targets/T'/reports`（T' 属于另一用户）→ backend-targetjob middleware 拦截 404；用户 B：⑨ `GET /reports/R1` → 404 REPORT_NOT_FOUND | （①）200 + `FeedbackReport{id:R1, status:'ready', preparednessLevel:非空, highlights:[..]非空, issues:[..], nextActions:[..], questionAssessments:[..]非空（每行含完整 dimension_results）, provenance:{promptVersion, rubricVersion, modelId, language, featureFlag, dataSourceVersion} 6 字段, errorCode:null/未声明}`；（②）200 + `FeedbackReport{status:'queued', preparednessLevel:null, highlights:[], issues:[], nextActions:[], questionAssessments:[], provenance:null, errorCode:null}`；（③）同 ② 但 status='generating'；（④）200 + `FeedbackReport{status:'failed', preparednessLevel:null, ..., errorCode:'AI_PROVIDER_TIMEOUT'}`；（⑤）200 + `PaginatedFeedbackReport{items:[20 行 by created_at DESC], pageInfo:{nextCursor:'<base64>', pageSize:20, hasMore:true}}`；（⑥）200 + items=2 + hasMore=false + nextCursor=null；（⑦）400 + `ApiError{code:'VALIDATION_FAILED'}`；（⑧）404 由 backend-targetjob middleware 返回；（⑨）404 + `ApiError{code:'REPORT_NOT_FOUND'}`；所有 response 的 `provenance` 字段（若非 null）严格只含 6 wire keys，任何 runtime 字段（`feature_key` / `model_profile_name` / provider / cost / latency / capability）零出现 | `backend/cmd/api/reports_http_scenario_test.go::TestE2EP0053ReportReadAndListing` |

## 4 Phase 6 — AI 失败 graceful + retry policy

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| `E2E.P0.054` | AI 失败 graceful 6 路径 + retry policy + permanent fail | 用户 A 拥有 7 个 `feedback_reports(status='queued')` + 7 个对应 `async_jobs(report_generate, status='queued', attempts=0, locked_at=null)`；按 fake 失败配置：① F3 `report.generate` ResolveActive 返回 `registry.ErrPromptUnsupported`；② F3 `report.question_assessment` ResolveActive 返回 `registry.ErrLanguageUnsupported`；③ A3 fake AIClient 返回 `AI_PROVIDER_SECRET_MISSING`；④ A3 fake AIClient 返回 timeout；⑤ A3 fake AIClient 返回 invalid JSON content；⑥ A3 成功但 parsed AI 输出空 highlights/issues/next_actions；⑦ 注入 `async_jobs.attempts=4`（lease 后将成为第 5 次尝试）+ A3 timeout；fake clock；log / metric / audit / ai_task_runs 收集器可读 | 对每个 report 让 review runner lease 并处理 | （①-⑥）每路径执行：（a）`feedback_reports.status='queued' → 'generating' → 'failed', error_code ∈ {AI_PROVIDER_CONFIG_INVALID（① + ②）, AI_PROVIDER_SECRET_MISSING（③）, AI_PROVIDER_TIMEOUT（④）, AI_OUTPUT_INVALID（⑤ + ⑥）}, generated_at=now()`；（b）lease 已把 `async_jobs.attempts` 从 0 更新为 1，failure finalize 不二次递增；使用 attempts=1 计算 `available_at=now()+backoff(attempts)`（60s），`status='queued', locked_at=null`（attempts < 5）；（c）outbox `report.generation.failed{reportId, sessionId, errorCode, retryable=true}` 发出；（d）`audit_events(event_type='report_generation_failed', metadata={error_code, session_id, target_job_id, attempts, is_permanent:false})`；（e）`ai_task_runs` 写入对应 task_type 失败行，每行 `status='failed'`（B4 enum，**非 'succeeded'**），`validation_status='invalid'` 仅⑤⑥（A3 invalid output / parse-after-success），其它路径 `validation_status=''` 或 `'ok'`；① 由 service 显式写 `task_type='report_generate'`（F3 主调失败，A3 decorator 未触发）；② 由 service 显式写 `task_type='report_assessment'`（F3 assessment 失败）；③④⑤ 由 A3 decorator 自动写；⑥ 由 service 显式写 `task_type='report_generate'` 因 parse-after-success；（f）HTTP 响应：客户端 `GET /reports/{reportId}` 在 generating → failed 推进时返回 200 + placeholder → 200 + failed + errorCode；不返回 502 / 503；（⑦）lease 后 attempts=5 时：`async_jobs.status='failed'`（permanent，B4 async_jobs.status enum）+ `locked_at=null`；`feedback_reports.status='failed'`；outbox `report.generation.failed{retryable:false}`；不再 reschedule；audit `is_permanent:true`；（隐私）所有失败路径 `ai_task_runs.error_code` 来自 B1 enum 不含 raw provider message；log structured fields / A3 metric label / audit_events.metadata / outbox payload 不含 `question_text` / `answer_text` / `hint_text` / AI prompt body / response body / provider secret | `backend/cmd/api/reports_http_scenario_test.go::TestE2EP0054ReportAIFailureAndRetry` |

## 5 Phase 6 — cross-user + privacy + legacy-negative

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| `E2E.P0.055` | cross-user 隔离 + 隐私红线 + retired 术语 0 出现 | 用户 A / 用户 B；用户 A 拥有 1 个 `feedback_reports(status='ready')` 与对应 `question_assessments` 3 行；fake F3 + fake AIClient 准备完整生成路径；log / metric label / audit_events / outbox_events / ai_task_runs / `feedback_reports.highlights/issues/next_actions` jsonb / `question_assessments.strengths/gaps/recommended_framework/dimension_results` jsonb 收集器可读；scoped legacy grep 工具可调 | （a）用户 B `GET /reports/R_A` → 404；（b）用户 B `GET /targets/T_A/reports` → 由 backend-targetjob 拦截 404；（c）让用户 A 跑一次完整 happy path（同 E2E.P0.052 简化版），完成后扫描所有持久化与运行时输出；（d）运行 `python3 scripts/lint/backend_review_legacy.py --repo-root . --phase all` | （a）404 + REPORT_NOT_FOUND；不泄露 reportX 存在性；（b）404 由 backend-targetjob middleware 返回；（c）扫描结果：`question_text` / `answer_text` / `hint_text` / AI prompt body / response body / provider secret 在 log / metric label / audit_events.metadata / outbox_events.payload / `feedback_reports.highlights/issues/next_actions` jsonb / `question_assessments.strengths/gaps/recommended_framework/dimension_results` jsonb / `ai_task_runs` typed columns 中零出现；A3 `ai_task_*` metric label 命中 F1 allowlist 且不含 `feature_key` / prompt-rubric version / provider raw model id；ai_task_runs 行 `task_type='report_generate'` / `'report_assessment'` 合法，`status` ∈ B4 CHECK enum {`success`,`failed`,`timeout`,`fallback`}（**非 'succeeded'**），`validation_status` ∈ {`ok`,`invalid`,空}；（d）scoped legacy grep 断言：`reportLayout` / `readinessScore` / `readiness_score` / `mistakes_queue` / `mistakesQueue` / `drill_builder` / `drillBuilder` / `growth_center` / `growthCenter` / `report_timeline` / `report_form` / 旧 next_action 取值（不在 enum 集合内的字面量）/ `review_method_version` / 错误列名 `leased_at` / `attempt_count` / `worker_id` / ai_task_runs 上下文的 `'succeeded'` 字面量 在实现 / runtime 输出范围（`backend/internal/review/` / `backend/internal/api/reports/` / `backend/internal/store/review/` / `openapi/fixtures/Reports/` / scenario runtime assets / generated tests）零出现；负向测试文档与禁止性说明（本 plan / checklist / bdd/test docs / spec §4.5 prohibition rows）允许枚举字面量作为禁止性断言；FeedbackReport.provenance JSON 严格只含 6 wire keys；任何 runtime 字段在 wire JSON 中零出现；6 wire 字段全部由 `feedback_reports` 单表读取（不 JOIN `ai_task_runs`） | `backend/cmd/api/reports_http_scenario_test.go::TestE2EP0055ReportPrivacyAndLegacy` |

## 6 数据隔离与污染恢复

每个场景按 `test/scenarios/e2e/README.md` §5 / §3 / §6 / §8 约定：

- 数据隔离：每个场景使用独立的 `user_id` / `report_id` / `session_id` / `target_job_id` / `async_job_id` 命名空间；不复用 `E2E.P0.038 ~ E2E.P0.051` 已占用的资源（含 backend-practice 002 + 003 范围）
- 清理顺序：cleanup 先删自身 `question_assessments` → `feedback_reports` → `async_jobs` → `practice_session_events` / `practice_turns` / `practice_sessions` / `practice_plans` → `idempotency_records` → `ai_task_runs` → `audit_events` → `outbox_events` → `target_jobs` → users
- 污染恢复：场景失败时按 README §8 顺序：① 清理场景自身资源；② 定位并恢复 shared 组件（F3 cache、review runner inflight；001 阶段 backend-async-runner 不存在，仅 review/runner + privacy/runner 两个 inline runner）；③ 仅在 ① ② 失败时 `test/scenarios/env-cleanup.sh && env-setup.sh` 全量重建
- 不预设 Helm chart / 外部 Git 平台名称；所有命令以本仓库脚本为真理源

## 7 与单元测试边界

本 BDD plan 验证用户可见行为切片（HTTP API + DB 状态 + ai_task_runs 写入 + log/metric/audit 红线 + provenance wire 边界 + scoped legacy grep + AI 失败 graceful + retry policy）；不重复内部接口签名、序列化结构、错误映射、ReadinessTier 算法、retry_focus 算法、next_action 决策矩阵等单元测试覆盖（详见 [test-plan](./test-plan.md)）。001 阶段不存在 runtime outbox→asynq dispatcher（与 backend-practice 002/003 一致），"dispatcher 集成测试" 不在本 BDD plan 范围内，由 future `backend-async-runner` plan 承接。

## 8 与 spec AC 映射

| spec AC | 覆盖场景 |
|---------|----------|
| C-1（report 主路径 queued → ready） | `E2E.P0.052` |
| C-2（per-question 维度评估写入） | `E2E.P0.052` 子断言（3 行 question_assessments） |
| C-3（ReadinessTier 计算） | `E2E.P0.052` 子断言（preparedness_level 非空 + 四档之一） |
| C-4（getFeedbackReport status placeholder） | `E2E.P0.053`（queued/generating/failed 三 case） |
| C-5（listTargetJobReports 分页 + 空） | `E2E.P0.053`（22 行多页 + cursor + 空列表通过 fixture） |
| C-6（AI 失败 graceful） | `E2E.P0.054` 6 路径 |
| C-7（retry policy + max attempts） | `E2E.P0.054` ⑦ 子断言（`async_jobs.attempts=5` → permanent；B4 列名） |
| C-8（FeedbackReport.provenance wire 6 字段） | `E2E.P0.053` 子断言 + `E2E.P0.055` 子断言 |
| C-9（cross-user 隔离） | `E2E.P0.053` ⑨ + `E2E.P0.055` （a）（b） |
| C-10（隐私红线） | `E2E.P0.055` （c） |
| C-11（ai_task_runs 行） | `E2E.P0.052` 子断言（4 行）+ `E2E.P0.054` 子断言（失败行）+ `E2E.P0.055` 子断言（typed columns 隐私反查） |
| C-12（单 async_jobs 双 worker 抢占） | 单元测试 `TestLeaseSkipLocked` 覆盖（不进 BDD scenario） |
| D-4 / D-5 / D-6（算法决策） | 单元测试 `TestComputeReadinessTier` / `TestSelectRetryFocusTurns` / `TestDecideNextAction` 覆盖（不进 BDD scenario）；BDD 验证最终输出 |
| D-8 / D-11 / D-15 / §4.5（隐私 + 错误码 + 通用文字） | `E2E.P0.054` + `E2E.P0.055` |
