#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-033-file-presign-register-roundtrip"
DATA="$OUT/data"
mkdir -p "$DATA"

{
  echo "E2E.P0.033 setup"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  echo "DATABASE_URL=${DATABASE_URL:+set}"
  echo "OBJECT_STORAGE_ENDPOINT=${OBJECT_STORAGE_ENDPOINT:+set}"
  echo "OBJECT_STORAGE_BUCKET=${OBJECT_STORAGE_BUCKET:+set}"
  echo "OBJECT_STORAGE_ACCESS_KEY=${OBJECT_STORAGE_ACCESS_KEY:+set}"
  echo "OBJECT_STORAGE_SECRET_KEY=${OBJECT_STORAGE_SECRET_KEY:+set}"
  cp "$ROOT/test/scenarios/e2e/p0-033-file-presign-register-roundtrip/data/expected-outcome.md" "$OUT/expected-outcome.md"
  dd if=/dev/zero of="$DATA/resume-small.pdf" bs=1024 count=1024 status=none
  dd if=/dev/zero of="$DATA/resume-boundary.bin" bs=1024 count=5120 status=none
  dd if=/dev/zero of="$DATA/resume-oversize.bin" bs=1024 count=11264 status=none
  wc -c "$DATA"/resume-small.pdf "$DATA"/resume-boundary.bin "$DATA"/resume-oversize.bin
} | tee "$OUT/setup.log"
