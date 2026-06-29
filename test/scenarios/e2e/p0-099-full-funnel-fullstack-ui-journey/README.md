# E2E.P0.099 full-funnel-fullstack-ui-journey

> **Status**: Ready
> **Owner plan**: [e2e-scenarios-p0/001](../../../docs/spec/e2e-scenarios-p0/plans/001-full-funnel-happy-journey/plan.md)
> **Spec acceptance**: C-1, C-2, C-4, C-5, C-6
> **Isolation**: shared-postgres, scenario-owned user email, dedicated localhost ports
> **Parallel-safe**: No

## Scope

Full-stack UI journey that drives the production-built frontend against a real
backend test server and Postgres. Playwright owns the transient backend and
frontend preview processes through `frontend/playwright.e2e.config.ts`.

## Given

- Dev-stack Postgres is reachable through `DATABASE_URL`.
- `frontend/playwright.e2e.config.ts` starts:
  - backend `TestE2EP0099ScenarioBackendServer` on `127.0.0.1:18099`
  - frontend production build/preview on `127.0.0.1:4174`
- The backend server seeds an authenticated user and ready resume through
  `registerResume` + `resume_parse`.

## When

Run:

```bash
bash scripts/setup.sh
bash scripts/trigger.sh
bash scripts/verify.sh
bash scripts/cleanup.sh
```

`trigger.sh` executes the Playwright spec with:

```bash
EI_PLAYWRIGHT_OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey/playwright" \
pnpm --filter @easyinterview/frontend exec playwright test --config=playwright.e2e.config.ts tests/e2e/full-funnel-journey.spec.ts
```

## Then

- UI navigation covers home -> parse -> workspace -> practice -> generating ->
  report -> next-round practice.
- Parse loading and report generating screens transition through real async job
  progress.
- The next-round CTA sends `createPracticePlan(next_round, sourceReportId)`,
  starts a fresh session, and lands on `/practice` with derived plan/session
  query params.
- Browser URL/storage/console surfaces do not expose private JD text, answer
  text, or report prose.
- TopBar, Home, routing, and report/practice UI surfaces do not expose debrief
  or user-profile navigation, screens, pending actions, or first-party mock
  transport branches.
- Playwright output is confined to `.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey/`.

## Cleanup

`cleanup.sh` deletes scenario-owned DB rows and stops any leftover scenario
backend/frontend processes on the dedicated commands. Output evidence under
`.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey/` is retained.
