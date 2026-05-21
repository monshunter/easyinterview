# E2E.P0.094 jd-match-profile-and-recommendations-list

> **Status**: scaffold (live DB + cmd/api drainer registration required)
> **Owner plan**: [backend-jobs-recommendations/001](../../../docs/spec/backend-jobs-recommendations/plans/001-jd-match-real-backend-baseline/plan.md)
> **Spec acceptance**: C-1, C-2, C-3, C-4, C-5, C-13, C-14, C-17, C-19

## Scope

Primary path covering:

- `GET /api/v1/jd-match/profile` aggregating 7 cross-owner internal APIs
  (`backend-auth.GetUserIdentityForUser`,
  `backend-profile.GetCandidateProfileForUser`,
  `backend-profile.CountExperienceCardsBySource`, 4 counter packages).
- `GET /api/v1/jd-match/agent-status` returning lazy idle baseline.
- `GET /api/v1/jd-match/recommendations` with cursor pagination + score sort.
- `GET /api/v1/jd-match/recommendations/{jobMatchId}` detail projection.
- `POST /api/v1/jd-match/recommendations/{jobMatchId}/dismiss` with IK.

Verifies spec D-18 sparse baseline (`avatarUrl` / `locationText` /
`compensationText` null, `skills` empty array), D-19 structural parity,
cross-user 404, IK replay, redaction of `freeNote` from log /
audit / outbox.

## Live-environment prerequisite

Scripts under `scripts/` require:

- A2 dev stack Postgres (`DATABASE_URL` reachable).
- cmd/api running with the jd_match_agent_scan drainer registered and
  the AIClient bound to a real or stubbed provider.

When live env is absent the trigger script exits with a non-zero
status (no silent skip per plan §3 substitute gate rules).

## Scripts

- `scripts/setup.sh` — boot the dev stack + seed user A + populate
  candidate_profile, 3 resumes, 5 target_jobs, 8 practice_sessions,
  2 debriefs, 25 jd_match_recommendations.
- `scripts/trigger.sh` — execute the 7-cross-owner trace + the
  cursor / dismiss / cross-user-404 calls; run focused go test
  `TestJDMatchHTTPScenario` against cmd/api.
- `scripts/verify.sh` — assert structural-parity body, fixture
  parity headers, freeNote / raw email negative grep.
- `scripts/cleanup.sh` — wipe user A from the 5 jd_match tables.

## Deferred

Implementation of these scripts lands together with cmd/api drainer
registration (Phase 5.7) and BDD-Gate verification (Phase 6.5).
