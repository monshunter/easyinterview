#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")/../../../../.."
LOG=".test-output/e2e/p0-071-practice-debrief-start-source-question/trigger.log"
test -s "$LOG"
grep -Fq "E2E.P0.071 RUNNER go test" "$LOG"
grep -Fq "=== RUN   TestE2EP0071PracticeDebriefStartUsesSourceQuestion" "$LOG"
grep -Fq -- "--- PASS: TestE2EP0071PracticeDebriefStartUsesSourceQuestion" "$LOG"
grep -Eq "^ok[[:space:]]+github.com/monshunter/easyinterview/backend/cmd/api([[:space:]]|$)" "$LOG"
! grep -Eq -- "--- FAIL:|^FAIL($|[[:space:]])|no tests to run|\\[no tests to run\\]" "$LOG"
! grep -Fq "__DEBRIEF_FIRST_QUESTION__" "$LOG"
echo "E2E.P0.071 PASS"
