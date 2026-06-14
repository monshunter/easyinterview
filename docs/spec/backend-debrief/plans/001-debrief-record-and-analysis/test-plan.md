# 001 Debrief Record and Analysis Test Plan

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-06-14

**关联计划**: [plan](./plan.md)
**关联 Spec**: [spec](../../spec.md)

## 0 目标

为 backend-debrief/001-debrief-record-and-analysis 定义单元测试与集成测试矩阵。每个测试项映射到具体测试文件与测试函数，覆盖主路径、边界、失败/恢复、跨层契约、隐私/观测、回归负向 8 类风险。

测试执行入口：
- 后端单元 + 集成：`cd backend && go test ./internal/debrief ./internal/api/debriefs ./internal/store/debrief ./cmd/api -count=1 -race`
- 后端全量：`cd backend && go test ./... -count=1`
- Lint script：`python3 -m pytest scripts/lint -q`
- Contract drift：`make codegen-check` / `make validate-fixtures` / `make lint-events` / `make codegen-events-check`

## 1 Coverage Matrix

| 行 | source | category | plan phase | verification | negative_scope |
|----|--------|----------|------------|--------------|----------------|
| R1 | spec D-1 / OpenAPI POST /debriefs | Primary | Phase 1-2 | Go unit + contract drift | — |
| R2 | spec D-1 / OpenAPI GET /debriefs/{id} | Primary | Phase 5 | Go unit + cross-user test | — |
| R3 | spec D-1 / OpenAPI POST /debriefs/question-suggestions (Phase 0 addendum) | Primary | Phase 3 | Go unit + contract drift | — |
| R4 | spec D-5 / drainer-registered worker handler happy | Primary | Phase 4 | Go unit + integration test | — |
| R5 | spec D-9 / AI failure graceful (F3/A3) | Failure/recovery | Phase 4 | Go unit + integration | — |
| R6 | spec D-9 / retry policy + permanent fail at attempts=5 | Failure/recovery | Phase 4-6 | Go integration | — |
| R7 | spec D-15 / Cross-user 404 isolation | Privacy / Boundary | Phase 5 | Go unit | — |
| R8 | spec D-16 / IK replay + mismatch | Boundary | Phase 1-2 | Go unit | — |
| R9 | spec D-4 / Empty questions array (422) | Boundary | Phase 1 | Go unit | — |
| R10 | spec D-4 / Max question count + length | Boundary | Phase 1 | Go unit | — |
| R11 | spec D-13 / Outbox payload schema (counts only, no raw text) | Cross-layer contract + Privacy | Phase 2 + Phase 4 | Go unit | raw_questions, notes, interviewerReaction, risk_items prose |
| R12 | spec D-14 / Cross-owner addendums (B1/B2/B3/B4/F3) gate | Cross-layer contract | Phase 0 | codegen-check + validate-fixtures + lint-events + migration checks | — |
| R13 | spec D-10 / ai_task_runs typed rows | Observability | Phase 3-4 | Go unit + integration | — |
| R14 | spec D-12 / Privacy redlines (no raw text in events/metric/log/audit) | Privacy | Phase 2 + Phase 4 + Phase 6 | Go unit + grep | — |
| R15 | spec D-11 / GenerationProvenance wire 6 fields only | Cross-layer contract + Privacy | Phase 5 | Go unit | runtime fields (feature_key, model_profile_name, provider, cost, latency) |
| R16 | spec D-2 / status state machine draft→completed only | Cross-layer + Boundary | Phase 4 | Go unit (CAS rejects) | failed / queued / generating / ready wire status |
| R17 | spec D-17 / Cross-owner Practice handoff verification | Cross-layer | Phase 0 | grep + backend-practice tests | — |
| R18 | spec D-3 / Frontend polling contract (getJob + getDebrief) | Cross-layer | Phase 5 | Go unit (getDebrief draft partial) + frontend-debrief side | — |
| R19 | spec D-8 / practiceFocusCount = len(risk_items) | Cross-layer | Phase 4 | Go unit (outbox payload semantic) | — |
| R20 | spec D-18 / DELETE /me CASCADE | Privacy / Cleanup | Phase 5 | migrations test + integration | — |
| R21 | spec §4.5 / Legacy negative: mistakes_count, generatedMistakeCount, experience_library, drill_builder, growth_center, star_editor, debrief_voice | Regression / Legacy-negative | Phase 6 | grep + lint script | mistakes_count, generatedMistakeCount, experience_library, drill_builder, growth_center, star_editor, debrief_voice |
| R22 | F1 metric registration boundary | Observability | Phase 6 | F1 owner co-author + grep | — |
| R23 | spec D-19 / D-20 resumeId suggestion context | Cross-layer contract + Privacy | Phase 8 | store/service/API/cmd-api focused tests + fixture parity + scenario wrapper | resumeVersionId / resume_version_id |
| R24 | E2E.P0.063 / sessionId suggestion context | Cross-layer contract + Privacy | Phase 3 | store/service/API focused tests + real prompt marker replacement + scenario wrapper | cross-user / wrong-target sessionId; unreplaced `{{mock_report_summary}}` |

## 2 测试项明细

### Phase 1: createDebrief handler 与 store 骨架

#### 1.1 TestCreateDebrief_ValidationError_EmptyQuestions
- 文件：`backend/internal/api/debriefs/handler_test.go`
- Given：POST /debriefs body `{targetJobId,questions:[],...}`
- When：handler 处理
- Then：返回 422 + B1 `VALIDATION_FAILED`；不写 debriefs / async_jobs / outbox
- 覆盖：R9

#### 1.2 TestCreateDebrief_ValidationError_LongQuestionText
- Given：单题 `questionText.length > 4000`
- When：handler 处理
- Then：返回 422 + B1 `VALIDATION_FAILED`
- 覆盖：R10

### Phase 2: createDebrief 完整事务 + outbox

#### 2.1 TestStoreCreateDebrief_HappyTransaction
- 文件：`backend/internal/store/debrief/store_test.go`
- Given：合法 CreateDebriefParams
- When：调 store.CreateDebrief
- Then：返回 `(debrief, async_job, nil)`；DB 含 debriefs(status='draft') + async_jobs(queued) + outbox_events(debrief.created)；事务原子
- 覆盖：R1

#### 2.2 TestStoreCreateDebrief_RollbackOnOutboxFailure
- Given：人为注入 outbox insert error
- When：调 store
- Then：事务回滚；debriefs / async_jobs 行不存在；返回 error
- 覆盖：R1, R11

#### 2.3 TestServiceCreateDebrief_AuditEmitted
- 文件：`backend/internal/debrief/service_test.go`
- Given：store mock 返回 happy
- When：service.CreateDebrief
- Then：调用 audit writer 一次 with action='create_debrief'
- 覆盖：R1, R13

#### 2.4 TestCreateDebrief_HappyResponse
- 文件：`backend/internal/api/debriefs/handler_test.go`
- Given：合法 request + auth user
- When：POST /debriefs
- Then：返回 202 + DebriefWithJob{debriefId, job:{jobType:'debrief_generate', status:'queued'}}
- 覆盖：R1

#### 2.5 TestCreateDebrief_IdempotencyReplay_SameBody
- Given：合法 request + IK + body hash A
- When：第二次发送相同 IK + 相同 body hash
- Then：返回相同 202 + 相同 debriefId + 相同 jobId（cached）
- 覆盖：R8

#### 2.6 TestCreateDebrief_IdempotencyMismatch_DifferentBody
- Given：IK X 已用 with body A
- When：发送 IK X with body B (different hash)
- Then：返回 409 + B1 `IDEMPOTENCY_KEY_MISMATCH`
- 覆盖：R8

#### 2.7 TestCreateDebrief_OutboxPayloadSchema
- Given：合法 createDebrief 已完成
- When：读 outbox_events row
- Then：payload keys 集合 = {debriefId, targetJobId, roundType, questionCount}；不含 raw_questions / notes / interviewerReaction 子串
- 覆盖：R11, R14

### Phase 3: suggestDebriefQuestions sync handler

#### 3.1 TestServiceSuggestQuestions_Happy
- 文件：`backend/internal/debrief/service_test.go`
- Given：用户已认证，target_job 属于用户，F3 active，A3 mock 返回有效 JSON
- When：service.SuggestQuestions({targetJobId, language:'zh', count:6})
- Then：返回 6 条 suggestions；写 ai_task_runs(task_type='debrief_suggest_questions', status='success')；audit 一行
- 覆盖：R3, R13

#### 3.2 TestServiceSuggestQuestions_CrossUserTargetJob_403
- Given：target_job 属于用户 B
- When：用户 A 调 service.SuggestQuestions
- Then：返回 B1 `FORBIDDEN`（或 backend-targetjob middleware 拦截 404）；不调用 AI；不写 ai_task_runs / audit
- 覆盖：R3, R7

#### 3.3 TestServiceSuggestQuestions_F3ResolveFailed
- Given：F3 ResolveActive 返回 error
- When：调 service
- Then：返回 B1 `AI_PROVIDER_CONFIG_INVALID`；写 ai_task_runs status='failed' + error_code；audit 一行 with error_code
- 覆盖：R5, R13

#### 3.4 TestServiceSuggestQuestions_A3Timeout
- Given：A3 AIClient mock 返回 timeout
- When：调 service
- Then：返回 B1 `AI_PROVIDER_TIMEOUT`；写 ai_task_runs status='timeout'
- 覆盖：R5, R13

#### 3.5 TestServiceSuggestQuestions_ParseFailed
- Given：A3 返回非 JSON 文本
- When：调 service
- Then：返回 B1 `AI_OUTPUT_INVALID`；写 ai_task_runs status='failed' + validation_status='invalid'
- 覆盖：R5, R13

#### 3.6 TestSuggestDebriefQuestions_CountBoundary
- 文件：`backend/internal/api/debriefs/handler_test.go`
- Given：request count = 0 / 11 / 1 / 10
- When：handler 处理
- Then：count<1 或 count>10 → 422 VALIDATION_FAILED；count=1 / 10 → 正常处理
- 覆盖：R3, R10

#### 3.7 TestSuggestDebriefQuestions_Unauthenticated_401
- Given：无 auth header
- When：handler 处理
- Then：返回 401 + B1 `AUTH_UNAUTHORIZED`
- 覆盖：R3

#### 3.8 TestStoreGetSuggestionContext_LoadsPracticeSessionSummary
- 文件：`backend/internal/store/debrief/store_test.go`
- Given：`target_jobs(user_id=A,id=T)` 与 `practice_sessions(user_id=A,target_job_id=T,id=S,status='completed')` 存在，turn/report derived summary ready
- When：`Repository.GetSuggestionContext({userID:A,targetJobID:T,sessionID:S})`
- Then：返回 `SuggestionContext.SessionSummary`，包含 session id、turn answer summary 与 ready report derived issues；不读取 raw `answer_text`
- 覆盖：R24

#### 3.9 TestServiceSuggestQuestions_SessionContextInPrompt
- 文件：`backend/internal/debrief/service_test.go`
- Given：context store 返回 `SessionSummary`，F3 active prompt 使用真实 marker `{{role_title}}` / `{{job_summary}}` / `{{mock_report_summary}}`
- When：`Service.SuggestQuestions({sessionID:S})`
- Then：`sessionId` 传入 context store；AI prompt 替换真实 marker 且包含 mock-report/session derived summary；返回 source=`mock_report`
- 覆盖：R24

#### 3.10 TestStoreGetSuggestionContext_CrossUserSessionNotFound
- 文件：`backend/internal/store/debrief/store_test.go`
- Given：`sessionId` 不属于当前用户、目标岗位不匹配或 session 未完成
- When：`Repository.GetSuggestionContext({userID:A,targetJobID:T,sessionID:S})`
- Then：返回 `ErrDebriefPrerequisite`；不调用 AI
- 覆盖：R24

#### 3.11 TestSuggestDebriefQuestions_MapsSessionIDToService
- 文件：`backend/internal/api/debriefs/handler_test.go`
- Given：generated `SuggestDebriefQuestionsRequest.sessionId` 非空
- When：handler 处理 request
- Then：domain `SuggestQuestionsRequest.SessionID` 等于 generated request `sessionId`
- 覆盖：R24

### Phase 4: debrief_generate worker handler

#### 4.1 TestGenerateHandler_HappyResolution
- 文件：`backend/internal/debrief/generate_handler_test.go`
- Given：job payload + F3 active + A3 mock 返回 valid JSON {aiAnalyses, riskItems}
- When：handler.Handle(ctx, job)
- Then：返回 nil error；store 调用 UpdateDebriefCompleted；ai_task_runs 写 task_type='debrief_generate' success
- 覆盖：R4

#### 4.2 TestGenerateHandler_PromptContextAssembled
- Given：debrief.raw_questions = N items + target_job 摘要
- When：handler 内部组装 F3 prompt context
- Then：prompt input vars 含 {questions[N], targetJobTitle, jdHighlights, language}；不含 raw notes（D-12 redaction 已在 prompt 层应用如必要）
- 覆盖：R4

#### 4.3 TestDrainer_DebriefGenerateRegistered
- 文件：`backend/cmd/api/main_test.go` 或等价集成测试
- Given：backend 启动
- When：drainer instance 检查 RegisteredHandlers
- Then：含 "debrief_generate" → debrief.GenerateHandler
- 覆盖：R4

#### 4.4 TestStoreUpdateDebriefCompleted_HappyTransaction
- 文件：`backend/internal/store/debrief/store_test.go`
- Given：debriefs(status='draft', id=X) 行已存在
- When：UpdateDebriefCompleted({id:X, raw_questions:[with aiAnalysis], risk_items, provenance...})
- Then：UPDATE 成功；status='completed'；outbox debrief.completed payload={debriefId,targetJobId,riskItemCount,practiceFocusCount}；事务原子
- 覆盖：R4, R11, R16

#### 4.5 TestStoreUpdateDebriefCompleted_CASRejectsCompleted
- Given：debriefs(status='completed', id=X) 行已存在（重复 worker 触发）
- When：UpdateDebriefCompleted({id:X,...})
- Then：CAS UPDATE 影响 0 行；返回 ErrAlreadyCompleted；不发 outbox
- 覆盖：R16

#### 4.6 TestStoreUpdateDebriefCompleted_OutboxRollback
- Given：人为注入 outbox insert error
- When：UpdateDebriefCompleted
- Then：事务回滚；debriefs.status 保持 'draft'；返回 error
- 覆盖：R4

#### 4.7 TestGenerateHandler_OutboxPayloadSchema
- Given：worker 完成
- When：读 outbox debrief.completed row
- Then：payload keys = {debriefId, targetJobId, riskItemCount, practiceFocusCount}；不含 notes / questionText / interviewerReaction / risk_items prose 子串；practiceFocusCount = len(risk_items)
- 覆盖：R11, R14, R19

#### 4.8 TestGenerateHandler_F3ResolveFailed
- Given：F3 ResolveActive 返回 error
- When：handler.Handle
- Then：返回 `targetjob.RetryableError`；ai_task_runs 写 failed row + B1 `AI_PROVIDER_CONFIG_INVALID`；debriefs.status 保持 'draft'；不发 outbox
- 覆盖：R5

#### 4.9 TestGenerateHandler_A3Timeout
- Given：A3 mock 返回 timeout
- When：handler.Handle
- Then：返回 RetryableError；ai_task_runs status='timeout'
- 覆盖：R5

#### 4.10 TestGenerateHandler_ParseEmpty
- Given：A3 返回 {aiAnalyses:[], riskItems:[]}
- When：handler.Handle
- Then：返回 RetryableError；ai_task_runs status='failed' + validation_status='invalid' + B1 `AI_OUTPUT_INVALID`
- 覆盖：R5

#### 4.11 TestGenerateHandler_PermanentFailAt5Attempts
- 文件：`backend/internal/targetjob/drainer_integration_test.go`（或新建 `debrief_runner_integration_test.go`）
- Given：async_jobs(debrief_generate, attempts=4) 行存在；mock handler 持续返回 RetryableError
- When：drainer lease + handle 第 5 次失败
- Then：async_jobs.status='failed'（permanent）+ locked_at=null；debriefs.status='draft' 保持；outbox debrief.completed 不发出
- 覆盖：R6

### Phase 5: getDebrief read handler

#### 5.1 TestStoreGetDebrief_DraftPartial
- 文件：`backend/internal/store/debrief/store_test.go`
- Given：debriefs(status='draft') 行存在
- When：GetDebrief(user_id, id)
- Then：返回 Debrief{status:'draft', questions:[原始 raw_questions, aiAnalysis:null], riskItems:[], nextRoundChecklist:[], thankYouDraft:null, provenance:null}
- 覆盖：R2, R15, R18

#### 5.2 TestStoreGetDebrief_CompletedFull
- Given：debriefs(status='completed', risk_items 非空) 行存在
- When：GetDebrief
- Then：返回 Debrief 全字段；questions[*].aiAnalysis 注入；riskItems 填充；provenance 6 字段
- 覆盖：R2, R15

#### 5.3 TestStoreGetDebrief_CrossUserNotFound
- Given：debriefs.user_id = A，请求方 user_id = B
- When：GetDebrief(B, debriefId)
- Then：返回 ErrNotFound
- 覆盖：R7

#### 5.4 TestServiceGetDebrief_ProvenanceWireOnly
- 文件：`backend/internal/debrief/service_test.go`
- Given：debriefs(status='completed') 行
- When：service.GetDebrief
- Then：返回的 Debrief.provenance keys = {promptVersion, rubricVersion, modelId, language, featureFlag, dataSourceVersion} 6 个；不含 feature_key / model_profile_name / provider / cost_micros / latency_ms 运行时字段
- 覆盖：R15

#### 5.5 TestGetDebrief_DraftResponse
- 文件：`backend/internal/api/debriefs/handler_test.go`
- Given：debriefs(status='draft')
- When：GET /debriefs/{id}
- Then：返回 200 + Debrief schema with status='draft' + 空字段
- 覆盖：R2, R18

#### 5.6 TestGetDebrief_CompletedResponse
- Given：debriefs(status='completed')
- When：GET /debriefs/{id}
- Then：返回 200 + Debrief 全字段
- 覆盖：R2

#### 5.7 TestGetDebrief_CrossUser404
- Given：跨用户访问
- When：GET /debriefs/{id}
- Then：返回 404 + B1 `DEBRIEF_NOT_FOUND`
- 覆盖：R7

#### 5.8 TestGetDebrief_NotFound404
- Given：debriefId 不存在
- When：GET /debriefs/{id}
- Then：返回 404 + B1 `DEBRIEF_NOT_FOUND`
- 覆盖：R7

### Phase 6: 隐私 / 观测 / retry / legacy negative

#### 6.1 TestOutboxPayload_NoRawText
- 文件：`backend/internal/store/debrief/store_test.go`
- Given：createDebrief + worker complete 完整流程
- When：读 outbox debrief.created + debrief.completed
- Then：payload 序列化字符串不含 questionText / myAnswerSummary / interviewerReaction / notes / risk_items prose 任一子串（mock 时用特定 marker string `__SECRET_RAW_TEXT__` 验证）
- 覆盖：R11, R14

#### 6.2 TestAuditEvents_NoRawText
- Given：createDebrief + worker complete + suggestDebriefQuestions 完整流程
- When：读 audit_events metadata
- Then：metadata 序列化字符串不含 raw text；keys 集合 ⊆ {debrief_id, target_job_id, status, language, error_code, suggestion_count, question_count}
- 覆盖：R14

#### 6.3 TestAITaskRunsWritten
- Given：完整流程（createDebrief → worker → suggestDebriefQuestions）
- When：查询 ai_task_runs
- Then：含两类 task_type 各至少 1 行；行字段 feature_key / model_profile_name / status / input_tokens / output_tokens / latency_ms / validation_status / error_code 完整
- 覆盖：R13

#### 6.4 TestAuditEventsWritten
- Given：完整流程
- When：查询 audit_events
- Then：含三种 action: create_debrief / debrief_completed / suggest_debrief_questions（命名以实际实现为准）
- 覆盖：R13

#### 6.5 TestRetryPolicy_BackoffBelowMax
- Given：async_jobs(debrief_generate, attempts=2) 失败
- When：drainer finalize
- Then：attempts=3, available_at=now()+backoff(3), status='queued', locked_at=null
- 覆盖：R6

#### 6.6 TestRetryPolicy_PermanentFailAtMax
- Given：async_jobs(attempts=4) 失败
- When：drainer finalize
- Then：attempts=5, status='failed' (permanent), locked_at=null
- 覆盖：R6

### Phase 8: D-20 resumeId suggestion context gate

#### 8.1 TestStoreGetSuggestionContext_LoadsResumeStructuredProfile
- 文件：`backend/internal/store/debrief/store_test.go`
- Given：SuggestionContextRequest 带 `resumeId`
- When：Repository.GetSuggestionContext 按 `(user_id, target_job_id)` 拉 target job 后继续按 `(user_id, resume_id)` 查询 `resumes.structured_profile`
- Then：SuggestionContext.ResumeSummary 包含扁平 resume structured_profile JSON
- 覆盖：R23

#### 8.2 TestStoreGetSuggestionContext_CrossUserResumeNotFound
- Given：target job 属于当前用户，但 `resumeId` 不存在或不属于当前用户
- When：Repository.GetSuggestionContext
- Then：返回 ErrDebriefPrerequisite，避免 cross-user resume 进入 prompt context
- 覆盖：R7, R23

#### 8.3 TestServiceSuggestQuestions_ResumeContextInPrompt
- 文件：`backend/internal/debrief/service_test.go`
- Given：SuggestionContext.ResumeSummary 已由 store 提供
- When：Service.SuggestQuestions 渲染 F3 prompt template
- Then：AI Complete payload 的 user message 包含 resume structured_profile 摘要，且 response 可返回 `source='resume'`
- 覆盖：R3, R23

#### 8.4 TestSuggestDebriefQuestions_MapsResumeIDToService
- 文件：`backend/internal/api/debriefs/handler_test.go`
- Given：generated `SuggestDebriefQuestionsRequest` body 使用 `resumeId`
- When：Handler.SuggestDebriefQuestions
- Then：domain `SuggestQuestionsRequest.ResumeID` 等于 request `resumeId`
- 覆盖：R3, R23

#### 8.5 TestBuildAPIHandlerMountsDebriefSuggestQuestionsBehindSessionMiddleware
- 文件：`backend/cmd/api/main_test.go`
- Given：cmd/api route table 挂载 debrief handler
- When：未认证请求 `POST /api/v1/debriefs/question-suggestions`，body 含 `resumeId`
- Then：返回 `AUTH_UNAUTHORIZED`，证明真实 route 被 session middleware 包裹并由 cmd/api 挂载
- 覆盖：R3, R23

#### 8.6 E2E.P0.063 scenario wrapper
- 文件：`test/scenarios/e2e/p0-063-debrief-suggest-questions/scripts/trigger.sh` 与 `verify.sh`
- Given：fixture `openapi/fixtures/Debriefs/suggestDebriefQuestions.json` default request 使用 `resumeId`
- When：运行 setup -> trigger -> verify -> cleanup
- Then：执行 store/service/API/cmd-api focused tests、`make validate-fixtures`、fixture `resumeId` marker 与 `resumeVersionId` 负向 gate
- 覆盖：R3, R12, R23

## 3 集成测试与覆盖率说明

- 覆盖率以 plan 列出的测试项达成度衡量；不引入 raw coverage 百分比作为 hard gate（与 §4 设计约束一致）。
- 集成测试 `TestGenerateHandler_PermanentFailAt5Attempts` 使用 testcontainer / dev-stack postgres 验证 drainer 真实并发行为。
- E2E scenario（P0.060-064）覆盖见 [bdd-plan.md](./bdd-plan.md) 与 [bdd-checklist.md](./bdd-checklist.md)。
