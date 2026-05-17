#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-007-cascaded-voice-turn"
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"
echo "E2E.P0.007 setup complete" >"$OUTPUT_DIR/setup.log"
