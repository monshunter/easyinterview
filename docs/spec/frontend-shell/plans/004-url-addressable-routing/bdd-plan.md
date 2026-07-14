# URL-Addressable Routing BDD Plan

> **版本**: 1.9
> **状态**: completed
> **更新日期**: 2026-07-14

## 1 Scenario Map

| 场景 ID | 场景 | 覆盖 Phase | 对应验收 | Checklist Gate |
|---------|------|------------|----------|----------------|
| E2E.P0.088 | canonical path deep-link / reload / back-forward | Phase 2 + Phase 4 + Phase 11 + Phase 12 | C-9 / C-11 / C-13 | Phase 4.3 / Phase 11.3 / Phase 12.4 |
| E2E.P0.089 | auth pendingAction + URL privacy redline | Phase 3 + Phase 11 + Phase 12 | C-2 / C-7 / C-11 / C-13 | Phase 3.3 / Phase 11.3 / Phase 12.4 |
| E2E.P0.090 | hash routing + unsupported route regression | Phase 1 + Phase 4 + Phase 11 + Phase 12 | C-4 / C-9 / C-11 / C-13 | Phase 4.4 / Phase 11.3 / Phase 12.4 |

## 2 Scenario Details

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.088 | Canonical path deep-link / reload / browser history | Frontend app uses Browser History router; workspace may be query-free list or carry one valid targetJobId; Parse target can be queued/processing/ready; Reports may carry a valid、missing or invalid targetJobId | Open representative canonical URLs, reload, navigate and use back/forward; let a Parse target become ready | Workspace preserves targetJobId only for detail and remains list without it；ready Parse uses replace to workspace detail so Back does not replay animation；Reports behavior stays targetJobId-only；report/generating remain reportId-only；history has no double-push | `test/scenarios/e2e/p0-088-url-addressable-routing-canonical/` |
| E2E.P0.089 | Auth pendingAction + URL privacy redline | User is unauthenticated and directly opens workspace/Parse/Reports with targetJobId plus hostile extra/raw/secret params | Complete email-code mock auth and restore the protected route；also exercise hostile history input | Workspace/Parse/Reports restore only targetJobId；raw/secret、planId、resumeId、auto-start and incompatible params have zero hits in URL、history、pendingAction、storage and logs | `test/scenarios/e2e/p0-089-url-routing-auth-privacy/` |
| E2E.P0.090 | Hash routing + unsupported route regression | Static preview / parity inputs use workspace/parse/reports hashes with targetJobId; legacy section and incompatible plan/resume/auto-start params are available | Open hash and direct URLs, then run routing/fallback regressions | Workspace/Parse/Reports hashes reach equivalent targetJobId-only canonical URLs；known frontend paths fall back correctly；legacy/incompatible params are stripped；unsupported inputs do not create screens or TopBar entries | `test/scenarios/e2e/p0-090-url-routing-hash-out-of-scope-negative/` |

## 3 Regression References

| 场景 ID | 场景 | 复用目的 | 验证入口 |
|---------|------|----------|----------|
| E2E.P0.001 | 默认首页与三入口 Shell | Prove canonical routing preserves default route and TopBar contract | `test/scenarios/e2e/p0-001-default-home-shell/` |
| E2E.P0.002 | 登录打断后恢复原业务动作 | Prove pendingAction canonical restore remains aligned with D1 auth gate | `test/scenarios/e2e/p0-002-auth-pending-action-resume/` |
| E2E.P0.006 | real-browser ui-design pixel parity gate | Prove the hash adapter continues to serve the pixel parity harness | `test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/` |
