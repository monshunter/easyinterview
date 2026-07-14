#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-018-workspace-default-render"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
"$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG_FILE" "${SCENARIO_ID:-$(basename "$OUTPUT_DIR")}"
grep -Fq 'TopBar.test.tsx' "$LOG_FILE" || { echo "E2E.P0.018: TopBar label test did not run" >&2; exit 1; }
grep -Fq 'p0-004-app-shell-language-switch.test.tsx' "$LOG_FILE" || { echo "E2E.P0.018: app shell language scenario did not run" >&2; exit 1; }
grep -Fq 'WorkspaceEmptyState.test.tsx' "$LOG_FILE" || { echo "E2E.P0.018: workspace no-context landing test did not run" >&2; exit 1; }
grep -Fq 'useWorkspaceTargetJobs.test.tsx' "$LOG_FILE" || { echo "E2E.P0.018: StrictMode list request-count test did not run" >&2; exit 1; }
grep -Fq 'ParseFlow.test.tsx' "$LOG_FILE" || { echo "E2E.P0.018: workspace detail route/read test did not run" >&2; exit 1; }
grep -Fq 'ParseResumeBinding.test.tsx' "$LOG_FILE" || { echo "E2E.P0.018: workspace detail resume/start test did not run" >&2; exit 1; }
grep -Fq 'ParseRoundStates.test.tsx' "$LOG_FILE" || { echo "E2E.P0.018: workspace round-state test did not run" >&2; exit 1; }
grep -Fq 'ready plan card opens workspace detail without Parse animation or route-side mutation' "$LOG_FILE" || { echo "E2E.P0.018: ready-card browser gate did not run" >&2; exit 1; }
grep -Fq 'workspace detail round states match the UI truth at desktop and mobile' "$LOG_FILE" || { echo "E2E.P0.018: round-state browser parity did not run" >&2; exit 1; }
testid_count="$(
  rg -o 'data-testid=' \
    "$REPO_ROOT/frontend/src/app/screens/workspace/WorkspaceScreen.tsx" \
    "$REPO_ROOT/frontend/src/app/screens/home/MockInterviewCard.tsx" \
    | wc -l | tr -d ' '
)"
if [ "$testid_count" -lt 14 ]; then
  echo "E2E.P0.018: expected >=14 workspace list/shared-card runtime testids, got $testid_count" >&2
  exit 1
fi
if rg -n 'practice-mode-card-|growth-center|drill-builder|mistake-queue' "$REPO_ROOT/frontend/src/app/screens/workspace" -g '!*.test.tsx'; then
  echo "E2E.P0.018: forbidden out-of-scope runtime testid leaked" >&2
  exit 1
fi
if rg -n 'workspace-resume-modal-disabled-note|resumePicker\.disabledNote' \
  "$REPO_ROOT/frontend/src/app/screens/workspace" \
  -g '!*.test.tsx'; then
  echo "E2E.P0.018: out-of-scope disabled resume picker wording leaked" >&2
  exit 1
fi
grep -Fq 'workspace-plan-list' "$REPO_ROOT/frontend/src/app/screens/workspace/WorkspaceScreen.tsx" || {
  echo "E2E.P0.018: workspace no-context plan-list anchor missing" >&2
  exit 1
}
grep -Fq 'name: "workspace"' "$REPO_ROOT/frontend/src/app/screens/workspace/WorkspaceScreen.tsx" || {
  echo "E2E.P0.018: workspace plan cards must open target-scoped workspace detail" >&2
  exit 1
}
if rg -n 'name: "parse"' "$REPO_ROOT/frontend/src/app/screens/workspace/WorkspaceScreen.tsx"; then
  echo "E2E.P0.018: ready workspace card still routes through Parse" >&2
  exit 1
fi
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
grep -Fq 'workspace-plan-list-start-' "$REPO_ROOT/frontend/src/app/screens/workspace/WorkspaceScreen.tsx" || {
  echo "E2E.P0.018: workspace plan-list quick-start action missing" >&2
  exit 1
}
grep -Fq 'workspace-plan-list-delete-' "$REPO_ROOT/frontend/src/app/screens/workspace/WorkspaceScreen.tsx" || {
  echo "E2E.P0.018: workspace plan-list delete action missing" >&2
  exit 1
}
grep -Fq 'position: "absolute"' "$REPO_ROOT/frontend/src/app/screens/home/MockInterviewCard.tsx" || {
  echo "E2E.P0.018: workspace plan-list delete action is not positioned at card top-right" >&2
  exit 1
}
grep -Fq 'right: 14' "$REPO_ROOT/frontend/src/app/screens/home/MockInterviewCard.tsx" || {
  echo "E2E.P0.018: workspace plan-list delete action is not anchored to the card right edge" >&2
  exit 1
}
if rg -n 'workspace-plan-list-open-' "$REPO_ROOT/frontend/src/app/screens/workspace" -g '!*.test.tsx'; then
  echo "E2E.P0.018: visible Open plan footer button returned to workspace runtime" >&2
  exit 1
fi
grep -Fq 'MockInterviewCard' "$REPO_ROOT/frontend/src/app/screens/workspace/WorkspaceScreen.tsx" || {
  echo "E2E.P0.018: workspace plan-list cards do not reuse the Home recent card body" >&2
  exit 1
}
grep -Fq 'railTestId' "$REPO_ROOT/frontend/src/app/screens/home/MockInterviewCard.tsx" || {
  echo "E2E.P0.018: shared interview card lacks mini round rail test hook" >&2
  exit 1
}
if rg -n 'workspace\.planList\.cardMeta|job\.targetLanguage\?\.toUpperCase|job\.sourceType \? formatSourceType' "$REPO_ROOT/frontend/src/app/screens/workspace/WorkspaceScreen.tsx"; then
  echo "E2E.P0.018: workspace plan-list cards leaked source/language metadata" >&2
  exit 1
fi
if rg -n '"workspace\.planList\.cardMeta"' "$REPO_ROOT/frontend/src/app/i18n/locales"; then
  echo "E2E.P0.018: out-of-scope plan-list cardMeta locale key remains" >&2
  exit 1
fi
grep -Fq 'background: "var(--ei-color-accent)"' "$REPO_ROOT/frontend/src/app/screens/home/MockInterviewCard.tsx" || {
  echo "E2E.P0.018: workspace plan-list quick-start CTA is not theme accent" >&2
  exit 1
}
if rg -n 'autoStartPractice|useStartPractice|PlanSwitcherModal|ResumePickerModal|WorkspaceInsightCard|useWorkspaceTargetJob\W|useWorkspaceResume|useWorkspacePracticePlan' \
  "$REPO_ROOT/frontend/src/app/screens/workspace" \
  -g '!*.test.tsx'; then
  echo "E2E.P0.018: workspace list module leaked out-of-scope detail/start/modal context" >&2
  exit 1
fi
grep -Fq 'startPracticeFromParams' "$REPO_ROOT/frontend/src/app/screens/parse/ParseScreen.tsx" || {
  echo "E2E.P0.018: shared workspace detail no longer owns start-practice handoff" >&2
  exit 1
}
grep -Fq 'workspace-plan-list-card-footer-' "$REPO_ROOT/ui-design/src/screen-workspace.jsx" || {
  echo "E2E.P0.018: ui-design plan-list card footer source missing" >&2
  exit 1
}
grep -Fq 'Icon name="trash"' "$REPO_ROOT/ui-design/src/screen-workspace.jsx" || {
  echo "E2E.P0.018: ui-design plan-list delete icon missing" >&2
  exit 1
}
grep -Fq 'position: "absolute", top: 20, right: 20' "$REPO_ROOT/ui-design/src/screen-workspace.jsx" || {
  echo "E2E.P0.018: ui-design plan-list delete icon is not fixed at card top-right" >&2
  exit 1
}
grep -Fq 'workspace-plan-list-rail-' "$REPO_ROOT/ui-design/src/screen-workspace.jsx" || {
  echo "E2E.P0.018: ui-design plan-list card mini round rail missing" >&2
  exit 1
}
grep -Fq 'nav("workspace"' "$REPO_ROOT/ui-design/src/screen-workspace.jsx" || {
  echo "E2E.P0.018: ui-design workspace cards must open workspace detail" >&2
  exit 1
}
if rg -n 'nav\("parse"' "$REPO_ROOT/ui-design/src/screen-workspace.jsx"; then
  echo "E2E.P0.018: ui-design ready card still opens Parse" >&2
  exit 1
fi
grep -Fq 'data-round-state' "$REPO_ROOT/frontend/src/app/screens/parse/ParseScreen.tsx" || {
  echo "E2E.P0.018: workspace detail round-state contract missing" >&2
  exit 1
}
grep -Fq 'sequence=done,current,pending distinctBackgrounds=3 distinctBorders=3' "$LOG_FILE" || {
  echo "E2E.P0.018: distinct round-state browser evidence missing" >&2
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
