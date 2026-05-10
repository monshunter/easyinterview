#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-024-jd-match-watchlist-and-signals"
mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT"
  pnpm --filter @easyinterview/frontend exec vitest run \
    src/app/screens/jd_match/WatchlistTab.test.tsx \
    src/app/screens/jd_match/WatchlistChevron.test.tsx \
    src/app/screens/jd_match/WatchlistEmpty.test.tsx \
    src/app/screens/jd_match/MarketSignals.test.tsx \
    src/app/screens/jd_match/MarketSignalsPartial.test.tsx \
    src/app/screens/jd_match/WatchlistPrivacy.test.tsx \
    src/app/screens/jd_match/useWatchlist.test.tsx
) | tee "$OUTPUT_DIR/trigger.log"
