# URL-Addressable Routing Checklist

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-07-07

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
- [x] 4.2 Add unsupported route regression; verification: nonCurrentRouteNegative and routeUrl tests prove unsupported inputs do not become canonical paths or materialized screens.
- [x] 4.3 BDD-Gate: `E2E.P0.088` canonical path deep-link / reload / back-forward passes.
- [x] 4.4 BDD-Gate: `E2E.P0.090` hash compatibility + unsupported route regression passes.

## Phase 5: Handoff and closeout

- [x] 5.1 Update implementation handoff docs only where operators need route/fallback guidance.
- [x] 5.2 Run route codec, browser history, auth privacy, host fallback and scenario focused gates.
- [x] 5.3 Reconcile spec/history/plan/checklist/context/BDD/index files to current implementation evidence before completion.
