# E2E.P0.097 jd-match-agent-scan-and-privacy-delete

> **Status**: scaffold (live DB + cmd/api drainer registration required)
> **Owner plan**: [backend-jobs-recommendations/001](../../../docs/spec/backend-jobs-recommendations/plans/001-jd-match-real-backend-baseline/plan.md)
> **Spec acceptance**: C-12, C-15, C-16, C-18

## Scope

- `jd_match_agent_scan` job lifecycle: scanning -> idle / error.
- Recommendation generator inline upsert + outbox
  `jd_match.recommendation.completed` event (PII boundary: no
  reasons / risks / source_url / interview_hypotheses).
- `DeleteJobMatchDataForUser` cascade order watchlist_items ->
  saved_searches -> jd_match_search_runs -> jd_match_recommendations ->
  agent_scans + audit_events tombstone with only userId / counts /
  timestamps / job_id.
- Failure rollback: simulated recommendation delete error short-
  circuits and writes a partial tombstone.

## Live-environment prerequisite

cmd/api in-process drainer registered with the
`jd_match_agent_scan` handler + stub AIClient with success /
output_invalid variants.

## Deferred

Script bodies land with Phase 5.7 / Phase 6.8.
