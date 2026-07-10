# URL-Addressable Routing BDD Checklist

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.088 canonical path deep-link / reload / browser history

- [x] 创建场景目录 `test/scenarios/e2e/p0-088-url-addressable-routing-canonical/`
- [x] 准备测试数据：workspace / practice / generating / report / resume workshop 的 stable safe ID combinations、auth states and direct-open frontend paths
- [x] 实现 setup / trigger / verify / cleanup；trigger 覆盖 direct open、reload、App navigation、back、forward、unknown/malformed query fallback and current handoff keys
- [x] 验证 TopBar active route、chrome hidden state、InterviewContext hydration、URL canonical output、safe params survival and no double-push behavior
- [x] 执行并通过场景验证，记录 trigger.log and route-state evidence

## E2E.P0.089 auth pendingAction + URL privacy redline

- [x] 创建场景目录 `test/scenarios/e2e/p0-089-url-routing-auth-privacy/`
- [x] 准备隐私标记数据：raw content markers、AI output markers、auth secret markers and hostile history markers
- [x] 实现 auth-gated workflow：未登录打开 canonical workflow URL、触发 protected action、完成 email-code mock auth、恢复原 route，并处理 hostile direct-open / popstate input
- [x] 捕获 URL、history.state、pendingAction、localStorage、sessionStorage、console and mock transport logs，并断言 raw/secret markers zero-hit，同时证明 legal handoff keys 未被 allowlist 误删
- [x] 执行并通过场景验证，记录 restored route、safe params and zero-hit evidence

## E2E.P0.090 hash compatibility + unsupported route regression

- [x] 创建场景目录 `test/scenarios/e2e/p0-090-url-routing-hash-out-of-scope-negative/`
- [x] 准备 hash inputs：current route hashes、unknown hash and unsupported route inputs
- [x] 验证 hash adapter：static preview / pixel parity entrypoint can bootstrap and normal browser mode reaches canonical route/path
- [x] 验证 unsupported route regression：unsupported inputs do not produce canonical paths, screen files, TopBar entries or materialized routes
- [x] 验证 host fallback：known frontend paths return `index.html`; API/static/script paths are not swallowed
- [x] 执行并通过场景验证，记录 routing and fallback evidence
