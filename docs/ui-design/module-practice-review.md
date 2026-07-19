# 模拟面试与报告模块

> **版本**: 1.35
> **状态**: active
> **更新日期**: 2026-07-19

## 1 目标

模拟面试采用连续文本聊天。系统不展示或维护题号、题目总数、当前题、追问/下一题分类、题目地图或专用提示；AI 根据 JD、简历、轮次和对话历史自然推进。用户主动结束后生成会话级报告。

## 2 Practice 页面结构

进入 `PracticeScreen` 前，Home 最近面试、Workspace 列表、Workspace 详情、Report 复练与下一轮入口统一展示全屏面试准备过渡态。该状态只表达 `createPracticePlan/getPracticePlan/startPracticeSession` 尚未完成，不展示虚假 opening message、百分比或可验证不了的阶段进度；成功后由服务端 session/opening message 进入正式会话，失败时关闭过渡态并回到原入口错误。过渡态必须阻断背景交互、提供 `role=status` / `aria-live=polite` / `aria-busy=true`，并在 `prefers-reduced-motion` 下停用非必要循环动画。

```text
PracticeScreen(sessionId)
├─ Global App TopBar（首页 / 面试 / 简历 / 暗色 / 语言 / 设置）
├─ Practice Session Header
│  ├─ 公司 / 岗位
│  ├─ 面试官角色
│  ├─ 计时
│  ├─ 暂停 / 恢复
│  ├─ disabled 电话图标（暂未开放）
│  └─ 结束并生成报告
└─ Conversation
   ├─ Transcript
   │  ├─ assistant message
   │  ├─ user message
   │  │  └─ failed-only retry icon
   │  ├─ pending-only interviewer-thinking row
   │  └─ terminal-only generic recovery state + current-plan CTA
   ├─ loader/completion error state
   └─ Composer
      ├─ text input
      └─ send
```

必须删除：

- 左侧“本轮题目”与所有 SessionMap DOM。
- Practice Session Header 题号/总题数。
- 主体 QuestionCard、题目 badge/topic/prompt。
- 专用 hint button/banner/count。
- PhoneSurface、字幕、麦克风、VAD、TTS、barge-in、hangup。
- 右侧辅助栏和任何会话内 persona/mode switch。

## 3 连续聊天规则

- opening assistant message 与后续 assistant reply 都是普通 message，不标记为问题。
- 用户输入是普通 message，不标记为回答。
- persisted user/assistant `message.text` 统一通过 `react-markdown + remark-gfm` 渲染；启用 `skipHtml` 且不使用 `rehypeRaw`。Raw HTML/event handler 不执行，remote image 不创建网络请求，`javascript:`/unsafe URI 被拒绝，安全外链使用 hardened `rel`。
- Markdown 仅是 view projection；send 与 same-ID retry 必须逐字节复用原始 `message.text/clientMessageId`，不得从渲染 DOM、纯文本或 normalized Markdown 反推 payload。
- UI 不显示“第 N 题”“追问”“回答”“下一题”等类别标签。
- transcript 来自 server `PracticeSession.messages`；user message 同时返回原 `clientMessageId/replyStatus`。刷新必须恢复完整有序会话及 pending/retryable/terminal/complete 回复状态，不使用本地 fixture、URL 或 browser storage 恢复业务状态。
- 用户请求提示时直接输入普通聊天内容，不存在专用 hint 行为。
- 提交后立即清空 composer，并先把该条 user message 加入 transcript；不得等 assistant response 返回后才一起显示。
- pending/retrying 期间禁用 composer，并显示可访问的面试官思考动画；此时不显示 retry。
- backend pending lease 固定为 90 秒；frontend POST 最多等待 95 秒。95 秒到达时 abort 本次请求并用同一 `clientMessageId` 调用 `getPracticeSession` 对账；不得盲目重发或改用新 ID，迟到旧 response 不得覆盖新事实。
- 失败时移除思考动画，保留原 user message，只在该消息底部显示 retry icon；retry 复用原文与同一 `clientMessageId`，不得重复追加。
- retry icon 仅用于无 HTTP response 的网络错误、API 明确 `retryable=true`，或刷新后 server message 为 `replyStatus=retryable_failed`；`pending` 继续 thinking，`terminal_failed` 以及 validation/auth/not-found/conflict/mismatch 不渲染 retry。`terminal_failed` 显示通用安全说明和唯一“返回当前面试规划”CTA，精确进入 `/workspace?targetJobId=...` 只读详情；不得进入 Parse 命令进度，auth/session-lost 仍走各自全局恢复。
- 失败后 textarea 可保存下一条草稿，但 submit 保持 disabled 并说明需先重试失败消息；retry 不得消费或覆盖该草稿。成功后以 server messages 收敛，optimistic row 不作为持久事实或 Finish 资格。
- timeout 对账读到 pending/retryable/terminal/complete 时采用 server truth；读失败或尚无该 ID 时保留原 optimistic row/ID 为 unresolved，继续锁定新 ID submit 与 Finish，不能因为 loader error 解锁 composer。
- 任一 optimistic pending/retryable-failed/retrying/terminal-recovery 状态都必须禁用 Finish，直到 server messages 完成收敛。

## 4 Practice Session Header

全局 App TopBar 在 Practice 保持可见，且与其他页面消费同一个内存 runtime/display context；进入、离开或切换 Practice route 不得重新调用 `getMe`。下列条目只描述其下方的会话控制栏。

- 公司/岗位优先来自 session.targetJobId 对应 generated `getTargetJob`。
- 面试官角色来自当前 round/plan，只读展示。
- 保留计时、暂停、结束。计时预算必须显示当前 `PracticePlan.timeBudgetMinutes`；该值在启动时来自所选 `TargetJob.summary.interviewRounds[]` 的 `durationMinutes`，不得写死 `25:00` 或其他默认分钟数。
- elapsed 是本地正计时；达到或超过预算不会自动结束，会话仍由用户点击“结束并生成报告”完成。plan budget loading/failure 时不得伪造一个默认预算。
- opening assistant message 不计为用户作答；提交首条 user message 前，“结束并生成报告”必须 disabled，并通过本地化 `aria-describedby` / 可见辅助文案说明“请先完成至少一次回答”。后端 `VALIDATION_FAILED` 仍是权威失败，前端不得仅靠本地计数宣称完成成功。
- 电话图标使用原生 disabled control：`disabled` + `aria-disabled=true`，灰色，无 click handler，title/aria-label 为“电话模式暂未开放 / Phone mode unavailable”。
- 不展示题号、总题数、text/phone segment、live chip 或 mode 文案。

## 5 Layout

- Desktop：Conversation 占满全局 App TopBar 与 Practice Session Header 下方可用宽度；内容列使用可读 max-width 居中，transcript 自适应增长，composer 固定在会话区底部。
- Mobile：单列；两层顶栏都可按各自合同换行，结束 CTA 可达；transcript 与 composer 不横向溢出。Markdown pre/code/table 只能在消息容器内局部滚动或安全换行，document `scrollWidth` 不得超过 viewport。
- 不保留空白 sidebar grid column。

## 6 报告边界

报告生成页只陈述真实 queued / generating / failed / timeout / ready 状态，不展示固定百分比、自动完成阶段、固定“实时观察”、未实现的通知或 records 承诺。timeout/network 才能继续检查；failed/not-found/invalid-contract/`REPORT_CONTEXT_TOO_LARGE` 是返回型终态，超限时引导返回规划并缩短输入后开启新会话。

Ready 报告只展示：

- 顶部两个数量指标、两行共四个能力/证据/行动区块，以及最低端全宽“面试总评”；准备度与服务端 `summary` 只在面试总评中出现。
- 报告内 code + 用户可见 label 的能力维度，以及本地化 status / confidence。
- 有候选人消息 grounding anchor 的会话证据摘要。
- 服务端排序的下一步行动。

报告不得展示题目回顾、逐题评分、题数、raw enum/code、turn-based retry 或 session UUID 等内部 locator。`reportId` 是唯一路由 locator，但不得作为用户界面字段；Context Strip/status/CTA identity 来自 frozen report context，Context Strip 只显示目标岗位、轮次和简历。复练/下一轮只提交 goal + sourceReportId，由后端从 source report/plan 派生 settings/round；复练有可靠 issue-backed dimension 时投影 focus，否则使用空 focus 开始通用同轮复练。

每份已创建报告附属一份只读会话记录。`/report-conversation?reportId=...` 复用本页 user/assistant Markdown/GFM message body 和 role visual language，但只消费 report-owned projection，不调用 `getPracticeSession`，也不呈现 Composer、thinking、retry、结束、暂停、计时、电话或任何实时状态控件。报告 queued / generating / ready / failed 都共享该访问边界；failed 报告可从 ReportsScreen 查看记录，并在非 `REPORT_CONTEXT_TOO_LARGE` 时手动把同一 report 重新排队。产品不提供 `listPracticeSessions`、会话历史列表或 `sessionId` 用户路由。

## 7 UI 实施与验证

- Practice：由本文档连续对话、回复状态、结束条件和 phone mode 约束承接。
- Report：由本文档报告结构与 `report-dashboard.md` 承接。
- Generating：由本文档 honest wait、失败恢复和 Back 行为承接。
- Shared primitives：使用正式 frontend token、component 和 accessibility contract。

正式 frontend 的验证分为：

1. DOM/control/a11y component contract。
2. responsive layout、viewport containment 与必要 browser smoke。
3. stale question/phone/hint positive-contract negative search。

## 8 验收标准

| ID | Given | When | Then |
|----|-------|------|------|
| U-1 | session 有 opening message | 进入 practice | 看到全局 App TopBar + Practice Session Header + 全宽聊天 + composer |
| U-2 | running session | 用户提交消息并等待 | user message 立即显示、composer 清空/禁用、面试官思考；成功后无重复 |
| U-3 | phone disabled | 查看/键盘操作电话图标 | 图标置灰且不能改变模式 |
| U-4 | send failure | 查看失败消息并点击其底部 retry icon | retry 只在失败后显示，复用原文与同一 clientMessageId；下一条草稿保留，成功后 user message 不重复且只有一个 reply |
| U-5 | 仅有 opening、尚无 user message | 查看/操作结束 CTA | CTA disabled，显示本地化可访问原因；绕过前端调用仍由后端 `VALIDATION_FAILED` 拒绝且不生成报告 |
| U-6 | desktop/mobile | responsive/browser gate | 无 sidebar 空白、无 document 横向溢出，关键操作在 viewport 内可见可用 |
| U-7 | 当前结构化轮次为 60 分钟 | 启动/刷新 Practice | plan 保存 60 分钟预算且 Top Bar 显示 `elapsed / 60:00`；不存在固定 `25:00` |
| U-8 | 已提交至少一条 user message | 点击结束并生成报告 | 进入 generating，随后会话级报告 |
| U-9 | AI 失败后刷新或组件重挂载 | `getPracticeSession` 返回 user `clientMessageId/replyStatus` | pending 恢复 thinking；retryable failure 在原 row 恢复同 ID retry；terminal failure 无 retry；成功后只有一个 assistant reply |
| U-10 | send POST 持续无响应 | 等待 95 秒 | abort 后按同一 ID 对账；pending/failed/complete 采用 server truth；对账失败时原 row/ID 保留且不能提交新消息；迟到 response 被忽略 |
| U-11 | server message 为 `terminal_failed` | 查看并点击恢复动作 | 无 row retry；显示通用安全说明；唯一 CTA 返回 `/workspace?targetJobId` 当前面试规划只读详情；无 current-scope `parse(targetJobId)` 恢复路径 |
| U-12 | persisted user/assistant text 含 GFM 与恶意 HTML/image/link/code/table | 渲染、retry 并在 390px 查看 | 两种角色安全渲染 GFM；HTML/remote image/unsafe URI 不执行；safe link hardened；retry exact raw text/ID；code/table 不撑破 document |
| U-13 | 报告资源已创建且 Practice 已结束 | 从报告打开会话记录 | 只读页按 sequence 显示同一安全 Markdown transcript；无 live controls、无 sessionId、无会话列表，并返回同一报告状态页 |
| U-14 | 报告生成失败且会话已结束 | 从 ReportsScreen 查看记录或重新生成 | 记录继续绑定同一 report；普通失败重新排队同一 report，超限失败只允许查看记录 |

## 9 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-19 | 1.35 | Practice 恢复全局 App TopBar，并把公司/角色/计时/会话动作明确命名为 Practice Session Header；route 切换不得重复读取账号偏好。 |
| 2026-07-18 | 1.34 | 在所有正式会话启动入口与 Practice 之间增加统一、诚实、可访问的面试准备过渡态。 |
| 2026-07-16 | 1.33 | 补齐已结束会话的 failed report 恢复：同 report 手动重新生成与任意状态只读面试记录。 |
| 2026-07-15 | 1.32 | 合并 report-owned 只读会话记录，复用正式 Practice 安全 Markdown message renderer，不恢复会话列表、sessionId route 或已删 Demo runtime。 |
| 2026-07-15 | 1.31 | 对齐报告 `3/2/2/2/1` 信息层级：准备度与服务端 summary 下移为底部全宽面试总评，顶部只保留两个数量指标。 |
| 2026-07-14 | 1.30 | Practice user/assistant 增加安全 Markdown/GFM view projection 与 mobile overflow 边界；terminal CTA 改为 Workspace targetJobId 只读详情。 |
| 2026-07-14 | 1.29 | T-B/P-A：90 秒服务端 lease 对应 95 秒前端 timeout + 同 ID 对账；terminal failure 增加精确返回当前 `parse(targetJobId)` 规划的通用 CTA。 |
| 2026-07-13 | 1.28 | 用户确认方案 A：Practice reply state 与原 clientMessageId 由后端读模型恢复，刷新后仍可在原消息下同 ID 重试。 |
| 2026-07-13 | 1.27 | Practice 增加即时 user row、pending 思考/锁定与失败消息底部同 ID retry；Report Context Strip 删除 session UUID 等内部 locator。 |
| 2026-07-12 | 1.26 | 增加零回答 Finish 禁用/后端权威拒绝，明确空 focus 通用同轮复练，并补输入超限报告的诚实终态。 |
| 2026-07-12 | 1.25 | 明确 Generating 终态动作、无 records 承诺、reportId-only frozen context 事实源与 report-local dimension focus。 |
| 2026-07-12 | 1.24 | 报告接入 grounded direct semantic shape；生成页删除伪实时语义，ready 页补齐 summary、code+label、本地化状态、服务端 replay focus 与强截图验收。 |
| 2026-07-12 | 1.23 | Practice 计时预算改为读取所选结构化轮次写入的 PracticePlan 时间快照；禁止固定 25 分钟和预算到点自动结束。 |
| 2026-07-12 | 1.22 | Practice 改为全宽连续文本聊天；删除题目、hint、phone surface；报告改为会话级。 |
