#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")/../../../../.."
OUT=".test-output/e2e/p0-072-practice-derived-source-isolation-privacy"
rm -f "$OUT/setup.env"
