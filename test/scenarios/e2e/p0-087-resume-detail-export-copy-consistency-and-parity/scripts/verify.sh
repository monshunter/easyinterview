#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
SCENARIO_ID="$(basename "$(dirname "$SCRIPT_DIR")")"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/$SCENARIO_ID"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
"$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG_FILE" "${SCENARIO_ID:-$(basename "$OUTPUT_DIR")}"
grep -Eq '^[[:space:]]*RUN[[:space:]]+v[0-9]' "$LOG_FILE" || { echo "$SCENARIO_ID: vitest runner marker missing" >&2; exit 1; }
if grep -Eiq 'No test files found|No tests found|No test suite found|No test cases found' "$LOG_FILE"; then echo "$SCENARIO_ID: no-test marker found" >&2; exit 1; fi
if grep -Eq '^[[:space:]]*Test Files[[:space:]].*failed|^[[:space:]]*Tests[[:space:]].*failed' "$LOG_FILE"; then echo "$SCENARIO_ID: failing vitest summary found" >&2; exit 1; fi
grep -Eq '^[[:space:]]*Test Files[[:space:]]+[1-9][0-9]*[[:space:]]+passed' "$LOG_FILE" || { echo "$SCENARIO_ID: no passing test files" >&2; exit 1; }
grep -Eq '^[[:space:]]*Tests[[:space:]]+[1-9][0-9]*[[:space:]]+passed' "$LOG_FILE" || { echo "$SCENARIO_ID: no passing tests" >&2; exit 1; }
for spec in ResumeDetailExport.test.tsx ResumeDetailFixtureParity.test.tsx ResumeDetailView.test.tsx ResumeRewritesTab.test.tsx ResumeEditTab.test.tsx; do
  grep -qF "$spec" "$LOG_FILE" || { echo "$SCENARIO_ID: spec $spec not exercised" >&2; exit 1; }
done
grep -qF '## frontend-build' "$LOG_FILE" || { echo "$SCENARIO_ID: frontend build marker missing" >&2; exit 1; }
grep -Eq '✓ built in|built in [0-9.]+s' "$LOG_FILE" || { echo "$SCENARIO_ID: frontend build success marker missing" >&2; exit 1; }
grep -qF '## playwright-pixel-parity-axe' "$LOG_FILE" || { echo "$SCENARIO_ID: playwright marker missing" >&2; exit 1; }
grep -qF 'tests/pixel-parity/resume-workshop-branch-rewrites-edit.spec.ts' "$LOG_FILE" || { echo "$SCENARIO_ID: playwright spec not exercised" >&2; exit 1; }
grep -Eq 'Running[[:space:]]+[1-9][0-9]*[[:space:]]+tests?[[:space:]]+using' "$LOG_FILE" || { echo "$SCENARIO_ID: playwright runner marker missing" >&2; exit 1; }
grep -Eq '^[[:space:]]*[1-9][0-9]*[[:space:]]+passed[[:space:]]+\([0-9.]+s\)' "$LOG_FILE" || { echo "$SCENARIO_ID: playwright passing summary missing" >&2; exit 1; }
if grep -Eiq 'failed|timed out|Timeout|Error:|ERR_PNPM' "$LOG_FILE"; then echo "$SCENARIO_ID: failing build/playwright marker found" >&2; exit 1; fi
cd "$REPO_ROOT"
if rg -n "welcome|mistake|growth|drill|followup|STAR|ExperiencesScreen|experiences-route|voice|OnboardingScreen|onboarding=true|ResumeBranchFlow|branchResumeVersion|seedStrategy|updateResumeVersion|acceptResumeTailorSuggestion|rejectResumeTailorSuggestion" frontend/src/app/screens/resume-workshop --glob '!**/*.test.ts' --glob '!**/*.test.tsx' > "$OUTPUT_DIR/retired-modules-grep.log"; then
  echo "$SCENARIO_ID: retired modules grep matched" >&2; exit 1
fi
if rg -n "(^|[^A-Za-z0-9_])(inline|rewrite|mirror)([^A-Za-z0-9_]|$)" frontend/src/app/screens/resume-workshop/tabs --glob '!**/*.test.ts' --glob '!**/*.test.tsx' > "$OUTPUT_DIR/retired-tailor-mode-grep.log"; then
  echo "$SCENARIO_ID: retired tailor mode grep matched" >&2; exit 1
fi
if rg -n "ui-design/src/(data|screen-resume-workshop)" frontend/src/app/screens/resume-workshop/tabs --glob '!**/*.test.ts' --glob '!**/*.test.tsx' > "$OUTPUT_DIR/prototype-import-grep.log"; then
  echo "$SCENARIO_ID: prototype runtime import detected" >&2; exit 1
fi
