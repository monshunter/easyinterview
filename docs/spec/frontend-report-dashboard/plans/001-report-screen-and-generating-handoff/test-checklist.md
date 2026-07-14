# Honest Grounded Report Screen Test Checklist

> **版本**: 3.4
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Test Plan**: [test-plan](./test-plan.md)

## Generating and report behavior

- [x] Polling schedule、pause/resume fences、single-run cap and truthful typed state tests pass.
- [x] Direct report contract、empty/non-empty focus、CTA request and mixed UI/report language tests pass.
- [x] Invalid/over-limit payloads fail closed without raw label, truncation or rewrite.

## Layout and privacy

- [x] Exact 24/64 deterministic fixtures wrap completely at desktop/mobile; 25/65 fail closed.
- [x] Prototype/formal DOM/style/bbox/viewport/pixel tests pass as code-level visual regression.
- [x] Report/session UUID sentinel DOM/a11y negatives and target/round/resume positives pass.

## ReportsScreen and routing

- [x] Current-target isolation、canonical join、current/latest-only、loading/empty/error and stale-response fence tests pass.
- [x] Trusted/untrusted Back matrix、reportId-only route and direct Workspace detail without Parse detour tests pass.
- [x] ReportsScreen-only list consumer and no TopBar/global/history/compatibility entry negatives pass.

## E2E separation and full regression

- [x] P0.099 independently binds current real report/generating API/DB/screenshot evidence; mock/component outputs cannot satisfy it.
- [x] Provider/eval and deterministic parity remain independent code gates, not E2E steps.
- [x] Root `make test` runs the complete frontend/backend unit regression for phase completion.
