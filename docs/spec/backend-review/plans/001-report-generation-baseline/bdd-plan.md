# Conversation-level Report BDD Plan

> **版本**: 2.1
> **状态**: completed
> **更新日期**: 2026-07-12

## Scenario Matrix
| ID | Type | Given | When | Then |
|----|------|-------|------|------|
| E2E.P0.056 | primary/contract | completed conversation with valid 1.0-5.0 scores | generating/report load | ready session-level report with deterministic readiness, no question rows |
| E2E.P0.057 | alternate | ready report | retry/next CTA | fresh session receives competency/round context |
| E2E.P0.058 | failure/recovery | provider/invalid score/empty or duplicate candidate dimension/missing session | load/retry | typed failure and no partial ready data |
| E2E.P0.099 | real integration | real conversation completed | runner and browser flow | real conversation-level report is visible |
