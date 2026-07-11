# E2E.P0.046 Practice text loop — failure & recovery

> **场景 ID**: E2E.P0.046
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom)
> **状态**: Ready

## 1 Given

practice fixture 数据就绪：`getPracticeSession` 提供 `default / missing-session`；`appendSessionEvent` 提供 `default / ai-timeout / mismatch / hint-conflict / replay`，其中 provider timeout 已按当前后端策略降级为 `200 session_wait`，replay 保留原成功快照；`createPracticeVoiceTurn` 提供 `chat-output-invalid`。后端 fake AI 还可连续返回两次语言/结构不合法问题。

## 2 When

- follow-up provider timeout：`appendSessionEvent` 返回 `200 session_wait`，保留输入且不追加伪成功 transcript；用户再次提交时生成新的 `clientEventId`
- replay：相同请求返回原始 `ask_follow_up` 成功快照，不制造另一种 assistant action
- 404：任一 practice operation 返回 404 → 渲染 `PracticeSessionLostState`，CTA 返回 workspace
- 409 mismatch：同一 `clientEventId` 改 payload → renderer 显示「同步异常，请刷新」
- 409 hint conflict：hint 触发后端冲突 → 普通 session conflict 错误，不出现 strict UI
- follow-up 连续两次语言/结构校验失败：后端返回既有 `session_wait`，保留输入且不重复追加 transcript；用户确认后用新的 `clientEventId` 重试
- voice chat 连续两次语言校验失败：前端映射顶层 `AI_OUTPUT_INVALID` 为本地化错误，留在同一 session，可直接切回文本继续

## 3 Then

- `session_wait` 保留输入且再次提交使用新的 `clientEventId`；网络/5xx transport retry 仍由 hook 的 inflight cache 保持原请求身份
- 404 触发 PracticeSessionLostState；CTA nav workspace 携带 targetJobId / jdId / planId / resumeId
- 409 mismatch / hint conflict 通过 `practice.errors.sessionConflict` 文案在 ErrorState 渲染
- `session_wait` 不把未接受答案伪装成已提交 transcript，重试不复用已结算的 `clientEventId`；replay 返回原成功快照
- `AI_OUTPUT_INVALID` 不展示 provider/raw mock 内容，中英文均使用 `practice.errors.aiOutputInvalid`
- raw `answerText` / `questionText` / `hint` / `provenance.modelId` 不出现在 console.log / URL / localStorage
- AI provider key / prompt registry / AIClient / LLM endpoint 0 命中

## 4 执行

```bash
./scripts/setup.sh && ./scripts/trigger.sh && ./scripts/verify.sh && ./scripts/cleanup.sh
```

## 5 关联需求

`frontend-workspace-and-practice` C-4, C-12
