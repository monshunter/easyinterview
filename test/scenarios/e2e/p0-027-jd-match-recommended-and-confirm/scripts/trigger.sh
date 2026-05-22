#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-027-jd-match-recommended-and-confirm"
mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT"
  REAL_API_MODE="${VITE_EI_API_MODE:-real}"
  REAL_API_BASE_URL="${VITE_EI_API_BASE_URL:-http://localhost:8080/api/v1}"
  printf 'VITE_EI_API_MODE=%s\nVITE_EI_API_BASE_URL=%s\n' "$REAL_API_MODE" "$REAL_API_BASE_URL"
  VITE_EI_API_MODE="$REAL_API_MODE" VITE_EI_API_BASE_URL="$REAL_API_BASE_URL" pnpm --filter @easyinterview/frontend exec vitest run \
    src/api/jdMatch.realApiMode.test.ts
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
