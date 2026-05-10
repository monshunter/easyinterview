#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-023-jd-match-search-and-saved"
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
)
for spec in "${required_specs[@]}"; do
  if ! grep -Fq "$spec" "$LOG_FILE"; then
    echo "missing required spec in trigger log: $spec" >&2
    exit 1
  fi
done

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
  'setInterval.*step'
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
