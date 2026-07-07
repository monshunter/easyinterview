# Expected Outcome

- `cmd/api` starts a resume runtime with routes, idempotency middleware, deterministic test AI, and an in-process drainer handling `resume_parse`.
- The drainer claims queued jobs and finalizes async job attempts through the shared `targetjob` drainer contract.
- Upload and paste sources are read from their own persisted source columns;
  unsupported source types fail instead of deserializing non-current guided fields.
- Upload PDF / DOCX / Markdown / text sources write readable prompt input and `parsed_text_snapshot`, not file names or binary bytes.
- Queued resume creation keeps `display_name` empty; successful parse writes ready state, parsed summary, parsed text snapshot, LLM-derived `displayName`, typed AI task run metadata, and one `resume.parse.completed` outbox event with only envelope fields.
- Invalid output, timeout, and retry-exhausted paths do not emit completed events and do not introduce `failed_retryable` into `resume_assets.parse_status`.
- Parse completion before Preview Confirm does not create `resume_versions` rows.
- Logs, audit rows, task-run metadata, and outbox payloads do not contain raw
  resume content, prompt bodies, or model raw responses.
