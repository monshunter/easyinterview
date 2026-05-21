#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-091-candidate-profile-seed-and-patch"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
grep -Eq 'ok[[:space:]]+github.com/monshunter/easyinterview/backend/cmd/api' "$LOG_FILE"
grep -Fq 'PASS: TestProfileHTTPScenario' "$LOG_FILE"

if grep -Eqi 'SKIP|no-op' "$LOG_FILE"; then
  echo "scenario must not skip; live env evidence missing" >&2
  exit 1
fi

# Negative: legacy module shorthands MUST NOT appear in profile module source.
if git -C "$REPO_ROOT" grep -nE 'mistake|growth|drill|experiences|star' backend/internal/profile/ > "$OUTPUT_DIR/negative-grep.log"; then
  echo "legacy module shorthand leaked into backend/internal/profile/" >&2
  cat "$OUTPUT_DIR/negative-grep.log" >&2
  exit 1
fi

echo "method=cmd-api-http" >> "$OUTPUT_DIR/setup.env"
