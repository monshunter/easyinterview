#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-034-resume-register-and-list"
mkdir -p "$OUT"

{
  echo "E2E.P0.034 setup"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  cp "$ROOT/test/scenarios/e2e/p0-034-resume-register-and-list/data/expected-outcome.md" "$OUT/expected-outcome.md"
  echo "database_url=${DATABASE_URL:-postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable}"
} | tee "$OUT/setup.log"
