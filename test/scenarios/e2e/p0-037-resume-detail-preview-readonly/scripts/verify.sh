#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-037-resume-detail-preview-readonly"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
grep -Fq "src/app/scenarios/p0-037-resume-detail-preview-readonly.test.tsx" "$LOG_FILE"
grep -Eq 'Tests +5 passed \(5\)' "$LOG_FILE"
grep -Eq 'Test Files +1 passed \(1\)' "$LOG_FILE"

# Negative: scenario evidence must not surface retired-route testid leakage.
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
  'D2-D6'; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then
    echo "forbidden legacy entry leaked into scenario evidence: $forbidden" >&2
    exit 1
  fi
done
