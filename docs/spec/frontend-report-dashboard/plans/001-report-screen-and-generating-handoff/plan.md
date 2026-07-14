# 001 — Honest Grounded Report Screen and Handoff

> **版本**: 3.3
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

在既有 report UI owner 内交付规划范围的独立 ReportsScreen、grounded direct semantic report 与诚实 generating：列表只展示当前 TargetJob canonical rounds 的 current report 与 latest attempt，不做全局中心或完整历史；详情保持三指标 + 四常驻区块，并修复 generating 伪实时语义、raw enum、CTA 推荐优先级、长内容/mobile 可读性与假强 parity gate。

## 2 Operation Matrix

| operationId | fixture | consumer | backend | persistence | AI | scenario |
|-------------|---------|----------|---------|-------------|----|----------|
| getTargetJob | `TargetJobs/getTargetJob.json` | ReportsScreen current target + canonical round display | targetjob handler/store | target_jobs.summary read | read none | P0.059 |
| listTargetJobReports | `Reports/listTargetJobReports.json` | ReportsScreen current/latest pointers only | reports handler/store | feedback_reports + canonical target read | read none | P0.059 |
| getFeedbackReport | `Reports/getFeedbackReport.json`: queued/generating/ready-needs-practice/ready-well-prepared/ready-empty-focus/failed/invalid-focus/long-content | generating/report; only status/context truth | reports handler/store | feedback_reports + frozen context | read none | P0.056/P0.057/P0.058/P0.059/P0.099 |
| createPracticePlan | `PracticePlans/createPracticePlan.json`: retry/next/mismatch | replay handler; no focus input | practice handler/store | practice_plans + source report projection | none | P0.057/P0.070/P0.072 |
| startPracticeSession | `PracticeSessions/startPracticeSession.json` | replay handler | practice handler/store | session/messages | practice.session.chat | P0.057 |

## 2.1 Owner / Dependency Order

| Gate | Dependency | Rule |
|------|------------|------|
| Phase 6 | ownership transfer recorded in `frontend/README.md` + workspace/practice spec | this plan exclusively edits prototype/formal GeneratingScreen |
| Phase 7.1 | `backend-review/001 6.1` + OpenAPI 001/002 codegen/fixture PASS | frontend RED/GREEN uses the real generated contract, not handwritten types |
| Phase 7.4 | `backend-practice/004 Phase 3` server-derived request/focus PASS | frontend removes all derived settings/identity/focus route authority |
| Phase 8 | backend 6-8 + frontend 6-7 PASS | scenario owners compose named markers; this plan owns 056-059 only |

## 3 质量门禁分类

- **Plan 类型**: user-visible UI + API consumer + contract migration + UX truthfulness。
- **TDD 策略**: /implement → /tdd；先修改 prototype/source tests，再以 component/hook/i18n/request negative/Playwright parity tests 建立 RED，最后迁移正式 frontend。
- **BDD 策略**: 原地加强 P0.056 happy report、P0.057 replay/next、P0.058 failure、P0.059 parity/negative、P0.099 real full-stack screenshots；P0.070/P0.072 验证服务端 focus。
- **替代验证 gate**: source traceability、i18n exact set、typecheck/build、computed-style/bbox/screenshot difference 与 full-page visual artifacts 补充 BDD。

## 4 Coverage Matrix

| Source | Category | Phase | Verification | UI anchor | Negative |
|--------|----------|-------|--------------|-----------|----------|
| spec C-1 | UX truthfulness | 6 | prototype/formal generating tests + P0.056 | screens-p0-complete::ReportGeneratingScreen | fake progress/live observation/notify |
| spec C-2 | primary/contract | 7 | component/API fixture tests | screen-report::ReportDashboard | raw enum/code, missing summary |
| spec C-3 | interaction | 7 | CTA variant/a11y tests + P0.057 | ReportHeader | fixed replay-primary styling |
| spec C-4 | business truth | 7 | request negative + P0.070/P0.072 | buildReplayPayload/startPractice | URL/client focus authority |
| spec C-5 | UX boundary | 7 | 1440/390 bbox/overflow/full-content tests | ContextStrip/DetailGrid | unrecoverable ellipsis/two-column mobile |
| spec C-6 | source parity | 8 | formal-vs-prototype DOM/style/bbox/screenshot diff | screen-report.jsx | non-empty-buffer-only gate |
| spec C-7 | real integration | 8 | P0.099 full-page screenshots | real ReportDashboard | cropped top-half-only evidence |
| spec C-8 | regression | 8 | repo-wide active negative scan | report/generating/scenarios/fixtures | stale question fields/fake-live/raw enum |
| spec C-9 | business truth | 7 | deep-link/route-tamper tests | ReportScreen/ContextStrip/CTA | route status/identity authority |
| spec C-10 | i18n boundary | 7 | mixed UI/report language tests | chrome vs semantic content | client translation of model labels/evidence/actions |
| spec C-12 | privacy/UX negative | 9 | sentinel-ID DOM/a11y tests + P0.059 + browser acceptance manifest | ContextStrip target/round/resume | visible session/report UUID; deleted internal API/CTA identity |
| spec C-13 | current-plan list | 10 | ReportsScreen target/overview table tests + P0.059 | ReportsScreen canonical round list | cross-target rows / full history / Parse list consumer |
| spec C-14 | navigation recovery | 10 | Report/Generating Back table tests + P0.058/P0.059 | Back control | workspace-only back / route target authority / detail-screen list consumer |

## 5 实施步骤

### Phase 1-5: Conversation-level baseline（已交付）

既有 prototype/formal data states、replay/next 与基础 parity 保留为历史；Phase 6 起修订真实性和 direct semantic contract。

### Phase 6: UI truth source and honest generating

#### 6.1 Reconcile report truth sources

统一 docs 与 prototype 为三项 summary metrics + 四个常驻区块，无 tab；readiness metric 增加服务端 summary，保留现有视觉语言。

#### 6.2 Remove simulated runtime claims

Prototype/Formal Generating 删除固定百分比、自动完成阶段、固定观察流、“通知我”与“稍后从记录查看”承诺。queued/generating 自动轮询；timeout/network 才能继续检查；failed/not-found/invalid-contract/`REPORT_CONTEXT_TOO_LARGE` 为终态并只返回，不把再次 GET 伪装成重新生成。超限文案说明返回规划、缩短材料并开启新会话这一真实恢复方向。

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

### Phase 8: Strong parity and real acceptance

#### 8.1 Source/geometry/screenshot parity

Playwright 使用同一 deterministic fixture 同时加载 prototype/formal：locale/timezone/Date/deviceScaleFactor=1 固定，等待 `document.fonts.ready`，关闭 animation/transition；分别验证 DOM、computed style、关键 bbox、390/1440 layout，并用 pixelmatch threshold 0.1、changed-pixel ratio ≤0.5% 判定 screenshot diff。失败保留 prototype/formal/diff 三件套，不以 buffer 非空收口。

#### 8.2 Full-page real UAT

P0.099 为当前 run 创建 en/zh ready rows 并捕获六图，不依赖 P0.100 output digest。每个 ready row 必须绑定 DB/API `canonical_report_content_digest`、`action_length_audit`、`content_audit`、`screenshot_sha256` 与 report/session/context digest。两张 390x844 mobile report `fullPage: true` 图完整覆盖 action 区域，证明实际 zh-CN / English label 满足 `<=64 Unicode code points` / `<=24 whitespace words`，完整可见、无 clipping/ellipsis/hidden content 且 `scrollWidth=390`。恰好 64/24 的换行由确定性 ui-design/OpenAPI fixture + prototype/formal pixel parity 独立证明。

#### 8.3 Active stale negative

扩展 fixture/scenario/docs/runtime/lint 扫描，删除旧 question fields、fake-live copy、raw enum surface 与客户端 focus authority；历史 bug/report/journal 作为证据保留。

### Phase 9: User-visible internal locator removal

先更新 `ui-design/src/screen-report.jsx` 与 UI 文档，把 Context Strip 从 session/target/round/resume 收敛为 target/round/resume。随后 RED-GREEN 同步 `ReportContextStrip` 与 direct tests；session/report IDs 继续作为 generated API/内部动作关联事实，但不进入可见 DOM、title、tooltip 或 accessible name。

门禁：fixture 使用彼此不同的 `report.id` / `sessionId` sentinel，focused tests 逐值拒绝它们出现在 textContent、title/tooltip、`aria-*` 或 accessible name，同时正向证明 OpenAPI/report contract 仍要求合法 UUID、CTA 仍使用 `sourceReportId`。删除孤儿 `report.context.session` locale key 并做 zero-reference。`E2E.P0.059` 更新 README/INDEX owner metadata 后执行 deterministic 1440/390 DOM/style/bbox/viewport/pixel parity；该场景保持 PASS cleanup 合同。

面向用户的成功证据必须由 `/agent-browser` 从同一真实 backend ready report 的 formal frontend 另存到 `.test-output/acceptance/report-context-strip/<run-id>/`，不依赖场景临时输出。目录只允许三项：`report-context-strip-desktop-1440x1200.png`（exact 1440x1200）、`report-context-strip-mobile-390x844.png`（exact 390x844）与 `manifest.json`；两图都必须 `fullPage: true`，不得使用 prototype、fixture-only、裁剪图或额外状态图。manifest 逐图记录 relative path、SHA-256、`state=ready`、viewport、`fullPage=true`、同一 report 的脱敏 locator/digest、`reportSentinelAbsent=true`、`sessionSentinelAbsent=true`，并链接 DOM/a11y negative audit；截图正向显示 target/round/resume，负向不存在 report/session sentinel。

### Phase 10: Independent current-plan reports list and Back recovery

#### 10.1 Prototype, independent route and state model

在 prototype/formal 新增独立 ReportsScreen：页面保留 App chrome，route 为 `/reports?targetJobId=<uuid>`，但不加入 TopBar。它以 `getTargetJob(targetJobId)` 提供当前规划与 canonical round display，以 `listTargetJobReports(targetJobId)` 提供 current/latest；loading、empty、network/contract error、ready 均有可访问状态。Parse 只提供入口，不承载列表。

#### 10.2 Isolation, current/latest and no-history contract

RED-GREEN 覆盖两个 TargetJob 数据集、请求 target 精确绑定、跨用户/target mismatch、missing/extra/noncanonical rounds、target switch 首帧清旧链接与 stale response fence。每个 report locator 只归属一个 canonical round；同轮 current/latest 同 ID 只允许 latest ready，latest ready 不得缺 current，current 不得缺 latest。每轮只显示 current ready 和 latest attempt：queued/generating 链接 Generating，failed typed 且无 Retry，latest ready 与 current 同 ID 不重复；latest ready 与 current 不同只表达“最近一次生成已完成”，仍只提供 current report 入口，不扩展为历史版本列表。

#### 10.3 Back matrix and owner boundary

ReportsScreen Back 精确返回 `parse?targetJobId=<current trusted id>`。缺失或非法 `/reports` target locator 以 `replaceRoute(workspace)` 安全替换当前 history entry，不得 push 后让 browser Back 反复重建坏链接。Report/Generating 的 ready、pending、failed、timeout/network 与正常 queued/generating 页面只要当前或最后可信 response 有 targetJobId，Back 都进入 `/reports?targetJobId=...`；missing reportId、404、首读网络失败或 invalid payload 无可信 identity 时回 workspace。Report/Generating route 保持 reportId-only，且二者不调用 `listTargetJobReports`；禁止从 route/title/reportId 推断 target identity。

P0.059 证明独立列表的 current-plan isolation、四态与 1440/390 parity，并覆盖 ready report Back；P0.058 证明 generating/failure/fallback。P0.016 仅证明 Parse 入口与零嵌入，不再组合列表证据。

## 6 验收标准

- Generating 对用户只陈述真实状态和真实可用动作。
- Report 三指标四区块完整展示 direct semantic summary/dimensions/evidence/actions，用户看不到 raw enum/code。
- 推荐 action 与 CTA 主次一致，功能仍允许用户选择；retry focus 由服务端投影。
- 空 retry focus 不阻塞通用同轮 Replay；非空 focus 必须与 needs-work dimension 和 issue 一一闭合。
- Desktop/mobile 长内容完整可读；formal/prototype DOM/style/bbox/screenshot difference gate 通过。
- P0.099 desktop+390 截图与 current-run audit 闭环合法24/64可读性；200-code-point fuse、18/52 repair margin 或P0.100内容PASS都不能替代。
- Report Context Strip 只显示 target/round/resume；session/report UUID 不出现在用户可见或可访问 UI，既有 API frozen context 与 CTA 行为保持不变。
- Context Strip 正式验收只有同一 ready report 的 exact 1440x1200 / 390x844 两张 formal real UI full-page 图与固定 manifest；path/hash/state/viewport/fullPage、target/round/resume 可见和 report/session DOM/a11y/screenshot sentinel absence 全部可校验。
- ReportsScreen 只展示当前 target 的 canonical rounds、current report 与 latest attempt，覆盖 loading/empty/error/identity mismatch/stale response；不展示其他规划或完整历史，desktop/mobile 与原型一致。
- ReportsScreen Back 返回当前规划详情；Report/Generating Back 在 trusted target context 存在时进入 `/reports?targetJobId=...`，缺失可信 identity 时进入 workspace；detail/generating route 保持 reportId-only，且列表 operation 只有 ReportsScreen 消费。

## 7 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 为修复 UX 重新设计页面 | 保留三指标四常驻区块，只做原型先行的真实性/可读性修订 |
| CTA 主次被误当业务权限 | variant 只表达推荐；可用性仍由 round/state/replay lock 决定 |
| 截图 gate 假绿 | 强制 prototype/formal 双端 DOM/style/bbox/diff 与 full-page artifact |
| 用户验收截图来源或状态漂移 | 固定 formal real ready state、两种 exact viewport、两张固定文件名和逐图 SHA-256/DOM-a11y sentinel manifest；拒绝 prototype、裁剪图和额外状态图 |
| 前端重新成为业务事实源 | URL/request negative gate，后端 source report integration proof |
| 空 focus 被误判为不可复练，或非法非空 focus 被静默删改 | 空数组显式正向 fixture；非空 cross-reference fail-closed table tests；前端不补默认 focus |
| 把动作内产品retry、基础设施attempt或客户端窗口误当服务端失败 | 锁定action-local report / runner lease / outbox infra三层所有权；maxAttempts49耗尽只进入可继续检查态，terminal failed只来自API |
| 列表混入其他规划或 stale response | `getTargetJob` 与 overview target/canonical round 双重闭合；target switch 首帧清旧 rows，request sequence fence 拒绝旧响应 |
| Back 使用旧 route 或标题猜测 TargetJob | 只接受当前/最后可信 API response 的 targetJobId；缺失即 workspace fallback，并对 route target 与 detail-screen list consumer 做负向测试 |

## 8 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-14 | 3.3 | Add report-locator ownership/cross-field fail-closed gates and replace-only recovery for invalid Reports deep links. |
| 2026-07-14 | 3.2 | Revise Phase 10 in place for an independent target-scoped ReportsScreen, current/latest-only isolation, complete states, trusted Back to `/reports`, and workspace fallback without API/schema changes. |
| 2026-07-14 | 3.1 | Add unchecked Phase 10 for trusted-target Back to Parse reports, workspace fallback, reportId-only routes and no reports-list consumer. |
| 2026-07-13 | 3.0 | Reopen in place to remove session/report UUID from the user-visible Context Strip and refresh deterministic desktop/mobile parity screenshots. |
| 2026-07-13 | 2.9 | Correct report timing ownership to action-local initial+3 with 10s/20s/40s; async attempts are infrastructure-only. Keep maxAttempts49 math and prohibit unsupported failed-report regenerate claims. |
| 2026-07-13 | 2.8 | L2：preserve poll attempt and next schedule across timer/in-flight hidden or blur；resume never resets to1 and one run remains capped at49. |
| 2026-07-13 | 2.7 | Lock report use of business10s/20s/40s under durable max4 and frontend maxAttempts49 (~6m04s)；separate business async cap80 from infra delivery and hide internal attempts/progress. |
| 2026-07-13 | 2.6 | Finalize A：wire fuse200；frontend semantic/UX 24 whitespace words / 64 Unicode code points；18/52 remains upstream targeted-repair margin；reopen boundary evidence. |
| 2026-07-13 | 2.5 | A-200 fuse；keep14/40 typed-invalid/no-raw gate and require desktop+390 complete wrapping. |
| 2026-07-13 | 2.4 | Bind each P0.099 ready row to its current-run canonical content/action/content-audit/screenshot/report/session/context digests；keep P0.100 independent and exact 14/40 in deterministic parity. |
| 2026-07-12 | 2.3 | Make empty focus a valid generic same-round replay and require issue-backed validation only for non-empty focus. |
| 2026-07-12 | 2.2 | Resolve Generating ownership, frozen report context/status truth, terminal action honesty, language split, dependency order and deterministic six-image acceptance. |
| 2026-07-12 | 2.1 | Reopen for honest generating, direct semantic summary/code+label, localized states, action-driven CTA priority, server-owned focus, responsive readability and strong screenshot acceptance. |
| 2026-07-12 | 2.0 | Reopen for conversation-level report and competency replay. |
