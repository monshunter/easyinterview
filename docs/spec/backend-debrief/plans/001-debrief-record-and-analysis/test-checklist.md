# 001 Debrief Record and Analysis Test Checklist

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-05-16

**关联 Test Plan**: [test-plan](./test-plan.md)
**关联计划**: [plan](./plan.md)
**关联 Spec**: [spec](../../spec.md)

## Phase 0: Cross-owner pre-launch addendums 验证

- [x] 0.A `make codegen-check` 通过（B1 enum / B2 operation / generated types 同步）
  <!-- verified: 2026-05-16 make codegen-check -->
- [x] 0.B `make validate-fixtures` 通过（含 createDebrief / getDebrief / suggestDebriefQuestions 三套 fixtures）
  <!-- verified: 2026-05-16 make validate-fixtures -> OK 56 fixtures -->
- [x] 0.C `make lint-events` 通过（events.yaml roundType 已修正）
  <!-- verified: 2026-05-16 make lint-events; events_inventory.py DebriefRoundType gate -->
- [x] 0.D `make codegen-events-check` 通过（events baseline JSON / schemas 同步）
  <!-- verified: 2026-05-16 make codegen-events-check -->
- [x] 0.E F3 baseline test：prompt-rubric-registry baseline test 含 `debrief.suggest_questions` v0.1.0 active 断言
  <!-- verified: 2026-05-16 go test ./internal/ai/registry focused; make lint-prompts; make lint-rubrics; make lint-ai-profile-coverage -->
- [x] 0.F backend-practice Q-3 验证：现状已支持 `goal='debrief'` + 合法 `mode IN ('assisted','strict')`（grep + test names 已记录）；如不支持，依赖 PR 已合入
  <!-- verified: 2026-05-16 dependency backend-practice/004-derived-plans-debrief; `cd backend && go test ./internal/practice ./internal/store/practice ./internal/api/practice ./cmd/api -run 'Derived|Debrief|Source' -count=1`; `cd backend && go test ./cmd/api -run 'TestE2EP0070|TestE2EP0071|TestE2EP0072|TestE2EP0073' -count=1`; scoped legacy grep no runtime/generated/fixture matches -->
- [x] 0.G `migrations/lint.sh` + `make migrate-check`（dev-stack .env）通过
  <!-- verified: 2026-05-16 ./migrations/lint.sh; set -a; . deploy/dev-stack/.env; set +a; make migrate-check -->

## Phase 1: createDebrief handler 与 store 骨架

- [x] 1.A TestCreateDebrief_ValidationError_EmptyQuestions 通过（[test-plan §1.1](./test-plan.md#11-testcreatedebrief_validationerror_emptyquestions)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/api/debriefs -run 'TestCreateDebrief' -count=1 -->
- [x] 1.B TestCreateDebrief_ValidationError_LongQuestionText 通过（[test-plan §1.2](./test-plan.md#12-testcreatedebrief_validationerror_longquestiontext)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/api/debriefs -run 'TestCreateDebrief' -count=1 -->
- [x] 1.C TestStoreInterface_Compiles 通过（store interface 接口定义完整）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/debrief -run TestStoreInterface_Compiles -count=1 -->

## Phase 2: createDebrief 完整事务 + outbox

- [x] 2.A TestStoreCreateDebrief_HappyTransaction 通过（[test-plan §2.1](./test-plan.md#21-teststorecreatedebrief_happytransaction)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/store/debrief -run 'TestStoreCreateDebrief' -count=1 -->
- [x] 2.B TestStoreCreateDebrief_RollbackOnOutboxFailure 通过（[test-plan §2.2](./test-plan.md#22-teststorecreatedebrief_rollbackonoutboxfailure)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/store/debrief -run 'TestStoreCreateDebrief' -count=1 -->
- [x] 2.C TestServiceCreateDebrief_AuditEmitted 通过（[test-plan §2.3](./test-plan.md#23-testservicecreatedebrief_auditemitted)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/debrief -run TestServiceCreateDebrief_AuditEmitted -count=1 -->
- [x] 2.D TestCreateDebrief_HappyResponse 通过（[test-plan §2.4](./test-plan.md#24-testcreatedebrief_happyresponse)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/api/debriefs -run 'TestCreateDebrief_HappyResponse|TestCreateDebrief_IdempotencyReplay_SameBody|TestCreateDebrief_IdempotencyMismatch_DifferentBody' -count=1 -->
- [x] 2.E TestCreateDebrief_IdempotencyReplay_SameBody 通过（[test-plan §2.5](./test-plan.md#25-testcreatedebrief_idempotencyreplay_samebody)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/api/debriefs -run 'TestCreateDebrief_HappyResponse|TestCreateDebrief_IdempotencyReplay_SameBody|TestCreateDebrief_IdempotencyMismatch_DifferentBody' -count=1 -->
- [x] 2.F TestCreateDebrief_IdempotencyMismatch_DifferentBody 通过（[test-plan §2.6](./test-plan.md#26-testcreatedebrief_idempotencymismatch_differentbody)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/api/debriefs -run 'TestCreateDebrief_HappyResponse|TestCreateDebrief_IdempotencyReplay_SameBody|TestCreateDebrief_IdempotencyMismatch_DifferentBody' -count=1 -->
- [x] 2.G TestCreateDebrief_OutboxPayloadSchema 通过（[test-plan §2.7](./test-plan.md#27-testcreatedebrief_outboxpayloadschema)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/store/debrief -run 'TestCreateDebrief_OutboxPayloadSchema|TestStoreCreateDebrief' -count=1 -->

## Phase 3: suggestDebriefQuestions sync handler

- [x] 3.A TestServiceSuggestQuestions_Happy 通过（[test-plan §3.1](./test-plan.md#31-testservicesuggestquestions_happy)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/debrief -run 'TestServiceSuggestQuestions_Happy|TestServiceSuggestQuestions_CrossUserTargetJob_403' -count=1 -->
- [x] 3.B TestServiceSuggestQuestions_CrossUserTargetJob_403 通过（[test-plan §3.2](./test-plan.md#32-testservicesuggestquestions_crossusertargetjob_403)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/debrief -run 'TestServiceSuggestQuestions_Happy|TestServiceSuggestQuestions_CrossUserTargetJob_403' -count=1 -->
- [x] 3.C TestServiceSuggestQuestions_F3ResolveFailed 通过（[test-plan §3.3](./test-plan.md#33-testservicesuggestquestions_f3resolvefailed)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/debrief -run 'TestServiceSuggestQuestions_F3ResolveFailed|TestServiceSuggestQuestions_A3Timeout|TestServiceSuggestQuestions_ParseFailed' -count=1 -->
- [x] 3.D TestServiceSuggestQuestions_A3Timeout 通过（[test-plan §3.4](./test-plan.md#34-testservicesuggestquestions_a3timeout)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/debrief -run 'TestServiceSuggestQuestions_F3ResolveFailed|TestServiceSuggestQuestions_A3Timeout|TestServiceSuggestQuestions_ParseFailed' -count=1 -->
- [x] 3.E TestServiceSuggestQuestions_ParseFailed 通过（[test-plan §3.5](./test-plan.md#35-testservicesuggestquestions_parsefailed)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/debrief -run 'TestServiceSuggestQuestions_F3ResolveFailed|TestServiceSuggestQuestions_A3Timeout|TestServiceSuggestQuestions_ParseFailed' -count=1 -->
- [x] 3.F TestSuggestDebriefQuestions_CountBoundary 通过（[test-plan §3.6](./test-plan.md#36-testsuggestdebriefquestions_countboundary)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/api/debriefs -run 'TestSuggestDebriefQuestions_CountBoundary|TestSuggestDebriefQuestions_Unauthenticated_401' -count=1 -->
- [x] 3.G TestSuggestDebriefQuestions_Unauthenticated_401 通过（[test-plan §3.7](./test-plan.md#37-testsuggestdebriefquestions_unauthenticated_401)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/api/debriefs -run 'TestSuggestDebriefQuestions_CountBoundary|TestSuggestDebriefQuestions_Unauthenticated_401' -count=1 -->

## Phase 4: debrief_generate worker handler

- [x] 4.A TestGenerateHandler_HappyResolution 通过（[test-plan §4.1](./test-plan.md#41-testgeneratehandler_happyresolution)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/debrief -run 'TestGenerateHandler_HappyResolution|TestGenerateHandler_PromptContextAssembled' -count=1 -->
  <!-- verified: 2026-05-16 debrief.generate prompt/rubric baseline aligned with worker parser schema; prompt_lint, rubric_lint, and migrations/lint.sh passed -->
- [x] 4.B TestGenerateHandler_PromptContextAssembled 通过（[test-plan §4.2](./test-plan.md#42-testgeneratehandler_promptcontextassembled)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/debrief -run 'TestGenerateHandler_HappyResolution|TestGenerateHandler_PromptContextAssembled' -count=1 -->
- [x] 4.C TestDrainer_DebriefGenerateRegistered 通过（[test-plan §4.3](./test-plan.md#43-testdrainer_debriefgenerateregistered)）
  <!-- verified: 2026-05-16 cd backend && go test ./cmd/api -run 'TestDrainer_DebriefGenerateRegistered' -count=1 -->
- [x] 4.D TestStoreUpdateDebriefCompleted_HappyTransaction 通过（[test-plan §4.4](./test-plan.md#44-teststoreupdatedebriefcompleted_happytransaction)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/store/debrief -run 'TestStoreUpdateDebriefCompleted' -count=1 -->
- [x] 4.E TestStoreUpdateDebriefCompleted_CASRejectsCompleted 通过（[test-plan §4.5](./test-plan.md#45-teststoreupdatedebriefcompleted_casrejectscompleted)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/store/debrief -run 'TestStoreUpdateDebriefCompleted' -count=1 -->
- [x] 4.F TestStoreUpdateDebriefCompleted_OutboxRollback 通过（[test-plan §4.6](./test-plan.md#46-teststoreupdatedebriefcompleted_outboxrollback)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/store/debrief -run 'TestStoreUpdateDebriefCompleted' -count=1 -->
- [x] 4.G TestGenerateHandler_OutboxPayloadSchema 通过（[test-plan §4.7](./test-plan.md#47-testgeneratehandler_outboxpayloadschema)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/store/debrief -run 'TestStoreLoadGenerateContext|TestStoreUpdateDebriefCompleted|TestGenerateHandler_OutboxPayloadSchema' -count=1 -->
- [x] 4.H TestGenerateHandler_F3ResolveFailed 通过（[test-plan §4.8](./test-plan.md#48-testgeneratehandler_f3resolvefailed)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/debrief -run 'TestGenerateHandler' -count=1 -->
- [x] 4.I TestGenerateHandler_A3Timeout 通过（[test-plan §4.9](./test-plan.md#49-testgeneratehandler_a3timeout)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/debrief -run 'TestGenerateHandler' -count=1 -->
- [x] 4.J TestGenerateHandler_ParseEmpty 通过（[test-plan §4.10](./test-plan.md#410-testgeneratehandler_parseempty)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/debrief -run 'TestGenerateHandler' -count=1 -->
- [x] 4.K TestGenerateHandler_PermanentFailAt5Attempts 通过（[test-plan §4.11](./test-plan.md#411-testgeneratehandler_permanentfailat5attempts)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/debrief -run 'TestGenerateHandler' -count=1 -->

## Phase 5: getDebrief read handler

- [x] 5.A TestStoreGetDebrief_DraftPartial 通过（[test-plan §5.1](./test-plan.md#51-teststoregetdebrief_draftpartial)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/store/debrief -run 'TestStoreGetDebrief|TestStoreGetSuggestionContext' -count=1 -->
- [x] 5.B TestStoreGetDebrief_CompletedFull 通过（[test-plan §5.2](./test-plan.md#52-teststoregetdebrief_completedfull)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/store/debrief -run 'TestStoreGetDebrief|TestStoreGetSuggestionContext' -count=1 -->
- [x] 5.C TestStoreGetDebrief_CrossUserNotFound 通过（[test-plan §5.3](./test-plan.md#53-teststoregetdebrief_crossusernotfound)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/store/debrief -run 'TestStoreGetDebrief|TestStoreGetSuggestionContext' -count=1 -->
- [x] 5.D TestServiceGetDebrief_ProvenanceWireOnly 通过（[test-plan §5.4](./test-plan.md#54-testservicegetdebrief_provenancewireonly)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/debrief -run 'TestServiceGetDebrief_ProvenanceWireOnly' -count=1 -->
- [x] 5.E TestGetDebrief_DraftResponse 通过（[test-plan §5.5](./test-plan.md#55-testgetdebrief_draftresponse)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/api/debriefs -run 'TestGetDebrief' -count=1 -->
- [x] 5.F TestGetDebrief_CompletedResponse 通过（[test-plan §5.6](./test-plan.md#56-testgetdebrief_completedresponse)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/api/debriefs -run 'TestGetDebrief' -count=1 -->
- [x] 5.G TestGetDebrief_CrossUser404 通过（[test-plan §5.7](./test-plan.md#57-testgetdebrief_crossuser404)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/api/debriefs -run 'TestGetDebrief' -count=1 -->
- [x] 5.H TestGetDebrief_NotFound404 通过（[test-plan §5.8](./test-plan.md#58-testgetdebrief_notfound404)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/api/debriefs -run 'TestGetDebrief' -count=1 -->

## Phase 6: 隐私 / 观测 / retry / legacy negative

- [x] 6.A TestOutboxPayload_NoRawText 通过（[test-plan §6.1](./test-plan.md#61-testoutboxpayload_norawtext)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/store/debrief -run 'TestOutboxPayload_NoRawText' -count=1 -->
- [x] 6.B TestAuditEvents_NoRawText 通过（[test-plan §6.2](./test-plan.md#62-testauditevents_norawtext)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/debrief -run 'TestAuditEvents_NoRawText' -count=1 -->
- [x] 6.C TestAITaskRunsWritten 通过（[test-plan §6.3](./test-plan.md#63-testaitaskrunswritten)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/debrief -run 'TestAITaskRunsWritten|TestAuditEventsWritten|TestAuditEvents_NoRawText' -count=1 -->
- [x] 6.D TestAuditEventsWritten 通过（[test-plan §6.4](./test-plan.md#64-testauditeventswritten)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/debrief -run 'TestAITaskRunsWritten|TestAuditEventsWritten|TestAuditEvents_NoRawText' -count=1 -->
- [x] 6.E TestRetryPolicy_BackoffBelowMax 通过（[test-plan §6.5](./test-plan.md#65-testretrypolicy_backoffbelowmax)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/targetjob -run 'TestRetryPolicy_BackoffBelowMax|TestRetryPolicy_PermanentFailAtMax' -count=1 -->
- [x] 6.F TestRetryPolicy_PermanentFailAtMax 通过（[test-plan §6.6](./test-plan.md#66-testretrypolicy_permanentfailatmax)）
  <!-- verified: 2026-05-16 cd backend && go test ./internal/targetjob -run 'TestRetryPolicy_BackoffBelowMax|TestRetryPolicy_PermanentFailAtMax' -count=1 -->
- [x] 6.G Legacy negative grep gate 通过（plan checklist 6.4 + 6.5）
  <!-- verified: 2026-05-16 python3 scripts/lint/backend_debrief_legacy.py --phase all; python3 -m pytest scripts/lint/backend_debrief_legacy_test.py -q -->

## Phase 7: 全计划单元/集成测试全量回归

- [x] 7.A `cd backend && go test ./internal/debrief ./internal/api/debriefs ./internal/store/debrief ./cmd/api -count=1 -race` 通过
  <!-- verified: 2026-05-16 cd backend && go test ./internal/debrief ./internal/api/debriefs ./internal/store/debrief ./cmd/api -count=1 -race -->
- [x] 7.B `cd backend && go test ./... -count=1` 通过（全 backend 单元 + 集成测试）
  <!-- verified: 2026-05-16 cd backend && go test ./... -count=1 -->
- [x] 7.C `python3 -m pytest scripts/lint -q` 通过（含 backend_debrief_legacy.py 与既有 lint）
  <!-- verified: 2026-05-16 python3 -m pytest scripts/lint -q -->
- [x] 7.D `make codegen-check` / `make validate-fixtures` / `make lint-events` / `make codegen-events-check` 通过
  <!-- verified: 2026-05-16 make codegen-check; make validate-fixtures; make lint-events; make codegen-events-check -->
- [x] 7.E `migrations/lint.sh` + `make migrate-check`（dev-stack .env）通过
  <!-- verified: 2026-05-16 ./migrations/lint.sh; set -a; . deploy/dev-stack/.env; set +a; make migrate-check -->
- [x] 7.F Phase 7 本计划定义的单元测试项全部通过
  <!-- verified: 2026-05-16 Phase 1-7 checklist and test-checklist items are checked with executable command evidence -->
