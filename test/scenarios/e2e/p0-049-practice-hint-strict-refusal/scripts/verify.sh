#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")/../../../../.."
LOG=.test-output/e2e/p0-049/trigger.log
grep -Fq "=== RUN   TestE2EP0049PracticeHintStrictRefusalAcrossGoals" "$LOG"
grep -Fq -- "--- PASS: TestE2EP0049PracticeHintStrictRefusalAcrossGoals" "$LOG"
grep -Eq "^ok[[:space:]]+github.com/monshunter/easyinterview/backend/cmd/api([[:space:]]|$)" "$LOG"
! grep -Eq -- "--- FAIL:|^FAIL($|[[:space:]])|no tests to run" "$LOG"
