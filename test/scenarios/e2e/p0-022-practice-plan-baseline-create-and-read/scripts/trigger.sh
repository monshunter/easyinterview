#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-022-practice-plan-baseline-create-and-read"

mkdir -p "$OUTPUT_DIR"

(
  cd "$REPO_ROOT/backend"
  go test -v ./cmd/api -run 'TestE2EP0022PracticePlanBaselineCreateAndRead' -count=1
) | tee "$OUTPUT_DIR/trigger.log"
