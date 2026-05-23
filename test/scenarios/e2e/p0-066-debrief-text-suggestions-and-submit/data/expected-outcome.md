# Expected Outcome

- Trigger log includes the `E2E.P0.066 RUNNER pnpm vitest` marker.
- Frontend real backend gate evidence is present for debrief owner generated-client calls.
- Debrief text suggestion, guided/manual/voice record, submit CTA, reducer, auth pending-action, and privacy boundary tests run and report passed tests.
- Entry source paths such as AI-confirmed, AI-edited, and manual compile to the expected wire shape.
- Privacy tests prove debrief field names do not leak through localStorage, sessionStorage, or console logging.
- `frontend_debrief_legacy.py --phase 8.9` passes.
- Verify rejects no-test and failed-test output and prints `E2E.P0.066 PASS`.
