# Expected Outcome

- `targetJob.realApiMode.test.ts` proves TargetJobs home/import/parse operations use the configured base URL, cookie credentials, side-effect `Idempotency-Key`, exact `{rawText,targetLanguage,resumeId}` import body, and provenance roundtrip through stub fetch.
- `HomeScreen.test.tsx`, `HomeLayout.test.tsx`, `HomeResumeSelection.test.tsx`, `HomeRecentMocks.test.tsx`, `useRecentTargetJobs.test.tsx`, and `MockInterviewCard.test.tsx` run and report passed tests.
- Authenticated StrictMode Home emits exactly one underlying `listTargetJobs` GET and one `listResumes` GET; the ready card body enters targetJobId-only Workspace detail and never Parse.
- Home hero title, paste-only JD textarea, compact existing-resume dropdown row, create-resume CTA, submit row and English copy render in the fixture-backed UI; old source controls, upload / URL triggers, and assist modal are absent.
- The out-of-scope Home hero sub copy and `解析并确认面试` button copy are absent; the main CTA renders as `立即面试` / `Start interview now`.
- Empty, one-job, and twelve-plus fixture states render deterministically; twelve-plus remains capped at 3 sorted cards and exposes `home-recent-more` / `更多` navigation to `workspace`.
- Out-of-scope entries such as `route-welcome`, `topbar-nav-mistakes`, `topbar-nav-growth`, `topbar-nav-drill`, `topbar-nav-voice`, and prototype JD Match card IDs do not appear in the trigger log.
- `tests/pixel-parity/home.spec.ts` runs `paste-only Home matches the UI truth and captures desktop/mobile evidence`, proving formal/prototype DOM, computed-style, bounding-box, viewport and screenshot parity at 1440×900 and 390×844 and emitting one `E2E.P0.014 home paste-only browser gate` marker per viewport.
