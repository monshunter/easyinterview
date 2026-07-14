#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-037-resume-detail-preview-readonly"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
grep -Fq "src/app/scenarios/p0-037-resume-detail-preview-readonly.test.tsx" "$LOG_FILE"
grep -Eq 'Tests +8 passed \(8\)' "$LOG_FILE"
grep -Eq 'Test Files +1 passed \(1\)' "$LOG_FILE"
grep -Fq 'E2E.P0.037 ready detail transport PASS initial=1 maxInFlight=1' "$LOG_FILE"
grep -Fq 'E2E.P0.037 detail rejection retry transport PASS initialRejected=1 retrySucceeded=2 maxInFlight=1' "$LOG_FILE"
grep -Fq 'E2E.P0.037 pending serial poll transport PASS initial=1 poll=2 maxInFlight=1' "$LOG_FILE"

if grep -Fq 'not wrapped in act' "$LOG_FILE"; then
  echo "unwrapped React update leaked into scenario evidence" >&2
  exit 1
fi

for forbidden in 'FAIL' 'SKIP' 'no tests to run'; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then
    echo "forbidden no-op marker leaked into scenario evidence: $forbidden" >&2
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
  'TARGET_JOB_NOT_FOUND' \
  "$fallback_marker"; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then
    echo "forbidden out-of-scope entry leaked into scenario evidence: $forbidden" >&2
    exit 1
  fi
done
