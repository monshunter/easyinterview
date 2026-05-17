#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-078-resume-tailor-failure-and-retry"
mkdir -p "$OUT"
{
  echo "E2E.P0.078 cleanup"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  echo "logs preserved under $OUT"
} | tee "$OUT/cleanup.log"
