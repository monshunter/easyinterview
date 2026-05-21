# E2E.P0.095 jd-match-watchlist-and-saved-search-lifecycle

> **Status**: scaffold (live DB + cmd/api drainer registration required)
> **Owner plan**: [backend-jobs-recommendations/001](../../../docs/spec/backend-jobs-recommendations/plans/001-jd-match-real-backend-baseline/plan.md)
> **Spec acceptance**: C-6, C-7, C-8, C-10, C-13, C-14, C-17

## Scope

Watchlist + saved-search lifecycle covering:

- `POST /api/v1/jd-match/watchlist` (IK replay + UNIQUE conflict
  returns first item per spec C-6).
- `GET /api/v1/jd-match/watchlist` tone derivation
  (score 92 / 78 / 45 -> ok / warn / muted per Q-4).
- `DELETE /api/v1/jd-match/watchlist/{jobMatchId}` (IK + 204).
- `GET /api/v1/jd-match/saved-searches` + `POST` create with
  `label / query` required, redacted from log / audit / outbox.

## Live-environment prerequisite

Same as E2E.P0.094 README. Scripts exit non-zero in the absence of
live cmd/api + Postgres.

## Deferred

Script bodies land together with cmd/api drainer registration
(Phase 5.7) and BDD-Gate verification (Phase 6.6).
