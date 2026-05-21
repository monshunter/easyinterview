# Expected Outcome — E2E.P0.083

- `DeleteCandidateProfileForUser(userA)` removes ALL of user A's
  experience_cards (in that order, before the profile row) and then the
  candidate_profiles row; audit_events records exactly one row with
  `action='profile.privacy_delete'` for user A.
- audit_events metadata payload contains ONLY `experienceCardCount`,
  `deletedAt`, and optional `jobId`; verify.sh greps for raw card content
  strings (`Drove design-system`, `Reduced UI defects`, `Acme`, `RFC + 6-week rollout`)
  and fails if present.
- After delete:
  - `candidate_profiles where user_id = userA` count = 0
  - `experience_cards where user_id = userA` count = 0
  - subsequent `GET /profiles/me` for user A re-seeds (D-1) — verified inside
    `TestProfileHTTPScenario`
  - `CountExperienceCardsBySource(userA)` returns all-zero map
- Other users (B / C) and their rows are unaffected.
- `git grep -E 'mistake|growth|drill|experiences|star' backend/internal/profile/` → 0 hits.
