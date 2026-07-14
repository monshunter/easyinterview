# 002 Conversation Message Loop BDD Plan

> **版本**: 2.9
> **状态**: completed
> **更新日期**: 2026-07-14

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
| E2E.P0.047 | reportable completion/context owner | 9 | zero-answer, pending-reply or one-answer running conversation | run the three exact `TestE2EP0047*` owner tests and finish | invalid completion has no writes; one answer atomically freezes/replays report-context.v1; schema-valid owner artifact exposes markers for P0.056/058 consumption |
| E2E.P0.044 | immediate send/pending | 10 | mutable session and deterministic deferred AI | submit one message and read session while pending, then allow success | persisted user row has original clientMessageId + pending status; one assistant reply completes it |
| E2E.P0.046 | reload recovery | 10 | AI returns retryable or terminal failure after user reservation | reload/get session, then retry only the retryable row with the same ID | read projection restores exact status; retryable failure converges to one reply, terminal failure has no retry, new IDs remain blocked while unresolved |
| E2E.P0.044 | lease-bounded pending | 11 | one immediate request and one reloaded server pending row are covered by a current tracked-source fingerprint | hold before 90 seconds, then read after expiry | immediate/persisted pending UX remains locked before expiry；GET lazily converges expiry without a duplicate send；fresh screenshot/source hashes prove current evidence |
| E2E.P0.046 | concurrent expiry and stale worker | 11 | G1 is pending, lease expires, two same-ID retries race and the old worker later returns | GET/reserve G2, release stale G1 Commit/Fail, then commit G2 | one retry owns G2；stale G1 writes nothing；one assistant reply exists；terminal recovery exposes the current-plan CTA；95-second client timeout reconciles the same ID |
| E2E.P0.046 | configured text boundaries | 12 | runtime config exposes 32KiB message / 256KiB session limits and persisted history is below/at boundary | send ASCII and multibyte limit/limit+1 messages, reload and retry accepted ID | boundary values persist and call AI；single/aggregate +1 returns typed validation with zero new row/provider；reload stays consistent |
