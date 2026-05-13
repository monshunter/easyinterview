#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-035-resume-parse-async-job-lifecycle"
mkdir -p "$OUT"

{
  echo "E2E.P0.035 setup"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  cp "$ROOT/test/scenarios/e2e/p0-035-resume-parse-async-job-lifecycle/data/expected-outcome.md" "$OUT/expected-outcome.md"
  echo "database_url=${DATABASE_URL:-postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable}"
} | tee "$OUT/setup.log"
