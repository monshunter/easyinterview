# 001 Debrief Record and Analysis Checklist

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-05-16

**关联计划**: [plan](./plan.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 Test Checklist**: [test-checklist](./test-checklist.md)

## Phase 0: Cross-owner pre-launch addendums + 前置验证

- [x] 0.1 B1 addendum：`shared/conventions.yaml` 新增 `DEBRIEF_NOT_FOUND` 错误码 + 通用 `IDEMPOTENCY_KEY_MISMATCH` 错误码 + `DebriefRoundType` enum + `DebriefQuestionSource` enum；确认不新增未登记 AI 旧别名；`make codegen-check` 同步 generated Go + TS + fixture parity 通过；commit `feat(shared-conventions): add debrief enums and error code` 落地（验证：`grep -rn "DEBRIEF_NOT_FOUND\|IDEMPOTENCY_KEY_MISMATCH\|DebriefRoundType\|DebriefQuestionSource" backend/internal/shared frontend/src/lib/conventions shared/fixtures openapi/openapi.yaml backend/internal/api/generated frontend/src/api/generated` 命中所有 4 个标识符）
  <!-- verified: 2026-05-16 red-green; commands: go test ./cmd/codegen/conventions focused, python3 scripts/lint/conventions_yaml.py, python3 scripts/lint/error_codes.py, python3 scripts/lint/conventions_drift.py, go test ./internal/shared/types ./internal/shared/errors focused, pnpm --dir frontend test src/lib/conventions/conventions-parity.test.ts, make lint-openapi -->
- [x] 0.2 B2 addendum：`openapi/openapi.yaml` 新增 `POST /debriefs/question-suggestions` operation + `SuggestDebriefQuestionsRequest`/`Response` schema + 修复 `Debrief.roundType` / `CreateDebriefRequest.roundType` 引用 `$ref DebriefRoundType`；扩展既有 fixtures `Debriefs/createDebrief.json` / `Debriefs/getDebrief.json`，新增 `Debriefs/suggestDebriefQuestions.json`；`make codegen-check` + `make validate-fixtures` 通过；commit `feat(openapi): add suggestDebriefQuestions operation and align debrief round type` 落地（验证：generated `frontend/src/api/generated/` 含 `suggestDebriefQuestions`；generated `backend/internal/api/generated/` 含 `SuggestDebriefQuestions` handler interface）
  <!-- verified: 2026-05-16 red-green; commands: make codegen-openapi, make lint-openapi, make validate-fixtures, python3 -m pytest scripts/lint/openapi_inventory_test.py -q, rg generated surfaces -->
- [x] 0.3 B3 addendum：修复 `shared/events.yaml` `debrief.created.roundType: $ref:b1.InterviewerRole` → `$ref:b1.DebriefRoundType`；同步 `shared/events/baseline/events.v1.json` + `shared/events/schemas/debrief.created.v1.json` + `shared/events/__fixtures__/envelopes.json`；`make codegen-events-check` + `make lint-events` 通过；commit `fix(events): align debrief.created roundType reference` 落地
  <!-- verified: 2026-05-16 red-green; commands: make codegen-events, make lint-events, python3 scripts/lint/events_inventory.py shared/events.yaml shared/jobs.yaml, python3 -m pytest scripts/lint/events_inventory_test.py -q, rg generated DebriefRoundType -->
- [x] 0.4 B4 addendum：`migrations/000001_create_baseline.up.sql` 与 `migrations/enum-sources.yaml` 将 `debrief_suggest_questions` 加入 `ai_task_runs.task_type` allowlist；不新增 B4 列；`migrations/lint.sh` + `make migrate-check` 通过；commit `fix(migrations): allow debrief suggestion ai task type` 落地（验证：migration replay 允许 `debrief_generate` / `debrief_suggest_questions`，拒绝未登记 task_type）
  <!-- verified: 2026-05-16 red-green; commands: go test ./internal/migrations focused, python3 -m pytest scripts/lint/migrations_lint_test.py -q, python3 scripts/lint/migrations_lint.py --repo-root ., ./migrations/lint.sh, set -a; . deploy/dev-stack/.env; set +a; make migrate-check -->
- [x] 0.5 F3 addendum：新增 `debrief.suggest_questions` feature_key + `debrief.suggest_questions.default` model profile + 基线 prompt/rubric v0.1.0 + `config/prompts/debrief.suggest_questions/v0.1.0*.{yaml,md}` / `config/rubrics/debrief.suggest_questions/v0.1.0*.yaml`；同步 prompt-rubric-registry seed migration 与 baseline tests；commit `feat(prompt-rubric-registry): seed debrief.suggest_questions baseline` 落地（验证：`backend/internal/ai/registry` 测试通过；`grep -rn "debrief.suggest_questions" config/prompts config/rubrics config/ai-profiles.yaml backend/internal/ai/registry backend/internal/shared/featurekeys` 命中）
  <!-- verified: 2026-05-16 red-green; commands: go test ./internal/ai/registry focused, go test ./internal/ai/aiclient/profile ./internal/shared/featurekeys focused, make lint-prompts, make lint-rubrics, make lint-ai-profile-coverage, python3 -m pytest scripts/lint/prompt_lint_test.py scripts/lint/rubric_lint_test.py scripts/lint/ai_profile_coverage_test.py -q, set -a; . deploy/dev-stack/.env; set +a; make migrate-check, rg debrief.suggest_questions/debrief_suggest_questions -->
- [x] 0.6 backend-practice 现状验证（Q-3）：`grep -rn "goal.*debrief\|mode.*debrief\|PracticeGoal.*debrief\|PracticeMode.*debrief" backend/internal/practice` 找出现状；若已支持，记录 verifying test names；若未支持，暂停 plan 001 转 backend-practice owner addendum，记录依赖 PR 链接；恢复后在本 checklist 行下方追加 `[依赖] backend-practice/<plan>` 注脚
  <!-- verified: 2026-05-16 dependency backend-practice/004-derived-plans-debrief landed in current branch; `cd backend && go test ./internal/practice ./internal/store/practice ./internal/api/practice ./cmd/api -run 'Derived|Debrief|Source' -count=1`; `cd backend && go test ./cmd/api -run 'TestE2EP0070|TestE2EP0071|TestE2EP0072|TestE2EP0073' -count=1`; scoped legacy grep shows runtime/generated/fixtures have no mode=debrief or PracticeModeDebrief. -->
- [x] 0.7 Phase 0 收口 quality gates：`cd backend && go test ./... -count=1` / `make codegen-check` / `make validate-fixtures` / `make lint-events` / `make codegen-events-check` / `migrations/lint.sh` / `make migrate-check`（dev-stack .env）/ `python3 -m pytest scripts/lint -q` / `make docs-check` / `git diff --check` 全部通过；commit `chore(backend-debrief): close phase 0 cross-owner addendums and validation` 落地
  <!-- verified: 2026-05-16 `cd backend && go test ./... -count=1`; `make codegen-check`; `make validate-fixtures`; `make lint-events`; `make codegen-events-check`; `./migrations/lint.sh`; `set -a; . deploy/dev-stack/.env; set +a; make migrate-check`; `python3 -m pytest scripts/lint -q`; `make docs-check`; `git diff --check`; `git diff --cached --check` -->

## Phase 1: createDebrief handler 与 store 骨架

- [x] 1.1 `backend/internal/debrief/` 包结构：`service.go` / `service_test.go` / `errors.go`；`backend/internal/store/debrief/` 包结构：`store.go` / `store_test.go`；`backend/internal/api/debriefs/` 包结构：`handler.go` / `handler_test.go`。验证测试：`cd backend && go build ./internal/debrief ./internal/store/debrief ./internal/api/debriefs`
  <!-- verified: 2026-05-16 red: package directories missing; green: cd backend && go build ./internal/debrief ./internal/store/debrief ./internal/api/debriefs -->
- [x] 1.2 Store 接口定义：`Store{CreateDebrief, GetDebrief, UpdateDebriefCompleted}` 完整 method signature；空 mock 实现；`TestStoreInterface_Compiles` 通过
  <!-- verified: 2026-05-16 red-note: Store interface already existed from 1.1 skeleton; green: cd backend && go test ./internal/debrief -run TestStoreInterface_Compiles -count=1 -->
- [x] 1.3 createDebrief handler skeleton：注入 user_id；解析 generated `CreateDebriefRequest`；校验 questions / 长度边界；返回 stub `DebriefWithJob` 202。测试断言：`TestCreateDebrief_ValidationError_EmptyQuestions`（[test-plan §1.1](./test-plan.md#11-testcreatedebrief_validationerror_emptyquestions)）+ `TestCreateDebrief_ValidationError_LongQuestionText`（[test-plan §1.2](./test-plan.md#12-testcreatedebrief_validationerror_longquestiontext)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/api/debriefs -run 'TestCreateDebrief' -count=1 -->
- [x] 1.4 idempotency middleware 接线：确认 `Idempotency-Key` 对 `POST /debriefs` 生效；resource_type='debrief' 通过中间件；测试 `TestCreateDebrief_IdempotencyEnabled` 通过；如 backend-practice IK middleware 未支持 resource_type='debrief'，扩展中间件并记录 commit
  <!-- verified: 2026-05-16 cd backend && go test ./internal/api/debriefs -run 'TestCreateDebrief_IdempotencyEnabled|TestCreateDebrief' -count=1 -->
- [x] 1.5 BDD-Gate: phase 1 不产生独立 BDD 场景，由后续 phase scenario 验证 createDebrief 入口
  <!-- verified: 2026-05-16 Phase 1 creates package skeleton + handler/middleware seams only; user-visible createDebrief scenario remains deferred to Phase 2/4/6 E2E.P0.060. Phase gates: cd backend && go test ./internal/debrief ./internal/api/debriefs ./internal/store/debrief -count=1; cd backend && go build ./internal/debrief ./internal/store/debrief ./internal/api/debriefs; make docs-check; git diff --check -->

## Phase 2: createDebrief 完整事务 + outbox

- [x] 2.1 store.CreateDebrief 实现：单一 SQL transaction 同事务写 `debriefs(status='draft')` + `async_jobs(debrief_generate, queued, dedupe_key=debriefId)` + `outbox_events(debrief.created, payload={debriefId,targetJobId,roundType,questionCount})`；事务失败回滚整体。测试断言：`TestStoreCreateDebrief_HappyTransaction`（[test-plan §2.1](./test-plan.md#21-teststorecreatedebrief_happytransaction)）+ `TestStoreCreateDebrief_RollbackOnOutboxFailure`（[test-plan §2.2](./test-plan.md#22-teststorecreatedebrief_rollbackonoutboxfailure)）
  <!-- verified: 2026-05-16 red: Repository.CreateDebrief undefined; green: cd backend && go test ./internal/store/debrief -run 'TestStoreCreateDebrief' -count=1 -->
- [x] 2.2 service.CreateDebrief 实现：调 store；事务外写 `audit_events(action='create_debrief', resource_type='debrief', metadata={...})`；返回给 handler。测试：`TestServiceCreateDebrief_AuditEmitted`（[test-plan §2.3](./test-plan.md#23-testservicecreatedebrief_auditemitted)）
  <!-- verified: 2026-05-16 red: ServiceOptions.Audit/CreateDebrief undefined; green: cd backend && go test ./internal/debrief -run TestServiceCreateDebrief_AuditEmitted -count=1; store audit writer covered by cd backend && go test ./internal/store/debrief -run 'TestRepositoryRecordDebriefAuditEvent|TestStoreCreateDebrief' -count=1 -->
- [x] 2.3 handler 串联：handler 调 service → 返回 generated `DebriefWithJob` 202。测试：`TestCreateDebrief_HappyResponse`（[test-plan §2.4](./test-plan.md#24-testcreatedebrief_happyresponse)） + `TestCreateDebrief_IdempotencyReplay_SameBody`（[test-plan §2.5](./test-plan.md#25-testcreatedebrief_idempotencyreplay_samebody)） + `TestCreateDebrief_IdempotencyMismatch_DifferentBody`（[test-plan §2.6](./test-plan.md#26-testcreatedebrief_idempotencymismatch_differentbody)）
  <!-- verified: 2026-05-16 red: mismatch returned PRACTICE_SESSION_CONFLICT; green: cd backend && go test ./internal/api/debriefs -run 'TestCreateDebrief_HappyResponse|TestCreateDebrief_IdempotencyReplay_SameBody|TestCreateDebrief_IdempotencyMismatch_DifferentBody' -count=1; shared middleware regression: cd backend && go test ./internal/middleware/idempotency ./internal/api/practice ./internal/resume/handler ./internal/upload/handler -count=1 -->
- [x] 2.4 fixture parity：`make validate-fixtures` 确认 `Debriefs/createDebrief.json` 与 handler 真实响应一致；`cd backend && go test ./internal/api/debriefs -count=1` 通过
  <!-- verified: 2026-05-16 make validate-fixtures; cd backend && go test ./internal/api/debriefs -count=1 -->
- [x] 2.5 outbox payload schema 校验：单元测试反序列化 outbox row 后 assert keys 集合 = {debriefId, targetJobId, roundType, questionCount}；keys 不含 raw text 字段。测试：`TestCreateDebrief_OutboxPayloadSchema`（[test-plan §2.7](./test-plan.md#27-testcreatedebrief_outboxpayloadschema)）
  <!-- verified: 2026-05-16 red: buildDebriefCreatedPayload undefined; green: cd backend && go test ./internal/store/debrief -run 'TestCreateDebrief_OutboxPayloadSchema|TestStoreCreateDebrief' -count=1 -->
- [x] 2.6 BDD-Gate: E2E.P0.060 覆盖 createDebrief 主路径 → drainer handle → completed（待 Phase 4 完成后整体验证）
  <!-- verified: 2026-05-16 test/scenarios/e2e/p0-060-debrief-create-worker-happy/scripts/setup.sh -> trigger.sh -> verify.sh -> cleanup.sh -->

## Phase 3: suggestDebriefQuestions sync handler

- [x] 3.1 service.SuggestQuestions 实现：拉 target_job 摘要 + 可选 session + 可选 resume；F3 Resolve + A3 Complete；解析输出；写 ai_task_runs + audit。测试：`TestServiceSuggestQuestions_Happy`（[test-plan §3.1](./test-plan.md#31-testservicesuggestquestions_happy)） + `TestServiceSuggestQuestions_CrossUserTargetJob_403`（[test-plan §3.2](./test-plan.md#32-testservicesuggestquestions_crossusertargetjob_403)）
  <!-- verified: 2026-05-16 red: SuggestionContext/Registry/AI service surface undefined; green: cd backend && go test ./internal/debrief -run 'TestServiceSuggestQuestions_Happy|TestServiceSuggestQuestions_CrossUserTargetJob_403' -count=1. Cross-user target follows current targetjob isolation semantics through ErrDebriefPrerequisite/404 boundary; no unregistered FORBIDDEN code introduced. -->
- [x] 3.2 AI 失败映射：F3 ResolveActive/config 失败 → B1 `AI_PROVIDER_CONFIG_INVALID`；A3 secret missing → `AI_PROVIDER_SECRET_MISSING`；A3 timeout → `AI_PROVIDER_TIMEOUT`；fallback exhausted/provider unreachable → `AI_FALLBACK_EXHAUSTED`；A3 invalid JSON / parsed empty → `AI_OUTPUT_INVALID`；写 ai_task_runs status='failed' + error_code。测试：`TestServiceSuggestQuestions_F3ResolveFailed`（[test-plan §3.3](./test-plan.md#33-testservicesuggestquestions_f3resolvefailed)） + `TestServiceSuggestQuestions_A3Timeout`（[test-plan §3.4](./test-plan.md#34-testservicesuggestquestions_a3timeout)） + `TestServiceSuggestQuestions_ParseFailed`（[test-plan §3.5](./test-plan.md#35-testservicesuggestquestions_parsefailed)）
  <!-- verified: 2026-05-16 red: F3 resolve failure did not write ai_task_runs row; green: cd backend && go test ./internal/debrief -run 'TestServiceSuggestQuestions_F3ResolveFailed|TestServiceSuggestQuestions_A3Timeout|TestServiceSuggestQuestions_ParseFailed|TestServiceSuggestQuestions_Happy|TestServiceSuggestQuestions_CrossUserTargetJob_403' -count=1 -->
- [x] 3.3 handler skeleton：注入 user_id；解析 `SuggestDebriefQuestionsRequest`；校验 count ∈ [1,10]；调 service；返回 200 或 5xx。测试：`TestSuggestDebriefQuestions_CountBoundary`（[test-plan §3.6](./test-plan.md#36-testsuggestdebriefquestions_countboundary)）+ `TestSuggestDebriefQuestions_Unauthenticated_401`（[test-plan §3.7](./test-plan.md#37-testsuggestdebriefquestions_unauthenticated_401)）
  <!-- verified: 2026-05-16 red: Handler.SuggestDebriefQuestions undefined; green: cd backend && go test ./internal/api/debriefs -run 'TestSuggestDebriefQuestions_CountBoundary|TestSuggestDebriefQuestions_Unauthenticated_401' -count=1 -->
- [x] 3.4 fixture parity：`make validate-fixtures` 确认 `suggestDebriefQuestions.json` `default` / `empty` / `prototype-baseline` variants 与 handler 一致
  <!-- verified: 2026-05-16 make validate-fixtures; cd backend && go test ./internal/debrief ./internal/api/debriefs -count=1; cd backend && go test ./internal/ai/aiclient -count=1 -->
- [x] 3.5 BDD-Gate: E2E.P0.063 覆盖 suggestDebriefQuestions 主路径 + AI failure（待 Phase 6 整体验证）
  <!-- verified: 2026-05-16 test/scenarios/e2e/p0-063-debrief-suggest-questions/scripts/setup.sh -> trigger.sh -> verify.sh -> cleanup.sh -->
  <!-- pending: service/handler unit and fixture gates are complete; scenario execution remains blocked until Phase 6 scenario run phase by plan design. -->

## Phase 4: debrief_generate worker handler

- [x] 4.1 `debrief.GenerateHandler` 实现 `targetjob.JobHandler` interface：解析 job.payload → 拉 debrief.raw_questions + target_job 摘要 → F3 `debrief.generate` Resolve + A3 Complete → 解析 aiAnalyses + risk_items。测试：`TestGenerateHandler_HappyResolution`（[test-plan §4.1](./test-plan.md#41-testgeneratehandler_happyresolution)） + `TestGenerateHandler_PromptContextAssembled`（[test-plan §4.2](./test-plan.md#42-testgeneratehandler_promptcontextassembled)）
  <!-- verified: 2026-05-16 red: NewGenerateHandler/GenerateContext undefined; green: cd backend && go test ./internal/debrief -run 'TestGenerateHandler_HappyResolution|TestGenerateHandler_PromptContextAssembled' -count=1 -->
  <!-- verified: 2026-05-16 debrief.generate prompt/rubric baseline and seed migration aligned to worker `questions`/`riskItems` schema; commands: python3 scripts/lint/prompt_lint.py --prompts-dir config/prompts --migrations-dir migrations; python3 scripts/lint/rubric_lint.py --rubrics-dir config/rubrics; ./migrations/lint.sh -->
- [x] 4.2 cmd/api bootstrap 注册：`drainer.RegisterHandler("debrief_generate", debrief.NewGenerateHandler(...))`；启动 backend 时 drainer 含 `debrief_generate` 路由；测试 `TestDrainer_DebriefGenerateRegistered`（[test-plan §4.3](./test-plan.md#43-testdrainer_debriefgenerateregistered)）
  <!-- verified: 2026-05-16 cd backend && go test ./cmd/api -run 'TestDrainer_DebriefGenerateRegistered' -count=1 -->
  <!-- supporting: 2026-05-16 cd backend && go test ./internal/store/debrief -run 'TestStoreLoadGenerateContext|TestStoreUpdateDebriefCompleted' -count=1 -->
- [x] 4.3 store.UpdateDebriefCompleted：单一 SQL transaction CAS（draft→completed only）+ raw_questions jsonb patch + risk_items + provenance 4 列写值 + outbox `debrief.completed`；事务失败回滚。测试：`TestStoreUpdateDebriefCompleted_HappyTransaction`（[test-plan §4.4](./test-plan.md#44-teststoreupdatedebriefcompleted_happytransaction)） + `TestStoreUpdateDebriefCompleted_CASRejectsCompleted`（[test-plan §4.5](./test-plan.md#45-teststoreupdatedebriefcompleted_casrejectscompleted)） + `TestStoreUpdateDebriefCompleted_OutboxRollback`（[test-plan §4.6](./test-plan.md#46-teststoreupdatedebriefcompleted_outboxrollback)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/store/debrief -run 'TestStoreUpdateDebriefCompleted' -count=1 -->
- [x] 4.4 outbox payload schema：`debrief.completed.payload` 严格 = {debriefId, targetJobId, riskItemCount, practiceFocusCount}；不含 notes / question text / answer text / risk_items prose。测试：`TestGenerateHandler_OutboxPayloadSchema`（[test-plan §4.7](./test-plan.md#47-testgeneratehandler_outboxpayloadschema)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/store/debrief -run 'TestStoreLoadGenerateContext|TestStoreUpdateDebriefCompleted|TestGenerateHandler_OutboxPayloadSchema' -count=1 -->
- [x] 4.5 AI failure semantics：F3/A3 失败 → handler 返回 `targetjob.RetryableError`；ai_task_runs 写 failed row + B1 error_code；drainer 自动 attempts+1 + backoff。测试：`TestGenerateHandler_F3ResolveFailed`（[test-plan §4.8](./test-plan.md#48-testgeneratehandler_f3resolvefailed)） + `TestGenerateHandler_A3Timeout`（[test-plan §4.9](./test-plan.md#49-testgeneratehandler_a3timeout)） + `TestGenerateHandler_ParseEmpty`（[test-plan §4.10](./test-plan.md#410-testgeneratehandler_parseempty)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/debrief -run 'TestGenerateHandler' -count=1 -->
- [x] 4.6 permanent fail at attempts=5：模拟 attempts=4 + 第 5 次失败 → drainer 自动置 `async_jobs.status='failed'`；`debriefs.status='draft'` 保持；outbox `debrief.completed` 不发出。测试：`TestGenerateHandler_PermanentFailAt5Attempts`（[test-plan §4.11](./test-plan.md#411-testgeneratehandler_permanentfailat5attempts)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/debrief -run 'TestGenerateHandler' -count=1 -->
- [x] 4.7 BDD-Gate: E2E.P0.060 + E2E.P0.062 覆盖 worker happy + worker failure + retry
  <!-- verified: 2026-05-16 test/scenarios/e2e/p0-060-debrief-create-worker-happy and p0-062-debrief-worker-retry-failure four-step scripts passed -->

## Phase 5: getDebrief read handler

- [x] 5.1 store.GetDebrief：SQL `SELECT ... FROM debriefs WHERE id=$1 AND user_id=$2`；map to `Debrief` schema；status='draft' 时 questions[*].aiAnalysis=null + riskItems=[] + provenance=null；status='completed' 时全字段填充；未找到返回 ErrNotFound。测试：`TestStoreGetDebrief_DraftPartial`（[test-plan §5.1](./test-plan.md#51-teststoregetdebrief_draftpartial)） + `TestStoreGetDebrief_CompletedFull`（[test-plan §5.2](./test-plan.md#52-teststoregetdebrief_completedfull)） + `TestStoreGetDebrief_CrossUserNotFound`（[test-plan §5.3](./test-plan.md#53-teststoregetdebrief_crossusernotfound)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/store/debrief -run 'TestStoreGetDebrief|TestStoreGetSuggestionContext' -count=1 -->
- [x] 5.2 service.GetDebrief：hydrate provenance 6 字段 (P0: featureFlag=null, dataSourceVersion='debrief/<debriefId>@v1' 字面量；其他来自 debriefs 列)。测试：`TestServiceGetDebrief_ProvenanceWireOnly`（[test-plan §5.4](./test-plan.md#54-testservicegetdebrief_provenancewireonly)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/debrief -run 'TestServiceGetDebrief_ProvenanceWireOnly' -count=1 -->
- [x] 5.3 handler.GetDebrief：注入 user_id；解析 path param；调 service；map ErrNotFound → `404 DEBRIEF_NOT_FOUND`；返回 generated `Debrief` 200。测试：`TestGetDebrief_DraftResponse`（[test-plan §5.5](./test-plan.md#55-testgetdebrief_draftresponse)） + `TestGetDebrief_CompletedResponse`（[test-plan §5.6](./test-plan.md#56-testgetdebrief_completedresponse)） + `TestGetDebrief_CrossUser404`（[test-plan §5.7](./test-plan.md#57-testgetdebrief_crossuser404)） + `TestGetDebrief_NotFound404`（[test-plan §5.8](./test-plan.md#58-testgetdebrief_notfound404)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/api/debriefs -run 'TestGetDebrief' -count=1 -->
  <!-- supporting: 2026-05-16 cd backend && go test ./cmd/api -run 'TestBuildDebriefRoutesWiresHandlerAndIdempotency|TestDrainer_DebriefGenerateRegistered' -count=1 -->
- [x] 5.4 fixture parity：`make validate-fixtures` 确认 `Debriefs/getDebrief.json` `default` / `debrief-draft` / `prototype-baseline` variants 与 handler 一致
  <!-- verified: 2026-05-16 make validate-fixtures -->
- [x] 5.5 BDD-Gate: E2E.P0.061 覆盖 getDebrief draft/completed 双态 + cross-user 隔离
  <!-- verified: 2026-05-16 test/scenarios/e2e/p0-061-debrief-get-isolation/scripts/setup.sh -> trigger.sh -> verify.sh -> cleanup.sh -->

## Phase 6: 隐私 / 观测 / retry / legacy negative + BDD 收口

- [x] 6.1 隐私红线：单元测试断言 `debrief.created` / `debrief.completed` outbox payload 不含 raw text；audit_events metadata 不含 raw text；`grep -rn "questionText\|myAnswerSummary\|interviewerReaction\|notes" backend/internal/debrief/service.go backend/internal/store/debrief/store.go shared/events.yaml shared/events/baseline/events.v1.json | grep -v "_test.go\|piiBoundary\|// "` 不命中实际 prose 输出。测试：`TestOutboxPayload_NoRawText`（[test-plan §6.1](./test-plan.md#61-testoutboxpayload_norawtext)） + `TestAuditEvents_NoRawText`（[test-plan §6.2](./test-plan.md#62-testauditevents_norawtext)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/store/debrief -run 'TestOutboxPayload_NoRawText' -count=1; cd backend && go test ./internal/debrief -run 'TestAuditEvents_NoRawText' -count=1 -->
- [x] 6.2 观测红线：`ai_task_runs` 写两种 task_type；audit_events 三种 action；F1 metric label 不含 raw text；如新增 metric 已在 F1 owner co-author addendum 中登记。测试：`TestAITaskRunsWritten`（[test-plan §6.3](./test-plan.md#63-testaitaskrunswritten)） + `TestAuditEventsWritten`（[test-plan §6.4](./test-plan.md#64-testauditeventswritten)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/debrief -run 'TestAITaskRunsWritten|TestAuditEventsWritten|TestAuditEvents_NoRawText' -count=1 -->
- [x] 6.3 Retry policy unit：drainer behavior 测试 attempts<5 vs attempts>=5；测试：`TestRetryPolicy_BackoffBelowMax`（[test-plan §6.5](./test-plan.md#65-testretrypolicy_backoffbelowmax)） + `TestRetryPolicy_PermanentFailAtMax`（[test-plan §6.6](./test-plan.md#66-testretrypolicy_permanentfailatmax)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/targetjob -run 'TestRetryPolicy_BackoffBelowMax|TestRetryPolicy_PermanentFailAtMax' -count=1 -->
- [x] 6.4 Legacy negative grep：`python3 scripts/lint/backend_debrief_legacy.py --phase all` 扫描 runtime / shared event-job contract / fixtures / scenario runtime，不扫描负向门禁说明文档；retired terms `mistakes_count` / `generatedMistakeCount` / `experience_library` / `drill_builder` / `growth_center` / `star_editor` / `debrief_voice` 0 命中
  <!-- verified: 2026-05-16 python3 scripts/lint/backend_debrief_legacy.py --phase all -->
- [x] 6.5 Legacy negative lint script：新增 `scripts/lint/backend_debrief_legacy.py --phase all` 覆盖 backend/internal/debrief + openapi/fixtures/Debriefs + scenario runtime；`python3 -m pytest scripts/lint -q` 通过
  <!-- verified: 2026-05-16 python3 -m pytest scripts/lint/backend_debrief_legacy_test.py -q; full scripts/lint pytest runs in Phase 7 -->
- [x] 6.6 BDD-Gate: 执行 P0.060 scenario 四段脚本（`scripts/setup.sh` → `scripts/trigger.sh` → `scripts/verify.sh` → `scripts/cleanup.sh`）通过；verify.sh 含 backend runner pass marker + 旧口径 grep 反查
  <!-- verified: 2026-05-16 p0-060-debrief-create-worker-happy four-step scripts passed -->
- [x] 6.7 BDD-Gate: 执行 P0.061 scenario 四段脚本通过
  <!-- verified: 2026-05-16 p0-061-debrief-get-isolation four-step scripts passed -->
- [x] 6.8 BDD-Gate: 执行 P0.062 scenario 四段脚本通过
  <!-- verified: 2026-05-16 p0-062-debrief-worker-retry-failure four-step scripts passed -->
- [x] 6.9 BDD-Gate: 执行 P0.063 scenario 四段脚本通过
  <!-- verified: 2026-05-16 p0-063-debrief-suggest-questions four-step scripts passed -->
- [x] 6.10 BDD-Gate: 执行 P0.064 scenario 四段脚本通过
  <!-- verified: 2026-05-16 p0-064-debrief-privacy-legacy four-step scripts passed -->

## Phase 7: Plan 收口

- [x] 7.1 全局回归：`cd backend && go test ./... -count=1` / `make codegen-check` / `make validate-fixtures` / `make lint-events` / `make codegen-events-check` / `make migrate-check`（dev-stack .env）/ `migrations/lint.sh` / `python3 -m pytest scripts/lint -q` / `make docs-check` / `git diff --check` 全部通过
  <!-- verified: 2026-05-16 `cd backend && go test ./internal/debrief ./internal/api/debriefs ./internal/store/debrief ./cmd/api -count=1 -race`; `cd backend && go test ./... -count=1`; `python3 -m pytest scripts/lint -q`; `python3 scripts/lint/prompt_lint.py --prompts-dir config/prompts --migrations-dir migrations`; `python3 scripts/lint/rubric_lint.py --rubrics-dir config/rubrics`; `make codegen-check`; `make validate-fixtures`; `make lint-events`; `make codegen-events-check`; `./migrations/lint.sh`; `set -a; . deploy/dev-stack/.env; set +a; make migrate-check`; `make docs-check`; `git diff --check` -->
- [x] 7.2 plans/INDEX.md 把 001 从 active 移到 completed，记录完成日期 2026-05-16
  <!-- verified: 2026-05-16 docs/spec/backend-debrief/plans/INDEX.md completed projection updated -->
- [x] 7.3 backend-debrief/history.md 增加 completion 行
  <!-- verified: 2026-05-16 docs/spec/backend-debrief/history.md records plan 001 completion -->
- [x] 7.4 提交 commit `feat(backend-debrief): close 001 debrief record and analysis baseline`；记录工作日志 `/work-journal`
  <!-- verified: 2026-05-16 work-journal entry prepared for the same commit message; commit is this close-out change set -->
