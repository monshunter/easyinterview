# E2E.P0.068 Debrief Failure States + Replay Handoff

> **场景 ID**: E2E.P0.068
> **关联需求**: frontend-debrief C-9, C-10, C-13
> **隔离级别**: shared-cluster
> **parallel-safe**: No
> **自动化入口**: `scripts/setup.sh` -> `scripts/trigger.sh` -> `scripts/verify.sh` -> `scripts/cleanup.sh`

## Given

`DebriefFailureState`, `DebriefMissingContextState`, `DebriefTimeoutState`
+ Step 2 launcher + `useRequestAuth` handoff are wired.

## When

Run the Vitest contracts that pin the failure states + handoff path
through `DebriefScreen.test.tsx` plus the InterviewContext reducer round
trip + the privacy boundary scan. Additionally lint scoped grep on the
debrief replay handler to require `createPracticePlan`,
`startPracticeSession`, and `sourceDebriefId`, because the CTA must create
a fresh debrief practice session before `nav('practice', ...)`.

## Then

Vitest exits 0. Scoped grep confirms the debrief replay CTA creates a
fresh practice plan/session sourced from the current debrief.
