# 001 BDD Plan

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-08

**关联 Plan**: [plan](./plan.md)

## 1 场景矩阵

| 场景 ID | 类别 | 关联 Phase | 关联 Spec C-* | 关联 BDD-Gate（主 checklist） |
|---------|------|-----------|--------------|----------------------------|
| E2E.P0.018 | primary path · workspace 默认渲染 + Plan Switcher / Resume Picker | Phase 1 + 2 + 3 | C-2, C-7, C-8, C-9 | Phase 1.6、Phase 2.8、Phase 3.8 |
| E2E.P0.019 | primary + boundary + failure · context loading + getPracticePlan refresh + WorkspaceEmptyState / WorkspaceMissingResumeState | Phase 2 + 3 + 4 | C-2, C-3, C-8, C-9 | Phase 3.8、Phase 4.9 |
| E2E.P0.020 | primary + alternate · 立即面试 双步契约 + Idempotency-Key + pendingAction(start_practice) + 未登录恢复 | Phase 4 | C-1, C-3, C-12 | Phase 4.9 |
| E2E.P0.021 | regression / legacy-negative · session history → report handoff + company intel handoff + 隐私红线 + 旧 route/testid 反向 grep | Phase 5 + 6 | C-7, C-9, C-10, C-12 | Phase 5.5、Phase 6.9 |

---

## Phase 1 + 2 + 3: workspace 默认渲染 + Plan Switcher / Resume Picker

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.018 | Workspace 默认渲染（happy path）+ Plan Switcher / Resume Picker | 用户已登录；`getTargetJob` fixture 配置 `with-rounds` variant；`getResume` fixture 返回 `default`；`listTargetJobs` fixture 配置 `default`（多 plan）+ `single-plan` 两 variant；`InterviewContext` 通过 route param 传入 `targetJobId / jdId / planId / resumeVersionId / roundId` | 用户从 home `Recent mock interviews` 卡片点击进入 workspace；分别打开 `Plan Switcher Modal` 与 `Resume Picker Modal`；切换 plan 后再次回到 workspace | （1）route 渲染 `workspace` 命中 `WorkspaceScreen` 而非 PlaceholderScreen；TopBar `topbar-nav-workspace` 高亮；（2）testid `workspace-crumbs` / `workspace-plan-eyebrow-{label,title,status,sub}` / `workspace-plan-action-{switch,create}` / `workspace-header-{tag,level,updated,title,subtitle,prep}` / `workspace-launcher-*` / `workspace-round-rail-*` / `workspace-binding-{jd,resume}` / `workspace-cta-start` / `workspace-companyintel-{summary,open}` / `workspace-jd-block-{must,nice,hidden}` / `workspace-prep-{strong,risk}-${idx}` 全部命中；（3）Plan Switcher Modal 通过 `listTargetJobs` 拉取候选 plan 列表（`default` variant 多 plan + `single-plan` variant 单 plan），DOM 渲染对应卡片；选择不同 plan 后 Modal 关闭，`InterviewContext` 切换并触发 `getTargetJob` / `getResume` / `getPracticePlan` 重新拉取；`从新 JD 创建规划` CTA 调 `nav("home")`；（4）Resume Picker Modal 渲染当前绑定 resume 选中态 + 其余 disabled 占位卡 + i18n 文案 `resumePicker.disabledNote`；generated client `listResumes` 调用次数为 0；（5）两 Modal 均支持 ESC 关闭 / 外层遮罩点击关闭 / X 按钮 / focus trap；关闭后 focus 回到触发按钮；（6）切换 zh/en、warm/light → dark → customAccent，关键文本 / computed background 出现可见变化；（7）mobile (390×844) viewport 下 Main 双列折叠为单列、Round Rail 横向滚动、Modal 全屏化、CTA 与 BindingPill 不溢出 | `test/scenarios/e2e/p0-018-workspace-default-render/` |

## Phase 2 + 3 + 4: Context loading + 空态 + getPracticePlan refresh

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.019 | Workspace context loading 三种 variant + getPracticePlan refresh | 用户已登录；分四子场景：（A）`getTargetJob=with-rounds` + `getResume=default` + `getPracticePlan=default(ready)`；（B）`getTargetJob=not-found`；（C）`getTargetJob=with-rounds` + `getResume=not-found`；（D）`getTargetJob=with-rounds` + `getResume=default` + `getPracticePlan=archived` 与 `not-found` 两 variant | 用户分别加载 workspace 路由（携带 `InterviewContext` route params） | （A 主路径）`WorkspaceScreen` 完整渲染、所有 testid 命中、`InterviewContext` 通过 `MERGE_TARGET_JOB / MERGE_RESUME / MERGE_PRACTICE_PLAN` 完整 hydrate；（B 缺 JD）渲染 `WorkspaceEmptyState`，testid `workspace-empty-{eyebrow,title,desc,cta}` 命中，CTA `导入 JD` 调 `nav("home")` 并 focus `home-jd-textarea`；（C 缺简历）渲染 `WorkspaceMissingResumeState`，testid `workspace-missing-resume-{eyebrow,title,desc,cta}` 命中，CTA `创建简历` 调 `nav("resume_versions", { flow: "create" })`；（D plan refresh）`archived` variant → `InterviewContext.planId=null`、UI 状态为 ready（不展示错误，因为 createPracticePlan 在用户点立即面试时执行）；`not-found` variant → 同上行为；（E 通用）`getTargetJob` 5xx → header 退化为只读旧 context + `重试` 按钮，不渲染假数据；JD 原文 / 简历正文 / `questionText` 不出现在 console.log / URL / localStorage / telemetry | `test/scenarios/e2e/p0-019-workspace-context-loading/` |

## Phase 4: 立即面试双步契约 + auth pendingAction

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.020 | 立即面试主路径 + 错误重试 + 未登录恢复 | 分三子场景：（A）已登录 + `InterviewContext.planId` 不存在（首次面试）；（B）已登录 + `InterviewContext.planId` 存在 + `getPracticePlan=default(ready)`；（C）未登录 + `InterviewContext` 完整；fixture：`createPracticePlan=default` + `missing-resume` + `validation-422`，`startPracticeSession=default` + `ai-timeout-502` | 用户点击 `立即面试`（`workspace-cta-start`）；分别触发：（A1）首次成功；（A2）`createPracticePlan` 422 (`missing-resume`)；（A3）`startPracticeSession` 502 (`AI_PROVIDER_TIMEOUT`) + 重试；（B1）跳过 createPracticePlan 直接 startPracticeSession；（C1）未登录 → `requestAuth` → 登录恢复 | （A1）（1）调 `createPracticePlan` body 含 `targetJobId / goal='baseline' / mode='assisted' / interviewerPersona='hiring_manager' / difficulty='standard' / language=zh-CN / questionBudget=6 / timeBudgetMinutes=30 / resumeAssetId / focusCompetencyCodes=[]`；（2）再调 `startPracticeSession` body `{ planId, hintsEnabled: false }`（practiceMode='strict' → hintsEnabled=false）；（3）两次调用 `Idempotency-Key` header 来自同一 batch（`create` / `start` 双键稳定）；（4）成功后 `nav("practice", { sessionId, planId, targetJobId, jdId, resumeVersionId, roundId, roundName, mode='text', modality='text', practiceMode='strict', practiceGoal='baseline', hintUsed='false', hintCount='0' })`；practice route 仍渲染 PlaceholderScreen（plan 002 替换前）；（5）`questionText` 不在 workspace 屏 DOM 中渲染；（A2）（1）`createPracticePlan` 422 触发 inline 错误 + 滚动并 focus `更换简历` 按钮；（2）不进入 `startPracticeSession`；（3）保留输入；（A3）（1）`startPracticeSession` 502 显示翻译错误码 + 重试按钮；（2）点击重试，`Idempotency-Key` header 与首次一致（dedupe）；（3）3 次失败后展示 `回到首页` fallback CTA；（B1）（1）跳过 `createPracticePlan`，仅调 `startPracticeSession`；（C1）（1）`requestAuth({ type: "start_practice", route: "workspace", params: { ..., autoStartPractice: "1" } })` 触发；（2）`auth_login` 路由携带 `pendingRoute=workspace` + `pendingType=start_practice` + 完整 InterviewContext keys + PracticeDisplayContext keys；（3）verify 完成后 workspace 检测并清理 `autoStartPractice=1` → 自动调 `useStartPractice().start()` → 跳 `practice`；（4）`pendingAction.params` 不携带 `answerText / hintText / promptHash` 等敏感字段；（5）StrictMode 下 generated client `createPracticePlan + startPracticeSession` 调用次数 ≤ 2（无双触发）；（6）负向断言 workspace 不产出 `debrief_replay` | `test/scenarios/e2e/p0-020-workspace-start-practice/` |

## Phase 5 + 6: Handoff + 隐私红线 + 旧入口反向 grep

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.021 | Workspace handoff + 隐私红线 + 旧入口负向 | 用户已登录；`getTargetJob=with-rounds` + `getResume=default` + `getPracticePlan=default(ready)`；当前 typed contract 不提供 session history | 用户点击：（A）`workspace-companyintel-open` 公司情报入口；（B）`workspace-history-empty` / disabled history placeholder | （A）（1）调 `nav("company_intel", { targetJobId, jdId })`；（2）`company_intel` route 仍渲染 PlaceholderScreen（外部 owner 替换）；（3）generated client `getCompanyIntel` 调用次数为 0；（B）（1）history 区域渲染 `EmptyHistory` / disabled placeholder；（2）点击不触发 `nav("report", ...)`；（3）不读取未声明 `TargetJob.recentSessions` / fixture extension，不调用 `getFeedbackReport`；（D 隐私）（1）JD 原文 / 简历正文 / `questionText` / AI prompt-response / `answerText` / `hintText` 不出现在 console / URL / localStorage / telemetry / fixture transport 日志 0 命中；（2）`pendingAction.params` 仅含 IDs / route / `PracticeDisplayContext` / `autoStartPractice`；（E 旧入口负向）（1）旧 prototype workspace 业务 testid（`practice-mode-card-*` / `growth-*` / `drill-builder-*` / `mistake-queue-*` / `workspace-mocked-*`）grep 0 命中；（2）旧 route alias（`welcome` / `growth` / `mistakes` / `drill` / `followup` / `experiences` / `star` / 独立 `voice`）在 workspace 模块 grep 0 命中（除 `normalizeRoute` alias map）；（3）`frontend/src/app/screens/workspace/` + `frontend/src/app/interview-context/` 不 import `ui-design/src/data.jsx` / `window.EI_DATA` / `getWorkspace*` prototype helper；（4）generated client `listResumes` / `getCompanyIntel` 调用次数为 0；（F regression）D1+D2+D3 已存在 `E2E.P0.001 / 002 / 004 / 005 / 006` 全部 PASS；home plan `E2E.P0.014 / 015 / 016 / 017` 仅在场景资产存在且 INDEX Ready 时执行 | `test/scenarios/e2e/p0-021-workspace-handoff/` |
