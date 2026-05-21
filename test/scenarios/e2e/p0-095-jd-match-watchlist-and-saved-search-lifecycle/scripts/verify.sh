#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-095-jd-match-watchlist-and-saved-search-lifecycle"
LOG="$OUTPUT_DIR/trigger.log"
if [ ! -s "$LOG" ]; then
  echo "verify: missing trigger.log" >&2
  exit 1
fi
if grep -Eq -- "--- SKIP:|skipping live|\\[no tests to run\\]|no tests to run" "$LOG"; then
  echo "verify: scenario log contains skip/no-test marker" >&2
  exit 1
fi
if grep -Eq -- "--- FAIL:|^FAIL$|^FAIL[[:space:]]" "$LOG"; then
  echo "verify: scenario log contains fail marker" >&2
  exit 1
fi
for marker in \
  "--- PASS: TestJDMatchHTTPScenario" \
  "--- PASS: TestJDMatchFixtureParity/watchlist_list" \
  "--- PASS: TestJDMatchFixtureParity/watchlist_add" \
  "--- PASS: TestJDMatchFixtureParity/watchlist_remove" \
  "--- PASS: TestJDMatchFixtureParity/saved_searches_list" \
  "--- PASS: TestJDMatchFixtureParity/saved_search_create"; do
  if ! grep -q -- "$marker" "$LOG"; then
    echo "verify: missing pass marker $marker" >&2
    exit 1
  fi
done
if ! grep -Eq -- "^ok[[:space:]]+github.com/monshunter/easyinterview/backend/cmd/api[[:space:]]" "$LOG"; then
  echo "verify: missing package-level ok marker" >&2
  exit 1
fi
if grep -Eq "alice@example\.com|bob@example\.com|jdmatch-fixture@example\.com" "$LOG"; then
  echo "verify: raw email leaked into trigger log" >&2
  exit 1
fi
echo "verify: ok"
