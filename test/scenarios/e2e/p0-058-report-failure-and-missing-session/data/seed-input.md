# Seed input

- Existing schema-valid P0.047 `completion-backend-evidence.json`; the backend evidence test receives its path through `PRACTICE_COMPLETION_EVIDENCE_PATH`.
- Isolated backend cases: missing/mismatched frozen coordinate, an in-memory exactly 917,505-byte framed input, invalid→valid multi-attempt recovery, four-invalid terminal failure, exact action-local `10s/20s/40s` waits, first-action exhaustion followed by a second independent invocation at attempt one, and async-job attempt values that do not change product retry state. No `input-*.json` boundary file is required.
- API/frontend states keyed only by `reportId`: missing ID, queued/generating, ready, not-found, timeout/network transport exhaustion, terminal failed, `REPORT_CONTEXT_TOO_LARGE` and malformed direct-ready shape.
- Route noise may contain conflicting session/status/target/resume/error values; tests prove the API projection remains authoritative.
