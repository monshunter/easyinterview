# 多 JD 与多简历目标管理结构

> **版本**: 3.6
> **状态**: active
> **更新日期**: 2026-07-14

## 1 文档目的

本文档定义当前 UI 中多份 JD、多份简历以及二者绑定关系的管理方式。JD 和简历都是一等公民；模拟面试规划负责把两者和面试轮次组合成一次模拟面试输入。

## 2 设计原则

1. 多 JD 以 `TargetJob` 管理，入口包括首页 JD 导入和最近模拟面试；岗位推荐模块不属于当前范围（D-17）。
2. 多简历以平铺 `Resume` 资产管理，入口在一级简历模块；不存在版本树、主版本或岗位定制版本（D-20）。
3. 每个模拟面试规划绑定一个 `TargetJob/JD`、一份 `Resume` 和一个 `InterviewRound`。
4. 系统必须同时保留每份简历的原始来源、解析文本和结构化内容。
5. 简历完成态名称由 backend parse 根据 LLM 结构化结果生成；上传 / 粘贴的通用标题只作为解析前来源信息，不作为完成态名称。
6. 简历缺失时不阻断用户看 JD，但在创建个性化模拟面试规划前触发补全和登录；历史规划缺少绑定简历时只阻断开始。
7. 模拟面试规划不是简历资产管理中心，只承载创建规划时已选择的简历快照。
8. Home 的 JD intake 只有粘贴文本；请求形态固定为 `{ rawText, targetLanguage, resumeId }`。Resume 模块仍可上传或粘贴简历，两类入口不得混用。

## 3 数据对象关系

```text
User
├─ TargetJobs
│  ├─ TargetJob(jobId=A)
│  │  ├─ jdText
│  │  ├─ jdAnalysis
│  │  ├─ interviewRounds
│  │  └─ latestMockPlanId
│  └─ TargetJob(jobId=B)
├─ Resumes（平铺）
│  ├─ Resume(resumeId=r-cn)
│  │  ├─ name: 刘哲_前端_中文_2026.pdf
│  │  ├─ sourceType: upload
│  │  ├─ languageTag: 中
│  │  ├─ originalFileRef
│  │  ├─ parsedTextSnapshot
│  │  ├─ structuredContent
│  │  └─ summary
│  ├─ Resume(resumeId=r-en)
│  │  ├─ name: LiuZhe_Frontend_EN_2026.pdf
│  │  ├─ sourceType: upload
│  │  └─ languageTag: EN
│  └─ Resume(resumeId=r-fullstack)
│     ├─ name: 刘哲_全栈方向_2024.pdf
│     └─ sourceType: upload
└─ MockInterviewPlans
   ├─ planId=P1 -> jobId=A + resumeId=r-cn + round=技术一面
   └─ planId=P2 -> jobId=B + resumeId=r-en + round=HR 初筛
```

## 4 多 JD 管理

### 4.1 首页

首页不是岗位资产管理页，而是开始或继续面试的入口，也是 JD 获取的唯一入口。

```text
Home
├─ Start New Interview
│  ├─ JD 输入卡
│  │  └─ 粘贴 JD 输入框（唯一入口）
│  ├─ 选择已有 ready 简历（适度宽度下拉框）
│  │  └─ 还没有简历？1 分钟创建（下拉框右侧同行）
│  └─ 立即面试（简历选择下方）
└─ Recent Mock Interviews
   ├─ 最多 3 张最近卡片，固定最大列宽
   ├─ 岗位名称 / 公司 / 状态 / 面试轮次节点
   ├─ 与 Interview 列表卡片共用卡片主体、mini round rail 和立即面试主按钮
   ├─ 点击卡片主体进入 Interview Plan
   ├─ 不展示删除按钮
   └─ 更多 -> Interview
```

### 4.2 模拟面试规划切换

```text
Mock Interview Plan Header
├─ 当前面试规划
├─ 切换规划
└─ 新建规划
```

切换规划时，页面上下文必须完整切换，包括 JD、简历绑定、当前轮次、会话记录和报告。

## 5 多简历管理

### 5.1 简历入口

```text
Resume / resume_versions
├─ Resume Workshop List（平铺）
│  ├─ 简历名称 / 语言
│  ├─ 来源（上传 / 粘贴）
│  ├─ 最近编辑
│  └─ 打开
├─ 新建简历
│  ├─ 上传文件
│  ├─ 粘贴文本
│  └─ 注册成功后直接打开详情
└─ Resume Detail
   └─ 只读原始简历正文

Workspace Plan Detail(targetJobId)
└─ 绑定简历
   ├─ 只读展示创建规划时已保存的简历摘要
   └─ 缺失或无效时阻断开始；不提供 picker / in-place rebind
```

详情页不再提供原件预览弹层、导出、复制、改写建议或手动编辑；原始简历预览就是当前只读简历正文。

## 6 JD 与简历绑定

```text
MockInterviewPlan
├─ planId
├─ targetJobId
├─ jdId
├─ resumeId
├─ roundId
└─ source: home / report_replay / report_next_round
```

同一份 JD 可以和不同简历形成不同面试规划。同一份简历也可以被多个 JD 使用。

## 7 UI 行为

| 场景 | 目标行为 |
|------|----------|
| 用户新增 JD | 首页先选择已有 ready 简历，再点击「立即面试」POST import；只进入 `/parse?targetJobId` queued/processing 进度，ready 后 replace 到 `/workspace?targetJobId` 只读 JD / 简历 / 轮次详情，唯一成功 CTA 是立即面试 |
| JD 正在解析 | `parse` 只展示四步进度与等待说明；不得向用户展示 model/provider、rubric/prompt/version/hash、provenance 或典型耗时等内部实现元数据 |
| 用户有多份 JD | 首页最多显示最近 3 条模拟面试；卡片主体点击进入规划详情，`立即面试` 主按钮直接启动 practice，不展示删除按钮；更多内容通过“更多”进入一级 `面试` 列表页 |
| 用户不想继续当前规划 | 在面试页点击切换规划或新建规划 |
| 用户首次无简历 | 首页提示创建简历；首页不提供上传简历入口，只跳转到 `resume_versions(flow=create)`；Workspace 详情若发现历史规划缺少绑定简历，只阻断开始，不在当前规划上补绑 |
| 用户上传新简历 | 创建新的 `Resume`，注册成功后直接打开详情 |
| 用户粘贴简历 | 创建新的 `Resume`，保留粘贴文本，根据内容派生临时标题，并注册成功后直接打开详情 |
| 用户查看简历资产 | 平铺列表查看全部简历，打开详情后只阅读原始简历正文 |
| 用户查看原始简历 | 简历详情正文即原始简历预览；不打开独立原件弹层 |
| 用户想更换面试绑定简历 | 回到首页用同一 JD 和新的 ready `Resume` 创建新面试规划；当前规划详情不做 in-place rebind |
| 用户在首页新建面试规划 | 在首页唯一 JD 文本框粘贴 JD；通过适度宽度下拉框选择已有 ready `Resume`，创建简历入口在下拉框右侧同排；未选择简历或未提供 JD 前「立即面试」禁用，按钮位于简历选择下方；提交 `{ rawText, targetLanguage, resumeId }` 后，route 只携带 `targetJobId`，真实 `resumeId` 由 TargetJob 后端事实恢复 |

## 8 不做的事

```text
不做:
├─ 岗位推荐 / 搜岗聚合（D-17）
├─ 简历版本树 / 主版本 / 岗位定制版本 / 分叉（D-20）
├─ 轻量问答建档（D-20）
├─ 独立经历库
├─ 要求用户先维护完整画像才能开始
├─ 多岗位信息混合在一个模拟面试规划里
├─ 让 Settings 承担简历管理职责
├─ 在面试规划详情中更换绑定简历
├─ 在简历详情页提供导出 / 复制 / 编辑 / 改写 / 原件弹层
├─ 从 Home 以文件、岗位链接或结构化表单导入 JD
└─ 覆盖原始来源快照
```

原始来源是用户提供的证据，任何结构化分析都不能覆盖它；详情页只展示当前简历正文。

## 9 修订记录

| 版本 | 日期 | 修订内容 |
|------|------|----------|
| 3.6 | 2026-07-14 | 新导入仅以 Parse 展示 queued/processing，ready replace 到 targetJobId-only Workspace 详情；绑定简历在详情只读，不携带 resumeId route 或提供 picker/rebind。 |
| 3.5 | 2026-07-13 | Home JD intake 收敛为唯一粘贴文本框与 `{ rawText, targetLanguage, resumeId }` 请求合同；Resume 上传 / 粘贴保持不变。 |
| 3.4 | 2026-07-13 | Parse loading 删除 model/rubric/provenance/latency 等内部调试信息，只保留用户可理解的进度与等待状态。 |
| 3.2 | 2026-07-09 | 将 Home 最近模拟面试卡片同步为复用 Interview 列表卡片动作模型：保留立即面试主按钮和卡片点击进入规划，隐藏删除按钮。 |
| 3.1 | 2026-07-09 | 固化 Home 最近模拟面试卡片固定最大列宽，并与 Interview 面试列表卡片共用卡片主体和 mini round rail。 |
| 3.0 | 2026-07-08 | 将最近模拟面试更多入口和多 JD 管理说明同步到一级 `面试` / `Interview` 列表入口，保留完整模拟面试 session 语义。 |
| 2.9 | 2026-07-07 | 未闭环回归修订：简历创建改为注册成功后直接打开详情，删除解析动画和预览确认页；详情展示原始正文。 |
| 2.8 | 2026-07-07 | 多简历管理收敛为平铺列表 + 创建 + 只读详情；删除详情改写、编辑、导出、复制和原件弹层流程，并补充 LLM-derived displayName 完成态合同。 |
| 2.7 | 2026-07-07 | 清理修订记录中的状态说明，保留当前 D-17/D-20 平铺简历与绑定合同。 |
| 2.6 | 2026-07-06 | 修订记录改为当前布局表述。 |
| 2.2 | 2026-07-06 | 首页简历选择控件改为下拉框；最近模拟面试仅展示 3 条并通过“更多”进入模拟面试列表页。 |
| 2.1 | 2026-07-06 | 首页新建模拟面试规划快捷入口前移简历选择：JD 输入卡旁选择已有 ready 简历并保留“1 分钟创建”入口，主按钮为「立即面试」。 |
| 2.0 | 2026-06-12 | 按 D-17/D-20 重写：删除岗位推荐入口；简历从 `ResumeSource + ResumeVersion` 树形结构改为平铺 `Resume` 资产；绑定键 `resumeVersionId` 改为 `resumeId`；新增改写采纳覆盖 / 另存收口 |
| 1.6 | 2026-06-12 | 树形多简历管理基线。 |
