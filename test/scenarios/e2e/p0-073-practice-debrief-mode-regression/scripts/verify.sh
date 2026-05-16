#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")/../../../../.."
LOG=".test-output/e2e/p0-073-practice-debrief-mode-regression/trigger.log"
test -s "$LOG"
grep -Fq "E2E.P0.073 RUNNER go test" "$LOG"
grep -Fq "=== RUN   TestE2EP0073PracticeDebriefAssistedStrictAndLegacyNegative" "$LOG"
grep -Fq -- "--- PASS: TestE2EP0073PracticeDebriefAssistedStrictAndLegacyNegative" "$LOG"
grep -Eq "^ok[[:space:]]+github.com/monshunter/easyinterview/backend/cmd/api([[:space:]]|$)" "$LOG"
! grep -Eq -- "--- FAIL:|^FAIL($|[[:space:]])|no tests to run|\\[no tests to run\\]" "$LOG"
if rg -n "PracticeModeDebrief|mode='debrief'|mode=debrief|legacy debrief replay value|startDebriefInterview" \
  backend/internal/practice backend/internal/api/practice backend/internal/store/practice \
  openapi frontend/src/api/generated backend/internal/api/generated \
  openapi/fixtures/PracticePlans openapi/fixtures/PracticeSessions \
  -g '!**/*_test.go' -g '!**/verify.sh'; then
  echo "E2E.P0.073: legacy debrief mode surface found" >&2
  exit 1
fi
echo "E2E.P0.073 PASS"
