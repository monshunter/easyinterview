# Expected Outcome

- `targetJob.realApiMode.test.ts` proves TargetJobs home/import/parse operations use the configured base URL, cookie credentials, side-effect `Idempotency-Key`, and provenance roundtrip through stub fetch.
- `HomeScreen.test.tsx`, `HomeLayout.test.tsx`, `HomeResumeSelection.test.tsx`, `HomeRecentMocks.test.tsx`, and `MockInterviewCard.test.tsx` run and report passed tests.
- Home hero title, integrated JD input card with upload/URL source actions, compact existing-resume dropdown row, create-resume CTA, submit row and English copy render in the fixture-backed UI.
- The out-of-scope Home hero sub copy and `解析并确认面试` button copy are absent; the main CTA renders as `立即面试` / `Start interview now`.
- Empty, one-job, and twelve-plus fixture states render deterministically; twelve-plus remains capped at 3 sorted cards and exposes `home-recent-more` / `更多` navigation to `workspace`.
- Out-of-scope entries such as `route-welcome`, `topbar-nav-mistakes`, `topbar-nav-growth`, `topbar-nav-drill`, `topbar-nav-voice`, and prototype JD Match card IDs do not appear in the trigger log.
