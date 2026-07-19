# Interview 面试规划目标模块

> **版本**: 1.48
> **状态**: completed
> **更新日期**: 2026-07-20

## 1 文档目的

本文档定义当前静态 UI 中 `面试` 一级模块的目标结构。`/workspace` 无 `targetJobId` 时展示可继续的面试规划列表，`/workspace?targetJobId=...` 时展示该规划的统一只读“面试规划详情 / 面试上下文确认”母版；列表卡片主体和 Home ready 最近记录都直接进入该 Workspace 详情，不播放解析动画。`/parse?targetJobId=...` 仅承接首页新建 JD 后的 queued/processing 命令进度；分析 ready 后以 replace 导航到 Workspace 详情。卡片右上角展示删除图标按钮，卡片底部只展示 `立即面试` 主按钮，不再展示可见的 `进入规划` 按钮；删除图标调用 generated `archiveTargetJob` 持久软归档，成功后卡片移出列表且刷新后不得回灌。该模块是既有面试规划的回访入口，不是“当前岗位”页。首页最近模拟面试只展示 3 条全宽横向记录，复用同一个 TargetJob round/progress/action mapper，但使用 Home 专属 record presentation，不展示删除按钮；有记录时固定显示「查看全部」进入 Workspace 列表。首次导入新 JD 时，首页以一个白色 intake card 组合 JD textarea/runtime count、selectable 简历下拉框、创建入口、「立即面试」CTA 与隐私提示；selectable 指未归档且 `parseStatus=ready` 或已有可读正文/结构化证据。提交 `{ rawText, targetLanguage, resumeId }` 后只进入 Parse 命令进度，ready 后由 Workspace 详情只读展示 JD / 简历 / 轮次上下文。缺少或无效简历的历史规划属于异常数据：Start、Reports、复练和下一轮全部 fail closed，不在当前规划上补绑，不默认选择最近简历，也不提供无简历训练或报告降级路径。

Home 的唯一 JD textarea 横向继续占满 intake card，不允许横向随文本增长。默认可视高度由 `106px` 提升为 `212px`；输入或粘贴内容超过默认高度后，textarea 按实际内容高度自动增高，完整显示当前文本且不出现内部纵向滚动条。删除内容时允许自动回缩，但不得低于 `212px`；desktop/mobile 均保持 `width: 100%` 与页面无横向溢出。该视觉行为不改变 runtime UTF-8 byte limit、计数器、提交请求、route、隐私或简历前置合同。

## 2 模块职责

当前面试规划模块负责：

- 展示面试规划列表，并让一级 `面试` 入口默认有可继续对象。
- 列表候选只来自 ready 且标题非空的 TargetJob。
- 点击卡片主体导航到 `/workspace?targetJobId=...` 统一只读详情母版，不触发 import、poll 或 Parse animation。
- 通过 `立即面试` 主按钮快速启动该规划的模拟面试。
- 卡片 mini round rail 与 `立即面试` 当前轮只使用 TargetJob API 的 `practiceProgress`；完成一轮后刷新/回访必须显示 backend 投影的下一轮。
- 通过删除图标调用 `archiveTargetJob` 持久归档，并在成功后从当前列表中移除卡片。
- 引导用户回首页导入新的 JD。

当前面试规划不负责：

- 作为“当前岗位”一级模块存在。
- 接受 `planId` / `resumeId` 作为详情 locator；Workspace 详情只接受 `targetJobId`，其余事实从受保护 API 读取。
- 作为首次 JD 导入链路中 Parse 命令进度之外的第二套解析动画或重复确认。
- 在 ready 卡片回访时重新调用 import、启动 polling 或 materialize Parse 命令进度。
- 提供保存规划按钮、切换简历入口或持久化删除 TargetJob。
- 管理多岗位资产的全部生命周期。
- 管理简历资产。
- 展示练习模式卡片。
- 提供热身、反问、单题深钻或针对性复练。
- 承载无上下文的面试报告入口。

## 3 页面框架

```text
[Interview / Plan List]
├─ Header
│  ├─ Title: 面试规划
│  ├─ Subtitle: 选择一个已有 JD / 规划继续准备
│  └─ 新建规划 -> 返回首页导入新的 JD
├─ Plan Cards
│  ├─ 状态 / 更新时间
│  ├─ 岗位
│  ├─ 公司 · 地点
│  ├─ 右上角删除图标
│  └─ 立即面试
└─ Empty
   └─ 暂无规划 -> 返回首页导入 JD

[Parse / Import Command Progress]
├─ Loading（仅新建 JD 的 queued / processing）
│  ├─ 解析中标题 / 等待说明
│  └─ 四步进度（当前步显示处理中动画）
│     └─ 不展示 model/provider、rubric/prompt/version/hash、provenance 或典型耗时
├─ ready -> replace `/workspace?targetJobId=...`
└─ 不渲染 ready 详情、不承接卡片回访

[Workspace / Unified Read-only Plan Detail]
├─ Back
│  └─ 返回面试规划列表 `/workspace`
├─ Header
│  ├─ Title Cluster: 面试规划详情 / 面试上下文确认 + 绑定简历链接
│  ├─ 来源 / 更新时间
│  ├─ 说明: 面试对话只使用本页确认的 JD、简历和轮次上下文
│  └─ Leading Action Row: 立即面试 + 面试报告
├─ Basic Fields
│  ├─ 岗位名
│  ├─ 公司
│  ├─ 职级 / 语言
│  └─ 地点
├─ Requirements
│  ├─ 必需项
│  └─ 加分项
├─ Hidden Signals
│  └─ 隐性关注点
├─ Round Assumptions
│  ├─ R1（name / type / duration / focus 来自 TargetJob.summary.interviewRounds[0]）
│  ├─ R2（name / type / duration / focus 来自 TargetJob.summary.interviewRounds[1]）
│  ├─ R3（name / type / duration / focus 来自 TargetJob.summary.interviewRounds[2]）
│  └─ Rn（轮次数量由 TargetJob.summary.interviewRounds.length 决定）
└─ 无独立 Interview Launch / 绑定简历大卡片 / Footer Actions
```

Workspace 详情的 desktop 目标构图使用约 `1250px` 居中内容列：Header 左侧组合“第 2 / 2 步”、标题、绑定简历链接和说明，右侧同行放置“立即面试 / 面试报告”；其后按基本信息、必需项/加分项、隐性关注点和轮次安排形成四层白色卡面。基本信息使用紧凑字段网格，必需项与加分项在同一卡内分区，隐性关注点使用主题强调面，轮次按真实数组形成响应式卡片网格。mobile 保持 Header、动作与四层内容的原顺序并改为单列，不得引入第二套 ready 页面、静态轮次或页尾动作。

## 4 关键交互

### 4.1 进入面试规划列表与当前规划

```text
入口:
├─ Home 最近模拟面试记录（最多 3 条）或“查看全部”
├─ Home 新建规划快捷入口（粘贴 JD + 选择已有简历 + JD import）
├─ 一级导航 面试
├─ Report 的复练当前轮
└─ Report 的进入下一轮

进入:
  /workspace -> Interview Plan List
  点击卡片主体 -> /workspace?targetJobId=...
  Home POST import -> /parse?targetJobId=... -> ready replace /workspace?targetJobId=...
  点击立即面试 -> 全屏面试准备过渡态 -> start practice session
```

一级 `面试` 入口不得默认展示“没有 JD 上下文”的死胡同；query-free `/workspace` 展示面试规划列表，列表候选来自当前 `listTargetJobs(analysisStatus=ready)` 契约。Workspace list 按参考稿使用宽卡承载规划；Home recent 则以全宽横向 record 承载同一个公司/岗位、动态 mini round rail、最近使用时间和 quick-start 语义。两种 presentation 必须共用 TargetJob/round/progress/action mapper，但不要求视觉盒型相同；均不得展示 TargetJob lifecycle `status` 文案或空地点占位。Workspace 卡片底部左侧显示由 `updatedAt` 派生的本地化“上次保存”，右侧保留唯一 `开始模拟面试` 主按钮，删除图标固定在卡片右上角。页面背景层必须从 TopBar 下沿连续覆盖整个可视窗口宽度与剩余高度，不得被 1508px 内容容器或 `overflow` 裁剪；标题、按钮和卡片另由居中内容层约束。1916px desktop 内容区约 1508px，实际 header 与 card grid 共享 1456px 布局右边界，使“新建面试规划”按钮右侧与第二张卡片右侧精确对齐；卡片使用两列等宽宽卡与约 28px 列间距，1 张卡片仍保持单列宽度，mobile 收敛为同序单列。卡片信息必须保持简洁，不得展示 `sourceType`、目标语言、`手动输入` 等导入元信息；解析失败、非 ready 或空标题 JD 不得显示。主体点击进入 `/workspace?targetJobId=...` 统一只读详情；启动按钮使用主题 accent 样式并直接启动 practice。首次导入主路径为 `Home 粘贴 JD -> 选择已有简历 -> POST import -> Parse queued/processing -> ready replace Workspace 详情 -> practice`；回访既有 ready 规划直接进入 Workspace 详情，不经过 Parse 动画。

### 4.2 切换或新建规划

```text
用户不想继续当前规划
  -> 返回 workspace 规划列表
     ├─ 选择另一张 ready TargetJob 卡片
     └─ 进入 `/workspace?targetJobId=...` 统一只读详情
  -> 或点击导入 JD
     └─ 回到首页导入新的 JD
```

这解决用户从一级 `面试` 进入后，不想继续最近规划时无路可走的问题。`workspace` 不提供 Plan Switcher Modal；列表本身就是切换入口。

### 4.3 简历绑定只读

```text
workspace 只读详情标题旁的“绑定简历”链接
  -> 使用 TargetJob API 已保存的 resumeId
  -> 跳转 resume_versions?resumeId=...
  -> 缺失或无效时阻断立即面试
  -> 用户想换简历时回到 Home 用目标 JD + 新简历创建新规划
```

简历绑定不属于当前规划详情的可变字段。Home 导入时已经强制选择 ready 简历；解析成功即保存该上下文快照。标题旁的“绑定简历”只用于查看对应简历详情，不触发 `getResume` 预读、不提供 resume picker、创建简历兜底或 in-place rebind。缺失绑定时显示非链接的异常状态，并禁用“立即面试”和“面试报告”；复练/下一轮也不得从该规划继续。不得用 route/list item/最近简历补齐，不存在无简历兼容模式。

### 4.4 立即面试

```text
workspace 只读详情 / report 复练入口
  -> 立即面试
  -> createPracticePlan（必要时）
  -> startPracticeSession
  -> Interview Session(sessionId)
     ├─ 文本面试
     └─ 电话模式
```

`立即面试` 必须携带已保存的 `planId + targetJobId + jdId + resumeId + roundId`，并通过 generated REST client 创建/启动 session。`workspace` 不读取 `autoStartPractice`，也不作为启动副作用路由。面试形式可在面试页选择或切换。规划列表页不展示模式卡片，也不让用户选择专项练习。

Home 最近面试、Workspace 列表、Workspace 详情以及 Report 复练/下一轮都必须复用同一面试准备过渡态。用户触发启动后立即以覆盖当前页面的阻断层展示诚实的 indeterminate 动画和“正在结合岗位、简历与本轮重点准备开场”的本地化说明；不得等待 `startPracticeSession` 返回后才反馈，也不得伪造百分比、阶段完成或 opening message。过渡态使用 `role=status`、`aria-live=polite` 与 `aria-busy=true`，阻断重复点击并支持 `prefers-reduced-motion`。成功后直接进入 `practice`；失败时关闭过渡态并在原入口附近保留可恢复错误，不新增 route、浏览器持久化或取消后继续运行的隐式行为。未登录时仍先进入认证流程，不提前显示启动过渡态。

在 Workspace 详情中，“立即面试”与“面试报告”组成标题下方的首行动作行，均从左侧开始排列；前者为 primary，后者为 secondary。desktop 保持同一行，mobile 优先保持同序横排，空间不足时允许整组换行但不得调换顺序、右对齐或把任一动作移回页尾。启动错误紧邻该动作行呈现。

### 4.5 公司情报嵌入卡片

```text
公司情报摘要（详情或报告 owner）
├─ 一句话画像
├─ 近期公开信号（精选）
├─ 反问建议（精选）
└─ 合规来源数量与刷新时间说明
```

公司信号只使用合规公开来源，不展示雇主评分聚合、不抓登录后内容、不使用私域数据。query-free Workspace 列表不展示公司情报卡片；若 Workspace 详情或报告 owner 需要摘要，必须使用 typed TargetJob / Report 字段重新接入。

### 4.6 查看当前规划的轮次报告

```text
workspace 统一详情标题下方首行动作行中的“面试报告”
  -> reports?targetJobId=...
  -> 独立 ReportsScreen 按当前 TargetJob canonical rounds 展示
     ├─ currentReport -> report?reportId=...
     ├─ latestAttempt queued/generating -> generating?reportId=...
     ├─ latestAttempt failed -> 普通失败提供同 report 重新生成 + 查看记录；超限失败只查看记录
     └─ 都为空 -> 该轮暂无报告
```

query-free Workspace 列表仍不展示模拟面试记录，也不新增无上下文或跨规划报告中心。唯一页面级入口位于 Workspace 详情标题下方首行动作行，与“立即面试”左对齐且同排，不进入全局 TopBar；Parse 只承接新导入 queued/processing，不渲染 ready 详情或报告入口、不调用 `listTargetJobReports`、不保留 `section=reports`。独立 ReportsScreen 同时读取当前 TargetJob 与 typed `listTargetJobReports(targetJobId)`，把每个 `PracticeRoundRef` 与当前 `TargetJob.summary.interviewRounds[]` join 后展示；只允许当前规划。`currentReport` 与 `latestAttempt` 是独立指针；较新的失败/生成尝试不能隐藏较早可用报告，latest ready 与 current 相同时不得重复，ID 不同时只说明最近生成已完成而不展开历史。

ReportsScreen 的 loading/empty/error/ready 是独立页面状态，target/round mismatch、跨用户或 stale response 必须整页 fail closed，不渲染其他规划 sentinel。Reports Back 返回 `/workspace?targetJobId=...` 只读详情。从 Report / Generating 返回时，若 trusted API context 能提供 `targetJobId`，Back 导航到 `/reports?targetJobId=...`；缺失可信 TargetJob identity、首读 404/网络失败或 invalid payload 时返回 query-free `/workspace`。route 不得自行拼接或覆盖 report identity。

## 5 数据对象

```text
MockInterviewPlan
├─ planId
├─ targetJobId
├─ jdId
├─ resumeId
├─ roundId
├─ status
├─ updatedAt
├─ latestSessionId
└─ previousReportSignals

TargetJob
├─ title
├─ company
├─ location
├─ level
├─ source
├─ jdText
├─ jdAnalysis
├─ match
└─ interviewRounds

TargetJobReportsOverview
├─ targetJobId
└─ rounds[]
   ├─ round: PracticeRoundRef
   ├─ currentReport: id + generatedAt | null
   └─ latestAttempt: id + status + errorCode + createdAt | null

Resume
├─ resumeId
├─ displayName
├─ source
└─ matchSummary
```

## 6 信息层级

```text
最高优先级:
├─ 有哪些可继续的 ready 面试规划
├─ 每个规划的岗位 / 公司 / 地点 / 更新时间
├─ 如何进入某个规划详情
├─ 如何立即开始某个规划的面试
└─ 如何导入新的 JD

次级信息:
├─ JD 拆解（workspace 详情）
├─ 标题旁绑定简历查看入口（workspace 详情）
├─ 当前轮次（workspace 详情）
├─ 每轮最后可用报告与最新生成状态（独立 ReportsScreen）
└─ 删除列表卡片（generated archiveTargetJob -> target_jobs.deleted_at）

低频信息:
├─ JD 原文
├─ 数据来源
└─ 更新时间
```

## 7 范围边界

| 范围外模块或流程 | 当前面试规划内是否保留相关能力 | 边界 |
|------------------|-------------------------------|------|
| 成长 | 否 | 不展示长期成长中心 |
| 多轮计划 | 否 | 只展示当前目标岗位的面试轮次节点，不维护独立计划系统 |
| 经历库 | 否 | 面试可读取简历和画像，但不要求用户维护经历库 |
| 追问树 | 否 | 追问在面试会话内发生，不作为入口 |
| 针对性复练 | 否 | 报告可发起复练当前轮，但不是单题 Drill |
| 练习模式选择 | 否 | workspace 列表只提供卡片进入 Workspace 详情和 `立即面试` 快速启动，不提供模式选择 |
| 全局/跨规划报告中心 / timeline | 否 | 允许 Workspace 详情内容区的页面级入口与 target-scoped ReportsScreen；完整报告仍是 reportId-only Dashboard，不展示完整历史 |

## 8 后续实现输入

1. 顶部导航文案为 `面试`；英文为 `Interview`。
2. `/workspace` 无 `targetJobId` 时展示 `面试规划列表`；`/workspace?targetJobId=...` 展示该规划只读详情。`planId` / `resumeId` 不是详情 locator，必须清理或忽略；不得再用 `当前岗位` 表示一级模块。
3. 面试规划列表必须是固定列宽卡片式；Home recent 必须是全宽横向 record。两者共用公司、岗位、动态 mini round rail、progress 与 quick-start mapper，但 Home 增加最近使用时间并不展示删除，Workspace 在卡片底部追加主题 accent `立即面试` 且右上角展示删除。两者都不得展示 TargetJob lifecycle `status`、空地点占位、来源类型、目标语言或 `手动输入` 等低价值元信息；只展示 `analysisStatus=ready` 且标题非空的 TargetJob。
4. 列表卡片不展示可见的 `进入规划` / `Open plan` 按钮；点击卡片主体进入 `/workspace?targetJobId=...` 详情且不得触发 import/poll/Parse animation，点击 `立即面试` 启动 practice，点击右上角删除图标调用 generated `archiveTargetJob`，成功后隐藏当前卡片且刷新后不回灌。
5. 真实面试轮次、已绑定简历和启动面试只出现在 Workspace 只读详情或后续 owner；Workspace 详情 round assumptions 与 Home 最近模拟面试卡片的迷你轮次轨道遵循本文档的同一视觉语义，但轮次数量、type/name、duration 和 focus 必须来自同一个 `TargetJob.summary.interviewRounds[]` mapper。该数组由后端 LLM 根据 JD、岗位级别、公司/行业性质、团队/业务上下文和招聘流程线索推断；前端不得用静态 4 轮、静态 HR/技术/经理面或静态分钟数 fallback。Workspace 规划列表保持紧凑卡片，但进入详情的 handoff 不得生成另一套静态 round name。
   当前/已完成状态必须来自 `TargetJob.practiceProgress`：`completedRounds` 画为完成态，`currentRound` 画为当前态，全部完成时所有节点为完成态且 `立即面试` disabled。缺失、跳轮、重复或 pair 不匹配时不高亮/不启动；禁止读取 lifecycle `status`、自由文本 `nextRound`、URL 或浏览器存储做轮次 fallback。mini rail 的 DOM、间距、颜色、节点几何以正式前端当前 token 和 component contract 为准。Workspace 详情的 round assumption 卡同步表达同一事实：done 显示“已进行”并使用 success-soft 背景/成功色边框，current 显示“即将进行”并使用 accent-soft 背景/主题色边框，pending 显示“未进行”并使用 neutral-soft 背景/规则线边框；三态必须有 `data-round-state`，不能只靠颜色传达。
6. Workspace detail 标题 cluster 在“面试规划详情”旁展示“绑定简历”查看链接，点击精确进入 `resume_versions?resumeId=<TargetJob.resumeId>`；不得渲染独立绑定简历/Interview Launch 大卡片，不得从 route/list item/最近简历推断绑定。缺失绑定显示非链接异常状态，Start、Reports、复练和下一轮全部 fail closed。
7. Workspace detail 是首次导入 ready 后和既有规划回访的同一只读母版；不得另设第二个 ready 确认页面，详情页不提供“仅保存规划”。
8. `/parse?targetJobId=...` 只保留新导入 queued/processing 的四步进度、当前步处理动画与面向用户的等待说明；ready 必须 replace 到 `/workspace?targetJobId=...`，既有 ready 卡片不得进入 Parse。内部 model/provider、rubric/prompt/version/hash、provenance、typical latency 不得出现在 prototype、formal DOM 或 desktop/mobile 截图中。
9. Home JD intake 只接受粘贴文本；prototype、formal DOM、OpenAPI 请求与 desktop/mobile 截图不得出现平行 JD 导入控件、弹窗或 source discriminator。Resume 上传能力属于 Resume owner，必须继续保留。
10. Workspace detail 标题下方首行动作行从左依次提供“立即面试”和“面试报告”，后者精确导航 `/reports?targetJobId=...`；两者 desktop 同排，mobile 同序响应式换行，均不进入 TopBar 或页尾。Parse ready 详情、内嵌列表、列表请求和 `section=reports` 兼容逻辑必须为零。独立 ReportsScreen 逐项覆盖当前 TargetJob canonical rounds；display 只来自当前 TargetJob，overview 只提供 `PracticeRoundRef/currentReport/latestAttempt`；ready 链接 report、queued/generating 链接 generating、普通 failed 提供同报告重新生成与查看记录、超限 failed 只查看记录；loading/empty/error/identity mismatch 完整 fail closed，且不展开历史版本。

## 9 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-20 | 1.48 | Home JD textarea 默认可视高度翻倍为 212px，并随输入内容自动增高/回缩以完整显示文本；横向宽度、runtime limit 与业务合同不变。 |
| 2026-07-19 | 1.47 | 按参考稿锁定 Workspace 详情的 1250px Header 动作区与基本信息、要求、隐性关注点、轮次四层卡面构图；业务 owner 与动态轮次合同不变。 |
| 2026-07-19 | 1.46 | 明确 Workspace 背景层全视口覆盖、内容层独立限宽，并让 header CTA 与双列卡片网格共享右边界；禁止背景裁剪空白带和按钮右侧错位。 |
| 2026-07-19 | 1.45 | Workspace list 按提供的桌面参考稿改为 1508px 内容区、双列宽卡、上次保存 footer 与参考级动作几何；mobile 保持单列，业务 mapper/route/API 不变。 |
| 2026-07-19 | 1.44 | Home 按桌面参考图重组为单一 intake card 与全宽横向 recent records；Workspace 仍保留固定列宽卡片，两者共用业务 mapper 而非强制同一盒型。 |
| 2026-07-18 | 1.43 | 所有正式面试启动入口统一增加可访问、诚实且 reduced-motion 兼容的全屏面试准备过渡态；成功进入 Practice，失败回到原入口错误。 |
| 2026-07-17 | 1.42 | 面试规划卡片移除重复且不可操作的 lifecycle status；地点缺失时不再显示 `Location not set` 占位。 |
| 2026-07-16 | 1.41 | 普通失败报告提供同 report 重新生成与查看记录；超限失败只保留只读面试记录恢复入口。 |
| 2026-07-15 | 1.40 | 将 selectable 简历锁定为 Workspace 及其报告后动作的强制上下文；历史缺绑规划视为异常并全链路 fail closed，不提供无简历降级。 |
| 2026-07-15 | 1.39 | 删除 Workspace 详情独立 Interview Launch/绑定简历大卡片；标题旁新增绑定简历详情链接，并将立即面试与面试报告移到左对齐首行动作行。 |
| 2026-07-14 | 1.38 | Workspace 详情轮次假设复用列表 rail 的 persisted progress，增加已进行/即将进行/未进行三种背景、边框、标签与状态属性。 |
| 2026-07-14 | 1.37 | 将 Workspace 明确拆为无参列表与 targetJobId 只读详情；ready 卡片直达详情，Parse 仅承接新导入 queued/processing 并在 ready 后 replace，Reports Back 返回 Workspace 详情。 |
| 2026-07-14 | 1.36 | 将报告列表从 Parse 内嵌区迁移到 target-scoped ReportsScreen；Parse 只保留内容区右上入口，并删除列表请求与 section 兼容。 |
| 2026-07-14 | 1.35 | 在 Parse 统一详情增加按 canonical round 的报告入口，锁定 current/latest 独立状态、report/generating 链接、独立失败边界与 Report Back 锚点；不新增顶层报告中心。 |
| 2026-07-13 | 1.34 | Home JD intake 收敛为唯一粘贴文本框、ready 简历下拉框与主 CTA，并固定 `{ rawText, targetLanguage, resumeId }` handoff。 |
| 2026-07-13 | 1.33 | 补齐 Parse loading 页面树并删除用户界面中的模型、rubric、provenance 与典型耗时等内部元数据。 |
