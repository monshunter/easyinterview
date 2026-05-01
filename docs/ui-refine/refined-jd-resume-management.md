# 多 JD 与多简历目标管理结构

> **版本**: 1.3
> **状态**: active
> **更新日期**: 2026-05-01

## 1 文档目的

本文档定义当前 UI 中多份 JD、多份简历以及二者绑定关系的管理方式。JD 和简历都是一等公民；模拟面试规划负责把两者和面试轮次组合成一次模拟面试输入。

## 2 设计原则

1. 多 JD 以 `TargetJob` 管理，入口包括首页 JD 导入、岗位推荐和最近模拟面试。
2. 多简历以 `ResumeSource` 和 `ResumeVersion` 管理，入口在一级简历模块。
3. 每个模拟面试规划绑定一个 `TargetJob/JD`、一个 `ResumeVersion` 和一个 `InterviewRound`。
4. 岗位定制简历是从结构化主版本派生出的目标岗位版本。
5. 系统必须同时保留原始简历内容、解析文本和结构化分析结果。
6. 简历缺失时不阻断用户看岗位，但在需要保存、绑定或进入个性化模拟面试前触发补全和登录。
7. 模拟面试规划不是简历资产管理中心，只负责选择当前面试使用哪份简历。

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
│  ├─ ResumeSource(sourceId=upload-1)
│  │  ├─ name: 刘哲-前端简历.pdf
│  │  ├─ originalFileRef
│  │  ├─ parsedTextSnapshot
│  │  └─ sourceType: upload
│  └─ ResumeSource(sourceId=guided-1)
│     ├─ name: 轻量问答简历 v1
│     ├─ originalText: Q&A transcript
│     └─ sourceType: guided
├─ ResumeVersions
│  ├─ ResumeVersion(resumeId=master-v3)
│  │  ├─ parentSourceId: upload-1
│  │  ├─ versionType: structured_master
│  │  └─ structuredProfile
│  └─ ResumeVersion(resumeId=job-A-v1)
│     ├─ parentVersionId: master-v3
│     ├─ targetJobId: A
│     ├─ versionType: targeted
│     └─ tailoredContent
└─ MockInterviewPlans
   ├─ planId=P1 -> jobId=A + resumeId=job-A-v1 + round=技术一面
   └─ planId=P2 -> jobId=B + resumeId=master-v3 + round=HR 初筛
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
Resume
├─ 原始简历
├─ 结构化主版本
├─ 岗位定制版本
├─ 上传 / 粘贴新简历
├─ 轻量问答创建简历
├─ 预览原始文件 / 解析文本
└─ 查看岗位定制版本

Mock Interview Plan
└─ 绑定简历
   ├─ 当前绑定简历
   └─ 更换 -> Resume Picker Modal
```

### 5.2 简历版本类型

| 类型 | 来源 | 用途 |
|------|------|------|
| 原始简历 | 上传文件、粘贴文本或问答记录 | 作为用户提供的证据只读保留 |
| 结构化主版本 | 系统解析原始简历或问答生成 | 用于岗位推荐、面试生成和报告归因 |
| 岗位定制版本 | 基于某份简历和某个 JD 派生 | 只绑定到特定岗位或面试规划 |
| 轻量问答简历 | 首次无简历时 3-5 轮问答生成 | 作为可继续完善的初始版本 |

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

同一份 JD 可以和不同简历版本形成不同面试规划。同一份简历也可以被多个 JD 使用。

## 7 UI 行为

| 场景 | 目标行为 |
|------|----------|
| 用户新增 JD | 进入 `parse` 核对并确认 JD，再创建或确认 `TargetJob`，进入模拟面试规划 |
| 用户有多份 JD | 首页显示最近模拟面试，岗位推荐提供新的 JD 来源 |
| 用户不想继续当前规划 | 在模拟面试页点击切换规划或新建规划 |
| 用户首次无简历 | 首页或模拟面试规划提示创建简历 |
| 用户上传新简历 | 创建新的 `ResumeSource` 和结构化主版本 |
| 用户粘贴简历 | 创建新的 `ResumeSource`，保留粘贴文本 |
| 用户轻量问答生成简历 | 创建 `sourceType=guided` 的 `ResumeSource` 和 v1 主版本 |
| 用户更换面试绑定简历 | 在弹窗选择已有 `ResumeVersion` 并绑定 |
| 用户为岗位定制简历 | 创建带 `targetJobId` 的派生版本 |

## 8 不做的事

```text
不做:
├─ 独立经历库
├─ 要求用户先维护完整画像才能开始
├─ 多岗位信息混合在一个模拟面试规划里
├─ 让 Settings 承担简历管理职责
├─ 让更换简历直接跳出当前面试规划
└─ 覆盖原始简历内容
```

原始简历是用户提供的证据，任何结构化分析、改写或岗位定制都不能覆盖它。
