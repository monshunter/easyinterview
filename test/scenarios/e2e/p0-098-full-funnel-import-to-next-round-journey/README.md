# E2E.P0.098 Persisted Interview Round Journey

> **Status**: Ready
> **Owner plan**: [e2e-scenarios-p0/001](../../../../docs/spec/e2e-scenarios-p0/plans/001-full-funnel-happy-journey/plan.md)
> **Execution**: automated
> **Parallel-safe**: No
> **Isolation**: fixed scenario user in shared local PostgreSQL; targeted cleanup only

## Scope

This scenario composes the current production store, service, generated client,
and UI contracts for the path from a ready resume and TargetJob through exact
round plan creation, session completion, TargetJob progress refresh, report
handoff, and next-round start. PostgreSQL is the business-state truth source;
the scenario deliberately keeps no second progress cursor in browser storage or
fixture-only orchestration.

## Contract

- `practice_plans` persists a paired `roundId + roundSequence`; baseline,
  retry-current, and next-round selection are validated against canonical
  TargetJob rounds and immutable completion facts.
- `startPracticeSession` and `sendPracticeMessage` use
  `practice_messages`; the AI context uses the full persisted resume snapshot
  and the exact persisted round name/type/focus.
- completion writes exactly one `session_completed` lifecycle fact and queues
  one report job; replay cannot advance progress twice.
- TargetJob Get/List project an ordered completed prefix plus the first
  incomplete round from persisted facts, independent of TargetJob/report
  lifecycle status. The current ready plan must match the exact current round
  and current resume.
- Home, Workspace list/detail, Report, and quick-start consume that backend
  projection. A ready Home/Workspace card body opens the read-only
  `/workspace?targetJobId=<id>` detail directly; `targetJobId` is the only URL
  locator. Parse remains an import command/progress route for queued or
  processing imports and is not used to read a ready plan. Equal-duration
  wrong-round and legacy-null plans are not reused; final/invalid progress
  fails closed.
- report persistence writes PostgreSQL `text[]` values correctly and report
  retries may re-enter the `generating` state idempotently.
- resume parsing uses the standard AI observability wrapper.
- non-test frontend source may persist only display preferences; interview
  progress/plan/session/report state is never written to browser storage.
- a real Playwright browser logs in through the host-run email-code API, sees
  round 1 as current, completes the persisted round-1 session through the real
  API, reloads Workspace, and sees round 1 done plus round 2 current. It then
  clicks the ready Workspace-list card and ready Home card in turn; both land
  directly on the target-scoped Workspace detail with no Parse animation,
  import command, or polling.
- the Workspace detail renders three distinct persisted round states after
  completion: `done/current/pending`, Chinese labels
  `已进行/即将进行/未进行`, and pairwise-distinct computed backgrounds and
  borders. The URL, labels, state attributes, and treatments survive reload.
- Workspace quick-start sends a real `POST /practice/plans` with the backend
  current `roundId`; the 201 response and a subsequent real GET both prove the
  normalized `roundSequence`. Only `POST /practice/sessions` is intercepted so
  this progress gate does not invoke the AI interviewer opening turn.

## Runtime prerequisites

The shared host-run environment must already be current and healthy:

- frontend origin from `FRONTEND_HOST_PORT` (default `127.0.0.1:5173`);
- backend API from `API_HOST_PORT` (default `127.0.0.1:8080/api/v1`);
- Mailpit API from `MAILPIT_WEB_HOST_PORT`;
- PostgreSQL with migration `000017_practice_plan_round_identity` applied.

`setup.sh` reads only `deploy/dev-stack/.env`, runs endpoint smoke checks, and
inserts one fixed `.example.test` user plus its isolated resume, TargetJob,
round-1 plan, and round-1 session with one completed user/assistant turn. It
does not bootstrap, restart, or reset the shared environment.

## Given / When / Then

- **Given** canonical TargetJob rounds `1, 2, 4`, an exact ready round-1 plan,
  and a waiting reportable round-1 session with one completed turn for the
  fixed scenario user.
- **When** Playwright logs in via Mailpit, calls the real completion endpoint,
  reloads Workspace, clicks its ready card body, reloads the resulting detail,
  then repeats the direct-detail navigation from the ready Home card before
  returning to the Workspace list and clicking its start button.
- **Then** the rail changes from `current,pending,pending` to
  `done,current,pending`; TargetJob returns completed round 1 and current round
  2; the same state survives real Home and Workspace-detail reloads. The detail
  URL is exactly `/workspace?targetJobId=<id>` from both card sources,
  `getTargetJob` executes once per detail visit, and no Parse/import/poll
  request occurs. Its three cards retain
  `已进行/即将进行/未进行` plus distinct background/border treatments; real plan
  creation uses `round-2-technical` and persists sequence 2.
- **And** the browser proceeds to the practice route using an intercepted
  session-start response, proving no AI interviewer opening call is required by
  this gate.

## Cleanup

`cleanup.sh` deletes only rows connected to user
`019f6098-0000-7000-8000-000000000001`, including dynamic report jobs,
outbox/audit evidence, auth challenge/session rows, the real round-2 plan, and
the seeded business rows. It does not clear Mailpit globally, stop services,
delete volumes, or touch other users.

Run `scripts/setup.sh`, `trigger.sh`, `verify.sh`, and `cleanup.sh` in order.
