# 多 JD 与多简历目标管理结构

> **版本**: 1.5
> **状态**: active
> **更新日期**: 2026-05-02

## 1 文档目的

本文档定义当前 UI 中多份 JD、多份简历以及二者绑定关系的管理方式。JD 和简历都是一等公民；模拟面试规划负责把两者和面试轮次组合成一次模拟面试输入。

## 2 设计原则

1. 多 JD 以 `TargetJob` 管理，入口包括首页 JD 导入、岗位推荐和最近模拟面试。
2. 多简历以 `ResumeSource` 和 `ResumeVersion` 管理，入口在一级简历模块。
3. 每个模拟面试规划绑定一个 `TargetJob/JD`、一个 `ResumeVersion` 和一个 `InterviewRound`。
4. 岗位定制简历是从某棵原始简历树的结构化主版本派生出的目标岗位版本。
5. 每份原始简历是一棵独立树，树内包含原始来源、主版本和多个岗位定制版本。
6. 系统必须同时保留原始简历内容、解析文本和结构化分析结果。
7. 简历缺失时不阻断用户看岗位，但在需要保存、绑定或进入个性化模拟面试前触发补全和登录。
8. 模拟面试规划不是简历资产管理中心，只负责选择当前面试使用哪份简历。

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
├─ ResumeSources
│  ├─ ResumeSource(sourceId=src-cn)
│  │  ├─ name: 刘哲_前端_中文_2026.pdf
│  │  ├─ sourceType: upload
│  │  ├─ languageTag: 中
│  │  ├─ status: active
│  │  ├─ originalFileRef
│  │  ├─ parsedTextSnapshot
│  │  └─ summary
│  ├─ ResumeSource(sourceId=src-en)
│  │  ├─ name: LiuZhe_Frontend_EN_2026.pdf
│  │  ├─ sourceType: upload
│  │  ├─ languageTag: EN
│  │  └─ status: active
│  └─ ResumeSource(sourceId=src-fullstack)
│     ├─ name: 刘哲_全栈方向_2024.pdf
│     └─ status: archived
├─ ResumeVersions
│  ├─ ResumeVersion(resumeId=v-cn-master)
│  │  ├─ originalId: src-cn
│  │  ├─ versionType: structured_master
│  │  └─ structuredProfile
│  ├─ ResumeVersion(resumeId=v-cn-bd)
│  │  ├─ originalId: src-cn
│  │  ├─ parentVersionId: v-cn-master
│  │  ├─ targetJobId: A
│  │  ├─ versionType: targeted
│  │  ├─ focusAngle
│  │  ├─ seedStrategy
│  │  └─ acceptedChanges
│  └─ ResumeVersion(resumeId=v-en-stripe)
│     ├─ originalId: src-en
│     ├─ parentVersionId: v-en-master
│     └─ versionType: targeted
└─ MockInterviewPlans
   ├─ planId=P1 -> jobId=A + resumeId=v-cn-bd + round=技术一面
   └─ planId=P2 -> jobId=B + resumeId=v-en-master + round=HR 初筛
```

## 4 多 JD 管理

### 4.1 首页

首页不是岗位资产管理页，而是开始或继续面试的入口。

```text
Home
├─ Start New Interview
│  ├─ 粘贴 JD
│  ├─ 上传 JD 文件
│  └─ URL 导入
└─ Recent Mock Interviews
   ├─ 岗位名称 / 公司
   ├─ 状态
   ├─ 面试轮次节点
   └─ Open Mock Interview Plan
```

### 4.2 岗位推荐

```text
Job Picks
├─ 推荐 JD 列表
├─ 匹配原因
├─ 风险提示
└─ 选择 JD -> 确认面试 -> JD 解析确认 -> Mock Interview Plan
```

岗位推荐是一级模块，用于发现新的目标 JD。

### 4.3 模拟面试规划切换

```text
Mock Interview Plan Header
├─ 当前面试规划
├─ 切换规划
└─ 新建规划
```

切换规划时，页面上下文必须完整切换，包括 JD、简历绑定、当前轮次、历史会话和报告。

## 5 多简历管理

### 5.1 简历版本入口

```text
Resume / resume_versions
├─ Resume Workshop List
│  ├─ 按原始分组
│  │  ├─ Original Resume Tree
│  │  ├─ Master Version
│  │  └─ Targeted Versions
│  └─ 按版本平铺
│     ├─ Version
│     ├─ From Original
│     ├─ Target
│     ├─ Match
│     └─ Last Edit
├─ 新建原始简历
│  ├─ 上传文件
│  ├─ 粘贴文本
│  ├─ 轻量问答
│  ├─ Agent 解析
│  └─ 预览确认保存 v1 -> 加入当前列表
├─ 基于这棵树新建版本
│  ├─ 版本名称
│  ├─ 目标岗位 / 公司
│  ├─ 侧重方向
│  ├─ Bullet 初始化方式
│  └─ 创建后返回列表并展示新版本
└─ Version Detail
   ├─ 预览: 原件预览 / 导出 PDF / 复制纯文本
   ├─ 改写建议: 定制版本默认进入
   └─ 手动编辑: 保存改动

Mock Interview Plan
└─ 绑定简历
   ├─ 当前绑定简历
   └─ 更换 -> Resume Picker Modal
```

### 5.2 简历版本类型

| 类型 | 来源 | 用途 |
|------|------|------|
| 原始简历 | 上传文件、粘贴文本或问答记录 | 作为用户提供的证据只读保留，是一棵简历树的根 |
| 结构化主版本 | 系统解析原始简历或问答生成 | 用于岗位推荐、面试生成和报告归因，也是定制版本的默认分叉基线 |
| 岗位定制版本 | 基于某份原始简历树和目标 JD 派生 | 只绑定到特定岗位或面试规划，改写采纳只作用于当前版本 |
| 轻量问答简历 | 首次无简历时 3-5 轮问答生成 | 作为可继续完善的初始版本 |

### 5.3 列表与详情切换

```text
Resume Workshop
├─ Group by original
│  ├─ 管理底稿与分叉关系
│  ├─ 选择某棵树作为新版本底稿
│  └─ 查看每棵树下的主版本和定制版本
└─ Flat by version
   ├─ 按匹配分和更新时间排序
   ├─ 快速挑选可投递或可绑定版本
   └─ 打开版本详情

Version Detail(versionId)
├─ Resume Branch Map
├─ Preview: 原件预览 / 导出 PDF / 复制纯文本
├─ Rewrites(定制版本默认进入)
└─ Edit
```

主版本可以进入预览和手动编辑，但 `改写建议` 应禁用或说明“主版本保持干净”。岗位定制版本中的拒绝、编辑、采纳不影响主版本或兄弟版本。用户从列表打开岗位定制版本时默认进入 `改写建议`；从列表打开主版本时默认进入 `预览`。

### 5.4 岗位定制分叉

```text
Original Tree
  -> 选为底稿
  -> 基于这棵树新建版本
     ├─ 版本名称
     ├─ 目标岗位 / 公司
     ├─ 侧重方向
     │  ├─ 前端平台 / 基建方向
     │  ├─ 协作影响力
     │  ├─ 全栈广度
     │  ├─ 技术 Lead 视角
     │  └─ 自定义
     └─ Bullet 初始化方式
        ├─ 从主版本复制
        ├─ 空白起步
        └─ AI 选 bullet
```

定制分叉必须明确来源树、主版本、目标岗位和初始化策略。只填写版本名称但没有目标岗位时不能创建。

## 6 JD 与简历绑定

```text
MockInterviewPlan
├─ planId
├─ targetJobId
├─ jdId
├─ resumeVersionId
├─ roundId
└─ source: home / job_picks / report_replay / report_next_round
```

同一份 JD 可以和不同简历版本形成不同面试规划。同一份简历也可以被多个 JD 使用。同一棵原始简历树可以派生出多个面向不同 JD 的定制版本。

## 7 UI 行为

| 场景 | 目标行为 |
|------|----------|
| 用户新增 JD | 进入 `parse` 核对并确认 JD，再创建或确认 `TargetJob`，进入模拟面试规划 |
| 用户有多份 JD | 首页显示最近模拟面试，岗位推荐提供新的 JD 来源 |
| 用户不想继续当前规划 | 在模拟面试页点击切换规划或新建规划 |
| 用户首次无简历 | 首页或模拟面试规划提示创建简历 |
| 用户上传新简历 | 创建新的 `ResumeSource`，解析后经预览确认保存结构化主版本 |
| 用户粘贴简历 | 创建新的 `ResumeSource`，保留粘贴文本，并经预览确认保存 v1 |
| 用户轻量问答生成简历 | 创建 `sourceType=guided` 的 `ResumeSource` 和 v1 主版本 |
| 用户查看简历资产 | 默认按原始简历树查看，也可切换为版本平铺 |
| 用户为岗位定制简历 | 从某棵树分叉，填写目标岗位 / 公司、侧重方向和 bullet 初始化方式 |
| 用户打开岗位定制版本 | 直接进入 `改写建议` 决策面；主版本仍进入 `预览` |
| 用户查看原件 | 在版本详情页打开只读弹层，展示来源关系、原始文件视图和解析文本快照 |
| 用户采纳改写建议 | 只写入当前岗位定制版本，不影响原始简历、主版本或其它定制版本 |
| 用户导出、复制或保存简历 | 给出可见反馈；创建或复制版本时应在当前列表态体现结果 |
| 用户更换面试绑定简历 | 在弹窗选择已有 `ResumeVersion` 并绑定 |

## 8 不做的事

```text
不做:
├─ 独立经历库
├─ 要求用户先维护完整画像才能开始
├─ 多岗位信息混合在一个模拟面试规划里
├─ 让 Settings 承担简历管理职责
├─ 让更换简历直接跳出当前面试规划
├─ 覆盖原始简历内容
├─ 把岗位定制改写写回原始简历或主版本
└─ 把旧 ResumeVersionsScreen 当成目标运行时页面
```

原始简历是用户提供的证据，任何结构化分析、改写或岗位定制都不能覆盖它。
