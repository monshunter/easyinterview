#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-035-resume-parse-async-job-lifecycle"
mkdir -p "$OUT"

{
  echo "E2E.P0.035 cleanup"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  echo "No external resources created by this scenario wrapper; Go tests clean their own rows."
} | tee "$OUT/cleanup.log"
