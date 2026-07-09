# Frontend Workspace and Practice Spec

> **版本**: 1.15
> **状态**: active
> **更新日期**: 2026-07-09

## 1 背景与目标

`frontend-workspace-and-practice` 是 `engineering-roadmap` §5.2 `Mock Interview + Practice` workstream 的前端业务 subspec，承接 `frontend-shell` 已交付的 App 壳、TopBar、route normalization、`requestAuth(pendingAction)`、fixture-backed generated client 与 UI parity gate，以及 `frontend-home-job-picks-and-parse` 已交付的 parse confirm 跳转契约。

本 subspec 的终稿范围收敛为三条当前 owner 路由：

- `workspace`：纯面试规划列表。一级 `面试` / `/workspace` 始终展示可继续的 ready TargetJob 卡片列表，不读取、不继承、不解释 `targetJobId` / `planId` / `resumeId` 等详情上下文；列表卡片进入 `parse?targetJobId=...` 的统一“面试规划详情 / 面试上下文确认”母版。
- `practice`：文本 / 语音共享的 Interview Session 外层骨架、状态机消费、提示/严格模拟显隐、答题事件提交与结束动作。
- `generating`：`completePracticeSession` 后的报告生成过渡态，轮询 `getFeedbackReport(reportId)`，并把完成/失败状态 handoff 给下游 `report` owner。

`report` 不并入本 subspec。`report` 的 dashboard、复练当前轮、进入下一轮与报告详情交互由 `frontend-report-dashboard` / `backend-review` owner 承接。公司轻情报仅作为 `workspace` 的嵌入摘要，由 TargetJob 摘要字段和当前路由上下文驱动。

本 subspec 通过 generated client + fixture-backed transport 消费已经存在的 TargetJobs / PracticePlans / PracticeSessions / Reports OpenAPI 契约；截至 2026-05-23，backend-resume、backend-practice、practice-voice-mvp 与 backend-review 已经落地本 spec 主路径依赖的真实 handler，前端 owner 的 completed plan 必须保留 fixture-backed UI variants，同时通过 `VITE_EI_API_MODE=real` generated-client gate 证明 production bootstrap 指向真实 backend base URL。任何新增或缺失 operation 仍须先回到 B2 / 对应 backend owner spec 修订，不能在前端手写 ad hoc fetch 或复制 `ui-design` mock data。

## 2 范围

### 2.1 In Scope

- `workspace` 屏（`route=workspace`）：
  - 面试规划列表 landing：顶部一级 `面试` 入口和任何 `/workspace` route 都展示已有规划卡片列表，使用 generated `listTargetJobs(analysisStatus=ready)`；每个规划必须有独立卡片背景、1px 边框、轻阴影、内部分区和底部操作区，不能退化成无容器文本列；卡片只展示状态、更新时间、岗位、公司和地点，不展示来源类型 / 目标语言 / `手动输入` 等导入元信息；失败解析、非 ready、空标题 TargetJob 不得进入列表；点击主题强调色 CTA 导航到 `parse` 统一面试规划详情；无规划时引导回首页导入 JD。
  - 路由纯度：`workspace` 不拥有 TargetJob 详情、Resume Picker、Plan Switcher、`autoStartPractice` 或 session 创建；即使 URL / stale context 带有 `targetJobId` / `planId` / `resumeId`，也必须清理/忽略并继续渲染列表。
  - 规划详情与启动：统一详情、保存规划、切换/绑定简历和 `立即面试` 由 `frontend-home-job-picks-and-parse` / practice/report handoff owner 通过 generated REST client 承接；`workspace` 不作为副作用中转页。
- `practice` 屏（`route=practice`，`mode/modality∈{text,voice}` × `practiceMode∈{assisted,strict}`；`practiceGoal∈{baseline,retry_current_round,next_round}`）：
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
  - `InterviewContext` 在 `practice / generating / report` owner route 内传递 `{planId, targetJobId, jdId, resumeId, roundId, practiceMode, practiceGoal}`；`workspace` 是纯列表，不参与 InterviewContext carry。
  - `parse` 详情页或报告复练 CTA 触发 `立即面试` 时直接通过 generated client 执行 `createPracticePlan`（必要时）→ `startPracticeSession`，成功后导航 `practice`；未登录时 pendingAction 回到原 owner 页面重试，不回到 `workspace(autoStartPractice=1)`。
  - `PracticeDisplayContext = {mode, modality, practiceMode, practiceGoal, hintUsed, hintCount}` 仅作为 practice → generating → report handoff 的路由展示上下文；`completePracticeSession` request body 严格使用 B2 `CompletePracticeSessionRequest{clientCompletedAt}`，不得把展示参数塞进 backend request。
- 契约消费形态：
  - `createPracticePlan`：parse / report handoff owner 创建 baseline / retry / next-round plan，写 `Idempotency-Key` header。
  - `getPracticePlan`：parse / report handoff owner 校验已有 plan 是否仍匹配当前 target/resume context。
  - `startPracticeSession`：parse / report handoff owner 点击「立即面试 / 复练 / 下一轮」时调用，写 `Idempotency-Key`；返回 `PracticeSession{currentTurn}` 直接驱动 practice 首屏。
  - `getPracticeSession`：practice 刷新 / 断网恢复。
  - `appendSessionEvent`：practice 屏用户操作通过单 endpoint + `kind` 路由；body 必须带 `clientEventId`，不得携带 `Idempotency-Key` header。
  - `completePracticeSession`：practice 屏点击「结束并生成报告」时调用，写 `Idempotency-Key`；返回 `ReportWithJob` 后进入 `generating`。
  - `getFeedbackReport`：generating 按 `reportId` 轮询，不按 `sessionId` 直接读取报告。

### 2.2 Out of Scope

- `ReportScreen` / Report Dashboard：Header、Context Strip、Summary Cards、Detail Surface、题目回顾、证据详情、复练当前轮、进入下一轮、无 `sessionId` 空态和报告失败态，由 `frontend-report-dashboard` / `backend-review` 承接。
- 公司轻情报独立页面、独立刷新 API 与公开来源详情页：不属于本 owner；本 spec 只消费 workspace 内嵌摘要卡片所需的 TargetJob 摘要字段。
- Home / Parse 与 JD 导入解析业务，由 `frontend-home-job-picks-and-parse` 承接。
- Auth / TopBar / Sidebar / Theme / I18n bootstrap / requestAuth 接线 / generated client bootstrap，由 `frontend-shell` 承接。
- TargetJobs / PracticePlans / PracticeSessions / Reports 真实 backend handler / service / store / event 发射，由 `backend-targetjob`、`backend-practice`、`event-and-outbox-contract`、`db-migrations-baseline`、`backend-review` 承接。
- 真实 STT / TTS provider 调用、prompt registry、AIClient orchestration，由 `ai-provider-and-model-routing`、`prompt-rubric-registry`、`practice-voice-mvp`、`backend-practice` 承接。
- Resume create / parse / tailor / edit，由 `backend-resume` 与对应前端 owner 承接；本 spec 只消费当前 plan 绑定的 flat Resume 只读字段。
- 不新增非当前 live UI：独立 `voice` route alias、独立 `PlanScreen`、独立 `VoicePracticeScreen`、入口前练习模式卡片、错题本/成长中心/单题深钻/追问树独立入口、报告时间线、报告一级导航、Inbox。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | Route owner 范围 | 本 subspec 只接管 `workspace / practice / generating`；`report` 是外部 owner handoff | 消除与 `frontend-report-dashboard` 的边界冲突，避免 plan 把报告详情混入 workspace/practice |
| D-2 | workspace 语义 | `workspace` = 面试规划列表，不是岗位资产管理中心，也不是当前规划详情/启动页；不展示练习模式卡片，不提供专项练习入口 | 与 `module-job-workspace.md` §1-§2 + product-scope §5.2 一致 |
| D-3 | practice 三轴分离 | 形式 `mode/modality∈{text,voice}`、辅助度 `practiceMode∈{assisted,strict}`、数据来源 `practiceGoal/goal∈{baseline,retry_current_round,next_round}` 分离 | 与 backend-practice D-22 / product-scope D-22 一致；正式前端只产出当前三类 practice goal |
| D-4 | TopBar 隐藏 | `practice` 与 `generating` 路由隐藏 chrome；`workspace` 保留 chrome | 与 `routes.ts::NO_CHROME_ROUTES` 和 `ui-design/src/app.jsx::hideTopBar` 一致 |
| D-5 | Route 最小上下文 | `workspace` 无业务上下文；`practice/generating/report` 使用各自最小上下文，`report` 最小键是 `sessionId/reportId` | 不把可继承字段误判为所有 route 必填，避免无谓空态 |
| D-6 | Report handoff | generating 成功后只导航到 `report?sessionId&reportId`；ReportScreen 内部渲染、复练和下一轮动作由外部 owner 实现 | 本 spec 只验证 handoff 参数与生成态，不验证 report dashboard |
| D-7 | voice 入口唯一 | 语音面试只能通过 `practice?mode=voice&modality=voice` 进入；文本输入框麦克风是 `语音转文字` | 不恢复独立 voice route 或“切到语音”文案 |
| D-8 | 公司轻情报嵌入边界 | workspace 可显示公司轻情报摘要和刷新/查看提示按钮；按钮停留在 `workspace` 并只携带当前 safe params；数据来自 TargetJob 摘要字段 | 防止前端在无 contract 的情况下扩展独立情报页面或 API |
| D-9 | 立即面试契约 | `parse` / report handoff owner 无匹配 ready plan 时先 `createPracticePlan`，再 `startPracticeSession`；两步均携带 `Idempotency-Key`；`workspace` 不执行 session 创建 | 与 `module-job-workspace.md` §4.4、frontend-shell pendingAction、backend-practice D-13 对齐；不依赖 route 副作用页 |
| D-10 | backend 契约消费 | 只通过 B2 generated client 消费 OpenAPI operation；字段变化先回 B2/backend owner 修订 | 防止 screen 内自造 endpoint 或复制 fixture JSON |
| D-11 | voice 协作面 | 本 spec 拥有 voice surface React 组件、DOM/a11y/parity；`practice-voice-mvp` 拥有 `createPracticeVoiceTurn`、STT/LLM/TTS、committed context、barge-in | voice UI 与 voice orchestration 不双 owner |
| D-12 | appendSessionEvent 单 endpoint | 提交回答 / 请求提示 / 跳过 / 暂停 / 恢复都通过 `appendSessionEvent` + `kind`；仅 `practiceMode='strict'` 不渲染提示按钮；`practiceGoal` 仅表达 `baseline / retry_current_round / next_round` 数据来源，不改变辅助度显隐 | 与 backend-practice D-7/D-16/D-21/D-22 一致 |
| D-13 | 完成是异步流 | `completePracticeSession` 返回 202 + `ReportWithJob{reportId,job}`；generating 用 `reportId` 轮询 `getFeedbackReport`，完成后 handoff 到 report owner | 前端不阻塞等待报告，不伪造 LLM 进度 |
| D-15 | 简历扁平化绑定（product-scope D-20） | `parse` / practice / report handoff context 使用 `resumeId`；workspace 列表只展示 target job 已绑定摘要，不选择或更换简历 | 与 [B2 D-26](../openapi-v1-contract/spec.md) / [frontend-resume-workshop D-8](../frontend-resume-workshop/spec.md) 同步 |
| D-14 | fixture-backed + real-backend gate 红线 | completed frontend owner plan 可以保留 fixture-backed UI variants，但当对应 backend owner 已落地真实 handler 时，必须原地补 `VITE_EI_API_MODE=real` generated-client gate + scenario verify marker；缺失 operation 或 fixture 时仍先回 B2 / mock-contract-suite / backend owner，不用本地 mock 兜底 | 保护前后端分离契约，避免 fixture UI PASS 被误判为真实 backend 闭环 |
| D-16 | 面试入口列表化 | TopBar 文案为 `面试` / `Interview`；`workspace` 无 `targetJobId` / `planId` 等上下文时展示面试规划列表，不再直接落到缺 JD 空态 | 与 product-scope D-23、module-job-workspace v1.19 一致；列表消费现有 `listTargetJobs`，不新增 `MockInterviewPlan` API 或独立多轮计划 |
| D-17 | 面试规划卡片信息取舍 | 列表卡片只保留继续规划所需的状态、更新时间、岗位、公司和地点；导入来源、目标语言和 `手动输入` 等字段不在卡片展示；进入动作使用主题强调色 CTA | 与 module-job-workspace v1.21 一致；避免把低价值技术元信息放大为主要阅读负担 |
| D-18 | 详情页归一化 | `workspace` 不再拥有详情视觉或上下文 route；Parse-derived “面试规划详情 / 面试上下文确认”母版是唯一详情入口 | 减少首次导入和回访的认知分叉，避免维护两套 JD/简历/轮次确认页面 |
| D-19 | 列表准入与无上下文路由 | `workspace` no-context 判定只看当前 route params，不继承 stale InterviewContext；列表请求必须携带 `analysisStatus=ready` 并防御性过滤空标题 / failed TargetJob | 与 backend-targetjob v2.2 失败解析不持久化合同一致；防止解析失败脏数据或详情页残留上下文污染一级 `面试` 入口 |
| D-20 | workspace route purity | `workspace` 即使收到 legacy query params 也清理 InterviewContext 并渲染列表；规划卡片导航 `parse`，`parse` / report owner 直接启动 practice | 回答“workspace 为什么有参数上下文”的实现偏差并防止回归 |

### 3.2 当前执行约束

- Resume Picker 由 Parse-derived 详情 owner 承接；workspace owner 不再维护 `ResumePickerModal` 或 `listResumes` gate。
- `createPracticeVoiceTurn` 已由 practice-voice-mvp / backend-practice voice extension 进入 generated client 与 fixture；voice surface 的完整 STT/LLM/TTS orchestration 仍归对应 voice owner，但正式前端不得继续把 operation 写作缺失。

## 4 设计约束

- 视觉与交互必须以 `ui-design/src/screen-workspace.jsx::WorkspacePlanList`、`ui-design/src/screens-p0-complete.jsx::ParseScreen`、`ui-design/src/screen-practice.jsx`、`ui-design/src/screens-p0-complete.jsx::ReportGeneratingScreen`、`ui-design/src/app.jsx`（route mapping / `INTERVIEW_CONTEXT_ROUTES` / `hideTopBar`）、`ui-design/src/primitives.jsx` 为唯一真理源进行源级复刻；不得二次设计。
- `WorkspacePlanList` 必须与 `screen-workspace.jsx` 当前结构一致；规划详情必须通过 `screens-p0-complete.jsx::ParseScreen` 的统一详情母版进入；workspace 不保留 `PlanSwitcherModal` / `ResumePickerModal`。
- workspace 中的公司轻情报只作为规划页嵌入卡片；workspace 不保留 URL 上下文参数，数据只来自当前 generated TargetJob list consumer。
- `PracticeScreen` 的 TopBar 工具区 / Left Panel / Main / Right Panel / 固定底部 CTA / 文本-语音 surface 切换 / 严格模拟开关必须与 `screen-practice.jsx` 当前结构一致；`practiceMode='strict'` 必须隐藏提示按钮、左侧实时观察、可调用经历、语音现场提示；`practiceGoal` 只允许当前核心闭环 goals，不改变辅助度显隐。
- `ReportGeneratingScreen` 的 4 步进度态、文案、节奏和 layout 必须与 `screens-p0-complete.jsx::ReportGeneratingScreen` 一致；轮询使用 generated `getFeedbackReport(reportId)`，不得在前端引入 AI provider / prompt registry / LLM key。
- route context 最小键必须按下表执行：

| Route | 本 spec owner | 最小上下文 | 缺失处理 |
|-------|---------------|------------|----------|
| `workspace` | 是 | 无 | 始终显示面试规划列表；忽略/清理 `targetJobId/planId/resumeId/autoStartPractice` 等 legacy params；列表为空显示友好空态并返回首页导入 JD |
| `practice` | 是 | `sessionId` 或可启动的 `planId`；推荐携带完整 InterviewContext | 缺 session/plan 时回 workspace 空态，不展示假问题 |
| `generating` | 是 | `sessionId + reportId` | 缺 `reportId` 显示生成态错误/返回 workspace |
| `report` | 否 | `sessionId + reportId` | 由 report owner 处理 |

- `PracticeDisplayContext = {mode, modality, practiceMode, practiceGoal, hintUsed, hintCount}` 只在 owner route/handoff context 中传递；`completePracticeSession` body 只发送 B2 `clientCompletedAt`，报告 owner 只展示练习方式和提示记录，不得推导通过率。
- 隐私红线：raw audio / TTS audio / transcript 明文 / LLM prompt-response 明文 / JD 原文 / 简历正文不得进入 console.log / URL query / localStorage / telemetry payload；fixture transport 不得在日志中泄漏。
- 暗色 / customAccent / 主题切换必须在 owner 三屏和 workspace 公司轻情报摘要卡片中通过 root `data-theme/data-mode/data-custom-accent` 生效。
- I18n 必须支持 zh / en；新增 `workspace.*` / `practice.*` / `generating.*` 命名空间；report 和公司轻情报扩展文案归对应 owner。
- Pixel parity gate 必须在 desktop (1440×900) + mobile (390×844) 两个 viewport 下断言 owner 三屏的 DOM 锚点 / computed style / bounding box / 截图差异；workspace 公司轻情报摘要卡片随 workspace gate 覆盖。
- Mobile 响应式：workspace 主左右列折叠；practice 三栏折叠为单列 + 底部 sheet；generating 居中进度态不溢出视口。
- `data-testid` 遵循 D1/D2 命名，使用 `workspace-*` / `practice-*` / `generating-*` 前缀；report 和公司轻情报扩展前缀归对应 owner。
- Current-scope negative gate 必须确认非当前 route/module 名称不作为 live route、TopBar 项、正向 testid、正向 scenario 或用户可见入口出现。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| workspace list / practice / generating UI | `frontend-workspace-and-practice`（本 spec） | 面试规划列表、PracticeSession 消费、source parity、visual parity、i18n、a11y、responsive；workspace 不拥有详情或启动副作用 |
| Report Dashboard UI | `frontend-report-dashboard` / `backend-review` | `ReportScreen`、报告详情、复练当前轮、进入下一轮、report 空态/失败态 |
| 公司轻情报摘要 | `frontend-workspace-and-practice` | workspace 内嵌摘要卡片；消费 TargetJob 摘要字段，不拥有独立刷新 API |
| App shell / routes / auth / runtime / theme | `frontend-shell` | TopBar、NO_CHROME_ROUTES、requestAuth、generated client bootstrap、mock transport、display preferences |
| Home / Parse / Unified Plan Detail | `frontend-home-job-picks-and-parse` | JD 导入、Parse loading、统一面试规划详情母版；workspace 列表卡片进入该母版 |
| TargetJobs backend | `backend-targetjob` | `listTargetJobs/getTargetJob/updateTargetJob/importTargetJob` handler/store/event |
| Practice backend | `backend-practice` | 6 个 Practice operation handler/service/store、state machine、AssistantAction、outbox、idempotency |
| Voice orchestration | `practice-voice-mvp` + `backend-practice` voice extension | `createPracticeVoiceTurn` contract/handler、STT/LLM/TTS、barge-in、committed-context |
| Report generation data | `backend-review` | `feedback_reports`、question assessments、readiness、report job result |
| Resume data | [`backend-resume`](../backend-resume/spec.md) | 扁平 Resume list / detail（D-20，无 version）；workspace 消费绑定 resume（`resumeId`）只读字段，active picker 消费 `listResumes` |
| OpenAPI / fixtures / codegen | `openapi-v1-contract` + `mock-contract-suite` | `openapi/openapi.yaml`、fixtures、generated Go/TS artifacts、fixture-backed mock transport |

### 5.1 Operation Matrix

| operationId | Fixture | Frontend consumer | Backend handler | Persistence | AI dependency | Scenario / status |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `listTargetJobs` | `openapi/fixtures/TargetJobs/listTargetJobs.json` (`default`, `prototype-baseline`) | `WorkspaceScreen` pure plan list only | `backend/internal/targetjob` implemented | `target_jobs` | none in frontend | `001-workspace-and-interview-context` |
| `getTargetJob` | `openapi/fixtures/TargetJobs/getTargetJob.json` (`default`, `prototype-baseline`) | Parse unified detail JD / requirements / source context | `backend/internal/targetjob` implemented | `target_jobs`, requirements/sources | none in frontend | `frontend-home-job-picks-and-parse/001` + parse tests |
| `getResume` | `openapi/fixtures/Resumes/getResume.json` (`default`) | Parse / resume owners only | backend-resume real handler | resume assets | none | external owner gates |
| `listResumes` | `openapi/fixtures/Resumes/listResumes.json` (`default`) | Parse resume picker list / resume workshop | backend-resume real handler | resume assets | none | external owner gates |
| `createPracticePlan` | `openapi/fixtures/PracticePlans/createPracticePlan.json` (`default`, `missing-resume`) | Parse detail start; report-derived retry / next round paths | backend-practice real handler | `practice_plans` | backend-only first-question prep | parse/report focused gates + `frontendOwners.realApiMode.test.ts` |
| `getPracticePlan` | `openapi/fixtures/PracticePlans/getPracticePlan.json` (`default`) | Parse/report handoff verifies existing plan matches target/resume context | backend-practice real handler | `practice_plans` | none | parse/report focused gates + real-mode gate |
| `startPracticeSession` | `openapi/fixtures/PracticeSessions/startPracticeSession.json` (`default`) | Parse detail start + report-derived replay/next-round | backend-practice real handler | `practice_sessions`, first turn | backend-only `practice.session.first_question` | parse/report focused gates + real-mode gate |
| `listPracticeSessions` | `openapi/fixtures/PracticeSessions/listPracticeSessions.json` (`default`) | Workspace session records handoff owner | backend-practice real handler | `practice_sessions` | none | workspace records + real-mode gate |
| `getPracticeSession` | `openapi/fixtures/PracticeSessions/getPracticeSession.json` (`default`, `prototype-baseline`, `missing-session`) | Practice refresh / recovery | backend-practice real handler | `practice_sessions`, turns/events | none in frontend | `002` + real-mode gate |
| `appendSessionEvent` | `openapi/fixtures/PracticeSessions/appendSessionEvent.json` (`default`) | Practice answer/hint/skip/pause/resume | backend-practice real handler | `practice_session_events`, `practice_turns` | backend-only follow-up/hint | `002` + real-mode gate；仍不带 Idempotency-Key |
| `completePracticeSession` | `openapi/fixtures/PracticeSessions/completePracticeSession.json` (`default`) | Practice finish CTA | backend-practice real handler | session status + outbox | none in frontend | `002` + real-mode gate |
| `getFeedbackReport` | `openapi/fixtures/Reports/getFeedbackReport.json` (`default`, `report-generating`, `prototype-baseline`) | Generating poll by `reportId` only；report owner consumes dashboard | backend-review real handler | `feedback_reports` + job result | backend-review only | report dashboard + real-mode gate |
| `createPracticeVoiceTurn` | `openapi/fixtures/PracticeSessions/createPracticeVoiceTurn.json` | Voice surface turn submission | practice-voice/backend-practice real handler | voice session events | STT/LLM/TTS backend-only | practice-voice owner + real-mode gate |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | owner route 专属 Screen 接管 | `frontend-shell` D1 已交付，owner route 当前由正式 screen 或外部 owner screen 接管 | 进入 `workspace` / `practice` / `generating` | `workspace` / `practice` 渲染正式 Screen；`practice/generating` 隐藏 chrome；`report` 不由本 spec 实现 | 001 / 002 / frontend-report-dashboard |
| C-2 | Workspace 渲染 + 空态 | 用户从一级 `面试` 进入，或 legacy URL 带有 `targetJobId/planId/resumeId` | 进入 `workspace` | 始终渲染面试规划列表且清理 stale InterviewContext；不调用 `getTargetJob`；不显示 `parse-error` / “缺少目标岗位 ID”；点击规划进入 `parse` 统一面试规划详情；不展示假数据 | 001 |
| C-2a | 面试规划列表卡片化与简化 | `listTargetJobs` 返回至少一条 ready 规划，并可能混入历史失败/空标题脏数据 | 进入无上下文 `workspace` | 列表请求带 `analysisStatus=ready`；列表项以响应式卡片呈现，卡片拥有背景、边框、轻阴影、body/footer 分区和主题强调色操作按钮；desktop 多列，mobile 单列，不出现无样式文本列；卡片不展示来源类型、目标语言或 `手动输入` 等低价值导入元信息；failed / blank-title TargetJob 不渲染 | 001 |
| C-3 | Workspace 交互闭环 | 已渲染 workspace 列表 | 用户点击 `进入规划` | 列表进入 `parse` 统一详情母版，携带真实 `targetJobId` 和可选真实 `currentPracticePlanId/resumeId`；不伪造 `jobId` / `jdId` / plan / resume / report id；更换简历、保存、立即面试和 session 创建由 parse/report/practice owner 执行 | 001 |
| C-4 | Practice 文本 happy path | 用户进入 `practice?mode=text&modality=text&practiceMode=assisted`，session=`running` | 用户输入回答、请求提示/跳过/暂停/恢复、提交事件、结束 | TextSurface 源级复刻；操作通过 `appendSessionEvent({clientEventId,kind,payload})`；AssistantAction 驱动下一题/追问/完成；结束调用 `completePracticeSession` 后进入 `generating?sessionId&reportId` | 002 |
| C-5 | Practice 语音 surface + core-goal 显隐 | 用户进入 `practice?mode=voice&modality=voice&practiceMode=strict`，以及 `practiceGoal=baseline/retry_current_round/next_round` 分别组合 assisted/strict | 用户进行语音回答或切换形式 | VoiceSurface 源级复刻；strict 隐藏提示、实时观察、可调用经历和现场提示；practiceGoal 不改变辅助度显隐；不直连 STT/TTS provider；voice turn flow 由 practice-voice owner gate 验证 | practice-voice-mvp/001 |
| C-6 | Generating 轮询 + report handoff | Practice 已 `completePracticeSession` 收到 `ReportWithJob{reportId,job}` | 用户在 generating 屏等待 | 4 步进度态与 `ReportGeneratingScreen` 一致；`queued/running` 保持等待，`succeeded` 导航 `report?sessionId&reportId`，`failed` 显示错误/重试/返回 workspace；不渲染 Report Dashboard | frontend-report-dashboard / backend-review |
| C-7 | Downstream handoff 参数 | workspace 公司轻情报卡片存在；generating 成功；records row 由 workspace records owner gate 接管 | 用户点击情报按钮 / 生成完成 / 查看记录占位 | 情报按钮保持在 `workspace` 并携带 safe `targetJobId/jdId`；generating 对 `report` handoff 携带 `sessionId/reportId`；plan 001 的 workspace records row 保持 disabled placeholder，不伪造 `sessionId/reportId`；报告目标屏渲染归外部 owner | 001 / frontend-report-dashboard |
| C-8 | UI source structure parity | C-1~C-7 通过 | Vitest+jsdom 加载 owner Screen | DOM 锚点、控件类型、icon、aria、keyboard、menu/modal 层级可追溯到 `screen-workspace.jsx` / `screen-practice.jsx` / `ReportGeneratingScreen` / `primitives.jsx` | 001 / 002 / external owner gates |
| C-9 | UI visual geometry parity | C-8 通过 | Playwright desktop + mobile 加载 owner 三屏 | 关键区块不重叠且 stays in viewport；theme/dark/customAccent 可见；workspace/practice/generating mobile 布局符合原型 | 001 / 002 / external owner gates |
| C-10 | UI current-scope negative search | C-8 + C-9 通过 | lint/grep gate 扫描 active runtime、positive tests、README、scenario | 非当前 route/module 不作为 live route、TopBar 项、正向 testid、正向 scenario 或用户入口出现；负向断言/禁止清单命中被分类允许 | 001 / 002 / product-scope gate |
| C-11 | BDD 主流程 + 关键分支 | owner route + parity gate 已就绪 | 创建并执行 E2E 场景 | 覆盖 workspace 渲染/切换/更换简历决策、未登录立即面试恢复、practice 文本、practice 语音 surface、strict 显隐、generating report handoff、非当前入口负向 | 001 / 002 / external owner gates |
| C-12 | Privacy 红线 | 用户完成 workspace→practice→generating 流程（文本 + 语音 surface 各一） | 检查 URL/localStorage/log/telemetry/fixture transport | raw audio / TTS audio / transcript 明文 / LLM prompt-response 明文 / JD 原文 / 简历正文不泄漏；只允许 IDs、状态、摘要和必要 route context | 001 / 002 / external owner gates |
| C-13 | 详情页归一化回归 | `parse?targetJobId=...` 可加载 TargetJob，`workspace?targetJobId=...` 是 legacy URL | 分别进入 parse 详情和 workspace | parse 共享统一详情 DOM/文案/布局；workspace canonicalize 为 `/workspace` 并仍为列表，不出现独立 `workspace-header` / `workspace-launcher` / `workspace-jd-card` 全页确认锚点，也不执行 `autoStartPractice` | 001 + frontend-home-job-picks-and-parse 001 |

## 7 关联计划

当前已完成 owner plan：

- `001-workspace-and-interview-context` — workspace 纯列表接管 + `listTargetJobs(analysisStatus=ready)` 消费 + 失败/空标题准入防线 + stale context 清理 + workspace BDD。
- `002-practice-text-event-loop` — PracticeScreen 文本 surface + `getPracticeSession/appendSessionEvent/completePracticeSession` 消费 + assisted/strict 辅助度策略 + current core-loop goals 显隐回归 + generating 入口。

语音 turn、generating/report dashboard 与 report-derived practice actions 由 `practice-voice-mvp`、`frontend-report-dashboard`、`backend-review` 和 `backend-practice` owner gate 承接；本 subspec 不保存 sibling plan 占位。

## 8 关联文档

- 上游 spec：[`engineering-roadmap`](../engineering-roadmap/spec.md) §5.2、[`product-scope`](../product-scope/spec.md) §5.2-§5.3、[`frontend-shell`](../frontend-shell/spec.md)、[`frontend-home-job-picks-and-parse`](../frontend-home-job-picks-and-parse/spec.md)、[`backend-practice`](../backend-practice/spec.md)、[`practice-voice-mvp`](../practice-voice-mvp/spec.md)、[`backend-targetjob`](../backend-targetjob/spec.md)、[`backend-auth`](../backend-auth/spec.md)、[`openapi-v1-contract`](../openapi-v1-contract/spec.md)、[`event-and-outbox-contract`](../event-and-outbox-contract/spec.md)、[`db-migrations-baseline`](../db-migrations-baseline/spec.md)、[`shared-conventions-codified`](../shared-conventions-codified/spec.md)、[`mock-contract-suite`](../mock-contract-suite/spec.md)
- UI 真理源：`ui-design/src/screen-workspace.jsx`、`ui-design/src/screen-practice.jsx`、`ui-design/src/screens-p0-complete.jsx::ReportGeneratingScreen`、`ui-design/src/app.jsx`（route mapping / `INTERVIEW_CONTEXT_ROUTES` / `hideTopBar`）、`ui-design/src/primitives.jsx`、[`docs/ui-design/module-job-workspace.md`](../../ui-design/module-job-workspace.md)、[`docs/ui-design/module-practice-review.md`](../../ui-design/module-practice-review.md)、[`docs/ui-design/module-map.md`](../../ui-design/module-map.md)、[`docs/ui-design/INDEX.md`](../../ui-design/INDEX.md)
- 当前正式前端入口：`frontend/src/app/{routes.ts,App.tsx,screens/PlaceholderScreen.tsx}`、`frontend/src/api/{generated/client.ts,mockTransport.ts}`、`frontend/src/app/runtime/AppRuntimeProvider.tsx`、`frontend/src/app/auth/pendingAction.ts`、`frontend/src/app/i18n/locales/{zh,en}.ts`、`frontend/src/app/theme/`、`frontend/src/app/scenarios/p0-005-app-shell-visual-system-smoke.test.tsx`、`frontend/tests/pixel-parity/`
- Fixture：`openapi/fixtures/TargetJobs/`（listTargetJobs / getTargetJob）、`openapi/fixtures/Resumes/`（getResume / registerResume）、`openapi/fixtures/PracticePlans/`（createPracticePlan / getPracticePlan）、`openapi/fixtures/PracticeSessions/`（startPracticeSession / getPracticeSession / appendSessionEvent / completePracticeSession）、`openapi/fixtures/Reports/getFeedbackReport.json`
- 治理 / 流程：[`AGENTS.md`](../../../AGENTS.md)、[`docs/development.md`](../../development.md) §2、[`docs/spec/README.md`](../README.md)、[`docs/spec/TEMPLATES.md`](../TEMPLATES.md)、[`test/scenarios/README.md`](../../../test/scenarios/README.md)
- 修订记录：[history.md](./history.md)

## 9 修订记录

| 版本 | 日期 | 说明 |
|------|------|------|
| 1.15 | 2026-07-09 | 固化 workspace route purity：`workspace` 是纯列表页，不承接详情参数上下文、不执行 `autoStartPractice`；规划详情和启动副作用由 parse/report/practice owner 承接。 |
| 1.14 | 2026-07-09 | 固化 workspace 面试列表准入：no-context 只看 route params，列表请求 `analysisStatus=ready` 并过滤 failed / 空标题 TargetJob，防止解析失败脏数据进入面试列表。 |
| 1.12 | 2026-07-08 | 固化面试规划列表卡片的信息取舍：移除来源/语言/手动输入等低价值元信息，进入规划 CTA 使用主题强调色并保持卡片/page 层次。 |
| 1.13 | 2026-07-09 | 过渡性归一 workspace 详情到 Parse-derived 母版；该过渡合同已由 1.15 的 workspace 纯列表语义取代。 |
| 1.11 | 2026-07-08 | 固化无上下文面试规划列表的卡片视觉合同，防止退化成无容器文本列。 |
| 1.10 | 2026-07-08 | 将 `workspace` 一级入口拆为面试规划列表 landing 与面试规划详情；同步 TopBar `面试` 命名和无上下文友好入口。 |
| 1.9 | 2026-07-07 | 将关联计划章节收敛为当前 completed 001/002 owner，并把 voice/report/generating 边界改为当前外部 owner gate。 |
| 1.8 | 2026-07-07 | 将 workspace 会话区域统一表述为 records，避免 active spec 使用过期口径描述当前记录占位和 handoff。 |
