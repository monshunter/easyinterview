#!/usr/bin/env bash
set -euo pipefail

LOG_FILE="${1:?log file required}"
SCENARIO_ID="${2:-frontend-real-backend}"
OWNER_TEST="${3:-clientFactory.test.ts}"

grep -Fq 'VITE_EI_API_MODE=real' "$LOG_FILE" || {
  echo "$SCENARIO_ID: frontend real-backend mode marker missing" >&2
  exit 1
}
grep -Fq 'VITE_EI_API_BASE_URL=http://localhost:8080/api/v1' "$LOG_FILE" || {
  echo "$SCENARIO_ID: frontend real-backend base URL marker missing" >&2
  exit 1
}
grep -Fq "$OWNER_TEST" "$LOG_FILE" || {
  echo "$SCENARIO_ID: $OWNER_TEST did not run" >&2
  exit 1
}
grep -Eq '^[[:space:]]*RUN[[:space:]]+v[0-9]' "$LOG_FILE" || {
  echo "$SCENARIO_ID: vitest runner marker missing" >&2
  exit 1
}
if grep -Eiq 'No test files found|No tests found|No test suite found|No test cases found' "$LOG_FILE"; then
  echo "$SCENARIO_ID: no-test marker found" >&2
  exit 1
fi
if grep -Eq '^[[:space:]]*Test Files[[:space:]].*failed|^[[:space:]]*Tests[[:space:]].*failed' "$LOG_FILE"; then
  echo "$SCENARIO_ID: failing vitest summary found" >&2
  exit 1
fi
grep -Eq '^[[:space:]]*Test Files[[:space:]]+[1-9][0-9]*[[:space:]]+passed' "$LOG_FILE" || {
  echo "$SCENARIO_ID: no passing test files" >&2
  exit 1
}
grep -Eq '^[[:space:]]*Tests[[:space:]]+[1-9][0-9]*[[:space:]]+passed' "$LOG_FILE" || {
  echo "$SCENARIO_ID: no passing tests" >&2
  exit 1
}
