# 001 Real API/UI Journeys BDD Checklist

> **版本**: 4.4
> **状态**: active
> **更新日期**: 2026-07-15

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
- [x] Privacy review rejects project user data and secrets, while benign development metadata is handled by independent file-integrity/digest gates rather than classified as private data.
- [x] The real flow opens report-owned Conversation using reportId only, binds real API rows to the owned report/session/messages in PostgreSQL, verifies strict source ordering and returns to the original Report.
  <!-- verified: 2026-07-15 method=scenario-asset-tdd evidence="test_report_conversation_evidence.py 7 PASS; setup/trigger/verify/cleanup scripts use real host-run API + PostgreSQL and never intercept requests" -->
- [x] Conversation evidence contains no transcript prose and no extra image；the screenshot directory/manifest/manual audit remain exactly six, and network evidence contains zero public session-list requests.
  <!-- verified: 2026-07-15 method=scenario-asset-static evidence="validator enforces exact six screenshots, redacted digest-only evidence and public-session-list count zero" -->

### E2E.P0.101

- [x] The tracked flow requires real frontend/backend/Mailpit and drives email-code, first profile setup and logout/relogin without request interception.
- [x] Business behavior remains owned by backend-auth/frontend-shell; this suite records only the executable asset and current-run result.
- [x] The tracked flow opens Settings through the sole gear, verifies current-run displayName and complete account email from the real backend, rejects the old account dropdown/settings tab surface, and keeps that email redacted from scenario logs/evidence.
- [x] Logout is entered from Settings and relogin skips profile setup；the flow never calls deleteMe and stores no full email/code/cookie evidence.

## 当前真实环境运行证据

- [ ] Run `E2E.P0.098` against the current real environment and record current-run PASS.
- [x] Run `E2E.P0.099` against the current real environment, complete the exact-six no-OCR audit plus bounded conversation/API/DB/back evidence, and record current-run PASS.
  <!-- verified: 2026-07-15 run_id="e2e-p0-099-20260715T021319Z-57232" result="PASS" evidence="exact six Chrome full-page screenshots; no-OCR visual audit; live API/PostgreSQL/conversation/back binding; bounded redaction" -->
- [x] Run `E2E.P0.101` against the current real environment and record current-run PASS.
  <!-- verified: 2026-07-15 run_id="e2e-p0-101-20260715114513-19516" result="PASS" evidence="real Mailpit lifecycle, full Settings email, logout/relogin, deleteMe=0, plain+URL-encoded email redaction, cleanup PASS" -->

本轮已在当前 host-run 真实环境完成 P0.099 与 P0.101 的 setup → trigger → verify → cleanup，current-run 结果均为 `PASS`；P0.101 另经 Chrome 1440/390 验证完整邮箱、单设置齿轮与零横向溢出。P0.098 仍保持未运行。

## Independent regression gates

- [x] Root `make test` remains the complete backend/frontend unit regression outside E2E；this classification does not claim a scenario run.
- [x] Codegen, migration, lint, build and provider reliability/eval gates are reported separately and never become scenario steps or PASS markers.
