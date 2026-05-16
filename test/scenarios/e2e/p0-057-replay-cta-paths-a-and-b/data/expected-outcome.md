# Expected outcome

- Path A: `nav("workspace", {…})` is invoked with `practiceGoal=retry_current_round`, `replayItems` containing the retry-focus turn IDs, and `autoStartPractice=1`; workspace then creates a fresh practice session before entering `practice`.
- Path B: `nav("workspace", {…})` is invoked with `practiceGoal=next_round`, `autoStartPractice=1`, and a rotated `roundId` derived from the canonical interview ladder; workspace then creates a fresh practice session.
- Unauthenticated branch routes to `auth_login` carrying the `replay_practice` pendingAction; `route=workspace` and `autoStartPractice=1` are preserved across encode → decode.
- No raw answer / question / hint text appears in nav payload, console.log, or route params.
