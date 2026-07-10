# 002 — Practice Text Event Loop Test Plan

> **版本**: 1.14
> **状态**: active
> **更新日期**: 2026-07-10

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)

## 1 Test Strategy

This owner uses focused frontend unit/integration tests, fixture contract gates, route/handoff privacy checks, scenario scripts, real-environment browser smoke and screenshot evidence. It verifies the current text / phone session UI; bottom-layer STT/LLM/TTS orchestration and report polling are owner boundaries.

## 2 Test Matrix

| Area | Files / Commands | Assertions |
|------|------------------|------------|
| Screen parity | `src/app/screens/practice/PracticeScreen.test.tsx`, `src/app/screens/practice/__tests__/PracticeScreenIntegration.test.tsx`, `frontend/tests/pixel-parity/practice.spec.ts` | current DOM anchors, text/phone segmented controls, no independent side panel, global finish CTA, responsive shell, theme/customAccent smoke |
| Session load | `hooks/usePracticeSessionLoader.test.tsx`, `__tests__/practiceSessionLost.test.tsx` | generated `getPracticeSession`, refresh triggers, 404 lost state, workspace CTA params |
| Events | `hooks/usePracticeEvents.test.tsx`, `__tests__/appendSessionEventBody.test.tsx`, `__tests__/idempotencyContract.test.tsx` | answer/hint/pause/resume event kinds, UUIDv7 `clientEventId`, retry reuse, append has no `Idempotency-Key`, no positive `turn_skipped` path |
| Assistant actions | `components/AssistantActionRenderer.test.tsx`, `__tests__/practiceCompletion.test.tsx` | 5 action types, transcript update, finish CTA state |
| Policy | `__tests__/practiceGoalParity.test.tsx`, `__tests__/practiceHints.test.tsx`, `__tests__/practiceModeSwitch.test.tsx`, `__tests__/outOfScopeNegative.test.ts` | rendered optional hint usage, current three practice goals, no user-visible strict switch |
| Controls | `__tests__/practicePauseResume.test.tsx`, `__tests__/SessionMap.test.tsx`, phone-mode focused tests | pause/resume disabling, session map states, phone captions, hang-up, restart, no skip, no role switch |
| Current UI boundary | `PracticeScreen.test.tsx`, `__tests__/outOfScopeNegative.test.ts`, scenario verify scripts | no independent side-panel controls, dictation, skip, in-session persona switch, strict switch, voice expression metrics or manual transcript fallback UI |
| Completion | `hooks/useCompletePracticeSession.test.tsx`, `__tests__/completePracticeSessionBody.test.tsx`, `utils/practiceHandoffParams.test.ts` | `completePracticeSession` body, idempotency replay, `resumeId` handoff, forbidden-key guard |
| Privacy / boundary | `__tests__/practicePrivacy.test.tsx`, `__tests__/outOfScopeNegative.test.ts`, P0.044/P0.047 verify scripts | no raw text in URL/storage/log, no `getFeedbackReport` in practice runtime, voice turn confined to owner hook |
| Fixtures | `make validate-fixtures` | PracticeSessions fixtures match OpenAPI envelope and variants |
| Type safety | `pnpm --filter @easyinterview/frontend exec tsc --noEmit` | generated types and current route params compile |

## 3 Scenario Gates

| Scenario | Scripts | Scope |
|----------|---------|-------|
| `E2E.P0.044` | `test/scenarios/e2e/p0-044-practice-text-loop-assisted-happy-path/scripts/` | assisted text happy path, runtime negative grep |
| `E2E.P0.045` | `test/scenarios/e2e/p0-045-practice-text-loop-mode-policy-display/scripts/` | text/phone policy, current goal parity, current-boundary negative gates |
| `E2E.P0.046` | `test/scenarios/e2e/p0-046-practice-text-loop-failure-and-recovery/scripts/` | AI timeout, 404, 409 mismatch, hint retry/recovery, no strict-mode conflict path |
| `E2E.P0.047` | `test/scenarios/e2e/p0-047-practice-text-loop-complete-and-generating-handoff/scripts/` | complete 202, replay, privacy handoff |

## 4 Closeout Gates

Run:

```bash
python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/frontend-workspace-and-practice/plans/002-practice-text-event-loop/context.yaml --target frontend
corepack pnpm --filter @easyinterview/frontend test src/app/screens/practice src/app/interview-context/InterviewContext.test.tsx src/api/frontendOwners.realApiMode.test.ts
corepack pnpm --filter @easyinterview/frontend exec tsc --noEmit
make validate-fixtures
```

Then run P0.044-P0.047 scenario scripts and global doc gates.

Final acceptance additionally requires local real-environment browser screenshots for text and phone practice flows under `.test-output/`.
