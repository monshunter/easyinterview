#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-044-practice-text-loop-assisted-happy-path"
LOG_FILE="$OUTPUT_DIR/trigger.log"
PRACTICE_DIR="$REPO_ROOT/frontend/src/app/screens/practice"
test -s "$LOG_FILE"
grep -Eq 'Test Files +[0-9]+ passed \([0-9]+\)' "$LOG_FILE" || { echo "E2E.P0.044: no passing test files found" >&2; exit 1; }
grep -Fq 'PracticeScreen.test.tsx' "$LOG_FILE" || { echo "E2E.P0.044: PracticeScreen.test.tsx did not run" >&2; exit 1; }
grep -Fq 'usePracticeEvents.test.tsx' "$LOG_FILE" || { echo "E2E.P0.044: usePracticeEvents.test.tsx did not run" >&2; exit 1; }
grep -Fq 'AssistantActionRenderer.test.tsx' "$LOG_FILE" || { echo "E2E.P0.044: AssistantActionRenderer.test.tsx did not run" >&2; exit 1; }
testid_count="$(rg -o 'data-testid=' "$PRACTICE_DIR/PracticeScreen.tsx" "$PRACTICE_DIR/components/" | wc -l | tr -d ' ')"
if [ "$testid_count" -lt 20 ]; then
  echo "E2E.P0.044: expected >=20 practice runtime testids, got $testid_count" >&2
  exit 1
fi
if rg -n 'VoiceSessionSurface|PracticeWaveformBars|PracticeAnnotatedWaveform|VoiceExpressionPanel' "$PRACTICE_DIR" -g '!*.test.*' -g '!__tests__/**'; then
  echo "E2E.P0.044: forbidden voice surface DOM import in practice runtime" >&2
  exit 1
fi
if rg -n 'window\.EI_DATA|getPracticeSampleQuestions|getPracticeSampleTranscript|getPracticeWaveformSamples' "$PRACTICE_DIR" -g '!*.test.*' -g '!__tests__/**'; then
  echo "E2E.P0.044: forbidden prototype data helper in practice runtime" >&2
  exit 1
fi
if rg -n 'practice-mode-card-|growth-summary|drill-builder-|mistakes-queue-' "$PRACTICE_DIR" -g '!*.test.*' -g '!__tests__/**'; then
  echo "E2E.P0.044: forbidden legacy testid leaked into practice runtime" >&2
  exit 1
fi
if rg -n '\bgetFeedbackReport\b|\bcreatePracticeVoiceTurn\b' "$PRACTICE_DIR" -g '!*.test.*' -g '!__tests__/**'; then
  echo "E2E.P0.044: out-of-scope generated client method called from practice runtime" >&2
  exit 1
fi
echo "E2E.P0.044 PASS"
