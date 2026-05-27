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

Copy `env-template/dev-real.env.example` to
`.test-output/e2e/p0-100-real-provider-full-funnel-hybrid/dev-real.env` and fill
only local, ignored values there. The required real secrets are:

- `SESSION_COOKIE_SECRET`
- `AUTH_CHALLENGE_TOKEN_PEPPER`
- `AI_PROVIDER_API_KEY`
