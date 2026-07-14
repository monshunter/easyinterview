# E2E.P0.098 Persisted Interview Round Journey

> **Status**: Ready
> **Owner plan**: [e2e-scenarios-p0/001](../../../../docs/spec/e2e-scenarios-p0/plans/001-real-api-ui-journeys/plan.md)
> **Execution**: automated Playwright against the host-run real stack
> **Parallel-safe**: No
> **Isolation**: fixed scenario user in shared local PostgreSQL; targeted cleanup only

## Scope

This scenario retains only the real user journey that requires a running
frontend, backend, PostgreSQL, and Mailpit. The browser signs in through the
production email-code flow, reads and mutates business state through live HTTP,
reloads user-visible pages, and verifies the persisted round projection.

The frontend must be running with `VITE_EI_API_MODE=real` and
`VITE_EI_API_BASE_URL` equal to the scenario backend API. A Playwright process,
fixture transport, source check, package test, build, or direct database query
does not qualify by itself. Database commands in `setup.sh` and `cleanup.sh`
only isolate scenario data; they are not E2E evidence.

## Given / When / Then

- **Given** a fixed real user with a ready resume, a three-round TargetJob, an
  exact round-1 plan, and a reportable waiting session in the shared local
  PostgreSQL database.
- **When** Playwright opens the host-run frontend, logs in through Mailpit and
  the real auth API, completes round 1 through
  `POST /practice/sessions/{sessionId}/complete`, reloads Workspace and Home,
  and opens the target detail from both cards.
- **Then** the browser observes `current,pending,pending` becoming
  `done,current,pending`; Home, Workspace list, and Workspace detail agree after
  reload; both cards route directly to
  `/workspace?targetJobId=019f6098-0000-7000-8000-000000000003`; and the real
  TargetJob response persists
  `currentRound={roundId: round-2-technical, roundSequence: 2}`.

The Playwright spec does not intercept or fulfill application requests. Every
request used by the scenario reaches the host-run backend.

## Runtime prerequisites

- the shared host-run frontend and backend are current and healthy;
- Mailpit is reachable through the port in `deploy/dev-stack/.env`;
- PostgreSQL has migration `000017_practice_plan_round_identity` applied;
- `deploy/dev-stack/.env` selects the real frontend API transport.

`setup.sh` checks those endpoints and inserts only the fixed scenario rows. It
does not bootstrap, restart, or reset the shared environment.

## Cleanup

`cleanup.sh` deletes only rows connected to user
`019f6098-0000-7000-8000-000000000001`. It does not clear Mailpit globally,
stop services, delete volumes, or touch other users.

Run `scripts/setup.sh`, `trigger.sh`, `verify.sh`, and `cleanup.sh` in order.
