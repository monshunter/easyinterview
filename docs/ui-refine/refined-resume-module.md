# Resume 一级模块目标结构

> **版本**: 1.3
> **状态**: active
> **更新日期**: 2026-05-02

## 1 文档目的

本文档定义当前静态 UI 中简历作为一级模块的目标结构。简历和 JD 一样是一等公民：它们不是模拟面试规划页的附属信息，而是岗位推荐、模拟面试、报告分析和真实复盘的重要输入。

## 2 核心决策

1. 简历是一级模块，对应顶部导航 `简历`，当前运行时入口是 `resume_versions`。
2. 简历模块管理原始简历、结构化主版本和岗位定制版本。
3. 用户上传、粘贴或问答生成的简历都必须有可识别名称。
4. 系统必须保留原始文件或原始文本；解析、编辑和岗位定制不能覆盖原件。
5. 模拟面试规划页只选择当前面试使用哪份简历，不承担完整简历管理职责。
6. 更换绑定简历应在模拟面试规划页打开简历选择弹窗。

## 3 模块结构

```text
Resume
├─ Header
│  ├─ 简历工坊 · 版本
│  ├─ 新版本
│  └─ 导出 PDF
├─ Source Map
│  ├─ 原始简历
│  │  ├─ 文件名 / 来源名
│  │  ├─ 保留状态
│  │  └─ 预览原件
│  ├─ 结构化主版本
│  │  ├─ 解析成可编辑字段
│  │  └─ 不覆盖原始简历
│  └─ 岗位定制版本
│     ├─ 面向某个目标岗位
│     └─ 只保存采纳的改写
├─ 原始简历预览 Modal
│  ├─ 原始文件
│  ├─ 解析文本
│  └─ 来源关系
├─ Version Tabs
│  ├─ 主版本
│  └─ 各目标岗位定制版本
├─ Version Summary
│  ├─ 目标岗位
│  ├─ 原始来源
│  ├─ 改写 bullet
│  └─ 匹配度变化
└─ Rewrite Diff
   ├─ 原句
   ├─ 改写
   ├─ 为什么这么改
   ├─ 采纳
   ├─ 拒绝
   └─ 编辑
```

## 4 数据对象

```text
ResumeSource
├─ sourceId
├─ name
├─ sourceType: upload / paste / guided
├─ originalFileRef
├─ originalText
├─ parsedTextSnapshot
├─ createdAt
└─ retainedPolicy

ResumeVersion
├─ resumeVersionId
├─ displayName
├─ parentSourceId
├─ parentVersionId
├─ versionType: structured_master / targeted
├─ targetJobId
├─ structuredProfile
├─ rewriteSuggestions
├─ acceptedChanges
├─ createdAt
├─ updatedAt
├─ model / prompt / parserVersion
└─ userConfirmedFields
```

原始内容只读保存。岗位定制内容应创建派生版本，不能改写原始文件或主版本的证据来源。

## 5 与 JD / 模拟面试规划的关系

```text
MockInterviewPlan
├─ targetJobId
├─ jdId
├─ resumeVersionId
└─ roundId

Resume
├─ 管理全部 ResumeSource / ResumeVersion
└─ 提供 Resume Picker 数据
```

模拟面试输入至少包括：

- 当前目标岗位 / JD。
- 当前绑定简历。
- 当前面试轮次。
- 当前岗位历史模拟面试结果。
- 当前岗位真实复盘记录。

## 6 首次无简历入口

当用户没有任何简历时，首页和简历模块都可以触发同一套 `Resume Intake`。

```text
No Resume
├─ 还没有简历？1 分钟创建
├─ 上传 / 粘贴简历
└─ 轻量问答
   ├─ 目标岗位
   ├─ 最近工作 / 职位
   ├─ 主要方向
   ├─ 代表项目
   └─ 核心技能
```

## 7 一级导航入口

```text
[Top Navigation]
└─ 简历
   ├─ 查看简历版本
   ├─ 新建版本
   ├─ 预览原始简历
   ├─ 查看结构化主版本
   ├─ 查看岗位定制版本
   └─ 导出
```

`Settings` 只保留账号、隐私、导出和删除等账户级操作，不承担简历资产管理职责。

旧 `resume` route 仍可直达历史简历单页，`screens-rest.jsx::ResumeScreen` 也仍保留为历史代码，但不作为顶部导航或当前目标入口。新文档和新交互不得引用旧 `ResumeScreen` 作为目标结构，必须以 `resume_versions` 的版本工坊、创建流程和原始简历预览为准。

## 8 后续实现输入

1. 简历列表和简历选择弹窗都必须展示简历名称。
2. 原始简历预览必须能看到原始文件和解析文本快照。
3. 结构化主版本、岗位定制版本和改写建议必须可追溯到原始来源。
4. 模拟面试规划页更换简历时不应直接跳转到简历模块。
