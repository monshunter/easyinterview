#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-033-file-presign-register-roundtrip"
mkdir -p "$OUT"

{
  echo "E2E.P0.033 trigger"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  missing=()
  for key in DATABASE_URL OBJECT_STORAGE_ENDPOINT OBJECT_STORAGE_BUCKET OBJECT_STORAGE_ACCESS_KEY OBJECT_STORAGE_SECRET_KEY; do
    if [[ -z "${!key:-}" ]]; then
      missing+=("$key")
    fi
  done
  if ((${#missing[@]} > 0)); then
    echo "ERROR: E2E.P0.033 requires live database and object storage env; missing: ${missing[*]}"
    exit 2
  fi
  cd "$ROOT/backend"
  go test ./internal/upload/handler -run 'TestCreateUploadPresign' -count=1 -v
  go test ./internal/upload/service -run 'Test(CreateUploadPresign|RegisterFileObject|DeleteFileObjectsForUser)' -count=1 -v
  go test ./internal/upload/store -run 'TestRepository(Create|Mark|Lock|RegisterUploaded|HardDelete|DeleteFileObjectsForUser|InsertAuditTombstone)' -count=1 -v
  go test ./internal/upload/objectstore -count=1 -v
  go test ./cmd/api -run TestBuildAPIHandlerMountsUploadPresignBehindSessionMiddleware -count=1 -v
  go test ./internal/upload/store -tags=integration -run 'TestInsertAuditTombstoneIntegrationDoesNotPersistObjectKey|TestFileObjectsIntegrationDatabaseAvailable' -count=1 -v
  go test ./internal/upload/objectstore -tags=integration -run TestMinIO -count=1 -v
} | tee "$OUT/trigger.log"
