#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-017-jd-match-placeholder"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
grep -Fq "JDMatchPlaceholder.test.tsx" "$LOG_FILE"
grep -Eq 'Test Files +[0-9]+ passed' "$LOG_FILE"
grep -Eq 'Tests +[0-9]+ passed' "$LOG_FILE"
# Negative: old prototype jd_match business testids must not appear
for forbidden in \
  'jdmatch-card-' \
  'jdmatch-saved-search-' \
  'jdmatch-watchlist-' \
  'jdmatch-market-signal-' \
  'jdmatch-search-bar' \
  'jdmatch-search-results' \
  'jdmatch-jd-detail-' \
  'jdmatch-agent-status' \
  'route-welcome' \
  'topbar-nav-mistakes' \
  'topbar-nav-growth' \
  'topbar-nav-drill' \
  'topbar-nav-voice'; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then
    echo "forbidden legacy entry leaked: $forbidden" >&2
    exit 1
  fi
done

# Source-level negative gate: hidden old business anchors must not remain in
# the placeholder implementation even if the Vitest log does not print them.
if grep -R --exclude='*.test.ts' --exclude='*.test.tsx' -E 'jdmatch-card-|jdmatch-saved-search-|jdmatch-watchlist-|jdmatch-market-signal-|jdmatch-search-bar|jdmatch-search-results|jdmatch-jd-detail-|jdmatch-agent-status' \
  "$REPO_ROOT/frontend/src/app/screens/jd_match"; then
  echo "forbidden legacy jd_match anchor leaked in frontend source" >&2
  exit 1
fi
