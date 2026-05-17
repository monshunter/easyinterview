# 001 Debrief Screen and Handoff Test Checklist

> **版本**: 1.3
> **状态**: completed
> **更新日期**: 2026-05-17

**关联 Test Plan**: [test-plan](./test-plan.md)
**关联计划**: [plan](./plan.md)
**关联 Spec**: [spec](../../spec.md)

## 2026-05-17 L2 Close-out Evidence

- Review/fix red evidence: `pnpm --filter @easyinterview/frontend exec vitest run src/app/screens/debrief/DebriefScreen.test.tsx` initially failed on route/InterviewContext hydration, UI-language submit, immediate Step 1 advance, submit auth gate, and replay language handoff assertions.
- Focused green evidence: `pnpm --filter @easyinterview/frontend exec vitest run src/app/screens/debrief/DebriefScreen.test.tsx` passed (6 tests) after runtime fixes.
- BDD evidence: P0.065-P0.069 ran sequentially through `setup.sh -> trigger.sh -> verify.sh -> cleanup.sh` and passed; P0.069 includes frontend build, `tests/pixel-parity/debrief.spec.ts`, legacy lint, and scenario-tree legacy grep.
- Gate evidence: final close-out re-runs `make validate-fixtures`, `pnpm --filter @easyinterview/frontend typecheck`, `pnpm --filter @easyinterview/frontend lint`, `python3 -m pytest scripts/lint -q`, `make docs-check`, and `git diff --check`.
- Review-fix evidence: follow-up red tests covered invalid JD filter, resume picker asset-phase skip, missing manual fallback, missing `myAnswerSummary`, and replay CTA without a fresh session; focused green runs cover debrief Vitest regressions plus backend-practice list endpoint tests.

## Phase 0: 依赖验证

- [x] 0.A Generated client 含 `createDebrief` / `getDebrief` / `suggestDebriefQuestions`
- [x] 0.B Fixtures `Debriefs/createDebrief.json` / `getDebrief.json` / `suggestDebriefQuestions.json` 通过 `make validate-fixtures`
- [x] 0.C `frontend/src/lib/conventions/` 与 generated client 含 `DebriefRoundType` / `DebriefQuestionSource` / `DEBRIEF_NOT_FOUND` / `IDEMPOTENCY_KEY_MISMATCH`（原 `shared/ts/conventions/` 文案为旧路径漂移）
- [x] 0.D Generated client 含 `listPracticeSessions`（Phase 0 addendum）；`listResumes` / `listResumeVersions(resumeAssetId)` 可用；`PracticeSessions/listPracticeSessions.json` fixture 通过 `make validate-fixtures`
- [x] 0.E `pnpm --filter @easyinterview/frontend typecheck` 通过

## Phase 1: DebriefScreen shell + Header + ContextStrip + Stepper

- [x] 1.A TestDebriefScreen_DefaultRender 通过（[test-plan §1.1](./test-plan.md#11-testdebriefscreen_defaultrender)）
- [x] 1.A2 TestRoutes_DebriefAliasNormalization 通过（[test-plan §1.1b](./test-plan.md#11b-testroutes_debriefaliasnormalization)）
- [x] 1.B TestDebriefHeader_RenderWithContext 通过（[test-plan §1.2](./test-plan.md#12-testdebriefheader_renderwithcontext)）
- [x] 1.C TestDebriefHeader_FallbackOnMissingContext 通过（[test-plan §1.3](./test-plan.md#13-testdebriefheader_fallbackonmissingcontext)）
- [x] 1.D TestContextStrip_OpenPicker 通过（[test-plan §1.4](./test-plan.md#14-testcontextstrip_openpicker)）
- [x] 1.E TestContextStrip_DisplayNameFetch 通过（[test-plan §1.5](./test-plan.md#15-testcontextstrip_displaynamefetch)）
- [x] 1.F TestContextStrip_FallbackOnAPIError 通过（[test-plan §1.6](./test-plan.md#16-testcontextstrip_fallbackonapierror)）
- [x] 1.G TestStepper_NavigationLogic 通过（[test-plan §1.7](./test-plan.md#17-teststepper_navigationlogic)）

## Phase 2: 3 个 in-page picker modal

- [x] 2.A TestPickerModal_BaseInteraction 通过（[test-plan §2.1](./test-plan.md#21-testpickermodal_baseinteraction)）
- [x] 2.B TestJDPicker_ListAndConfirm 通过（[test-plan §2.2](./test-plan.md#22-testjdpicker_listandconfirm)）
- [x] 2.C TestMockSessionPicker_ListAndOptional 通过（[test-plan §2.3](./test-plan.md#23-testmocksessionpicker_listandoptional)）
- [x] 2.D TestMockSessionPicker_FilterFallback 通过（[test-plan §2.4](./test-plan.md#24-testmocksessionpicker_filterfallback)）
- [x] 2.E TestResumePicker_ListAndConfirm 通过（[test-plan §2.5](./test-plan.md#25-testresumepicker_listandconfirm)）
- [x] 2.F TestContextStrip_AutoTriggerSuggestions 通过（[test-plan §2.6](./test-plan.md#26-testcontextstrip_autotriggersuggestions)）

## Phase 3: Step 0 + 跨模式共享 entries + Voice UI shell

- [x] 3.A TestRecordSummaryBar_Counts 通过（[test-plan §3.1](./test-plan.md#31-testrecordsummarybar_counts)）
- [x] 3.B TestModeToggle_PreservesEntries 通过（[test-plan §3.2](./test-plan.md#32-testmodetoggle_preservesentries)）
- [x] 3.C TestGuidedRecord_OccurredCTA 通过（[test-plan §3.3](./test-plan.md#33-testguidedrecord_occurredcta)）
- [x] 3.D TestGuidedRecord_SkipCTA 通过（[test-plan §3.4](./test-plan.md#34-testguidedrecord_skipcta)）
- [x] 3.E TestGuidedRecord_EditCTA 通过（[test-plan §3.5](./test-plan.md#35-testguidedrecord_editcta)）
- [x] 3.F TestGuidedRecord_ManualAddCTA 通过（[test-plan §3.6](./test-plan.md#36-testguidedrecord_manualaddcta)）
- [x] 3.G TestVoiceRecord_UIShellOnly 通过（[test-plan §3.7](./test-plan.md#37-testvoicerecord_uishellonly)）
- [x] 3.H TestVoiceRecord_NoSTTBinding 通过（[test-plan §3.8](./test-plan.md#38-testvoicerecord_nosttbinding)）+ grep negative 验证
- [x] 3.I TestSubmitCTA_DisabledState 通过（[test-plan §3.9](./test-plan.md#39-testsubmitcta_disabledstate)）
- [x] 3.J TestSubmitCTA_EnabledAndClick 通过（[test-plan §3.10](./test-plan.md#310-testsubmitcta_enabledandclick)）

## Phase 4: suggestDebriefQuestions 集成

- [x] 4.A TestUseSuggestQuestions_AutoTrigger 通过（[test-plan §4.1](./test-plan.md#41-testusesuggestquestions_autotrigger)）
- [x] 4.B TestUseSuggestQuestions_Refetch 通过（[test-plan §4.2](./test-plan.md#42-testusesuggestquestions_refetch)）
- [x] 4.C TestUseSuggestQuestions_Debounce 通过（[test-plan §4.3](./test-plan.md#43-testusesuggestquestions_debounce)）
- [x] 4.D TestDebriefScreen_SuggestionsIntegration 通过（[test-plan §4.4](./test-plan.md#44-testdebriefscreen_suggestionsintegration)）
- [x] 4.E TestSuggestions_AIFailureDegradation 通过（[test-plan §4.5](./test-plan.md#45-testsuggestions_aifailuredegradation)）

## Phase 5: createDebrief + 双轨 polling + 失败态

- [x] 5.A TestUseSubmitDebrief_Happy202 通过（[test-plan §5.1](./test-plan.md#51-testusesubmitdebrief_happy202)）
- [x] 5.B TestUseSubmitDebrief_422ValidationFailed 通过（[test-plan §5.2](./test-plan.md#52-testusesubmitdebrief_422validationfailed)）
- [x] 5.C TestUseSubmitDebrief_409IKMismatchRetry 通过（[test-plan §5.3](./test-plan.md#53-testusesubmitdebrief_409ikmismatchretry)）
- [x] 5.D TestUseSubmitDebrief_401AuthGate 通过（[test-plan §5.4](./test-plan.md#54-testusesubmitdebrief_401authgate)）
- [x] 5.E TestUseDebriefPolling_HappySuccess 通过（[test-plan §5.5](./test-plan.md#55-testusedebriefpolling_happysuccess)）
- [x] 5.F TestUseDebriefPolling_JobFailed 通过（[test-plan §5.6](./test-plan.md#56-testusedebriefpolling_jobfailed)）
- [x] 5.G TestUseDebriefPolling_MaxAttemptsTimeout 通过（[test-plan §5.7](./test-plan.md#57-testusedebriefpolling_maxattemptstimeout)）
- [x] 5.H TestUseDebriefPolling_VisibilityPause 通过（[test-plan §5.8](./test-plan.md#58-testusedebriefpolling_visibilitypause)）
- [x] 5.I TestDebriefFailureState_Render 通过（[test-plan §5.9](./test-plan.md#59-testdebrieffailurestate_render)）
- [x] 5.J TestDebriefMissingContextState_Render 通过（[test-plan §5.10](./test-plan.md#510-testdebriefmissingcontextstate_render)）
- [x] 5.K TestDebriefTimeoutState_Render 通过（[test-plan §5.11](./test-plan.md#511-testdebrieftimeoutstate_render)）
- [x] 5.L TestInterviewContext_SetDebriefContext 通过（[test-plan §5.12](./test-plan.md#512-testinterviewcontext_setdebriefcontext)）
- [x] 5.M TestInterviewContext_DoesNotOverwriteJobId 通过（[test-plan §5.13](./test-plan.md#513-testinterviewcontext_doesnotoverwritejobid)）
- [x] 5.N TestPendingAction_DebriefParamsRoundTrip 通过（[test-plan §5.14](./test-plan.md#514-testpendingaction_debriefparamsroundtrip)）
- [x] 5.O TestInterviewContext_OtherActionsNotAffected 通过（[test-plan §5.15](./test-plan.md#515-testinterviewcontext_otheractionsnotaffected)）

## Phase 6: Step 1 分析渲染 + Step 2 复盘面试 launcher + handoff

- [x] 6.A TestStep1Analysis_RiskItemsRender 通过（[test-plan §6.1](./test-plan.md#61-teststep1analysis_riskitemsrender)）
- [x] 6.B TestStep1Analysis_DimensionsRender 通过（[test-plan §6.2](./test-plan.md#62-teststep1analysis_dimensionsrender)）
- [x] 6.C TestStep1Analysis_ProvenanceExpansion 通过（[test-plan §6.3](./test-plan.md#63-teststep1analysis_provenanceexpansion)）
- [x] 6.D TestStep1Analysis_P1FieldsNotRendered 通过（[test-plan §6.4](./test-plan.md#64-teststep1analysis_p1fieldsnotrendered)）
- [x] 6.E TestStep2Launcher_ContentPreview 通过（[test-plan §6.5](./test-plan.md#65-teststep2launcher_contentpreview)）
- [x] 6.F TestStartDebriefInterview_NavPayload 通过（[test-plan §6.6](./test-plan.md#66-teststartdebriefinterview_navpayload)）
- [x] 6.G TestStartDebriefInterview_AuthGate 通过（[test-plan §6.7](./test-plan.md#67-teststartdebriefinterview_authgate)）
- [x] 6.H TestStartDebriefInterview_CreatesFreshPracticeSession 通过（[test-plan §6.8](./test-plan.md#68-teststartdebriefinterview_createsfreshpracticesession)）

## Phase 7: i18n + 主题 + 响应式

- [x] 7.A TestI18n_DebriefNamespaceComplete 通过（[test-plan §7.1](./test-plan.md#71-testi18n_debriefnamespacecomplete)）
- [x] 7.B TestTheme_DebriefScreen 通过（[test-plan §7.2](./test-plan.md#72-testtheme_debriefscreen)）
- [x] 7.C TestResponsive_Mobile 通过（[test-plan §7.3](./test-plan.md#73-testresponsive_mobile)）

## Phase 8: Pixel parity + 隐私 + Legacy negative

- [x] 8.A Debrief Playwright pixel parity desktop project 通过：DOM anchors、desktop viewport bounding boxes、theme/customAccent computed values 与非空 screenshot smoke 均通过
- [x] 8.B Debrief Playwright pixel parity mobile project 通过：mobile viewport bounding boxes、horizontal overflow negative 与非空 screenshot smoke 均通过
- [x] 8.C 主题 pixel parity（light / dark / customAccent）通过：`tests/pixel-parity/debrief.spec.ts` 覆盖 dark mode 与 customAccent computed values
- [x] 8.D TestPrivacy_NoRawTextInLocalStorage 通过（[test-plan §8.1](./test-plan.md#81-testprivacy_norawtextinlocalstorage)）
- [x] 8.E TestPrivacy_NoRawTextInConsoleLog 通过（[test-plan §8.2](./test-plan.md#82-testprivacy_norawtextinconsolelog)）
- [x] 8.F 隐私 grep gate 通过（plan checklist 8.5）
- [x] 8.G Legacy negative grep gate 通过（plan checklist 8.6）
- [x] 8.H `scripts/lint/frontend_debrief_legacy.py` 通过 `python3 -m pytest scripts/lint -q`
- [x] 8.I BDD scenarios P0.065-069 全部通过（详见 [bdd-checklist.md](./bdd-checklist.md)）

## Phase 9: 全计划单元/集成测试全量回归

- [x] 9.A `pnpm --filter @easyinterview/frontend test -- src/app/screens/debrief --run` 通过
- [x] 9.B `pnpm --filter @easyinterview/frontend test -- --run` 通过（全 frontend 单元测试）
- [x] 9.C `pnpm --filter @easyinterview/frontend lint` 通过
- [x] 9.D `pnpm --filter @easyinterview/frontend exec playwright test tests/pixel-parity/debrief.spec.ts`（debrief Playwright gate）通过
- [x] 9.E `python3 -m pytest scripts/lint -q` 通过
- [x] 9.F `make docs-check` + `git diff --check` 通过
- [x] 9.G Phase 9 本计划定义的单元测试项全部通过
