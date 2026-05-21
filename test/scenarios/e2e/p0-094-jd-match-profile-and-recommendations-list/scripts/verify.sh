#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-094-jd-match-profile-and-recommendations-list"
LOG="$OUTPUT_DIR/trigger.log"
if [ ! -s "$LOG" ]; then
  echo "verify: missing trigger.log" >&2
  exit 1
fi
if ! grep -q "^--- PASS: TestJDMatchHTTPScenario" "$LOG"; then
  echo "verify: TestJDMatchHTTPScenario did not pass" >&2
  exit 1
fi
if grep -Eq "alice@example\\.com|bob@example\\.com" "$LOG"; then
  echo "verify: raw email leaked into trigger log" >&2
  exit 1
fi
echo "verify: ok"
