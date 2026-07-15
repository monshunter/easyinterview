# 001 Real API/UI Journeys BDD Plan

> **版本**: 3.11
> **状态**: active
> **更新日期**: 2026-07-15

**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 Scenario Matrix

| ID | 类型 | Given | When | Then |
|----|------|-------|------|------|
| E2E.P0.098 | real API/UI persistence | shared real stack with a pre-seeded round-1 plan and waiting session | browser logs in, calls the real completion flow, reloads Home/Workspace and opens TargetJob detail | canonical progress persists as `done,current,pending`; requests reach the backend; no round-2 plan is created |
| E2E.P0.099 | real report/generating/conversation UI/API | current-run en/zh ready reports plus honest generating resource and owned report-session-message rows | browser captures exact six full-page report/generating images, opens Conversation from Report, calls real reportId-only API, verifies read-only DB binding, then goes Back | each screenshot row binds current state/digests；bounded non-image evidence proves same report/context, strict message ordering, correct back target and zero public session-list requests without storing transcript prose |
| E2E.P0.101 | auth/settings-owned real API/UI session | shared real frontend/backend/Mailpit with a fresh email identity | browser verifies code, completes profile, opens Settings via the sole gear, checks the complete account email and display name, logs out from Settings and signs in again | first session requires profile setup；Settings matches the same account with no old menu/tab；logs/evidence redact the email；logout clears session；relogin skips setup；no request is intercepted/mocked and deleteMe is never called |

## 2 Failure and privacy

- Missing real frontend/backend/Mailpit/PostgreSQL prerequisites fail closed or return `MANUAL_REQUIRED`; they are not skipped as PASS.
- Route interception, fixture transport, dev mock, stale run identity, digest mismatch or non-exact screenshots fail the relevant scenario.
- Evidence excludes project user data and secrets: codes, cookies, raw JD/resume/transcript, complete prompt/response and report prose copies. Benign development metadata such as PNG color profiles is not private user data and remains subject only to integrity/digest validation.
- Code tests and eval gates are reported separately and cannot replace any real-environment scenario.
- P0.101 business assertions remain owned by backend-auth/frontend-shell; this suite owns only the executable asset and current-run result。账号删除由 domain/contract tests 承接，不进入共享 E2E。
- P0.099 conversation must not add screenshots: its directory, manifest and manual audit remain exactly six images; route/status/count/sequence/binding/back-target digests are the only added evidence fields.
