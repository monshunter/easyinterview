# Expected Outcome

- `createUploadPresign.default` returns `201` with `fileObjectId`, `uploadUrl`, `method`, `headers`, and `expiresAt`.
- Idempotency replay returns the original successful response without creating a second `file_objects` row.
- Unknown purpose returns `422 VALIDATION_FAILED` and points at `purpose`.
- Cross-user register cannot reveal or mutate another user's `fileObjectId`.
- Privacy deletion deletes object storage keys before hard deleting `file_objects` rows.
- `trigger.sh` includes a live `TestUploadPresignRegisterPrivacyDeleteLiveRoundtrip` gate that drives HTTP presign, signed PUT, internal register, `DELETE /api/v1/me`, and privacy drainer processing.
- Object delete failure returns a retryable error and leaves DB rows untouched.
- Audit tombstone metadata contains `fileObjectId`, `purpose`, and `deletedAt`; it must not contain `objectKey`.
- No `registered` or `deleted_pending` upload status is introduced.
