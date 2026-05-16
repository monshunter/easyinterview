#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")/../../../../.."
OUT=".test-output/e2e/p0-070-practice-derived-plan-create-read-replay"
rm -f "$OUT/setup.env"
