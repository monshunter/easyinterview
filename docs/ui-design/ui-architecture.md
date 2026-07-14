# EasyInterview UI 目标总体架构

> **版本**: 2.30
> **状态**: active
> **更新日期**: 2026-07-14

## 1 文档目的

本文档定义当前静态 UI 原型对应的目标信息架构。当前 UI 范围的核心入口为首页、面试和简历；真实面试复盘和用户画像不属于当前 UI 范围。

目标 UI 必须与 `ui-design/index.html` 和 `ui-design/src/app.jsx` 当前运行时交互一致。

## 2 已确认决策

1. App 默认进入首页；未登录状态由当前页面内的登录入口和业务前置登录处理。
2. 顶部导航为：`首页`、`面试`、`简历`。
3. 用户菜单为：`设置与隐私`、`退出登录`；未登录时只显示登录入口。
4. `复盘` 和 `用户画像` 不属于当前 UI 范围，不是一级导航、用户菜单入口、目标 route、静态原型页面或后续默认 workstream。
5. `debrief`、`debrief_full`、`profile` 等范围外 hash / route 输入在静态原型中归一到 `home`，不得 materialize 范围外页面。
6. `auth_profile_setup` 仍保留为首次登录资料补全页；这是账号资料补全，不是用户画像。
7. 报告内容只有 session-scoped Dashboard；允许从规划详情内容区进入 target-scoped ReportsScreen 索引当前轮次报告，但不加入 TopBar、不形成全局中心或第二种报告内容形态。报告后续开练动作只有 `复练当前轮` 与 `进入下一轮`。
8. 简历是一级模块：平铺列表、上传 / 粘贴创建、注册后直接详情、只读原始正文。
9. 当前只开放连续文本面试；电话入口置灰，不产生 `phone` / `voice` route state，通用 speech 基础设施留待后续重新评审。
10. 顶栏主题色、暗色模式、语言下拉和设置页字体预设是全局显示控制，不属于业务模块。
11. Desktop TopBar 保持 58px 单行节奏；`<=720px` 使用内容驱动的响应式换行，primary nav 独占下一行，`<=460px` 收起品牌文字并限制语言标签宽度。移动端页面内容必须从 TopBar 实际底部开始，所有控件与导航都留在 viewport 内，不允许用固定 58px 或横向页面溢出来伪造对齐。

## 3 目标产品骨架

```text
[EasyInterview App]
├─ TopBar
│  ├─ Brand: E mark + EasyInterview
│  ├─ Primary nav: 首页 / 面试 / 简历
│  ├─ Theme / dark / language
│  └─ User menu: 设置与隐私 / 退出登录
├─ Home / 首页
│  ├─ 粘贴 JD 输入框（唯一 JD intake）
│  ├─ 选择已有简历（适度宽度下拉框）
│  │  └─ 还没有简历？1 分钟创建（右侧同行）
│  ├─ 立即面试（简历选择下方）
│  └─ 最近模拟面试（最多 3 条 + 更多）
├─ Interview / 面试
│  ├─ 面试规划列表（一级入口默认 landing）
│  ├─ 面试规划详情 / 面试上下文确认（Parse 母版统一承接首次核对与回访；右上角进入当前规划报告）
│  ├─ JD / 简历 / InterviewRound
│  └─ 立即面试
├─ Interview Session
│  ├─ 全宽连续文本聊天
│  ├─ 电话入口置灰
│  ├─ 即时 user row + pending interviewer thinking
│  ├─ server-owned reply state + failed-row same-ID retry
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
   ├─ 设置与隐私
   └─ 退出登录
```

## 4 顶部导航

```text
[Top Navigation]
├─ 首页
├─ 面试
├─ 简历
├─ 主题色菜单
├─ 暗色模式
├─ 语言下拉
└─ 用户区
   ├─ 未登录: 登录
   └─ 已登录:
      ├─ 设置与隐私
      └─ 退出登录
```

顶部导航或用户菜单范围外能力：

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

- Desktop：TopBar 单行、58px 高、左右 32px padding。
- Mobile：TopBar 可按当前语言和已登录用户名称换行，左右 14px padding；primary nav 独占一行并可在自身容器内横向滚动，但不得扩大 document 宽度。
- 报告等带 App Shell 的页面从 TopBar 实际 `getBoundingClientRect().bottom` 开始；中英文或登录态引起的合法高度差不能用页面局部 offset 抹平。

## 5 目标模块关系

```text
Home
├─ 粘贴 JD（唯一文本输入）
├─ 选择已有 ready 简历（定宽下拉框）
│  └─ 还没有简历？1 分钟创建（同排）
│     └─ Resume Intake
│
├─ 立即面试（选择简历下方）
│  └─ Interview Plan Detail / Context Confirm(resumeId 已绑定)
│     └─ Interview Session 或保存后回到面试规划列表
├─ 最近模拟面试
│  ├─ 最多 3 条快捷卡片
│  └─ 更多 -> 面试规划列表

Interview / 面试
├─ 面试规划列表
│  ├─ 已有 TargetJob / JD 候选规划
│  ├─ 打开统一面试规划详情
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
├─ Back -> InterviewPlanDetail(targetJobId)
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

### 6.3 用户菜单和认证页面

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

1. 正式前端 TopBar 必须源级复刻当前三入口静态原型，包括 desktop 58px 单行与 mobile 响应式换行、无 document 横向溢出的状态。
2. `frontend` 不得注册 `debrief` / `profile` RouteName、primary nav、user menu 或 screen 分支。
3. OpenAPI、backend、migrations、shared、config 和 scenario 的复盘 / 用户画像范围收敛由 product-scope/001-core-loop-module-pruning 承接。
4. `auth_profile_setup` 保留为账号资料补全，不得写成用户画像。
5. Pixel parity 和 route tests 必须覆盖范围外入口负向断言。
6. Home 必须只保留 JD textarea、ready Resume 下拉框和主 CTA；正式前端与静态原型都不得保留其他 JD intake 控件或弹窗，desktop 1440px 与 mobile 390px 截图必须证明该单一路径。
7. Practice transient optimistic row 不得成为跨刷新事实源；`getPracticeSession` 必须恢复 user `clientMessageId/replyStatus`，pending/retryable/terminal/complete UI 由该服务端投影收敛。
8. `ReportsScreen(targetJobId)` 是受保护、chrome-visible 的上下文页面，入口仅在规划详情内容区右上角；TopBar 仍严格为三入口。Parse 不嵌入列表或保留 `section=reports`，Report/Generating trusted Back 返回 ReportsScreen。

## 9 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-14 | 2.30 | 增加 target-scoped ReportsScreen 上下文层级；Parse 仅保留内容区入口，TopBar 仍三入口，报告详情仍是唯一内容形态。 |
| 2026-07-13 | 2.29 | Practice conversation 增加服务端可恢复 reply state、即时消息/思考态与 failed-row same-ID retry 架构。 |
| 2026-07-13 | 2.28 | Home JD intake 收敛为唯一粘贴文本框，保留 ready 简历选择与主 CTA，并要求 desktop/mobile 截图验收。 |
| 2026-07-12 | 2.27 | 将正式前端既有的 mobile TopBar 两行/多行响应式、内容驱动高度和无 document 横向溢出规则回写为 UI 真理源，并要求带 App Shell 页面使用真实 TopBar 底部做绝对 viewport parity。 |
