#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-098-full-funnel-import-to-next-round-journey"
LOG="$OUTPUT_DIR/trigger.log"

if [ ! -s "$LOG" ]; then
  echo "verify: missing trigger.log" >&2
  exit 1
fi

if grep -Eq -- "--- SKIP:|\\[no tests to run\\]|no tests to run|0 tests|0 passed" "$LOG"; then
  echo "verify: scenario log contains skip/no-test marker" >&2
  exit 1
fi

if grep -Eq -- "--- FAIL:|^FAIL$|^FAIL[[:space:]]" "$LOG"; then
  echo "verify: scenario log contains fail marker" >&2
  exit 1
fi

for marker in \
  "--- PASS: TestE2EP0098FullFunnelImportToNextRound" \
  "--- PASS: TestE2EP0098CreatePracticePlanAcceptsEmptyFocusCodes" \
  "--- PASS: TestE2EP0098FullFunnelOutOfScopeNegativeRoutePattern" \
  "job_type=resume_parse" \
  "job_type=target_import" \
  "job_type=report_generate" \
  "outcome=succeeded"; do
  if ! grep -q -- "$marker" "$LOG"; then
    echo "verify: missing marker $marker" >&2
    exit 1
  fi
done

if ! grep -Eq -- "^ok[[:space:]]+github.com/monshunter/easyinterview/backend/cmd/api[[:space:]]" "$LOG"; then
  echo "verify: missing package-level ok marker" >&2
  exit 1
fi

for private_marker in \
  "Full funnel private JD text" \
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

python3 - "$REPO_ROOT" <<'PY'
import pathlib
import re
import sys

root = pathlib.Path(sys.argv[1])
pattern = re.compile(r"""(^|[\s'"'/#?&=:-])(welcome|growth|mistakes|drill|followup|experiences|star(_editor)?|onboarding)([\s'"'/#?&=:-]|$)|mode=debrief|name=['"](plan|resume|voice)['"]|route=['"](plan|resume|voice)['"]|#route=(plan|resume|voice)([\s'"'/#?&=:-]|$)""")
d22_out_of_scope = re.compile(
    r"""Debriefs|/debriefs\b|createDebrief|suggestDebriefQuestions|getDebrief|"""
    r"""getMyProfile|updateMyProfile|listExperienceCards|createExperienceCard|updateExperienceCard|"""
    r"""CandidateProfile|ExperienceCard|sourceDebriefId|source_debrief_id|"""
    r"""candidate_profiles|experience_cards|debrief_generate|debrief\.created|"""
    r"""debrief\.completed|profile\.update"""
)
allowed = [
    "startPracticeSession",
    "createPracticePlan",
    "practice_plans",
    "resumeId",
    "resume_assets",
    "/api/v1/practice/sessions/{sessionId}/voice-turns",
]
for sample in allowed:
    if pattern.search(sample):
        print(f"verify: out-of-scope regex falsely matched canonical token {sample}", file=sys.stderr)
        sys.exit(1)

scan_paths = [
    "backend/cmd/api/main.go",
    "backend/internal/api/generated",
    "backend/internal/api/practice",
    "backend/internal/api/reports",
    "openapi/fixtures/Jobs/getJob.json",
    "openapi/fixtures/PracticePlans/createPracticePlan.json",
    "openapi/fixtures/PracticeSessions/appendSessionEvent.json",
    "openapi/fixtures/PracticeSessions/completePracticeSession.json",
    "openapi/fixtures/PracticeSessions/startPracticeSession.json",
    "openapi/fixtures/Reports/getFeedbackReport.json",
    "openapi/fixtures/Resumes/registerResume.json",
    "openapi/fixtures/TargetJobs/getTargetJob.json",
    "openapi/fixtures/TargetJobs/importTargetJob.json",
]
for rel in scan_paths:
    path = root / rel
    files = [path] if path.is_file() else [p for p in path.rglob("*") if p.is_file()]
    for file_path in files:
        text = file_path.read_text(encoding="utf-8", errors="ignore")
        if pattern.search(text):
            print(f"verify: out-of-scope route vocabulary found in {file_path.relative_to(root)}", file=sys.stderr)
            sys.exit(1)
        if d22_out_of_scope.search(text):
            print(f"verify: D-22 out-of-scope debrief/profile token found in {file_path.relative_to(root)}", file=sys.stderr)
            sys.exit(1)
PY

echo "verify: ok"
