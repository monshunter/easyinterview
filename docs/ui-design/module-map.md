# EasyInterview UI 目标模块地图

> **版本**: 2.9
> **状态**: active
> **更新日期**: 2026-07-06

## 1 文档目的

本文档把当前静态 UI 页面整理成目标模块，明确当前能力归属、合并入口和范围边界。模块地图以当前 `ui-design/index.html` 原型和 `src/app.jsx` 运行时路由为准。

## 2 当前核心模块

| 模块 | 用户任务 | 页面/能力 | 说明 |
|------|----------|----------------|------|
| Home / 首页 | 粘贴 JD 或继续最近模拟面试 | JD 输入、JD 文件/URL 弹窗、最近模拟面试、创建简历入口 | 默认入口；JD 获取唯一入口 |
| Mock Interview / 模拟面试 | 回访既有面试规划并再次发起 session | 当前面试规划、切换/新建规划、JD/简历绑定、面试轮次、公司情报嵌入卡片、立即面试、会话记录 | 一级导航 |
| Interview Session | 完成一场完整模拟面试 | 文本面试、语音面试、语音转文字、带提示练习 / 严格模拟、问题推进、结束生成报告 | 会话级页面 |
| Report Dashboard | 查看一次已完成模拟面试的报告 | 仪表盘、上下文条、准备度、维度、题目回顾、证据、复练计划；Header 唯一一对复练 / 下一轮 CTA | 隶属于 session，不是一级导航 |
| Resume / 简历 | 管理简历资产 | 平铺简历列表、上传/粘贴创建、解析预览确认、简历详情、改写仅采纳、采纳后覆盖或另存 | 一级导航 |
| Account & Settings / 设置与隐私 | 管理账号基础信息、登录安全、界面偏好和隐私 | 个人基础信息、登录方式、字体预设、导出、删除 | 用户菜单入口 |
| Auth / 认证 | 登录和退出 | 邮箱验证码登录、邮箱验证、首次账号资料补全、退出登录 | 操作级触发，不是默认入口 |
| Global Display Controls / 全局显示控制 | 调整 UI 呈现 | 顶栏主题色、暗色模式、语言下拉，设置页字体预设 | 横切能力 |

## 3 当前归属与合并入口

| 当前能力 | 目标归属 | 调整方式 |
|----------|----------|----------|
| `workspace` | Mock Interview / 当前面试规划 | 回访枢纽；首次导入启动决策由 `parse` 承载 |
| 公司情报 | Mock Interview | 只保留模拟面试规划页内嵌轻量卡片 |
| `resume_versions` | Resume | 一级简历模块当前入口 |
| `practice` | Interview Session | 文本面试与语音面试共享会话页面 |
| `generating` | Interview / Report 过渡态 | 报告生成状态，不作为顶部导航 |
| `report` | Report Dashboard | 会话级报告详情，不作为顶部导航 |
| `settings` | Account & Settings | 用户菜单入口 |
| `auth_*` | Auth | 认证流程页面 |

## 4 当前范围外模块和流程

| 范围外模块或流程 | 当前边界 | 原因 |
|----------------|----------|------|
| Debrief / 复盘 | 不作为 route/API/DB/event/job/config/scenario 正向资产 | 当前核心闭环是 JD / 简历 -> 模拟面试 -> 报告 -> 复练当前轮 / 进入下一轮 |
| User Profile / 用户画像 | 不沉淀独立候选人画像产品或数据模型；账号资料补全保留 | 账号资料和候选人画像是不同边界 |
| Job Picks / 岗位推荐 | 不作为独立模块；JD 获取唯一入口是首页导入 | 避免平行入口 |
| 简历版本树 / 主版本 / 岗位定制版本 / 分叉流程 | 不作为当前简历管理形态 | 简历按平铺资产管理 |
| 简历轻量问答建档 | 不作为创建入口 | 创建简历只保留上传 / 粘贴 |
| 设置页通知 / 订阅占位 tab | 不作为当前设置页 tab | 占位能力未来按需重新设计 |
| Growth / Multi-round Plan / Experience Library / Drill / Mistakes | 不作为当前模块 | 不增强当前 JD -> 模拟 -> 报告闭环 |

## 5 当前路由到目标模块映射

### 5.1 当前目标路由

| 当前 Route | 目标归属 | 目标状态 |
|------------|----------|----------|
| `home` | Home / 首页 | 默认入口 |
| `parse` | JD Parse & Confirm | JD 解析确认与启动页 |
| `workspace` | Mock Interview / 当前面试规划 | 一级导航 |
| `practice` | Interview Session | 会话级页面 |
| `generating` | ReportGenerating | 报告生成过渡态 |
| `report` | Report Dashboard(sessionId) | 会话级详情 |
| `resume_versions` | Resume / 简历 | 一级导航 |
| `settings` | Account & Settings | 用户菜单入口 |
| `auth_login` | Auth | 登录页 |
| `auth_verify` | Auth | 邮箱验证页 |
| `auth_profile_setup` | Auth | 首次账号资料补全页 |
| `auth_logout` | Auth | 退出登录页 |

### 5.2 非当前原型路由归一

| 非当前 Route | 运行时折回 | 当前边界 |
|----------|------------|----------|
| `welcome` | `home` | 不 materialize 默认欢迎页 |
| `growth` | `home` | 不 materialize 独立成长中心 |
| `plan` | `workspace` | 不 materialize 独立多轮计划 |
| `mistakes` | `report` | 不 materialize 独立错题队列 |
| `drill` | `practice` | 不 materialize 独立单题 Drill |
| `followup` | `practice` | 不 materialize 独立追问树 |
| `experiences` | `resume_versions` | 不 materialize 独立经历库 |
| `star` | `resume_versions` | 不 materialize 独立 STAR 编辑器 |
| `resume` | `resume_versions` | 不 materialize 非当前简历单页 |
| `onboarding` | `resume_versions` | 不 materialize onboarding 单页 |
| `auth_register` | `auth_login` | 不 materialize 独立注册页 |
| `auth_reset` | `auth_login` | 不 materialize 独立重置登录页 |
| `jd_match` | `home` | 不 materialize 岗位推荐 |
| `debrief` / `debrief_full` | `home` | 不 materialize 复盘 |
| `profile` | `home` | 不 materialize 用户画像 |

## 6 目标数据依赖

```text
User
├─ AuthIdentity
├─ Resumes
├─ TargetJobs
├─ MockInterviewPlans
├─ InterviewSessions
└─ ReportDashboards
```

当前目标数据依赖之外：

- `DebriefRecords`
- `CandidateProfile`
- `ExperienceCard`

## 7 一致性约束

1. 顶部导航只出现 `首页 / 模拟面试 / 简历`。
2. 用户菜单只出现 `设置与隐私 / 退出登录`。
3. `debrief`、`debrief_full`、`profile` 不得作为目标 route、screen key、data-testid 正向锚点或场景正向入口。
4. `auth_profile_setup` 是账号资料补全，不是用户画像。
5. 复盘和用户画像不得作为静态源码、设计文档、正式前端、OpenAPI、backend、migrations、shared、config、scenario 正向资产。
