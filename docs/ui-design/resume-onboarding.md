# 首次无简历用户引导流程

> **版本**: 1.14
> **状态**: active
> **更新日期**: 2026-07-08

## 1 文档目的

本文档定义用户首次使用且没有简历时的轻量引导流程。目标是让用户尽快形成可用于模拟面试的第一份简历，而不是被迫填写冗长画像。

当前静态 UI 中，该流程由 `resume_versions(flow=create)` 进入，并在 `screen-resume-workshop.jsx` 内完成 `输入 -> 注册 -> 解析等待 -> 来源格式自适应只读详情 / 失败态`。创建输入只有上传文件和粘贴文本两种；轻量问答建档不属于当前流程。

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
     ├─ 选择 PDF / Markdown / TXT（默认最大 2MiB）
     ├─ 记录 sourceLabel = 文件名（仅作为来源信息）
     └─ 注册成功后进入解析等待
```

原始文件会作为只读来源保存。PDF 详情页使用原始 PDF 文件渲染出的纵向页面栈，不展示浏览器 PDF 工具栏；Markdown / TXT 详情页展示提取后的 Markdown 正文 / 解析文本快照，且 Markdown body 不额外注入 `displayName`、header 名称、summary 或来源元数据；PDF 与 Markdown 使用统一阅读背景板和白色页面；文件名不得作为简历内容或完成态名称。DOCX 不属于当前上传支持范围。

### 3.2 粘贴简历文本

```text
Resume Create Flow
  -> Paste
     ├─ 粘贴原始文本
     ├─ 文本为空时禁用解析
     ├─ 记录中性来源标题
     └─ 注册成功后进入解析等待
```

系统必须同时保留：

- 原始粘贴文本。
- 解析文本快照。
- 结构化分析结果。
- 简历名称、来源、时间、语言和模型版本。

## 4 解析等待与详情流程

```text
Submit Source
  -> registerResume
  -> resume_versions?resumeId=<id>
  -> Resume Parse Waiting
     ├─ parse_status=queued/processing: 等待动画
     ├─ parse_status=ready: Source-adaptive Resume Detail
     └─ parse_status=failed: 失败态
```

注册成功即创建简历资产。后台解析继续更新结构化字段、LLM-derived `displayName` 和 Markdown 正文快照；前端展示等待态，但不展示结构化草稿确认页或确认保存页；解析前不得把原文第一行或文件名当作列表 / 详情名称。最终详情渲染由来源格式自动决定：PDF 用纵向页面栈，粘贴 / Markdown / TXT 用 Markdown 渲染；两类 renderer 共用同一阅读背景板，Markdown 正文页只渲染 body 本身。

## 5 引导页面框架

```text
[Resume Create Flow]
├─ Header
│  ├─ 创建第一份简历
│  └─ 说明: 保留原始来源并解析为结构化简历
├─ Input Tabs
│  ├─ 上传文件
│  └─ 粘贴内容
└─ Inline Error / Progress
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
- 用户可以在详情页查看只读简历正文，PDF 来源在正文区域以页面栈展示原始 PDF，粘贴 / Markdown / TXT 来源用 Markdown 渲染；两者共用阅读背景板，Markdown body 不额外展示 `displayName` 或来源元数据；原始来源快照由系统保留；详情页不提供导出、浏览器 PDF 工具栏或原件弹层。

## 9 后续实现输入

1. 简历引导不应成为首页前置 onboarding。
2. 上传和粘贴必须落到同一套平铺 `Resume` 结构。
3. 原始内容、解析文本和结构化内容都必须保留。
4. 首页的 `还没有简历？1 分钟创建` 应进入 `resume_versions(flow=create)`。
5. 完整面试和报告应根据是否有简历调整个性化程度。
6. 不展示结构化草稿确认页。
7. 注册成功后必须进入解析等待态；成功后详情页按来源格式自动展示 PDF 页面栈或 Markdown 正文，失败后展示失败态。
8. 非当前 `onboarding` route 通过 `routeAliases` 折回 `resume_versions`，不作为当前目标入口；新文档和新交互不得引入画像前置流程或轻量问答建档。

## 10 修订记录

| 版本 | 日期 | 修订内容 |
|------|------|----------|
| 1.14 | 2026-07-08 | 修订详情阅读面：Markdown body 只渲染正文，不注入 displayName/header 元数据，并与 PDF 共用背景板。 |
| 1.13 | 2026-07-08 | 将 PDF 详情合同从原件预览改为无浏览器工具栏的纵向页面栈。 |
| 1.12 | 2026-07-07 | 修订上传格式与详情渲染：Resume 上传仅支持 PDF / Markdown / TXT；PDF 用 PDF 原件预览，粘贴 / Markdown / TXT 用 Markdown 渲染。 |
| 1.11 | 2026-07-07 | 本轮优化：创建后增加解析等待/失败态，详情渲染 Markdown；删除创建页右侧说明 rail，并同步 2MiB 上传上限。 |
| 1.10 | 2026-07-07 | 修订命名与上传原文展示回归：粘贴路径只记录中性来源标题，列表/详情名称等待 LLM-derived displayName；上传文件详情展示提取后的原文正文。 |
| 1.9 | 2026-07-07 | 未闭环回归修订：首次简历创建删除解析动画和预览确认页，注册后直接打开只读详情；粘贴路径派生临时标题并避免展示通用“粘贴的简历”。 |
| 1.8 | 2026-07-07 | 隐私提示对齐只读详情：保存原始来源快照，但详情页不再提供导出或原件弹层。 |
| 1.7 | 2026-07-06 | 当前入口和非当前 route 边界改为正向合同表述。 |
| 1.6 | 2026-06-12 | 输入路径收敛为上传 / 粘贴两种；数据落点改为平铺 `Resume` 结构。 |
| 1.5 | 2026-05-02 | 上传 / 粘贴 / 轻量问答三路径基线。 |
