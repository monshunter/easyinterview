# URL-Addressable Routing Checklist

> **版本**: 1.16
> **状态**: completed
> **更新日期**: 2026-07-19

**关联计划**: [plan](./plan.md)

## Phase 1: Route codec and canonical bootstrap

- [x] 1.1 Implement route-to-URL table for every current `Route.name`; verification: routeUrl tests cover serialize/parse for primary nav, context routes, user-menu routes and auth routes.
- [x] 1.2 Implement shared safe-param allowlist; verification: routeUrl and pendingAction tests prove legal handoff params survive while unknown/raw payload params are dropped.
- [x] 1.3 Historical implementation evidence: the former Demo bootstrap adapter normalized hash inputs；Phase 14 supersedes this contract after the Demo/toolchain removal.

## Phase 2: Browser History integration

- [x] 2.1 Move initial route bootstrap to browser-aware route store; verification: routeStore tests cover priority order and default fallback.
- [x] 2.2 Keep `NavigationProvider.navigate(next)` API stable; verification: AppRoutingHistory tests cover push/replace, no duplicate push and TopBar active route.
- [x] 2.3 Implement `popstate` handling; verification: AppRoutingHistory / AppRoutingPrivacy tests cover route params, chrome state, InterviewContext hydration and hostile history cleanup.

## Phase 3: Auth and privacy

- [x] 3.1 Restore auth pendingAction through canonical route; verification: pendingAction tests cover safe route identity and param filtering.
- [x] 3.2 Add URL/privacy redline tests; verification: AppRoutingPrivacy and pendingAction tests capture URL, history state, pendingAction, storage, console and mock transport logs.

## Phase 4: Host fallback and routing regression

- [x] 4.1 Add host fallback coverage; verification: spaFallback tests prove known frontend paths return `index.html` and API/static/script paths are not swallowed.
- [x] 4.2 Add unsupported route regression; verification: outOfScopeRouteNegative and routeUrl tests prove unsupported inputs do not become canonical paths or materialized screens.

## Phase 5: Handoff and closeout

- [x] 5.1 Update implementation handoff docs only where operators need route/fallback guidance.
- [x] 5.2 Run route codec, browser history, auth privacy, host fallback and scenario focused gates.
- [x] 5.3 Reconcile spec/history/plan/checklist/context/BDD/index files to current implementation evidence before completion.

## Phase 14: UI Demo pruning owner reconcile

- [x] 14.1 Supersede the historical hash-adapter contract with canonical path/query routing and remove deleted browser-harness discovery from plan/checklist/context.（验证：owner context PASS；deleted path/symbol search=0）
- [x] 14.2 Verify route-store fragment rejection, canonical replacement, owner context, docs links and active UI Demo residual gates before restoring `completed`.（验证：route focused=`4 files / 42 tests`；004 context PASS；deleted symbol search=0；docs links PASS；active residuals=0）





## Phase 8: route-table evidence reconciliation

- [x] 8.1 Align the owner route table and context discovery with the current `routeUrl.ts` safe-param sets.
  <!-- verified: 2026-07-10 method=route-table-evidence-reconciliation evidence="Owner table now matches WORKSPACE_SAFE empty and RESUME_VERSIONS_SAFE resumeId/flow/createMode/targetJobId; stale auto-start/tailor/voice discovery keywords were replaced with current workspace-strip and phone-mode terms." -->



## Phase 10: unconsumed route helper removal

- [x] 10.1 Add a focused source-surface RED assertion for the route helper with zero repository consumers.
- [x] 10.2 Delete `routeUrlsEqual` and its false route-store consumer comment without adding a replacement.
  <!-- verified: 2026-07-10 method=unconsumed-route-helper-removal evidence="Deleted the wrapper and comment with no replacement. Route codec/store/history pass 3 files/55 tests and non-test frontend symbol inventory is empty; formatRouteUrl and routeStore code are unchanged." -->
- [x] 10.3 Run focused route codec/store tests, typecheck, symbol inventory, owner/product contexts, docs, diff and pruning gates; then restore the owner to `completed`.

## Phase 11: target-scoped Reports route

- [x] 11.1 RED-GREEN: add `reports` to context routes and canonical `/reports` with `targetJobId` as its only safe param；keep chrome visible and `PRIMARY_NAV_ROUTES` / TopBar unchanged at three entries.
- [x] 11.2 RED-GREEN: Parse drops hostile `section=reports`/report/status/round inputs；report and generating serialize/restore reportId only；missing/invalid Reports target uses replace-only workspace recovery with no push/back-loop, and a child mount redirect cannot be overwritten by stale bootstrap canonicalization.

## Phase 12: `/workspace` list/detail and command-only Parse

- [x] 12.1 RED-GREEN route codec: allow only optional `targetJobId` on workspace; keep query-free `/workspace` as list and `/workspace?targetJobId=<uuid>` as detail; strip `planId`、`resumeId`、`autoStartPractice` and unknown/raw inputs.
- [x] 12.2 RED-GREEN Parse route: keep only `targetJobId`; define it as queued/processing command progress, and prove ready initial read or poll transition calls replace—not push—to workspace detail.
- [x] 12.3 Update pendingAction/hash/history/host-fallback matrices so workspace and Parse restore targetJobId only; Back after ready replace must not return to Parse animation.
- [x] 12.5 Run focused route/store/App/auth/privacy/fallback tests, root `make test`, frontend typecheck/build, owner contexts, docs/diff and old workspace-zero-query positive-claim searches before restoring `completed`; code tests must not enter `test/scenarios/e2e/`.

## BDD Gate

- [x] BDD-Gate: `BDD.SHELL.ROUTING.001` 由 [BDD checklist](./bdd-checklist.md) 关联 URL/history/auth-guard owner behavior tests；不创建或声明真实 E2E PASS。

## Phase 13: Practice chrome visibility

- [x] 13.1 RED-GREEN: `isChromeHidden("practice")` 从 true 改为 false，`generating` 保持 true；direct/reload/back/forward 与 App tests 覆盖相同结果。
- [x] 13.2 CROSS-OWNER BDD: `frontend-workspace-and-practice/001` 的 `BDD.PRACTICE.GLOBAL_CHROME.005` 承接用户行为；本 owner 只验证 route/chrome codec，不新增 E2E。
- [x] 13.3 REGRESSION: focused route/App、根 `make test`、typecheck/build、contexts/docs/diff 与 old positive claim 搜索通过后恢复 completed。
## Phase 15: Generating shared chrome correction

- [x] 15.1 RED: route/App history tests 将 Generating 从 hidden chrome 改为 visible，并证明当前 `NO_CHROME_ROUTES` 实现失败。<!-- verified: 2026-07-19 method=focused-vitest-red evidence="canonical direct-open and popstate tests failed on absent TopBar; TopBar unit failed isChromeHidden(generating)=true" -->
- [x] 15.2 GREEN: 删除 Generating no-chrome 例外；保留 canonical path/query、auth guard、initial open、refresh 与 Back/Forward 合同。<!-- verified: 2026-07-19 method=focused-vitest-green evidence="TopBar 17, canonical routing 11 and App history 15 tests PASS" -->
- [x] 15.3 BDD-Gate: `BDD.SHELL.ROUTING.001` 覆盖所有 canonical route 的共享 TopBar 和安全恢复，不新增业务 route。<!-- verified: 2026-07-19 evidence="App/route/TopBar focused suites PASS; current-run Practice, Resume, Parse and Generating all retained shared chrome and context-route Interview highlight." -->
- [x] 15.4 REGRESSION: focused routing、App、typecheck/build、根 `make test` 与 context/docs/diff gates 通过后恢复 completed。<!-- verified: 2026-07-19 evidence="Final focused 89 PASS; production build PASS; root make test 615 / 4615 PASS; context/docs/index/diff gates pass." -->
