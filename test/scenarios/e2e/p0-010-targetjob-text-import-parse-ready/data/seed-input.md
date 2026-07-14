# Seed Input

- User id: deterministic authenticated test user.
- Request: exact `{rawText,targetLanguage,resumeId}` for a Senior Frontend Engineer JD and an existing Resume.
- Headers: `Idempotency-Key` for import and update.
- AI: deterministic fake `target.import.default` response in `APP_ENV=test`.
