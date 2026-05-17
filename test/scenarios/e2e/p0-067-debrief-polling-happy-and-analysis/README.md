# E2E.P0.067 Debrief Polling Happy + Analysis Render

> **场景 ID**: E2E.P0.067
> **关联需求**: frontend-debrief C-8, C-12
> **隔离级别**: shared-cluster
> **parallel-safe**: No
> **自动化入口**: `scripts/setup.sh` -> `scripts/trigger.sh` -> `scripts/verify.sh` -> `scripts/cleanup.sh`

## Given

`useDebriefPolling` + `DebriefAnalysisStep` Phase 5-6 surfaces are
wired. Generated client surfaces `getJob` + `getDebrief`.

## When

Run the Vitest contracts that cover the polling state machine and the
analysis renderer (DebriefScreen integration test + InterviewContext
reducer + privacy boundary scan).

## Then

Vitest exits 0 with all debrief suites passing. Legacy negative gate
holds.
