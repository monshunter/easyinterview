# UI-Design Pixel Parity Gate

> **版本**: 1.5
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

本 plan 已完成真实浏览器 UI parity gate：使用 Playwright + Chromium 在 desktop (1440x900) 与 mobile (390x844) 两个 viewport 下加载正式 `frontend/dist` 和 `ui-design/` golden preview，验证当前 App Shell 与已迁移业务屏的 DOM anchors、computed style、bounding box、responsive geometry、theme/dark/customAccent 和 screenshot smoke。

当前 gate 只覆盖 current route catalog 和 current UI truth source：

- TopBar 三入口：`home / workspace / resume_versions`。
- Auth / settings / user menu / logout browser geometry。
- Home / parse / workspace / resume workshop / practice / generating / report parity specs。
- Workspace full-state 通过 server-bound route bootstrap 进入，不依赖 synthetic route params。
- Screenshot 常规 gate 使用非空 screenshot smoke；baseline diff 只作为显式 `--update-snapshots` 维护流程。
- Non-current route/module entries 只作为负向断言，不作为正向屏幕或可见入口。

## 2 当前合同

### 2.1 Executable Surface

| surface | executable | contract |
|---------|------------|----------|
| Playwright config | `frontend/playwright.config.ts` | desktop + mobile projects, `tests/pixel-parity`, static web server, `.playwright-output` |
| Static server | `frontend/scripts/serve-pixel-parity.mjs` | serves `frontend/dist`, `ui-design/`, and `/health`; fails loudly when required dirs are missing |
| Pixel specs | `frontend/tests/pixel-parity/*.spec.ts` | current 13-spec parity suite for shell and migrated screens |
| Scenario | `test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/` | setup/trigger/verify/cleanup wrapper for `test:pixel-parity` |
| Handoff docs | `frontend/README.md` §2.7 | install, build, run, screenshot baseline maintenance and offline limits |

### 2.2 Current Gate Expectations

- `pnpm --filter @easyinterview/frontend build` produces `frontend/dist/index.html`.
- `pnpm --filter @easyinterview/frontend test:pixel-parity` runs current parity specs in both viewport projects.
- `E2E.P0.006` verify requires passing summary, desktop/mobile markers, and all current spec markers.
- Non-current entries such as standalone `jd_match`, `debrief`, `profile`, `mistakes`, `growth`, `drill`, and standalone `voice` must not appear as live route/testid parity failures.
- `ui-design/index.html` may need CDN access unless assets are vendored; this is documented in the scenario README.

## 3 质量门禁

- **Plan 类型**: `frontend + tooling + visual parity + BDD`。
- **TDD 策略**: 适用。Playwright config, server fixture and specs are executable tests. The owner gate fails when selectors, geometry, build output, browser install or server paths are missing.
- **BDD 策略**: 适用。`E2E.P0.006` is the real-browser parity scenario and is tracked by [bdd-plan](./bdd-plan.md) / [bdd-checklist](./bdd-checklist.md).
- **替代验证 gate**:
  - `pnpm --filter @easyinterview/frontend build`
  - `pnpm --filter @easyinterview/frontend test:pixel-parity`
  - `test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/scripts/setup.sh && test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/scripts/trigger.sh && test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/scripts/verify.sh && test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/scripts/cleanup.sh`
  - `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/frontend-shell/plans/003-ui-design-pixel-parity-gate/context.yaml --target frontend`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `make docs-check`

## 4 实施结果

### Phase 1: Playwright infrastructure

- Added `@playwright/test`, `test:pixel-parity`, and `test:pixel-parity:install`.
- Added `frontend/playwright.config.ts` with desktop/mobile projects and static web server integration.
- Added `frontend/scripts/serve-pixel-parity.mjs` with health check and explicit missing-dir failures.

### Phase 2: Shell parity

- Added TopBar DOM, computed style and current-entry assertions.
- Added auth/settings shell DOM and geometry assertions.
- Added layout checks for desktop and mobile viewport containment.

### Phase 3: Screenshot and theme parity

- Added screenshot smoke for clean checkout behavior.
- Added dark and customAccent token/paint assertions.
- Kept snapshot baseline regeneration as explicit maintenance, not normal PASS criteria.

### Phase 4: Business-screen parity expansion

- Added current Home, Parse, Workspace, Resume Workshop, Practice, Generating and Report parity specs.
- Added workspace server-bound initial route bootstrap for full-state checks.
- Added authenticated user menu browser parity and logout flow.

### Phase 5: Scenario and docs handoff

- Added `E2E.P0.006` scenario assets and verification wrapper.
- Updated frontend README with install/build/run/baseline maintenance instructions.
- Preserved jsdom smoke (`E2E.P0.005`) as fast feedback while Playwright owns real browser parity.

## 5 验收标准

| ID | 验收点 | 验证 |
|----|--------|------|
| A-1 | Playwright config and static server are executable | scaffold/server tests, P0.006 setup |
| A-2 | TopBar, auth/settings shell and user menu match current UI source in real browsers | `topbar.spec.ts`, `screens.spec.ts`, `layout.spec.ts` |
| A-3 | Current business screens have viewport-safe parity coverage | home/parse/workspace/resume/practice/generating/report specs |
| A-4 | Screenshot smoke, dark mode and customAccent work without ignored baseline dependency | `screenshot.spec.ts` and per-screen smoke |
| A-5 | Non-current routes/modules do not become live parity surfaces | pixel specs and P0.006 verify negative checks |
| A-6 | Scenario/documentation handoff is complete | `E2E.P0.006`, `frontend/README.md`, docs-check |

## 6 变更记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-07 | 1.5 | Compress owner docs to the current 13-spec Playwright pixel parity contract and remove staged implementation narrative. |
| 2026-07-06 | 1.4 | Reconcile current positive screen markers and browser-menu parity scope. |
