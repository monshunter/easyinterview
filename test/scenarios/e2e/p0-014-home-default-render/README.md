# E2E.P0.014 — Home Default Render

> **Scenario ID**: E2E.P0.014
> **Owner**: frontend-home-job-picks-and-parse/001-home-jd-import-and-parse
> **Status**: Ready
> **Execution**: automated

## Scope

Verifies the current Home Vitest contract:
- JD input quick start with existing-resume selector and `立即面试` CTA
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
- JD input card integrates paste textarea with upload/URL source actions
- Upload/URL source actions live inside `home-jd-input-card`; independent upload source panel is absent
- Existing ready resume dropdown is compact, with create-resume CTA on the same row
- Main CTA sits below resume selection rather than inside the JD input card
- Main CTA copy is `立即面试` / `Start interview now`
- Out-of-scope aux cards (JOB PICKS, POST-INTERVIEW) remain absent
- Real-mode generated-client gate for TargetJobs home/import/parse operations using stub fetch
- English i18n copy
- Ready TargetJob filtering, card-detail navigation, quick-start and out-of-scope negatives

## Scripts

- `scripts/setup.sh` — initialize the scenario output marker
- `scripts/trigger.sh` — run the generated-client routing test and five Home Vitest files
- `scripts/verify.sh` — verify runner markers and out-of-scope log/source negatives
- `scripts/cleanup.sh` — remove the setup marker while retaining the trigger log

## Real-Mode Generated-Client Gate

- The trigger runs `src/api/targetJob.realApiMode.test.ts` with
  `VITE_EI_API_MODE=real` and
  `VITE_EI_API_BASE_URL=http://localhost:8080/api/v1`. Its stub fetch proves the
  generated client routes `listTargetJobs`, `createUploadPresign`,
  `importTargetJob`, `getTargetJob`, and `updateTargetJob` to the configured
  base URL with cookie credentials, Idempotency-Key side effects, and provenance
  roundtrip without making a network request.
- UI variants remain fixture-backed so this scenario can keep deterministic
  DOM, filtering, sorting, layout, i18n and navigation assertions.
