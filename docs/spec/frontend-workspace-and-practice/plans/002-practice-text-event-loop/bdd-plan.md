# 002 Practice Continuous Conversation BDD Plan

> **版本**: 2.0
> **状态**: active
> **更新日期**: 2026-07-12

## Scenario Matrix
| ID | Type | Phase | Given | When | Then |
|----|------|-------|-------|------|------|
| E2E.P0.044 | primary | 3 | running session | exchange messages | ordered full-width chat, no question classification |
| E2E.P0.045 | alternate/regression | 2 | text session + phone params | load/operate UI | no side/question/hint; phone disabled; still text |
| E2E.P0.046 | failure/recovery | 3 | message AI failure | retry | user message retained once, one eventual reply |
| E2E.P0.047 | primary | 4 | running conversation | finish | generating handoff with stable IDs |
| E2E.P0.099 | real integration | 5 | real local stack/data | complete browser flow | conversation/report path and screenshots prove runtime behavior |
