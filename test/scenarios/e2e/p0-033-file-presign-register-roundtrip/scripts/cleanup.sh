#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-033-file-presign-register-roundtrip"

{
  echo "E2E.P0.033 cleanup"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  rm -rf "$OUT/data"
  echo "temporary binary inputs removed; logs preserved"
} | tee "$OUT/cleanup.log"
