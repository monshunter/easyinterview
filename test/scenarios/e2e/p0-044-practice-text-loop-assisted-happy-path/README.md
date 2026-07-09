# E2E.P0.044 Practice text loop — assisted happy path

> **场景 ID**: E2E.P0.044
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom)
> **状态**: Ready

## 1 Given

用户已登录，practice fixture 数据就绪：`getPracticeSession=default` 提供 currentTurn；`appendSessionEvent=default` 提供 `assistantAction.type=ask_follow_up`；assisted + baseline 模式 + sessionId 通过 route param 传入。

## 2 When

进入 practice route，PracticeScreen 渲染静态壳 + 动态 currentTurn；用户在输入框输入答案 → 点击 Send → 触发 `appendSessionEvent({kind:"answer_submitted", payload:{turnId, answerText}})`；AssistantActionRenderer 派发 `ask_follow_up` 写入 transcript。

## 3 Then

- 顶部 TopBar 渲染当前控件 testid（company / title / question / timer / pause / mode-text / mode-phone / finish CTA）
- 中部 grid 渲染 SessionMap / QuestionCard / Transcript / InputBar，且不出现 RightPanel
- send 触发 `appendSessionEvent` body 含 UUIDv7 `clientEventId`，request init 不含 `Idempotency-Key` header
- `assistantAction.type=ask_follow_up` 在 transcript 追加 follow-up 标记的 AI 消息
- assisted 模式下 LIVE NOTES、hint button、experience cards 全部渲染
- 非当前 prototype testid `practice-mode-card-*` / `growth-*` / `drill-builder-*` / `mistakes-queue-*` 0 命中

## 4 执行

```bash
./scripts/setup.sh && ./scripts/trigger.sh && ./scripts/verify.sh && ./scripts/cleanup.sh
```

## 5 关联需求

`frontend-workspace-and-practice` C-4, C-8, C-9
