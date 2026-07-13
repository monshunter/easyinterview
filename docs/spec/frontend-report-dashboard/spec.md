# Frontend Report Dashboard Spec

> **版本**: 1.21
> **状态**: active
> **更新日期**: 2026-07-13

## 1 背景与目标

`frontend-report-dashboard` 承接一次连续模拟面试结束后的诚实生成态与证据化报告。页面帮助用户理解总体准备度、能力维度、优势/风险证据和下一步建议，并从当前轮复练或进入下一轮。

报告不按题目组织，也不展示伪精确分数。前端只展示服务端持久化的最终语义，不自行拆分维度、推断 readiness、改写证据或生成建议。

## 2 范围

### 2.1 In Scope

- `generating`：轮询真实 report status，展示诚实的异步等待说明；不伪造百分比、实时观察或通知订阅。
- `report`：Header、Context Strip、三项 Summary Metrics、四个常驻内容区（Dimensions / Strength Evidence / Risks / Next Actions）。
- Readiness metric 同时展示服务端 `summary`，避免用户只看到一个档位而无解释。
- Dimension 使用动态 `label`，status/confidence 走完整 zh/en i18n，不泄漏 raw enum/code。
- Header 保留唯一一对 CTA；`nextActions[0].type` 决定现有按钮主次，不新增 CTA。
- Replay focus 由后端 source report 投影；URL/前端 request 不承载 focus/evidence-gap 业务事实。
- `reportId` 是唯一 locator；status/error、Context Strip 和 CTA identity 全部来自 `getFeedbackReport` 的 frozen `context`，route 中冲突值一律忽略。
- 长上下文可读、明确 mobile 单列、keyboard/a11y、desktop/mobile full-page parity。

### 2.2 Out of Scope

- Questions tab、题目回顾、逐题评分、题号/总题数、per-question replay。
- candidate numeric score、录用概率、排名、timeline、独立错题本。
- 前端推导报告语义或在客户端持久化报告业务状态。
- 真正的通知订阅；若未来需要，必须由独立后端通知合同承接。

## 3 用户决策

| ID | 决策 | 当前结论 |
|----|------|----------|
| D-1 | 报告粒度 | 整场 conversation，不按题目/turn |
| D-2 | 页面骨架 | 三项指标 + 四个常驻区块，无 tab；以当前原型源码为准 |
| D-3 | 报告语义 | LLM direct semantic API；前端只展示 summary、label/status/confidence、evidence、actions |
| D-4 | CTA 推荐 | first next action 只切换两枚既有 CTA 主次，用户仍可自主选择 |
| D-5 | Replay focus | 后端 source report 是唯一事实源，客户端不透传 focus |
| D-6 | Generating | 仅表达真实 queued/generating 状态，不伪造阶段进度、实时证据或通知 |
| D-7 | 页面 owner | GeneratingScreen 与 ReportScreen 均由本 owner 独占；practice owner 只负责 stable reportId handoff |
| D-8 | 语言/长度职责 | 上游 schema200 code points 只作 malformed fuse。Frontend 按 report language 镜像 English 24 whitespace words / zh-CN 64 Unicode code points；English delimiter 以 ECMAScript `/\s/u` 为唯一口径（含 U+FEFF、不含 U+0085），backend/evalkit 必须精确同构。超界 ready payload 进入 typed invalid 且不回显 raw label。targeted repair 的18/52只是上游内部余量，不是 UI 边界。合法边界在desktop+390完整换行；不截断/ellipsis/改写 |
| D-9 | Replay focus | `retryFocusDimensionCodes=[]` 是合法的通用同轮复练，不因空 focus 禁用 Replay；只有非空 focus 才要求每个 code 同时引用 `needs_work` dimension 与至少一条同 code issue，非法非空引用按 direct-contract failure fail closed |
| D-10 | Generating 等待窗口 | 单次`GenerateReport`动作在后端调用内执行initial+最多3次retry并等待10s/20s/40s；动作结束销毁retry context，新的独立动作清零，`async_jobs.attempts/max_attempts`仅作基础设施执行。Frontend固定`maxAttempts=49`、初始1.5s、×1.5、cap8s，总约6m04s，覆盖4×60s+70s=5m10s并留约54s；queued/generating不展示attempt/retry/progress，窗口耗尽只进入可继续检查态，不伪装服务端failed |
| D-11 | Poll pause/resume | hidden/blur只暂停同一poll run；timer等待与in-flight请求都保存current/next attempt和delay，visible/focus从n+1继续，不回1、不重复n。单run累计调用<=49；只有显式continue-check或reportId/client identity变化重置 |
| D-12 | Context Strip 正式截图验收 | 每次验收只保留同一 ready report 的两张 formal real UI `fullPage: true` 图：1440x1200 与 390x844；固定目录、文件名与 manifest schema，逐图绑定 SHA-256、ready state、viewport/fullPage 和 report/session sentinel DOM/a11y absence。Prototype、裁剪图、额外状态图不能替代或混入这组成功证据。 |

## 4 UI 真理源

- `ui-design/src/screen-report.jsx::ReportScreen`
- `ui-design/src/screens-p0-complete.jsx::ReportGeneratingScreen`
- `docs/ui-design/report-dashboard.md`
- `docs/ui-design/module-practice-review.md`

正式前端必须源级复刻修订后的 prototype。验收拆为 DOM/control/a11y、computed style/bounding box/responsive、formal-vs-prototype screenshot difference 和真实 full-page UAT；非空 screenshot buffer 不再作为 parity 完成依据。

## 5 Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `getFeedbackReport` | `Reports/getFeedbackReport.json`: queued/generating/ready-needs-practice/ready-well-prepared/ready-empty-focus/failed/invalid-focus/long-content | generating poll + ReportDashboard；唯一状态/上下文事实源 | backend-review reports handler/store | `feedback_reports` + frozen context | read none | `E2E.P0.056`, `E2E.P0.057`, `E2E.P0.058`, `E2E.P0.059`, `E2E.P0.099` |
| `createPracticePlan` | `PracticePlans/createPracticePlan.json`: retry/next/mismatch | replay/next CTA；不发送 focus | backend-practice handler/store | `practice_plans` + source report projection | none | `E2E.P0.057`, `E2E.P0.070`, `E2E.P0.072` |
| `startPracticeSession` | `PracticeSessions/startPracticeSession.json` | replay/next CTA | backend-practice handler/store | session + opening message | `practice.session.chat` | `E2E.P0.057` |

Frontend Phase 7.1 必须等待 `backend-review/001 6.1` 的 generated contract；Phase 7.4 必须等待 `backend-review/001 8.1` 的 server-owned projection。`getTargetJob/getResume` 不属于本页读取链：冻结 label/identity 直接随 report 返回，避免深链刷新读取可变实体。

## 6 页面结构

### 6.1 Generating

```text
GeneratingScreen
├─ honest queued/generating status
├─ static explanation of analysis work
├─ polling / timeout / typed failure
└─ Continue checking (timeout/network only) / Back to Home
```

- 不显示与后端无关的百分比或自动逐项完成动画。
- 不显示固定“实时观察”内容。
- 不承诺“好了通知我”或“稍后从记录查看”；当前没有 records consumer。
- queued/generating 自动继续检查；timeout/network 允许“继续检查”重启轮询；failed/not-found/invalid-contract/`REPORT_CONTEXT_TOO_LARGE` 是终态，只提供返回，不把再次 GET 伪装成重新生成。超限文案应说明本次材料与对话过长、可返回规划后缩短输入并开启新会话，不承诺当前 report 可重试恢复。
- Report generation的10s/20s/40s是单次后端动作内的等待，不来自runner business requeue；`async_jobs.attempts/max_attempts`与outbox/infra投递的30s/2m/10m/1h/6h都不属于本页面的产品retry事实。Frontend只消费queued/generating/ready/failed，不展示内部attempt_count/retry_count/reason/scope。当前OpenAPI没有failed report regenerate operation，因此本页不得宣称或新增同report Retry入口；独立动作清零只由后端两次`GenerateReport` invocation证明。
- Polling 固定为`maxAttempts=49`、1.5s初始退避、乘数1.5、8s cap，总窗口约6m04s。它覆盖report最坏4×60s provider timeout +10+20+40s=5m10s并保留约54s调度/网络余量；不得恢复旧30次窗口。窗口耗尽不改变服务端状态，只显示可继续检查的typed client timeout。

### 6.2 Report Dashboard

```text
ReportDashboard
├─ Back
├─ Header
│  ├─ title / subtitle
│  ├─ replay current round
│  └─ next round
├─ ContextStrip
│  ├─ target
│  ├─ round
│  └─ resume
├─ SummaryMetrics
│  ├─ readiness + summary
│  ├─ dimension count
│  └─ evidence count
└─ DetailGrid
   ├─ Dimensions
   ├─ Strength Evidence
   ├─ Risks
   └─ Next Actions
```

### 6.3 可读性与响应式

- target/round/resume 只读 frozen report context，可换行或通过 title/accessible description 查看完整值。session/report UUID 等内部 locator 不渲染为用户字段，也不通过 title、tooltip 或 accessible description 暴露。
- Desktop detail 使用双列；mobile 390px 明确单列，长 label/evidence/action 不横向溢出。
- Frontend consumer 在 render 前执行 24/64 semantic boundary；English 按 ECMAScript `/\s/u` whitespace words（U+FEFF 是 delimiter，U+0085 不是）、zh-CN 按 Unicode code points 计数。超界 payload 不进入 ReportDashboard，不得利用 CSS 截断把 invalid 内容伪装为可用。
- Deterministic ui-design/OpenAPI fixture 使用恰好 24-word/64-code-point actions；prototype/formal 1440x1200 与 390x844 full-page parity 均覆盖 action 区域并证明完整换行。200-code-point malformed fixture只证明 typed invalid/no-raw-output，不得替代 UX gate；18/52 targeted-repair margin 也不得替代边界 fixture。
- Context Strip 用户验收固定写入 `.test-output/acceptance/report-context-strip/<run-id>/`，且成功目录只包含 `report-context-strip-desktop-1440x1200.png`、`report-context-strip-mobile-390x844.png` 与 `manifest.json`。两图必须来自同一 formal frontend 的真实 backend ready report，分别使用 exact viewport 1440x1200 / 390x844 和 `fullPage: true`；不得用 `ui-design` prototype、fixture-only 页面、裁剪图或额外 loading/error 图冒充。
- `manifest.json` 必须逐图记录相对路径、SHA-256、`state=ready`、viewport width/height、`fullPage=true`、同一 report 的脱敏 locator/digest，以及 `reportSentinelAbsent=true`、`sessionSentinelAbsent=true`；同时绑定该页面的 DOM/a11y negative audit，证明 report/session sentinel 在 text、title/tooltip、任意 `aria-*` 与 accessible name 中均不存在。截图中 target/round/resume 必须可见，且 report/session sentinel 不能以用户文案、调试标记或可访问名称出现。
- status/confidence、readiness、CTA chrome、empty/error/loading 随 UI locale 本地化；summary、dimension label、evidence 与 action label 按 report language 原样展示。未知 enum fail closed 到 typed error，不回显 raw token。

### 6.4 CTA

- `retry_current_round` 为 first action 时 Replay 使用 accent，Next 为 secondary。
- `next_round` 为 first action 时 Next 使用 accent，Replay 为 secondary。
- `review_evidence` 时两者均 secondary，正文行动建议保持可见。
- next round unavailable 时 disabled 并提供可访问原因；两条启动路径仍共享重复点击锁。
- `retryFocusDimensionCodes=[]` 不代表报告或 Replay 无效：Replay 仍创建通用同轮 plan。非空 focus 才做 issue-backed 校验；任一 code 未命中 `needs_work` dimension 或同 code issue 时整份 ready contract fail closed，前端不删错项、不猜 focus。
- derived request 只带 goal + sourceReportId；persona/difficulty/language/time budget/target/resume/round/focus 全由后端从 source report 派生。`context.hasNextRound=false` 时下一轮 disabled；route 中 identity/status/error 即使存在也不得覆盖 API。

## 7 状态与错误

- 缺 session/report：专用空态，不展示假报告。
- queued/generating：留在 honest generating。
- ready：渲染完整 direct semantic dashboard。
- timeout/network：typed recoverable error、继续检查/back；failed/not found/unknown enum/invalid contract/`REPORT_CONTEXT_TOO_LARGE`：typed terminal error、back，不显示虚假 Retry；超限文案保持本地化且给出返回规划后的可执行方向。
- empty dimensions/evidence 或缺 summary：视为后端合同失败，不由前端补假数据。
- 空 `retryFocusDimensionCodes` 单独不构成合同失败；非法非空 focus cross-reference 才进入 typed invalid-contract 终态。

## 8 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | honest generating | report queued/generating | 页面轮询 | 无假百分比、假观察、假通知 | 001 |
| C-2 | ready dashboard | direct semantic report | 打开 report | 三指标四区块，summary 与本地化语义完整 | 001 |
| C-3 | recommended action | retry/next/review first action | 查看 Header | 仅切换现有 CTA 主次且功能可用 | 001 |
| C-4 | server-owned replay | report 含 retry focus | 点击复练 | request 不传 focus，服务端 plan/session 得到 focus | 001 |
| C-5 | long/mobile | 长 target/round/resume/evidence | desktop/mobile 打开 | 完整可读、mobile 单列、无横向溢出 | 001 |
| C-6 | deterministic boundary parity | ui-design/OpenAPI fixtures 含恰好 24-whitespace-word / 64-Unicode-code-point actions | 运行 browser gate | prototype/formal DOM/style/bbox/viewport/pixel difference 通过，边界 label 完整换行且无截断/省略/横溢 | 001 |
| C-7 | current-run canonical mobile UAT | P0.099 当前 run 的 en/zh ready rows | exact six images | 每个 row 绑定 DB/API `canonical_report_content_digest`、`action_length_audit`、`content_audit`、`screenshot_sha256` 与 report/session/context digest；390x844 report 图覆盖 action 区域，实际 `<=24-word` / `<=64-code-point` label 完整可见且无截断/省略/横溢 | 001 |
| C-8 | stale negative | 全仓 active assets | 扫描 | 无 raw enum UI、fake-live copy、客户端 focus 与旧 question fields | 001 |
| C-9 | route tamper / deep link | 只有 reportId，或 route 带冲突 status/target/resume/round | 刷新/读取/点击 CTA | API frozen context/status 获胜，route 不能改变展示与动作 | 001 |
| C-10 | language split | UI locale 与 report language 不同 | 查看报告 | chrome 随 UI locale；模型语义原文不翻译、不改写 | 001 |
| C-11 | empty / issue-backed focus | ready report 的 focus 为空，或非空 focus 引用 needs-work/issue | 点击 Replay / 校验 direct contract | 空 focus 合法创建通用同轮复练；非空 focus 必须逐项 issue-backed，非法引用 fail closed | 001 |
| C-12 | internal locator cleanup + formal acceptance | 同一真实 backend ready report 含 distinct frozen session/report sentinel | 以 formal frontend 分别在 exact 1440x1200 / 390x844 执行 `fullPage: true` 捕获并生成固定 manifest | Context Strip 只显示 target/round/resume；成功目录只有固定两图+manifest；逐图 path/hash/state/viewport/fullPage 与 DOM/a11y sentinel absence 闭合，API 事实源与 CTA 行为不变 | 001 |

## 9 关联计划

- [001-report-screen-and-generating-handoff](./plans/001-report-screen-and-generating-handoff/plan.md)

## 10 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-13 | 1.21 | Remove session/report UUID from the user-visible Context Strip while retaining API-frozen context as the internal action authority. |
| 2026-07-13 | 1.20 | Supersede 1.18 timing ownership: 10s/20s/40s belong to one action-local GenerateReport retry context; async job attempts are infrastructure-only. Keep maxAttempts49/6m04s honest polling and do not claim a failed-report regenerate API. |
| 2026-07-13 | 1.19 | L2：in-flight/timer pause preserves attempt and schedule；resume never resets1 or exceeds49；run35622 is aborted7/11 not PASS. |
| 2026-07-13 | 1.18 | 锁定maxAttempts49（约6m04s），覆盖report复用business policy的10s/20s/40s与4×60s调用；分离business async与infra投递退避，不展示内部attempt/progress。 |
| 2026-07-13 | 1.17 | 方案 A 最终边界：schema fuse200；frontend semantic/UX 24 whitespace words / 64 Unicode code points；18/52仅为上游targeted-repair余量；desktop+390合法边界完整换行。 |
| 2026-07-13 | 1.16 | A-200：wire fuse改200；14/40仍为frontend typed-invalid gate，desktop+390合法边界完整换行，超限不回显raw。 |
| 2026-07-13 | 1.15 | 区分 120-char malformed fuse、P0.099 current-run canonical screenshot audit 与 deterministic 14/40 boundary parity；解除对 P0.100 output digest 的人为依赖。 |
| 2026-07-12 | 1.14 | 明确空 focus 是合法通用同轮 Replay；非空 focus 才执行 needs-work + issue-backed cross-reference 校验。 |
| 2026-07-12 | 1.13 | 将 GeneratingScreen 唯一 owner 转入本模块；补 immutable report context、route tamper、终态动作矩阵与 UI/report 双语言边界。 |
| 2026-07-12 | 1.12 | 重新打开 001：锁定三指标四常驻区块，接入 direct semantic summary/code+label，删除 generating 伪实时语义，增加 i18n/CTA/mobile/readability 与真实截图 gate。 |
| 2026-07-12 | 1.11 | 报告改为会话级四部分，删除逐题模型、hint/phone 展示与 turn-based replay。 |
