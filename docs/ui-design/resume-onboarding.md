# 首次无简历用户引导流程

> **版本**: 1.9
> **状态**: active
> **更新日期**: 2026-07-07

## 1 文档目的

本文档定义用户首次使用且没有简历时的轻量引导流程。目标是让用户尽快形成可用于模拟面试的第一份简历，而不是被迫填写冗长画像。

当前静态 UI 中，该流程由 `resume_versions(flow=create)` 进入，并在 `screen-resume-workshop.jsx` 内完成 `输入 -> 注册 -> 直接打开只读详情`。创建输入只有上传文件和粘贴文本两种；轻量问答建档不属于当前流程。

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

简历引导不挡在首页之前。用户可以先输入 JD 或浏览静态页面，再在需要个性化准备时补全简历。当前静态 UI 的目标入口是 `resume_versions` 内的 `flow=create`；非当前 `onboarding` 路由折回 `resume_versions`。

## 3 两种输入路径

### 3.1 上传简历文件

```text
Resume Create Flow
  -> Upload
     ├─ 选择 PDF / DOCX / Markdown / TXT
     ├─ 记录 sourceLabel = 文件名
     └─ 注册成功后直接打开简历详情
```

原始文件会作为只读来源保存。详情页直接展示可读的原始文本 / 解析文本快照；后台解析出的结构化内容不再要求用户通过确认页保存。

### 3.2 粘贴简历文本

```text
Resume Create Flow
  -> Paste
     ├─ 粘贴原始文本
     ├─ 文本为空时禁用解析
     ├─ 根据原始文本派生临时标题
     └─ 注册成功后直接打开简历详情
```

系统必须同时保留：

- 原始粘贴文本。
- 解析文本快照。
- 结构化分析结果。
- 简历名称、来源、时间、语言和模型版本。

## 4 直接详情流程

```text
Submit Source
  -> registerResume
  -> resume_versions?resumeId=<id>
  -> Resume Detail
     └─ 只读原始简历正文
```

注册成功即创建简历资产。后台解析可以继续更新结构化字段和 LLM-derived `displayName`，但前端不展示解析动画页、结构化草稿确认页或确认保存页。

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
│  │  └─ 只读详情
│  └─ 接下来
│     └─ 直接打开简历详情
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
  -> Register Source
  -> Open Resume Detail
```

隐私提示应说明：

- 会保存原始简历文件或原始文本。
- 会生成结构化简历。
- 用户可以在详情页查看只读简历正文，原始来源快照由系统保留；详情页不提供导出或原件弹层。

## 9 后续实现输入

1. 简历引导不应成为首页前置 onboarding。
2. 上传和粘贴必须落到同一套平铺 `Resume` 结构。
3. 原始内容、解析文本和结构化内容都必须保留。
4. 首页的 `还没有简历？1 分钟创建` 应进入 `resume_versions(flow=create)`。
5. 完整面试和报告应根据是否有简历调整个性化程度。
6. 不展示解析阶段动画、解析失败重试页或结构化草稿确认页。
7. 注册成功后必须直接打开简历详情，详情页看到的就是简历原始内容本身。
8. 非当前 `onboarding` route 通过 `routeAliases` 折回 `resume_versions`，不作为当前目标入口；新文档和新交互不得引入画像前置流程或轻量问答建档。

## 10 修订记录

| 版本 | 日期 | 修订内容 |
|------|------|----------|
| 1.9 | 2026-07-07 | 未闭环回归修订：首次简历创建删除解析动画和预览确认页，注册后直接打开只读详情；粘贴路径派生临时标题并避免展示通用“粘贴的简历”。 |
| 1.8 | 2026-07-07 | 隐私提示对齐只读详情：保存原始来源快照，但详情页不再提供导出或原件弹层。 |
| 1.7 | 2026-07-06 | 当前入口和非当前 route 边界改为正向合同表述。 |
| 1.6 | 2026-06-12 | 输入路径收敛为上传 / 粘贴两种；数据落点改为平铺 `Resume` 结构。 |
| 1.5 | 2026-05-02 | 上传 / 粘贴 / 轻量问答三路径基线。 |
