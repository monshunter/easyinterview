#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"; LOG="$ROOT/.test-output/e2e/p0-072-practice-derived-source-isolation-privacy/trigger.log"
grep -Fq -- '=== RUN   TestCreateDerivedPracticePlanRejectsCopiedServerFields' "$LOG"
grep -Fq -- '--- PASS: TestCreateDerivedPracticePlanRejectsCopiedServerFields' "$LOG"
grep -Fq -- '=== RUN   TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy' "$LOG"
grep -Fq -- '--- PASS: TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy' "$LOG"
for marker in REPORT_DERIVED_ISOLATION_PASS REPORT_DERIVED_PRIVACY_PASS REPORT_DERIVED_POSTGRES_PASS; do grep -Fq -- "$marker" "$LOG"; done
grep -Eq -- '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/practice' "$LOG"
grep -Eq -- '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/store/practice' "$LOG"
! grep -Eq -- '--- FAIL:|^FAIL($|[[:space:]])|no tests to run|\[no tests to run\]|DATABASE_URL is required' "$LOG"
OUT="$ROOT/.test-output/e2e/p0-072-practice-derived-source-isolation-privacy"
LEGACY_HITS="$OUT/legacy-focus-positive-hits.txt"
LEGACY_PATTERN='focusCompetencyCodes|focus_competency_codes|retryFocusCompetencyCodes|retry_focus_competency_codes|retryFocusTurnIds|retry_focus_turn_ids|retry_round|\{\{semantic_focus\}\}'
if rg -n "$LEGACY_PATTERN" \
  "$ROOT/backend/internal/practice" \
  "$ROOT/backend/internal/store/practice" \
  "$ROOT/backend/internal/api/practice" \
  "$ROOT/backend/cmd/api" \
  "$ROOT/openapi/openapi.yaml" \
  "$ROOT/backend/internal/api/generated" \
  "$ROOT/frontend/src/api/generated" \
  "$ROOT/openapi/fixtures" \
  "$ROOT/config/prompts/practice.session.chat/v0.2.0.md" \
  "$ROOT/test/scenarios/e2e/p0-070-practice-derived-plan-create-read-replay" \
  "$ROOT/test/scenarios/e2e/p0-072-practice-derived-source-isolation-privacy" \
  --glob '!**/*_test.go' --glob '!**/scripts/verify.sh' --glob '!**/PROTOTYPE_MAPPING.md' > "$LEGACY_HITS"; then
  cat "$LEGACY_HITS" >&2
  echo 'legacy focus identifier remains in a positive final-active consumer' >&2
  exit 1
fi
rm -f "$LEGACY_HITS"
echo 'REPORT_DERIVED_LEGACY_IDENTIFIER_NEGATIVE_PASS'
echo 'E2E.P0.072 PASS'
