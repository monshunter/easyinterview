#!/usr/bin/env sh
set -eu
ROOT="$(git -C "$(dirname "$0")" rev-parse --show-toplevel)"
rm -f "$ROOT/.test-output/e2e/p0-051-practice-assistance-stale-contract-negative/setup.env"
