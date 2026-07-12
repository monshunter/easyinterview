# E2E.P0.099 Real Continuous-Conversation Full-Stack Journey

> **Status**: Ready
> **Owner plan**: [e2e-scenarios-p0/001](../../../../docs/spec/e2e-scenarios-p0/plans/001-full-funnel-happy-journey/plan.md)
> **Execution**: hybrid
> **Isolation**: shared host-run environment, synthetic account/data
> **Parallel-safe**: No

## Scope

Browser acceptance against the shared real frontend, backend, PostgreSQL,
Mailpit login flow, and configured real AI provider. The journey covers resume
creation, JD import, continuous practice chat, completion, asynchronous report
generation, and the conversation-level report.

## Required observations

- practice is one full-width chat window with no question sidebar, question
  counter, or current-question heading;
- the phone control is visible but natively disabled;
- one user message produces one natural assistant continuation through
  `sendPracticeMessage`;
- completion reaches a ready report with dimensions, conversation evidence,
  risks, and next action;
- PostgreSQL contains non-null empty plan focus codes, ordered
  `practice_messages`, a completed report job, and a ready report;
- desktop and mobile screenshots exist for practice and report.

Run the four scripts in order. `setup.sh` creates a run ID. A browser agent or
human then writes a redacted `evidence.md` with that run ID and the observations
above. `trigger.sh` runs current code gates and returns `PASS` only when the
current evidence and all four screenshots exist.

Evidence stays under:

```text
.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey/
```
