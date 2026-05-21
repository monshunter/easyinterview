# Seed Input — E2E.P0.093

Reuses user / settings fixtures from E2E.P0.091 / E2E.P0.092 (users A / B / C).
The cmd/api scenario test seeds user A with a candidate_profile + 1 experience
card immediately before invoking `DeleteCandidateProfileForUser` so the
delete-order assertion is meaningful.

`TestPrivacyDeleteOrderAndAudit` (service unit test) covers fake-store-driven
chain order + audit tombstone semantics in isolation.
