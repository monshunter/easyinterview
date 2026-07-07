# UI-Design Pixel Parity Gate Checklist

> **版本**: 1.6
> **状态**: completed
> **更新日期**: 2026-07-07

**关联计划**: [plan](./plan.md)

## Phase 1: Playwright infrastructure

- [x] 1.1 `@playwright/test` and npm scripts are available（验证：`frontend/package.json` includes `test:pixel-parity` and `test:pixel-parity:install`）
- [x] 1.2 `frontend/playwright.config.ts` declares desktop and mobile projects, pixel testDir, outputDir and static webServer（验证：scaffold tests and `pnpm exec playwright test --list`）
- [x] 1.3 `frontend/scripts/serve-pixel-parity.mjs` serves `frontend/dist`, `ui-design/` and `/health`, and fails loudly when required paths are missing（验证：server scaffold tests and P0.006 setup）

## Phase 2: Shell parity

- [x] 2.1 TopBar real-browser parity covers current three entries, language menu, display controls and computed style（验证：`frontend/tests/pixel-parity/topbar.spec.ts`）
- [x] 2.2 Auth/login shell and settings shell parity cover source-level DOM anchors and geometry（验证：`frontend/tests/pixel-parity/screens.spec.ts`）
- [x] 2.3 Desktop/mobile layout checks prove TopBar, auth shell and user area stay in viewport without overlap（验证：`frontend/tests/pixel-parity/layout.spec.ts`）

## Phase 3: Screenshot and theme parity

- [x] 3.1 Clean-checkout screenshot smoke uses non-empty screenshot buffers and does not require ignored local snapshot baselines（验证：`frontend/tests/pixel-parity/screenshot.spec.ts`）
- [x] 3.2 Dark mode and customAccent mutate expected root tokens and visible paint（验证：`screenshot.spec.ts` plus per-screen specs）
- [x] 3.3 Explicit `--update-snapshots` workflow is documented for local/CI baseline maintenance（验证：`frontend/README.md` and P0.006 README）

## Phase 4: Current screen parity expansion

- [x] 4.1 Home, Parse and Workspace parity specs cover current DOM anchors, responsive geometry, theme and screenshot smoke（验证：`home.spec.ts`, `parse.spec.ts`, `workspace.spec.ts`）
- [x] 4.2 Resume Workshop parity specs cover flat list, create flow, rewrites/edit/detail and non-current tree/branch negative gates（验证：resume-workshop pixel specs）
- [x] 4.3 Practice, Generating and Report parity specs cover current DOM anchors, layout, theme and screenshot smoke（验证：`practice.spec.ts`, `generating.spec.ts`, `report.spec.ts`）
- [x] 4.4 Workspace full-state uses server-bound initial route bootstrap rather than synthetic route params（验证：`workspace.spec.ts`）
- [x] 4.5 Authenticated user-menu browser parity covers avatar chip, dropdown geometry, mobile viewport containment and logout flow（验证：`topbar.spec.ts`）

## Phase 5: Scenario and docs handoff

- [x] 5.1 `E2E.P0.006` scenario assets exist with setup/trigger/verify/cleanup scripts（验证：`test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/`）
- [x] 5.2 BDD-Gate: `E2E.P0.006` executes `pnpm --filter @easyinterview/frontend test:pixel-parity` and verifies all current spec markers（验证：scenario trigger/verify）
- [x] 5.3 `frontend/README.md` documents Playwright install, frontend build, parity run, screenshot baseline maintenance and offline CDN limits（验证：docs-check）
- [x] 5.4 Non-current route/module entries are negative-only and do not materialize as live parity surfaces（验证：pixel specs and scenario verify）

## Phase 6: closeout

- [x] 6.1 `pnpm --filter @easyinterview/frontend build` passes before pixel parity（验证：owner closeout）
- [x] 6.2 `pnpm --filter @easyinterview/frontend test:pixel-parity` passes for the current parity suite（验证：owner closeout）
- [x] 6.3 Owner context and docs/index are current（验证：`validate_context.py frontend-shell/003 frontend`; `sync-doc-index --check`; `make docs-check`）
