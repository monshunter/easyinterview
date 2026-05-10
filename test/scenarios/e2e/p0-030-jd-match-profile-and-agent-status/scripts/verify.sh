#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-030-jd-match-profile-and-agent-status"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
grep -Eq 'Test Files +[0-9]+ passed' "$LOG_FILE"
grep -Eq 'Tests +[0-9]+ passed' "$LOG_FILE"

required_specs=(
  'JDMatchScreen.test.tsx'
  'JDMatchScreen.dataDriven.test.tsx'
  'JDMatchScreen.fetchBehavior.test.tsx'
  'JDMatchScreen.placeholderRemoved.test.tsx'
  'useJobMatchProfile.test.tsx'
  'useAgentScanStatus.test.tsx'
  'JDMatchAuthGate.test.tsx'
  'JDMatchAutoResume.test.tsx'
)
for spec in "${required_specs[@]}"; do
  if ! grep -Fq "$spec" "$LOG_FILE"; then
    echo "missing required spec in trigger log: $spec" >&2
    exit 1
  fi
done

grep -Fq 'Responsive geometry matches jd_match layout contracts' "$LOG_FILE"
grep -Fq 'dark mode and customAccent visibly affect jd_match computed colors' "$LOG_FILE"
grep -Eq '[0-9]+ passed' "$LOG_FILE"
PIXEL_SPEC="$REPO_ROOT/frontend/tests/pixel-parity/jd_match.spec.ts"
test -s "$PIXEL_SPEC"
grep -Fq 'jdmatch-market-signals-inner' "$PIXEL_SPEC"
grep -Fq 'toHaveScreenshot' "$PIXEL_SPEC"

# Source-level negative gate: D-10 forbids polling/streaming for AGENT
# status. Use git ls-files + filter for portable BSD / GNU / ugrep behaviour.
SCAN_FILES=$(cd "$REPO_ROOT" && git ls-files \
  'frontend/src/app/screens/jd_match/*.ts' \
  'frontend/src/app/screens/jd_match/*.tsx' \
  | grep -Ev '\.test\.(ts|tsx)$|\.spec\.(ts|tsx)$' || true)
forbidden_polling_patterns=(
  'setInterval.*getAgentScanStatus'
  'EventSource.*agent-status'
  'new WebSocket'
)
for pattern in "${forbidden_polling_patterns[@]}"; do
  if [[ -n "$SCAN_FILES" ]]; then
    while IFS= read -r f; do
      [[ -z "$f" ]] && continue
      if grep -Eq "$pattern" "$REPO_ROOT/$f"; then
        echo "forbidden polling/streaming leaked in $f: $pattern" >&2
        exit 1
      fi
    done <<<"$SCAN_FILES"
  fi
done
