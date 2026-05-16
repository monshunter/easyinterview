# 001 Debrief Screen and Handoff Test Plan

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-16

**关联计划**: [plan](./plan.md)
**关联 Spec**: [spec](../../spec.md)

## 0 目标

为 frontend-debrief/001-debrief-screen-and-handoff 定义单元测试（Vitest + jsdom）与端到端 UI 测试（Playwright）矩阵。每个测试项映射到具体测试文件与测试函数，覆盖主路径、跨模式状态、AI 推荐、轮询、失败态、UI source/visual parity、隐私、回归负向 8 类风险。

测试执行入口：
- Vitest 单元测试：`pnpm --filter @easyinterview/frontend test -- src/app/screens/debrief --run`
- Vitest 全量：`pnpm --filter @easyinterview/frontend test -- --run`
- Lint：`pnpm --filter @easyinterview/frontend lint`
- Playwright pixel parity：`pnpm --filter @easyinterview/frontend test:pixel-parity`
- Lint script：`python3 -m pytest scripts/lint -q`

## 1 Coverage Matrix

| 行 | source | category | plan phase | verification | negative_scope | ui_source_anchor |
|----|--------|----------|------------|--------------|----------------|------------------|
| R1 | spec D-1 / D-2 / route owner + `debrief_full` normalization | Primary | Phase 0-1 | Vitest route + DOM render | `debrief_full` must not be formal RouteName / TopBar entry | screens-p1-depth.jsx:38-368 DebriefFullScreen |
| R2 | spec D-2 / Header source parity | UI source parity | Phase 1 | Vitest jsdom DOM | — | screens-p1-depth.jsx:122-144 |
| R3 | spec D-2 / ContextStrip source parity | UI source parity | Phase 1-2 | Vitest jsdom DOM + interaction | — | screens-p1-depth.jsx:412-432 DebriefContextStrip |
| R4 | spec D-3 / Stepper navigation | Primary + Boundary | Phase 1 | Vitest interaction | — | screens-p1-depth.jsx:148-156 |
| R5 | spec D-8 / 3 picker modal in-page | Primary | Phase 2 | Vitest + scenario | — | screens-p1-depth.jsx:434-518 DebriefContextPickerModal |
| R6 | spec D-7 / Cross-mode shared entries | Boundary | Phase 3 | Vitest reducer | — | review-module.md §5.2 |
| R7 | spec D-2 / Text mode GuidedDebriefRecord parity | UI source parity + Primary | Phase 3 | Vitest jsdom + interaction | — | screens-p1-depth.jsx:519-619 |
| R8 | spec D-6 / Voice UI shell only | Alternate + Boundary | Phase 3 | Vitest visual + negative grep | Web Audio API binding, SpeechRecognition listener, STT endpoint call | screens-p1-depth.jsx:656-870 |
| R9 | spec D-4 / D-5 / suggestDebriefQuestions auto + manual | Primary + Failure/recovery | Phase 4 | Vitest hook + scenario | — | — |
| R10 | spec D-9 / D-10 / createDebrief + IK + 4 response classes | Primary + Failure | Phase 5 | Vitest hook + scenario | jobId field must not receive debrief async job id | — |
| R11 | spec D-10 / Polling getJob + getDebrief | Primary | Phase 5 | Vitest hook + scenario | — | — |
| R12 | spec D-12 / 3 failure states (Failure / Missing / Timeout) | Failure/recovery | Phase 5 | Vitest render + scenario | — | — |
| R13 | spec D-14 / InterviewContext SET_DEBRIEF_CONTEXT action | Cross-layer + Boundary | Phase 5 | Vitest reducer | jobId collision, pendingAction key omission | — |
| R14 | spec D-11 / Step 2 nav practice with practiceGoal=debrief, NO createPracticePlan call | Cross-layer | Phase 6 | Vitest + scenario + negative grep | createPracticePlan, startPracticeSession (must NOT be called in debrief module) | screens-p1-depth.jsx:1388-1421 DebriefReplayPlan |
| R15 | spec §4 / Step 1 analysis render (risk_items, dimensions, provenance, no P1 fields) | Primary | Phase 6 | Vitest + scenario | nextRoundChecklist, thankYouDraft (must remain empty in P0) | screens-p1-depth.jsx:320-362 |
| R16 | spec D-13 / i18n debrief.* namespace | UX | Phase 7 | Vitest i18n keys | workspace.*, practice.*, report.* (must NOT cross namespace) | — |
| R17 | spec §4 / Theme dark+customAccent | UX | Phase 7 | Vitest + Playwright | — | — |
| R18 | spec §4 / Mobile responsive | UX | Phase 7 | Vitest + Playwright | — | — |
| R19 | spec §4 / UI visual geometry parity (desktop + mobile) | UI visual parity | Phase 8 | Playwright pixel diff | — | screens-p1-depth.jsx:38-2180 全屏 |
| R20 | spec D-16 / Privacy redlines (raw text in URL/localStorage/sessionStorage/console.log/telemetry) | Privacy | Phase 8 | Vitest spy + grep | questionText, myAnswerSummary, interviewerReaction, notes (must NOT appear in browser persistence) | — |
| R21 | spec D-18 / Legacy negative grep | Regression/Legacy-negative | Phase 8 | grep + pytest lint | experience_library, star_editor, drill_builder, mistakes_book, growth_center, report_timeline | — |
| R22 | spec §5.1 / Operation Matrix negative scope: createPracticePlan / startPracticeSession / getFeedbackReport / getCompanyIntel zero calls in debrief module | Cross-layer + Regression | Phase 6 | Vitest spy + grep | createPracticePlan, startPracticeSession, getFeedbackReport, getCompanyIntel | — |

## 2 测试项明细

### Phase 1: DebriefScreen shell + Header + ContextStrip + Stepper

#### 1.1 TestDebriefScreen_DefaultRender
- 文件：`frontend/src/app/screens/debrief/DebriefScreen.test.tsx`
- Given：route='debrief'，user 已认证，InterviewContext 含 `targetJobId='tj-1'`
- When：mount DebriefScreen
- Then：渲染 Header + ContextStrip + Stepper(current=0) + Step 0 Record panel；DOM 含 testid `debrief-screen` / `debrief-header` / `debrief-context-strip` / `debrief-stepper-step-0`
- 覆盖：R1

#### 1.1b TestRoutes_DebriefAliasNormalization
- 文件：`frontend/src/app/normalizeRoute.test.ts` + `frontend/src/app/App.test.tsx`
- Given：initial route name = `debrief_full`
- When：调用 `normalizeRouteName("debrief_full")` 并 mount App
- Then：normalize 结果为 `debrief`；App 渲染 DebriefScreen；TopBar 高亮 `debrief`；`routes.ts` 正式 RouteName 不包含 `debrief_full`
- 覆盖：R1

#### 1.2 TestDebriefHeader_RenderWithContext
- Given：InterviewContext 含 targetJob，roundType='behavioral'
- When：mount DebriefHeader
- Then：eyebrow 显示 "复盘 · {companyName} · 行为面 · {date}"；H1 + 副文案；右上 meta time / interviewer / modality 占位
- 覆盖：R2

#### 1.3 TestDebriefHeader_FallbackOnMissingContext
- Given：InterviewContext 缺 targetJobId
- When：mount
- Then：eyebrow 显示 default "复盘 · 未选择目标岗位"；其他字段 fallback；不抛错
- 覆盖：R2

#### 1.4 TestContextStrip_OpenPicker
- 文件：`frontend/src/app/screens/debrief/components/DebriefContextStrip.test.tsx`
- Given：DebriefContextStrip 渲染
- When：用户点击 targetJob 卡片
- Then：触发 `setPickerType('targetJob')` 回调；不离开当前页
- 覆盖：R3

#### 1.5 TestContextStrip_DisplayNameFetch
- Given：selectedContext.targetJob='tj-1'
- When：mount
- Then：调用 `getTargetJob('tj-1')` mock；卡片显示返回的 title + companyName
- 覆盖：R3

#### 1.6 TestContextStrip_FallbackOnAPIError
- Given：`getTargetJob` mock 抛 5xx
- When：mount
- Then：卡片 fallback 显示 targetJobId 字面量 + "无法加载详情"；不阻塞 ContextStrip 其他卡片
- 覆盖：R3

#### 1.7 TestStepper_NavigationLogic
- Given：DebriefStepper 渲染 currentStep=1（已访问过 step 0）
- When：用户点击 step 0 按钮
- Then：currentStep 回到 0；entries 状态保留；点击 step 2（未访问）无效
- 覆盖：R4

### Phase 2: 3 个 in-page picker modal

#### 2.1 TestPickerModal_BaseInteraction
- 文件：`frontend/src/app/screens/debrief/components/DebriefContextPickerModal.test.tsx`
- Given：modal open with kind='targetJob'
- When：(1) 点击外部 (2) 按 Esc (3) 点击 cancel (4) 点击 confirm with selectedId
- Then：(1)(2)(3) 触发 onClose；(4) 触发 onConfirm(selectedId) 且 onClose
- 覆盖：R5

#### 2.2 TestJDPicker_ListAndConfirm
- Given：fixture `listTargetJobs` 返回 3 个 ready jobs
- When：JD picker 渲染 → 用户选 tj-2 → 确认
- Then：onConfirm('tj-2') 触发；reducer dispatch SET_DEBRIEF_CONTEXT with targetJob='tj-2'
- 覆盖：R5

#### 2.3 TestMockSessionPicker_ListAndOptional
- Given：Phase 0 已生成 `listPracticeSessions`；fixture 返回 2 个 completed sessions；"暂不关联" option 渲染
- When：用户点击 "暂不关联" 确认
- Then：onConfirm(null) 触发；reducer dispatch SET_DEBRIEF_CONTEXT with mockSession=null
- 覆盖：R5

#### 2.4 TestMockSessionPicker_FilterFallback
- Given：mock generated client 已含 `listPracticeSessions`，但 operation 不接受 `status` filter parameter
- When：Mock picker 调用
- Then：client-side filter `session.status === 'completed'`；只渲染 completed sessions；记录 fallback warning
- 覆盖：R5

#### 2.5 TestResumePicker_ListAndConfirm
- Given：fixture `listResumes` 返回 2 个 active assets；fixture `listResumeVersions(resumeAssetId)` 返回 selected asset 的 2 个 ready versions
- When：Resume picker 渲染 → 选 resume asset → 选 resume-v3 → 确认
- Then：onConfirm({resumeAssetId, resumeVersionId:'resume-v3'})；reducer dispatch SET_DEBRIEF_CONTEXT with `resumeVersionId`
- 覆盖：R5

#### 2.6 TestContextStrip_AutoTriggerSuggestions
- Given：selectedContext.targetJob 改变到非空 + selectedContext.resume 非空
- When：useEffect 触发
- Then：debounce 500ms 后 enable useSuggestDebriefQuestions hook with targetJobId + resumeVersionId
- 覆盖：R9

### Phase 3: Step 0 + 跨模式共享 entries + Voice UI shell

#### 3.1 TestRecordSummaryBar_Counts
- Given：entries=[3 items with sources ai_confirmed / ai_edited / manual]
- When：渲染 summary bar
- Then：total=3；chips 显示 recorded:3 / text:2 / voice:0 / manual:1
- 覆盖：R7

#### 3.2 TestModeToggle_PreservesEntries
- Given：text 模式 entries.length=2
- When：用户切换到 voice 模式
- Then：entries 不变；mode='voice'；切回 text 后 entries 仍有 2 条
- 覆盖：R6

#### 3.3 TestGuidedRecord_OccurredCTA
- Given：suggestions=[6 items]，activeGuide=0
- When：用户点击 "遇到过，记录" CTA on suggestion[0]
- Then：entries.length += 1，新 entry source='ai_confirmed'，q/a/follow 来自 suggestion；activeGuide += 1
- 覆盖：R7

#### 3.4 TestGuidedRecord_SkipCTA
- When：用户点击 "没问到，跳过"
- Then：entries 不变；activeGuide += 1
- 覆盖：R7

#### 3.5 TestGuidedRecord_EditCTA
- When：用户点击 "改成真实问题" → inline edit → save
- Then：entries += 1 with source='ai_edited'，questionText 来自 user edit；activeGuide += 1
- 覆盖：R7

#### 3.6 TestGuidedRecord_ManualAddCTA
- When：用户点击 "手动添加真实问题" → 填表 → save
- Then：entries += 1 with source='manual'
- 覆盖：R7

#### 3.7 TestVoiceRecord_UIShellOnly
- 文件：`frontend/src/app/screens/debrief/components/VoiceDebriefRecord.test.tsx`
- Given：渲染 VoiceDebriefRecord
- Then：DOM 含 testid `debrief-voice-status` / `debrief-voice-not-implemented` / 「空格暂停/继续」键盘提示文案；entries 列表渲染；UI 视觉锚点 100% 对齐 prototype
- 覆盖：R8

#### 3.8 TestVoiceRecord_NoSTTBinding
- Given：mount VoiceDebriefRecord
- When：检查源码 + runtime
- Then：源码不 import Web Audio API / SpeechRecognition / navigator.mediaDevices；runtime 不调用任何 STT endpoint；spy 不接收音频流；grep `"navigator.mediaDevices\|SpeechRecognition\|webkitSpeechRecognition\|AudioContext"` in `frontend/src/app/screens/debrief/` 0 命中
- 覆盖：R8

#### 3.9 TestSubmitCTA_DisabledState
- Given：entries=[] 或 selectedContext.targetJob=null
- When：检查 Submit CTA
- Then：CTA disabled
- 覆盖：R10

#### 3.10 TestSubmitCTA_EnabledAndClick
- Given：entries=[2 items]，selectedContext.targetJob='tj-1'
- When：用户点击 Submit CTA
- Then：触发 useSubmitDebrief hook
- 覆盖：R10

### Phase 4: suggestDebriefQuestions 集成

#### 4.1 TestUseSuggestQuestions_AutoTrigger
- 文件：`frontend/src/app/screens/debrief/hooks/useSuggestDebriefQuestions.test.ts`
- Given：targetJobId 改变 from null → 'tj-1'，enabled=true
- When：渲染 + wait 500ms
- Then：mock `suggestDebriefQuestions` 调用 1 次 with targetJobId='tj-1'；suggestions 含 6 items；loading false；error null
- 覆盖：R9

#### 4.2 TestUseSuggestQuestions_Refetch
- Given：hook 已 fetch 一次
- When：调 refetch()
- Then：mock 再次调用；suggestions 更新
- 覆盖：R9

#### 4.3 TestUseSuggestQuestions_Debounce
- Given：enabled=true
- When：targetJobId 100ms 内连续改变 3 次
- Then：mock 只调用 1 次（最终值）
- 覆盖：R9

#### 4.4 TestDebriefScreen_SuggestionsIntegration
- Given：完整 DebriefScreen with ContextStrip 三选完成
- When：suggestions hook 返回 6 items
- Then：GuidedDebriefRecord 渲染 suggestions；activeGuide=0 显示 suggestions[0]
- 覆盖：R9

#### 4.5 TestSuggestions_AIFailureDegradation
- Given：mock `suggestDebriefQuestions` 返回 502 `AI_PROVIDER_TIMEOUT`
- When：触发 hook
- Then：error.code='AI_PROVIDER_TIMEOUT'；UI 显示 inline error "推荐生成失败，可手动添加问题"；不阻塞 step 0；"重新生成推荐" 按钮启用；parameterized 子用例覆盖 `AI_PROVIDER_CONFIG_INVALID` / `AI_PROVIDER_SECRET_MISSING` / `AI_OUTPUT_INVALID` / `AI_FALLBACK_EXHAUSTED`
- 覆盖：R9

### Phase 5: createDebrief + 双轨 polling + 失败态

#### 5.1 TestUseSubmitDebrief_Happy202
- 文件：`frontend/src/app/screens/debrief/hooks/useSubmitDebrief.test.ts`
- Given：合法 payload
- When：调用 submit
- Then：generated `createDebrief` 调用 with Idempotency-Key (UUIDv4)；返回 202 + DebriefWithJob；reducer dispatch SET_DEBRIEF_CONTEXT；setStep(1) 触发；polling 启动
- 覆盖：R10

#### 5.2 TestUseSubmitDebrief_422ValidationFailed
- Given：mock 返回 422 VALIDATION_FAILED with details=[{field:'questions', message:'must have at least 1 question'}]
- When：submit
- Then：error.code='VALIDATION_FAILED'；UI inline error 显示 field 失败信息；不 setStep
- 覆盖：R10

#### 5.3 TestUseSubmitDebrief_409IKMismatchRetry
- Given：mock 第一次返回 409 IDEMPOTENCY_KEY_MISMATCH，第二次返回 202
- When：submit
- Then：自动重生 IK 并重试 1 次；最终 202 + 正常流程
- 覆盖：R10

#### 5.4 TestUseSubmitDebrief_401AuthGate
- Given：mock 返回 401 AUTH_UNAUTHORIZED
- When：submit
- Then：useRequestAuth 触发 type='submit_debrief'，params=full submit payload；不抛错给用户；登录后 pendingAction 回调 submit
- 覆盖：R10

#### 5.5 TestUseDebriefPolling_HappySuccess
- 文件：`frontend/src/app/screens/debrief/hooks/useDebriefPolling.test.ts`
- Given：mock `getJob` 返回 status='queued' → 'running' → 'succeeded'；mock `getDebrief` 返回 completed Debrief
- When：启动 polling with debriefJobId+debriefId
- Then：phase A 调用 `getJob` 3 次（按指数退避）；status='succeeded' 后 phase B 调用 `getDebrief` 1 次；pollingState='succeeded'
- 覆盖：R11

#### 5.6 TestUseDebriefPolling_JobFailed
- Given：mock `getJob` 返回 status='failed', errorCode='AI_PROVIDER_TIMEOUT'
- When：polling
- Then：phase B 不触发；pollingState='failed' with errorCode
- 覆盖：R11, R12

#### 5.7 TestUseDebriefPolling_MaxAttemptsTimeout
- Given：mock `getJob` 永久返回 'queued'
- When：polling 达到 max attempts (30)
- Then：pollingState='timeout'；停止轮询
- 覆盖：R11

#### 5.8 TestUseDebriefPolling_VisibilityPause
- Given：polling running
- When：document.visibilityState='hidden'
- Then：暂停 polling；再次 'visible' 时恢复
- 覆盖：R11

#### 5.9 TestDebriefFailureState_Render
- 文件：`frontend/src/app/screens/debrief/components/DebriefFailureState.test.tsx`
- Given：errorCode='AI_PROVIDER_TIMEOUT'
- When：mount
- Then：DOM 含 testid `debrief-failure-state`；显示 errorCode 文案；CTA「返回 step 0 编辑」+「重试生成」
- 覆盖：R12

#### 5.10 TestDebriefMissingContextState_Render
- Given：targetJobId=undefined
- When：mount DebriefScreen
- Then：渲染 DebriefMissingContextState；自动打开 JD picker modal
- 覆盖：R12

#### 5.11 TestDebriefTimeoutState_Render
- Given：pollingState='timeout'
- When：mount
- Then：DOM 含 testid `debrief-timeout-state`；CTA「重试」+「返回 step 0」
- 覆盖：R12

#### 5.12 TestInterviewContext_SetDebriefContext
- 文件：`frontend/src/app/interview-context/reducer.test.ts`
- Given：existing InterviewContext state
- When：dispatch SET_DEBRIEF_CONTEXT with {debriefId, debriefJobId, targetJobId, sessionId, resumeVersionId, practiceGoal, language}
- Then：state 含新字段；其他字段保留；`jobId` 不被写入
- 覆盖：R13

#### 5.13 TestInterviewContext_DoesNotOverwriteJobId
- Given：existing state 含 `jobId='target-job-alias'`
- When：dispatch SET_DEBRIEF_CONTEXT with `debriefJobId='async-job-1'`
- Then：`state.jobId` 仍为 `target-job-alias`；`state.debriefJobId='async-job-1'`
- 覆盖：R13

#### 5.14 TestPendingAction_DebriefParamsRoundTrip
- Given：pendingAction params 含 `practiceGoal='debrief'` / `debriefId='D'` / `debriefJobId='J'`
- When：encodePendingAction → decodePendingAction
- Then：三项全部保留；登录恢复后 `requestAuth` 可回到 debrief 或 practice handoff
- 覆盖：R13

#### 5.15 TestInterviewContext_OtherActionsNotAffected
- Given：existing state
- When：dispatch 既有 SET_PRACTICE_CONTEXT
- Then：SET_DEBRIEF_CONTEXT 字段不被覆盖
- 覆盖：R13

### Phase 6: Step 1 分析渲染 + Step 2 复盘面试 launcher + handoff

#### 6.1 TestStep1Analysis_RiskItemsRender
- 文件：`frontend/src/app/screens/debrief/DebriefScreen.test.tsx`
- Given：fixture `getDebrief=default` (completed + riskItems=[3 items])
- When：step=1 渲染
- Then：风险项列表显示 3 项；每项 label + severity color tag (low=gray, medium=amber, high=red)；testid `debrief-analysis-risk-item-{i}`
- 覆盖：R15

#### 6.2 TestStep1Analysis_DimensionsRender
- Given：同上
- When：step=1
- Then：3 张维度对比卡片渲染（与模拟面试 / JD / 简历）；testid `debrief-analysis-dimension-{mock,jd,resume}`
- 覆盖：R15

#### 6.3 TestStep1Analysis_ProvenanceExpansion
- Given：getDebrief 返回 provenance 6 字段
- When：用户点击 "关于本次分析" 展开
- Then：显示 6 字段 (promptVersion / rubricVersion / modelId / language / featureFlag / dataSourceVersion)；不显示运行时字段
- 覆盖：R15

#### 6.4 TestStep1Analysis_P1FieldsNotRendered
- Given：getDebrief 返回 nextRoundChecklist=[]，thankYouDraft=null（P0 留空）
- When：step=1
- Then：DOM 不含 nextRoundChecklist 区块 / thankYouDraft 区块；grep `data-testid="debrief-next-round-checklist"` 0 命中
- 覆盖：R15

#### 6.5 TestStep2Launcher_ContentPreview
- Given：step=2，debrief.questions + riskItems 已就绪
- When：mount
- Then：DebriefReplayPlan 内容预览渲染；显示复现真实问题 + 薄弱处追问 + 简历对比预览
- 覆盖：R14

#### 6.6 TestStartDebriefInterview_NavPayload
- Given：用户已认证；step=2 渲染
- When：用户点击 "开始复盘面试" CTA
- Then：nav 调用 with payload `{practiceGoal:'debrief', mode:'text', modality:'text', sessionId:'mock-24', targetJobId:'tj-1', resumeVersionId:'resume-v3', debriefId:'D', language:'zh'}`
- 覆盖：R14

#### 6.7 TestStartDebriefInterview_AuthGate
- Given：用户未认证
- When：用户点击 "开始复盘面试" CTA
- Then：useRequestAuth 触发 type='start_debrief_interview'，params=full payload；登录后 pendingAction 恢复 nav practice
- 覆盖：R14

#### 6.8 TestStartDebriefInterview_NoCreatePracticePlanCall
- Given：完整 step 2 流程
- When：用户点击 "开始复盘面试" → spy 监控 generated client method calls
- Then：`createPracticePlan` / `startPracticeSession` 在 frontend/src/app/screens/debrief/ 内 0 调用（spy assert）；只调 `nav`
- 覆盖：R14, R22

### Phase 7: i18n + 主题 + 响应式

#### 7.1 TestI18n_DebriefNamespaceComplete
- 文件：`frontend/src/app/i18n/debrief.test.ts`
- Given：zh.ts + en.ts 的 debrief.* keys
- When：grep 列出
- Then：包含 ~30 个关键 keys 完整覆盖 header / contextStrip / stepper / step0-2 / pickers / failureStates / suggestions / voice；zh + en 字段数对齐
- 覆盖：R16

#### 7.2 TestTheme_DebriefScreen
- Given：root data-theme='dark' / data-custom-accent='blue'
- When：mount DebriefScreen
- Then：Header / ContextStrip / Stepper / picker modal 应用 CSS variables；computed style 与 prototype 一致
- 覆盖：R17

#### 7.3 TestResponsive_Mobile
- Given：viewport 390×844
- When：mount DebriefScreen
- Then：Header 紧凑布局；ContextStrip 单列；Stepper 横向滑动；Step 0 单列 + tab；Step 2 sticky bottom CTA；picker 全屏 sheet；不出现水平滚动
- 覆盖：R18

### Phase 8: Pixel parity + 隐私 + Legacy negative

#### 8.1 TestPrivacy_NoRawTextInLocalStorage
- 文件：`frontend/src/app/screens/debrief/privacy.test.tsx`
- Given：完整 debrief 流程 with marker `__SECRET_RAW_TEXT__` 注入 entries
- When：流程完成 → 检查 localStorage / sessionStorage
- Then：marker 0 命中；localStorage 仅含 stable IDs + display knobs；sessionStorage 同款
- 覆盖：R20

#### 8.2 TestPrivacy_NoRawTextInConsoleLog
- Given：同上
- When：检查 console.log spy 输出
- Then：marker 0 命中；fixture transport spy 接收 raw body 但不应记录到 logger
- 覆盖：R20

## 3 集成测试与覆盖率说明

- 覆盖率以 plan 列出的测试项达成度衡量；不引入 raw coverage 百分比作为 hard gate。
- Playwright pixel parity 测试详见 `frontend/tests/pixel-parity/debrief-desktop.spec.ts` + `debrief-mobile.spec.ts`。
- E2E scenario（P0.065-069）覆盖见 [bdd-plan.md](./bdd-plan.md) 与 [bdd-checklist.md](./bdd-checklist.md)。
