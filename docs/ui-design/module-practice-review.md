# 模拟面试与报告模块

> **版本**: 1.23
> **状态**: active
> **更新日期**: 2026-07-12

## 1 目标

模拟面试采用连续文本聊天。系统不展示或维护题号、题目总数、当前题、追问/下一题分类、题目地图或专用提示；AI 根据 JD、简历、轮次和对话历史自然推进。用户主动结束后生成会话级报告。

## 2 Practice 页面结构

```text
PracticeScreen(sessionId)
├─ TopBar
│  ├─ 公司 / 岗位
│  ├─ 面试官角色
│  ├─ 计时
│  ├─ 暂停 / 恢复
│  ├─ disabled 电话图标（暂未开放）
│  └─ 结束并生成报告
└─ Conversation
   ├─ Transcript
   │  ├─ assistant message
   │  └─ user message
   ├─ Error / retry state
   └─ Composer
      ├─ text input
      └─ send
```

必须删除：

- 左侧“本轮题目”与所有 SessionMap DOM。
- TopBar 题号/总题数。
- 主体 QuestionCard、题目 badge/topic/prompt。
- 专用 hint button/banner/count。
- PhoneSurface、字幕、麦克风、VAD、TTS、barge-in、hangup。
- 右侧辅助栏和任何会话内 persona/mode switch。

## 3 连续聊天规则

- opening assistant message 与后续 assistant reply 都是普通 message，不标记为问题。
- 用户输入是普通 message，不标记为回答。
- UI 不显示“第 N 题”“追问”“回答”“下一题”等类别标签。
- transcript 来自 server `PracticeSession.messages`；刷新必须恢复完整有序会话，不使用本地 fixture transcript。
- 用户请求提示时直接输入普通聊天内容，不存在专用 hint 行为。
- 发送期间禁用 composer；失败保留用户消息和 retry，不重复追加。

## 4 Top Bar

- 公司/岗位优先来自 session.targetJobId 对应 generated `getTargetJob`。
- 面试官角色来自当前 round/plan，只读展示。
- 保留计时、暂停、结束。计时预算必须显示当前 `PracticePlan.timeBudgetMinutes`；该值在启动时来自所选 `TargetJob.summary.interviewRounds[]` 的 `durationMinutes`，不得写死 `25:00` 或其他默认分钟数。
- elapsed 是本地正计时；达到或超过预算不会自动结束，会话仍由用户点击“结束并生成报告”完成。plan budget loading/failure 时不得伪造一个默认预算。
- 电话图标使用原生 disabled control：`disabled` + `aria-disabled=true`，灰色，无 click handler，title/aria-label 为“电话模式暂未开放 / Phone mode unavailable”。
- 不展示题号、总题数、text/phone segment、live chip 或 mode 文案。

## 5 Layout

- Desktop：Conversation 占满 TopBar 下方可用宽度；内容列使用可读 max-width 居中，transcript 自适应增长，composer 固定在会话区底部。
- Mobile：单列；TopBar 控件可换行但结束 CTA 可达；transcript 与 composer 不横向溢出。
- 不保留空白 sidebar grid column。

## 6 报告边界

报告只展示：

- 准备度。
- 能力维度。
- 会话证据。
- 下一步行动。

报告不得展示题目回顾、逐题评分、题数或 turn-based retry；复练当前轮携带能力重点而不是题目 ID。

## 7 UI 真理源

- Practice：`ui-design/src/screen-practice.jsx::PracticeScreen`
- Report：`ui-design/src/screen-report.jsx::ReportScreen`
- Generating：`ui-design/src/screens-p0-complete.jsx::ReportGeneratingScreen`
- Shared primitives：`ui-design/src/primitives.jsx`

正式 frontend 必须源级复刻上述当前原型。验证分为：

1. DOM/control/a11y source structure parity。
2. computed style/bounding box/responsive/screenshot geometry parity。
3. stale question/phone/hint positive-contract negative search。

## 8 验收标准

| ID | Given | When | Then |
|----|-------|------|------|
| U-1 | session 有 opening message | 进入 practice | 只看到 TopBar + 全宽聊天 + composer |
| U-2 | 多轮 messages | 用户连续发送 | 消息按序追加，无题目分类 |
| U-3 | phone disabled | 查看/键盘操作电话图标 | 图标置灰且不能改变模式 |
| U-4 | send failure | 重试同一 clientMessageId | user message 不重复，成功后只有一个 reply |
| U-5 | 用户结束 | 点击结束并生成报告 | 进入 generating，随后会话级报告 |
| U-6 | desktop/mobile | parity gate | 无 sidebar 空白、无溢出、截图与原型一致 |
| U-7 | 当前结构化轮次为 60 分钟 | 启动/刷新 Practice | plan 保存 60 分钟预算且 Top Bar 显示 `elapsed / 60:00`；不存在固定 `25:00` |

## 9 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-12 | 1.23 | Practice 计时预算改为读取所选结构化轮次写入的 PracticePlan 时间快照；禁止固定 25 分钟和预算到点自动结束。 |
| 2026-07-12 | 1.22 | Practice 改为全宽连续文本聊天；删除题目、hint、phone surface；报告改为会话级。 |
