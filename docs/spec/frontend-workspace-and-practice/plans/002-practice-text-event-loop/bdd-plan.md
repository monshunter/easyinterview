# 002 Practice Continuous Conversation BDD Plan

> **版本**: 2.4
> **状态**: active
> **更新日期**: 2026-07-12

## Scenario Matrix
| ID | Type | Phase | Given | When | Then |
|----|------|-------|-------|------|------|
| E2E.P0.044 | primary | 3 | running session | exchange messages | ordered full-width chat, no question classification |
| E2E.P0.045 | alternate/regression | 2 | text session + phone params | load/operate UI | no side/question/hint; phone disabled; still text |
| E2E.P0.046 | failure/recovery | 3/6 | loader or message AI failure | retry | loader refreshes or same message retries; no duplicate user message |
| E2E.P0.047 | primary/recovery/boundary | 4/6/7/8 | opening-only zero-answer session, one-answer session, or transient completion failure | inspect Finish, direct-call completion, finish/retry | zero-answer UI is disabled with localized accessible reason and backend independently rejects with zero side effects; one-answer completion/replay reaches generating with reportId as the only URL/history/API locator and no copied business identity/status |
| E2E.P0.099 | real integration | 5 | real local stack/data | complete browser flow | conversation/report path and screenshots prove runtime behavior |
