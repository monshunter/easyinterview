# E2E.P0.014 — Home Default Render

> **Scenario ID**: E2E.P0.014
> **Owner**: frontend-home-job-picks-and-parse/001-home-jd-import-and-parse
> **Status**: Ready
> **Execution**: automated
> **Isolation**: repo-tracked Vitest / Playwright with stub fetch
> **parallel-safe**: No

## Scope

Verifies the current Home Vitest contract:
- Paste-only JD quick start with an existing-resume selector and `立即面试` CTA
- The intake surface contains one textarea and no JD URL, file-upload, manual-form, source-control, or assist-modal branch
- Empty TargetJob variant renders no recent cards
- One/default variants render shared `MockInterviewCard` content
- Twelve-plus variant is sorted by `updatedAt desc`, capped at 3, and exposes More navigation
- Recent cards open plan detail or use the structured-round quick-start action

## Fixture Variants

`openapi/fixtures/TargetJobs/listTargetJobs.json`:
- `empty` — zero items
- `one-job` — single TargetJob
- `twelve-plus` — 15 items, only first 3 rendered and `更多` shown

## Verification Points

- Hero label/title DOM anchors, with out-of-scope hero sub copy absent
- `home-jd-input-card` contains only the paste textarea; old source controls, upload / URL triggers, and assist modal are absent
- Existing ready resume dropdown is compact, with create-resume CTA on the same row
- Main CTA sits below resume selection rather than inside the JD input card
- Main CTA copy is `立即面试` / `Start interview now`
- Out-of-scope aux cards (JOB PICKS, POST-INTERVIEW) remain absent
- Real-mode generated-client gate for TargetJobs home/import/parse operations using stub fetch
- English i18n copy
- Ready TargetJob filtering, card-detail navigation, quick-start and out-of-scope negatives
- 1440×900 and 390×844 formal/prototype DOM, computed-style, bounding-box, viewport, and screenshot parity

## Scripts

- `scripts/setup.sh` — initialize the scenario output marker
- `scripts/trigger.sh` — run the generated-client routing test, five Home Vitest files, build, and Home Playwright parity gate
- `scripts/verify.sh` — verify runner markers, desktop/mobile screenshot markers, and old-intake/out-of-scope negatives
- `scripts/cleanup.sh` — remove the setup marker while retaining the trigger log

## Real-Mode Generated-Client Gate

- The trigger runs `src/api/targetJob.realApiMode.test.ts` with
  `VITE_EI_API_MODE=real` and
  `VITE_EI_API_BASE_URL=http://localhost:8080/api/v1`. Its stub fetch proves the
  generated client routes `listTargetJobs`, paste-only `importTargetJob`,
  `getTargetJob`, and `updateTargetJob` to the configured base URL with cookie
  credentials, Idempotency-Key side effects, and provenance roundtrip without
  making a network request. Its `createUploadPresign(purpose=resume)` assertion
  is an explicit positive guard for the independent Resume upload capability;
  it is not a JD intake path.
- UI variants remain fixture-backed so this scenario can keep deterministic
  DOM, filtering, sorting, layout, i18n and navigation assertions.
