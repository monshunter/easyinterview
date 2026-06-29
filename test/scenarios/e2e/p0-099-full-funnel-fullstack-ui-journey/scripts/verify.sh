#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey"
LOG="$OUTPUT_DIR/trigger.log"
SETUP_MARKER="$OUTPUT_DIR/setup.env"

if [ ! -s "$LOG" ]; then
  echo "verify: missing trigger.log" >&2
  exit 1
fi

if grep -Eq -- "0 passed|0 tests|no tests to run|No tests found|skipped" "$LOG"; then
  echo "verify: scenario log contains skip/no-test marker" >&2
  exit 1
fi

if grep -Eq -- "failed|timed out|Error:" "$LOG"; then
  echo "verify: scenario log contains failure marker" >&2
  exit 1
fi

for marker in \
  "tests/e2e/full-funnel-journey.spec.ts" \
  "E2E.P0.099 full funnel import to next-round practice" \
  "P0.099 backend server listening" \
  "job_type=resume_parse" \
  "job_type=target_import" \
  "job_type=report_generate" \
  "1 passed"; do
  if ! grep -q -- "$marker" "$LOG"; then
    echo "verify: missing marker $marker" >&2
    exit 1
  fi
done

if [ ! -s "$OUTPUT_DIR/state.json" ]; then
  echo "verify: missing backend state.json" >&2
  exit 1
fi

if [ ! -d "$OUTPUT_DIR/playwright" ]; then
  echo "verify: missing Playwright output directory" >&2
  exit 1
fi

if ! grep -q -- "VITE_EI_API_MODE" "$REPO_ROOT/frontend/playwright.e2e.config.ts"; then
  echo "verify: Playwright config does not set real API mode" >&2
  exit 1
fi

for private_marker in \
  "Scenario confidential JD" \
  "I split migration risk" \
  "add tradeoff" \
  "prompt body" \
  "response body" \
  "provider-secret"; do
  if grep -Fq -- "$private_marker" "$LOG"; then
    echo "verify: private marker leaked into trigger log: $private_marker" >&2
    exit 1
  fi
done

if [ -s "$SETUP_MARKER" ]; then
  for misplaced_dir in "$REPO_ROOT/frontend/.playwright-output" "$REPO_ROOT/frontend/test-results"; do
    if [ -d "$misplaced_dir" ] && find "$misplaced_dir" -type f -newer "$SETUP_MARKER" | grep -q .; then
      echo "verify: Playwright artifacts written outside scenario output: $misplaced_dir" >&2
      exit 1
    fi
  done
fi

python3 - "$REPO_ROOT" <<'PY'
import pathlib
import re
import sys

root = pathlib.Path(sys.argv[1])
pattern = re.compile(r"""route-(welcome|growth|mistakes|drill|followup|experiences|star(_editor)?|onboarding|voice)\b|/[#]?(welcome|growth|mistakes|drill|followup|experiences|star|onboarding|voice)([\s'"?/#:]|$)|#route=(welcome|growth|mistakes|drill|followup|experiences|star|onboarding|plan|resume|voice)([\s'"'/#?&=:-]|$)|\b(name|route)\s*[:=]\s*['"](welcome|growth|mistakes|drill|followup|experiences|star|onboarding|plan|resume|voice)['"]|\bmode\s*=\s*['"]debrief['"]|mode=debrief""")
d22_retired_ui = re.compile(
    r"""DebriefScreen|ProfileScreen|topbar-nav-debrief|topbar-user-profile|"""
    r"""home-aux-debrief|debriefId|debriefJobId|SET_DEBRIEF_CONTEXT|"""
    r"""createDebrief|suggestDebriefQuestions|getDebrief|Debriefs/|Profile/"""
)
allowed = [
    "startPracticeSession",
    "createPracticePlan",
    "practice_plans",
    "resumeId",
    "resumes",
    "/api/v1/practice/sessions/{sessionId}/voice-turns",
]
for sample in allowed:
    if pattern.search(sample):
        print(f"verify: legacy regex falsely matched canonical token {sample}", file=sys.stderr)
        sys.exit(1)

scan_paths = [
    "frontend/src/app/routes.ts",
    "frontend/src/app/routeUrl.ts",
    "frontend/src/app/navigation",
    "frontend/src/app/interview-context",
    "frontend/src/app/screens/home",
    "frontend/src/app/screens/parse",
    "frontend/src/app/screens/workspace",
    "frontend/src/app/screens/practice",
    "frontend/src/app/screens/generating",
    "frontend/src/app/screens/report",
    "frontend/tests/e2e/full-funnel-journey.spec.ts",
]
for rel in scan_paths:
    path = root / rel
    files = [path] if path.is_file() else [p for p in path.rglob("*") if p.is_file()]
    for file_path in files:
        if "__tests__" in file_path.parts or ".test." in file_path.name:
            continue
        text = file_path.read_text(encoding="utf-8", errors="ignore")
        if pattern.search(text):
            print(f"verify: legacy route vocabulary found in {file_path.relative_to(root)}", file=sys.stderr)
            sys.exit(1)
        if d22_retired_ui.search(text):
            print(f"verify: D-22 retired debrief/profile UI token found in {file_path.relative_to(root)}", file=sys.stderr)
            sys.exit(1)
PY

echo "verify: ok"
