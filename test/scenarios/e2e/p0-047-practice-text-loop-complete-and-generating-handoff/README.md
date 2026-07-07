# E2E.P0.047 Practice text loop — complete & generating handoff

> **场景 ID**: E2E.P0.047
> **执行方式**: automated
> **隔离级别**: in-process (vitest jsdom)
> **状态**: Ready

## 1 Given

practice fixture 数据就绪：`getPracticeSession=default / completing`；`appendSessionEvent=completed` 提供 `assistantAction.type=session_completed`；`completePracticeSession=default / replay / mismatch / session-already-completed / cross-user-not-found` 5 variants。

## 2 When

- 用户点击「结束并生成报告」CTA → useCompletePracticeSession 派生 `Idempotency-Key` → POST `completePracticeSession({clientCompletedAt})`
- 服务端 202 返回 `ReportWithJob{reportId, job}`
- 前端通过 `buildPracticeHandoffParams` 派生 generating route params 并 nav

## 3 Then

- body 仅 `{clientCompletedAt}`，display 字段（mode/modality/practiceMode/practiceGoal/hintUsed/hintCount）不出现在 body
- request 必含 `Idempotency-Key` header；retry 复用同一 key；replay scenario 同 key 二次返回首次 response
- nav `generating` 携带稳定 InterviewContext ID（planId / targetJobId / jdId / resumeId / roundId / sessionId / reportId）+ PracticeDisplayContext（mode / modality / practiceMode / practiceGoal / hintUsed / hintCount）
- URL params 不含 raw `answerText / questionText / hint / promptVersion / rubricVersion / modelId`
- 双击 finish CTA 仅一次 POST + 一次 nav
- 负向断言：`getFeedbackReport` / `createPracticeVoiceTurn` 调用次数 = 0；非当前 testid / 非当前 route alias / `Idempotency-Key.*appendSessionEvent` 全部 0 命中

## 4 执行

```bash
./scripts/setup.sh && ./scripts/trigger.sh && ./scripts/verify.sh && ./scripts/cleanup.sh
```

## 5 关联需求

`frontend-workspace-and-practice` C-4, C-6, C-12
