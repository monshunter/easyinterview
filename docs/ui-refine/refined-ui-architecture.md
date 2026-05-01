# EasyInterview UI 目标总体架构

> **版本**: 1.5
> **状态**: active
> **更新日期**: 2026-05-01

## 1 文档目的

本文档定义当前阶段静态 UI 原型对应的目标信息架构。目标不是复刻旧版所有页面，而是按用户任务组织页面、导航和模块边界，并保持与 `easyinterview-ui/EasyInterview.html` + `easyinterview-ui/src/app.jsx` 当前运行时交互一致。

## 2 已确认决策

1. App 默认进入首页，不再有未登录欢迎页或登录前置页。
2. 顶部导航为：`首页`、`岗位推荐`、`模拟面试`、`简历`、`真实复盘`。
3. `当前岗位` 不再作为一级模块；用户从 `模拟面试` 进入当前面试规划，规划由 `JD/目标岗位 + 简历 + 当前面试轮次` 组成。
4. `面试报告` 不再作为一级模块；报告隶属于某一次已完成的模拟面试，只能从面试结束页、会话历史或相关入口进入。
5. `岗位推荐` 是一级模块，用于基于简历、用户画像和岗位偏好推荐 JD；点击 `确认面试` 后先进入 JD 解析确认 `parse`，再进入模拟面试前确认 `workspace`。
6. 简历是一级模块，管理原始简历、结构化主版本、岗位定制版本和导入/创建流程。
7. 用户画像在用户菜单中，是 AI 根据简历、JD、模拟面试、真实复盘和用户修正沉淀出的结构化资料；个人设置只保留账号基础信息和隐私安全。
8. 面试是一个整体过程，不提供“热身、单题、反问、针对性复练”等练习模式选择。
9. 文本面试和语音面试是整场面试的形式；文本输入框里的麦克风是“语音转文字”，不是切换到语音面试。
10. 面试规划页顶部的线条节点展示当前目标岗位的真实面试轮次进度，例如 HR 初筛、技术一面、技术二面、经理面，而不是模拟面试内部题目流程。
11. 报告只有一种运行时形态：仪表盘。`reportLayout` hash 参数和设计画板上的 Editorial / Timeline 标签当前不改变 `ReportScreen`，时间线和独立刊物式报告页不进入当前目标。
12. 报告必须显式区分 `复练当前轮` 与 `进入下一轮`：前者重复当前轮并带入报告中的问题，后者创建下一轮面试规划。
13. 真实复盘是一级模块，按 `复盘记录 -> 复盘分析 -> 复盘面试` 递进，用于真实面试而非模拟报告错题本。
14. 公司情报不是一级导航；当前 UI 在模拟面试规划页内展示轻量嵌入卡片，并可打开 `company_intel` 详情页后返回面试前确认。
15. `welcome`、`growth`、`plan`、`mistakes`、`drill`、`followup`、`experiences`、`star` 等旧路由不代表当前目标模块；运行时通过 `routeAliases` 折回 `home`、`workspace`、`report`、`practice` 或 `resume_versions`。
16. 成长、多轮计划、经历库、追问树、单题 Drill、独立错题复练队列仍不进入当前主流程。

## 3 目标产品骨架

```text
[EasyInterview App]
├─ Home / 首页
│  ├─ JD 粘贴输入
│  ├─ JD 文件 / URL 导入弹窗
│  ├─ 还没有简历？1 分钟创建
│  ├─ 最近模拟面试列表
│  │  └─ 每张卡片展示目标岗位和面试轮次节点
│  ├─ 岗位推荐辅助入口
│  └─ 真实复盘辅助入口
├─ Job Picks / 岗位推荐
│  ├─ 用户画像摘要
│  ├─ 推荐 JD 列表
│  ├─ 匹配原因 / 不匹配原因
│  └─ 确认面试 -> JD 解析确认 -> 模拟面试规划
├─ Mock Interview / 模拟面试
│  ├─ 当前面试规划
│  │  ├─ TargetJob / JD
│  │  ├─ ResumeVersion
│  │  └─ InterviewRound
│  ├─ 切换规划 / 新建规划
│  ├─ 更换简历弹窗
│  ├─ 面试轮次节点
│  ├─ JD 拆解 / 公司情报轻量卡片 / 我的准备
│  │  └─ 打开公司情报详情(company_intel)
│  ├─ 右侧模拟面试历史
│  └─ 立即面试
├─ Interview Session(sessionId)
│  ├─ 面试形式: 文本面试 / 语音面试
│  ├─ 文本回答
│  │  └─ 语音转文字填入输入框
│  ├─ 实时语音面试
│  ├─ 问题推进
│  └─ 结束并生成报告
├─ Report Dashboard(sessionId)
│  ├─ 会话 / 岗位 / 简历 / 轮次 / 形式上下文
│  ├─ 准备度详情
│  ├─ 维度详情
│  ├─ 题目回顾页
│  ├─ 证据详情
│  ├─ 复练当前轮
│  └─ 进入下一轮
├─ Resume / 简历
│  ├─ 原始简历
│  ├─ 结构化主版本
│  ├─ 岗位定制版本
│  ├─ 上传 / 粘贴 / 轻量问答创建
│  └─ 原始简历预览
├─ Debrief / 真实复盘
│  ├─ 选择目标岗位 / JD
│  ├─ 选择关联模拟面试
│  ├─ 选择绑定简历
│  ├─ 复盘记录
│  ├─ 复盘分析
│  └─ 复盘面试
└─ User Menu
   ├─ 登录 / 注册(未登录)
   ├─ 用户画像(已登录)
   ├─ 设置与隐私(已登录)
   └─ 退出登录(已登录)
```

## 4 顶部导航

### 4.1 当前导航结构

```text
[Top Navigation]
├─ 首页
├─ 岗位推荐
├─ 模拟面试
├─ 简历
├─ 真实复盘
└─ 用户区
   ├─ 未登录: 登录 / 注册
   └─ 已登录:
      ├─ 用户画像
      ├─ 设置与隐私
      └─ 退出登录
```

### 4.2 不进入顶部导航的能力

```text
不作为一级导航:
├─ 当前岗位 -> 并入模拟面试的当前面试规划
├─ 面试报告 -> 隶属于一次模拟面试 session
├─ 公司情报 -> 模拟面试规划页轻量卡片和详情页
├─ 简历原件预览 -> 简历模块内弹窗
├─ 用户画像 -> 用户菜单入口
├─ 设置 -> 用户菜单入口
├─ 登录 / 注册 / 验证 / 重置 / 退出 -> 用户区或认证页
├─ 成长 -> 移除
├─ 多轮计划 -> 移除
├─ 经历库 -> 移除
├─ 追问树 -> 移除
├─ 针对性复练 / 单题 Drill -> 移除为独立流程
└─ 报告时间线 / 刊物式报告页 -> 移除为独立形态
```

## 5 目标模块关系

```text
Home
├─ 粘贴 / 上传 / URL 导入 JD
│  └─ Parse & Confirm Interview
│     └─ Mock Interview Plan
├─ 最近模拟面试
│  └─ Mock Interview Plan(jobId, resumeVersionId, round)
├─ 1 分钟创建简历
│  └─ Resume Intake
├─ 岗位推荐
│  └─ Job Picks
└─ 真实面试复盘
   └─ Debrief

Job Picks
└─ 选择推荐 JD
   └─ Parse & Confirm Interview
      └─ Mock Interview Plan

Mock Interview Plan
├─ 绑定 TargetJob / JD
├─ 绑定 ResumeVersion
├─ 确认 InterviewRound
├─ 发起 InterviewSession
└─ 查看当前岗位下的模拟面试历史
   └─ Report Dashboard(sessionId)

Report Dashboard(sessionId)
├─ 题目回顾和证据详情
├─ 复练当前轮
│  └─ Mock Interview Plan(same round)
└─ 进入下一轮
   └─ Mock Interview Plan(next round)

Resume
├─ 原始简历
├─ 结构化主版本
└─ 岗位定制版本

Debrief
└─ 目标岗位 + 关联模拟面试 + 简历
   └─ 复盘记录 -> 复盘分析 -> 复盘面试
```

## 6 页面层级规则

### 6.1 一级页面

```text
Home
JobPicks
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
VoiceInterview(sessionId)
CompanyIntelDetail
```

这些页面不作为顶部导航入口，只能从具体上下文进入。

### 6.3 用户菜单页面

```text
UserProfile
SettingsPrivacy
AuthLogin
AuthRegister
AuthVerify
AuthReset
AuthLogout
```

用户画像不是个人设置。个人设置只保存账号身份、登录安全、隐私与偏好；画像保存系统推断、证据来源和用户纠偏层。

## 7 后续实现输入

1. App 默认路由进入 `home`，不进入 `welcome`。
2. 顶部导航不得恢复 `当前岗位` 或 `面试报告` 一级入口。
3. `workspace` 的产品语义是 `MockInterviewPlan`，即当前模拟面试规划，不是独立岗位工作台。
4. `workspace` 必须允许切换规划和新建规划；如果用户不想继续当前规划，可以从这里切到新的 `JD + 简历 + 轮次` 组合。
5. `workspace` 的主 CTA 文案是 `立即面试`。
6. `workspace` 的流程线展示面试轮次，不展示模拟面试内部题目流程。
7. 更换简历应打开选择简历弹窗，不直接跳转到简历模块。
8. 面试页的顶部 `语音面试` 是整场面试形式；输入框中的麦克风是语音转文字。
9. 报告必须显示所属会话、目标岗位、绑定简历、面试轮次、完成时间和沟通形式。
10. 报告的后续动作必须拆成 `复练当前轮` 和 `进入下一轮` 两条路径。
11. 简历模块必须保留原始简历预览和解析文本快照，结构化或岗位定制不能覆盖原件。
12. 真实复盘必须先确认目标岗位、关联模拟面试和简历，再进入复盘记录。
13. `routeAliases` 是当前静态 UI 的兼容层：旧 hash 路由可渲染，但不得据此恢复旧导航或旧模块。
14. `easyinterview-canvas.html` 的旧分区标题和报告变体标签不改变当前目标架构；文档以 `app.jsx` 实际渲染为准。
