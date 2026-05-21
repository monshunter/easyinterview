# E2E.P0.094 seed input

Two test users (A / B). User A is provisioned with:

- candidate_profile row (headline + yearsOfExperience set).
- 3 resume_assets rows.
- 5 target_jobs rows.
- 8 practice_sessions rows (any status).
- 2 debriefs rows.
- 25 jd_match_recommendations rows with mixed scores so cursor
  pagination is non-trivial.
- 1 agent_scans row (status='idle', last_scan_at = 2h ago).

User B is fully provisioned at the auth layer but has no JD-Match
data, so cross-user reads return 404 RESOURCE_NOT_FOUND.
