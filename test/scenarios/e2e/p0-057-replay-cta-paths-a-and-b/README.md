# E2E.P0.057 — Replay CTA path A + path B

> **Owner**: frontend-report-dashboard/001-report-screen-and-generating-handoff
> **Coverage tags**: C-3 / C-6
> **Execution**: Vitest

Verifies the replay flow:
- Path A `goReplay()` carries `retry_current_round` payload with the report-derived `retryFocusTurnIds`.
- Path B `goNextRound()` rotates `roundId` via `inferNextRoundId` and carries `practiceGoal=next_round`.
- Authenticated users land directly on practice; unauthenticated users route through `useRequestAuth({type:'replay_practice'})`.
- Privacy red lines: no raw answer / question / hint text in payload.
