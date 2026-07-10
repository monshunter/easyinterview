#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
SCENARIO_ID="$(basename "$(dirname "$SCRIPT_DIR")")"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/$SCENARIO_ID"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
"$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG_FILE" "${SCENARIO_ID:-$(basename "$OUTPUT_DIR")}"
for spec in \
  ResumeWorkshopScreen.test.tsx \
  ResumeDetailView.test.tsx \
  ResumeCreateFlow.test.tsx \
  ResumeWorkshopAuthGate.test.tsx \
  ; do
  grep -qF "$spec" "$LOG_FILE" || { echo "$SCENARIO_ID: spec $spec not exercised" >&2; exit 1; }
done

# Flat-resume regression gate: out-of-scope form and operation tokens stay absent.
cd "$REPO_ROOT"
if rg -n "welcome|mistake|growth|drill|followup|STAR|ExperiencesScreen|experiences-route|voice|OnboardingScreen|onboarding=true|ResumeBranchFlow|branchResumeVersion|seedStrategy|acceptResumeTailorSuggestion|rejectResumeTailorSuggestion|updateResumeVersion" frontend/src/app/screens/resume-workshop --glob '!**/*.test.ts' --glob '!**/*.test.tsx' > "$OUTPUT_DIR/out-of-scope-modules-grep.log"; then
  echo "$SCENARIO_ID: out-of-scope modules grep matched something (see out-of-scope-modules-grep.log)" >&2
  exit 1
fi
if rg -n "(^|[^A-Za-z0-9_-])(inline|rewrite|mirror)([^A-Za-z0-9_-]|$)" frontend/src/app/screens/resume-workshop --glob '!**/*.test.ts' --glob '!**/*.test.tsx' > "$OUTPUT_DIR/out-of-scope-tailor-mode-grep.log"; then
  echo "$SCENARIO_ID: out-of-scope tailor mode grep matched something" >&2
  exit 1
fi
if rg -n "ui-design/src/(data|screen-resume-workshop)" frontend/src/app/screens/resume-workshop --glob '!**/*.test.ts' --glob '!**/*.test.tsx' > "$OUTPUT_DIR/prototype-import-grep.log"; then
  echo "$SCENARIO_ID: prototype runtime import detected" >&2
  exit 1
fi
