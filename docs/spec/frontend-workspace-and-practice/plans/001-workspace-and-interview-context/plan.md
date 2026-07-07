# 001 Workspace + InterviewContext + Start Practice Contract

> **版本**: 1.11
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

本 plan 已完成 `workspace` 当前面试规划页、`InterviewContext`、规划/简历切换、立即面试启动和未登录恢复合同：

- `workspace` route 渲染正式 `WorkspaceScreen`，复刻 `ui-design/src/screen-workspace.jsx` 的当前面试规划、Interview Launcher、公司信号嵌入卡片、JD 拆解、我的准备和当前规划记录区域。
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
| `listTargetJobs` | `openapi/fixtures/TargetJobs/listTargetJobs.json` | `PlanSwitcherModal` 候选规划列表 | `backend-targetjob/001` | `target_jobs` | none | `E2E.P0.018` |
| `getTargetJob` | `openapi/fixtures/TargetJobs/getTargetJob.json` | Workspace header、Launcher、JD 拆解、准备信号、空态 | `backend-targetjob/001` | `target_jobs` + requirements/sources | none | `E2E.P0.018`, `E2E.P0.019` |
| `getResume` | `openapi/fixtures/Resumes/getResume.json` | 绑定简历摘要与缺简历空态 | `backend-resume/001` | `resume_assets` | none | `E2E.P0.018`, `E2E.P0.019` |
| `listResumes` | `openapi/fixtures/Resumes/listResumes.json` | `ResumePickerModal` flat active-list | `backend-resume/001` | `resume_assets` | none | `E2E.P0.018` |
| `getPracticePlan` | `openapi/fixtures/PracticePlans/getPracticePlan.json` | workspace refresh / ready-plan reuse / not-found recovery | `backend-practice/001` | `practice_plans` | none | `E2E.P0.019`, `E2E.P0.020` |
| `createPracticePlan` | `openapi/fixtures/PracticePlans/createPracticePlan.json` | 无可用 plan 时创建 baseline plan | `backend-practice/001` | `practice_plans` | backend-only first question prep | `E2E.P0.020` |
| `startPracticeSession` | `openapi/fixtures/PracticeSessions/startPracticeSession.json` | 启动 practice session 并 handoff 到 `practice` | `backend-practice/001` | `practice_sessions` + first turn | backend-only `practice.session.first_question` | `E2E.P0.020` |
| `getFeedbackReport` | N/A | 本 plan 不消费；records placeholder 不拼 report row | external owner | external | none | `E2E.P0.021` negative |

### 2.2 UI / Route Boundary

- `workspace` 保留 App chrome；`practice` 和 `generating` 由下游 owner 隐藏 chrome。
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

## 5 验收标准

| ID | 验收点 | 验证 |
|----|--------|------|
| A-1 | `workspace` route renders current planning page and preserves UI parity | WorkspaceScreen/Header/modal tests, pixel parity, `E2E.P0.018` |
| A-2 | InterviewContext route hydration and carry/clear behavior are deterministic | `InterviewContext.test.tsx`, `App.test.tsx` |
| A-3 | Plan Switcher consumes `listTargetJobs`; Resume Picker consumes flat `listResumes` active-list | modal tests, `E2E.P0.018` verify |
| A-4 | Start practice runs `getPracticePlan` / `createPracticePlan` / `startPracticeSession` with idempotency and auth recovery | `WorkspaceStartPractice.test.tsx`, `WorkspaceAuthGate.test.tsx`, `E2E.P0.020` |
| A-5 | Workspace insight stays embedded and records placeholder does not synthesize report rows | `WorkspaceHandoff.test.tsx`, `E2E.P0.021` |
| A-6 | Privacy and non-current route/module gates have zero runtime residuals | scenario verify scripts, pruning-surface lint |

## 6 变更记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-07 | 1.11 | Compress owner docs to the current workspace, flat Resume Picker, start-practice, embedded insight, records placeholder and privacy contract. |
| 2026-07-07 | 1.10 | Reconcile workspace owner handoff, completion, fixture and active-list boundary wording. |
