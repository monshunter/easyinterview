# Expected Outcome — E2E.P0.092

- `POST /experience-cards` (IK present, valid body) → `201 Created` + `ExperienceCard` JSON
- DB row `source_type='manual'` (regardless of body), `confidence='medium'`, `profile_id` resolves from user A's candidate_profile
- `POST /experience-cards` with same IK + same body → idempotency middleware replay; `201` + same `id`
- `POST /experience-cards` missing IK header → `422 Unprocessable Entity` + `error.code='VALIDATION_FAILED'`
- `PATCH /experience-cards/{A.id}` from user B → `404 Not Found` + `error.code='RESOURCE_NOT_FOUND'`
- `CountExperienceCardsBySource(userA)` → `{ manual:>=1, resume_parse:0, practice_report:0, debrief:0 }`
- Cross-owner internal `GetCandidateProfileForUser(userC)` returns nil (D-13 no-seed) BEFORE userC hits `GET /profiles/me`; subsequent call after userC seed returns non-nil `*CandidateProfile`
- `git grep -E 'mistake|growth|drill|experiences|star' backend/internal/profile/` → 0 hits
- HTTP responses byte-shape consistent with generated `CandidateProfile` / `ExperienceCard` / `PaginatedExperienceCard` DTOs (fixture default scenarios remain unchanged after B2 D-24 additive)
