# EasyInterview UI 目标总体架构

> **版本**: 2.11
> **状态**: active
> **更新日期**: 2026-06-12

## 1 文档目的

本文档定义当前阶段静态 UI 原型对应的目标信息架构。目标不是复刻旧版所有页面，而是按用户任务组织页面、导航和模块边界，并保持与 `ui-design/index.html` + `ui-design/src/app.jsx` 当前运行时交互一致。

## 2 已确认决策

1. App 默认进入首页，不再有未登录欢迎页或登录前置页。
2. 顶部导航为：`首页`、`模拟面试`、`简历`、`复盘`；岗位推荐一级模块已删除（D-17）。
3. 顶栏固定提供全局显示控制：主题色菜单、暗色模式、语言下拉；这些控制不属于业务模块，也不要求登录。
4. 主题色与 dark/light 模式正交：当前运行时主题为 `暖陶 / 苔林 / 深海 / 梅子 / 自定义 accent`，默认主题为 `深海`，暗色按钮只切换所选主题的明暗版本（D-21 v2.1：自定义 accent 按用户决策保留）。
5. 字体预设在 `设置与隐私` 的个人资料页维护，当前有 `编辑级 / 现代 / 杂志` 三组；切换时必须原子更新标题 serif 与正文 sans，等宽字体保持不变。
6. `当前岗位` 不再作为一级模块；用户从 `模拟面试` 进入当前面试规划，规划由 `JD/目标岗位 + 简历 + 当前面试轮次` 组成。规划页是回访枢纽：服务最近面试回访、切换/新建规划与再次发起面试；首次导入新 JD 的启动决策在 `parse` 完成。
7. `面试报告` 不再作为一级模块；报告隶属于某一次已完成的模拟面试，只能从面试结束页、会话历史或相关入口进入。
8. 岗位推荐模块已整体删除：`jd_match` 不再是目标 route（旧 hash 归一回 `home`），JD 获取唯一入口是首页导入（粘贴 / 上传 / URL）；首次导入链路不经过 `workspace` 第二次全页确认。
9. 简历是一级模块，当前运行时以 `screen-resume-workshop.jsx` 为准：平铺简历列表、上传 / 粘贴创建、解析预览确认和简历详情；不存在版本树 / 主版本 / 岗位定制版本 / 分叉流程（D-20）。改写建议每条仅有`采纳`；采纳后确认前预览整份结果，由用户选择覆盖原简历或保存为新简历。
10. 用户画像在用户菜单中，是 AI 根据简历、JD、模拟面试、复盘和用户修正沉淀出的结构化资料；个人设置只保留账号基础信息、界面偏好、隐私数据和登录安全。
11. 面试是一个整体过程，不提供“热身、单题、反问、针对性复练”等练习模式选择。
12. 文本面试和语音面试是整场面试的形式；二者共享 `PracticeScreen` 顶部控制、题目地图和会话上下文，文本输入框里的麦克风是“语音转文字”，不是切换到语音面试。
13. 面试会话内存在 `带提示练习` / `严格模拟` 辅助程度开关；它只控制提示、实时观察、可调用经历和语音现场提示的显隐，不恢复面试前练习模式卡片。
14. `结束并生成报告` 是面试页右侧底部固定动作；生成报告时必须带上 `practiceMode`、`hintUsed` 和 `hintCount`，供报告展示练习方式和提示记录。
15. 面试规划页顶部的线条节点展示当前目标岗位的真实面试轮次进度，例如 HR 初筛、技术一面、技术二面、经理面，而不是模拟面试内部题目流程。
16. 报告只有一种运行时形态：仪表盘。旧 `reportLayout` hash 参数、设计画板上的 Editorial / Timeline 标签以及 `ReportEditorial` / `ReportTimeline` 组件已从当前静态 UI 清理；时间线和独立刊物式报告页不进入当前目标。
17. 报告必须显式区分 `复练当前轮` 与 `进入下一轮`：前者直接进入当前轮复练 session 并带入报告中的问题，后者直接进入下一轮面试 session。这对 CTA 只在报告 Header 出现一次（D-19）；复练计划详情只承载路径说明与复练清单，题目回顾的`加入本轮复练`是计划标记动作。
18. 复盘是一级模块，按 `复盘记录 -> 复盘分析 -> 复盘面试` 递进，用于真实面试而非模拟报告错题本。
19. 复盘记录的 `文本` 和 `语音` 是同一份记录的两种添加方式；顶部汇总条、来源标记和问题卡片列表跨模式共享。
20. 语音复盘由 AI 连续对话引导，支持 [空格] 暂停、实时提取待确认卡片、手动加一条、确认 / 编辑 / 删除 / 写入记录；切换回文本不得丢失语音会话状态。
21. 公司情报不是一级导航；当前 UI 只在模拟面试规划页内展示轻量嵌入卡片。`company_intel` 独立详情页与 route 已删除（D-18），旧 hash 归一回 `workspace`。
22. `welcome`、`growth`、`plan`、`mistakes`、`drill`、`followup`、`experiences`、`star`、`resume`、`onboarding`、`jd_match`、`company_intel` 等旧路由不代表独立目标模块；运行时通过 `ROUTE_ALIASES` 归一到 `home`、`workspace`、`report`、`practice` 或 `resume_versions`。`voice` 不再保留 route alias；语音面试只能通过 `practice` 显式携带 `mode=voice` 与 `modality=voice` 进入，不据此恢复独立语音页面骨架。
23. 成长、多轮计划、经历库、追问树、单题 Drill、独立错题复练队列仍不进入当前主流程。
24. 模拟面试规划页的历史列表只展示当前 `MockInterviewPlan` / `TargetJob` / `JD` 范围内的会话，不混入其他公司、岗位或 JD 的历史。
25. 复盘页的 `目标岗位 / JD`、`关联模拟面试`、`绑定简历` 三个上下文卡片都通过本页弹窗选择，不跳转到模拟面试、报告或简历页面。上下文遵循“选一带二”：任一上下文选定后自动带出可推导项，标注“已自动带入”且可逐项更换。
26. 旧 `screens-p1-depth.jsx::ResumeVersionsScreen`（`_LegacyResumeVersionsScreen`）dead code 已删除；`ui-design/index.html` 后加载 `screen-resume-workshop.jsx` 并覆盖 `window.ResumeVersionsScreen`，平铺简历工坊是唯一实现。
27. 路由层维护 `InterviewContext`，在 `parse`、`workspace`、`practice`、`generating`、`report`、`debrief` 间贯通 `planId / targetJobId / jdId / resumeId / roundId / sessionId`；`parse` 立即面试时直接产出完整上下文进入 `practice`。
28. 登录打断使用 `pendingAction` 恢复原动作；当前静态稿已覆盖立即面试（`parse` 与 `workspace` 两处入口）、复练当前轮和进入下一轮。
29. TopBar 品牌区只保留 `E` mark 与 `EasyInterview`；`EasyInterview` 作为产品名不翻译，解释性定位文案和版本号不在 TopBar 常驻展示。版本号作为产品元数据放在设置页产品信息区。

## 3 目标产品骨架

```text
[EasyInterview App]
├─ TopBar / 全局控制
│  ├─ Brand: E mark + EasyInterview
│  ├─ 主题色: 暖陶 / 苔林 / 深海（默认） / 梅子 / 自定义
│  ├─ Dark / Light
│  ├─ 语言下拉: Globe icon + 当前语言标签
│  └─ 用户区
├─ Home / 首页
│  ├─ JD 粘贴输入
│  ├─ JD 文件 / URL 导入弹窗
│  ├─ 还没有简历？1 分钟创建
│  ├─ 最近模拟面试列表
│  │  └─ 每张卡片展示目标岗位和面试轮次节点
│  └─ 复盘辅助入口
├─ Mock Interview / 模拟面试
│  ├─ 当前面试规划
│  │  ├─ TargetJob / JD
│  │  ├─ 绑定简历
│  │  └─ InterviewRound
│  ├─ 切换规划 / 新建规划
│  ├─ 更换简历弹窗
│  ├─ 面试轮次节点
│  ├─ JD 拆解 / 公司情报嵌入卡片 / 我的准备
│  ├─ 右侧当前规划的模拟面试历史
│  └─ 立即面试
├─ Interview Session(sessionId)
│  ├─ 面试形式: 文本面试 / 语音面试
│  ├─ 练习辅助: 带提示练习 / 严格模拟
│  ├─ 文本回答
│  │  └─ 语音转文字填入输入框
│  ├─ 实时语音面试
│  │  ├─ 正在说话波形
│  │  ├─ 本次回答标注波形
│  │  ├─ 实时转写
│  │  └─ 表达层指标
│  ├─ 问题推进
│  └─ 右侧底部固定结束并生成报告
├─ Report Dashboard(sessionId)
│  ├─ 会话 / 岗位 / 简历 / 轮次 / 形式上下文
│  ├─ 准备度详情
│  ├─ 维度详情
│  ├─ 题目回顾页
│  ├─ 证据详情
│  ├─ 复练当前轮
│  └─ 进入下一轮
├─ Resume / 简历
│  ├─ 简历工坊列表（平铺）
│  ├─ 新建简历
│  │  ├─ 上传 / 粘贴
│  │  ├─ Agent 解析
│  │  └─ 预览确认保存
│  └─ 简历详情
│     ├─ 预览
│     ├─ 改写建议（仅采纳 -> 预览 -> 覆盖 / 另存）
│     └─ 手动编辑
├─ Debrief / 复盘
│  ├─ 目标岗位 / JD 选择弹窗
│  ├─ 关联模拟面试选择弹窗
│  ├─ 绑定简历选择弹窗
│  ├─ 复盘记录
│  │  ├─ 文本添加
│  │  ├─ 语音添加
│  │  └─ 共享问题卡片
│  ├─ 复盘分析
│  └─ 复盘面试
└─ User Menu
   ├─ 登录(未登录)
   ├─ 用户画像(已登录)
   ├─ 设置与隐私(已登录)
   └─ 退出登录(已登录)
```

## 4 顶部导航

### 4.1 当前导航结构

```text
[Top Navigation]
├─ 首页
├─ 模拟面试
├─ 简历
├─ 复盘
├─ 主题色菜单
├─ 暗色模式
├─ 语言下拉
└─ 用户区
   ├─ 未登录: 登录
   └─ 已登录:
      ├─ 用户画像
      ├─ 设置与隐私
      └─ 退出登录
```

### 4.2 不进入顶部导航的能力

```text
不作为一级导航:
├─ 岗位推荐 -> 移除（D-17）
├─ 当前岗位 -> 并入模拟面试的当前面试规划
├─ 面试报告 -> 隶属于一次模拟面试 session
├─ 公司情报 -> 模拟面试规划页嵌入卡片（独立详情页已删除）
├─ 简历原件预览 -> 简历模块内弹窗
├─ 用户画像 -> 用户菜单入口
├─ 设置 -> 用户菜单入口
├─ 登录 / 验证 / 首次资料补全 / 退出 -> 用户区或认证页
├─ 成长 -> 移除
├─ 多轮计划 -> 移除
├─ 经历库 -> 移除
├─ 追问树 -> 移除
├─ 针对性复练 / 单题 Drill -> 移除为独立流程
└─ 报告时间线 / 刊物式报告页 -> 移除为独立形态
```

### 4.3 全局显示控制

```text
TopBar Display Controls
├─ Theme menu
│  ├─ 暖陶
│  ├─ 苔林
│  ├─ 深海（默认）
│  ├─ 梅子
│  └─ 自定义 accent
│     ├─ 色相
│     ├─ 饱和度
│     └─ 恢复主题默认色
├─ Dark / Light toggle
└─ Language dropdown
   ├─ 中文
   └─ English

Settings -> 个人资料 -> 界面偏好
└─ Font preset
   ├─ 编辑级: Noto Serif SC + Inter
   ├─ 现代: Source Serif Pro + Geist
   └─ 杂志: Cormorant Garamond + IBM Plex Sans
```

这些控制影响 UI 呈现，不改变当前业务路由、模块归属或认证状态。
语言下拉必须按 locale 元数据渲染选项；当前静态原型在 `ui-design/src/app.jsx`
通过 `LANGUAGE_OPTIONS` 维护 `key`、展示标签、短标签和别名，按钮只显示 globe icon
与当前语言标签（如 `中文` / `English`），后续新增语言时扩展该列表而不是把
TopBar 改回二选一 toggle 或把多个候选语言拼在按钮上。Locale 初始化优先级为
用户显式选择（`localStorage["ei-lang"]`）> 浏览器 locale > English fallback。

## 5 目标模块关系

```text
Home
├─ 粘贴 / 上传 / URL 导入 JD
│  └─ Parse & Confirm Interview
│     └─ Mock Interview Plan
├─ 最近模拟面试
│  └─ Mock Interview Plan(jobId, resumeId, round)
├─ 1 分钟创建简历
│  └─ Resume Intake
└─ 复盘
   └─ Debrief

Mock Interview Plan
├─ 绑定 TargetJob / JD
├─ 绑定简历
├─ 确认 InterviewRound
├─ 发起 InterviewSession
└─ 查看当前面试规划下的模拟面试历史
   └─ Report Dashboard(sessionId)

Report Dashboard(sessionId)
├─ 题目回顾和证据详情
├─ 复练当前轮
│  └─ Interview Session(same round)
└─ 进入下一轮
   └─ Interview Session(next round)

Resume
├─ Flat Resume List
├─ Create Flow
│  └─ 上传 / 粘贴 -> Agent 解析 -> 预览确认 -> 保存
└─ Resume Detail
   ├─ Preview
   ├─ Rewrites（仅采纳 -> 预览 -> 覆盖 / 另存）
   └─ Edit

Debrief
└─ 目标岗位 + 关联模拟面试 + 简历
   ├─ 点击上下文卡片打开对应选择弹窗
   └─ 复盘记录 -> 复盘分析 -> 复盘面试
```

## 6 页面层级规则

### 6.1 一级页面

```text
Home
MockInterviewPlan
Resume
Debrief
```

### 6.2 上下文 / 会话级页面

```text
ParseAndConfirmInterview
InterviewSession(sessionId)
ReportGenerating(sessionId)
ReportDashboard(sessionId)
```

这些页面不作为顶部导航入口，只能从具体上下文进入。

### 6.3 用户菜单页面

```text
UserProfile
SettingsPrivacy
SettingsInterfacePreferences
AuthLogin
AuthVerify
AuthProfileSetup
AuthLogout
```

用户画像不是个人设置。个人设置只保存账号身份、登录安全、隐私与偏好；画像保存系统推断、证据来源和用户纠偏层。

### 6.4 历史 / 废弃代码

```text
Historical prototype routes normalized only for the static UI
├─ ROUTE_ALIASES 已折返:
│  ├─ welcome -> home
│  ├─ growth -> home
│  ├─ plan -> workspace
│  ├─ mistakes -> report
│  ├─ drill -> practice
│  ├─ followup -> practice
│  ├─ experiences -> resume_versions
│  ├─ star -> resume_versions
│  ├─ resume -> resume_versions
│  ├─ onboarding -> resume_versions
│  ├─ auth_register -> auth_login
│  ├─ auth_reset -> auth_login
│  ├─ jd_match -> home
│  └─ company_intel -> workspace
├─ voice route alias 已删除:
│  └─ 语音面试必须使用 practice(mode=voice, modality=voice)
└─ 已从当前静态 UI 源码清理:
   ├─ screens-rest.jsx
   ├─ screens-completion.jsx
   ├─ screens-p2.jsx
   ├─ screen-jd-match.jsx / JDMatchScreen
   ├─ CompanyIntelScreen（仅保留 CompanyIntelEmbed）
   ├─ PlanScreen
   ├─ VoicePracticeScreen
   ├─ ExperienceLibraryScreen
   ├─ Legacy ResumeVersionsScreen（dead code 已删除）
   ├─ ResumeTreeView / ResumeBranchFlow
   ├─ OnboardingScreen
   ├─ ReportEditorial / ReportTimeline
   ├─ IssueRow / PerQBlock / KVInline
   ├─ SettingsNotif / SettingsBilling
   ├─ AuthResetScreen
   └─ ThankYouLetter
```

判断目标架构时不得以旧 hash、旧画板标签或已删除组件为准。目标页面必须同时符合当前顶部导航、`normalizeRoute` 后的 `activeRouteName` 和实际渲染内容。

## 7 后续实现输入

1. App 默认路由进入 `home`，不进入 `welcome`。
2. 顶部导航不得恢复 `岗位推荐`、`当前岗位` 或 `面试报告` 一级入口。
3. `workspace` 的产品语义是 `MockInterviewPlan`，即当前模拟面试规划，不是独立岗位工作台；定位是回访枢纽，不得作为首次导入链路中 `parse` 与 session 之间的必经确认页。
4. `workspace` 必须允许切换规划和新建规划；如果用户不想继续当前规划，可以从这里切到新的 `JD + 简历 + 轮次` 组合。
5. `workspace` 的主 CTA 文案是 `立即面试`。
6. `workspace` 的流程线展示面试轮次，不展示模拟面试内部题目流程。
7. 更换简历应打开选择简历弹窗，不直接跳转到简历模块。
8. 面试页的顶部 `语音面试` 是整场面试形式；切换后仍停留在 `PracticeScreen` 外层骨架内，只替换为语音波形、实时转写和表达指标 Surface；输入框中的麦克风是语音转文字。
9. `严格模拟` 是会话内辅助程度开关，必须隐藏提示、实时观察、可调用经历和语音现场提示；不得恢复面试前练习模式卡片。
10. `结束并生成报告` 必须保持为右侧底部固定动作，并传递 `practiceMode`、`hintUsed` 和 `hintCount`。
11. 报告必须显示 `sessionId`、目标岗位、绑定简历、面试轮次、沟通形式、练习方式和提示使用记录；无 `sessionId` 或生成失败时必须进入明确状态页。
12. 报告的后续动作必须拆成 `复练当前轮` 和 `进入下一轮` 两条路径；两者都直接进入对应面试 session，只有返回按钮回到模拟面试规划。这对 CTA 只在 Header 出现一次，复练计划详情和题目回顾不得重复开练按钮。
13. 简历模块必须保留原始来源预览和解析文本快照；结构化编辑和改写采纳不能覆盖原件快照。
14. 复盘必须先确认目标岗位、关联模拟面试和简历；三个上下文变更动作必须在当前复盘页打开选择弹窗，不直接跳转到其他一级或会话页。
15. 复盘记录必须把文本添加和语音添加写入同一组问题卡片，并保留来源标记；切换添加方式不得丢状态。
16. 语音复盘提取的问题必须先进入待确认卡片，用户确认后才写入复盘记录。
17. `ROUTE_ALIASES` 只归一除 `voice` 外的历史原型 hash route，不构成线上兼容承诺；旧 route 不得据此恢复旧导航或旧模块，语音面试必须使用 `practice` 显式参数。
18. `canvas.html` 不应保留旧分区标题、旧单页简历画板、旧 onboarding 画板或报告变体画板；文档以 `app.jsx` 实际渲染为准。
19. 顶栏主题色、暗色和语言下拉必须保持为横切显示控制，不进入任何业务模块。
20. 字体预设必须在设置页作为界面偏好维护，并原子切换 serif/sans 字体组合。
21. 简历模块目标实现以 `screen-resume-workshop.jsx` 为准；旧树形 / 版本工坊实现不得重新驱动文档、画板或运行时入口。
22. 简历创建必须经过解析进度和预览确认；创建输入只有上传文件和粘贴文本两种。
23. `screens` 映射不得重新注册 `welcome`、`mistakes`、`growth`、`plan`、`experiences`、`drill`、`followup`、`star`、`resume`、`onboarding`、`jd_match`、`company_intel` 等旧页面 key。
24. `ResumeScreen`、旧 `ResumeVersionsScreen`、`OnboardingScreen`、`ReportEditorial`、`ReportTimeline`、`PlanScreen`、`ExperienceLibraryScreen`、`JDMatchScreen`、`CompanyIntelScreen` 和旧 `DebriefScreen` 不应进入当前源码或目标文档的页面结构。
25. 模拟面试历史列表必须按当前面试规划 / 当前 TargetJob / JD 过滤，不展示其他公司或岗位的历史。
26. 改写建议每条仅提供 `采纳`；采纳后确认前预览整份结果，由用户选择覆盖原简历或保存为新简历；不得恢复逐条 `拒绝 / 编辑`。
27. 设置页只保留 `个人资料` 与 `隐私与数据` 两个 tab，不得恢复通知 / 订阅占位；主题菜单为四个预设 + 自定义 accent，默认主题为 `深海`。
