# 报告仪表盘目标结构

> **版本**: 1.47
> **状态**: active
> **更新日期**: 2026-07-19

## 1 目标

报告列表按当前面试规划组织，报告详情以整场 conversation 为分析单位。列表只展示当前 TargetJob canonical rounds 的当前可用报告与最新生成状态；详情帮助用户理解准备度、能力表现、证据与下一步。页面只展示后端持久化事实，不按题目组织、不展示隐藏数值分，也不由前端推断报告含义。

## 2 页面结构

```text
ReportsScreen(targetJobId)
├─ Back -> Workspace current plan detail
├─ Header / current TargetJob
├─ loading / empty / error
└─ Canonical round rows
   ├─ currentReport -> ReportDashboard(reportId)
   │  └─ 查看面试记录 -> ReportConversation(reportId)
   └─ latestAttempt
      ├─ queued/generating -> ReportGenerating(reportId) + 查看面试记录
      ├─ failed -> 重新生成同一 reportId + 查看面试记录
      └─ ready -> same-ID 与 current 去重；不同 ID 仅显示独立面试记录入口

ReportDashboard(reportId)
├─ Back -> ReportsScreen / Workspace fallback
├─ Header
│  ├─ breadcrumb / title / subtitle
│  ├─ 复练当前轮
│  └─ 进入下一轮
├─ ContextStrip
│  ├─ target
│  ├─ round
│  ├─ resume -> /resume-versions?resumeId=<frozen resumeId>
│  └─ interview record action -> report-conversation route
├─ SummaryMetrics
│  ├─ 能力维度数量
│  └─ 会话证据数量
├─ DetailGrid
│  ├─ 能力维度
│  ├─ 优势证据
│  ├─ 风险 / 待加强证据
│  └─ 下一步行动
└─ OverallSummary
   ├─ 面试总评
   ├─ localized readiness tier
   └─ LLM summary

ReportConversation(reportId)
├─ Back -> ReportDashboard / ReportGenerating
├─ frozen target / round / resume
└─ ordered readonly Markdown transcript
```

Desktop `2048×917` 目标视图在全局 TopBar 下使用浅蓝全视口背景和约 `1432px` 居中内容面。Back、Header、Context Strip、Summary、Detail 与 Overall Summary 共享同一左右边界；Header 固定以“复练当前轮”为蓝色实心主按钮并带刷新 icon，“进入下一轮”为白色描边次按钮并带右箭头，不再根据 `nextActions[0]` 交换视觉层级；可用性与点击行为仍由服务端状态控制。Context Strip 必须是一张共享外框的横向整卡，四个上下文子项由内部竖线分隔，不得渲染成四张相互独立的小卡。两张数量指标卡与两行内容卡使用 10-12px 圆角、轻边框和克制阴影；四张内容卡都带与含义一致的圆形语义 icon，正文区使用紧凑的标题、行与列表层级。优势/风险卡不重复渲染已在能力维度行出现的置信度文案。典型两维度、两条证据的合法 ready 报告应在 `2048×917` 首屏内完整露出底部总评卡；更长合法内容仍自然增高并完整换行，不通过截断、隐藏或省略号换取首屏。Mobile 保持同一阅读顺序并单列收敛。

ReportsScreen 是规划范围的导航/索引页，不是第二种报告内容形态。ReportDashboard 的当前设计合同是 desktop 自上而下 `4/2/2/2/1`：同一个 Context Strip block 内包含目标岗位、轮次、简历、面试记录四个同级子项；其后是两项数量指标、两行各两个内容区，最后一个全宽“面试总评”大卡片；无 tab。Mobile 保持相同 DOM 与阅读顺序，每组收敛为单列。不得根据旧文档恢复顶部“准备度 + summary”指标、四卡或四 tab。

ReportsScreen 与 ReportConversation 也必须采用同一套目标稿页面语言，而不是保留窄的工具页布局：desktop 内容面约 `1372px`，Back、蓝色 eyebrow、公司与岗位主标题、说明文案共享左边界，右侧使用低对比度、不可交互且 `aria-hidden` 的线性几何插画平衡标题区。ReportsScreen 在 Header 下先显示一张目标岗位摘要卡，再把 canonical rounds 渲染为左侧两位编号和贯穿竖线、右侧独立圆角轮次卡的时间线；每个轮次卡展示真实 round name/duration、当前 report 日期或真实状态，主报告动作使用蓝色实心按钮，面试记录使用白色描边按钮。ReportConversation 在 Header 下显示一张三列分隔的 target/round/resume Context Strip，随后按原始顺序显示消息：assistant 与 user 都使用相同的整行浅色卡片、描边、圆角、内边距和约 `60px` 方形头像轮廓；只通过蓝色 `AI` / 灰色“我”的头像色彩和角色文案区分身份。两类正文均完整换行，无 composer。

### 2.1 当前规划报告列表

- 入口位于 `/workspace?targetJobId=...` 只读详情标题下方首行动作行，与“立即面试”从左同排，进入 `/reports?targetJobId=<uuid>`；不加入全局 TopBar 或页尾。
- 同时读取 `getTargetJob(targetJobId)` 与 `listTargetJobReports(targetJobId)`。前者提供当前规划与 canonical round display，后者只提供 `currentReport/latestAttempt`；target、round ID/sequence/count/order 任一不一致即整页 fail closed。
- 每轮只展示当前可用报告和最新生成状态：queued/generating 同时提供“查看生成进度”和“查看面试记录”；failed latest attempt 保留同一 `reportId`，除 `REPORT_CONTEXT_TOO_LARGE` 外提供“重新生成报告”和“查看面试记录”。面试一旦结束，记录入口不受报告 queued/generating/ready/failed 状态影响；重新生成只把同一报告重新排队，不创建第二份报告或历史入口；latest ready 与 current 同 ID 时去重，ID 不同时只说明最近一次生成已完成，不展开完整历史版本。
- 若同一轮同时存在旧 `currentReport` 与更新的 failed `latestAttempt`，旧报告的“查看报告/查看面试记录”仍绑定 current report；失败恢复动作绑定 latest attempt，并使用可区分的 accessible name。`REPORT_CONTEXT_TOO_LARGE` 只保留查看面试记录与返回动作，因为冻结输入不会因重复调用变小。
- loading、empty、network/contract error、跨用户/target mismatch 和 stale response 均有明确状态；target 切换首帧清空旧 rows，其他规划 sentinel 不得出现。
- Reports Back 返回 `/workspace?targetJobId=<当前可信 id>` 只读详情；缺失/非法 target 提供安全返回 query-free `/workspace`，不从其他 query 推断规划，也不进入 Parse 命令进度。
- Header 下的目标岗位摘要卡只消费现有 TargetJob 事实：公司、岗位、可选 location、canonical round 数量和 `createdAt`；缺失字段留空或降级为已有本地化文案，不伪造部门、面试日期或业务状态。轮次时间线的时长只来自 canonical round `durationMinutes`，报告日期只来自 `currentReport.generatedAt` / latest attempt 真实时间。

## 3 诚实生成态

Generating 与其它长耗时流程共用 `AsyncTransitionScene`：页面保留共享 TopBar，并在约 1090px 居中的轻量白色主面板中展示报告插画、真实 `queued/generating` 状态标签、衬线标题、说明、indeterminate rule 和返回 Reports/Workspace 的既有动作。视觉上的局部高亮只表达“仍在处理”，不携带 `aria-valuenow`，不映射 provider attempt、重试次数、固定百分比或 SLA。ready 自动进入 Report；timeout/network/failed/invalid 的 typed error 仍使用既有恢复状态，不被动画覆盖。

Generating 只表达后端真实的 queued / generating / failed / timeout / ready 状态：

- 可以说明系统正在整理上下文、核对证据、形成建议，但不得显示与后端无关的百分比或自动完成阶段。
- 不得展示固定示例作为“实时观察”。
- 不得使用“好了通知我”或虚构后台通知。当前 records 合同位于独立 ReportsScreen；Generating 只能在 trusted response 提供 `targetJobId` 时通过 Back 返回该页面，不能自行维护列表。
- queued/generating 自动继续轮询；timeout/network 可“继续检查”；ready 后进入 report；failed/not found/invalid contract/`REPORT_CONTEXT_TOO_LARGE` 对当前 Generating 页面是终态，返回 Reports 后才可执行受控的用户恢复动作，不能把重新 GET 伪装成重新生成。超限态说明本次材料与对话过长，并引导返回规划、缩短输入后开启新会话。
- Report 单次 `GenerateReport` 动作在后端调用内执行 initial + 最多 3 次 retry 与 10s/20s/40s 等待；动作返回销毁 retry context，新一次用户“重新生成报告”动作重新获得独立预算。Runner 的 `async_jobs.attempts/max_attempts` 与 outbox/infra 的 30s/2m/10m/1h/6h 都不是产品 retry 事实。Frontend 使用 `maxAttempts=49`、初始 1.5s、×1.5、cap 8s，总约 6m04s，覆盖 4×60s+10+20+40=5m10s 并留约 54s。整个 queued/generating 期间不显示 attempt/retry/progress；轮询窗口耗尽只提供“继续检查”，不能显示为服务端 failed。

## 4 Summary Metrics 与面试总评

| Metric | 内容 |
|--------|------|
| 能力维度 | `dimensionAssessments.length` |
| 会话证据 | `highlights.length + issues.length` |

Summary Metrics 只承载两个可扫描数量，不展示 readiness 或 `summary`。

页面最底部的全宽 `OverallSummary` 是唯一“面试总评”容器：先展示 localized readiness tier，再展示服务端 LLM `summary`。该卡片在 desktop 跨越整个内容网格，在 mobile 位于所有证据与行动之后；`summary` 全页只出现一次，前端不得改写、提炼、复制或移动回顶部指标区。不得在前端补默认数量、默认 summary 或假报告。

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

### 6.3 报告附属的只读面试记录

- Report Context Strip 把“面试记录”作为目标岗位、轮次、简历之后的第四个同级子项，值为“查看本次面试记录”；不再在 strip 下方保留游离入口。ReportsScreen 的 current report 与任意不同 locator 的 queued/generating/ready/failed latest attempt 都必须提供“查看面试记录”快捷入口，并进入各自 `/report-conversation?reportId=...`；生成进度或失败恢复动作不得替代记录入口，同 ID current/latest 只显示一次。
- 简历子项使用冻结 `resumeId` 构造 canonical `/resume-versions?resumeId=...` URL，显示冻结 `resumeDisplayName`，支持点击、复制链接与新标签页打开本次面试使用的简历副本；Report 不额外调用 `getResume`，不把可变简历正文复制进报告响应。
- 页面只消费 `getReportConversation`，按 `sequence` 显示 user/assistant 安全 Markdown/GFM；不展示 Composer、thinking、retry、pause、计时、电话或 session/message/client IDs。
- ready 返回 ReportDashboard，queued/generating 返回 ReportGenerating；failed 以 frozen context 的可信 `targetJobId` 直接返回 ReportsScreen，便于继续在同一失败行手动处理，缺失可信 target 时回 Workspace。missing/跨用户/乱序/非法 role/stale response 整体 fail closed，不显示 partial transcript。
- 不设立 session history 列表、`sessionId` 用户路由或新关系表；已删除的并行 Demo runtime 不作为实现/验收来源。
- Desktop 记录页使用约 `1372px` 内容面、Header 右侧 decorative illustration、一张三列内部竖线分隔的 Context Strip，以及明确区分 assistant/user 的消息行。Assistant 与 user 行共享整行浅色卡片、描边、圆角和同宽约 `60px` 方形头像轮廓；角色只由蓝色 `AI` / 灰色“我”的色彩与文案区分，不允许任一角色直接落在主体背景上。Mobile 保持 target/round/resume/message 的同一 DOM 顺序，Context Strip 与消息行收敛为单列且不横溢。

## 7 可读性与响应式

- frozen target / round / resume 允许换行或通过 title/accessible description 读取完整值；简历以 link、面试记录以 button action semantics 暴露，均可键盘访问且具有明确 accessible name；只有简历 URL 可复制/新标签页打开，reportId 不写入 DOM 属性。
- session/report UUID 等内部 locator 不渲染为用户字段，也不进入 title、tooltip 或 accessible description；它们只保留在 API/动作内部关联中。
- Desktop 的 ready 内容严格按 `4/2/2/2/1` 排列：Context Strip 四列、Summary Metrics 两列、Dimensions/Strength Evidence 两列、Risks/Next Actions 两列、Overall Summary 全宽一列。390px mobile 保持相同 DOM 顺序并全部收敛为单列。
- Desktop 同一双列行的两个 panel 外边界必须上下对齐并等高；内容较少的一侧在卡片内部自然留白，不得让内层边框提前结束形成“一高一低”。该规则同时覆盖 Dimensions/Strength Evidence 与 Risks/Next Actions 两行。
- 长 dimension/evidence/action 必须换行，不横向溢出、不被不可恢复截断。
- 1440x1200 desktop 与 390x844 mobile full-page 都必须覆盖 action 区域，并证明合法 24/64 label 完整换行、无截断/ellipsis/隐藏/横溢。恰好 24/64 由 deterministic fixture-backed responsive test 证明；200-code-point malformed fixture只用于 typed invalid/no-raw-output 测试，不能充当 UX PASS。18/52 只用于 targeted repair 内部生成，不替代边界 fixture。
- 能力维度行在宽度足够时保持 `label` 与本地化 status/confidence 左右对齐；空间不足时整项换为两段可读行。英文长 label 优先按单词换行，禁止为了保留右侧状态而压缩成逐字符竖排。
- Report 保留 App Shell TopBar：desktop 内容从 58px TopBar 后开始；390px mobile 内容从响应式 TopBar 的实际底部开始。TopBar 可因 UI locale 与已登录设置按钮产生合法换行，但 document `scrollWidth` 不得超过 viewport，报告局部布局也不得用相对坐标掩盖共享 TopBar 的绝对纵向偏差。

## 8 状态

- Missing session/report：专用空态。
- Queued/generating：诚实等待态。
- Timeout/network：typed recoverable error，可继续检查；Failed/not found/invalid contract/`REPORT_CONTEXT_TOO_LARGE`：当前 Generating 页 typed terminal error，只能返回。ReportsScreen 对普通 failed latest attempt 提供同 report regenerate；超限态不得出现 regenerate。
- Ready：dimensions/evidence 数量指标、四个证据/行动区块以及底部唯一面试总评完整。
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
- readiness/summary 继续位于顶部 Summary Metrics、`summary` 重复渲染，或 Overall Summary 未位于四个内容区之后。
- user-visible or accessibility-exposed session/report UUID/internal locator。
- global/cross-target Report Center、完整历史版本列表、Parse/Report/Generating reports-list consumer 或 route-provided targetJobId authority。
- `listPracticeSessions` / `getPracticeSession` 记录入口、会话历史列表、`sessionId` 用户路由或只对 ready report 显示记录。

## 10 验收标准

| ID | Given | When | Then |
|----|-------|------|------|
| R-1 | queued/generating | 打开生成页 | 无假进度、假观察、假通知 |
| R-2 | ready direct report | 打开报告 | desktop 按 `4/2/2/2/1` 展示；四项上下文同属一个 block，顶部只有两个数量指标，四个常驻区块之后是全宽面试总评，localized readiness 与唯一 `summary` 完整 |
| R-3 | retry/next/review first action | 查看 Header | 现有 CTA 主次与建议一致 |
| R-4 | needs-work / well-prepared report | 点击复练 | source report 服务端投影 issue-backed focus，或在无可支持 focus 时创建空 focus 的通用同轮复练；客户端不携带 focus |
| R-5 | 长内容 desktop/mobile | 打开报告 | 完整可读、mobile 单列、无横向溢出 |
| R-6 | deterministic boundary fixture | 运行正式前端 desktop+390 responsive test | 恰好 24-whitespace-word / 64-Unicode-code-point label 在 1440x1200 与 390x844 均完整换行；超 24/64 fixture 进入 typed invalid 且不回显 raw |
| R-7 | real provider zh/en | P0.099 当前 run 的 en/zh ready rows | 六图 manifest 对每个 row 绑定 DB/API `canonical_report_content_digest`、`action_length_audit`、`content_audit`、`screenshot_sha256` 与 report/session/context digest；两张 390x844 report full-page 截图完整覆盖 action 区域，实际 label 分别满足 `<=24 whitespace words` / `<=64 Unicode code points` 且完整可见、无截断/省略/横溢 |
| R-8 | reportId-only / conflicting route | 深链刷新/点击 CTA | API frozen status/context 获胜 |
| R-9 | UI locale != report language | 打开报告 | chrome 本地化，模型语义保持报告原文 |
| R-10 | ready report has internal IDs | 打开 desktop/mobile 报告 | Context Strip 显示 target/round/resume/interview record 四项；resume 使用 frozen resumeId URL，conversation 使用无 reportId DOM 属性的 action；可见 DOM、可访问名称与截图都不暴露 session/report UUID |
| R-11 | trusted target context / no trusted identity | 从 ready/pending/failed/recoverable generating 点击 Back，或在 missing/first-load failure 点击 Back | 有 trusted target 时进入 `/reports?targetJobId=...`；否则进入 workspace；report/generating route 仍只含 reportId |
| R-12 | 当前 TargetJob overview populated/empty/loading/error | 直开或刷新 `/reports?targetJobId=...` | 只展示当前规划 canonical rounds 的 current/latest，不展示其他规划或完整历史；mismatch/stale fail closed，desktop/mobile responsive/state tests 通过 |
| R-13 | owned report 为 queued/generating/ready/failed | 从 Report Context Strip 第四项或 ReportsScreen 打开面试记录 | 同一 reportId-only 页显示严格有序的只读 Markdown transcript，返回正确父页，无会话列表/live controls/internal IDs |
| R-14 | ready report 有冻结 resumeId | 点击或复制简历子项 URL | canonical `/resume-versions?resumeId=...` 打开本次面试使用的简历副本；Report 不额外读取简历正文 |
| R-15 | desktop 双列内容长度不同 | 打开报告 | 同一行两个 panel 上下边界对齐且等高，短内容侧仅在卡片内部留白 |
| R-16 | latest attempt 为普通 failed / `REPORT_CONTEXT_TOO_LARGE` | 打开 ReportsScreen | 普通 failed 显示同 report 的重新生成与面试记录；超限只显示面试记录；旧 ready 与新 failed 并存时动作 locator 和 accessible name 均可区分 |
| R-17 | ready report at 2048×917 / 390×844 | 打开报告并检查布局 | desktop 主体约 1432px 且与 Header/Context/Detail/Overall 共用网格；Context 为一张四列分隔整卡，四个内容卡带语义 icon，典型合法内容的总评在首屏完整出现；按钮、圆角、字体层级与目标图一致，mobile 无横向溢出 |
| R-18 | 当前规划有 canonical rounds 与 current/latest 报告 | 打开 ReportsScreen | desktop 约 1372px，Header 带 decorative illustration，现有 TargetJob 事实组成目标摘要卡；轮次以编号时间线 + 独立卡展示，报告/记录/生成/恢复动作主次明确且 locator 不变，mobile 同序无横溢 |
| R-19 | report-owned transcript 有 assistant/user 消息 | 打开 ReportConversation | desktop 约 1372px，Header 与三列 Context Strip 对齐；assistant/user 共享浅色整行卡片、描边、圆角、内边距和同宽方形头像轮廓，只以蓝色 AI / 灰色“我”区分身份；Markdown 完整换行且无 composer/internal IDs，mobile 同序无横溢 |

## 11 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-19 | 1.46 | 按用户验收反馈统一面试官与“我”的消息卡外框、圆角、内边距和头像轮廓；角色只以色彩与文案区分。 |
| 2026-07-19 | 1.47 | 按生成报告参考稿重建诚实等待场景：恢复共享 TopBar，使用中心报告插画、白色主面板、indeterminate rule 与既有返回动作，继续禁止假百分比/阶段/SLA。 |
| 2026-07-19 | 1.45 | 按提供的报告列表与面试记录参考稿锁定共享 Header 插画、目标摘要卡、编号时间线、三列上下文与 assistant/user 消息行；只消费现有持久化事实，不扩展 API。 |
| 2026-07-19 | 1.44 | 用户确认上一轮只做了外层宽度/圆角包装，未完成目标稿结构改造；原地重开报告页，锁定 1432px 整页重构、单体 Context Strip、内容卡语义 icon、紧凑首屏构图与长内容非截断边界。 |
| 2026-07-16 | 1.42 | 补齐 latest-ready 列表树：同 ID 去重，不同 ID 保留独立面试记录入口且不扩展为历史报告列表。 |
| 2026-07-19 | 1.43 | 按参考图锁定报告页 1336px 主体、浅蓝背景、共享网格、圆角卡片、语义 icon 与 Header CTA 层级。 |
| 2026-07-16 | 1.41 | 固化已结束面试的记录入口不依赖报告状态：queued/generating 与生成进度并列展示“查看面试记录”，failed/ready 延续同一规则。 |
| 2026-07-16 | 1.40 | Failed conversation Back 直接回可信 target-scoped ReportsScreen，并要求多标签状态冲突刷新当前 overview，避免停留在 stale failure。 |
| 2026-07-16 | 1.39 | 失败报告增加同 `reportId` 的手动重新生成与只读面试记录恢复；`REPORT_CONTEXT_TOO_LARGE` 保持不可重复生成。 |
| 2026-07-15 | 1.38 | 将 ready 报告修订为 `4/2/2/2/1`：面试记录并入 Context Strip 第四项，简历增加 frozen-resume canonical URL；desktop 两个 detail 行锁定同排等高与内部留白。 |
| 2026-07-15 | 1.37 | 对齐方案 A 的 TopBar 响应式合同：已登录账号区为设置按钮，不再按用户名 chip 计算换行。 |
| 2026-07-15 | 1.36 | 合并 report-owned 只读面试记录：Report 主入口 + ReportsScreen 快捷入口、reportId-only route、安全 Markdown、四状态父页 Back；保留 `3/2/2/2/1` 和正式 frontend 单实现原则。 |
| 2026-07-15 | 1.35 | 将 Workspace 详情的报告入口从标题右上角移到标题下方左对齐首行动作行，与“立即面试”同排。 |
| 2026-07-15 | 1.34 | 用户确认报告信息层级改为 `3/2/2/2/1`：顶部准备度卡片下移为底部全宽“面试总评”，与服务端 `summary` 只在该处展示；mobile 保持同序单列。 |
| 2026-07-14 | 1.33 | 将 ReportsScreen 入口与 Back 锚定到 Workspace targetJobId 只读详情；Parse 只保留新导入命令进度。 |
| 2026-07-14 | 1.32 | 增加独立 target-scoped ReportsScreen，锁定 current/latest-only、规划隔离、四态与 desktop/mobile 响应式合同；Report/Generating trusted Back 改回该列表。 |
| 2026-07-14 | 1.31 | 将 Report/Generating Back 收敛为 trusted target -> Parse reports anchor、无可信 identity -> Workspace fallback，并禁止顶层报告中心与 route target authority。 |
| 2026-07-13 | 1.30 | Context Strip 删除 session/report UUID 等内部 locator，只保留 target/round/resume，并要求 desktop/mobile 可见与可访问负向验收。 |
| 2026-07-13 | 1.29 | Correct report timing ownership to action-local initial+3 with10s/20s/40s; async attempts are infrastructure-only. Keep maxAttempts49/6m04s and no unsupported failed-report regenerate UI. |
| 2026-07-13 | 1.28 | Lock report use of business10s/20s/40s under durable max4 and frontend maxAttempts49 (~6m04s)；separate business cap80 from infra delivery and expose no internal attempt/progress. |
| 2026-07-13 | 1.27 | Clarify product full-validator repair scope：sole-label targeted，all other/mixed whole-report，one-budget full revalidation；visible UI boundary unchanged. |
| 2026-07-13 | 1.26 | 方案 A 最终边界：200 code-point wire fuse；24-word/64-code-point semantic/UX；targeted repair 内部余量18/52；desktop+390完整换行，超限typed invalid/no raw。 |
| 2026-07-13 | 1.25 | A-200：wire/schema fuse改为200；14/40仍为UX gate，desktop+390合法边界完整换行，超限typed invalid且不回显raw。 |
| 2026-07-13 | 1.24 | 归一化 action-label schema120/语言14-40 violation set；即使 label>120 导致 schema-invalid 仍使用 action_labels，修复同时满足两层上限。 |
| 2026-07-13 | 1.23 | Runtime 使用一次总预算下的整报告 / 唯一 action-length label-only LLM repair，labels-only 原样 merge并全量复验；evalkit 分界由 F3/P0.100 owner 承接。 |
| 2026-07-13 | 1.22 | 区分 120-char wire/schema fuse、P0.099 current-run canonical audit chain 与确定性 14/40 boundary fixture 响应式验证；P0.100 内容可靠性不与六图强绑 output digest。 |
| 2026-07-12 | 1.21 | 修复两套页面同时错误导致的 mobile 英文能力维度逐字符竖排，定义 label 与 status 可读换行契约。 |
| 2026-07-12 | 1.20 | 固化 Report mobile TopBar 响应式换行、内容起点和无横向溢出的绝对 viewport 契约。 |
| 2026-07-12 | 1.19 | 明确 Replay 在无可支持 focus 时创建空 focus 的通用同轮复练，并补 `REPORT_CONTEXT_TOO_LARGE` 的诚实终态与可执行返回指引。 |
| 2026-07-12 | 1.18 | 补 frozen context/reportId-only 事实源、终态动作矩阵、records 负向边界与 UI/report 双语言契约。 |
| 2026-07-12 | 1.17 | 统一三指标四常驻区块；接入 direct semantic summary/code+label，删除 generating 伪实时语义，补齐 enum i18n、CTA 推荐、server-owned focus、mobile 可读性与强截图 gate。 |
| 2026-07-12 | 1.16 | 下一轮只使用 TargetJob 有序结构化轮次的紧邻后一项；异常状态 fail closed。 |
