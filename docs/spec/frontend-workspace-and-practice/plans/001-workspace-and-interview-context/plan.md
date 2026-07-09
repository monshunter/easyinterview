# 001 Workspace + InterviewContext + Start Practice Contract

> **版本**: 1.25
> **状态**: active
> **更新日期**: 2026-07-09

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

本 plan 当前完成 `workspace` 纯面试规划列表、ready TargetJob 列表准入、stale context 清理和 `parse` 详情跳转合同。历史上由 workspace 承接的规划/简历切换、立即面试启动和未登录恢复合同已在 v1.19 退出 workspace owner。本次 v1.12 原地修订追加一级 `面试` 入口优化：`workspace` 无上下文时不再展示缺 JD 死胡同，而是展示面试规划列表。
本次 v1.13 原地修订追加面试规划列表卡片化视觉收口：列表项必须是清晰可感知的卡片，而不是无容器的文本列。
本次 v1.14 原地修订追加面试规划列表卡片简化：卡片不得展示 `sourceType` / `targetLanguage` 等低价值技术字段，底部只保留主题色进入 CTA，并使用现有卡片 / 边框 / elevation token 拉开与页面背景的层次。
本次 v1.15 原地修订修复 `进入规划` 回归：列表卡片必须消费后端投影的当前 practice plan / bound resume 信息，导航到统一面试规划详情时携带真实 `planId` / `resumeId`，不得回落到缺简历空态或合成不存在的 plan/resume id。
本次 v1.16 原地修订修复方案 A 数据持久化缺口：创建 JD/规划时选中的简历必须作为 `target_jobs.resume_id` 持久化；`listTargetJobs` / `getTargetJob` 的 `resumeId` 优先来自 target job 级绑定，列表卡片重新进入规划时即使尚未创建 `practice_plans` 也必须携带真实 `resumeId`。
本次 v1.17 原地修订曾将 workspace 普通回访复用 `frontend-home-job-picks-and-parse` 的 Parse-derived 母版；该过渡合同已被 v1.19 取代，当前 workspace 不再渲染详情或执行 create/start session。
本次 v1.18 原地修订修复面试列表准入与顶栏导航回归：no-context workspace 必须只以当前 route params 判定，不得继承 stale `InterviewContext`；列表请求 `listTargetJobs` 必须带 `analysisStatus=ready`，并过滤 failed / 空标题 TargetJob，防止解析失败脏数据进入面试列表。
本次 v1.19 原地修订收敛 workspace route purity：`workspace` 是纯列表页，不再承接 `targetJobId/planId/resumeId/autoStartPractice` 参数上下文；规划卡片导航 `parse`，`parse` / report owner 直接创建 practice plan / session。
本次 v1.20 原地修订修复面试列表卡片规格回归：desktop plan-list grid 必须使用固定最大列宽，1/2/3 张卡片的规格保持稳定，不得因单卡数量被拉伸为整行宽卡。
本次 v1.21 原地修订融合 Home 最近模拟面试与 workspace 面试列表卡片：workspace 卡片必须复用 Home recent card 的主体结构、公司/状态 eyebrow、岗位/地点层级和 mini round rail。本次 v1.22 原地修订把列表卡片的 `进入规划` 可见 footer CTA 改为点击卡片主体承接，并增加 `立即面试` 主按钮和使用简历列表 trash 图标样式的删除能力；Home recent 复用同一卡片动作模型但不展示删除按钮。
本次 v1.23 原地修订把删除按钮从本地列表隐藏升级为 generated `archiveTargetJob` 持久软归档；删除成功后卡片移除，刷新后不得回灌，删除失败时保留卡片并展示可恢复错误。本次 v1.24 原地修订把删除图标移到卡片右上角，footer 只保留 `立即面试` 主按钮。v1.25 review remediation 要求 workspace list quick-start 必须把结构化 `roundId/roundName` 带入 practice route。

- TopBar `workspace` 文案改为 `面试` / `Interview`，route/testid 仍保持 `workspace`。
- `workspace` 始终渲染面试规划列表，使用 generated `listTargetJobs(analysisStatus=ready)`，点击卡片主体导航 `parse` 统一面试规划详情。
- 面试规划列表每个 plan item 必须具备独立卡片容器，并以 Home 最近模拟面试卡片主体为主：公司/状态 eyebrow、岗位、地点和 mini round rail 保持同源；desktop 使用固定最大列宽的响应式多列，1/2/3 张卡片规格保持稳定，mobile 折叠为单列；workspace 只在同一卡片底部追加 `立即面试` 主按钮，删除图标固定在卡片右上角。
- 面试规划列表卡片只展示对继续规划有决策价值的信息：状态、更新时间、岗位、公司和地点；不展示 `手动输入` / 来源类型 / 目标语言等导入元信息；不展示可见的 `进入规划` footer button。
- 面试规划列表只展示已解析成功且具备岗位标题的 TargetJob：generated `listTargetJobs` 请求必须带 `analysisStatus=ready`，UI 层必须防御性排除 failed / processing / queued / 空标题记录。
- 顶栏 `面试` 或 legacy `/workspace?targetJobId=...` 都必须 canonicalize 为 `/workspace`、清理 stale InterviewContext 并展示面试规划列表，不得渲染缺目标岗位错误页。
- 面试规划列表卡片进入统一面试规划详情时必须使用 `listTargetJobs` 返回的当前 `currentPracticePlanId` / `resumeId`；`resumeId` 是 target job 创建时的持久绑定，若当前还没有 ready practice plan，也必须携带该 `resumeId` 进入详情页。
- `workspace` 不渲染统一详情母版，不拥有 `autoStartPractice` route side effect；列表页 `立即面试` 使用 shared generated practice handoff (`getPracticePlan` / `createPracticePlan` / `startPracticeSession`) 显式启动 session，并携带 saved TargetJob 的 `targetJobId/resumeId/currentPracticePlanId` 与结构化 `roundId/roundName`。
- 列表页删除图标使用 generated `archiveTargetJob` 和 `Idempotency-Key` 持久软归档 TargetJob；成功后从当前列表移除，失败时不导航、不隐藏卡片，并展示错误；不得继续使用本地-only hidden set 作为删除合同。
- `InterviewContext` 不在 `workspace` route carry；`practice / generating / report` owner route 按各自最小上下文携带 `targetJobId / jdId / resumeId / roundId / planId / practiceMode / practiceGoal / hintUsed / hintCount`。
- Plan Switcher、Resume Picker、WorkspaceInsightCard 和 workspace start-practice hooks 已退出本 owner 当前 runtime。
- 当前规划记录区只展示 typed records placeholder，不从 `TargetJob` fixture extension、`any` 或 report API 拼接记录行。
- JD 原文、简历正文、题目文本、答案、提示、prompt/response 不进入 URL、localStorage、console 或 fixture transport 日志。

## 2 当前合同

### 2.1 Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|----------|
| `listTargetJobs` | `openapi/fixtures/TargetJobs/listTargetJobs.json` | `WorkspacePlanList` 一级面试规划列表 landing | `backend-targetjob/001` | `target_jobs.resume_id` + optional latest ready `practice_plans` | none | `E2E.P0.018` |
| `archiveTargetJob` | `openapi/fixtures/TargetJobs/archiveTargetJob.json` | `WorkspacePlanList` 删除图标 | `backend-targetjob/001` Phase 12 | `target_jobs.status='archived'` + `deleted_at` | none | `E2E.P0.018` persistent delete gate |
| `getTargetJob` | `openapi/fixtures/TargetJobs/getTargetJob.json` | Parse unified detail owner, not workspace | `backend-targetjob/001` | `target_jobs.resume_id` + requirements/sources + optional latest ready `practice_plans` | none | parse owner + P0.018 focused gate |
| `getResume` | `openapi/fixtures/Resumes/getResume.json` | Parse / resume owners only | `backend-resume/001` | `resume_assets` | none | external owner gates |
| `listResumes` | `openapi/fixtures/Resumes/listResumes.json` | Home select + Parse bound resume display / resume workshop | `backend-resume/001` | `resume_assets` | none | parse owner gate |
| `getPracticePlan` | `openapi/fixtures/PracticePlans/getPracticePlan.json` | Workspace list quick start and parse/report start handoff validate existing plan context | `backend-practice/001` | `practice_plans` | none | workspace + parse/report focused gates |
| `createPracticePlan` | `openapi/fixtures/PracticePlans/createPracticePlan.json` | Workspace list quick start and parse/report handoff create baseline / retry / next-round plan when needed | `backend-practice/001` | `practice_plans` | backend-only first question prep | workspace + parse/report focused gates |
| `startPracticeSession` | `openapi/fixtures/PracticeSessions/startPracticeSession.json` | Workspace list quick start and parse/report handoff start practice and navigate `practice` | `backend-practice/001` | `practice_sessions` + first turn | backend-only `practice.session.first_question` | workspace + parse/report focused gates |
| `getFeedbackReport` | N/A | 本 plan 不消费；report owner handles replay/next-round CTA | external owner | external | none | external owner gates |

### 2.2 UI / Route Boundary

- `workspace` 保留 App chrome；`practice` 和 `generating` 由下游 owner 隐藏 chrome。
- TopBar 显示 `面试` / `Interview`；`workspace` 始终展示面试规划列表，详情入口是 `parse`。
- 当前面试规划不展示练习模式卡片、成长中心、单题深钻、专项练习、独立 voice route 或独立公司信号页面。
- `resumeId` 是当前简历绑定键；`resumeVersionId` 不作为本 plan 正向 route/context 字段。
- Records placeholder 只说明当前规划下的模拟面试记录区域存在；真实记录行必须来自 typed records contract owner。

## 3 质量门禁

- **Plan 类型**: `feature-behavior + contract + frontend-ui + BDD`。
- **TDD 策略**: 适用。Vitest 覆盖 route hydration、InterviewContext reducer、WorkspaceScreen list DOM anchors、parse/report direct start handoff、generated client body/header、auth pendingAction、privacy and non-current negative gates。
- **BDD 策略**: 适用。`E2E.P0.018` - `E2E.P0.021` 覆盖 workspace 默认渲染、context loading、立即面试、embedded-only handoff、privacy 和 non-current negative gates。
- **替代验证 gate**:
  - `pnpm --filter @easyinterview/frontend test src/app/screens/workspace src/app/screens/parse/ParseResumeBinding.test.tsx src/app/screens/report/__tests__/ReplayCta.test.tsx src/app/App.test.tsx`
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

### Phase 3: Plan and resume switching (retired from workspace owner)

- Historical implementation added workspace-local plan/resume modals.
- v1.19 removes those runtime files from workspace owner; parse detail owns resume selection and workspace cards only open `parse`.
- Modal keyboard / focus behavior is no longer a workspace acceptance gate.

### Phase 4: Start practice and auth recovery (moved out of workspace)

- Historical implementation added workspace-local start-practice hooks.
- v1.19 moves the two-step launch to parse/report handoff owners through shared `startPracticeFromParams`.
- `workspace(autoStartPractice=1)` is no longer a valid side-effect contract; URL codec strips all workspace context params.

### Phase 5: Embedded insight, records placeholder and privacy (retired from workspace owner)

- Historical implementation added workspace-local insight and records placeholders.
- v1.19 removes those runtime files from current workspace owner; report/records owners must provide typed consumers before reintroducing user-visible rows.
- Negative coverage remains for sensitive fields and non-current runtime routes/testids.

### Phase 6: Verification closeout

- Frontend focused tests, workspace pixel parity, BDD scenario scripts, fixture validation, docs/index checks and negative grep gates passed in owner closeout.

### Phase 7: Interview nav and plan-list landing revision

- Update product/UI truth sources and static prototype so TopBar uses `面试` / `Interview` and `workspace` separates plan-list landing from current-plan detail.
- Add `WorkspacePlanList` backed by generated `listTargetJobs`; v1.19 changes plan-card navigation to `parse?targetJobId=...` without fabricating resume or report data.
- Detail, not-found and missing-resume states are parse owner responsibilities, not workspace list states.
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
- Keep `WorkspacePlanList` as the `面试` landing; v1.19 supersedes this phase by routing ordinary detail traffic to `parse` instead of `workspace?targetJobId=...`.
- Start-practice side effects are now owned by parse/report handoff code, not workspace `autoStartPractice`.
- Add regression coverage that the old independent workspace detail anchors (`workspace-header`, `workspace-launcher`, `workspace-jd-card`, `workspace-prep-card`, `workspace-history-card`) do not appear on ordinary detail re-entry, while `workspace-plan-list` and start-practice gates remain live.

### Phase 13: Plan-list admission and stale-context navigation remediation

- Reopen the completed owner because runtime evidence showed parse-failed / blank TargetJob rows could appear as interview-plan cards, and TopBar `面试` from a detail page reused stale context.
- Require `useWorkspaceTargetJobs` consumers to request `analysisStatus=ready` and require `WorkspacePlanList` to ignore non-ready or blank-title records if legacy data still appears.
- Align route split with `ui-design/src/screen-workspace.jsx`: workspace ignores stale `InterviewContext` and legacy route params, so it cannot produce “缺少目标岗位 ID”.
- Add regression coverage for failed/blank record exclusion, ready-query assertion, TopBar no-context list landing after a hydrated detail context, and E2E.P0.018 scenario wording.

### Phase 14: Workspace route purity remediation

- Reopen the completed owner because `workspace` is a REST list page, but historical detail/start hooks left it acting as a parameterized context route.
- Remove workspace detail/start/modal runtime files and tests from the current owner; keep only `WorkspacePlanList` + `useWorkspaceTargetJobs`.
- Change plan-card navigation and static UI truth source to open `parse`, not `workspace?targetJobId=...`.
- Remove `workspace` from `INTERVIEW_CONTEXT_ROUTES`; App route sync clears context whenever route name is `workspace`.
- Move start-practice side effects to parse/report handoff owners through generated `getPracticePlan` / `createPracticePlan` / `startPracticeSession`, with `Idempotency-Key` on side effects and no `autoStartPractice` workspace hop.

### Phase 15: Plan-list card size stability

- Reopen the completed owner because screenshot review showed the desktop list grid stretched a single plan card across the full content width.
- Update `docs/ui-design/module-job-workspace.md` and `ui-design/src/screen-workspace.jsx` so the plan-list grid uses `auto-fill` with a fixed maximum column width instead of `auto-fit + 1fr`.
- Update formal `WorkspacePlanList` and focused tests so the grid contract rejects `1fr` desktop stretching while keeping compact mobile single-column behavior.
- Verify with focused Vitest, source/UI design contract tests, typecheck, build, and browser screenshots.

### Phase 16: Home recent card / workspace list card fusion

- Reopen the completed owner because user review asked the Home "最近模拟面试" card and Interview list card to become one visual object instead of two separate card systems.
- Update `docs/ui-design/module-job-workspace.md` and `ui-design/src/screen-workspace.jsx` so workspace plan cards use the Home recent card body, including mini round rail driven by `TargetJob.summary.interviewRounds[]`.
- Update formal `WorkspacePlanList` to reuse the Home recent card component/body while appending the workspace-specific footer CTA; the card grid keeps the fixed `360px` desktop max width from Phase 15.
- Add focused regression coverage that fails when workspace cards lose the home recent mini rail or reintroduce a separate workspace-only body.

### Phase 17: Plan-list action row and card-click planning

- Reopen the completed owner because user review asks the visible `进入规划` footer CTA to become invisible and be replaced by clicking the card itself.
- Update `docs/ui-design/module-job-workspace.md` and `ui-design/src/screen-workspace.jsx` so workspace cards append `立即面试` and a top-right trash icon delete action; the card root remains the planning-detail navigation control.
- Update formal `WorkspacePlanList` to start practice directly through shared generated practice handoff with structured `roundId/roundName` when `立即面试` is clicked, keep delete isolated from card navigation, and let Phase 18 own backend-persistent archive behavior.
- Add focused regression coverage that fails when `进入规划` appears as a visible footer button, when Home recent shows a delete action, when delete triggers navigation/backend deletion, or when `立即面试` opens the planning detail instead of starting practice.

### Phase 18: Persistent TargetJob archive integration

- Reopen the completed owner because backend-targetjob now owns `archiveTargetJob`, so workspace delete must no longer be local-only.
- Update `WorkspacePlanList` to call generated `archiveTargetJob(targetJobId)` with an `Idempotency-Key`; only remove the card after the backend returns success.
- Preserve card-click planning and quick-start propagation boundaries: delete and quick-start must stop bubbling to card navigation.
- Add focused regression coverage that fails when delete is implemented via local-only hidden state, when `archiveTargetJob` is missing from generated client usage, when delete success does not remove the card, when delete failure hides the card, or when Home recent renders a delete control.
- Add real-backend browser smoke and screenshot proof that an archived TargetJob disappears from workspace after refresh.

### Phase 9: Plan-list card simplification and theme consistency

- Reopen the completed owner after screenshot review because Phase 8 still rendered low-value source/language metadata and a secondary CTA that visually competed with the theme.
- Update `docs/ui-design/module-job-workspace.md` and `ui-design/src/screen-workspace.jsx` so no-context plan cards are concise: status + updated date, title, company/location, and a theme accent `进入规划` / `Open plan` CTA only.
- Update formal `WorkspacePlanList` to remove source/language display, keep footer as a minimal action row, and strengthen separation from page background via existing card/rule/elevation tokens.
- Add jsdom, Playwright and E2E.P0.018 assertions that fail when `workspace.planList.cardMeta`, source labels, target language text, or secondary button styling return to the card.

## 5 验收标准

| ID | 验收点 | 验证 |
|----|--------|------|
| A-1 | `workspace` route renders the pure plan-list page and preserves UI parity | WorkspaceScreen tests, pixel parity, `E2E.P0.018` |
| A-2 | InterviewContext route hydration and carry/clear behavior are deterministic | `InterviewContext.test.tsx`, `App.test.tsx` |
| A-3 | Workspace owner no longer contains Plan Switcher / Resume Picker runtime; parse detail displays the saved resume binding readonly | source negative gate + `ParseResumeBinding.test.tsx` |
| A-4 | Start practice runs from parse/report owner through `getPracticePlan` / `createPracticePlan` / `startPracticeSession` with idempotency | `ParseResumeBinding.test.tsx`, `ReplayCta.test.tsx` |
| A-5 | Workspace owner no longer contains embedded insight / records placeholder runtime | source negative gate |
| A-6 | Privacy and non-current route/module gates have zero runtime residuals | scenario verify scripts, pruning-surface lint |
| A-7 | TopBar shows `面试` / `Interview`; no-context `workspace` shows a plan list landing; plan cards open current-plan detail | TopBar tests, WorkspaceScreen tests, `E2E.P0.018`, pixel parity workspace spec |
| A-8 | Workspace plan-list cards keep stable desktop width regardless of 1/2/3 card count | `WorkspaceScreen.test.tsx`, browser screenshot |
| A-8 | Plan-list landing visually renders as Home recent-style cards with mini round rail, not loose text columns or a separate workspace-only body | `WorkspaceEmptyState.test.tsx`, `WorkspaceScreen.test.tsx`, `frontend/tests/pixel-parity/workspace.spec.ts`, `E2E.P0.018` |
| A-9 | Plan-list cards stay concise and theme-consistent: no source/language metadata, accent quick-start CTA, clear card/page separation | `WorkspaceEmptyState.test.tsx`, `frontend/tests/pixel-parity/workspace.spec.ts`, `E2E.P0.018` |
| A-10 | Target job import persists selected resume binding and workspace plan-list re-entry carries `resumeId` even before any `practice_plans` row exists | backend targetjob tests, Home import tests, Workspace plan-list regression, local API smoke |
| A-11 | Workspace list card re-entry navigates to `parse` detail and no longer shows the independent workspace detail page or starts sessions | `WorkspaceScreen.test.tsx`, `WorkspaceEmptyState.test.tsx`, `frontend/tests/pixel-parity/workspace.spec.ts`, `E2E.P0.018` |
| A-12 | Workspace plan list requests ready TargetJobs only, filters failed / blank-title records defensively, and TopBar / legacy-param workspace navigation clears stale detail context | `WorkspaceEmptyState.test.tsx`, `WorkspaceScreen.test.tsx`, `App.test.tsx`, `E2E.P0.018` |
| A-13 | Parse/report handoff owners start practice directly and do not route through `workspace(autoStartPractice=1)` | `ParseResumeBinding.test.tsx`, `ReplayCta.test.tsx` |
| A-14 | Workspace card click opens planning detail while footer provides quick start carrying structured `roundId/roundName`, and top-right delete performs persistent `archiveTargetJob`; Home recent reuses quick start and omits delete | `MockInterviewCard.test.tsx`, `HomeRecentMocks.test.tsx`, `WorkspaceScreen.test.tsx`, `WorkspaceEmptyState.test.tsx`, browser screenshots |
| A-15 | Workspace delete is durable across refresh and never implemented as local-only hiding | generated-client tests, real-backend smoke, `E2E.P0.018`, screenshot acceptance |

## 6 变更记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-09 | 1.25 | Review remediation: preserve structured round params in workspace list quick-start practice handoff. |
| 2026-07-09 | 1.24 | Move the workspace card delete icon to the card top-right; keep footer for `立即面试` only. |
| 2026-07-09 | 1.23 | Reopen owner plan to integrate generated `archiveTargetJob` for persistent workspace card delete. |
| 2026-07-09 | 1.22 | Reopen owner plan to replace visible Open plan footer CTA with card-click planning plus quick-start and delete actions. |
| 2026-07-09 | 1.21 | Reopen owner plan to fuse Home recent mock cards and workspace plan-list cards into one shared card body with a workspace footer CTA. |
| 2026-07-09 | 1.16 | Reopen owner plan for target job-level resume binding persistence so workspace plan-list re-entry no longer loses the resume selected during JD import. |
| 2026-07-09 | 1.17 | Reopen owner plan to route workspace current-plan detail into the unified Parse-derived Interview Plan Detail / Context Confirm mother page while preserving workspace start-practice ownership. |
| 2026-07-09 | 1.18 | Reopen owner plan for parse-failure dirty-data admission defense and stale InterviewContext-free TopBar workspace navigation. |
| 2026-07-09 | 1.19 | Reopen owner plan to make workspace a pure list route, delete old detail/start/modal runtime, route cards to parse, and move practice start side effects to parse/report owners. |
| 2026-07-08 | 1.14 | Reopen owner plan for plan-list card simplification, metadata removal and theme-consistent CTA styling after screenshot review. |
| 2026-07-08 | 1.13 | Reopen owner plan for interview plan-list card visual hardening after screenshot review. |
| 2026-07-08 | 1.12 | Reopen owner plan for Interview nav naming and workspace plan-list landing revision. |
| 2026-07-07 | 1.11 | Compress owner docs to the current workspace, flat Resume Picker, start-practice, embedded insight, records placeholder and privacy contract. |
| 2026-07-07 | 1.10 | Reconcile workspace owner handoff, completion, fixture and active-list boundary wording. |
