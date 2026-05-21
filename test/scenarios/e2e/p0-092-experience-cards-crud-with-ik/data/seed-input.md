# Seed Input — E2E.P0.092

Shares same user / settings fixtures with E2E.P0.091 (users A / B / C) — see
`../p0-091-candidate-profile-seed-and-patch/data/seed-input.md`.

`TestProfileHTTPScenario` is the single Go test that exercises both P0.091 and
P0.092 trigger paths; per-scenario verify scripts pivot on different output
expectations.
