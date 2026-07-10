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
- mode segment 只保留 text / phone，不能出现 role switch、strict switch、skip button

## 3 Then

- mode / goal 组合的显隐快照 pairwise 一致
- assisted 模式下 hint button 渲染 + 点击触发 `appendSessionEvent({kind:"hint_requested"})`
- out-of-scope strict input 参数下 hint button 仍渲染，strict switch / strict banner DOM 0 命中
- pause 后 send / hint disabled，skip button DOM 0 命中，再点击 0 个 POST
- 范围外输入负向 grep（out-of-scope practice goal / `切到语音` / `Switch to voice` / out-of-scope voice imports / strict switch / skip button / `Idempotency-Key.*appendSessionEvent` / 独立 voice route）全部 0 命中

## 4 执行

```bash
./scripts/setup.sh && ./scripts/trigger.sh && ./scripts/verify.sh && ./scripts/cleanup.sh
```

## 5 关联需求

`frontend-workspace-and-practice` C-4, C-10, C-12
