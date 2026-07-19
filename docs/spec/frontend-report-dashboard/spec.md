# Frontend Report Dashboard Spec

> **版本**: 1.35
> **状态**: completed
> **更新日期**: 2026-07-19

## 1 背景与目标

`frontend-report-dashboard` 承接当前面试规划的独立报告列表、一次连续模拟面试结束后的诚实生成态与证据化报告。列表帮助用户按当前规划的 canonical rounds 找到当前可用报告与最新生成状态；详情帮助用户理解总体准备度、能力维度、优势/风险证据和下一步建议，并从当前轮复练或进入下一轮。

报告不按题目组织，也不展示伪精确分数。前端只展示服务端持久化的最终语义，不自行拆分维度、推断 readiness、改写证据或生成建议。

## 2 范围

### 2.1 In Scope

- `reports`：受保护的独立页面，唯一上下文参数为 `targetJobId`；读取当前 TargetJob 与 `listTargetJobReports(targetJobId)`，只展示该规划 canonical rounds 的 `currentReport/latestAttempt`，不做全局中心或完整版本历史。Desktop 使用约 `1372px` 内容面、Header decorative illustration、只消费现有事实的目标摘要卡，以及“左侧两位编号/贯穿竖线 + 右侧独立轮次卡”的时间线构图；主报告、记录、生成和恢复动作按状态保持明确主次。
- `reports` 的 loading / empty / error / ready 状态彼此完备；target/round identity 漂移、跨规划响应和 stale request 均 fail closed，不渲染其他规划 sentinel 或错链。
- `generating`：轮询真实 report status，展示诚实的异步等待说明；不伪造百分比、实时观察或通知订阅。
- `report`：Header、四项 Context Strip（目标岗位 / 轮次 / 可链接简历副本 / 面试记录）、两项 Summary Metrics、两行各两个常驻内容区（Dimensions / Strength Evidence / Risks / Next Actions），以及底部一个全宽 Overall Summary。
- `report` desktop `2048×917` 目标视图使用约 `1432px` 居中内容面与浅蓝全视口背景；Back、标题、CTA、Context Strip、两列卡片和底部总评共享同一横向网格。Context Strip 是一张由三条内部竖线分成四列的整卡；四张内容卡必须具备语义圆形 icon 与紧凑正文结构，优势/风险不重复显示能力维度行已经表达的 confidence。典型两维度/两证据合法报告的底部总评应完整进入首屏；长合法内容仍完整换行且不得截断。不得把只调整 max-width、圆角、背景或无溢出当作目标稿改造完成。
- `report-conversation`：仅以 `reportId` 读取报告附属的 ordered user/assistant transcript，安全渲染 Markdown/GFM，并返回同一报告状态页；queued/generating/ready/failed 都可访问。Desktop 使用约 `1372px` 内容面、Header decorative illustration、三列分隔 Context Strip；assistant/user 共享浅色整行卡片、描边、圆角、内边距与同宽方形头像轮廓，只以蓝色 AI / 灰色“我”的色彩和角色文案区分身份。ReportsScreen 只要存在代表已结束会话的 current report 或 latest attempt，就必须独立展示“查看面试记录”，不能被“查看生成进度”或重新生成动作替代。
- failed latest attempt recovery：除 `REPORT_CONTEXT_TOO_LARGE` 外，ReportsScreen 以同一 `reportId` 调用 `regenerateFeedbackReport` 并进入 Generating；所有 failed report 都保留“查看面试记录”。
- Overall Summary 使用“面试总评”标题，同时展示 localized readiness tier 与服务端 `summary`；二者不得继续出现在顶部指标区，`summary` 全页只展示一次。
- Dimension 使用动态 `label`，status/confidence 走完整 zh/en i18n，不泄漏 raw enum/code。
- Header 保留唯一一对 CTA；按目标稿固定“复练当前轮”为 accent 主按钮、“进入下一轮”为描边次按钮，并分别显示刷新/右箭头 icon。`nextActions[0]` 仍作为报告正文建议展示，但不再交换 Header 视觉层级；按钮可用性、请求与导航语义不变。
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
| D-13 | Back 恢复目标 | trusted `getFeedbackReport` / poll response 提供 `targetJobId` 时，ready、failed 与 recoverable generating 均返回 `/reports?targetJobId=...`；missing reportId、404/首读网络失败或 invalid payload 无可信 TargetJob identity 时返回 `workspace`；failed conversation 在 owner read 仍 resolving 时不渲染可点击 Back，避免短暂默认跳错 workspace | Report/Generating route 保持 reportId-only，不把 targetJobId 复制进其 URL，也不让 route identity 覆盖 API |
| D-14 | 当前规划报告列表 | 页面级 `/reports?targetJobId=...` 只展示该 TargetJob canonical rounds 的当前可用报告与最新生成状态；不跨规划、不展示完整历史、也不加入全局 TopBar | `getTargetJob` 提供当前规划/轮次展示事实，`listTargetJobReports` 只提供 report locator 与 attempt 状态；两者 identity 必须闭合 |
| D-15 | Reports Back 与解析职责分离 | ReportsScreen 的可信目标返回 `/workspace?targetJobId=...` 只读详情；Workspace query 只携带 `targetJobId`，不得增加 `resumeId`、`planId`、`reportId` 或 `section` | `parse` 命令/进度路由只承接新 JD 解析；读取既有规划不得展示解析动画，也不得触发 import 或 polling |
| D-16 | 报告附属会话记录 | Report Context Strip 第四个同级子项是主入口；ReportsScreen 当轮 current report 与 queued/generating/failed latest attempt 都必须提供快捷入口，生成进度/恢复动作只能并列存在；canonical route 为 `/report-conversation?reportId=...` | 只消费 generated `getReportConversation`；不增加会话列表、第三个 Header CTA、新关系表或浏览器业务存储 |
| D-17 | 报告上下文链接 | 面试记录并入 Context Strip 第四项；简历显示冻结 `resumeDisplayName` 并链接 frozen `resumeId` 的 canonical `/resume-versions?resumeId=...` | Report 不额外调用 `getResume`；URL 可复制、可新标签页打开，目标页复用既有简历详情读取链 |
| D-18 | Failed report recovery | ReportsScreen 对非超限 failed latest attempt 显示“重新生成报告”与“查看面试记录”；regenerate 复用同一 reportId 和稳定 IK，成功只进入同 ID Generating | `REPORT_CONTEXT_TOO_LARGE` 不显示 regenerate；旧 current ready 与更新 failed 并存时两组动作分别绑定各自 locator，并有可区分 accessible name；双击、stale target response 与 malformed job fail closed。若另一页面已改变状态，`REPORT_INVALID_STATE_TRANSITION` / `REPORT_NOT_READY` 触发 current target + overview 重读，不能保留 stale failed 操作 |

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
| `getReportConversation` | `Reports/getReportConversation.json`: ready/queued/generating/failed/missing/invalid-order/unsafe-markdown | ReportConversation + ReportsScreen current-report / every distinct latest-attempt shortcut locator | backend-review reports handler/store | `feedback_reports.session_id` -> `practice_messages ORDER BY seq_no` | none | focused generated-client/component tests + `BDD.REPORT.CONVERSATION.001` + `E2E.P0.099` click/load/back |
| `regenerateFeedbackReport` | `Reports/regenerateFeedbackReport.json` | ReportsScreen failed latest-attempt recovery；same reportId | backend-review regeneration handler/service/store | same report row + fresh job + idempotency/audit | async `report.generate` after 202 | focused generated-client/component tests + `BDD.REPORT.REGENERATE.UI.001` |
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
   └─ latestAttempt -> conversation action + generating progress / failed regenerate by status
```

- `/reports` 是受保护的上下文页面，query 只允许一个合法 UUID `targetJobId`；缺失或非法时以 replace-only 方式安全回到 workspace，不新增 history entry 或形成 Back 循环；无权访问时显示可理解错误，不把其他 route 参数当业务权威。
- `getTargetJob(targetJobId)` 提供标题及 2~5 条 canonical round 的 name/type/duration/focus；`listTargetJobReports(targetJobId)` 只提供 current report 与 latest attempt。响应 target、round ID、sequence、数量和顺序必须与 TargetJob 精确一致，否则整页列表 fail closed。
- 非空 report locator 必须全局唯一归属一个 canonical round；同一 locator 只允许在同一轮的 `currentReport` 与 `latestAttempt` 双占位，且此时 latest 必须为 `ready`。`latestAttempt.status=ready` 必须存在 `currentReport`，非空 `currentReport` 也必须存在 `latestAttempt`；任一不可能组合整页 fail closed。
- 每轮只显示当前可用 ready 报告和最新生成尝试；latest ready 与 current 相同只显示一个报告与记录入口。任意不同 ID 的 latest attempt 都显示自己的面试记录；queued/generating 另行进入 Generating，普通 failed 另行显示同 report regenerate，`REPORT_CONTEXT_TOO_LARGE` 不显示 regenerate；both-null 表达该轮尚无报告，不虚构记录入口，也不展开完整历史版本。
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
- queued/generating 自动继续检查；timeout/network 允许“继续检查”重启轮询；failed/not-found/invalid-contract/`REPORT_CONTEXT_TOO_LARGE` 对当前 Generating 页面是终态，只提供返回，不把再次 GET 伪装成重新生成。返回 Reports 后，普通 failed 才显示显式的新生成动作；超限文案说明缩短输入并开启新会话。
- Report generation 的10s/20s/40s是单次后端动作内的等待，不来自runner business requeue；`async_jobs.attempts/max_attempts`与outbox/infra投递的30s/2m/10m/1h/6h都不属于页面产品 retry。Frontend 只消费 queued/generating/ready/failed，不展示内部 attempt/retry/reason/scope。用户显式 regenerate 是新的后端动作并重置独立预算，不是暴露内部 retry。
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
├─ Back -> ready report / pending generating / failed target-scoped reports list
├─ Frozen target / round / resume summary
└─ Ordered readonly Markdown messages
```

- `reportId` 变化的首帧清空旧 transcript；只有 closed `ReportConversation` payload 校验通过后才渲染。
- `reportStatus=ready` 返回 `/report?reportId=...`；queued/generating 返回 `/generating?reportId=...`；failed 在 frozen context 提供可信 `targetJobId` 时直接返回 `/reports?targetJobId=...`，使用户继续在同一行处理重新生成，缺失可信 target 时回 `/workspace`。缺失/无权/首读失败且无可信父状态时同样回 `/workspace`。
- 消息的 `sequence` 必须严格升序且唯一，role 只允许 user/assistant；正文复用 Practice 的安全 Markdown/GFM 策略，不读取 live `replyStatus/clientMessageId`。
- loading/error/empty-corruption/hidden-404 状态不展示 partial transcript；desktop/mobile 均为单列，390px code/table 只在消息容器内滚动。

## 7 状态与错误

- 缺 session/report：专用空态，不展示假报告。
- queued/generating：留在 honest generating。
- ready：渲染完整 direct semantic dashboard。
- timeout/network：typed recoverable error、继续检查/back；failed/not found/unknown enum/invalid contract/`REPORT_CONTEXT_TOO_LARGE`：Generating typed terminal error、back。ReportsScreen 只对普通 failed 显示显式 regenerate，对超限只显示 conversation；若当前/最后可信 response 已给出 `targetJobId`，Back 返回当前 Reports 列表，否则回 `/workspace` 列表。
- empty dimensions/evidence 或缺 summary：视为后端合同失败，不由前端补假数据。
- 空 `retryFocusDimensionCodes` 单独不构成合同失败；非法非空 focus cross-reference 才进入 typed invalid-contract 终态。
- ReportConversation 对 queued/generating/ready/failed 使用同一只读合同；report 不存在、跨用户、消息空内容/乱序/非法 role 或 unsafe contract 时整体 fail closed。

## 8 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | honest generating | report queued/generating | 页面轮询 | 无假百分比、假观察、假通知 | 001 |
| C-2 | ready dashboard | direct semantic report | 打开 report | desktop 为 `4/2/2/2/1`；四项上下文同属一个 block，简历具有 canonical URL、面试记录使用不暴露 reportId 的 action，两组 detail panel 同行等高；顶部只有两个数量指标，四个内容区之后是全宽面试总评，localized readiness 与唯一 `summary` 完整 | 001 |
| C-3 | fixed Header hierarchy | retry/next/review first action | 查看 Header | 复练始终为带刷新 icon 的 accent 主按钮，下一轮始终为带右箭头的描边次按钮；功能与禁用规则保持可用 | 001 |
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
| C-14 | Back to current reports | ready/failed/recoverable generating response 有 trusted targetJobId，或首读失败无可信 identity | 点击 Back | 前者到 `/reports?targetJobId=...`；后者到 `/workspace` 列表；failed conversation owner resolving 时 Back 不可点击/不可见；report/generating route 仍是 reportId-only，不调用 list operation、不信任 route target identity | 001 |
| C-15 | Reports Back 直达只读规划详情 | ReportsScreen 已闭合可信 `targetJobId` | 点击 Back | 直接到 `/workspace?targetJobId=...`，query 只有 `targetJobId`；不进入 Parse、不展示解析动画、不触发 import/polling；无可信 identity 时回 `/workspace` 列表 | 001 |
| C-16 | failed report recovery | latest attempt 为普通 failed、超限 failed，或旧 ready + 更新 failed | 点击重新生成/查看记录，可能双击、网络未知或切换 target | 普通 failed 同 ID 生成且进入 matching Generating；同 key 复用、双击单请求、stale/malformed response 不导航；超限只有记录；旧/新 locator 与 accessible name 可区分 | 001 |
| C-17 | 已结束面试记录始终可用 | current report 或 latest attempt 已存在，报告处于 queued/generating/ready/failed | 打开 ReportsScreen | 每个不同 reportId 都有“查看面试记录”；queued/generating 同时保留“查看生成进度”，failed 同时保留允许的恢复动作，same-ID ready 不重复，empty round 不虚构记录入口 | 001 |
| C-18 | 目标稿整页结构 | 合法 ready report 在 2048×917 / 390×844 展示 | 逐块检查 Header、Context、Metrics、Detail 与 Overall | desktop 约 1432px 居中；Context 为一张四列分隔整卡；四个 Detail card 带语义 icon 并使用紧凑列表结构；典型内容总评完整进入首屏；mobile 同序单列且长内容不截断 | 001 |
| C-19 | 报告列表目标构图 | 当前 TargetJob 有 canonical rounds 和 current/latest 报告事实 | 打开 `/reports?targetJobId=...` | desktop 约 1372px，Header 插画、现有事实目标摘要卡、编号时间线和独立轮次卡完整；报告/记录/生成/恢复动作保持各自 locator、状态与 a11y；mobile 同序无横溢 | [001 Phase 18](./plans/001-report-screen-and-generating-handoff/plan.md) |
| C-20 | 面试记录目标构图 | report-owned transcript 含合法 assistant/user 消息 | 打开 report-conversation | desktop 约 1372px，Header、三列 Context Strip 和消息列共享边界；assistant/user 共享浅色整行卡片、描边、圆角、内边距和同宽方形头像轮廓，只以蓝色 AI / 灰色“我”区分身份；Markdown、安全边界、Back 与无 composer/internal IDs 合同不变，mobile 同序无横溢 | [001 Phase 18](./plans/001-report-screen-and-generating-handoff/plan.md) |

## 9 关联计划

- [001-report-screen-and-generating-handoff](./plans/001-report-screen-and-generating-handoff/plan.md)

## 10 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-19 | 1.35 | Align assistant/user transcript cards to one shared outlined surface and avatar silhouette per acceptance feedback. |
| 2026-07-19 | 1.34 | Reopen ReportsScreen and ReportConversation for the supplied target composition: shared illustrated Header, real-fact summary card, numbered round timeline, three-column context and differentiated assistant/user rows. |
| 2026-07-19 | 1.33 | Reopen the ready dashboard for a complete target-composition rebuild after the prior width-only alignment failed to implement the supplied UI structure. |
| 2026-07-16 | 1.31 | Hide failed-conversation Back until the trusted report owner resolves, preventing a transient workspace misroute. |
| 2026-07-16 | 1.30 | Require the interview-record action for every completed-session report projection, including queued/generating rows alongside generation progress. |
| 2026-07-16 | 1.29 | Add same-report failed regeneration and failed-transcript actions to ReportsScreen, with oversize exclusion, idempotency and stale-response fences. |
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
