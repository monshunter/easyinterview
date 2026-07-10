# URL-Addressable Routing BDD Plan

> **版本**: 1.7
> **状态**: completed
> **更新日期**: 2026-07-10

## 1 Scenario Map

| 场景 ID | 场景 | 覆盖 Phase | 对应验收 | Checklist Gate |
|---------|------|------------|----------|----------------|
| E2E.P0.088 | canonical path deep-link / reload / back-forward | Phase 2 + Phase 4 | C-9 | Phase 4.3 |
| E2E.P0.089 | auth pendingAction + URL privacy redline | Phase 3 | C-2 / C-7 | Phase 3.3 |
| E2E.P0.090 | hash routing + unsupported route regression | Phase 1 + Phase 4 | C-4 / C-9 | Phase 4.4 |

## 2 Scenario Details

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.088 | Canonical path deep-link / reload / browser history | Frontend app uses Browser History router; workspace input carries hostile detail/start params, while practice/generating/report/resume workshop inputs carry route-specific params | Open representative canonical URLs, reload, navigate across current routes, and use browser back/forward | Workspace canonicalizes to a query-free list route; other current routes preserve only their own safe params; TopBar/chrome state, reload and back/forward behavior remain correct without double-push | `test/scenarios/e2e/p0-088-url-addressable-routing-canonical/` |
| E2E.P0.089 | Auth pendingAction + URL privacy redline | User is unauthenticated; the positive practice workflow contains safe IDs, while hostile workspace inputs carry route-incompatible params plus representative raw and secret markers | User triggers the auth-gated practice action, completes email-code mock auth, returns to practice, then opens hostile workspace auth/history inputs | Practice restore preserves its safe handoff IDs; hostile workspace direct-open and popstate inputs normalize to query-free workspace state; raw and secret markers have zero hits in URL, history, pendingAction, storage and logs | `test/scenarios/e2e/p0-089-url-routing-auth-privacy/` |
| E2E.P0.090 | Hash routing + unsupported route regression | Static preview / pixel parity inputs use `#route=...`; unsupported paths and hash routes are available | Open representative hash URLs and unsupported direct URLs, then run routing/fallback regressions | Hash routes bootstrap through `normalizeRoute` and land on equivalent canonical state; unsupported inputs normalize to current routes or Home; canonical output never emits unsupported paths; frontend fallback does not swallow API/static/script paths | `test/scenarios/e2e/p0-090-url-routing-hash-out-of-scope-negative/` |

## 3 Regression References

| 场景 ID | 场景 | 复用目的 | 验证入口 |
|---------|------|----------|----------|
| E2E.P0.001 | 默认首页与三入口 Shell | Prove canonical routing preserves default route and TopBar contract | `test/scenarios/e2e/p0-001-default-home-shell/` |
| E2E.P0.002 | 登录打断后恢复原业务动作 | Prove pendingAction canonical restore remains aligned with D1 auth gate | `test/scenarios/e2e/p0-002-auth-pending-action-resume/` |
| E2E.P0.006 | real-browser ui-design pixel parity gate | Prove the hash adapter continues to serve the pixel parity harness | `test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/` |
