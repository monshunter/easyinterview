#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
SCENARIO_ID="$(basename "$(dirname "$SCRIPT_DIR")")"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/$SCENARIO_ID"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
"$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG_FILE" "${SCENARIO_ID:-$(basename "$OUTPUT_DIR")}"
grep -Fq 'WorkspaceScreen.test.tsx' "$LOG_FILE" || { echo "$SCENARIO_ID: workspace boundary test did not run" >&2; exit 1; }
grep -Fq 'buildCreatePlanRequest.test.ts' "$LOG_FILE" || { echo "$SCENARIO_ID: plan budget request test did not run" >&2; exit 1; }
grep -Fq 'startPractice.test.ts' "$LOG_FILE" || { echo "$SCENARIO_ID: structured-round start test did not run" >&2; exit 1; }
grep -Fq 'ReplayCta.test.tsx' "$LOG_FILE" || { echo "$SCENARIO_ID: report replay handoff test did not run" >&2; exit 1; }
if rg -n '\.getCompany[A-Za-z]*Insight\(|\.getFeedbackReport\(|recentSessions|console\.log\(' \
  "$REPO_ROOT/frontend/src/app/screens/workspace" \
  "$REPO_ROOT/frontend/src/app/interview-context" \
  -g '!*.test.tsx'; then
  echo "$SCENARIO_ID: forbidden runtime call or out-of-scope field leaked" >&2
  exit 1
fi
if rg -n 'questionText|answerText|hintText|promptHash|rawTranscript|jdRaw|resumeRaw' \
  "$REPO_ROOT/frontend/src/app/screens/workspace" \
  "$REPO_ROOT/frontend/src/app/interview-context" \
  -g '!*.test.tsx'; then
  echo "$SCENARIO_ID: forbidden runtime privacy field leaked" >&2
  exit 1
fi
if rg -n 'practice-mode-card-|growth-center|drill-builder|mistake-queue|workspace-mocked-' \
  "$REPO_ROOT/frontend/src/app/screens/workspace" \
  "$REPO_ROOT/frontend/src/app/interview-context" \
  -g '!*.test.tsx'; then
  echo "$SCENARIO_ID: forbidden out-of-scope runtime testid leaked" >&2
  exit 1
fi
if rg -n 'ui-design/src/data|window\.EI_DATA|getWorkspace' \
  "$REPO_ROOT/frontend/src/app/screens/workspace" \
  "$REPO_ROOT/frontend/src/app/interview-context" \
  -g '!*.test.tsx'; then
  echo "$SCENARIO_ID: prototype data import/helper leaked into runtime" >&2
  exit 1
fi
echo "$SCENARIO_ID PASS"
