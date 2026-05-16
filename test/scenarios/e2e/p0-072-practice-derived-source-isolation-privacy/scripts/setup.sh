#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")/../../../../.."
OUT=".test-output/e2e/p0-072-practice-derived-source-isolation-privacy"
mkdir -p "$OUT"
printf 'scenario=E2E.P0.072\nsetup_at=%s\n' "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" > "$OUT/setup.env"
