#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-045-practice-text-loop-mode-policy-display"
LOG_FILE="$OUTPUT_DIR/trigger.log"
PRACTICE_DIR="$REPO_ROOT/frontend/src/app/screens/practice"
test -s "$LOG_FILE"
"$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" \
  "$LOG_FILE" \
  "${SCENARIO_ID:-$(basename "$OUTPUT_DIR")}"
grep -Fq 'PracticeScreen.test.tsx' "$LOG_FILE" || { echo "E2E.P0.045: PracticeScreen.test.tsx did not run" >&2; exit 1; }
grep -Fq 'practiceGoalParity.test.tsx' "$LOG_FILE" || { echo "E2E.P0.045: practiceGoalParity.test.tsx did not run" >&2; exit 1; }
grep -Fq 'practiceHints.test.tsx' "$LOG_FILE" || { echo "E2E.P0.045: practiceHints.test.tsx did not run" >&2; exit 1; }
grep -Fq 'practicePauseResume.test.tsx' "$LOG_FILE" || { echo "E2E.P0.045: practicePauseResume.test.tsx did not run" >&2; exit 1; }
grep -Fq 'practiceTargetDisplay.test.tsx' "$LOG_FILE" || { echo "E2E.P0.045: practiceTargetDisplay.test.tsx did not run" >&2; exit 1; }
grep -Fq 'practiceVoiceTurn.test.tsx' "$LOG_FILE" || { echo "E2E.P0.045: practiceVoiceTurn.test.tsx did not run" >&2; exit 1; }
grep -Fq 'practiceSessionContinuity.test.tsx' "$LOG_FILE" || { echo "E2E.P0.045: practiceSessionContinuity.test.tsx did not run" >&2; exit 1; }
grep -Fq 'practiceModeSwitch.test.tsx' "$LOG_FILE" || { echo "E2E.P0.045: practiceModeSwitch.test.tsx did not run" >&2; exit 1; }
grep -Fq 'SessionMap.test.tsx' "$LOG_FILE" || { echo "E2E.P0.045: SessionMap.test.tsx did not run" >&2; exit 1; }
grep -Fq 'usePracticeTargetDisplay.test.tsx' "$LOG_FILE" || { echo "E2E.P0.045: usePracticeTargetDisplay.test.tsx did not run" >&2; exit 1; }
grep -Fq 'usePracticePhoneController.test.tsx' "$LOG_FILE" || { echo "E2E.P0.045: usePracticePhoneController.test.tsx did not run" >&2; exit 1; }
grep -Fq 'phoneVad.test.ts' "$LOG_FILE" || { echo "E2E.P0.045: phoneVad.test.ts did not run" >&2; exit 1; }
grep -Fq 'phoneVadMonitor.test.ts' "$LOG_FILE" || { echo "E2E.P0.045: phoneVadMonitor.test.ts did not run" >&2; exit 1; }
grep -Fq 'usePracticeVoicePlayback.test.tsx' "$LOG_FILE" || { echo "E2E.P0.045: usePracticeVoicePlayback.test.tsx did not run" >&2; exit 1; }
grep -Fq 'usePracticeVoiceTurn.lifecycle.test.tsx' "$LOG_FILE" || { echo "E2E.P0.045: usePracticeVoiceTurn.lifecycle.test.tsx did not run" >&2; exit 1; }
grep -Fq 'usePracticeSessionLoader.test.tsx' "$LOG_FILE" || { echo "E2E.P0.045: usePracticeSessionLoader.test.tsx did not run" >&2; exit 1; }
HANDSET_COUNT="$({ rg -o 'data-testid="practice-topbar-phone-toggle"' "$PRACTICE_DIR/components/TopBar.tsx" || true; } | wc -l | tr -d '[:space:]')"
if [[ "$HANDSET_COUNT" != "1" ]]; then
  echo "E2E.P0.045: expected exactly one production handset toggle, found $HANDSET_COUNT" >&2
  exit 1
fi
grep -Fq 'data-testid="practice-phone-hangup"' "$PRACTICE_DIR/components/PracticePhoneSurface.tsx" || { echo "E2E.P0.045: center hang-up control missing" >&2; exit 1; }
grep -Fq 'activeMode === "phone" ? exitPhoneMode : enterPhoneMode' "$PRACTICE_DIR/PracticeScreen.tsx" || { echo "E2E.P0.045: top-bar phone toggle does not share exitPhoneMode" >&2; exit 1; }
grep -Fq 'onHangUp={exitPhoneMode}' "$PRACTICE_DIR/PracticeScreen.tsx" || { echo "E2E.P0.045: center hang-up does not share exitPhoneMode" >&2; exit 1; }
rg -Uq 'interruptPlaybackRef\s*\.current\("mode_switch"\)' "$PRACTICE_DIR/usePracticePhoneController.ts" || { echo "E2E.P0.045: phone exit does not use mode-switch interruption" >&2; exit 1; }
if rg -n 'practice-topbar-mode-segment|practice-topbar-mode-text|practice-topbar-mode-phone|practice-topbar-live|practice-phone-restart|callEnded|onRestartCall' "$PRACTICE_DIR" -g '!*.test.*' -g '!__tests__/**'; then
  echo "E2E.P0.045: superseded phone-mode testids or restart/call-ended state leaked" >&2
  exit 1
fi
if rg -n -i 'phone.?restart|restart.?phone|restart.?call|call.?restart|onRestartCall|callEnded|practice-phone-restart' \
  "$REPO_ROOT/backend" "$REPO_ROOT/openapi" "$REPO_ROOT/shared" "$REPO_ROOT/migrations" "$REPO_ROOT/config"; then
  echo "E2E.P0.045: backend or contract phone-restart residue leaked" >&2
  exit 1
fi
if rg -n 'questionIntent' "$PRACTICE_DIR" -g '!*.test.*' -g '!__tests__/**'; then
  echo "E2E.P0.045: raw questionIntent reached production Practice rendering" >&2
  exit 1
fi
if rg -n "practiceMode\s*[=:]\s*['\"]debrief['\"]|PracticeGoalDebrief|goal\s*[=:]\s*['\"]debrief['\"]" "$PRACTICE_DIR" -g '!*.test.*' -g '!__tests__/**'; then
  echo "E2E.P0.045: out-of-scope practice goal literal leaked" >&2
  exit 1
fi
if rg -n '切到语音|Switch to voice' "$PRACTICE_DIR" -g '!*.test.*' -g '!__tests__/**'; then
  echo "E2E.P0.045: out-of-scope mode-switch copy leaked" >&2
  exit 1
fi
if rg -n 'practice-input-skip|practice-topbar-strict|practice-topbar-role|语音转文字|Speech-to-text|插入转写|表达层指标|口头禅|长停顿|语速|音量' "$PRACTICE_DIR" -g '!*.test.*' -g '!__tests__/**'; then
  echo "E2E.P0.045: out-of-scope practice controls or copy leaked" >&2
  exit 1
fi
# usePracticeEvents must NOT set the Idempotency-Key header on the request.
# Detect any header assignment that reaches the wire.
if rg -n '"Idempotency-Key"\s*:|idempotencyKey\s*:|setIdempotencyKey\b|opts\.idempotencyKey' "$PRACTICE_DIR/hooks/usePracticeEvents.ts"; then
  echo "E2E.P0.045: usePracticeEvents must NOT set Idempotency-Key on appendSessionEvent" >&2
  exit 1
fi
echo "E2E.P0.045 PASS"
