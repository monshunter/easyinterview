# E2E.P0.080 Seed Input

- E2E.P0.074 through E2E.P0.079 are expected to have passed in the same Phase 9 close-out.
- The tailor privacy fixture injects private resume summary, structured profile text, JD context, prompt body markers, match summary values, and suggested bullet values.
- Live store integration uses the same `resume_tailor_runs`, `resume_version_suggestions`, `ai_task_runs`, and `outbox_events` tables as the backend resume runtime.
- Retired vocabulary searches are scoped to `backend/internal/resume/` and exclude scenario `verify.sh` files to avoid matching the regression script itself.
