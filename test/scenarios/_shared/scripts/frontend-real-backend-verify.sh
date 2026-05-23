#!/usr/bin/env bash
set -euo pipefail

LOG_FILE="${1:?log file required}"
SCENARIO_ID="${2:-frontend-real-backend}"

grep -Fq 'VITE_EI_API_MODE=real' "$LOG_FILE" || {
  echo "$SCENARIO_ID: frontend real-backend mode marker missing" >&2
  exit 1
}
grep -Fq 'VITE_EI_API_BASE_URL=http://localhost:8080/api/v1' "$LOG_FILE" || {
  echo "$SCENARIO_ID: frontend real-backend base URL marker missing" >&2
  exit 1
}
grep -Fq 'frontendOwners.realApiMode.test.ts' "$LOG_FILE" || {
  echo "$SCENARIO_ID: frontendOwners.realApiMode.test.ts did not run" >&2
  exit 1
}
