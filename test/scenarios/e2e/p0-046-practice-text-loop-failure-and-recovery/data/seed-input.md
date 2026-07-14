# Seed input

Use the checked-in Practice `reply-retryable-failed`, retry-success, and `reply-terminal-failed` OpenAPI fixtures for browser evidence. The trigger derives a unique empty database from the server credentials in `DATABASE_URL`, applies current migrations, and runs the previous recovery test plus four exact independent-connection lease/generation/concurrency tests before force-dropping that database. All browser/backend evidence is bound to the shared Practice source manifest.
