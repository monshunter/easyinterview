#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
SCENARIO="$ROOT/test/scenarios/e2e/p0-074-resume-confirm-master-and-version-reads"
OUT="$ROOT/.test-output/e2e/p0-074-resume-confirm-master-and-version-reads"
mkdir -p "$OUT"

{
  echo "E2E.P0.074 setup"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  cp "$SCENARIO/data/seed-input.md" "$OUT/seed-input.md"
  cp "$SCENARIO/data/expected-outcome.md" "$OUT/expected-outcome.md"
  echo "scenario_dir=$SCENARIO"
  echo "output_dir=$OUT"
} | tee "$OUT/setup.log"
