# 001 Real API/UI Journeys BDD Checklist

> **版本**: 3.8
> **状态**: active
> **更新日期**: 2026-07-14

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 静态场景资产

### E2E.P0.098

- [x] The tracked flow requires real frontend/backend/Mailpit authentication and a real completion request.
- [x] The assertion contract requires reload-visible `done,current,pending` across Home、Workspace and TargetJob detail.
- [x] No application request may be intercepted/fulfilled, and the scenario does not create a round-2 plan.

### E2E.P0.099

- [x] The tracked runbook requires real frontend/backend/provider、authenticated report API and read-only PostgreSQL evidence.
- [x] The capture contract requires exactly six current-run full-page screenshots for en/zh ready reports and honest generating at desktop/mobile.
- [x] The evidence/manual contracts bind current API/DB state and canonical digests, reject stale/cross-run evidence, and require no-OCR visible-state/privacy review.

### E2E.P0.101

- [x] The tracked flow requires real frontend/backend/Mailpit and drives email-code, first profile setup and logout/relogin without request interception.
- [x] Business behavior remains owned by backend-auth/frontend-shell; this suite records only the executable asset and current-run result.

## 当前真实环境运行证据

- [ ] Run `E2E.P0.098` against the current real environment and record current-run PASS.
- [ ] Run `E2E.P0.099` against the current real environment, complete the exact-six no-OCR audit and record current-run PASS.
- [ ] Run `E2E.P0.101` against the current real environment and record current-run PASS.

本轮只审计静态资产与证据合同，没有执行上述真实环境场景；场景状态保持 `Ready`。

## Independent regression gates

- [x] Root `make test` remains the complete backend/frontend unit regression outside E2E；this classification does not claim a scenario run.
- [x] Codegen, migration, lint, build and provider reliability/eval gates are reported separately and never become scenario steps or PASS markers.
