# Resume 一级模块目标结构

> **版本**: 4.2
> **状态**: active
> **更新日期**: 2026-07-19

## 1 文档目的

本文档定义简历作为一级模块的目标结构。简历和 JD 一样是一等公民：它们不是模拟面试规划页的附属信息，而是模拟面试和报告分析的重要输入。

当前运行时入口是 `resume_versions`，目标实现由本文档与对应 frontend spec/plan 承接。按 product-scope D-20，简历资产是扁平 IA：不存在原始简历树、结构化主版本、岗位定制版本、分叉流程或版本管理；范围外旧设计或组件不得作为树形实现依据。

## 2 核心决策

1. 简历是一级模块，对应顶部导航 `简历`，当前运行时入口是 `resume_versions`。
2. 简历是平铺列表中的独立资产：每份简历自带原始来源（文件或粘贴文本，只读）、解析文本快照和结构化内容，不区分原始简历 / 主版本 / 岗位定制版本（D-20）。
3. 简历首页是 `Resume Workshop` 平铺卡片列表，无树形分组、无视图切换、无“选为底稿”；桌面端每行排列两张等宽卡，移动端使用同序单列并占满可用宽度。
4. 创建新简历经历 `上传 / 粘贴 -> 注册成功 -> 解析等待 -> 来源格式自适应详情 / 失败态`；不展示结构化草稿确认页或确认保存页；轻量问答建档不属于当前流程。
5. 简历详情页是只读正文页：上传 PDF 使用同源 source endpoint 渲染为从上到下平铺的 PDF 页面栈；粘贴、Markdown 文件和 TXT 文件使用 Markdown 渲染引擎展示正文；Markdown body 区只渲染简历正文本身，不额外注入 `displayName`、header 名称、summary 或来源元数据；PDF 与 Markdown 使用统一阅读背景板和白色 page surface。Desktop `1916×821` 下详情内容面约 `1512px`，Back、蓝色 eyebrow、名称 kicker、主标题和 meta 共享左边界；阅读背景板约 `1310px`，其中白色 PDF/Markdown 纸张约 `1150px` 并居中。详情不提供导出、复制、编辑、预览标签、改写建议、结构化草稿确认、浏览器 PDF 工具栏或查看原始简历弹层。
6. 原件预览不是独立二级入口，而是详情正文区域根据来源格式自动选择的渲染方式，对用户透明。
7. 用户上传或粘贴生成的简历都必须有可识别名称；最终 `displayName` 优先由 backend parse 根据 LLM `displayName` / 结构化结果生成。若 LLM 输出失败但 backend 已抽取可读正文，backend 必须写入非通用 fallback 名称；不以“上传的简历”“粘贴的简历”、文件名或 raw resume 第一行作为完成态名称。
8. 系统必须保留原始文件或原始文本；上传 PDF / Markdown / text 的 prompt input 来自文件可读正文提取，粘贴和 Markdown/text 成功态详情正文来自 LLM 生成的 Markdown 快照，PDF 成功态详情正文使用原始 PDF 文件渲染出的纵向页面栈，解析结果不能覆盖原始来源快照。DOCX 不属于当前上传支持范围。
9. 用户最多可保留默认 10 份 active 简历，上传文件默认最大 10MiB；用户可以从列表删除简历以释放数量上限。
10. 模拟面试规划页和解析确认页只选择当前面试使用哪份简历，不承担完整简历管理职责。
11. 新建规划时选择或更换绑定简历仍由对应创建流程承接；既有 Workspace 只读详情中的“绑定简历”是查看入口，点击后使用已保存的 `resumeId` 跳转到对应简历详情，不提供 in-place rebind。

## 3 模块结构

```text
Resume / resume_versions
├─ Resume Workshop List（平铺）
│  ├─ Header
│  │  ├─ 简历工坊
│  │  └─ 新建简历（唯一创建入口）
│  └─ Resume Cards
│     ├─ 简历名称 / 摘要 / 语言标签
│     ├─ 来源（上传文件 / 粘贴文本）/ 最近编辑
│     ├─ Footer: 打开
│     └─ 右上角: 删除
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

原始内容只读保存。详情页根据来源格式自动选择渲染方式：PDF 上传使用原始文件的同源 source endpoint 渲染纵向页面栈；粘贴、Markdown 文件和 TXT 文件优先展示 `parsedTextSnapshot` 形成的 Markdown 正文，`parsedTextSnapshot` 为空时才回退到 `originalText`。PDF 与 Markdown 共用同一浅色阅读背景板，并在背景板内放置白色页面；Markdown 页面内不得额外 prepend `displayName`、详情 header 名称、summary 或来源元数据。它不提供结构化草稿确认、二次编辑、改写采纳、复制、导出、浏览器 PDF 工具栏或原件弹层。结构化内容只作为无原文时的降级兜底。

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

简历列表 Header 的唯一“新建简历”按钮必须使用与 Workspace “新建面试规划”一致的 22px 圆圈加号 SVG、线宽和图标间距，不得退回裸 `+` 文本或缩小图标。

范围外 `resume` / `experiences` / `star` / `onboarding` route 通过 `ROUTE_ALIASES` 折回 `resume_versions`；当前目标入口以 `screen-resume-workshop.jsx` 为准，范围外画板或范围外组件不得作为树形 / 分叉流程依据。

## 8 后续实现输入

- `queued/processing` 详情使用共享 `AsyncTransitionScene` 的 resume variant：TopBar 保持“简历”高亮，中心文件插画由多层轨道和单一运动点围绕，标题/说明位于同一轴线；返回控件统一显示“返回 / Back”，目标仍为简历工坊。
- 后台 poll 必须持续保留这一 DOM；首次尚无 Resume 数据的通用 loading 与解析等待态分离。循环轨道只改变不参与布局的 opacity/transform，并在 reduced-motion 下停用；desktop/mobile 均不得因动画造成 geometry 抖动或 document 横向溢出。

1. 简历卡片列表和简历选择弹窗都必须展示简历名称、来源和最近编辑时间；列表卡片可展示 closed `ResumeSummary.summaryHeadline`，但参考稿列表不重复展示语言 tag，且不得读取详情正文。
2. 简历列表是单层卡片列表，不得出现表头、表格行、树形分组、主版本 / 定制版本标签或“选为底稿”动作；desktop 页面内容区约 1408px，每行排列两张等宽卡；mobile 保持同一 DOM/阅读顺序并占满可用宽度。卡片 header 使用文件 icon + 名称/摘要，右上角是独立删除动作；meta 与 footer 由规则线分隔，footer 只在右侧保留“打开”。
3. 简历详情必须只展示简历内容本身；Header 使用 Back、蓝色 eyebrow、名称 kicker、主标题与来源/日期 meta 的参考稿层级；desktop 内容面约 `1512px`，阅读背景板约 `1310px`，内部白色纸张约 `1150px` 且居中；PDF 上传用从上到下平铺的 PDF 页面栈，粘贴 / Markdown / TXT 用 Markdown 渲染；Markdown body 不得额外注入 `displayName`、header 名称、summary 或来源元数据；PDF 与 Markdown 必须使用统一阅读背景板；mobile 收敛为满宽背景板和可读纸张，且不得出现预览 / 改写 / 编辑 tabs、导出、复制、原件弹层、浏览器 PDF 工具栏或二次编辑动作。
4. 创建新简历时，不展示结构化预览确认页；注册成功后必须进入解析等待态，成功后打开来源格式自适应只读详情，失败后展示失败态。
5. parse 成功后的完成态名称必须优先使用 LLM 根据简历内容生成的 `displayName`；parse 失败但已有可读正文时必须使用 backend 生成的非通用 fallback 名称。通用上传 / 粘贴标题、文件名和 raw resume 第一行只能作为来源信息或正文内容，前端不得展示为列表或详情名称。
6. 创建面试规划前通过 Home 显式选择 ready 简历；解析确认页与模拟面试规划页不提供更换简历入口。
7. 简历列表不得出现重复创建 CTA；每张卡片底部保留明确“打开”动作，删除图标固定在右上角；删除成功必须从 active 列表隐藏简历并释放默认 10 份数量上限，失败时保留原卡片并展示可恢复错误。

## 9 修订记录

| 版本 | 日期 | 修订内容 |
|------|------|----------|
| 4.2 | 2026-07-20 | 将简历详情与解析等待态的返回控件统一为“返回 / Back”，保持简历工坊目标不变。 |
| 4.1 | 2026-07-19 | 按解析简历参考稿统一等待场景：共享蓝白画布、中心文件轨道插画、稳定几何、返回入口与 reduced-motion；保持后台轮询无 loading 闪现。 |
| 4.0 | 2026-07-19 | 按简历预览参考稿补齐详情 Header 与 `1512/1310/1150px` 内容面、阅读背景板和白色纸张构图；保持来源格式 renderer、只读行为和数据合同不变。 |
| 3.9 | 2026-07-19 | 简历 Header 创建入口改用与 Workspace 一致的 22px 圆圈加号，保持创建 route 与双列卡片布局不变。 |
| 3.8 | 2026-07-19 | 按提供的简历列表参考稿锁定标题区、文件 icon、meta 分隔、删除与 footer 打开层级，并根据用户补充将 desktop 明确为每行两张等宽卡；数据与路由合同不变。 |
| 3.7 | 2026-07-15 | 将 Resume Workshop 列表从表格行修订为与面试规划一致的响应式卡片网格：桌面固定最大列宽、移动单列，并保留打开/删除及 closed ResumeSummary 边界。 |
| 3.6 | 2026-07-14 | 将上传文件默认上限从历史 2MiB 修订为 RuntimeConfig 投影的 10MiB。 |
| 3.5 | 2026-07-10 | 将 route / 画板 / 组件负向边界统一为范围外口径；行为不变。 |
| 3.4 | 2026-07-10 | 收敛解析前命名口径：使用中性 fallback 名称与来源信息，删除旧命名描述。 |
| 3.2 | 2026-07-08 | 修订来源格式阅读面：Markdown body 禁止注入 displayName/header 元数据，并与 PDF 共用阅读背景板和白色页面。 |
| 3.1 | 2026-07-08 | 将 PDF 详情合同从原生 PDF 预览收敛为无工具栏的纵向页面栈。 |
| 3.0 | 2026-07-07 | 修订来源格式渲染合同：PDF 上传使用同源 PDF 原件预览，粘贴 / Markdown / TXT 使用 Markdown 渲染；DOCX 退出当前 Resume 上传支持范围。 |
| 2.9 | 2026-07-07 | 本轮优化：新增解析等待/失败态、Markdown 正文渲染、删除简历、默认 10 份数量上限与 2MiB 上传上限；删除重复创建 CTA。 |
| 2.8 | 2026-07-07 | 修订失败态命名与轮询合同：LLM 输出失败但已有可读正文时 backend 写入非通用 fallback 名称，详情不因 failed-with-snapshot 长期停留在解析前名称。 |
| 2.7 | 2026-07-07 | 修订命名与上传原文展示回归：禁止从 raw 第一行或文件名派生可见名称，上传 PDF / Markdown / text 详情正文必须来自文件可读正文提取。 |
| 2.6 | 2026-07-07 | 未闭环回归修订：删除创建流解析动画和预览确认页，注册成功后直接打开只读详情；详情优先展示原始内容快照，前端不得展示通用上传/粘贴名称。 |
| 2.5 | 2026-07-07 | 简历详情收敛为只读正文，移除预览 / 改写 / 编辑 tabs、导出、复制、原件弹层与二次编辑入口；补充 LLM-derived displayName 完成态合同。 |
| 2.4 | 2026-07-07 | 清理修订记录中的状态说明，保留当前 D-20 平铺简历合同。 |
| 2.2 | 2026-07-06 | 将当前目标结构改为范围外边界表述，避免用范围外组件说明解释当前简历模块行为。 |
| 2.1 | 2026-07-06 | 将当前目标结构改为范围外边界表述。 |
| 2.0 | 2026-06-12 | 按 product-scope D-20 扁平化重写：列表平铺；改写建议仅采纳；新增采纳后覆盖 / 另存收口；当前目标不包含原始简历树、主版本、岗位定制版本、分叉流程或轻量问答 |
| 1.7 | 2026-05-02 | 树形简历工坊基线。 |
