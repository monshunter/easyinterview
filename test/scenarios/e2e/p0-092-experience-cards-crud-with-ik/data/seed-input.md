# Seed Input — E2E.P0.082

Shares same user / settings fixtures with E2E.P0.081 (users A / B / C) — see
`../p0-081-candidate-profile-seed-and-patch/data/seed-input.md`.

`TestProfileHTTPScenario` is the single Go test that exercises both P0.081 and
P0.082 trigger paths; per-scenario verify scripts pivot on different output
expectations.
