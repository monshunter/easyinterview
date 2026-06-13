# Frontend Report Dashboard Spec

> **版本**: 1.3
> **状态**: active
> **更新日期**: 2026-06-13

> **2026-06-12 product-scope v2.1 D-19 对齐**：报告 CTA 单点收敛——报告页只保留 Header 一对 CTA（`复练当前轮` / `进入下一轮`）；复练计划详情 tab（next）只承载路径说明与复练清单，不重复 CTA 按钮；题目回顾的 `加入本轮复练` 改为本地标记动作（per-question toggle，不直接开启 session）。见 §3.1 D-19 与 §10。

## 1 背景与目标

`frontend-report-dashboard` 是 [engineering-roadmap §5.2](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) `Report Dashboard` workstream 的前端业务 subspec，承接 [frontend-shell](../frontend-shell/spec.md) 已交付的 App 壳、TopBar、route normalization、`requestAuth(pendingAction)`、fixture-backed generated client、UI parity gate；承接 [frontend-workspace-and-practice](../frontend-workspace-and-practice/spec.md) 的 `practice → generating` handoff（plan 002 已交付 `completePracticeSession` 真实调用 + 13 字段 generating route params + ReportWithJob 缓存）；同时作为 [backend-review](../backend-review/spec.md) `getFeedbackReport` / `listTargetJobReports` schema 的前端 consumer。

本 subspec 的终稿范围收敛为两条当前 owner 路由：

- `generating`：`completePracticeSession` 后的报告生成过渡态。承接 frontend-workspace-and-practice 已经设置好的 generating route params + reportId，轮询 `getFeedbackReport(reportId)` 直到 `status='ready'`（→ nav `report`）或 `status='failed'`（→ nav `report?reportStatus=failed`）。
- `report`：完整的证据化报告 dashboard。源级复刻 `ui-design/src/screen-report.jsx::ReportScreen` 的三态：`ReportDashboard`（正常报告：Header + ContextStrip + 4 个 Summary Cards + 5 个 Detail Tabs + 维度卡片行 + 优先级 + 复练重点 + 题目回顾概览 + 风险亮点）、`ReportFailureState`（reportStatus='failed'）、`ReportMissingSessionState`（缺 sessionId）。`复练当前轮` / `进入下一轮` CTA 触发 `nav("workspace", {..., autoStartPractice:'1'})`，由 workspace owner 创建全新 practice session 后再进入 `practice`。

`workspace` / `practice` / `company_intel` / `debrief` 不在本 subspec 范围。`workspace` / `practice` 由 [frontend-workspace-and-practice](../frontend-workspace-and-practice/spec.md) owner 承接；`company_intel` 与 `debrief` 由 external company-intel / future `frontend-debrief` owner 承接。本 subspec 只在 ReportScreen 的复练 CTA 中发起 workspace auto-start handoff；不实现 practice 任何 UI，也不直接创建或复用 practice session。

本 subspec 通过 generated client + fixture-backed transport 消费已经存在的 Reports OpenAPI 契约；截至 2026-05-23，backend-review/001 已落地 `getFeedbackReport` / `listTargetJobReports` 真实 handler，frontend plan 001 的 completed UI variants 必须配套 `VITE_EI_API_MODE=real` generated-client gate，证明 production bootstrap 使用真实 backend base URL、cookie credentials、无 fixture `Prefer` header，并保持 dashboard-only D-7 不消费列表 UI。任何新增或缺失 operation 先回到 [B2](../openapi-v1-contract/spec.md) / [backend-review](../backend-review/spec.md) 修订，不能在前端手写 ad hoc fetch 或复制 `ui-design` mock data。

## 2 范围

### 2.1 In Scope

- `generating` 屏（`route=generating`，chrome 隐藏）：
  - 源级复刻 `ui-design/src/screens-p0-complete.jsx::ReportGeneratingScreen`（lines 269-399）：页头（标题 + 副文案 "Reading every turn. Evidence first."）+ 进度条 + 5 阶段列表（转写并对齐对话 / 逐题抽取证据 / 按 rubric 评分 / 聚类题目回顾信号 / 生成建议；每个 done/active/pending 状态圆圈）+ 实时观察流（live evidence snippets fade-in）+ 底部提示（P95 SLA "<12s target" + 「通知我」按钮 UI-only）。
  - 轮询 `getFeedbackReport(reportId)` hook：指数退避（初始 1.5s，每次 × 1.5，上限 8s）+ max attempts（30 次约 4 分钟）+ visibility 暂停 / 恢复轮询；status='queued' / 'generating' 保持等待；status='ready' → `nav("report", { sessionId, reportId, ...passThroughContext })`；status='failed' → `nav("report", { sessionId, reportId, reportStatus:'failed', errorCode, ...passThroughContext })`；max attempts 达到 → ErrorState 「报告生成超时，请重试」+ retry / 返回 workspace CTA。
- `report` 屏（`route=report`，保留默认 App chrome / TopBar；不进入一级导航）：
  - 源级复刻 `ui-design/src/screen-report.jsx::ReportScreen`（lines 1-516）三态：
    - `ReportDashboard`（params.reportStatus !== 'failed' && sessionId 存在；通过 `getFeedbackReport(reportId)` 拉取数据后渲染）：返回按钮 + Header + `ReportContextStrip`（sessionId / targetJob / round / resume / modality / practiceMode / hints）+ 4 个 Summary Cards（准备度 / 维度 / 题目 / 下一步）+ `ReportDetailSurface` 5 个 tab（readiness / dimensions / questions / evidence / next，默认 `questions`）+ 维度卡片行 + 优先级 + 复练重点 + 题目回顾概览（5 题快速状态）+ 风险 & 亮点 + 复练 CTA（路径 A / 路径 B）。
    - `ReportFailureState`（params.reportStatus === 'failed'）：卡片 + 文案「报告生成失败」+ errorCode 显示 + CTA「重新生成」（nav `generating`）/ 「返回 workspace」。
    - `ReportMissingSessionState`（缺 sessionId）：卡片 + 文案「会话已结束或不存在」+ CTA「返回 workspace」。
  - 5 个 detail tab 源级复刻：
    - `readiness`（lines 335-357）：拨号盘 + 二级详情（JD 对齐、证据密度、下一档门槛）。
    - `dimensions`（lines 360-382）：二级维度卡片网格（覆盖 D-4 ReadinessTier 四档对应的维度详情）。
    - `questions`（lines 385-442）：题目列表侧栏 + 当前题回答分析（有效点、缺口、建议框架、证据片段、下次追问）。
    - `evidence`（lines 445-467）：风险证据详情 + 可复用亮点证据。
    - `next`（lines 470-514）：路径 A（复练当前轮）vs 路径 B（进入下一轮）对比展示 + 复练清单；**不含 CTA 按钮**（D-19 单点收敛）：路径卡片以 footer 文案「开练入口在页面顶部：复练当前轮 / 进入下一轮」引导用户使用 Header CTA，next tab 自身不渲染 `report-next-cta-a` / `report-next-cta-b`。
    - 题目回顾 `加入本轮复练`（D-19）：是 per-question 本地标记动作（`report-questions-add-to-replay` toggle，文案 `加入本轮复练` ↔ `已加入本轮复练`），只改本地 state，不 `nav`、不开启 session、不调 API；标记仅在报告内表达复练意图，实际开练仍由 Header `复练当前轮` CTA 承载。
  - 复练 CTA 行为（仅 Header 一对 CTA，D-19）：
    - 路径 A `goReplay()` → `nav("workspace", { sourceSessionId: sessionId, replayItems: retryFocusTurnIds, evidenceGaps: focusGaps, planId, targetJobId, jdId, resumeVersionId, roundId, mode:'text', modality:'text', practiceMode: lastPracticeMode, practiceGoal:'retry_current_round', autoStartPractice:'1' })`；workspace owner 随后调用 practice-plan/session start 契约并 `nav("practice", { sessionId:newSessionId, ...sameContext })`；未登录走 `useRequestAuth({type:'replay_practice', route:'workspace', params:{...sameParams, autoStartPractice:'1'}})`。
    - 路径 B `goNextRound()` → `nav("workspace", { nextRoundId, roundName, roundId: nextRoundId, planId, targetJobId, jdId, resumeVersionId, mode:'text', modality:'text', practiceMode: lastPracticeMode, practiceGoal:'next_round', autoStartPractice:'1' })`；workspace owner 创建新 session 后进入 `practice`；未登录走 `useRequestAuth({type:'replay_practice', route:'workspace', params:{...sameParams, autoStartPractice:'1'}})`。
- 跨路由共享：
  - `InterviewContext` 在本 subspec owner route 内传递 `{planId, targetJobId, jdId, resumeVersionId, roundId, sessionId, reportId, mode, modality, practiceMode, practiceGoal, hintUsed, hintCount}` 与 frontend-workspace-and-practice 一致 13 字段（与 `buildPracticeHandoffParams` 输出完全一致；`roundName` 不在该 13 字段内，由 ContextStrip 在本 spec owner route 内通过 `InterviewContext.roundId` + i18n 本地推导显示）；本 spec 只 read（不 mutate session 字段），但在 generating 屏成功 nav report 时把 reportId 留在 context；在复练 CTA 触发后由 workspace / practice owner reducer 接管。
  - 未登录用户点击复练 CTA 时通过 `useRequestAuth({type:'replay_practice', route:'workspace', params:{...InterviewContext, autoStartPractice:'1'}})` 触发鉴权；登录后 pendingAction 回到 `workspace`，由 workspace auto-start 机制创建新 session 后进入 `practice`。
  - 隐私：route params 仅传 13 个 handoff params（7 个稳定 owner IDs + 6 个 display knobs，与 `buildPracticeHandoffParams` 输出一致）；不传 raw answer/question/hint/prompt/provenance；与 frontend-workspace-and-practice plan 002 隐私红线一致。
- 契约消费形态：
  - `getFeedbackReport`：generating 轮询 + report dashboard 详情；按 OpenAPI `GET /reports/{reportId}` 仅写 path param `reportId`；不写 Idempotency-Key。
  - `listTargetJobReports`：本 spec plan 001 不消费（dashboard-only D-7，无报告列表导航入口）；schema parity 必须保证 future plan 002+ 接入。

### 2.2 Out of Scope

- `WorkspaceScreen` / Interview Launcher / Resume Picker / Plan Switcher：由 [frontend-workspace-and-practice](../frontend-workspace-and-practice/spec.md) 承接。
- `PracticeScreen` 任何 UI / 状态机消费 / 文本 surface / voice surface / 完成动作：由 [frontend-workspace-and-practice](../frontend-workspace-and-practice/spec.md) 承接；本 spec 只在复练 CTA 中交给 workspace auto-start，等待该 owner 创建新 session 后进入 practice。
- `CompanyIntelScreen` / `getCompanyIntel`：external company-intel owner 承接。
- `DebriefScreen` / `getDebrief`：future `frontend-debrief` owner 承接。
- Home / Parse / JD Match shell 与 JD 导入解析：由 [frontend-home-job-picks-and-parse](../frontend-home-job-picks-and-parse/spec.md) 承接。
- Auth / TopBar / Sidebar / Theme / I18n bootstrap / requestAuth 接线：由 [frontend-shell](../frontend-shell/spec.md) 承接。
- Reports / FeedbackReport 真实 backend handler / service / store / event 发射：由 [backend-review](../backend-review/spec.md) 承接。
- AI provider / prompt registry / 模型路由：由 [ai-provider-and-model-routing](../ai-provider-and-model-routing/spec.md) / [prompt-rubric-registry](../prompt-rubric-registry/spec.md) / backend-review 承接。
- 报告列表 UI（report timeline / multiple report dashboards 历史浏览）：product-scope D-7 dashboard-only 边界，不在本 spec 范围；plan 002 future 可考虑增加但需先修订 product-scope。
- 报告导出 / 分享：由 future 隐私 spec / 平台 owner 承接。
- 报告评分质量反馈（用户对报告打分）：由 future [prompt-rubric-registry/005-grayscale-and-quality-feedback](../prompt-rubric-registry/spec.md) 承接。
- 不新增或恢复弃用模块 / 路由 / 术语作为 live UI：报告时间线、刊物式报告页、独立错题本、Drill builder、Growth center、5 档 readiness、多形态 report、独立 `report` 一级导航入口、独立 `mistakes` route。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | Route owner 范围 | 本 subspec 只接管 `generating / report`；`workspace / practice / company_intel / debrief` 是外部 owner | 消除与 frontend-workspace-and-practice / external owner 的边界冲突 |
| D-2 | UI 真理源 | `ui-design/src/screen-report.jsx` + `ui-design/src/screens-p0-complete.jsx::ReportGeneratingScreen` + `docs/ui-design/report-dashboard.md` + `ui-design/src/app.jsx`（route mapping / `INTERVIEW_CONTEXT_ROUTES` / `hideTopBar`）+ `ui-design/src/primitives.jsx` 为唯一真理源进行源级复刻；不得二次设计 | 与 frontend-workspace-and-practice D-2 一致；保护 ui-design parity gate |
| D-3 | GeneratingScreen 轮询节奏 | 指数退避（初始 1.5s × 1.5 上限 8s）+ max attempts=30（约 4 分钟）+ visibility/focus 暂停-恢复 + status='ready' 自动 nav report / status='failed' nav report?reportStatus=failed + max attempts 达到显示超时态 | 与 backend-review worker P95 latency observation + B3 outbox `report.generated` 异步时延一致；P95 < 12s 时 ~7 次轮询命中 ready |
| D-4 | 状态分支 | GeneratingScreen 渲染 status ∈ {`queued`,`generating`} 等待态 + 5 阶段进度动画；status='ready' nav report；status='failed' nav report?reportStatus=failed；ReportScreen 渲染 ReportDashboard / ReportFailureState / ReportMissingSessionState 三态；status='ready' 必须含完整 FeedbackReport 字段 | 与 B2 `FeedbackReport` schema + backend-review D-7 状态机一致 |
| D-5 | 复练 CTA payload | 路径 A nav workspace auto-start payload：{sourceSessionId, replayItems:retryFocusTurnIds, evidenceGaps, planId, targetJobId, jdId, resumeVersionId, roundId, mode:'text', modality:'text', practiceMode:lastPracticeMode, practiceGoal:'retry_current_round', autoStartPractice:'1'}；路径 B nav workspace auto-start payload：{nextRoundId, roundName, roundId:nextRoundId, planId, targetJobId, jdId, resumeVersionId, mode:'text', modality:'text', practiceMode:lastPracticeMode, practiceGoal:'next_round', autoStartPractice:'1'}；其中 `lastPracticeMode` 来自 InterviewContext.practiceMode；workspace owner 必须创建 fresh session 后进入 practice，不能复用 source session；未登录走 `useRequestAuth({type:'replay_practice', route:'workspace', ...})`（**deliberate divergence from UI source**：UI 真理源 `ui-design/src/screen-report.jsx:114` 使用通用 `requestAuth({type:'create_session', ...})`，生产前端按 product-scope §4.1 "复练优先" + backend-practice D-4 plan goal 四值的要求区分 `replay_practice` 与 `start_practice`，以便 pendingAction 恢复时正确派发到 `goReplay`/`goNextRound` 不同 payload；此分叉只影响 auth/routing 契约，不影响视觉源级复刻 gate） | 与 backend-practice D-4 plan goal 四值 + frontend-workspace-and-practice D-9 立即面试契约 + product-scope §4.1 "复练优先" 一致 |
| D-6 | 报告失败状态语义 | ReportFailureState 渲染失败卡片 + errorCode 文案映射（按 B1 `AI_*` enum 各自文案）+ CTA「重新生成」（nav `generating`）+「返回 workspace」；不在 ReportDashboard 内 inline 渲染失败态；不暴露 raw provider error 给用户 | 与 backend-review D-8 graceful failed + B1 error_code 一致；用户能感知失败但不暴露内部细节 |
| D-7 | i18n 命名空间约定 | 新增 `report.*` 与 `generating.*` 命名空间；不复用 `workspace.*` 与 `practice.*`；外部 `workspace.reportReady` / `workspace.lastReport` / `workspace.gotoReport` 已存在的 key 保留不动（由 workspace owner 维护） | 命名空间独立避免与其他 owner 冲突 |
| D-8 | InterviewContext reducer 扩展边界 | 在 001 plan 已有 `InterviewContext` reducer 基础上**仅 read**；不新增 `SET_REPORT_ID` 或 `MERGE_REPORT_DISPLAY` reducer action（reportId 在 frontend-workspace-and-practice plan 002 buildPracticeHandoffParams 时已通过 route params 写入并由 InterviewContext.useEffect 同步）；本 spec 通过 route params + URL search params 读取 reportId / sessionId / reportStatus / errorCode | 不破坏 frontend-workspace-and-practice reducer 边界；不在多 owner 间双重 write context |
| D-9 | 5 个 detail tab DOM 锚点 | `report-detail-tab-{readiness,dimensions,questions,evidence,next}` 触发按钮 + `report-detail-panel-{readiness,dimensions,questions,evidence,next}` 内容容器 + 各 tab 内具体 testid（如 `report-readiness-dial / report-dimensions-grid / report-questions-list / report-questions-detail / report-evidence-card-* / report-next-path-{a,b}`） | DOM anchor 锁定让源级 parity test 可执行 |
| D-10 | 准备度 tier 4 档文案 | `not_ready` / `needs_practice` / `basically_ready` / `well_prepared` 各档 zh/en 文案与拨号盘 visual 段落颜色映射；与 ui-design 与 docs/ui-design/report-dashboard.md 一致；不引入 5 档 readiness 旧 numeric score 字面量 | 与 backend-review D-4 + product-scope §6.9 + ui-design `data.jsx` `readinessLabel` 一致 |
| D-11 | 维度卡片状态映射 | `strong` / `meets_bar` / `needs_work` 三态文案与色调；不引入旧 `acceptable` / `weak` / `developing` / `proficient` 字面量混淆（注：rubric score_levels 四档是 AI 内部评分 label，不暴露到 UI；用户可见的是 dimensions.state 三态） | 与 B1 `DimensionStatus` enum + backend-review D-4 ↔ rubric score_levels 不同层 |
| D-12 | retired 术语 | 旧 `reportLayout` / 5 档 readiness / 独立 mistakes 错题本 / Drill builder / Growth center / 报告时间线 / 多形态 report / 独立 `report` 一级导航 / 旧 reportLayout fixture variant 不得作为 live route / TopBar 项 / 正向 testid / 正向 scenario / 用户可见入口出现 | 与 product-scope D-6/D-7 + backend-review §4.5 一致 |
| D-13 | 隐私红线 | route params / URL search params / InterviewContext 不传 raw `answerText` / `questionText` / `hint` / `promptHash` / `modelId` raw value；仅传 13 个 handoff params（7 个稳定 owner IDs + 6 个 display knobs）；fixture transport spy 不泄漏 body；console.log / localStorage / telemetry 同款约束 | 与 frontend-workspace-and-practice plan 002 隐私红线 + product-scope §9.3 一致 |
| D-14 | backend 契约消费 | 只通过 B2 generated client 消费 OpenAPI operation；字段变化先回 B2/backend-review 修订；不在前端自造 endpoint 或复制 fixture JSON | 与 frontend-workspace-and-practice D-10 一致 |
| D-19 | 报告 CTA 单点收敛 | 报告页只保留 Header 一对 CTA（`复练当前轮` / `进入下一轮`）作为唯一开练入口；`next`（复练计划）tab 只承载路径 A/B 说明与复练清单，**不渲染** `report-next-cta-a` / `report-next-cta-b`，改以 footer 文案引导用户使用 Header CTA；题目回顾 `加入本轮复练`（`report-questions-add-to-replay`）是 per-question 本地标记 toggle（`加入本轮复练` ↔ `已加入本轮复练`），只改本地 state，不 `nav` / 不开 session / 不调 API；不引入任何二级开练按钮 | 对齐 product-scope v2.1 D-19；纯前端 UI 收敛，无 OpenAPI/backend 变更；`ui-design/src/screen-report.jsx` 当前真理源即此形态（Header `goReplay`/`goNextRound` + NextTab 无按钮 + QuestionsTab `toggleQueued` 本地 per-question state） |

### 3.2 待确认事项

- 「通知我」CTA 是否在 plan 001 实现真实邮件订阅 / 留作 UI-only：plan 001 默认 UI-only（disabled or local state only）；若需要真实邮件通知，先回 backend-review / platform owner 修订 contract。
- ReportScreen 加载失败（`getFeedbackReport` 网络 5xx 或 timeout）的恢复策略：plan 001 默认 retry + InlineError；若需要更复杂的 recovery（如本地缓存 + offline mode）由 plan 002+ 处理。

## 4 设计约束

- 视觉与交互必须以 `ui-design/src/screen-report.jsx`、`ui-design/src/screens-p0-complete.jsx::ReportGeneratingScreen`、`ui-design/src/app.jsx`（route mapping / `INTERVIEW_CONTEXT_ROUTES` / `hideTopBar`）、`ui-design/src/primitives.jsx` 为唯一真理源进行源级复刻；不得二次设计。
- `ReportDashboard` 的 Header / ContextStrip / 4 Summary Cards / 5 Detail Tabs / 维度卡片 / 优先级 / 复练重点 / 题目回顾 / 风险亮点必须与 `screen-report.jsx` 当前结构一致；`ReportStatButton` / `DimRow` / `StatCard` 等 primitives 直接复用 `ui-design/src/primitives.jsx`。
- `ReportGeneratingScreen` 的 5 阶段进度态、文案、节奏（每阶段 700-1200ms）、layout 必须与 `screens-p0-complete.jsx::ReportGeneratingScreen` 一致；实时观察流文案 fade-in 动画与原型一致；轮询使用 generated `getFeedbackReport(reportId)`，不得在前端引入 AI provider / prompt registry / LLM key。
- route context 最小键必须按下表执行：

| Route | 本 spec owner | 最小上下文 | 缺失处理 |
|-------|---------------|------------|----------|
| `generating` | 是 | `reportId`（必填）+ `sessionId`（推荐携带 + 全套 13 字段） | 缺 `reportId` 显示 ErrorState「报告 ID 缺失，返回 workspace」 |
| `report` | 是 | `sessionId + reportId`（必填）+ 全套 13 字段（推荐） | 缺 `sessionId` 显示 `ReportMissingSessionState`；缺 `reportId` 显示同上；`reportStatus='failed'` 显示 `ReportFailureState` |
| `workspace` | 否 | `targetJobId` | 由 frontend-workspace-and-practice 处理 |
| `practice` | 否 | `sessionId` 或 `planId` | 由 frontend-workspace-and-practice 处理（复练 CTA 先走 workspace auto-start，再由该 owner 进入 practice） |
| `company_intel` | 否 | `targetJobId` + `jdId` | 由 company-intel owner 处理 |

- 隐私红线：raw answer / question / hint / prompt-response 明文 / JD 原文 / 简历正文不得进入 console.log / URL query / localStorage / telemetry payload；fixture transport 不得在日志中泄漏。
- 暗色 / customAccent / 主题切换必须在 owner 两屏（generating / report）通过 root `data-theme/data-mode/data-custom-accent` 生效。
- I18n 必须支持 zh / en；新增 `report.*` / `generating.*` 命名空间；workspace/practice/companyIntel/debrief 文案归外部 owner。
- Pixel parity gate 必须在 desktop (1440×900) + mobile (390×844) 两个 viewport 下断言 owner 两屏的 DOM 锚点 / computed style / bounding box / 截图差异。
- Mobile 响应式：generating 居中进度态不溢出视口；report 主屏三列折叠为单列 + Detail Surface 切 collapsible Accordion + 复练 CTA sticky bottom。
- `data-testid` 遵循 D9 命名，使用 `generating-*` / `report-*` 前缀；workspace / practice / companyIntel / debrief 前缀归外部 owner。
- stale-contract negative gate 必须区分"禁止作为 live UI/runtime 正向入口"和"允许出现在负向断言/禁止清单/历史说明"。旧 route/module 名称不得作为 active route、TopBar 项、正向 testid、正向 scenario 或用户可见入口重新出现。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| generating / report UI | `frontend-report-dashboard`（本 spec） | 两屏 React 组件、轮询 hook、状态分支（dashboard/failure/missing）、复练 CTA、source parity、visual parity、i18n、a11y、responsive |
| Workspace / Practice UI | [`frontend-workspace-and-practice`](../frontend-workspace-and-practice/spec.md) | workspace 屏、practice 屏、generating handoff 入口；本 spec 在复练 CTA nav 回该 owner |
| App shell / routes / auth / runtime / theme | [`frontend-shell`](../frontend-shell/spec.md) | TopBar、NO_CHROME_ROUTES、requestAuth、generated client bootstrap、mock transport、display preferences |
| Home / Parse / JD Match | [`frontend-home-job-picks-and-parse`](../frontend-home-job-picks-and-parse/spec.md) | parse confirm 跳转 workspace |
| Company Intel UI | external company-intel owner | `CompanyIntelScreen` |
| Debrief UI | future `frontend-debrief` | `DebriefScreen` |
| Reports backend | [`backend-review`](../backend-review/spec.md) | `getFeedbackReport` / `listTargetJobReports` handler / service / store / inline review runner / report 生成 / 维度评估 / readiness / retry_focus / next_action |
| Practice backend | [`backend-practice`](../backend-practice/spec.md) | 6 Practice operation handler / state machine / outbox / complete handoff `ReportWithJob` 提供方 |
| OpenAPI / fixtures / codegen | [`openapi-v1-contract`](../openapi-v1-contract/spec.md) + [`mock-contract-suite`](../mock-contract-suite/spec.md) | `openapi/openapi.yaml`、fixtures `Reports/getFeedbackReport.json` / `listTargetJobReports.json`、generated Go/TS artifacts、fixture-backed mock transport |
| Resume data | [`backend-resume`](../backend-resume/spec.md) | 简历 binding 字段；本 spec 通过 generated `getResumeVersion(resumeVersionId)` 只读 displayName 用于 ContextStrip，缺失时回退显示 resumeVersionId |
| TargetJob data | [`backend-targetjob`](../backend-targetjob/spec.md) | `target_jobs` 行；本 spec 通过 generated `getTargetJob(targetJobId)` 只读 title/companyName 用于 ContextStrip，缺失时回退显示 targetJobId |

### 5.1 Operation Matrix

| operationId | Fixture | Frontend consumer | Backend handler | Persistence | AI dependency | Scenario / status |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `getFeedbackReport` | `openapi/fixtures/Reports/getFeedbackReport.json`（`default` = ready 完整字段 / `report-generating` = generating 占位 / `prototype-baseline` = 中文示例 / `report-failed` = failed） | GeneratingScreen 轮询 + ReportScreen 拉取 | backend-review real handler | `feedback_reports` + `question_assessments` read | none in frontend；report 内容生成由 backend-review 完成 | `001-report-screen-and-generating-handoff` + `frontendOwners.realApiMode.test.ts` |
| `getTargetJob` | `openapi/fixtures/TargetJobs/getTargetJob.json`（`default`） | ReportContextStrip 只读 target job title/companyName；失败时显示 targetJobId fallback，不阻塞报告正文 | backend-targetjob 既有 handler | `target_jobs` read | none | `E2E.P0.056` ContextStrip 子断言 |
| `getResumeVersion` | `openapi/fixtures/Resumes/getResumeVersion.json`（`default`） | ReportContextStrip 只读 resume version displayName；失败时显示 resumeVersionId fallback，不读取 raw resume text | backend-resume 既有 handler | `resume_versions` read | none | `E2E.P0.056` ContextStrip 子断言 + privacy negative |
| `listTargetJobReports` | `openapi/fixtures/Reports/listTargetJobReports.json`（`default` = 分页 + pageInfo / `empty` variant） | 本 plan 001 **不消费 UI**（dashboard-only D-7，无一级列表导航入口）；schema parity 和 `frontendOwners.realApiMode.test.ts` 证明 generated client 可指向真实 backend；Vitest 单元测试 + scenario verify.sh + scoped legacy grep 三层断言 `listTargetJobReports` 在 `frontend/src/app/screens/{report,generating}/` 0 调用 | backend-review real handler | `feedback_reports` cursor read | none | real-mode gate + 负向 UI 断言 |
| `completePracticeSession` | 由 frontend-workspace-and-practice 002 消费；本 spec 不调用 | — | — | — | — | 负向断言（在 generating / report 模块零调用） |
| `appendSessionEvent` | 由 frontend-workspace-and-practice 002 消费；本 spec 不调用 | — | — | — | — | 负向断言 |
| `getPracticeSession` / `startPracticeSession` / 其他 Practice operation | 由 frontend-workspace-and-practice 消费；本 spec 不调用 | — | — | — | — | 负向断言 |
| `getCompanyIntel` | N/A | 本 spec **不消费**；company-intel owner 承接 | external owner | — | — | 负向断言 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 两条 owner route 专属 Screen 接管 | frontend-shell D1 已交付；frontend-workspace-and-practice 已完成 generating handoff；owner route 当前由 PlaceholderScreen 占位 | 进入 `generating` 与 `report` | 两条 route 渲染正式 Screen；`generating` 隐藏 chrome；`report` 保留默认 App chrome / TopBar 且不进入一级导航；不展示 PlaceholderScreen；P0.056-P0.059 trigger 前置 real-mode generated-client gate | 001 |
| C-2 | GeneratingScreen 轮询 happy path | InterviewContext 携带 `reportId, sessionId, ...passThrough`；fixture `getFeedbackReport` 配置为 `report-generating` 几次轮询后切换到 `default`（ready） | 进入 `generating` | 渲染 5 阶段进度动画 + 实时观察流；轮询调用 `getFeedbackReport(reportId)` 多次（按指数退避节奏）；status='ready' 时自动 `nav("report", { sessionId, reportId, ...passThrough })`；轮询期间不重复 nav | 001 |
| C-3 | GeneratingScreen 失败处理 | fixture `getFeedbackReport=report-failed`（status='failed' + errorCode='AI_PROVIDER_TIMEOUT'） | 进入 `generating`，轮询命中 failed | 自动 `nav("report", { sessionId, reportId, reportStatus:'failed', errorCode:'AI_PROVIDER_TIMEOUT', ...passThrough })`；不展示 ReportDashboard | 001 |
| C-4 | GeneratingScreen 超时 | fixture 永久返回 `report-generating`（模拟 backend 卡住） | 进入 `generating`，轮询达到 max attempts | 渲染 ErrorState「报告生成超时，请重试」+ retry / 返回 workspace CTA；retry 重启轮询 | 001 |
| C-5 | ReportDashboard 渲染（ready） | InterviewContext 携带完整 13 字段；fixture `getFeedbackReport=default` 返回完整 FeedbackReport | 进入 `report` | 渲染 Header + ContextStrip + 4 Summary Cards + 5 Detail Tabs（默认 `questions`，readiness 通过显式切换覆盖）+ 维度卡片行 + 优先级 + 复练重点 + 题目回顾 + 风险亮点；准备度 tier 文案与色调按 D-10 映射；维度卡片状态按 D-11 三态映射 | 001 |
| C-6 | ReportFailureState | `report?reportStatus=failed&errorCode=AI_PROVIDER_TIMEOUT&sessionId=S&reportId=R&...` | 进入 `report` | 渲染 ReportFailureState 卡片 + errorCode 文案映射 + CTA「重新生成」（nav `generating?reportId&sessionId&...`）+ 「返回 workspace」 | 001 |
| C-7 | ReportMissingSessionState | `report?reportId=R`（缺 sessionId） | 进入 `report` | 渲染 ReportMissingSessionState 卡片 + CTA「返回 workspace」（nav workspace with targetJobId）；不调用 `getFeedbackReport` | 001 |
| C-8 | 5 detail tab 切换 | C-5 已渲染 ReportDashboard | 用户点击 tab 切换按钮 | 5 个 tab（readiness / dimensions / questions / evidence / next）panel 切换；每个 tab 内容源级复刻；testid `report-detail-tab-{key}` + `report-detail-panel-{key}` 命中；其他 panel 不渲染（或 display:none） | 001 |
| C-9 | 复练 CTA 路径 A（Header 唯一入口） | C-5 已渲染 ReportDashboard，准备度 = needs_practice，retry_focus_turn_ids 非空 | 用户点击 Header「复练当前轮」CTA | `nav("workspace", { sourceSessionId, replayItems: retryFocusTurnIds, evidenceGaps, planId, targetJobId, jdId, resumeVersionId, roundId, mode:'text', modality:'text', practiceMode:lastPracticeMode, practiceGoal:'retry_current_round', autoStartPractice:'1' })`；workspace owner 调用 `startPracticeSession` 后 `nav("practice", { sessionId:newSessionId, ... })`；未登录走 useRequestAuth 后恢复同一 workspace auto-start payload；CTA 仅在 Header 渲染（D-19） | 001 |
| C-10 | 复练 CTA 路径 B（Header 唯一入口） | C-5 已渲染 ReportDashboard，准备度 = basically_ready，next_action='next_round' | 用户点击 Header「进入下一轮」CTA | `nav("workspace", { nextRoundId, roundName, roundId:nextRoundId, planId, targetJobId, jdId, resumeVersionId, mode:'text', modality:'text', practiceMode:lastPracticeMode, practiceGoal:'next_round', autoStartPractice:'1' })`；workspace owner 调用 `startPracticeSession` 后进入 practice；未登录走 useRequestAuth 后恢复同一 workspace auto-start payload；CTA 仅在 Header 渲染（D-19） | 001 |
| C-16 | next tab 无重复 CTA（D-19） | C-5 已渲染 ReportDashboard | 切到 `next`（复练计划）tab | `report-detail-panel-next` 渲染路径 A/B 说明 + 复练清单 + footer「开练入口在页面顶部」引导文案；`report-next-cta-a` / `report-next-cta-b` testid 在 DOM **0 命中**；不存在任何二级开练按钮触发 `goReplay`/`goNextRound` | 001 |
| C-17 | 题目回顾本地标记（D-19） | C-5 已渲染 ReportDashboard，切到 `questions` tab | 用户点击 `report-questions-add-to-replay` | 仅 toggle 当前题目本地标记 state（文案 `加入本轮复练` ↔ `已加入本轮复练`），不触发 `nav`、不调用 `startPracticeSession`/任何 API、不改 URL/InterviewContext；切换不同题目各自独立标记；实际开练仍只由 Header CTA 承载 | 001 |
| C-11 | UI source structure parity | C-1~C-10 通过 | Vitest+jsdom 加载 owner Screen | DOM 锚点、控件类型、icon、aria、keyboard、menu/modal 层级可追溯到 `screen-report.jsx` / `ReportGeneratingScreen` / `primitives.jsx`；testid 命名一致 | 001 |
| C-12 | UI visual geometry parity | C-11 通过 | Playwright desktop + mobile 加载 owner 两屏 | 关键区块不重叠且 stays in viewport；theme/dark/customAccent 可见；generating mobile 居中不溢出；report mobile 三列折叠为单列 + Accordion + sticky CTA | 001 |
| C-13 | UI stale-contract negative search | C-11 + C-12 通过 | lint/grep gate 扫描 active runtime、positive tests、README、scenario | 旧 `reportLayout` / 5 档 readiness numeric / `mistakes` route / `drill_builder` testid / `growth_center` / 报告时间线 / 多形态 report 不作为 live route / TopBar / 正向 testid / 正向 scenario / 用户入口出现；负向断言/禁止清单命中被分类允许 | 001 |
| C-14 | BDD 主流程 + 关键分支 | 两条 owner route + parity gate 已就绪 | 创建并执行 E2E 场景 | 覆盖 generating 轮询 ready + failed + 超时、report dashboard 渲染、复练 CTA 路径 A/B、ReportFailureState、ReportMissingSessionState、旧口径负向 | 001 |
| C-15 | Privacy 红线 | 用户完成 generating + report 流程 | 检查 URL/localStorage/log/telemetry/fixture transport | raw answer/question/hint/prompt-response 明文 / JD 原文 / 简历正文不泄漏；只允许 13 个 handoff params（7 个稳定 owner IDs + 6 个 display knobs） | 001 |

## 7 关联计划

本 spec v1.0 已创建首个 active plan 目录 `001-report-screen-and-generating-handoff`；其余计划编号仍为预留，后续通过 `/design` 创建对应 plan/context 后再进入 `/implement`：

- `001-report-screen-and-generating-handoff` — GeneratingScreen 轮询 + ReportScreen 5 tab dashboard + 复练 CTA + 失败/缺 sessionId 兜底 + i18n + Playwright pixel parity + 旧口径负向 + BDD `E2E.P0.056-059`。
- `002-report-quality-feedback-and-listing` — 用户对报告打分 / 反馈（依赖 prompt-rubric-registry 003 grayscale） + listTargetJobReports 列表 UI（如产品决定开启 dashboard 历史浏览，需先修订 product-scope D-7）。
- `003-report-export-and-share` — 报告导出 / 分享（依赖未来隐私 spec）。

## 8 关联文档

- 上游 spec：[`engineering-roadmap`](../engineering-roadmap/spec.md) §5.2、[`product-scope`](../product-scope/spec.md) §6.9（M4 证据化报告）、[`product-scope`](../product-scope/spec.md) §4.1（产品原则）、[`frontend-shell`](../frontend-shell/spec.md)、[`frontend-workspace-and-practice`](../frontend-workspace-and-practice/spec.md)、[`backend-review`](../backend-review/spec.md)、[`backend-practice`](../backend-practice/spec.md)、[`openapi-v1-contract`](../openapi-v1-contract/spec.md)、[`mock-contract-suite`](../mock-contract-suite/spec.md)、[`shared-conventions-codified`](../shared-conventions-codified/spec.md)、[`prompt-rubric-registry`](../prompt-rubric-registry/spec.md)
- UI 真理源：`ui-design/src/screen-report.jsx`、`ui-design/src/screens-p0-complete.jsx::ReportGeneratingScreen`、`ui-design/src/app.jsx`（route mapping / `INTERVIEW_CONTEXT_ROUTES` / `hideTopBar`）、`ui-design/src/primitives.jsx`、`ui-design/src/data.jsx`（`report` sample 数据结构）、[`docs/ui-design/report-dashboard.md`](../../ui-design/report-dashboard.md)、[`docs/ui-design/module-map.md`](../../ui-design/module-map.md)、[`docs/ui-design/INDEX.md`](../../ui-design/INDEX.md)
- 当前正式前端入口：`frontend/src/app/{routes.ts,App.tsx,screens/PlaceholderScreen.tsx}`、`frontend/src/api/{generated/client.ts,mockTransport.ts}`、`frontend/src/app/runtime/AppRuntimeProvider.tsx`、`frontend/src/app/auth/pendingAction.ts`、`frontend/src/app/i18n/locales/{zh,en}.ts`、`frontend/src/app/theme/`、`frontend/src/app/interview-context/`、`frontend/tests/pixel-parity/`
- Fixture：`openapi/fixtures/Reports/getFeedbackReport.json`、`openapi/fixtures/Reports/listTargetJobReports.json`
- 治理 / 流程：[`AGENTS.md`](../../../AGENTS.md)、[`docs/development.md`](../../development.md) §2、[`docs/spec/README.md`](../README.md)、[`docs/spec/TEMPLATES.md`](../TEMPLATES.md)、[`test/scenarios/README.md`](../../../test/scenarios/README.md)
- 历史：[history.md](./history.md)

## 10 修订记录

| 版本 | 日期 | 说明 |
|------|------|------|
| 1.3 | 2026-06-13 | 对齐 product-scope v2.1 D-19 报告 CTA 单点收敛：新增 D-19 决策、§2.1 next tab 去 CTA + 题目本地标记、C-16/C-17 验收行；CTA 唯一入口收敛到 Header（C-9/C-10 标注）；plan 001 重开承接，纯前端无契约变更 |
