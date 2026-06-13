# Frontend Home / Job Picks / Parse Spec

> **版本**: 2.0
> **状态**: active
> **更新日期**: 2026-06-13

> **2026-06-12 product-scope v2.1 对齐声明**：本 spec v2.0 承接两项锁定决策——
> **D-17**：岗位推荐模块（`jd_match` / Job Picks）整体删除，JD 获取唯一入口是首页导入，一级导航收敛为 `首页 / 模拟面试 / 简历 / 复盘` 四项。本 spec 中所有 jd_match 屏 / 三 tab / JobMatch 契约消费 / Job Picks aux card 相关条目（§2.1 JD Match 屏、D-1 / D-8 / D-9 / D-10 / D-11 / D-12、C-8 / C-12~C-16）自 v2.0 起退役为历史记录，不得作为新实现依据；前端删除范围与验收见 §9，由 [plan 002](./plans/002-jd-match-recommendations/plan.md) 原地重开承接。
> **D-14**：JD 导入单次确认——`parse` 解析确认页同时承载启动决策（核对解析结果、绑定简历、确认 InterviewRound、立即面试 / 仅保存规划），首次导入链路只允许一次全页确认；旧「确认并进入面试前确认」二次确认按钮删除，`workspace` 不再是首次导入必经第二确认页。目标交互契约见 §10，由 [plan 001](./plans/001-home-jd-import-and-parse/plan.md) 原地重开承接；C-5 / D-4 中与二次确认相关的口径以 §10 为准。

## 1 背景与目标

`frontend-home-job-picks-and-parse` 是 `engineering-roadmap` S1 工程蓝图中明确预占的前端业务 subspec，承接 `frontend-shell`（D1+D2+D3 视觉系统）已交付的 App 壳、TopBar、五个一级入口、route normalization、`requestAuth(pendingAction)` 与 fixture-backed mock transport，落地用户首次进入并产出一次模拟面试上下文的入口闭环。

本 subspec 的目标是把当前 `ui-design/` 静态原型中 `home`、`jd_match`、`parse` 三个屏幕从 mock data 迁移到正式前端工程，并通过 generated client + fixture-backed transport 消费已存在的 `TargetJobs` OpenAPI 契约，使「带着 JD 来的用户」能够在最短路径完成「粘贴/上传/URL 导入 JD → 解析确认 → 进入模拟面试规划」的 P0 主路径。2026-05-22 起，plan 001 在保留 fixture-backed UI variants 的同时，用 `VITE_EI_API_MODE=real` generated-client gate + backend TargetJob live scenarios 证明 TargetJobs/import/parse 真实 backend 联调闭环。

`frontend-shell/spec.md` §2.1 `parse` 路由壳与 `eiCreateInterviewContext` 等价契约由本 subspec 承接业务内容；`backend-targetjob`（handler / service / store）和 `backend-upload`（upload presign）由独立 owner 承接并已完成对应真实 handler。

## 2 范围

### 2.1 In Scope

- Home 屏（`route=home`）：
  - Hero（label / title / sub）按 `ui-design/src/screen-home.jsx` 源级复刻
  - JD 导入卡片：textarea + upload modal + URL modal 三 source variants
  - Recent mock interviews 列表：消费 `listTargetJobs`，渲染 `MockInterviewCard` + `MiniRoundRail`
  - Empty state：当 `listTargetJobs` 返回空数组时引导粘贴/上传 JD，不展示占位面试数据
  - Auxiliary cards：`JOB PICKS`（→ `jd_match`）+ `POST-INTERVIEW`（→ `debrief`）
  - Resume create CTA：未登录态可见，未登录点击触发 `requestAuth(pendingAction)`
  - i18n `zh` / `en` 全文案接入 D1 typed locale helper
  - 2026-05-22 plan 001 L2 remediation：P0.014-P0.016 trigger 前置 `targetJob.realApiMode.test.ts`，证明 `listTargetJobs` / `createUploadPresign` / `importTargetJob` / `getTargetJob` / `updateTargetJob` 的 production generated client 指向真实 backend base URL；fixture-backed UI variants 继续用于确定性 DOM / failure / privacy 覆盖
- Parse 屏（`route=parse`）：
  - Loading 阶段：4 步进度条 + footer 模型 / rubric / prompt hash 占位文案
  - Preview/Confirm 阶段：Basic fields 行内可编辑、Must Have / Nice to Have requirements 块带 hit/partial/gap toggle、Hidden signals 块（`TargetJobSummary.interviewHypotheses` 推断）、Round assumptions 4 卡
  - Footer actions：Cancel → `home`、Re-parse、Confirm → `workspace` 携带完整 `eiCreateInterviewContext` 等价 params
  - 通过 `analysisStatus` 状态机（`queued` / `processing` / `ready` / `failed`）驱动 loading→preview 切换；不假装"正在调用 LLM"
- JD Match 屏（`route=jd_match`）：
  - **plan 001**：P1 placeholder shell（保留 TopBar nav 高亮 + 路由可达；Hero + Profile snapshot chip 静态渲染；三 tab 显示 placeholder copy；不消费任何 mock 数据）
  - **plan 002**：完整三 tab 业务（Recommended / Search / Watchlist）+ Profile snapshot chip 数据驱动 + AGENT scan status badge + Save/Mark not relevant/Confirm interview/Open source 闭环 + 自然语言 Search + Saved searches + Watchlist + Market signals；通过 generated `JobMatch` client + fixture-backed transport 消费 OpenAPI 新增 12 个 operationId（`getJobMatchProfile` / `getAgentScanStatus` / `listJobRecommendations` / `getJobRecommendation` / `addToWatchlist` / `removeFromWatchlist` / `markJobNotRelevant` / `searchJobs` / `listSavedSearches` / `createSavedSearch` / `listWatchlist` / `getMarketSignals`）；side-effect 操作均带 `Idempotency-Key`；2026-05-22 起 `backend-jobs-recommendations/001-jd-match-real-backend-baseline` 已落地真实 handler，frontend plan 002 通过 `VITE_EI_API_MODE=real` generated-client gate 证明 12 个 operation 指向真实 backend base URL，同时保留 fixture-backed UI variants；与 D-2 模式一致，前端不直连 LLM/provider/外部招聘平台
- 与 D1 `requestAuth(pendingAction)` 集成：未登录用户提交 import 与 confirm interview 时触发 pendingAction，登录后恢复
- 与 D2/D3 `ui-design/` parity gate 集成：home / parse 两屏新增 Vitest+jsdom smoke 与 Playwright desktop+mobile pixel parity 测试

### 2.2 Out of Scope

- `jd_match` backend handler / service / store / agent scan pipeline / AI-backed search / 候选池抓取 / market signals 计算的代码实现 — 已由独立 subspec `backend-jobs-recommendations/001-jd-match-real-backend-baseline` 承接并完成；本 frontend subspec 不拥有 backend 代码，但 plan 002 必须通过 `VITE_EI_API_MODE=real` generated-client gate + backend E2E.P0.094-P0.097 证明真实 API 联调闭环
- 真实联网搜索（LinkedIn / Boss / 脉脉 / 拉勾 / 公司官网 API 直连） — 由 `backend-jobs-recommendations` 承接
- AGENT 真实定时扫描调度 — 仅前端展示 backend 返回的 `lastScanAt` / `nextScanAt`，frontend 不实现真实定时器或 SSE/WebSocket 推送
- Watchlist 与 Saved Searches 客户端持久化 — 锁定为服务端持久化（D-9），frontend 不写 localStorage / sessionStorage
- jd_match → parse 反向状态闭环（推荐已使用标记、`markJobMatchAsConsumed` 等）— plan 002 frontend 仅锁定 nav 出口 params 携带 `sourceJobMatchId`，反向更新由后续 plan 承接（D-8）
- `workspace` 屏内业务（mock plan 状态、轮次切换、简历绑定、公司情报）— 由 `frontend-workspace-and-practice` subspec 承接
- `practice` / `report` / `debrief` / `resume_versions` 屏业务 — 各自独立 subspec
- 真实 LLM 调用、JD 抓取、URL fetch、文件上传二进制处理的 backend 代码所有权 — 由 `backend-targetjob`、`backend-upload` 与 `backend-runtime-topology` 承接；本 frontend subspec 不实现 backend 代码，但 plan 001 必须通过 real-mode generated-client gate + backend E2E.P0.010-P0.013 / upload route-handler focused tests 证明当前联调不再停留在 fixture-only 状态
- 数据库 schema、event/outbox、AI provider profile 接入 — 由 B4 / B3 / A3 承接
- 不新增旧 `welcome` / `growth` / `mistakes` / `drill` / `followup` / `experiences` / `star` / 独立 `voice` route 别名

## 3 用户决策 / 待确认事项

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | jd_match 业务范围 | 采用「契约先行 + frontend fixture 消费 + real-mode generated-client 联调 gate」模式，与 D-2 一致：plan 001 落地 P1 placeholder shell；plan 002 先扩展 OpenAPI `JobMatch` tag + 12 operationId + fixture，frontend 通过 generated client + fixture-backed transport 一次性源级复刻 ui-design 三 tab 完整业务；2026-05-22 `backend-jobs-recommendations/001-jd-match-real-backend-baseline` 已完成真实 backend handler / service / store / agent scan pipeline / AI-backed search / candidate pool / market signals，plan 002 原地补 `VITE_EI_API_MODE=real` gate 证明 12 operation 的 production generated client 指向真实 backend base URL；新增 `JobMatch` tag 时已同步 B2 owner truth source（`openapi-v1-contract` endpoint inventory、OpenAPI/fixture README、inventory linter、fixture validator 和 mock-contract coverage 口径），把 12 tag / 34 endpoint gate additive 升级到 13 tag / 46 endpoint | 与 D-2 generated client + fixture-backed transport 模式对齐，同时避免 fixture-backed UI 被误判为真实 backend 闭环；frontend 永远不直连 LLM/provider/外部招聘平台；真实 backend semantics 由 backend P0.094-P0.097 live scenarios 配对证明 |
| D-2 | Parse loading 进度驱动 | 前端只通过 generated API client 调 `getTargetJob` 轮询 backend 返回的 `analysisStatus` 状态机：`queued` → `processing` → `ready` / `failed`；UI 4 步进度条与 footer 只源级复刻 `ui-design` 的解析节奏与版式，不代表前端直接调用 LLM；2026-05-22 plan 001 增加 `VITE_EI_API_MODE=real` generated-client gate，证明 production bootstrap 对 TargetJobs/import/parse operation 指向真实 backend base URL；2026-05-24 regression gate 固化：即使首次 `getTargetJob` 已返回 `ready`，正式前端也必须先展示 `ui-design` 4 步 loading 演示并按 tick 完成后再进入 preview，禁止直接跳过 loading 到 parsed preview；同一 mounted `ParseScreen` 收到新的 `targetJobId` 时必须清空旧 preview/edit state，回到 loading gate，并在 loading 完成后 hydrate 新 TargetJob | 当 fixture transport / backend 返回 `ready` 时进度条以可观察节奏快速完成但不可被跳过；返回 `failed` 时切错误态而非伪装继续；前端不得接入 AI provider、prompt registry、LLM key 或任何 provider-specific endpoint；真实 backend semantics 由 backend-targetjob E2E.P0.010-P0.013 配对证明 |
| D-3 | Hidden signals 来源 | 前端只展示 backend/API 返回的 `TargetJobSummary.interviewHypotheses`（对象级 `provenance` 必须存在）+ `TargetJobSummary.coreThemes`；`fitSummary.riskSignals` 用于 "WHERE IT'S A STRETCH" 类风险呈现；结构与 icon/置信度 tag 必须与 `ui-design` Hidden signals 卡片一致 | 不在前端凭 JD 文本推断、补写、改写或重新生成 hidden signals；所有 AI-generated 字段必须通过 OpenAPI fixture / backend response 可追溯到 `GenerationProvenance` |
| D-4 | Confirm 跳转契约 | `nav("workspace", interviewContextFromTargetJob(targetJob))`；完整参数为 `targetJobId`、`jobId`、`jdId`、`planId`、`resumeVersionId`、`roundId`、`roundName`；`planId` 由 D1 `eiCreateInterviewContext` 等价契约从 `targetJobId` 推导（`plan-${targetJobId}`） | 真实 `createPracticePlan` API 调用由 `frontend-target-job-workspace` 承接；本 subspec 不主动创建 PracticePlan |
| D-5 | i18n locale 拆分 | 在 `frontend/src/app/i18n/locales/zh.ts` / `en.ts` 中新增 `home.*`、`parse.*`、`jdMatch.*` 三个命名空间；不混入 messages.ts 类型聚合层 | 与 D1 D-7 i18n 规则一致；新增命名空间需通过 D1 typed helper test |
| D-6 | Auth gate 触发点 | Paste/Upload/URL 三种 import 提交在未登录时触发 `requestAuth({ type: "import_jd", route: "home", params: { pendingImportId, source }, label })`；`pendingImportId` 只引用当前 SPA 会话内存中的待提交 source payload，不包含 JD 原文或 source URL；Parse Confirm 在未登录时触发 `requestAuth({ type: "confirm_interview", route: "workspace", params: interviewContextFromTargetJob(targetJob) })`，必须携带与已登录跳转一致的 7 个 workspace params | 已登录用户直接执行；未登录跳 `auth_login`，登录后恢复目标 route 与 params；import 恢复先回 home 自动提交，再跳 parse；与 frontend-shell C-2 一致 |
| D-7 | Privacy 红线 | JD 原文（rawText / rawDescription / sourceUrl）不进入 logger 字符串、URL query、localStorage、telemetry payload；只通过 generated client request body 与 React state 传递；fixture redact lint 必须覆盖 | 与 product-scope spec §1.6 隐私默认保守一致；observability redact rule 已在 D1 接入 |
| D-8 | jd_match → parse 反向数据流 | plan 002 frontend 仅锁定 nav 出口契约 `nav("parse", { source: "jd_match", sourceJobMatchId })`；parse 屏是否反向更新推荐已用状态（如 `markJobMatchAsConsumed` operation、列表显示 used badge 等）由后续 plan / 独立 subspec 承接 | plan 002 不修改 parse subspec 文档与代码；parse 屏可在未来选择性接入 source 标识；E2E.P0.031 仅断言出口 params 完整性，不断言反向状态 |
| D-9 | Watchlist 与 Saved Searches 持久化策略 | 服务端持久化（backend 承接 `watchlist_items` + `saved_searches` 表与对应 operationId）；frontend 通过 generated client 调用 `addToWatchlist` / `removeFromWatchlist` / `listWatchlist` / `createSavedSearch` / `listSavedSearches` 等 operation；不写 localStorage / sessionStorage / IndexedDB | 与 D-7 隐私默认保守一致；frontend 不引入 client-side persistence 依赖；plan 002 frontend 实现仅消费 generated client，backend 持久化由 backend-jobs-recommendations 承接 |
| D-10 | Agent scan 状态来源 | 通过单独 operationId `getAgentScanStatus` 返回 `{ status: enum<idle\|scanning\|error>, lastScanAt, nextScanAt, scannedSourceCount, ... }`；frontend 在进入 jd_match + 切回 Recommended tab 时各调一次；不引入 SSE / WebSocket / 真实定时器；refreshIntervalHours 由 backend 决定，frontend 不硬编码 4h | 与 D-2 polling 模式一致；frontend 不实现后台任务；状态变化慢，UI 仅在主动切 tab 时刷新 |
| D-11 | jd_match 隐私红线扩展 | 自然语言搜索 query / saved-search label / watchlist label / sourceJobUrl / 任意 reason freeNote / linkedJobMatchId 不进入 logger 字符串、URL query、localStorage、telemetry payload；只通过 generated client request body 与 React state 传递；fixture redact lint 必须覆盖；`window.open` 调用 source URL 时必须带 `noopener,noreferrer` flags | 与 D-7 一致并扩展；jd_match 比 home/parse 多一组隐私字段（query / watchlist / source URL / freeNote），需在 redact 列表与 negative grep 中显式锁定 |
| D-12 | Search loading 形态 | 以 ui-design `screen-jd-match.jsx::SearchTab` `searching=true` 区域**当前形态**为真理源（保留 5 步 AGENT panel + opacity 渐变 + accent 标签）：保留 `● AGENT SCANNING` / `● AGENT 扫描中` accent label、1px accent 外边框、3px accent 左边框、5 行 step 文案 + `opacity: i <= 2 ? 1 : 0.4` 渐变（前 3 步 active 1.0、后 2 步 dim 0.4）；正式前端通过 i18n 5 个 step key（典型命名 `jdMatch.search.searchingStep1` … `searchingStep5` 或 `jdMatch.search.searchingSteps[]`，最终命名以 `messages.ts` typed helper 落地为准）源级复刻 5 步文案；ui-design 中的动态 JD 数字（`248 → 87 unique postings` / `248 → 87 条唯一岗位`）必须在前端 i18n 里替换为不含具体数字的静态语义文案（如 `→ Dedup by JD hash · removing duplicate postings` / `→ 按 JD 哈希去重 · 移除重复岗位`），以避免前端在缺失真实统计时硬编码假数据；frontend loading 由 `searchJobs` 请求 in-flight 状态驱动，从 Run 起持续显示 5 步 panel 直到 results 渲染或 inline error 出现；不展示真实步骤进度切换，不依赖 backend `progress` 字段；不引入 setInterval / SSE / WebSocket 真步骤推进；保留 ui-design 现有 opacity 渐变作为唯一动画效果，不新增 keyframes / transition / 真步骤切换动画；plan 002 不修改 ui-design 静态文件；若未来 ui-design owner 修订 SearchTab `searching` 区域形态（包括但不限于：简化为单一文案、加入真步骤切换动画、改写 step 文案），由后续 plan 接力，不影响 plan 002 当前实施 | `searchJobs` response 不再包含 progress 字段；前端动画/文案自治；正式前端严格保留 ui-design 当前 5 步 AGENT panel 真理源；前端不写动态计数也不在 i18n 文案中嵌入硬编码数字；防止前端二次设计或先于 ui-design 改写 |

## 4 设计约束

- 视觉与交互必须以 `ui-design/src/screen-home.jsx`、`ui-design/src/screen-jd-match.jsx`、`ui-design/src/screens-p0-complete.jsx::ParseScreen`、`ui-design/src/primitives.jsx` 为唯一真理源进行源级复刻；DOM 构图、控件类型、icon、菜单/弹层层级、aria 状态、主要交互路径必须可追溯到对应 jsx 函数；不得二次设计或重新解释视觉
- Parse 屏的 4 步 loading 文案、模型/rubric/prompt hash footer 的 DOM 构图、节奏、层级与可见文案必须与 `screens-p0-complete.jsx::ParseScreen` lines 10-104 一致；正式前端只能把这些值作为 backend parse metadata / fixture metadata 的展示，不得因此接入前端 LLM 调用或 provider 配置；任何视觉或文案修改必须先改 `ui-design/` 真理源
- `MockInterviewCard` 的 status pill / round rail / company meta slot / 标题 / 地点 DOM 必须与 `screen-home.jsx::MockInterviewCard` lines 148-216 一致；数据只能来自 generated `TargetJob` schema：company meta slot 显示 `companyName · status-derived label`，`statusTone` 从 `TargetJob.status` 派生，round rail 使用本 plan 明确的 P0 默认轮次与 currentIndex fallback；不得从 `ui-design/src/data.jsx`、未声明 fixture 字段或本地 mock 补 `level` / `nextRound` / `statusTone`
- `JDAssistModal` 的 upload / URL 双模态、关闭按钮、Continue/Cancel actions 必须与 `screen-home.jsx::JDAssistModal` lines 218-262 一致
- 所有 import source variants 必须通过 generated `ImportTargetJobRequest` schema 提交；`type` discriminator + 必填字段在前端 form-level 校验
- Parse 屏 hit/partial/gap toggle 状态在前端是 ephemeral UI state；用户保存（Confirm）时不写回 `TargetJobRequirement.evidenceLevel`；evidenceLevel、summary、fitSummary、hidden signals 均是 backend/API 返回的 AI-generated 只读字段，前端不得用本地规则或 LLM 重新推断
- Confirm 时调用 `updateTargetJob` 写回用户编辑的 title/companyNameHint/locationText/notes；不调用 `createPracticePlan`
- listTargetJobs 必须按 `updatedAt desc` 取最近 N 条（N=12 默认；服务端 pagination 由 generated client `cursor` 接力）；当前 generated client 通过 `RequestOptions.query.pageSize=12` 传参
- `createUploadPresign` / `importTargetJob` / `updateTargetJob` 都是 side-effect operation，前端必须通过 generated client 传 `idempotencyKey`，并在测试中断言 `Idempotency-Key` header 存在
- Job Picks aux card 即使 `jd_match` 在 P1，也必须保持点击可达 — 路由可见即"已交付"
- i18n 必须支持 zh/en；初始 UI 语言跟随浏览器 locale，未知或缺失 fallback `en`；与 D1 D-7 一致
- 暗色 / customAccent 必须在 home + parse 两屏均通过 root level `data-theme` / `data-mode` / `data-custom-accent` 切换生效，不允许硬编码颜色
- Pixel parity gate 必须在 desktop (1440×900) + mobile (390×844) 两个 viewport 下断言 home + parse 的 DOM 锚点 / 关键 computed style / bounding box / 截图差异；任何 parity 失败必须修到与 `ui-design/` 一致或先修订真理源
- Mobile 响应式：home 与 parse 在 ≤768px viewport 下不能溢出视口；parse 屏 Requirements 双列在 mobile 折叠为单列；MockInterviewCard 网格 `repeat(auto-fill, minmax(320px, 1fr))` 在 mobile 自然成单列
- 所有可见用户字段（status pill / round rail / requirements label）必须有 `data-testid`，遵循 D1/D2 既有命名（`home-recent-mock-card-${id}` / `parse-requirement-${kind}-${idx}` 等）

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| frontend home / parse / jd_match shell | `frontend-home-job-picks-and-parse` | 三屏 React 组件、路由壳业务内容、i18n、源级 parity 测试 |
| App shell / route normalization / requestAuth / locale helper / generated client / fixture transport bootstrap | `frontend-shell` | D1 已交付，本 subspec 直接消费 |
| TargetJobs OpenAPI 契约 | `openapi-v1-contract` | `importTargetJob` / `listTargetJobs` / `getTargetJob` / `updateTargetJob` schema 与 fixture |
| Mock transport / fixture-backed response | `mock-contract-suite` | 本 subspec 通过 generated client mock transport 消费 fixture；fixture-backed UI variants 用于稳定覆盖 DOM / failure / privacy，不再作为真实 backend 完成证据 |
| TargetJob persistence / runner / event 发射 | `backend-targetjob` + `event-and-outbox-contract` + `db-migrations-baseline` | 真实 backend handler / store / event 已由 `backend-targetjob/001-targetjob-import-and-parse-bootstrap` 完成；plan 001 用 `targetJob.realApiMode.test.ts` + backend E2E.P0.010-P0.013 配对证明 frontend generated client 与真实 TargetJobs backend 对齐 |
| AI parsing 调用 | `ai-provider-and-model-routing` + `prompt-rubric-registry` | 真实 LLM 调用通过 backend；本 subspec 不直接消费 AI |
| jd_match frontend 三 tab | `frontend-home-job-picks-and-parse/plans/002-jd-match-recommendations` | 三 tab React 组件、子组件、`JobMatch` OpenAPI 契约扩展（12 operationId）、fixture、i18n、源级 parity 测试、5 个 BDD 场景；通过 generated client + fixture-backed transport 闭环 |
| jd_match `JobMatch` tag handler / store / agent scan pipeline / AI-backed search / 候选池抓取 / market signals 计算 | `backend-jobs-recommendations/001-jd-match-real-backend-baseline` | 真实 backend 已完成；frontend plan 002 保留 OpenAPI 契约 + fixture UI variants，并通过 `VITE_EI_API_MODE=real` generated-client gate + backend E2E.P0.094-P0.097 证明真实 API 联调 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | Home 默认渲染 | 用户未登录或已登录，无 TargetJob | 打开 `home` 路由 | Hero / textarea card / upload + URL 入口 / Job Picks aux card / Post-interview aux card / Resume create CTA / empty state 完整渲染；TopBar 高亮 home；i18n zh/en 文案均通过 typed helper；CSS variables `--ei-bg`/`--ei-ink`/`--ei-accent`/`data-mode` 切换生效；DOM 锚点能追溯到 `ui-design/src/screen-home.jsx` | 001-home-jd-import-and-parse |
| C-2 | Paste JD → Parse | 用户在 Home 输入框粘贴 JD 文本 | 点击「解析并确认面试」 | 调用 `importTargetJob`（`source.type=manual_text`、`source.rawText=JD 文本`、`targetLanguage=当前 UI locale`）；成功后路由跳 `parse?targetJobId=…`；Parse 屏先显示 4 步 loading，再轮询 `getTargetJob` 至 `analysisStatus=ready` 切到 preview；preview 渲染 fixture 中 title/companyName/locationText/requirements/summary.interviewHypotheses/fitSummary.riskSignals；JD 原文不写入 URL/localStorage/telemetry | 001-home-jd-import-and-parse |
| C-3 | Upload / URL 双 source variant | 用户在 Home 点击 upload 或 URL 入口 | 在 JDAssistModal 中确认 | Upload 路径先调用 generated `createUploadPresign`（`purpose=target_job_attachment`，带 `Idempotency-Key`），fixture 返回 `fileObjectId` 后再提交 `importTargetJob` `source.type=file`；URL 路径提交 `source.type=url`（`url` 字段）；后续流程与 C-2 一致；DOM 与 `screen-home.jsx::JDAssistModal` 行为一致（关闭、Continue、Cancel）；frontend tests 不真实上传二进制到 object storage | 001-home-jd-import-and-parse |
| C-4 | Recent mock interviews 列表 | 用户已登录，listTargetJobs 返回 N 条 TargetJob | 进入 home | 渲染最多 12 张 `MockInterviewCard`，按 `updatedAt desc` 排序；卡片显示 `companyName · status-derived label` / title / locationText / status pill（statusTone 从 `TargetJob.status` 派生）/ MiniRoundRail P0 fallback 当前轮次圆点；点击卡片调 `nav("workspace", interviewContextFromTargetJob(targetJob))`，默认补齐 `targetJobId / jobId / planId / jdId / resumeVersionId / roundId / roundName`；列表为空时显示 `HomeEmptyState`，`回到 JD 输入` 按钮 focus textarea | 001-home-jd-import-and-parse |
| C-5 | Parse 编辑与 Confirm | 用户在 Parse preview 编辑 OpenAPI 当前允许保存的 title / company / location / notes 字段，并切换若干 hit toggle | 点击「确认并进入面试前确认」 | 调用 `updateTargetJob`（仅 supplied fields，例：`titleHint` / `companyNameHint` / `locationText` / `notes`，带 `Idempotency-Key`）；level / language 槽位按 `ui-design` DOM 展示但为 read-only，直到 B2 扩展 `UpdateTargetJobRequest`；hit toggle 不写后端；成功后路由跳 `workspace?targetJobId=&jobId=&jdId=&planId=&resumeVersionId=&roundId=&roundName=`，使用 `interviewContextFromTargetJob(targetJob)` 推导 D1 `eiCreateInterviewContext` 等价默认值；默认 `resume-unbound` 时 Workspace 可先进入 `workspace-missing-resume` 下一步状态；任何 `updateTargetJob` 4xx 显示 inline 错误并保留编辑态；Cancel 跳 `home`；Re-parse 重置 `stage=loading` 并重新轮询 `getTargetJob` | 001-home-jd-import-and-parse |
| C-6 | Parse 失败态 | `getTargetJob.analysisStatus=failed` 或 polling 超时 | 用户在 Parse loading 阶段等待 | 切到 error state：显示「JD 解析失败」标题 + 失败原因 + 重新解析按钮 + 返回首页按钮；不展示伪造的 preview 数据；测试通过 fixture variant 锁定 | 001-home-jd-import-and-parse |
| C-7 | Auth pending action 恢复 | 未登录用户在 Home paste JD 并提交 / 在 Parse 点击 Confirm | 进入 auth_login 完成登录 | 登录成功后回到原 route 与 params：paste/upload/url import 流先回到 Home，通过 opaque `pendingImportId` 消费当前 SPA 会话内存中的待提交 source payload 并自动重新发起 `importTargetJob`，成功后跳 Parse；pending route params 不携带 JD 原文、source URL 或 rawDescription；Confirm 流回到 workspace 携带完整 7 字段 interview context（`targetJobId` / `jobId` / `jdId` / `planId` / `resumeVersionId` / `roundId` / `roundName`）；与 `frontend-shell` C-2 一致 | 001-home-jd-import-and-parse |
| C-8 | jd_match P1 placeholder shell | 用户点击 TopBar Job Picks 或 home aux card「打开岗位推荐」 | 路由进入 `jd_match` | 渲染 hero + profile snapshot chip 静态版本 + 三 tab 标签（推荐 / 联网搜索 / 关注列表）；tab 内容区显示 "Coming in P1" 占位文案 + 关联 plan 002 引用；TopBar 高亮 jd_match；DOM 保留主要锚点（`jdmatch-hero` / `jdmatch-tab-${k}` / `jdmatch-placeholder`）便于后续 plan 002 接力；不展示任何 mock recommendation 数据 | 001-home-jd-import-and-parse |
| C-9 | UI source structure parity | D1+D2+D3 已交付，新增 home / parse / jd_match shell | Vitest+jsdom 测试 | DOM 锚点、控件类型（textarea / button / modal / 自定义弹层）、icon name、aria 状态、主要交互路径必须能追溯到 `ui-design/src/screen-home.jsx` / `screens-p0-complete.jsx::ParseScreen` / `screen-jd-match.jsx`（仅 hero+tab 部分）；旧 prototype 中存在但当前真理源已移除的 testid / control 类型负向断言不命中；任何 parity 失败必须修到与原型一致或先修订 `ui-design/` 真理源 | 001-home-jd-import-and-parse |
| C-10 | UI visual geometry parity | C-9 通过 | Playwright 在 desktop (1440×900) + mobile (390×844) 双 viewport 下加载 `frontend/dist` home 与 parse 路由 | 关键区块 bounding box 不重叠且 stays in viewport；warm/light + dark + customAccent 三态切换关键元素 computed background / color 出现可见变化；mobile viewport 下 Requirements 双列折叠为单列、textarea card 不溢出；新增 `tests/pixel-parity/home.spec.ts` 与 `tests/pixel-parity/parse.spec.ts`；与 D2/D3 现有 21 个 spec 累加；CI / 本地 `pnpm --filter @easyinterview/frontend test:pixel-parity` 通过 | 001-home-jd-import-and-parse |
| C-11 | Privacy 红线 | 用户提交 JD 原文 | observability / log / URL / localStorage / telemetry 输出 | JD raw text / rawDescription / sourceUrl 不出现在 console.log、不出现在 URL query、不写入 localStorage、不进入任何 telemetry payload；前端 redact lint 反查通过；fixture transport 不在 mockTransport 日志中泄漏 raw 内容 | 001-home-jd-import-and-parse |
| C-12 | jd_match Recommended tab + Profile chip + AGENT 状态完整渲染 | 用户进入 `jd_match` 路由（已登录或未登录均可），`getJobMatchProfile` / `getAgentScanStatus` / `listJobRecommendations` fixture 配置多种 variant（idle/scanning/error；empty/one/many/failed） | 默认 Recommended tab 加载 | （1）Hero / Profile snapshot chip（驱动来源 `getJobMatchProfile`）/ AGENT badge（驱动来源 `getAgentScanStatus`）/ 三 tab 标签（带数量 badge）/ Recommended tab 双列布局（左 1.1fr / 右 1.4fr）/ JobMatchCard 列表（驱动来源 `listJobRecommendations`）/ 底部 next scan footer 全部渲染并 testid 命中；（2）plan 001 placeholder 文案与 testid 在 DOM 0 命中；（3）i18n zh/en 切换；（4）warm/light → dark → customAccent 三态切换关键 computed 颜色变化；（5）TopBar `topbar-nav-jd_match` 高亮 | 002-jd-match-recommendations |
| C-13 | jd_match JobMatchCard 详情 + Save / Mark not relevant / Confirm interview / Open source 闭环 | 用户在 Recommended tab 选中某张卡 | 点击 JDDetail 4 button | （1）JDDetail header / + Why matches / ⚠ Where stretch / Role snapshot / INTEL 条件渲染 / Action bar 渲染；（2）Save → `addToWatchlist` + `Idempotency-Key` + window.eiToast + button 状态切；再点 → `removeFromWatchlist` + toast；（3）Mark not relevant → `markJobNotRelevant` + `Idempotency-Key` + reason enum + 卡片隐藏 + 自动选下一张 + toast；（4）Confirm interview → `nav("parse", { source: "jd_match", sourceJobMatchId })`；（5）Source → `window.open(url, "_blank", "noopener,noreferrer")`；（6）4xx → revert + error toast，不破坏 UI；（7）jobMatchId / sourceUrl / freeNote 不进 URL / localStorage / telemetry | 002-jd-match-recommendations |
| C-14 | jd_match Search tab 自然语言搜索 + savedSearches + filter + 5 步 AGENT panel loading + failure | 用户切换 Search tab，`searchJobs` / `listSavedSearches` / `createSavedSearch` fixture 配置多种 variant（default / empty / failed / slow-response） | 输入 query 点 Run；切换 4 个 chip filter；点 Save current as watch；切 5xx variant 重跑 | （1）切 tab → `searchJobs` 调用 0 次；`listSavedSearches` 调 1 次；Search 输入区按 ui-design 源级复刻 `NATURAL LANGUAGE SEARCH` / `自然语言搜索` label、search icon、`SOURCES` / `数据源` label 与 5 个 source chips（LinkedIn / Boss 直聘 / 脉脉 / 拉勾 / Company sites 或 公司官网）；（2）Run → `searchJobs` + `Idempotency-Key`；in-flight 期间渲染 5 步 AGENT panel（`● AGENT SCANNING` accent label + 5 个 step 文案通过 i18n typed helper 渲染 + opacity 渐变前 3 步 active 1.0 / 后 2 步 dim 0.4）持续显示直到 results 渲染或 failure inline error 出现；不引入 setInterval / 真步骤切换动画；负向断言：动态 JD 数字（`248` / `87` / `unique postings` / `唯一岗位`）在 SearchTab DOM 与前端 i18n 中 0 命中；（3）results 2 列网格 cap 6；4 个 chip filter（all/strong/remote/unseen）纯 client-side 切换；（4）Save current → `createSavedSearch` + `Idempotency-Key` + toast；（5）5xx → 失败 inline error + 保留输入；empty → no-results 空态；（6）query / saved-search label / filter state 不进 URL / localStorage / telemetry / console | 002-jd-match-recommendations |
| C-15 | jd_match Auth pending action | 未登录用户在 jd_match 点击 Save / Not relevant / Confirm interview / Run search / Save current as watch 等 side-effect | 进入 auth_login 完成登录 | 登录后回到原 route 与 params 并自动重新触发 side-effect；Recommended action 使用 `params: { tab: "recommended", selectedJobMatchId, action }` 恢复选中卡与动作；Search action 使用 `params: { tab: "search", action, pendingJdMatchActionId }` 携带仅当前 SPA 会话内有效的 opaque payload id，登录后从内存消费 query / label 再执行；query 文本 / watchlist label / saved-search label / sourceUrl / freeNote 等不进入 pendingAction params / URL / localStorage；与 frontend-shell C-2 + plan 001 C-7 一致 | 002-jd-match-recommendations |
| C-16 | jd_match Watchlist tab + Market signals + chevron handoff | 用户切 Watchlist tab，`listWatchlist` / `getMarketSignals` fixture 配置多种 variant（empty / few / partial-data / failed） | 列表 / Market signals 渲染；点 chevron 切回 Recommended | （1）`listWatchlist` + `getMarketSignals` 各调 1 次；（2）`jdmatch-watchlist-item-${id}` + `jdmatch-market-signal-${k}` testid 全命中；3 tone（ok/warn/muted）+ refresh footer i18n；（3）chevron → tab 切 Recommended + selected = `WatchlistItem.linkedJobMatchId`（backend 提供，不在前端 string match）；（4）empty variant → empty state；partial-data variant → 部分卡片渲染 + 缺失值显示 fallback；（5）watchlist label / sourceJobUrl / linkedJobMatchId 不进 URL / localStorage / telemetry | 002-jd-match-recommendations |

## 7 关联计划

- [001-home-jd-import-and-parse](./plans/001-home-jd-import-and-parse/plan.md) — Home + Parse 端到端 + jd_match P1 placeholder shell；2026-05-22 L2 remediation 补 TargetJobs/upload/import/parse real-mode generated-client gate + backend E2E.P0.010-P0.013 配对证据；2026-05-24 regression remediation 固化 ready 响应也必须先展示 `ui-design` loading 演示，并把 P0.016 Confirm → Workspace browser gate 升级为 7 字段 route/context + `workspace-missing-resume` screenshot marker；同日补 same-route `targetJobId` switch regression，防止已 mounted Parse preview 继续显示旧 TargetJob（completed 2026-05-24）
- [002-jd-match-recommendations](./plans/002-jd-match-recommendations/plan.md) — jd_match 三 tab 完整 frontend 业务 + JobMatch OpenAPI 12 operationId + fixture + real-mode generated-client gate + 5 BDD 场景（completed L2 remediation 2026-05-22）

## 8 关联文档

- 上游 spec：[`engineering-roadmap`](../engineering-roadmap/spec.md)、[`product-scope`](../product-scope/spec.md)、[`frontend-shell`](../frontend-shell/spec.md)、[`openapi-v1-contract`](../openapi-v1-contract/spec.md)、[`mock-contract-suite`](../mock-contract-suite/spec.md)
- UI 真理源：`ui-design/src/screen-home.jsx`、`ui-design/src/screens-p0-complete.jsx::ParseScreen`、`ui-design/src/app.jsx`、`ui-design/src/primitives.jsx`、[`docs/ui-design/jd-resume-management.md`](../../ui-design/jd-resume-management.md)、[`docs/ui-design/ui-architecture.md`](../../ui-design/ui-architecture.md)、[`docs/ui-design/module-job-workspace.md`](../../ui-design/module-job-workspace.md)、[`docs/ui-design/removed-modules-and-scope.md`](../../ui-design/removed-modules-and-scope.md) §15（旧 `ui-design/src/screen-jd-match.jsx` 已随 2026-06-12 第二批裁剪删除）

## 9 D-17 前端删除范围与零残留验收（plan 002 active scope）

### 9.1 删除范围

| 资产 | 处置 |
|------|------|
| `frontend/src/app/screens/jd_match/` 全目录（screens / tabs / hooks / 子组件 / 测试） | 删除 |
| `routes.ts` `PRIMARY_NAV_ROUTES` 中 `jd_match` 项与 route 定义、`routeUrl.ts` `/jd-match` path 与 `JD_MATCH_SAFE`、`App.tsx` 渲染分支、`TopBar` `NAV_LABEL_KEYS` / `NAV_ICONS` 条目 | 删除；`jd_match` route key 与 `/jd-match` path 归一回 `home`（normalize alias + legacy path），一级导航收敛为 `home / workspace / resume_versions / debrief` 四项 |
| Home 屏 `JOB PICKS` aux card（→ jd_match 入口） | 删除；`POST-INTERVIEW` aux card 维持，与当前 `ui-design/src/screen-home.jsx` 对齐 |
| i18n `jdMatch.*` 命名空间与 `nav.jd_match` 词条（zh/en） | 删除 |
| `frontend/tests/pixel-parity/jd_match.spec.ts` 与 topbar parity 五入口 golden 断言 | 删除 / 改为四入口断言 |
| jd_match 相关 Vitest（unit / scenario / 路由测试中的 jd_match 用例） | 删除或改写为负向断言 |
| `test/scenarios/e2e/p0-017 / p0-027..031` 前端 jd_match 场景目录与 INDEX 行 | 删除 |
| `frontend/src/lib/jobs/jobs.ts` 等生成常量中的 jd_match 条目 | 随 shared codegen 再生成（上游删除归 backend plan 001 Phase 9） |

### 9.2 验收标准（v2.0 新增）

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-R1 | 四入口导航收敛 | D-17 删除完成 | 打开 App / 切换语言 / 切换主题 | TopBar 一级导航仅 `home / workspace / resume_versions / debrief` 四项；`topbar-nav-jd_match` 零命中；pixel parity 与 golden 预览四入口一致 | 002 removal phase |
| C-R2 | 旧 route 归一 | 用户直开 `/jd-match` 或旧 `jd_match` route key（含 localStorage 残留） | App normalize / parse URL | 归一到 `home`，不渲染独立 jd_match 屏；`legacyRouteNegative` 断言 jd_match 不在 live route 目录 | 002 removal phase |
| C-R3 | 前端零残留 | 删除完成 | `rg -i "jd[-_]?match|job picks"` 于 `frontend/src frontend/tests`（normalize alias / legacy path / 负向断言除外）；`pnpm test` / `typecheck` / `build` / `test:pixel-parity` | 零残留；全套件通过；Home 仅保留 `POST-INTERVIEW` aux card | 002 removal phase |

## 10 D-14 单次确认漏斗目标契约（plan 001 active scope）

以 `ui-design/src/screens-p0-complete.jsx::ParseScreen` 与 [docs/ui-design/module-job-workspace.md](../../ui-design/module-job-workspace.md) v1.10 为唯一 UI 真理源：

1. parse 确认页在解析 preview 基础上同时承载启动决策：轮次假设卡可点选确认 InterviewRound；绑定简历 pill（复用 ResumePickerModal）选择 / 更换简历；底部主操作为「立即面试」与「仅保存规划」。
2. 「立即面试」走 `requestAuth(create_session)` 登录拦截，pendingAction 恢复后直达 `practice`；「仅保存规划」进入 `workspace`。
3. 旧「确认并进入面试前确认」二次确认按钮删除；首次导入链路 parse 与 session 之间不存在平行确认页；`workspace` 定位为回访枢纽。
4. 详细 DOM 锚点、operation matrix 与验收行在 plan 001 v2.0 修订中固化（C-5 旧口径以本节为准失效）。

## 11 修订记录

| 版本 | 日期 | 说明 |
|------|------|------|
| 2.0 | 2026-06-13 | 对齐 product-scope v2.1：D-17 删除 jd_match 模块（新增 §9 删除范围与 C-R1~C-R3，jd_match 相关历史条目退役）；D-14 单次确认漏斗目标契约（新增 §10，plan 001 重开承接）；UI 真理源列表移除 screen-jd-match.jsx |
- 历史：[history.md](./history.md)
