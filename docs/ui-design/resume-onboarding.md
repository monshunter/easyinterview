# 首次无简历用户引导流程

> **版本**: 1.4
> **状态**: active
> **更新日期**: 2026-05-02

## 1 文档目的

本文档定义用户首次使用且没有简历时的轻量引导流程。目标是让用户尽快形成可用于岗位推荐和模拟面试的第一版简历，而不是被迫填写冗长画像。

## 2 触发时机

```text
Home
  -> 还没有简历？1 分钟创建
  -> resume_versions(flow=create)
  -> Resume Intake

Mock Interview Plan
  -> 检测无绑定简历
  -> Resume Intake Prompt

Resume
  -> 新版本 / 导入新简历
  -> Resume Intake
```

简历引导不挡在首页之前。用户可以先输入 JD、看岗位推荐或浏览静态页面，再在需要个性化准备时补全简历。当前静态 UI 的目标入口是 `resume_versions` 内的 `flow=create`，不是旧 `onboarding` 路由；旧 `onboarding` 现在折回 `resume_versions`，历史 `screens-p0-complete.jsx::OnboardingScreen` 已清理。

## 3 两种引导路径

### 3.1 上传或粘贴简历

```text
Resume Intake
  -> Upload / Paste Resume
     ├─ 上传文件或粘贴原始文本
     ├─ 保存原始文件 / 原始文本
     ├─ 解析结构化简历
     ├─ 生成可识别名称
     ├─ 展示解析摘要
     ├─ 允许用户确认 / 修改关键字段
     └─ 生成 v1 主版本
```

系统必须同时保留：

- 原始文件或粘贴内容。
- 解析文本快照。
- 结构化分析结果。
- 简历名称、来源、时间、语言和模型版本。

### 3.2 轻量问答生成简历

```text
Resume Intake
  -> Guided Resume Q&A
     ├─ Q1: 你现在想准备什么类型的岗位
     ├─ Q2: 你最近一段工作在哪里，职位是什么
     ├─ Q3: 你主要做什么方向或负责什么模块
     ├─ Q4: 你最能代表能力的一个项目是什么
     └─ Q5: 你常用技术栈或核心技能是什么
  -> Generate Structured Resume Draft
  -> 用户确认
  -> 生成 v1 主版本
```

问答控制在 3-5 轮。它生成的是可继续完善的初始简历版本，不要求用户一次填完所有经历。

## 4 引导页面框架

```text
[Resume Intake]
├─ Header
│  ├─ 创建第一版简历
│  └─ 说明: 用于岗位推荐和模拟面试上下文
├─ Tabs
│  ├─ 上传文件
│  ├─ 粘贴内容
│  └─ 轻量问答
├─ Guided Flow
│  ├─ 左侧步骤
│  ├─ 当前问题
│  ├─ 文本回答
│  ├─ 上一步 / 下一步
│  └─ 生成 v1
├─ What Gets Saved
│  ├─ 原始版本
│  ├─ 结构化简历
│  └─ 版本基线
└─ After V1
   ├─ 岗位推荐
   └─ 开始面试
```

## 5 跳过策略

用户可以暂时跳过简历引导，但系统需要降低个性化承诺。

| 场景 | 行为 |
|------|------|
| 无简历但只看 JD | 允许继续 |
| 无简历打开岗位推荐 | 可展示通用或低置信结果，提示补简历会更准确 |
| 无简历开始模拟面试 | 允许但提示问题会更多依赖 JD，较少结合个人经历 |
| 无简历生成报告 | 报告只基于本场回答和 JD，不声称了解完整背景 |
| 用户稍后补简历 | 更新岗位推荐、面试规划和后续报告分析 |

## 6 数据落点

```text
ResumeSource
├─ sourceId
├─ name
├─ sourceType: upload / paste / guided
├─ originalContent
├─ originalFileRef
├─ parsedTextSnapshot
└─ createdAt

ResumeVersion
├─ resumeVersionId
├─ displayName
├─ parentSourceId
├─ versionType: structured_master
├─ structuredProfile
├─ model / prompt / parserVersion
└─ userConfirmedFields
```

轻量问答也生成 `ResumeSource`，其中 `originalContent` 保存问答记录。

## 7 登录与隐私

上传、粘贴和问答都会产生敏感个人资料，必须触发登录或明确的保存授权。

```text
Resume Intake
  -> Auth Gate(如未登录)
  -> Privacy Notice
  -> Save ResumeSource
  -> Save ResumeVersion
```

隐私提示应说明：

- 会保存原始简历或问答记录。
- 会生成结构化简历。
- 用户可以查看原件、导出和删除。

## 8 后续实现输入

1. 简历引导不应成为首页前置 onboarding。
2. 上传、粘贴和轻量问答必须落到同一套 `ResumeSource + ResumeVersion` 结构。
3. 原始内容、解析文本和结构化内容都必须保留。
4. 首页的 `还没有简历？1 分钟创建` 应进入简历创建流程。
5. 完整面试和报告应根据是否有简历调整个性化程度。
6. 旧 `onboarding` route 通过 `routeAliases` 折回 `resume_versions`，不作为当前目标入口；当前入口以 `resume_versions(flow=create)` 为准，新文档和新交互不得恢复旧 `OnboardingScreen` 的画像前置流程。
