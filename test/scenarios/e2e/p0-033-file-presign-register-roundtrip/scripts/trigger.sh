#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-033-file-presign-register-roundtrip"
mkdir -p "$OUT"

{
  echo "E2E.P0.033 trigger"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  cd "$ROOT/backend"
  go test ./internal/upload/handler -run 'TestCreateUploadPresign' -count=1 -v
  go test ./internal/upload/service -run 'Test(CreateUploadPresign|RegisterFileObject|DeleteFileObjectsForUser)' -count=1 -v
  go test ./internal/upload/store -run 'TestRepository(Create|Mark|Lock|RegisterUploaded|HardDelete|DeleteFileObjectsForUser|InsertAuditTombstone)' -count=1 -v
  go test ./internal/upload/objectstore -count=1 -v
  go test ./cmd/api -run TestBuildAPIHandlerMountsUploadPresignBehindSessionMiddleware -count=1 -v
  go test ./internal/upload/store -tags=integration -run 'TestInsertAuditTombstoneIntegrationDoesNotPersistObjectKey|TestFileObjectsIntegrationDatabaseAvailable' -count=1 -v
  go test ./internal/upload/objectstore -tags=integration -run TestMinIO -count=1 -v
} | tee "$OUT/trigger.log"
