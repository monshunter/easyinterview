#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-057-replay-cta-paths-a-and-b"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
grep -Eq 'Test Files +[0-9]+ passed' "$LOG_FILE" || { echo "E2E.P0.057: no passing test files" >&2; exit 1; }
grep -Fq 'pendingActionReplayPractice.test.ts' "$LOG_FILE" || { echo "E2E.P0.057: pendingAction replay test did not run" >&2; exit 1; }
grep -Fq 'ReplayCta.test.tsx' "$LOG_FILE" || { echo "E2E.P0.057: ReplayCta test did not run" >&2; exit 1; }

# Replay CTA payload must not assign raw text literals into payload keys.
# Match key-style usages (`answerText:`, `answerText =`, `"answerText"`).
if grep -RnE '("answerText"|''answerText''|answerText\s*[:=]|"questionText"|''questionText''|questionText\s*[:=]|promptHash\s*[:=]|"promptHash")' \
    "$REPO_ROOT/frontend/src/app/screens/report" \
    --include='*.tsx' --include='*.ts' --exclude-dir=__tests__; then
  echo "E2E.P0.057: raw text leaked into replay payload code" >&2
  exit 1
fi

# practiceGoal values surface in both paths.
grep -Fq 'retry_current_round' "$REPO_ROOT/frontend/src/app/screens/report/handoff.ts"
grep -Fq 'next_round' "$REPO_ROOT/frontend/src/app/screens/report/handoff.ts"
