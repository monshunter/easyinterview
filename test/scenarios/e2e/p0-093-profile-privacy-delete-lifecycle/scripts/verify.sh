#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-083-profile-privacy-delete-lifecycle"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
grep -Eq 'ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/profile/service' "$LOG_FILE"
grep -Eq 'ok[[:space:]]+github.com/monshunter/easyinterview/backend/cmd/api' "$LOG_FILE"
grep -Fq 'PASS: TestPrivacyDeleteOrderAndAudit' "$LOG_FILE"
grep -Fq 'PASS: TestProfileHTTPScenario' "$LOG_FILE"

if grep -Eqi 'SKIP|no-op' "$LOG_FILE"; then
  echo "scenario must not skip; live env evidence missing" >&2
  exit 1
fi

# Privacy redline: trigger.log must NOT leak raw card content (the cmd/api
# scenario writes a real card before invoking privacy delete; if any sensitive
# field shows up in the log, audit/log pipeline is leaking PII).
for forbidden in \
  'Drove design-system migration' \
  'Reduced UI defects by 38%' \
  'RFC + 6-week rollout' ; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then
    echo "raw experience card content leaked into trigger log: $forbidden" >&2
    exit 1
  fi
done

if git -C "$REPO_ROOT" grep -nE 'mistake|growth|drill|experiences|star' backend/internal/profile/ > "$OUTPUT_DIR/negative-grep.log"; then
  echo "legacy module shorthand leaked into backend/internal/profile/" >&2
  cat "$OUTPUT_DIR/negative-grep.log" >&2
  exit 1
fi

echo "method=internal-api+cmd-api-http" >> "$OUTPUT_DIR/setup.env"
