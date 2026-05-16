#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")/../../../../.."
LOG=".test-output/e2e/p0-072-practice-derived-source-isolation-privacy/trigger.log"
test -s "$LOG"
grep -Fq "E2E.P0.072 RUNNER go test" "$LOG"
grep -Fq "=== RUN   TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy" "$LOG"
grep -Fq -- "--- PASS: TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy" "$LOG"
grep -Eq "^ok[[:space:]]+github.com/monshunter/easyinterview/backend/cmd/api([[:space:]]|$)" "$LOG"
! grep -Eq -- "--- FAIL:|^FAIL($|[[:space:]])|no tests to run|\\[no tests to run\\]" "$LOG"
! grep -Fq "__PRIVATE_DEBRIEF_TEXT__" "$LOG"
echo "E2E.P0.072 PASS"
