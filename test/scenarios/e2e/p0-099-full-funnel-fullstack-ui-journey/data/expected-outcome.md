# Expected Outcome

- Backend test server writes `.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey/state.json`.
- Frontend production build succeeds and preview serves on `127.0.0.1:4174`.
- Playwright executes `frontend/tests/e2e/full-funnel-journey.spec.ts`.
- The spec observes:
  - target import accepted
  - ParseScreen loading steps and ready confirmation
  - Workspace start practice
  - PracticeScreen first session and answer submission
  - Generating screen before ready report
  - ReportDashboard with report id
  - next-round CTA creating a derived plan and fresh session
- Network capture includes baseline and `next_round` `POST /practice/plans`
  bodies.
- Browser privacy checks pass for URL, localStorage, sessionStorage, and console.
- UI source surfaces scanned by the wrapper do not contain `DebriefScreen`,
  `ProfileScreen`, `topbar-nav-debrief`, `topbar-user-profile`,
  `home-aux-debrief`, `debriefId`, or `debriefJobId`.
- Trigger logs contain a Playwright pass marker and no skip/no-test/fail marker.
- Playwright artifacts stay under `.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey/`.
