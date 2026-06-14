# E2E.P0.076 Expected Outcome

## API Outcomes

- `POST /api/v1/resumes/{resumeId}:duplicate` returns 201 and a new flat
  resume.
- Duplicate preserves source snapshot fields and applies any allowed
  `structuredProfile` overlay while replacing server-owned provenance.
- Same idempotency key and fingerprint replays the first response without a second service call.
- Invalid input returns `422 VALIDATION_FAILED`.
- Cross-user source resume returns 404 without exposing record existence.
- Failed persistence rolls back without orphan rows.

## Evidence Outcomes

- `method=cmd-api-http` appears in verify output.
- Fixture parity covers `duplicateResume` `default`, `idempotency-replay`,
  and `validation-error-422`.
- Store/service tests cover source snapshot copy, profile overlay, cross-user
  source 404, and rollback no-orphan assertion.
- Trigger log contains no skipped or no-op focused gates.
- Retired vocabulary and privacy negative greps return zero hits for backend resume implementation and scenario evidence.
