#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec "$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)/test/scenarios/_shared/scripts/scenario-evidence-cleanup.sh" "$(dirname "$SCRIPT_DIR")"
