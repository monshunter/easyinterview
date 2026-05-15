# 001 — Report Generation Baseline BDD Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-15

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.052 report 主路径

- [ ] 在 `backend/cmd/api/reports_http_scenario_test.go` 新增 `TestE2EP0052ReportGenerationHappyPath`：1 个 user/report/session/target_job 命名空间（不与 backend-practice 002+003 已占用编号冲突）
- [ ] 准备 fake F3 RegistryClient + fake AIClient + fake AITaskRunWriter；fake AIClient 为 `report.generate` 返回合法 JSON（含 highlights/issues/next_actions 各 ≥ 1 行 + content_summary），为 `report.question_assessment` 对每 turn 返回 dimension_results map (4 维度 × score_level；写入 wire 前映射为 B2 DimensionStatus)
- [ ] 实现 setup：用真实 SQL fake / mem store 写入 1 个 user + 1 个 target_job + 1 个 practice_plan + 1 个 practice_session（status='completed', turn_count=3, language='en'）+ 3 个 practice_turns（turn_index=1/2/3, status='assessed', answer_summary='...'）+ 1 个 feedback_reports(status='queued') placeholder + 1 个 async_jobs(report_generate, status='queued', dedupe_key=session_id, attempts=0, locked_at=null)
- [ ] 实现 trigger：启动 review runner（用 fake clock 触发 1 次 poll loop）
- [ ] 实现 verify：
  - lease：async_jobs.status='running'，`attempts=1`（B4 列名 `attempts`，**非 attempt_count**），`locked_at` 非空（**不断言 `worker_id` 列；该列在 B4 baseline 不存在**）
  - status 推进：feedback_reports.status='queued' → 'generating' → 'ready'
  - F3 + A3 调用：fake AIClient 计数 `report.generate`=1 + `report.question_assessment`=3，按 turn_index 升序
  - feedback_reports.preparednessLevel 非空且 ∈ {not_ready, needs_practice, basically_ready, well_prepared}
  - feedback_reports.highlights / issues / next_actions jsonb 非空数组，第一行 next_actions.type ∈ {retry_current_round, next_round, review_evidence}
  - feedback_reports.retry_focus_turn_ids jsonb 数组（Phase 0.5 新增列；可能为空但字段存在）
  - feedback_reports.language='en' / feature_flag 非空 / data_source_version 非空（Phase 0.5 新增列，wire provenance 6 字段来源）
  - feedback_reports.prompt_version / rubric_version / model_id / provider 非空（生成 provenance 余下 4 字段）
  - feedback_reports.generated_at 非空
  - question_assessments 写入 3 行（turn_index 1/2/3）每行含 dimension_results map + overall_status + confidence + strengths + gaps + recommended_framework + review_status + included_in_retry_plan；dimension_results.status 仅允许 needs_work / meets_bar / strong，weak/developing/proficient 仅作为内部 score_level 输入
  - outbox_events 1 行 event_name='report.generated' aggregate_id=report_id payload schema 一致（reportId, sessionId, targetJobId, preparednessLevel, questionIssueCount, promptVersion, rubricVersion, modelId）
  - audit_events 1 行 event_type='report_generated' metadata 含 status / preparedness_level / language / target_job_id
  - async_jobs.status='running' → 'succeeded' + completed_at 非空 + `locked_at=null`（B4 enum 用 'succeeded'，与 `targetjob.SQLStore.FinalizeAsyncJob` 一致）
  - ai_task_runs 写入 4 行：1 行 `task_type='report_generate', status='success'`（B4 enum，**非 'succeeded'**）+ `validation_status='ok'` + 3 行 `task_type='report_assessment', status='success', validation_status='ok'`；每行含 model_profile_name + `input_tokens` + `output_tokens`（B4 列名，非旧 prompt/completion token 列名）+ latency_ms
  - 隐私红线：feedback_reports.highlights / issues / next_actions / question_assessments.strengths / gaps / recommended_framework / dimension_results jsonb 不含 raw `question_text` / `answer_text` / `hint_text` literal（用 grep 字段 byte content）
  - provenance wire 6 字段断言：`GET /reports/{reportId}` 返回的 `provenance` JSON keys 严格 = {`promptVersion`,`rubricVersion`,`modelId`,`language`,`featureFlag`,`dataSourceVersion`}；6 字段值全部来自 `feedback_reports` 单表，read 路径不 JOIN `ai_task_runs`
- [ ] 实现 cleanup：按 [bdd-plan §6](./bdd-plan.md#6-数据隔离与污染恢复) 顺序删除自身资源
- [ ] 执行 `cd backend && go test ./cmd/api -run TestE2EP0052ReportGenerationHappyPath -count=1`
- [ ] 记录验证证据到 plan §3.6 L2 修订说明（如经过 L2 review）或本 checklist 收口段

## E2E.P0.053 getFeedbackReport status placeholder + listTargetJobReports 分页 + cross-user 404

- [ ] 在 `backend/cmd/api/reports_http_scenario_test.go` 新增 `TestE2EP0053ReportReadAndListing`：2 个 user (A, B) + 1 个 target_job(T owned by A) + 22 个 reports owned by A on T（含 4 个 status 状态：R1 ready 完整 / R2 queued / R3 generating / R4 failed + errorCode='AI_PROVIDER_TIMEOUT'，另 18 个 ready 用于分页）+ 1 个 target_job(T' owned by B) + 0 个 reports on T'
- [ ] 准备 fake auth middleware 让用户 A / 用户 B 分别 authenticate
- [ ] 实现 setup：写入上述资源；ensure created_at 间隔可被 cursor 区分
- [ ] 实现 trigger：作为 user A 调 ① `GET /reports/R1` ② `GET /reports/R2` ③ `GET /reports/R3` ④ `GET /reports/R4` ⑤ `GET /targets/T/reports?pageSize=20`（取 nextCursor） ⑥ `GET /targets/T/reports?cursor=<nextCursor>&pageSize=20` ⑦ `GET /targets/T/reports?pageSize=20&cursor=<篡改值>` ⑧ `GET /targets/T'/reports`（T' 属于 B，预期 backend-targetjob middleware 拦截）；作为 user B 调 ⑨ `GET /reports/R1`
- [ ] 实现 verify：
  - ① 200 + FeedbackReport{id=R1, status='ready', preparednessLevel 非空, highlights/issues/nextActions/questionAssessments 非空, provenance={6 字段}, errorCode null/未声明}
  - ② 200 + status='queued' + 内容字段=空 + provenance=null + errorCode=null
  - ③ 200 + status='generating' + 同 ② 内容
  - ④ 200 + status='failed' + errorCode='AI_PROVIDER_TIMEOUT' + 内容字段=空
  - ⑤ 200 + items=20（created_at DESC 排序）+ pageInfo.nextCursor 非空 + pageSize=20 + hasMore=true
  - ⑥ 200 + items=2 + hasMore=false + nextCursor=null
  - ⑦ 400 + ApiError{code='VALIDATION_FAILED'}
  - ⑧ 404（由 backend-targetjob middleware；不在 backend-review handler 内）
  - ⑨ 404 + ApiError{code='REPORT_NOT_FOUND'}
  - 所有 response.provenance（若非 null）JSON keys 集合严格 = {promptVersion, rubricVersion, modelId, language, featureFlag, dataSourceVersion}
  - 任何 runtime 字段（`feature_key` / `model_profile_name` / `provider` / `cost` / `latency` / `capability`）在 wire JSON 中零出现
- [ ] 实现 cleanup：按隔离顺序删除资源
- [ ] 执行 `cd backend && go test ./cmd/api -run TestE2EP0053ReportReadAndListing -count=1`
- [ ] 记录验证证据

## E2E.P0.054 AI 失败 graceful + retry policy

- [ ] 在 `backend/cmd/api/reports_http_scenario_test.go` 新增 `TestE2EP0054ReportAIFailureAndRetry`：7 个 user/report/session/async_job 命名空间（每个对应一种失败模式）
- [ ] 准备 fake F3 + fake AIClient + fake AITaskRunWriter + fake clock；按场景分支配置：
  - report A：F3 `report.generate` ResolveActive 返回 `registry.ErrPromptUnsupported`
  - report B：F3 `report.question_assessment` ResolveActive 返回 `registry.ErrLanguageUnsupported`
  - report C：A3 fake AIClient 返回 `AI_PROVIDER_SECRET_MISSING`
  - report D：A3 fake AIClient 返回 timeout
  - report E：A3 fake AIClient 返回 invalid JSON content
  - report F：A3 成功但 parsed 出 highlights=[], issues=[], next_actions=[]（empty content）
  - report G：注入 `async_jobs.attempts=4`（B4 列名）+ A3 timeout（第 5 次失败模拟）
- [ ] 实现 setup：写入 7 套 user/session/turns/feedback_reports(status='queued')/async_jobs(status='queued', attempts=0, locked_at=null)；report G 的 `async_jobs.attempts=4`
- [ ] 实现 trigger：启动 review runner（fake clock 触发 poll loop 多次以覆盖所有 7 个）
- [ ] 实现 verify：
  - report A：feedback_reports.status='failed', error_code='AI_PROVIDER_CONFIG_INVALID'; lease 后 `async_jobs.attempts=1`，failure finalize 不二次递增；available_at=now()+60s（2^1×30s）, status='queued', locked_at=null; outbox `report.generation.failed{retryable:true}`; audit `event_type='report_generation_failed', is_permanent:false`; ai_task_runs 显式写一行 `task_type='report_generate', status='failed'（B4 enum）, validation_status='' 或省略, error_code='AI_PROVIDER_CONFIG_INVALID'`（由 service 显式写，因 Complete 未被调用）
  - report B：同 A 但 error_code='AI_PROVIDER_CONFIG_INVALID'（F3 错误统一映射），ai_task_runs 写 `task_type='report_assessment'`（评估阶段失败）
  - report C：error_code='AI_PROVIDER_SECRET_MISSING'；ai_task_runs 由 A3 decorator 自动写 `task_type='report_generate', status='failed', error_code='AI_PROVIDER_SECRET_MISSING'`
  - report D：error_code='AI_PROVIDER_TIMEOUT'；A3 decorator 写 `status='failed'`（或 B4 enum `'timeout'`，按 decorator 实际语义；测试断言列入 status enum 集合）
  - report E：error_code='AI_OUTPUT_INVALID'；A3 decorator 写 `status='failed'`, `validation_status='invalid'`
  - report F：error_code='AI_OUTPUT_INVALID'；service 显式写额外 row `status='failed', validation_status='invalid'`（parse-after-success 路径）
  - report G：初始 attempts=4，lease 后 `async_jobs.attempts=5`；failure finalize 不二次递增，status='failed'（permanent，B4 async_jobs.status enum 用 'succeeded'/'failed'）+ locked_at=null; feedback_reports.status='failed'; outbox `report.generation.failed{retryable:false}`; audit `is_permanent:true`; 不再 reschedule（fake clock 推进 30min 也不触发 retry）
  - 所有 7 个 HTTP 客户端查询 `GET /reports/{reportId}` 时返回 200 + placeholder 或 200 + failed + errorCode，**永不**返回 502 / 503
  - 所有失败路径：ai_task_runs.`error_code` 来自 B1 enum 不含 raw provider message；ai_task_runs.`status` ∈ B4 CHECK enum {`success`,`failed`,`timeout`,`fallback`}（**严格非 'succeeded'**）；log structured fields / A3 metric label / audit_events.metadata / outbox payload 不含 `question_text` / `answer_text` / `hint_text` / AI prompt body / response body / provider secret
- [ ] 实现 cleanup
- [ ] 执行 `cd backend && go test ./cmd/api -run TestE2EP0054ReportAIFailureAndRetry -count=1`
- [ ] 记录验证证据

## E2E.P0.055 cross-user + privacy + legacy-negative

- [ ] 在 `backend/cmd/api/reports_http_scenario_test.go` 新增 `TestE2EP0055ReportPrivacyAndLegacy`
- [ ] 准备 fake F3 RegistryClient + fake AIClient + fake AITaskRunWriter + log / metric / audit 收集器
- [ ] 实现 setup：2 个 user (A, B)；用户 A 拥有 1 个 feedback_reports(status='ready') + 3 行 question_assessments + 完整 outbox + audit 行（由 happy path 简化版生成）；用户 B 存在但对 user A 资源无权限
- [ ] 实现 trigger：
  - 作为 user B 调 `GET /reports/R_A` → 404
  - 作为 user B 调 `GET /targets/T_A/reports` → 404（由 backend-targetjob middleware）
  - 让 user A 跑一次完整 happy path 生成 report，收集所有持久化与运行时输出
  - 运行 `python3 scripts/lint/backend_review_legacy.py --repo-root . --phase all`
- [ ] 实现 verify：
  - cross-user 404 + REPORT_NOT_FOUND envelope 不泄露 R_A 存在性
  - 完整 happy path 后扫描结果：`question_text` / `answer_text` / `hint_text` / AI prompt body / response body / provider secret 在以下持久化与运行时输出中零出现：
    - log structured fields（结构化日志收集器）
    - A3 `ai_task_*` metric label（命中 F1 allowlist；不含 `feature_key` / prompt-rubric version / provider raw model id）
    - audit_events.metadata（jsonb 字段 byte content scan）
    - outbox_events.payload（jsonb）
    - feedback_reports.highlights / issues / next_actions（jsonb）
    - question_assessments.strengths / gaps / recommended_framework / dimension_results（jsonb）
    - ai_task_runs typed columns（含 model_profile_name / input_tokens / output_tokens / latency_ms / validation_status / error_code 但不含 raw text）
  - ai_task_runs 行 task_type ∈ {'report_generate', 'report_assessment'}（合法 B4 CHECK）
  - FeedbackReport.provenance JSON 严格只含 6 wire keys；任何 runtime 字段在 wire JSON 中零出现
  - scoped legacy grep 工具断言（在 `backend/internal/review/` / `backend/internal/api/reports/` / `backend/internal/store/review/` / `openapi/fixtures/Reports/` / scenario runtime assets / generated tests）：
    - `reportLayout` / `report_layout` 零出现
    - 5 档 readiness（旧字面量）零出现
    - `readiness_score` / `readinessScore` numeric 字段零出现
    - `mistakes_queue` / `mistakesQueue` / `mistake_queue` 零出现
    - `drill_builder` / `drillBuilder` 零出现
    - `growth_center` / `growthCenter` 零出现
    - `report_timeline` / `reportTimeline` 零出现
    - `report_form` / `reportForm` 零出现
    - 旧 next_action 取值（不在 enum {`retry_current_round`,`next_round`,`review_evidence`} 内的字面量）零出现
    - `review_method_version` 旧字段零出现
    - 错误列名 `leased_at` / `attempt_count` / `worker_id` 在实现 / runtime / 测试范围零出现（B4 baseline 列名是 `locked_at` / `attempts`，且没有 `worker_id`）
    - ai_task_runs 上下文的 `'succeeded'` 字面量零出现（B4 `ai_task_runs.status` enum 用 `success`，`succeeded` 仅在 `async_jobs.status` 合法）
    - 本 plan / BDD / test docs / spec §4.5 prohibition rows / `scripts/lint/backend_review_legacy.py` 自身允许枚举字面量作为禁止性断言
- [ ] 实现 cleanup
- [ ] 执行 `cd backend && go test ./cmd/api -run TestE2EP0055ReportPrivacyAndLegacy -count=1`
- [ ] 记录验证证据

## 收口

- [ ] `cd backend && go test ./cmd/api -run 'TestE2EP0052|TestE2EP0053|TestE2EP0054|TestE2EP0055' -count=1` 全绿
- [ ] `python3 scripts/lint/backend_review_legacy.py --repo-root . --phase all` 通过（含 001 scoped legacy 反查项）
- [ ] `python3 -m pytest scripts/lint/backend_review_legacy_test.py -q` 通过
