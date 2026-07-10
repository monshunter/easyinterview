# E2E.P0.057 — Replay CTA path A + path B

> **Owner**: frontend-report-dashboard/001-report-screen-and-generating-handoff
> **Coverage tags**: C-3 / C-6
> **Execution**: Vitest

Verifies the replay flow:
- Path A `goReplay()` carries `retry_current_round` payload with the report-derived `retryFocusTurnIds`.
- Path B `goNextRound()` rotates `roundId` via `inferNextRoundId` and carries `practiceGoal=next_round`.
- Authenticated users create a derived plan, start a fresh session and land directly on practice through the generated client.
- A signed-out report route is auth-gated before CTA mount; the separate `replay_practice` pending-action contract returns to the same report route for retry.
- Privacy red lines: no raw answer / question / hint text in payload.

The trigger also runs the report-owner preflight so the spec, plan/BDD/test documents, direct-start source and this scenario cannot drift back to a workspace route side effect.
