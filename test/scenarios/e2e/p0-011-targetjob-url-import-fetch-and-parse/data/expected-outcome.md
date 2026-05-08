# Expected Outcome

- Stored `target_job_sources.url` omits query, fragment, and userinfo.
- URL fetch persists sanitized URL, snapshot text, fetched timestamp, and `freshness_status=fresh`.
- Parse succeeds and emits privacy-safe `target.parsed`.
- `source_refresh` placeholder is enqueued without exposing source URL.
- Invalid and unavailable URL paths map to B1 `TARGET_IMPORT_SOURCE_INVALID` / `TARGET_IMPORT_SOURCE_UNAVAILABLE`.
