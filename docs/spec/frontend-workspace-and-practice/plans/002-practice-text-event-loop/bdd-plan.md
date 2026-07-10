# 002 — Practice Text Event Loop BDD Plan

> **版本**: 1.14
> **状态**: active
> **更新日期**: 2026-07-10

**关联 Plan**: [plan](./plan.md)

## 1 Scenario Matrix

| 场景 ID | 类别 | 当前 owner 行为 | 关联 Gate |
|---------|------|----------------|-----------|
| `E2E.P0.044` | primary path | text session loads without independent side-panel controls / dictation / skip / role switch / strict switch; answers submit through `appendSessionEvent`, AssistantAction updates transcript, practice runtime negative grep passes | C-4 / C-8 / C-10 |
| `E2E.P0.045` | alternate path | text/phone visibility matrix remains stable across `baseline / retry_current_round / next_round`; hint optional, pause/resume, phone captions, hang-up/restart and current-boundary negative gates pass | C-4 / C-5 / C-10 / C-12 |
| `E2E.P0.046` | failure path | AI timeout, 404 lost state, 409 mismatch, hint retry/recovery and retry reuse render current recovery UI; no strict-mode conflict path is exposed | C-4 / C-12 |
| `E2E.P0.047` | completion path | complete 202 uses `Idempotency-Key`, body only has `clientCompletedAt`, handoff carries `resumeId` and display context, privacy grep passes | C-4 / C-6 / C-12 |
| `REAL.ENV.SCREENSHOT` | final acceptance | local real backend/frontend environment renders text and phone practice flows, and screenshots prove current UI boundaries | C-4 / C-5 / C-8 / C-9 / C-10 |

## 2 Behavior Details

### E2E.P0.044

Given a signed-in user with `sessionId / planId / targetJobId / jdId / resumeId / roundId`, when they open text `practice`, submit an answer and receive `ask_follow_up` / `ask_question`, then the transcript, session map, assistant action renderer, generated client spy and runtime negative grep all match the current text-loop contract, and no independent side-panel controls, dictation, skip, role switch, visible strict switch or voice-analysis anchors render.

### E2E.P0.045

Given `mode/modality in {text, phone}` and `practiceGoal in {baseline, retry_current_round, next_round}`, when the user triggers optional hint, pause/resume, phone captions, hang-up and restart behavior, then current goals do not change visibility, out-of-scope controls remain absent, phone mode is user-visible as Phone/电话模式, and append requests still avoid `Idempotency-Key`.

### E2E.P0.046

Given fixture variants for timeout, missing session, mismatch and hint retry/recovery, when event or session calls fail, then the UI keeps user input where appropriate, reuses retry IDs, refreshes server-wins state on mismatch, keeps hint usage optional, does not surface strict-mode conflicts, and keeps raw answer/question/hint/provenance out of URL, storage, console and telemetry.

### E2E.P0.047

Given a completed practice session, when the user clicks finish, then `completePracticeSession` sends only `clientCompletedAt` with `Idempotency-Key`, replay is idempotent, mismatch shows recovery UI, and navigation to `generating` carries only stable IDs plus display context.

### REAL.ENV.SCREENSHOT

Given the local dev dependencies, backend and frontend are running from the current branch, when a browser opens the real `practice` page in text and phone modes, then screenshots show the current UI without independent side-panel controls, speech-to-text, skip, role switch, visible strict switch or voice-analysis panels, and the captured evidence is stored under `.test-output/`.
