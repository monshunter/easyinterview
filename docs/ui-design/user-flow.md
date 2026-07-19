# EasyInterview 目标用户流程

> **版本**: 2.34
> **状态**: active
> **更新日期**: 2026-07-19

## 1 文档目的

本文档定义当前目标用户流程。当前流程按 D-22 收敛为三条主线：首页启动、面试、简历管理；报告后的下一步只保留复练当前轮和进入下一轮。

## 2 主流程总览

```text
Home
  -> JD 导入
  -> Parse command progress（queued / processing）
  -> ready replace Workspace Plan Detail(targetJobId)
     └─ 面试报告 -> Reports(current target)
  -> Interview Session
  -> Generating
  -> Report Dashboard
     ├─ 复练当前轮
     └─ 进入下一轮

Resume
  -> 上传 / 粘贴创建
  -> 直接打开简历详情
  -> 只读简历详情
  -> Home 新建规划时显式绑定；Workspace detail 只读展示

Interview
  -> 浏览面试规划列表
  -> 回访当前面试规划
  -> 查看当前规划报告
  -> 立即面试
```

## 3 首页启动流程

```text
Home
├─ 粘贴 JD
└─ 还没有简历？1 分钟创建
```

首页不提供复盘辅助入口。用户有 JD 且已显式选择 selectable 简历时才能创建 TargetJob 并进入解析进度；selectable 指未归档且 `parseStatus=ready` 或已有可读正文/结构化证据。用户没有该类简历时只能进入简历创建，不能提交 JD 或进入无简历训练/报告降级链路。

## 4 JD 解析进度、规划详情与启动

```text
Parse Command Progress(targetJobId)
├─ queued / processing 四步进度
├─ failed / timeout 恢复
└─ ready -> replace Workspace Plan Detail(targetJobId)

Workspace Plan Detail(targetJobId)
├─ 标题旁 绑定简历 -> Resume Detail(resumeId)
├─ 首行动作行 立即面试 + 面试报告 -> Reports(targetJobId)
├─ 核对 JD 基础信息
├─ 核对必需项 / 加分项 / 隐性关注点
└─ 确认 InterviewRound
```

首次导入链路只有一个 ready 详情母版。Parse 只展示新导入的 queued/processing 命令进度；解析成功即代表规划已保存并 replace 到 `/workspace?targetJobId=...` 只读详情。既有 ready 规划回访也直接进入同一 Workspace 详情，不播放 Parse 动画。

## 5 面试规划回访

```text
Workspace
├─ 面试规划列表（无上下文一级入口）
│  ├─ 打开已有规划
│  └─ 从新 JD 创建规划 -> Home
└─ 当前面试规划（只带 targetJobId 上下文）
   ├─ JD / 简历 / 轮次
   ├─ 公司轻情报卡片
   ├─ 当前规划会话记录
   └─ 立即面试
```

Workspace 服务回访、切换规划和再次发起面试，不是泛岗位资产管理中心。

## 6 面试与报告

```text
Practice
├─ 连续文本聊天
├─ 电话入口置灰
├─ 用户提交后立即显示自己的消息
├─ pending：锁输入 + 面试官思考
├─ retryable failure：原消息下同 ID retry
├─ refresh：从 server reply state 恢复
└─ 结束并生成报告

Report Dashboard
├─ 准备度
├─ 维度详情
├─ 证据详情
├─ 复练当前轮
└─ 进入下一轮

Reports(targetJobId)
├─ 只显示当前规划 canonical rounds
├─ 当前可用报告 -> Report Dashboard
├─ 最新生成中 -> Generating
├─ 最新生成失败 -> 同 report 重新生成 + 查看面试记录
├─ `REPORT_CONTEXT_TOO_LARGE` -> 仅查看面试记录 / 返回
├─ 空 / 加载 / 请求失败
└─ 返回 -> `/workspace?targetJobId` current plan detail
```

报告详情必须隶属于 session；Reports 只是当前规划范围的索引，不是全局中心或第二种报告内容形态。`复练当前轮` 和 `进入下一轮` 是唯一一对开练 CTA。Report/Generating 有可信 target context 时返回 Reports，无可信上下文时安全返回 Workspace。

## 7 简历流程

```text
Resume
├─ 平铺简历列表
├─ 新建简历
│  ├─ 上传文件
│  ├─ 粘贴文本
│  └─ 注册成功后直接打开详情
└─ 简历详情
   └─ 只读原始正文
```

简历创建不提供轻量问答建档，不展示解析动画或结构化草稿确认页。解析成功后，backend parse 负责根据 LLM `displayName` / 结构化结果生成可识别 `displayName`；若解析失败但可读正文已抽取，backend 必须写入非通用 fallback 名称。解析前或旧数据不得从原文第一行或文件名派生可见名称，只能显示中性 fallback 名称、来源信息或 LLM/结构化 headline。详情页只阅读原始简历正文，不提供导出、复制、编辑、改写或原件弹层。

## 8 认证与设置

```text
登录触发点
├─ 立即面试
├─ 打开当前规划报告深链
├─ 复练当前轮
├─ 进入下一轮
├─ 保存简历
└─ 设置

Auth
├─ auth_login
├─ auth_verify
├─ auth_profile_setup
└─ auth_logout

已登录 TopBar
└─ 设置齿轮 -> Settings
   ├─ Appearance: 账号级主题本地预览 / 单次保存
   ├─ Account: 姓名 / 完整账号邮箱 / 退出登录
   └─ Privacy: 导出暂不可用 / 删除账号二次确认
```

`auth_profile_setup` 是首次账号资料补全，用于登录后接续 pendingAction 前的账号基础信息确认，不是用户画像。Settings 为无 tab 的受保护单页，复用 bootstrap runtime `/me` 字段；页面切换不重复获取 `/me`，主题拖动只本地预览，保存只发送一次账号更新。页面不为手机号、界面语言、时区、登录安全、字体预设或产品信息制造静态占位。

## 9 范围外流程

以下流程不属于目标用户流程或正式前端入口：

- 真实面试复盘
- 复盘面试
- 用户画像查看和纠偏
- CandidateProfile / ExperienceCard 维护
- 岗位推荐
- 公司情报独立详情页
- 独立成长中心
- 独立错题本 / 单题 Drill
- 独立 Voice 页面

## 10 异常和边界

| 情况 | 处理 |
|------|------|
| 用户没有 JD | 首页提示在唯一文本框粘贴 JD |
| 用户没有 selectable 简历 | 首页只提供创建简历入口，未形成可读证据并显式选择前不调用 import；历史 Workspace 缺失/无效绑定属于异常数据，Start、Reports、复练和下一轮全部 fail closed，不提供 rebind 或 fallback |
| 用户未登录执行写入动作 | 进入邮箱验证码登录，成功后接续 pendingAction |
| Practice 请求等待 / 失败 / 刷新 | pending 锁输入并显示思考；只有 server retryable failure 在原 row 显示 retry；刷新通过 `getPracticeSession` 恢复原 `clientMessageId/replyStatus` |
| `/reports` 缺失/非法/无权 targetJobId | 不展示其他规划或 stale rows，提供安全返回 Workspace；未登录时只用合法 targetJobId 接续鉴权 |
| URL fragment 携带 `route` / 业务参数 | fragment 不参与 routing，按 canonical path/query 解析并在首次 replace 时移除 |
| 范围外 `auth_reset` | 归一到 `auth_login` |

四条长耗时流程使用一致的过渡体验，但不共享业务进度事实：

1. 开始面试：有效启动请求发出后立即进入全屏准备场景，保留 TopBar，成功进入 Practice，失败回到原入口错误。
2. 注册简历后：详情 route 在 `queued/processing` 持续显示解析场景并提供统一“返回 / Back”控件，目标仍为简历工坊；轮询不得闪回通用 loading。
3. 完成面试后：Generating 保留 TopBar 与统一“返回 / Back”控件，目标仍由可信报告上下文决定；只展示 API `queued/generating` 事实和 indeterminate 进度，ready 自动进入报告。
4. 导入 JD 后：Parse 以四项步骤列表表达当前客户端等待节奏，ready 后 replace 到 Workspace 详情；不得展示 provider、prompt、rubric、耗时或百分比。

## 11 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-20 | 2.34 | 统一长耗时流程与二三级页面返回控件文案为“返回 / Back”，不改变目标路由。 |
| 2026-07-19 | 2.33 | 将进入面试、简历解析、报告生成和 JD 解析统一为共享过渡体验，同时保留各自真实状态、返回路径与无假进度边界。 |
| 2026-07-19 | 2.32 | 设置入口统一更名为“设置”，新增账号级主题保存流程与页面切换零重复 `/me` 读取约束。 |
| 2026-07-16 | 2.31 | Reports 增加已结束会话的 failed report 同 ID 重生成与只读面试记录恢复，超限失败保持不可重试。 |
| 2026-07-15 | 2.30 | 将 selectable 简历设为 JD import、训练、报告及报告后动作的永久强制前置；历史缺绑规划统一 fail closed。 |
| 2026-07-15 | 2.29 | 采用设置简化方案 A：已登录 TopBar 只保留设置齿轮，Settings 只承接真实账号字段、退出、导出不可用与账号删除。 |
| 2026-07-15 | 2.28 | Workspace 详情将绑定简历改为标题旁详情链接，并把立即面试/面试报告合并为左对齐首行动作行。 |
| 2026-07-14 | 2.27 | 将 Parse 收窄为新导入 queued/processing 进度；ready replace 与既有规划、Reports Back 均进入 targetJobId-only Workspace 详情。 |
| 2026-07-14 | 2.26 | 增加从规划详情进入 target-scoped Reports 的流程，并锁定 current/latest-only、Back 路径、鉴权与非全局入口边界。 |
| 2026-07-13 | 2.25 | Practice 增加即时 user row、pending thinking/输入锁、failed-row retry 与服务端 reply-state 刷新恢复。 |
