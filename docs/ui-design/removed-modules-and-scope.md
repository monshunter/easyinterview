# EasyInterview UI 移除模块与范围裁剪

> **版本**: 2.5
> **状态**: active
> **更新日期**: 2026-06-12

## 1 文档目的

本文档记录当前阶段 UI 梳理中确认不进入目标架构的模块、页面职责和交互形态。这里的“移除”指目标信息架构不再保留对应模块、导航和默认流程；当前静态 UI 源码也已清理对应废弃 screen 注册、组件和画板入口。

当前 `src/app.jsx` 还通过 `ROUTE_ALIASES` 把若干旧原型 route 归一到新目标模块。旧 route 不再进入对应历史页面；能通过 hash 访问这些旧 route 只代表静态原型仍有历史入口，不代表旧模块仍是目标架构的一部分，也不代表线上兼容承诺。`voice` route alias 已删除，语音面试只能通过 `practice?mode=voice&modality=voice` 或等价显式参数进入。

## 2 已确认移除

```text
移除模块和流程:
├─ 成长
├─ 多轮计划
├─ 经历库
├─ 未登录落地页
├─ 练习模式选择
├─ 追问树
├─ 针对性复练独立模块
├─ 单题 Drill / 错题复练队列
├─ 面试报告一级导航
├─ 当前岗位一级导航
├─ 收件箱 / Inbox 式旧入口文案
├─ 报告时间线
├─ 独立刊物式报告页
├─ 首次导入链路中 parse 与 session 之间的第二个全页确认
├─ 独立重置登录页（auth_reset / 忘记密码 / 密码 / 两步验证口径）
└─ 复盘感谢信草稿（ThankYouLetter 死代码）
```

不移除但重新定义的能力：

```text
保留但重新定义:
├─ 岗位推荐 -> 一级导航 Job Picks
├─ 语音 -> 面试形式: 语音面试
├─ 麦克风输入 -> 文本面试中的语音转文字
├─ 带提示练习 / 严格模拟 -> 面试会话内辅助程度开关
├─ 复练 -> 报告里的复练当前轮
├─ 下一轮 -> 报告里的进入下一轮
├─ 用户画像 -> 用户菜单里的 AI 画像详情页
└─ 外观偏好 -> 顶栏显示控制和设置页字体预设，不是业务模块
```

## 3 成长模块

### 3.1 当前问题

成长模块像长期趋势看板，但当前用户的关键任务是围绕一份 JD 完成准备。它不能清楚回答用户此刻应该做什么。

### 3.2 目标处理

```text
Growth
└─ 从目标导航和默认流程移除；运行时旧 route 折回 Home
```

若需要展示进展，只在模拟面试规划、报告仪表盘或用户画像中展示与当前任务相关的信号，不提供独立成长中心。

## 4 多轮计划模块

### 4.1 当前问题

多轮计划把“下一步练什么”包装成独立计划系统，增加用户理解成本。当前 UI 只需要在模拟面试规划中展示真实面试轮次节点，在报告中区分复练当前轮和进入下一轮。

### 4.2 目标处理

```text
Plan
└─ 从目标导航和默认流程移除；运行时旧 route 折回 Mock Interview Plan

Mock Interview Plan
└─ 面试轮次节点
   ├─ HR 初筛
   ├─ 技术一面
   ├─ 技术二面
   └─ 经理面
```

面试轮次不是独立计划系统。

## 5 经历库模块

### 5.1 当前问题

经历库要求用户维护一套长期资产，但当前看不到它和“一份 JD 的准备闭环”之间的直接关系。

### 5.2 目标处理

```text
Experience Library
├─ ExperienceLibrary
├─ StarEditor
└─ Onboarding 中的经历沉淀
```

以上能力从当前主流程移除。运行时旧 `experiences` / `star` / `onboarding` route 折回 `resume_versions`；当前“1 分钟创建简历”目标入口是 `resume_versions(flow=create)`。

替代归属：

```text
Resume
└─ 简历原始内容 / 结构化主版本 / 岗位定制版本

User Profile
└─ 系统根据简历、JD、模拟面试、复盘沉淀经历证据
```

旧 `screens-p1-depth.jsx::ResumeVersionsScreen` 中的单页版本工坊实现也不再作为目标入口。当前运行时由后加载的 `screen-resume-workshop.jsx` 覆盖 `window.ResumeVersionsScreen`；旧实现仅以 `_LegacyResumeVersionsScreen` dead code 形式保留，避免在本次简历 IA 调整中改动 1300 行历史文件。

## 6 未登录落地页

### 6.1 当前问题

未登录页把“开始准备”之前又放了一层产品入口。用户应该默认进入首页，直接开始输入 JD。

### 6.2 目标处理

```text
Welcome / SignIn landing
└─ 从默认路由移除；运行时旧 route 折回 Home

Home
└─ 成为默认入口

Auth Pages
├─ 登录（邮箱验证码，单入口承担首次与后续登录）
├─ 邮箱验证
├─ 首次资料补全
└─ 退出登录
```

登录能力保留，但只在操作需要身份、保存、同步、导出或敏感数据处理时出现。

## 7 练习模式选择

### 7.1 当前问题

模式卡片把面试拆成热身、反问、语音、单题深钻、多轮计划等入口，用户会以为自己需要先判断练哪一种。但面试是整体过程，模拟面试规划应该直接进入一场完整面试。

当前静态 UI 保留的是会话内的 `严格模拟` 开关，而不是入口前模式选择。它只控制提示、实时观察、可调用经历和语音现场提示是否展示；用户仍在同一场完整面试里完成练习。

### 7.2 目标处理

```text
Practice mode cards
└─ 移除

Mock Interview Plan Primary CTA
└─ 立即面试

PracticeScreen Assistance Toggle
├─ 带提示练习: 展示提示、实时观察、可调用经历和现场提示
└─ 严格模拟: 隐藏上述辅助信息
```

文本和语音保留，但它们是面试形式，不是模式卡片。带提示练习和严格模拟保留，但它们是会话内辅助程度，不是新的练习入口。

## 8 收件箱式旧入口文案

### 8.1 当前问题

`收件箱` / `Inbox` 会把复盘页误表达成消息或任务收件箱。当前目标架构没有收件箱模块，也没有以收件箱作为复盘入口的设计。

### 8.2 目标处理

```text
Inbox / 收件箱
└─ 从复盘页和其他当前目标页面文案中移除

Debrief Back
└─ 返回首页 / Back home
```

## 9 追问树与针对性复练

### 9.1 当前问题

追问树和针对性复练会把用户带入单点突破流程，和“从 0 进入一场完整面试”的目标相冲突。

### 9.2 目标处理

```text
Follow-up Tree
Targeted Drill
Mistake Retry
Drill Builder
└─ 从目标流程移除；旧 route 折回 Practice 或 Report
```

报告中可以展示题目回顾、回答分析、证据缺口和建议，并允许用户 `加入本轮复练`。但这条路径是重复当前面试轮次的一场完整模拟面试，不是单题 Drill 或错题队列。

## 10 当前岗位一级导航

### 10.1 当前问题

`当前岗位` 容易让用户误以为这是一个岗位资产管理模块。但当前流程中用户从这里进入的是模拟面试前确认：目标岗位、JD、简历和面试轮次的组合。

### 10.2 目标处理

```text
Top Navigation
├─ 删除 当前岗位
└─ 保留 模拟面试
   └─ 当前面试规划
```

页面可以展示当前目标岗位，但模块名称和一级导航语义必须是 `模拟面试`。

## 11 面试报告一级导航

### 11.1 当前问题

用户从一级导航进入报告详情时，会不知道报告属于哪个岗位、哪份简历、哪一轮、哪一次会话。

### 11.2 目标处理

```text
Top Navigation
└─ 删除 面试报告

Report Dashboard(sessionId)
└─ 只能从以下入口进入:
   ├─ 面试结束后
   ├─ 模拟面试历史
   └─ 相关会话入口
```

报告顶部必须展示完整会话上下文。

## 12 报告时间线与刊物式报告

### 12.1 当前问题

时间线和刊物式报告会制造多种报告形态。用户需要的是一个稳定的面试报告入口，而不是在多个视图之间理解差异。

### 12.2 目标处理

```text
Report
└─ 只保留 Dashboard
   ├─ 顶层指标
   ├─ 上下文条
   ├─ 准备度详情
   ├─ 维度详情
   ├─ 题目回顾页
   ├─ 证据详情
   └─ 复练计划
```

刊物式长内容拆入仪表盘模块详情，不再作为独立页面或独立报告模式。

当前 `screen-report.jsx` 只保留 Dashboard。Editorial / Timeline 历史组件定义、`reportLayout` 参数和设计画板报告变体标签已从静态 UI 清理。

## 13 首次导入双重确认结构

### 13.1 当前问题

旧链路中，用户导入 JD 后先在 `parse` 完成“第 2/2 步 · 核对并确认”，又被带到 `workspace` 再做一次“开始前确认这场模拟面试的上下文”，JD 必需项 / 加分项 / 隐性关注点与轮次信息在两页连续重复展示。对带着 JD 来、想最快开练的用户，这是主漏斗里最大的一处冗余摩擦。

### 13.2 目标处理

```text
parse 之后的第二个全页确认
└─ 移除；启动决策（绑定简历、确认轮次、立即面试）收拢进 parse 解析确认页

workspace
└─ 重新定位为回访枢纽
   ├─ Home 最近模拟面试 / Report / 会话历史 / 一级导航进入
   ├─ parse 的仅保存规划进入
   ├─ 切换 / 新建规划
   └─ 立即面试（回访再练）
```

`workspace` 页面能力保留，移除的是它在首次导入链路中的必经确认职责。

## 14 独立重置登录页与复盘感谢信残留

### 14.1 当前问题

产品已统一为邮箱验证码登录，任何页面都没有密码，但 `auth_reset` 仍以“密码重置”页面存在，功能与登录页发送验证码完全重叠；`screens-p1-depth.jsx` 中的 `ThankYouLetter` 感谢信草稿组件定义后从未被渲染，也不属于当前复盘范围。

### 14.2 目标处理

```text
auth_reset
└─ 页面与组件删除；route 归一回 auth_login
   ├─ 登录页保留收不到验证码的帮助说明
   └─ 验证码重发 / 更换邮箱在 auth_verify 完成

ThankYouLetter
└─ 死代码删除；不作为复盘或任何模块的目标能力恢复
```

## 15 受影响页面

| 当前页面 | 目标处理 |
|----------|----------|
| `welcome` | 移除默认入口职责；运行时折回 `home` |
| `growth` | 移除独立成长中心；运行时折回 `home` |
| `plan` | 移除独立多轮计划；运行时折回 `workspace` |
| `experiences` | 移除独立经历库；运行时折回 `resume_versions` |
| `star` | 移除独立 STAR 编辑器；运行时折回 `resume_versions` |
| `onboarding` | 运行时折回 `resume_versions`；移除旧经历库前置职责，简历创建目标入口是 `resume_versions(flow=create)` |
| `resume` | 运行时折回 `resume_versions`；不作为顶部导航或目标入口 |
| `debrief_full` | 同 `debrief` 的历史别名；不新增独立模块 |
| `followup` | 移除独立追问树；运行时折回 `practice` |
| `mistakes` | 移除独立错题复练流程；运行时折回 `report` |
| `drill` | 移除独立单题 Drill；运行时折回 `practice` |
| `voice` | route alias 已删除；语音面试形式保留在 `practice?mode=voice&modality=voice`，不保留独立页面骨架 |
| `report` | 保留为 session-scoped 仪表盘报告 |
| `resume_versions` | 保留为一级简历模块当前入口 |
| `jd_match` | 保留为一级岗位推荐 |
| `profile` | 保留为用户菜单里的用户画像 |
| `settings` | 保留为用户菜单里的账号、隐私、界面偏好入口 |
| `auth_reset` | 移除独立重置登录页；运行时折回 `auth_login`，验证码重发与更换邮箱由 `auth_verify` 承担 |
| `auth_login` / `auth_verify` / `auth_profile_setup` / `auth_logout` | 保留为认证流程页面 |

## 16 当前静态 UI 中已清理或失效的废弃 / 历史代码

以下清单用于约束后续文档和实现判断：这些组件、画板入口或旧实现已从当前目标运行时清理 / 归一 / dead code 化，后续不得作为目标页面恢复。当前目标以顶部导航、`ROUTE_ALIASES` 归一后的 `activeRouteName`、实际渲染内容和后加载覆盖关系为准；`voice` 不经过 `ROUTE_ALIASES`，必须使用 `practice` 显式参数。

| 旧代码位置 / 组件 | 关联 route 或入口 | 清理后状态 | 目标处理 |
|----------------|-------------------|--------------|----------|
| `screens-completion.jsx::WelcomeScreen` | `welcome` | 文件已删除；route 折回 `home` | 废弃未登录欢迎页职责 |
| `screens-rest.jsx::MistakesScreen` | `mistakes` | 文件已删除；route 折回 `report` | 废弃独立错题本和错题复练队列 |
| `screens-completion.jsx::DrillBuilderScreen` | `drill` | 文件已删除；route 折回 `practice` | 废弃单题 Drill 构建器 |
| `screens-completion.jsx::FollowUpTreeScreen` | `followup` | 文件已删除；route 折回 `practice` | 废弃独立追问树 |
| `screens-rest.jsx::GrowthScreen` | `growth` | 文件已删除；route 折回 `home` | 废弃独立成长中心 |
| `screens-p2.jsx::PlanScreen` | `plan` | 文件已删除；route 折回 `workspace` | 废弃独立多轮计划页；`workspace` 内的轮次节点仍保留 |
| `screens-p1-depth.jsx::ExperienceLibraryScreen` | `experiences` | 组件已删除；route 折回 `resume_versions` | 废弃独立经历库 |
| `screens-p1-depth.jsx::ResumeVersionsScreen` | `resume_versions` 旧实现 | 导出改为 `_LegacyResumeVersionsScreen`；`screen-resume-workshop.jsx` 后加载并覆盖 `window.ResumeVersionsScreen` | 废弃旧单页版本工坊；以新简历工坊列表 / 详情 / 分叉 IA 为准 |
| `screens-completion.jsx::StarEditorScreen` | `star` | 文件已删除；route 折回 `resume_versions` | 废弃独立 STAR 编辑器 |
| `screen-report.jsx::ReportEditorial` / `ReportTimeline` | 报告变体标签 / `reportLayout` | 组件、参数和画板变体已删除；`report` 只渲染 Dashboard | 废弃独立刊物式报告和时间线报告形态 |
| `screens-rest.jsx::DebriefScreen` | 旧复盘实现 | 文件已删除；当前目标使用 `DebriefFullScreen` | 废弃旧复盘页 |
| `screens-rest.jsx::ResumeScreen` | `resume` | 文件已删除；route 折回 `resume_versions` | 废弃为目标入口；以 `resume_versions` 为准 |
| `screens-p0-complete.jsx::OnboardingScreen` | `onboarding` | 组件已删除；route 折回 `resume_versions` | 废弃为当前简历创建入口；以 `resume_versions(flow=create)` 为准 |
| `screens-p2.jsx::VoicePracticeScreen` | `voice` | 文件已删除；`voice` route alias 已删除 | 语音能力保留为 `PracticeScreen` 内的语音 Surface，不恢复独立语音页骨架 |
| `screen-auth.jsx::AuthResetScreen` | `auth_reset` | 组件已删除；route 归一回 `auth_login` | 废弃独立重置登录页；无密码产品不保留“密码重置”形态 |
| `screens-p1-depth.jsx::ThankYouLetter` | 无（定义后从未渲染） | 组件已删除 | 废弃复盘感谢信草稿；不进入当前复盘范围 |

语音能力不是废弃能力；废弃的是脱离 `PracticeScreen` 外层骨架的独立语音页面呈现。

## 17 外观偏好不是业务模块

### 17.1 当前处理

```text
外观偏好
├─ 顶栏主题色
├─ 暗色模式
├─ 语言下拉
└─ 设置页字体预设
```

这些控制保留，因为它们是横切的 UI 呈现能力；但它们不属于岗位推荐、模拟面试、报告、简历或复盘的业务模块，也不应该成为新的一级导航或 onboarding 步骤。

## 18 未来重新引入条件

被移除模块未来如需重新引入，必须先回答：

1. 用户在什么时刻会主动需要它。
2. 它解决的是哪一个具体任务。
3. 它和首页、岗位推荐、模拟面试、报告、简历、复盘闭环是什么关系。
4. 它是否值得成为一级导航或独立模块。

在这些问题没有明确答案前，不应把这些模块放回主流程。
