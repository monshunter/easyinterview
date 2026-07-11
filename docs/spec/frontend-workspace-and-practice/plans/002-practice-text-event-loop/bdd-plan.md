# 002 — Practice Text Event Loop BDD Plan

> **版本**: 1.15
> **状态**: active
> **更新日期**: 2026-07-11

**关联 Plan**: [plan](./plan.md)

## 1 Scenario Matrix

| 场景 ID | 类别 | 当前 owner 行为 | 关联 Gate |
|---------|------|----------------|-----------|
| `E2E.P0.044` | primary path | text session loads without independent side-panel controls / dictation / skip / role switch / strict switch; answers submit through `appendSessionEvent`, AssistantAction updates transcript, practice runtime negative grep passes | C-4 / C-8 / C-10 |
| `E2E.P0.045` | alternate path | text/phone matrix uses one handset icon and one same-session hang-up; no segmented/live/cut-off/restart/callEnded; getTargetJob company/title is real and raw questionIntent stays hidden | C-4 / C-5 / C-5a / C-10 / C-12 |
| `E2E.P0.046` | failure path | AI timeout, 404, 409, text repair `session_wait` retains input and retries with a new event ID without duplicate transcript; voice typed error keeps current session and text-mode exit; no canned question or strict-mode conflict path | C-4 / C-5a / C-12 |
| `E2E.P0.047` | completion path | complete 202 uses `Idempotency-Key`, body only has `clientCompletedAt`, handoff carries `resumeId` and display context, privacy grep passes | C-4 / C-6 / C-12 |
| `REAL.ENV.SCREENSHOT` | final acceptance | local real backend/frontend environment renders text and phone practice flows, and screenshots prove current UI boundaries | C-4 / C-5 / C-8 / C-9 / C-10 |

## 2 Behavior Details

### E2E.P0.044

Given a signed-in user with `sessionId / planId / targetJobId / jdId / resumeId / roundId`, when they open text `practice`, submit an answer and receive `ask_follow_up` / `ask_question`, then the transcript, session map, assistant action renderer, generated client spy and runtime negative grep all match the current text-loop contract, and no independent side-panel controls, dictation, skip, role switch, visible strict switch or voice-analysis anchors render.

### E2E.P0.045

Given `mode/modality in {text, phone}`, a real `targetJobId` and any current practice goal, when the user clicks the single handset, uses captions or hangs up, then text enters phone, phone exits to text for the same session, microphone/TTS stop, no later phone TTS plays, and segmented/live/cut-off/restart/callEnded controls are absent. Top Bar displays generated `getTargetJob` company/title and no raw `questionIntent` or fixture transcript appears.

### E2E.P0.046

Given variants for timeout, missing session, mismatch, text follow-up repair failure and voice follow-up repair failure, when event/session/voice calls fail, then network retries reuse their in-flight ID, but an acknowledged text `session_wait` retains the answer and retries with a new `clientEventId` without duplicate transcript; mismatch refreshes server-wins state; the existing top-level typed voice error keeps the same session and text-mode exit; no path invents a canned question or leaks raw answer/question/hint/provenance into URL, storage, console or telemetry.

### E2E.P0.047

Given a completed practice session, when the user clicks finish, then `completePracticeSession` sends only `clientCompletedAt` with `Idempotency-Key`, replay is idempotent, mismatch shows recovery UI, and navigation to `generating` carries only stable IDs plus display context.

### REAL.ENV.SCREENSHOT

Given the local dev dependencies, backend and frontend are running from the current branch in real API mode, when a browser opens the same `practice` session in text and phone modes, then screenshots and runtime evidence show real company/title, one handset icon, the center red hang-up and same-session return to text, while segmented/live/cut-off/restart/callEnded, fixture dialogue, raw questionIntent and other out-of-scope controls remain absent.
