#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")/../../../../.."
cd backend
go test -v ./cmd/api -run '^TestE2EP0050PracticeAssistantActionProvenanceAndTaskRuns$' -count=1 | tee ../.test-output/e2e/p0-050/trigger.log
