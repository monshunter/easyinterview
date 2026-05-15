#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")/../../../../.."
mkdir -p .test-output/e2e/p0-050
cd backend
tmp_log="$(mktemp)"
trap 'rm -f "$tmp_log"' EXIT
set +e
go test -v ./cmd/api -run '^TestE2EP0050PracticeAssistantActionProvenanceAndTaskRuns$' -count=1 >"$tmp_log" 2>&1
go_test_status=$?
set -e
tee ../.test-output/e2e/p0-050/trigger.log <"$tmp_log"
exit "$go_test_status"
