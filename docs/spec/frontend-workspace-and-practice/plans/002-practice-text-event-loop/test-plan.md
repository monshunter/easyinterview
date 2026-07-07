# 002 — Practice Text Event Loop Test Plan

> **版本**: 1.10
> **状态**: completed
> **更新日期**: 2026-07-07

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)

## 1 Test Strategy

This owner uses focused frontend unit/integration tests, fixture contract gates, route/handoff privacy checks and scenario scripts. It verifies the current text event loop only; voice turn orchestration and report polling are owner boundaries.

## 2 Test Matrix

| Area | Files / Commands | Assertions |
|------|------------------|------------|
| Screen parity | `src/app/screens/practice/PracticeScreen.test.tsx`, `src/app/screens/practice/__tests__/PracticeScreenIntegration.test.tsx`, `frontend/tests/pixel-parity/practice.spec.ts` | current DOM anchors, segmented controls, strict switch, responsive shell, theme/customAccent smoke |
| Session load | `hooks/usePracticeSessionLoader.test.tsx`, `__tests__/practiceSessionLost.test.tsx` | generated `getPracticeSession`, refresh triggers, 404 lost state, workspace CTA params |
| Events | `hooks/usePracticeEvents.test.tsx`, `__tests__/appendSessionEventBody.test.tsx`, `__tests__/idempotencyContract.test.tsx` | 5 event kinds, UUIDv7 `clientEventId`, retry reuse, append has no `Idempotency-Key` |
| Assistant actions | `components/AssistantActionRenderer.test.tsx`, `__tests__/practiceCompletion.test.tsx` | 5 action types, transcript update, finish CTA state |
| Policy | `hooks/usePracticeAssistance.test.ts`, `__tests__/practiceGoalParity.test.tsx`, `__tests__/practiceHints.test.tsx`, `__tests__/practiceStrictToggleLocked.test.tsx` | strict/assisted visibility, current three practice goals, strict lock |
| Controls | `__tests__/practiceSkip.test.tsx`, `__tests__/practicePauseResume.test.tsx`, `__tests__/SessionMap.test.tsx`, `__tests__/RoleDropdown.test.tsx` | skip, pause/resume disabling, session map states, UI-only role switch |
| Completion | `hooks/useCompletePracticeSession.test.tsx`, `__tests__/completePracticeSessionBody.test.tsx`, `utils/practiceHandoffParams.test.ts` | `completePracticeSession` body, idempotency replay, `resumeId` handoff, forbidden-key guard |
| Privacy / boundary | `__tests__/practicePrivacy.test.tsx`, `__tests__/nonCurrentNegative.test.ts`, P0.044/P0.047 verify scripts | no raw text in URL/storage/log, no `getFeedbackReport` in practice runtime, voice turn confined to owner hook |
| Fixtures | `make validate-fixtures` | PracticeSessions fixtures match OpenAPI envelope and variants |
| Type safety | `pnpm --filter @easyinterview/frontend exec tsc --noEmit` | generated types and current route params compile |

## 3 Scenario Gates

| Scenario | Scripts | Scope |
|----------|---------|-------|
| `E2E.P0.044` | `test/scenarios/e2e/p0-044-practice-text-loop-assisted-happy-path/scripts/` | assisted text happy path, runtime negative grep |
| `E2E.P0.045` | `test/scenarios/e2e/p0-045-practice-text-loop-mode-policy-display/scripts/` | strict/assisted policy matrix and current goal parity |
| `E2E.P0.046` | `test/scenarios/e2e/p0-046-practice-text-loop-failure-and-recovery/scripts/` | AI timeout, 404, 409 mismatch, strict conflict |
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
