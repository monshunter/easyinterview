# E2E.P0.065 Debrief Default Render + 3 Picker Modal

> **场景 ID**: E2E.P0.065
> **关联需求**: frontend-debrief C-1, C-2, C-3, C-11, C-14
> **隔离级别**: shared-cluster
> **parallel-safe**: No
> **自动化入口**: `scripts/setup.sh` -> `scripts/trigger.sh` -> `scripts/verify.sh` -> `scripts/cleanup.sh`

## Given

User session authenticated; generated client surfaces
`listTargetJobs / listPracticeSessions / listResumes / getTargetJob /
getResume / getPracticeSession`; route
normalization maps the historical `debrief_full` alias to `debrief`.

## When

Run the frontend Vitest suite scoped to the debrief screen so the four
unit suites (`DebriefScreen`, `DebriefHeader`, `DebriefContextStrip`,
`DebriefStepper`) and the route alias regression exercise the default
render path plus picker open/close behaviour without launching a real
browser.

## Then

Vitest reports each suite as `passed`, no `failed` / `no tests to run`
banner appears, and the legacy negative lint stays clean.

## Cleanup

Remove the per-scenario output directory.
