#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
SCENARIO_ID="$(basename "$(dirname "$SCRIPT_DIR")")"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/$SCENARIO_ID"
mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT"
  "$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-gate.sh" "$REPO_ROOT"
  pnpm --filter @easyinterview/frontend exec vitest run --reporter=verbose \
    src/app/screens/resume-workshop/create/ResumeCreateFlow.test.tsx \
    src/app/screens/resume-workshop/create/UploadTab.test.tsx \
    src/app/screens/resume-workshop/create/PreviewStage.test.tsx \
    src/app/screens/resume-workshop/create/hooks/useResumePresignUpload.test.tsx \
    src/app/screens/resume-workshop/create/hooks/useResumeRegistration.test.tsx \
    src/app/screens/resume-workshop/create/ParsingStage.test.tsx \
    src/app/screens/resume-workshop/create/CreateFlowLegacyNegative.test.ts
) | tee "$OUTPUT_DIR/trigger.log"
