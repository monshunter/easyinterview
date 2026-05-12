#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)

python3 "$REPO_ROOT/scripts/lint/migrations_lint.py" --repo-root "$REPO_ROOT"
