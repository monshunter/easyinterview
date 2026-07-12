# Expected outcome

- Path A: `createPracticePlan` receives `goal=retry_current_round` and the source report; `startPracticeSession` creates a fresh session; navigation goes directly to `practice`. The payload carries competency focus codes and evidence gaps.
- Path B: `createPracticePlan` receives `goal=next_round`; the immediate structured successor supplies `roundId` and `timeBudgetMinutes`; a fresh session starts and navigation goes directly to `practice`.
- Final/single/empty/unknown/loading/failure and duplicate-derived-ID state keeps next-round disabled. Either in-flight CTA disables both buttons, and repeated clicks create at most one plan/session.
- A signed-out report route enters `auth_login` before report data or CTA side effects run. The standalone `replay_practice` pending-action round trip preserves `route=report` and the original safe params.
- No raw answer / question / hint text appears in nav payload, console.log, or route params.
