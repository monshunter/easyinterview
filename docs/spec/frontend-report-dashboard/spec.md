# Frontend Report Dashboard Spec

> **版本**: 1.28
> **状态**: completed
> **更新日期**: 2026-07-15

## 1 背景与目标

`frontend-report-dashboard` 承接当前面试规划的独立报告列表、一次连续模拟面试结束后的诚实生成态与证据化报告。列表帮助用户按当前规划的 canonical rounds 找到当前可用报告与最新生成状态；详情帮助用户理解总体准备度、能力维度、优势/风险证据和下一步建议，并从当前轮复练或进入下一轮。

报告不按题目组织，也不展示伪精确分数。前端只展示服务端持久化的最终语义，不自行拆分维度、推断 readiness、改写证据或生成建议。

## 2 范围

### 2.1 In Scope

- `reports`：受保护的独立页面，唯一上下文参数为 `targetJobId`；读取当前 TargetJob 与 `listTargetJobReports(targetJobId)`，只展示该规划 canonical rounds 的 `currentReport/latestAttempt`，不做全局中心或完整版本历史。
- `reports` 的 loading / empty / error / ready 状态彼此完备；target/round identity 漂移、跨规划响应和 stale request 均 fail closed，不渲染其他规划 sentinel 或错链。
- `generating`：轮询真实 report status，展示诚实的异步等待说明；不伪造百分比、实时观察或通知订阅。
- `report`：Header、四项 Context Strip（目标岗位 / 轮次 / 可链接简历副本 / 面试记录）、两项 Summary Metrics、两行各两个常驻内容区（Dimensions / Strength Evidence / Risks / Next Actions），以及底部一个全宽 Overall Summary。
- `report-conversation`：仅以 `reportId` 读取报告附属的 ordered user/assistant transcript，安全渲染 Markdown/GFM，并返回同一报告状态页；queued/generating/ready/failed 都可访问。
- Overall Summary 使用“面试总评”标题，同时展示 localized readiness tier 与服务端 `summary`；二者不得继续出现在顶部指标区，`summary` 全页只展示一次。
- Dimension 使用动态 `label`，status/confidence 走完整 zh/en i18n，不泄漏 raw enum/code。
- Header 保留唯一一对 CTA；`nextActions[0].type` 决定现有按钮主次，不新增 CTA。
- Replay focus 由后端 source report 投影；URL/前端 request 不承载 focus/evidence-gap 业务事实。
- `reportId` 是唯一 locator；status/error、Context Strip 和 CTA identity 全部来自 `getFeedbackReport` 的 frozen `context`，route 中冲突值一律忽略。
- Reports 的 Back 直接返回当前规划只读详情 `/workspace?targetJobId=...`，不进入 Parse 命令/进度页，不触发解析动画、import 或 polling；Report / Generating 的 Back 使用 API trusted context 中的 `targetJobId` 返回 `/reports?targetJobId=...`，只有无法取得可信 TargetJob identity 时才回 `/workspace` 列表。
- 长上下文可读、明确 mobile 单列、keyboard/a11y、desktop/mobile full-page parity。

### 2.2 Out of Scope

- Questions tab、题目回顾、逐题评分、题号/总题数、per-question replay。
- candidate numeric score、录用概率、排名、timeline、独立错题本。
- 前端推导报告语义或在客户端持久化报告业务状态。
- 真正的通知订阅；若未来需要，必须由独立后端通知合同承接。
- 全局/跨规划 Report Center、报告 timeline、完整历史版本列表，或由 Parse/Report/Generating 消费 `listTargetJobReports`。
- 独立会话历史列表、`sessionId` 用户路由、`listPracticeSessions` / `getPracticeSession` 记录入口，或 Composer/retry/thinking 等实时控件。

## 3 用户决策

| ID | 决策 | 当前结论 |
|----|------|----------|
| D-1 | 报告粒度 | 整场 conversation，不按题目/turn |
| D-2 | 页面骨架 | 无 tab；desktop 自上而下严格采用 `4/2/2/2/1`：Context Strip 四项、数量指标两项、两行各两个常驻区块、底部全宽面试总评；mobile 保持同一 DOM 顺序并收敛为单列；desktop 同一双列行的两个 panel 必须等高，短内容侧在卡片内部留白 |
| D-3 | 报告语义 | LLM direct semantic API；前端只展示 summary、label/status/confidence、evidence、actions |
| D-4 | CTA 推荐 | first next action 只切换两枚既有 CTA 主次，用户仍可自主选择 |
| D-5 | Replay focus | 后端 source report 是唯一事实源，客户端不透传 focus |
| D-6 | Generating | 仅表达真实 queued/generating 状态，不伪造阶段进度、实时证据或通知 |
| D-7 | 页面 owner | GeneratingScreen 与 ReportScreen 均由本 owner 独占；practice owner 只负责 stable reportId handoff |
| D-8 | 语言/长度职责 | 上游 schema200 code points 只作 malformed fuse。Frontend 按 report language 镜像 English 24 whitespace words / zh-CN 64 Unicode code points；English delimiter 以 ECMAScript `/\s/u` 为唯一口径（含 U+FEFF、不含 U+0085），backend/evalkit 必须精确同构。超界 ready payload 进入 typed invalid 且不回显 raw label。targeted repair 的18/52只是上游内部余量，不是 UI 边界。合法边界在desktop+390完整换行；不截断/ellipsis/改写 |
| D-9 | Replay focus | `retryFocusDimensionCodes=[]` 是合法的通用同轮复练，不因空 focus 禁用 Replay；只有非空 focus 才要求每个 code 同时引用 `needs_work` dimension 与至少一条同 code issue，非法非空引用按 direct-contract failure fail closed |
| D-10 | Generating 等待窗口 | 单次`GenerateReport`动作在后端调用内执行initial+最多3次retry并等待10s/20s/40s；动作结束销毁retry context，新的独立动作清零，`async_jobs.attempts/max_attempts`仅作基础设施执行。Frontend固定`maxAttempts=49`、初始1.5s、×1.5、cap8s，总约6m04s，覆盖4×60s+70s=5m10s并留约54s；queued/generating不展示attempt/retry/progress，窗口耗尽只进入可继续检查态，不伪装服务端failed |
| D-11 | Poll pause/resume | hidden/blur只暂停同一poll run；timer等待与in-flight请求都保存current/next attempt和delay，visible/focus从n+1继续，不回1、不重复n。单run累计调用<=49；只有显式continue-check或reportId/client identity变化重置 |
| D-12 | Context Strip 正式截图验收 | 每次验收只保留同一 ready report 的两张正式 frontend real UI `fullPage: true` 图：1440x1200 与 390x844；固定目录、文件名与 manifest schema，逐图绑定 SHA-256、ready state、viewport/fullPage 和 report/session sentinel DOM/a11y absence。Fixture-only 页面、裁剪图、额外状态图不能替代或混入这组成功证据。 |
| D-13 | Back 恢复目标 | trusted `getFeedbackReport` / poll response 提供 `targetJobId` 时，ready、failed 与 recoverable generating 均返回 `/reports?targetJobId=...`；missing reportId、404/首读网络失败或 invalid payload 无可信 TargetJob identity 时返回 `workspace` | Report/Generating route 保持 reportId-only，不把 targetJobId 复制进其 URL，也不让 route identity 覆盖 API |
| D-14 | 当前规划报告列表 | 页面级 `/reports?targetJobId=...` 只展示该 TargetJob canonical rounds 的当前可用报告与最新生成状态；不跨规划、不展示完整历史、也不加入全局 TopBar | `getTargetJob` 提供当前规划/轮次展示事实，`listTargetJobReports` 只提供 report locator 与 attempt 状态；两者 identity 必须闭合 |
| D-15 | Reports Back 与解析职责分离 | ReportsScreen 的可信目标返回 `/workspace?targetJobId=...` 只读详情；Workspace query 只携带 `targetJobId`，不得增加 `resumeId`、`planId`、`reportId` 或 `section` | `parse` 命令/进度路由只承接新 JD 解析；读取既有规划不得展示解析动画，也不得触发 import 或 polling |
| D-16 | 报告附属会话记录 | Report Context Strip 第四个同级子项是主入口，ReportsScreen 当轮 current report 行可提供快捷入口；canonical route 为 `/report-conversation?reportId=...` | 只消费 generated `getReportConversation`；不增加会话列表、第三个 Header CTA、新关系表或浏览器业务存储 |
| D-17 | 报告上下文链接 | 面试记录并入 Context Strip 第四项；简历显示冻结 `resumeDisplayName` 并链接 frozen `resumeId` 的 canonical `/resume-versions?resumeId=...` | Report 不额外调用 `getResume`；URL 可复制、可新标签页打开，目标页复用既有简历详情读取链 |

## 4 UI 设计与正式实现

- `docs/ui-design/report-dashboard.md`
- `docs/ui-design/module-practice-review.md`
- `frontend/src/app/screens/report/components/ReportDashboard.tsx`
- `frontend/src/app/screens/report-conversation/ReportConversationScreen.tsx`
- `frontend/src/app/screens/report/__tests__/ConversationReport.test.tsx`
- `frontend/src/app/i18n/locales/`

正式 `frontend/` 是唯一可运行实现。验收拆为 DOM 顺序/control/a11y、computed style/bounding box/responsive、正式 frontend 的确定性视觉回归和真实 full-page UAT；非空 screenshot buffer 不作为布局完成依据，也不维护平行 prototype 运行时。

## 5 Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | verification |
|-------------|---------|-------------------|-----------------|-------------|---------------|--------------|
| `getTargetJob` | `TargetJobs/getTargetJob.json` | ReportsScreen 当前规划标题与 canonical round display；不作为 Report/Generating 详情补读 | backend-targetjob handler/store | `target_jobs.summary` read | read none | focused generated-client/consumer tests |
| `listTargetJobReports` | `Reports/listTargetJobReports.json` | ReportsScreen 当前规划 current/latest 列表，唯一 UI consumer | backend-review reports handler/store | `feedback_reports` + current TargetJob canonical summary read | read none | focused handler/store/consumer tests |
| `getFeedbackReport` | `Reports/getFeedbackReport.json`: queued/generating/ready-needs-practice/ready-well-prepared/ready-empty-focus/failed/invalid-focus/long-content | generating poll + ReportDashboard；唯一状态/上下文事实源 | backend-review reports handler/store | `feedback_reports` + frozen context | read none | focused consumer tests + `E2E.P0.099` real API/UI |
| `getReportConversation` | `Reports/getReportConversation.json`: ready/queued/generating/failed/missing/invalid-order/unsafe-markdown | ReportConversation + ReportsScreen current-report shortcut locator | backend-review reports handler/store | `feedback_reports.session_id` -> `practice_messages ORDER BY seq_no` | none | focused generated-client/component tests + `BDD.REPORT.CONVERSATION.001` + `E2E.P0.099` click/load/back |
| `createPracticePlan` | `PracticePlans/createPracticePlan.json`: retry/next/mismatch | replay/next CTA；不发送 focus | backend-practice handler/store | `practice_plans` + source report projection | none | focused request/consumer tests |
| `startPracticeSession` | `PracticeSessions/startPracticeSession.json` | replay/next CTA | backend-practice handler/store | session + opening message | `practice.session.chat` | focused request/consumer tests |

Frontend Phase 7.1 必须等待 `backend-review/001 6.1` 的 generated contract；Phase 7.4 必须等待 `backend-review/001 8.1` 的 server-owned projection。`getResume` 不属于本 owner 的读取链；`getTargetJob` 只属于 ReportsScreen，不属于 Report/Generating 详情读取链。详情页冻结 label/identity 直接随 report 返回，避免深链刷新读取可变实体。

`listTargetJobReports` 只属于本 owner 的 ReportsScreen，不属于 Parse、Report 或 Generating 读取链。ReportsScreen 以同一个 `targetJobId` 并行读取 `getTargetJob` 与 overview，先以当前 TargetJob canonical rounds 建 display，再验证 overview target/round identity；Report/Generating 仅根据当前/最后可信 `getFeedbackReport` response 的 `targetJobId` 构造 Back destination。

`listPracticeSessions` 不属于当前产品入口。ReportConversation 不通过 TargetJob 或 session 列表定位；它只使用 `reportId` 调用 `getReportConversation`，由后端反查唯一 session。

## 6 页面结构

### 6.1 Reports list

```text
ReportsScreen
├─ Back -> /workspace?targetJobId=... read-only plan detail
├─ Header / current plan identity
├─ loading / empty / error
└─ CanonicalRoundList
   ├─ currentReport -> report?reportId=...
   │  └─ conversation shortcut -> report-conversation?reportId=...
   └─ latestAttempt -> generating?reportId=... / typed failed state
```

- `/reports` 是受保护的上下文页面，query 只允许一个合法 UUID `targetJobId`；缺失或非法时以 replace-only 方式安全回到 workspace，不新增 history entry 或形成 Back 循环；无权访问时显示可理解错误，不把其他 route 参数当业务权威。
- `getTargetJob(targetJobId)` 提供标题及 2~5 条 canonical round 的 name/type/duration/focus；`listTargetJobReports(targetJobId)` 只提供 current report 与 latest attempt。响应 target、round ID、sequence、数量和顺序必须与 TargetJob 精确一致，否则整页列表 fail closed。
- 非空 report locator 必须全局唯一归属一个 canonical round；同一 locator 只允许在同一轮的 `currentReport` 与 `latestAttempt` 双占位，且此时 latest 必须为 `ready`。`latestAttempt.status=ready` 必须存在 `currentReport`，非空 `currentReport` 也必须存在 `latestAttempt`；任一不可能组合整页 fail closed。
- 每轮只显示当前可用 ready 报告和最新生成尝试；latest ready 与 current 相同只显示一个报告入口。queued/generating 进入 Generating，failed 显示 typed localized 状态且无同 report Retry，both-null 表达该轮尚无报告；不展开完整历史版本。
- target 切换时立即清空旧 rows，旧请求响应不得覆盖新 target；两个规划的 report sentinel 必须严格隔离。loading、empty、network/contract error 均有独立可访问状态，失败时不保留 stale rows。
- 页面保留 App chrome，但不在 TopBar 增加“面试报告”入口；Back 精确回 `/workspace?targetJobId=<当前可信 id>`，直接读取规划只读详情。该导航不得经过 `/parse`，不得展示解析动画，也不得触发 import 或 polling；Workspace query 只允许 `targetJobId`。

### 6.2 Generating

```text
GeneratingScreen
├─ honest queued/generating status
├─ static explanation of analysis work
├─ polling / timeout / typed failure
└─ Continue checking (timeout/network only) / Back to current Reports or Workspace fallback
```

- 不显示与后端无关的百分比或自动逐项完成动画。
- 不显示固定“实时观察”内容。
- 不承诺“好了通知我”或后台通知。当前 records consumer 是独立 ReportsScreen；Generating 只能在 response 已提供可信 `targetJobId` 时通过 Back 返回该页面，不能伪造通知或自行读取报告列表。
- queued/generating 自动继续检查；timeout/network 允许“继续检查”重启轮询；failed/not-found/invalid-contract/`REPORT_CONTEXT_TOO_LARGE` 是终态，只提供返回，不把再次 GET 伪装成重新生成。超限文案应说明本次材料与对话过长、可返回规划后缩短输入并开启新会话，不承诺当前 report 可重试恢复。
- Report generation的10s/20s/40s是单次后端动作内的等待，不来自runner business requeue；`async_jobs.attempts/max_attempts`与outbox/infra投递的30s/2m/10m/1h/6h都不属于本页面的产品retry事实。Frontend只消费queued/generating/ready/failed，不展示内部attempt_count/retry_count/reason/scope。当前OpenAPI没有failed report regenerate operation，因此本页不得宣称或新增同report Retry入口；独立动作清零只由后端两次`GenerateReport` invocation证明。
- Polling 固定为`maxAttempts=49`、1.5s初始退避、乘数1.5、8s cap，总窗口约6m04s。它覆盖report最坏4×60s provider timeout +10+20+40s=5m10s并保留约54s调度/网络余量；不得恢复旧30次窗口。窗口耗尽不改变服务端状态，只显示可继续检查的typed client timeout。

### 6.3 Report Dashboard

```text
ReportDashboard
├─ Back -> current Reports / Workspace fallback
├─ Header
│  ├─ title / subtitle
│  ├─ replay current round
│  └─ next round
├─ ContextStrip
│  ├─ target
│  ├─ round
│  ├─ resume link
│  └─ interview record link
├─ SummaryMetrics
│  ├─ dimension count
│  └─ evidence count
├─ DetailGrid
│  ├─ Dimensions
│  ├─ Strength Evidence
│  ├─ Risks
│  └─ Next Actions
└─ OverallSummary (full width)
   ├─ localized readiness tier
   └─ summary
```

### 6.4 可读性与响应式

- target/round/resume/interview record 是同一 Context Strip 的四个同级子项；target/round/resume 来自只读 frozen report context，resume 名称以 frozen `resumeId` canonical URL 打开副本，interview record 通过不携带 reportId DOM 属性的站内 action 打开。session/report UUID 等内部 locator 不渲染为用户字段，也不通过 title、tooltip 或 accessible description 暴露。
- Desktop ready 内容严格按 `4/2/2/2/1`：Context Strip 四列、Summary Metrics 两列、Dimensions/Strength Evidence 两列、Risks/Next Actions 两列、Overall Summary 跨两列全宽。Mobile 390px 保持相同 DOM 顺序并全部单列，长 label/evidence/action/summary 不横向溢出。
- Desktop 两个 detail row 都以较高 panel 决定行高，同行较短 panel 的可见卡片边框填满该行并在内部留白；不得只拉伸无边框 wrapper，导致卡片本体仍一高一低。
- Frontend consumer 在 render 前执行 24/64 semantic boundary；English 按 ECMAScript `/\s/u` whitespace words（U+FEFF 是 delimiter，U+0085 不是）、zh-CN 按 Unicode code points 计数。超界 payload 不进入 ReportDashboard，不得利用 CSS 截断把 invalid 内容伪装为可用。
- Deterministic frontend/OpenAPI fixture 使用恰好 24-word/64-code-point actions；正式 frontend 的 1440x1200 与 390x844 full-page responsive/geometry gate 必须覆盖 action 区域和底部 Overall Summary，证明完整换行、顺序稳定且 `summary` 只出现一次。200-code-point malformed fixture只证明 typed invalid/no-raw-output，不得替代 UX gate；18/52 targeted-repair margin 也不得替代边界 fixture。
- Context Strip 用户验收固定写入 `.test-output/acceptance/report-context-strip/<run-id>/`，且成功目录只包含 `report-context-strip-desktop-1440x1200.png`、`report-context-strip-mobile-390x844.png` 与 `manifest.json`。两图必须来自同一正式 frontend 的真实 backend ready report，分别使用 exact viewport 1440x1200 / 390x844 和 `fullPage: true`；不得用 fixture-only 页面、裁剪图或额外 loading/error 图冒充。
- `manifest.json` 必须逐图记录相对路径、SHA-256、`state=ready`、viewport width/height、`fullPage=true`、同一 report 的脱敏 locator/digest，以及 `reportSentinelAbsent=true`、`sessionSentinelAbsent=true`；同时绑定该页面的 DOM/a11y negative audit，证明 report/session sentinel 在 text、title/tooltip、任意 `aria-*` 与 accessible name 中均不存在。截图中 target/round/resume/interview record 必须可见，resume link 与 conversation action 必须通过当前行为审计，且 report/session sentinel 不能以用户文案、调试标记或可访问名称出现。
- status/confidence、readiness、CTA chrome、empty/error/loading 随 UI locale 本地化；summary、dimension label、evidence 与 action label 按 report language 原样展示。未知 enum fail closed 到 typed error，不回显 raw token。

### 6.5 CTA

- `retry_current_round` 为 first action 时 Replay 使用 accent，Next 为 secondary。
- `next_round` 为 first action 时 Next 使用 accent，Replay 为 secondary。
- `review_evidence` 时两者均 secondary，正文行动建议保持可见。
- next round unavailable 时 disabled 并提供可访问原因；两条启动路径仍共享重复点击锁。
- `retryFocusDimensionCodes=[]` 不代表报告或 Replay 无效：Replay 仍创建通用同轮 plan。非空 focus 才做 issue-backed 校验；任一 code 未命中 `needs_work` dimension 或同 code issue 时整份 ready contract fail closed，前端不删错项、不猜 focus。
- derived request 只带 goal + sourceReportId；persona/difficulty/language/time budget/target/resume/round/focus 全由后端从 source report 派生。`context.hasNextRound=false` 时下一轮 disabled；route 中 identity/status/error 即使存在也不得覆盖 API。

### 6.6 Back 与报告记录恢复

- ReportsScreen 在 `getTargetJob` 与 overview identity 闭合后，其 Back 精确导航 `/workspace?targetJobId=<id>` 并进入只读规划详情；缺失/非法/不可验证 target 的 Reports deep link 仍 replace 到 `/workspace` 列表。该链路不得进入 Parse 命令/进度状态机，也不得触发 import、解析动画或 polling。
- trusted response context 有 `targetJobId` 时，Report/Generating Back 精确导航 `/reports?targetJobId=<id>`；ReportsScreen 再以该 ID 重新读取并验证当前规划与 overview。
- missing reportId、404、首读网络失败或 invalid payload 无可信 `targetJobId` 时，Back 回 `/workspace` 列表。不得从 route、旧状态、target title 或 reportId 推导 TargetJob identity。
- Report/Generating route 继续只使用 `reportId`；不得为了 Back 在其 URL 增加 `targetJobId`，也不得调用 `listTargetJobReports`。ReportsScreen 是唯一列表 consumer。

### 6.7 Report Conversation

```text
ReportConversation(reportId)
├─ Back -> /report or /generating with the same reportId
├─ Frozen target / round / resume summary
└─ Ordered readonly Markdown messages
```

- `reportId` 变化的首帧清空旧 transcript；只有 closed `ReportConversation` payload 校验通过后才渲染。
- `reportStatus=ready` 返回 `/report?reportId=...`；queued/generating/failed 返回 `/generating?reportId=...`。缺失/无权/首读失败且无可信父状态时回 `/workspace`。
- 消息的 `sequence` 必须严格升序且唯一，role 只允许 user/assistant；正文复用 Practice 的安全 Markdown/GFM 策略，不读取 live `replyStatus/clientMessageId`。
- loading/error/empty-corruption/hidden-404 状态不展示 partial transcript；desktop/mobile 均为单列，390px code/table 只在消息容器内滚动。

## 7 状态与错误

- 缺 session/report：专用空态，不展示假报告。
- queued/generating：留在 honest generating。
- ready：渲染完整 direct semantic dashboard。
- timeout/network：typed recoverable error、继续检查/back；failed/not found/unknown enum/invalid contract/`REPORT_CONTEXT_TOO_LARGE`：typed terminal error、back，不显示虚假 Retry；若当前/最后可信 response 已给出 `targetJobId`，Back 返回当前 Reports 列表，否则回 `/workspace` 列表。超限文案保持本地化且给出返回规划后的可执行方向。
- empty dimensions/evidence 或缺 summary：视为后端合同失败，不由前端补假数据。
- 空 `retryFocusDimensionCodes` 单独不构成合同失败；非法非空 focus cross-reference 才进入 typed invalid-contract 终态。
- ReportConversation 对 queued/generating/ready/failed 使用同一只读合同；report 不存在、跨用户、消息空内容/乱序/非法 role 或 unsafe contract 时整体 fail closed。

## 8 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | honest generating | report queued/generating | 页面轮询 | 无假百分比、假观察、假通知 | 001 |
| C-2 | ready dashboard | direct semantic report | 打开 report | desktop 为 `4/2/2/2/1`；四项上下文同属一个 block，简历具有 canonical URL、面试记录使用不暴露 reportId 的 action，两组 detail panel 同行等高；顶部只有两个数量指标，四个内容区之后是全宽面试总评，localized readiness 与唯一 `summary` 完整 | 001 |
| C-3 | recommended action | retry/next/review first action | 查看 Header | 仅切换现有 CTA 主次且功能可用 | 001 |
| C-4 | server-owned replay | report 含 retry focus | 点击复练 | request 不传 focus，服务端 plan/session 得到 focus | 001 |
| C-5 | long/mobile | 长 target/round/resume/evidence | desktop/mobile 打开 | 完整可读、mobile 单列、无横向溢出 | 001 |
| C-6 | deterministic responsive contract | frontend/OpenAPI fixtures 含恰好 24-whitespace-word / 64-Unicode-code-point actions | 运行正式 frontend component/browser gate | DOM 顺序、style、bbox 与 viewport 证明 desktop `4/2/2/2/1`、两组 detail panel 同行等高、mobile 同序单列、底部总评全宽/可读、边界 label 完整换行且无截断/省略/横溢 | 001 |
| C-7 | current-run canonical mobile UAT | P0.099 当前 run 的 en/zh ready rows | exact six images | 每个 row 绑定 DB/API `canonical_report_content_digest`、`action_length_audit`、`content_audit`、`screenshot_sha256` 与 report/session/context digest；390x844 report 图覆盖 action 区域与其后的面试总评，实际 `<=24-word` / `<=64-code-point` label 与唯一 summary 完整可见且无截断/省略/横溢 | 001 |
| C-8 | stale negative | 全仓 active assets | 扫描 | 无 raw enum UI、fake-live copy、客户端 focus 与旧 question fields | 001 |
| C-9 | route tamper / deep link | 只有 reportId，或 route 带冲突 status/target/resume/round | 刷新/读取/点击 CTA | API frozen context/status 获胜，route 不能改变展示与动作 | 001 |
| C-10 | language split | UI locale 与 report language 不同 | 查看报告 | chrome 随 UI locale；模型语义原文不翻译、不改写 | 001 |
| C-11 | empty / issue-backed focus | ready report 的 focus 为空，或非空 focus 引用 needs-work/issue | 点击 Replay / 校验 direct contract | 空 focus 合法创建通用同轮复练；非空 focus 必须逐项 issue-backed，非法引用 fail closed | 001 |
| C-12 | internal locator cleanup + formal acceptance | 同一真实 backend ready report 含 distinct frozen session/report sentinel | 以 Chrome 分别在 exact 1440x1200 / 390x844 执行 `fullPage: true` 捕获并生成固定 manifest | Context Strip 显示 target/round/resume/interview record 四项；resume URL 与 conversation action 可用；成功目录只有固定两图+manifest；逐图 path/hash/state/viewport/fullPage、responsive geometry 与 DOM/a11y sentinel absence 闭合，API 事实源与 CTA 行为不变 | 001 |
| C-13 | 当前规划报告列表 | `/reports` 有合法且可访问的 `targetJobId`，overview 覆盖 both-null、prior-ready+newer-failed、generating-only、latest-ready | 直开/刷新/切换状态并查看轮次 | 只显示当前 TargetJob canonical rounds 的 current report 与 latest attempt；无其他规划或完整历史；loading/empty/error/identity mismatch/stale response fail closed；ready/generating 链接正确；1440/390 正式 frontend 响应式 gate 通过 | 001 |
| C-14 | Back to current reports | ready/failed/recoverable generating response 有 trusted targetJobId，或首读失败无可信 identity | 点击 Back | 前者到 `/reports?targetJobId=...`；后者到 `/workspace` 列表；report/generating route 仍是 reportId-only，不调用 list operation、不信任 route target identity | 001 |
| C-15 | Reports Back 直达只读规划详情 | ReportsScreen 已闭合可信 `targetJobId` | 点击 Back | 直接到 `/workspace?targetJobId=...`，query 只有 `targetJobId`；不进入 Parse、不展示解析动画、不触发 import/polling；无可信 identity 时回 `/workspace` 列表 | 001 |

## 9 关联计划

- [001-report-screen-and-generating-handoff](./plans/001-report-screen-and-generating-handoff/plan.md)

## 10 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-15 | 1.26 | 用户确认 ready 报告采用 `3/2/2/2/1`：顶部 readiness+summary 下移为底部全宽面试总评，顶部只保留两个数量指标，mobile 同序单列；同步正式 frontend 单实现验收口径，不改 OpenAPI/backend/persistence。 |
| 2026-07-14 | 1.25 | Separate read-only plan navigation from parsing: Reports Back now opens `/workspace?targetJobId=...` directly, while Report/Generating trusted Back remains `/reports` and untrusted fallback remains the workspace list. |
| 2026-07-14 | 1.24 | Close report-locator round ownership and current/latest cross-field invariants; invalid/missing Reports deep links replace to workspace without a browser Back loop. |
| 2026-07-14 | 1.23 | Move per-round records into an independent target-scoped ReportsScreen, keep current/latest only, and redirect trusted Report/Generating Back actions to `/reports`; no global center, schema, or persistence change. |
| 2026-07-14 | 1.22 | Lock Report/Generating Back to the Parse per-round reports section when trusted targetJobId exists, with workspace fallback and reportId-only route/list-consumer negatives. |
| 2026-07-13 | 1.21 | Remove session/report UUID from the user-visible Context Strip while retaining API-frozen context as the internal action authority. |
| 2026-07-13 | 1.20 | Supersede 1.18 timing ownership: 10s/20s/40s belong to one action-local GenerateReport retry context; async job attempts are infrastructure-only. Keep maxAttempts49/6m04s honest polling and do not claim a failed-report regenerate API. |
| 2026-07-13 | 1.19 | L2：in-flight/timer pause preserves attempt and schedule；resume never resets1 or exceeds49；run35622 is aborted7/11 not PASS. |
| 2026-07-13 | 1.18 | 锁定maxAttempts49（约6m04s），覆盖report复用business policy的10s/20s/40s与4×60s调用；分离business async与infra投递退避，不展示内部attempt/progress。 |
| 2026-07-13 | 1.17 | 方案 A 最终边界：schema fuse200；frontend semantic/UX 24 whitespace words / 64 Unicode code points；18/52仅为上游targeted-repair余量；desktop+390合法边界完整换行。 |
| 2026-07-13 | 1.16 | A-200：wire fuse改200；14/40仍为frontend typed-invalid gate，desktop+390合法边界完整换行，超限不回显raw。 |
| 2026-07-13 | 1.15 | 区分 malformed fuse、P0.099 current-run canonical screenshot audit 与 deterministic boundary tests；provider/eval output digest 不作为 UI 前置条件。 |
| 2026-07-12 | 1.14 | 明确空 focus 是合法通用同轮 Replay；非空 focus 才执行 needs-work + issue-backed cross-reference 校验。 |
| 2026-07-12 | 1.13 | 将 GeneratingScreen 唯一 owner 转入本模块；补 immutable report context、route tamper、终态动作矩阵与 UI/report 双语言边界。 |
| 2026-07-12 | 1.12 | 重新打开 001：锁定三指标四常驻区块，接入 direct semantic summary/code+label，删除 generating 伪实时语义，增加 i18n/CTA/mobile/readability 与真实截图 gate。 |
| 2026-07-12 | 1.11 | 报告改为会话级四部分，删除逐题模型、hint/phone 展示与 turn-based replay。 |
