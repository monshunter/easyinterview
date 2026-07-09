# Resume 一级模块目标结构

> **版本**: 3.3
> **状态**: active
> **更新日期**: 2026-07-09

## 1 文档目的

本文档定义当前静态 UI 中简历作为一级模块的目标结构。简历和 JD 一样是一等公民：它们不是模拟面试规划页的附属信息，而是模拟面试和报告分析的重要输入。

当前运行时入口是 `resume_versions`，目标实现以 `ui-design/src/screen-resume-workshop.jsx` 为准。按 product-scope D-20，简历资产是扁平 IA：不存在原始简历树、结构化主版本、岗位定制版本、分叉流程或版本管理；非当前画板或非当前组件不得作为树形实现依据。

## 2 核心决策

1. 简历是一级模块，对应顶部导航 `简历`，当前运行时入口是 `resume_versions`。
2. 简历是平铺列表中的独立资产：每份简历自带原始来源（文件或粘贴文本，只读）、解析文本快照和结构化内容，不区分原始简历 / 主版本 / 岗位定制版本（D-20）。
3. 简历首页是 `Resume Workshop` 平铺列表，无树形分组、无视图切换、无“选为底稿”。
4. 创建新简历经历 `上传 / 粘贴 -> 注册成功 -> 解析等待 -> 来源格式自适应详情 / 失败态`；不展示结构化草稿确认页或确认保存页；轻量问答建档不属于当前流程。
5. 简历详情页是只读正文页：上传 PDF 使用同源 source endpoint 渲染为从上到下平铺的 PDF 页面栈；粘贴、Markdown 文件和 TXT 文件使用 Markdown 渲染引擎展示正文；Markdown body 区只渲染简历正文本身，不额外注入 `displayName`、header 名称、summary 或来源元数据；PDF 与 Markdown 使用统一阅读背景板和白色 page surface；不提供导出、复制、编辑、预览标签、改写建议、结构化草稿确认、浏览器 PDF 工具栏或查看原始简历弹层。
6. 原件预览不是独立二级入口，而是详情正文区域根据来源格式自动选择的渲染方式，对用户透明。
7. 用户上传或粘贴生成的简历都必须有可识别名称；最终 `displayName` 优先由 backend parse 根据 LLM `displayName` / 结构化结果生成。若 LLM 输出失败但 backend 已抽取可读正文，backend 必须写入非通用 fallback 名称；不以“上传的简历”“粘贴的简历”、文件名或 raw resume 第一行作为完成态名称。
8. 系统必须保留原始文件或原始文本；上传 PDF / Markdown / text 的 prompt input 来自文件可读正文提取，粘贴和 Markdown/text 成功态详情正文来自 LLM 生成的 Markdown 快照，PDF 成功态详情正文使用原始 PDF 文件渲染出的纵向页面栈，解析结果不能覆盖原始来源快照。DOCX 不属于当前上传支持范围。
9. 用户最多可保留默认 10 份 active 简历，上传文件默认最大 2MiB；用户可以从列表删除简历以释放数量上限。
10. 模拟面试规划页和解析确认页只选择当前面试使用哪份简历，不承担完整简历管理职责。
11. 更换绑定简历应在对应页面打开简历选择弹窗，不跳转到简历模块。

## 3 模块结构

```text
Resume / resume_versions
├─ Resume Workshop List（平铺）
│  ├─ Header
│  │  ├─ 简历工坊
│  │  └─ 新建简历（唯一创建入口）
│  └─ 列表行
│     ├─ 简历名称 / 语言标签
│     ├─ 来源（上传文件 / 粘贴文本）
│     ├─ 最近编辑
│     ├─ 打开
│     └─ 删除
├─ Resume Create Flow
│  ├─ 上传文件
│  ├─ 粘贴内容
│  └─ 注册成功后进入解析等待
├─ Resume Parse Waiting(resumeId)
│  ├─ 解析中动画
│  └─ 成功后进入 Markdown 详情 / 失败后进入失败态
└─ Resume Detail(resumeId)
   └─ 只读简历正文（共享阅读背景板 + PDF 页面栈 / Markdown 页面）
      ├─ 简历名称（LLM-derived displayName）
      ├─ 来源 / 语言 / 最近更新时间
      └─ 原始 PDF / 原始文本 / 解析文本快照
```

## 4 数据对象

```text
Resume
├─ resumeId
├─ displayName
├─ sourceType: upload / paste
├─ originalFileRef
├─ originalText
├─ parsedTextSnapshot
├─ languageTag
├─ summary
├─ structuredContent
├─ createdAt
├─ updatedAt
├─ model / prompt / parserVersion
└─ userConfirmedFields
```

原始内容只读保存。详情页根据来源格式自动选择渲染方式：PDF 上传使用原始文件的同源 source endpoint 渲染纵向页面栈；粘贴、Markdown 文件和 TXT 文件优先展示 `parsedTextSnapshot` 形成的 Markdown 正文，`originalText` 只作为失败或历史数据兼容 fallback。PDF 与 Markdown 共用同一浅色阅读背景板，并在背景板内放置白色页面；Markdown 页面内不得额外 prepend `displayName`、详情 header 名称、summary 或来源元数据。它不提供结构化草稿确认、二次编辑、改写采纳、复制、导出、浏览器 PDF 工具栏或原件弹层。结构化内容只作为无原文时的降级兜底。

## 5 与 JD / 模拟面试规划的关系

```text
MockInterviewPlan
├─ targetJobId
├─ jdId
├─ resumeId
└─ roundId

Resume
├─ 管理全部平铺简历资产
└─ 提供 Resume Picker 数据
```

模拟面试输入至少包括：

- 当前目标岗位 / JD。
- 当前绑定简历。
- 当前面试轮次。
- 当前岗位过往模拟面试结果。

## 6 首次无简历入口

当用户没有任何简历时，首页和简历模块都可以触发同一套 `Resume Intake`。

```text
No Resume
├─ 还没有简历？1 分钟创建
├─ resume_versions(flow=create)
├─ 上传 / 粘贴
└─ 注册成功后进入解析等待，成功后打开来源格式自适应详情
```

## 7 一级导航入口

```text
[Top Navigation]
└─ 简历
   ├─ 查看简历工坊平铺列表
   ├─ 新建简历（上传 / 粘贴）
   ├─ 打开简历详情
   └─ 阅读只读简历正文（PDF 页面栈 / Markdown 渲染）
```

`Settings` 只保留账号、隐私、导出和删除等账户级操作，不承担简历资产管理职责。

非当前 `resume` / `experiences` / `star` / `onboarding` route 通过 `ROUTE_ALIASES` 折回 `resume_versions`；当前目标入口以 `screen-resume-workshop.jsx` 为准，非当前画板或非当前组件不得作为树形 / 分叉流程依据。

## 8 后续实现输入

1. 简历列表和简历选择弹窗都必须展示简历名称、来源和最近编辑时间。
2. 简历列表是单层平铺，不得出现树形分组、主版本 / 定制版本标签或“选为底稿”动作。
3. 简历详情必须只展示简历内容本身；PDF 上传用从上到下平铺的 PDF 页面栈，粘贴 / Markdown / TXT 用 Markdown 渲染；Markdown body 不得额外注入 `displayName`、header 名称、summary 或来源元数据；PDF 与 Markdown 必须使用统一阅读背景板；不得出现预览 / 改写 / 编辑 tabs、导出、复制、原件弹层、浏览器 PDF 工具栏或二次编辑动作。
4. 创建新简历时，不展示结构化预览确认页；注册成功后必须进入解析等待态，成功后打开来源格式自适应只读详情，失败后展示失败态。
5. parse 成功后的完成态名称必须优先使用 LLM 根据简历内容生成的 `displayName`；parse 失败但已有可读正文时必须使用 backend 生成的非通用 fallback 名称。通用上传 / 粘贴标题、文件名和 raw resume 第一行只能作为来源信息或正文内容，前端不得展示为列表或详情名称。
6. 创建面试规划前通过 Home 显式选择 ready 简历；解析确认页与模拟面试规划页不提供更换简历入口。
7. 简历列表不得出现重复创建 CTA；删除动作必须从 active 列表隐藏简历，并释放默认 10 份数量上限。

## 9 修订记录

| 版本 | 日期 | 修订内容 |
|------|------|----------|
| 3.2 | 2026-07-08 | 修订来源格式阅读面：Markdown body 禁止注入 displayName/header 元数据，并与 PDF 共用阅读背景板和白色页面。 |
| 3.1 | 2026-07-08 | 将 PDF 详情合同从原生 PDF 预览收敛为无工具栏的纵向页面栈。 |
| 3.0 | 2026-07-07 | 修订来源格式渲染合同：PDF 上传使用同源 PDF 原件预览，粘贴 / Markdown / TXT 使用 Markdown 渲染；DOCX 退出当前 Resume 上传支持范围。 |
| 2.9 | 2026-07-07 | 本轮优化：新增解析等待/失败态、Markdown 正文渲染、删除简历、默认 10 份数量上限与 2MiB 上传上限；删除重复创建 CTA。 |
| 2.8 | 2026-07-07 | 修订失败态命名与轮询合同：LLM 输出失败但已有可读正文时 backend 写入非通用 fallback 名称，详情不因 failed-with-snapshot 长期停留在占位。 |
| 2.7 | 2026-07-07 | 修订命名与上传原文展示回归：禁止从 raw 第一行或文件名派生可见名称，上传 PDF / Markdown / text 详情正文必须来自文件可读正文提取。 |
| 2.6 | 2026-07-07 | 未闭环回归修订：删除创建流解析动画和预览确认页，注册成功后直接打开只读详情；详情优先展示原始内容快照，前端不得展示通用上传/粘贴名称。 |
| 2.5 | 2026-07-07 | 简历详情收敛为只读正文，移除预览 / 改写 / 编辑 tabs、导出、复制、原件弹层与二次编辑入口；补充 LLM-derived displayName 完成态合同。 |
| 2.4 | 2026-07-07 | 清理修订记录中的状态说明，保留当前 D-20 平铺简历合同。 |
| 2.2 | 2026-07-06 | 将当前目标结构改为非当前范围边界表述，避免用非当前组件说明解释当前简历模块行为。 |
| 2.1 | 2026-07-06 | 将当前目标结构改为非当前范围边界表述。 |
| 2.0 | 2026-06-12 | 按 product-scope D-20 扁平化重写：列表平铺；改写建议仅采纳；新增采纳后覆盖 / 另存收口；当前目标不包含原始简历树、主版本、岗位定制版本、分叉流程或轻量问答 |
| 1.7 | 2026-05-02 | 树形简历工坊基线。 |
