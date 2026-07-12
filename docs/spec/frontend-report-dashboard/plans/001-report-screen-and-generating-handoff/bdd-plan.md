# Conversation Report Screen BDD Plan

> **版本**: 2.0
> **状态**: completed
> **更新日期**: 2026-07-12

## Scenario Matrix
| ID | Type | Given | When | Then |
|----|------|-------|------|------|
| E2E.P0.056 | primary | report queued→ready | generating/report load | conversation progress + four-surface dashboard |
| E2E.P0.057 | primary | ready report | replay/next CTA | fresh session with competency/round context |
| E2E.P0.058 | failure/recovery | failed/missing/timeout | load/retry | typed state, no fake report |
| E2E.P0.059 | regression/visual | zh/en desktop/mobile | parity gates | source/geometry match and no question surface |
| E2E.P0.099 | real integration | real stack/data | full browser flow | conversation report and screenshot evidence |
