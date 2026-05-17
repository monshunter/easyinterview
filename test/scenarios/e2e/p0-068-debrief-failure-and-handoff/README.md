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
debrief module for any direct `createPracticePlan` / `startPracticeSession`
calls (must be zero — the handoff goes through `nav('practice', ...)`).

## Then

Vitest exits 0. Direct grep finds zero forbidden `createPracticePlan` /
`startPracticeSession` calls in the debrief module.
