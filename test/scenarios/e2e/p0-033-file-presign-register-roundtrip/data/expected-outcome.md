# Expected Outcome

- `createUploadPresign.default` returns `201` with `fileObjectId`, `uploadUrl`, `method`, `headers`, and `expiresAt`.
- Idempotency replay returns the original successful response without creating a second `file_objects` row.
- Unknown purpose returns `422 VALIDATION_FAILED` and points at `purpose`.
- The removed `target_job_attachment` purpose returns live HTTP `422 VALIDATION_FAILED` and creates zero `file_objects` rows.
- `privacy_export` accepts the 5 MB boundary and creates one `pending` file object without restoring any TargetJob upload path.
- Cross-user register cannot reveal or mutate another user's `fileObjectId`.
- Privacy data-erasure deletes object storage keys before hard deleting `file_objects` rows.
- `trigger.sh` includes a live `TestUploadPresignRegisterPrivacyDeleteLiveRoundtrip` gate that drives HTTP presign, signed PUT, internal register, `DELETE /api/v1/me`, and privacy runner kernel processing.
- Object delete failure returns a retryable error and leaves DB rows untouched.
- Audit tombstone metadata contains `fileObjectId`, `purpose`, and `deletedAt`; it must not contain `objectKey`.
- No `registered` or `deleted_pending` upload status is introduced.
