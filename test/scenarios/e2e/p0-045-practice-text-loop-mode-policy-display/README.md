# E2E.P0.045 Practice text loop mode policy display

> **场景 ID**: E2E.P0.045
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom)
> **状态**: Ready

## 1 Given

practice fixture 数据就绪：`getPracticeSession=default`；`appendSessionEvent` 多个 variant 可切换（`default / show-hint / hint-strict-conflict / turn-skipped / pause-resume`）。strict / assisted × baseline / retry_current_round / next_round 组合驱动显隐对照。

## 2 When

- assisted + baseline / retry_current_round / next_round：渲染 LIVE NOTES + hint button
- strict + baseline / retry_current_round / next_round：上述辅助区隐藏 + 渲染 strict-mode banner
- 点击 hint（assisted）触发 `hint_requested`，渲染 HintBanner、`INCREMENT_HINT_COUNT`
- 点击 skip 触发 `turn_skipped`，SessionMap 标记 `skipped`
- 点击 pause / resume 触发 `session_paused / session_resumed`，三按钮 disable / enable
- 点击 strict toggle 弹 lock toast，不调 backend

## 3 Then

- 6 个组合的显隐快照（assisted/strict × baseline/retry_current_round/next_round）pairwise 一致
- assisted 模式下 hint button 渲染 + 点击触发 `appendSessionEvent({kind:"hint_requested"})`
- strict 模式下 hint button DOM 0 命中 + strict-mode banner 渲染
- pause 后 send / hint / skip 三按钮 disabled，再点击 0 个 POST
- RoleDropdown 切换 0 个 generated client 调用
- 旧口径负向 grep（retired practice goal / `切到语音` / `Switch to voice` / voice imports / 旧 testid / `Idempotency-Key.*appendSessionEvent` / 独立 voice route）全部 0 命中

## 4 执行

```bash
./scripts/setup.sh && ./scripts/trigger.sh && ./scripts/verify.sh && ./scripts/cleanup.sh
```

## 5 关联需求

`frontend-workspace-and-practice` C-4, C-10, C-12
