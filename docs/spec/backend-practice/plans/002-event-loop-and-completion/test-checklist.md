# 002 Conversation Message Loop Test Checklist

> **版本**: 2.4
> **状态**: completed
> **更新日期**: 2026-07-12

**关联 Test Plan**: [test-plan](./test-plan.md)

## Phase 1: Store
- [x] Store reservation/replay/concurrency/uniqueness tests pass.
## Phase 2: Service/API
- [x] Happy message loop and contract tests pass.
## Phase 3: Failure
- [x] Failure/repair/retry tests pass.
## Phase 4: Completion
- [x] Completion/report handoff tests pass.
## Phase 5: Privacy
- [x] Isolation/redaction/full gates pass.
## Phase 6: Review remediation
- [x] Send/complete race and typed-conflict tests pass. (`go test ./backend/internal/practice -count=1`; `go test ./backend/internal/store/practice -count=1`)
- [x] P0.046/P0.047 scenario marker contract tests pass. (`python3 -m pytest scripts/lint/scenario_script_contract_test.py -q -k practice_failure_and_completion`; focused Go tests; `bash -n`)
## Phase 7: Resume grounding
- [x] Send store precedence and long-input tail-marker tests pass.<!-- verified: 2026-07-12 method=go-test -->
- [x] Empty context returns typed validation with zero AI/assistant reply while preserving user-message retry semantics.<!-- verified: 2026-07-12 method=go-test -->
- [x] P0.044/P0.046 scenario gates pass.<!-- verified: 2026-07-12 method=scenario -->
- [x] System-role policy, JSON untrusted-context escaping and persona-style-only follow-up payload/lint/eval gates pass.<!-- verified: 2026-07-12 method=go+pytest -->

## Phase 8: Completion ledger projection

- [x] Completion atomicity/replay tests prove exactly one lifecycle fact.<!-- verified: 2026-07-12 method=P0.047 result=PASS -->
- [x] Wrong-resume exclusion, duplicate session/event and report-status-independence projection tests pass.<!-- verified: 2026-07-12 method=P0.098 real-postgres marker=wrong-resume-completion-ignored=PASS -->
- [x] P0.047/P0.098 scenario gates pass with persisted backend evidence.<!-- verified: 2026-07-12 method=scenario-run -->
