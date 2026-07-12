# E2E.P0.098 Conversation Funnel Contract Composition

> **Status**: Ready
> **Owner plan**: [e2e-scenarios-p0/001](../../../../docs/spec/e2e-scenarios-p0/plans/001-full-funnel-happy-journey/plan.md)
> **Execution**: automated
> **Parallel-safe**: No

## Scope

This scenario composes the current backend contract gates for the path from a
ready resume and target job through practice-plan creation, continuous chat,
session completion, and conversation-level report persistence. It deliberately
does not maintain a second scenario-only API server or duplicate product
orchestration.

## Contract

- `createPracticePlan` accepts omitted/empty focus codes and stores `{}`, never
  `NULL`.
- `startPracticeSession` and `sendPracticeMessage` use
  `practice_messages`; no question/turn/event endpoint is part of the loop.
- completion writes only lifecycle event columns and queues one report job.
- report persistence writes PostgreSQL `text[]` values correctly and report
  retries may re-enter the `generating` state idempotently.
- resume parsing uses the standard AI observability wrapper.

Run `scripts/setup.sh`, `trigger.sh`, `verify.sh`, and `cleanup.sh` in order.
