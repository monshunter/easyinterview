#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-033-file-presign-register-roundtrip"
mkdir -p "$OUT"

{
  echo "E2E.P0.033 verify"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  test -s "$OUT/trigger.log"
  if grep -E -- '--- SKIP:|DATABASE_URL is not set|OBJECT_STORAGE_[A-Z_]+ is not set|skipping (file_objects integration test|MinIO smoke)' "$OUT/trigger.log"; then
    echo "ERROR: live integration skip detected; E2E.P0.033 cannot be marked PASS without live DB and MinIO evidence"
    exit 1
  fi
  if grep -E -- 'no tests to run|\\[no tests to run\\]' "$OUT/trigger.log"; then
    echo "ERROR: focused scenario gate matched no tests"
    exit 1
  fi
  grep -q 'TestCreateUploadPresign' "$OUT/trigger.log"
  grep -q 'TestCreateUploadPresignCreatesPendingFileObjectAndPresignsObject' "$OUT/trigger.log"
  grep -q 'TestRepositoryRegisterUploadedChecksObjectWhileRowLocked' "$OUT/trigger.log"
  grep -q 'TestBuildAPIHandlerMountsUploadPresignBehindSessionMiddleware' "$OUT/trigger.log"
  grep -q 'TestUploadPresignRegisterPrivacyDeleteLiveRoundtrip' "$OUT/trigger.log"
  grep -q 'TestDeleteFileObjectsForUser' "$OUT/trigger.log"
  grep -q 'TestInsertAuditTombstoneIntegrationDoesNotPersistObjectKey' "$OUT/trigger.log"
  cd "$ROOT/backend"
  go test ./internal/upload/handler -run TestCreateUploadPresignFixtureParity -count=1
  cd "$ROOT"
  if rg -n 'registered|deleted_pending' backend/internal/upload migrations; then
    exit 1
  fi
  if rg -n 'upload-route-frontend-signed|hardcode S3 SDK|frontend-signed' backend test/scenarios/e2e/p0-033-file-presign-register-roundtrip --glob '!**/verify.sh'; then
    exit 1
  fi
  if rg -n 'user-1/resume/file-1.pdf|object_key' "$OUT"; then
    exit 1
  fi
  echo "fixture byte diff: covered by TestCreateUploadPresignFixtureParity"
  echo "DB state machine: covered by store transition tests"
  echo "live HTTP presign -> MinIO PUT -> RegisterFileObject -> DELETE /api/v1/me: covered by TestUploadPresignRegisterPrivacyDeleteLiveRoundtrip"
  echo "object key before/after: covered by live roundtrip and DeleteFileObjectsForUser object-first unit test"
  echo "privacy tombstone: covered by integration-tag audit tombstone test"
} | tee "$OUT/verify.log"
