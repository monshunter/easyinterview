# Interview 面试规划目标模块

> **版本**: 1.36
> **状态**: active
> **更新日期**: 2026-07-14

## 1 文档目的

本文档定义当前静态 UI 中 `面试` 一级模块的目标结构。`workspace` / 顶部导航 `面试` 是纯列表页，只展示可继续的面试规划列表，不读取也不解释 `targetJobId` / `planId` / `resumeId` 等详情上下文。规划详情由 `parse` route 的统一“面试规划详情 / 面试上下文确认”母版承接，列表卡片点击卡片主体时导航到 `parse?targetJobId=...`；卡片右上角展示删除图标按钮，卡片底部只展示 `立即面试` 主按钮，不再展示可见的 `进入规划` 按钮；删除图标调用 generated `archiveTargetJob` 持久软归档，成功后卡片移出列表且刷新后不得回灌。该模块是既有面试规划的回访入口，不是“当前岗位”页：用户从首页最近模拟面试、报告、会话记录或一级导航回到这里，只浏览规划列表并选择一个规划继续。首页最近模拟面试只展示 3 条快捷卡片，复用同一卡片主体和 `立即面试` 主按钮但不展示删除按钮，`更多` 进入本页列表。首次导入新 JD 时，首页作为新建面试规划快捷入口，只保留 JD textarea、ready 简历下拉框与「立即面试」CTA；`还没有简历？1 分钟创建 →` 与下拉框同一行水平对齐。提交 `{ rawText, targetLanguage, resumeId }` 后进入统一详情页，后者只读展示 JD / 简历 / 轮次上下文；缺少或无效简历时只阻断开始，不在当前规划上补绑或默认替用户选中最近简历。

## 2 模块职责

当前面试规划列表负责：

- 展示面试规划列表，并让一级 `面试` 入口默认有可继续对象。
- 列表候选只来自 ready 且标题非空的 TargetJob。
- 点击卡片主体导航到 `parse` route 统一详情母版。
- 通过 `立即面试` 主按钮快速启动该规划的模拟面试。
- 卡片 mini round rail 与 `立即面试` 当前轮只使用 TargetJob API 的 `practiceProgress`；完成一轮后刷新/回访必须显示 backend 投影的下一轮。
- 通过删除图标调用 `archiveTargetJob` 持久归档，并在成功后从当前列表中移除卡片。
- 引导用户回首页导入新的 JD。

当前面试规划不负责：

- 作为“当前岗位”一级模块存在。
- 承接 `targetJobId` / `planId` / `resumeId` 等详情上下文。
- 作为首次 JD 导入链路中 `parse` 与 session 之间的第二个全页确认。
- 维护一套独立于 `JD 解析结果` 母版之外的 workspace 详情页视觉。
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

[Parse / Unified Plan Detail]
├─ Loading（仅首次导入）
│  ├─ 解析中标题 / 等待说明
│  └─ 四步进度（当前步显示处理中动画）
│     └─ 不展示 model/provider、rubric/prompt/version/hash、provenance 或典型耗时
├─ Back
│  ├─ 首次导入: 返回首页
│  └─ 规划回访: 返回面试规划列表
├─ Header
│  ├─ 面试规划详情 / 面试上下文确认
│  ├─ 来源 / 更新时间
│  ├─ 说明: 面试对话只使用本页确认的 JD、简历和轮次上下文
│  └─ 右上角 面试报告 -> reports?targetJobId=...
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
├─ Interview Launch
│  ├─ 已绑定简历
│  └─ 缺简历时阻断开始
└─ Footer Actions
   └─ 立即面试
```

## 4 关键交互

### 4.1 进入面试规划列表与当前规划

```text
入口:
├─ Home 最近模拟面试卡片（最多 3 条）或“更多”
├─ Home 新建规划快捷入口（粘贴 JD + 选择已有简历 + JD import）
├─ 一级导航 面试
├─ Report 的复练当前轮
└─ Report 的进入下一轮

进入:
  workspace -> Interview Plan List
  点击卡片主体 -> parse?targetJobId=...(+planId/resumeId)
  点击立即面试 -> start practice session
```

一级 `面试` 入口不得默认展示“没有 JD 上下文”的死胡同；`workspace` 永远展示面试规划列表，列表候选来自当前 `listTargetJobs(analysisStatus=ready)` 契约。列表必须以卡片承载每个规划：卡片主体与 Home 最近模拟面试卡片同源，包含公司/状态 eyebrow、岗位、地点和 mini round rail；workspace 只在卡片底部追加 `立即面试` 主按钮，并把删除图标固定在卡片右上角。桌面端响应式多列，移动端单列；桌面端卡片列必须使用固定最大列宽，1 张、2 张或 3 张规划卡片的规格保持一致，不得因为 `auto-fit + 1fr` 拉伸成整行宽卡；不得退化为没有容器感的文本列。卡片信息必须保持简洁，不得展示 `sourceType`、目标语言、`手动输入` 等导入元信息；解析失败、非 ready 或空标题 JD 不得显示为面试规划卡片。卡片主体点击进入 `parse` 统一面试规划详情页；底部 `立即面试` 按钮使用主题 accent 样式并直接启动 practice；右上角删除按钮使用简历列表同款 trash 图标样式，调用 generated `archiveTargetJob` 软归档 TargetJob，成功后从当前列表移除，失败时保留卡片并展示错误。首次导入主路径为 `Home 粘贴 JD -> 选择已有简历 -> 立即面试 -> 面试规划详情核对 -> practice`；回访既有规划、点击列表卡片也进入同一个只读详情母版，不再经过另一套 workspace 全页确认。

### 4.2 切换或新建规划

```text
用户不想继续当前规划
  -> 返回 workspace 规划列表
     ├─ 选择另一张 ready TargetJob 卡片
     └─ 进入 parse 统一详情
  -> 或点击导入 JD
     └─ 回到首页导入新的 JD
```

这解决用户从一级 `面试` 进入后，不想继续最近规划时无路可走的问题。`workspace` 不提供 Plan Switcher Modal；列表本身就是切换入口。

### 4.3 简历绑定只读

```text
parse 统一详情中的绑定简历卡片
  -> 展示已保存的 Resume 摘要
  -> 缺失或无效时阻断立即面试
  -> 用户想换简历时回到 Home 用目标 JD + 新简历创建新规划
```

简历绑定不属于当前规划详情的可变字段。Home 导入时已经强制选择 ready 简历；解析成功即保存该上下文快照。详情页不提供 resume picker、创建简历兜底或 in-place rebind，避免同一个规划在面试前后出现不同上下文。

### 4.4 立即面试

```text
parse 统一详情 / report 复练入口
  -> 立即面试
  -> createPracticePlan（必要时）
  -> startPracticeSession
  -> Interview Session(sessionId)
     ├─ 文本面试
     └─ 电话模式
```

`立即面试` 必须携带已保存的 `planId + targetJobId + jdId + resumeId + roundId`，并通过 generated REST client 创建/启动 session。`workspace` 不读取 `autoStartPractice`，也不作为启动副作用路由。面试形式可在面试页选择或切换。规划列表页不展示模式卡片，也不让用户选择专项练习。

### 4.5 公司情报嵌入卡片

```text
公司情报摘要（详情或报告 owner）
├─ 一句话画像
├─ 近期公开信号（精选）
├─ 反问建议（精选）
└─ 合规来源数量与刷新时间说明
```

公司信号只使用合规公开来源，不展示雇主评分聚合、不抓登录后内容、不使用私域数据。当前 `workspace` 纯列表不展示公司情报卡片；若后续详情或报告 owner 需要摘要，必须使用 typed TargetJob / Report 字段重新接入。

### 4.6 查看当前规划的轮次报告

```text
parse 统一详情内容区右上角“面试报告”
  -> reports?targetJobId=...
  -> 独立 ReportsScreen 按当前 TargetJob canonical rounds 展示
     ├─ currentReport -> report?reportId=...
     ├─ latestAttempt queued/generating -> generating?reportId=...
     ├─ latestAttempt failed -> 本地化失败状态，无同 report Retry
     └─ 都为空 -> 该轮暂无报告
```

`workspace` 列表仍不展示模拟面试记录，也不新增无上下文或跨规划报告中心。唯一页面级入口位于 `parse` 内容区标题行右上角，不进入全局 TopBar；Parse 自身不渲染列表、不调用 `listTargetJobReports`、不保留 `section=reports`。独立 ReportsScreen 同时读取当前 TargetJob 与 typed `listTargetJobReports(targetJobId)`，把每个 `PracticeRoundRef` 与当前 `TargetJob.summary.interviewRounds[]` join 后展示；只允许当前规划。`currentReport` 与 `latestAttempt` 是独立指针；较新的失败/生成尝试不能隐藏较早可用报告，latest ready 与 current 相同时不得重复，ID 不同时只说明最近生成已完成而不展开历史。

ReportsScreen 的 loading/empty/error/ready 是独立页面状态，target/round mismatch、跨用户或 stale response 必须整页 fail closed，不渲染其他规划 sentinel。Reports Back 返回 `parse?targetJobId=...`。从 Report / Generating 返回时，若 trusted API context 能提供 `targetJobId`，Back 导航到 `/reports?targetJobId=...`；缺失可信 TargetJob identity、首读 404/网络失败或 invalid payload 时返回 `workspace`。route 不得自行拼接或覆盖 report identity。

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
├─ JD 拆解（parse 详情）
├─ 绑定简历（parse 详情）
├─ 当前轮次（parse 详情）
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
| 练习模式选择 | 否 | workspace 列表只提供卡片进入规划详情和 `立即面试` 快速启动，不提供模式选择 |
| 全局/跨规划报告中心 / timeline | 否 | 允许 Parse 内容区的页面级入口与 target-scoped ReportsScreen；完整报告仍是 reportId-only Dashboard，不展示完整历史 |

## 8 后续实现输入

1. 顶部导航文案为 `面试`；英文为 `Interview`。
2. `workspace` 永远展示 `面试规划列表`；即使 URL 或历史上下文残留 `targetJobId` / `planId`，也必须清空或忽略这些上下文，不进入详情页；不得再用 `当前岗位` 表示一级模块。
3. 面试规划列表必须是列表卡片式：每个规划复用 Home 最近模拟面试卡片主体（公司/状态 eyebrow、岗位、地点、mini round rail），并在底部追加明确的主题 accent `立即面试` 按钮，删除图标按钮固定在卡片右上角；桌面卡片采用固定最大列宽，单卡不得铺满整行，数量从 1 到 3 变化时卡片规格保持稳定；卡片不展示来源类型、目标语言或 `手动输入` 等导入元信息；列表只展示 `analysisStatus=ready` 且标题非空的 TargetJob。
4. 列表卡片不展示可见的 `进入规划` / `Open plan` 按钮；点击卡片主体进入 `parse` 详情，点击 `立即面试` 启动 practice，点击右上角删除图标调用 generated `archiveTargetJob`，成功后隐藏当前卡片且刷新后不回灌。
5. 真实面试轮次、已绑定简历和启动面试只出现在 parse 只读详情或后续 owner；parse round assumptions 与 Home 最近模拟面试卡片的迷你轮次轨道保持 UI 真理源样式，但轮次数量、type/name、duration 和 focus 必须来自同一个 `TargetJob.summary.interviewRounds[]` mapper。该数组由后端 LLM 根据 JD、岗位级别、公司/行业性质、团队/业务上下文和招聘流程线索推断；前端不得用静态 4 轮、静态 HR/技术/经理面或静态分钟数 fallback。Workspace 规划列表保持紧凑卡片，但进入详情的 handoff 不得生成另一套静态 round name。
   当前/已完成状态必须来自 `TargetJob.practiceProgress`：`completedRounds` 画为完成态，`currentRound` 画为当前态，全部完成时所有节点为完成态且 `立即面试` disabled。缺失、跳轮、重复或 pair 不匹配时不高亮/不启动；禁止读取 lifecycle `status`、自由文本 `nextRound`、URL 或浏览器存储做轮次 fallback。mini rail 的 DOM、间距、颜色、节点几何保持现有原型值不变。
6. 已绑定简历展示、启动面试、公司信号、记录区等详情能力由 `parse` / practice / report 对应 owner 承接，不属于 workspace 列表页。
7. 本页是回访入口：不得把首次 JD 导入用户从 `parse` 强制带回本页做第二次全页确认；解析成功即代表规划已保存，详情页不再提供“仅保存规划”。
8. `parse` 首次导入 loading 只保留四步进度、当前步处理动画与面向用户的等待说明；内部 model/provider、rubric/prompt/version/hash、provenance、typical latency 不得出现在 prototype、formal DOM 或 desktop/mobile 截图中。
9. Home JD intake 只接受粘贴文本；prototype、formal DOM、OpenAPI 请求与 desktop/mobile 截图不得出现平行 JD 导入控件、弹窗或 source discriminator。Resume 上传能力属于 Resume owner，必须继续保留。
10. `parse` 内容区标题行右上角提供“面试报告”入口并精确导航 `/reports?targetJobId=...`；它不进入 TopBar。Parse 内嵌列表、列表请求和 `section=reports` 兼容逻辑必须为零。独立 ReportsScreen 逐项覆盖当前 TargetJob canonical rounds；display 只来自当前 TargetJob，overview 只提供 `PracticeRoundRef/currentReport/latestAttempt`；ready 链接 report、queued/generating 链接 generating、failed 无 Retry，loading/empty/error/identity mismatch 完整 fail closed，且不展开历史版本。

## 9 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-14 | 1.36 | 将报告列表从 Parse 内嵌区迁移到 target-scoped ReportsScreen；Parse 只保留内容区右上入口，并删除列表请求与 section 兼容。 |
| 2026-07-14 | 1.35 | 在 Parse 统一详情增加按 canonical round 的报告入口，锁定 current/latest 独立状态、report/generating 链接、独立失败边界与 Report Back 锚点；不新增顶层报告中心。 |
| 2026-07-13 | 1.34 | Home JD intake 收敛为唯一粘贴文本框、ready 简历下拉框与主 CTA，并固定 `{ rawText, targetLanguage, resumeId }` handoff。 |
| 2026-07-13 | 1.33 | 补齐 Parse loading 页面树并删除用户界面中的模型、rubric、provenance 与典型耗时等内部元数据。 |
