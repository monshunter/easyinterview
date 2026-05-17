# E2E.P0.066 Text Mode AI Suggestions + Entries + createDebrief Submit

> **场景 ID**: E2E.P0.066
> **关联需求**: frontend-debrief C-4, C-5, C-7
> **隔离级别**: shared-cluster
> **parallel-safe**: No
> **自动化入口**: `scripts/setup.sh` -> `scripts/trigger.sh` -> `scripts/verify.sh` -> `scripts/cleanup.sh`

## Given

Generated client surfaces `suggestDebriefQuestions` + `createDebrief`;
`DebriefScreen` Phase 3-5 surfaces are wired (`GuidedDebriefRecord`,
`DebriefSubmitCTA`, `useSubmitDebrief`).

## When

Run the Vitest contracts that pin Phase 3-5 behaviour:
GuidedDebriefRecord / DebriefSubmitCTA / VoiceDebriefRecord smoke +
`InterviewContext.test.tsx > TestInterviewContext_SetDebriefContext`
(Phase 5.4 reducer extension) + the privacy boundary scan.

## Then

Vitest exits 0, every entry source path (ai_confirmed / ai_edited /
manual) compiles to wire shape, and the privacy boundary tests assert
no localStorage / sessionStorage / console.log writes leak debrief
field names.

## Cleanup

Remove the per-scenario output directory.
