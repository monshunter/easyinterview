# Expected outcome

- Path A: `nav("practice", {…})` invoked with `practiceGoal=retry_current_round`, `replayItems` containing the retry-focus turn IDs.
- Path B: `nav("practice", {…})` invoked with `practiceGoal=next_round` and a rotated `roundId` derived from the canonical interview ladder.
- Unauthenticated branch routes to `auth_login` carrying the `replay_practice` pendingAction; `autoReplay=1` is preserved across encode → decode.
- No raw answer / question / hint text appears in nav payload, console.log, or route params.
