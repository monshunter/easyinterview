#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-023-jd-match-search-and-saved"
mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT"
  pnpm --filter @easyinterview/frontend exec vitest run \
    src/app/screens/jd_match/SearchTab.test.tsx \
    src/app/screens/jd_match/SearchTabRun.test.tsx \
    src/app/screens/jd_match/SearchTabSavedSearches.test.tsx \
    src/app/screens/jd_match/SearchTabFilter.test.tsx \
    src/app/screens/jd_match/SearchTabFailure.test.tsx \
    src/app/screens/jd_match/SearchTabPrivacy.test.tsx \
    src/app/screens/jd_match/SearchTabAuthGate.test.tsx \
    src/app/screens/jd_match/useSearchJobs.test.tsx \
    src/app/screens/jd_match/useSavedSearches.test.tsx
) | tee "$OUTPUT_DIR/trigger.log"
