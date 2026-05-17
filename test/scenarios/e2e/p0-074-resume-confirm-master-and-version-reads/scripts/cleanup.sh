#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-074-resume-confirm-master-and-version-reads"
mkdir -p "$OUT"

{
  echo "E2E.P0.074 cleanup"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  echo "No scenario-owned long-lived resources remain; Go tests clean their own database rows."
} | tee "$OUT/cleanup.log"
