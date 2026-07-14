# URL-Addressable Routing Checklist

> **版本**: 1.12
> **状态**: completed
> **更新日期**: 2026-07-14

**关联计划**: [plan](./plan.md)

## Phase 1: Route codec and hash adapter

- [x] 1.1 Implement route-to-URL table for every current `Route.name`; verification: routeUrl tests cover serialize/parse for primary nav, context routes, user-menu routes and auth routes.
- [x] 1.2 Implement shared safe-param allowlist; verification: routeUrl and pendingAction tests prove legal handoff params survive while unknown/raw payload params are dropped.
- [x] 1.3 Preserve `#route=...` adapter; verification: bootstrapRoute tests prove hash inputs normalize through the same route contract and normal browser mode can replace with canonical path.

## Phase 2: Browser History integration

- [x] 2.1 Move initial route bootstrap to browser-aware route store; verification: routeStore tests cover priority order and default fallback.
- [x] 2.2 Keep `NavigationProvider.navigate(next)` API stable; verification: AppRoutingHistory tests cover push/replace, no duplicate push and TopBar active route.
- [x] 2.3 Implement `popstate` handling; verification: AppRoutingHistory / AppRoutingPrivacy tests cover route params, chrome state, InterviewContext hydration and hostile history cleanup.

## Phase 3: Auth and privacy

- [x] 3.1 Restore auth pendingAction through canonical route; verification: pendingAction tests cover safe route identity and param filtering.
- [x] 3.2 Add URL/privacy redline tests; verification: AppRoutingPrivacy and pendingAction tests capture URL, history state, pendingAction, storage, console and mock transport logs.
- [x] 3.3 BDD-Gate: `E2E.P0.089` auth pendingAction + URL privacy redline passes.

## Phase 4: Host fallback and routing regression

- [x] 4.1 Add host fallback coverage; verification: spaFallback tests prove known frontend paths return `index.html` and API/static/script paths are not swallowed.
- [x] 4.2 Add unsupported route regression; verification: outOfScopeRouteNegative and routeUrl tests prove unsupported inputs do not become canonical paths or materialized screens.
- [x] 4.3 BDD-Gate: `E2E.P0.088` canonical path deep-link / reload / back-forward passes.
- [x] 4.4 BDD-Gate: `E2E.P0.090` hash routing + unsupported route regression passes.

## Phase 5: Handoff and closeout

- [x] 5.1 Update implementation handoff docs only where operators need route/fallback guidance.
- [x] 5.2 Run route codec, browser history, auth privacy, host fallback and scenario focused gates.
- [x] 5.3 Reconcile spec/history/plan/checklist/context/BDD/index files to current implementation evidence before completion.

## Phase 6: Hash routing current-contract wording

- [x] 6.1 Replace old hash labels in frontend tests, owner/BDD docs and E2E.P0.090 assets with current hash adapter/routing wording; preserve behavior and rerun focused routing gates.
  <!-- verified: 2026-07-10 commands="focused routeUrl/bootstrapRoute/scope/P0.090 Vitest; frontend-shell and product contexts; docs/diff/pruning gates" result="58 tests pass; scoped old-label search zero; real_residuals=0" -->

## Phase 7: URL privacy test lifecycle isolation

- [x] 7.1 P0.089 hostile-query 同步负向用例在断言后显式 unmount，清除无关 runtime provider update（验证：P0.089 focused 无 act warning、routing owner/full frontend tests、build、owner context/docs gates）
  <!-- verified: 2026-07-10 method=url-privacy-test-lifecycle-isolation evidence="Focused red preserved one AppRuntimeProvider act warning while all 3 assertions passed. Added explicit unmount after the final hostile-query privacy assertion without changing routing or production code. P0.089 3 tests and routing owner 12 files/125 tests pass warning-free; frontend build and owner/product contexts pass. Full frontend 137 files/829 tests pass and P0.089 is absent from the remaining warning list; completed-state docs/diff/pruning gates rerun during closeout." -->

## Phase 8: route-table evidence reconciliation

- [x] 8.1 Align the owner route table and context discovery with the current `routeUrl.ts` safe-param sets.
  <!-- verified: 2026-07-10 method=route-table-evidence-reconciliation evidence="Owner table now matches WORKSPACE_SAFE empty and RESUME_VERSIONS_SAFE resumeId/flow/createMode/targetJobId; stale auto-start/tailor/voice discovery keywords were replaced with current workspace-strip and phone-mode terms." -->
- [x] 8.2 Reconcile P0.088 test comments, scenario assets and BDD wording so workspace detail/start params are hostile strip inputs rather than current handoff keys.
  <!-- verified: 2026-07-10 method=route-table-evidence-reconciliation evidence="P0.088 docs now expect workspace-plan-list and query-free /workspace for direct, malformed and hash inputs; route-specific practice/report params remain positive. Focused routeUrl and P0.088 tests pass 44/44." -->
- [x] 8.3 Run focused route/P0.088 tests, the P0.088 wrapper, owner/product contexts, docs, diff and pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=route-table-evidence-reconciliation evidence="Focused routeUrl/P0.088 passes 44/44 and the P0.088 four-stage wrapper passes 16/16. Full scripts/lint, shell/product contexts, docs/index/link/diff and pruning gates pass; no runtime behavior or environment data changed." -->

## Phase 9: P0.089 workspace-zero-query evidence reconciliation

- [x] 9.1 Record the mismatch between current P0.089 executable assertions and scenario prose that still claims workspace query params survive canonicalization.
  <!-- verified: 2026-07-10 method=p0-089-evidence-drift-scan evidence="Executable assertions require hostile auth/login workspace input to drop planId/targetJobId and hostile popstate to become /workspace, while expected-outcome.md and README still claimed those params survived; the positive test title also retained workspace auto-start wording." -->
- [x] 9.2 Align the P0.089 test title, BDD wording, README and data assets with positive practice restore plus hostile query-free workspace normalization.
  <!-- verified: 2026-07-10 method=p0-089-workspace-zero-query-reconciliation evidence="Renamed the positive test to practice pending action; aligned BDD plan/checklist, README, seed and expected outcome so workspace planId/targetJobId are hostile inputs and canonical output is /workspace. Focused P0.089 passes 3/3 warning-free." -->
- [x] 9.3 Run focused P0.089 tests, its four-stage wrapper, owner/product contexts, docs, diff and pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=p0-089-workspace-zero-query-reconciliation evidence="Focused P0.089 and its setup/trigger/verify/cleanup wrapper pass 3/3 warning-free. Stale positive-claim search is empty; shell/product contexts, docs/index/link/diff and pruning gates pass with real_residuals=0. No routing runtime or environment data changed." -->

## Phase 10: unconsumed route helper removal

- [x] 10.1 Add a focused source-surface RED assertion for the route helper with zero repository consumers.
  <!-- verified: 2026-07-10 method=unconsumed-route-helper-source-red evidence="Focused routeUrl failed 1/36 solely because routeUrl.ts still exported routeUrlsEqual; all 35 behavior assertions passed and repository inventory found no consumer." -->
- [x] 10.2 Delete `routeUrlsEqual` and its false route-store consumer comment without adding a replacement.
  <!-- verified: 2026-07-10 method=unconsumed-route-helper-removal evidence="Deleted the wrapper and comment with no replacement. Route codec/store/history pass 3 files/55 tests and non-test frontend symbol inventory is empty; formatRouteUrl and routeStore code are unchanged." -->
- [x] 10.3 Run focused route codec/store tests, typecheck, symbol inventory, owner/product contexts, docs, diff and pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=unconsumed-route-helper-removal evidence="Route codec/store/history pass 3 files/55 tests; P0.088/P0.089/P0.090 pass 3 files/23 tests; typecheck and non-test symbol inventory pass. Shell/product contexts and docs/index/link/diff/pruning gates pass with real_residuals=0." -->

## Phase 11: target-scoped Reports route

- [x] 11.1 RED-GREEN: add `reports` to context routes and canonical `/reports` with `targetJobId` as its only safe param；keep chrome visible and `PRIMARY_NAV_ROUTES` / TopBar unchanged at three entries.
  <!-- verified: 2026-07-14 method=route-tdd evidence="Routing/App/auth/privacy focused suite 152/152 passes with reports as a chrome-visible protected context route, targetJobId-only serialization, App dispatch and an unchanged three-entry TopBar." -->
- [x] 11.2 RED-GREEN: Parse drops hostile `section=reports`/report/status/round inputs；report and generating serialize/restore reportId only；missing/invalid Reports target uses replace-only workspace recovery with no push/back-loop, and a child mount redirect cannot be overwritten by stale bootstrap canonicalization.
  <!-- verified: 2026-07-14 method=route-tdd evidence="Reports/App history/route-store focused suite 40/40 proves replace without push for missing/invalid target plus canonical URL stability; route/privacy regressions keep hostile and legacy business parameters out of URL/history/pendingAction." -->
- [x] 11.3 BDD-Gate: P0.088 covers `/reports` direct/reload/navigation/back-forward；P0.089 covers unauthenticated deep-link auth continuation with targetJobId-only restore and privacy zero-hit；P0.090 covers hash bootstrap、host fallback、unsupported params and TopBar negative.
  <!-- verified: 2026-07-14 evidence="P0.088 86 tests, P0.089 15 tests and P0.090 87 tests pass; invalid/missing reports target uses replaceState, auth restores targetJobId only, hash/host fallback strips unsupported params and TopBar remains three items." -->
- [x] 11.4 POST-PASS: focused route/store/App/auth/privacy/fallback tests、P0.088/P0.089/P0.090、frontend typecheck/build、owner contexts、docs/diff/pruning pass before restoring plan/checklist/BDD to `completed`.
  <!-- verified: 2026-07-14 method=current-aggregate evidence="Focused routing/auth/privacy, full frontend 121 files/977 tests, typecheck/build, all three scenarios, owner context, docs, diff and pruning gates pass." -->

## Phase 12: `/workspace` list/detail and command-only Parse

- [x] 12.1 RED-GREEN route codec: allow only optional `targetJobId` on workspace; keep query-free `/workspace` as list and `/workspace?targetJobId=<uuid>` as detail; strip `planId`、`resumeId`、`autoStartPractice` and unknown/raw inputs.<!-- verified: 2026-07-14 method=vitest-red-green evidence="RED failed 12 route assertions against the former param-free Workspace and multi-param Parse allowlists; GREEN passed routeUrl plus HomeImport 2 files / 51 tests after narrowing both codecs and import handoff." -->
- [x] 12.2 RED-GREEN Parse route: keep only `targetJobId`; define it as queued/processing command progress, and prove ready initial read or poll transition calls replace—not push—to workspace detail.<!-- verified: 2026-07-14 method=vitest-red-green evidence="ParseFlow RED exposed stale preview-delay semantics and cache-driven duplicate GETs; GREEN passed 7 polling/handoff tests plus 2 App detail tests with immediate replace and ref-backed no-N+1 handoff." -->
- [x] 12.3 Update pendingAction/hash/history/host-fallback matrices so workspace and Parse restore targetJobId only; Back after ready replace must not return to Parse animation.
- [x] 12.4 BDD-Gate: update `E2E.P0.088` for workspace list/detail direct/reload/history and Parse ready replace; update `E2E.P0.089` for targetJobId-only auth restore/privacy; update `E2E.P0.090` for hash/fallback and incompatible-param stripping.
- [x] 12.5 Run focused route/store/App/auth/privacy/fallback tests, all three scenario wrappers, frontend typecheck/build, owner contexts, docs/diff and old workspace-zero-query positive-claim searches before restoring `completed`.
  <!-- verified: 2026-07-14 evidence="P0.088/P0.089/P0.090 setup/trigger/verify/cleanup PASS; targetJobId-only workspace/Parse/Reports restore, ready replace history and incompatible-param stripping remain green with full frontend/typecheck/build." -->
