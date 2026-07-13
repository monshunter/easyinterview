# Honest Grounded Report Screen BDD Plan

> **版本**: 2.9
> **状态**: completed
> **更新日期**: 2026-07-13

## Scenario Matrix

| ID | Type | Phase | Given | When | Then |
|----|------|-------|-------|------|------|
| E2E.P0.056 | primary/composed | 6-7 | report queued/generating during action-local retry waits and tab hidden/blur during timer or in-flight request | pause/resume under one maxAttempts49 polling run | polling current/next attempt+delay preserved，resume n+1，no duplicate/concurrency/reset1；ready transitions honestly |
| E2E.P0.057 | primary/boundary | 7 | ready report with empty generic focus or valid non-empty issue-backed focus | replay/next CTA | Replay remains legal in both cases, sends no client focus, and a server-derived fresh session starts once |
| E2E.P0.058 | failure/recovery/composed | 7 | report action attempt4/nonretryable API failed，or client poll-window/network timeout，plus invalid direct shape/focus | load/action | v3 proves action-local10s/20s/40s/reset and async-attempt separation；only API terminal failed/invalid is back-only；maxAttempts49/network exhaustion is continue-check；no raw enum/fake report/internal attempts |
| E2E.P0.059 | regression/visual | 8 | identical deterministic zh/en prototype/formal fixtures at 1440/390 | parity runner | DOM/style/bbox and pixelmatch changed ratio ≤0.5% plus active negatives pass |
| E2E.P0.070 | cross-layer | 7 | ready needs-work source report | replay plan create/read | focus is projected by backend, not URL/request |
| E2E.P0.072 | security/failure | 7 | invalid/cross-user source report | replay request | fail closed without focus/privacy leakage |
| E2E.P0.099 | current-run canonical real UX + deterministic boundary parity | 8 | current-run en/zh ready rows + exact 24/64 ui-design/OpenAPI fixtures | capture exact six images；bind per-row canonical content/action/content-audit/screenshot/report/session/context digests；run prototype/formal parity | 390x844 real action regions show actual `<=24-whitespace-word` / `<=64-Unicode-code-point` labels fully；deterministic fixtures prove exact boundary wrapping with no clipping/ellipsis/hiding/overflow |
