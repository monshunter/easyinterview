#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-025-practice-idempotency-and-isolation-matrix"

mkdir -p "$OUTPUT_DIR"

(
  cd "$REPO_ROOT/backend"
  go test -v ./cmd/api -run 'TestE2EP0025PracticeIdempotencyAndIsolationMatrix' -count=1
) | tee "$OUTPUT_DIR/trigger.log"
