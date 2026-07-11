# E2E.P0.045 Practice text loop assistance policy display

> **场景 ID**: E2E.P0.045
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom)
> **状态**: Ready

## 1 Given

practice fixture 数据就绪：`getPracticeSession=default`；`appendSessionEvent` 多个 variant 可切换（`default / show-hint / hint-conflict / pause-resume`）。out-of-scope strict input / assisted × baseline / retry_current_round / next_round 组合只用于确认当前 UI 不按范围外模式隐藏提示。

## 2 When

- assisted + baseline / retry_current_round / next_round：渲染 hint button
- out-of-scope strict input + baseline / retry_current_round / next_round：hint button 仍渲染，不出现 strict-mode banner
- 点击 hint（assisted）触发 `hint_requested`，渲染 HintBanner、`INCREMENT_HINT_COUNT`
- 点击 pause / resume 触发 `session_paused / session_resumed`，send / hint disable / enable
- TopBar 只保留一个电话图标：文本模式点击进入电话模式，电话模式再次点击立即回到同一 session 文本模式
- 电话主体只保留中心挂断按钮，点击后与 TopBar 电话图标走同一退出路径；不能出现 text/phone 分段、`live` 标记、重新开始或 call-ended 中间态
- 同一会话的文本/电话回答都采用服务端返回的新 session snapshot；连续两次电话回答必须从 turn A 前进到 turn B
- route 从 session A 切换为 session B 时，先清空 A 的草稿、暂停、提示、错误、标注和计时，再只显示 B 的首题

## 3 Then

- mode / goal 组合的显隐快照 pairwise 一致；电话切换始终是一个图标，中心按钮只负责挂断并回到文本模式
- loader 只接纳当前 session snapshot；会话内 turn 连续前进，会话间本地状态完全隔离
- assisted 模式下 hint button 渲染 + 点击触发 `appendSessionEvent({kind:"hint_requested"})`
- out-of-scope strict input 参数下 hint button 仍渲染，strict switch / strict banner DOM 0 命中
- pause 后 send / hint disabled，skip button DOM 0 命中，再点击 0 个 POST
- 目标岗位显示优先使用 server session / TargetJob 数据，生产渲染不暴露内部 `questionIntent`
- 范围外输入负向 grep（out-of-scope practice goal / `切到语音` / `Switch to voice` / out-of-scope voice imports / strict switch / skip button / `Idempotency-Key.*appendSessionEvent` / 独立 voice route）全部 0 命中

## 4 执行

```bash
./scripts/setup.sh && ./scripts/trigger.sh && ./scripts/verify.sh && ./scripts/cleanup.sh
```

## 5 关联需求

`frontend-workspace-and-practice` C-4, C-10, C-12
