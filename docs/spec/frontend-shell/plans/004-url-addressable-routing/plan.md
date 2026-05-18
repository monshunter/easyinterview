# URL-Addressable Routing

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-05-18

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

把 `frontend-shell` 从仅内存 route + `#route=` bootstrap 升级为可复制、可刷新、可直开的 URL-addressable SPA routing。正式前端继续保留单页应用和现有 `Route { name, params }` / `LooseRoute` 内部合约，但 canonical 用户地址改为 Browser History path + query。

本计划不把前端 URL 做成 REST API 的 1:1 镜像。URL 表达的是用户所在页面、稳定服务端资源 ID 和轻量 display hint；后端 action、raw payload、AI prompt / response、表单草稿和 auth secret 都不进入 URL。

## 2 背景

当前 `App` 以 React state 保存 route，`navigate()` 只调用 `setRoute(normalizeRoute(next))`；启动入口只解析 `window.__EASYINTERVIEW_INITIAL_ROUTE__` 或 `#route=...`。这足以支撑 static preview 和 pixel parity，但无法支撑用户复制业务页链接、刷新当前任务、直接打开 report / workspace / resume detail、或用浏览器 back / forward 还原路线。

EasyInterview 的核心工作流已经围绕真实 JD、target job、resume version、practice session、report 和 debrief 展开。随着 `frontend-resume-workshop`、`frontend-workspace-and-practice`、`frontend-report-dashboard` 等页面接入，路由需要表达稳定资源上下文，否则用户在刷新或分享链接后会丢失当前工作台状态。

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior` + `frontend` + `routing`。本计划改变用户可见导航行为、deep-link 能力、auth 恢复路径和部署 fallback。
- **TDD 策略**: 通过 `/implement frontend-shell/004-url-addressable-routing frontend` -> `/tdd` 执行。每个 checklist item 先写失败断言，再实现最小改动：route codec / path table / hash adapter 用 Vitest 单测；History push/replace/popstate 用 jsdom 集成测试；auth pendingAction + privacy 用 focused component / hook test；server fallback 与 pixel harness 用现有 Playwright / scenario wrapper 验证。
- **BDD 策略**: 适用。本计划引入用户可感知 URL 行为，维护 [bdd-plan](./bdd-plan.md) / [bdd-checklist](./bdd-checklist.md)，主 checklist 使用 `BDD-Gate:` 引用 `E2E.P0.088`、`E2E.P0.089`、`E2E.P0.090`。
- **替代验证 gate**: 不适用；BDD 是本计划的用户流 gate。补充 gate 包括 route codec 单测、privacy negative grep、`test:pixel-parity` hash adapter regression、`make docs-check` 和 context validator。

## 4 Operation Matrix / Contract Boundary

| operationId / Boundary | Fixture / Contract | Frontend Consumer | Backend Handler | Persistence | AI dependency | Scenario Coverage |
|------------------------|--------------------|-------------------|-----------------|-------------|---------------|-------------------|
| N/A - Browser History router | 前端 route codec，不新增 OpenAPI operation | `frontend/src/app/*` route adapter、NavigationProvider、TopBar、auth pendingAction | n/a | browser history only | none | E2E.P0.088 / E2E.P0.089 / E2E.P0.090 |
| N/A - Static preview hash adapter | `#route=...` adapter，继续走 `normalizeRoute` | `bootstrapRoute.ts` / pixel parity harness | n/a | none | none | E2E.P0.090 + existing E2E.P0.006 |
| Existing generated API client | 不变；URL route 不新增 API、fixture 或 generated client contract | route params 只作为已有 screen/hook 的输入 | 各 owner handler 不变 | 各 owner store 不变 | none in frontend；owner 后端 AI 依赖不由 URL router 新增 | E2E.P0.088 / owner scenarios |
| N/A - Server fallback rewrite | 部署 / dev server 配置：known frontend path -> `index.html`，`/api/*` 不被吞掉 | direct open / reload | backend API routes remain API-owned | n/a | none | E2E.P0.088 / E2E.P0.090 |

## 5 Canonical URL Contract

| Route | Canonical URL | Safe Params | Chrome |
|-------|---------------|-------------|--------|
| `home` | `/` | `pendingImportId`, `source` | visible |
| `jd_match` | `/jd-match` | `tab`, `selectedJobMatchId`, `action`, `pendingJdMatchActionId` | visible |
| `workspace` | `/workspace` | `targetJobId`, `jobId`, `resumeVersionId`, `planId`, `roundId`, `roundName`, `jdId`, `sessionId`, `sourceSessionId`, `replayItems`, `evidenceGaps`, `nextRoundId`, `mode`, `modality`, `practiceMode`, `practiceGoal`, `hintUsed`, `hintCount`, `autoStartPractice`, `language`, `debriefId` | visible |
| `resume_versions` | `/resume-versions` | `resumeAssetId`, `versionId`, `flow`, `tab`, `createMode`, `targetJobId`, `branchOriginalId`, `tailorRunId` | visible |
| `debrief` | `/debrief` | `targetJobId`, `jobId`, `jdId`, `sessionId`, `resumeVersionId`, `roundId`, `roundName`, `mode`, `modality`, `practiceMode`, `practiceGoal`, `language`, `debriefId`, `debriefJobId` | visible |
| `parse` | `/parse` | `jdId`, `targetJobId`, `importId`, `source`, `sourceJobMatchId` | visible |
| `practice` | `/practice` | `sessionId`, `planId`, `targetJobId`, `jobId`, `jdId`, `resumeVersionId`, `roundId`, `roundName`, `mode`, `modality`, `practiceMode`, `practiceGoal`, `language`, `debriefId` | hidden |
| `generating` | `/generating` | `sessionId`, `reportId`, `planId`, `targetJobId`, `jobId`, `jdId`, `resumeVersionId`, `roundId`, `roundName`, `mode`, `modality`, `practiceMode`, `practiceGoal`, `hintUsed`, `hintCount` | hidden |
| `report` | `/report` | `sessionId`, `reportId`, `targetJobId`, `jobId`, `jdId`, `resumeVersionId`, `roundId`, `roundName`, `mode`, `modality`, `practiceMode`, `practiceGoal`, `hintUsed`, `hintCount`, `reportStatus`, `errorCode` | visible |
| `company_intel` | `/company-intel` | `targetJobId`, `jobId`, `companyId`, `jdId` | visible |
| `profile` | `/profile` | none | visible |
| `settings` | `/settings` | `tab` | visible |
| `auth_login` | `/auth/login` | `next`, `email` display hint only, plus encoded pendingAction safe params (`pendingRoute`, `pendingType`, `pendingLabel` and target-route safe params) | visible |
| `auth_register` | `/auth/register` | `next`, `email` display hint only, plus encoded pendingAction safe params | visible |
| `auth_verify` | `/auth/verify` | `email` display hint only, short-lived verification code handled by the auth form/client boundary, plus encoded pendingAction safe params | visible |
| `auth_reset` | `/auth/reset` | `next`, `email` display hint only | visible |
| `auth_logout` | `/auth/logout` | `next` | visible |

`voice` is intentionally absent. Voice interview remains `practice?mode=voice&modality=voice`.

The allowlist above is a current cross-owner contract, not an aspirational sample. It is derived from `frontend/src/app/interview-context/InterviewContext.tsx`, `frontend/src/app/auth/pendingAction.ts`, and active owner specs for home / parse / jd_match, workspace / practice / generating, report, resume workshop and debrief. `jd_match` search query and saved-search label remain SPA session memory only; Search auth restore uses opaque `pendingJdMatchActionId`, and Confirm interview hands off to `parse` through `source=jd_match&sourceJobMatchId=...`. Implementation may only remove a key after the owning spec removes the workflow that produces it.

## 6 实施步骤

### Phase 1: Route codec and compatibility adapter

#### 1.1 Route-to-URL table

Create a typed route codec next to the existing route catalog. It must convert every retained `Route.name` to canonical path + sorted query params, and parse canonical URL back into `LooseRoute`. Unknown routes fallback to `home`; old aliases still pass through `normalizeRouteName` and never become canonical route names.

#### 1.2 Safe param allowlist

Introduce a route-param allowlist shared by URL serialization, auth pendingAction serialization, and tests. Unknown params are dropped from canonical URLs unless explicitly allowed by the route owner. The allowlist must preserve current owner handoff keys such as `autoStartPractice`, `practiceMode`, `practiceGoal`, `reportId`, `reportStatus`, `errorCode`, `tailorRunId`, `pendingImportId`, `selectedJobMatchId`, `pendingJdMatchActionId`, `sourceJobMatchId`, `debriefId` and `debriefJobId`. Raw payload fields such as `rawText`, `rawDescription`, `sourceUrl`, `query`, `label`, `guidedAnswers`, `parsedSummary`, `structuredProfile`, `suggestion`, `originalBullet`, `suggestedBullet`, `questionText`, `answerText`, `notes`, `prompt`, `response`, `file`, `token`, `password` and auth/session secrets must fail privacy tests if they appear in URL / storage / history.

#### 1.3 Hash adapter preservation

Keep `#route=...` parsing for static preview, pixel parity and old harness entrypoints. The adapter must produce the same `LooseRoute` input as canonical URL parsing, run through `normalizeRoute`, and then allow the router to replace browser state with the canonical path when mounted in normal browser mode.

### Phase 2: Browser History integration

#### 2.1 Router state source

Replace App-only in-memory navigation with a browser-aware route store. Initial route priority must be: explicit test harness `window.__EASYINTERVIEW_INITIAL_ROUTE__` > canonical path > `#route=` adapter > `DEFAULT_ROUTE`. The route store owns `pushState`, `replaceState`, `popstate`, and URL equality checks.

#### 2.2 NavigationProvider handoff

Keep the public `navigate(next: LooseRoute)` contract so existing screens do not need cross-owner rewrites. Internally, navigation normalizes route, serializes safe params to canonical URL, pushes/replaces history, and updates React state exactly once.

#### 2.3 Back/forward and chrome parity

Browser back / forward must update route state, TopBar active route, chrome hidden behavior and InterviewContext hydration without double renders or lost params. `practice` and `generating` remain chrome-hidden even when opened directly by canonical URL.

### Phase 3: Auth and privacy

#### 3.1 pendingAction canonical restore

`requestAuth(pendingAction)` stores canonical route identity and safe params only. Login success restores the same canonical URL and route params; logout / auth pages must not preserve raw form body or auth secrets beyond the auth contract.

#### 3.2 URL privacy negative tests

Add automated negative checks for URL, `window.history.state`, `pendingAction`, `localStorage`, `sessionStorage`, console/log capture and mock transport logs. The checks must use representative raw JD, resume text, guided answers, parsed summary, suggestion text and AI prompt tokens, and assert zero leakage during navigation, auth restore, direct-open and browser `popstate` recovery from hostile history entries.

### Phase 4: Host fallback and regression

#### 4.1 Dev / static host fallback

Ensure dev server, preview server and pixel parity server can serve known frontend paths by returning `index.html`; `/api/*`, generated client fixture endpoints and scenario script paths must not be swallowed by frontend fallback.

#### 4.2 Legacy route negative regression

Retired routes (`welcome`, `growth`, `plan`, `mistakes`, `drill`, `followup`, `experiences`, `star`, `onboarding`, standalone `voice`) must still normalize to retained routes or `home`, and must not appear as screen files, TopBar entries, route path constants, scenario names or canonical path outputs.

#### 4.3 Documentation and handoff

Update frontend README / route comments only where implementation needs operator guidance. The handoff must tell implementers how to run route codec tests, History integration tests, pixel parity hash regression and E2E.P0.088-090 scenario wrappers.

## 7 验收标准

- `frontend-shell` spec C-11 / C-12 / C-13 have executable coverage.
- Every retained route has canonical path parsing and serialization tests, including unknown / malformed / old alias fallback.
- Direct open, reload, App navigation, back, forward and `replaceState` behave consistently across primary nav, context routes, auth pages and chrome-hidden routes.
- `#route=...` static preview and pixel parity entrypoints remain supported until those harnesses are explicitly migrated.
- URL / history / pendingAction / storage privacy negative tests block raw JD, resume, guided answers, parsed result, suggestion, AI prompt / response and auth secret leakage.
- No OpenAPI / generated client / fixture contract changes are introduced by this plan.
- `E2E.P0.088`、`E2E.P0.089`、`E2E.P0.090` are created and pass before this plan is marked completed.

## 8 风险与应对

| 风险 | 应对措施 |
|------|----------|
| Treating frontend URL as REST API mirror creates brittle paths tied to backend implementation | Keep URL contract route-centric and user-centric; only stable resource IDs enter URL; action verbs remain API/client concerns |
| URL leaks sensitive payloads through query, history, storage or pendingAction | Shared safe-param allowlist + negative grep / runtime capture; auth restore stores route identity, not raw draft |
| Hash migration breaks ui-design / pixel parity harness | Preserve `#route=` adapter and add E2E.P0.090 regression; migrate harness only after owner plans update |
| Server fallback swallows API paths | Explicit fallback tests distinguish known frontend paths from `/api/*` and fixture/script paths |
| Components bypass router and mutate `window.location` directly | Route adapter becomes the only write path; tests grep for direct writes outside adapter/tests |
| Legacy route compatibility resurrects removed modules | Old aliases normalize only to retained route names or `home`; canonical output never emits retired route path |
