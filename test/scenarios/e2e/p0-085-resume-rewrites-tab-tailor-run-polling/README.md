# E2E.P0.085 Resume Rewrites Tab + Tailor Run Polling + Rerun + Ready/Failed/Timeout

> **场景 ID**: E2E.P0.085
> **执行方式**: automated (vitest jsdom)
> **隔离级别**: in-process (vitest worker)
> **状态**: Ready

## 1 Given

- Fixture-backed mock-first client: `Resumes/branchResumeVersion.json ai-select-202-with-job` + `ResumeTailor/requestResumeTailor.json default / idempotency-replay` + `ResumeTailor/getResumeTailorRun.json queued / generating / default(ready) / failed` + `Resumes/getResumeVersion.json targeted-with-suggestions`.
- Deterministic harness: `useResumeTailorRunPolling` consumes ordered fixture responses via Vitest fake-timer chain (no synthetic schema).
- User authenticated; lang default; MASTER + targetJobId already resolved (Phase 1 source state).
- Phase 0 real-backend preflight: `requestResumeTailor` + `getResumeTailorRun` generated client + server interface + handler + cmd/api route real.

## 2 When

- ai_select branch submission lands → URL carries `tailorRunId` → Rewrites Tab mounts and starts polling.
- Polling sequence queued → generating → ready; the `onReady` callback refetches the version.
- User clicks `重新运行改写` with `mode='gap_review'` → `useRequestResumeTailor` requests a new tailor run.
- Second sequence: queued → generating → failed → user clicks Retry → restart polling.
- Third sequence: max-attempts overflow → timeout.
- Switch to Edit tab and back to Rewrites tab.
- Component unmount (back to list / route change).

## 3 Then

- Polling banner testid `resume-rewrites-polling-banner` renders while polling, replaced by failed banner (`resume-rewrites-failed-banner` + role=alert) on terminal failure / timeout.
- `getResumeTailorRun` is called with a single positional arg (no `Idempotency-Key`), proving read-only no-IK contract.
- `requestResumeTailor` carries `Idempotency-Key` (v1 wire format); same body fingerprint replays the key, mode change rotates it.
- After ready, `getResumeVersion` is refetched (covered by ResumeRewritesTabContainer.onVersionRefreshed via versionQuery.retry).
- Failed / timeout banners surface a localized retry CTA (`resume-rewrites-polling-retry`); the underlying hook `retry()` resets the attempt counter and re-enters polling.
- Unmount cancels the active setTimeout — Vitest fake timers verify no further `getResumeTailorRun` calls after unmount.
- Privacy: originalBullet / suggestedBullet / matchSummary text never leaks into URL / localStorage / fetch transport log.
- `method=fixture-backed-frontend` with backend real route preflight evidence.

## 4 Verification Entry

`scripts/trigger.sh` runs:

- `src/app/screens/resume-workshop/tabs/ResumeRewritesTab.test.tsx`
- `src/app/screens/resume-workshop/tabs/hooks/useResumeTailorRunPolling.test.tsx`
- `src/app/screens/resume-workshop/tabs/hooks/useRequestResumeTailor.test.tsx`

## 5 Output

- `.test-output/e2e/p0-085-resume-rewrites-tab-tailor-run-polling/trigger.log` Vitest pass output.
- verify.sh asserts vitest summary + spec presence + retired-grep gates.

## 6 Baseline

`make codegen-check` clean: `requestResumeTailor` request type + `getResumeTailorRun` response type + status enum `queued|generating|ready|failed` per shared-conventions §5.

## 7 离线限制

Pure fixture-backed Vitest path; offline-friendly.

## 8 方法标注

`method=fixture-backed-frontend`. Backend real route preflight evidence: plan 003 checklist 0.3.
