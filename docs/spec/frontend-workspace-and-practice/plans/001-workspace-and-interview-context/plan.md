# 001 Workspace + InterviewContext + Start Practice Contract

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-05-09

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

在 `frontend-shell` D1+D2+D3 已交付的 App 壳、视觉系统、pixel parity gate、`requestAuth(pendingAction)` 与 fixture-backed generated client，以及 `frontend-home-job-picks-and-parse/001` active plan 已定义的 `parse confirm → workspace` 跳转契约之上，把 `workspace` 屏从 D1 `PlaceholderScreen` 迁到正式前端，端到端跑通 P0 主路径「确认 JD/岗位/绑定简历/轮次 → 立即面试 → backend 创建 plan + session → 跳到 practice 路由」；同时落地跨路由共享的 `InterviewContext` store/hook，作为 `practice / generating` 后续 plan 的统一上下文真理源。

完成本计划后，用户在 frontend dev server 上能够：

1. 通过 TopBar `模拟面试` 入口或 home/parse 跳转进入 `workspace`，看到当前面试规划页头（公司·岗位 / 状态 / `当前轮次·绑定简历` / `切换规划` / `新建规划`）、Interview Launcher（轮次节点条 + 面试前确认 + `立即面试` CTA + 目标岗位/JD + 绑定简历）、Main Left（公司轻情报 handoff 卡片 + JD 拆解）、Main Right（我的准备 + 当前规划的模拟面试历史占位）
2. 通过 `切换规划` 打开 Plan Switcher Modal（消费 `listTargetJobs`），选择规划后更新 `InterviewContext`，关闭 modal
3. 通过 `更换简历` 打开 Resume Picker Modal（受 `listResumes` 缺契约约束，本 plan 仅展示「当前绑定简历 + disabled 列表」模式，附 spec §3.2 待确认事项的可见说明）
4. 点击 `立即面试`：当 `InterviewContext.planId` 不存在或 `getPracticePlan` 返回 404 时先调 `createPracticePlan(goal='baseline')`（带 `Idempotency-Key`），再调 `startPracticeSession`（带 `Idempotency-Key`），成功后 `nav("practice", { sessionId, planId, targetJobId, jdId, resumeVersionId, roundId, mode, modality, practiceMode, hintUsed, hintCount })`；practice route 仍渲染 D1 `PlaceholderScreen`，由 plan 002 替换
5. 未登录用户点 `立即面试` 触发 `requestAuth({ type: "start_practice", route: "workspace", params: { ...InterviewContext, ...PracticeDisplayContext, autoStartPractice: "1" } })`，登录成功后回到 workspace，自动恢复 startPractice 双步流程并跳到 practice
6. 缺 JD/target 渲染 `WorkspaceEmptyState`（CTA 跳 home）；缺简历渲染 `WorkspaceMissingResumeState`（CTA 跳 `resume_versions?flow=create`）
7. 公司轻情报摘要卡片只渲染 `CompanyIntelEmbed`（源级复刻 `screen-company-intel.jsx::CompanyIntelEmbed`），点击 `打开公司情报` 调 `nav("company_intel", { targetJobId, jdId })` handoff 给外部 owner；不在本 plan 实现 `CompanyIntelScreen`
8. 模拟面试历史因当前 OpenAPI/generated `TargetJob` 未声明 typed history 字段，本 plan 只渲染 `EmptyHistory` / disabled placeholder，不读取 fixture extension；真实历史行与 `nav("report", { sessionId, reportId })` handoff 交给后续 `listPracticeSessions` 或等价 typed contract owner
9. 全部用户行为通过 generated client + fixture-backed mock transport 闭环；JD 原文 / 简历正文 / hint / answer / AI prompt-response 不出现在 console / URL / localStorage / telemetry；i18n zh/en 完整切换；dark + customAccent 三态可见变化；desktop (1440×900) + mobile (390×844) pixel parity 通过

## 2 背景

`frontend-shell` D1+D2+D3 已交付：默认 `home` route + 五入口 TopBar + route normalization + `requestAuth(pendingAction)` 与登录恢复 + generated client + fixture transport bootstrap + warm/forest/ocean/plum 四主题 + dark + customAccent + Vitest+jsdom smoke gate（`E2E.P0.001/002/004/005`）+ Playwright pixel parity gate（`E2E.P0.006`）。

`frontend-home-job-picks-and-parse/001-home-jd-import-and-parse` 当前仍为 active plan（参见 [home plan §1 / §3.7](../../../frontend-home-job-picks-and-parse/plans/001-home-jd-import-and-parse/plan.md)）：它定义 home / parse / jd_match 屏、`interviewContextFromTargetJob(targetJob)` 与 parse confirm → workspace 跳转契约，但 `test/scenarios/e2e/p0-014` ~ `p0-017` 场景资产尚未出现在当前仓库。本 plan 必须接住同款 `nav("workspace", { targetJobId, jdId, planId, resumeVersionId, roundId, ... })` params；在 home plan Ready 前，P0.018 使用直接 workspace route/hash seed 验证，不把 P0.014-017 作为无条件完成 gate。

`backend-practice` v1.3 spec 已锁 6 个 Practice operation 与 D-13 `startPracticeSession` 同步首题语义，并记录当前 OpenAPI/generated `PracticeMode` 仍有旧 `debrief_replay` enum 漂移；`backend-targetjob` 实施 plan `001` 已交付 `listTargetJobs / getTargetJob / updateTargetJob / importTargetJob` 真实 handler。本 plan 只通过 generated client 消费已存在的 OpenAPI operation；`createPracticePlan / getPracticePlan / startPracticeSession` 真实 backend handler 由 `backend-practice/001-plan-and-session-orchestration`（待派生）承接，本 plan 阶段保持 fixture-backed `not-yet-implemented` 状态。

`workspace / practice / generating` 三屏在 D1 仍由 `PlaceholderScreen` 占位；本 plan 是 `frontend-workspace-and-practice` 子 spec 的首个计划，承接 spec §7 预留编号 `001-workspace-and-interview-context`。`practice` 与 `generating` 屏继续保持 PlaceholderScreen，由 plan `002` / `003` / `004` 在 backend-practice handler / voice / report 契约就位后替换。

## 2.1 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-05-09 | 1.2 | L2 follow-up 修复 workspace generated API server-bound id 归一化、target-job stale/error recovery、target 切换竞态与英文本地化派生标签。 |

## 3 质量门禁分类

- **Plan 类型**: feature-behavior（用户可感知 UI + API 行为 + 业务流程 + 端到端功能）
- **TDD 策略**: Red-Green-Refactor 入口为 `pnpm --filter @easyinterview/frontend test`（Vitest）；每个 Phase 在新增组件 / hook / 路由壳前先写失败测试，覆盖 DOM 锚点、控件类型、props/state、generated client 调用断言（method、path、body schema、`Idempotency-Key` header）、URL/state 隐私反查与负向旧 testid 断言；`pnpm --filter @easyinterview/frontend test:pixel-parity` 在 Phase 6 扩展为 workspace + workspace empty/missing-resume 状态的 desktop + mobile spec；新增组件位于 `frontend/src/app/screens/workspace/`、`frontend/src/app/interview-context/`、`frontend/src/app/screens/workspace/modals/`；测试文件与组件 colocate（`*.test.tsx`）。
- **BDD 策略**: Feature plan requires BDD；本 plan 在 [bdd-plan.md](./bdd-plan.md) 定义 4 个场景 `E2E.P0.018 / E2E.P0.019 / E2E.P0.020 / E2E.P0.021`，[bdd-checklist.md](./bdd-checklist.md) 跟踪每个场景资产创建与执行；主 [checklist.md](./checklist.md) 在每个 Phase 末尾保留 `BDD-Gate:` 项引用对应场景 ID。
- **替代验证 gate**: 不适用（feature plan，已有完整 BDD + TDD 双层覆盖）

## 3.5 Coverage Matrix

| 类别 | 覆盖描述 | UI Source Anchor | Phase | 验证入口 |
|------|----------|------------------|-------|---------|
| Primary path · workspace 默认渲染 | 进入 workspace，渲染 plan eyebrow + Interview Launcher + Main Left + Main Right；TopBar 高亮 `workspace` | `screen-workspace.jsx::WorkspaceScreen` lines 116-302 + `app.jsx::App.screens.workspace` line 271 | 1+2 | E2E.P0.018 + Vitest `workspace/WorkspaceScreen.test.tsx` |
| Primary path · 立即面试 双步契约 | 立即面试：plan 不存在时 `createPracticePlan(goal='baseline')` → `startPracticeSession`；两步均带 `Idempotency-Key`；成功后 nav `practice` | `screen-workspace.jsx::WorkspaceScreen.startInterview` lines 102-114 | 4 | E2E.P0.020 + Vitest `workspace/WorkspaceStartPractice.test.tsx` |
| Alternate path · Plan Switcher | `切换规划` 打开 `PlanSwitcherModal`，选择不同 plan 后 `InterviewContext` 更新；新建规划 CTA 跳 home | `screen-workspace.jsx::PlanSwitcherModal` lines 587-666 | 3 | E2E.P0.018 + Vitest `workspace/modals/PlanSwitcherModal.test.tsx` |
| Alternate path · Resume Picker (disabled list) | `更换简历` 打开 `ResumePickerModal`；本 plan 仅渲染当前绑定简历 + disabled 列表 + 待解锁说明（spec §3.2 `listResumes` 缺契约） | `screen-workspace.jsx::ResumePickerModal` lines 517-585 | 3 | Vitest `workspace/modals/ResumePickerModal.test.tsx` + 负向断言：generated client `listResumes` 不被调用 |
| Alternate path · 未登录立即面试 | 未登录点 `立即面试` 触发 `requestAuth({ type: "start_practice" })`；登录后自动恢复 startPractice 双步契约 | `app.jsx::App.requestAuth` + `screen-workspace.jsx::WorkspaceScreen.startInterview` | 4 | E2E.P0.020 + Vitest `workspace/WorkspaceAuthGate.test.tsx` |
| Alternate path · 现有 plan 复用 | `InterviewContext.planId` 存在且 `getPracticePlan` 返回 `status=ready` → 跳过 `createPracticePlan` 直接 `startPracticeSession` | spec D-9 + `screen-workspace.jsx::startInterview` | 4 | Vitest 双 fixture variant + `Idempotency-Key` 反查 |
| Failure / recovery · createPracticePlan 4xx | `createPracticePlan` 422 (`missing-resume`) 显示 inline 错误，保留输入；不进入 `startPracticeSession` | n/a (error state) | 4 | Vitest fixture variant `missing-resume` + inline error UI |
| Failure / recovery · startPracticeSession AI 失败 | `startPracticeSession` 502 `AI_PROVIDER_TIMEOUT` 显示重试 UI；同 `Idempotency-Key` 重试调用相同 endpoint | spec D-9 + backend-practice C-5 | 4 | Vitest fixture variant + retry button reuses same key |
| Failure / recovery · getPracticePlan 404 | 当前 `InterviewContext.planId` 在 backend 找不到（DELETE /me cascade 等）→ 重新创建 plan | spec D-5 + backend-practice C-13 | 4 | Vitest fixture variant + flow falls back to `createPracticePlan` |
| Failure / recovery · listTargetJobs 失败 | 网络/5xx → 头部退化为只读旧 context（缓存的 InterviewContext params）+ `重试` 按钮，不渲染假数据 | n/a (error state) | 2 | Vitest |
| Boundary · 空 sessionHistory | 当前 plan 没有历史 session → 显示 `EmptyHistory` 占位 + `首场面试将出现在这里` 文案 | `screen-workspace.jsx::WorkspaceScreen` lines 242-264 (sessionHistory map) | 2 | Vitest |
| Boundary · listTargetJobs 仅 1 条 | 仅 1 个 target → Plan Switcher Modal 中只有 1 张 plan 卡 + `从新 JD 创建规划` CTA | `screen-workspace.jsx::PlanSwitcherModal` lines 608-647 | 3 | Vitest fixture variant |
| Boundary · 缺 JD context | `InterviewContext.targetJobId` 缺失 / `getTargetJob` 404 → `WorkspaceEmptyState` | `screen-workspace.jsx::WorkspaceEmptyState` lines 305-318 | 2+5 | E2E.P0.019 + Vitest `workspace/WorkspaceEmptyState.test.tsx` |
| Boundary · 缺简历 | `InterviewContext.resumeVersionId` 缺失 / `getResume` 404 → `WorkspaceMissingResumeState` | `screen-workspace.jsx::WorkspaceMissingResumeState` lines 320-333 | 2+5 | E2E.P0.019 + Vitest `workspace/WorkspaceMissingResumeState.test.tsx` |
| Boundary · plan eyebrow 文案 | `plan.round` / `selectedResume.name` ellipsis；`status` tone 切换 (`draft/preparing` muted, `applied/interviewing` amber, `offer/rejected/archived` neutral) | `screen-workspace.jsx::WorkspaceScreen` lines 122-149 + status pill mapping | 2 | Vitest computed style |
| Cross-layer contract · CreatePracticePlanRequest schema | body 字段 `targetJobId / goal='baseline' / mode='assisted'(默认) / interviewerPersona / difficulty / language / questionBudget / timeBudgetMinutes / resumeAssetId / focusCompetencyCodes` 与 OpenAPI `CreatePracticePlanRequest` 一致；side-effect 调用带 `Idempotency-Key` | OpenAPI `CreatePracticePlanRequest` schema + `openapi.yaml` line 586 | 4 | mock-contract-suite parity + Vitest body 反查 |
| Cross-layer contract · StartPracticeSessionRequest schema | body `planId / hintsEnabled` 与 OpenAPI `StartPracticeSessionRequest` 一致；side-effect 调用带 `Idempotency-Key`；hintsEnabled 由正式前端二值 `practiceMode` 派生（`assisted` → true，`strict` → false）；旧 generated enum 值 `debrief_replay` 只作为 negative drift 搜索，不得由 workspace 产出 | OpenAPI `StartPracticeSessionRequest` + spec D-3 / D-12 | 4 | Vitest body 反查 + spec D-3 衍生映射断言 |
| Cross-layer contract · `practice` 跳转 params | nav `practice` 携带 `sessionId / planId / targetJobId / jdId / resumeVersionId / roundId / roundName / mode / modality / practiceMode / practiceGoal / hintUsed / hintCount`；与 spec §2.1 `PracticeDisplayContext` + InterviewContext 一致 | spec §2.1 + `screen-workspace.jsx::WorkspaceScreen` line 94-101 | 4 | Vitest nav stub 断言完整字段集 |
| Cross-layer contract · listTargetJobs 复用 | `Plan Switcher Modal` 通过 generated `listTargetJobs` 拉数据；`pageSize` 复用 `frontend-home-job-picks-and-parse/001` 同款 viewmodel mapping | `screen-workspace.jsx::PlanSwitcherModal` + home plan §3.7 | 3 | Vitest |
| Cross-layer contract · getPracticePlan refresh | workspace mount 时若 `InterviewContext.planId` 存在 → 调 `getPracticePlan(planId)`；`status='ready'` 复用，`status='archived'` 或 404 视为缺 plan 走 createPracticePlan 路径；不得假设 OpenAPI 未声明的 plan status | spec §2.1 + OpenAPI `PracticePlan.status` | 4 | Vitest fixture 多 variant |
| Cross-layer contract · mode/practiceMode 协议 | spec D-3：`mode/modality∈{text,voice}` 与 `practiceMode∈{assisted,strict}` 独立；本 plan 立即面试默认 `mode='text', modality='text', practiceMode='strict'`（与 ui-design `screen-workspace.jsx` line 99 一致）；negative gate 确认 workspace 不产出 `debrief_replay` | `screen-workspace.jsx::startContext` lines 94-101 | 4 | Vitest startContext 字段断言 + negative grep |
| Cross-layer contract · CompanyIntelEmbed handoff | 卡片仅展示 fixture 中 `target_jobs` 公开摘要；`打开公司情报` 调 `nav("company_intel", { targetJobId, jdId })`；不调 `getCompanyIntel` | spec §2.2 + `screen-company-intel.jsx::CompanyIntelEmbed` | 5 | Vitest negative + nav stub 断言 |
| Cross-layer contract · session history placeholder | 因当前 typed contract 缺 `recentSessions` / `listPracticeSessions`，workspace 只渲染 `EmptyHistory` / disabled placeholder；不从 fixture extension 或 `any` 读取历史行；真实 report handoff 等 typed history contract 落地 | spec §2.1 + `screen-workspace.jsx::WorkspaceScreen.sessionHistory` visual shell | 5 | Vitest negative + E2E.P0.021 |
| Privacy / security · JD 原文 + 简历正文 | 原文 / parsed body 不进 console / URL / localStorage / telemetry / fixture transport 日志；header eyebrow 仅展示 `company / title / status / sourceType / fitSummary availability` 等 generated 结构化字段 | spec §4 隐私红线 | 2-5 | Vitest 反查 + redact lint |
| Privacy / security · pendingAction params | `pendingAction.params` 不携带 `answerText / hintText / promptHash` 等敏感字段；只允许 IDs / 状态 / route 与 `PracticeDisplayContext` 结构化字段，以及 `autoStartPractice=1` 控制位 | `frontend-shell/spec.md::pendingAction` + `pendingAction.ts` | 4 | Vitest |
| Privacy / security · Idempotency-Key 来源 | `Idempotency-Key` 通过 `frontend/src/lib/conventions/idempotency.ts` 派生；不复用 sessionId / userId 明文；POST 调用必带 | OpenAPI parameters `IdempotencyKey` | 4 | Vitest header 反查 |
| Observability | mockTransport spy 仅记录 status / latency / 4xx code，不带 body | n/a | 2-5 | Vitest mockTransport spy 断言 |
| UX · loading state | workspace mount 阶段 `listTargetJobs / getTargetJob / getResume / getPracticePlan` 并行加载占位（≥1 viewport 不闪烁）；立即面试按钮在请求中显示 spinner + disabled | n/a | 2+4 | Vitest fake timer |
| UX · empty state | 无 JD context / 无简历 / 无 sessionHistory 三种空态独立 | `WorkspaceEmptyState / WorkspaceMissingResumeState` + sessionHistory empty | 2+5 | Vitest |
| UX · error state | listTargetJobs 5xx / getTargetJob 5xx / createPracticePlan 4xx / startPracticeSession 5xx 各自显示 inline 错误 | n/a | 2+4 | Vitest |
| UX · i18n zh/en | 全文案通过 typed helper；新增 `workspace.*` namespace；切换立即重绘 | D1 typed locale helper | 1-5 | Vitest `i18n` namespace test |
| UX · dark + customAccent | workspace 三态切换：plan eyebrow / Interview Launcher / Main 双列 / CompanyIntelEmbed / Modal 关键元素 computed 颜色变化 | D2 `data-theme` / `data-mode` / `data-custom-accent` | 1+6 | Playwright + Vitest computed style |
| UX · responsive layout | mobile 390×844：Main 双列折叠为单列；Interview Launcher CTA 与 BindingPill 不溢出；Round Rail 横向滚动；Modal 全屏化 | n/a | 1+6 | Playwright mobile project |
| UI source structure parity · plan eyebrow | crumbs (`返回首页`) + plan card (`公司·岗位` / status tag / `当前轮次·绑定简历` / `切换规划` / `新建规划`) | `screen-workspace.jsx::WorkspaceScreen` lines 119-149 | 1 | Vitest DOM + testid `workspace-crumbs` / `workspace-plan-eyebrow-{label,title,status,sub}` / `workspace-plan-action-{switch,create}` |
| UI source structure parity · header summary | status tag + level + updatedAt + 标题 + 公司·地点·来源 + 准备状态/匹配度 | `screen-workspace.jsx::WorkspaceScreen` lines 152-172 | 2 | Vitest + testid `workspace-header-{tag,level,updated,title,subtitle,prep}` |
| UI source structure parity · Interview Launcher | Round Rail + 面试前确认文案 + `立即面试` 主 CTA + JD/Resume BindingPill + note line | `screen-workspace.jsx::WorkspaceScreen` lines 174-196 + `InterviewRoundRail` lines 677-722 + `BindingPill` lines 724-740 | 2+4 | Vitest + testid `workspace-launcher-*` / `workspace-round-rail-*` / `workspace-binding-{jd,resume}` / `workspace-cta-start` |
| UI source structure parity · Main Left CompanyIntelEmbed | `CompanyIntelEmbed` 卡片 + `打开公司情报` button → `nav("company_intel", ...)` | `screen-workspace.jsx::WorkspaceScreen` line 203 + `screen-company-intel.jsx::CompanyIntelEmbed` | 5 | Vitest + testid `workspace-companyintel-{summary,open}` |
| UI source structure parity · Main Left JD breakdown | Card + Must Have / Nice to Have / Hidden signals 三 ReqBlock；hits 命中圆点 | `screen-workspace.jsx::WorkspaceScreen` lines 206-219 + `ReqBlock` lines 742-759 | 2 | Vitest + testid `workspace-jd-block-{must,nice,hidden}` |
| UI source structure parity · Main Right risks/strengths | `我的准备` Card + 直接命中 / 风险提示 list | `screen-workspace.jsx::WorkspaceScreen` lines 224-239 | 2 | Vitest + testid `workspace-prep-{strong,risk}-${idx}` |
| UI source structure parity · Main Right session history | sessionHistory visual shell + `EmptyHistory` / disabled placeholder；不得 import `getWorkspaceSessionHistory` prototype helper | `screen-workspace.jsx::WorkspaceScreen` lines 241-265 | 5 | Vitest + testid `workspace-history-empty` + negative grep |
| UI source structure parity · ResumePickerModal | 模态层 + 简历卡列表 + footer Cancel / Use；本 plan 列表项除当前绑定外 disabled | `screen-workspace.jsx::ResumePickerModal` lines 517-585 | 3 | Vitest + testid `workspace-resume-modal-{card-${id},disabled-note,confirm,cancel,close}` |
| UI source structure parity · PlanSwitcherModal | 模态层 + plan 卡列表 + footer `从新 JD 创建规划` / Cancel / Use | `screen-workspace.jsx::PlanSwitcherModal` lines 587-666 | 3 | Vitest + testid `workspace-plan-modal-{card-${id},create,confirm,cancel,close}` |
| UI source structure parity · WorkspaceEmptyState / WorkspaceMissingResumeState | 单卡 + 标题 + 描述 + CTA | `screen-workspace.jsx::WorkspaceEmptyState` + `WorkspaceMissingResumeState` lines 305-333 | 2+5 | Vitest + testid `workspace-empty-{eyebrow,title,desc,cta}` / `workspace-missing-resume-*` |
| UI visual geometry parity · desktop | 1440×900 workspace + 两 modal + 两空态 bounding box stays in viewport, no overlap | n/a | 6 | Playwright `tests/pixel-parity/workspace.spec.ts` desktop project |
| UI visual geometry parity · mobile | 390×844 Main 单列、Round Rail 横向滚动、Modal 全屏化、CTA 不溢出 | n/a | 6 | Playwright mobile project |
| UI visual geometry parity · dark / customAccent | 三态切换关键元素 computed background / color 可见变化 | n/a | 6 | Playwright |
| UI visual geometry parity · screenshot regression | toHaveScreenshot baseline maxDiffPixels 阈值 | n/a | 6 | Playwright + frontend baseline |
| UI stale-contract negative · 旧 route alias | 旧 `welcome` / `growth` / `mistakes` / `drill` / `followup` / `experiences` / `star` / 独立 `voice` route 在 workspace 新代码中不出现（除 `normalizeRoute` alias map 与对应 D1 测试） | spec §4 + frontend-shell D1 alias 表 | 全 phase | Vitest + scenario verify negative grep |
| UI stale-contract negative · 旧画板标签 | `practiceModeCard` / `warmup` / `single_drill` / `drill_builder` / `mistake_queue` / `growth_center` 不出现在 workspace 模块和 i18n key | spec §4 + product-scope §4.5 | 全 phase | grep negative |
| UI stale-contract negative · 不直接 import prototype data | `frontend/src/app/screens/workspace/` 不 import `ui-design/src/data.jsx` / `window.EI_DATA` / `getWorkspaceJDSample` 等 prototype helper | n/a | 全 phase | Vitest + tsc grep |
| UI stale-contract negative · 不调 getCompanyIntel / listResumes | workspace 模块不调 `getCompanyIntel`（缺 contract） / `listResumes`（缺 contract）；调用尝试触发 typecheck 错误 | spec §3.2 + §5.1 operation matrix | 全 phase | Vitest spy + tsc |
| Regression / legacy-negative · D1+D2+D3 + 已存在场景 gate | `E2E.P0.001/002/004/005/006` 全部 PASS；`E2E.P0.014/015/016/017` 仅在对应目录已存在且 home plan INDEX 标记 Ready 后作为条件回归 gate 执行 | n/a | 6 | scenario rerun |
| Regression / legacy-negative · 不直接调用 LLM/provider | workspace 模块不出现 AI provider key / provider registry / prompt registry / AIClient / LLM endpoint / bypass generated client 的 ad hoc fetch | n/a | 全 phase | Vitest + grep negative |

### 高风险类别 N/A 说明

- **隐私 / 安全 · audio**：本 plan 不涉及语音 surface / STT / TTS / barge-in，audio buffer 不进入 workspace；audio 红线由 plan 003 落地。N/A 原因记录在此。
- **隐私 / 安全 · LLM prompt-response**：workspace 不直接调 LLM；首题生成由 backend `startPracticeSession` 同步处理（spec D-13）。前端只接收 `currentTurn.questionText`，本 plan 不在 workspace 屏渲染或缓存 questionText（属于 practice 屏 plan 002）。N/A 原因记录在此。

## 3.6 Frontend / Backend Operation Matrix

本 plan 走 `docs/development.md` §2.2 Frontend-First Path：正式前端先对齐 `ui-design/` 并通过 generated client + fixture-backed transport 完成 P0 UI/BDD；真实 handler、store、AI 调用由 `backend-practice/001-plan-and-session-orchestration`（待派生）独立 owner 落地前，以下 Practice 类 backend 状态保持 `not-yet-implemented`，不得把 fixture PASS 宣称为真实 backend 闭环。

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `listTargetJobs` | `openapi/fixtures/TargetJobs/listTargetJobs.json` scenarios: `default` / `prototype-baseline` / 计划新增 `single-plan` / `12-plus` | `PlanSwitcherModal` 通过 generated client 拉所有候选 plan；与 home plan §3.7 viewmodel mapping 一致 | implemented (`backend-targetjob/001` 已交付) | `target_jobs` + `target_job_requirements` | none in frontend | E2E.P0.018 |
| `getTargetJob` | `openapi/fixtures/TargetJobs/getTargetJob.json` scenarios: `default` / `prototype-baseline` / 计划新增 `with-rounds` / `not-found` | workspace mount 时通过 generated client 拉当前 `targetJobId`；驱动 header / Interview Launcher / JD 拆解 / risks-strengths；缺失或 404 → `WorkspaceEmptyState` | implemented (`backend-targetjob/001` 已交付) | `target_jobs` + `target_job_requirements` + `target_job_sources` | none in frontend | E2E.P0.018 / E2E.P0.019 |
| `getResume` | `openapi/fixtures/Resumes/getResume.json` scenarios: `default` / 计划新增 `not-found` | workspace mount 时通过 generated client 拉绑定 resume；驱动 BindingPill resume 段；缺失或 404 → `WorkspaceMissingResumeState` | `not-yet-implemented`（future `backend-resume`） | `resume_assets` | none | E2E.P0.018 / E2E.P0.019 |
| `listResumes` | N/A（缺 OpenAPI operation） | **本 plan 不消费**：Resume Picker 仅展示当前绑定 resume + disabled 列表；`spec §3.2` 待解锁前不调用 | missing operation；`openapi-v1-contract` + future `backend-resume` 修订前不实现 | resume assets | none | 负向断言：generated client `listResumes` 不存在或调用为 0 |
| `createPracticePlan` | `openapi/fixtures/PracticePlans/createPracticePlan.json` scenarios: `default` / `missing-resume` / 计划新增 `validation-422` | workspace 立即面试：`InterviewContext.planId` 不存在或 `getPracticePlan` 失败时调用；body 含 `targetJobId / goal='baseline' / mode='assisted' / interviewerPersona / difficulty / language / questionBudget / timeBudgetMinutes / resumeAssetId / focusCompetencyCodes`；side-effect 调用带 `Idempotency-Key` | `not-yet-implemented`（owned by `backend-practice/001-plan-and-session-orchestration`） | `practice_plans` | none in frontend；backend-only `practice.session.first_question` 由 `startPracticeSession` 阶段触发 | E2E.P0.020 |
| `getPracticePlan` | `openapi/fixtures/PracticePlans/getPracticePlan.json` scenarios: `default` / 计划新增 `archived` / `not-found` | workspace mount 时若 `InterviewContext.planId` 存在 → 拉取以确认 `status='ready'`；非 ready 视为缺 plan 走 createPracticePlan 路径 | `not-yet-implemented` | `practice_plans` | none | E2E.P0.019 |
| `startPracticeSession` | `openapi/fixtures/PracticeSessions/startPracticeSession.json` scenarios: `default` / 计划新增 `ai-timeout-502` | workspace 立即面试 plan 就绪后调用；body `{ planId, hintsEnabled }`；`Idempotency-Key` 与 createPracticePlan 同一 batch；成功响应携带 `currentTurn{turnIndex:1, questionText, askedAt, status:'asked'}`，前端只缓存 sessionId / planId / turnIndex 进 InterviewContext，**不在 workspace 屏渲染 questionText** | `not-yet-implemented` | `practice_sessions` + 第 1 个 turn + `session_started` event | backend-only `practice.session.first_question` (spec D-13) | E2E.P0.020 |
| `getCompanyIntel` | N/A | 本 plan **不消费**；CompanyIntelEmbed 卡片仅渲染 `target_jobs.companyName / location / source / summary`，handoff 由 `nav("company_intel", { targetJobId, jdId })` 触发外部 owner | external | external | none | 负向断言 + E2E.P0.021 |
| `getFeedbackReport` | N/A | 本 plan **不消费**；workspace session history 仅渲染 disabled placeholder，不通过 fixture extension 获取 `sessionId/reportId`；真实 history → report handoff 等 `listPracticeSessions` 或等价 typed contract 落地 | external (`frontend-report-dashboard` / future `backend-review`) | external | none | E2E.P0.021 negative |

## 3.7 InterviewContext View-Model Mapping

正式前端不得从 `ui-design/src/data.jsx` 或未声明 fixture 字段补齐 `InterviewContext` 之外的数据。`frontend/src/app/interview-context/InterviewContext.tsx` 锁定字段集合与 fallback：

| InterviewContext field | Source | Rule |
|-----------------------|--------|------|
| `planId` | `getPracticePlan(planId).id` 或 `createPracticePlan` 响应 `id` 或 fallback `plan-${targetJobId}` | route param 优先；fallback 仅在 plan 未创建时使用，不能进入 backend body |
| `targetJobId` / `jobId` | `getTargetJob.id` / route param | 必填；缺失 → `WorkspaceEmptyState` |
| `jdId` | route param 或 fallback `jd-${targetJobId}` | 用于 company intel handoff |
| `resumeVersionId` | route param 或 `getResume.id` | 缺失 → `WorkspaceMissingResumeState` |
| `roundId` | route param 或 `screen-workspace.jsx::getWorkspaceRoundId` 同款映射（`HR / 技术一面 / 技术二面 / 经理面 / round-draft`） | 不依赖 OpenAPI 未声明字段 |
| `roundName` | `getTargetJob` 中对应轮次或 locale fallback | i18n zh/en |
| `mode` / `modality` | startContext 默认 `text`/`text` | 跳 practice 时携带 |
| `practiceMode` | startContext 默认 `strict`（与 ui-design line 99 一致） | spec D-3 二值；正式前端只允许 `assisted` / `strict`，`debrief_replay` 作为旧 generated enum 负向搜索 |
| `practiceGoal` | route param 或 plan/session 数据 fallback `baseline` | 数据来源维度，不能塞进 `practiceMode` |
| `hintUsed` / `hintCount` | startContext 默认 `'false'` / `'0'` | spec §2.1 `PracticeDisplayContext` |
| `autoStartPractice` | pendingAction route param | 仅未登录恢复使用；值为 `'1'` 时 workspace mount 后清理该控制位并调用 `useStartPractice().start()` |
| `sessionId` | `startPracticeSession.id` | 启动后写入；workspace 不读取未声明的 history session 字段 |

`InterviewContext` 通过 React Context + 受控 reducer 管理，单元测试覆盖：rehydrate from route param / merge with fetched plan / default fallback / clear on `nav("home")` 回到无 context route。

## 4 实施步骤

### Phase 1: WorkspaceScreen 静态壳 + 路由壳 + InterviewContext store + i18n（无数据）

#### 1.1 新增 `frontend/src/app/interview-context/InterviewContext.tsx`

实现 `InterviewContextProvider` + `useInterviewContext()` + `useStartPracticeContext()`：

- `InterviewContextValue = { planId?, targetJobId, jobId, jdId, resumeVersionId?, roundId?, roundName?, mode, modality, practiceMode, hintUsed, hintCount, sessionId? }`
- reducer actions：`HYDRATE_FROM_ROUTE` / `MERGE_TARGET_JOB` / `MERGE_RESUME` / `MERGE_PRACTICE_PLAN` / `MERGE_SESSION` / `CLEAR`
- 默认值与 §3.7 mapping 一致
- 在 `frontend/src/app/routes.ts` 新增并导出 `INTERVIEW_CONTEXT_ROUTES` / `shouldCarryInterviewContext(routeName)`（集合与 `ui-design/src/app.jsx` 中 `INTERVIEW_CONTEXT_ROUTES` 保持 parity），再通过 `useNavigation` 协调跨路由参数传递；不得只引用当前不存在的常量

#### 1.2 新增 `frontend/src/app/screens/workspace/WorkspaceScreen.tsx`

按 `ui-design/src/screen-workspace.jsx::WorkspaceScreen` lines 116-302 源级复刻渲染 plan eyebrow、header summary、Interview Launcher（Round Rail + 面试前确认 + CTA + BindingPill 双卡 + note line）、Main Left（CompanyIntelEmbed placeholder + JD 拆解 placeholder）、Main Right（risks/strengths placeholder + sessionHistory placeholder）。本 phase 不接入数据：所有动态字段渲染 placeholder skeleton；`startInterview` callback 仅记录调用次数；`切换规划` / `更换简历` / `打开公司情报` / sessionHistory 行点击仅记录 nav stub。

#### 1.3 路由壳替换

在 `frontend/src/app/App.tsx` `renderRouteScreen` 中绑定 `workspace` → `<WorkspaceScreen route={route} />`（替换 D1 `PlaceholderScreen`）。`practice` / `generating` 仍渲染 `PlaceholderScreen`，不在本 plan 改动。

#### 1.4 i18n locale 文件扩展

在 `frontend/src/app/i18n/locales/zh.ts` / `en.ts` 中新增 `workspace.*` 命名空间（与 `screen-workspace.jsx::L` zh/en 字典等价 ≥ 50 key：crumbs / overview / requirements / prep / practices / timeline / startCore / launchTitle / launchSub / flow / roundStatus / jdBound / resumeBound / changeResume / prepStatus / jdMatch / sessionTag / reportReady / planEyebrow / planSub / switchPlan / createPlan / must / nice / hidden / risks / strongs / lastReport / gotoReport / notePractice / empty.* / missingResume.* / planSwitcher.* / resumePicker.* 等）；`messages.ts` 类型聚合补齐。

#### 1.5 Vitest 红灯 → 绿灯

新增 `workspace/WorkspaceScreen.test.tsx`：测 i18n zh/en 切换重绘、≥ 20 个 testid 存在（按 §3.5 UI source structure parity rows）、所有可点击 placeholder 触发 nav/start stub、控件类型断言（button / textarea / 非 select；resume picker 必须是 list of buttons 而非 `<select>`）、负向断言旧 prototype 中存在但当前真理源已移除的 testid（如 `practice-mode-card-*` / `growth-*` / `drill-builder-*`）不命中。

新增 `interview-context/InterviewContext.test.tsx`：测 reducer 各 action、route param hydration、清理逻辑、跨路由 carry 参数。

#### 1.6 BDD-Gate

- BDD-Gate: 验证 `E2E.P0.018` 中 workspace 静态部分（plan eyebrow + Interview Launcher + Main 双列 + TopBar 高亮）资产构建到 ready 态

### Phase 2: TargetJob 数据消费（listTargetJobs / getTargetJob）+ Header / Launcher / JD 拆解 / risks-strengths

#### 2.1 `useWorkspaceTargetJob()` hook

通过 D1 generated client 调 `getTargetJob(targetJobId)`；React state 跟踪 loading / data / error 三态；mount 时若 `InterviewContext.targetJobId` 缺失 → 立即返回 `empty` 状态（不发请求）；返回数据写入 `InterviewContext` 通过 `MERGE_TARGET_JOB` action。

#### 2.2 Header summary + Interview Launcher 数据接入

按 §3.7 mapping 把 generated `TargetJob` 字段注入：plan eyebrow 标题、status tag、subtitle；header tag/updatedAt/title/subtitle；Interview Launcher 的 Round Rail 数据（`TargetJob.summary` 当前是字符串，round 不从未声明结构读取；用 §3.7 fallback `[HR, Technical 1, Technical 2, Manager]`，roundId 通过 `getWorkspaceRoundId` 同款 helper 派生）；BindingPill JD 段只用 `title / companyName / locationText / sourceType` 与 route/context 派生字段。`prepStatus` / `jdMatch` 文案从 `TargetJob.fitSummary.strengths / gaps / riskSignals` 与 `openQuestionIssueCount` 派生，缺失时 fallback `—`；不得读取不存在的 `level / match / readinessLabel / statusTone / nextRound`。

#### 2.3 Main Left JD 拆解

按 `screen-workspace.jsx::ReqBlock` lines 742-759 源级复刻 `Must Have` / `Nice to Have` / `Hidden signals` 三 ReqBlock；数据来源 `TargetJob.requirements`（按 OpenAPI declared requirement fields 分组）+ `TargetJob.summary` 文本；hits 命中圆点逻辑保留为 view-model 派生：`TargetJob.fitSummary.strengths` 可映射命中提示，缺失时全部为空圆。不得使用未声明的 `summary.interviewHypotheses` / `summary.coreThemes` / `fitSummary.directHits`。

#### 2.4 Main Right risks/strengths

按 `screen-workspace.jsx::WorkspaceScreen` lines 224-239 源级复刻；数据来源 `TargetJob.fitSummary.strengths` + `TargetJob.fitSummary.riskSignals`，`gaps` 可并入风险/待补强列表（缺失时空态）。

#### 2.5 sessionHistory placeholder

本 phase 仅渲染 placeholder skeleton / `EmptyHistory` 文案（点击行 disabled）。真实 history 数据接入由后续 plan / future `listPracticeSessions` 或等价 typed history operation 落地；本 plan 不从 fixture extension 或 `any` 派生历史行。

#### 2.6 fixture variant

扩展 `openapi/fixtures/TargetJobs/getTargetJob.json` 新增 `with-rounds` / `not-found` variants（`with-rounds` 提供完整 requirements / fitSummary / coreThemes 用于 happy path；`not-found` 用于 `WorkspaceEmptyState`）；`listTargetJobs.json` 新增 `single-plan` variant（用于 Phase 3 Plan Switcher boundary）；`make validate-fixtures` 通过。

#### 2.7 Vitest

新增 `workspace/WorkspaceHeader.test.tsx`：测 fixture 三态渲染（loading skeleton → ready → 4xx error 占位）；header 各字段映射；JD 拆解 ReqBlock 数据驱动；risks/strengths 列表渲染；负向断言不读取不存在的字段（不依赖 `level` 等不属于 OpenAPI schema 的 key）；空数组与缺失字段 fallback。

#### 2.8 BDD-Gate

- BDD-Gate: 验证 `E2E.P0.018` 中 workspace 数据接入部分（header / Launcher / JD 拆解 / risks-strengths）通过

### Phase 3: 简历绑定 + Resume Picker + Plan Switcher Modal

#### 3.1 `useWorkspaceResume()` hook

通过 generated client 调 `getResume(resumeVersionId)`；React state 跟踪 loading / data / error 三态；缺失或 404 → 写入 `InterviewContext.resumeVersionId=null` 并触发 `WorkspaceMissingResumeState`。

#### 3.2 BindingPill Resume 段接入

按 `screen-workspace.jsx::WorkspaceScreen` line 191 源级复刻；title=`getResume.title`，meta 通过 `readResumeSummary(parsedSummary: Record<string, unknown> | null)` narrow helper 安全读取 `headline / yearsOfExperience` 后拼接；点击 `更换` 打开 `ResumePickerModal`。

#### 3.3 `ResumePickerModal.tsx` (disabled-list 模式)

按 `screen-workspace.jsx::ResumePickerModal` lines 517-585 源级复刻 DOM 结构 + 模态层 + footer 按钮；列表项：仅当前绑定 resume 启用 + 选中态；其余位置渲染 `disabled` 占位卡 + i18n 文案 `resumePicker.disabledNote`（中文：「更多简历版本将在 backend 开放 `listResumes` 接口后启用」），并在底部显示 spec §3.2 链接。`Use this resume` 按钮在仅有当前绑定 resume 时仍可点击但实际只关闭 modal（因为没有可切换项）。Vitest 负向断言 generated client `listResumes` 不存在或调用次数为 0。

#### 3.4 `PlanSwitcherModal.tsx`

按 `screen-workspace.jsx::PlanSwitcherModal` lines 587-666 源级复刻；通过 `useWorkspaceTargetJobs()` hook（调用 `listTargetJobs`，复用 home plan §3.7 viewmodel mapping）拉取候选 plan 列表；`从新 JD 创建规划` 调 `nav("home")`；`Use this plan` 调用 `updateInterviewContextFromPlan(plan)` 切换 `InterviewContext`，再触发 `useWorkspaceTargetJob()` 与 `useWorkspaceResume()` 重新拉取。

#### 3.5 keyboard / focus / a11y

ResumePickerModal 与 PlanSwitcherModal 必须实现：

- ESC 关闭 modal
- 外层遮罩点击关闭
- 关闭按钮（X）关闭
- focus trap 在 modal 内（首次打开 focus 第一个 focusable 元素；Tab 循环）
- 关闭后 focus 回到触发按钮

新增 `frontend/src/app/screens/workspace/modals/useModalA11y.ts` hook 封装。

#### 3.6 fixture variant

`getResume.json` 新增 `not-found` variant；`createPracticePlan.json` 已有 `default` / `missing-resume` 直接复用；`getPracticePlan.json` 新增 `archived` / `not-found` variants（用于 Phase 4）。

#### 3.7 Vitest

新增 `workspace/modals/ResumePickerModal.test.tsx`：测 DOM、disabled 列表渲染、文案 zh/en、a11y 行为、generated client `listResumes` 调用次数为 0；新增 `workspace/modals/PlanSwitcherModal.test.tsx`：测 DOM、`listTargetJobs` 接入、boundary（1 条 / 12+ 条）、a11y、`从新 JD 创建规划` 跳 home、`Use this plan` 切换 `InterviewContext` 并触发 refetch；新增 `workspace/WorkspaceMissingResumeState.test.tsx`：测缺简历空态。

#### 3.8 BDD-Gate

- BDD-Gate: 验证 `E2E.P0.018` 中 Plan Switcher / Resume Picker 部分；验证 `E2E.P0.019` 中 `WorkspaceMissingResumeState` 路径

### Phase 4: 立即面试双步契约 + getPracticePlan refresh + auth pendingAction

#### 4.1 `useWorkspacePracticePlan()` hook

mount 时若 `InterviewContext.planId` 存在 → 调 `getPracticePlan(planId)`；`status='ready'` → `MERGE_PRACTICE_PLAN`；`status='archived'` 或 404 → 重置 `InterviewContext.planId=null`（驱动后续 createPracticePlan）。不得假设 OpenAPI 未声明的 `cancelled/failed` plan status。

#### 4.2 `useStartPractice()` hook

实现立即面试双步契约：

```ts
async function start() {
  let planId = ctx.planId && planStatus === 'ready' ? ctx.planId : null;
  const idempotencyBatch = newIdempotencyBatch(); // frontend/src/lib/conventions/idempotency.ts
  if (!planId) {
    const plan = await client.createPracticePlan(buildCreatePlanRequest(ctx), {
      idempotencyKey: idempotencyBatch.create,
    });
    planId = plan.id;
    dispatch({ type: 'MERGE_PRACTICE_PLAN', plan });
  }
  const session = await client.startPracticeSession({ planId, hintsEnabled: ctx.practiceMode === 'assisted' }, {
    idempotencyKey: idempotencyBatch.start,
  });
  dispatch({ type: 'MERGE_SESSION', session });
  navigate({ name: 'practice', params: buildPracticeRouteParams(ctx, session) });
}
```

`buildCreatePlanRequest(ctx)` body 字段映射在 `frontend/src/app/interview-context/buildCreatePlanRequest.ts` 中实现；`questionBudget` 默认 6，`timeBudgetMinutes` 默认 30，`difficulty` 默认 `standard`，`interviewerPersona` 默认 `hiring_manager`，`focusCompetencyCodes` baseline 为空数组，`language` 取当前 UI locale，`mode` 默认 `assisted`（与 fixture default 一致）。

#### 4.3 立即面试 CTA 接线

`screen-workspace.jsx::WorkspaceScreen.startInterview` 同款视觉触发（line 102-114），但正式前端必须按当前 `useRequestAuth` 可执行契约接线：未登录调 `requestAuth({ type: 'start_practice', label: t('workspace.startCore'), route: 'workspace', params: { ...buildWorkspaceRouteParams(ctx), ...buildPracticeDisplayContext(ctx), autoStartPractice: '1' } })`；登录后 verify 完成回到 workspace，`WorkspaceScreen` 读取并清理 `autoStartPractice` 控制位，确认 `InterviewContext` 完整后自动调 `useStartPractice().start()`，成功后再 `nav("practice", buildPracticeRouteParams(ctx, session))`。Vitest 必须证明 `route='practice'` 不用于未登录 pendingAction。

#### 4.4 ButtonState

CTA 在请求中 disabled + spinner；4xx / 5xx 显示 inline 错误（在 Interview Launcher CTA 下方）+ 重试按钮；重试复用同一 `idempotencyBatch` 实现 dedupe；3 次失败后展示 fallback CTA `回到首页`。

#### 4.5 错误映射

- `createPracticePlan` 422 (`missing-resume`) → 提示去绑定简历，自动滚动并 focus `更换简历` 按钮
- `createPracticePlan` 422 (其他校验) → 通用 inline 错误 + retry
- `startPracticeSession` 502 `AI_PROVIDER_TIMEOUT` / `AI_OUTPUT_INVALID` / `AI_PROVIDER_SECRET_MISSING` → 显示 backend 同款错误码翻译 + retry
- 网络错误 → 通用错误占位

#### 4.6 fixture variant

`startPracticeSession.json` 新增 `ai-timeout-502` variant（响应 status=502, error.code=`AI_PROVIDER_TIMEOUT`）；`getPracticePlan.json` 新增 `archived` / `not-found` variants。

#### 4.7 Vitest

新增 `workspace/WorkspaceStartPractice.test.tsx`：测 happy path（无 plan → createPracticePlan → startPracticeSession → nav practice）；测 happy path（有 plan + ready → 跳过 createPracticePlan）；测 happy path（有 plan + archived → 重新 createPracticePlan）；测 createPracticePlan 4xx + 5xx；测 startPracticeSession 5xx + retry；测 `Idempotency-Key` 在 retry 复用；测 nav practice 携带完整 InterviewContext + `PracticeDisplayContext`（含 `practiceGoal`）字段；测 hintsEnabled 由二值 practiceMode 派生，并负向断言 workspace 不产出 `debrief_replay`。

新增 `workspace/WorkspaceAuthGate.test.tsx`：测未登录立即面试 → `requestAuth` 触发 → `auth_login` 携带 `pendingRoute=workspace` / `pendingType=start_practice` / `autoStartPractice=1` → 登录恢复 workspace → 自动 startPractice → nav practice；测 pendingAction params 仅含 IDs、route、`PracticeDisplayContext` 与 `autoStartPractice`，不含敏感字段。

#### 4.8 BDD-Gate

- BDD-Gate: 验证 `E2E.P0.020` 立即面试主路径 + 未登录恢复 + `E2E.P0.019` getPracticePlan 恢复路径

### Phase 5: CompanyIntelEmbed handoff + Session History handoff + 空态收口

#### 5.1 `CompanyIntelEmbed` 组件

按 `ui-design/src/screen-company-intel.jsx::CompanyIntelEmbed` 源级复刻；数据来源仅限 `getTargetJob.companyName / locationText / sourceType / summary`；不调 `getCompanyIntel`；`打开公司情报` 调 `nav("company_intel", { targetJobId, jdId })` handoff（route 仍由 D1 `PlaceholderScreen` 占位，由外部 owner 后续替换）。

#### 5.2 sessionHistory placeholder

本 plan 不实现 `listPracticeSessions`（缺契约），也不读取 `getTargetJob` fixture extension 中未声明的 `recentSessions[]`。session history 区域固定渲染 `EmptyHistory` / disabled placeholder + 文案 `首场面试将出现在这里`；点击不触发 `nav("report", ...)`。真实历史 session 行、缺 `reportId` disabled 行、以及 workspace → report handoff 必须等待 `listPracticeSessions` 或等价 typed history contract 在对应 owner plan 落地。

> **Phase 5 contract note**：当前 generated `TargetJob` 不含 `recentSessions`；本 plan 的正向测试必须锁定“不得通过 `any` / fixture extension 读取 history”。后续 owner 若要启用历史行，应先修订 OpenAPI + fixtures + generated client，再回到对应 plan 更新 BDD。

#### 5.3 WorkspaceEmptyState / WorkspaceMissingResumeState 收口

补齐两空态的 a11y 与跳转 CTA：

- `WorkspaceEmptyState` `导入 JD` CTA → `nav("home")`，并在 home `home-jd-textarea` 自动 focus
- `WorkspaceMissingResumeState` `创建简历` CTA → `nav("resume_versions", { flow: "create" })`

新增 `workspace/WorkspaceHandoff.test.tsx`：测 CompanyIntelEmbed 不调 `getCompanyIntel`；测 nav handoff 携带 `targetJobId / jdId`；测 sessionHistory 区域为 `EmptyHistory` / disabled placeholder 且点击不触发 report nav；负向断言不读取 `TargetJob.recentSessions` / 不调用 `getFeedbackReport`。

#### 5.4 BDD-Gate

- BDD-Gate: 验证 `E2E.P0.021` handoff 主路径 + 隐私红线 + 旧入口反向 grep

### Phase 6: 验证收口（pixel parity + scenario + regression rerun）

#### 6.1 Playwright pixel parity 扩展

新增 `frontend/tests/pixel-parity/workspace.spec.ts` 覆盖 desktop (1440×900) + mobile (390×844) 两 chromium project：

- DOM 锚点（plan eyebrow / header summary / Round Rail / Interview Launcher CTA / BindingPill 双卡 / Main Left CompanyIntelEmbed + JD 拆解 / Main Right risks-strengths + EmptyHistory / 两 modal）
- 关键元素 bounding box stays in viewport, no overlap
- mobile：Main 双列折叠为单列、Round Rail 横向滚动、Modal 全屏化
- warm/light → dark → customAccent 三态切换 computed background / color 可见变化
- toHaveScreenshot baseline 区域：workspace 主屏 / Plan Switcher Modal / Resume Picker Modal / WorkspaceEmptyState / WorkspaceMissingResumeState

`pnpm --filter @easyinterview/frontend test:pixel-parity` 全 PASS（在 D2/D3 + home plan 现有基础上累加）。

#### 6.2 Scenario 资产

派生 4 个新 scenario 目录：

- `test/scenarios/e2e/p0-018-workspace-default-render/`
- `test/scenarios/e2e/p0-019-workspace-context-loading/`
- `test/scenarios/e2e/p0-020-workspace-start-practice/`
- `test/scenarios/e2e/p0-021-workspace-handoff/`

每个目录含 `README.md`（§6 baseline 维护、§7 离线运行限制）+ `scripts/{setup,trigger,verify,cleanup}.sh`（按 `test/scenarios/README.md` + `test/scenarios/e2e/README.md` 规范）+ `data/seed-input.md` + `data/expected-outcome.md`。

#### 6.3 Scenario INDEX 更新

`test/scenarios/e2e/INDEX.md` P0 表追加 4 行（`E2E.P0.018` / `E2E.P0.019` / `E2E.P0.020` / `E2E.P0.021`），关联需求列指向 `frontend-workspace-and-practice C-1 ~ C-12`（按 §3.5 矩阵）；状态 Ready；执行方式 automated。

#### 6.4 Regression 重跑

`E2E.P0.001 / 002 / 004 / 005 / 006` 的 `setup → trigger → verify → cleanup` 全部 PASS；`E2E.P0.014 / 015 / 016 / 017` 仅在对应目录已存在且 home plan INDEX 标记 Ready 时执行，否则记录为“上游 active plan 条件 gate，当前不适用”。`pnpm --filter @easyinterview/frontend test`（全量 Vitest）+ `pnpm --filter @easyinterview/frontend typecheck` + `pnpm --filter @easyinterview/frontend build` + `make build` 全 PASS。

#### 6.5 文档与索引同步

`/sync-doc-index --fix-index` 把 `docs/spec/INDEX.md` 与 `docs/spec/frontend-workspace-and-practice/plans/INDEX.md` 同步到 Header 当前；`make docs-check` zero drift；`check_md_links` 双 OK。

#### 6.6 负向搜索

- `frontend/src/app/screens/workspace/` + `frontend/src/app/interview-context/` 不 import `ui-design/src/data.jsx` / `window.EI_DATA` / `getWorkspaceJDSample` / `getWorkspaceResumeOptions` / `getWorkspacePlanOptions` / `getWorkspaceSessionHistory` 等 prototype helper（0 命中）
- 旧 prototype workspace 业务 testid（`workspace-mocked-*` / `practice-mode-card-*` / `growth-*` / `drill-builder-*` / `mistake-queue-*`）grep 0 命中（除负向断言文件）
- 旧 route alias（`welcome` / `growth` / `mistakes` / `drill` / `followup` / `experiences` / `star` / 独立 `voice`）在 workspace 模块中 grep 0 命中（除 `app/normalizeRoute.ts` alias map）
- JD 原文 / 简历正文 grep — 仅出现在 React state / generated client request body / fixture，不出现在 `console.log` / URL / `localStorage` / telemetry 调用
- LLM/provider grep — workspace 模块不出现 provider key、provider registry、prompt registry、AIClient、LLM endpoint 或 ad hoc 绕过 generated client 的 fetch；`questionText` 不在 workspace 屏 DOM 中渲染
- generated client `listResumes` / `getCompanyIntel` 调用次数为 0（断言入口：`mockTransport` spy + tsc 类型检查）

#### 6.7 BDD-Gate

- BDD-Gate: 验证 `E2E.P0.018` / `E2E.P0.019` / `E2E.P0.020` / `E2E.P0.021` 全部通过 + D1+D2+D3 已存在 `E2E.P0.001/002/004/005/006` regression PASS；home plan `E2E.P0.014/015/016/017` 仅在场景资产存在且 INDEX Ready 时作为条件 regression gate

## 5 验收标准

- 本计划列出的 Phase 1-6 全部 checklist 项通过
- spec C-1 / C-2 / C-3 / C-7 / C-8 / C-9 / C-10 / C-12 全部覆盖且通过对应测试；C-11 的 workspace 子集（含 BDD 主流程 + 关键分支）通过；C-4 / C-5 / C-6 由后续 plan 002 / 003 / 004 承接，本 plan 不验证
- 关联 BDD-Gate（`E2E.P0.018 / 019 / 020 / 021`）全部通过；D1+D2+D3 已存在 regression（`P0.001/002/004/005/006`）全部 PASS；home plan regression（`P0.014/015/016/017`）仅在场景资产存在且 INDEX Ready 时执行并 PASS
- pixel parity 在 desktop + mobile 两 viewport 下 workspace 主屏 + 两 modal + 两空态新增 spec 全 PASS
- `make docs-check` zero drift；`check_md_links` 双 OK；`pnpm typecheck` 0 错；`pnpm build` + `make build` PASS
- 负向搜索（旧 prototype 业务 testid / 旧 route alias / prototype data 直接 import / JD 原文 / 简历正文泄漏 / `listResumes` / `getCompanyIntel` 调用）全部 0 命中

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| `getTargetJob` / Practice contract 缺 typed history 字段 → sessionHistory 无法落地 | Phase 5.2 固定只渲染 `EmptyHistory` / disabled placeholder；negative gate 确认不读取 `recentSessions` fixture extension、不通过 `any` 拼出 report handoff；后续 owner 先补 `listPracticeSessions` 或等价 typed contract |
| `listResumes` 缺 contract → Resume Picker 仅可显示当前绑定 resume | Phase 3.3 落 disabled-list 模式；Vitest 负向断言 generated client `listResumes` 不存在；UI 文案明示「待解锁」并指向 spec §3.2；plan 002+ 不直接复用 disabled-list 行为，由后续 contract 修订后切换 |
| `Idempotency-Key` 在 createPracticePlan + startPracticeSession 双步分配策略不一致导致重试不幂等 | 抽 `frontend/src/lib/conventions/idempotency.ts::newIdempotencyBatch()` 返回稳定 `{create, start}` 双键；retry 复用同一 batch；Vitest 锁定 retry 第二次调用 header 与首次一致 |
| `useStartPractice` 在 React StrictMode 双触发导致重复 createPracticePlan | hook 内部使用 `inFlightRef` + `Promise.resolve` 缓存当前 batch；首次调用入栈后 deduplicate；Vitest 在 StrictMode 下断言 generated client 调用次数 = 1 |
| pendingAction round-trip 时 startContext 未持久化导致登录后启动失败 | `requestAuth` params 携带完整 `InterviewContext` keys、`PracticeDisplayContext` 与 `autoStartPractice=1`；pending route 固定为 `workspace`，登录恢复后 workspace mount 检查并清理控制位再触发 `useStartPractice().start()`；新增 Vitest 锁定 `pendingRoute=workspace` 且不使用不可执行 callback |
| Plan Switcher Modal 切换 plan 后 `useWorkspaceTargetJob / Resume / PracticePlan` 三 hook 触发顺序不稳定导致中间态闪烁 | 在 `InterviewContext` reducer 中引入 `loadingTargetJob/loadingResume/loadingPlan` 三个 boolean，UI 在切换期间显示统一 skeleton；Vitest 用 fake timer 锁定渲染顺序 |
| pixel parity 跨字体子像素差异（D3 retrospective 经验） | 沿用 D3：`workspace.spec.ts` toHaveScreenshot 仅作 frontend 内部 regression（含 maxDiffPixels 阈值），不与 ui-design golden 跨字体源做硬 diff |
| Modal a11y focus trap 与 Vitest jsdom 不完全兼容 | 使用 `useModalA11y` hook + `aria-modal="true"` + `data-testid="..."`；Vitest 仅断言 attributes + Tab 焦点行为通过手动 `userEvent.tab()` 验证；E2E 留 Playwright 端到端 a11y 验证 |
| 旧 prototype data 渗透（开发者从 `screen-workspace.jsx` 复制粘贴时把 `getWorkspaceResumeOptions` / `getWorkspaceJDSample` 一并带过来） | Vitest negative grep + `eslint-rules` 反查（`no-restricted-imports` 限制 `ui-design/`）；scenario verify 阶段 grep `EI_DATA` / `getWorkspace*` literal |
| 历史 fixture `prototype-baseline` variant 与新 D 系列 contract 漂移导致 backend-targetjob 已交付 handler 与 fixture 字段不一致 | Phase 2.6 / 6.4 在 `make validate-fixtures` 后再跑 backend `make test`（如可达），并在 PR 描述中显式列出本 plan 修改的 fixture 文件；漂移由 `mock-contract-suite` parity test 兜底 |
| backend-practice handler 落地后真实响应 schema 微调导致 fixture / 前端反序列化错位 | 本 plan fixture-only；plan 提交时在 retrospective 中提示 backend-practice/001 owner：变更 `CreatePracticePlanRequest` / `StartPracticeSessionRequest` schema 必须先回 OpenAPI + 同步本 plan fixture 与 viewmodel mapping |
