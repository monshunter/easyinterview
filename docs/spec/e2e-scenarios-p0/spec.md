# E2E Scenarios P0 Spec

> **版本**: 2.12
> **状态**: active
> **更新日期**: 2026-07-13

## 1 目标

为当前核心漏斗维护两层证据：可重复的跨层契约组合门禁，以及共享真实环境中的浏览器验收。当前核心漏斗是：简历与 JD 就绪 → 创建 practice plan → 连续聊天 → 完成会话 → 生成 conversation-level report。

## 2 当前产品不变量

- Practice 是一个连续聊天窗口，不存在题目侧栏、题号计数、当前题目、逐题分类或题目预算。
- `startPracticeSession` 创建开场 assistant message；`sendPracticeMessage` 追加普通 user/assistant message pair。
- `practice_messages` 是聊天内容真理源；`practice_session_events` 只保留 session started/completed 生命周期事实。
- Report 以完成时 frozen context 和整场对话生成 summary、code+label dimensions、grounded evidence/risks/actions，不生成逐题 assessment；`reportId` 是浏览器唯一 locator。
- Voice 暂时 fail-closed：前端原生 disabled，后端返回 unsupported capability，且不触达 provider 或持久化音频。

## 3 场景分层

| ID | Owner plan | 类型 | 目的 | 环境 |
|----|------------|------|------|------|
| `E2E.P0.098` | `001-full-funnel-happy-journey` | automated | 组合当前 handler/service/store/report persistence/retry 契约 | repo tests；不维护场景专属后端 |
| `E2E.P0.099` | `001-full-funnel-happy-journey` | hybrid | 验证真实 Mailpit 登录、真实前后端、PostgreSQL 与精确六图 UI 证据 | shared host-run environment |
| `E2E.P0.100` | `002-manual-uat-real-provider-full-funnel` | hybrid | 真实 provider 内容可靠性、context-aware judge、逐项因果核对与 3x 临界重复 | shared host-run environment |

## 4 设计决策记录

| ID | 决策 | 理由 |
|----|------|------|
| D-1 | 删除 P0.099 专属 Playwright test server/config | 对应 Go server 已不存在；继续保留会形成第二套漂移运行时 |
| D-2 | P0.098 使用当前 focused gates 组合 | 复用 owner 测试，不复制业务 orchestration |
| D-3 | P0.099 使用 agent-browser/human hybrid 真实验收 | 能验证真实登录、provider、DB 与可见 UI，并留存截图 |
| D-4 | 共享环境由顶层 `test/scenarios/env-*.sh` 管理 | 场景目录不私有化环境 bootstrap |
| D-5 | P0.099 使用精确六图非笛卡尔矩阵 | 同时证明 zh/en、needs-practice/well-prepared、desktop/mobile、generating 与长内容，不用少量截图冒充覆盖 |
| D-6 | 报告内容可靠性不能只验 schema，唯一 owner 为 plan 002 | 每个样本仍按 fact→judgment→action 的 supported/partial/unsupported、阈值与 zero-tolerance 判定；产品验收以固定五类代表场景至少 4/5（80%）表达概率性语义置信度，不替换失败样本；plan 001 不重复承接其 Phase/BDD |
| D-7 | P0.100 bounded retry + fencing | 每次product `GenerateReport` invocation独立拥有initial+最多3次retry与10s/20s/40s，返回销毁且下一invocation清零；`async_jobs.attempts/max_attempts`仅作基础设施执行；running+claimed-attempt fence持久化副作用；frontend in-flight pause恢复n+1且单run cap49；evalkit generation/judge各自max4。run25849(10/11)与run35622(7/11)均保留为aborted/not-PASS |
| D-8 | P0.099/P0.100 最小充分分界 | 200-code-point 只作 malformed-output fuse；所有最终输出先通过24/64及完整机械合同。固定五类代表场景承接约80%产品语义验收；P0.100 仍保留更严格的5类/11次、关键3/3与blind-review诊断，失败不得冒充PASS。18/52仅是targeted-repair余量。P0.099 不消费其 output digest；每个 ready row 绑定 current-run DB/API `canonical_report_content_digest`、`action_length_audit`、`content_audit`、`screenshot_sha256` 与 report/session/context digest，desktop/390 action 区域完整可见且无截断/省略/横溢 |

## 5 Operation Matrix

| operationId | fixture | consumer | persistence | AI | gate |
|-------------|---------|----------|-------------|----|------|
| `registerResume` | current fixture | Resume Workshop | resumes/jobs | `resume.parse.default` | P0.099/P0.100 |
| `importTargetJob` | current fixture | Home/Parse | target jobs/jobs | `target.import.default` | P0.099/P0.100 |
| `createPracticePlan` | current fixture | Parse/Workspace/Report CTA | `practice_plans` | none | P0.022/P0.098/P0.099/P0.100 |
| `startPracticeSession` | current fixture | Practice entry | sessions/messages/lifecycle/outbox | `practice.chat.default` | P0.023/P0.098/P0.099/P0.100 |
| `sendPracticeMessage` | current fixture | Practice chat | `practice_messages`, task-runs | `practice.chat.default` | P0.044/P0.046/P0.098/P0.099/P0.100 |
| `completePracticeSession` | current fixture | Finish report | session/report/job/lifecycle/outbox | report job | P0.047/P0.098/P0.099/P0.100 |
| `getFeedbackReport` | current fixture | Generating/Report | `feedback_reports` | none on read | P0.056/P0.058/P0.099/P0.100 |

## 6 验收标准

- P0.098 在当前代码上运行真实测试项且拒绝 no-test/skip-as-pass。
- P0.099 只在 exact 六张 full-page 截图的每个 ready row 绑定当前 run 的 DB/API `canonical_report_content_digest`、`action_length_audit`、`content_audit`、`screenshot_sha256` 与 report/session/context digest 时 PASS；跨 run、只换 screenshot 或缺 canonical audit 的 row 无效。
- 六图中的两张 390x844 report 图完整覆盖 action 区域，实际 English / zh-CN label 分别满足 `<=24 whitespace words` / `<=64 Unicode code points` 且完整可见，无 clipping/ellipsis/hidden content/横溢。恰好 24/64 只由确定性 ui-design/OpenAPI fixture + prototype/formal pixel parity 证明。
- P0.100 只由 plan 002 承接。Product generation每次invocation新建内存retry context，最多3 retry、精确等待10s/20s/40s并在返回时销毁；新的独立invocation从0开始，async job attempt不参与产品计数。Evalkit分别维护独立generation/judge max4 budgets；judge只对provider/protocol invalid重试，valid negative终端FAIL。每个最终输出的机械合同必须100%通过；五类固定代表场景至少4/5才满足当前产品语义置信度，不因失败替换样本。
- 更严格的 P0.100 5类/11次、关键3/3与blind-review继续作为稳定性诊断；它只有全部完成才输出PASS。最终 prompt run `e2e-p0-100-20260713T101214Z-59381` 为机械9/9、语义8/9、代表场景4/5，满足当前产品验收但严格场景保持FAIL。该诊断不向P0.099提供output-digest前置条件；200-code-point schema PASS或18/52 repair margin不得冒充内容质量或UX PASS。
- active runtime/scenario 资产不存在 append-event、question counter/sidebar/current-question 合同。
- 真实凭证、邮箱验证码、cookie、完整 prompt/response 不进入 tracked docs 或验收 evidence。

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-13 | 2.11 | Supersede durable report/job max4 semantics: product retry is per-GenerateReport invocation initial+3 with10s/20s/40s, destroyed on return and reset for a new action; async attempts remain infrastructure-only while lease side-effect fencing stays current. |
| 2026-07-13 | 2.10 | L2：add report job max4、lease-generation fencing and frontend in-flight poll-resume gates；run35622 aborted7/11 not PASS. |
| 2026-07-13 | 2.9 | User-approved independent max4 generation/judge budgets；durable product cap、typed judge retry boundary、attempt audit matrix；run25849 aborted10/11，not PASS. |
| 2026-07-13 | 2.8 | Evalkit reuses product full semantic validator；sole-label targeted repair，all other/mixed whole-report repair，one budget/full revalidation/second-invalid zero judge；focused live fixes recorded，final matrix pending. |
| 2026-07-13 | 2.7 | Finalize scheme A：wire/schema fuse200 code points；semantic/UX24 whitespace words/64 Unicode code points；targeted repair margin18/52；reopen P0.099/P0.100 evidence. |
| 2026-07-13 | 2.6 | Normalize all action-label schema120/14-40 violations to action_labels，including label>120 schema-invalid；record current P0.100 live FAIL. |
| 2026-07-13 | 2.5 | Split evalkit schema+14/40 scoped repair from runner cross-field/focus/action no-repair gate；judge remains one-shot after both pass. |
| 2026-07-13 | 2.4 | Separate P0.100 5-case/11-attempt reliability from P0.099 current-run canonical screenshot audit；keep exact 14/40 boundary proof deterministic. |
| 2026-07-13 | 2.3 | Lock P0.100 evalkit same-source output-schema validation, one bounded product repair, aggregate generation usage/latency + repair_used and one-shot judge；bound action length and multi-focus semicolon fragments. |
| 2026-07-12 | 2.2 | Split active scenario ownership: plan 001 owns P0.099 exact six-image acceptance; plan 002 exclusively owns all new P0.100 content-reliability phases and BDD gates. |
| 2026-07-12 | 2.1 | Reopen P0.099/100 for frozen direct reports, exact six-image acceptance and content-level reliability audit. |
| 2026-07-12 | 2.0 | Rebase E2E owners onto continuous conversation and shared real-browser acceptance. |
