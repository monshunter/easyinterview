#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")/../../../../.."
ROOT="$(pwd)"
OUT="$ROOT/.test-output/e2e/p0-072-practice-derived-source-isolation-privacy"
mkdir -p "$OUT"
tmp_log="$(mktemp)"
trap 'rm -f "$tmp_log"' EXIT
set +e
{
  echo "E2E.P0.072 RUNNER go test"
  cd backend
  go test -v ./cmd/api -run '^TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy$' -count=1
} >"$tmp_log" 2>&1
go_test_status=$?
set -e
tee "$OUT/trigger.log" <"$tmp_log"
exit "$go_test_status"
