# E2E.P0.016 — Parse readonly plan, reports entry, and Start handoff

> **Scenario ID**: E2E.P0.016
> **Owner**: frontend-home-job-picks-and-parse/001-home-jd-import-and-parse
> **Status**: Ready
> **Execution**: Vitest + source contract + frontend build + Playwright

## Given / When / Then

- **Given** an authenticated user opens a saved, readonly interview plan with a backend-owned `resumeId` and 2–5 canonical `TargetJob.summary.interviewRounds[]`.
- **When** the plan detail renders at desktop 1440x900 or mobile 390x844, the user can open `parse-reports-entry` from the content header or start the current round.
- **Then** the entry navigates exactly to `/reports?targetJobId=<uuid>`; Parse does not embed a reports list, does not call `listTargetJobReports`, does not accept the retired `section=reports` parameter, and the global TopBar still has only Home / Interview / Resume. The existing readonly and direct Start behavior remains unchanged.

## Scope

- Keeps Basic fields, requirement evidence, hidden signals, saved round assumptions and bound resume display readonly.
- Keeps Start disabled when the saved plan has no bound resume and prevents route-only resume authority.
- Keeps direct Start on `getPracticePlan` / `createPracticePlan` / `startPracticeSession`, with no `updateTargetJob` call.
- Adds one page-level `parse-reports-entry` in the plan-detail header; the entry is not a TopBar item.
- Treats `/parse?...&section=reports` as hostile legacy input: the canonical URL drops `section`, the page does not scroll/focus a report section, and no embedded report DOM appears.
- Leaves report overview loading and rendering to independent `ReportsScreen`; P0.059 owns its current-plan list and state parity.

## Deterministic evidence

- `scripts/source_contract_test.py` proves `ReportsScreen.tsx` is the only production screen consumer of `listTargetJobReports`, Parse has exactly one entry and no embedded list, `PARSE_SAFE` has no `section`, and TopBar has no reports entry.
- `ParseReports.test.tsx`, route and TopBar unit tests prove the exact handoff and negative contracts.
- `frontend/tests/pixel-parity/parse.spec.ts` compares the formal/prototype entry at 1440x900 and 390x844 using DOM text, computed style, absolute bounding boxes and pixelmatch ≤0.5%; it attaches formal/prototype entry screenshots.
- The same Playwright run retains readonly plan and direct Start regression coverage.

## Scripts

- `scripts/setup.sh` — resets the scenario-owned output and writes `setup.env`.
- `scripts/trigger.sh` — runs source, focused Vitest, frontend build and desktop/mobile Playwright gates.
- `scripts/verify.sh` — binds actual runner/pass markers and the entry/no-request/no-TopBar/no-section evidence.
- `scripts/cleanup.sh` — removes only the transient setup marker and leaves current-run logs for inspection.

## Environment

This scenario is self-contained and uses fixture-backed Vitest/Playwright plus the static parity server. It does not require the shared Docker or host-run backend/frontend environment.
