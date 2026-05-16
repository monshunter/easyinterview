# 001 Debrief Record and Analysis

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-16

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 Test Plan**: [test-plan](./test-plan.md)

## 1 目标

落地 backend-debrief P0 闭环的全部 backend 实现：3 个 API operation (`createDebrief` / `getDebrief` / `suggestDebriefQuestions`) + `debrief_generate` 异步 worker handler（通过 backend-targetjob drainer 注册）+ AI 调用编排 + 单向状态机 (`draft → completed`) + outbox `debrief.created` / `debrief.completed` + cross-user 隔离 + 隐私红线 + 观测红线 + cross-owner pre-launch addendum 收口（B1/B2/B3/B4/F3）+ legacy negative gate。

落地完成后，前端 `frontend-debrief/001` 可基于本 plan 暴露的 schema + fixture + scenario asset 完成 DebriefScreen 真实数据接入。

## 2 背景

`backend-debrief` spec v1.0 在 2026-05-16 由本 plan 同时段派生；spec §7 已声明 plan 001 落地全部 D-1~D-18 决策。spec §1 已确认 B2 OpenAPI 已冻结 2 个既有 operation（`createDebrief` / `getDebrief`），但需要通过 Phase 0 跨 owner pre-launch addendum 完成：

- **B1 addendum**：新增 `DEBRIEF_NOT_FOUND` 错误码（D-15 前置）+ 通用 `IDEMPOTENCY_KEY_MISMATCH` 错误码 + `DebriefRoundType` enum（替代 `shared/events.yaml` 中错误的 `b1.InterviewerRole` 引用）+ `DebriefQuestionSource` enum（覆盖 `jd` / `resume` / `mock_report` / `manual`，用于 `suggestDebriefQuestions` 响应）+ 同步 generated Go/TS 字面量与 fixture parity；AI 失败映射只使用当前 B1 canonical `AI_*` code。
- **B2 addendum**：新增 `POST /debriefs/question-suggestions` `suggestDebriefQuestions` operation + `SuggestDebriefQuestionsRequest` / `SuggestDebriefQuestionsResponse` schema + fixtures `Debriefs/suggestDebriefQuestions.json`；同步修复 `Debrief.roundType` / `CreateDebriefRequest.roundType` 字段引用（若 B1 owner 选择独立 enum `DebriefRoundType`）；扩展既有 fixtures `Debriefs/createDebrief.json` + `Debriefs/getDebrief.json`（含 `default` + `debrief-draft` + `prototype-baseline` variants）。
- **B3 addendum**：修复 `shared/events.yaml` `debrief.created.roundType: $ref:b1.InterviewerRole` → `$ref:b1.DebriefRoundType`；同步 `shared/events/baseline/events.v1.json` + `shared/events/schemas/debrief.created.v1.json`；`make lint-events` + `make codegen-events-check` 通过。
- **B4 addendum**：新增 `ai_task_runs.task_type='debrief_suggest_questions'` CHECK / enum-source 字面量；同步 `migrations/000001_create_baseline.up.sql`、`migrations/enum-sources.yaml`、migration lint / replay tests；不新增 B4 列。
- **F3 addendum**：新增 `debrief.suggest_questions` feature_key + `debrief.suggest_questions.default` model profile + 基线 prompt/rubric v0.1.0；同步 `config/prompts/debrief.suggest_questions/` / `config/rubrics/debrief.suggest_questions/` 目录与 `seed_baseline_prompt_rubric_versions` migration（如有）。
- **backend-practice 验证**：Q-3 spec 已要求验证 backend-practice 现状是否支持 `goal='debrief'` plan 派生 + 合法 `mode IN ('assisted','strict')` session start；如未支持，需 backend-practice owner 同步 addendum，否则 frontend-debrief step 2 "复盘面试" handoff 无法闭环。

backend-targetjob drainer 已在 `backend/internal/targetjob/drainer.go` 注册 `debrief_generate` job_type CHECK，但当前实现假设：drainer 通过 `JobHandler` interface 路由不同 job_type 到对应处理器；本 plan 在 cmd/api bootstrap 注册 `debrief.GenerateHandler` 完成 wiring。

## 3 质量门禁分类

- **Plan 类型**: contract + code-internal + feature-behavior（混合：跨 owner 契约 addendum + backend domain 代码实现 + 用户可见 API 行为）
- **TDD 策略**: Code plan requires TDD。所有 handler / service / store / worker handler 必须先写测试（红/绿/重构）；测试文件：`backend/internal/debrief/*_test.go` + `backend/internal/api/debriefs/*_test.go` + `backend/internal/store/debrief/*_test.go`；测试命令：`cd backend && go test ./internal/debrief ./internal/api/debriefs ./internal/store/debrief -count=1`；Phase 1-6 每个 checklist item 必须命名其测试断言来源（见 test-plan.md 与 test-checklist.md）。
- **BDD 策略**: Feature plan requires BDD。本 plan 引入 3 个用户可见 API operation + 1 个异步 worker 触发的 outbox / DB state 变化，前端 frontend-debrief 直接消费；BDD scenarios `E2E.P0.060-064` 已在 [bdd-plan.md](./bdd-plan.md) 分配，主 [checklist.md](./checklist.md) 在 Phase 6 含 `BDD-Gate:` 项引用每个 scenario ID；执行必须使用当前场景框架的 `scripts/setup.sh` → `scripts/trigger.sh` → `scripts/verify.sh` → `scripts/cleanup.sh` 四段入口，cleanup 在失败时也必须执行。
- **替代验证 gate**:
  - Phase 0 cross-owner addendum：`make codegen-check` / `make validate-fixtures` / `make lint-events` / `make codegen-events-check` / `migrations/lint.sh` / 单元测试 `python3 -m pytest scripts/lint -q`
  - Phase 0 backend-practice 验证：`cd backend && go test ./internal/practice -run "TestPracticePlan.*Debrief|TestPracticeSession.*Debrief" -count=1`（如无，则补 backend-practice owner skill addendum）
  - Privacy + legacy negative gate：`grep -rn "mistakes_count\|generatedMistakeCount\|experience_library\|drill_builder\|growth_center" backend/internal/debrief shared/events.yaml shared/jobs.yaml docs/spec/backend-debrief/ openapi/fixtures/Debriefs/ test/scenarios/e2e/p0-06*` 命中即失败

## 4 实施步骤

### Phase 0: Cross-owner pre-launch addendums + 前置验证

#### 0.1 B1 addendum：新增错误码与 enum

在 `shared/conventions.yaml` 新增 `DEBRIEF_NOT_FOUND` 错误码 + 通用 `IDEMPOTENCY_KEY_MISMATCH` 错误码 + `DebriefRoundType` enum（values: hr_screen / hiring_manager / behavioral / technical / culture / custom）+ `DebriefQuestionSource` enum（values: jd / resume / mock_report / manual）；不新增任何未登记 AI 旧别名，后续失败映射必须使用当前 B1 canonical `AI_PROVIDER_CONFIG_INVALID` / `AI_PROVIDER_SECRET_MISSING` / `AI_PROVIDER_TIMEOUT` / `AI_OUTPUT_INVALID` / `AI_FALLBACK_EXHAUSTED`；运行 `make codegen-check` 同步 generated Go (`shared/go/conventions/v1/`) + TS (`shared/ts/conventions/v1/`) + fixture parity (`shared/fixtures/conventions-parity.json`)；提交 commit `feat(shared-conventions): add debrief enums and error code`.

#### 0.2 B2 addendum：新增 suggestDebriefQuestions operation + 修复 roundType 引用

在 `openapi/openapi.yaml` 新增：
- `POST /debriefs/question-suggestions` operation `suggestDebriefQuestions`：request body `SuggestDebriefQuestionsRequest{targetJobId(uuid required), sessionId(uuid optional), resumeVersionId(uuid optional), language(string required), count(int default 6 max 10)}`；response 200 `SuggestDebriefQuestionsResponse{suggestions: array of SuggestedDebriefQuestion{stage(optional), questionText(string required), whyLikelyAsked(string required), source($ref DebriefQuestionSource required)}}`。
- 修复 `Debrief.roundType` / `CreateDebriefRequest.roundType` 字段引用为 `$ref: '#/components/schemas/DebriefRoundType'`（替代原 inline enum 数组），保持 wire 字面量不变（仍是 6 个 hr_screen/hiring_manager/behavioral/technical/culture/custom）。

扩展 / 新增 fixtures：
- 扩展既有 `openapi/fixtures/Debriefs/createDebrief.json`（`default` = 202 + DebriefWithJob + idempotency-key example）
- 扩展既有 `openapi/fixtures/Debriefs/getDebrief.json`（`default` = completed full / `debrief-draft` = draft + 空字段 / `prototype-baseline` = 中文示例）
- 新增 `openapi/fixtures/Debriefs/suggestDebriefQuestions.json`（`default` = 6 suggestions / `empty` = 0 / `prototype-baseline`）

运行 `make codegen-check` + `make validate-fixtures`；提交 commit `feat(openapi): add suggestDebriefQuestions operation and align debrief round type`.

#### 0.3 B3 addendum：修复 events.yaml roundType 引用

修改 `shared/events.yaml` 行 201：
```diff
-      roundType: { type: $ref:b1.InterviewerRole, source: spec:3.1.4 }
+      roundType: { type: $ref:b1.DebriefRoundType, source: spec:3.1.4 }
```

运行 `make codegen-events-check` + `make lint-events`；同步 `shared/events/baseline/events.v1.json` + `shared/events/schemas/debrief.created.v1.json` + `shared/events/__fixtures__/envelopes.json`（如 envelope fixture 引用旧 enum）；提交 commit `fix(events): align debrief.created roundType reference`.

#### 0.4 B4 addendum：新增 ai_task_runs debrief_suggest_questions task_type

在 db-migrations-baseline 真理源新增 `ai_task_runs.task_type='debrief_suggest_questions'` 字面量：
- 修改 `migrations/000001_create_baseline.up.sql` 中 `ai_task_runs.task_type` CHECK allowlist，加入 `debrief_suggest_questions`
- 修改 `migrations/enum-sources.yaml` 对应 enum-source，保证 lint / drift check 可追溯
- 补 migration lint / replay 断言：`debrief_generate` 与 `debrief_suggest_questions` 均允许写入，未登记 task_type 仍被 CHECK 拒绝
- 不新增列、不修改 `debriefs` 表结构

运行 `migrations/lint.sh` + `make migrate-check`；提交 commit `fix(migrations): allow debrief suggestion ai task type`.

#### 0.5 F3 addendum：新增 debrief.suggest_questions feature_key

在 prompt-rubric-registry 真理源新增 feature_key `debrief.suggest_questions` + profile `debrief.suggest_questions.default` + 基线 prompt v0.1.0：
- Prompt 输入变量：`{{targetJobTitle}}` / `{{jdHighlights}}` / `{{resumeBullets?}}` / `{{practiceSessionSummary?}}` / `{{language}}` / `{{count}}`
- Prompt 输出 schema：`{suggestions: [{stage?, questionText, whyLikelyAsked, source}]}`（与 B2 SuggestDebriefQuestionsResponse 对齐）
- 路径：`config/prompts/debrief.suggest_questions/v0.1.0*.{yaml,md}` + `config/rubrics/debrief.suggest_questions/v0.1.0*.yaml`
- 同步 `migrations/000002_seed_baseline_prompt_rubric_versions.up.sql` 或新增独立 migration（由 F3 owner 决定）

提交 commit `feat(prompt-rubric-registry): seed debrief.suggest_questions baseline`.

#### 0.6 backend-practice 现状验证（Q-3）

运行 `grep -rn "goal.*debrief\|mode.*debrief" backend/internal/practice` 找出当前 plan 派生 / session start handler 是否分支处理 `goal='debrief'`，并确认 `mode` 仍为 backend-practice D-5 收敛后的 `assisted|strict`：
- 如果已支持：在 plan history 记录验证证据（grep 输出 + 关联测试名）。
- 如果未支持：暂停 plan 001 实施，回到 backend-practice owner 同步 addendum（`practice_plans.goal='debrief'` 派生默认逻辑 + `PracticeMode IN ('assisted','strict')` session start 默认 handler）；恢复 plan 001 后记录依赖 commit。

#### 0.7 整体 Phase 0 收口

运行所有 quality gates：
- `cd backend && go test ./... -count=1`
- `make codegen-check` + `make validate-fixtures` + `make lint-events` + `make codegen-events-check`
- `migrations/lint.sh` + `make migrate-check`（dev-stack .env）
- `python3 -m pytest scripts/lint -q`
- `make docs-check` + `git diff --check`

未通过任何 gate 时 BLOCK Phase 1 实施；提交单一 commit `chore(backend-debrief): close phase 0 cross-owner addendums and validation`.

### Phase 1: createDebrief handler 与 store 骨架

#### 1.1 包结构与 store 接口

在 `backend/internal/debrief/` 新建包：`service.go` / `service_test.go` / `errors.go`；在 `backend/internal/store/debrief/` 新建包：`store.go` / `store_test.go`；在 `backend/internal/api/debriefs/` 新建：`handler.go` / `handler_test.go`。

定义 store 接口：
```go
type Store interface {
    CreateDebrief(ctx, params CreateDebriefParams) (*Debrief, *AsyncJob, error)
    GetDebrief(ctx, userID, debriefID uuid.UUID) (*Debrief, error)
    UpdateDebriefCompleted(ctx, params UpdateDebriefCompletedParams) error
}
```

#### 1.2 createDebrief handler skeleton

实现 `CreateDebrief` handler：
- 注入 `user_id` from auth middleware（B7 backend-auth）
- 解析 generated `CreateDebriefRequest` types
- 校验 `questions.length >= 1`、单题 `questionText.length <= 4000`、`myAnswerSummary.length <= 4000`、`interviewerReaction.length <= 1000`、`notes.length <= 10000`；失败返回 `422 VALIDATION_FAILED`
- 调用 service 层 `CreateDebrief`（暂返回 stub）
- 返回 generated `DebriefWithJob` 202

#### 1.3 idempotency middleware 接线

确认 `Idempotency-Key` 中间件已对 `POST /debriefs` 生效（复用 backend-practice 既有 IK 实现）：
- 同 IK + 同 user_id + 相同 request body hash → 返回 cached 202 + 相同 debriefId + 相同 jobId
- 同 IK + 不同 body hash → 返回 409 `IDEMPOTENCY_KEY_MISMATCH`
- 不同 IK → 正常处理

如 backend-practice IK middleware 不支持自定义 resource type，本 phase 必须扩展中间件支持 `resource_type='debrief'`。

### Phase 2: createDebrief 完整事务 + outbox

#### 2.1 store.CreateDebrief 实现

单一 SQL transaction 内：
- `INSERT INTO debriefs (id, user_id, target_job_id, status, round_type, interviewer_role, language, raw_questions, notes) VALUES (...)` (status='draft', id 为 generated uuidv7)
- `INSERT INTO async_jobs (id, job_type, resource_type, resource_id, status, dedupe_key, payload, available_at) VALUES ('debrief_generate', 'debrief', debrief.id, 'queued', debrief.id::text, jsonb_payload, now())`
- `INSERT INTO outbox_events (id, event_name, aggregate_type, aggregate_id, payload, ...) VALUES ('debrief.created', 'debrief', debrief.id, {debriefId, targetJobId, roundType, questionCount}, ...)`
- 失败任一回滚整个事务；返回 `(debrief, async_job, nil)` 给 service 层

#### 2.2 service 层组装 + audit

在 `service.CreateDebrief` 内：
- 调用 store
- 写一行 `audit_events(action='create_debrief', resource_type='debrief', resource_id=debrief.id, user_id, metadata={target_job_id, language, question_count})` 在事务外（best-effort）
- 返回给 handler

#### 2.3 handler 串联

handler 调用 service，处理 IK middleware 返回，返回 generated `DebriefWithJob{debriefId, job:{...}}`。

#### 2.4 fixture parity 验证

`make validate-fixtures` 确认 `Debriefs/createDebrief.json` 与 handler 真实响应一致；`cd backend && go test ./internal/api/debriefs -run "TestCreateDebrief_Happy|TestCreateDebrief_IdempotencyReplay|TestCreateDebrief_IdempotencyMismatch|TestCreateDebrief_ValidationError" -count=1` 通过。

### Phase 3: suggestDebriefQuestions sync handler

#### 3.1 service.SuggestQuestions 实现

在 `backend/internal/debrief/service.go` 新增 `SuggestQuestions(ctx, params SuggestParams) (*SuggestResult, error)`：
- 拉 target_job 摘要（generated `getTargetJob(targetJobId, user_id)`），越权返回 `403 FORBIDDEN`
- 可选拉 session 摘要 + resume version 摘要
- 组装 F3 prompt 上下文：`{targetJobTitle, jdHighlights, resumeBullets?, practiceSessionSummary?, language, count}`
- 调 `RegistryClient.Resolve("debrief.suggest_questions", language)` → A3 `AIClient.Complete`
- 解析 AI 输出 JSON 到 `SuggestResult{suggestions: []SuggestedQuestion{questionText, whyLikelyAsked, source(B1 enum)}}`
- AI 失败：F3 resolve/config 失败返回 B1 `AI_PROVIDER_CONFIG_INVALID`；A3 secret missing 返回 `AI_PROVIDER_SECRET_MISSING`；A3 timeout 返回 `AI_PROVIDER_TIMEOUT`；fallback exhausted / provider unreachable 返回 `AI_FALLBACK_EXHAUSTED`；invalid JSON / parsed empty 返回 `AI_OUTPUT_INVALID`
- 成功：写 `ai_task_runs(task_type='debrief_suggest_questions', status='success', feature_key, model_profile_name, input_tokens, output_tokens, latency_ms)` + audit 一行
- 失败：写 `ai_task_runs(status='failed', error_code, validation_status='invalid' 或 nil)` + audit 一行 with error_code

#### 3.2 handler skeleton

实现 `SuggestDebriefQuestions` handler：
- 注入 user_id
- 解析 generated `SuggestDebriefQuestionsRequest`
- 校验 `count` 边界 (1-10)
- 调 service.SuggestQuestions
- 返回 generated `SuggestDebriefQuestionsResponse` 200 或 5xx with B1 error_code
- 不要求 IK

#### 3.3 fixture parity + 失败映射

`make validate-fixtures` 确认 `suggestDebriefQuestions.json` variants 一致；`TestSuggestDebriefQuestions_Happy|TestSuggestDebriefQuestions_AIFailure|TestSuggestDebriefQuestions_CrossUserTargetJob|TestSuggestDebriefQuestions_CountBoundary` 通过。

### Phase 4: debrief_generate worker handler

#### 4.1 GenerateHandler 实现 targetjob.JobHandler interface

在 `backend/internal/debrief/generate_handler.go` 实现：
```go
type GenerateHandler struct {
    store        Store
    aiClient     aiclient.AIClient
    registry     promptregistry.Client
    auditWriter  audit.Writer
    aiTaskWriter aiclient.AITaskRunWriter
}

func (h *GenerateHandler) Handle(ctx context.Context, job targetjob.Job) error {
    // 1. 解析 job.payload 到 GenerateJobPayload{debriefId, targetJobId, language, questionCount}
    // 2. 拉 debrief.raw_questions 全文 + target_job 摘要 + 可选 resume_version
    // 3. Resolve("debrief.generate", language) -> AIClient.Complete
    // 4. 解析 AI 输出: aiAnalyses[] + risk_items[]
    // 5. UpdateDebriefCompleted (status=completed, raw_questions in-place patch, risk_items, prompt_version, rubric_version, model_id, provider) + outbox emit (debrief.completed) in single transaction
    // 6. AI failure: return targetjob.RetryableError or PermanentError as appropriate; drainer handles retry/permanent fail
    // 7. ai_task_writer.Write success/failed row (decorator-based by AIClient already)
}
```

#### 4.2 cmd/api bootstrap 注册

在 `backend/cmd/api/main.go` 或等价 bootstrap 注册 `debrief.GenerateHandler` 到既有 `targetjob.Drainer`：
```go
drainer.RegisterHandler("debrief_generate", debrief.NewGenerateHandler(...))
```

#### 4.3 store.UpdateDebriefCompleted 实现

单一 SQL transaction：
- `UPDATE debriefs SET status='completed', raw_questions=$1::jsonb, risk_items=$2::jsonb, prompt_version=$3, rubric_version=$4, model_id=$5, provider=$6, updated_at=now() WHERE id=$7 AND status='draft'`（CAS）
- `INSERT INTO outbox_events (event_name='debrief.completed', payload={debriefId,targetJobId,riskItemCount,practiceFocusCount}, ...)`
- 失败任一回滚

#### 4.4 outbox payload 校验

确保 `debrief.completed.payload` 严格按 D-13 4 字段（不含 notes / question text / answer text）；写一个 unit test 反序列化 payload 后 assert allowed keys 集合。

#### 4.5 AI failure semantics

- F3 ResolveActive 失败 / A3 timeout / A3 invalid JSON / parsed empty → handler 返回 `targetjob.RetryableError`；drainer 自动 attempts+1 + backoff
- attempts >= 5 时 drainer 自动置 `async_jobs.status='failed'`；`debriefs.status` 保持 `'draft'`
- 不发 `debrief.completed` outbox
- `ai_task_runs` 写 failed row

Tests: `TestGenerateHandler_Happy|TestGenerateHandler_F3ResolveFailed|TestGenerateHandler_A3Timeout|TestGenerateHandler_ParseEmpty|TestGenerateHandler_PermanentFail`.

### Phase 5: getDebrief read handler

#### 5.1 handler.GetDebrief

实现 `GetDebrief` handler：
- 注入 user_id
- 解析 path param debriefId
- 调 service.GetDebrief(user_id, debriefId)
- 返回 generated `Debrief` schema 200

#### 5.2 store.GetDebrief

`SELECT ... FROM debriefs WHERE id=$1 AND user_id=$2` → map to `Debrief` schema：
- `status='draft'`：`questions=[{questionText,myAnswerSummary,interviewerReaction,aiAnalysis:null}]`、`riskItems=[]`、`nextRoundChecklist=[]`（D-7 P0 留空）、`thankYouDraft=null`、`provenance=null`
- `status='completed'`：`questions[*].aiAnalysis` 注入；`riskItems` 填充；`provenance` 6 字段填充
- 未找到 / cross-user → ErrNotFound → handler 返回 `404 DEBRIEF_NOT_FOUND`

#### 5.3 service.GetDebrief

业务层 hydrate Provenance 6 字段（其中 `featureFlag` / `dataSourceVersion` 在 D-11 中决定走 jsonb 子键 / audit-only / hard-coded constant；本 plan 选择：`featureFlag` = `null`（P0 无 feature flag 区分）；`dataSourceVersion` = `'debrief/<debriefId>@v1'` 字面量）。

#### 5.4 fixture parity + cross-user 隔离

Tests: `TestGetDebrief_DraftPartialReturn|TestGetDebrief_CompletedFullReturn|TestGetDebrief_CrossUser404|TestGetDebrief_ProvenanceWireOnly`.

### Phase 6: 隐私 / 观测红线 + retry / legacy negative + BDD gate

#### 6.1 隐私红线断言

- 单元测试断言 `debrief.created` / `debrief.completed` outbox payload 不含 `questionText` / `myAnswerSummary` / `interviewerReaction` / `notes` 子串
- 单元测试断言 audit_events metadata 不含上述字段
- grep gate：`grep -rn "questionText\|myAnswerSummary" backend/internal/debrief/service.go shared/events.yaml shared/events/baseline/events.v1.json | grep -v "_test.go"` 不命中
- 单元测试断言 F1 metric label 不含 raw text

#### 6.2 观测红线

- 单元测试断言 `ai_task_runs` 写入两个 task_type (`debrief_generate` / `debrief_suggest_questions`) 的正确字段
- 断言 audit_events 记录 createDebrief / worker complete / suggestDebriefQuestions 三种 action
- F1 metric 字典登记（如有新 metric）由 F1 owner co-author

#### 6.3 Retry policy + permanent fail

- 单元测试断言 attempts < 5 时 async_jobs.status='queued' + available_at>now()
- 单元测试断言 attempts >= 5 时 async_jobs.status='failed' + debriefs.status 保持 'draft'
- 集成测试断言 drainer lease + retry 行为

#### 6.4 Legacy negative gate

- grep `mistakes_count` / `generatedMistakeCount` / `experience_library` / `drill_builder` / `growth_center` / `star_editor` / `debrief_voice` 在 `backend/internal/debrief` / `shared/events.yaml` / `shared/jobs.yaml` / `docs/spec/backend-debrief/` / `openapi/fixtures/Debriefs/` 不命中
- scenario verify.sh 在每个 P0.060-064 scenario 中含 grep 反查
- `backend_debrief_legacy.py --phase all` lint script（可复用 backend_practice_legacy.py 模式）

#### 6.5 BDD gate 收口

执行所有 P0.060-064 scenario；每个目录必须按顺序执行 `scripts/setup.sh`、`scripts/trigger.sh`、`scripts/verify.sh`，并在成功或失败后执行 `scripts/cleanup.sh`：
- `test/scenarios/e2e/p0-060-debrief-create-and-generate/`
- `test/scenarios/e2e/p0-061-debrief-get-and-cross-user/`
- `test/scenarios/e2e/p0-062-debrief-worker-failure-and-retry/`
- `test/scenarios/e2e/p0-063-suggest-debrief-questions/`
- `test/scenarios/e2e/p0-064-debrief-privacy-and-legacy-negative/`

所有 scenario 必须 `--- PASS` + `ok` + verify.sh 通过。

#### 6.6 Plan 收口

- `cd backend && go test ./... -count=1` 通过
- `make codegen-check` / `make validate-fixtures` / `make lint-events` / `make codegen-events-check` / `make migrate-check`（dev-stack .env）/ `make docs-check` / `git diff --check` 通过
- 更新 plans/INDEX.md 把 001 移到 completed
- 更新 backend-debrief/history.md 增加 1.1 completion 行

## 5 验收标准

- C-1 ~ C-17（[spec §6](../../spec.md#6-验收标准)）全部通过
- 本 plan 列出的 Phase 0-6 实现项全部按 checklist 勾选
- [BDD-Gate](./bdd-checklist.md) `E2E.P0.060-064` 全部通过
- [test-checklist](./test-checklist.md) 单元测试与集成测试项全部通过
- Phase 0 跨 owner addendum 全部在主分支落地（B1/B2/B3/B4/F3 + backend-practice 验证）
- legacy negative gate 通过

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| backend-practice 现状未支持 `goal='debrief'` 与合法 `mode IN ('assisted','strict')` session start，导致 frontend-debrief step 2 复盘面试 handoff 闭环失败 | Phase 0 Q-3 验证发现未支持时，立即暂停 plan 001，回 backend-practice owner 同步 addendum（future `004-derived-plans-debrief` 或等价原地修订）；恢复后记录依赖 commit |
| B2 owner 对 `POST /debriefs/question-suggestions` 路径或 schema 命名持不同意见 | Phase 0 与 B2 owner co-design：默认建议是 collection-action `POST /debriefs/question-suggestions`，备选 `POST /debriefs/_suggestions`；schema 命名 `SuggestDebriefQuestionsRequest`/`Response`；最终命名以 B2 owner 决定为准，本 plan 跟随更新 |
| F3 owner 决定 `debrief.suggest_questions` 必须含 rubric（Q-2），增加 Phase 0 工作量 | Phase 0 与 F3 owner co-design rubric schema（如 6 条 suggestion 必须覆盖 stage 多样性）；如复杂度过高，降级到 prompt-only baseline，rubric 留 plan 003+ |
| AI 输出 JSON 解析失败率高（worker 路径触发频繁 permanent fail） | Phase 4 单元测试覆盖 5 种 parse failure pattern；如生产观测发现 parse 失败率 > 5%，触发 F3 owner 修订 prompt strict mode 或加 structured output validator |
| suggestDebriefQuestions 被滥用（前端死循环 / 用户疯狂点击） | Q-5 默认 P0 不做 rate limit；如生产观测发现滥用，plan 002 增加 rate limit middleware（如每用户每分钟 5 次）+ Quota |
| `Debrief.provenance` 的 `featureFlag` / `dataSourceVersion` 持久化锚点不清晰（B4 无对应列）| D-11 决策行 + Phase 5 实现明确：`featureFlag=null`（P0 无 feature flag 区分），`dataSourceVersion='debrief/<debriefId>@v1'` 字面量；如未来引入 feature flag 区分 prompt 版本，再决定走 jsonb 子键还是新增 B4 列 |
| 跨 owner addendum 数量多（B1/B2/B3/B4/F3 + backend-practice 验证），Phase 0 工作量超预期 | 接受 Phase 0 较重的现实；按 0.1-0.7 顺序串行执行；每个 sub-phase 独立 commit 便于回滚；如某 owner addendum 卡住，暂停 plan 001 直到 unblock |
