# 首次无简历用户引导流程

> **版本**: 1.20
> **状态**: active
> **更新日期**: 2026-07-20

## 1 文档目的

本文档定义用户首次使用且没有简历时的轻量引导流程。目标是让用户尽快形成可用于模拟面试的第一份简历，而不是被迫填写冗长画像。

当前静态 UI 中，该流程由 `resume_versions(flow=create)` 进入，并在 `screen-resume-workshop.jsx` 内完成 `输入 -> 注册 -> 解析等待 -> 来源格式自适应只读详情 / 失败态`。创建输入只有上传文件和粘贴文本两种；轻量问答建档不属于当前流程。

## 2 触发时机

```text
Home
  -> 还没有简历？1 分钟创建
  -> resume_versions(flow=create)
  -> Resume Create Flow

Resume
  -> 新建简历
  -> resume_versions(flow=create)
```

简历引导不挡在首页浏览和 JD 录入之前，但 selectable Resume 是提交 JD、创建模拟面试规划以及进入训练/报告链路的强制前置。selectable 指未归档且 `parseStatus=ready` 或已有可读正文/结构化证据。用户没有该类简历时只能进入 `resume_versions` 内的 `flow=create`；创建并形成可读证据后返回 Home，由用户显式选择该简历再提交。范围外 `onboarding` 路由折回 `resume_versions`。

## 3 两种输入路径

### 3.1 上传简历文件

```text
Resume Create Flow
  -> Upload
     ├─ 拖放或选择一个 PDF / Markdown / TXT（默认最大 10MiB）
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
     ├─ parse_status=queued/processing: 等待动画（图标与内容几何位置固定，仅用透明度/光晕表达进行中状态）
     ├─ parse_status=ready: Source-adaptive Resume Detail
     └─ parse_status=failed: 失败态
```

注册成功即创建简历资产。后台解析继续更新结构化字段、LLM-derived `displayName` 和 Markdown 正文快照；前端展示等待态，但不展示结构化草稿确认页或确认保存页。等待动画必须保持图标容器、SVG、标题、说明和返回动作的几何位置稳定，不得通过循环 `scale` / `translate` 造成亚像素抖动；可使用透明度或不参与布局的柔和光晕表达进行中状态。解析前不得把原文第一行或文件名当作列表 / 详情名称。最终详情渲染由来源格式自动决定：PDF 用纵向页面栈，粘贴 / Markdown / TXT 用 Markdown 渲染；两类 renderer 共用同一阅读背景板，Markdown 正文页只渲染 body 本身。

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

Desktop 参考布局在全局 TopBar 下使用浅蓝背景和约 `1470px` 居中内容面。返回入口、eyebrow、36px 主标题、说明和输入卡共享左边界；右上角仅允许不承载业务事实的低对比文件插画。输入卡顶部是两项 tab rail，主体上传区高度约 370px，使用浅蓝虚线边框、标题/说明/主按钮和格式、大小、隐私三个能力标签形成克制的工具型层级，不再显示中央 72px 圆形上传 icon。Desktop 上传区必须真正支持拖放单个 PDF / Markdown / TXT，并在携带文件的 dragenter/dragover 时用边框、背景和“松开以上传”文案反馈；dragleave/drop 后复位。显式“选择文件”按钮继续作为键盘、触屏和不支持拖放环境的等价入口。多文件、格式错误、超限、runtime limit 未就绪或 submitting 状态必须复用当前本地 guard，失败时不得发出 presign/register。Paste 模式沿用同一外框和 tab 高度，不创建第二套页面骨架。`<=720px` 时装饰隐藏、标题与卡片单列、按钮可键盘操作且页面无横向溢出；移动端不依赖拖放完成上传。

## 6 强制前置策略

用户可以暂时不创建简历并继续浏览或编辑尚未提交的 JD，但不能跳过简历进入任何模拟面试业务链路。

| 场景 | 行为 |
|------|------|
| 无简历浏览 Home 或录入 JD 草稿 | 允许继续，但「立即面试」保持禁用且不调用 `importTargetJob` |
| 无简历提交 JD / 创建规划 | 不允许；进入 `resume_versions(flow=create)` 创建简历 |
| 无简历开始、复练或进入下一轮 | 不允许；不存在 JD-only 训练降级模式 |
| 无简历生成或查看异常规划报告 | 不允许；不存在仅基于回答和 JD 的报告降级模式 |
| 简历创建并形成可读证据 | 返回 Home，由用户显式选择后才能提交 JD；不自动绑定最近简历 |
| 历史 TargetJob 缺失或无效绑定 | 作为异常数据 fail closed；不在 Workspace 补绑，不从 route 或浏览器存储恢复 |

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
5. Home、Practice、Reports 和报告后动作都必须以已持久化的 selectable 简历绑定为前提，不实现无简历降级分支。
6. 不展示结构化草稿确认页。
7. 注册成功后必须进入解析等待态；成功后详情页按来源格式自动展示 PDF 页面栈或 Markdown 正文，失败后展示失败态。
8. 范围外 `onboarding` route 通过 `routeAliases` 折回 `resume_versions`，不作为当前目标入口；新文档和新交互不得引入画像前置流程或轻量问答建档。

## 10 修订记录

| 版本 | 日期 | 修订内容 |
|------|------|----------|
| 1.20 | 2026-07-20 | 删除上传区中央大图标并把视觉 dropzone 落为真实单文件拖放；保留选择文件等价入口、drag-active 反馈与失败零请求边界。 |
| 1.18 | 2026-07-18 | 锁定解析等待态的稳定性：后台轮询不得闪回通用 loading，图标与文案几何位置保持固定。 |
| 1.19 | 2026-07-19 | 按参考图锁定 CreateFlow 的全视口背景、1470px 内容面、大型输入卡、上传能力标签与移动端收敛。 |
| 1.17 | 2026-07-15 | 将 selectable 简历锁定为 JD 提交、模拟面试和报告链路的永久强制前置；删除无简历训练与降级报告承诺，历史缺绑统一 fail closed。 |
| 1.16 | 2026-07-14 | 将上传文件默认上限从历史 2MiB 修订为 RuntimeConfig 投影的 10MiB。 |
| 1.15 | 2026-07-10 | 将 onboarding route 负向边界统一为范围外口径；行为不变。 |
| 1.14 | 2026-07-08 | 修订详情阅读面：Markdown body 只渲染正文，不注入 displayName/header 元数据，并与 PDF 共用背景板。 |
| 1.13 | 2026-07-08 | 将 PDF 详情合同从原件预览改为无浏览器工具栏的纵向页面栈。 |
| 1.12 | 2026-07-07 | 修订上传格式与详情渲染：Resume 上传仅支持 PDF / Markdown / TXT；PDF 用 PDF 原件预览，粘贴 / Markdown / TXT 用 Markdown 渲染。 |
| 1.11 | 2026-07-07 | 本轮优化：创建后增加解析等待/失败态，详情渲染 Markdown；删除创建页右侧说明 rail，并同步 2MiB 上传上限。 |
| 1.10 | 2026-07-07 | 修订命名与上传原文展示回归：粘贴路径只记录中性来源标题，列表/详情名称等待 LLM-derived displayName；上传文件详情展示提取后的原文正文。 |
| 1.9 | 2026-07-07 | 未闭环回归修订：首次简历创建删除解析动画和预览确认页，注册后直接打开只读详情；粘贴路径派生临时标题并避免展示通用“粘贴的简历”。 |
| 1.8 | 2026-07-07 | 隐私提示对齐只读详情：保存原始来源快照，但详情页不再提供导出或原件弹层。 |
| 1.7 | 2026-07-06 | 当前入口和范围外 route 边界改为正向合同表述。 |
| 1.6 | 2026-06-12 | 输入路径收敛为上传 / 粘贴两种；数据落点改为平铺 `Resume` 结构。 |
| 1.5 | 2026-05-02 | 上传 / 粘贴 / 轻量问答三路径基线。 |
