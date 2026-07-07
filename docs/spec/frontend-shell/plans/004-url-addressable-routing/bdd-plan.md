# URL-Addressable Routing BDD Plan

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-07-07

## 1 Scenario Map

| 场景 ID | 场景 | 覆盖 Phase | 对应验收 | Checklist Gate |
|---------|------|------------|----------|----------------|
| E2E.P0.088 | canonical path deep-link / reload / back-forward | Phase 2 + Phase 4 | C-9 | Phase 4.3 |
| E2E.P0.089 | auth pendingAction + URL privacy redline | Phase 3 | C-2 / C-7 | Phase 3.3 |
| E2E.P0.090 | hash compatibility + unsupported route regression | Phase 1 + Phase 4 | C-4 / C-9 | Phase 4.4 |

## 2 Scenario Details

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.088 | Canonical path deep-link / reload / browser history | Frontend app uses Browser History router; fixture-backed mock runtime is available; safe IDs exist for workspace, practice, generating, report and resume workshop paths | Open representative canonical URLs, reload, navigate across current routes, and use browser back/forward | Current URLs parse to expected `Route` + safe params; TopBar active route and chrome state match route catalog; InterviewContext hydrates; reload/back/forward preserve safe context without double-push behavior | `test/scenarios/e2e/p0-088-url-addressable-routing-canonical/` |
| E2E.P0.089 | Auth pendingAction + URL privacy redline | User is unauthenticated; workflow URLs contain safe IDs; test harness seeds representative raw and secret markers outside the safe-param allowlist | User triggers auth-gated actions, completes email-code mock auth, returns to the target route, and restores hostile browser history input through `popstate` | Restored route and canonical URL contain only current route names and safe IDs/hints; raw and secret markers have zero hits in URL, history, pendingAction, storage and logs; hostile entries are replaced with canonical safe state | `test/scenarios/e2e/p0-089-url-routing-auth-privacy/` |
| E2E.P0.090 | Hash compatibility + unsupported route regression | Static preview / pixel parity inputs still use `#route=...`; unsupported paths and hash routes are available | Open representative hash URLs and unsupported direct URLs, then run routing/fallback regressions | Hash routes bootstrap through `normalizeRoute` and land on equivalent canonical state; unsupported inputs normalize to current routes or Home; canonical output never emits unsupported paths; frontend fallback does not swallow API/static/script paths | `test/scenarios/e2e/p0-090-url-routing-hash-non-current-negative/` |

## 3 Regression References

| 场景 ID | 场景 | 复用目的 | 验证入口 |
|---------|------|----------|----------|
| E2E.P0.001 | 默认首页与三入口 Shell | Prove canonical routing preserves default route and TopBar contract | `test/scenarios/e2e/p0-001-default-home-shell/` |
| E2E.P0.002 | 登录打断后恢复原业务动作 | Prove pendingAction canonical restore remains compatible with D1 auth gate | `test/scenarios/e2e/p0-002-auth-pending-action-resume/` |
| E2E.P0.006 | real-browser ui-design pixel parity gate | Prove the hash adapter remains compatible with pixel parity harness | `test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/` |
