# 001 Debrief Screen and Handoff Test Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-16

**关联 Test Plan**: [test-plan](./test-plan.md)
**关联计划**: [plan](./plan.md)
**关联 Spec**: [spec](../../spec.md)

## Phase 0: 依赖验证

- [ ] 0.A Generated client 含 `createDebrief` / `getDebrief` / `suggestDebriefQuestions`
- [ ] 0.B Fixtures `Debriefs/createDebrief.json` / `getDebrief.json` / `suggestDebriefQuestions.json` 通过 `make validate-fixtures`
- [ ] 0.C `shared/ts/conventions/` 含 `DebriefRoundType` / `DebriefQuestionSource` / `DEBRIEF_NOT_FOUND`
- [ ] 0.D `pnpm --filter frontend tsc -- --noEmit` 通过

## Phase 1: DebriefScreen shell + Header + ContextStrip + Stepper

- [ ] 1.A TestDebriefScreen_DefaultRender 通过（[test-plan §1.1](./test-plan.md#11-testdebriefscreen_defaultrender)）
- [ ] 1.B TestDebriefHeader_RenderWithContext 通过（[test-plan §1.2](./test-plan.md#12-testdebriefheader_renderwithcontext)）
- [ ] 1.C TestDebriefHeader_FallbackOnMissingContext 通过（[test-plan §1.3](./test-plan.md#13-testdebriefheader_fallbackonmissingcontext)）
- [ ] 1.D TestContextStrip_OpenPicker 通过（[test-plan §1.4](./test-plan.md#14-testcontextstrip_openpicker)）
- [ ] 1.E TestContextStrip_DisplayNameFetch 通过（[test-plan §1.5](./test-plan.md#15-testcontextstrip_displaynamefetch)）
- [ ] 1.F TestContextStrip_FallbackOnAPIError 通过（[test-plan §1.6](./test-plan.md#16-testcontextstrip_fallbackonapierror)）
- [ ] 1.G TestStepper_NavigationLogic 通过（[test-plan §1.7](./test-plan.md#17-teststepper_navigationlogic)）

## Phase 2: 3 个 in-page picker modal

- [ ] 2.A TestPickerModal_BaseInteraction 通过（[test-plan §2.1](./test-plan.md#21-testpickermodal_baseinteraction)）
- [ ] 2.B TestJDPicker_ListAndConfirm 通过（[test-plan §2.2](./test-plan.md#22-testjdpicker_listandconfirm)）
- [ ] 2.C TestMockSessionPicker_ListAndOptional 通过（[test-plan §2.3](./test-plan.md#23-testmocksessionpicker_listandoptional)）
- [ ] 2.D TestMockSessionPicker_FilterFallback 通过（[test-plan §2.4](./test-plan.md#24-testmocksessionpicker_filterfallback)）
- [ ] 2.E TestResumePicker_ListAndConfirm 通过（[test-plan §2.5](./test-plan.md#25-testresumepicker_listandconfirm)）
- [ ] 2.F TestContextStrip_AutoTriggerSuggestions 通过（[test-plan §2.6](./test-plan.md#26-testcontextstrip_autotriggersuggestions)）

## Phase 3: Step 0 + 跨模式共享 entries + Voice UI shell

- [ ] 3.A TestRecordSummaryBar_Counts 通过（[test-plan §3.1](./test-plan.md#31-testrecordsummarybar_counts)）
- [ ] 3.B TestModeToggle_PreservesEntries 通过（[test-plan §3.2](./test-plan.md#32-testmodetoggle_preservesentries)）
- [ ] 3.C TestGuidedRecord_OccurredCTA 通过（[test-plan §3.3](./test-plan.md#33-testguidedrecord_occurredcta)）
- [ ] 3.D TestGuidedRecord_SkipCTA 通过（[test-plan §3.4](./test-plan.md#34-testguidedrecord_skipcta)）
- [ ] 3.E TestGuidedRecord_EditCTA 通过（[test-plan §3.5](./test-plan.md#35-testguidedrecord_editcta)）
- [ ] 3.F TestGuidedRecord_ManualAddCTA 通过（[test-plan §3.6](./test-plan.md#36-testguidedrecord_manualaddcta)）
- [ ] 3.G TestVoiceRecord_UIShellOnly 通过（[test-plan §3.7](./test-plan.md#37-testvoicerecord_uishellonly)）
- [ ] 3.H TestVoiceRecord_NoSTTBinding 通过（[test-plan §3.8](./test-plan.md#38-testvoicerecord_nosttbinding)）+ grep negative 验证
- [ ] 3.I TestSubmitCTA_DisabledState 通过（[test-plan §3.9](./test-plan.md#39-testsubmitcta_disabledstate)）
- [ ] 3.J TestSubmitCTA_EnabledAndClick 通过（[test-plan §3.10](./test-plan.md#310-testsubmitcta_enabledandclick)）

## Phase 4: suggestDebriefQuestions 集成

- [ ] 4.A TestUseSuggestQuestions_AutoTrigger 通过（[test-plan §4.1](./test-plan.md#41-testusesuggestquestions_autotrigger)）
- [ ] 4.B TestUseSuggestQuestions_Refetch 通过（[test-plan §4.2](./test-plan.md#42-testusesuggestquestions_refetch)）
- [ ] 4.C TestUseSuggestQuestions_Debounce 通过（[test-plan §4.3](./test-plan.md#43-testusesuggestquestions_debounce)）
- [ ] 4.D TestDebriefScreen_SuggestionsIntegration 通过（[test-plan §4.4](./test-plan.md#44-testdebriefscreen_suggestionsintegration)）
- [ ] 4.E TestSuggestions_AIFailureDegradation 通过（[test-plan §4.5](./test-plan.md#45-testsuggestions_aifailuredegradation)）

## Phase 5: createDebrief + 双轨 polling + 失败态

- [ ] 5.A TestUseSubmitDebrief_Happy202 通过（[test-plan §5.1](./test-plan.md#51-testusesubmitdebrief_happy202)）
- [ ] 5.B TestUseSubmitDebrief_400ValidationError 通过（[test-plan §5.2](./test-plan.md#52-testusesubmitdebrief_400validationerror)）
- [ ] 5.C TestUseSubmitDebrief_409IKMismatchRetry 通过（[test-plan §5.3](./test-plan.md#53-testusesubmitdebrief_409ikmismatchretry)）
- [ ] 5.D TestUseSubmitDebrief_401AuthGate 通过（[test-plan §5.4](./test-plan.md#54-testusesubmitdebrief_401authgate)）
- [ ] 5.E TestUseDebriefPolling_HappySuccess 通过（[test-plan §5.5](./test-plan.md#55-testusedebriefpolling_happysuccess)）
- [ ] 5.F TestUseDebriefPolling_JobFailed 通过（[test-plan §5.6](./test-plan.md#56-testusedebriefpolling_jobfailed)）
- [ ] 5.G TestUseDebriefPolling_MaxAttemptsTimeout 通过（[test-plan §5.7](./test-plan.md#57-testusedebriefpolling_maxattemptstimeout)）
- [ ] 5.H TestUseDebriefPolling_VisibilityPause 通过（[test-plan §5.8](./test-plan.md#58-testusedebriefpolling_visibilitypause)）
- [ ] 5.I TestDebriefFailureState_Render 通过（[test-plan §5.9](./test-plan.md#59-testdebrieffailurestate_render)）
- [ ] 5.J TestDebriefMissingContextState_Render 通过（[test-plan §5.10](./test-plan.md#510-testdebriefmissingcontextstate_render)）
- [ ] 5.K TestDebriefTimeoutState_Render 通过（[test-plan §5.11](./test-plan.md#511-testdebrieftimeoutstate_render)）
- [ ] 5.L TestInterviewContext_SetDebriefContext 通过（[test-plan §5.12](./test-plan.md#512-testinterviewcontext_setdebriefcontext)）
- [ ] 5.M TestInterviewContext_OtherActionsNotAffected 通过（[test-plan §5.13](./test-plan.md#513-testinterviewcontext_otheractionsnotaffected)）

## Phase 6: Step 1 分析渲染 + Step 2 复盘面试 launcher + handoff

- [ ] 6.A TestStep1Analysis_RiskItemsRender 通过（[test-plan §6.1](./test-plan.md#61-teststep1analysis_riskitemsrender)）
- [ ] 6.B TestStep1Analysis_DimensionsRender 通过（[test-plan §6.2](./test-plan.md#62-teststep1analysis_dimensionsrender)）
- [ ] 6.C TestStep1Analysis_ProvenanceExpansion 通过（[test-plan §6.3](./test-plan.md#63-teststep1analysis_provenanceexpansion)）
- [ ] 6.D TestStep1Analysis_P1FieldsNotRendered 通过（[test-plan §6.4](./test-plan.md#64-teststep1analysis_p1fieldsnotrendered)）
- [ ] 6.E TestStep2Launcher_ContentPreview 通过（[test-plan §6.5](./test-plan.md#65-teststep2launcher_contentpreview)）
- [ ] 6.F TestStartDebriefInterview_NavPayload 通过（[test-plan §6.6](./test-plan.md#66-teststartdebriefinterview_navpayload)）
- [ ] 6.G TestStartDebriefInterview_AuthGate 通过（[test-plan §6.7](./test-plan.md#67-teststartdebriefinterview_authgate)）
- [ ] 6.H TestStartDebriefInterview_NoCreatePracticePlanCall 通过（[test-plan §6.8](./test-plan.md#68-teststartdebriefinterview_nocreatepracticeplancall)）

## Phase 7: i18n + 主题 + 响应式

- [ ] 7.A TestI18n_DebriefNamespaceComplete 通过（[test-plan §7.1](./test-plan.md#71-testi18n_debriefnamespacecomplete)）
- [ ] 7.B TestTheme_DebriefScreen 通过（[test-plan §7.2](./test-plan.md#72-testtheme_debriefscreen)）
- [ ] 7.C TestResponsive_Mobile 通过（[test-plan §7.3](./test-plan.md#73-testresponsive_mobile)）

## Phase 8: Pixel parity + 隐私 + Legacy negative

- [ ] 8.A Playwright pixel parity desktop 1440×900 通过；diff < 0.5%
- [ ] 8.B Playwright pixel parity mobile 390×844 通过；diff < 0.5%
- [ ] 8.C 主题 pixel parity（light / dark / customAccent）通过
- [ ] 8.D TestPrivacy_NoRawTextInLocalStorage 通过（[test-plan §8.1](./test-plan.md#81-testprivacy_norawtextinlocalstorage)）
- [ ] 8.E TestPrivacy_NoRawTextInConsoleLog 通过（[test-plan §8.2](./test-plan.md#82-testprivacy_norawtextinconsolelog)）
- [ ] 8.F 隐私 grep gate 通过（plan checklist 8.5）
- [ ] 8.G Legacy negative grep gate 通过（plan checklist 8.6）
- [ ] 8.H `scripts/lint/frontend_debrief_legacy.py` 通过 `python3 -m pytest scripts/lint -q`
- [ ] 8.I BDD scenarios P0.065-069 全部通过（详见 [bdd-checklist.md](./bdd-checklist.md)）

## Phase 9: 全计划单元/集成测试全量回归

- [ ] 9.A `pnpm --filter frontend test -- src/app/screens/debrief --run` 通过
- [ ] 9.B `pnpm --filter frontend test -- --run` 通过（全 frontend 单元测试）
- [ ] 9.C `pnpm --filter frontend lint` 通过
- [ ] 9.D `pnpm --filter frontend run pixel-parity`（Playwright）通过
- [ ] 9.E `python3 -m pytest scripts/lint -q` 通过
- [ ] 9.F `make docs-check` + `git diff --check` 通过
- [ ] 9.G Phase 9 本计划定义的单元测试项全部通过
