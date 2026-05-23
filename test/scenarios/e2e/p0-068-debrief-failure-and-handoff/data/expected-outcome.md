# Expected Outcome

- Trigger log includes the `E2E.P0.068 RUNNER pnpm vitest` marker.
- Frontend real backend gate evidence is present for debrief owner generated-client calls.
- Failure, missing-context, timeout, auth handoff, privacy, and interview context tests run and report passed tests.
- Source-level handoff gate confirms the debrief replay CTA creates a fresh practice plan and practice session with `sourceDebriefId`.
- Trigger log includes `DEBRIEF HANDOFF SESSION GATE OK`.
- `frontend_debrief_legacy.py --phase 8.11` passes.
- Verify rejects no-test and failed-test output and prints `E2E.P0.068 PASS`.
