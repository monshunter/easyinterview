#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")/../../../../.."
cd backend
go test -v ./cmd/api -run '^TestE2EP0048PracticeHintAssistedAcrossGoals$' -count=1 | tee ../.test-output/e2e/p0-048/trigger.log
