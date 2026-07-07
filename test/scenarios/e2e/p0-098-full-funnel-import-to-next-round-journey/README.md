# E2E.P0.098 full-funnel-import-to-next-round-journey

> **Status**: Ready
> **Owner plan**: [e2e-scenarios-p0/001](../../../docs/spec/e2e-scenarios-p0/plans/001-full-funnel-happy-journey/plan.md)
> **Spec acceptance**: C-1, C-2, C-3, C-4, C-5, C-6, C-7
> **Isolation**: shared-postgres, scenario-owned user email
> **Parallel-safe**: No

## Scope

API-level full funnel journey from resume seed and JD import to a generated
report and a next-round practice plan. The scenario runs the real `cmd/api`
handler stack, SQL stores, idempotency middleware, async runner, outbox, and
Postgres persistence with deterministic scenario AI.

## Given

- Dev-stack Postgres is reachable through `DATABASE_URL`.
- Migrations are current.
- Scenario-owned test users use `full-funnel-journey@example.com` and
  `full-funnel-seed@example.com`.

## When

Run:

```bash
bash scripts/setup.sh
bash scripts/trigger.sh
bash scripts/verify.sh
bash scripts/cleanup.sh
```

`trigger.sh` executes:

```bash
cd backend && DATABASE_URL=... go test -v ./cmd/api -run '^TestE2EP0098' -count=1
```

## Then

- `registerResume`, `target_import`, `createPracticePlan`, `startPracticeSession`,
  `appendSessionEvent`, `completePracticeSession`, `report_generate`, and
  `createPracticePlan(next_round)` all run through real handlers/stores.
- Runner logs show `resume_parse`, `target_import`, and `report_generate`
  completed.
- Idempotency replay returns the same resource IDs without duplicate side
  effects.
- Empty `focusCompetencyCodes` persists as an empty Postgres array, not NULL.
- Observable logs do not contain private JD text, answer text, report prose, or
  provider/prompt payloads.
- Non-current route vocabulary is rejected with a route-aware pattern while canonical
  operation names remain allowed.
- D-22 non-current debrief/profile API, table, event, job, feature-key, fixture, and
  generated-contract tokens are absent from the API-level core loop.

## Cleanup

`cleanup.sh` deletes scenario-owned users and dependent observable rows. Output
evidence under `.test-output/e2e/p0-098-full-funnel-import-to-next-round-journey/`
is retained for inspection.
