# Seed Input

- User A: `01918fa0-0000-7000-8000-0000000000a1`, cookie `raw-session-token-a`.
- Ready plan: `practice-plan-p0-023`, `goal=baseline`, `mode=assisted`, `language=zh-CN`.
- Target job fixture: `target-job-p0-023-a`, owned by User A.
- Resume asset fixture: `resume-asset-p0-023-a`, owned by User A.
- F3 fake resolution: `featureKey=practice.session.first_question`, `promptVersion=prompt.v1`, `rubricVersion=rubric.v1`, `modelProfileName=practice.first_question.default`.
- A3 fake AI response: JSON with `questionText` and `questionIntent=behavioral.leadership.design_system`.
- Header: `Idempotency-Key: e2e-p0-023-start-session`.
