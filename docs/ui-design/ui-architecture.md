# EasyInterview UI 目标总体架构

> **版本**: 2.21
> **状态**: active
> **更新日期**: 2026-07-07

## 1 文档目的

本文档定义当前静态 UI 原型对应的目标信息架构。当前 UI 范围的核心入口为首页、模拟面试和简历；真实面试复盘和用户画像不属于当前 UI 范围。

目标 UI 必须与 `ui-design/index.html` 和 `ui-design/src/app.jsx` 当前运行时交互一致。

## 2 已确认决策

1. App 默认进入首页；未登录状态由当前页面内的登录入口和业务前置登录处理。
2. 顶部导航为：`首页`、`模拟面试`、`简历`。
3. 用户菜单为：`设置与隐私`、`退出登录`；未登录时只显示登录入口。
4. `复盘` 和 `用户画像` 不属于当前 UI 范围，不是一级导航、用户菜单入口、目标 route、静态原型页面或后续默认 workstream。
5. `debrief`、`debrief_full`、`profile` 等非当前 hash / route 输入在静态原型中归一到 `home`，不得 materialize 非当前页面。
6. `auth_profile_setup` 仍保留为首次登录资料补全页；这是账号资料补全，不是用户画像。
7. 报告只有 session-scoped Dashboard；报告后续动作只有 `复练当前轮` 与 `进入下一轮`。
8. 简历是一级模块：平铺列表、上传 / 粘贴创建、注册后直接详情、只读原始正文。
9. 语音是面试形式，只能通过 `practice` 显式参数进入；不得暴露独立 `voice` route。
10. 顶栏主题色、暗色模式、语言下拉和设置页字体预设是全局显示控制，不属于业务模块。

## 3 目标产品骨架

```text
[EasyInterview App]
├─ TopBar
│  ├─ Brand: E mark + EasyInterview
│  ├─ Primary nav: 首页 / 模拟面试 / 简历
│  ├─ Theme / dark / language
│  └─ User menu: 设置与隐私 / 退出登录
├─ Home / 首页
│  ├─ JD 输入源
│  │  ├─ 粘贴 JD 输入框
│  │  └─ 上传文件 / URL 导入入口（输入卡底部 source actions）
│  ├─ 选择已有简历（适度宽度下拉框）
│  │  └─ 还没有简历？1 分钟创建（右侧同行）
│  ├─ 立即面试（简历选择下方）
│  └─ 最近模拟面试（最多 3 条 + 更多）
├─ Mock Interview / 模拟面试
│  ├─ 当前面试规划
│  ├─ JD / 简历 / InterviewRound
│  ├─ 公司情报嵌入卡片
│  ├─ 会话记录
│  └─ 立即面试
├─ Interview Session
│  ├─ 文本面试 / 语音面试
│  ├─ 带提示练习 / 严格模拟
│  ├─ 问题推进
│  └─ 结束并生成报告
├─ Report Dashboard
│  ├─ 会话 / 岗位 / 简历 / 轮次 / 形式上下文
│  ├─ 准备度、维度、题目回顾、证据详情
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
├─ 模拟面试
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

## 5 目标模块关系

```text
Home
├─ JD 输入源
│  ├─ 粘贴 JD
│  └─ 上传文件 / URL 导入（同一输入卡底部 source actions）
├─ 选择已有 ready 简历（定宽下拉框）
│  └─ 还没有简历？1 分钟创建（同排）
│     └─ Resume Intake
│
├─ 立即面试（选择简历下方）
│  └─ Parse & Confirm Interview(resumeId 已绑定)
│     └─ Interview Session 或 Mock Interview Plan
├─ 最近模拟面试
│  ├─ 最多 3 条快捷卡片
│  └─ 更多 -> Mock Interview Plan 列表

Mock Interview Plan
├─ TargetJob / JD
├─ Resume
├─ InterviewRound
├─ InterviewSession
└─ ReportDashboard

ReportDashboard
├─ 题目回顾和证据详情
├─ 复练当前轮 -> Interview Session(same round)
└─ 进入下一轮 -> Interview Session(next round)
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
ParseAndConfirmInterview
InterviewSession(sessionId)
ReportGenerating(sessionId)
ReportDashboard(sessionId)
```

### 6.3 用户菜单和认证页面

```text
SettingsPrivacy
AuthLogin
AuthVerify
AuthProfileSetup
AuthLogout
```

## 7 非当前 route 输入归一

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

`voice` 不保留 route alias。判断目标架构时以 `normalizeRoute` 后的 `activeRouteName` 和实际渲染内容为准，不以非当前 hash、非当前画板标签或非当前组件为准。

## 8 后续实现输入

1. 正式前端 TopBar 必须源级复刻当前三入口静态原型。
2. `frontend` 不得注册 `debrief` / `profile` RouteName、primary nav、user menu 或 screen 分支。
3. OpenAPI、backend、migrations、shared、config 和 scenario 的复盘 / 用户画像范围收敛由 product-scope/001-core-loop-module-pruning 承接。
4. `auth_profile_setup` 保留为账号资料补全，不得写成用户画像。
5. Pixel parity 和 route tests 必须覆盖非当前入口负向断言。
