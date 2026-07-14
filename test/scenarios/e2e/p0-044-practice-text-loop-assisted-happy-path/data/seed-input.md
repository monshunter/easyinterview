# Seed input

Use the checked-in Practice default and `reply-pending` OpenAPI fixtures plus the test-local persisted user/assistant GFM projection for browser evidence; no API fixture or generated client is changed. Focused loader/screen/API/service/repository tests use only their existing in-memory or SQL-mock fixtures, including the exact pre-90-second pending boundary; this scenario does not mutate PostgreSQL. All evidence is bound to the shared Practice source manifest.
