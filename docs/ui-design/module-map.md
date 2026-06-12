# EasyInterview UI 目标模块地图

> **版本**: 2.7
> **状态**: active
> **更新日期**: 2026-06-12

## 1 文档目的

本文档把当前静态 UI 页面整理成目标模块，明确哪些能力保留、合并、降级或移除。模块地图以当前 `ui-design/index.html` 原型和 `src/app.jsx` 运行时路由为准。

## 2 保留的核心模块

| 模块 | 用户任务 | 保留页面/能力 | 说明 |
|------|----------|----------------|------|
| Home / 首页 | 粘贴 JD 或继续最近模拟面试 | JD 输入、JD 文件/URL 弹窗、最近模拟面试、创建简历入口、复盘辅助入口 | 默认入口，不需要登录页前置；JD 获取唯一入口（D-17） |
| Mock Interview / 模拟面试 | 回访既有面试规划并再次发起 session | 当前面试规划（回访枢纽）、切换/新建规划、JD/简历绑定、面试轮次、公司情报嵌入卡片、立即面试、会话历史 | 一级导航，不再叫当前岗位；首次导入的启动决策在 `parse` 完成，本页不是首次导入的必经确认页 |
| Interview Session | 完成一场完整模拟面试 | 文本面试、语音面试、语音转文字、带提示练习 / 严格模拟、问题推进、右侧底部固定结束生成报告 | 会话级页面 |
| Report Dashboard | 查看一次已完成模拟面试的报告 | 仪表盘、上下文条、准备度、维度、题目回顾、证据、复练计划；Header 唯一一对复练 / 下一轮 CTA | 隶属于 session，不是一级导航 |
| Resume / 简历 | 管理简历资产 | 平铺简历列表、上传/粘贴创建、解析预览确认、简历详情（预览/改写建议/手动编辑）、改写仅采纳、采纳后覆盖或另存、原件预览、导出/复制反馈 | 一级导航；无版本树 / 主版本 / 岗位定制概念（D-20） |
| Debrief / 复盘 | 复盘真实面试并生成复盘面试 | 选择目标岗位/JD、关联模拟面试、绑定简历、文本 / 语音添加共享复盘记录、复盘分析、复盘面试 | 一级导航 |
| User Profile / 用户画像 | 查看和修正系统理解用户的结构化画像 | 来源统计、画像维度、证据来源、用户纠偏、模块使用开关 | 用户菜单入口；画像唯一呈现处 |
| Account & Settings / 设置与隐私 | 管理账号基础信息、登录安全、界面偏好和隐私 | 个人基础信息、登录方式、字体预设、导出、删除 | 用户菜单入口；只有个人资料与隐私数据两个 tab（D-21） |
| Auth / 认证 | 登录和退出 | 邮箱验证码登录、邮箱验证、首次资料补全、退出登录 | 操作级触发，不是默认入口；无密码、无独立重置登录页 |
| Global Display Controls / 全局显示控制 | 调整 UI 呈现 | 顶栏主题色（四预设 + 自定义 accent，默认深海）、暗色模式、语言下拉，设置页字体预设 | 横切能力，不是业务模块 |

## 3 合并或降级的能力

| 当前能力 | 目标归属 | 调整方式 |
|----------|----------|----------|
| `workspace` | Mock Interview / 当前面试规划 | 产品语义从“当前岗位”改为“模拟面试规划”，定位回访枢纽；首次导入启动决策由 `parse` 承载 |
| 公司情报 | Mock Interview | 只保留模拟面试规划页内嵌轻量卡片；`company_intel` 独立详情页与 route 已删除（D-18） |
| `resume_versions` | Resume | 作为一级简历模块的当前入口；`flow=create`、`resumeId` 和 `tab` 驱动创建和详情子状态 |
| `resume` | Resume / Mock Interview | 简历资产在 Resume 管，模拟面试页只选择绑定简历 |
| `practice` | Interview Session | 文本面试与语音面试共享的会话页面；由 `mode/modality` 决定中间 Surface，由 `practiceMode` 决定带提示练习或严格模拟；语音面试必须通过 `practice?mode=voice&modality=voice` 或等价显式参数进入 |
| `generating` | Interview / Report 过渡态 | 报告生成状态，不作为顶部导航 |
| `report` | Report Dashboard | 会话级报告详情，不作为顶部导航 |
| `debrief` / `debrief_full` | Debrief | 一级复盘流程 |
| `profile` | User Profile | 用户菜单入口，承载 AI 画像详情和纠偏 |
| `settings` | Account & Settings | 用户菜单入口，只保存账号基础信息和隐私安全 |
| `auth_*` | Auth | 认证流程页面 |
| 顶栏主题 / 暗色 / 语言 | Global Display Controls | 只改变显示，不改变业务模块 |
| 设置页字体预设 | Account & Settings / Interface Preferences | 原子切换 serif/sans 字体组合 |

## 4 移除的模块和流程

| 当前模块或流程 | 目标处理 | 原因 |
|----------------|----------|------|
| Job Picks / 岗位推荐 | 移除（D-17） | 超出 MVP 闭环；JD 获取唯一入口是首页导入 |
| Company Intel 独立详情页 | 移除（D-18） | 轻量情报由模拟面试规划页嵌入卡片承载 |
| 简历版本树 / 主版本 / 岗位定制版本 / 分叉流程 | 移除（D-20） | 简历按平铺资产管理，不做版本继承 |
| 简历轻量问答建档 | 移除（D-20） | 创建简历只保留上传 / 粘贴 |
| 设置页通知 / 订阅占位 tab | 移除（D-21） | 占位能力未来按需重新设计，不以空 tab 预留 |
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
| `parse` | JD Parse & Confirm | JD 解析确认与启动页；来自首页 JD 导入；核对解析结果、绑定简历、确认轮次后立即面试，或仅保存规划进入 `workspace` |
| `workspace` | Mock Interview / 当前面试规划 | 一级导航；回访枢纽，不插入首次导入链路 |
| `practice` | Interview Session / 文本或语音面试 | 会话级页面；`mode/modality` 决定文本输入区或语音波形 / 表达指标区，`practiceMode` 决定辅助信息显隐 |
| `generating` | ReportGenerating | 报告生成过渡态 |
| `report` | Report Dashboard(sessionId) | 会话级详情；当前运行时只渲染仪表盘 |
| `resume_versions` | Resume / 简历 | 一级导航当前入口 |
| `debrief` / `debrief_full` | Debrief / 复盘 | 一级导航 |
| `profile` | User Profile / 用户画像 | 用户菜单入口 |
| `settings` | Account & Settings | 用户菜单入口 |
| `auth_login` | Auth | 登录页 |
| `auth_verify` | Auth | 邮箱验证页 |
| `auth_profile_setup` | Auth | 首次登录资料补全页 |
| `auth_logout` | Auth | 退出登录页 |

### 5.2 历史原型路由归一

`src/app.jsx` 当前通过 `ROUTE_ALIASES` 把部分旧原型路由归一到目标模块。它们可被 hash 引用，但不代表恢复对应独立模块，也不构成线上兼容承诺。`activeRouteName` 会先经过 `normalizeRoute`，因此旧 route 只作为静态原型的历史入口；历史组件即使保留，也不得重新绕过目标 route 骨架。`voice` 不再保留 route alias，语音面试只能从 `practice` 显式参数进入。

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
| `resume` | `resume_versions` | 移除历史简历单页；目标入口是新简历工坊 |
| `onboarding` | `resume_versions` | 移除历史 5 分钟画像 / 经历卡片页；当前简历创建走 `flow=create` |
| `auth_register` | `auth_login` | 移除独立注册页；邮箱验证码登录统一承担首次登录和后续登录 |
| `auth_reset` | `auth_login` | 移除独立重置登录页；无密码产品中验证码重发与更换邮箱由 `auth_verify` 承担 |
| `jd_match` | `home` | 移除岗位推荐一级模块（D-17）；JD 获取唯一入口是首页导入 |
| `company_intel` | `workspace` | 移除公司情报独立详情页（D-18）；轻量情报由模拟面试规划页嵌入卡片承载 |

### 5.3 历史别名但不新增模块的页面

| Route | 当前处理 | 说明 |
|-------|----------|------|
| `debrief_full` | 同 `debrief` 的历史别名 | 与 `debrief` 渲染同一目标复盘页面，不新增模块 |

### 5.4 已清理废弃代码清单

| 代码位置 / 组件 | 关联 route 或入口 | 当前状态 |
|----------------|-------------------|----------|
| `screens-rest.jsx` | `mistakes` / `resume` / `growth` / 旧复盘 | 文件已删除 |
| `screens-completion.jsx` | `welcome` / `drill` / `followup` / `star` | 文件已删除 |
| `screens-p2.jsx` | `plan` / `voice` | 文件已删除；历史 `PlanScreen` 和 `VoicePracticeScreen` 不再作为目标骨架 |
| `screens-p1-depth.jsx::ExperienceLibraryScreen` | `experiences` | 组件已删除；保留 `DebriefFullScreen` |
| `screens-p1-depth.jsx::ResumeVersionsScreen`（`_LegacyResumeVersionsScreen`） | `resume_versions` 旧实现 | dead code 已删除；当前唯一实现是 `screen-resume-workshop.jsx` 的平铺简历工坊 |
| `screens-p0-complete.jsx::OnboardingScreen` | `onboarding` | 组件已删除；保留 `ParseScreen`、`ReportGeneratingScreen` 和 `SettingsScreen` |
| `screen-report.jsx::ReportEditorial` / `ReportTimeline` | 报告变体标签 / `reportLayout` | 组件和参数已删除；`ReportScreen` 只返回 Dashboard |
| `screen-auth.jsx::AuthResetScreen` | `auth_reset` | 组件已删除；route 归一回 `auth_login`，验证码重发由 `auth_verify` 承担 |
| `screens-p1-depth.jsx::ThankYouLetter` | 无（定义后从未渲染） | 死代码已删除；感谢信草稿不属于当前复盘范围 |
| `screen-jd-match.jsx::JDMatchScreen` | `jd_match` | 文件已删除；route 归一回 `home`（D-17） |
| `screen-company-intel.jsx::CompanyIntelScreen` | `company_intel` | 独立详情页组件已删除；文件仅保留 `CompanyIntelEmbed` 嵌入卡片（D-18） |
| `screen-resume-workshop.jsx::ResumeTreeView` / `ResumeBranchFlow` | 简历版本树 / 分叉流程 | 组件已删除；列表只保留平铺视图（D-20） |
| `screens-p0-complete.jsx::SettingsNotif` / `SettingsBilling` | 设置通知 / 订阅占位 tab | 组件已删除（D-21） |
| `screen-report.jsx::IssueRow` / `PerQBlock` / `KVInline` | 无（定义后从未渲染） | 死代码已删除 |

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
├─ Resumes（平铺资产）
│  ├─ originalSource（文件 / 粘贴文本 · 只读）
│  ├─ parsedTextSnapshot
│  └─ structuredContent
├─ TargetJobs
│  ├─ JD
│  ├─ MatchAnalysis
│  └─ InterviewRounds
├─ MockInterviewPlans
│  ├─ targetJobId
│  ├─ resumeId
│  └─ roundId
├─ InterviewSessions
│  ├─ modality: text / voice
│  ├─ practiceMode: assisted / strict
│  ├─ hintUsed
│  ├─ hintCount
│  └─ transcript
├─ ReportDashboards
└─ DebriefRecords
   ├─ entries
   │  └─ source: confirmed / text_guided / voice_extracted / manual
   └─ DebriefInterview
```

## 7 一致性约束

1. 顶部导航只出现 `首页 / 模拟面试 / 简历 / 复盘`。
2. 不再使用 `当前岗位` 表示一级模块；如需表达当前上下文，使用 `当前面试规划`。
3. 不再使用 `面试报告` 表示一级模块；报告必须带 `sessionId` 或等价上下文。
4. 语音面试的目标入口是 `practice?mode=voice&modality=voice` 或等价显式参数；不得保留或新增 `voice` route，文本输入框麦克风必须表述为语音转文字。
5. `jd_match` 与 `company_intel` 不再是目标 route：旧 hash 分别归一回 `home` 与 `workspace`，不得据此恢复岗位推荐模块或独立公司情报页；公司情报只出现在模拟面试规划页嵌入卡片。
6. 除 `voice` 外的旧原型 route 可通过 `ROUTE_ALIASES` 归一到当前模块，但文档不得把旧画板标签当作目标导航或独立模块。
7. `reportLayout`、报告变体组件和报告变体画板不得恢复；目标报告仍是 Dashboard。
8. 主题色、暗色模式、语言下拉和字体预设是全局显示控制，不得被写成岗位、面试、报告或认证模块。
9. 设置页可以维护界面偏好，但不得把目标岗位、年限、薪资偏好等画像信息移入个人资料。
10. 判断当前目标模块时，以 `normalizeRoute` 后的 `activeRouteName` 和 TopBar 一级导航为准，不以旧 hash 或旧画板标签为准。
11. 已清理或 dead code 化的废弃组件不得重新驱动 `docs/ui-design` 的目标设计、导航或用户流程。
12. 简历模块的当前目标以 `screen-resume-workshop.jsx` 为准：平铺简历列表、上传 / 粘贴创建、解析预览确认、简历详情中的预览 / 改写建议 / 手动编辑。改写建议每条仅有 `采纳`（无逐条拒绝 / 编辑）；采纳后确认前预览整份结果，由用户选择覆盖原简历或保存为新简历；原件预览、导出、复制和保存都必须有可见反馈。
13. `带提示练习` / `严格模拟` 是 `PracticeScreen` 内的辅助程度开关，不是面试前模式卡片；严格模拟必须隐藏提示、实时观察、可调用经历和语音现场提示。
14. `结束并生成报告` 必须保持为面试页右侧底部固定动作，并把 `practiceMode`、`hintUsed` 和 `hintCount` 传入报告上下文。
15. `debrief` 的文本添加和语音添加必须共享同一份 `entries` 列表；语音提取卡片确认后才写入正式复盘记录，并保留来源标记。
16. `parse` 是 JD 解析确认与启动页：核对解析结果、绑定简历、确认轮次、立即面试在同一页完成；首次导入链路在 `parse` 与 session 之间不得插入第二个全页确认。
17. `workspace` 是回访枢纽：服务最近面试回访、切换/新建规划与再次发起面试，不得恢复为首次导入的必经确认页。
18. `debrief` 上下文选择遵循“选一带二”：任一上下文选定后自动带出可推导项，标注“已自动带入”且可逐项更换；选择仍全部在复盘页弹窗内完成。
19. 认证只有邮箱验证码一种登录方式：不得保留或新增 `auth_reset` 目标 route、独立重置登录页或“忘记密码 / 密码 / 两步验证”口径；验证码重发与更换邮箱在 `auth_verify` 内完成。
20. 报告页只有 Header 一对 `复练当前轮 / 进入下一轮` CTA；复练计划详情与题目回顾不得重复出现开练按钮，题目回顾的`加入本轮复练`只是计划标记动作。
21. 主题菜单为四个预设 + 自定义 accent，默认主题为深海；设置页只保留个人资料与隐私数据两个 tab，不得恢复通知或订阅占位。
