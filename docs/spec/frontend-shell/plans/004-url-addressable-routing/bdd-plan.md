# URL-Addressable Routing BDD Plan

> **版本**: 1.8
> **状态**: completed
> **更新日期**: 2026-07-14

## 1 Scenario Map

| 场景 ID | 场景 | 覆盖 Phase | 对应验收 | Checklist Gate |
|---------|------|------------|----------|----------------|
| E2E.P0.088 | canonical path deep-link / reload / back-forward | Phase 2 + Phase 4 + Phase 11 | C-9 / C-11 | Phase 4.3 / Phase 11.3 |
| E2E.P0.089 | auth pendingAction + URL privacy redline | Phase 3 + Phase 11 | C-2 / C-7 / C-11 | Phase 3.3 / Phase 11.3 |
| E2E.P0.090 | hash routing + unsupported route regression | Phase 1 + Phase 4 + Phase 11 | C-4 / C-9 / C-11 | Phase 4.4 / Phase 11.3 |

## 2 Scenario Details

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.088 | Canonical path deep-link / reload / browser history | Frontend app uses Browser History router; `/reports` may carry a valid、missing or invalid targetJobId while workspace and legacy report params include hostile inputs | Open representative canonical URLs including Reports, reload, navigate, and use browser back/forward | Valid Reports preserves only targetJobId and remains chrome-visible/protected；missing/invalid target replaces to workspace without push/back-loop；report/generating preserve only reportId；Parse strips legacy section；TopBar remains three entries and history has no double-push | `test/scenarios/e2e/p0-088-url-addressable-routing-canonical/` |
| E2E.P0.089 | Auth pendingAction + URL privacy redline | User is unauthenticated and directly opens `/reports?targetJobId=...` with hostile extra/raw/secret params | Complete email-code mock auth and restore the protected route；also exercise hostile history input | Restore reaches Reports with only targetJobId；raw/secret and incompatible params have zero hits in URL、history、pendingAction、storage and logs | `test/scenarios/e2e/p0-089-url-routing-auth-privacy/` |
| E2E.P0.090 | Hash routing + unsupported route regression | Static preview / parity inputs use `#route=reports&targetJobId=...`; legacy `section=reports` and unsupported paths/params are available | Open hash and direct URLs, then run routing/fallback regressions | Reports hash reaches equivalent canonical URL；legacy section/report/status params are stripped；known `/reports` host fallback works；unsupported inputs do not create screens or TopBar entries | `test/scenarios/e2e/p0-090-url-routing-hash-out-of-scope-negative/` |

## 3 Regression References

| 场景 ID | 场景 | 复用目的 | 验证入口 |
|---------|------|----------|----------|
| E2E.P0.001 | 默认首页与三入口 Shell | Prove canonical routing preserves default route and TopBar contract | `test/scenarios/e2e/p0-001-default-home-shell/` |
| E2E.P0.002 | 登录打断后恢复原业务动作 | Prove pendingAction canonical restore remains aligned with D1 auth gate | `test/scenarios/e2e/p0-002-auth-pending-action-resume/` |
| E2E.P0.006 | real-browser ui-design pixel parity gate | Prove the hash adapter continues to serve the pixel parity harness | `test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/` |
