# E2E.P0.087 Resume Detail Export PDF + Copy Text Consistency + Three-Screen UI Parity + Retired-Module Negative

> **场景 ID**: E2E.P0.087
> **执行方式**: automated (vitest jsdom + frontend build + Playwright pixel parity / axe + repo greps)
> **隔离级别**: in-process (vitest worker) + static dist server (Playwright)
> **状态**: Ready

## 1 Given

- Plan 001 P0.037 verification still holds: `exportResumeVersion.json p0-501-not-available` toast and `buildResumePlainText` clipboard fallback are tested in `ResumeDetailExport.test.tsx`.
- Plan 003 mounts ResumeBranchFlow / ResumeRewritesTab / ResumeEditTab; the source-level mirror is enforced via Vitest DOM testid assertions plus Playwright viewport/style/screenshot/axe checks.
- Authenticated user; lang default; targeted version v1 ready.
- Phase 0 real-backend preflight: branch / suggestion / update / tailor / version-read ops all real on backend-resume/002.

## 2 When

- Render BranchFlow / Rewrites Tab / Edit Tab in jsdom (desktop default viewport) and assert the source-level mirror DOM anchors.
- Build `frontend/dist`, then render BranchFlow / Rewrites Tab / Edit Tab in Playwright desktop + mobile projects and assert DOM anchors, computed style, viewport-safe bounding boxes, screenshot buffers, and scoped axe-core checks.
- Export PDF button click on the detail view → existing ResumeDetailExport.test asserts Idempotency-Key + 501 toast.
- Copy plain text click → existing ResumeDetailExport.test asserts clipboard write + fallback message.
- Run repo-wide greps that mirror plan 003 §7.10-7.12 closure within `branch/` and `tabs/` write-scope.

## 3 Then

- Plan 001 Export PDF P0 stub behaviour is unchanged after Plan 003 lands: `exportResumeVersion` request still carries Idempotency-Key, response 501 maps to `PDF 导出能力即将开放` toast, no blob is written.
- Copy plain text continues to call `navigator.clipboard.writeText` with the `buildResumePlainText` projection, falling back to the `Clipboard write unavailable` toast on errors.
- BranchFlow / RewritesTab / EditTab DOM anchor + state attributes (`data-edit-dirty`, `data-bullet-count`, `data-branch-can-submit`, ...) prove the source-level mirror.
- Playwright proves the three screens are non-blank, viewport-safe, and axe-clean on desktop + mobile, using `frontend/tests/pixel-parity/resume-workshop-branch-rewrites-edit.spec.ts`.
- Retired grep: `welcome|mistake|growth|drill|followup|STAR|experiences|voice|OnboardingScreen|onboarding=true` 0 hits in `branch/` + `tabs/`.
- Retired tailor mode grep: `(inline|rewrite|mirror)` 0 hits in `branch/` + `tabs/` (B3 D-14 alignment).
- Prototype import grep: `ui-design/src/(data|screen-resume-workshop)` 0 hits in `branch/` + `tabs/`.
- Privacy: structured profile / suggestion text never appears in URL / localStorage / fetch transport log (covered by per-tab privacy specs).

## 4 Verification Entry

`scripts/trigger.sh` runs:

- `src/app/screens/resume-workshop/components/ResumeDetailExport.test.tsx`
- `src/app/screens/resume-workshop/components/ResumeDetailFixtureParity.test.tsx`
- `src/app/screens/resume-workshop/branch/ResumeBranchFlow.test.tsx`
- `src/app/screens/resume-workshop/tabs/ResumeRewritesTab.test.tsx`
- `src/app/screens/resume-workshop/tabs/ResumeEditTab.test.tsx`
- `pnpm --filter @easyinterview/frontend build`
- `frontend/tests/pixel-parity/resume-workshop-branch-rewrites-edit.spec.ts`

## 5 Output

- `.test-output/e2e/p0-087-resume-detail-export-copy-consistency-and-parity/trigger.log` Vitest + build + Playwright pass.
- verify.sh asserts Vitest summary, build marker, Playwright runner/pass summary, spec presence, and all three grep gates land at 0 hits.

## 6 Baseline

- `exportResumeVersion` fixture `p0-501-not-available` byte-stable.
- `buildResumePlainText` adapter byte-stable; covered by plan 001 ResumeDetailExport.test.tsx.

## 7 离线限制

Vitest + local build + Playwright static-server path; no external network dependency.

## 8 方法标注

`method=fixture-backed-frontend`. Backend real route preflight evidence: plan 003 checklist 0.1-0.4 + plan 001 P0.037 baseline.
