# 002 Conversation Message Loop BDD Plan

> **版本**: 2.4
> **状态**: completed
> **更新日期**: 2026-07-12

## 1 Scenario Matrix
| ID | Type | Phase | Given | When | Then |
|----|------|-------|-------|------|------|
| E2E.P0.044 | primary/UI | 2 | Practice loaded | exchange multiple messages | ordered chat only |
| E2E.P0.046 | failure/recovery | 3/6 | AI failure, replay or mismatch | retry same/new message | no duplicate user row, one eventual reply, deterministic conflict |
| E2E.P0.047 | primary/lifecycle | 4/6 | running conversation with possible in-flight reply | finish | one report job; late assistant commit rolls back and cannot reopen session |
| E2E.P0.044 | grounding primary | 7 | complete source snapshot exists while structured profile is empty | send follow-up | AI payload contains the complete tail marker and no invented resume facts |
| E2E.P0.046 | grounding failure/recovery | 7 | all resume content fields are empty after user reservation | generate reply then retry | typed validation, zero AI/assistant reply, and same client message remains recoverable |
| E2E.P0.047 | completion ledger primary/replay | 8 | current plan has canonical round identity | complete and replay completion | one committed `session_completed` fact, one report/job handoff, no duplicate round contribution |
| E2E.P0.098 | cross-layer round progression | 8 | multi-round TargetJob and real persisted plan/session, plus a same-user plan bound to another resume | complete first then final rounds and reload TargetJob | only TargetJob-bound-resume facts advance the canonical prefix; projection advances to the next existing round, then completed/null; report state and frontend storage are not facts |
