# E2E.P0.075 Expected Outcome

## API Outcomes

- `PATCH /api/v1/resumes/{resumeId}` returns 200 for editable `displayName`,
  `focusAngle`, `matchScore`, and `structuredProfile` updates.
- `structuredProfile` updates are persisted as flat resume state after server
  provenance is stripped.
- Client-supplied `structuredProfile.provenance` and top-level server-owned fields are rejected with `422 VALIDATION_FAILED`.
- Same idempotency key and same fingerprint replays the first 200 response without a second service call.
- Same idempotency key and different fingerprint returns `409 IDEMPOTENCY_KEY_MISMATCH`.
- Cross-user and missing resume updates return 404 without exposing record existence.

## Evidence Outcomes

- `method=cmd-api-http` appears in verify output.
- Fixture parity covers `updateResume` `default`, `idempotency-replay`, and
  `validation-error-422`.
- Store/service tests cover profile overwrite, server-owned provenance reset,
  cross-user 404, and rollback on missing rows.
- Trigger log contains no skipped or no-op focused gates.
- Retired vocabulary and privacy negative greps return zero hits for backend resume implementation and scenario evidence.
