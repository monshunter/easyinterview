# EasyInterview UI 目标总体架构

> **版本**: 2.37
> **状态**: active
> **更新日期**: 2026-07-19

## 1 文档目的

本文档定义当前目标信息架构。当前 UI 范围的核心入口为首页、面试和简历；真实面试复盘和用户画像不属于当前 UI 范围。

目标 UI 由本文档、模块 UI 文档和对应产品 spec 共同约束，并由正式 `frontend/` 直接实施。

## 2 已确认决策

1. App 默认进入首页；未登录状态由当前页面内的登录入口和业务前置登录处理。
2. 顶部导航为：`首页`、`面试`、`简历`。
3. 未登录时 TopBar 显示登录入口；已登录时账号区只显示一个直接进入 `settings` 的圆形 `E` initial-mark 按钮。它是视觉化设置入口，不是用户头像，不显示账号 chip 或 dropdown；退出登录位于设置页。
4. `复盘` 和 `用户画像` 不属于当前 UI 范围，不是一级导航、账号设置入口、目标 route、正式页面或后续默认 workstream。
5. `debrief`、`debrief_full`、`profile` 等范围外 route 输入归一到 `home`，不得 materialize 范围外页面。
6. `auth_profile_setup` 仍保留为首次登录资料补全页；这是账号资料补全，不是用户画像。
7. 报告内容只有 session-scoped Dashboard；允许从规划详情内容区进入 target-scoped ReportsScreen 索引当前轮次报告，但不加入 TopBar、不形成全局中心或第二种报告内容形态。报告后续开练动作只有 `复练当前轮` 与 `进入下一轮`。
8. 简历是一级模块：平铺列表、上传 / 粘贴创建、注册后直接详情、只读原始正文。
9. 当前只开放连续文本面试；电话入口置灰，不产生 `phone` / `voice` route state，通用 speech 基础设施留待后续重新评审。
10. TopBar 只保留暗色模式、语言下拉和设置入口；主题色移入“设置”页的“外观”区并保存为账号级偏好。产品字体采用固定默认栈，不提供字体预设。
11. Desktop TopBar 使用与 Home 参考图一致的 76px 单行节奏；`<=720px` 使用内容驱动的响应式换行，primary nav 独占下一行，`<=460px` 收起品牌文字并限制语言标签宽度。移动端页面内容必须从 TopBar 实际底部开始，所有控件与导航都留在 viewport 内，不允许用固定 desktop 高度或横向页面溢出来伪造对齐。
12. “设置 > 外观”的 Ocean / Plum / Custom 是始终可见的一级主题选择器；custom accent picker 是仅在选择 Custom 后于一级下方展开的二级编辑器，只保留色相与饱和度两个调整维度。二级编辑器不得覆盖或替换一级选择器；色相轨道必须展示完整光谱，彩度轨道必须以当前色相从低彩到高彩渐变，不增加 preview/value 区或“恢复主题默认色 / Reset to theme accent”按钮。选择 Ocean 或 Plum 是退出自定义色的唯一清晰路径。调整只做本地预览，点击保存才发送一次账号更新请求。
13. `/workspace` 是无参规划列表，`/workspace?targetJobId=...` 是统一只读规划详情；ready 卡片直接进入详情。`/parse?targetJobId=...` 只承接新导入 queued/processing 命令进度，ready 后 replace 到 Workspace 详情。
14. Practice 的 persisted user/assistant text 通过 `react-markdown + remark-gfm` 安全投影；`skipHtml`、no `rehypeRaw`、no remote image、safe link，send/retry 仍使用原始 text/clientMessageId。

## 3 目标产品骨架

```text
[EasyInterview App]
├─ TopBar
│  ├─ Brand: E mark + EasyInterview
│  ├─ Primary nav: 首页 / 面试 / 简历
│  ├─ Dark / language
│  └─ Account: 已登录圆形 E 设置入口 / 未登录登录入口
├─ Home / 首页
│  ├─ 粘贴 JD 输入框（唯一 JD intake）
│  ├─ 选择已有简历（适度宽度下拉框）
│  │  └─ 还没有简历？1 分钟创建（右侧同行）
│  ├─ 立即面试（简历选择下方）
│  └─ 最近模拟面试（最多 3 条全宽记录 + 查看全部）
├─ Interview / 面试
│  ├─ 面试规划列表（一级入口默认 landing）
│  ├─ 面试规划详情 / 面试上下文确认（Workspace targetJobId 只读母版；右上角进入当前规划报告）
│  ├─ Parse 命令进度（只承接新导入 queued/processing，ready 后 replace 到 Workspace 详情）
│  ├─ JD / 简历 / InterviewRound
│  └─ 立即面试
├─ Interview Session
│  ├─ Global App TopBar
│  ├─ Practice Session Header（公司 / 岗位 / 角色 / 计时 / 暂停 / 电话 / 结束）
│  ├─ 全宽连续文本聊天
│  ├─ 电话入口置灰
│  ├─ 即时 user row + pending interviewer thinking
│  ├─ server-owned reply state + failed-row same-ID retry
│  ├─ user/assistant safe Markdown/GFM projection
│  └─ 结束并生成报告
├─ Reports / 当前规划报告
│  ├─ 当前 TargetJob canonical rounds
│  ├─ 每轮 current report + latest generation state
│  └─ 不展示其他规划或完整历史
├─ Report Dashboard
│  ├─ 会话 / 岗位 / 简历 / 轮次上下文
│  ├─ 准备度、维度、证据详情、下一步
│  ├─ 复练当前轮
│  └─ 进入下一轮
├─ Resume / 简历
│  ├─ 平铺简历列表
│  ├─ 上传 / 粘贴创建
│  ├─ 注册成功后直接打开详情
│  └─ 详情：只读原始简历正文
└─ Settings / Auth
   ├─ 邮箱验证码登录
   ├─ 首次账号资料补全
   └─ 设置（外观账号级主题 / 账号真实字段 / 退出 / 导出不可用 / 删除账号）
```

## 4 顶部导航

```text
[Top Navigation]
├─ 首页
├─ 面试
├─ 简历
├─ 暗色模式
├─ 语言下拉
└─ 用户区
   ├─ 未登录: 登录
   └─ 已登录: 圆形 E initial-mark 设置按钮 -> settings
```

顶部导航或设置入口范围外能力：

- `复盘 / Debrief`
- `用户画像 / User Profile`
- `岗位推荐 / Job Picks`
- `当前岗位`
- `面试报告`
- `成长`
- `经历库`
- `单题 Drill`
- `独立 Voice`

响应式约束：

- Desktop：TopBar 单行、76px 高、左右 24px padding；品牌、导航与右侧 controls 按参考图保持同一垂直中心线。
- Mobile：TopBar 可按当前语言和已登录设置按钮换行，左右 14px padding；primary nav 独占一行并可在自身容器内横向滚动，但不得扩大 document 宽度。
- 报告等带 App Shell 的页面从 TopBar 实际 `getBoundingClientRect().bottom` 开始；中英文或登录态引起的合法高度差不能用页面局部 offset 抹平。
- Practice 也属于带 App Shell 的页面：全局 TopBar 下方再渲染 Practice Session Header；不得把会话控制栏冒充全局导航，也不得因进入或切换 Practice route 重新读取 `/me`。

## 5 目标模块关系

```text
Home
├─ 粘贴 JD（唯一文本输入）
├─ 选择已有 ready 简历（定宽下拉框）
│  └─ 还没有简历？1 分钟创建（同排）
│     └─ Resume Intake
│
├─ 立即面试（选择简历下方）
│  └─ Parse queued/processing(targetJobId)
│     └─ ready replace -> Workspace Interview Plan Detail(targetJobId, resume 已绑定)
│        └─ Interview Session
├─ 最近模拟面试
│  ├─ 最多 3 条快捷卡片
│  └─ 查看全部 -> 面试规划列表

Interview / 面试
├─ 面试规划列表
│  ├─ 已有 TargetJob / JD 候选规划
│  ├─ 卡片直达 Workspace 统一只读面试规划详情
│  └─ 从新 JD 创建规划 -> Home

Mock Interview Plan
├─ TargetJob / JD
├─ Resume
├─ InterviewRound
├─ InterviewSession
├─ ReportsScreen(targetJobId)
└─ ReportDashboard

ReportDashboard
├─ 准备度、维度、证据详情和下一步
├─ 复练当前轮 -> Interview Session(same round)
└─ 进入下一轮 -> Interview Session(next round)

ReportsScreen(targetJobId)
├─ Back -> Workspace InterviewPlanDetail(targetJobId)
├─ current report -> ReportDashboard(reportId)
└─ latest generating -> ReportGenerating(reportId)
```

## 6 页面层级规则

### 6.1 一级页面

```text
Home
MockInterviewPlan
Resume
```

### 6.2 上下文 / 会话级页面

```text
InterviewPlanDetail / ContextConfirm
InterviewSession(sessionId)
ReportsScreen(targetJobId)
ReportGenerating(reportId)
ReportDashboard(reportId)
```

### 6.3 账号设置和认证页面

```text
SettingsPrivacy
AuthLogin
AuthVerify
AuthProfileSetup
AuthLogout
```

## 7 范围外 route 输入归一

```text
ROUTE_ALIASES
├─ welcome -> home
├─ growth -> home
├─ plan -> workspace
├─ mistakes -> report
├─ drill -> practice
├─ followup -> practice
├─ experiences -> resume_versions
├─ star -> resume_versions
├─ resume -> resume_versions
├─ onboarding -> resume_versions
├─ auth_register -> auth_login
├─ auth_reset -> auth_login
├─ jd_match -> home
├─ debrief -> home
├─ debrief_full -> home
└─ profile -> home
```

`voice` 不保留 route alias。判断目标架构时以 `normalizeRoute` 后的 `activeRouteName` 和实际渲染内容为准，不以范围外 hash、范围外画板标签或范围外组件为准。

## 8 后续实现输入

1. 正式前端 TopBar 必须实现当前三入口设计，包括 desktop 76px 单行与 mobile 响应式换行、无 document 横向溢出的状态。
2. `frontend` 不得注册 `debrief` / `profile` RouteName、primary nav、user menu 或 screen 分支。
3. OpenAPI、backend、migrations、shared、config 和 scenario 的复盘 / 用户画像范围收敛由 product-scope/001-core-loop-module-pruning 承接。
4. `auth_profile_setup` 保留为账号资料补全，不得写成用户画像。
5. Route 与 component tests 必须覆盖范围外入口负向断言。
6. Home 必须只保留 JD textarea、ready Resume 下拉框和主 CTA；正式前端不得保留其他 JD intake 控件或弹窗，desktop 1440px 与 mobile 390px responsive/browser smoke 必须证明该单一路径。
7. Practice transient optimistic row 不得成为跨刷新事实源；`getPracticeSession` 必须恢复 user `clientMessageId/replyStatus`，pending/retryable/terminal/complete UI 由该服务端投影收敛。
8. `ReportsScreen(targetJobId)` 是受保护、chrome-visible 的上下文页面，入口仅在 Workspace 规划详情内容区右上角；TopBar 仍严格为三入口。Parse 不渲染 ready 详情、不嵌入列表或保留 `section=reports`，Reports Back 返回 Workspace detail，Report/Generating trusted Back 返回 ReportsScreen。
9. `CustomAccentPicker` 只允许 hue/saturation DOM 与更新回调。Ocean / Plum / Custom 一级选择器始终可见；custom editor 只在 Custom 激活时挂载于同列下方，且不得与一级选择器共享覆盖式 grid area。component/source negative gate 必须证明 preview/value/reset 区、“恢复主题默认色 / Reset to theme accent”双语文案及仅为旧 UI 服务的 `onClear` / `active` props 零引用；选择 Ocean / Plum 恢复预定义主题并隐藏二级编辑器。
10. Settings Appearance 的 1440 desktop 与 390 mobile tests 必须覆盖 Ocean / Plum / custom accent 草稿预览、一级常驻/二级按需展开、保存、DOM、computed style、viewport containment 与必要 screenshot smoke；TopBar 删除旧主题区域后不得留下空白占位或横向溢出。
11. Route/component gate 必须证明 query-free Workspace 列表、targetJobId Workspace 详情和 Parse command-progress 三态互斥；ready 卡片详情执行一次同 key `getTargetJob`，不得 import、poll、播放 Parse animation 或在 route side 启动 session。
12. Practice message renderer 必须同时覆盖 user/assistant GFM、raw HTML/remote image/unsafe URI 负向、安全 link、exact raw same-ID retry，以及 390px pre/code/table 局部滚动且 document 无横向溢出。
13. TopBar 已登录态只渲染圆形 `E` initial-mark 设置按钮；它不得读取或暗示用户头像数据。component/responsive/a11y gate 必须证明姓名、caret、backdrop、dropdown 与 TopBar logout 零引用，且 desktop/mobile 点击区域和 focus ring 可用。
14. Settings 为无 tab 单页：Account 只读展示 runtime `/me.displayName` / 完整 `email` 并进入既有 logout 确认；Privacy 只展示导出暂不可用与账号删除。完整 email 只用于 authenticated 页面显示，不写入日志/场景证据。删除流程覆盖确认、pending、失败重试；`202` 后调用现有 `refreshAuth()` 重探测 `/me`（预期 401），提交 unauthenticated 状态并 replace Home；不得重复实现清 session 方法、挂载时重复调用 `/me`、保留 `emailMasked` alias 或伪静态字段。
15. 字体固定为 Noto Serif SC（标题）、Inter（正文）与 JetBrains Mono（标签/代码）；删除其它 font preset 数据、包、CSS imports、locale 文案和兼容状态。

## 9 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-19 | 2.37 | 锁定 Settings Appearance 两层主题结构：一级预定义/Custom 选择器常驻，二级 hue/chroma 仅在 Custom 激活时于下方展开；hue 使用完整光谱，chroma 使用当前色相的低彩到高彩渐变。 |
| 2026-07-19 | 2.36 | Home 参考图成为当前首页视觉方向：desktop TopBar 调整为 76px，已登录设置入口使用单一圆形 E initial mark，不恢复账号菜单。 |
| 2026-07-19 | 2.35 | 修正设置账号主题验收 owner：desktop/mobile 主题交互由 Settings Appearance 承接，TopBar 不再承接 Theme menu。 |
| 2026-07-19 | 2.34 | 将主题色移入“设置 > 外观”并按账号保存；TopBar 保留暗色/语言/设置齿轮；Practice 恢复全局 App TopBar，并区分会话控制栏。 |
| 2026-07-15 | 2.33 | 采用设置简化方案 A：TopBar 已登录账号区收敛为设置齿轮；设置页改为无 tab 的真实账号/隐私单页；字体收敛为固定默认栈。 |
| 2026-07-14 | 2.32 | 将 Workspace 拆为无参列表与 targetJobId 只读详情，Parse 收窄为新导入命令进度；Reports/terminal 返回详情，并加入 Practice 安全 Markdown/GFM 投影边界。 |
| 2026-07-14 | 2.31 | 将 CustomAccentPicker 收敛为色相与饱和度，并以 Ocean / Plum 作为退出自定义色的唯一清晰路径。 |
| 2026-07-14 | 2.30 | 增加 target-scoped ReportsScreen 上下文层级；Parse 仅保留内容区入口，TopBar 仍三入口，报告详情仍是唯一内容形态。 |
| 2026-07-13 | 2.29 | Practice conversation 增加服务端可恢复 reply state、即时消息/思考态与 failed-row same-ID retry 架构。 |
| 2026-07-13 | 2.28 | Home JD intake 收敛为唯一粘贴文本框，保留 ready 简历选择与主 CTA，并要求 desktop/mobile 截图验收。 |
| 2026-07-12 | 2.27 | 将正式前端既有的 mobile TopBar 两行/多行响应式、内容驱动高度和无 document 横向溢出规则写入 UI 设计，并要求带 App Shell 页面使用真实 TopBar 底部做绝对 viewport 验证。 |
