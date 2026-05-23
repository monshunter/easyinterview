# Expected Outcome

- `targetJob.realApiMode.test.ts` proves TargetJobs home/import/parse operations use the real backend base URL, cookie credentials, side-effect `Idempotency-Key`, and provenance roundtrip.
- `HomeScreen.test.tsx`, `HomeRecentMocks.test.tsx`, and `MockInterviewCard.test.tsx` run and report passed tests.
- Home hero, textarea card, upload/URL actions, JOB PICKS, POST-INTERVIEW, and TopBar home highlight render in the fixture-backed UI.
- Empty, one-job, and twelve-plus fixture states render deterministically; twelve-plus remains capped at 12 sorted cards.
- Legacy entries such as `route-welcome`, `topbar-nav-mistakes`, `topbar-nav-growth`, `topbar-nav-drill`, `topbar-nav-voice`, and prototype JD Match card IDs do not appear in the trigger log.
