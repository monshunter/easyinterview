#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")/../../../../.."
grep -q "TestE2EP0050PracticeAssistantActionProvenanceAndTaskRuns" .test-output/e2e/p0-050/trigger.log
grep -q "github.com/monshunter/easyinterview/backend/cmd/api" .test-output/e2e/p0-050/trigger.log
! grep -q "no tests to run" .test-output/e2e/p0-050/trigger.log
