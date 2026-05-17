# E2E.P0.086 Resume Suggestion Accept/Reject/Manual-Edit + Edit Tab updateResumeVersion + Terminal State Machine

> **场景 ID**: E2E.P0.086
> **执行方式**: automated (vitest jsdom)
> **隔离级别**: in-process (vitest worker)
> **状态**: Ready

## 1 Given

- Fixture-backed mock-first client: `Resumes/getResumeVersion.json targeted-with-suggestions` + `Resumes/acceptResumeTailorSuggestion.json default / idempotency-replay / already-decided-409` + `Resumes/rejectResumeTailorSuggestion.json default / idempotency-replay / already-decided-409` + `Resumes/updateResumeVersion.json default / idempotency-replay / validation-error-422`.
- 409 envelope: `error.code='VALIDATION_FAILED'` + `error.details.reason='SUGGESTION_ALREADY_DECIDED'`.
- User A authenticated; lang default. User B authenticated (cross-user).
- Phase 0 real-backend preflight: 6 ops (accept/reject/update + suggestion deps) generated client/server/handler/route real.

## 2 When

- Accept b1 → replay same IK → fresh IK accept again (409 already-decided).
- Reject b3 → replay same IK → fresh IK reject again (409).
- Manual edit b2: update structuredProfile.manualEdits[] → bodyless accept; replay; update success + accept fail → manualPendingFor → retry.
- Edit Tab: change headline + summary → save → replay → 422 inline.
- Disallowed field push (versionType / parentVersionId / etc.) → mapper throws + lint signal.
- User B accepts version owned by user A → 404 cross-user.

## 3 Then

- accept/reject requests are bodyless (argv length = 3, no `manualEditText`).
- IK behaviour: same `(versionId, suggestionId)` replays the cached key; switching suggestions rotates; 422 clears cache.
- 409 SUGGESTION_ALREADY_DECIDED maps to `kind=already_decided` with a localized toast; 404 maps to `kind=cross_user` without leaking ownership info; 422 maps to `kind=validation` inline alert.
- Manual edit: update fires first, accept fires second; saved-manual-pending state shown when accept fails; retry only re-fires accept (no double-write of edit).
- D-12: accept/reject never call updateResumeVersion; structured_profile DOM not mutated by accept alone.
- Edit Tab `updateResumeVersion` payload only carries allowed fields (filterUpdateResumeVersionPayload throws on disallowed); 422 surfaces in-form error; 409 surfaces idempotency-conflict copy; success triggers toast + versionQuery.retry refetch.
- Privacy: originalBullet / suggestedBullet / matchSummary / structuredProfile / manual edit text never leak to URL / pendingAction / localStorage / fetch transport log / toast content.

## 4 Verification Entry

`scripts/trigger.sh` runs:

- `src/app/screens/resume-workshop/tabs/hooks/useTailorSuggestionDecision.test.tsx`
- `src/app/screens/resume-workshop/tabs/hooks/useResumeRewritesActions.test.tsx`
- `src/app/screens/resume-workshop/tabs/hooks/useUpdateResumeVersion.test.tsx`
- `src/app/screens/resume-workshop/tabs/ResumeEditTab.test.tsx`

## 5 Output

- `.test-output/e2e/p0-086-resume-suggestion-accept-reject-edit-and-update-version/trigger.log` Vitest pass.
- verify.sh asserts vitest summary + spec presence + retired-grep gates.

## 6 Baseline

`make codegen-check` clean: accept/reject bodyless POST + updateResumeVersion request type per shared-conventions §5.

## 7 离线限制

Pure fixture-backed Vitest path; offline-friendly.

## 8 方法标注

`method=fixture-backed-frontend`. Backend real route preflight evidence: plan 003 checklist 0.1 + 0.2.
