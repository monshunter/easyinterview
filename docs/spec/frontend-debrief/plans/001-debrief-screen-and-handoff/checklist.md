# 001 Debrief Screen and Handoff Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-16

**关联计划**: [plan](./plan.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 Test Checklist**: [test-checklist](./test-checklist.md)

## Phase 0: 依赖验证 + ui-design source map + 包结构

- [ ] 0.1 backend-debrief Phase 0 完成验证：`grep -rn "suggestDebriefQuestions\|createDebrief\|getDebrief" frontend/src/api/generated/` 命中 3 个 method；`ls openapi/fixtures/Debriefs/` 含 `createDebrief.json` / `getDebrief.json` / `suggestDebriefQuestions.json` 三个 fixture；`grep -rn "DebriefRoundType\|DebriefQuestionSource\|DEBRIEF_NOT_FOUND\|IDEMPOTENCY_KEY_MISMATCH" shared/ts/conventions/` 命中；backend-practice 现状已支持 `goal='debrief'` + `mode='debrief'`（验证证据已在 backend-debrief/001 Phase 0.5 记录）
- [ ] 0.2 ui-design source map 记录到 plan history 与本 checklist 注脚（6 个组件 source anchor）
- [ ] 0.3 创建包结构 `frontend/src/app/screens/debrief/{DebriefScreen.tsx, components/, hooks/, reducer.ts, types.ts, i18n/}`；空 stub 编译通过：`pnpm --filter @easyinterview/frontend typecheck`
- [ ] 0.4 route 接线：`App.tsx` 的 `case "debrief"` → `<DebriefScreen>`；`normalizeRoute.ts` 把历史 alias `debrief_full` normalize 到 `debrief`；不在 `routes.ts` 正式 `RouteName` / primary nav / `INTERVIEW_CONTEXT_ROUTES` 中新增 `debrief_full`；移除原 PlaceholderScreen 对 debrief 的占位；TopBar 一级导航高亮逻辑保留；测试：`TestRoutes_DebriefAliasNormalization` 通过

## Phase 1: DebriefScreen shell + Header + ContextStrip + Stepper

- [ ] 1.1 `<DebriefScreen>` container：接收 InterviewContext；internal state 初始化（step=0, inputMode='text', entries=[], selectedContext=null, pickerType=null, suggestions=null）；render Header + ContextStrip + Stepper + step panel；测试：`TestDebriefScreen_DefaultRender`（[test-plan §1.1](./test-plan.md#11-testdebriefscreen_defaultrender)）
- [ ] 1.2 `<DebriefHeader>`：复刻 prototype lines 122-144；从 InterviewContext 派生 eyebrow；缺失数据 fallback；测试：`TestDebriefHeader_RenderWithContext`（[test-plan §1.2](./test-plan.md#12-testdebriefheader_renderwithcontext)）+ `TestDebriefHeader_FallbackOnMissingContext`（[test-plan §1.3](./test-plan.md#13-testdebriefheader_fallbackonmissingcontext)）
- [ ] 1.3 `<DebriefContextStrip>`：复刻 lines 412-432；三卡片点击 → setPickerType；通过 `getTargetJob` / `getResumeVersion` / `getPracticeSession` 拉 display name；失败 fallback 显示 ID；测试：`TestContextStrip_OpenPicker`（[test-plan §1.4](./test-plan.md#14-testcontextstrip_openpicker)）+ `TestContextStrip_DisplayNameFetch`（[test-plan §1.5](./test-plan.md#15-testcontextstrip_displaynamefetch)）+ `TestContextStrip_FallbackOnAPIError`（[test-plan §1.6](./test-plan.md#16-testcontextstrip_fallbackonapierror)）
- [ ] 1.4 `<DebriefStepper>`：复刻 lines 148-156；3 步骤；当前 step 高亮；点击已访问 step 返回；测试：`TestStepper_NavigationLogic`（[test-plan §1.7](./test-plan.md#17-teststepper_navigationlogic)）

## Phase 2: 3 个 in-page picker modal

- [ ] 2.1 `<DebriefContextPickerModal>` 通用骨架：复刻 lines 434-518；接收 kind / options / selectedId / onClose / onConfirm；外部点击 / Esc 关闭；mobile 全屏 sheet；测试：`TestPickerModal_BaseInteraction`（[test-plan §2.1](./test-plan.md#21-testpickermodal_baseinteraction)）
- [ ] 2.2 JD picker：调 `listTargetJobs({status:'ready'})`；单选；onConfirm 触发 `SET_DEBRIEF_CONTEXT` reducer action；测试：`TestJDPicker_ListAndConfirm`（[test-plan §2.2](./test-plan.md#22-testjdpicker_listandconfirm)）
- [ ] 2.3 Mock Session picker：调 `listPracticeSessions({targetJobId, status:'completed'})`；如 server-side filter 不支持 client-side filter；"暂不关联" option；测试：`TestMockSessionPicker_ListAndOptional`（[test-plan §2.3](./test-plan.md#23-testmocksessionpicker_listandoptional)）+ `TestMockSessionPicker_FilterFallback`（[test-plan §2.4](./test-plan.md#24-testmocksessionpicker_filterfallback)）
- [ ] 2.4 Resume picker：调 `listResumeVersions({status:'ready'})`；单选；测试：`TestResumePicker_ListAndConfirm`（[test-plan §2.5](./test-plan.md#25-testresumepicker_listandconfirm)）
- [ ] 2.5 ContextStrip 三选完成 detect + 自动触发 suggestions：useEffect debounce 500ms；当 targetJob + resume 都选定（mockSession optional）时 enable suggestions hook；测试：`TestContextStrip_AutoTriggerSuggestions`（[test-plan §2.6](./test-plan.md#26-testcontextstrip_autotriggersuggestions)）

## Phase 3: Step 0 复盘记录 + 跨模式共享 entries + Voice UI shell

- [ ] 3.1 顶部统一汇总条：复刻 lines 162-185；total + source chips；测试：`TestRecordSummaryBar_Counts`（[test-plan §3.1](./test-plan.md#31-testrecordsummarybar_counts)）
- [ ] 3.2 Mode toggle：复刻 lines 187-210；切换不丢 entries；测试：`TestModeToggle_PreservesEntries`（[test-plan §3.2](./test-plan.md#32-testmodetoggle_preservesentries)）
- [ ] 3.3 `<GuidedDebriefRecord>`：复刻 lines 519-619；接收 suggestions + entries；4 个 CTA（occurred / skipped / edit / manual）；entries 写入 with source；测试：`TestGuidedRecord_OccurredCTA`（[test-plan §3.3](./test-plan.md#33-testguidedrecord_occurredcta)）+ `TestGuidedRecord_SkipCTA`（[test-plan §3.4](./test-plan.md#34-testguidedrecord_skipcta)）+ `TestGuidedRecord_EditCTA`（[test-plan §3.5](./test-plan.md#35-testguidedrecord_editcta)）+ `TestGuidedRecord_ManualAddCTA`（[test-plan §3.6](./test-plan.md#36-testguidedrecord_manualaddcta)）
- [ ] 3.4 `<VoiceDebriefRecord>` UI shell：复刻 lines 656-870 视觉；toggle highlight + idle 占位 + 待确认卡片列表 + 「空格暂停/继续」提示文案；显示 "语音复盘集成中，敬请期待" 占位；**不**绑定真实 Web Audio API；测试：`TestVoiceRecord_UIShellOnly`（[test-plan §3.7](./test-plan.md#37-testvoicerecord_uishellonly)） + `TestVoiceRecord_NoSTTBinding`（[test-plan §3.8](./test-plan.md#38-testvoicerecord_nosttbinding)）
- [ ] 3.5 Submit CTA：复刻 lines 314-316；disabled 条件 entries.length===0 或 targetJob===null；点击触发 Phase 5 createDebrief；测试：`TestSubmitCTA_DisabledState`（[test-plan §3.9](./test-plan.md#39-testsubmitcta_disabledstate)） + `TestSubmitCTA_EnabledAndClick`（[test-plan §3.10](./test-plan.md#310-testsubmitcta_enabledandclick)）

## Phase 4: suggestDebriefQuestions 集成

- [ ] 4.1 `useSuggestDebriefQuestions` hook 实现：接收 targetJobId/sessionId?/resumeVersionId?/language/count/enabled；返回 suggestions/loading/error/refetch；自动 debounce 500ms；测试：`TestUseSuggestQuestions_AutoTrigger`（[test-plan §4.1](./test-plan.md#41-testusesuggestquestions_autotrigger)）+ `TestUseSuggestQuestions_Refetch`（[test-plan §4.2](./test-plan.md#42-testusesuggestquestions_refetch)）+ `TestUseSuggestQuestions_Debounce`（[test-plan §4.3](./test-plan.md#43-testusesuggestquestions_debounce)）
- [ ] 4.2 DebriefScreen 整合：ContextStrip 三选完成后 enable hook；GuidedDebriefRecord 渲染 suggestions；用户可点击 "重新生成推荐" 调 refetch；测试：`TestDebriefScreen_SuggestionsIntegration`（[test-plan §4.4](./test-plan.md#44-testdebriefscreen_suggestionsintegration)）
- [ ] 4.3 失败降级：error.code 为 B1 canonical `AI_PROVIDER_CONFIG_INVALID` / `AI_PROVIDER_SECRET_MISSING` / `AI_PROVIDER_TIMEOUT` / `AI_OUTPUT_INVALID` / `AI_FALLBACK_EXHAUSTED` → 显示 inline error；不阻塞 step 0；"重新生成推荐" 按钮启用；测试：`TestSuggestions_AIFailureDegradation`（[test-plan §4.5](./test-plan.md#45-testsuggestions_aifailuredegradation)）

## Phase 5: createDebrief + 双轨 polling + 失败态

- [ ] 5.1 `useSubmitDebrief` hook：接收 payload；生成 IK UUIDv4；调 generated `createDebrief`；handle 4 类响应（202 / 422 / 409 / 401 / 5xx）；202 后写 `debriefId` + `debriefJobId`，不得写现有 `jobId`；测试：`TestUseSubmitDebrief_Happy202`（[test-plan §5.1](./test-plan.md#51-testusesubmitdebrief_happy202)）+ `TestUseSubmitDebrief_422ValidationFailed`（[test-plan §5.2](./test-plan.md#52-testusesubmitdebrief_422validationfailed)）+ `TestUseSubmitDebrief_409IKMismatchRetry`（[test-plan §5.3](./test-plan.md#53-testusesubmitdebrief_409ikmismatchretry)）+ `TestUseSubmitDebrief_401AuthGate`（[test-plan §5.4](./test-plan.md#54-testusesubmitdebrief_401authgate)）
- [ ] 5.2 `useDebriefPolling` hook：双轨 polling `getJob(debriefJobId)` + getDebrief；指数退避；visibility/focus 暂停-恢复；pollingState 状态机；测试：`TestUseDebriefPolling_HappySuccess`（[test-plan §5.5](./test-plan.md#55-testusedebriefpolling_happysuccess)）+ `TestUseDebriefPolling_JobFailed`（[test-plan §5.6](./test-plan.md#56-testusedebriefpolling_jobfailed)）+ `TestUseDebriefPolling_MaxAttemptsTimeout`（[test-plan §5.7](./test-plan.md#57-testusedebriefpolling_maxattemptstimeout)）+ `TestUseDebriefPolling_VisibilityPause`（[test-plan §5.8](./test-plan.md#58-testusedebriefpolling_visibilitypause)）
- [ ] 5.3 `<DebriefFailureState>`：失败卡片 + errorCode 映射 + CTA 重试 / 返回；测试：`TestDebriefFailureState_Render`（[test-plan §5.9](./test-plan.md#59-testdebrieffailurestate_render)）
- [ ] 5.4 `<DebriefMissingContextState>`：缺 targetJobId；CTA 自动开 JD picker；测试：`TestDebriefMissingContextState_Render`（[test-plan §5.10](./test-plan.md#510-testdebriefmissingcontextstate_render)）
- [ ] 5.5 `<DebriefTimeoutState>`：polling 超时；CTA 重试 / 返回；测试：`TestDebriefTimeoutState_Render`（[test-plan §5.11](./test-plan.md#511-testdebrieftimeoutstate_render)）
- [ ] 5.6 InterviewContext reducer 扩展 `SET_DEBRIEF_CONTEXT` action；写 `debriefId` / `debriefJobId` / `practiceGoal?`，不得写现有 `jobId`；扩展 `PENDING_ACTION_INTERVIEW_KEYS` 覆盖 `practiceGoal` / `debriefId` / `debriefJobId`；不破坏既有 action；测试：`TestInterviewContext_SetDebriefContext`（[test-plan §5.12](./test-plan.md#512-testinterviewcontext_setdebriefcontext)）+ `TestInterviewContext_DoesNotOverwriteJobId`（[test-plan §5.13](./test-plan.md#513-testinterviewcontext_doesnotoverwritejobid)）+ `TestPendingAction_DebriefParamsRoundTrip`（[test-plan §5.14](./test-plan.md#514-testpendingaction_debriefparamsroundtrip)）+ `TestInterviewContext_OtherActionsNotAffected`（[test-plan §5.15](./test-plan.md#515-testinterviewcontext_otheractionsnotaffected)）

## Phase 6: Step 1 分析渲染 + Step 2 复盘面试 launcher + handoff

- [ ] 6.1 Step 1 分析渲染：风险项列表 + 维度对比卡片 3 张 + provenance 展开区 6 字段；不渲染 nextRoundChecklist / thankYouDraft；测试：`TestStep1Analysis_RiskItemsRender`（[test-plan §6.1](./test-plan.md#61-teststep1analysis_riskitemsrender)）+ `TestStep1Analysis_DimensionsRender`（[test-plan §6.2](./test-plan.md#62-teststep1analysis_dimensionsrender)）+ `TestStep1Analysis_ProvenanceExpansion`（[test-plan §6.3](./test-plan.md#63-teststep1analysis_provenanceexpansion)）+ `TestStep1Analysis_P1FieldsNotRendered`（[test-plan §6.4](./test-plan.md#64-teststep1analysis_p1fieldsnotrendered)）
- [ ] 6.2 Step 2 复盘面试 launcher：复刻 `<DebriefReplayPlan>` lines 1388-1421；内容预览从 debrief.questions + riskItems 派生；测试：`TestStep2Launcher_ContentPreview`（[test-plan §6.5](./test-plan.md#65-teststep2launcher_contentpreview)）
- [ ] 6.3 "开始复盘面试" CTA：nav practice with payload；未登录走 useRequestAuth；不调 createPracticePlan/startPracticeSession；测试：`TestStartDebriefInterview_NavPayload`（[test-plan §6.6](./test-plan.md#66-teststartdebriefinterview_navpayload)）+ `TestStartDebriefInterview_AuthGate`（[test-plan §6.7](./test-plan.md#67-teststartdebriefinterview_authgate)）+ `TestStartDebriefInterview_NoCreatePracticePlanCall`（[test-plan §6.8](./test-plan.md#68-teststartdebriefinterview_nocreatepracticeplancall)）

## Phase 7: i18n + 主题 + 响应式

- [ ] 7.1 i18n `debrief.*` namespace：新增 zh.ts / en.ts keys 完整（header / contextStrip / stepper / step0-2 / pickers / failureStates / suggestions / voice）；测试：`TestI18n_DebriefNamespaceComplete`（[test-plan §7.1](./test-plan.md#71-testi18n_debriefnamespacecomplete)）
- [ ] 7.2 主题适配：dark / customAccent 在 DebriefScreen 各 step + picker modal 中正常；Vitest + Playwright 验证 root `data-theme` 应用；测试：`TestTheme_DebriefScreen`（[test-plan §7.2](./test-plan.md#72-testtheme_debriefscreen)）
- [ ] 7.3 Mobile 响应式：viewport 390×844 测试 Header / ContextStrip / Stepper / Step 0 双栏折叠 / Step 1 单列 / Step 2 sticky CTA / picker 全屏 sheet；测试：`TestResponsive_Mobile`（[test-plan §7.3](./test-plan.md#73-testresponsive_mobile)）

## Phase 8: Playwright pixel parity + 隐私 + legacy negative + BDD

- [ ] 8.1 Playwright pixel parity desktop 1440×900：`frontend/tests/pixel-parity/debrief-desktop.spec.ts` 通过；diff < 0.5%
- [ ] 8.2 Playwright pixel parity mobile 390×844：`frontend/tests/pixel-parity/debrief-mobile.spec.ts` 通过；diff < 0.5%
- [ ] 8.3 主题 pixel parity：light / dark / customAccent 各跑一次
- [ ] 8.4 隐私 + telemetry 验证：Vitest fixture spy 注入 marker；submit 后 spy 接收 raw body 但 console.log / localStorage / sessionStorage / telemetry 不写；测试：`TestPrivacy_NoRawTextInLocalStorage`（[test-plan §8.1](./test-plan.md#81-testprivacy_norawtextinlocalstorage)）+ `TestPrivacy_NoRawTextInConsoleLog`（[test-plan §8.2](./test-plan.md#82-testprivacy_norawtextinconsolelog)）
- [ ] 8.5 隐私 grep gate：`grep -rn "questionText\|myAnswerSummary\|interviewerReaction\|notes" frontend/src/app/screens/debrief/ frontend/src/app/i18n/locales/ | grep -v "_test\|generated\|.types\|// privacy reviewed"` 仅命中合理位置
- [ ] 8.6 Legacy negative grep：`grep -rn "experience_library\|star_editor\|drill_builder\|mistakes_book\|growth_center\|report_timeline" frontend/src/app/screens/debrief/ frontend/src/app/i18n/locales/ test/scenarios/e2e/p0-06[56789]-*` 0 命中
- [ ] 8.7 Legacy negative lint script：新增 `scripts/lint/frontend_debrief_legacy.py`；`python3 -m pytest scripts/lint -q` 通过
- [ ] 8.8 BDD-Gate E2E.P0.065：`bash test/scenarios/e2e/p0-065-debrief-default-render-and-pickers/run.sh` 通过
- [ ] 8.9 BDD-Gate E2E.P0.066：`bash test/scenarios/e2e/p0-066-debrief-text-suggestions-and-submit/run.sh` 通过
- [ ] 8.10 BDD-Gate E2E.P0.067：`bash test/scenarios/e2e/p0-067-debrief-polling-happy-and-analysis/run.sh` 通过
- [ ] 8.11 BDD-Gate E2E.P0.068：`bash test/scenarios/e2e/p0-068-debrief-failure-and-handoff/run.sh` 通过
- [ ] 8.12 BDD-Gate E2E.P0.069：`bash test/scenarios/e2e/p0-069-debrief-pixel-parity-and-legacy-negative/run.sh` 通过

## Phase 9: Plan 收口

- [ ] 9.1 全局回归：`pnpm --filter @easyinterview/frontend test -- --run` / `pnpm --filter @easyinterview/frontend lint` / `pnpm --filter @easyinterview/frontend test:pixel-parity` / `python3 -m pytest scripts/lint -q` / `make docs-check` / `git diff --check` 全部通过
- [ ] 9.2 plans/INDEX.md 把 001 从 active 移到 completed，记录完成日期 2026-MM-DD
- [ ] 9.3 frontend-debrief/history.md 增加 1.1 completion 行
- [ ] 9.4 提交 commit `feat(frontend-debrief): close 001 debrief screen and handoff baseline`；记录工作日志 `/work-journal`
