# Expected Outcome

- `cmd/api` starts a resume runtime with routes, idempotency middleware, deterministic test AI, and an in-process runner kernel handling `resume_parse`.
- The runner kernel claims queued jobs and finalizes async job attempts through the shared runner kernel contract.
- The tracked `resume.parse.default` profile keeps at least 8192 output tokens for structured JSON safety; complete resume text is not echoed by the model.
- A long input's unique tail marker is present in both the AI prompt and the deterministic `parsed_text_snapshot`; business code performs no character/token truncation.
- `finish_reason=length` fails closed as `AI_OUTPUT_INVALID`, preserves the complete source snapshot, and never writes ready state or a completed outbox event.
- Upload and paste sources are read from their own persisted source columns;
  unsupported source types fail validation instead of entering parse.
- Upload PDF / Markdown / text sources write readable prompt input and `parsed_text_snapshot`, not file names, binary bytes, or PDF literal乱码.
- DOCX upload input is rejected before AI and does not write a parse snapshot.
- Unreadable PDF fallback is rejected before AI and never writes a garbage snapshot.
- Queued resume creation keeps `display_name` empty; successful structured-only parse writes ready state, parsed summary, deterministic parsed text snapshot, LLM-derived `displayName`, typed AI task run metadata, and one `resume.parse.completed` outbox event with only envelope fields.
- Invalid output, timeout, and retry-exhausted paths do not emit completed events, keep extracted readable snapshots when available, write failed-with-snapshot fallback `display_name`, and do not introduce `failed_retryable` into `parse_status`.
- Logs, audit rows, task-run metadata, and outbox payloads do not contain raw
  resume content, prompt bodies, or model raw responses.
