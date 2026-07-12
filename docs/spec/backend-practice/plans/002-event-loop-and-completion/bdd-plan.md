# 002 Conversation Message Loop BDD Plan

> **版本**: 2.1
> **状态**: completed
> **更新日期**: 2026-07-12

## 1 Scenario Matrix
| ID | Type | Phase | Given | When | Then |
|----|------|-------|-------|------|------|
| E2E.P0.044 | primary/UI | 2 | Practice loaded | exchange multiple messages | ordered chat only |
| E2E.P0.046 | failure/recovery | 3/6 | AI failure, replay or mismatch | retry same/new message | no duplicate user row, one eventual reply, deterministic conflict |
| E2E.P0.047 | primary/lifecycle | 4/6 | running conversation with possible in-flight reply | finish | one report job; late assistant commit rolls back and cannot reopen session |
