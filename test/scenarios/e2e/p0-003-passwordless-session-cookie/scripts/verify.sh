#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-003-passwordless-session-cookie"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
grep -Eq 'ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/auth' "$LOG_FILE"

for forbidden in \
  'scenario-magic-token-1' \
  'scenario-magic-token-2' \
  'scenario-session-token-1' \
  'scenario-session-token-2' \
  'scenario.user+e2e-p0-003@example.test' \
  'scenario-pepper' \
  'scenario-session-secret' \
  'http://api.test/api/v1/auth/email/verify'; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then
    echo "forbidden scenario evidence leaked: $forbidden" >&2
    exit 1
  fi
done
