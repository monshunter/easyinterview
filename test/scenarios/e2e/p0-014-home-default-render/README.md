# E2E.P0.014 — Home Default Render

> **Scenario ID**: E2E.P0.014
> **Owner**: frontend-home-job-picks-and-parse/001-home-jd-import-and-parse
> **Status**: Ready
> **Execution**: automated

## Scope

Verifies the home screen renders correctly in three states:
- Empty state (no TargetJobs) → textarea focused, empty state CTA
- Non-empty state (1-3 TargetJobs) → MockInterviewCards rendered
- 12+ items capped at 12, sorted by `updatedAt desc`

## Fixture Variants

`openapi/fixtures/TargetJobs/listTargetJobs.json`:
- `empty` — zero items
- `one-job` — single TargetJob
- `twelve-plus` — 15 items, only first 12 rendered

## Verification Points

- Hero (label/title/sub) DOM anchors
- Textarea card + upload/URL buttons
- Aux cards (JOB PICKS, POST-INTERVIEW)
- TopBar highlights home
- i18n zh/en switching
- Theme switching (warm/dark/customAccent)
- Mobile responsive layout

## Scripts

- `scripts/setup.sh` — ensure frontend dist exists, select fixture variant
- `scripts/trigger.sh` — run home screen verification via Vitest/Playwright
- `scripts/verify.sh` — assert testid anchors, sorting, card cap, empty state
- `scripts/cleanup.sh` — reset test state

## Offline Limitations

- Requires `pnpm build` output at `frontend/dist/`
- Playwright chromium must be installed
- UI-design golden preview may fail offline (CDN fonts)
