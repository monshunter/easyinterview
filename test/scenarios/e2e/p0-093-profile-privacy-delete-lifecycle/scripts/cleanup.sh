#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-083-profile-privacy-delete-lifecycle"

# Trigger test owns DB cleanup via t.Cleanup; this script trims local logs.
if [[ -f "$OUTPUT_DIR/negative-grep.log" ]]; then
  rm -f "$OUTPUT_DIR/negative-grep.log"
fi
