# Seed Input

- User A: `01918fa0-0000-7000-8000-0000000000a1`, cookie `raw-session-token-a`.
- User B: `01918fa0-0000-7000-8000-0000000000b1`, cookie `raw-session-token-b`.
- Target job fixture: `target-job-p0-022-a`, owned by User A.
- Resume asset fixture: `resume-asset-p0-022-a`, owned by User A.
- Request: `goal=baseline`, `mode=assisted`, `interviewerPersona=hiring_manager`, `difficulty=standard`, `language=zh-CN`.
- Header: `Idempotency-Key: e2e-p0-022-create-plan`.
