# 002 — Practice Text Event Loop BDD Plan

> **版本**: 1.10
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 Plan**: [plan](./plan.md)

## 1 Scenario Matrix

| 场景 ID | 类别 | 当前 owner 行为 | 关联 Gate |
|---------|------|----------------|-----------|
| `E2E.P0.044` | primary path | assisted text session loads, answers submit through `appendSessionEvent`, AssistantAction updates transcript, practice runtime negative grep passes | C-4 / C-8 / C-10 |
| `E2E.P0.045` | alternate path | assisted/strict visibility matrix remains stable across `baseline / retry_current_round / next_round`; hint, skip, pause/resume and strict lock pass | C-4 / C-10 / C-12 |
| `E2E.P0.046` | failure path | AI timeout, 404 lost state, 409 mismatch, strict conflict and retry reuse render current recovery UI | C-4 / C-12 |
| `E2E.P0.047` | completion path | complete 202 uses `Idempotency-Key`, body only has `clientCompletedAt`, handoff carries `resumeId` and display context, privacy grep passes | C-4 / C-6 / C-12 |

## 2 Behavior Details

### E2E.P0.044

Given a signed-in user with `sessionId / planId / targetJobId / jdId / resumeId / roundId` and `practiceMode=assisted`, when they open `practice`, submit an answer and receive `ask_follow_up` / `ask_question`, then the transcript, session map, assistant action renderer, generated client spy and runtime negative grep all match the current text-loop contract.

### E2E.P0.045

Given `practiceMode in {assisted, strict}` and `practiceGoal in {baseline, retry_current_round, next_round}`, when the user triggers hint, skip, pause/resume and strict toggle behavior, then assisted UI exposes helper controls, strict UI hides helper controls, current goals do not change visibility, and append requests still avoid `Idempotency-Key`.

### E2E.P0.046

Given fixture variants for timeout, missing session, mismatch and strict conflict, when event or session calls fail, then the UI keeps user input where appropriate, reuses retry IDs, refreshes server-wins state on mismatch, and keeps raw answer/question/hint/provenance out of URL, storage, console and telemetry.

### E2E.P0.047

Given a completed practice session, when the user clicks finish, then `completePracticeSession` sends only `clientCompletedAt` with `Idempotency-Key`, replay is idempotent, mismatch shows recovery UI, and navigation to `generating` carries only stable IDs plus display context.
