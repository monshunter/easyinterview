#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-097-jd-match-agent-scan-and-privacy-delete"
mkdir -p "$OUTPUT_DIR"
export DATABASE_URL="${DATABASE_URL:-postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable}"
cd "$REPO_ROOT/backend"
go test ./cmd/api -run '^(TestJDMatchHTTPScenario|TestJDMatchAgentScanDrainerScenario)$' -count=1 -v | tee "$OUTPUT_DIR/trigger.log"
printf 'scenario=E2E.P0.097\nmethod=cmd-api-http+live-postgres\ntrigger_at=%s\n' "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" > "$OUTPUT_DIR/trigger.env"
echo "trigger: ok"
