# Seed input

- Route entry: `generating?reportId=01918fa0-0000-7000-8000-000000007000&sessionId=01918fa0-0000-7000-8000-000000005000&targetJobId=01918fa0-0000-7000-8000-000000002000&resumeId=01918fa0-0000-7000-8000-000000004000&planId=plan-1&roundId=round-tech-1&mode=text&modality=text&practiceMode=strict&hintUsed=false&hintCount=0`.
- Deterministic test clients: each focused file controls its own `getFeedbackReport`, `getTargetJob` and `getResume` responses; the scenario does not impose one cross-file request sequence.
- Auth state: signed-in user from `openapi/fixtures/Auth/getMe.json`.
