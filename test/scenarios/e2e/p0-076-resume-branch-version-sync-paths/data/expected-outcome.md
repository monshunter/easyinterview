# E2E.P0.076 Expected Outcome

## API Outcomes

- `POST /api/v1/resume-versions` with `seedStrategy=copy_master` returns 201 and a targeted `ResumeVersion`.
- Copy-master preserves parent profile content while replacing `structuredProfile.provenance` with server-owned branch provenance.
- `seedStrategy=blank` returns 201 and a targeted `ResumeVersion` with empty editable profile fields.
- Same idempotency key and fingerprint replays the first response without a second service call.
- Invalid seed strategy returns `422 VALIDATION_FAILED`.
- Cross-user parent version and foreign target job return 404 without exposing record existence.
- Synchronous strategies do not create `resume_tailor_runs` or `async_jobs`.

## Evidence Outcomes

- `method=cmd-api-http` appears in verify output.
- Fixture parity covers branch `default`, `copy-master-sync`, `blank-sync`, `idempotent-replay`, and `validation-error-422`.
- Live DB integration covers copy profile provenance reset, blank profile, cross-user parent 404, foreign target 404, and rollback no-orphan assertion.
- Trigger log contains no skipped or no-op focused gates.
- Retired vocabulary and privacy negative greps return zero hits for backend resume implementation and scenario evidence.
