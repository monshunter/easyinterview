#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-028-jd-match-search-and-saved"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
grep -Eq 'Test Files +[0-9]+ passed' "$LOG_FILE"
grep -Eq 'Tests +[0-9]+ passed' "$LOG_FILE"

required_specs=(
  'SearchTab.test.tsx'
  'SearchTabRun.test.tsx'
  'SearchTabSavedSearches.test.tsx'
  'SearchTabFilter.test.tsx'
  'SearchTabFailure.test.tsx'
  'SearchTabPrivacy.test.tsx'
  'SearchTabAuthGate.test.tsx'
  'JDMatchAutoResume.test.tsx'
)
for spec in "${required_specs[@]}"; do
  if ! grep -Fq "$spec" "$LOG_FILE"; then
    echo "missing required spec in trigger log: $spec" >&2
    exit 1
  fi
done

PENDING_STATE_FILE="$REPO_ROOT/frontend/src/app/screens/jd_match/pendingJdMatchActionState.ts"
SCREEN_FILE="$REPO_ROOT/frontend/src/app/screens/jd_match/JDMatchScreen.tsx"
AUTO_RESUME_TEST="$REPO_ROOT/frontend/src/app/screens/jd_match/JDMatchAutoResume.test.tsx"
SEARCH_TAB_TEST="$REPO_ROOT/frontend/src/app/screens/jd_match/SearchTab.test.tsx"
PIXEL_TEST="$REPO_ROOT/frontend/tests/pixel-parity/jd_match.spec.ts"
test -s "$PENDING_STATE_FILE"
test -s "$SCREEN_FILE"
test -s "$AUTO_RESUME_TEST"
test -s "$SEARCH_TAB_TEST"
test -s "$PIXEL_TEST"
grep -Fq 'pending-jd-match-' "$PENDING_STATE_FILE"
grep -Fq 'pendingJdMatchActionId' "$SCREEN_FILE"
grep -Fq 'consumePendingJdMatchAction' "$SCREEN_FILE"
grep -Fq 'secret frontend remote' "$AUTO_RESUME_TEST"
grep -Fq 'not.toContain(secretQuery)' "$AUTO_RESUME_TEST"
grep -Fq 'NATURAL LANGUAGE SEARCH' "$SEARCH_TAB_TEST"
grep -Fq 'jdmatch-search-input-icon' "$SEARCH_TAB_TEST"
grep -Fq 'jdmatch-search-source-company' "$SEARCH_TAB_TEST"
grep -Fq 'jdmatch-search-source-company' "$PIXEL_TEST"
if grep -Eq 'params: *\{[^}]*query|params: *\{[^}]*label' "$SCREEN_FILE"; then
  echo 'Search pendingAction params must not carry query or label directly' >&2
  exit 1
fi

# Source-level negative gate: dynamic JD numbers and prototype-only step
# index advancement must not appear in the implementation files. Use git
# ls-files + filter for portable BSD / GNU / ugrep behaviour.
SCAN_SRC=$(cd "$REPO_ROOT" && git ls-files \
  'frontend/src/app/screens/jd_match/*.ts' \
  'frontend/src/app/screens/jd_match/*.tsx' \
  | grep -Ev '\.test\.(ts|tsx)$|\.spec\.(ts|tsx)$' || true)
forbidden_dynamic_data=(
  '"248"'
  '"87"'
  'unique postings'
  'setInterval\('
  'setTimeout\('
)
for pattern in "${forbidden_dynamic_data[@]}"; do
  if [[ -n "$SCAN_SRC" ]]; then
    while IFS= read -r f; do
      [[ -z "$f" ]] && continue
      if grep -Eq "$pattern" "$REPO_ROOT/$f"; then
        echo "forbidden dynamic-data / step-advance leaked in $f: $pattern" >&2
        exit 1
      fi
    done <<<"$SCAN_SRC"
  fi
done

# Negative gate on i18n locale source files: dynamic JD numbers must not
# appear in any frontend i18n value (search-tab step labels in particular).
SCAN_I18N=$(cd "$REPO_ROOT" && git ls-files \
  'frontend/src/app/i18n/locales/*.ts' \
  | grep -Ev '\.test\.ts$' || true)
i18n_forbidden=(
  '"248"'
  '"87"'
  'unique postings'
)
for pattern in "${i18n_forbidden[@]}"; do
  if [[ -n "$SCAN_I18N" ]]; then
    while IFS= read -r f; do
      [[ -z "$f" ]] && continue
      if grep -Eq "$pattern" "$REPO_ROOT/$f"; then
        echo "forbidden dynamic-data leaked in i18n locale source $f: $pattern" >&2
        exit 1
      fi
    done <<<"$SCAN_I18N"
  fi
done
