# E2E.P0.086 Rewrites Accept-Only + Edit Tab updateResume

> **场景 ID**: E2E.P0.086
> **执行方式**: automated (vitest jsdom)
> **隔离级别**: in-process (vitest worker)
> **状态**: Ready

## 1 Given

- Fixture-backed mock-first client: flat `getResume`, `updateResume`,
  `duplicateResume`, and current `requestResumeTailor/getResumeTailorRun`
  fixtures.
- User A authenticated; lang default.
- D-20: suggestion accept/reject/updateVersion routes are retired; Rewrites
  suggestions are ephemeral and accept-only until saved.

## 2 When

- Accept a rewrite locally → Preview & save opens modal → overwrite or save
  as new through flat handlers.
- Edit Tab: change headline + summary → save → 422 inline.
- Disallowed legacy operations are searched out of runtime source.

## 3 Then

- No accept/reject request is sent; accept only updates local Rewrites state.
- Save paths call current flat handlers (`updateResume` overwrite,
  `duplicateResume` save-as-new).
- Edit Tab `updateResume` payload only carries allowed fields; 422 surfaces
  in-form error; success triggers toast + detail refresh.
- Privacy: originalBullet / suggestedBullet / matchSummary / structuredProfile
  / manual edit text never leak to URL / pendingAction / localStorage / fetch
  transport log / toast content.

## 4 Verification Entry

`scripts/trigger.sh` runs:

- `src/app/screens/resume-workshop/create/PreviewStage.test.tsx`
- `src/app/screens/resume-workshop/tabs/ResumeRewritesTab.test.tsx`
- `src/app/screens/resume-workshop/components/ResumeDetailView.test.tsx`
- `src/app/screens/resume-workshop/tabs/ResumeEditTab.test.tsx`

## 5 Output

- `.test-output/e2e/p0-086-resume-suggestion-accept-reject-edit-and-update-version/trigger.log` Vitest pass.
- verify.sh asserts vitest summary + spec presence + retired-grep gates.

## 6 Baseline

`make codegen-check` clean: retired accept/reject/updateVersion operations stay absent; flat `updateResume` / `duplicateResume` / tailor operations remain valid.

## 7 离线限制

Pure fixture-backed Vitest path; offline-friendly.

## 8 方法标注

`method=fixture-backed-frontend`. Backend real route preflight evidence: plan 003 checklist 0.1 + 0.2.
