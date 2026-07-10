# Expected outcome

- Path A: `createPracticePlan` receives `goal=retry_current_round` and the source report; `startPracticeSession` creates a fresh session; navigation goes directly to `practice`. The payload carries the retry-focus turn IDs and evidence gaps.
- Path B: `createPracticePlan` receives `goal=next_round`; the canonical interview ladder rotates `roundId`; a fresh session starts and navigation goes directly to `practice`.
- A signed-out report route enters `auth_login` before report data or CTA side effects run. The standalone `replay_practice` pending-action round trip preserves `route=report` and the original safe params.
- No raw answer / question / hint text appears in nav payload, console.log, or route params.
