#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-078-resume-tailor-failure-and-retry"
mkdir -p "$OUT"
cp "$ROOT/test/scenarios/e2e/p0-078-resume-tailor-failure-and-retry/data/seed-input.md" "$OUT/seed-input.md"
cp "$ROOT/test/scenarios/e2e/p0-078-resume-tailor-failure-and-retry/data/expected-outcome.md" "$OUT/expected-outcome.md"
{
  echo "E2E.P0.078 setup"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  echo "output=$OUT"
} | tee "$OUT/setup.log"
