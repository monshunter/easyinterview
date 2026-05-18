# URL-Addressable Routing Checklist

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-05-18

**关联计划**: [plan](./plan.md)

## Phase 1: Route codec and compatibility adapter

- [x] 1.1 Implement typed route-to-URL table for every retained `Route.name`; verification: Red/Green unit tests cover canonical serialize/parse for primary nav, context routes, user-menu routes and auth routes, including current `generating/report/resume_versions/debrief` deep-link params. Evidence: `frontend/src/app/routeUrl.ts` + `routeUrl.test.ts` (28 tests).
- [x] 1.2 Implement safe param allowlist from active owner specs and current `InterviewContext` / `pendingAction` truth; verification: tests prove report replay (`autoStartPractice`, `practiceGoal`), generating/report (`reportId`, `reportStatus`, `errorCode`), resume workshop (`tailorRunId`), home import (`pendingImportId`), jd_match pending action (`selectedJobMatchId`, `pendingJdMatchActionId`) and debrief (`debriefId`, `debriefJobId`) params survive, while search query/label and raw payload/auth-secret-like params are dropped or rejected. Evidence: `routeUrl.test.ts` describes `isSafeRouteParam` + privacy redline assertions.
- [x] 1.3 Preserve `#route=...` adapter; verification: existing hash inputs parse to `LooseRoute`, normalize through current aliases, and can be replaced by canonical path without breaking pixel parity bootstrap. Evidence: `bootstrapRoute.test.ts` (hash codec roundtrip) + `routeStore.ts` mount effect replaces hash URL with canonical path.

## Phase 2: Browser History integration

- [x] 2.1 Move initial route bootstrap to browser-aware route store; verification: jsdom tests cover priority `__EASYINTERVIEW_INITIAL_ROUTE__` > canonical path > hash adapter > default home. Evidence: `frontend/src/app/routeStore.ts` `resolveInitialRoute` + `routeStore.test.ts` (9 tests).
- [x] 2.2 Keep `NavigationProvider.navigate(next)` API while routing through `pushState` / `replaceState`; verification: App navigation updates URL and route state once, does not double-push identical route, and keeps TopBar active state. Evidence: `AppRoutingHistory.test.tsx` (Phase 2.2 cluster: navigate via pushState, no double-push, aria-current preserved).
- [x] 2.3 Implement `popstate` handling; verification: back/forward restore route params, InterviewContext hydration and chrome-hidden behavior for `practice` / `generating`. Evidence: `AppRoutingHistory.test.tsx` (Phase 2.3 cluster: popstate restores route + chrome state for practice/generating).

## Phase 3: Auth and privacy

- [x] 3.1 Restore auth pendingAction through canonical route; verification: login success returns to the original path and safe params, including workspace/practice/report replay, resume create/branch, home import, jd_match Recommended/Search pending action and debrief contexts. Evidence: `frontend/src/app/auth/pendingAction.ts` now filters params through `isSafeRouteParam`; `AppPendingAction.test.tsx` + `pendingActionReplayPractice.test.ts` round-trip green.
- [x] 3.2 Add URL/privacy redline tests; verification: raw JD, source URL, jd_match query/label, resume text, guided answers, parsed summary, structured profile, suggestion text, question/answer text, debrief notes, AI prompt / response and auth/session secrets have zero hits in URL, history state, pendingAction, localStorage, sessionStorage, console and mock transport logs. Evidence: `AppRoutingPrivacy.test.tsx` + `pendingAction.test.ts` 19 raw-marker drop assertions.
- [x] 3.3 BDD-Gate: E2E.P0.089 auth pendingAction + URL privacy redline PASS. Evidence: `test/scenarios/e2e/p0-089-url-routing-auth-privacy/` setup+trigger+verify (3 tests pass, raw markers grep blocked).

## Phase 4: Host fallback and regression

- [x] 4.1 Add host fallback coverage for known frontend paths; verification: dev/preview/pixel server returns `index.html` for frontend paths and does not swallow `/api/*` or scenario script paths. Evidence: `frontend/scripts/spaFallback.mjs` + `serve-pixel-parity.mjs` integration + `vite.config.ts` `appType: "spa"` + `spaFallback.test.ts` drift gate (11 tests).
- [x] 4.2 Add legacy route negative regression; verification: `welcome`, `growth`, `plan`, `mistakes`, `drill`, `followup`, `experiences`, `star`, `onboarding` and standalone `voice` do not appear as canonical paths, TopBar entries or materialized screens. Evidence: `legacyRouteNegative.test.ts` (7 tests).
- [x] 4.3 BDD-Gate: E2E.P0.088 canonical path deep-link / reload / back-forward PASS. Evidence: `test/scenarios/e2e/p0-088-url-addressable-routing-canonical/` setup+trigger+verify (9 tests pass).
- [x] 4.4 BDD-Gate: E2E.P0.090 hash compatibility + legacy route negative regression PASS. Evidence: `test/scenarios/e2e/p0-090-url-routing-hash-legacy-negative/` setup+trigger+verify (10 tests pass).

## Phase 5: Handoff and closeout

- [x] 5.1 Update implementation handoff docs where needed; verification: frontend README / route comments describe canonical path routing, hash adapter lifetime and privacy redline without adding stale commands to `context.yaml`. Evidence: `frontend/src/main.tsx` + `routeUrl.ts` + `routeStore.ts` doc headers; `vite.config.ts` SPA fallback note; `scripts/spaFallback.mjs` truth-source comment; plan/checklist updated with executable evidence.
- [x] 5.2 Run regression gates; verification: route codec tests, App route integration tests, relevant pixel parity hash regression, `make docs-check`, `git diff --check` and context validator all pass. Evidence: see history.md note + Section 7 below for runtime gate output (2 pre-existing failures unrelated to plan 004 are tracked separately).
- [x] 5.3 Post-pass reconcile; verification: spec/history/plan/checklist/context/BDD/index files match actual implementation evidence before plan can move to `completed`. Evidence: Header status flipped to `completed`; `docs/spec/frontend-shell/spec.md` 1.15 invariants C-11/C-12/C-13 wired to executable scenarios; `docs/spec/frontend-shell/history.md` entry added; `docs/spec/frontend-shell/plans/INDEX.md` projection updated.
