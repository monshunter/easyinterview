# Expected Outcome — E2E.P0.081

## HTTP responses

- A1 `GET /profiles/me` → `200 OK` + CandidateProfile JSON `{ headline:null, yearsOfExperience:null, currentRole:null, preferredPracticeLanguage:"en", uiLanguage:"zh-CN", region:"CN-SH" }`
- A2 `GET /profiles/me` → `200 OK` + same body
- A3 `PATCH /profiles/me { headline, yearsOfExperience }` → `200 OK` + patched profile
- A5 `PATCH /profiles/me { yearsOfExperience: -1 }` → `422 Unprocessable Entity` + `error.code = VALIDATION_FAILED`
- B1 / C1 `GET /profiles/me` → `200 OK` + per-user seed profile

## DB invariants

- `candidate_profiles`: 1 row per user A / B / C after their first call; profile_version monotonically increments only on successful patch (A3 → 2)
- A5 rejection does NOT bump profile_version
- cross-user isolation: A row != B row != C row

## Evidence redaction

- `audit_events.metadata` 不含 headline / currentRole / yearsOfExperience 等 PII
- backend log / outbox / URL / cookie 不携带 raw profile field 值
- `grep mistake|growth|drill|experiences|star backend/internal/profile/` 0 hit
