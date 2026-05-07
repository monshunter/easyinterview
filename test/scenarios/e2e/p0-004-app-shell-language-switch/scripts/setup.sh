#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-004-app-shell-language-switch"

mkdir -p "$OUTPUT_DIR"
date -u +"%Y-%m-%dT%H:%M:%SZ" > "$OUTPUT_DIR/setup.marker"
