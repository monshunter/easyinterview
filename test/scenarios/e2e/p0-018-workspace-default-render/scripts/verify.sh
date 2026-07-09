#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-018-workspace-default-render"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
"$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG_FILE" "${SCENARIO_ID:-$(basename "$OUTPUT_DIR")}"
grep -Eq 'Test Files +[0-9]+ passed \([0-9]+\)' "$LOG_FILE" || { echo "E2E.P0.018: no passing test files found" >&2; exit 1; }
grep -Fq 'TopBar.test.tsx' "$LOG_FILE" || { echo "E2E.P0.018: TopBar label test did not run" >&2; exit 1; }
grep -Fq 'p0-004-app-shell-language-switch.test.tsx' "$LOG_FILE" || { echo "E2E.P0.018: app shell language scenario did not run" >&2; exit 1; }
grep -Fq 'WorkspaceEmptyState.test.tsx' "$LOG_FILE" || { echo "E2E.P0.018: workspace no-context landing test did not run" >&2; exit 1; }
grep -Fq 'ParseResumeBinding.test.tsx' "$LOG_FILE" || { echo "E2E.P0.018: parse detail resume/start test did not run" >&2; exit 1; }
testid_count="$(
  rg -o 'data-testid=' \
    "$REPO_ROOT/frontend/src/app/screens/workspace/WorkspaceScreen.tsx" \
    | wc -l | tr -d ' '
)"
if [ "$testid_count" -lt 12 ]; then
  echo "E2E.P0.018: expected >=12 workspace list runtime testids, got $testid_count" >&2
  exit 1
fi
if rg -n 'practice-mode-card-|growth-center|drill-builder|mistake-queue' "$REPO_ROOT/frontend/src/app/screens/workspace" -g '!*.test.tsx'; then
  echo "E2E.P0.018: forbidden non-current runtime testid leaked" >&2
  exit 1
fi
if rg -n 'workspace-resume-modal-disabled-note|resumePicker\.disabledNote' \
  "$REPO_ROOT/frontend/src/app/screens/workspace" \
  -g '!*.test.tsx'; then
  echo "E2E.P0.018: non-current disabled resume picker wording leaked" >&2
  exit 1
fi
grep -Fq 'workspace-plan-list' "$REPO_ROOT/frontend/src/app/screens/workspace/WorkspaceScreen.tsx" || {
  echo "E2E.P0.018: workspace no-context plan-list anchor missing" >&2
  exit 1
}
grep -Fq 'name: "parse"' "$REPO_ROOT/frontend/src/app/screens/workspace/WorkspaceScreen.tsx" || {
  echo "E2E.P0.018: workspace plan cards must open parse detail, not workspace params" >&2
  exit 1
}
if rg -n -F -e 'jobId: job.id' -e 'jdId: `jd-${job.id}`' -e 'plan-${job.id}' "$REPO_ROOT/frontend/src/app/screens/workspace/WorkspaceScreen.tsx"; then
  echo "E2E.P0.018: workspace plan cards fabricated route context ids" >&2
  exit 1
fi
grep -Fq 'workspace-plan-list-card-body-' "$REPO_ROOT/frontend/src/app/screens/workspace/WorkspaceScreen.tsx" || {
  echo "E2E.P0.018: workspace plan-list card body section missing" >&2
  exit 1
}
grep -Fq 'workspace-plan-list-card-footer-' "$REPO_ROOT/frontend/src/app/screens/workspace/WorkspaceScreen.tsx" || {
  echo "E2E.P0.018: workspace plan-list card footer section missing" >&2
  exit 1
}
grep -Fq 'boxShadow: "var(--ei-shadow-elev2)"' "$REPO_ROOT/frontend/src/app/screens/workspace/WorkspaceScreen.tsx" || {
  echo "E2E.P0.018: workspace plan-list cards lack elevation token" >&2
  exit 1
}
if rg -n 'workspace\.planList\.cardMeta|job\.targetLanguage\?\.toUpperCase|job\.sourceType \? formatSourceType' "$REPO_ROOT/frontend/src/app/screens/workspace/WorkspaceScreen.tsx"; then
  echo "E2E.P0.018: workspace plan-list cards leaked source/language metadata" >&2
  exit 1
fi
if rg -n '"workspace\.planList\.cardMeta"' "$REPO_ROOT/frontend/src/app/i18n/locales"; then
  echo "E2E.P0.018: obsolete plan-list cardMeta locale key remains" >&2
  exit 1
fi
grep -Fq 'background: "var(--ei-color-accent)"' "$REPO_ROOT/frontend/src/app/screens/workspace/WorkspaceScreen.tsx" || {
  echo "E2E.P0.018: workspace plan-list open CTA is not theme accent" >&2
  exit 1
}
if rg -n 'autoStartPractice|useStartPractice|PlanSwitcherModal|ResumePickerModal|WorkspaceInsightCard|useWorkspaceTargetJob\W|useWorkspaceResume|useWorkspacePracticePlan' \
  "$REPO_ROOT/frontend/src/app/screens/workspace" \
  -g '!*.test.tsx'; then
  echo "E2E.P0.018: workspace list module leaked retired detail/start/modal context" >&2
  exit 1
fi
grep -Fq 'startPracticeFromParams' "$REPO_ROOT/frontend/src/app/screens/parse/ParseScreen.tsx" || {
  echo "E2E.P0.018: parse detail no longer owns start-practice handoff" >&2
  exit 1
}
grep -Fq 'workspace-plan-list-card-footer-' "$REPO_ROOT/ui-design/src/screen-workspace.jsx" || {
  echo "E2E.P0.018: ui-design plan-list card footer source missing" >&2
  exit 1
}
grep -Fq 'nav("parse"' "$REPO_ROOT/ui-design/src/screen-workspace.jsx" || {
  echo "E2E.P0.018: ui-design workspace cards must open parse detail" >&2
  exit 1
}
grep -Fq '"nav.workspace": "面试"' "$REPO_ROOT/frontend/src/app/i18n/locales/zh.ts" || {
  echo "E2E.P0.018: zh TopBar workspace label is not concise 面试" >&2
  exit 1
}
grep -Fq '"nav.workspace": "Interview"' "$REPO_ROOT/frontend/src/app/i18n/locales/en.ts" || {
  echo "E2E.P0.018: en TopBar workspace label is not concise Interview" >&2
  exit 1
}
echo "E2E.P0.018 PASS"
