# Seed Input

- User A: `01918fa0-0000-7000-8000-0000000000a1`, cookie `raw-session-token-a`.
- Ready plan: `practice-plan-p0-024`, `goal=baseline`, `mode=assisted`, `language=zh-CN`.
- Header: `Idempotency-Key: e2e-p0-024-start-session`.
- First A3 fake AI result: `AI_PROVIDER_TIMEOUT`.
- Retry A3 fake AI result: first-question JSON with `questionText` and `questionIntent`.
