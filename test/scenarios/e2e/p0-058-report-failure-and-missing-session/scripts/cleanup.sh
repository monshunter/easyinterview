#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-058-report-failure-and-missing-session"
rm -f "$OUT/setup.env" "$OUT"/*.log
