# Frontend Resume Workshop Create Flow BDD Checklist

> **版本**: 1.13
> **状态**: active
> **更新日期**: 2026-07-14

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.081 CreateFlow Direct-to-detail Happy Path

- [x] Scenario assets exist for E2E.P0.081.
- [x] Verify script covers upload / paste happy path, IK, direct detail navigation, parser/preview absence, privacy, UI parity assertions, and raw-first-line naming negative. <!-- verified: 2026-07-07 method=scenario scenario=E2E.P0.081 -->
- [ ] P0.081 当前 gate 覆盖 RuntimeConfig/default 10MiB upload、384KiB paste 的 UTF-8 limit/limit+1、zero request、sidebar absent 与 waiting/detail handoff。
- [x] P0.081 或 focused substitute gate 覆盖 Resume upload 仅支持 PDF / Markdown / TXT，DOCX 在 presign/register 前被拒绝。<!-- verified: 2026-07-07 method=focused-substitute tests=UploadTab.test.tsx -->

## E2E.P0.082 Parser UI Absence

- [x] Scenario assets exist under `test/scenarios/e2e/p0-082-resume-create-flow-direct-detail-only/` as direct detail and parser/preview absence evidence.
- [x] Verify script proves parser failure / timeout / retry UI is absent from current create flow.

## E2E.P0.083 CTA Direct-create Handoff

- [x] Scenario assets exist under `test/scenarios/e2e/p0-083-resume-create-flow-direct-create-handoff/`.
- [x] Verify script covers Home CTA, direct detail navigation and auth pending action boundary without preview confirm.

## E2E.P0.084 Historical Home Existing Resume Picker Regression

These checked items are historical evidence. Active 001 Phase 19 replaces the full-Resume/Parse-shared predicate with Home-only `ResumeSummary.parseStatus/hasReadableContent` coverage.

- [x] Focused Home regression test covers readable non-archived `listResumes` records remaining selectable even when `parseStatus` is not `ready`.<!-- verified: 2026-07-08 method=vitest test=HomeResumeSelection.test.tsx -->
- [x] Focused Parse regression test covers the same selectable-resume rule after JD parse handoff.<!-- verified: 2026-07-08 method=vitest test=ParseResumeBinding.test.tsx -->
- [x] Browser screenshot evidence shows the Home resume select populated and not empty.<!-- verified: 2026-07-08 method=playwright-screenshot artifact=.test-output/screenshots/home-resume-picker-fixed-2026-07-08.png -->

## Internal Cleanup Substitute Gate

- [x] Phase 9 source negative, focused create-flow and typecheck gates pass without changing E2E.P0.081-P0.084 behavior.<!-- verified: 2026-07-10 method=substitute-gate evidence="Source gate passed 3/3, create-flow passed 6 files/32 tests, and frontend typecheck passed." -->
