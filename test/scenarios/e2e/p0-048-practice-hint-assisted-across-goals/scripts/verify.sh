#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")/../../../../.."
grep -q "TestE2EP0048PracticeHintAssistedAcrossGoals" .test-output/e2e/p0-048/trigger.log
grep -q "github.com/monshunter/easyinterview/backend/cmd/api" .test-output/e2e/p0-048/trigger.log
! grep -q "no tests to run" .test-output/e2e/p0-048/trigger.log
