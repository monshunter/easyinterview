# 首次无简历用户引导流程

> **版本**: 1.6
> **状态**: active
> **更新日期**: 2026-06-12

## 1 文档目的

本文档定义用户首次使用且没有简历时的轻量引导流程。目标是让用户尽快形成可用于模拟面试的第一份简历，而不是被迫填写冗长画像。

当前静态 UI 中，该流程由 `resume_versions(flow=create)` 进入，并在 `screen-resume-workshop.jsx` 内完成 `输入 -> Agent 解析 -> 预览确认 -> 保存`。轻量问答建档已按 product-scope D-20 删除；创建输入只有上传文件和粘贴文本两种。

## 2 触发时机

```text
Home
  -> 还没有简历？1 分钟创建
  -> resume_versions(flow=create)
  -> Resume Create Flow

Mock Interview Plan / Parse
  -> 检测无绑定简历
  -> Resume Intake Prompt

Resume
  -> 新建简历
  -> resume_versions(flow=create)
```

简历引导不挡在首页之前。用户可以先输入 JD 或浏览静态页面，再在需要个性化准备时补全简历。当前静态 UI 的目标入口是 `resume_versions` 内的 `flow=create`，不是旧 `onboarding` 路由；旧 `onboarding` 现在折回 `resume_versions`，历史 `screens-p0-complete.jsx::OnboardingScreen` 已清理。

## 3 两种输入路径

### 3.1 上传简历文件

```text
Resume Create Flow
  -> Upload
     ├─ 选择 PDF / DOCX / Markdown / TXT
     ├─ 记录 sourceLabel = 文件名
     ├─ 进入 Agent Parsing
     ├─ 展示结构化草稿预览
     └─ 用户确认后保存为一份新简历
```

原始文件会作为只读来源保存。解析出的工作经历、项目、技能和教育经历进入可编辑结构化简历。

### 3.2 粘贴简历文本

```text
Resume Create Flow
  -> Paste
     ├─ 粘贴原始文本
     ├─ 文本为空时禁用解析
     ├─ 记录 sourceLabel = 粘贴文本
     ├─ 进入 Agent Parsing
     ├─ 展示结构化草稿预览
     └─ 用户确认后保存为一份新简历
```

系统必须同时保留：

- 原始粘贴文本。
- 解析文本快照。
- 结构化分析结果。
- 简历名称、来源、时间、语言和模型版本。

## 4 解析与确认流程

```text
Submit Source
  -> ResumeParseFlow
     ├─ 提取原始文本
     ├─ 识别个人信息
     ├─ 解析工作经历
     ├─ 识别代表项目
     ├─ 聚合技能
     ├─ 提取教育背景
     └─ 生成结构化简历
  -> ResumePreviewConfirm
     ├─ 身份信息
     ├─ 个人简介
     ├─ 工作经历
     ├─ 代表项目
     ├─ 技能
     ├─ 教育经历
     ├─ 确认后保存什么
     └─ 确认并保存
```

保存必须发生在用户确认预览之后。解析阶段可以展示动态进度，但不应在用户未确认前把草稿当成正式简历。

## 5 引导页面框架

```text
[Resume Create Flow]
├─ Header
│  ├─ 创建第一份简历
│  └─ 说明: 保留原始来源并解析为结构化简历
├─ Input Tabs
│  ├─ 上传文件
│  └─ 粘贴内容
├─ Right Rail
│  ├─ 会保存什么
│  │  ├─ 原始来源
│  │  └─ 结构化简历
│  └─ 接下来
│     ├─ 动态解析原始内容
│     └─ 进入预览确认页
├─ Agent Parsing
│  └─ 逐步展示解析状态
└─ Preview Confirm
   ├─ 结构化草稿
   ├─ 保存内容说明
   ├─ 返回上一步
   └─ 确认并保存
```

## 6 跳过策略

用户可以暂时跳过简历引导，但系统需要降低个性化承诺。

| 场景 | 行为 |
|------|------|
| 无简历但只看 JD | 允许继续 |
| 无简历开始模拟面试 | 允许但提示问题会更多依赖 JD，较少结合个人经历 |
| 无简历生成报告 | 报告只基于本场回答和 JD，不声称了解完整背景 |
| 用户稍后补简历 | 更新面试规划和后续报告分析 |

## 7 数据落点

```text
Resume
├─ resumeId
├─ displayName
├─ sourceType: upload / paste
├─ sourceLabel
├─ originalContent
├─ originalFileRef
├─ parsedTextSnapshot
├─ structuredContent
├─ parseRunMetadata
├─ previewConfirmedAt
├─ model / prompt / parserVersion
├─ userConfirmedFields
└─ createdAt
```

## 8 登录与隐私

上传和粘贴都会产生敏感个人资料，必须触发登录或明确的保存授权。

```text
Resume Create Flow
  -> Auth Gate(如未登录)
  -> Privacy Notice
  -> Parse Source
  -> Preview Confirm
  -> Save Resume
```

隐私提示应说明：

- 会保存原始简历文件或原始文本。
- 会生成结构化简历。
- 用户可以查看原件、导出和删除。

## 9 后续实现输入

1. 简历引导不应成为首页前置 onboarding。
2. 上传和粘贴必须落到同一套平铺 `Resume` 结构。
3. 原始内容、解析文本和结构化内容都必须保留。
4. 首页的 `还没有简历？1 分钟创建` 应进入 `resume_versions(flow=create)`。
5. 完整面试和报告应根据是否有简历调整个性化程度。
6. 解析阶段应展示可解释进度，并允许用户取消回到输入。
7. 预览确认页必须让用户在正式保存前检查结构化草稿。
8. 旧 `onboarding` route 通过 `routeAliases` 折回 `resume_versions`，不作为当前目标入口；新文档和新交互不得恢复旧 `OnboardingScreen` 的画像前置流程，也不得恢复轻量问答建档。

## 10 修订记录

| 版本 | 日期 | 修订内容 |
|------|------|----------|
| 1.6 | 2026-06-12 | 按 product-scope D-20 删除轻量问答建档：输入路径收敛为上传 / 粘贴两种；数据落点改为平铺 `Resume` 结构；同步删除岗位推荐相关跳过策略行（D-17） |
| 1.5 | 2026-05-02 | 上传 / 粘贴 / 轻量问答三路径基线（已被 1.6 取代） |
