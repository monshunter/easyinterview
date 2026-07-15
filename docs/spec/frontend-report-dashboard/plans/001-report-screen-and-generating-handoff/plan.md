# 001 — Honest Grounded Report Screen and Handoff

> **版本**: 3.8
> **状态**: completed
> **更新日期**: 2026-07-15

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

在既有 report UI owner 内交付规划范围的独立 ReportsScreen、grounded direct semantic report、诚实 generating 与 report-owned 只读 conversation：列表只展示当前 TargetJob canonical rounds 的 current report 与 latest attempt，不做全局中心或完整历史；ready 详情采用 `4/2/2/2/1` 当前设计合同；会话记录以 reportId-only 附属页复用安全 Markdown message renderer，不建立 session 列表或实时控制。

## 2 Operation Matrix

| operationId | fixture | consumer | backend | persistence | AI | verification |
|-------------|---------|----------|---------|-------------|----|--------------|
| getTargetJob | `TargetJobs/getTargetJob.json` | ReportsScreen current target + canonical round display | targetjob handler/store | target_jobs.summary read | read none | focused client/consumer tests |
| listTargetJobReports | `Reports/listTargetJobReports.json` | ReportsScreen current/latest pointers only | reports handler/store | feedback_reports + canonical target read | read none | focused handler/store/consumer tests |
| getFeedbackReport | `Reports/getFeedbackReport.json`: queued/generating/ready-needs-practice/ready-well-prepared/ready-empty-focus/failed/invalid-focus/long-content | generating/report; only status/context truth | reports handler/store | feedback_reports + frozen context | read none | focused consumer tests + P0.099 real API/UI |
| getReportConversation | `Reports/getReportConversation.json`: ready/queued/generating/failed/missing/invalid-order/unsafe-markdown | report-conversation; ReportsScreen 只传 current report locator | reports handler/store | feedback_reports.session_id -> practice_messages ordered read | none | focused client/component/route tests + BDD.REPORT.CONVERSATION.001 + P0.099 click/load/back |
| createPracticePlan | `PracticePlans/createPracticePlan.json`: retry/next/mismatch | replay handler; no focus input | practice handler/store | practice_plans + source report projection | none | focused request/consumer tests |
| startPracticeSession | `PracticeSessions/startPracticeSession.json` | replay handler | practice handler/store | session/messages | practice.session.chat | focused request/consumer tests |

## 2.1 Owner / Dependency Order

| Gate | Dependency | Rule |
|------|------------|------|
| Phase 6 | ownership transfer recorded in `frontend/README.md` + workspace/practice spec | this plan exclusively edits the formal GeneratingScreen |
| Phase 7.1 | `backend-review/001 6.1` + OpenAPI 001/002 codegen/fixture PASS | frontend RED/GREEN uses the real generated contract, not handwritten types |
| Phase 7.4 | `backend-practice/004 Phase 3` server-derived request/focus PASS | frontend removes all derived settings/identity/focus route authority |
| Phase 8 | backend 6-8 + frontend 6-7 PASS | code gates close deterministic behavior；P0.099 is reserved for an explicitly run real report/generating UI acceptance |
| Phase 13 | OpenAPI 001/002/003 + backend-review report-owned read PASS | formal frontend 只消费 generated `getReportConversation`；不恢复 `listPracticeSessions` 或已删除 Demo runtime |
| Phase 12 | current OpenAPI `FeedbackReport.context.resumeId`、`summary` + `preparednessLevel` and frontend report owner | layout/navigation-only change；do not modify API、backend、persistence or model prompt |

## 3 质量门禁分类

- **Plan 类型**: user-visible UI + API consumer + contract migration + UX truthfulness。
- **TDD 策略**: `/implement` → `/tdd`；先在当前正式 frontend component/responsive tests 建立 `4/2/2/2/1`、上下文 link semantics、双列等高、唯一 summary 与 bottom full-width Overall Summary 的 RED，再做最小 UI/i18n/CSS 实现。
- **BDD 策略**: `BDD.REPORT.UI.001` 由代码层 owner tests 验证 generating/ready/failed、ready `4/2/2/2/1`、上下文链接、同行等高、replay/next/back 与 fail-closed 行为，并由仓库根 `make test` 统一回归；P0.099 仅作为真实 full-stack report/generating UI/API 与 exact-six screenshots 的独立 handoff，只有显式真实运行后才产生 PASS。
- **替代验证 gate**: source traceability、i18n exact set、typecheck/build、DOM order、computed-style/bbox/responsive/accessibility 与正式 frontend deterministic visual artifacts 作为独立 code/UI gates，不包装为 E2E。

## 4 Coverage Matrix

| Source | Category | Phase | Verification | UI anchor | Negative |
|--------|----------|-------|--------------|-----------|----------|
| spec C-1 | UX truthfulness | 6 | generating component/polling tests | screens-p0-complete::ReportGeneratingScreen | fake progress/live observation/notify |
| spec C-2 | primary/layout | 12 | component/DOM-order/responsive tests | ReportDashboard `4/2/2/2/1` | detached conversation entry、missing resume URL、unequal same-row panels、top readiness/summary、duplicate summary、bottom card not full width |
| spec C-3 | interaction | 7 | CTA variant/a11y tests | ReportHeader | fixed replay-primary styling |
| spec C-4 | business truth | 7 | request/server-projection contract tests | buildReplayPayload/startPractice | URL/client focus authority |
| spec C-5 | UX boundary | 12 | 1440/390 bbox/overflow/full-content tests | ContextStrip/DetailGrid/OverallSummary | unrecoverable ellipsis、wrong mobile order、two-column mobile |
| spec C-6 | formal implementation contract | 12 | formal frontend DOM/style/bbox/responsive visual gate | ReportDashboard | non-empty-buffer-only or parallel-prototype gate |
| spec C-7 | real integration | 12 | P0.099 full-page screenshots after implementation | real ReportDashboard | cropped evidence、missing bottom summary |
| spec C-8 | regression | 12 | repo-wide active negative scan | report/generating/scenarios/fixtures | stale question fields、fake-live、raw enum、old top summary layout |
| spec C-9 | business truth | 7 | deep-link/route-tamper tests | ReportScreen/ContextStrip/CTA | route status/identity authority |
| spec C-10 | i18n boundary | 7 | mixed UI/report language tests | chrome vs semantic content | client translation of model labels/evidence/actions |
| spec C-11 | replay boundary | 7 | empty/non-empty focus table tests | Replay CTA/direct contract | empty-focus rejection、unsupported non-empty focus normalization |
| spec C-12 | privacy/UX negative | 12 | sentinel-ID DOM/a11y tests + real-backend Chrome acceptance manifest | ContextStrip target/round/resume/interview record; frozen resume URL; privacy-preserving conversation action | visible session/report UUID; detached entry; missing resume URL; deleted internal API/CTA identity |
| spec C-13 | current-plan list | 10 | ReportsScreen target/overview table tests | ReportsScreen canonical round list | cross-target rows / full history / Parse list consumer |
| spec C-14 | navigation recovery | 10 | Report/Generating Back table tests | Back control | workspace-only back / route target authority / detail-screen list consumer |
| spec C-15 | command/read separation | 11 | ReportsScreen Back route tests + current-scope route negative | Reports Back control | Parse navigation/animation/import/poll; extra Workspace query params |
| spec C-16/C-17 | report conversation primary/failure | 13 | generated-client + component/route/Markdown tests; P0.099 real click/load/back handoff | formal `ReportConversationScreen` + ReportsScreen/Report entries | session list, live controls, third Header CTA, stale/partial transcript |

## 5 实施步骤

### Phase 1-5: Conversation-level baseline（已交付）

既有 data states、replay/next 与基础视觉回归保留为历史实现证据；Phase 6 起修订真实性和 direct semantic contract，当前 ready 信息层级由 Phase 12 取代旧布局结论。

### Phase 6: UI design document and honest generating

#### 6.1 Reconcile report truth sources

Phase 6 当时统一为三项 summary metrics + 四个常驻区块；该历史布局在 2026-07-15 被 Phase 12 明确取代，不再作为当前实现依据。

#### 6.2 Remove simulated runtime claims

正式 GeneratingScreen 删除固定百分比、自动完成阶段、固定观察流、“通知我”与“稍后从记录查看”承诺。queued/generating 自动轮询；timeout/network 才能继续检查；failed/not-found/invalid-contract/`REPORT_CONTEXT_TOO_LARGE` 为终态并只返回，不把再次 GET 伪装成重新生成。超限文案说明返回规划、缩短材料并开启新会话这一真实恢复方向。

单次`GenerateReport`动作内部为provider/protocol恢复执行initial+最多3次retry，等待10s/20s/40s；动作返回即销毁retry context，新动作从0开始。Runner的`async_jobs.attempts/max_attempts`只作基础设施lease/finalize，outbox/infra仍独立使用30s/2m/10m/1h/6h。Frontend只看服务端status，不展示attempt_count/retry_count/reason/scope或假进度。Polling用`maxAttempts=49`、1.5s×1.5、cap8s，总约6m04s；覆盖4×60s+10+20+40=5m10s并留约54s。窗口耗尽只能表达客户端等待超时并允许继续检查，不能改写为report failed。当前OpenAPI没有failed-report regenerate operation，本plan不新增或宣称Retry UI。

Visibility/focus暂停是同一poll run内的调度暂停，不是retry/reset边界。无论hidden/blur发生在timer等待还是request in-flight，poller都必须保存当前attempt与下一attempt/delay；恢复后从`n+1`继续且不重复n。只有reportId/client identity改变或用户显式“继续检查”才创建新run并重置count。重复hidden/visible/blur/focus不得产生并行请求，单run总调用仍`<=49`。

### Phase 7: Direct semantic dashboard and server-owned handoff

#### 7.1 Consume code+label direct report

接入 summary、immutable context、dimension code+label/status/confidence、dimensionCode evidence 与 retryFocusDimensionCodes；reportId 是唯一 locator，未知/缺失合同 fail closed，不前端补值。空 focus 是合法通用同轮 Replay；仅当 focus 非空时逐项校验其同时命中 `needs_work` dimension 与同 code issue，非法非空引用整份 fail closed。

#### 7.2 Localize and prioritize

status/confidence/readiness 与固定 chrome 全量 i18n；first next action 只改变两枚现有 CTA variant，disabled reason 可访问。模型生成的 summary/dimension/evidence/action label 按 report language 原样显示，即使 UI locale 不同也不翻译。

#### 7.3 Readability and responsive layout

长内容在 desktop/mobile 可靠换行。action schema200 code points 只作 malformed fuse；frontend 以 English 24 whitespace words / zh-CN 64 Unicode code points 为 semantic/UX gate，超界 ready payload 进入 typed invalid 且不回显 raw；合法边界在 1440x1200 与 390x844 完整换行。18/52 仅是上游 targeted-repair 内部余量，UI 不显示或校验为更小边界。

#### 7.4 Remove client focus authority

Replay/Next URL/request 删除 settings/identity/focus/evidence gaps；closed derived request 只有 goal + sourceReportId，后端派生 plan/round/focus。空 server focus 仍合法创建通用同轮 plan；非空 focus 的 issue-backed 合法性由 backend 与 frontend direct-contract gate 共同拒绝漂移。`context.hasNextRound` 控制 Next disabled，保留 fresh session 与重复点击锁。

### Phase 8: Deterministic visual regression and real acceptance

#### 8.1 Formal source/geometry/screenshot regression

正式 frontend 使用 deterministic fixture 固定 locale/timezone/Date/deviceScaleFactor=1，等待 `document.fonts.ready` 并关闭 animation/transition；验证 DOM、computed style、关键 bbox、390/1440 layout 与受控 screenshot baseline。失败保留正式 frontend 的 actual/expected/diff 证据，不以 buffer 非空收口，也不依赖平行 prototype 运行时。

#### 8.2 Full-page real UAT

当被显式运行时，P0.099 为当前 run 创建 en/zh ready rows 并捕获六图，不依赖 provider/eval output digest。每个 ready row 必须绑定 DB/API `canonical_report_content_digest`、`action_length_audit`、`content_audit`、`screenshot_sha256` 与 report/session/context digest。两张 390x844 mobile report `fullPage: true` 图完整覆盖 action 区域，证明当前合法 zh-CN / English label 完整可见、无 clipping/ellipsis/hidden content 且 `scrollWidth=390`。恰好 64/24 的边界由代码层 deterministic fixture tests 独立证明，不作为 E2E 步骤。本轮未运行，状态保持 `Ready`。

#### 8.3 Active stale negative

扩展 fixture/scenario/docs/runtime/lint 扫描，删除旧 question fields、fake-live copy、raw enum surface 与客户端 focus authority；历史 bug/report/journal 作为证据保留。

### Phase 9: User-visible internal locator removal

先更新 `frontend/src` 与 UI 文档，把 Context Strip 从 session/target/round/resume 收敛为 target/round/resume。随后 RED-GREEN 同步 `ReportContextStrip` 与 direct tests；session/report IDs 继续作为 generated API/内部动作关联事实，但不进入可见 DOM、title、tooltip 或 accessible name。

门禁：fixture 使用彼此不同的 `report.id` / `sessionId` sentinel，focused tests 逐值拒绝它们出现在 textContent、title/tooltip、`aria-*` 或 accessible name，同时正向证明 OpenAPI/report contract 仍要求合法 UUID、CTA 仍使用 `sourceReportId`。删除孤儿 `report.context.session` locale key 并做 zero-reference。Deterministic 1440/390 DOM/style/viewport component assertions 作为 frontend code gate 执行，不登记为 E2E。

面向用户的成功证据必须由 Chrome 从同一真实 backend ready report 的正式 frontend 另存到 `.test-output/acceptance/report-context-grid/`，不依赖场景临时输出。目录只允许三项：`report-desktop-1440x1200-full.png`（exact 1440x1200 viewport）、`report-mobile-390x844-full.png`（exact 390x844 viewport）与 `manifest.json`；两图都必须 `fullPage: true`，不得使用 fixture-only 页面、裁剪图或额外状态图。manifest 逐图记录 relative path、SHA-256、`state=ready`、viewport、`fullPage=true`、同一 report 的脱敏 locator digest、`reportSentinelAbsent=true`、`sessionSentinelAbsent=true`，并绑定 DOM/a11y/geometry audit；截图正向显示 target/round/resume/interview record，简历 URL 与 conversation action 行为通过，负向不存在 report/session sentinel。

### Phase 10: Independent current-plan reports list and Back recovery

#### 10.1 Independent route and state model

在正式 frontend 新增独立 ReportsScreen：页面保留 App chrome，route 为 `/reports?targetJobId=<uuid>`，但不加入 TopBar。它以 `getTargetJob(targetJobId)` 提供当前规划与 canonical round display，以 `listTargetJobReports(targetJobId)` 提供 current/latest；loading、empty、network/contract error、ready 均有可访问状态。Parse 只提供入口，不承载列表。

#### 10.2 Isolation, current/latest and no-history contract

RED-GREEN 覆盖两个 TargetJob 数据集、请求 target 精确绑定、跨用户/target mismatch、missing/extra/noncanonical rounds、target switch 首帧清旧链接与 stale response fence。每个 report locator 只归属一个 canonical round；同轮 current/latest 同 ID 只允许 latest ready，latest ready 不得缺 current，current 不得缺 latest。每轮只显示 current ready 和 latest attempt：queued/generating 链接 Generating，failed typed 且无 Retry，latest ready 与 current 同 ID 不重复；latest ready 与 current 不同只表达“最近一次生成已完成”，仍只提供 current report 入口，不扩展为历史版本列表。

#### 10.3 Back matrix and owner boundary

Phase 10 当时交付的 Reports Back destination 仅作为历史实现证据保留，当前 route contract 由 Phase 11 取代。缺失或非法 `/reports` target locator 以 `replaceRoute(workspace)` 安全替换当前 history entry，不得 push 后让 browser Back 反复重建坏链接。Report/Generating 的 ready、pending、failed、timeout/network 与正常 queued/generating 页面只要当前或最后可信 response 有 targetJobId，Back 都进入 `/reports?targetJobId=...`；missing reportId、404、首读网络失败或 invalid payload 无可信 identity 时回 workspace。Report/Generating route 保持 reportId-only，且二者不调用 `listTargetJobReports`；禁止从 route/title/reportId 推断 target identity。

Focused frontend tests 证明独立列表的 current-plan isolation、四态、1440/390 parity、ready report Back 以及 generating/failure/fallback；这些代码层断言不再组合成场景证据。

### Phase 11: Reports Back direct read-only workspace detail

#### 11.1 Separate command and read navigation

先以 ReportsScreen route/component tests 建立 RED：当 `getTargetJob` 与 overview 已闭合可信 `targetJobId` 时，Back 必须精确导航 `/workspace?targetJobId=<id>`；Workspace query 只保留 `targetJobId`，不携带 `resumeId`、`planId`、`reportId` 或 `section`。GREEN 后该链路直接进入只读规划详情，不挂载 Parse 命令/进度页，不展示解析动画，也不触发 JD import 或 parse polling。缺失或非法 Reports target 仍 replace 到 `/workspace` 列表。

#### 11.2 Preserve report recovery and prove current-route negatives

Report/Generating 恢复合同不变：当前/最后可信 API response 有 `targetJobId` 时 Back 到 `/reports?targetJobId=...`；missing reportId、404、首读网络失败或 invalid payload 无可信 identity 时回 `/workspace` 列表，route 仍为 reportId-only。Focused route/browser tests 覆盖 Reports Back 直达 Workspace 详情与 history，不经过 Parse；source gate 对当前实现和 owner docs 执行旧 Reports-to-Parse 正向引用零命中，并证明只读链没有 Parse animation/import/poll。

### Phase 12: Bottom full-width interview summary hierarchy

#### 12.1 RED: lock semantic ownership and order

在 `ConversationReport` owner tests 中锁定 ready DOM 顺序：Context Strip 四项、两个数量指标、Dimensions/Strength Evidence、Risks/Next Actions、最后一个 Overall Summary；`summary` 全页只出现一次，readiness 不在 Summary Metrics，Overall Summary 标题使用本地化“面试总评”。新增 RED 证明游离的 conversation entry、无 canonical resume URL 与未填满行高的 panel card 违反当前合同。

#### 12.2 GREEN: implement `4/2/2/2/1`

只修改正式 `ReportDashboard` / `ReportContextStrip`、必要样式与 zh/en locale：desktop 按 `4/2/2/2/1` 分组，把 conversation entry 并入 Context Strip 第四项；resume 使用 frozen `resumeId` 生成 canonical URL；两个 detail row 的 panel card 填满各自行高。Overall Summary 位于最低端并跨满内容网格；mobile 保持相同 DOM 顺序且各组单列。Overall Summary 展示 localized readiness tier 与服务端 `summary` 原文，不改写、不生成第二份摘要。现有 OpenAPI context/summary/preparedness 足够，本阶段禁止修改 backend、schema、fixture contract、persistence 或 prompt。

#### 12.3 Responsive, accessibility and current evidence

正式 frontend component/browser tests 在 1440 与 390 宽度验证卡片数量、DOM 顺序、desktop 全宽跨度、两组 detail panel 同行等高、mobile 单列、完整换行与无横向溢出；简历 URL、面试记录 action、标题、readiness 与 summary 均可访问，且 reportId 不进入 DOM 属性。`BDD.REPORT.UI.001` 的 ready 分支必须覆盖新层级。实现阶段同步对齐 P0.099 README/manual visual audit 与 capture/verification contract，使其后续完整矩阵明确检查四项上下文、两组响应式 panel、行动区之后的底部面试总评；本阶段的当前行为验收由一个真实 backend 持久化 ready report 的 Chrome desktop/mobile 截图、几何测量和两条导航实测闭环。由于本次不改变 generating/failure、语言状态或报告生成语义，不为 CSS/入口层级修订重复制造三套 provider 资源，也不把历史六图 PASS 当作当前截图证据。

### Phase 13: Integrate report-owned readonly conversation

#### 13.1 Contract and route

将 generated `getReportConversation` 与 `/report-conversation?reportId=...` 接入正式 frontend，路由只允许 `reportId`。报告 Context Strip 的第四个同级子项保留主入口，ReportsScreen 的 current report 行保留快捷入口；不改 Header 两 CTA 和 Phase 12 的 `4/2/2/2/1` 目标布局。

#### 13.2 Readonly projection and failure boundary

按 `sequence` 严格升序渲染 user/assistant 安全 Markdown/GFM；禁止 Composer、retry、thinking、pause、session/message/client IDs 与 browser storage。Report status 决定 Back 到 report 或 generating；hidden 404、invalid order/role/closed projection 与 stale response 整体 fail closed。

#### 13.3 Current integration gates

运行 report-conversation/ReportsScreen focused frontend tests、reports service/store/handler focused Go tests、OpenAPI/codegen/fixture drift gates 与根 `make test`。P0.099 脚本仍只驱动真实 API/UI，不嵌入 Go/Vitest/pytest/lint/build；未显式运行时不记录新 E2E PASS。

## 6 验收标准

- Generating 对用户只陈述真实状态和真实可用动作。
- Report desktop 按 `4/2/2/2/1` 展示：同一个 Context Strip 内有四个同级子项，简历具有 canonical URL，面试记录为 privacy-preserving SPA action；顶部两个数量指标、两行四个等高配对内容区、最低端全宽面试总评；readiness 与服务端 `summary` 仅在总评出现，用户看不到 raw enum/code。
- 推荐 action 与 CTA 主次一致，功能仍允许用户选择；retry focus 由服务端投影。
- 空 retry focus 不阻塞通用同轮 Replay；非空 focus 必须与 needs-work dimension 和 issue 一一闭合。
- Desktop/mobile 长内容完整可读；正式 frontend DOM/style/bbox/responsive visual gate 通过，mobile 保持同序单列。
- 当前 Chrome desktop+390 截图与几何/导航 audit 闭环真实合法 ready 内容的可读性；P0.099 完整矩阵合同同步承接相同检查，200-code-point fuse、exact 24/64 code test、18/52 repair margin 或 provider eval 都不能替代真实 UI。
- Report Context Strip 显示 target/round/resume/interview record 四项；resume 仅携带 frozen resumeId canonical URL，session/report UUID 不出现在用户可见或可访问 DOM 属性，既有 API frozen context 与 CTA 行为保持不变。
- Context Strip 当前正式验收使用同一 ready report 的 Chrome desktop / exact 390x844 两张 real UI full-page 图；四项上下文、resume URL、conversation action、report/session DOM sentinel absence 与响应式几何全部可校验。
- ReportsScreen 只展示当前 target 的 canonical rounds、current report 与 latest attempt，覆盖 loading/empty/error/identity mismatch/stale response；不展示其他规划或完整历史，desktop/mobile 响应式合同一致。
- ReportsScreen Back 直接进入 `/workspace?targetJobId=...` 只读规划详情，query 只有 `targetJobId`，不经过 Parse、不展示解析动画、不触发 import/poll；Report/Generating Back 在 trusted target context 存在时进入 `/reports?targetJobId=...`，缺失可信 identity 时进入 `/workspace` 列表；detail/generating route 保持 reportId-only，且列表 operation 只有 ReportsScreen 消费。
- Report 主入口与 ReportsScreen 快捷入口都打开同一 reportId-only 只读记录；queued/generating/ready/failed 可返回正确父页，失败/隐私边界不泄露 partial transcript 或内部 ID。

## 7 风险与应对

| 风险 | 应对措施 |
|------|----------|
| summary 被继续误当顶部准备度说明或重复展示 | 用 DOM 顺序与唯一文本断言锁定底部 Overall Summary 为唯一 owner；Summary Metrics 只允许两个数量指标 |
| CTA 主次被误当业务权限 | variant 只表达推荐；可用性仍由 round/state/replay lock 决定 |
| 截图 gate 假绿 | 强制正式 frontend DOM/style/bbox/responsive baseline 与 full-page artifact，并断言底部总评可见 |
| 用户验收截图来源或状态漂移 | 固定 formal real ready state、两种 exact viewport、两张固定文件名和逐图 SHA-256/DOM-a11y sentinel manifest；拒绝 fixture-only 页面、裁剪图和额外状态图 |
| 前端重新成为业务事实源 | URL/request negative gate，后端 source report integration proof |
| 空 focus 被误判为不可复练，或非法非空 focus 被静默删改 | 空数组显式正向 fixture；非空 cross-reference fail-closed table tests；前端不补默认 focus |
| 把动作内产品retry、基础设施attempt或客户端窗口误当服务端失败 | 锁定action-local report / runner lease / outbox infra三层所有权；maxAttempts49耗尽只进入可继续检查态，terminal failed只来自API |
| 列表混入其他规划或 stale response | `getTargetJob` 与 overview target/canonical round 双重闭合；target switch 首帧清旧 rows，request sequence fence 拒绝旧响应 |
| Back 使用旧 route 或标题猜测 TargetJob | 只接受当前/最后可信 API response 的 targetJobId；缺失即 workspace fallback，并对 route target 与 detail-screen list consumer 做负向测试 |
| 读取既有规划误入解析状态机 | Reports Back 固定为 targetJobId-only Workspace detail；route/browser tests 与 source negative 同时拒绝 Parse navigation、动画、import 和 polling |
| 并行分支合并恢复已删 Demo 或丢失会话入口 | 保留 Demo zero-directory gate，并以 route/list/detail tests 反向锁定两处入口和唯一 generated API |

## 8 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-15 | 3.8 | Revise Phase 12 to `4/2/2/2/1`: integrate conversation into Context Strip, link the frozen resume copy, and require equal-height desktop detail pairs with internal whitespace. |
| 2026-07-15 | 3.7 | Integrate the previously isolated report-owned readonly conversation implementation as Phase 13, preserving the active Phase 12 `3/2/2/2/1` layout contract and deleted Demo boundary. |
| 2026-07-15 | 3.6 | Reopen Phase 12 for the confirmed `3/2/2/2/1` report hierarchy: move readiness and the existing server summary into one bottom full-width interview-summary card, keep mobile order, replace stale prototype parity wording, and make no API/backend change. |
| 2026-07-14 | 3.5 | Separate code-owned report UI BDD from the explicitly run Ready-only P0.099 real acceptance handoff. |
| 2026-07-14 | 3.4 | Reopen Phase 11 so Reports Back opens the targetJobId-only read-only Workspace detail directly, without Parse animation/import/polling; preserve Report/Generating trusted recovery. |
| 2026-07-14 | 3.3 | Add report-locator ownership/cross-field fail-closed gates and replace-only recovery for invalid Reports deep links. |
| 2026-07-14 | 3.2 | Revise Phase 10 in place for an independent target-scoped ReportsScreen, current/latest-only isolation, complete states, trusted Back to `/reports`, and workspace fallback without API/schema changes. |
| 2026-07-14 | 3.1 | Add unchecked Phase 10 for trusted-target Back to Parse reports, workspace fallback, reportId-only routes and no reports-list consumer. |
| 2026-07-13 | 3.0 | Reopen in place to remove session/report UUID from the user-visible Context Strip and refresh deterministic desktop/mobile parity screenshots. |
| 2026-07-13 | 2.9 | Correct report timing ownership to action-local initial+3 with 10s/20s/40s; async attempts are infrastructure-only. Keep maxAttempts49 math and prohibit unsupported failed-report regenerate claims. |
| 2026-07-13 | 2.8 | L2：preserve poll attempt and next schedule across timer/in-flight hidden or blur；resume never resets to1 and one run remains capped at49. |
| 2026-07-13 | 2.7 | Lock report use of business10s/20s/40s under durable max4 and frontend maxAttempts49 (~6m04s)；separate business async cap80 from infra delivery and hide internal attempts/progress. |
| 2026-07-13 | 2.6 | Finalize A：wire fuse200；frontend semantic/UX 24 whitespace words / 64 Unicode code points；18/52 remains upstream targeted-repair margin；reopen boundary evidence. |
| 2026-07-13 | 2.5 | A-200 fuse；keep14/40 typed-invalid/no-raw gate and require desktop+390 complete wrapping. |
| 2026-07-13 | 2.4 | Bind each P0.099 ready row to its current-run canonical content/action/content-audit/screenshot/report/session/context digests；keep provider eval and deterministic boundary tests independent. |
| 2026-07-12 | 2.3 | Make empty focus a valid generic same-round replay and require issue-backed validation only for non-empty focus. |
| 2026-07-12 | 2.2 | Resolve Generating ownership, frozen report context/status truth, terminal action honesty, language split, dependency order and deterministic six-image acceptance. |
| 2026-07-12 | 2.1 | Reopen for honest generating, direct semantic summary/code+label, localized states, action-driven CTA priority, server-owned focus, responsive readability and strong screenshot acceptance. |
| 2026-07-12 | 2.0 | Reopen for conversation-level report and competency replay. |
