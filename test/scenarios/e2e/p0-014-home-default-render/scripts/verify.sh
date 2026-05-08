#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-014-home-default-render"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
grep -Fq "HomeScreen.test.tsx" "$LOG_FILE"
grep -Fq "HomeRecentMocks.test.tsx" "$LOG_FILE"
grep -Fq "MockInterviewCard.test.tsx" "$LOG_FILE"
grep -Eq 'Test Files +[0-9]+ passed' "$LOG_FILE"
grep -Eq 'Tests +[0-9]+ passed' "$LOG_FILE"
# Negative: legacy testids not in output
for forbidden in 'route-welcome' 'topbar-nav-mistakes' 'topbar-nav-growth' 'topbar-nav-drill' 'topbar-nav-voice' 'home-pasted-success' 'home-mocked-recent' 'jdmatch-card-' 'jdmatch-market-signal-'; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then
    echo "forbidden legacy entry leaked: $forbidden" >&2
    exit 1
  fi
done
