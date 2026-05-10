#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-022-jd-match-recommended-and-confirm"
mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT"
  pnpm --filter @easyinterview/frontend exec vitest run \
    src/app/screens/jd_match/RecommendedTab.test.tsx \
    src/app/screens/jd_match/JobMatchCard.test.tsx \
    src/app/screens/jd_match/JDDetail.test.tsx \
    src/app/screens/jd_match/RecommendedToggleWatchlist.test.tsx \
    src/app/screens/jd_match/RecommendedDismiss.test.tsx \
    src/app/screens/jd_match/RecommendedConfirmInterview.test.tsx \
    src/app/screens/jd_match/RecommendedOpenSource.test.tsx \
    src/app/screens/jd_match/RecommendedPrivacy.test.tsx \
    src/app/screens/jd_match/JDMatchAuthGate.test.tsx \
    src/app/screens/jd_match/JDMatchAutoResume.test.tsx \
    src/app/screens/jd_match/JDMatchDetailFetch.test.tsx \
    src/app/screens/jd_match/useJobRecommendation.test.tsx \
    src/app/screens/jd_match/useJobMatchRecommendations.test.tsx \
    src/app/screens/jd_match/useToggleWatchlist.test.tsx \
    src/app/screens/jd_match/useDismissRecommendation.test.tsx
) | tee "$OUTPUT_DIR/trigger.log"
