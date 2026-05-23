# Frontend Workspace and Practice Spec

> **版本**: 1.4
> **状态**: active
> **更新日期**: 2026-05-23

## 1 背景与目标

`frontend-workspace-and-practice` 是 `engineering-roadmap` §5.2 `Mock Interview + Practice` workstream 的前端业务 subspec，承接 `frontend-shell` 已交付的 App 壳、TopBar、route normalization、`requestAuth(pendingAction)`、fixture-backed generated client 与 UI parity gate，以及 `frontend-home-job-picks-and-parse` 已锁定但仍在 active plan 中实现的 parse confirm 跳转契约。

本 subspec 的终稿范围收敛为三条当前 owner 路由：

- `workspace`：当前面试规划，确认 JD、目标岗位、绑定简历、轮次与会话历史，并发起 practice session。
- `practice`：文本 / 语音共享的 Interview Session 外层骨架、状态机消费、提示/严格模拟显隐、答题事件提交与结束动作。
- `generating`：`completePracticeSession` 后的报告生成过渡态，轮询 `getFeedbackReport(reportId)`，并把完成/失败状态 handoff 给下游 `report` owner。

`report` 和 `company_intel` 不并入本 subspec。`report` 的 dashboard、复练当前轮、进入下一轮与报告详情交互由 `frontend-report-dashboard` / future `backend-review` owner 承接；`company_intel` 详情页与数据源由外部 company-intel owner 承接。本 subspec 只在 `workspace` 中保留公司轻情报入口/摘要卡片和导航 handoff，不实现 `CompanyIntelScreen`。

本 subspec 通过 generated client + fixture-backed transport 消费已经存在的 TargetJobs / PracticePlans / PracticeSessions / Reports OpenAPI 契约；截至 2026-05-23，backend-resume、backend-practice、practice-voice-mvp 与 backend-review 已经落地本 spec 主路径依赖的真实 handler，前端 owner 的 completed plan 必须保留 fixture-backed UI variants，同时通过 `VITE_EI_API_MODE=real` generated-client gate 证明 production bootstrap 指向真实 backend base URL。任何新增或缺失 operation 仍须先回到 B2 / 对应 backend owner spec 修订，不能在前端手写 ad hoc fetch 或复制 `ui-design` mock data。

## 2 范围

### 2.1 In Scope

- `workspace` 屏（`route=workspace`）：
  - 当前面试规划页头：`公司·岗位` / 状态标签 / `当前轮次·绑定简历` / `切换规划` / `新建规划`，源级复刻 `ui-design/src/screen-workspace.jsx::WorkspaceScreen`。
  - Interview Launcher：面试轮次节点条 + 面试前确认 + `立即面试` 主 CTA + 目标岗位/JD + 绑定简历（`更换 → Resume Picker Modal`）。
  - Main Left：公司轻情报摘要入口（只渲染卡片与 `company_intel` handoff，不实现详情页）+ JD 拆解。
  - Main Right：我的准备（优势 / 风险提示）+ 当前规划的模拟面试历史占位。因当前 `TargetJob` / Practice OpenAPI 未声明 typed session history，本 subspec 的首个 plan 只渲染 `EmptyHistory` / disabled history placeholder；真实历史行和 report handoff 等待 `listPracticeSessions` 或等价 typed history contract 落地。
  - 缺 JD 显示 `WorkspaceEmptyState`，缺简历显示 `WorkspaceMissingResumeState`；不得用本地 mock 补字段。
- `practice` 屏（`route=practice`，`mode/modality∈{text,voice}` × `practiceMode∈{assisted,strict}`；`goal='debrief'` 由 plan/session 数据来源表达，不再作为 practiceMode）：
  - 顶部工具区（chrome 隐藏）：公司/岗位 + 面试官角色 + 题号/总数 + 计时 + 暂停 + 文本/语音形式切换 + 严格模拟开关。
  - 文本面试 `TextSurface`：对话记录 + 输入区 + `语音转文字` 麦克风 + 提示 + 跳过 + 提交。
  - 语音面试 `VoiceSurface`：`PracticeWaveformBars` + `PracticeAnnotatedWaveform` + 实时转写；具体 STT/LLM/TTS orchestration 归 `practice-voice-mvp`。
  - Left/Right Panel：题目地图、实时观察、JD 关联、可调用经历、AI 透明度、表达层指标、现场提示与音频留存说明，按 `practiceMode` 显隐。
  - PracticeSession 消费状态：`queued / running / waiting_user_input / completing / completed / failed / cancelled`（以 `shared/conventions.yaml` / `openapi/openapi.yaml` 当前 `SessionStatus` 七值为准）；前端不重复实现 backend 状态机，只消费 `PracticeSession` / `SessionEventResult` / `AssistantAction`。
- `generating` 屏（`route=generating`，chrome 隐藏）：
  - 源级复刻 `ui-design/src/screens-p0-complete.jsx::ReportGeneratingScreen` 的 4 步进度态、文案、节奏与布局。
  - 通过 `getFeedbackReport(reportId)` fixture variant `report-generating → default` 模拟轮询；`failed` variant 触发错误态。
  - 成功时导航 handoff 到 `report?sessionId&reportId`；失败时展示重试/返回 workspace 操作，但不渲染 Report Dashboard。
- 跨路由共享：
  - `InterviewContext` 在本 subspec owner route 内传递 `{planId, targetJobId, jdId, resumeVersionId, roundId, practiceMode, practiceGoal}`；外部 route 只要求自己的最小上下文。
  - 未登录用户点击 `立即面试` 时通过 `useRequestAuth({type:"start_practice", route:"workspace", params:{...InterviewContext, ...PracticeDisplayContext, autoStartPractice:"1"}})` 触发鉴权；登录后 `pendingAction` 回到 `workspace`，由 `WorkspaceScreen` 检测 `autoStartPractice=1` 并执行 `createPracticePlan` / `startPracticeSession`，成功后再跳转 `practice`。
  - `PracticeDisplayContext = {mode, modality, practiceMode, practiceGoal, hintUsed, hintCount}` 仅作为 practice → generating → report handoff 的路由展示上下文；`completePracticeSession` request body 严格使用 B2 `CompletePracticeSessionRequest{clientCompletedAt}`，不得把展示参数塞进 backend request。
- 契约消费形态：
  - `createPracticePlan`：workspace 创建 baseline plan，写 `Idempotency-Key` header。
  - `getPracticePlan`：workspace 拉取规划详情用于刷新 / 断网恢复。
  - `startPracticeSession`：workspace 点击「立即面试」时调用，写 `Idempotency-Key`；返回 `PracticeSession{currentTurn}` 直接驱动 practice 首屏。
  - `getPracticeSession`：practice 刷新 / 断网恢复。
  - `appendSessionEvent`：practice 屏用户操作通过单 endpoint + `kind` 路由；body 必须带 `clientEventId`，不得携带 `Idempotency-Key` header。
  - `completePracticeSession`：practice 屏点击「结束并生成报告」时调用，写 `Idempotency-Key`；返回 `ReportWithJob` 后进入 `generating`。
  - `getFeedbackReport`：generating 按 `reportId` 轮询，不按 `sessionId` 直接读取报告。

### 2.2 Out of Scope

- `ReportScreen` / Report Dashboard：Header、Context Strip、Summary Cards、Detail Surface、题目回顾、证据详情、复练当前轮、进入下一轮、无 `sessionId` 空态和报告失败态，由 `frontend-report-dashboard` / future `backend-review` 承接。
- `CompanyIntelScreen` / `getCompanyIntel` 数据源 / 合规来源详情页：由外部 company-intel owner 承接；本 spec 只保留 workspace 入口和导航 handoff。
- Home / Parse / JD Match shell 与 JD 导入解析业务，由 `frontend-home-job-picks-and-parse` 承接。
- Auth / TopBar / Sidebar / Theme / I18n bootstrap / requestAuth 接线 / generated client bootstrap，由 `frontend-shell` 承接。
- TargetJobs / PracticePlans / PracticeSessions / Reports 真实 backend handler / service / store / event 发射，由 `backend-targetjob`、`backend-practice`、`event-and-outbox-contract`、`db-migrations-baseline`、future `backend-review` 承接。
- 真实 STT / TTS provider 调用、prompt registry、AIClient orchestration，由 `ai-provider-and-model-routing`、`prompt-rubric-registry`、`practice-voice-mvp`、future `backend-practice` 承接。
- ResumeVersion CRUD、parse、占位简历资产，由 future `backend-resume` 与对应前端 owner 承接；本 spec 只消费当前 plan 绑定的 resume 只读字段。
- 不新增或恢复弃用模块 / 路由 / 术语作为 live UI：独立 `voice` route alias、独立 `PlanScreen`、独立 `VoicePracticeScreen`、入口前练习模式卡片、错题本/成长中心/单题深钻/追问树独立入口、报告时间线、报告一级导航、Inbox。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | Route owner 范围 | 本 subspec 只接管 `workspace / practice / generating`；`report / company_intel` 是外部 owner handoff | 消除与 `frontend-report-dashboard` / company-intel owner 的边界冲突，避免 plan 把报告详情和公司情报详情混入 workspace/practice |
| D-2 | workspace 语义 | `workspace` = 当前面试规划，不是岗位资产管理中心；不展示练习模式卡片，不提供专项练习入口；唯一主 CTA `立即面试` | 与 `module-job-workspace.md` §1-§2 + product-scope §5.2 一致 |
| D-3 | practice 三轴分离 | 形式 `mode/modality∈{text,voice}`、辅助度 `practiceMode∈{assisted,strict}`、数据来源 `practiceGoal/goal∈{baseline,retry_current_round,next_round,debrief}` 分离；`goal='debrief'` 可与 assisted 或 strict 组合，不再作为 practiceMode，也不天然禁用 hint | 与 backend-practice D-5/D-21 一致；当前 OpenAPI 仍可能残留 `legacy debrief replay value` generated enum 值，正式前端不得产出该旧值，B1/B2/backend-practice owner 负责收敛契约漂移 |
| D-4 | TopBar 隐藏 | `practice` 与 `generating` 路由隐藏 chrome；`workspace` 保留 chrome | 与 `routes.ts::NO_CHROME_ROUTES` 和 `ui-design/src/app.jsx::hideTopBar` 一致 |
| D-5 | Route 最小上下文 | 本 spec owner route 使用完整 InterviewContext；外部 `report` 最小键是 `sessionId/reportId`，外部 `company_intel` 最小键是 `targetJobId/jdId` | 不把可继承字段误判为所有 route 必填，避免无谓空态 |
| D-6 | Report handoff | generating 成功后只导航到 `report?sessionId&reportId`；ReportScreen 内部渲染、复练和下一轮动作由外部 owner 实现 | 本 spec 只验证 handoff 参数与生成态，不验证 report dashboard |
| D-7 | voice 入口唯一 | 语音面试只能通过 `practice?mode=voice&modality=voice` 进入；文本输入框麦克风是 `语音转文字` | 不恢复独立 voice route 或“切到语音”文案 |
| D-8 | company_intel handoff | workspace 可显示公司轻情报摘要和打开按钮；详情页、刷新、数据源、`getCompanyIntel` operation 不属于本 spec | 防止前端在无 contract 的情况下内嵌公司情报 mock |
| D-9 | 立即面试契约 | workspace 无 plan 时先 `createPracticePlan(goal='baseline')`，再 `startPracticeSession`；两步均携带 `Idempotency-Key`；未登录走 `requestAuth({type:'start_practice', route:'workspace', params:{autoStartPractice:'1', ...}})`，验证后由 workspace 自动执行同一双步契约 | 与 `module-job-workspace.md` §4.4、frontend-shell pendingAction、backend-practice D-13 对齐；不依赖 `useRequestAuth` callback |
| D-10 | backend 契约消费 | 只通过 B2 generated client 消费 OpenAPI operation；字段变化先回 B2/backend owner 修订 | 防止 screen 内自造 endpoint 或复制 fixture JSON |
| D-11 | voice 协作面 | 本 spec 拥有 voice surface React 组件、DOM/a11y/parity；`practice-voice-mvp` 拥有 `createPracticeVoiceTurn`、STT/LLM/TTS、committed context、barge-in | voice UI 与 voice orchestration 不双 owner |
| D-12 | appendSessionEvent 单 endpoint | 提交回答 / 请求提示 / 跳过 / 暂停 / 恢复都通过 `appendSessionEvent` + `kind`；仅 `practiceMode='strict'` 不渲染提示按钮，`goal='debrief'` 是否可提示由 practiceMode 决定 | 与 backend-practice D-7/D-16/D-21 一致 |
| D-13 | 完成是异步流 | `completePracticeSession` 返回 202 + `ReportWithJob{reportId,job}`；generating 用 `reportId` 轮询 `getFeedbackReport`，完成后 handoff 到 report owner | 前端不阻塞等待报告，不伪造 LLM 进度 |
| D-14 | fixture-backed + real-backend gate 红线 | completed frontend owner plan 可以保留 fixture-backed UI variants，但当对应 backend owner 已落地真实 handler 时，必须原地补 `VITE_EI_API_MODE=real` generated-client gate + scenario verify marker；缺失 operation 或 fixture 时仍先回 B2 / mock-contract-suite / backend owner，不用本地 mock 兜底 | 保护前后端分离契约，避免 fixture UI PASS 被误判为真实 backend 闭环 |

### 3.2 待确认事项

- Resume Picker 的 active-list 与版本展开契约已经由 backend-resume/B2 落地；completed workspace plan 若继续保留 disabled-list UX，必须在原 plan 中标为历史交付状态，并在后续 workspace owner 修订时切到 `listResumes` + `listResumeVersions(resumeAssetId)` real-mode gate。
- `createPracticeVoiceTurn` 已由 practice-voice-mvp / backend-practice voice extension 进入 generated client 与 fixture；voice surface 的完整 STT/LLM/TTS orchestration 仍归对应 voice owner，但正式前端不得继续把 operation 写作缺失。

## 4 设计约束

- 视觉与交互必须以 `ui-design/src/screen-workspace.jsx`、`ui-design/src/screen-practice.jsx`、`ui-design/src/screens-p0-complete.jsx::ReportGeneratingScreen`、`ui-design/src/screen-company-intel.jsx::CompanyIntelEmbed`、`ui-design/src/app.jsx`（route mapping / `INTERVIEW_CONTEXT_ROUTES` / `hideTopBar`）、`ui-design/src/primitives.jsx` 为唯一真理源进行源级复刻；不得二次设计。
- `WorkspaceScreen` 的页头 / Interview Launcher / Main Left / Main Right / 模态层必须与 `screen-workspace.jsx` 当前结构一致；`PlanSwitcherModal` / `ResumePickerModal` 的 keyboard close / focus trap 行为必须由正式前端补齐并测试。
- workspace 中的公司轻情报只复刻 `CompanyIntelEmbed` 卡片和 `nav('company_intel', context)` handoff；不得在本 spec 内实现 `CompanyIntelScreen` 或 `getCompanyIntel` 数据源。
- `PracticeScreen` 的 TopBar 工具区 / Left Panel / Main / Right Panel / 固定底部 CTA / 文本-语音 surface 切换 / 严格模拟开关必须与 `screen-practice.jsx` 当前结构一致；`practiceMode='strict'` 必须隐藏提示按钮、左侧实时观察、可调用经历、语音现场提示；`goal='debrief'` 只影响题目来源，不改变辅助度显隐。
- `ReportGeneratingScreen` 的 4 步进度态、文案、节奏和 layout 必须与 `screens-p0-complete.jsx::ReportGeneratingScreen` 一致；轮询使用 generated `getFeedbackReport(reportId)`，不得在前端引入 AI provider / prompt registry / LLM key。
- route context 最小键必须按下表执行：

| Route | 本 spec owner | 最小上下文 | 缺失处理 |
|-------|---------------|------------|----------|
| `workspace` | 是 | `targetJobId`；推荐携带 `planId/jdId/resumeVersionId/roundId` | 缺 target/JD 显示 `WorkspaceEmptyState`；缺 resume 显示 `WorkspaceMissingResumeState` |
| `practice` | 是 | `sessionId` 或可启动的 `planId`；推荐携带完整 InterviewContext | 缺 session/plan 时回 workspace 空态，不展示假问题 |
| `generating` | 是 | `sessionId + reportId` | 缺 `reportId` 显示生成态错误/返回 workspace |
| `report` | 否 | `sessionId + reportId` | 由 report owner 处理 |
| `company_intel` | 否 | `targetJobId`；推荐携带 `jdId` | 由 company-intel owner 处理 |

- `PracticeDisplayContext = {mode, modality, practiceMode, practiceGoal, hintUsed, hintCount}` 只在 owner route/handoff context 中传递；`completePracticeSession` body 只发送 B2 `clientCompletedAt`，报告 owner 只展示练习方式和提示记录，不得推导通过率。
- 隐私红线：raw audio / TTS audio / transcript 明文 / LLM prompt-response 明文 / JD 原文 / 简历正文不得进入 console.log / URL query / localStorage / telemetry payload；fixture transport 不得在日志中泄漏。
- 暗色 / customAccent / 主题切换必须在 owner 三屏和 workspace 公司轻情报摘要卡片中通过 root `data-theme/data-mode/data-custom-accent` 生效。
- I18n 必须支持 zh / en；新增 `workspace.*` / `practice.*` / `generating.*` 命名空间；report/companyIntel 文案归外部 owner。
- Pixel parity gate 必须在 desktop (1440×900) + mobile (390×844) 两个 viewport 下断言 owner 三屏的 DOM 锚点 / computed style / bounding box / 截图差异；workspace company-intel 摘要卡片随 workspace gate 覆盖。
- Mobile 响应式：workspace 主左右列折叠；practice 三栏折叠为单列 + 底部 sheet；generating 居中进度态不溢出视口。
- `data-testid` 遵循 D1/D2 命名，使用 `workspace-*` / `practice-*` / `generating-*` 前缀；report/company-intel 前缀归外部 owner。
- stale-contract negative gate 必须区分“禁止作为 live UI/runtime 正向入口”和“允许出现在负向断言/禁止清单/历史说明”。旧 route/module 名称不得作为 active route、TopBar 项、正向 testid、正向 scenario 或用户可见入口重新出现。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| workspace / practice / generating UI | `frontend-workspace-and-practice`（本 spec） | 三屏 React 组件、InterviewContext store/hook、PracticeSession 消费、source parity、visual parity、i18n、a11y、responsive |
| Report Dashboard UI | `frontend-report-dashboard` / future `backend-review` | `ReportScreen`、报告详情、复练当前轮、进入下一轮、report 空态/失败态 |
| Company Intel 详情 | external company-intel owner | `CompanyIntelScreen`、`getCompanyIntel`、刷新、公开来源 freshness、合规数据源 |
| App shell / routes / auth / runtime / theme | `frontend-shell` | TopBar、NO_CHROME_ROUTES、requestAuth、generated client bootstrap、mock transport、display preferences |
| Home / Parse / JD Match | `frontend-home-job-picks-and-parse` | parse confirm 跳转 workspace，JD 导入与推荐 |
| TargetJobs backend | `backend-targetjob` | `listTargetJobs/getTargetJob/updateTargetJob/importTargetJob` handler/store/event |
| Practice backend | `backend-practice` | 6 个 Practice operation handler/service/store、state machine、AssistantAction、outbox、idempotency |
| Voice orchestration | `practice-voice-mvp` + `backend-practice` voice extension | `createPracticeVoiceTurn` contract/handler、STT/LLM/TTS、barge-in、committed-context |
| Report generation data | future `backend-review` | `feedback_reports`、question assessments、readiness、report job result |
| Resume data | [`backend-resume`](../backend-resume/spec.md) | Resume asset list/detail/version semantics；workspace 可消费绑定 resume 只读字段，后续 active picker 消费 `listResumes` / `listResumeVersions` |
| OpenAPI / fixtures / codegen | `openapi-v1-contract` + `mock-contract-suite` | `openapi/openapi.yaml`、fixtures、generated Go/TS artifacts、fixture-backed mock transport |

### 5.1 Operation Matrix

| operationId | Fixture | Frontend consumer | Backend handler | Persistence | AI dependency | Scenario / status |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `listTargetJobs` | `openapi/fixtures/TargetJobs/listTargetJobs.json` (`default`, `prototype-baseline`) | `WorkspaceScreen` plan switcher / fallback recent plans | `backend/internal/targetjob` implemented | `target_jobs` | none in frontend | `001-workspace-and-interview-context` |
| `getTargetJob` | `openapi/fixtures/TargetJobs/getTargetJob.json` (`default`, `prototype-baseline`) | `WorkspaceScreen` JD / requirements / company meta | `backend/internal/targetjob` implemented | `target_jobs`, requirements/sources | none in frontend | `001-workspace-and-interview-context` |
| `getResume` | `openapi/fixtures/Resumes/getResume.json` (`default`) | Bound resume summary only | backend-resume real handler | resume assets | none | `001` bound summary + real-mode gate |
| `listResumes` | `openapi/fixtures/Resumes/listResumes.json` (`default`) | Resume picker list / debrief picker / resume workshop | backend-resume real handler | resume assets | none | completed owner plans must run real-mode gate before fixture UI variants |
| `listResumeVersions` | `openapi/fixtures/Resumes/listResumeVersions.json` (`default`) | Resume picker version expansion / debrief picker / resume workshop | backend-resume real handler | resume_versions | none | completed owner plans must run real-mode gate before fixture UI variants |
| `createPracticePlan` | `openapi/fixtures/PracticePlans/createPracticePlan.json` (`default`, `missing-resume`, `debrief-derived`) | Workspace `立即面试` / debrief replay when no plan | backend-practice real handler | `practice_plans` | backend-only first-question prep | `001` + `frontendOwners.realApiMode.test.ts` |
| `getPracticePlan` | `openapi/fixtures/PracticePlans/getPracticePlan.json` (`default`) | Workspace refresh / recovery | backend-practice real handler | `practice_plans` | none | `001` + real-mode gate |
| `startPracticeSession` | `openapi/fixtures/PracticeSessions/startPracticeSession.json` (`default`, `debrief-derived-first-question`) | Workspace start + auth resume + debrief replay | backend-practice real handler | `practice_sessions`, first turn | backend-only `practice.session.first_question` | `001` / debrief + real-mode gate |
| `listPracticeSessions` | `openapi/fixtures/PracticeSessions/listPracticeSessions.json` (`default`) | Debrief mock-session picker / history handoff owner | backend-practice real handler | `practice_sessions` | none | debrief + real-mode gate |
| `getPracticeSession` | `openapi/fixtures/PracticeSessions/getPracticeSession.json` (`default`, `prototype-baseline`, `missing-session`) | Practice refresh / recovery | backend-practice real handler | `practice_sessions`, turns/events | none in frontend | `002` + real-mode gate |
| `appendSessionEvent` | `openapi/fixtures/PracticeSessions/appendSessionEvent.json` (`default`) | Practice answer/hint/skip/pause/resume | backend-practice real handler | `practice_session_events`, `practice_turns` | backend-only follow-up/hint | `002` + real-mode gate；仍不带 Idempotency-Key |
| `completePracticeSession` | `openapi/fixtures/PracticeSessions/completePracticeSession.json` (`default`) | Practice finish CTA | backend-practice real handler | session status + outbox | none in frontend | `002` + real-mode gate |
| `getFeedbackReport` | `openapi/fixtures/Reports/getFeedbackReport.json` (`default`, `report-generating`, `prototype-baseline`) | Generating poll by `reportId` only；report owner consumes dashboard | backend-review real handler | `feedback_reports` + job result | backend-review only | report dashboard + real-mode gate |
| `createPracticeVoiceTurn` | `openapi/fixtures/PracticeSessions/createPracticeVoiceTurn.json` | Voice surface turn submission | practice-voice/backend-practice real handler | voice session events | STT/LLM/TTS backend-only | practice-voice owner + real-mode gate |
| `getCompanyIntel` | N/A | Out of scope for this spec | external owner; no contract declared in this spec | external intel store/source cache | possible backend-only summarization | external owner |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 三条 owner route 专属 Screen 接管 | `frontend-shell` D1 已交付，owner route 当前由 PlaceholderScreen 占位 | 进入 `workspace` / `practice` / `generating` | 三条 route 渲染正式 Screen；`practice/generating` 仍隐藏 chrome；`report/company_intel` 不由本 spec 实现 | 001 / 002 / 004 |
| C-2 | Workspace 渲染 + 空态 | InterviewContext 至少包含 `targetJobId`，可选 `planId/jdId/resumeVersionId/roundId` | 进入 `workspace` | 渲染当前规划、Interview Launcher、JD 拆解、绑定简历、公司轻情报摘要入口、准备信号、会话历史；缺 JD/target 或 resume 时进入对应空态；不展示假数据 | 001 |
| C-3 | Workspace 交互闭环 | 已渲染 workspace | 用户点击 `切换规划` / `更换简历` / `立即面试` | 切换规划更新 InterviewContext；更换简历按 B2/listResumes 决策执行；立即面试用 generated client 调 `createPracticePlan`（必要时）→ `startPracticeSession`，副作用请求带 `Idempotency-Key`；未登录 pendingAction 恢复到 workspace 后自动执行双步启动，再进入 practice | 001 |
| C-4 | Practice 文本 happy path | 用户进入 `practice?mode=text&modality=text&practiceMode=assisted`，session=`running` | 用户输入回答、请求提示/跳过/暂停/恢复、提交事件、结束 | TextSurface 源级复刻；操作通过 `appendSessionEvent({clientEventId,kind,payload})`；AssistantAction 驱动下一题/追问/完成；结束调用 `completePracticeSession` 后进入 `generating?sessionId&reportId` | 002 |
| C-5 | Practice 语音 surface + strict/debrief goal 显隐 | 用户进入 `practice?mode=voice&modality=voice&practiceMode=strict`，以及 `goal='debrief'` 分别组合 assisted/strict | 用户进行语音回答或切换形式 | VoiceSurface 源级复刻；strict 隐藏提示、实时观察、可调用经历和现场提示；assisted+debrief 仍显示可用提示，strict+debrief 隐藏提示；不直连 STT/TTS provider；full voice turn flow 等 `createPracticeVoiceTurn` contract landed 后启用 | 003 |
| C-6 | Generating 轮询 + report handoff | Practice 已 `completePracticeSession` 收到 `ReportWithJob{reportId,job}` | 用户在 generating 屏等待 | 4 步进度态与 `ReportGeneratingScreen` 一致；`queued/running` 保持等待，`succeeded` 导航 `report?sessionId&reportId`，`failed` 显示错误/重试/返回 workspace；不渲染 Report Dashboard | 004 |
| C-7 | Downstream handoff 参数 | workspace 公司轻情报入口存在；generating 成功；session history 仍缺 typed contract | 用户点击打开情报 / 生成完成 / 查看 history 占位 | 对 `company_intel` handoff 携带 `targetJobId/jdId`；generating 对 `report` handoff 携带 `sessionId/reportId`；plan 001 的 workspace history 行保持 disabled placeholder，不伪造 `sessionId/reportId`；目标屏渲染归外部 owner | 001 / 004 |
| C-8 | UI source structure parity | C-1~C-7 通过 | Vitest+jsdom 加载 owner Screen | DOM 锚点、控件类型、icon、aria、keyboard、menu/modal 层级可追溯到 `screen-workspace.jsx` / `screen-practice.jsx` / `ReportGeneratingScreen` / `CompanyIntelEmbed` / `primitives.jsx` | 001 / 002 / 003 / 004 |
| C-9 | UI visual geometry parity | C-8 通过 | Playwright desktop + mobile 加载 owner 三屏 | 关键区块不重叠且 stays in viewport；theme/dark/customAccent 可见；workspace/practice/generating mobile 布局符合原型 | 001 / 002 / 003 / 004 |
| C-10 | UI stale-contract negative search | C-8 + C-9 通过 | lint/grep gate 扫描 active runtime、positive tests、README、scenario | 旧 route/module 不作为 live route、TopBar 项、正向 testid、正向 scenario 或用户入口出现；负向断言/禁止清单命中被分类允许 | 001 / 002 / 003 / 004 |
| C-11 | BDD 主流程 + 关键分支 | 三条 owner route + parity gate 已就绪 | 创建并执行 E2E 场景 | 覆盖 workspace 渲染/切换/更换简历决策、未登录立即面试恢复、practice 文本、practice 语音 surface、strict 显隐、generating report handoff、旧口径负向 | 001 / 002 / 003 / 004 |
| C-12 | Privacy 红线 | 用户完成 workspace→practice→generating 流程（文本 + 语音 surface 各一） | 检查 URL/localStorage/log/telemetry/fixture transport | raw audio / TTS audio / transcript 明文 / LLM prompt-response 明文 / JD 原文 / 简历正文不泄漏；只允许 IDs、状态、摘要和必要 route context | 001 / 002 / 003 / 004 |

## 7 关联计划

本 spec v1.2 已创建首个 active plan 目录 `001-workspace-and-interview-context`；其余计划编号仍为预留，后续通过 `/design` 创建对应 plan/context 后再进入 `/implement`：

- `001-workspace-and-interview-context` — workspace 接管 + InterviewContext store/hook + `listTargetJobs/getTargetJob/getResume/getPracticePlan/createPracticePlan/startPracticeSession` 消费 + auth pendingAction + workspace BDD；2026-05-23 L2 remediation 将 P0.018-P0.021 trigger 接入 `frontendOwners.realApiMode.test.ts`。
- `002-practice-text-event-loop` — PracticeScreen 文本 surface + `getPracticeSession/appendSessionEvent/completePracticeSession` 消费 + assisted/strict 辅助度策略 + `goal='debrief'` 数据来源显隐回归 + generating 入口；2026-05-23 L2 remediation 将 P0.044-P0.047 trigger 接入 `frontendOwners.realApiMode.test.ts`。
- `003-practice-voice-surface` — VoiceSurface / WaveformBars / AnnotatedWaveform / SpeechToText source parity + audio capture/playback/barge-in UI event reporting；full `createPracticeVoiceTurn` 等 B2 contract landed 后启用。
- `004-generating-report-handoff` — GeneratingScreen 轮询 `getFeedbackReport(reportId)` + succeeded/failed handoff 到 external report owner + BDD。

不再预留本 subspec 内的 `report` 或 `company_intel` plan；相关工作由外部 owner spec 原地创建。

## 8 关联文档

- 上游 spec：[`engineering-roadmap`](../engineering-roadmap/spec.md) §5.2、[`product-scope`](../product-scope/spec.md) §5.2-§5.3、[`frontend-shell`](../frontend-shell/spec.md)、[`frontend-home-job-picks-and-parse`](../frontend-home-job-picks-and-parse/spec.md)、[`backend-practice`](../backend-practice/spec.md)、[`practice-voice-mvp`](../practice-voice-mvp/spec.md)、[`backend-targetjob`](../backend-targetjob/spec.md)、[`backend-auth`](../backend-auth/spec.md)、[`openapi-v1-contract`](../openapi-v1-contract/spec.md)、[`event-and-outbox-contract`](../event-and-outbox-contract/spec.md)、[`db-migrations-baseline`](../db-migrations-baseline/spec.md)、[`shared-conventions-codified`](../shared-conventions-codified/spec.md)、[`mock-contract-suite`](../mock-contract-suite/spec.md)
- UI 真理源：`ui-design/src/screen-workspace.jsx`、`ui-design/src/screen-practice.jsx`、`ui-design/src/screens-p0-complete.jsx::ReportGeneratingScreen`、`ui-design/src/screen-company-intel.jsx::CompanyIntelEmbed`、`ui-design/src/app.jsx`（route mapping / `INTERVIEW_CONTEXT_ROUTES` / `hideTopBar`）、`ui-design/src/primitives.jsx`、[`docs/ui-design/module-job-workspace.md`](../../ui-design/module-job-workspace.md)、[`docs/ui-design/module-practice-review.md`](../../ui-design/module-practice-review.md)、[`docs/ui-design/module-map.md`](../../ui-design/module-map.md)、[`docs/ui-design/INDEX.md`](../../ui-design/INDEX.md)
- 当前正式前端入口：`frontend/src/app/{routes.ts,App.tsx,screens/PlaceholderScreen.tsx}`、`frontend/src/api/{generated/client.ts,mockTransport.ts}`、`frontend/src/app/runtime/AppRuntimeProvider.tsx`、`frontend/src/app/auth/pendingAction.ts`、`frontend/src/app/i18n/locales/{zh,en}.ts`、`frontend/src/app/theme/`、`frontend/src/app/scenarios/p0-005-app-shell-visual-system-smoke.test.tsx`、`frontend/tests/pixel-parity/`
- Fixture：`openapi/fixtures/TargetJobs/`（listTargetJobs / getTargetJob）、`openapi/fixtures/Resumes/`（getResume / registerResume）、`openapi/fixtures/PracticePlans/`（createPracticePlan / getPracticePlan）、`openapi/fixtures/PracticeSessions/`（startPracticeSession / getPracticeSession / appendSessionEvent / completePracticeSession）、`openapi/fixtures/Reports/getFeedbackReport.json`
- 治理 / 流程：[`AGENTS.md`](../../../AGENTS.md)、[`docs/development.md`](../../development.md) §2、[`docs/spec/README.md`](../README.md)、[`docs/spec/TEMPLATES.md`](../TEMPLATES.md)、[`test/scenarios/README.md`](../../../test/scenarios/README.md)
- 历史：[history.md](./history.md)
