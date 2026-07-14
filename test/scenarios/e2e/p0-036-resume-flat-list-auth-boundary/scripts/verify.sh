#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-036-resume-flat-list-auth-boundary"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
grep -Fq "src/app/scenarios/p0-036-resume-flat-list-auth-boundary.test.tsx" "$LOG_FILE"
grep -Eq 'Tests +5 passed \(5\)' "$LOG_FILE"
grep -Eq 'Test Files +1 passed \(1\)' "$LOG_FILE"
grep -Fq 'E2E.P0.036 summary-only list/detail transport PASS summaryFields=9 listResumes=1 getResumeBeforeOpen=0 getResumeAfterOpen=1' "$LOG_FILE"
grep -Fq 'E2E.P0.036 list rejection retry transport PASS initialRejected=1 retrySucceeded=2' "$LOG_FILE"

for forbidden in 'FAIL' 'SKIP' 'no tests to run' 'not wrapped in act'; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then
    echo "forbidden no-op or React warning leaked into scenario evidence: $forbidden" >&2
    exit 1
  fi
done

# Negative: scenario evidence must not surface out-of-scope route testid leakage.
fallback_marker="D2""-D6"
for forbidden in \
  'route-welcome' \
  'route-mistakes' \
  'route-drill' \
  'route-followup' \
  'route-onboarding' \
  'route-experiences' \
  'route-star' \
  'route-voice' \
  "$fallback_marker"; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then
    echo "forbidden out-of-scope entry leaked into scenario evidence: $forbidden" >&2
    exit 1
  fi
done
