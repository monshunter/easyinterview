# 报告仪表盘目标结构

> **版本**: 1.32
> **状态**: active
> **更新日期**: 2026-07-14

## 1 目标

报告列表按当前面试规划组织，报告详情以整场 conversation 为分析单位。列表只展示当前 TargetJob canonical rounds 的当前可用报告与最新生成状态；详情帮助用户理解准备度、能力表现、证据与下一步。页面只展示后端持久化事实，不按题目组织、不展示隐藏数值分，也不由前端推断报告含义。

## 2 页面结构

```text
ReportsScreen(targetJobId)
├─ Back -> Parse current plan detail
├─ Header / current TargetJob
├─ loading / empty / error
└─ Canonical round rows
   ├─ currentReport -> ReportDashboard(reportId)
   └─ latestAttempt -> ReportGenerating(reportId) / typed status

ReportDashboard(reportId)
├─ Back -> ReportsScreen / Workspace fallback
├─ Header
│  ├─ breadcrumb / title / subtitle
│  ├─ 复练当前轮
│  └─ 进入下一轮
├─ ContextStrip
│  ├─ target
│  ├─ round
│  └─ resume
├─ SummaryMetrics
│  ├─ 准备度 + summary
│  ├─ 能力维度数量
│  └─ 会话证据数量
└─ DetailGrid
   ├─ 能力维度
   ├─ 优势证据
   ├─ 风险 / 待加强证据
   └─ 下一步行动
```

ReportsScreen 是规划范围的导航/索引页，不是第二种报告内容形态。ReportDashboard 的 UI 真理源仍是“三项指标 + 四个常驻区块，无 tab”，不得根据旧文档恢复四卡或四 tab。

### 2.1 当前规划报告列表

- 入口位于 Parse 内容区标题行右上角，进入 `/reports?targetJobId=<uuid>`；不加入全局 TopBar。
- 同时读取 `getTargetJob(targetJobId)` 与 `listTargetJobReports(targetJobId)`。前者提供当前规划与 canonical round display，后者只提供 `currentReport/latestAttempt`；target、round ID/sequence/count/order 任一不一致即整页 fail closed。
- 每轮只展示当前可用报告和最新生成状态：queued/generating 可进入 Generating，failed 为本地化状态且无同 report Retry，latest ready 与 current 同 ID 时去重；ID 不同时只说明最近一次生成已完成，不展开完整历史版本。
- loading、empty、network/contract error、跨用户/target mismatch 和 stale response 均有明确状态；target 切换首帧清空旧 rows，其他规划 sentinel 不得出现。
- Reports Back 返回 `parse?targetJobId=<当前可信 id>`；缺失/非法 target 提供安全返回 workspace，不从其他 query 推断规划。

## 3 诚实生成态

Generating 只表达后端真实的 queued / generating / failed / timeout / ready 状态：

- 可以说明系统正在整理上下文、核对证据、形成建议，但不得显示与后端无关的百分比或自动完成阶段。
- 不得展示固定示例作为“实时观察”。
- 不得使用“好了通知我”或虚构后台通知。当前 records 合同位于独立 ReportsScreen；Generating 只能在 trusted response 提供 `targetJobId` 时通过 Back 返回该页面，不能自行维护列表。
- queued/generating 自动继续轮询；timeout/network 可“继续检查”；ready 后进入 report；failed/not found/invalid contract/`REPORT_CONTEXT_TOO_LARGE` 是终态，只能返回，不把重新 GET 伪装成重新生成。超限态说明本次材料与对话过长，并引导返回规划、缩短输入后开启新会话。
- Report单次`GenerateReport`动作在后端调用内执行initial+最多3次retry与10s/20s/40s等待；动作返回销毁retry context，新动作清零。Runner的`async_jobs.attempts/max_attempts`与outbox/infra的30s/2m/10m/1h/6h都不是产品retry事实。Frontend使用`maxAttempts=49`、初始1.5s、×1.5、cap8s，总约6m04s，覆盖4×60s+10+20+40=5m10s并留约54s。整个queued/generating期间不显示attempt/retry/progress；轮询窗口耗尽只提供“继续检查”，不能显示为服务端failed。当前OpenAPI没有failed-report regenerate operation，不设计或宣称同report Retry入口。

## 4 Summary Metrics

| Metric | 内容 |
|--------|------|
| 准备度 | localized readiness tier + LLM `summary` |
| 能力维度 | `dimensionAssessments.length` |
| 会话证据 | `highlights.length + issues.length` |

不得在前端补默认数量、默认 summary 或假报告。

## 5 Detail Grid

### 5.1 Dimensions

- 使用 `label` 作为用户可见维度名称，`code` 只作为报告内关联标识。
- status / confidence 必须映射为当前 UI 语言文案，不能显示 `strong · high` 等 raw enum。模型生成的 dimension label 按 report language 原样显示。

### 5.2 Evidence

- highlights / issues 显示 report-language dimension label/evidence 与 localized confidence。
- 前端不展示内部 message anchors，不复制完整 transcript，不按题号或 turn 分组。
- 缺失/未知 dimensionCode 视为合同错误，不由前端猜测。

### 5.3 Next Actions

- 按服务端顺序展示，第一项是推荐优先级。
- action type 不直接作为用户文案；action label 按 report language 原样显示；未知类型 fail closed。
- `ReportNextAction.label.maxLength=200` code points 只是在 OpenAPI / JSON Schema 层拒绝 malformed model output 的技术保险丝，不是正常文案长度、设计目标或 UI 验收目标。真实用户体验只接受 English `<=24 whitespace words`、`zh-CN <=64 Unicode code points`；正式 UI 不展示以 200 为目标的文案。
- 产品完整validator若发现全部违规仅为`nextActions[i].label` schema maxLength200和/或语言24/64，即使schema-invalid也走`action_labels`；targeted generation使用内部生成余量18/52，只merge labels。其它任意schema、semantic或mixed violation走整报告generation。Initial及后续每轮输出均完整复验，最多4次调用；attempt4仍invalid才在judge前fail closed。Frontend 对超24/64 ready payload进入typed invalid，不回显raw label，不截断/代写/ellipsis伪装合法。

## 6 CTA

- first action 为 `retry_current_round`：复练使用 accent，下一轮为 secondary。
- first action 为 `next_round`：下一轮使用 accent，复练为 secondary。
- first action 为 `review_evidence`：两者均 secondary。
- 视觉主次只表达建议，不改变用户可选择的合法路径。
- 下一轮未知/末轮/加载失败/重复派生 ID 时 disabled 并提供可访问原因；任一 start 进行中时两枚 CTA 都 disabled。
- Replay/Next derived 请求只携带 goal + sourceReportId；后端从 source report/plan 投影全部 settings、round 与 focus。Replay 始终允许：有 issue-backed needs-work dimension 时使用服务端 focus，没有可支持 focus 时创建空 focus 的通用同轮复练。`context.hasNextRound=false` 时 Next disabled。

### 6.1 事实源与语言边界

- `reportId` 是唯一 locator；status/error、target/resume/round label 与 CTA identity 全部来自 `getFeedbackReport.context`。route 中冲突参数必须忽略。
- UI chrome、enum、固定 CTA/错误文案随 UI locale；LLM summary/dimension/evidence/action label 使用 report language，前端不得翻译或改写。

### 6.2 Back 与报告记录恢复

- Ready Report、pending/queued/generating、API failed terminal state、timeout/network continue-check state若已有 trusted `targetJobId`，Back 导航到 `/reports?targetJobId=<id>`。
- missing reportId、首读 404/网络失败、invalid payload 且没有可信 TargetJob identity 时，Back 导航到 `workspace`。
- Report / Generating route 始终只携带 `reportId`；不得把 targetJobId 写入当前 route、从 URL/标题反推 identity，或调用 `listTargetJobReports`。该 operation 的唯一 UI consumer 是 ReportsScreen。

## 7 可读性与响应式

- frozen target / round / resume 允许换行或通过 title/accessible description 读取完整值。
- session/report UUID 等内部 locator 不渲染为用户字段，也不进入 title、tooltip 或 accessible description；它们只保留在 API/动作内部关联中。
- Desktop 使用双列 DetailGrid；390px mobile 明确单列。
- 长 dimension/evidence/action 必须换行，不横向溢出、不被不可恢复截断。
- 1440x1200 desktop 与 390x844 mobile full-page 都必须覆盖 action 区域，并证明合法 24/64 label 完整换行、无截断/ellipsis/隐藏/横溢。恰好 24/64 由 deterministic fixture parity 证明；200-code-point malformed fixture只用于 typed invalid/no-raw-output 测试，不能充当 UX PASS。18/52 只用于 targeted repair 内部生成，不替代边界 fixture。
- 能力维度行在宽度足够时保持 `label` 与本地化 status/confidence 左右对齐；空间不足时整项换为两段可读行。英文长 label 优先按单词换行，禁止为了保留右侧状态而压缩成逐字符竖排。
- Report 保留 App Shell TopBar：desktop 内容从 58px TopBar 后开始；390px mobile 内容从响应式 TopBar 的实际底部开始。TopBar 可因 UI locale 与已登录用户名称产生合法换行，但 document `scrollWidth` 不得超过 viewport，报告局部布局也不得用相对坐标掩盖共享 TopBar 的绝对纵向偏差。

## 8 状态

- Missing session/report：专用空态。
- Queued/generating：诚实等待态。
- Timeout/network：typed recoverable error，可继续检查；Failed/not found/invalid contract/`REPORT_CONTEXT_TOO_LARGE`：typed terminal error，只能返回；超限态不得出现同 report 的 Retry。
- Ready：summary、dimensions、evidence、actions 完整。
- Empty required semantic fields：合同失败，不回退假内容。

## 9 负向边界

当前 UI、fixtures、tests、scenarios 和 active 文档不得正向保留：

- 四卡 / 四 tab 或 Questions tab。
- questionAssessments / retryFocusTurnIds / per-question replay。
- candidate numeric score、录用概率、timeline。
- raw enum/code 用户文案。
- fake progress / live observation / fake notify。
- focusCompetencyCodes / evidenceGaps URL 或客户端事实源。
- route status/error/target/resume/round 覆盖 API frozen facts。
- client translation/rewrite of model summary/dimension/evidence/action labels。
- user-visible or accessibility-exposed session/report UUID/internal locator。
- global/cross-target Report Center、完整历史版本列表、Parse/Report/Generating reports-list consumer 或 route-provided targetJobId authority。

## 10 验收标准

| ID | Given | When | Then |
|----|-------|------|------|
| R-1 | queued/generating | 打开生成页 | 无假进度、假观察、假通知 |
| R-2 | ready direct report | 打开报告 | 三指标四常驻区块，summary 与 localized semantic 完整 |
| R-3 | retry/next/review first action | 查看 Header | 现有 CTA 主次与建议一致 |
| R-4 | needs-work / well-prepared report | 点击复练 | source report 服务端投影 issue-backed focus，或在无可支持 focus 时创建空 focus 的通用同轮复练；客户端不携带 focus |
| R-5 | 长内容 desktop/mobile | 打开报告 | 完整可读、mobile 单列、无横向溢出 |
| R-6 | prototype/formal deterministic boundary fixture | 运行 desktop+390 parity | 恰好 24-whitespace-word / 64-Unicode-code-point label 在 1440x1200 与 390x844 均完整换行；超 24/64 fixture 进入 typed invalid 且不回显 raw |
| R-7 | real provider zh/en | P0.099 当前 run 的 en/zh ready rows | 六图 manifest 对每个 row 绑定 DB/API `canonical_report_content_digest`、`action_length_audit`、`content_audit`、`screenshot_sha256` 与 report/session/context digest；两张 390x844 report full-page 截图完整覆盖 action 区域，实际 label 分别满足 `<=24 whitespace words` / `<=64 Unicode code points` 且完整可见、无截断/省略/横溢 |
| R-8 | reportId-only / conflicting route | 深链刷新/点击 CTA | API frozen status/context 获胜 |
| R-9 | UI locale != report language | 打开报告 | chrome 本地化，模型语义保持报告原文 |
| R-10 | ready report has internal IDs | 打开 desktop/mobile 报告 | Context Strip 只显示 target/round/resume；可见 DOM、可访问名称与截图都不暴露 session/report UUID |
| R-11 | trusted target context / no trusted identity | 从 ready/pending/failed/recoverable generating 点击 Back，或在 missing/first-load failure 点击 Back | 有 trusted target 时进入 `/reports?targetJobId=...`；否则进入 workspace；report/generating route 仍只含 reportId |
| R-12 | 当前 TargetJob overview populated/empty/loading/error | 直开或刷新 `/reports?targetJobId=...` | 只展示当前规划 canonical rounds 的 current/latest，不展示其他规划或完整历史；mismatch/stale fail closed，desktop/mobile prototype/formal parity 通过 |

## 11 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-14 | 1.32 | 增加独立 target-scoped ReportsScreen，锁定 current/latest-only、规划隔离、四态与 desktop/mobile parity；Report/Generating trusted Back 改回该列表。 |
| 2026-07-14 | 1.31 | 将 Report/Generating Back 收敛为 trusted target -> Parse reports anchor、无可信 identity -> Workspace fallback，并禁止顶层报告中心与 route target authority。 |
| 2026-07-13 | 1.30 | Context Strip 删除 session/report UUID 等内部 locator，只保留 target/round/resume，并要求 desktop/mobile 可见与可访问负向验收。 |
| 2026-07-13 | 1.29 | Correct report timing ownership to action-local initial+3 with10s/20s/40s; async attempts are infrastructure-only. Keep maxAttempts49/6m04s and no unsupported failed-report regenerate UI. |
| 2026-07-13 | 1.28 | Lock report use of business10s/20s/40s under durable max4 and frontend maxAttempts49 (~6m04s)；separate business cap80 from infra delivery and expose no internal attempt/progress. |
| 2026-07-13 | 1.27 | Clarify product full-validator repair scope：sole-label targeted，all other/mixed whole-report，one-budget full revalidation；visible UI boundary unchanged. |
| 2026-07-13 | 1.26 | 方案 A 最终边界：200 code-point wire fuse；24-word/64-code-point semantic/UX；targeted repair 内部余量18/52；desktop+390完整换行，超限typed invalid/no raw。 |
| 2026-07-13 | 1.25 | A-200：wire/schema fuse改为200；14/40仍为UX gate，desktop+390合法边界完整换行，超限typed invalid且不回显raw。 |
| 2026-07-13 | 1.24 | 归一化 action-label schema120/语言14-40 violation set；即使 label>120 导致 schema-invalid 仍使用 action_labels，修复同时满足两层上限。 |
| 2026-07-13 | 1.23 | Runtime 使用一次总预算下的整报告 / 唯一 action-length label-only LLM repair，labels-only 原样 merge并全量复验；evalkit 分界由 F3/P0.100 owner 承接。 |
| 2026-07-13 | 1.22 | 区分 120-char wire/schema fuse、P0.099 current-run canonical audit chain 与确定性 14/40 boundary fixture pixel parity；P0.100 内容可靠性不与六图强绑 output digest。 |
| 2026-07-12 | 1.21 | 修复 prototype/formal 同时错误导致的 mobile 英文能力维度逐字符竖排，定义 label 与 status 可读换行契约。 |
| 2026-07-12 | 1.20 | 固化 Report mobile TopBar 响应式换行、内容起点和无横向溢出的绝对 viewport parity 契约。 |
| 2026-07-12 | 1.19 | 明确 Replay 在无可支持 focus 时创建空 focus 的通用同轮复练，并补 `REPORT_CONTEXT_TOO_LARGE` 的诚实终态与可执行返回指引。 |
| 2026-07-12 | 1.18 | 补 frozen context/reportId-only 事实源、终态动作矩阵、records 负向边界与 UI/report 双语言契约。 |
| 2026-07-12 | 1.17 | 统一三指标四常驻区块；接入 direct semantic summary/code+label，删除 generating 伪实时语义，补齐 enum i18n、CTA 推荐、server-owned focus、mobile 可读性与强截图 gate。 |
| 2026-07-12 | 1.16 | 下一轮只使用 TargetJob 有序结构化轮次的紧邻后一项；异常状态 fail closed。 |
