#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-006-ui-design-pixel-parity-gate"

DIST_HTML="$REPO_ROOT/frontend/dist/index.html"
UI_DESIGN_HTML="$REPO_ROOT/ui-design/index.html"

if [[ ! -f "$DIST_HTML" ]]; then
  echo "[setup] frontend dist missing at $DIST_HTML" >&2
  echo "        run 'pnpm --filter @easyinterview/frontend build' first" >&2
  exit 1
fi

if [[ ! -f "$UI_DESIGN_HTML" ]]; then
  echo "[setup] ui-design index.html missing at $UI_DESIGN_HTML" >&2
  exit 1
fi

# Detect Playwright chromium cache (macOS / Linux default).
PW_CACHE_CANDIDATES=(
  "$HOME/Library/Caches/ms-playwright"
  "$HOME/.cache/ms-playwright"
)
PW_CACHE_DIR=""
for candidate in "${PW_CACHE_CANDIDATES[@]}"; do
  if [[ -d "$candidate" ]]; then
    PW_CACHE_DIR="$candidate"
    break
  fi
done

if [[ -z "$PW_CACHE_DIR" ]]; then
  echo "[setup] Playwright cache not found in: ${PW_CACHE_CANDIDATES[*]}" >&2
  echo "        run 'pnpm --filter @easyinterview/frontend test:pixel-parity:install' first" >&2
  exit 1
fi

if ! ls "$PW_CACHE_DIR" 2>/dev/null | grep -q chromium; then
  echo "[setup] chromium not installed under $PW_CACHE_DIR" >&2
  echo "        run 'pnpm --filter @easyinterview/frontend test:pixel-parity:install' first" >&2
  exit 1
fi

mkdir -p "$OUTPUT_DIR"
printf 'scenario=E2E.P0.006\nsetup_at=%s\nplaywright_cache=%s\n' \
  "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" \
  "$PW_CACHE_DIR" > "$OUTPUT_DIR/setup.env"
