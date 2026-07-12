# 002 Conversation Message Loop BDD Plan

> **版本**: 2.0
> **状态**: active
> **更新日期**: 2026-07-12

## 1 Scenario Matrix
| ID | Type | Phase | Given | When | Then |
|----|------|-------|-------|------|------|
| E2E.P0.044 | primary/UI | 2 | Practice loaded | exchange multiple messages | ordered chat only |
| E2E.P0.046 | failure/recovery | 3 | AI failure | retry same message | no duplicate user row, one eventual reply |
| E2E.P0.047 | primary | 4 | running conversation | finish | one report job and generating handoff |
