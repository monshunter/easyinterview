# URL-Addressable Routing BDD Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-18

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.088 canonical path deep-link / reload / browser history

- [ ] 创建场景目录 `test/scenarios/e2e/p0-088-url-addressable-routing-canonical/`
- [ ] 准备测试数据：workspace / practice / generating / report / resume workshop / debrief 的稳定 safe ID 组合、unauthenticated 与 authenticated mock session、direct-open frontend paths
- [ ] 实现 setup / trigger / verify / cleanup：trigger 覆盖 direct open、reload、App navigation、back、forward、unknown/malformed query fallback，以及 `autoStartPractice` / `reportId` / `reportStatus` / `tailorRunId` / `debriefId` / `debriefJobId` 当前 handoff keys
- [ ] 验证 TopBar active route、chrome hidden state、InterviewContext hydration、URL canonical output、cross-owner safe params survival 和 history length / no double-push 行为
- [ ] 执行并通过场景验证，记录 trigger.log 与 route-state evidence

## E2E.P0.089 auth pendingAction + URL privacy redline

- [ ] 创建场景目录 `test/scenarios/e2e/p0-089-url-routing-auth-privacy/`
- [ ] 准备隐私标记数据：raw JD、source URL、jd_match search query/label、resume text、guided answers、parsed summary、structured profile、suggestion text、question/answer text、debrief notes、AI prompt / response marker、auth secret marker
- [ ] 实现 auth-gated workflow：未登录打开 canonical workflow URL、触发 workspace auto-start / report replay / home import / jd_match Recommended/Search pending action / resume workshop / debrief login、完成 mock passwordless、恢复原 route
- [ ] 捕获 URL、history.state、pendingAction、localStorage、sessionStorage、console、mock transport logs，并对隐私标记做 zero-hit 断言，同时证明合法 handoff keys（含 `selectedJobMatchId` / `pendingJdMatchActionId`）未被 allowlist 误删
- [ ] 执行并通过场景验证，记录 restored route、safe params、合法 handoff keys 和 zero-hit grep evidence

## E2E.P0.090 hash compatibility + legacy route negative regression

- [ ] 创建场景目录 `test/scenarios/e2e/p0-090-url-routing-hash-legacy-negative/`
- [ ] 准备 hash inputs：`#route=home`、`#route=workspace&targetJobId=...`、`#route=practice&mode=voice&modality=voice`、unknown hash、legacy aliases 和 standalone `voice`
- [ ] 验证 hash adapter：static preview / pixel parity entrypoint 仍能 bootstrap，且正常浏览器模式下得到 canonical route/path
- [ ] 验证 legacy negative：retired aliases 不产生 canonical paths、screen files、TopBar entries、scenario names 或 standalone `voice` route
- [ ] 验证 host fallback：known frontend paths 返回 `index.html`，`/api/*` 与 scenario script paths 不被 frontend fallback 吞掉
- [ ] 执行并通过场景验证，记录 retired-route grep 和 fallback evidence
