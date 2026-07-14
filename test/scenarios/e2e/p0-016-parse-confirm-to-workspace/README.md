# E2E.P0.016 — Parse terminal handoff and Workspace readonly plan

> **Scenario ID**: E2E.P0.016
> **Owner**: frontend-home-job-picks-and-parse/001-home-jd-import-and-parse
> **Status**: Ready
> **Execution**: Vitest + source contract + frontend build + Playwright

## Given / When / Then

- **Given** an authenticated user has a newly imported TargetJob that becomes ready, or opens an already-ready plan with a backend-owned `resumeId` and 2–5 canonical `TargetJob.summary.interviewRounds[]`.
- **When** Parse observes ready it history-replaces to `/workspace?targetJobId=<uuid>`; saved cards and direct links open that Workspace detail immediately at desktop 1440x900 or mobile 390x844.
- **Then** no Parse animation or saved-plan preview is shown for ready data. The Workspace detail can open `/reports?targetJobId=<uuid>` or start the current round, while report rows remain owned by `ReportsScreen` and the global TopBar still has only Home / Interview / Resume.

## Scope

- Keeps Basic fields, requirement evidence, hidden signals, saved round assumptions and bound resume display readonly on the target-scoped Workspace route.
- Parse is command/progress-only: queued/processing may poll; ready initial reads and ready poll results use replace navigation to Workspace, with no artificial preview delay.
- Keeps Start disabled when the saved plan has no bound resume and prevents route-only resume authority.
- Keeps direct Start on `getPracticePlan` / `createPracticePlan` / `startPracticeSession`, with no `updateTargetJob` call.
- Keeps one page-level `parse-reports-entry` DOM anchor in the shared plan-detail header; the entry is not a TopBar item and does not make the route a Parse page.
- Treats `/parse?...&section=reports` and `/workspace?...&section=reports` as hostile legacy input: canonical URLs drop `section`, and no embedded report DOM appears.
- Leaves report overview loading and rendering to independent `ReportsScreen`; P0.059 owns its current-plan list and state parity.

## Deterministic evidence

- `scripts/source_contract_test.py` proves `ReportsScreen.tsx` is the only production screen consumer of `listTargetJobReports`, the shared Workspace detail has one entry and no embedded list, Parse/Workspace safe params contain only `targetJobId`, and TopBar has no reports entry.
- `ParseFlow.test.tsx`, `ParseReports.test.tsx`, route and TopBar unit tests prove the terminal replace, exact report handoff and negative contracts.
- `frontend/tests/pixel-parity/parse.spec.ts` compares the Workspace detail/report entry and done/current/pending round-state surfaces at 1440x900 and 390x844 using DOM text, computed style, absolute bounding boxes and screenshots.
- The same Playwright run retains readonly plan and direct Start regression coverage.

## Scripts

- `scripts/setup.sh` — resets the scenario-owned output and writes `setup.env`.
- `scripts/trigger.sh` — runs source, focused Vitest, frontend build and desktop/mobile Playwright gates.
- `scripts/verify.sh` — binds actual runner/pass markers and the entry/no-request/no-TopBar/no-section evidence.
- `scripts/cleanup.sh` — removes only the transient setup marker and leaves current-run logs for inspection.

## Environment

This scenario is self-contained and uses fixture-backed Vitest/Playwright plus the static parity server. It does not require the shared Docker or host-run backend/frontend environment.
