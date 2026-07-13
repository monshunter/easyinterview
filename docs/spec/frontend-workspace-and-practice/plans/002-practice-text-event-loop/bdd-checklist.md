# 002 Practice Continuous Conversation BDD Checklist

> **版本**: 2.5
> **状态**: active
> **更新日期**: 2026-07-13

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.044
- [x] Revise/run/record conversation happy-path evidence.
- [ ] Revision 2026-07-13 holds send pending and proves immediate user row, composer clear/lock, accessible interviewer-thinking, then reloads a server `replyStatus=pending` row and proves same-ID/no-resend recovery, success dedupe and clean 1440/390 screenshots. (`pending-reply` fixture is planned, not current.)
## E2E.P0.045
- [x] Revise/run/record simplified UI and disabled-phone evidence.
## E2E.P0.046
- [x] Revise/run/record failure/retry evidence.
- [x] Remediation: execute loader refresh and same-message retry screen assertions. (E2E.P0.046 PASS)
- [ ] Revision 2026-07-13 proves typed transport/`ApiClientError.apiError.retryable` classification and thinking removal；AI failure → reload restores server original text/same clientMessageId as `retryable_failed`, row-local retry preserves draft and converges to one user/reply pair；terminal validation/auth/not-found/conflict/mismatch has no retry and re-reads truth；all unresolved states block Finish；1440/390 failed-state screenshots match source. (`retryable-failed-reply`, `terminal-failed-reply` and terminal error fixtures are planned, not current.)
## E2E.P0.047
- [x] Revise/run/record completion evidence.
- [x] Remediation: execute completion retry routing and Finish CTA lifecycle assertions. (E2E.P0.047 PASS)

## E2E.P0.047 Phase 7 zero-answer completion

- [ ] Prepare opening-only and one-committed-user-message sessions without raw message evidence.
- [ ] Trigger frontend native-disabled/zh-en described-reason assertions and consume backend authoritative rejection/no-side-effect markers.
- [ ] Verify `ZERO_ANSWER_FINISH_DISABLED_PASS`, `ZERO_ANSWER_COMPLETION_REJECTED_PASS`, one-answer stable reportId handoff and exact replay; cleanup scenario rows.

## E2E.P0.047 Phase 8 reportId-only handoff

- [ ] After one-answer completion, assert browser URL/history navigation and downstream report request contain only reportId; copied target/plan/session/resume/round/status/error fields are absent.
- [ ] Replay completion and prove the same reportId locator returns without duplicate report navigation state.
