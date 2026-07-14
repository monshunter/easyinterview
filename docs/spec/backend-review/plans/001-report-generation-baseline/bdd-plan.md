# Grounded Conversation Report BDD Plan

> **版本**: 2.19
> **状态**: active
> **更新日期**: 2026-07-14

## Scenario Matrix

| ID | Type | Phase | Given | When | Then |
|----|------|-------|-------|------|------|
| E2E.P0.056 | primary/contract | 7,11 | schema-valid P0.047 completion owner artifact + completed context-rich conversation + in-memory 62,397-byte regression/default-limit payloads | run exact `TestE2EP0056ReportBackendEvidence` then compose frontend runner | regression and 917,504-byte limit reach provider unchanged; backend artifact + frontend markers pass; frozen context, valid anchors/actions and no hidden score |
| E2E.P0.058 | failure/recovery | 6-9,11 | schema-valid completion owner artifact plus missing/mismatched/in-memory 917,505-byte oversized snapshot, action-local attempt2/3/4 recovery, attempt4 invalid/provider failure, second independent invocation and nonretryable failure | run exact `TestE2EP0058ReportFailureBackendEvidence` then compose typed frontend states | oversized input makes zero provider/repair calls and returns recoverable receipt；one action uses initial+3 with10s/20s/40s；new action resets；no partial ready |
| E2E.P0.070 | primary/replay | 8 | ready source report with empty or issue-backed needs-work focus | retry plan create/read/replay | server creates generic same-round retry for empty focus or atomically projects issue-backed focus |
| E2E.P0.072 | security/failure | 8 | missing/cross-user/wrong-target/resume/round/non-ready/invalid source | derived plan create | fail closed without privacy or focus leakage |
| E2E.P0.099 | real integration/UX + deterministic boundary parity | 8 | real shared env + current-run en/zh ready rows + exact 24/64 ui-design/OpenAPI fixtures | capture exact six full-page images；bind each ready row's canonical content/action/content-audit/screenshot/report/session/context digests；run prototype/formal pixel parity on boundary fixtures | 390x844 real report images prove actual labels satisfy `<=24 whitespace words` / `<=64 Unicode code points` and are fully visible；deterministic parity independently proves exact 24/64 wrapping with no clipping/ellipsis/hiding/overflow |
| E2E.P0.100 | real-provider quality + bounded retry | 8-9 | distinct fixed-five contexts + deterministic action-retry and takeover fixtures | product action-local initial+3 + lease fencing + evalkit generation/judge independent max4；strict diagnostic additionally requires 11/11 + blind audit | product waits10s/20s/40s and new invocation resets；mechanical finals100%；fixed-five semantic至少4/5；strict run59381 remains FAIL；async attempts stay infra-only；stale worker zero report/outbox/audit/job |
| E2E.P0.059 | cross-owner canonical-round overview | 10 | owned ready TargetJob has canonical rounds with no report, prior-ready+newer-failed, generating-only and latest-ready histories | call `listTargetJobReports`, then ReportsScreen joins `PracticeRoundRef` to current TargetJob summary | every round appears in canonical order；current/latest pointers remain independent and minimal；invalid ownership/context/round identity fails closed；only current target renders，with no full history or global report center |

Scenario entity ownership remains registry truth: frontend-report-dashboard owns 056/058/059, frontend-home-job-picks-and-parse owns 016, backend-practice owns 070/072, and e2e-scenarios-p0 owns 099/100. This plan supplies named backend evidence; it does not duplicate the scenario directories.

P0.056 backend artifact uses schema `report-backend-evidence.v1`; P0.058 uses `report-backend-evidence.v3` with separate database and runtime facts. Both use the exact keys/commands/markers defined in test-plan. Neither scenario may recreate `completePracticeSession`; both must consume `practice-completion-evidence.v1` from backend-practice/002. PASS requires exact Go test execution evidence, schema-valid redacted artifacts and absence of FAIL/no-test/raw-content markers before frontend composition.
