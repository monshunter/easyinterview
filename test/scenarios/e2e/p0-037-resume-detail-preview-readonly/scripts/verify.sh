#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-037-resume-detail-preview-readonly"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
grep -Fq "src/app/scenarios/p0-037-resume-detail-preview-readonly.test.tsx" "$LOG_FILE"
grep -Eq 'Tests +6 passed \(6\)' "$LOG_FILE"
grep -Eq 'Test Files +1 passed \(1\)' "$LOG_FILE"

if grep -Fq 'not wrapped in act' "$LOG_FILE"; then
  echo "unwrapped React update leaked into scenario evidence" >&2
  exit 1
fi

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
