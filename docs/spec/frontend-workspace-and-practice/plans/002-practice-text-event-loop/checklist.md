# 002 — Practice Text Event Loop Checklist

> **版本**: 1.10
> **状态**: completed
> **更新日期**: 2026-07-07

**关联计划**: [plan](./plan.md)

## Current Contract

- [x] `PracticeScreen` 渲染当前 text event loop shell：TopBar、SessionMap、QuestionCard、Transcript、InputBar、RightPanel、HintBanner、FinishCta、lost state 和 mobile layout 都有 DOM anchor / a11y / i18n / theme 覆盖。
- [x] `usePracticeSessionLoader` 只通过 generated `getPracticeSession` 读取 session，覆盖 loading、data、refresh、404 和 error 状态。
- [x] `usePracticeEvents` 只通过 generated `appendSessionEvent` 提交 answer / hint / skip / pause / resume，body 含 `clientEventId`，request 不带 `Idempotency-Key`。
- [x] AssistantAction renderer 覆盖 `ask_question / ask_follow_up / show_hint / session_wait / session_completed`，provenance 只进 AI transparency UI。
- [x] `usePracticeAssistance` 只由 `practiceMode` 决定辅助显隐；`baseline / retry_current_round / next_round` 对显隐无副作用。
- [x] `useCompletePracticeSession` 只通过 generated `completePracticeSession` 完成会话，body 只含 `clientCompletedAt`，side-effect request 带 `Idempotency-Key` 并处理 replay / mismatch / 5xx / StrictMode 防抖。
- [x] `buildPracticeHandoffParams` 使用 `resumeId` 与稳定 owner IDs handoff 到 `generating`，不携带 answer / question / hint / prompt / model provenance。
- [x] voice turn 只允许出现在 `hooks/usePracticeVoiceTurn.ts`；text event loop 不调用 `getFeedbackReport`、不直接轮询 report、不绕过 generated client。
- [x] BDD-Gate: `E2E.P0.044` assisted happy path、`E2E.P0.045` mode policy、`E2E.P0.046` failure recovery、`E2E.P0.047` completion handoff 场景资产和 verify scripts 齐全。

## Verification

- [x] `validate_context.py frontend-workspace-and-practice/002 frontend`
- [x] Focused practice Vitest suite covers screen, hooks, components, handoff utils and privacy gates.
- [x] `pnpm --filter @easyinterview/frontend exec tsc --noEmit`
- [x] `make validate-fixtures`
- [x] P0.044-P0.047 scenario scripts execute `setup -> trigger -> verify -> cleanup`.
- [x] Owner wording grep has no pre-D20 resume binding field, placeholder voice surface token, stepwise build narrative, plan-reservation note, or out-of-owner generated-client positive surface.
- [x] `sync-doc-index --check`
- [x] `make docs-check`
- [x] `git diff --check`
