#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-091-candidate-profile-seed-and-patch"

mkdir -p "$OUTPUT_DIR"
export DATABASE_URL="${DATABASE_URL:-postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable}"
(
  cd "$REPO_ROOT/backend"
  go test ./cmd/api -run TestProfileHTTPScenario -count=1 -v
) | tee "$OUTPUT_DIR/trigger.log"
