# E2E.P0.097 expected outcome

- agent_scan job state machine: scanning -> idle (last_scan_at +
  next_scan_at + recommendation_count updated) on success;
  scanning -> error (error_message redacted) on AI failure.
- outbox `jd_match.recommendation.completed` envelope contains only
  userId / agentScanId / recommendationCount / completedAt.
- DeleteJobMatchDataForUser cascade order recorded in audit_events
  tombstone with counts and userId only (no PII text).
- Cross-user data (B / C) untouched.
- Partial failure tombstone records the watchlist + saved_searches +
  search_runs counts that landed before the simulated rollback.
