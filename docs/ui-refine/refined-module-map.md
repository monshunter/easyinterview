# EasyInterview UI 目标模块地图

> **版本**: 1.5
> **状态**: active
> **更新日期**: 2026-05-01

## 1 文档目的

本文档把当前静态 UI 页面整理成目标模块，明确哪些能力保留、合并、降级或移除。模块地图以当前 `easyinterview-ui/EasyInterview.html` 原型和 `src/app.jsx` 运行时路由为准。

## 2 保留的核心模块

| 模块 | 用户任务 | 保留页面/能力 | 说明 |
|------|----------|----------------|------|
| Home / 首页 | 粘贴 JD 或继续最近模拟面试 | JD 输入、JD 文件/URL 弹窗、最近模拟面试、创建简历入口 | 默认入口，不需要登录页前置 |
| Job Picks / 岗位推荐 | 基于画像和简历找到值得准备的 JD | 推荐列表、匹配原因、画像摘要、从推荐进入模拟面试 | 一级导航 |
| Mock Interview / 模拟面试 | 确认一场模拟面试的上下文并立即开始 | 当前面试规划、切换/新建规划、JD/简历绑定、面试轮次、立即面试、会话历史 | 一级导航，不再叫当前岗位 |
| Interview Session | 完成一场完整模拟面试 | 文本面试、语音面试、语音转文字、问题推进、结束生成报告 | 会话级页面 |
| Report Dashboard | 查看一次已完成模拟面试的报告 | 仪表盘、上下文条、准备度、维度、题目回顾、证据、复练计划 | 隶属于 session，不是一级导航 |
| Resume / 简历 | 管理简历资产 | 原始简历、结构化主版本、岗位定制版本、上传/粘贴/问答、原件预览 | 一级导航 |
| Debrief / 真实复盘 | 复盘真实面试并生成复盘面试 | 选择目标岗位/JD、关联模拟面试、绑定简历、复盘记录、复盘分析、复盘面试 | 一级导航 |
| User Profile / 用户画像 | 查看和修正系统理解用户的结构化画像 | 来源统计、画像维度、证据来源、用户纠偏、模块使用开关 | 用户菜单入口 |
| Account & Settings / 设置与隐私 | 管理账号基础信息、登录安全和隐私 | 个人基础信息、登录方式、通知、订阅、导出、删除 | 用户菜单入口 |
| Auth / 认证 | 登录、注册和退出 | 登录、注册、邮箱验证、重置登录、退出登录 | 操作级触发，不是默认入口 |

## 3 合并或降级的能力

| 当前能力 | 目标归属 | 调整方式 |
|----------|----------|----------|
| `workspace` | Mock Interview / 当前面试规划 | 产品语义从“当前岗位”改为“模拟面试规划” |
| `company_intel` | Mock Interview | 从独立页降为模拟面试规划页的公司情报分区或详情 |
| `resume_versions` | Resume | 作为一级简历模块的当前入口 |
| `resume` | Resume / Mock Interview | 简历资产在 Resume 管，模拟面试页只选择绑定简历 |
| `jd_match` | Job Picks | 作为一级岗位推荐模块保留 |
| `practice` | Interview Session | 文本面试页面 |
| `voice` | Interview Session | 语音面试形式，不是独立练习类别 |
| `generating` | Interview / Report 过渡态 | 报告生成状态，不作为顶部导航 |
| `report` | Report Dashboard | 会话级报告详情，不作为顶部导航 |
| `debrief` / `debrief_full` | Debrief | 一级真实复盘流程 |
| `profile` | User Profile | 用户菜单入口，承载 AI 画像详情和纠偏 |
| `settings` | Account & Settings | 用户菜单入口，只保存账号基础信息和隐私安全 |
| `auth_*` | Auth | 认证流程页面 |

## 4 移除的模块和流程

| 当前模块或流程 | 目标处理 | 原因 |
|----------------|----------|------|
| Growth | 移除 | 当前阶段看不到对一次岗位准备闭环的直接作用 |
| Multi-round Plan | 移除 | 独立计划系统增加理解成本；面试轮次节点只在模拟面试规划中展示 |
| Experience Library | 移除 | 简历和用户画像承载经历素材，不恢复独立经历库 |
| Welcome / SignIn landing | 移除 | 默认进入首页，登录在操作过程中触发 |
| Practice mode cards | 移除 | 面试是整体过程，不让用户选择热身、反问、单题等模式 |
| Follow-up Tree | 移除 | 追问在面试会话内发生，不作为树形训练入口 |
| Targeted Drill / Single-question Drill | 移除 | 当前不做单点突破式训练 |
| Mistake retry queue | 移除 | 题目问题在报告题目回顾中呈现，不生成独立队列 |
| Report timeline | 移除 | 报告只通过仪表盘展示 |
| Magazine-style report page | 移除独立形态 | 长内容拆入仪表盘模块二级详情 |

## 5 当前路由到目标模块映射

### 5.1 当前目标路由

| 当前 Route | 目标归属 | 目标状态 |
|------------|----------|----------|
| `home` | Home / 首页 | 默认入口 |
| `parse` | JD Parse & Confirm | JD 解析确认步骤；来自首页 JD 导入或岗位推荐确认 |
| `jd_match` | Job Picks / 岗位推荐 | 一级导航；含为你推荐、联网搜索、关注列表 |
| `workspace` | Mock Interview / 当前面试规划 | 一级导航 |
| `practice` | Interview Session / 文本面试 | 会话级页面 |
| `voice` | Interview Session / 语音面试 | 会话级页面 |
| `generating` | ReportGenerating | 报告生成过渡态 |
| `report` | Report Dashboard(sessionId) | 会话级详情；当前运行时只渲染仪表盘 |
| `resume_versions` | Resume / 简历 | 一级导航当前入口 |
| `debrief` / `debrief_full` | Debrief / 真实复盘 | 一级导航 |
| `profile` | User Profile / 用户画像 | 用户菜单入口 |
| `settings` | Account & Settings | 用户菜单入口 |
| `auth_login` | Auth | 登录页 |
| `auth_register` | Auth | 注册页 |
| `auth_verify` | Auth | 邮箱验证页 |
| `auth_reset` | Auth | 重置登录页 |
| `auth_logout` | Auth | 退出登录页 |
| `company_intel` | Mock Interview / Company Intel | 从模拟面试规划页轻量卡片打开的详情页 |

### 5.2 兼容旧路由

`src/app.jsx` 当前通过 `routeAliases` 把部分旧路由折回目标模块。它们可被 hash 或设计画板引用，但不代表恢复对应独立模块。

| 旧 Route | 运行时折回 | 目标状态 |
|----------|------------|----------|
| `welcome` | `home` | 移除默认欢迎页职责 |
| `growth` | `home` | 移除独立成长中心 |
| `plan` | `workspace` | 移除独立多轮计划；轮次只在当前面试规划中展示 |
| `mistakes` | `report` | 移除独立错题队列；题目问题留在报告题目回顾 |
| `drill` | `practice` | 移除独立单题 Drill；运行时折回完整面试 |
| `followup` | `practice` | 移除独立追问树；运行时折回完整面试 |
| `experiences` | `resume_versions` | 移除独立经历库；经历证据并入简历和用户画像 |
| `star` | `resume_versions` | 移除独立 STAR 编辑器；简历改写在简历模块内完成 |

### 5.3 仍可直达但不属于目标入口的历史页面

| Route | 当前处理 | 说明 |
|-------|----------|------|
| `resume` | 历史简历单页 | 顶部导航不使用；目标入口是 `resume_versions` |
| `onboarding` | 历史 5 分钟画像 / 经历卡片页 | 仅作为旧画板或回溯页面存在；当前“1 分钟创建简历”走 `resume_versions` 的 `flow=create` |

## 6 目标模块依赖

```text
User
├─ AuthIdentity
├─ UserProfile
│  ├─ Resume signals
│  ├─ JD signals
│  ├─ Mock interview signals
│  ├─ Debrief signals
│  └─ User corrections
├─ ResumeVersions
│  ├─ OriginalResume
│  ├─ StructuredMaster
│  └─ TargetedVersion
├─ TargetJobs
│  ├─ JD
│  ├─ MatchAnalysis
│  └─ InterviewRounds
├─ MockInterviewPlans
│  ├─ targetJobId
│  ├─ resumeVersionId
│  └─ roundId
├─ InterviewSessions
│  ├─ modality: text / voice
│  └─ transcript
├─ ReportDashboards
└─ DebriefRecords
   └─ DebriefInterview
```

## 7 一致性约束

1. 顶部导航只出现 `首页 / 岗位推荐 / 模拟面试 / 简历 / 真实复盘`。
2. 不再使用 `当前岗位` 表示一级模块；如需表达当前上下文，使用 `当前面试规划`。
3. 不再使用 `面试报告` 表示一级模块；报告必须带 `sessionId` 或等价上下文。
4. `voice` 的保留语义是语音面试形式；文本输入框麦克风必须表述为语音转文字。
5. `jd_match` 不再被描述为首页辅助小入口；它是当前静态 UI 的一级岗位推荐模块。
6. 旧 route 可通过 `routeAliases` 折回当前模块，但文档不得把旧画板标签当作目标导航或独立模块。
7. `reportLayout` 当前不改变运行时报告形态；报告变体只保留为历史代码 / 画板标签，目标报告仍是 Dashboard。
