# Expected Outcome

- `targetJob.realApiMode.test.ts` proves TargetJobs home/import/parse operations use the real backend base URL, cookie credentials, side-effect `Idempotency-Key`, and provenance roundtrip.
- `HomeScreen.test.tsx`, `HomeResumeSelection.test.tsx`, `HomeRecentMocks.test.tsx`, and `MockInterviewCard.test.tsx` run and report passed tests.
- Home hero title, textarea card, upload/URL actions, existing-resume dropdown, create-resume CTA, and TopBar home highlight render in the fixture-backed UI.
- The retired Home hero sub copy and old `解析并确认面试` button copy are absent; the main CTA renders as `立即面试` / `Start interview now`.
- Empty, one-job, and twelve-plus fixture states render deterministically; twelve-plus remains capped at 3 sorted cards and exposes `home-recent-more` / `更多` navigation to `workspace`.
- Legacy entries such as `route-welcome`, `topbar-nav-mistakes`, `topbar-nav-growth`, `topbar-nav-drill`, `topbar-nav-voice`, and prototype JD Match card IDs do not appear in the trigger log.
