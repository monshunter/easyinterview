# 001 Real API/UI Journeys BDD Plan

> **版本**: 3.7
> **状态**: active
> **更新日期**: 2026-07-14

**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 Scenario Matrix

| ID | 类型 | Given | When | Then |
|----|------|-------|------|------|
| E2E.P0.098 | real API/UI persistence | shared real stack with a pre-seeded round-1 plan and waiting session | browser logs in, calls the real completion flow, reloads Home/Workspace and opens TargetJob detail | canonical progress persists as `done,current,pending`; requests reach the backend; no round-2 plan is created |
| E2E.P0.099 | real report/generating UI/API | current-run en/zh ready reports plus honest generating resource | browser captures exact six full-page images and runner binds authenticated API/read-only DB evidence | each row binds current state and digests; ready action regions are complete and generating never claims ready |
| E2E.P0.101 | auth-owned real API/UI session | shared real frontend/backend/Mailpit with a fresh email identity | browser receives the Mailpit code, verifies, completes profile, logs out and signs in again | the first session requires profile setup; the completed account relogin does not; no request is intercepted or mocked |

## 2 Failure and privacy

- Missing real frontend/backend/Mailpit/PostgreSQL prerequisites fail closed or return `MANUAL_REQUIRED`; they are not skipped as PASS.
- Route interception, fixture transport, dev mock, stale run identity, digest mismatch or non-exact screenshots fail the relevant scenario.
- Evidence excludes secrets, codes, cookies, raw JD/resume/transcript, complete prompt/response and report prose copies.
- Code tests and eval gates are reported separately and cannot replace any real-environment scenario.
- P0.101 business assertions remain owned by backend-auth/frontend-shell; this suite owns only the executable asset and current-run result.
