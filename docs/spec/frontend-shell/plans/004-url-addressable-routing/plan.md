# URL-Addressable Routing

> **版本**: 1.12
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

本计划把正式前端 route 从 in-memory state 扩展为 Browser History canonical URL，同时保留内部 `Route { name, params }` / `LooseRoute` 合同。URL 只表达用户所在页面、稳定资源 ID 和轻量 display hint；后端 action、表单正文、AI prompt/response、验证码和 session secret 不进入 URL、history state、pendingAction 或 browser storage。

## 2 当前合同

### 2.1 Route and history

- App 初始 route 优先级：test harness route > canonical path > hash adapter > default `home`。
- `NavigationProvider.navigate(next)` 继续作为屏幕层 API；内部统一 normalize、safe-param filter、serialize、`pushState` / `replaceState` and React state update。
- `popstate` 必须恢复 route params、TopBar active state、chrome hidden state and InterviewContext hydration，并清理 hostile `history.state`。
- `#route=...` adapter 继续服务 static preview、pixel parity and scenario harness；正常浏览器模式下替换为 canonical path。

### 2.2 Canonical route table

| Route | Canonical URL | Safe Params | Chrome |
|-------|---------------|-------------|--------|
| `home` | `/` | `pendingImportId`, `source`, `resumeId` | visible |
| `workspace` | `/workspace` | `targetJobId` | visible |
| `resume_versions` | `/resume-versions` | `resumeId`, `flow`, `createMode`, `targetJobId` | visible |
| `parse` | `/parse` | `targetJobId` | visible |
| `practice` | `/practice` | `sessionId`, `planId`, `targetJobId`, `jobId`, `jdId`, `resumeId`, `sourceReportId`, `roundId`, `roundName`, `mode`, `modality`, `practiceMode`, `practiceGoal`, `language` | hidden |
| `reports` | `/reports` | `targetJobId` | visible |
| `generating` | `/generating` | `reportId` | hidden |
| `report` | `/report` | `reportId` | visible |
| `settings` | `/settings` | `tab` | visible |
| `auth_login` | `/auth/login` | `next`, `email`, encoded pendingAction safe params | visible |
| `auth_verify` | `/auth/verify` | `email`, encoded pendingAction safe params | visible |
| `auth_profile_setup` | `/auth/profile` | `email`, encoded pendingAction safe params | visible |
| `auth_logout` | `/auth/logout` | `next` | visible |

Unsupported paths and malformed query input must normalize to the current route catalog or `home`; canonical output must never emit unsupported paths.

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior` + `frontend` + `routing`。
- **TDD 策略**: 本计划按 `/implement frontend-shell/004-url-addressable-routing frontend` -> `/tdd` 完成。Current regression gate covers route codec, hash adapter, route store, History integration, auth pendingAction serialization, privacy redline and host fallback tests.
- **BDD 策略**: 需要 BDD。本计划维护 [bdd-plan](./bdd-plan.md) / [bdd-checklist](./bdd-checklist.md)，主 checklist 使用 `BDD-Gate:` 引用 `E2E.P0.088`、`E2E.P0.089`、`E2E.P0.090`。
- **替代验证 gate**: 不适用；BDD 是用户行为 gate。Supplemental gates include focused Vitest, host fallback tests, context validator, `make docs-check` and `git diff --check`。

## 4 Operation Matrix / Contract Boundary

| Boundary | Contract | Frontend Consumer | Backend Handler | Persistence | AI dependency | Scenario Coverage |
|----------|----------|-------------------|-----------------|-------------|---------------|-------------------|
| Browser History router | Route codec + safe-param allowlist | route adapter, NavigationProvider, TopBar, auth pendingAction | N/A | browser history only | none | E2E.P0.088 / E2E.P0.089 / E2E.P0.090 |
| Hash adapter | `#route=...` -> `LooseRoute` -> normalize -> canonical replace | `bootstrapRoute.ts`, pixel parity harness, scenario harness | N/A | none | none | E2E.P0.090 + E2E.P0.006 |
| Generated API client | No new OpenAPI operation, fixture or generated client contract | Route params only feed existing screen hooks | owner handlers unchanged | owner stores unchanged | owner-specific only | E2E.P0.088 and owner scenarios |
| Host fallback | Known frontend paths return `index.html`; API/static/script paths stay owned by their handlers | direct open / reload / preview / pixel server | API routes unchanged | N/A | none | E2E.P0.088 / E2E.P0.090 |

## 5 Privacy Redline

The shared safe-param allowlist must drop unknown params and raw payload fields. Tests must prove zero leakage across canonical URL, `window.history.state`, pendingAction, `localStorage`, `sessionStorage`, console capture and mock transport logs.

Blocked payload categories:

- JD / resume original text, pasted content, source URLs and file bodies.
- Practice answers, question text, hints and report evidence text.
- Parsed summaries, structured resume content, rewrite suggestions and generated AI output.
- Prompt body, provider raw response, model raw payload and auth/session secret values.

## 6 Current Implementation Summary

- `routeUrl.ts` owns route-to-URL serialize/parse, route param allowlist and canonical path table.
- `bootstrapRoute.ts` preserves hash adapter input and feeds the same normalization path.
- `routeStore.ts` owns initial route resolution, `pushState`, `replaceState`, `popstate` and URL equality checks.
- `NavigationProvider` keeps screen-level `navigate(next)` stable while routing through Browser History.
- Auth pendingAction serialization shares the safe-param allowlist with URL serialization.
- `spaFallback.mjs`, Vite SPA config and pixel parity server tests separate known frontend paths from API/static/script paths.

### 6.1 Phase 8 route-table evidence reconciliation

- Reconcile the canonical route table with `routeUrl.ts`: workspace accepts no query params, while resume workshop accepts only `resumeId`, `flow`, `createMode` and `targetJobId`.
- Keep old workspace detail/start keys only as hostile P0.088 inputs that must be stripped; practice, generating and report continue to preserve their own current safe params.
- Remove stale discovery keywords and align P0.088 README/data/BDD wording with the executable jsdom assertions.
- Gate with focused routeUrl/P0.088 tests, the P0.088 wrapper, owner/product contexts and docs/diff/pruning checks. No routing runtime behavior changes.

### 6.2 Phase 9 P0.089 workspace-zero-query evidence reconciliation

- Treat the direct-open and popstate workspace payloads in P0.089 as hostile inputs; both must canonicalize to query-free `/workspace` while the positive auth continuation still restores safe practice params.
- Align the executable test title, BDD wording, scenario README and data assets with the current route table; remove claims that workspace retains `planId`, `targetJobId` or other query params.
- Gate with the focused P0.089 test, its four-stage scenario wrapper, owner/product contexts and docs/diff/pruning checks. No routing runtime behavior changes.

### 6.3 Phase 10 unconsumed route helper removal

- Delete `routeUrlsEqual`; repository inventory proves no production or test consumer, while `routeStore.ts` already compares cached `formatRouteUrl` strings directly.
- Remove the false route-store consumer comment and do not add a replacement wrapper.
- BDD is not applicable because the export has no executable caller. Alternative gates are a focused source-surface red/green test, route codec/store regressions, typecheck and owner/global checks.

### 6.4 Phase 11 target-scoped Reports route

- Register protected context route `reports` at `/reports`; its safe-param allowlist contains only `targetJobId`, chrome stays visible, and the route is deliberately absent from `PRIMARY_NAV_ROUTES` / TopBar.
- Missing or invalid Reports target identity automatically uses `replaceRoute(workspace)` so the bad deep link does not remain immediately behind a pushed workspace entry. The route-store bootstrap canonicalizer must not overwrite a newer child mount redirect with its stale initial URL.
- Keep `parse` free of `section=reports` and strip hostile `section`, `reportId`, status and round query inputs. Narrow `report` / `generating` to reportId-only locators; trusted target context for Back comes from API responses, never URL or pendingAction.
- Cover direct open, reload, App navigation, replace/back/forward, hash adapter and SPA host fallback in P0.088/P0.090. Cover unauthenticated direct-open and exact targetJobId-only pendingAction restoration in P0.089.
- Gate with route codec/store/App/auth/privacy/host fallback tests, TopBar negative, owner contexts, docs/diff and pruning checks. Existing route history remains regression evidence; this Phase reopens the completed owner in place.

### 6.5 Phase 12 command/query route split

- Supersede Phase 8/9 的 workspace-zero-query 结论：`/workspace` 无 `targetJobId` 时展示规划列表；`/workspace?targetJobId=<uuid>` 是受保护、可直开/刷新/历史恢复的只读详情 route。它只保留合法 `targetJobId`，并剔除 `planId`、`resumeId`、`autoStartPractice` 与其他业务状态。
- `/parse?targetJobId=<uuid>` 只承载刚导入 TargetJob 的 queued/processing 命令进度；`resumeId` 不再是 safe param。TargetJob 首读已 ready 或轮询转 ready 时，screen 必须 `replaceRoute({ name: "workspace", params: { targetJobId } })`，避免 Back 回到冗余动画。
- ready Home/Workspace card 直接 push 到 workspace detail；不得先进入 Parse。Workspace detail 复用统一只读详情组件，但不播放 Parse loading animation，也不触发 import/poll/start side effects。
- P0.088 覆盖 list/detail direct/reload/back-forward、Parse ready replace 与合法 targetJobId preservation；P0.089 覆盖 workspace/parse targetJobId-only pendingAction restore 和 raw/extra zero-hit；P0.090 覆盖 hash adapter、host fallback 与 planId/resumeId/auto-start stripping。

## 7 验收标准

- Every current route serializes to and parses from its canonical URL with sorted safe query params.
- Direct open, reload, App navigation, replace, back and forward preserve route state without double-push behavior.
- `practice` and `generating` stay chrome-hidden when opened by canonical URL.
- `reports` is protected and chrome-visible, accepts only `targetJobId`, survives direct/reload/history/auth restore, and never appears in TopBar.
- `/workspace` without target is the list and `/workspace?targetJobId` is read-only detail; only `targetJobId` survives normalization. `/parse?targetJobId` is command/progress only and ready state replace-navigates to workspace detail.
- Parse strips legacy `section=reports`; report/generating preserve only reportId and cannot use query state as trusted report context.
- Hash adapter inputs continue to work for static preview and pixel parity harness.
- Auth pendingAction restore returns to the original canonical route using safe params only.
- Hostile URL / hash / history state input is scrubbed into canonical safe state.
- Host fallback returns `index.html` for known frontend paths and does not swallow API/static/script paths.
- `E2E.P0.088`、`E2E.P0.089`、`E2E.P0.090` pass.

## 8 风险与应对

| 风险 | 应对措施 |
|------|----------|
| Frontend URL mirrors backend implementation too closely | Keep the URL route-centric and user-centric; action verbs remain API/client concerns |
| Sensitive payload leaks through query, history or pendingAction | Shared safe-param allowlist + runtime privacy redline tests |
| Hash adapter breaks preview or parity harness | Keep the current adapter covered by E2E.P0.090 |
| Host fallback swallows API paths | Explicit fallback tests distinguish known frontend paths from API/static/script paths |
| Components bypass router | Route adapter remains the only write path; focused tests cover navigation behavior |

## 9 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-14 | 1.12 | Supersede workspace-zero-query with `/workspace?targetJobId` read-only detail and make `/parse?targetJobId` command-only with ready replace. |
| 2026-07-14 | 1.11 | Use replace-only workspace recovery for invalid Reports deep links and prevent stale bootstrap canonicalization from recreating the bad URL. |
| 2026-07-14 | 1.10 | Reopen in place for protected `/reports`, targetJobId-only deep links/auth restore, no TopBar entry, no Parse section compatibility, and reportId-only report/generating routes. |
| 2026-07-10 | 1.9 | Remove the unconsumed routeUrlsEqual wrapper and false consumer comment. |
| 2026-07-10 | 1.8 | Reconcile P0.089 workspace hostile-input evidence with the query-free canonical route contract. |
| 2026-07-10 | 1.7 | Align the route owner and P0.088 with workspace zero-query and current resume-workshop safe params. |
| 2026-07-10 | 1.6 | Isolate the synchronous P0.089 hostile-query test lifecycle with explicit cleanup; keep routing and privacy behavior unchanged. |
| 2026-07-10 | 1.5 | Normalize hash adapter wording across owner, BDD, tests and E2E.P0.090 without changing routing behavior. |
| 2026-07-07 | 1.4 | Compress URL routing owner docs to the current canonical URL, safe-param, hash adapter, privacy and host fallback contract. |
