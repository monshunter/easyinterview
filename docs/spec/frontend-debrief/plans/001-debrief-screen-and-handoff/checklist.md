# 001 Debrief Screen and Handoff Checklist

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-05-17

**关联计划**: [plan](./plan.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 Test Checklist**: [test-checklist](./test-checklist.md)

## Phase 0: 依赖验证 + ui-design source map + 包结构

- [x] 0.0 backend-practice listPracticeSessions cross-owner addendum 落地（独立分支 `feat/backend-practice-list-sessions-0517`，commit `feat(openapi): add listPracticeSessions operation`）：`openapi/openapi.yaml` 新增 `GET /practice/sessions` operation + `PaginatedPracticeSession` schema；`openapi/fixtures/PracticeSessions/listPracticeSessions.json` 含 `default` / `empty` scenarios；`scripts/lint/openapi_inventory.py` EXPECTED_OPERATIONS 增加 listPracticeSessions；`docs/spec/openapi-v1-contract/spec.md` §3.1.1 新增 #57，history 1.22；regen Go/TS generated artifacts；gates `make codegen-openapi` / `make lint-openapi` / `make validate-fixtures` / `make docs-check` 通过
- [x] 0.1 cross-owner Phase 0 完成验证：`grep -rn "suggestDebriefQuestions\|createDebrief\|getDebrief\|listPracticeSessions\|listResumes\|listResumeVersions" frontend/src/api/generated/` 命中（client.ts + spec.ts）；`ls openapi/fixtures/Debriefs/` 含 `createDebrief.json` / `getDebrief.json` / `suggestDebriefQuestions.json` 三个 fixture；`ls openapi/fixtures/PracticeSessions/listPracticeSessions.json` 存在（由 0.0 addendum 落地）；`grep -rn "DebriefRoundType\|DebriefQuestionSource\|DEBRIEF_NOT_FOUND\|IDEMPOTENCY_KEY_MISMATCH" frontend/src/lib/conventions/ frontend/src/api/generated/` 命中（B1 codegen 产物路径校准；原 plan 文案 `shared/ts/conventions/` 为旧路径漂移）；backend-practice 现状已支持 `goal='debrief'` + 合法 `mode IN ('assisted','strict')`（验证证据记录在 backend-debrief/001 Phase 0.6 与 backend-practice/004）
- [x] 0.2 ui-design source map 记录到 plan history 与本 checklist 注脚（6 个组件 source anchor）
  <!-- verified 2026-05-17: ui-design/src/screens-p1-depth.jsx anchors confirmed via `grep -n "function DebriefFullScreen\|const DebriefContextStrip\|const DebriefContextPickerModal\|const GuidedDebriefRecord\|const VoiceDebriefRecord\|const DebriefReplayPlan"` → 38 / 412 / 434 / 519 / 656 / 1388. -->
- [x] 0.3 创建包结构 `frontend/src/app/screens/debrief/{DebriefScreen.tsx, components/{DebriefHeader,DebriefContextStrip,DebriefStepper}.tsx, types.ts, debrief.css}`；空 stub 编译通过：`pnpm --filter @easyinterview/frontend typecheck`（Phase 1.x 同步落地组件实体；`hooks/` / `reducer.ts` / `i18n/` 子目录由 Phase 4+/5+ 落地时再创建以避免空目录；i18n keys 走全局 `frontend/src/app/i18n/locales/{zh,en}.ts` 的 `debrief.*` 命名空间，与现存 `report.*` / `practice.*` 约定一致）
- [x] 0.4 route 接线：`App.tsx` 的 `case "debrief"` → `<DebriefScreen>`（与 ReportScreen 同档前置；移除 PlaceholderScreen 对 debrief 的占位）；`normalizeRoute.ts` 把历史 alias `debrief_full` normalize 到 `debrief`；不在 `routes.ts` 正式 `RouteName` / `PRIMARY_NAV_ROUTES` / `INTERVIEW_CONTEXT_ROUTES` 中新增 `debrief_full`；TopBar 一级导航 `debrief` 高亮逻辑保留；测试 `normalizeRoute.test.ts > normalizes the historical debrief_full alias to the current debrief route` 通过；route 渲染由 `DebriefScreen.test.tsx > TestDebriefScreen_DefaultRender` 覆盖（`screen.getByTestId("route-debrief")`）

## Phase 1: DebriefScreen shell + Header + ContextStrip + Stepper

- [x] 1.1 `<DebriefScreen>` container：route mount via App.tsx；internal state 初始化（step=0, maxVisited=0, inputMode='text', selectedContext=EMPTY_SELECTED_CONTEXT, pickerKind=null）；render Header + ContextStrip + Stepper + step panel；测试：`DebriefScreen.test.tsx > TestDebriefScreen_DefaultRender` 三例（shell + back nav + picker open）通过；`InterviewContext` 注入由 Phase 5.4 配合 reducer 扩展时接入（当前 container 已暴露 `data-step` / `data-input-mode` / `data-picker-kind` 数据属性供后续 hook + reducer 整合）。
- [x] 1.2 `<DebriefHeader>`：复刻 prototype lines 122-144（back / eyebrow / serif title / subcopy / 右上 meta 三列）；eyebrow 由 `selectedContext.targetJob.companyName + title` 派生；缺失数据走 `debrief.header.eyebrowFallback*` / `debrief.header.metaFallback`；i18n keys 通过全局 `useI18n` + `MessageKey`；测试：`DebriefHeader.test.tsx > TestDebriefHeader_RenderWithContext`（eyebrow + meta + onBack）+ `TestDebriefHeader_FallbackOnMissingContext`（unset eyebrow + 三个 dd fallback）通过。
- [x] 1.3 `<DebriefContextStrip>`：复刻 lines 412-432；三卡片（targetJob / mockSession / resume）点击触发 `onOpenPicker(kind)`；title 从 `selectedContext` 派生（`pickTitle` 处理 company+title / id / asset+versionId 组合），缺失 fallback 到 `debrief.contextStrip.unset`；测试：`DebriefContextStrip.test.tsx > TestContextStrip_OpenPicker`（三 kind 按序回调）+ `TestContextStrip_DisplayNameFetch`（已选择上下文显示派生标题）+ `TestContextStrip_FallbackOnAPIError`（未选择走 fallback 文案）通过。实时 `getTargetJob` / `getResumeVersion` / `getPracticeSession` fetch 由 Phase 2 picker 落地时统一接入 picker modal（context strip 自身保持 props-driven，符合 §3.1 "container 拥有 fetch，叶组件保持 pure" 约定）。
- [x] 1.4 `<DebriefStepper>`：复刻 lines 148-156；3 步骤，`data-active` / `data-reachable` 暴露状态；未访问的前向 step 通过 `disabled` 阻止跳跃；测试：`DebriefStepper.test.tsx > TestStepper_NavigationLogic` 三例（高亮 + 已访问可回 + 未访问不可跳）通过。

## Phase 2: 3 个 in-page picker modal

- [x] 2.1 `<DebriefContextPickerModal>` 通用骨架：复刻 lines 434-518；接收 kind / options / selectedId / loading / errorMessage / banner / emptyCopy / allowEmpty / noneOptionCopy / onClose / onConfirm；外部点击 + Esc 关闭（keydown listener）；mobile 全屏 sheet 由 debrief.css `@media (max-width: 640px)` 控制（`width:100%; height:100%`）；testid `debrief-picker-modal` / `debrief-picker-close` / `debrief-picker-cancel` / `debrief-picker-confirm` 已暴露；NONE_SENTINEL 处理 `allowEmpty` 的「暂不关联」选项。
- [x] 2.2 JD picker：`JDPicker` 调 `client.listTargetJobs({ query: { status: 'ready' } })`；通过 `usePickerOptions` loader 拉选项；onConfirm 把 `TargetJob` 写入 DebriefScreen `selectedContext.targetJob`（reducer 的 `SET_DEBRIEF_CONTEXT` 写发生在 Phase 5.1 createDebrief 响应后，而非 picker 提交时 —— spec §3.2 把 picker 选择视为 UI 草稿，不污染全局 InterviewContext）。
- [x] 2.3 Mock Session picker：`MockSessionPicker` 调 `client.listPracticeSessions({ query: { targetJobId, status: 'completed' } })`；当 server 响应非全 `completed` 时 hook 报告 `client-side-status-filter` fallback 并在 banner 中渲染 `debrief.picker.mockSession.filterFallback`；`allowEmpty=true` + `noneOptionCopy` 提供「暂不关联模拟面试」选项；`targetJobId === null` 时显示「请先选择目标岗位」empty state。
- [x] 2.4 Resume picker：`ResumePicker` 两段式 phase（asset → version）；第一阶段 `client.listResumes()` 拉 `parseStatus==='ready'` + `status==='active'` 资产；第二阶段 `client.listResumeVersions(assetId)` 拉版本；返回 `{ asset, version }` 到 DebriefScreen。
- [x] 2.5 ContextStrip 三选完成 detect + 自动触发 suggestions：DebriefScreen `suggestionsEnabled = Boolean(targetJob) && Boolean(resumeVersion)`；`useSuggestDebriefQuestions` hook 内部 `useEffect` 500ms debounce，targetJob / sessionId / resumeVersionId / language 改变时 coalesce 成一次请求；mockSession 保持可选。

## Phase 3: Step 0 复盘记录 + 跨模式共享 entries + Voice UI shell

- [x] 3.1 顶部统一汇总条：`DebriefRecordSummaryBar` 复刻 lines 162-232；entries.length + text / voice / manual 三 chip（real-recorded 暂未引入直到 Phase 4+ 真实 mock-extraction）；testid `debrief-record-summary` / `debrief-record-summary-count` / `debrief-chip-*` 暴露。
- [x] 3.2 Mode toggle：`DebriefModeToggle` 复刻 lines 187-210；`text|voice` tab + hint 文案；切换不修改 entries（DebriefScreen 持有 entries state，toggle 只 setInputMode）；testid `debrief-mode-toggle` + `data-mode` 数据属性。
- [x] 3.3 `<GuidedDebriefRecord>`：复刻 lines 519-619；接收 suggestions + entries + activeGuide；4 个 CTA testid `debrief-suggested-question-{occurred,skip,edit,manual}` + inline editor (`debrief-guided-editor`) + 4 种 entry source (`ai_confirmed` / `ai_edited` / `manual` / 后续 `voice_extracted` 由 Phase 8 真实集成填入)；entries 列表 testid `debrief-guided-entries`。
- [x] 3.4 `<VoiceDebriefRecord>` UI shell：复刻 lines 656-870 视觉占位 + 「空格暂停/继续」hint 文案；显示固定占位 `debrief-voice-not-implemented`；**不**绑定 Web Audio API / SpeechRecognition；切换回 text 模式 entries 仍保留（DebriefScreen `hidden={inputMode !== "voice"}` 双 panel 持续 mount）。
- [x] 3.5 Submit CTA：`DebriefSubmitCTA` 复刻 lines 314-316；disabled 条件 entries.length === 0 或 targetJob === null（reason 文案 inline）；点击调 DebriefScreen `handleSubmit` → Phase 5 `useSubmitDebrief.submit`；testid `debrief-submit-btn` / `debrief-submit-reason`。

## Phase 4: suggestDebriefQuestions 集成

- [x] 4.1 `useSuggestDebriefQuestions` hook 实现：参数 `{ targetJobId?, sessionId?, resumeVersionId?, language, count=6, enabled }`；返回 `{ suggestions, loading, error, refetch }`；`useEffect` 内 500ms `setTimeout` debounce + `fetchTokenRef` 取消乱序请求；`parseError` 提取 B1 canonical error code；hook 在 `enabled === false` 或缺 targetJobId 时清空 suggestions。
- [x] 4.2 DebriefScreen 整合：`suggestionsEnabled = Boolean(targetJob) && Boolean(resumeVersion)` 自动启用；`GuidedDebriefRecord` 读 hook 输出渲染 currentGuide；`debrief-guided-regenerate` 按钮调 `suggestions.refetch`。
- [x] 4.3 失败降级：`GuidedDebriefRecord` 检测 `errorCode != null` → 渲染 `debrief-guided-failure` inline error + regenerate 按钮；error 不阻塞 entries / submit / mode toggle / picker；error code 文案走 `debrief.failure.code.*`（B1 canonical 全部映射到 i18n keys 见 `debrief.failure.code.AI_*`）。

## Phase 5: createDebrief + 双轨 polling + 失败态

- [x] 5.1 `useSubmitDebrief` hook：`generateIdempotencyKey` 优先 `crypto.randomUUID`；调 `client.createDebrief(body, { headers: { 'Idempotency-Key': key } })`；映射 entries → `DebriefQuestionInput[]`；handle UNAUTHENTICATED / VALIDATION_FAILED / IDEMPOTENCY_KEY_MISMATCH (auto-retry 一次新 IK) / UNKNOWN；成功后 dispatch `SET_DEBRIEF_CONTEXT { debriefId, debriefJobId, practiceGoal:"debrief" }`，**绝不写 `jobId`**。
- [x] 5.2 `useDebriefPolling` hook：Phase A `getJob(debriefJobId)` 指数退避 1.5s × 1.5 上限 8s / max 30 attempts；`document.visibilitychange` + `window.focus/blur` 暂停 polling；`job.status === 'succeeded'` → Phase B 一次性 `getDebrief(debriefId)` 拉 enriched record；`failed` 收 errorCode；超过 max attempts → `timeout`；状态机 `idle → running → succeeded | failed | timeout`；`restart()` 重置。
- [x] 5.3 `<DebriefFailureState>`：errorCode 文案查 `debrief.failure.code.*` keys；CTA「返回 Step 0 编辑」+「重试生成」；testid `debrief-failure-state` / `debrief-failure-retry` / `debrief-failure-back` / `debrief-failure-message`。
- [x] 5.4 `<DebriefMissingContextState>`：缺 targetJob 时由 Step 0 panel 渲染；CTA 触发 `setPickerKind("targetJob")` 自动开 JD picker；testid `debrief-missing-context-state` / `debrief-missing-cta`。
- [x] 5.5 `<DebriefTimeoutState>`：polling.state === 'timeout' 时渲染；CTA 重启 polling / 返回 Step 0；testid `debrief-timeout-state` / `debrief-timeout-retry` / `debrief-timeout-back`。
- [x] 5.6 InterviewContext reducer 扩展 `SET_DEBRIEF_CONTEXT`：payload `{ debriefId?, debriefJobId?, practiceGoal? }`；不写 `jobId` / `targetJobId`；`HYDRATE_FROM_ROUTE` 增加 debriefId / debriefJobId 回填；`PENDING_ACTION_INTERVIEW_KEYS` 增加 `practiceGoal` / `debriefId` / `debriefJobId` / `sessionId`；测试：`InterviewContext.test.tsx > TestInterviewContext_SetDebriefContext` + `TestInterviewContext_DoesNotOverwriteJobId` + `TestPendingAction_DebriefParamsRoundTrip` + `TestInterviewContext_OtherActionsNotAffected` 通过（1002 个 vitest）。

## Phase 6: Step 1 分析渲染 + Step 2 复盘面试 launcher + handoff

- [x] 6.1 Step 1 分析渲染：`DebriefAnalysisStep` 渲染 `debrief.riskItems` 列表（severity 文案查 `debrief.severity.*`）+ 3 张维度对比卡片（target / mock / resume，body 从 `debrief.questions[*].aiAnalysis` 派生，缺失走 "—"）+ provenance 展开区 6 字段（promptVersion / rubricVersion / modelId / language / featureFlag / dataSourceVersion）；**不**渲染 `nextRoundChecklist` / `thankYouDraft`；testid `debrief-analysis-step` / `debrief-analysis-risk-item` / `debrief-analysis-dimension-*` / `debrief-analysis-provenance-toggle`。
- [x] 6.2 Step 2 复盘面试 launcher：`DebriefReplayPlan` 复刻 lines 1388-1421；预览取 `debrief.questions.map(q => q.questionText).slice(0,5)` 或 fallback 到 entries；riskItems.slice(0,3) 展示薄弱项预览；testid `debrief-replay-plan` / `debrief-replay-preview-questions` / `debrief-replay-preview-risks`。
- [x] 6.3 「开始复盘面试」CTA：`handleStartReplay` 组装 `{ practiceGoal:'debrief', mode:'text', modality:'text', targetJobId, resumeVersionId?, sessionId?, debriefId? }` 调 `useRequestAuth({ type:'start_debrief_interview', label, route:'practice', params })`；登录态 → 直接 `navigate('practice', params)`，未登录 → 自动 encode 为 pendingAction 跳 `auth_login`；本 plan **不**调 `createPracticePlan` / `startPracticeSession`；testid `debrief-start-interview-btn`。

## Phase 7: i18n + 主题 + 响应式

- [x] 7.1 i18n `debrief.*` namespace：`frontend/src/app/i18n/locales/{zh,en}.ts` 已覆盖 header / contextStrip / stepper / picker / record / failure / missing / timeout / analysis / replay / severity 共 80+ keys；测试 `frontend/src/app/i18n/__tests__/debriefI18nCoverage.test.ts > frontend-debrief/001 i18n coverage` 强制 zh ↔ en 字节同步 + AI_* / IDEMPOTENCY_KEY_MISMATCH / VALIDATION_FAILED 错误码覆盖（1007 vitest 通过）。
- [x] 7.2 主题适配：debrief 组件全部使用 `var(--ei-color-*)` / `var(--ei-font-*)` / `var(--ei-space-*)` / `var(--ei-radius-*)` 设计 token，由 root `data-theme` / `data-mode` / `data-custom-accent` 接管 dark + customAccent；frontend-debrief legacy lint 拒绝旧 token 漂移（`scripts/lint/frontend_debrief_legacy.py`）。
- [x] 7.3 Mobile 响应式：`debrief.css @media (max-width: 640px)` 已覆盖 Header / ContextStrip / Stepper / Step 0 双栏折叠 / Step 1 provenance 单列 / Step 2 sticky CTA / picker 全屏 sheet；E2E.P0.065 / P0.069 scenario 覆盖 DOM 锚点 + token 应用 + legacy negative grep。

## Phase 8: Playwright pixel parity + 隐私 + legacy negative + BDD

- [x] 8.1 Playwright pixel parity desktop 1440×900：`frontend/tests/pixel-parity/debrief.spec.ts` 已接入 P0.069；`pnpm --filter @easyinterview/frontend exec playwright test tests/pixel-parity/debrief.spec.ts` 通过，覆盖 `debrief_full` alias normalize、Step 0 source anchors、desktop viewport bounding boxes、theme/customAccent computed values 与非空 screenshot smoke。
- [x] 8.2 Playwright pixel parity mobile 390×844：同一 Playwright spec 的 `mobile` project 通过，覆盖 mobile viewport bounding boxes、horizontal overflow negative 与非空 screenshot smoke；desktop-only/mobile-only 用例按项目条件跳过。
- [x] 8.3 主题 pixel parity：同一 Playwright spec 覆盖 dark mode 与 customAccent computed style 变化；CSS token 派生继续由 `scripts/lint/frontend_debrief_legacy.py` 拒绝硬编码漂移。
- [x] 8.4 隐私 + telemetry 验证：`frontend/src/app/screens/debrief/__tests__/privacyBoundary.test.ts` 静态扫描 debrief 模块，断言任何源码都不会以 `localStorage.setItem(.*questionText)` / `sessionStorage.setItem(.*myAnswerSummary)` / `console.log(.*interviewerReaction)` / `navigate(.*questionText)` / `history.pushState(.*questionText)` 形态泄漏 raw entry text（zero offenders）。
- [x] 8.5 隐私 grep gate：仅命中 `types.ts` 字段声明 + `GuidedDebriefRecord` / `DebriefReplayPlan` 受控 DOM 显示 + i18n key（皆为合理位置）；privacyBoundary.test.ts 作为可执行口径替代静态 grep。
- [x] 8.6 Legacy negative grep：`scripts/lint/frontend_debrief_legacy.py` + P0.069 trigger.sh 联合覆盖 `frontend/src/app/screens/debrief/` / `frontend/src/app/i18n/locales/` / `test/scenarios/e2e/p0-06[56789]-*`，全部 0 命中（已用 `--exclude=trigger.sh` 排除自指断言文件）。
- [x] 8.7 Legacy negative lint script：`scripts/lint/frontend_debrief_legacy.py` + `scripts/lint/frontend_debrief_legacy_test.py`（4 测试：terms / clean-repo / offender / test-file-skip）；`python3 -m pytest scripts/lint -q` 249 个测试通过。
- [x] 8.8 BDD-Gate E2E.P0.065：`p0-065-debrief-default-render-and-pickers` 四段脚本通过（DebriefScreen / Header / ContextStrip / Stepper + route normalize 测试 + legacy lint）。
- [x] 8.9 BDD-Gate E2E.P0.066：`p0-066-debrief-text-suggestions-and-submit` 四段脚本通过（debrief 模块全部 vitest + InterviewContext reducer + pendingAction + privacy boundary + legacy lint）。
- [x] 8.10 BDD-Gate E2E.P0.067：`p0-067-debrief-polling-happy-and-analysis` 四段脚本通过（debrief 模块 vitest + legacy lint）。
- [x] 8.11 BDD-Gate E2E.P0.068：`p0-068-debrief-failure-and-handoff` 四段脚本通过（debrief 模块 vitest + InterviewContext reducer + 模块内 `createPracticePlan` / `startPracticeSession` 直接调用 0 命中负向断言 + legacy lint）。
- [x] 8.12 BDD-Gate E2E.P0.069：`p0-069-debrief-pixel-parity-and-legacy-negative` 四段脚本通过（i18n coverage + privacy boundary + dev-mock fixture coverage + frontend build + debrief Playwright pixel parity + legacy lint + scenario-tree legacy grep）。

## Phase 9: Plan 收口

- [x] 9.1 全局回归：2026-05-17 L2 close-out 重新验证 `pnpm --filter @easyinterview/frontend exec vitest run src/app/screens/debrief/DebriefScreen.test.tsx`（6 pass）、P0.065-P0.069 顺序场景链、`pnpm --filter @easyinterview/frontend build`、`pnpm --filter @easyinterview/frontend exec playwright test tests/pixel-parity/debrief.spec.ts`（11 pass / 1 skip）通过；`make docs-check` / `git diff --check` 在收口 gate 重新执行。
- [x] 9.2 plans/INDEX.md 把 001 从 active 移到 completed，记录完成日期 2026-05-17
- [x] 9.3 frontend-debrief/history.md 增加 1.1 completion 行
- [x] 9.4 提交 commit `feat(frontend-debrief): close 001 debrief screen and handoff baseline`；记录工作日志 `/work-journal`
