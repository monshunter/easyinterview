# Frontend Home / Job Picks / Parse Spec

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-05-08

## 1 背景与目标

`frontend-home-job-picks-and-parse` 是 `engineering-roadmap` S1 工程蓝图中明确预占的前端业务 subspec，承接 `frontend-shell`（D1+D2+D3 视觉系统）已交付的 App 壳、TopBar、五个一级入口、route normalization、`requestAuth(pendingAction)` 与 fixture-backed mock transport，落地用户首次进入并产出一次模拟面试上下文的入口闭环。

本 subspec 的目标是把当前 `ui-design/` 静态原型中 `home`、`jd_match`、`parse` 三个屏幕从 mock data 迁移到正式前端工程，并通过 generated client + fixture-backed transport 消费已存在的 `TargetJobs` OpenAPI 契约，使「带着 JD 来的用户」能够在最短路径完成「粘贴/上传/URL 导入 JD → 解析确认 → 进入模拟面试规划」的 P0 主路径。

`frontend-shell/spec.md` §2.1 `parse` 路由壳与 `eiCreateInterviewContext` 等价契约由本 subspec 承接业务内容；`backend-targetjob`（handler / service / store）单独立项。

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
- Parse 屏（`route=parse`）：
  - Loading 阶段：4 步进度条 + footer 模型 / rubric / prompt hash 占位文案
  - Preview/Confirm 阶段：Basic fields 行内可编辑、Must Have / Nice to Have requirements 块带 hit/partial/gap toggle、Hidden signals 块（`TargetJobSummary.interviewHypotheses` 推断）、Round assumptions 4 卡
  - Footer actions：Cancel → `home`、Re-parse、Confirm → `workspace` 携带完整 `eiCreateInterviewContext` 等价 params
  - 通过 `analysisStatus` 状态机（`queued` / `processing` / `ready` / `failed`）驱动 loading→preview 切换；不假装"正在调用 LLM"
- JD Match 屏（`route=jd_match`）P1 placeholder shell：
  - 保留 TopBar nav 高亮 + 路由可达
  - Hero + Profile snapshot chip 静态渲染
  - 三 tab 显示「Coming in P1 · backend recommendations API 落地后启用」placeholder copy
  - 不消费任何 mock recommendations / search / watchlist 数据
- 与 D1 `requestAuth(pendingAction)` 集成：未登录用户提交 import 与 confirm interview 时触发 pendingAction，登录后恢复
- 与 D2/D3 `ui-design/` parity gate 集成：home / parse 两屏新增 Vitest+jsdom smoke 与 Playwright desktop+mobile pixel parity 测试

### 2.2 Out of Scope

- `jd_match` Recommended / Search / Watchlist 三 tab 完整业务内容、自然语言搜索、saved searches、market signals、agent active status — 由后续 plan `002-jd-match-recommendations` 在 backend recommendations API 落地后承接
- `recommendJobs` / `searchJobs` / `listWatchlist` / `getMarketSignals` 等 OpenAPI operationId — 当前未声明，本 subspec 不预埋前端 mock
- `workspace` 屏内业务（mock plan 状态、轮次切换、简历绑定、公司情报）— 由后续 `frontend-target-job-workspace` subspec 承接
- `practice` / `report` / `debrief` / `resume_versions` 屏业务 — 各自独立 subspec
- 真实 LLM 调用、JD 抓取、URL fetch、文件上传二进制处理 — 由 `backend-targetjob` 与 `backend-runtime-topology` 承接
- 数据库 schema、event/outbox、AI provider profile 接入 — 由 B4 / B3 / A3 承接
- 不新增旧 `welcome` / `growth` / `mistakes` / `drill` / `followup` / `experiences` / `star` / 独立 `voice` route 别名

## 3 用户决策 / 待确认事项

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | jd_match 业务范围 | 本 subspec 首个 plan 仅做 P1 placeholder shell；完整 Recommended / Search / Watchlist 三 tab 等 backend recommendations API 落地后由后续 plan `002-jd-match-recommendations` 承接 | 当前 OpenAPI 没有 `listJobRecommendations` / `searchJobs` / `listWatchlist` 契约；强行实现需大量纯前端 mock，违背 D1 D-5「不直接 import prototype data」 |
| D-2 | Parse loading 进度驱动 | 前端只通过 generated API client 调 `getTargetJob` 轮询 backend 返回的 `analysisStatus` 状态机：`queued` → `processing` → `ready` / `failed`；UI 4 步进度条与 footer 只源级复刻 `ui-design` 的解析节奏与版式，不代表前端直接调用 LLM | 当 fixture transport / backend 返回 `ready` 时进度条以可观察节奏快速完成；返回 `failed` 时切错误态而非伪装继续；前端不得接入 AI provider、prompt registry、LLM key 或任何 provider-specific endpoint |
| D-3 | Hidden signals 来源 | 前端只展示 backend/API 返回的 `TargetJobSummary.interviewHypotheses`（对象级 `provenance` 必须存在）+ `TargetJobSummary.coreThemes`；`fitSummary.riskSignals` 用于 "WHERE IT'S A STRETCH" 类风险呈现；结构与 icon/置信度 tag 必须与 `ui-design` Hidden signals 卡片一致 | 不在前端凭 JD 文本推断、补写、改写或重新生成 hidden signals；所有 AI-generated 字段必须通过 OpenAPI fixture / backend response 可追溯到 `GenerationProvenance` |
| D-4 | Confirm 跳转契约 | `nav("workspace", { targetJobId, jdId, planId, resumeVersionId, roundId })`；`planId` 由 D1 `eiCreateInterviewContext` 等价契约从 `targetJobId` 推导（`plan-${targetJobId}`） | 真实 `createPracticePlan` API 调用由 `frontend-target-job-workspace` 承接；本 subspec 不主动创建 PracticePlan |
| D-5 | i18n locale 拆分 | 在 `frontend/src/app/i18n/locales/zh.ts` / `en.ts` 中新增 `home.*`、`parse.*`、`jdMatch.*` 三个命名空间；不混入 messages.ts 类型聚合层 | 与 D1 D-7 i18n 规则一致；新增命名空间需通过 D1 typed helper test |
| D-6 | Auth gate 触发点 | Paste/Upload/URL 三种 import 提交在未登录时触发 `requestAuth({ type: "import_jd", route: "home", params: { pendingImportId, source }, label })`；`pendingImportId` 只引用当前 SPA 会话内存中的待提交 source payload，不包含 JD 原文或 source URL；Parse Confirm 在未登录时触发 `requestAuth({ type: "confirm_interview", route: "workspace", params: { targetJobId, jdId, planId } })` | 已登录用户直接执行；未登录跳 `auth_login`，登录后恢复目标 route 与 params；import 恢复先回 home 自动提交，再跳 parse；与 frontend-shell C-2 一致 |
| D-7 | Privacy 红线 | JD 原文（rawText / rawDescription / sourceUrl）不进入 logger 字符串、URL query、localStorage、telemetry payload；只通过 generated client request body 与 React state 传递；fixture redact lint 必须覆盖 | 与 product-scope spec §1.6 隐私默认保守一致；observability redact rule 已在 D1 接入 |

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
| Mock transport / fixture-backed response | `mock-contract-suite` | 本 subspec 通过 generated client mock transport 消费 fixture |
| TargetJob persistence / runner / event 发射 | `backend-targetjob`（未来）+ `event-and-outbox-contract` + `db-migrations-baseline` | 真实 backend handler / store / event 实现，本 subspec 不依赖真实 backend |
| AI parsing 调用 | `ai-provider-and-model-routing` + `prompt-rubric-registry` | 真实 LLM 调用通过 backend；本 subspec 不直接消费 AI |
| jd_match 业务三 tab | `frontend-home-job-picks-and-parse/plans/002-jd-match-recommendations`（后续）+ `backend-jobs-recommendations`（未来） | placeholder shell 在本 plan，业务在后续 plan |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | Home 默认渲染 | 用户未登录或已登录，无 TargetJob | 打开 `home` 路由 | Hero / textarea card / upload + URL 入口 / Job Picks aux card / Post-interview aux card / Resume create CTA / empty state 完整渲染；TopBar 高亮 home；i18n zh/en 文案均通过 typed helper；CSS variables `--ei-bg`/`--ei-ink`/`--ei-accent`/`data-mode` 切换生效；DOM 锚点能追溯到 `ui-design/src/screen-home.jsx` | 001-home-jd-import-and-parse |
| C-2 | Paste JD → Parse | 用户在 Home 输入框粘贴 JD 文本 | 点击「解析并确认面试」 | 调用 `importTargetJob`（`source.type=manual_text`、`source.rawText=JD 文本`、`targetLanguage=当前 UI locale`）；成功后路由跳 `parse?targetJobId=…`；Parse 屏先显示 4 步 loading，再轮询 `getTargetJob` 至 `analysisStatus=ready` 切到 preview；preview 渲染 fixture 中 title/companyName/locationText/requirements/summary.interviewHypotheses/fitSummary.riskSignals；JD 原文不写入 URL/localStorage/telemetry | 001-home-jd-import-and-parse |
| C-3 | Upload / URL 双 source variant | 用户在 Home 点击 upload 或 URL 入口 | 在 JDAssistModal 中确认 | Upload 路径先调用 generated `createUploadPresign`（`purpose=target_job_attachment`，带 `Idempotency-Key`），fixture 返回 `fileObjectId` 后再提交 `importTargetJob` `source.type=file`；URL 路径提交 `source.type=url`（`url` 字段）；后续流程与 C-2 一致；DOM 与 `screen-home.jsx::JDAssistModal` 行为一致（关闭、Continue、Cancel）；frontend tests 不真实上传二进制到 object storage | 001-home-jd-import-and-parse |
| C-4 | Recent mock interviews 列表 | 用户已登录，listTargetJobs 返回 N 条 TargetJob | 进入 home | 渲染最多 12 张 `MockInterviewCard`，按 `updatedAt desc` 排序；卡片显示 `companyName · status-derived label` / title / locationText / status pill（statusTone 从 `TargetJob.status` 派生）/ MiniRoundRail P0 fallback 当前轮次圆点；点击卡片调 `nav("workspace", interviewContextFromTargetJob(targetJob))`，默认补齐 `targetJobId / jobId / planId / jdId / resumeVersionId / roundId / roundName`；列表为空时显示 `HomeEmptyState`，`回到 JD 输入` 按钮 focus textarea | 001-home-jd-import-and-parse |
| C-5 | Parse 编辑与 Confirm | 用户在 Parse preview 编辑 OpenAPI 当前允许保存的 title / company / location / notes 字段，并切换若干 hit toggle | 点击「确认并进入面试前确认」 | 调用 `updateTargetJob`（仅 supplied fields，例：`titleHint` / `companyNameHint` / `locationText` / `notes`，带 `Idempotency-Key`）；level / language 槽位按 `ui-design` DOM 展示但为 read-only，直到 B2 扩展 `UpdateTargetJobRequest`；hit toggle 不写后端；成功后路由跳 `workspace?targetJobId=&jdId=&planId=&resumeVersionId=&roundId=`，使用 D1 `eiCreateInterviewContext` 等价契约从 `targetJobId` 推导默认值；任何 `updateTargetJob` 4xx 显示 inline 错误并保留编辑态；Cancel 跳 `home`；Re-parse 重置 `stage=loading` 并重新轮询 `getTargetJob` | 001-home-jd-import-and-parse |
| C-6 | Parse 失败态 | `getTargetJob.analysisStatus=failed` 或 polling 超时 | 用户在 Parse loading 阶段等待 | 切到 error state：显示「JD 解析失败」标题 + 失败原因 + 重新解析按钮 + 返回首页按钮；不展示伪造的 preview 数据；测试通过 fixture variant 锁定 | 001-home-jd-import-and-parse |
| C-7 | Auth pending action 恢复 | 未登录用户在 Home paste JD 并提交 / 在 Parse 点击 Confirm | 进入 auth_login 完成登录 | 登录成功后回到原 route 与 params：paste/upload/url import 流先回到 Home，通过 opaque `pendingImportId` 消费当前 SPA 会话内存中的待提交 source payload 并自动重新发起 `importTargetJob`，成功后跳 Parse；pending route params 不携带 JD 原文、source URL 或 rawDescription；Confirm 流回到 workspace 携带 interview context；与 `frontend-shell` C-2 一致 | 001-home-jd-import-and-parse |
| C-8 | jd_match P1 placeholder shell | 用户点击 TopBar Job Picks 或 home aux card「打开岗位推荐」 | 路由进入 `jd_match` | 渲染 hero + profile snapshot chip 静态版本 + 三 tab 标签（推荐 / 联网搜索 / 关注列表）；tab 内容区显示 "Coming in P1" 占位文案 + 关联 plan 002 引用；TopBar 高亮 jd_match；DOM 保留主要锚点（`jdmatch-hero` / `jdmatch-tab-${k}` / `jdmatch-placeholder`）便于后续 plan 002 接力；不展示任何 mock recommendation 数据 | 001-home-jd-import-and-parse |
| C-9 | UI source structure parity | D1+D2+D3 已交付，新增 home / parse / jd_match shell | Vitest+jsdom 测试 | DOM 锚点、控件类型（textarea / button / modal / 自定义弹层）、icon name、aria 状态、主要交互路径必须能追溯到 `ui-design/src/screen-home.jsx` / `screens-p0-complete.jsx::ParseScreen` / `screen-jd-match.jsx`（仅 hero+tab 部分）；旧 prototype 中存在但当前真理源已移除的 testid / control 类型负向断言不命中；任何 parity 失败必须修到与原型一致或先修订 `ui-design/` 真理源 | 001-home-jd-import-and-parse |
| C-10 | UI visual geometry parity | C-9 通过 | Playwright 在 desktop (1440×900) + mobile (390×844) 双 viewport 下加载 `frontend/dist` home 与 parse 路由 | 关键区块 bounding box 不重叠且 stays in viewport；warm/light + dark + customAccent 三态切换关键元素 computed background / color 出现可见变化；mobile viewport 下 Requirements 双列折叠为单列、textarea card 不溢出；新增 `tests/pixel-parity/home.spec.ts` 与 `tests/pixel-parity/parse.spec.ts`；与 D2/D3 现有 21 个 spec 累加；CI / 本地 `pnpm --filter @easyinterview/frontend test:pixel-parity` 通过 | 001-home-jd-import-and-parse |
| C-11 | Privacy 红线 | 用户提交 JD 原文 | observability / log / URL / localStorage / telemetry 输出 | JD raw text / rawDescription / sourceUrl 不出现在 console.log、不出现在 URL query、不写入 localStorage、不进入任何 telemetry payload；前端 redact lint 反查通过；fixture transport 不在 mockTransport 日志中泄漏 raw 内容 | 001-home-jd-import-and-parse |

## 7 关联计划

- [001-home-jd-import-and-parse](./plans/001-home-jd-import-and-parse/plan.md) — Home + Parse 端到端 + jd_match P1 placeholder shell
- 002-jd-match-recommendations（保留编号；启动条件：backend recommendations API 落地、`listJobRecommendations` / `searchJobs` / `listWatchlist` operationId 进入 OpenAPI 并有 fixture）

## 8 关联文档

- 上游 spec：[`engineering-roadmap`](../engineering-roadmap/spec.md)、[`product-scope`](../product-scope/spec.md)、[`frontend-shell`](../frontend-shell/spec.md)、[`openapi-v1-contract`](../openapi-v1-contract/spec.md)、[`mock-contract-suite`](../mock-contract-suite/spec.md)
- UI 真理源：`ui-design/src/screen-home.jsx`、`ui-design/src/screen-jd-match.jsx`、`ui-design/src/screens-p0-complete.jsx`、`ui-design/src/app.jsx`、`ui-design/src/primitives.jsx`、[`docs/ui-design/jd-resume-management.md`](../../ui-design/jd-resume-management.md)、[`docs/ui-design/ui-architecture.md`](../../ui-design/ui-architecture.md)、[`docs/ui-design/module-job-workspace.md`](../../ui-design/module-job-workspace.md)
- 历史：[history.md](./history.md)
