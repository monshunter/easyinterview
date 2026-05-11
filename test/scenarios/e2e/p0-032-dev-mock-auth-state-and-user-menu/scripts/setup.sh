#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-032-dev-mock-auth-state-and-user-menu"

mkdir -p "$OUTPUT_DIR"
printf 'scenario=E2E.P0.032\nsetup_at=%s\n' "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" > "$OUTPUT_DIR/setup.env"
