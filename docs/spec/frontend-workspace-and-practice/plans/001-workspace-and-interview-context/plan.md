# 001 Workspace + InterviewContext + Start Practice Contract

> **版本**: 1.17
> **状态**: completed
> **更新日期**: 2026-07-09

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

本 plan 已完成 `workspace` 当前面试规划页、`InterviewContext`、规划/简历切换、立即面试启动和未登录恢复合同。本次 v1.12 原地修订追加一级 `面试` 入口优化：`workspace` 无上下文时不再展示缺 JD 死胡同，而是展示面试规划列表；带上下文时进入统一面试规划详情。
本次 v1.13 原地修订追加面试规划列表卡片化视觉收口：列表项必须是清晰可感知的卡片，而不是无容器的文本列。
本次 v1.14 原地修订追加面试规划列表卡片简化：卡片不得展示 `sourceType` / `targetLanguage` 等低价值技术字段，底部只保留主题色进入 CTA，并使用现有卡片 / 边框 / elevation token 拉开与页面背景的层次。
本次 v1.15 原地修订修复 `进入规划` 回归：列表卡片必须消费后端投影的当前 practice plan / bound resume 信息，导航到统一面试规划详情时携带真实 `planId` / `resumeId`，不得回落到缺简历空态或合成不存在的 plan/resume id。
本次 v1.16 原地修订修复方案 A 数据持久化缺口：创建 JD/规划时选中的简历必须作为 `target_jobs.resume_id` 持久化；`listTargetJobs` / `getTargetJob` 的 `resumeId` 优先来自 target job 级绑定，列表卡片重新进入规划时即使尚未创建 `practice_plans` 也必须携带真实 `resumeId`。
本次 v1.17 原地修订收敛规划详情体验：`workspace` 无上下文继续展示面试规划列表；带上下文普通回访复用 `frontend-home-job-picks-and-parse` 的 Parse-derived “面试规划详情 / 面试上下文确认”母版，不再渲染独立 workspace Header / Launcher / JD card 第二套详情页；`autoStartPractice=1` 仍由 workspace owner 执行 create/start session。

- TopBar `workspace` 文案改为 `面试` / `Interview`，route/testid 仍保持 `workspace`。
- `workspace` 无 `targetJobId` / `planId` 时渲染面试规划列表，使用 generated `listTargetJobs`，点击卡片进入统一面试规划详情。
- 面试规划列表每个 plan item 必须具备独立卡片容器、卡片背景、1px 边框、轻阴影、内部分区和底部操作区；desktop 使用响应式多列，mobile 折叠为单列。
- 面试规划列表卡片只展示对继续规划有决策价值的信息：状态、更新时间、岗位、公司和地点；不展示 `手动输入` / 来源类型 / 目标语言等导入元信息；`进入规划` 必须使用主题强调色按钮。
- 面试规划列表卡片进入统一面试规划详情时必须使用 `listTargetJobs` 返回的当前 `currentPracticePlanId` / `resumeId`；`resumeId` 是 target job 创建时的持久绑定，若当前还没有 ready practice plan，也必须携带该 `resumeId` 进入详情页。
- `workspace` route 无上下文渲染正式 `WorkspacePlanList`；带上下文普通回访复用统一详情母版；`autoStartPractice=1` 仍渲染/执行 workspace 启动合同并跳转 practice。
- `InterviewContext` 在 `workspace / practice / generating` owner route 内携带 `targetJobId / jdId / resumeId / roundId / planId / practiceMode / practiceGoal / hintUsed / hintCount`；离开 owner route 清理上下文。
- Plan Switcher 使用 generated `listTargetJobs` 切换当前规划；Resume Picker 使用当前 flat `listResumes` active-list 选择绑定简历。
- `立即面试` 使用 generated `getPracticePlan` / `createPracticePlan` / `startPracticeSession`，副作用请求带 `Idempotency-Key`；未登录时通过 `requestAuth({ type: "start_practice" })` 回到 `workspace` 后自动续跑。
- `WorkspaceInsightCard` 只作为 workspace 内嵌摘要存在；点击停留在 `workspace` 并携带 safe params，不调用独立公司信号 API。
- 当前规划记录区只展示 typed records placeholder，不从 `TargetJob` fixture extension、`any` 或 report API 拼接记录行。
- JD 原文、简历正文、题目文本、答案、提示、prompt/response 不进入 URL、localStorage、console 或 fixture transport 日志。

## 2 当前合同

### 2.1 Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|----------|
| `listTargetJobs` | `openapi/fixtures/TargetJobs/listTargetJobs.json` | `WorkspacePlanList` 一级面试规划列表 landing / `PlanSwitcherModal` 候选规划列表 | `backend-targetjob/001` | `target_jobs.resume_id` + optional latest ready `practice_plans` | none | `E2E.P0.018` |
| `getTargetJob` | `openapi/fixtures/TargetJobs/getTargetJob.json` | Workspace header、Launcher、JD 拆解、准备信号、空态 | `backend-targetjob/001` | `target_jobs.resume_id` + requirements/sources + optional latest ready `practice_plans` | none | `E2E.P0.018`, `E2E.P0.019` |
| `getResume` | `openapi/fixtures/Resumes/getResume.json` | 绑定简历摘要与缺简历空态 | `backend-resume/001` | `resume_assets` | none | `E2E.P0.018`, `E2E.P0.019` |
| `listResumes` | `openapi/fixtures/Resumes/listResumes.json` | `ResumePickerModal` flat active-list | `backend-resume/001` | `resume_assets` | none | `E2E.P0.018` |
| `getPracticePlan` | `openapi/fixtures/PracticePlans/getPracticePlan.json` | workspace refresh / ready-plan reuse / not-found recovery | `backend-practice/001` | `practice_plans` | none | `E2E.P0.019`, `E2E.P0.020` |
| `createPracticePlan` | `openapi/fixtures/PracticePlans/createPracticePlan.json` | 无可用 plan 时创建 baseline plan | `backend-practice/001` | `practice_plans` | backend-only first question prep | `E2E.P0.020` |
| `startPracticeSession` | `openapi/fixtures/PracticeSessions/startPracticeSession.json` | 启动 practice session 并 handoff 到 `practice` | `backend-practice/001` | `practice_sessions` + first turn | backend-only `practice.session.first_question` | `E2E.P0.020` |
| `getFeedbackReport` | N/A | 本 plan 不消费；records placeholder 不拼 report row | external owner | external | none | `E2E.P0.021` negative |
| `updateTargetJob` | `openapi/fixtures/TargetJobs/updateTargetJob.json` | 统一详情母版 Save/Start 前保存允许编辑字段 | `backend-targetjob/001` | `target_jobs` partial update | none | `E2E.P0.016`, `E2E.P0.018` shared detail |

### 2.2 UI / Route Boundary

- `workspace` 保留 App chrome；`practice` 和 `generating` 由下游 owner 隐藏 chrome。
- TopBar 显示 `面试` / `Interview`；`workspace` 无上下文时展示面试规划列表，带上下文时复用统一面试规划详情母版。
- 当前面试规划不展示练习模式卡片、成长中心、单题深钻、专项练习、独立 voice route 或独立公司信号页面。
- `resumeId` 是当前简历绑定键；`resumeVersionId` 不作为本 plan 正向 route/context 字段。
- Records placeholder 只说明当前规划下的模拟面试记录区域存在；真实记录行必须来自 typed records contract owner。

## 3 质量门禁

- **Plan 类型**: `feature-behavior + contract + frontend-ui + BDD`。
- **TDD 策略**: 适用。Vitest 覆盖 route hydration、InterviewContext reducer、WorkspaceScreen DOM anchors、Plan/Resume modal、generated client body/header、auth pendingAction、privacy and non-current negative gates。
- **BDD 策略**: 适用。`E2E.P0.018` - `E2E.P0.021` 覆盖 workspace 默认渲染、context loading、立即面试、embedded-only handoff、privacy 和 non-current negative gates。
- **替代验证 gate**:
  - `pnpm --filter @easyinterview/frontend test src/app/App.test.tsx src/app/screens/workspace/WorkspaceScreen.test.tsx src/app/screens/workspace/WorkspaceHeader.test.tsx src/app/screens/workspace/WorkspaceModalIntegration.test.tsx src/app/screens/workspace/modals/PlanSwitcherModal.test.tsx src/app/screens/workspace/modals/ResumePickerModal.test.tsx src/app/screens/workspace/modals/useModalA11y.test.tsx`
  - `pnpm --filter @easyinterview/frontend test src/app/screens/workspace/WorkspaceHandoff.test.tsx src/app/screens/workspace/WorkspaceScreen.test.tsx`
  - `pnpm --filter @easyinterview/frontend test src/app/screens/workspace/WorkspaceStartPractice.test.tsx src/app/screens/workspace/WorkspaceAuthGate.test.tsx`
  - `pnpm --filter @easyinterview/frontend test:pixel-parity`
  - `make validate-fixtures`
  - `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/frontend-workspace-and-practice/plans/001-workspace-and-interview-context/context.yaml --target frontend`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `make docs-check`

## 4 实施结果

### Phase 0: Contract preflight

- Confirmed `docs/development.md` §2 frontend/backend contract workflow.
- Confirmed UI truth source: `docs/ui-design/module-job-workspace.md`, `ui-design/src/screen-workspace.jsx`, `ui-design/src/app.jsx`, `ui-design/src/primitives.jsx`.
- Confirmed generated client operations and fixtures listed in §2.1.

### Phase 1: Workspace shell and InterviewContext

- Added `InterviewContextProvider`, reducer actions, route hydration and owner-route carry/clear behavior.
- Replaced `workspace` placeholder with `WorkspaceScreen`; kept `practice` / `generating` for their owner plans.
- Added zh/en `workspace.*` messages and DOM anchor coverage.

### Phase 2: TargetJob, resume and workspace data

- Wired `getTargetJob` into header, launcher, JD breakdown, preparation signals and empty/error states.
- Wired `getResume` into bound resume summary and missing-resume state.
- Preserved source-level layout parity for desktop and mobile.

### Phase 3: Plan and resume switching

- Added `PlanSwitcherModal` backed by `listTargetJobs`.
- Added `ResumePickerModal` backed by flat `listResumes` active-list.
- Added modal keyboard close, backdrop close, close button, focus trap and focus return.

### Phase 4: Start practice and auth recovery

- Added `getPracticePlan` refresh and ready/not-found recovery.
- Added `useStartPractice` two-step launch with idempotency batches and retry behavior.
- Added `requestAuth(start_practice)` pendingAction flow and `autoStartPractice=1` recovery.

### Phase 5: Embedded insight, records placeholder and privacy

- Added `WorkspaceInsightCard` embedded-only behavior.
- Kept records area as disabled placeholder until typed records consumer lands.
- Added negative coverage for report API calls, untyped fixture extensions, prototype helpers, sensitive fields and non-current runtime routes/testids.

### Phase 6: Verification closeout

- Frontend focused tests, workspace pixel parity, BDD scenario scripts, fixture validation, docs/index checks and negative grep gates passed in owner closeout.

### Phase 7: Interview nav and plan-list landing revision

- Update product/UI truth sources and static prototype so TopBar uses `面试` / `Interview` and `workspace` separates plan-list landing from current-plan detail.
- Add `WorkspacePlanList` backed by generated `listTargetJobs`; clicking a plan card navigates to `workspace?targetJobId=...` without fabricating resume or report data.
- Preserve current detail page behavior for route contexts that include `targetJobId` / `planId`; not-found detail and missing resume states remain separate from the no-context list landing.
- Update source-structure, i18n, route, scenario and pixel-parity tests so TopBar click proves the list landing and hydrated route still proves the current-plan detail.

### Phase 8: Plan-list card visual hardening

- Reopen the completed owner after screenshot review because Phase 7 verified list presence but did not assert visual card affordance.
- Update `docs/ui-design/module-job-workspace.md` and `ui-design/src/screen-workspace.jsx` so the plan list card contract includes visible card background, 1px border, subtle elevation, internal body/footer sections and responsive card grid.

### Phase 10: Plan-list bound resume navigation remediation

- Reopen the completed owner after runtime bug report because Phase 7-9 preserved card visuals but allowed list card navigation to omit the existing practice-plan resume binding.
- Add regression coverage that `WorkspacePlanList` uses declared TargetJob projection fields for current plan id and bound resume id, and does not synthesize `plan-${targetJobId}` or `resume-unbound`.
- Keep the no-binding fallback explicit: route carries only target job identity and detail page owns missing-resume recovery.
- Verified focused frontend suites and typecheck on 2026-07-08.

### Phase 11: TargetJob-level resume binding remediation

- Reopen the completed owner because runtime evidence showed imported target jobs can have no `practice_plans` rows yet, so deriving `resumeId` only from latest ready practice plan still loses the resume selected during Home import.
- Require `importTargetJob` to persist the selected `resumeId` to `target_jobs.resume_id`; `TargetJob.resumeId` must expose this target job binding even before a practice plan exists.
- Update `WorkspacePlanList` tests so a list item with `resumeId` but no `currentPracticePlanId` opens current-plan detail with the real resume binding and does not show the missing-resume state.
- Update formal `WorkspacePlanList` to source-level mirror that card treatment while keeping generated `listTargetJobs` data flow and safe plan navigation unchanged.
- Add jsdom and Playwright pixel-parity assertions that fail when plan items render as loose text columns or lose card elevation/sectioning.

### Phase 12: Unified detail route remediation

- Reopen the completed owner because user review confirmed that the standalone workspace detail and Parse result page duplicate the same JD / resume / round confirmation job.
- Keep `WorkspacePlanList` as the no-context `面试` landing, but route ordinary `workspace?targetJobId=...` detail traffic into the Parse-derived “面试规划详情 / 面试上下文确认” mother page.
- Preserve workspace-owned `autoStartPractice=1`: after unified detail Start navigates to workspace with auto-start, `useStartPractice` still performs `getPracticePlan` / `createPracticePlan` / `startPracticeSession` with idempotency and auth recovery.
- Add regression coverage that the old independent workspace detail anchors (`workspace-header`, `workspace-launcher`, `workspace-jd-card`, `workspace-prep-card`, `workspace-history-card`) do not appear on ordinary detail re-entry, while `workspace-plan-list` and start-practice gates remain live.

### Phase 9: Plan-list card simplification and theme consistency

- Reopen the completed owner after screenshot review because Phase 8 still rendered low-value source/language metadata and a secondary CTA that visually competed with the theme.
- Update `docs/ui-design/module-job-workspace.md` and `ui-design/src/screen-workspace.jsx` so no-context plan cards are concise: status + updated date, title, company/location, and a theme accent `进入规划` / `Open plan` CTA only.
- Update formal `WorkspacePlanList` to remove source/language display, keep footer as a minimal action row, and strengthen separation from page background via existing card/rule/elevation tokens.
- Add jsdom, Playwright and E2E.P0.018 assertions that fail when `workspace.planList.cardMeta`, source labels, target language text, or secondary button styling return to the card.

## 5 验收标准

| ID | 验收点 | 验证 |
|----|--------|------|
| A-1 | `workspace` route renders current planning page and preserves UI parity | WorkspaceScreen/Header/modal tests, pixel parity, `E2E.P0.018` |
| A-2 | InterviewContext route hydration and carry/clear behavior are deterministic | `InterviewContext.test.tsx`, `App.test.tsx` |
| A-3 | Plan Switcher consumes `listTargetJobs`; Resume Picker consumes flat `listResumes` active-list | modal tests, `E2E.P0.018` verify |
| A-4 | Start practice runs `getPracticePlan` / `createPracticePlan` / `startPracticeSession` with idempotency and auth recovery | `WorkspaceStartPractice.test.tsx`, `WorkspaceAuthGate.test.tsx`, `E2E.P0.020` |
| A-5 | Workspace insight stays embedded and records placeholder does not synthesize report rows | `WorkspaceHandoff.test.tsx`, `E2E.P0.021` |
| A-6 | Privacy and non-current route/module gates have zero runtime residuals | scenario verify scripts, pruning-surface lint |
| A-7 | TopBar shows `面试` / `Interview`; no-context `workspace` shows a plan list landing; plan cards open current-plan detail | TopBar tests, WorkspaceScreen tests, `E2E.P0.018`, pixel parity workspace spec |
| A-8 | Plan-list landing visually renders as list cards, not loose text columns | `WorkspaceEmptyState.test.tsx`, `frontend/tests/pixel-parity/workspace.spec.ts`, `E2E.P0.018` |
| A-9 | Plan-list cards stay concise and theme-consistent: no source/language metadata, accent CTA, clear card/page separation | `WorkspaceEmptyState.test.tsx`, `frontend/tests/pixel-parity/workspace.spec.ts`, `E2E.P0.018` |
| A-10 | Target job import persists selected resume binding and workspace plan-list re-entry carries `resumeId` even before any `practice_plans` row exists | backend targetjob tests, Home import tests, Workspace plan-list regression, local API smoke |
| A-11 | Workspace list card re-entry renders the unified Parse-derived plan detail and no longer shows the independent workspace detail page; `autoStartPractice=1` still launches practice through workspace start contract | `WorkspaceScreen.test.tsx`, `WorkspaceHandoff.test.tsx`, `frontend/tests/pixel-parity/workspace.spec.ts`, `E2E.P0.018`, `E2E.P0.020` |

## 6 变更记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-09 | 1.16 | Reopen owner plan for target job-level resume binding persistence so workspace plan-list re-entry no longer loses the resume selected during JD import. |
| 2026-07-09 | 1.17 | Reopen owner plan to route workspace current-plan detail into the unified Parse-derived Interview Plan Detail / Context Confirm mother page while preserving workspace start-practice ownership. |
| 2026-07-08 | 1.14 | Reopen owner plan for plan-list card simplification, metadata removal and theme-consistent CTA styling after screenshot review. |
| 2026-07-08 | 1.13 | Reopen owner plan for interview plan-list card visual hardening after screenshot review. |
| 2026-07-08 | 1.12 | Reopen owner plan for Interview nav naming and workspace plan-list landing revision. |
| 2026-07-07 | 1.11 | Compress owner docs to the current workspace, flat Resume Picker, start-practice, embedded insight, records placeholder and privacy contract. |
| 2026-07-07 | 1.10 | Reconcile workspace owner handoff, completion, fixture and active-list boundary wording. |
