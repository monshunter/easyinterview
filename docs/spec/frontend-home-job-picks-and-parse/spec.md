# Frontend Home / Parse Spec

> **版本**: 2.8
> **状态**: completed
> **更新日期**: 2026-07-06

> **2026-06-12 product-scope v2.1 对齐声明**：本 spec v2.0 承接两项锁定决策——
> **D-17**：岗位推荐模块（`jd_match` / Job Picks）整体删除，JD 获取唯一入口是首页导入，一级导航不再包含 Job Picks。本 spec 中所有 jd_match 屏 / 三 tab / JobMatch 契约消费 / Job Picks aux card 相关条目（§2.1 历史 JD Match 屏、D-1 / D-8 / D-9 / D-10 / D-11 / D-12、C-8 / C-12~C-16）自 v2.0 起不得作为新实现依据；前端删除范围与验收由 [product-scope/001-core-loop-module-pruning](../product-scope/plans/001-core-loop-module-pruning/plan.md) 承接，旧 002 plan 实体已删除。
> **D-14**：JD 导入单次确认——`parse` 解析确认页同时承载启动决策（核对解析结果、绑定简历、确认 InterviewRound、立即面试 / 仅保存规划），首次导入链路只允许一次全页确认；旧「确认并进入面试前确认」二次确认按钮删除，`workspace` 不再作为用户可见的首次导入二次确认页，仅允许作为 `autoStartPractice=1` 的会话创建技术桥接。目标交互契约见 §10，由 [plan 001](./plans/001-home-jd-import-and-parse/plan.md) 原地重开承接；C-5 / D-4 中与二次确认相关的口径以 §10 为准。
> **2026-07-06 product-scope D-22 再对齐**：复盘 / debrief 与 profile 独立模块也已删除；当前一级导航收敛为 `首页 / 模拟面试 / 简历`，Home 不再渲染 `JOB PICKS` 或 `POST-INTERVIEW` aux card。当前 active scope 只保留 Home + Parse 新建模拟面试入口，旧 `jd_match` / Job Picks / JobMatch / `debrief` 不得作为当前实现依据。

## 1 背景与目标

`frontend-home-job-picks-and-parse` 是 `engineering-roadmap` S1 工程蓝图中明确预占的前端业务 subspec，承接 `frontend-shell`（D1+D2+D3 视觉系统）已交付的 App 壳、当前三项一级入口、route normalization、`requestAuth(pendingAction)` 与 fixture-backed mock transport，落地用户首次进入并产出一次模拟面试上下文的入口闭环。

本 subspec 的当前目标是把 `ui-design/` 静态原型中仍 active 的 `home`、`parse` 两个屏幕迁移并保持到正式前端工程，通过 generated client + fixture-backed transport 消费已存在的 `TargetJobs` OpenAPI 契约，使「带着 JD 来的用户」能够在最短路径完成「粘贴/上传/URL 导入 JD → 解析确认 → 进入模拟面试规划」的 P0 主路径。2026-05-22 起，plan 001 在保留 fixture-backed UI variants 的同时，用 `VITE_EI_API_MODE=real` generated-client gate + backend TargetJob live scenarios 证明 TargetJobs/import/parse 真实 backend 联调闭环。

`frontend-shell/spec.md` §2.1 `parse` 路由壳与 `eiCreateInterviewContext` 等价契约由本 subspec 承接业务内容；`backend-targetjob`（handler / service / store）和 `backend-upload`（upload presign）由独立 owner 承接并已完成对应真实 handler。

## 2 范围

### 2.1 In Scope

- Home 屏（`route=home`）：
  - Hero（label / title）按 `ui-design/src/screen-home.jsx` 源级复刻；不再渲染旧 `home.heroSub`
  - JD 导入区：粘贴 JD 输入框是主输入；上传 JD 文件与 URL 导入是同一输入卡底部的 source actions，不再渲染独立上传 source panel
  - 首页新建模拟面试快捷入口必须先通过适度宽度的下拉框选择已有 ready 简历；`还没有简历？1 分钟创建 →` 必须在下拉框右侧同一行水平对齐，只导航到简历创建，不上传简历，也不得平铺所有简历为按钮列表
  - 主按钮文案为「立即面试」，位置在“选择已有简历”行下方，不得停留在 JD textarea 卡片右下角
  - Recent mock interviews 列表：消费 `listTargetJobs`，最多渲染 3 张 `MockInterviewCard` + `MiniRoundRail`；超过 3 条时展示“更多”并跳转到 `workspace` 模拟面试列表页
  - Empty state：当 `listTargetJobs` 返回空数组时引导粘贴/上传 JD，不展示占位面试数据
  - Retired auxiliary cards：`JOB PICKS`（→ `jd_match`）与 `POST-INTERVIEW`（→ `debrief`）均不得渲染；正式前端以负向测试和 pixel parity 锁定 0 命中
  - Resume create CTA：未登录态可见，未登录点击触发 `requestAuth(pendingAction)`
  - i18n `zh` / `en` 全文案接入 D1 typed locale helper
  - 2026-05-22 plan 001 L2 remediation：P0.014-P0.016 trigger 前置 `targetJob.realApiMode.test.ts`，证明 `listTargetJobs` / `createUploadPresign` / `importTargetJob` / `getTargetJob` / `updateTargetJob` 的 production generated client 指向真实 backend base URL；fixture-backed UI variants 继续用于确定性 DOM / failure / privacy 覆盖
- Parse 屏（`route=parse`）：
  - Loading 阶段：4 步进度条 + footer 模型 / rubric / prompt hash 占位文案
  - Preview/Launch 阶段：Basic fields 行内可编辑、Must Have / Nice to Have requirements 块带 hit/partial/gap toggle、Hidden signals 块（`TargetJobSummary.interviewHypotheses` 推断）、Round assumptions 4 卡、ready 简历绑定
  - Footer actions：Cancel → `home`、Re-parse、仅保存规划 → `workspace`、立即面试 → `workspace(autoStartPractice=1)`；两个成功出口均携带真实 `resumeId`
  - 通过 `analysisStatus` 状态机（`queued` / `processing` / `ready` / `failed`）驱动 loading→preview 切换；不假装"正在调用 LLM"
- Retired JD Match / Job Picks 屏：
  - `route=jd_match`、`/jd-match`、`frontend/src/app/screens/jd_match/`、`jdMatch.*` i18n、JobMatch generated client、P0.017 / P0.027-P0.031 场景均为历史资产；正式前端只保留 legacy route/key 归一到 `home` 的兼容入口和负向断言，不得恢复独立屏幕、TopBar 入口、Home aux card 或 JobMatch API 消费
- 与 D1 `requestAuth(pendingAction)` 集成：未登录用户提交 import 与 confirm interview 时触发 pendingAction，登录后恢复
- 与 D2/D3 `ui-design/` parity gate 集成：home / parse 两屏新增 Vitest+jsdom smoke 与 Playwright desktop+mobile pixel parity 测试

### 2.2 Out of Scope

- `jd_match` / Job Picks / JobMatch / Watchlist / Saved Searches / Agent scan / Market signals / AI-backed search 的任何正向实现或后续接力；这些模块已随 D-17 删除，旧 plan 实体不保留，删除证据由 product-scope owner 与负向 gate 承接
- `workspace` 屏内业务（mock plan 状态、轮次切换、简历绑定、公司情报）— 由 `frontend-workspace-and-practice` subspec 承接
- `practice` / `report` / `resume_versions` 屏业务 — 各自独立 subspec；`debrief` 独立前端入口已随 D-22 退役
- 真实 LLM 调用、JD 抓取、URL fetch、文件上传二进制处理的 backend 代码所有权 — 由 `backend-targetjob`、`backend-upload` 与 `backend-runtime-topology` 承接；本 frontend subspec 不实现 backend 代码，但 plan 001 必须通过 real-mode generated-client gate + backend E2E.P0.010-P0.013 / upload route-handler focused tests 证明当前联调不再停留在 fixture-only 状态
- 数据库 schema、event/outbox、AI provider profile 接入 — 由 B4 / B3 / A3 承接
- 不新增旧 `welcome` / `growth` / `mistakes` / `drill` / `followup` / `experiences` / `star` / 独立 `voice` route 别名

## 3 用户决策 / 待确认事项

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | jd_match 业务范围 | **Retired**：D-17 后 `jd_match` / Job Picks / JobMatch / P0.017 / P0.027-P0.031 均不是当前实现范围；旧 plan 001 placeholder 与 plan 002 三 tab 业务只作为历史记录。当前实现只允许 legacy route/key 归一回 `home` 与负向断言，禁止恢复 TopBar 入口、Home aux card、独立 screen、JobMatch generated client 或后端推荐链路 | 防止历史完成态计划被 `/implement` 重新拾取；后续若产品重新引入岗位推荐，必须先更新 product-scope 与 `ui-design/`，不能复用本历史口径直接开发 |
| D-2 | Parse loading 进度驱动 | 前端只通过 generated API client 调 `getTargetJob` 轮询 backend 返回的 `analysisStatus` 状态机：`queued` → `processing` → `ready` / `failed`；UI 4 步进度条与 footer 只源级复刻 `ui-design` 的解析节奏与版式，不代表前端直接调用 LLM；2026-05-22 plan 001 增加 `VITE_EI_API_MODE=real` generated-client gate，证明 production bootstrap 对 TargetJobs/import/parse operation 指向真实 backend base URL；2026-05-24 regression gate 固化：即使首次 `getTargetJob` 已返回 `ready`，正式前端也必须先展示 `ui-design` 4 步 loading 演示并按 tick 完成后再进入 preview，禁止直接跳过 loading 到 parsed preview；同一 mounted `ParseScreen` 收到新的 `targetJobId` 时必须清空旧 preview/edit state，回到 loading gate，并在 loading 完成后 hydrate 新 TargetJob | 当 fixture transport / backend 返回 `ready` 时进度条以可观察节奏快速完成但不可被跳过；返回 `failed` 时切错误态而非伪装继续；前端不得接入 AI provider、prompt registry、LLM key 或任何 provider-specific endpoint；真实 backend semantics 由 backend-targetjob E2E.P0.010-P0.013 配对证明 |
| D-3 | Hidden signals 来源 | 前端只展示 backend/API 返回的 `TargetJobSummary.interviewHypotheses`（对象级 `provenance` 必须存在）+ `TargetJobSummary.coreThemes`；`fitSummary.riskSignals` 用于 "WHERE IT'S A STRETCH" 类风险呈现；结构与 icon/置信度 tag 必须与 `ui-design` Hidden signals 卡片一致 | 不在前端凭 JD 文本推断、补写、改写或重新生成 hidden signals；所有 AI-generated 字段必须通过 OpenAPI fixture / backend response 可追溯到 `GenerationProvenance` |
| D-4 | Parse Launch 跳转契约 | `仅保存规划` 使用 `nav("workspace", interviewContextFromTargetJob(targetJob, { resumeId }))`；`立即面试` 使用同一 context 并附加 `autoStartPractice=1`，由 workspace 创建 session 后进入 practice；完整参数为 `targetJobId`、`jobId`、`jdId`、`planId`、真实 `resumeId`、`roundId`、`roundName`、`mode`、`modality`、`practiceMode` 等安全字段；`planId` 由 D1 `eiCreateInterviewContext` 等价契约从 `targetJobId` 推导（`plan-${targetJobId}`） | 真实 `createPracticePlan` / `startPractice` API 调用由 `frontend-workspace-and-practice` 承接；本 subspec 不直接创建 session，但必须传递真实 `resumeId` |
| D-5 | i18n locale 拆分 | 当前只保留 `home.*` 与 `parse.*` active 文案；历史 `jdMatch.*` 命名空间已随 D-17 删除，`nav.jd_match` 词条必须 0 命中 | 与 D1 D-7 i18n 规则一致；新增 active 命名空间需通过 D1 typed helper test |
| D-6 | Auth gate 触发点 | Paste/Upload/URL 三种 import 提交在未登录时触发 `requestAuth({ type: "import_jd", route: "home", params: { pendingImportId, source, resumeId }, label })`；`pendingImportId` 只引用当前 SPA 会话内存中的待提交 source payload，不包含 JD 原文或 source URL；首页未取得用户显式选择的 ready `resumeId` 时不得提交 import；Parse preview 若已收到首页显式选择的 `resumeId`，可把它作为当前绑定简历，但仍必须拒绝缺失或无效的 `resumeId`；具备真实 `resumeId` 的启动动作使用 `requestAuth({ type: "start_practice", route: "workspace", params: { ...context, autoStartPractice: "1" } })` | 已登录用户直接执行；未登录跳 `auth_login`，登录后恢复目标 route 与 params；import 恢复先回 home 自动提交，再跳 parse；与 frontend-shell C-2 一致 |
| D-7 | Privacy 红线 | JD 原文（rawText / rawDescription / sourceUrl）不进入 logger 字符串、URL query、localStorage、telemetry payload；只通过 generated client request body 与 React state 传递；fixture redact lint 必须覆盖 | 与 product-scope spec §1.6 隐私默认保守一致；observability redact rule 已在 D1 接入 |
| D-8 | jd_match → parse 反向数据流 | **Retired**：D-17 后不存在从岗位推荐进入 parse 的 active 出口；`source=jd_match` 与 `sourceJobMatchId` 只允许出现在历史文档、负向断言或旧数据迁移说明中 | Parse 当前只接收 Home import 传入的 `targetJobId` / `source` / `resumeId` |
| D-9 | Watchlist 与 Saved Searches 持久化策略 | **Retired**：Watchlist 与 Saved Searches 不再是当前产品模块，不得新增前端持久化、后端表、OpenAPI operation 或 fixture | 避免为已删模块保留 parallel contract surface |
| D-10 | Agent scan 状态来源 | **Retired**：`getAgentScanStatus`、scan badge、定时扫描与招聘源抓取均不是当前 scope | 当前前端不得接入 SSE/WebSocket/scan polling 或后台推荐状态 |
| D-11 | jd_match 隐私红线扩展 | **Retired**：query / saved-search label / watchlist label / sourceJobUrl / freeNote 等 jd_match 字段不得作为 active UI/API 输入出现；仅保留负向搜索约束，防止残留进入 URL / localStorage / telemetry | 当前隐私红线聚焦 JD 原文、source URL、resumeId 与 TargetJob provenance |
| D-12 | Search loading 形态 | **Retired**：旧 `screen-jd-match.jsx::SearchTab` 与 5 步 AGENT panel 已随 D-17 删除；任何搜索 loading 视觉不得从该历史原型复活 | 新视觉方向必须先改当前 `ui-design/`，再进入正式 frontend |

## 4 设计约束

- 视觉与交互必须以当前 active 的 `ui-design/src/screen-home.jsx`、`ui-design/src/screens-p0-complete.jsx::ParseScreen`、`ui-design/src/primitives.jsx` 为唯一真理源进行源级复刻；DOM 构图、控件类型、icon、菜单/弹层层级、aria 状态、主要交互路径必须可追溯到对应 jsx 函数；不得二次设计或重新解释视觉
- Parse 屏的 4 步 loading 文案、模型/rubric/prompt hash footer 的 DOM 构图、节奏、层级与可见文案必须与 `screens-p0-complete.jsx::ParseScreen` lines 10-104 一致；正式前端只能把这些值作为 backend parse metadata / fixture metadata 的展示，不得因此接入前端 LLM 调用或 provider 配置；任何视觉或文案修改必须先改 `ui-design/` 真理源
- `MockInterviewCard` 的 status pill / round rail / company meta slot / 标题 / 地点 DOM 必须与 `screen-home.jsx::MockInterviewCard` lines 148-216 一致；数据只能来自 generated `TargetJob` schema：company meta slot 显示 `companyName · status-derived label`，`statusTone` 从 `TargetJob.status` 派生，round rail 使用本 plan 明确的 P0 默认轮次与 currentIndex fallback；不得从 `ui-design/src/data.jsx`、未声明 fixture 字段或本地 mock 补 `level` / `nextRound` / `statusTone`
- `JDAssistModal` 的 upload / URL 双模态、关闭按钮、Continue/Cancel actions 必须与 `screen-home.jsx::JDAssistModal` lines 218-262 一致
- 所有 import source variants 必须通过 generated `ImportTargetJobRequest` schema 提交；`type` discriminator + 必填字段在前端 form-level 校验
- Parse 屏 hit/partial/gap toggle 状态在前端是 ephemeral UI state；用户保存或启动时不写回 `TargetJobRequirement.evidenceLevel`；evidenceLevel、summary、fitSummary、hidden signals 均是 backend/API 返回的 AI-generated 只读字段，前端不得用本地规则或 LLM 重新推断
- 保存规划 / 立即面试时调用 `updateTargetJob` 写回用户编辑的 title/companyNameHint/locationText/notes；不在 Parse 屏直接调用 `createPracticePlan`
- listTargetJobs 必须按 `updatedAt desc` 取最近 N 条；首页只展示前 3 条，超过 3 条时通过“更多”进入 `workspace` 模拟面试列表页；当前 generated client 可继续通过 `RequestOptions.query.pageSize=12` 预取以判断是否存在更多，服务端 pagination 由 generated client `cursor` 接力
- `createUploadPresign` / `importTargetJob` / `updateTargetJob` 都是 side-effect operation，前端必须通过 generated client 传 `idempotencyKey`，并在测试中断言 `Idempotency-Key` header 存在
- Job Picks / POST-INTERVIEW aux cards 属于退役入口，必须在 Home DOM 与 pixel parity 中 0 命中；旧 `/jd-match` 或 `jd_match` route key 只允许归一到 `home`
- i18n 必须支持 zh/en；初始 UI 语言跟随浏览器 locale，未知或缺失 fallback `en`；与 D1 D-7 一致
- 暗色 / customAccent 必须在 home + parse 两屏均通过 root level `data-theme` / `data-mode` / `data-custom-accent` 切换生效，不允许硬编码颜色
- Pixel parity gate 必须在 desktop (1440×900) + mobile (390×844) 两个 viewport 下断言 home + parse 的 DOM 锚点 / 关键 computed style / bounding box / 截图差异；任何 parity 失败必须修到与 `ui-design/` 一致或先修订真理源
- Mobile 响应式：home 与 parse 在 ≤768px viewport 下不能溢出视口；parse 屏 Requirements 双列在 mobile 折叠为单列；MockInterviewCard 网格 `repeat(auto-fill, minmax(320px, 1fr))` 在 mobile 自然成单列
- 所有可见用户字段（status pill / round rail / requirements label）必须有 `data-testid`，遵循 D1/D2 既有命名（`home-recent-mock-card-${id}` / `parse-requirement-${kind}-${idx}` 等）

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| frontend home / parse | `frontend-home-job-picks-and-parse` | Home / Parse React 组件、路由壳业务内容、i18n、源级 parity 测试 |
| App shell / route normalization / requestAuth / locale helper / generated client / fixture transport bootstrap | `frontend-shell` | D1 已交付，本 subspec 直接消费 |
| TargetJobs OpenAPI 契约 | `openapi-v1-contract` | `importTargetJob` / `listTargetJobs` / `getTargetJob` / `updateTargetJob` schema 与 fixture |
| Mock transport / fixture-backed response | `mock-contract-suite` | 本 subspec 通过 generated client mock transport 消费 fixture；fixture-backed UI variants 用于稳定覆盖 DOM / failure / privacy，不再作为真实 backend 完成证据 |
| TargetJob persistence / runner / event 发射 | `backend-targetjob` + `event-and-outbox-contract` + `db-migrations-baseline` | 真实 backend handler / store / event 已由 `backend-targetjob/001-targetjob-import-and-parse-bootstrap` 完成；plan 001 用 `targetJob.realApiMode.test.ts` + backend E2E.P0.010-P0.013 配对证明 frontend generated client 与真实 TargetJobs backend 对齐 |
| AI parsing 调用 | `ai-provider-and-model-routing` + `prompt-rubric-registry` | 真实 LLM 调用通过 backend；本 subspec 不直接消费 AI |
| retired jd_match / Job Picks / JobMatch / debrief entry cleanup | `product-scope/001-core-loop-module-pruning` + 本 subspec 历史计划 | 只保留历史记录、legacy route 归一和负向搜索；不得再作为 active frontend/backend/API owner |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | Home 默认渲染 | 用户未登录或已登录，无 TargetJob | 打开 `home` 路由 | Hero / textarea card / upload + URL 入口 / 选择已有简历 / Resume create CTA / empty state 完整渲染；不渲染旧 `home.heroSub`；TopBar 高亮 home；i18n zh/en 文案均通过 typed helper；CSS variables `--ei-bg`/`--ei-ink`/`--ei-accent`/`data-mode` 切换生效；DOM 锚点能追溯到 `ui-design/src/screen-home.jsx` | 001-home-jd-import-and-parse |
| C-2 | Paste JD → Parse | 用户在 Home 输入框粘贴 JD 文本，`listResumes` 返回 ready 简历 | 显式选择一份已有简历后点击「立即面试」 | 调用 `importTargetJob`（`source.type=manual_text`、`source.rawText=JD 文本`、`targetLanguage=当前 UI locale`）；成功后路由跳 `parse?targetJobId=…&resumeId=<selected ready resume id>`；未选择 ready 简历时按钮 disabled，不调用 import；Parse 屏先显示 4 步 loading，再轮询 `getTargetJob` 至 `analysisStatus=ready` 切到 preview；preview 渲染 fixture 中 title/companyName/locationText/requirements/summary.interviewHypotheses/fitSummary.riskSignals，并继承首页显式选择的真实 `resumeId`；JD 原文不写入 URL/localStorage/telemetry | 001-home-jd-import-and-parse |
| C-3 | Upload / URL 双 source variant | 用户在 Home 点击 upload 或 URL 入口 | 在 JDAssistModal 中确认 | Upload 路径先调用 generated `createUploadPresign`（`purpose=target_job_attachment`，带 `Idempotency-Key`），fixture 返回 `fileObjectId` 后再提交 `importTargetJob` `source.type=file`；URL 路径提交 `source.type=url`（`url` 字段）；后续流程与 C-2 一致；DOM 与 `screen-home.jsx::JDAssistModal` 行为一致（关闭、Continue、Cancel）；frontend tests 不真实上传二进制到 object storage | 001-home-jd-import-and-parse |
| C-4 | Recent mock interviews 列表 | 用户已登录，listTargetJobs 返回 N 条 TargetJob | 进入 home | 渲染最多 3 张 `MockInterviewCard`，按 `updatedAt desc` 排序；超过 3 条时渲染 `更多` CTA，点击跳转 `workspace` 模拟面试列表页；卡片显示 `companyName · status-derived label` / title / locationText / status pill（statusTone 从 `TargetJob.status` 派生）/ MiniRoundRail P0 fallback 当前轮次圆点；点击卡片调 `nav("workspace", interviewContextFromTargetJob(targetJob))`，默认补齐 `targetJobId / jobId / planId / jdId / resumeId / roundId / roundName`；列表为空时显示 `HomeEmptyState`，`回到 JD 输入` 按钮 focus textarea | 001-home-jd-import-and-parse |
| C-5 | Parse 编辑与保存/启动 | 用户在 Parse preview 编辑 OpenAPI 当前允许保存的 title / company / location / notes 字段，并切换若干 hit toggle | 选择 ready 简历后点击「仅保存规划」或「立即面试」 | 调用 `updateTargetJob`（仅 supplied fields，例：`titleHint` / `companyNameHint` / `locationText` / `notes`，带 `Idempotency-Key`）；level / language 槽位按 `ui-design` DOM 展示但为 read-only，直到 B2 扩展 `UpdateTargetJobRequest`；hit toggle 不写后端；「仅保存规划」成功后路由跳 `workspace?targetJobId=&jobId=&jdId=&planId=&resumeId=&roundId=&roundName=`；「立即面试」成功后跳 `workspace` 并携带 `autoStartPractice=1`，由 workspace 创建 session 后进入 `practice`；两个成功出口均必须携带真实 ready `resumeId`，不得写入 `resume-unbound`；任何 `updateTargetJob` 4xx 显示 inline 错误并保留编辑态；Cancel 跳 `home`；Re-parse 重置 `stage=loading` 并重新轮询 `getTargetJob` | 001-home-jd-import-and-parse |
| C-6 | Parse 失败态 | `getTargetJob.analysisStatus=failed` 或 polling 超时 | 用户在 Parse loading 阶段等待 | 切到 error state：显示「JD 解析失败」标题 + 失败原因 + 重新解析按钮 + 返回首页按钮；不展示伪造的 preview 数据；测试通过 fixture variant 锁定 | 001-home-jd-import-and-parse |
| C-7 | Auth pending action 恢复 | 未登录用户在 Home paste JD 并提交 / 已具备 verified ready `resumeId` 的启动动作需要登录恢复 | 进入 auth_login 完成登录 | 登录成功后回到原 route 与 params：paste/upload/url import 流先回到 Home，通过 opaque `pendingImportId` 消费当前 SPA 会话内存中的待提交 source payload 并自动重新发起 `importTargetJob`，成功后跳 Parse；pending route params 不携带 JD 原文、source URL 或 rawDescription；Parse 阶段未读取到 ready 简历时不得产生成功 pendingAction；如启动动作已有真实 `resumeId`，pending route 必须恢复到 `workspace` 并携带 `autoStartPractice=1` 与完整 interview context；与 `frontend-shell` C-2 一致 | 001-home-jd-import-and-parse |
| C-8 | Retired Job Picks / jd_match 负向 gate | 用户打开 Home、TopBar、旧 `/jd-match` URL 或旧 `jd_match` route key | App normalize / render | Home 不渲染 `home-aux-jobpicks`；TopBar 不渲染 `topbar-nav-jd_match`；旧 `/jd-match` / `jd_match` 归一到 `home`；`frontend/src/app/screens/jd_match/` 与 `frontend/tests/pixel-parity/jd_match.spec.ts` 0 命中 | 002 removal phase + product-scope pruning |
| C-9 | UI source structure parity | D1+D2+D3 已交付，新增/维护 home / parse active shell | Vitest+jsdom 测试 | DOM 锚点、控件类型（textarea / button / modal / 自定义弹层）、icon name、aria 状态、主要交互路径必须能追溯到 `ui-design/src/screen-home.jsx` / `screens-p0-complete.jsx::ParseScreen`；旧 prototype 中存在但当前真理源已移除的 testid / control 类型负向断言不命中；任何 parity 失败必须修到与原型一致或先修订 `ui-design/` 真理源 | 001-home-jd-import-and-parse |
| C-10 | UI visual geometry parity | C-9 通过 | Playwright 在 desktop (1440×900) + mobile (390×844) 双 viewport 下加载 `frontend/dist` home 与 parse 路由 | 关键区块 bounding box 不重叠且 stays in viewport；warm/light + dark + customAccent 三态切换关键元素 computed background / color 出现可见变化；mobile viewport 下 Requirements 双列折叠为单列、textarea card 不溢出；新增 `tests/pixel-parity/home.spec.ts` 与 `tests/pixel-parity/parse.spec.ts`；与 D2/D3 现有 21 个 spec 累加；CI / 本地 `pnpm --filter @easyinterview/frontend test:pixel-parity` 通过 | 001-home-jd-import-and-parse |
| C-11 | Privacy 红线 | 用户提交 JD 原文 | observability / log / URL / localStorage / telemetry 输出 | JD raw text / rawDescription / sourceUrl 不出现在 console.log、不出现在 URL query、不写入 localStorage、不进入任何 telemetry payload；前端 redact lint 反查通过；fixture transport 不在 mockTransport 日志中泄漏 raw 内容 | 001-home-jd-import-and-parse |
| C-12 | Retired JobMatch API 零消费 | D-17 删除完成 | 运行 OpenAPI/generated client、fixture、frontend source 负向搜索 | `getJobMatchProfile` / `getAgentScanStatus` / `listJobRecommendations` / `searchJobs` / `listWatchlist` 等 JobMatch operation 不作为 active frontend consumer 出现；fixture 与场景目录不得被当前 plan discovery 引用 | 002 removal phase + product-scope pruning |
| C-13 | Retired JobMatch side-effect 零恢复 | D-17 删除完成 | 搜索前端 action、auth pendingAction、route params | `addToWatchlist` / `removeFromWatchlist` / `markJobNotRelevant` / `sourceJobMatchId` / `jd_match_action` 不得作为 active side-effect 或登录恢复类型出现 | 002 removal phase + product-scope pruning |
| C-14 | Retired Search / Agent scan 零恢复 | D-17 删除完成 | 搜索 UI / i18n / timer / network 代码 | `screen-jd-match`、5 步 AGENT panel、saved search、watchlist、market signal、scan polling 等仅允许出现在历史文档或负向断言中；active `frontend/src` 0 命中 | 002 removal phase + product-scope pruning |
| C-15 | Retired jd_match pending action 零恢复 | D-17 删除完成 | 搜索 pendingAction 类型与 URL params | 当前登录恢复只覆盖 Home import 与 Parse start practice；不得恢复 `jd_match` side-effect auto-resume | 002 removal phase + product-scope pruning |
| C-16 | Retired Watchlist / Market signals 零恢复 | D-17 删除完成 | 搜索 fixture、schema、screen 与 scenario | watchlist / market signal 只作为历史计划记录；active source、OpenAPI operation matrix 与 context discovery 不得引用 | 002 removal phase + product-scope pruning |

## 7 关联计划

- [001-home-jd-import-and-parse](./plans/001-home-jd-import-and-parse/plan.md) — 当前 owner：Home + Parse 端到端；历史 Phase 5 / P0.017 jd_match placeholder 自 D-17 起只作为退役记录。2026-05-22 L2 remediation 补 TargetJobs/upload/import/parse real-mode generated-client gate + backend E2E.P0.010-P0.013 配对证据；2026-05-24 regression remediation 固化 ready 响应也必须先展示 `ui-design` loading 演示，并把 P0.016 Confirm → Workspace browser gate 升级为 7 字段 route/context + `workspace-missing-resume` screenshot marker；同日补 same-route `targetJobId` switch regression，防止已 mounted Parse preview 继续显示旧 TargetJob（completed 2026-05-24）

## 8 关联文档

- 上游 spec：[`engineering-roadmap`](../engineering-roadmap/spec.md)、[`product-scope`](../product-scope/spec.md)、[`frontend-shell`](../frontend-shell/spec.md)、[`openapi-v1-contract`](../openapi-v1-contract/spec.md)、[`mock-contract-suite`](../mock-contract-suite/spec.md)
- UI 真理源：`ui-design/src/screen-home.jsx`、`ui-design/src/screens-p0-complete.jsx::ParseScreen`、`ui-design/src/app.jsx`、`ui-design/src/primitives.jsx`、[`docs/ui-design/jd-resume-management.md`](../../ui-design/jd-resume-management.md)、[`docs/ui-design/ui-architecture.md`](../../ui-design/ui-architecture.md)、[`docs/ui-design/module-job-workspace.md`](../../ui-design/module-job-workspace.md)、[`docs/ui-design/removed-modules-and-scope.md`](../../ui-design/removed-modules-and-scope.md) §15（旧 `ui-design/src/screen-jd-match.jsx` 已随 2026-06-12 第二批裁剪删除）

## 9 D-17 前端删除范围与零残留验收（plan 002 active scope）

### 9.1 删除范围

| 资产 | 处置 |
|------|------|
| `frontend/src/app/screens/jd_match/` 全目录（screens / tabs / hooks / 子组件 / 测试） | 删除 |
| `routes.ts` `PRIMARY_NAV_ROUTES` 中 `jd_match` 项与 route 定义、`routeUrl.ts` `/jd-match` path 与 `JD_MATCH_SAFE`、`App.tsx` 渲染分支、`TopBar` `NAV_LABEL_KEYS` / `NAV_ICONS` 条目 | 删除；`jd_match` route key 与 `/jd-match` path 归一回 `home`（normalize alias + legacy path），一级导航收敛为 `home / workspace / resume_versions` 三项 |
| Home 屏 `JOB PICKS` aux card（→ jd_match 入口）与 `POST-INTERVIEW` aux card（→ debrief 入口） | 删除；当前 `ui-design/src/screen-home.jsx` 不再保留这两个 aux cards |
| i18n `jdMatch.*` 命名空间与 `nav.jd_match` 词条（zh/en） | 删除 |
| `frontend/tests/pixel-parity/jd_match.spec.ts` 与 topbar parity 五入口 golden 断言 | 删除 / 改为当前三入口断言 |
| jd_match 相关 Vitest（unit / scenario / 路由测试中的 jd_match 用例） | 删除或改写为负向断言 |
| `test/scenarios/e2e/p0-017 / p0-027..031` 前端 jd_match 场景目录与 INDEX 行 | 删除 |
| `frontend/src/lib/jobs/jobs.ts` 等生成常量中的 jd_match 条目 | 随 shared codegen 再生成（上游删除归 backend plan 001 Phase 9） |

### 9.2 验收标准（v2.0 新增）

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-R1 | 三入口导航收敛 | D-17 / D-22 删除完成 | 打开 App / 切换语言 / 切换主题 | TopBar 一级导航仅 `home / workspace / resume_versions` 三项；`topbar-nav-jd_match` 与 `topbar-nav-debrief` 零命中；pixel parity 与当前 golden 预览一致 | 002 removal phase + product-scope pruning |
| C-R2 | 旧 route 归一 | 用户直开 `/jd-match` 或旧 `jd_match` route key（含 localStorage 残留） | App normalize / parse URL | 归一到 `home`，不渲染独立 jd_match 屏；`legacyRouteNegative` 断言 jd_match 不在 live route 目录 | 002 removal phase |
| C-R3 | 前端零残留 | 删除完成 | `rg -i "jd[-_]?match|job picks|post-interview|debrief"` 于 `frontend/src frontend/tests`（normalize alias / legacy path / 负向断言除外）；`pnpm test` / `typecheck` / `build` / `test:pixel-parity` | 零残留；全套件通过；Home 不保留 retired aux cards | 002 removal phase + product-scope pruning |

## 10 D-14 单次确认漏斗目标契约（plan 001 active scope）

以 `ui-design/src/screens-p0-complete.jsx::ParseScreen` 与 [docs/ui-design/module-job-workspace.md](../../ui-design/module-job-workspace.md) v1.10 为唯一 UI 真理源：

1. parse 确认页在解析 preview 基础上同时承载启动决策：轮次假设卡可点选确认 InterviewRound；绑定简历 pill（复用 ResumePickerModal）选择 / 更换简历；底部主操作为「立即面试」与「仅保存规划」。
2. 「立即面试」走 `requestAuth(start_practice)` 登录拦截，pendingAction 恢复到 `workspace` 并携带 `autoStartPractice=1`，由 workspace 创建 session 后进入 `practice`；「仅保存规划」进入 `workspace`。
3. 旧「确认并进入面试前确认」二次确认按钮删除；首次导入链路 parse 与 session 之间不存在平行确认页；`workspace` 定位为回访枢纽。
4. 详细 DOM 锚点、operation matrix 与验收行在 plan 001 v2.0 修订中固化（C-5 旧口径以本节为准失效）。

### 10.1 2026-06-30 简历绑定强制修订

首次 JD 导入后的 `parse` 解析确认页不得再把 `resume-unbound` 作为可启动或可保存规划的默认上下文。正式前端必须在用户离开 `parse` 前完成以下约束：

1. 进入 preview 后读取 `listResumes`，只允许选择 `parseStatus=ready` 且未归档的简历。
2. 有可用简历时不得默认选中任何一份；必须展示可选 ready 简历并保持 `立即面试` / `仅保存规划` disabled，直到用户显式选择一份简历。
3. 没有可用简历或读取失败时，`立即面试` 与 `仅保存规划` 必须禁用，并显示“创建简历”入口；入口导航到 `resume_versions?flow=create`，不得静默继续。
4. `立即面试` 必须携带真实 `resumeId` 进入 `workspace` 的 `autoStartPractice=1` 会话创建链路，并使用 `requestAuth(start_practice)` 保护登录恢复；`仅保存规划` 必须携带同一真实 `resumeId` 进入 `workspace`。
5. P0.016 的浏览器 gate 必须反向证明 `workspace-missing-resume` 不再是 Parse 确认主路径的成功状态。

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-17 | Parse 强制绑定简历 | `parse` preview 已就绪，`listResumes` 返回 ready 简历 | 用户选择简历后点击 `立即面试` 或 `仅保存规划` | 两个出口均携带真实 `resumeId`；`立即面试` 通过 `requestAuth(start_practice)` / `workspace autoStartPractice=1` 创建 session 后进入 `practice`；没有 ready 简历时两个出口禁用并引导创建简历；不得生成 `resume-unbound` 面试规划 | 001-home-jd-import-and-parse |

### 10.2 2026-07-06 首页新建规划快捷入口修订

首页的 JD 输入卡是“新建模拟面试规划”的快捷入口，不再把简历选择延后到 parse 才出现。正式前端必须在首页完成以下约束：

1. 删除旧说明文案 `home.heroSub`，首页 Hero 只保留 label 与标题。
2. 主按钮文案从「解析并确认面试」改为「立即面试」；英文从 `Parse & confirm interview` 改为 `Start interview now`。
3. JD 输入卡下方展示“选择已有简历”下拉框，读取 `listResumes`，只允许选择 `parseStatus=ready` 且未归档的简历；不得平铺全部简历为静态列表，也不得出现上传简历入口。
4. `还没有简历？1 分钟创建 →` 与“选择已有简历”并排，点击进入 `resume_versions?flow=create`。
5. 用户未显式选择 ready 简历时，`立即面试` disabled，不调用 `importTargetJob`，也不产生 pending import。
6. 用户显式选择 ready 简历后，paste / upload / URL import 成功路由到 `parse` 时必须携带真实 `resumeId`；parse 可继承该显式选择，但仍必须保留缺失或无效 resume 时的阻断与创建入口。
7. 2026-07-06 布局修订：简历下拉框宽度不得撑满整页，创建入口在下拉框右侧同一行；「立即面试」主按钮放在简历选择行下方，避免与 textarea/source 行混在一起。
8. 2026-07-06 整合修订：上传 JD 文件与 URL 导入必须回收到同一个 `home-jd-input-card` 底部，作为 `home-jd-source-controls` source actions；不得继续渲染独立 `home-upload-source-panel` 或双栏 `home-source-layout`。

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-18 | Home 预绑定简历启动 | Home JD 输入卡已加载，`listResumes` 返回 ready 简历 | 用户通过下拉框显式选择一份已有简历、粘贴 JD 后点击「立即面试」 | `importTargetJob` 成功后进入 `parse`，route params 带真实 `resumeId`；未选择简历时按钮 disabled 且 import 不发生；首页不展示旧 hero sub、不展示上传简历入口，不平铺全部简历；创建简历 CTA 进入 `resume_versions?flow=create` | 001-home-jd-import-and-parse |
| C-19 | Home 最近模拟面试收敛 | Home 已登录且 `listTargetJobs` 返回超过 3 条 TargetJob | 用户进入 Home | 最近模拟面试只显示最近 3 张卡片；`更多` CTA 可见并点击跳转到 `workspace` 模拟面试列表页；少于或等于 3 条时不需要额外列表展开；卡片排序仍按 `updatedAt desc` | 001-home-jd-import-and-parse |
| C-20 | Home 新建规划布局收敛 | Home JD 输入卡已加载，存在 ready 简历 | 用户查看新建模拟面试入口 | `home-resume-select` 下拉框使用适度宽度并与 `还没有简历？1 分钟创建 →` 同行水平对齐；`立即面试` 位于简历选择行下方，仍保持未选简历或未输入 JD 时 disabled；paste / upload / URL import 继续携带真实 `resumeId`；JD source actions 的当前容器归属由 C-21 约束 | 001-home-jd-import-and-parse |
| C-21 | Home JD source actions 输入卡内整合 | Home JD 输入卡已加载，存在 ready 简历 | 用户查看新建模拟面试入口 | `home-jd-input-card` 同时承载 `home-jd-textarea` 与底部 `home-jd-source-controls`；`home-upload-trigger` 与 `home-url-trigger` 位于输入卡内，独立 `home-upload-source-panel` / `home-source-layout` 0 命中；`home-resume-select` 与 `home-resume-create` 仍同行，`home-jd-submit` 仍在 `home-submit-row` 且不在输入卡内；paste / upload / URL import 继续携带真实 `resumeId` | 001-home-jd-import-and-parse |

## 11 修订记录

| 版本 | 日期 | 说明 |
|------|------|------|
| 2.8 | 2026-07-06 | 删除 JD Match 002 plan 实体引用；无争议废弃模块不再保留退役说明或 historical plan，删除证据由 product-scope/001 当前 owner 承接。 |
| 2.7 | 2026-07-06 | 第二轮 scope pruning reconcile：把本 subspec 当前 active 范围收敛为 Home + Parse，Job Picks / jd_match / JobMatch / POST-INTERVIEW / debrief 正向描述全部改为历史退役与负向 gate。 |
| 2.6 | 2026-07-06 | 修订首页 JD source actions：上传 JD 文件与 URL 导入回收到输入卡底部，删除独立 upload source panel，保留简历下拉框与主按钮布局。 |
| 2.5 | 2026-07-06 | 修订首页新建规划布局：粘贴 JD 与上传文件分区，简历下拉框定宽并与创建入口同行，主按钮移到简历选择下方。 |
| 2.4 | 2026-07-06 | 修订首页选择已有简历控件为下拉框，并把最近模拟面试收敛为 3 张卡片 + “更多”跳转到模拟面试列表页。 |
| 2.3 | 2026-07-06 | 修订首页新建模拟面试规划快捷入口：删除冗余 hero sub，主按钮改为「立即面试」，并在首页预先选择已有 ready 简历后才允许提交 JD import。 |
| 2.2 | 2026-06-30 | 修订 D-14 简历绑定：Parse 不得默认选中最新 ready 简历，用户必须显式选择后才能保存规划或启动面试。 |
| 2.1 | 2026-06-30 | 修订 D-14：Parse 解析确认页必须在 `立即面试` / `仅保存规划` 前绑定 ready 简历，禁止 `resume-unbound` 成为成功 handoff。 |
| 2.0 | 2026-06-13 | 对齐 product-scope v2.1：D-17 删除 jd_match 模块（新增 §9 删除范围与 C-R1~C-R3，jd_match 相关历史条目退役）；D-14 单次确认漏斗目标契约（新增 §10，plan 001 重开承接）；UI 真理源列表移除 screen-jd-match.jsx |
- 历史：[history.md](./history.md)
