# Resume 一级模块目标结构

> **版本**: 1.7
> **状态**: active
> **更新日期**: 2026-05-02

## 1 文档目的

本文档定义当前静态 UI 中简历作为一级模块的目标结构。简历和 JD 一样是一等公民：它们不是模拟面试规划页的附属信息，而是岗位推荐、模拟面试、报告分析和复盘的重要输入。

当前运行时入口仍是 `resume_versions`，但目标实现已经从旧的单页版本工坊切换为 `ui-design/src/screen-resume-workshop.jsx`：它在 `screens-p1-depth.jsx` 之后加载并覆盖 `window.ResumeVersionsScreen`。旧 `screens-p1-depth.jsx::ResumeVersionsScreen` 只保留为 `_LegacyResumeVersionsScreen` dead code，不再驱动目标文档、导航或用户流程。

## 2 核心决策

1. 简历是一级模块，对应顶部导航 `简历`，当前运行时入口是 `resume_versions`。
2. 简历模块管理原始简历、结构化主版本和岗位定制版本。
3. 简历首页是 `Resume Workshop` 列表视图，默认按原始简历成树分组，也可切换为按版本平铺。
4. 每份原始简历是一棵独立树：原始文件 / 文本作为只读来源，解析得到主版本，再从主版本分叉出岗位定制版本。
5. 原始简历支持在用与归档状态；归档不等于删除，仍可回溯来源。
6. 用户可以从一棵原始简历树选择底稿并创建新的岗位定制版本。
7. 版本详情页包含 `预览`、`改写建议`、`手动编辑` 三个标签；主版本不显示可用改写建议，避免破坏基线。
8. 打开岗位定制版本时默认进入 `改写建议`；打开主版本时默认进入 `预览`。
9. 详情页必须能从当前版本回看原始简历，弹层展示来源关系、原始文件视图和解析文本快照。
10. `导出 PDF`、`复制纯文本`、`复制为新版本`、`保存改动` 和 `创建版本` 都必须有可见反馈；静态原型中的创建动作应在当前列表态体现。
11. `拒绝 / 编辑 / 采纳` 只作用于当前岗位定制版本，不改变原始简历、主版本或其它兄弟版本。
12. 用户上传、粘贴或问答生成的简历都必须有可识别名称。
13. 创建新原始简历必须经历 `输入 -> Agent 解析 -> 预览确认 -> 保存 v1`，不能在未确认前写入正式版本。
14. 系统必须保留原始文件或原始文本；解析、编辑和岗位定制不能覆盖原件。
15. 模拟面试规划页只选择当前面试使用哪份简历，不承担完整简历管理职责。
16. 更换绑定简历应在模拟面试规划页打开简历选择弹窗。

## 3 模块结构

```text
Resume / resume_versions
├─ Resume Workshop List
│  ├─ Header
│  │  ├─ 简历工坊
│  │  └─ 新建简历
│  ├─ Stats Strip
│  │  ├─ 原始简历 active / total
│  │  ├─ 全部版本
│  │  ├─ 最高匹配
│  │  └─ 最近编辑
│  ├─ View Switcher
│  │  ├─ 按原始分组
│  │  └─ 按版本平铺
│  ├─ Original Tree View
│  │  ├─ 原始简历行
│  │  │  ├─ 文件名 / 来源名
│  │  │  ├─ 语言标签
│  │  │  ├─ 来源类型 / 创建时间 / 摘要
│  │  │  ├─ active / archived
│  │  │  └─ 选为底稿
│  │  ├─ 主版本
│  │  ├─ 岗位定制版本
│  │  └─ 上传另一份原始简历
│  └─ Flat Version View
│     ├─ 版本
│     ├─ 来源原始
│     ├─ 目标岗位
│     ├─ 匹配分
│     ├─ 最近编辑
│     └─ 打开（岗位定制默认进入改写建议）
├─ Resume Create Flow
│  ├─ 上传文件
│  ├─ 粘贴内容
│  ├─ 轻量问答
│  ├─ Agent 解析中
│  └─ 预览确认并保存 v1 -> 加入当前列表
├─ Resume Branch Flow
│  ├─ 底稿来源: 原始简历 + 主版本
│  ├─ 版本名称
│  ├─ 目标岗位 / 公司
│  ├─ 侧重方向
│  ├─ Bullet 初始化方式
│  │  ├─ 从主版本复制
│  │  ├─ 空白起步
│  │  └─ AI 选 bullet
│  └─ 创建岗位定制版本 -> 返回列表并展示新版本
└─ Resume Detail(versionId)
   ├─ Breadcrumb: 原始简历 > 版本 > 当前版本
   ├─ Resume Branch Map
   │  ├─ 原始简历
   │  ├─ 主版本
   │  └─ 当前版本
   ├─ Preview Tab
   │  ├─ 只读简历预览
   │  ├─ 导出 PDF
   │  ├─ 复制纯文本
   │  └─ 查看原件弹层: 来源关系 / 原始文件 / 解析文本
   ├─ Rewrites Tab(仅岗位定制版本，打开定制版本的默认标签)
   │  ├─ 建议改写列表
   │  ├─ 原句 / AI 改写 diff
   │  ├─ 为什么这么改
   │  ├─ 拒绝
   │  ├─ 编辑
   │  └─ 采纳
   └─ Edit Tab
      ├─ 一句话标题
      ├─ 简介
      ├─ 工作经历
      ├─ 技能
      └─ 保存改动
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
├─ languageTag
├─ summary
├─ status: active / archived
├─ createdAt
└─ retainedPolicy

ResumeVersion
├─ resumeVersionId
├─ displayName
├─ originalId
├─ parentSourceId
├─ parentVersionId
├─ versionType: structured_master / targeted
├─ targetJobId
├─ targetLabel
├─ focusAngle
├─ seedStrategy: copy_master / blank / ai_select
├─ structuredProfile
├─ rewriteSuggestions
├─ acceptedChanges
├─ rejectedChanges
├─ manualEdits
├─ matchScore
├─ bulletCount
├─ createdAt
├─ updatedAt
├─ model / prompt / parserVersion
└─ userConfirmedFields
```

原始内容只读保存。岗位定制内容应创建派生版本，不能改写原始文件、原始文本或主版本的证据来源。

## 5 与 JD / 模拟面试规划的关系

```text
MockInterviewPlan
├─ targetJobId
├─ jdId
├─ resumeVersionId
└─ roundId

Resume
├─ 管理全部 ResumeSource / ResumeVersion
├─ 维护原始简历树和岗位定制分叉
└─ 提供 Resume Picker 数据
```

模拟面试输入至少包括：

- 当前目标岗位 / JD。
- 当前绑定简历。
- 当前面试轮次。
- 当前岗位历史模拟面试结果。
- 当前岗位复盘记录。

## 6 首次无简历入口

当用户没有任何简历时，首页和简历模块都可以触发同一套 `Resume Intake`。

```text
No Resume
├─ 还没有简历？1 分钟创建
├─ resume_versions(flow=create)
├─ 上传 / 粘贴 / 轻量问答
├─ Agent 解析原始内容
├─ 预览结构化草稿
└─ 用户确认后保存 v1 主版本
```

轻量问答控制在 3-5 轮，当前静态稿为 5 个问题：最近经历、主要方向、代表项目、量化结果、目标岗位。它生成的是可继续完善的初始简历版本，不要求用户一次填完所有经历。

## 7 一级导航入口

```text
[Top Navigation]
└─ 简历
   ├─ 查看简历工坊
   ├─ 按原始简历树管理版本
   ├─ 按版本平铺挑选简历
   ├─ 新建原始简历
   ├─ 从已有树分叉岗位定制版本
   ├─ 打开版本详情
   ├─ 预览 / 改写建议 / 手动编辑
   ├─ 查看原始文件
   ├─ 导出 PDF
   └─ 复制纯文本
```

`Settings` 只保留账号、隐私、导出和删除等账户级操作，不承担简历资产管理职责。

旧 `resume` route 现在通过 `routeAliases` 折回 `resume_versions`，历史 `screens-rest.jsx::ResumeScreen` 已清理。旧 `screens-p1-depth.jsx::ResumeVersionsScreen` 也不再作为运行时目标：它被重命名保留为 `_LegacyResumeVersionsScreen`，当前目标入口以 `screen-resume-workshop.jsx` 导出的 `ResumeWorkshopScreen` 为准。

## 8 后续实现输入

1. 简历列表和简历选择弹窗都必须展示简历名称。
2. 原始简历树必须清楚展示原始来源、主版本和岗位定制版本之间的关系。
3. 原始简历预览必须能看到来源关系、原始文件和解析文本快照。
4. 结构化主版本、岗位定制版本和改写建议必须可追溯到原始来源。
5. 主版本应保持干净：可手动编辑结构化字段，但不得把岗位定制改写写回主版本。
6. 岗位定制版本的 `拒绝 / 编辑 / 采纳` 只作用于当前版本。
7. 创建新原始简历时，必须先展示解析进度和结构化预览，用户确认后才保存为 v1。
8. 打开岗位定制版本时，应直接落到 `改写建议` 决策面；打开主版本时落到 `预览`。
9. 从已有树分叉岗位定制版本时，必须记录目标岗位 / 公司、侧重方向和 bullet 初始化方式，并在创建后返回列表展示新版本。
10. 导出、复制、复制为新版本、保存字段和创建版本等按钮不得是无反馈空动作；静态原型至少要给出 toast 或当前会话内列表变化。
11. 模拟面试规划页更换简历时不应直接跳转到简历模块。
