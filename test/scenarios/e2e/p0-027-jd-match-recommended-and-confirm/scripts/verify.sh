#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-027-jd-match-recommended-and-confirm"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
grep -Eq 'Test Files +[0-9]+ passed' "$LOG_FILE"
grep -Eq 'Tests +[0-9]+ passed' "$LOG_FILE"

required_specs=(
  'RecommendedTab.test.tsx'
  'JobMatchCard.test.tsx'
  'JDDetail.test.tsx'
  'RecommendedToggleWatchlist.test.tsx'
  'RecommendedDismiss.test.tsx'
  'RecommendedConfirmInterview.test.tsx'
  'RecommendedOpenSource.test.tsx'
  'RecommendedPrivacy.test.tsx'
  'JDMatchAuthGate.test.tsx'
  'JDMatchAutoResume.test.tsx'
  'JDMatchDetailFetch.test.tsx'
  'useJobRecommendation.test.tsx'
)
for spec in "${required_specs[@]}"; do
  if ! grep -Fq "$spec" "$LOG_FILE"; then
    echo "missing required spec in trigger log: $spec" >&2
    exit 1
  fi
done

HOOK_FILE="$REPO_ROOT/frontend/src/app/screens/jd_match/useJobRecommendation.ts"
SCREEN_FILE="$REPO_ROOT/frontend/src/app/screens/jd_match/JDMatchScreen.tsx"
DETAIL_TEST="$REPO_ROOT/frontend/src/app/screens/jd_match/JDMatchDetailFetch.test.tsx"
test -s "$HOOK_FILE"
test -s "$SCREEN_FILE"
test -s "$DETAIL_TEST"
grep -Fq 'getJobRecommendation(jobMatchId)' "$HOOK_FILE"
grep -Fq 'useJobRecommendation(selectedId)' "$SCREEN_FILE"
grep -Fq 'DETAIL_ONLY_TITLE' "$DETAIL_TEST"
grep -Fq 'toHaveBeenCalledWith(' "$DETAIL_TEST"
grep -Fq 'JDMatchAutoResume.test.tsx' "$LOG_FILE"

# Source-level negative gate: pendingAction must NEVER carry private fields.
# Use git ls-files + filter for portable BSD / GNU / ugrep behaviour.
SCAN_FILES=$(cd "$REPO_ROOT" && git ls-files \
  'frontend/src/app/screens/jd_match/*.ts' \
  'frontend/src/app/screens/jd_match/*.tsx' \
  | grep -Ev '\.test\.(ts|tsx)$|\.spec\.(ts|tsx)$' || true)
forbidden_pendingaction_fields=(
  'pendingAction.*query'
  'pendingAction.*freeNote'
  'pendingAction.*sourceUrl'
  'params:.*query'
  'params:.*freeNote'
  'params:.*sourceUrl'
)
for pattern in "${forbidden_pendingaction_fields[@]}"; do
  if [[ -n "$SCAN_FILES" ]]; then
    while IFS= read -r f; do
      [[ -z "$f" ]] && continue
      if grep -Eq "$pattern" "$REPO_ROOT/$f"; then
        echo "forbidden pendingAction field leaked in $f: $pattern" >&2
        exit 1
      fi
    done <<<"$SCAN_FILES"
  fi
done
