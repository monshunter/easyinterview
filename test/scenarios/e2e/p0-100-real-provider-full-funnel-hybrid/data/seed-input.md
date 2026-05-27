# E2E.P0.100 Seed Input

This hybrid scenario uses synthetic user-entered materials, not mock responses.

## Account

- Email: `manual-uat-full-funnel@example.test`
- Local mailbox: `http://127.0.0.1:8025`
- Auth flow: real passwordless challenge through Mailpit magic link

## Materials

- JD: `data/jd-backend-engineer.zh.md` or `data/jd-backend-engineer.en.md`
- Resume: `data/resume-backend-engineer.zh.md` or `data/resume-backend-engineer.en.md`
- Practice answer: `data/answer-sample-backend-engineer.zh.md` or `data/answer-sample-backend-engineer.en.md`

## Runtime Inputs

Use `deploy/dev-stack/.env` as the single local real-environment file for
shared dependencies, host-run backend, frontend real mode, and real AI provider
adaptation. The required real secrets are:

- `SESSION_COOKIE_SECRET`
- `AUTH_CHALLENGE_TOKEN_PEPPER`
- `AI_PROVIDER_API_KEY`
