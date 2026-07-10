#!/usr/bin/env bash
set -euo pipefail

SCENARIO_DIR="${1:?scenario directory required}"
REPO_ROOT="$(git -C "$SCENARIO_DIR" rev-parse --show-toplevel)"
SCENARIO_ID="$(basename "$SCENARIO_DIR")"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/$SCENARIO_ID"

rm -f "$OUTPUT_DIR/setup.env"
