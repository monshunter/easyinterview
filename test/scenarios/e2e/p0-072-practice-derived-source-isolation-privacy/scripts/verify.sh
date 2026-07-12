#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"; LOG="$ROOT/.test-output/e2e/p0-072-practice-derived-source-isolation-privacy/trigger.log"
grep -Fq -- '--- PASS: TestDerivedPracticePlanRequiresSourceReport' "$LOG"
echo 'E2E.P0.072 PASS'
