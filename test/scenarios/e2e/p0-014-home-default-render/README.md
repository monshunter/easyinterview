# E2E.P0.014 — Home Default Render

> **Scenario ID**: E2E.P0.014
> **Owner**: frontend-home-job-picks-and-parse/001-home-jd-import-and-parse
> **Status**: Ready
> **Execution**: automated

## Scope

Verifies the home screen renders correctly in three states:
- JD input quick start with existing-resume selector and `立即面试` CTA
- Empty state (no TargetJobs) → textarea focused, empty state CTA
- Non-empty state (1-3 TargetJobs) → MockInterviewCards rendered
- 12+ items capped at 12, sorted by `updatedAt desc`

## Fixture Variants

`openapi/fixtures/TargetJobs/listTargetJobs.json`:
- `empty` — zero items
- `one-job` — single TargetJob
- `twelve-plus` — 15 items, only first 12 rendered

## Verification Points

- Hero label/title DOM anchors, with retired hero sub copy absent
- Textarea card + upload/URL buttons
- Existing ready resume selector + create-resume CTA
- Main CTA copy is `立即面试` / `Start interview now`
- Retired aux cards (JOB PICKS, POST-INTERVIEW) remain absent
- Real backend mode generated-client gate for TargetJobs home/import/parse operations
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

## Real Backend Overlay

- The trigger first runs `src/api/targetJob.realApiMode.test.ts` with
  `VITE_EI_API_MODE=real` and
  `VITE_EI_API_BASE_URL=http://localhost:8080/api/v1`, proving the production
  generated client routes `listTargetJobs`, `createUploadPresign`,
  `importTargetJob`, `getTargetJob`, and `updateTargetJob` to the real backend
  base URL with cookie credentials, Idempotency-Key side effects, and
  provenance roundtrip.
- UI variants remain fixture-backed so this scenario can keep deterministic
  DOM, sorting, theme, i18n, and responsive assertions.
