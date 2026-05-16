# 001 Debrief Record and Analysis Test Checklist

> **版本**: 1.0
> **状态**: active
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

- [ ] 1.A TestCreateDebrief_ValidationError_EmptyQuestions 通过（[test-plan §1.1](./test-plan.md#11-testcreatedebrief_validationerror_emptyquestions)）
- [ ] 1.B TestCreateDebrief_ValidationError_LongQuestionText 通过（[test-plan §1.2](./test-plan.md#12-testcreatedebrief_validationerror_longquestiontext)）
- [ ] 1.C TestStoreInterface_Compiles 通过（store interface 接口定义完整）

## Phase 2: createDebrief 完整事务 + outbox

- [ ] 2.A TestStoreCreateDebrief_HappyTransaction 通过（[test-plan §2.1](./test-plan.md#21-teststorecreatedebrief_happytransaction)）
- [ ] 2.B TestStoreCreateDebrief_RollbackOnOutboxFailure 通过（[test-plan §2.2](./test-plan.md#22-teststorecreatedebrief_rollbackonoutboxfailure)）
- [ ] 2.C TestServiceCreateDebrief_AuditEmitted 通过（[test-plan §2.3](./test-plan.md#23-testservicecreatedebrief_auditemitted)）
- [ ] 2.D TestCreateDebrief_HappyResponse 通过（[test-plan §2.4](./test-plan.md#24-testcreatedebrief_happyresponse)）
- [ ] 2.E TestCreateDebrief_IdempotencyReplay_SameBody 通过（[test-plan §2.5](./test-plan.md#25-testcreatedebrief_idempotencyreplay_samebody)）
- [ ] 2.F TestCreateDebrief_IdempotencyMismatch_DifferentBody 通过（[test-plan §2.6](./test-plan.md#26-testcreatedebrief_idempotencymismatch_differentbody)）
- [ ] 2.G TestCreateDebrief_OutboxPayloadSchema 通过（[test-plan §2.7](./test-plan.md#27-testcreatedebrief_outboxpayloadschema)）

## Phase 3: suggestDebriefQuestions sync handler

- [ ] 3.A TestServiceSuggestQuestions_Happy 通过（[test-plan §3.1](./test-plan.md#31-testservicesuggestquestions_happy)）
- [ ] 3.B TestServiceSuggestQuestions_CrossUserTargetJob_403 通过（[test-plan §3.2](./test-plan.md#32-testservicesuggestquestions_crossusertargetjob_403)）
- [ ] 3.C TestServiceSuggestQuestions_F3ResolveFailed 通过（[test-plan §3.3](./test-plan.md#33-testservicesuggestquestions_f3resolvefailed)）
- [ ] 3.D TestServiceSuggestQuestions_A3Timeout 通过（[test-plan §3.4](./test-plan.md#34-testservicesuggestquestions_a3timeout)）
- [ ] 3.E TestServiceSuggestQuestions_ParseFailed 通过（[test-plan §3.5](./test-plan.md#35-testservicesuggestquestions_parsefailed)）
- [ ] 3.F TestSuggestDebriefQuestions_CountBoundary 通过（[test-plan §3.6](./test-plan.md#36-testsuggestdebriefquestions_countboundary)）
- [ ] 3.G TestSuggestDebriefQuestions_Unauthenticated_401 通过（[test-plan §3.7](./test-plan.md#37-testsuggestdebriefquestions_unauthenticated_401)）

## Phase 4: debrief_generate worker handler

- [ ] 4.A TestGenerateHandler_HappyResolution 通过（[test-plan §4.1](./test-plan.md#41-testgeneratehandler_happyresolution)）
- [ ] 4.B TestGenerateHandler_PromptContextAssembled 通过（[test-plan §4.2](./test-plan.md#42-testgeneratehandler_promptcontextassembled)）
- [ ] 4.C TestDrainer_DebriefGenerateRegistered 通过（[test-plan §4.3](./test-plan.md#43-testdrainer_debriefgenerateregistered)）
- [ ] 4.D TestStoreUpdateDebriefCompleted_HappyTransaction 通过（[test-plan §4.4](./test-plan.md#44-teststoreupdatedebriefcompleted_happytransaction)）
- [ ] 4.E TestStoreUpdateDebriefCompleted_CASRejectsCompleted 通过（[test-plan §4.5](./test-plan.md#45-teststoreupdatedebriefcompleted_casrejectscompleted)）
- [ ] 4.F TestStoreUpdateDebriefCompleted_OutboxRollback 通过（[test-plan §4.6](./test-plan.md#46-teststoreupdatedebriefcompleted_outboxrollback)）
- [ ] 4.G TestGenerateHandler_OutboxPayloadSchema 通过（[test-plan §4.7](./test-plan.md#47-testgeneratehandler_outboxpayloadschema)）
- [ ] 4.H TestGenerateHandler_F3ResolveFailed 通过（[test-plan §4.8](./test-plan.md#48-testgeneratehandler_f3resolvefailed)）
- [ ] 4.I TestGenerateHandler_A3Timeout 通过（[test-plan §4.9](./test-plan.md#49-testgeneratehandler_a3timeout)）
- [ ] 4.J TestGenerateHandler_ParseEmpty 通过（[test-plan §4.10](./test-plan.md#410-testgeneratehandler_parseempty)）
- [ ] 4.K TestGenerateHandler_PermanentFailAt5Attempts 通过（[test-plan §4.11](./test-plan.md#411-testgeneratehandler_permanentfailat5attempts)）

## Phase 5: getDebrief read handler

- [ ] 5.A TestStoreGetDebrief_DraftPartial 通过（[test-plan §5.1](./test-plan.md#51-teststoregetdebrief_draftpartial)）
- [ ] 5.B TestStoreGetDebrief_CompletedFull 通过（[test-plan §5.2](./test-plan.md#52-teststoregetdebrief_completedfull)）
- [ ] 5.C TestStoreGetDebrief_CrossUserNotFound 通过（[test-plan §5.3](./test-plan.md#53-teststoregetdebrief_crossusernotfound)）
- [ ] 5.D TestServiceGetDebrief_ProvenanceWireOnly 通过（[test-plan §5.4](./test-plan.md#54-testservicegetdebrief_provenancewireonly)）
- [ ] 5.E TestGetDebrief_DraftResponse 通过（[test-plan §5.5](./test-plan.md#55-testgetdebrief_draftresponse)）
- [ ] 5.F TestGetDebrief_CompletedResponse 通过（[test-plan §5.6](./test-plan.md#56-testgetdebrief_completedresponse)）
- [ ] 5.G TestGetDebrief_CrossUser404 通过（[test-plan §5.7](./test-plan.md#57-testgetdebrief_crossuser404)）
- [ ] 5.H TestGetDebrief_NotFound404 通过（[test-plan §5.8](./test-plan.md#58-testgetdebrief_notfound404)）

## Phase 6: 隐私 / 观测 / retry / legacy negative

- [ ] 6.A TestOutboxPayload_NoRawText 通过（[test-plan §6.1](./test-plan.md#61-testoutboxpayload_norawtext)）
- [ ] 6.B TestAuditEvents_NoRawText 通过（[test-plan §6.2](./test-plan.md#62-testauditevents_norawtext)）
- [ ] 6.C TestAITaskRunsWritten 通过（[test-plan §6.3](./test-plan.md#63-testaitaskrunswritten)）
- [ ] 6.D TestAuditEventsWritten 通过（[test-plan §6.4](./test-plan.md#64-testauditeventswritten)）
- [ ] 6.E TestRetryPolicy_BackoffBelowMax 通过（[test-plan §6.5](./test-plan.md#65-testretrypolicy_backoffbelowmax)）
- [ ] 6.F TestRetryPolicy_PermanentFailAtMax 通过（[test-plan §6.6](./test-plan.md#66-testretrypolicy_permanentfailatmax)）
- [ ] 6.G Legacy negative grep gate 通过（plan checklist 6.4 + 6.5）

## Phase 7: 全计划单元/集成测试全量回归

- [ ] 7.A `cd backend && go test ./internal/debrief ./internal/api/debriefs ./internal/store/debrief ./cmd/api -count=1 -race` 通过
- [ ] 7.B `cd backend && go test ./... -count=1` 通过（全 backend 单元 + 集成测试）
- [ ] 7.C `python3 -m pytest scripts/lint -q` 通过（含 backend_debrief_legacy.py 与既有 lint）
- [ ] 7.D `make codegen-check` / `make validate-fixtures` / `make lint-events` / `make codegen-events-check` 通过
- [ ] 7.E `migrations/lint.sh` + `make migrate-check`（dev-stack .env）通过
- [ ] 7.F Phase 7 本计划定义的单元测试项全部通过
