# E2E.P0.046 Practice text loop — failure & recovery

> **场景 ID**: E2E.P0.046
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom)
> **状态**: Ready

## 1 Given

practice fixture 数据就绪：`getPracticeSession` 提供 `default / missing-session`；`appendSessionEvent` 提供 `default / ai-timeout / mismatch / hint-conflict`。

## 2 When

- AI 502：`appendSessionEvent` 返回 502 `AI_PROVIDER_TIMEOUT` → 输入区下方 InlineError + retry；retry 复用同一 `clientEventId`
- 404：任一 practice operation 返回 404 → 渲染 `PracticeSessionLostState`，CTA 返回 workspace
- 409 mismatch：同一 `clientEventId` 改 payload → renderer 显示「同步异常，请刷新」
- 409 hint conflict：hint 触发后端冲突 → 普通 session conflict 错误，不出现 strict UI

## 3 Then

- 502 retry 复用同一 `clientEventId`（usePracticeEvents inflight cache）
- 404 触发 PracticeSessionLostState；CTA nav workspace 携带 targetJobId / jdId / planId / resumeId
- 409 mismatch / hint conflict 通过 `practice.errors.sessionConflict` 文案在 ErrorState 渲染
- raw `answerText` / `questionText` / `hint` / `provenance.modelId` 不出现在 console.log / URL / localStorage
- AI provider key / prompt registry / AIClient / LLM endpoint 0 命中

## 4 执行

```bash
./scripts/setup.sh && ./scripts/trigger.sh && ./scripts/verify.sh && ./scripts/cleanup.sh
```

## 5 关联需求

`frontend-workspace-and-practice` C-4, C-12
