# EasyInterview UI 目标模块地图

> **版本**: 2.20
> **状态**: active
> **更新日期**: 2026-07-15

## 1 文档目的

本文档把当前产品页面整理成目标模块，明确当前能力归属、合并入口和范围边界。模块地图以当前产品 spec、本文档 route catalog 和正式前端 route contract 为准。

## 2 当前核心模块

| 模块 | 用户任务 | 页面/能力 | 说明 |
|------|----------|----------------|------|
| Home / 首页 | 粘贴 JD 或继续最近模拟面试 | 单一 JD 文本框、ready 简历选择、最近模拟面试、创建简历入口 | 默认入口；JD 获取唯一入口，只接受粘贴文本 |
| Interview / 面试 | 浏览并回访既有面试规划，再次发起 session | 面试规划列表、当前面试规划、切换/新建规划、JD/简历绑定、面试轮次、公司情报嵌入卡片、立即面试、会话记录 | 一级导航 |
| Interview Session | 完成一场完整模拟面试 | 全宽连续文本聊天、user/assistant 安全 Markdown/GFM、普通消息自然推进、disabled 电话入口、结束生成报告 | 会话级页面 |
| Reports / 当前规划报告 | 查看一个面试规划各轮当前报告与最新生成状态 | target-scoped canonical round list、loading/empty/error、进入 report/generating | 规划上下文页面，不是 TopBar 一级导航或全局中心 |
| Report Dashboard | 查看一次已完成模拟面试的报告 | 仪表盘、上下文条、准备度、维度、证据、下一步；Header 唯一一对复练 / 下一轮 CTA | 隶属于 session，不是一级导航 |
| Resume / 简历 | 管理简历资产 | 平铺简历列表、上传/粘贴创建后直接打开详情、只读原始正文、LLM-derived displayName、禁止 raw 第一行/文件名命名 | 一级导航 |
| Account & Settings / 设置与隐私 | 查看真实账号信息并执行账号与隐私动作 | 只读姓名/完整账号邮箱、退出登录、导出暂不可用、删除账号 | 已登录 TopBar 设置齿轮入口 |
| Auth / 认证 | 登录和退出 | 邮箱验证码登录、邮箱验证、首次账号资料补全、退出登录 | 操作级触发，不是默认入口 |
| Global Display Controls / 全局显示控制 | 调整 UI 呈现 | 顶栏 Ocean / Plum / custom accent（仅色相、饱和度）、暗色模式、语言下拉 | 横切能力；custom accent 无 preview/value/reset，选择 Ocean / Plum 即退出自定义色；字体采用固定产品栈 |

## 3 当前归属与合并入口

| 当前能力 | 目标归属 | 调整方式 |
|----------|----------|----------|
| `workspace` | Interview / 面试规划列表 + 面试规划详情 | 无 `targetJobId` 时展示列表；只带 `targetJobId` 时展示统一只读详情；卡片直达详情，`planId` 不是 locator |
| 公司情报 | Interview | 只保留当前面试规划页内嵌轻量卡片 |
| `resume_versions` | Resume | 一级简历模块当前入口 |
| `practice` | Interview Session | 连续文本聊天；user/assistant 使用安全 Markdown/GFM view projection，retry 保持 raw payload；电话入口置灰 |
| `reports` | Reports / 当前规划报告 | 只接受 targetJobId 的规划范围索引；不作为顶部导航 |
| `generating` | Interview / Report 过渡态 | 报告生成状态，不作为顶部导航 |
| `report` | Report Dashboard | 会话级报告详情，不作为顶部导航 |
| `settings` | Account & Settings | 已登录 TopBar 设置齿轮入口；退出登录位于页面内 |
| `auth_*` | Auth | 认证流程页面 |

## 4 当前范围外模块和流程

| 范围外模块或流程 | 当前边界 | 原因 |
|----------------|----------|------|
| Debrief / 复盘 | 不作为 route/API/DB/event/job/config/scenario 正向资产 | 当前核心闭环是 JD / 简历 -> 模拟面试 -> 报告 -> 复练当前轮 / 进入下一轮 |
| User Profile / 用户画像 | 不沉淀独立候选人画像产品或数据模型；账号资料补全保留 | 账号资料和候选人画像是不同边界 |
| Job Picks / 岗位推荐 | 不作为独立模块；JD 获取唯一入口是首页导入 | 避免平行入口 |
| 简历版本树 / 主版本 / 岗位定制版本 / 分叉流程 | 不作为当前简历管理形态 | 简历按平铺资产管理 |
| 简历轻量问答建档 | 不作为创建入口 | 创建简历只保留上传 / 粘贴 |
| 设置页通知 / 订阅扩展 | 不作为当前设置页 tab | 如需通知或订阅能力，先更新本目录设计文档和 owner spec/plan，再进入前端实施 |
| Growth / Multi-round Plan / Experience Library / Drill / Mistakes | 不作为当前模块 | 不增强当前 JD -> 模拟 -> 报告闭环 |

## 5 当前路由到目标模块映射

### 5.1 当前目标路由

| 当前 Route | 目标归属 | 目标状态 |
|------------|----------|----------|
| `home` | Home / 首页 | 默认入口 |
| `parse` | JD Import Command Progress | 仅新导入 queued/processing；ready 后 replace 到 Workspace detail |
| `workspace` | Interview / 面试规划列表 + 当前面试规划 | `/workspace` 列表；`/workspace?targetJobId` 只读详情；一级导航 |
| `practice` | Interview Session | 会话级页面 |
| `reports` | ReportsScreen(targetJobId) | 规划范围上下文页面，chrome visible、非一级导航 |
| `generating` | ReportGenerating | 报告生成过渡态 |
| `report` | Report Dashboard(reportId) | 会话级详情 |
| `resume_versions` | Resume / 简历 | 一级导航 |
| `settings` | Account & Settings | 已登录设置齿轮直达的受保护单页 |
| `auth_login` | Auth | 登录页 |
| `auth_verify` | Auth | 邮箱验证页 |
| `auth_profile_setup` | Auth | 首次账号资料补全页 |
| `auth_logout` | Auth | 退出登录页 |

### 5.2 范围外路由归一

| 范围外 Route | 运行时折回 | 当前边界 |
|----------|------------|----------|
| `welcome` | `home` | 不 materialize 默认欢迎页 |
| `growth` | `home` | 不 materialize 独立成长中心 |
| `plan` | `workspace` | 不 materialize 独立多轮计划 |
| `mistakes` | `report` | 不 materialize 独立错题队列 |
| `drill` | `practice` | 不 materialize 独立单题 Drill |
| `followup` | `practice` | 不 materialize 独立追问树 |
| `experiences` | `resume_versions` | 不 materialize 独立经历库 |
| `star` | `resume_versions` | 不 materialize 独立 STAR 编辑器 |
| `resume` | `resume_versions` | 不 materialize 范围外简历单页 |
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
├─ TargetJobReportOverviews
└─ ReportDashboards
```

当前目标数据依赖之外：

- `DebriefRecords`
- `CandidateProfile`
- `ExperienceCard`

## 7 一致性约束

1. 顶部导航只出现 `首页 / 面试 / 简历`。
2. 已登录 TopBar 账号区只出现一个设置齿轮；退出登录只从 Settings 进入既有确认页。
3. `debrief`、`debrief_full`、`profile` 不得作为目标 route、screen key、data-testid 正向锚点或场景正向入口。
4. `auth_profile_setup` 是账号资料补全，不是用户画像。
5. 复盘和用户画像不得作为静态源码、设计文档、正式前端、OpenAPI、backend、migrations、shared、config、scenario 正向资产。
6. Home JD intake 只渲染 textarea、ready Resume 下拉框与「立即面试」CTA；不得出现其他 JD 导入控件、弹窗或并行请求形态。Resume 模块的文件上传不受此约束影响。
7. `reports` 只展示当前 `targetJobId` 的 canonical rounds、current report 与 latest attempt；入口位于 Workspace 详情内容区，不进入 TopBar。Parse 无 ready 详情、嵌入列表/section 兼容，Reports Back 返回 Workspace detail，Report/Generating route 仍 reportId-only。
8. Custom accent picker 只保留 hue/saturation；preview/value 区、“恢复主题默认色 / Reset to theme accent”与 `onClear` / `active` 冗余 props 必须零引用。Ocean / Plum 是退出自定义色的唯一预定义主题动作。
9. ready Home/Workspace 卡片只进入 `/workspace?targetJobId`，不得 import、poll、播放 Parse animation 或使用 `planId/resumeId` 做详情 locator；只有 Home POST import 进入 `/parse?targetJobId` 命令进度。
10. Practice user/assistant Markdown 投影启用 `skipHtml` 且不使用 `rehypeRaw`；remote image/unsafe URI 不执行，安全 link hardened，same-ID retry 使用原始 text/clientMessageId，mobile code/table 不造成 document overflow。
11. Settings 不保留 tab、登录安全、字体预设、产品信息或手机号/界面语言/时区等无当前数据源字段；Account 只读展示 runtime `/me` 姓名与完整账号邮箱，Privacy 只展示导出不可用和删除账号。

## 8 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-15 | 2.20 | 采用设置简化方案 A：账号入口收敛为设置齿轮，Settings 只保留真实账号/隐私动作，全局字体改为固定产品栈。 |
