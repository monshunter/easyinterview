# 001 Workspace + InterviewContext + Start Practice Contract Checklist

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-05-13

**关联计划**: [plan](./plan.md)

> L2 code review fix pass（2026-05-09）补齐 route → `InterviewContext` 自动 hydrate、Workspace modal 接线、`getPracticePlan` refresh/replacement、`autoStartPractice=1` 登录恢复、BDD 场景脚本强断言与 full workspace pixel parity。

> Handoff（2026-05-12）：`openapi-v1-contract/004-resume-additive-coverage` 已落地 `listResumes` operation + fixtures + generated client；本 completed plan 的 disabled-list 模式保持历史交付状态，后续 workspace owner 应原地修订 Resume Picker 为 active-list。

> Handoff（2026-05-13）：`backend-resume/001-asset-register-parse-and-listing` 已落地 `listResumes` 真实 `cmd/api` route、handler fixture parity 与 E2E.P0.034/P0.035；workspace owner 可启动 Resume Picker active-list 原地修订，移除 disabled-list 负向断言。关联提交：`1d1f69c feat(backend-resume): wire resume routes and listing`。

## Phase 1: WorkspaceScreen 静态壳 + 路由壳 + InterviewContext store + i18n（无数据）

- [x] 1.1 新增 `frontend/src/app/interview-context/InterviewContext.tsx`：`InterviewContextProvider` + `useInterviewContext()` + `useStartPracticeContext()`；reducer 支持 `HYDRATE_FROM_ROUTE` / `MERGE_TARGET_JOB` / `MERGE_RESUME` / `MERGE_PRACTICE_PLAN` / `MERGE_SESSION` / `CLEAR` / `CLEAR_RESUME` / `CLEAR_PRACTICE_PLAN` / `CLEAR_AUTO_START`；默认值与 plan §3.7 mapping 一致；同时在 `frontend/src/app/routes.ts` 新增并导出 `INTERVIEW_CONTEXT_ROUTES` / `shouldCarryInterviewContext(routeName)`，集合与 `ui-design/src/app.jsx` 保持 parity；`App.tsx` 通过 route sync 从 `workspace/practice/generating` route params 自动 dispatch `HYDRATE_FROM_ROUTE`，离开 carry route 时清空；新增 `interview-context/InterviewContext.test.tsx` 与 `App.test.tsx` 锁定 reducer、carry/clear 与真实 App route hydration 行为
- [x] 1.2 新增 `frontend/src/app/screens/workspace/WorkspaceScreen.tsx`，按 `ui-design/src/screen-workspace.jsx::WorkspaceScreen` lines 116-302 源级复刻 plan eyebrow + crumbs + header summary + Interview Launcher（Round Rail + 面试前确认 + CTA + BindingPill 双卡 + note line）+ Main Left（CompanyIntelEmbed placeholder + JD 拆解 placeholder）+ Main Right（risks/strengths placeholder + sessionHistory placeholder）；本 phase 全部动态字段渲染 placeholder skeleton；`startInterview` / 切换规划 / 更换简历 / 打开公司情报 / sessionHistory click 仅记录 nav stub；Vitest 断言 ≥ 20 个 testid 存在（按 plan §3.5 UI source structure parity rows）+ 控件类型断言（Resume Picker 必须是 list of buttons 而非 `<select>`，Plan Switcher 同理）
- [x] 1.3 在 `frontend/src/app/App.tsx` `renderRouteScreen` 中绑定 `workspace` → `<WorkspaceScreen route={route} />`，替换 D1 `PlaceholderScreen`；`practice` / `generating` 仍渲染 `PlaceholderScreen`，不在本 plan 改动；Vitest 断言 `App.tsx` 内 `workspace` route render 命中 `WorkspaceScreen` 而非 PlaceholderScreen，`practice` / `generating` 仍命中 PlaceholderScreen
- [x] 1.4 扩展 `frontend/src/app/i18n/locales/zh.ts` 与 `en.ts` 新增 `workspace.*` 命名空间（≥ 50 key 与 `screen-workspace.jsx::L` zh/en 字典等价：crumbs / overview / requirements / prep / practices / timeline / startCore / launchTitle / launchSub / flow / roundStatus / jdBound / resumeBound / changeResume / prepStatus / jdMatch / sessionTag / reportReady / planEyebrow / planSub / switchPlan / createPlan / must / nice / hidden / risks / strongs / lastReport / gotoReport / notePractice / empty.* / missingResume.* / planSwitcher.* / resumePicker.*）；`messages.ts` 类型聚合补齐；Vitest `i18n` 套件断言新 namespace zh/en 同步无缺漏
- [x] 1.5 新增 `workspace/WorkspaceScreen.test.tsx`：测 i18n zh/en 切换重绘、≥ 20 个 testid 存在、点击 placeholder 调用 nav/start stub、控件类型断言、负向断言旧 prototype testid（`practice-mode-card-*` / `growth-*` / `drill-builder-*` / `mistake-queue-*`）不命中
- [x] 1.6 BDD-Gate: 验证 `E2E.P0.018` 中 workspace 静态部分资产构建到 ready 态 <!-- verified: 2026-05-09 method=scenario bddChecklist=complete -->

## Phase 2: TargetJob 数据消费（listTargetJobs / getTargetJob）+ Header / Launcher / JD 拆解 / risks-strengths

- [x] 2.1 新增 `frontend/src/app/screens/workspace/hooks/useWorkspaceTargetJob.ts`：通过 D1 generated client 调 `getTargetJob(targetJobId)`；React state 跟踪 loading / data / error 三态；`InterviewContext.targetJobId` 缺失 → 立即返回 `empty` 状态（不发请求）；返回数据通过 `MERGE_TARGET_JOB` 写入 `InterviewContext`；Vitest `hooks/useWorkspaceTargetJob.test.tsx` 断言 generated client 调用次数 1、loading/data/error 三态、empty 短路
- [x] 2.2 在 `WorkspaceScreen` 中按 plan §3.7 mapping 把 generated `TargetJob` 字段注入：plan eyebrow 标题 / status tag / subtitle；header tag/updatedAt/title/subtitle；Interview Launcher Round Rail 使用 fallback `[HR, Technical 1, Technical 2, Manager]`，roundId 通过 `getWorkspaceRoundId` 同款 helper 派生；BindingPill JD 段只用 `title / companyName / locationText / sourceType` 与 route/context 派生字段；`prepStatus` / `jdMatch` 从 `fitSummary.strengths/gaps/riskSignals` 与 `openQuestionIssueCount` 派生（缺失时 fallback `—`）；不读取不存在的 `TargetJob.level` / `TargetJob.match` / `TargetJob.nextRound` / `TargetJob.statusTone` / `TargetJob.readinessLabel` 字段；Vitest 断言字段映射、负向不依赖 OpenAPI 未声明字段
- [x] 2.3 Main Left JD 拆解：按 `screen-workspace.jsx::ReqBlock` lines 742-759 源级复刻 `Must Have` / `Nice to Have` / `Hidden signals` 三 ReqBlock；数据来源 `TargetJob.requirements` declared fields + `TargetJob.summary` 文本；hits 命中圆点从 `fitSummary.strengths` view-model 派生，缺失时全部为空圆；不得使用未声明的 `summary.interviewHypotheses` / `summary.coreThemes` / `fitSummary.directHits`；Vitest 断言 `workspace-jd-block-{must,nice,hidden}` testid + 列表渲染 + hit 圆点切换
- [x] 2.4 Main Right risks/strengths：按 `screen-workspace.jsx::WorkspaceScreen` lines 224-239 源级复刻；数据来源 `TargetJob.fitSummary.strengths` + `TargetJob.fitSummary.riskSignals`，`gaps` 可并入风险/待补强列表（缺失时空态）；Vitest 断言 `workspace-prep-{strong,risk}-${idx}` testid + 列表渲染 + 缺失字段空态
- [x] 2.5 sessionHistory placeholder：本 phase 仅渲染 `EmptyHistory` placeholder + 文案 `首场面试将出现在这里`（zh/en）；点击行 disabled；Phase 5 再接 handoff 路径
- [x] 2.6 fixture variant：扩展 `openapi/fixtures/TargetJobs/getTargetJob.json` 新增 `with-rounds` / `not-found` variants；`listTargetJobs.json` `one-job` 已满足 `single-plan`；`make validate-fixtures` 通过
- [x] 2.7 新增 `workspace/WorkspaceHeader.test.tsx`：测 fixture 三态渲染、header 字段映射、JD 拆解数据驱动、risks/strengths 列表渲染
- [x] 2.8 BDD-Gate: 验证 `E2E.P0.018` 中 workspace 数据接入部分通过 <!-- verified: 2026-05-09 method=scenario -->

## Phase 3: 简历绑定 + Resume Picker + Plan Switcher Modal

- [x] 3.1 新增 `frontend/src/app/screens/workspace/hooks/useWorkspaceResume.ts`：通过 generated client 调 `getResume(resumeVersionId)`；React state 跟踪 loading / data / error 三态；缺失或 404 → 写入 `InterviewContext.resumeVersionId=null` 并触发 `WorkspaceMissingResumeState`；Vitest 断言 hook 行为
- [x] 3.2 BindingPill Resume 段：按 `screen-workspace.jsx::WorkspaceScreen` line 191 源级复刻；title=`getResume.title`，meta 通过 `readResumeSummary(parsedSummary: Record<string, unknown> | null)` type guard 安全读取 `headline / yearsOfExperience` 后拼接；点击 `更换` 打开 `ResumePickerModal`；Vitest 断言 testid `workspace-binding-resume` 与 `更换` 按钮，L2 复核补充 `WorkspaceModalIntegration.test.tsx` 覆盖正式 `WorkspaceScreen` 中按钮到 modal 的实际接线
- [x] 3.3 新增 `frontend/src/app/screens/workspace/modals/ResumePickerModal.tsx` (disabled-list 模式)：按 `screen-workspace.jsx::ResumePickerModal` lines 517-585 源级复刻 DOM + 模态层 + footer；列表项仅当前绑定 resume 启用 + 选中态；其余位置渲染 `disabled` 占位卡 + i18n 文案 `resumePicker.disabledNote` 指向 spec §3.2；testid `workspace-resume-modal-{card-${id},disabled-note,confirm,cancel,close}`；`Use this resume` 按钮在仅有当前绑定 resume 时关闭 modal；Vitest `workspace/modals/ResumePickerModal.test.tsx` 断言 DOM、disabled 列表、文案 zh/en、a11y、generated client `listResumes` 调用次数为 0
- [x] 3.4 新增 `frontend/src/app/screens/workspace/modals/PlanSwitcherModal.tsx`：按 `screen-workspace.jsx::PlanSwitcherModal` lines 587-666 源级复刻；通过新增 `useWorkspaceTargetJobs()` hook（调用 `listTargetJobs`，复用 home plan §3.7 viewmodel mapping）拉取候选 plan 列表；testid `workspace-plan-modal-{card-${id},create,confirm,cancel,close}`；`从新 JD 创建规划` 调 `nav("home")`；选择 plan 卡后由 `Use this plan` 确认切换 `InterviewContext` 并触发 `useWorkspaceTargetJob()` / `useWorkspaceResume()` / `useWorkspacePracticePlan()` 重新拉取；Vitest `workspace/modals/PlanSwitcherModal.test.tsx` 断言 DOM、`listTargetJobs` 接入、boundary（1 条 / 12+ 条）、a11y、CTA 行为
- [x] 3.5 新增 `frontend/src/app/screens/workspace/modals/useModalA11y.ts`：实现 ESC 关闭 / 外层遮罩点击关闭 / 关闭按钮 / focus trap（首次打开 focus 第一个 focusable 元素 + Tab 循环）/ 关闭后 focus 回到触发按钮；Vitest 用 `userEvent.tab()` 验证 + `aria-modal` attribute + Vitest 锁定四种关闭路径
- [x] 3.6 fixture variant：`getResume.json` 新增 `not-found`；`getPracticePlan.json` 新增 `archived` / `not-found`；`make validate-fixtures` 通过
- [x] 3.7 新增 `workspace/WorkspaceMissingResumeState.test.tsx` + `workspace/WorkspaceEmptyState.test.tsx`：测空态 CTA 跳转
- [x] 3.8 BDD-Gate: 验证 `E2E.P0.018` Plan Switcher / Resume Picker + `E2E.P0.019` `WorkspaceMissingResumeState` 路径 <!-- verified: 2026-05-09 method=scenario -->

## Phase 4: 立即面试双步契约 + getPracticePlan refresh + auth pendingAction

- [x] 4.1 新增 `frontend/src/app/screens/workspace/hooks/useWorkspacePracticePlan.ts`：mount 时若 `InterviewContext.planId` 存在 → 调 `getPracticePlan(planId)`；`status='ready'` → `MERGE_PRACTICE_PLAN`；`status='archived'` 或 404 → 重置 `InterviewContext.planId=null`（驱动后续 createPracticePlan）；不得假设 OpenAPI 未声明的 `cancelled/failed` plan status；Vitest 断言 ready / archived / not-found / 5xx 路径
- [x] 4.2 新增 `frontend/src/app/screens/workspace/hooks/useStartPractice.ts`：实现立即面试双步契约（plan §4.2 伪码）；`buildCreatePlanRequest(ctx)` 在 `frontend/src/app/interview-context/buildCreatePlanRequest.ts` 实现，body 字段映射含 `targetJobId / goal='baseline' / mode='assisted' / interviewerPersona='hiring_manager' / difficulty='standard' / language=ui-locale / questionBudget=6 / timeBudgetMinutes=30 / resumeAssetId=ctx.resumeVersionId / focusCompetencyCodes=[]`；`hintsEnabled` 由 `practiceMode==='assisted'` 派生；通过 `frontend/src/lib/conventions/idempotency.ts::newIdempotencyBatch()` 派生稳定 `{create, start}` 双键；`inFlightRef` + `Promise` 缓存防 StrictMode 双触发；retry 复用同一 batch；Vitest 断言 generated client 调用次数 = 1（StrictMode 下）+ retry 与首次同 `Idempotency-Key`
- [x] 4.3 立即面试 CTA 接线：WorkspaceScreen CTA 接入 `useStartPractice` + `handleStart`；成功导航使用 hook 返回的最新 `planId`，覆盖 archived/404 后重新创建 plan 的 replacement path。
- [x] 4.4 ButtonState：loading 态 disabled+spinner、error 态 inline 提示+retry、3 次失败 fallback CTA
- [x] 4.5 错误映射：error message 展示 + retry button；retry 复用 idempotencyBatch
- [x] 4.6 fixture variant：`startPracticeSession.json` 新增 `ai-timeout-502`；`make validate-fixtures` 通过
- [x] 4.7 新增 `workspace/WorkspaceStartPractice.test.tsx`：测 happy path（无 plan → createPracticePlan → startPracticeSession → nav practice）；测 happy path（有 plan + ready → 跳过 createPracticePlan）；测 happy path（有 plan + archived → 重新 createPracticePlan）；测 createPracticePlan 4xx + 5xx；测 startPracticeSession 5xx + retry；测 `Idempotency-Key` retry 复用；测 nav practice 携带完整 InterviewContext + `PracticeDisplayContext`（含 `practiceGoal`）字段；测 hintsEnabled 由二值 practiceMode 派生，并负向断言 workspace 不产出 `legacy debrief replay value`
- [x] 4.8 新增 `workspace/WorkspaceAuthGate.test.tsx`：测未登录立即面试 → `requestAuth` 触发 → `auth_login` 携带 `pendingRoute=workspace` / `pendingType=start_practice` / `autoStartPractice=1` → 登录恢复 workspace → route hydrate `InterviewContext` → 清理 `autoStartPractice` → 自动 startPractice → nav practice；测 pendingAction.params 仅含 IDs / route / `PracticeDisplayContext` / `autoStartPractice` 结构化字段，不含敏感字段
- [x] 4.9 BDD-Gate: 验证 `E2E.P0.020` 立即面试 + 未登录恢复 + `E2E.P0.019` getPracticePlan 恢复 <!-- verified: 2026-05-09 method=scenario -->

## Phase 5: CompanyIntelEmbed handoff + Session History handoff + 空态收口

- [x] 5.1 新增 `frontend/src/app/screens/workspace/CompanyIntelEmbed.tsx`：数据仅限 getTargetJob 字段，不调 getCompanyIntel
- [x] 5.2 sessionHistory placeholder：Phase 2 已实现 `EmptyHistory` / disabled placeholder
- [x] 5.3 WorkspaceEmptyState / WorkspaceMissingResumeState 收口：CTA 跳转已实现（home / resume_versions?flow=create）
- [x] 5.4 新增 `workspace/WorkspaceHandoff.test.tsx`：测 CompanyIntelEmbed 不调 `getCompanyIntel`、handoff 携带 `targetJobId / jdId`；测 sessionHistory 为 `EmptyHistory` / disabled placeholder 且点击不触发 report nav；测不读取 `TargetJob.recentSessions`、不调用 `getFeedbackReport`；测两空态 CTA 跳转
- [x] 5.5 BDD-Gate: 验证 `E2E.P0.021` handoff 主路径 + 隐私红线 + 旧入口反向 grep <!-- verified: 2026-05-09 method=scenario -->

## Phase 6: 验证收口（pixel parity + scenario + regression rerun）

- [x] 6.1 新增 `frontend/tests/pixel-parity/workspace.spec.ts` 覆盖 desktop (1440×900) + mobile (390×844) 两 chromium project：fixture-backed App route hydration、DOM 锚点（plan eyebrow / header summary / Round Rail / Interview Launcher CTA / BindingPill 双卡 / Main Left CompanyIntelEmbed + JD 拆解 / Main Right risks-strengths + EmptyHistory / 两 modal）+ bounding box stays in viewport, no overlap + warm/light → dark → customAccent 三态切换 computed 颜色变化 + toHaveScreenshot baseline；mobile 断言 Main 折单列、Round Rail 横向滚动、Modal 全屏化
- [x] 6.2 `pnpm --filter @easyinterview/frontend test:pixel-parity` 在 D2/D3 + home plan 现有基础上累加 workspace 新增 spec；96/96 PASS
- [x] 6.3 派生 4 个 scenario 目录 `test/scenarios/e2e/p0-018` ~ `p0-021`，各含 README.md + scripts/{setup,trigger,verify,cleanup}.sh + data/seed-input.md + data/expected-outcome.md
- [x] 6.4 `test/scenarios/e2e/INDEX.md` P0 表追加 4 行（P0.018-P0.021），关联需求 frontend-workspace-and-practice C-1~C-12，状态 Ready，automated
- [x] 6.5 Regression 重跑：`P0.001/002/004/005` 由 full Vitest + full pixel parity 等价覆盖；P0.006 pixel parity 96/96 PASS；workspace `P0.018/019/020/021` 均 `setup -> trigger -> verify -> cleanup` PASS；`P0.014/015/016/017` 场景目录存在但属 home plan active gate，记录为条件 gate（当前不适用）
- [x] 6.6 全量验证：`pnpm --filter @easyinterview/frontend test` (67 files, 432 tests PASS)、`pnpm --filter @easyinterview/frontend build` (PASS，包含 `tsc --noEmit` + `vite build`)；`make build` 未作为本次 L2 code review fix 的必要 gate 执行
- [x] 6.7 文档与索引同步：checklist、bdd-checklist 与 plans INDEX 已更新至最新；`make docs-check` + `/sync-doc-index --fix-index` post-fix zero drift gate 执行
- [x] 6.8 负向搜索：prototype helper imports、旧 testid、旧 route alias、JD 原文/简历正文泄漏、AI/LLM 直接调用、listResumes/getCompanyIntel/getFeedbackReport 运行时调用 → 全部 0 命中（仅注释与测试断言命中）
- [x] 6.9 BDD-Gate: 验证 `E2E.P0.018/019/020/021` 全部 `setup -> trigger -> verify -> cleanup` PASS；D1+D2+D3 regression 由 full Vitest + full pixel parity 覆盖；`P0.006` pixel parity 96/96 PASS；home plan `P0.014-017` 条件 gate 当前不适用 <!-- verified: 2026-05-09 method=scenario -->

## L2 Code Review Remediation（2026-05-09）

- [x] D-L2-001 route params 自动 hydrate：`App.tsx` 在 carry routes 内 dispatch `HYDRATE_FROM_ROUTE`，并在离开 carry route 时清空 `InterviewContext`；`App.test.tsx` 覆盖真实 App path。
- [x] P-L2-002 modal 接线：`WorkspaceScreen` 正式挂载 `PlanSwitcherModal` / `ResumePickerModal`；按钮点击到 modal DOM 由 `WorkspaceModalIntegration.test.tsx` 与 pixel parity modal anchors 覆盖。
- [x] P-L2-003 practice plan refresh/replacement：`useWorkspacePracticePlan` 与 `useStartPractice` 均处理 ready / archived / 404，archived/404 清 plan 后走 `createPracticePlan` replacement path，并用最新返回 planId 导航 practice。
- [x] P-L2-004 `autoStartPractice=1` 恢复：登录回到 workspace 后自动启动 startPractice，启动前清理控制位，pendingAction 敏感字段负向断言通过。
- [x] C-L2-005 BDD 场景脚本强断言：P0.018-P0.021 trigger/verify 脚本补齐真实 Vitest entry、runtime negative grep 与 completion marker。
- [x] D-L2-006 full workspace pixel parity：workspace pixel spec 从空态扩展到 fixture-backed full workspace、两 modal、desktop/mobile bounding boxes 与 screenshots；full `test:pixel-parity` 96/96 PASS。

## L2 Code Review Follow-up Remediation（2026-05-09）

- [x] P-L2-007 server-bound id normalization：synthetic `plan-${targetJobId}` / `resume-unbound` / invalid UUID route params must be treated as absent before `getPracticePlan` / `getResume` / `createPracticePlan` / `startPracticeSession`; Vitest must assert generated client methods are not called with synthetic ids. <!-- verified: 2026-05-09 method=vitest files=buildCreatePlanRequest.test.ts,useWorkspacePracticePlan.test.tsx,useWorkspaceResume.test.tsx,WorkspaceStartPractice.test.tsx -->
- [x] P-L2-008 target-job stale/error recovery：`getTargetJob` 404/5xx must render workspace empty/error recovery instead of the full workspace shell, and target changes must key/ignore stale in-flight completions; Vitest must cover stale completion ordering. <!-- verified: 2026-05-09 method=vitest files=useWorkspaceTargetJob.test.tsx,WorkspaceEmptyState.test.tsx -->
- [x] P-L2-009 workspace label localization：JD block labels, round fallback labels, target status labels, source labels, and derived prep labels must resolve through `workspace.*` locale keys; English Vitest must assert Chinese labels are absent. <!-- verified: 2026-05-09 method=vitest files=WorkspaceScreen.test.tsx,WorkspaceHeader.test.tsx -->
