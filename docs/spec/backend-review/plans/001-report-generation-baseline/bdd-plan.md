# Conversation-level Report BDD Plan

> **版本**: 2.0
> **状态**: active
> **更新日期**: 2026-07-12

## Scenario Matrix
| ID | Type | Given | When | Then |
|----|------|-------|------|------|
| E2E.P0.056 | primary/contract | completed conversation | generating/report load | ready session-level report, no question rows |
| E2E.P0.057 | alternate | ready report | retry/next CTA | fresh session receives competency/round context |
| E2E.P0.058 | failure/recovery | provider/invalid output/missing session | load/retry | typed failure and no partial ready data |
| E2E.P0.099 | real integration | real conversation completed | runner and browser flow | real conversation-level report is visible |
