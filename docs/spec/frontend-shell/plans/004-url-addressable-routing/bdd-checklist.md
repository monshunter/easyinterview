# URL-Addressable Routing BDD Checklist

> **版本**: 1.8
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.088 canonical path deep-link / reload / browser history

- [x] 创建场景目录 `test/scenarios/e2e/p0-088-url-addressable-routing-canonical/`
- [x] 准备测试数据：workspace hostile detail/start params，以及 practice / generating / report / resume workshop 的 current route-specific params
- [x] 实现 setup / trigger / verify / cleanup；trigger 覆盖 query-free workspace canonicalization、direct open、reload、App navigation、back、forward and unknown query filtering
- [x] 验证 TopBar active route、chrome hidden state、URL canonical output、route-specific safe-param survival, workspace param stripping and no double-push behavior
- [x] 执行并通过场景验证，记录 trigger.log and route-state evidence
- [x] Revision 2026-07-14 adds `/reports?targetJobId=...` direct/reload/navigation/back-forward、chrome-visible/protected behavior、targetJobId-only serialization，and reportId-only Report/Generating plus Parse no-section negatives.
  <!-- verified: 2026-07-14 scenario=E2E.P0.088 evidence="Current wrapper passes 86 tests including direct/reload/navigation/history and replace-only invalid-target recovery without a back loop." -->

## E2E.P0.089 auth pendingAction + URL privacy redline

- [x] 创建场景目录 `test/scenarios/e2e/p0-089-url-routing-auth-privacy/`
- [x] 准备隐私标记数据：raw content markers、AI output markers、auth secret markers and hostile history markers
- [x] 实现 auth-gated workflow：未登录打开 canonical workflow URL、触发 protected action、完成 email-code mock auth、恢复原 route，并处理 hostile direct-open / popstate input
- [x] 捕获 URL、history.state、pendingAction、localStorage、sessionStorage、console and mock transport logs，并断言 raw/secret markers zero-hit，同时证明 legal handoff keys 未被 allowlist 误删
- [x] 执行并通过场景验证，记录 restored route、safe params and zero-hit evidence
- [x] 校准场景证据：practice auth continuation 保留 safe handoff params；hostile workspace direct-open / popstate 均规范化为 query-free `/workspace`
  <!-- verified: 2026-07-10 method=p0-089-workspace-zero-query-reconciliation evidence="BDD details, scenario README/seed/expected outcome and the positive Vitest title now distinguish practice safe-param restore from hostile workspace query stripping." -->
- [x] Revision 2026-07-14 proves unauthenticated Reports deep link restores only targetJobId after auth；hostile section/report/status/round/raw/secret markers are absent from URL、history、pendingAction、storage and logs.
  <!-- verified: 2026-07-14 scenario=E2E.P0.089 evidence="Current wrapper passes 15 tests for protected Reports auth continuation, targetJobId-only restoration and privacy negatives." -->

## E2E.P0.090 hash routing + unsupported route regression

- [x] 创建场景目录 `test/scenarios/e2e/p0-090-url-routing-hash-out-of-scope-negative/`
- [x] 准备 hash inputs：current route hashes、unknown hash and unsupported route inputs
- [x] 验证 hash adapter：static preview / pixel parity entrypoint can bootstrap and normal browser mode reaches canonical route/path
- [x] 验证 unsupported route regression：unsupported inputs do not produce canonical paths, screen files, TopBar entries or materialized routes
- [x] 验证 host fallback：known frontend paths return `index.html`; API/static/script paths are not swallowed
- [x] 执行并通过场景验证，记录 routing and fallback evidence
- [x] Revision 2026-07-14 covers `#route=reports` canonical bootstrap、known `/reports` SPA fallback、legacy `section=reports` stripping and zero Reports TopBar entry.
  <!-- verified: 2026-07-14 scenario=E2E.P0.090 evidence="Current wrapper passes 87 tests for hash canonicalization, host fallback, legacy section stripping and TopBar negative." -->
